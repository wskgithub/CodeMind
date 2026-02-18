package llm

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// LoadBalancer 多后端负载均衡器
// 采用加权轮询 + 用户亲和性策略，在保证负载均衡的同时提高缓存命中率
type LoadBalancer struct {
	nodes       []*LBNode
	rdb         *redis.Client
	logger      *zap.Logger
	mu          sync.RWMutex
	counter     atomic.Uint64
	affinityTTL time.Duration
}

// LBNode 负载均衡节点
type LBNode struct {
	ID            int64
	Name          string
	Provider      Provider
	Weight        int
	MaxConn       int
	ModelPatterns []string // 该节点支持的模型匹配模式
	active        atomic.Int64
	healthy       atomic.Bool
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(rdb *redis.Client, logger *zap.Logger) *LoadBalancer {
	return &LoadBalancer{
		rdb:         rdb,
		logger:      logger,
		affinityTTL: 1 * time.Hour,
	}
}

// UpdateNodes 更新后端节点列表（全量替换）
func (lb *LoadBalancer) UpdateNodes(nodes []*LBNode) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, n := range nodes {
		n.healthy.Store(true)
	}
	lb.nodes = nodes

	names := make([]string, len(nodes))
	for i, n := range nodes {
		names[i] = fmt.Sprintf("%s(w=%d)", n.Name, n.Weight)
	}
	lb.logger.Info("负载均衡节点已更新", zap.Strings("nodes", names))
}

// SelectProvider 为指定用户和模型选择一个后端 Provider
// 策略：用户亲和性优先 → 加权轮询
func (lb *LoadBalancer) SelectProvider(ctx context.Context, userID int64, modelName string) (Provider, error) {
	lb.mu.RLock()
	candidates := lb.filterByModel(modelName)
	lb.mu.RUnlock()

	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的后端节点处理模型 '%s'", modelName)
	}

	// 1. 尝试用户亲和性：上次分配的节点
	if node := lb.getAffinityNode(ctx, userID, candidates); node != nil {
		node.active.Add(1)
		return &lbProviderWrapper{node: node, provider: node.Provider}, nil
	}

	// 2. 加权选择
	node := lb.weightedSelect(candidates)
	if node == nil {
		return nil, fmt.Errorf("所有后端节点不可用")
	}

	// 记录亲和性
	lb.setAffinity(ctx, userID, node.ID)

	node.active.Add(1)
	return &lbProviderWrapper{node: node, provider: node.Provider}, nil
}

// ReleaseConnection 释放连接计数
func (lb *LoadBalancer) ReleaseConnection(nodeID int64) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	for _, n := range lb.nodes {
		if n.ID == nodeID {
			if n.active.Add(-1) < 0 {
				n.active.Store(0)
			}
			return
		}
	}
}

// GetNodeStats 获取各节点的活跃连接数
func (lb *LoadBalancer) GetNodeStats() map[int64]int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	stats := make(map[int64]int64, len(lb.nodes))
	for _, n := range lb.nodes {
		stats[n.ID] = n.active.Load()
	}
	return stats
}

// SetNodeHealth 设置节点健康状态
func (lb *LoadBalancer) SetNodeHealth(nodeID int64, healthy bool) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	for _, n := range lb.nodes {
		if n.ID == nodeID {
			n.healthy.Store(healthy)
			return
		}
	}
}

// IsNodeHealthy 查询节点是否健康
func (lb *LoadBalancer) IsNodeHealthy(nodeID int64) bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	for _, n := range lb.nodes {
		if n.ID == nodeID {
			return n.healthy.Load()
		}
	}
	return false
}

// NodeCount 返回当前节点数量
func (lb *LoadBalancer) NodeCount() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return len(lb.nodes)
}

// filterByModel 筛选支持指定模型的健康节点
func (lb *LoadBalancer) filterByModel(modelName string) []*LBNode {
	var result []*LBNode
	for _, n := range lb.nodes {
		if !n.healthy.Load() {
			continue
		}
		if n.active.Load() >= int64(n.MaxConn) {
			continue
		}
		if matchesPatterns(n.ModelPatterns, modelName) {
			result = append(result, n)
		}
	}
	return result
}

// getAffinityNode 获取用户亲和性节点
func (lb *LoadBalancer) getAffinityNode(ctx context.Context, userID int64, candidates []*LBNode) *LBNode {
	if lb.rdb == nil {
		return nil
	}

	key := fmt.Sprintf("codemind:lb:affinity:%d", userID)
	nodeIDStr, err := lb.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var nodeID int64
	if _, err := fmt.Sscanf(nodeIDStr, "%d", &nodeID); err != nil {
		return nil
	}

	for _, n := range candidates {
		if n.ID == nodeID && n.healthy.Load() && n.active.Load() < int64(n.MaxConn) {
			// 续期亲和性
			lb.rdb.Expire(ctx, key, lb.affinityTTL)
			return n
		}
	}
	return nil
}

