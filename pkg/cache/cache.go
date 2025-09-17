package cache

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("cache: key not found")

type Lock interface {
	Release(ctx context.Context) error
}

type Cache interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (Lock, error)
	Close() error
}

type Provider func(cfg Config) (Cache, error)

var providers = map[string]Provider{}

func Register(name string, provider Provider) {
	providers[name] = provider
}

func New(cfg Config) (Cache, error) {
	if cfg.Provider == "" {
		cfg.Provider = "memory"
	}

	provider, ok := providers[cfg.Provider]
	if !ok {
		return nil, errors.New("cache: unsupported provider " + cfg.Provider)
	}

	return provider(cfg)
}
