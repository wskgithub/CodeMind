package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// ═══════════════════════════════════════════════════════════════
// Mock Provider Implementation
// ═══════════════════════════════════════════════════════════════

type mockProvider struct {
	mock.Mock
	name   string
	format ProviderFormat
}

func newMockProvider(name string, format ProviderFormat) *mockProvider {
	return &mockProvider{
		name:   name,
		format: format,
	}
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Format() ProviderFormat {
	return m.format
}

func (m *mockProvider) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	args := m.Called(ctx, req)
	if resp, ok := args.Get(0).(*ChatCompletionResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (io.ReadCloser, error) {
	args := m.Called(ctx, req)
	if rc, ok := args.Get(0).(io.ReadCloser); ok {
		return rc, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) ChatCompletionRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	args := m.Called(ctx, rawBody)
	if data, ok := args.Get(0).([]byte); ok {
		if usage, ok2 := args.Get(1).(*Usage); ok2 {
			return data, usage, args.Error(2)
		}
		return data, nil, args.Error(2)
	}
	return nil, nil, args.Error(2)
}

func (m *mockProvider) ChatCompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	args := m.Called(ctx, rawBody)
	if rc, ok := args.Get(0).(io.ReadCloser); ok {
		return rc, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	args := m.Called(ctx, req)
	if resp, ok := args.Get(0).(*CompletionResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) CompletionStream(ctx context.Context, req *CompletionRequest) (io.ReadCloser, error) {
	args := m.Called(ctx, req)
	if rc, ok := args.Get(0).(io.ReadCloser); ok {
		return rc, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) CompletionRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	args := m.Called(ctx, rawBody)
	if data, ok := args.Get(0).([]byte); ok {
		if usage, ok2 := args.Get(1).(*Usage); ok2 {
			return data, usage, args.Error(2)
		}
		return data, nil, args.Error(2)
	}
	return nil, nil, args.Error(2)
}

func (m *mockProvider) CompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	args := m.Called(ctx, rawBody)
	if rc, ok := args.Get(0).(io.ReadCloser); ok {
		return rc, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) ListModels(ctx context.Context) (*ModelListResponse, error) {
	args := m.Called(ctx)
	if resp, ok := args.Get(0).(*ModelListResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) RetrieveModel(ctx context.Context, modelID string) (*ModelInfo, error) {
	args := m.Called(ctx, modelID)
	if info, ok := args.Get(0).(*ModelInfo); ok {
		return info, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) EmbeddingRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	args := m.Called(ctx, rawBody)
	if data, ok := args.Get(0).([]byte); ok {
		if usage, ok2 := args.Get(1).(*Usage); ok2 {
			return data, usage, args.Error(2)
		}
		return data, nil, args.Error(2)
	}
	return nil, nil, args.Error(2)
}

func (m *mockProvider) ResponsesRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	args := m.Called(ctx, rawBody)
	if data, ok := args.Get(0).([]byte); ok {
		if usage, ok2 := args.Get(1).(*Usage); ok2 {
			return data, usage, args.Error(2)
		}
		return data, nil, args.Error(2)
	}
	return nil, nil, args.Error(2)
}

func (m *mockProvider) ResponsesStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	args := m.Called(ctx, rawBody)
	if rc, ok := args.Get(0).(io.ReadCloser); ok {
		return rc, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) AnthropicMessagesRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	args := m.Called(ctx, rawBody)
	if data, ok := args.Get(0).([]byte); ok {
		if usage, ok2 := args.Get(1).(*Usage); ok2 {
			return data, usage, args.Error(2)
		}
		return data, nil, args.Error(2)
	}
	return nil, nil, args.Error(2)
}

func (m *mockProvider) AnthropicMessagesStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	args := m.Called(ctx, rawBody)
	if rc, ok := args.Get(0).(io.ReadCloser); ok {
		return rc, args.Error(1)
	}
	return nil, args.Error(1)
}

// mockReadCloser 用于模拟流式响应.
type mockReadCloser struct {
	io.Reader
	onClose func()
	closed  bool
}

func newMockReadCloser(data string) *mockReadCloser {
	return &mockReadCloser{Reader: strings.NewReader(data)}
}

func newMockReadCloserWithCallback(data string, onClose func()) *mockReadCloser {
	return &mockReadCloser{Reader: strings.NewReader(data), onClose: onClose}
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	if m.onClose != nil {
		m.onClose()
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════
// ProviderManager Tests
// ═══════════════════════════════════════════════════════════════

func TestNewProviderManager(t *testing.T) {
	pm := NewProviderManager("default-provider")
	assert.NotNil(t, pm)
	assert.Equal(t, "default-provider", pm.GetDefaultProviderName())
}

func TestProviderManager_Register(t *testing.T) {
	pm := NewProviderManager("default")
	provider := newMockProvider("test-provider", FormatOpenAI)

	pm.Register(provider)

	assert.True(t, pm.HasProvider("test-provider"))
	assert.False(t, pm.HasProvider("non-existent"))
}

func TestProviderManager_GetProvider(t *testing.T) {
	pm := NewProviderManager("default")
	provider := newMockProvider("test-provider", FormatOpenAI)
	pm.Register(provider)

	tests := []struct {
		name         string
		providerName string
		errMsg       string
		wantErr      bool
	}{
		{
			name:         "existing provider",
			providerName: "test-provider",
			wantErr:      false,
		},
		{
			name:         "non-existent provider",
			providerName: "non-existent",
			wantErr:      true,
			errMsg:       "Provider 'non-existent' 未注册",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := pm.GetProvider(tt.providerName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				assert.Nil(t, p)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, p)
				assert.Equal(t, tt.providerName, p.Name())
			}
		})
	}
}

func TestProviderManager_GetDefault(t *testing.T) {
	pm := NewProviderManager("default")

	// Test when default provider is not registered
	p, err := pm.GetDefault()
	assert.Error(t, err)
	assert.Nil(t, p)

	// Register default provider
	provider := newMockProvider("default", FormatOpenAI)
	pm.Register(provider)

	p, err = pm.GetDefault()
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "default", p.Name())
}