// setAffinity 记录用户亲和性
func (lb *LoadBalancer) setAffinity(ctx context.Context, userID int64, nodeID int64) {
	if lb.rdb == nil {
		return
	}
	key := fmt.Sprintf("codemind:lb:affinity:%d", userID)
	lb.rdb.Set(ctx, key, fmt.Sprintf("%d", nodeID), lb.affinityTTL)
}

// weightedSelect 加权选择算法
// 综合权重和当前负载进行选择，负载越低被选中概率越高
func (lb *LoadBalancer) weightedSelect(candidates []*LBNode) *LBNode {
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	// 计算每个节点的有效权重：配置权重 × (1 - 当前负载率)
	type scored struct {
		node   *LBNode
		weight float64
	}
	var items []scored
	var totalWeight float64

	for _, n := range candidates {
		loadRatio := float64(n.active.Load()) / float64(n.MaxConn)
		if loadRatio > 1 {
			loadRatio = 1
		}
		w := float64(n.Weight) * (1 - loadRatio*0.7) // 负载因子最多降低 70% 权重
		if w < 1 {
			w = 1
		}
		items = append(items, scored{node: n, weight: w})
		totalWeight += w
	}

	// 加权随机选择
	r := rand.Float64() * totalWeight
	for _, item := range items {
		r -= item.weight
		if r <= 0 {
			return item.node
		}
	}

	return items[len(items)-1].node
}

// matchesPatterns 检查模型名是否匹配任一模式
func matchesPatterns(patterns []string, modelName string) bool {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "*" {
			return true
		}
		if matched, _ := filepath.Match(p, modelName); matched {
			return true
		}
	}
	return false
}

// lbProviderWrapper 包装 Provider 以在请求完成时释放连接计数
type lbProviderWrapper struct {
	node     *LBNode
	provider Provider
}

func (w *lbProviderWrapper) Name() string          { return w.provider.Name() }
func (w *lbProviderWrapper) Format() ProviderFormat { return w.provider.Format() }

func (w *lbProviderWrapper) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	defer w.release()
	return w.provider.ChatCompletion(ctx, req)
}

func (w *lbProviderWrapper) ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (io.ReadCloser, error) {
	body, err := w.provider.ChatCompletionStream(ctx, req)
	if err != nil {
		w.release()
		return nil, err
	}
	return &releaseOnCloseReader{ReadCloser: body, release: w.release}, nil
}

func (w *lbProviderWrapper) ChatCompletionRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	defer w.release()
	return w.provider.ChatCompletionRaw(ctx, rawBody)
}

func (w *lbProviderWrapper) ChatCompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	body, err := w.provider.ChatCompletionStreamRaw(ctx, rawBody)
	if err != nil {
		w.release()
		return nil, err
	}
	return &releaseOnCloseReader{ReadCloser: body, release: w.release}, nil
}

func (w *lbProviderWrapper) Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	defer w.release()
	return w.provider.Completion(ctx, req)
}

func (w *lbProviderWrapper) CompletionStream(ctx context.Context, req *CompletionRequest) (io.ReadCloser, error) {
	body, err := w.provider.CompletionStream(ctx, req)
	if err != nil {
		w.release()
		return nil, err
	}
	return &releaseOnCloseReader{ReadCloser: body, release: w.release}, nil
}

func (w *lbProviderWrapper) ListModels(ctx context.Context) (*ModelListResponse, error) {
	defer w.release()
	return w.provider.ListModels(ctx)
}

func (w *lbProviderWrapper) AnthropicMessages(ctx context.Context, req *AnthropicMessagesRequest) (*AnthropicMessagesResponse, error) {
	defer w.release()
	return w.provider.AnthropicMessages(ctx, req)
}

func (w *lbProviderWrapper) AnthropicMessagesStream(ctx context.Context, req *AnthropicMessagesRequest) (io.ReadCloser, error) {
	body, err := w.provider.AnthropicMessagesStream(ctx, req)
	if err != nil {
		w.release()
		return nil, err
	}
	return &releaseOnCloseReader{ReadCloser: body, release: w.release}, nil
}

func (w *lbProviderWrapper) release() {
	if w.node.active.Add(-1) < 0 {
		w.node.active.Store(0)
	}
}

// releaseOnCloseReader 包装 io.ReadCloser，在 Close 时自动释放负载均衡连接计数。
// 通过 sync.Once 保证 release 只执行一次，即使 Close 被多次调用也安全。
type releaseOnCloseReader struct {
	io.ReadCloser
	release func()
	once    sync.Once
}

func (r *releaseOnCloseReader) Close() error {
	err := r.ReadCloser.Close()
	r.once.Do(r.release)
	return err
}
