package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type memoryCache struct {
	mu    sync.RWMutex
	items map[string]memoryItem
}

type memoryItem struct {
	value    []byte
	expireAt time.Time
}

type memoryLock struct {
	key   string
	timer *time.Timer
}

var (
	memoryLocks sync.Map
)

func init() {
	Register("memory", func(cfg Config) (Cache, error) {
		return newMemoryCache(), nil
	})
}

func newMemoryCache() *memoryCache {
	return &memoryCache{items: make(map[string]memoryItem)}
}

func (m *memoryCache) Get(ctx context.Context, key string, value interface{}) error {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()

	if !ok || (item.expireAt != (time.Time{}) && time.Now().After(item.expireAt)) {
		if ok {
			m.mu.Lock()
			delete(m.items, key)
			m.mu.Unlock()
		}
		return ErrNotFound
	}

	if value == nil {
		return nil
	}

	return json.Unmarshal(item.value, value)
}

func (m *memoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var (
		data []byte
		err  error
	)

	if value == nil {
		data = []byte("null")
	} else {
		data, err = json.Marshal(value)
		if err != nil {
			return err
		}
	}

	item := memoryItem{value: data}
	if ttl > 0 {
		item.expireAt = time.Now().Add(ttl)
	}

	m.mu.Lock()
	m.items[key] = item
	m.mu.Unlock()

	return nil
}

func (m *memoryCache) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	for _, key := range keys {
		delete(m.items, key)
	}
	m.mu.Unlock()
	return nil
}

func (m *memoryCache) AcquireLock(ctx context.Context, key string, ttl time.Duration) (Lock, error) {
	for {
		if _, loaded := memoryLocks.LoadOrStore(key, struct{}{}); !loaded {
			lock := &memoryLock{key: key}
			if ttl > 0 {
				lock.timer = time.AfterFunc(ttl, func() {
					memoryLocks.Delete(key)
				})
			}
			return lock, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func (m *memoryCache) Close() error {
	m.mu.Lock()
	m.items = make(map[string]memoryItem)
	m.mu.Unlock()
	return nil
}

func (l *memoryLock) Release(ctx context.Context) error {
	if l.timer != nil {
		l.timer.Stop()
	}
	memoryLocks.Delete(l.key)
	return nil
}