func TestProviderManager_SetModelRoutes(t *testing.T) {
	pm := NewProviderManager("default")
	openai := newMockProvider("openai", FormatOpenAI)
	anthropic := newMockProvider("anthropic", FormatAnthropic)

	pm.Register(openai)
	pm.Register(anthropic)

	routes := map[string]string{
		"gpt-*":    "openai",
		"claude-*": "anthropic",
		"gpt-4":    "openai",
		"*":        "openai",
	}

	pm.SetModelRoutes(routes)

	// Test exact match
	p, err := pm.RouteByModel("gpt-4")
	assert.NoError(t, err)
	assert.Equal(t, "openai", p.Name())

	// Test wildcard match
	p, err = pm.RouteByModel("gpt-3.5")
	assert.NoError(t, err)
	assert.Equal(t, "openai", p.Name())

	// Test another wildcard
	p, err = pm.RouteByModel("claude-sonnet-4")
	assert.NoError(t, err)
	assert.Equal(t, "anthropic", p.Name())

	// Test global wildcard fallback
	p, err = pm.RouteByModel("unknown-model")
	assert.NoError(t, err)
	assert.Equal(t, "openai", p.Name())
}

func TestProviderManager_RouteByModel_Priority(t *testing.T) {
	pm := NewProviderManager("default")
	openai := newMockProvider("openai", FormatOpenAI)
	anthropic := newMockProvider("anthropic", FormatAnthropic)

	pm.Register(openai)
	pm.Register(anthropic)

	// Exact match should take priority over wildcard
	routes := map[string]string{
		"gpt-4": "anthropic", // exact match
		"gpt-*": "openai",    // wildcard
		"*":     "default",
	}
	pm.SetModelRoutes(routes)

	p, err := pm.RouteByModel("gpt-4")
	assert.NoError(t, err)
	assert.Equal(t, "anthropic", p.Name()) // exact match priority

	p, err = pm.RouteByModel("gpt-3.5")
	assert.NoError(t, err)
	assert.Equal(t, "openai", p.Name()) // wildcard match
}

func TestProviderManager_RouteByModel_FallbackToDefault(t *testing.T) {
	pm := NewProviderManager("default")
	defaultProvider := newMockProvider("default", FormatOpenAI)
	pm.Register(defaultProvider)

	// No routes set, should fallback to default
	p, err := pm.RouteByModel("any-model")
	assert.NoError(t, err)
	assert.Equal(t, "default", p.Name())
}

func TestProviderManager_ListProviders(t *testing.T) {
	pm := NewProviderManager("default")

	// Empty list
	names := pm.ListProviders()
	assert.Empty(t, names)

	// Register providers
	pm.Register(newMockProvider("provider1", FormatOpenAI))
	pm.Register(newMockProvider("provider2", FormatAnthropic))
	pm.Register(newMockProvider("provider3", FormatOpenAI))

	names = pm.ListProviders()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "provider1")
	assert.Contains(t, names, "provider2")
	assert.Contains(t, names, "provider3")
}

