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

// LoadBalancer distributes requests across multiple backend nodes using weighted round-robin
// with user affinity to improve cache hit rates.
type LoadBalancer struct {
	rdb         *redis.Client
	logger      *zap.Logger
	nodes       []*LBNode
	affinityTTL time.Duration
	mu          sync.RWMutex
}

// LBNode represents a load balancer node.
type LBNode struct {
	Provider      Provider
	Name          string
	ModelPatterns []string
	ID            int64
	Weight        int
	MaxConn       int
	active        atomic.Int64
	healthy       atomic.Bool
}

// NewLoadBalancer creates a new load balancer.
func NewLoadBalancer(rdb *redis.Client, logger *zap.Logger) *LoadBalancer {
	return &LoadBalancer{
		rdb:         rdb,
		logger:      logger,
		affinityTTL: 1 * time.Hour,
	}
}

// UpdateNodes replaces the backend node list.
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
	lb.logger.Info("load balancer nodes updated", zap.Strings("nodes", names))
}

// SelectProvider selects a backend Provider for the given user and model.
// Strategy: user affinity first, then weighted selection.
func (lb *LoadBalancer) SelectProvider(ctx context.Context, userID int64, modelName string) (Provider, error) {
	lb.mu.RLock()
	candidates := lb.filterByModel(modelName)
	lb.mu.RUnlock()

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available backend node for model '%s'", modelName)
	}

	if node := lb.getAffinityNode(ctx, userID, candidates); node != nil {
		node.active.Add(1)
		return &lbProviderWrapper{node: node, provider: node.Provider}, nil
	}

	node := lb.weightedSelect(candidates)
	if node == nil {
		return nil, fmt.Errorf("all backend nodes unavailable")
	}

	lb.setAffinity(ctx, userID, node.ID)

	node.active.Add(1)
	return &lbProviderWrapper{node: node, provider: node.Provider}, nil
}

// ReleaseConnection decrements the active connection count for a node.
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

// GetNodeStats returns active connection counts for each node.
func (lb *LoadBalancer) GetNodeStats() map[int64]int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	stats := make(map[int64]int64, len(lb.nodes))
	for _, n := range lb.nodes {
		stats[n.ID] = n.active.Load()
	}
	return stats
}

// SetNodeHealth sets the health status of a node.
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

// IsNodeHealthy returns whether a node is healthy.
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

// NodeCount returns the current number of nodes.
func (lb *LoadBalancer) NodeCount() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return len(lb.nodes)
}

// filterByModel filters healthy nodes that support the specified model.
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
			lb.rdb.Expire(ctx, key, lb.affinityTTL)
			return n
		}
	}
	return nil
}

// setAffinity records user affinity to a node.
func (lb *LoadBalancer) setAffinity(ctx context.Context, userID int64, nodeID int64) {
	if lb.rdb == nil {
		return
	}
	key := fmt.Sprintf("codemind:lb:affinity:%d", userID)
	lb.rdb.Set(ctx, key, fmt.Sprintf("%d", nodeID), lb.affinityTTL)
}

// weightedSelect performs weighted selection based on configured weight and current load.
func (lb *LoadBalancer) weightedSelect(candidates []*LBNode) *LBNode {
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

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
		w := float64(n.Weight) * (1 - loadRatio*0.7)
		if w < 1 {
			w = 1
		}
		items = append(items, scored{node: n, weight: w})
		totalWeight += w
	}

	r := rand.Float64() * totalWeight
	for _, item := range items {
		r -= item.weight
		if r <= 0 {
			return item.node
		}
	}

	return items[len(items)-1].node
}

// matchesPatterns checks if a model name matches any of the patterns.
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

// lbProviderWrapper wraps a Provider to release connection count when request completes.
type lbProviderWrapper struct {
	node     *LBNode
	provider Provider
}

func (w *lbProviderWrapper) Name() string           { return w.provider.Name() }
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

func (w *lbProviderWrapper) CompletionRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	defer w.release()
	return w.provider.CompletionRaw(ctx, rawBody)
}

func (w *lbProviderWrapper) CompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	body, err := w.provider.CompletionStreamRaw(ctx, rawBody)
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

func (w *lbProviderWrapper) RetrieveModel(ctx context.Context, modelID string) (*ModelInfo, error) {
	defer w.release()
	return w.provider.RetrieveModel(ctx, modelID)
}

func (w *lbProviderWrapper) EmbeddingRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	defer w.release()
	return w.provider.EmbeddingRaw(ctx, rawBody)
}

func (w *lbProviderWrapper) ResponsesRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	defer w.release()
	return w.provider.ResponsesRaw(ctx, rawBody)
}

func (w *lbProviderWrapper) ResponsesStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	body, err := w.provider.ResponsesStreamRaw(ctx, rawBody)
	if err != nil {
		w.release()
		return nil, err
	}
	return &releaseOnCloseReader{ReadCloser: body, release: w.release}, nil
}

func (w *lbProviderWrapper) AnthropicMessagesRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	defer w.release()
	return w.provider.AnthropicMessagesRaw(ctx, rawBody)
}

func (w *lbProviderWrapper) AnthropicMessagesStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	body, err := w.provider.AnthropicMessagesStreamRaw(ctx, rawBody)
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

// releaseOnCloseReader wraps io.ReadCloser to release connection count on Close.
// Uses sync.Once to ensure release is called exactly once even if Close is called multiple times.
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
