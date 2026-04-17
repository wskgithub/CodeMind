package llm

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// ProviderManager manages multiple LLM providers with model routing.
type ProviderManager struct {
	providers       map[string]Provider
	modelRoutes     map[string]string
	defaultProvider string
	mu              sync.RWMutex
}

// NewProviderManager creates a provider manager.
func NewProviderManager(defaultProvider string) *ProviderManager {
	return &ProviderManager{
		providers:       make(map[string]Provider),
		modelRoutes:     make(map[string]string),
		defaultProvider: defaultProvider,
	}
}

// Register registers a provider.
func (m *ProviderManager) Register(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// SetModelRoutes sets model routing rules (supports wildcards like "claude-*").
func (m *ProviderManager) SetModelRoutes(routes map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelRoutes = routes
}

// GetProvider returns a provider by name.
func (m *ProviderManager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not registered", name)
	}
	return p, nil
}

// GetDefault returns the default provider.
func (m *ProviderManager) GetDefault() (Provider, error) {
	return m.GetProvider(m.defaultProvider)
}

// RouteByModel routes to the appropriate provider based on model name.
func (m *ProviderManager) RouteByModel(modelName string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if providerName, ok := m.modelRoutes[modelName]; ok {
		if p, ok := m.providers[providerName]; ok {
			return p, nil
		}
	}

	for pattern, providerName := range m.modelRoutes {
		if pattern == "*" {
			continue
		}
		matched, err := filepath.Match(pattern, modelName)
		if err == nil && matched {
			if p, ok := m.providers[providerName]; ok {
				return p, nil
			}
		}
	}

	if providerName, ok := m.modelRoutes["*"]; ok {
		if p, ok := m.providers[providerName]; ok {
			return p, nil
		}
	}

	return m.GetDefault()
}

// ListProviders returns all registered provider names.
func (m *ProviderManager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// GetProviderByFormat returns the first provider matching the format.
func (m *ProviderManager) GetProviderByFormat(format ProviderFormat) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.providers {
		if p.Format() == format {
			return p, nil
		}
	}
	return nil, fmt.Errorf("no provider with format '%s'", format)
}

// GetDefaultProviderName returns the default provider name.
func (m *ProviderManager) GetDefaultProviderName() string {
	return m.defaultProvider
}

// HasProvider checks if a provider exists.
func (m *ProviderManager) HasProvider(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.providers[name]
	return ok
}

// ProviderInfo contains provider metadata for debugging.
type ProviderInfo struct {
	Name   string         `json:"name"`
	Format ProviderFormat `json:"format"`
}

// ListProviderInfo returns info for all providers.
func (m *ProviderManager) ListProviderInfo() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(m.providers))
	for _, p := range m.providers {
		infos = append(infos, ProviderInfo{
			Name:   p.Name(),
			Format: p.Format(),
		})
	}
	return infos
}

// DebugRoutes returns routing rules for debugging.
func (m *ProviderManager) DebugRoutes() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	parts := make([]string, 0, len(m.modelRoutes)+1)
	for pattern, provider := range m.modelRoutes {
		parts = append(parts, fmt.Sprintf("  %s -> %s", pattern, provider))
	}
	parts = append(parts, fmt.Sprintf("  [default] -> %s", m.defaultProvider))
	return "Model routing rules:\n" + strings.Join(parts, "\n")
}