func TestProviderManager_GetProviderByFormat(t *testing.T) {
	pm := NewProviderManager("default")
	openai := newMockProvider("openai", FormatOpenAI)
	anthropic := newMockProvider("anthropic", FormatAnthropic)

	pm.Register(openai)
	pm.Register(anthropic)

	// Test finding OpenAI format
	p, err := pm.GetProviderByFormat(FormatOpenAI)
	assert.NoError(t, err)
	assert.Equal(t, FormatOpenAI, p.Format())

	// Test finding Anthropic format
	p, err = pm.GetProviderByFormat(FormatAnthropic)
	assert.NoError(t, err)
	assert.Equal(t, FormatAnthropic, p.Format())

	// Test not found
	p, err = pm.GetProviderByFormat("unknown-format")
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestProviderManager_ListProviderInfo(t *testing.T) {
	pm := NewProviderManager("default")
	pm.Register(newMockProvider("openai", FormatOpenAI))
	pm.Register(newMockProvider("anthropic", FormatAnthropic))

	infos := pm.ListProviderInfo()
	assert.Len(t, infos, 2)

	infoMap := make(map[string]ProviderFormat)
	for _, info := range infos {
		infoMap[info.Name] = info.Format
	}

	assert.Equal(t, FormatOpenAI, infoMap["openai"])
	assert.Equal(t, FormatAnthropic, infoMap["anthropic"])
}

func TestProviderManager_DebugRoutes(t *testing.T) {
	pm := NewProviderManager("default")
	pm.Register(newMockProvider("openai", FormatOpenAI))
	pm.Register(newMockProvider("anthropic", FormatAnthropic))

	routes := map[string]string{
		"gpt-*":    "openai",
		"claude-*": "anthropic",
	}
	pm.SetModelRoutes(routes)

	debug := pm.DebugRoutes()
	assert.Contains(t, debug, "gpt-* -> openai")
	assert.Contains(t, debug, "claude-* -> anthropic")
	assert.Contains(t, debug, "[default] -> default")
}

func TestProviderManager_ConcurrentAccess(t *testing.T) {
	pm := NewProviderManager("default")

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent registrations with unique names
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			p := newMockProvider(fmt.Sprintf("provider-%d", idx), FormatOpenAI)
			pm.Register(p)
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	wg.Add(numGoroutines * 2)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			pm.ListProviders()
		}()
		go func(idx int) {
			defer wg.Done()
			pm.HasProvider(fmt.Sprintf("provider-%d", idx))
		}(i)
	}
	wg.Wait()

	// Should have all 100 registered providers
	assert.Equal(t, numGoroutines, len(pm.ListProviders()))
}

// ═══════════════════════════════════════════════════════════════
// LoadBalancer Tests
// ═══════════════════════════════════════════════════════════════

func TestNewLoadBalancer(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	assert.NotNil(t, lb)
	assert.Equal(t, 0, lb.NodeCount())
	assert.NotNil(t, lb.logger)
}

func TestLoadBalancer_UpdateNodes(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	provider2 := newMockProvider("provider2", FormatOpenAI)

	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
		{ID: 2, Name: "node2", Provider: provider2, Weight: 20, MaxConn: 200, ModelPatterns: []string{"gpt-*"}},
	}

	lb.UpdateNodes(nodes)

	assert.Equal(t, 2, lb.NodeCount())

	// Verify nodes are healthy after update
	assert.True(t, lb.IsNodeHealthy(1))
	assert.True(t, lb.IsNodeHealthy(2))
}

func TestLoadBalancer_SelectProvider_NoNodes(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	ctx := context.Background()
	provider, err := lb.SelectProvider(ctx, 1, "gpt-4")

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "没有可用的后端节点")
}

func TestLoadBalancer_SelectProvider_NoMatchingModel(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider, Weight: 10, MaxConn: 100, ModelPatterns: []string{"claude-*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()
	selected, err := lb.SelectProvider(ctx, 1, "gpt-4")

	assert.Error(t, err)
	assert.Nil(t, selected)
	assert.Contains(t, err.Error(), "没有可用的后端节点")
}

func TestLoadBalancer_SelectProvider_UnhealthyNode(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)
	lb.SetNodeHealth(1, false)

	ctx := context.Background()
	selected, err := lb.SelectProvider(ctx, 1, "gpt-4")

	assert.Error(t, err)
	assert.Nil(t, selected)
}

func TestLoadBalancer_SelectProvider_MaxConnReached(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider, Weight: 10, MaxConn: 1, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()

	// First selection should succeed
	selected1, err := lb.SelectProvider(ctx, 1, "gpt-4")
	assert.NoError(t, err)
	assert.NotNil(t, selected1)

	// Second selection should fail (max conn reached)
	selected2, err := lb.SelectProvider(ctx, 2, "gpt-4")
	assert.Error(t, err)
	assert.Nil(t, selected2)
}

