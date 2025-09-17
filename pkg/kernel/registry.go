package kernel

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// DiscoveryProvider 定义节点配置发现能力。
type DiscoveryProvider interface {
	Name() string
	FetchNodeConfig(ctx context.Context, nodeID string) (NodeConfig, error)
	Close() error
}

// Registry 维护协议到 Provider 的映射。
type Registry struct {
	mu              sync.RWMutex
	defaultProtocol string
	providers       map[string]DiscoveryProvider
}

// NewRegistry 根据配置初始化 Provider 注册表。
func NewRegistry(opts Options) (*Registry, error) {
	providers := make(map[string]DiscoveryProvider)

	if opts.HTTP.BaseURL != "" {
		httpProvider, err := NewHTTPProvider(opts.HTTP)
		if err != nil {
			return nil, fmt.Errorf("init http provider: %w", err)
		}
		providers["http"] = httpProvider
	}

	if opts.GRPC.Endpoint != "" {
		grpcProvider, err := NewGRPCDiscoveryProvider(opts.GRPC)
		if err != nil {
			return nil, fmt.Errorf("init grpc provider: %w", err)
		}
		providers["grpc"] = grpcProvider
	}

	registry := &Registry{
		providers: providers,
	}

	registry.defaultProtocol = registry.chooseDefault(strings.ToLower(opts.DefaultProtocol))

	return registry, nil
}

func (r *Registry) chooseDefault(candidate string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if candidate != "" {
		if _, ok := r.providers[candidate]; ok {
			return candidate
		}
	}

	if _, ok := r.providers["http"]; ok {
		return "http"
	}

	for protocol := range r.providers {
		return protocol
	}

	return ""
}

// Provider 返回指定协议的发现实现。
func (r *Registry) Provider(protocol string) (DiscoveryProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	proto := strings.ToLower(protocol)
	if proto == "" {
		proto = r.defaultProtocol
	}

	provider, ok := r.providers[proto]
	if !ok {
		return nil, ErrProviderNotFound
	}

	return provider, nil
}

// DefaultProtocol 返回默认协议。
func (r *Registry) DefaultProtocol() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.defaultProtocol
}

// Protocols 返回当前可用协议列表。
func (r *Registry) Protocols() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	protocols := make([]string, 0, len(r.providers))
	for protocol := range r.providers {
		protocols = append(protocols, protocol)
	}

	sort.Strings(protocols)
	return protocols
}

// Close 关闭所有 Provider。
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var firstErr error
	for key, provider := range r.providers {
		if provider == nil {
			continue
		}
		if err := provider.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close provider %s: %w", key, err)
		}
	}

	r.providers = nil
	return firstErr
}
