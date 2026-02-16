package llm

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// ProviderManager 多 Provider 管理器
// 负责 Provider 注册、模型路由和默认 Provider 选择
type ProviderManager struct {
	mu              sync.RWMutex
	providers       map[string]Provider        // name -> Provider
	modelRoutes     map[string]string          // pattern -> provider name
	defaultProvider string                     // 默认 Provider 名称
}

// NewProviderManager 创建 Provider 管理器
func NewProviderManager(defaultProvider string) *ProviderManager {
	return &ProviderManager{
		providers:       make(map[string]Provider),
		modelRoutes:     make(map[string]string),
		defaultProvider: defaultProvider,
	}
}

// Register 注册一个 Provider
func (m *ProviderManager) Register(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// SetModelRoutes 设置模型路由规则
// routes 格式: map["claude-*"] = "anthropic-cloud"
func (m *ProviderManager) SetModelRoutes(routes map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelRoutes = routes
}

// GetProvider 根据名称获取 Provider
func (m *ProviderManager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("Provider '%s' 未注册", name)
	}
	return p, nil
}

// GetDefault 获取默认 Provider
func (m *ProviderManager) GetDefault() (Provider, error) {
	return m.GetProvider(m.defaultProvider)
}

// RouteByModel 根据模型名称路由到合适的 Provider
// 匹配规则支持通配符: "claude-*" 匹配所有以 "claude-" 开头的模型
func (m *ProviderManager) RouteByModel(modelName string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 精确匹配优先
	if providerName, ok := m.modelRoutes[modelName]; ok {
		if p, ok := m.providers[providerName]; ok {
			return p, nil
		}
	}

	// 通配符匹配
	for pattern, providerName := range m.modelRoutes {
		if pattern == "*" {
			continue // 通配符 "*" 最后处理
		}
		matched, err := filepath.Match(pattern, modelName)
		if err == nil && matched {
			if p, ok := m.providers[providerName]; ok {
				return p, nil
			}
		}
	}

	// 全局通配符
	if providerName, ok := m.modelRoutes["*"]; ok {
		if p, ok := m.providers[providerName]; ok {
			return p, nil
		}
	}

	// 回退到默认 Provider
	return m.GetDefault()
}

// ListProviders 列出所有已注册的 Provider 名称
func (m *ProviderManager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var names []string
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// GetProviderByFormat 根据格式获取第一个匹配的 Provider
func (m *ProviderManager) GetProviderByFormat(format ProviderFormat) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.providers {
		if p.Format() == format {
			return p, nil
		}
	}
	return nil, fmt.Errorf("没有 '%s' 格式的 Provider", format)
}

// GetDefaultProviderName 返回默认 Provider 名称
func (m *ProviderManager) GetDefaultProviderName() string {
	return m.defaultProvider
}

// HasProvider 检查是否存在指定名称的 Provider
func (m *ProviderManager) HasProvider(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.providers[name]
	return ok
}

// ProviderInfo Provider 简要信息（用于日志和调试）
type ProviderInfo struct {
	Name   string         `json:"name"`
	Format ProviderFormat `json:"format"`
}

// ListProviderInfo 列出所有 Provider 的详细信息
func (m *ProviderManager) ListProviderInfo() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var infos []ProviderInfo
	for _, p := range m.providers {
		infos = append(infos, ProviderInfo{
			Name:   p.Name(),
			Format: p.Format(),
		})
	}
	return infos
}

// DebugRoutes 返回路由规则的调试信息
func (m *ProviderManager) DebugRoutes() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var parts []string
	for pattern, provider := range m.modelRoutes {
		parts = append(parts, fmt.Sprintf("  %s -> %s", pattern, provider))
	}
	parts = append(parts, fmt.Sprintf("  [default] -> %s", m.defaultProvider))
	return "模型路由规则:\n" + strings.Join(parts, "\n")
}