func TestLoadBalancer_SelectProvider_Success(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()
	selected, err := lb.SelectProvider(ctx, 1, "gpt-4")

	assert.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestLoadBalancer_SelectProvider_WeightedDistribution(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	provider2 := newMockProvider("provider2", FormatOpenAI)

	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 90, MaxConn: 1000, ModelPatterns: []string{"*"}},
		{ID: 2, Name: "node2", Provider: provider2, Weight: 10, MaxConn: 1000, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()

	// Run multiple selections to verify weight distribution
	counts := map[int64]int{1: 0, 2: 0}
	iterations := 1000

	for i := 0; i < iterations; i++ {
		selected, err := lb.SelectProvider(ctx, int64(i), "gpt-4")
		if err != nil {
			continue
		}

		// Use wrapper to get the underlying provider name
		if wrapper, ok := selected.(*lbProviderWrapper); ok {
			counts[wrapper.node.ID]++
			lb.ReleaseConnection(wrapper.node.ID)
		}
	}

	// Node1 should have significantly more selections than node2
	// (90% weight vs 10% weight, though load balancing may affect this)
	assert.Greater(t, counts[1], counts[2])
	t.Logf("Distribution: node1=%d (%.1f%%), node2=%d (%.1f%%)",
		counts[1], float64(counts[1])*100/float64(iterations),
		counts[2], float64(counts[2])*100/float64(iterations))
}

func TestLoadBalancer_SelectProvider_Affinity(t *testing.T) {
	// This test requires Redis, skip if not available
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available")
	}

	logger := zap.NewNop()
	lb := NewLoadBalancer(redisClient, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()
	userID := int64(12345)

	// First selection
	selected1, err := lb.SelectProvider(ctx, userID, "gpt-4")
	assert.NoError(t, err)
	assert.NotNil(t, selected1)

	// Release and select again for same user
	lb.ReleaseConnection(1)

	// Second selection should prefer the same node (affinity)
	selected2, err := lb.SelectProvider(ctx, userID, "gpt-4")
	assert.NoError(t, err)
	assert.NotNil(t, selected2)
}

func TestLoadBalancer_ReleaseConnection(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()

	// Select and release multiple times
	for i := 0; i < 5; i++ {
		selected, err := lb.SelectProvider(ctx, int64(i), "gpt-4")
		assert.NoError(t, err)

		if wrapper, ok := selected.(*lbProviderWrapper); ok {
			lb.ReleaseConnection(wrapper.node.ID)
		}
	}

	// Stats should show 0 active connections (all released)
	stats := lb.GetNodeStats()
	assert.Equal(t, int64(0), stats[1])
}

func TestLoadBalancer_GetNodeStats(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	provider2 := newMockProvider("provider2", FormatOpenAI)

	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
		{ID: 2, Name: "node2", Provider: provider2, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()

	// Select multiple providers
	selected1, _ := lb.SelectProvider(ctx, 1, "gpt-4")
	selected2, _ := lb.SelectProvider(ctx, 2, "gpt-4")

	stats := lb.GetNodeStats()

	// Total active should be 2
	totalActive := int64(0)
	for _, count := range stats {
		totalActive += count
	}
	assert.Equal(t, int64(2), totalActive)

	// Release connections
	if wrapper, ok := selected1.(*lbProviderWrapper); ok {
		lb.ReleaseConnection(wrapper.node.ID)
	}
	if wrapper, ok := selected2.(*lbProviderWrapper); ok {
		lb.ReleaseConnection(wrapper.node.ID)
	}

	// After release, all should be 0
	stats = lb.GetNodeStats()
	for _, count := range stats {
		assert.Equal(t, int64(0), count)
	}
}

func TestLoadBalancer_SetNodeHealth(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	// Initially healthy
	assert.True(t, lb.IsNodeHealthy(1))

	// Set unhealthy
	lb.SetNodeHealth(1, false)
	assert.False(t, lb.IsNodeHealthy(1))

	// Set healthy again
	lb.SetNodeHealth(1, true)
	assert.True(t, lb.IsNodeHealthy(1))

	// Non-existent node
	lb.SetNodeHealth(999, false)
	assert.False(t, lb.IsNodeHealthy(999))
}

func TestLoadBalancer_NodeCount(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	assert.Equal(t, 0, lb.NodeCount())

	provider := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)
	assert.Equal(t, 1, lb.NodeCount())

	// Update with more nodes
	nodes = []*LBNode{
		{ID: 1, Name: "node1", Provider: provider, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
		{ID: 2, Name: "node2", Provider: provider, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)
	assert.Equal(t, 2, lb.NodeCount())
}

func TestLoadBalancer_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	provider2 := newMockProvider("provider2", FormatOpenAI)

	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 1000, ModelPatterns: []string{"*"}},
		{ID: 2, Name: "node2", Provider: provider2, Weight: 10, MaxConn: 1000, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent selections
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			userID := int64(idx)
			model := "gpt-4"

			selected, err := lb.SelectProvider(ctx, userID, model)
			if err != nil {
				return
			}

			if wrapper, ok := selected.(*lbProviderWrapper); ok {
				// Simulate some work
				time.Sleep(time.Millisecond)
				lb.ReleaseConnection(wrapper.node.ID)
			}
		}(i)
	}
	wg.Wait()

	// All connections should be released
	stats := lb.GetNodeStats()
	for _, count := range stats {
		assert.Equal(t, int64(0), count)
	}
}

// ═══════════════════════════════════════════════════════════════
// Helper Function Tests (filterByModel, matchesPatterns, etc.)
// ═══════════════════════════════════════════════════════════════

