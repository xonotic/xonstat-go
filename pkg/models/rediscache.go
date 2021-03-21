package models

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

// RedisCache is an implementation of the caching layer in Redis.
type RedisCache struct {
	// The underlying Redis client
	client *redis.Client

	// The underlying cache store
	cache *cache.Cache
}

// NewRedisCache creates a new Redis cache instance.
func NewRedisCache(options *redis.Options) *RedisCache {
	client := redis.NewClient(options)

	cache := cache.New(&cache.Options{
		Redis:      client,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return &RedisCache{
		client: client,
		cache:  cache,
	}
}

// Get retrieves a key from the cache if it exists.
func (r *RedisCache) Get(key string) ([]byte, error) {
	ctx := context.TODO()
	var content []byte

	err := r.cache.Get(ctx, key, &content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// Set stores content in the cache for a given duration of time.
func (r *RedisCache) Set(key string, content []byte, duration time.Duration) error {
	ctx := context.TODO()

	err := r.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: content,
		TTL:   duration,
	})

	if err != nil {
		return err
	}

	return nil
}
