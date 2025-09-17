package cache

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type redisCache struct {
	client *redis.Redis
}

type redisLock struct {
	lock *redis.RedisLock
}

func init() {
	Register("redis", func(cfg Config) (Cache, error) {
		if cfg.Redis.Host == "" {
			return nil, errors.New("cache: redis host is required")
		}

		conf := redis.RedisConf{
			Host:        cfg.Redis.Host,
			Type:        cfg.Redis.Type,
			Pass:        cfg.Redis.Password,
			Tls:         cfg.Redis.TLS,
			NonBlock:    cfg.Redis.NonBlock,
			PingTimeout: cfg.Redis.PingTimeout,
		}
		if conf.Type == "" {
			conf.Type = redis.NodeType
		}
		if conf.PingTimeout == 0 {
			conf.PingTimeout = time.Second
		}

		client, err := redis.NewRedis(conf)
		if err != nil {
			return nil, err
		}
		return &redisCache{client: client}, nil
	})
}

func (r *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	reply, err := r.client.GetCtx(ctx, key)
	if err == redis.Nil {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(reply), value)
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if ttl > 0 {
		seconds := int(math.Ceil(ttl.Seconds()))
		if seconds <= 0 {
			seconds = 1
		}
		return r.client.SetexCtx(ctx, key, string(data), seconds)
	}

	return r.client.SetCtx(ctx, key, string(data))
}

func (r *redisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	_, err := r.client.DelCtx(ctx, keys...)
	return err
}

func (r *redisCache) AcquireLock(ctx context.Context, key string, ttl time.Duration) (Lock, error) {
	lock := redis.NewRedisLock(r.client, key)
	seconds := 30
	if ttl > 0 {
		seconds = int(math.Ceil(ttl.Seconds()))
		if seconds <= 0 {
			seconds = 1
		}
	}
	lock.SetExpire(seconds)

	ok, err := lock.AcquireCtx(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}

	return &redisLock{lock: lock}, nil
}

func (r *redisCache) Close() error {
	return nil
}

func (l *redisLock) Release(ctx context.Context) error {
	_, err := l.lock.ReleaseCtx(ctx)
	return err
}