func TestMatchesPatterns(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		patterns  []string
		wantMatch bool
	}{
		{
			name:      "exact match",
			patterns:  []string{"gpt-4"},
			modelName: "gpt-4",
			wantMatch: true,
		},
		{
			name:      "wildcard all",
			patterns:  []string{"*"},
			modelName: "any-model",
			wantMatch: true,
		},
		{
			name:      "prefix wildcard",
			patterns:  []string{"gpt-*"},
			modelName: "gpt-4-turbo",
			wantMatch: true,
		},
		{
			name:      "prefix wildcard no match",
			patterns:  []string{"gpt-*"},
			modelName: "claude-3",
			wantMatch: false,
		},
		{
			name:      "multiple patterns match first",
			patterns:  []string{"gpt-*", "claude-*"},
			modelName: "gpt-4",
			wantMatch: true,
		},
		{
			name:      "multiple patterns match second",
			patterns:  []string{"gpt-*", "claude-*"},
			modelName: "claude-3-sonnet",
			wantMatch: true,
		},
		{
			name:      "no patterns",
			patterns:  []string{},
			modelName: "gpt-4",
			wantMatch: false,
		},
		{
			name:      "whitespace trimming",
			patterns:  []string{"  gpt-*  "},
			modelName: "gpt-4",
			wantMatch: true,
		},
		{
			name:      "question mark wildcard",
			patterns:  []string{"gpt-?"},
			modelName: "gpt-4",
			wantMatch: true,
		},
		{
			name:      "character class",
			patterns:  []string{"gpt-[0-9]"},
			modelName: "gpt-4",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesPatterns(tt.patterns, tt.modelName)
			assert.Equal(t, tt.wantMatch, got)
		})
	}
}

func TestLoadBalancer_FilterByModel(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	provider2 := newMockProvider("provider2", FormatOpenAI)
	provider3 := newMockProvider("provider3", FormatOpenAI)

	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100, ModelPatterns: []string{"gpt-*"}},
		{ID: 2, Name: "node2", Provider: provider2, Weight: 10, MaxConn: 100, ModelPatterns: []string{"claude-*"}},
		{ID: 3, Name: "node3", Provider: provider3, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	// Use reflection to test private method through public behavior
	ctx := context.Background()

	// Select for gpt model - should only match node1 or node3
	selected, err := lb.SelectProvider(ctx, 1, "gpt-4")
	assert.NoError(t, err)
	assert.NotNil(t, selected)

	if wrapper, ok := selected.(*lbProviderWrapper); ok {
		// Should be node1 or node3, not node2
		assert.NotEqual(t, int64(2), wrapper.node.ID)
	}
}

// ═══════════════════════════════════════════════════════════════
// lbProviderWrapper Tests
// ═══════════════════════════════════════════════════════════════

func TestLBProviderWrapper_Name(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	assert.Equal(t, "test-provider", wrapper.Name())
}

func TestLBProviderWrapper_Format(t *testing.T) {
	provider := newMockProvider("test-provider", FormatAnthropic)
	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	assert.Equal(t, FormatAnthropic, wrapper.Format())
}

func TestLBProviderWrapper_ChatCompletion(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedResp := &ChatCompletionResponse{ID: "test-123"}
	provider.On("ChatCompletion", mock.Anything, mock.Anything).Return(expectedResp, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	req := &ChatCompletionRequest{Model: "gpt-4"}

	resp, err := wrapper.ChatCompletion(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	provider.AssertExpectations(t)
}

func TestLBProviderWrapper_ChatCompletionStream_Success(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	mockBody := newMockReadCloser("stream data")
	provider.On("ChatCompletionStream", mock.Anything, mock.Anything).Return(mockBody, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	req := &ChatCompletionRequest{Model: "gpt-4"}

	body, err := wrapper.ChatCompletionStream(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, body)

	// Close should release connection
	err = body.Close()
	assert.NoError(t, err)
}

func TestLBProviderWrapper_ChatCompletionStream_Error(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	provider.On("ChatCompletionStream", mock.Anything, mock.Anything).Return(nil, errors.New("stream error"))

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	req := &ChatCompletionRequest{Model: "gpt-4"}

	body, err := wrapper.ChatCompletionStream(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestLBProviderWrapper_ListModels(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedResp := &ModelListResponse{Object: "list"}
	provider.On("ListModels", mock.Anything).Return(expectedResp, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	resp, err := wrapper.ListModels(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
}

func TestLBProviderWrapper_EmbeddingRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedData := []byte(`{"data": []}`)
	expectedUsage := &Usage{TotalTokens: 10}
	provider.On("EmbeddingRaw", mock.Anything, []byte("test")).Return(expectedData, expectedUsage, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	data, usage, err := wrapper.EmbeddingRaw(ctx, []byte("test"))

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedUsage, usage)
}

func TestLBProviderWrapper_ChatCompletionRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedData := []byte(`{"id": "test"}`)
	expectedUsage := &Usage{TotalTokens: 10}
	provider.On("ChatCompletionRaw", mock.Anything, []byte("raw")).Return(expectedData, expectedUsage, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	data, usage, err := wrapper.ChatCompletionRaw(ctx, []byte("raw"))

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedUsage, usage)
}

func TestLBProviderWrapper_ChatCompletionStreamRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	mockBody := newMockReadCloser("stream data")
	provider.On("ChatCompletionStreamRaw", mock.Anything, []byte("raw")).Return(mockBody, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	body, err := wrapper.ChatCompletionStreamRaw(ctx, []byte("raw"))

	assert.NoError(t, err)
	assert.NotNil(t, body)
	body.Close()
}

func TestLBProviderWrapper_ChatCompletionStreamRaw_Error(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	provider.On("ChatCompletionStreamRaw", mock.Anything, []byte("raw")).Return(nil, errors.New("stream error"))

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	body, err := wrapper.ChatCompletionStreamRaw(ctx, []byte("raw"))

	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestLBProviderWrapper_Completion(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedResp := &CompletionResponse{ID: "test-123"}
	provider.On("Completion", mock.Anything, mock.Anything).Return(expectedResp, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	req := &CompletionRequest{Model: "gpt-4"}

	resp, err := wrapper.Completion(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
}

func TestLBProviderWrapper_CompletionStream(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	mockBody := newMockReadCloser("stream data")
	provider.On("CompletionStream", mock.Anything, mock.Anything).Return(mockBody, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	req := &CompletionRequest{Model: "gpt-4"}

	body, err := wrapper.CompletionStream(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, body)
	body.Close()
}

func TestLBProviderWrapper_CompletionStream_Error(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	provider.On("CompletionStream", mock.Anything, mock.Anything).Return(nil, errors.New("stream error"))

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	req := &CompletionRequest{Model: "gpt-4"}

	body, err := wrapper.CompletionStream(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestLBProviderWrapper_CompletionRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedData := []byte(`{"id": "test"}`)
	expectedUsage := &Usage{TotalTokens: 10}
	provider.On("CompletionRaw", mock.Anything, []byte("raw")).Return(expectedData, expectedUsage, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	data, usage, err := wrapper.CompletionRaw(ctx, []byte("raw"))

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedUsage, usage)
}

func TestLBProviderWrapper_CompletionStreamRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	mockBody := newMockReadCloser("stream data")
	provider.On("CompletionStreamRaw", mock.Anything, []byte("raw")).Return(mockBody, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	body, err := wrapper.CompletionStreamRaw(ctx, []byte("raw"))

	assert.NoError(t, err)
	assert.NotNil(t, body)
	body.Close()
}

func TestLBProviderWrapper_CompletionStreamRaw_Error(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	provider.On("CompletionStreamRaw", mock.Anything, []byte("raw")).Return(nil, errors.New("stream error"))

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	body, err := wrapper.CompletionStreamRaw(ctx, []byte("raw"))

	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestLBProviderWrapper_RetrieveModel(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedInfo := &ModelInfo{ID: "gpt-4"}
	provider.On("RetrieveModel", mock.Anything, "gpt-4").Return(expectedInfo, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	info, err := wrapper.RetrieveModel(ctx, "gpt-4")

	assert.NoError(t, err)
	assert.Equal(t, expectedInfo, info)
}

func TestLBProviderWrapper_ResponsesRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedData := []byte(`{"id": "test"}`)
	expectedUsage := &Usage{TotalTokens: 10}
	provider.On("ResponsesRaw", mock.Anything, []byte("raw")).Return(expectedData, expectedUsage, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	data, usage, err := wrapper.ResponsesRaw(ctx, []byte("raw"))

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedUsage, usage)
}

func TestLBProviderWrapper_ResponsesStreamRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	mockBody := newMockReadCloser("stream data")
	provider.On("ResponsesStreamRaw", mock.Anything, []byte("raw")).Return(mockBody, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	body, err := wrapper.ResponsesStreamRaw(ctx, []byte("raw"))

	assert.NoError(t, err)
	assert.NotNil(t, body)
	body.Close()
}

func TestLBProviderWrapper_ResponsesStreamRaw_Error(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	provider.On("ResponsesStreamRaw", mock.Anything, []byte("raw")).Return(nil, errors.New("stream error"))

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	body, err := wrapper.ResponsesStreamRaw(ctx, []byte("raw"))

	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestLBProviderWrapper_AnthropicMessagesRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	expectedData := []byte(`{"id": "test"}`)
	expectedUsage := &Usage{TotalTokens: 10}
	provider.On("AnthropicMessagesRaw", mock.Anything, []byte("raw")).Return(expectedData, expectedUsage, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	data, usage, err := wrapper.AnthropicMessagesRaw(ctx, []byte("raw"))

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedUsage, usage)
}

func TestLBProviderWrapper_AnthropicMessagesStreamRaw(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	mockBody := newMockReadCloser("stream data")
	provider.On("AnthropicMessagesStreamRaw", mock.Anything, []byte("raw")).Return(mockBody, nil)

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	body, err := wrapper.AnthropicMessagesStreamRaw(ctx, []byte("raw"))

	assert.NoError(t, err)
	assert.NotNil(t, body)
	body.Close()
}

func TestLBProviderWrapper_AnthropicMessagesStreamRaw_Error(t *testing.T) {
	provider := newMockProvider("test-provider", FormatOpenAI)
	provider.On("AnthropicMessagesStreamRaw", mock.Anything, []byte("raw")).Return(nil, errors.New("stream error"))

	node := &LBNode{ID: 1, Name: "node1", Provider: provider}
	wrapper := &lbProviderWrapper{node: node, provider: provider}

	ctx := context.Background()
	body, err := wrapper.AnthropicMessagesStreamRaw(ctx, []byte("raw"))

	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestReleaseOnCloseReader(t *testing.T) {
	releaseCalled := false
	releaseFunc := func() {
		releaseCalled = true
	}

	mockBody := newMockReadCloserWithCallback("test data", releaseFunc)
	reader := &releaseOnCloseReader{
		ReadCloser: mockBody,
		release:    releaseFunc,
	}

	// Read data
	data, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "test data", string(data))

	// Close should trigger release
	err = reader.Close()
	assert.NoError(t, err)
	assert.True(t, releaseCalled)

	// Close again should not panic (sync.Once protection)
	err = reader.Close()
	assert.NoError(t, err)
}

// ═══════════════════════════════════════════════════════════════
// Redis Integration Tests (optional, skipped if Redis unavailable)
// ═══════════════════════════════════════════════════════════════

func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil
	}

	// Clean up test keys
	client.Del(ctx, "codemind:lb:affinity:*")

	return client
}

func TestLoadBalancer_AffinityWithRedis(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available")
	}
	defer redisClient.Close()

	logger := zap.NewNop()
	lb := NewLoadBalancer(redisClient, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	provider2 := newMockProvider("provider2", FormatOpenAI)

	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
		{ID: 2, Name: "node2", Provider: provider2, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()
	userID := int64(99999)

	// Clean up any existing affinity
	redisClient.Del(ctx, "codemind:lb:affinity:99999")

	// First selection
	selected1, err := lb.SelectProvider(ctx, userID, "gpt-4")
	assert.NoError(t, err)

	var firstNodeID int64
	if wrapper, ok := selected1.(*lbProviderWrapper); ok {
		firstNodeID = wrapper.node.ID
	}

	// Release connection
	lb.ReleaseConnection(firstNodeID)

	// Multiple subsequent selections should prefer the same node
	for i := 0; i < 5; i++ {
		selected, err := lb.SelectProvider(ctx, userID, "gpt-4")
		assert.NoError(t, err)

		if wrapper, ok := selected.(*lbProviderWrapper); ok {
			assert.Equal(t, firstNodeID, wrapper.node.ID,
				"Expected affinity to same node, got different node on iteration %d", i)
			lb.ReleaseConnection(wrapper.node.ID)
		}
	}

	// Clean up
	redisClient.Del(ctx, "codemind:lb:affinity:99999")
}

func TestLoadBalancer_SetAffinity(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available")
	}
	defer redisClient.Close()

	logger := zap.NewNop()
	lb := NewLoadBalancer(redisClient, logger)

	ctx := context.Background()
	userID := int64(88888)
	nodeID := int64(1)

	// Clean up
	redisClient.Del(ctx, "codemind:lb:affinity:88888")

	// Set affinity
	lb.setAffinity(ctx, userID, nodeID)

	// Verify in Redis
	val, err := redisClient.Get(ctx, "codemind:lb:affinity:88888").Result()
	assert.NoError(t, err)
	assert.Equal(t, "1", val)

	// Check TTL is set
	ttl, err := redisClient.TTL(ctx, "codemind:lb:affinity:88888").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 1*time.Hour)

	// Clean up
	redisClient.Del(ctx, "codemind:lb:affinity:88888")
}

func TestLoadBalancer_GetAffinityNode(t *testing.T) {
	redisClient := setupTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available")
	}
	defer redisClient.Close()

	logger := zap.NewNop()
	lb := NewLoadBalancer(redisClient, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()
	userID := int64(77777)

	// Clean up
	redisClient.Del(ctx, "codemind:lb:affinity:77777")

	// No affinity set yet
	node := lb.getAffinityNode(ctx, userID, nodes)
	assert.Nil(t, node)

	// Set affinity
	redisClient.Set(ctx, "codemind:lb:affinity:77777", "1", time.Hour)

	// Now should get the node
	node = lb.getAffinityNode(ctx, userID, nodes)
	assert.NotNil(t, node)
	assert.Equal(t, int64(1), node.ID)

	// Clean up
	redisClient.Del(ctx, "codemind:lb:affinity:77777")
}

func TestLoadBalancer_NoRedis(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger) // No Redis

	provider := newMockProvider("provider1", FormatOpenAI)
	nodes := []*LBNode{
		{ID: 1, Name: "node1", Provider: provider, Weight: 10, MaxConn: 100, ModelPatterns: []string{"*"}},
	}
	lb.UpdateNodes(nodes)

	ctx := context.Background()

	// Should work without Redis (no affinity, just weighted selection)
	selected, err := lb.SelectProvider(ctx, 1, "gpt-4")
	assert.NoError(t, err)
	assert.NotNil(t, selected)

	// setAffinity should not panic with nil Redis
	lb.setAffinity(ctx, 1, 1)
}

// ═══════════════════════════════════════════════════════════════
// Provider Implementation Tests (via mocks)
// ═══════════════════════════════════════════════════════════════

func TestOpenAIProvider_Methods(t *testing.T) {
	// These are integration tests that would require a real Client
	// For unit tests, we verify the provider interface is properly defined

	var _ Provider = (*OpenAIProvider)(nil)
	var _ Provider = (*AnthropicProvider)(nil)
}

// ═══════════════════════════════════════════════════════════════
// Weighted Selection Tests
// ═══════════════════════════════════════════════════════════════

func TestLoadBalancer_WeightedSelect(t *testing.T) {
	logger := zap.NewNop()
	lb := NewLoadBalancer(nil, logger)

	provider1 := newMockProvider("provider1", FormatOpenAI)
	provider2 := newMockProvider("provider2", FormatOpenAI)
	provider3 := newMockProvider("provider3", FormatOpenAI)

	t.Run("empty candidates", func(t *testing.T) {
		node := lb.weightedSelect([]*LBNode{})
		assert.Nil(t, node)
	})

	t.Run("single candidate", func(t *testing.T) {
		nodes := []*LBNode{
			{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100},
		}
		node := lb.weightedSelect(nodes)
		assert.NotNil(t, node)
		assert.Equal(t, int64(1), node.ID)
	})

	t.Run("equal weights distribution", func(t *testing.T) {
		nodes := []*LBNode{
			{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 1000},
			{ID: 2, Name: "node2", Provider: provider2, Weight: 10, MaxConn: 1000},
		}

		counts := map[int64]int{1: 0, 2: 0}
		iterations := 1000

		for i := 0; i < iterations; i++ {
			node := lb.weightedSelect(nodes)
			if node != nil {
				counts[node.ID]++
			}
		}

		// With equal weights, both should be selected roughly equally
		ratio := float64(counts[1]) / float64(counts[2])
		assert.InDelta(t, 1.0, ratio, 0.3, "Selection ratio should be close to 1.0 for equal weights")
	})

	t.Run("unequal weights distribution", func(t *testing.T) {
		nodes := []*LBNode{
			{ID: 1, Name: "node1", Provider: provider1, Weight: 90, MaxConn: 1000},
			{ID: 2, Name: "node2", Provider: provider2, Weight: 10, MaxConn: 1000},
		}

		counts := map[int64]int{1: 0, 2: 0}
		iterations := 1000

		for i := 0; i < iterations; i++ {
			node := lb.weightedSelect(nodes)
			if node != nil {
				counts[node.ID]++
			}
		}

		// Node 1 should be selected much more frequently
		assert.Greater(t, counts[1], counts[2])
		// Roughly 90% vs 10%
		ratio := float64(counts[1]) / float64(counts[1]+counts[2])
		assert.InDelta(t, 0.9, ratio, 0.1)
	})

	t.Run("load factor reduces weight", func(t *testing.T) {
		// Create nodes with same weight but different load
		node1 := &LBNode{ID: 1, Name: "node1", Provider: provider1, Weight: 100, MaxConn: 100}
		node1.active.Store(0) // 0% load

		node2 := &LBNode{ID: 2, Name: "node2", Provider: provider2, Weight: 100, MaxConn: 100}
		node2.active.Store(70) // 70% load

		node3 := &LBNode{ID: 3, Name: "node3", Provider: provider3, Weight: 100, MaxConn: 100}
		node3.active.Store(90) // 90% load

		nodes := []*LBNode{node1, node2, node3}

		counts := map[int64]int{1: 0, 2: 0, 3: 0}
		iterations := 1000

		for i := 0; i < iterations; i++ {
			node := lb.weightedSelect(nodes)
			if node != nil {
				counts[node.ID]++
			}
		}

		// Node 1 (0% load) should be selected most frequently
		// Node 3 (90% load) should be selected least frequently
		assert.Greater(t, counts[1], counts[2])
		assert.Greater(t, counts[2], counts[3])
	})

	t.Run("high load minimum weight", func(t *testing.T) {
		// Even at 100% load, should still get minimum weight of 1
		node1 := &LBNode{ID: 1, Name: "node1", Provider: provider1, Weight: 10, MaxConn: 100}
		node1.active.Store(100) // 100% load

		node2 := &LBNode{ID: 2, Name: "node2", Provider: provider2, Weight: 10, MaxConn: 100}
		node2.active.Store(0) // 0% load

		nodes := []*LBNode{node1, node2}

		// Node 1 should still be occasionally selected
		node1Selected := false
		for i := 0; i < 100; i++ {
			node := lb.weightedSelect(nodes)
			if node != nil && node.ID == 1 {
				node1Selected = true
				break
			}
		}

		assert.True(t, node1Selected, "Node at 100% load should still be occasionally selected")
	})
}
