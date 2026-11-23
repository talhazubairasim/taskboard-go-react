// Package cache provides a Redis-backed JSON cache with hit/miss statistics.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss is returned by Get when the key is absent.
var ErrCacheMiss = errors.New("cache miss")

// Stats counts cache lookups so the hit rate can be reported on /metrics.
type Stats struct {
	hits   atomic.Int64
	misses atomic.Int64
}

// Hit records a cache hit.
func (s *Stats) Hit() { s.hits.Add(1) }

// Miss records a cache miss.
func (s *Stats) Miss() { s.misses.Add(1) }

// Snapshot returns hits, misses and the hit rate (0..1).
func (s *Stats) Snapshot() (hits, misses int64, rate float64) {
	hits, misses = s.hits.Load(), s.misses.Load()
	if total := hits + misses; total > 0 {
		rate = float64(hits) / float64(total)
	}
	return
}

// RedisCache stores JSON values in Redis.
type RedisCache struct {
	client *redis.Client
	Stats  *Stats
}

// NewRedisCache connects to Redis and verifies the connection.
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return &RedisCache{client: client, Stats: &Stats{}}, nil
}

// Set stores a value with expiration.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves and unmarshals a value, recording a hit or miss.
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		c.Stats.Miss()
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get value: %w", err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		c.Stats.Miss()
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	c.Stats.Hit()
	return nil
}

// Delete removes keys.
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// DeletePattern deletes all keys matching a glob pattern.
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

// Ping checks the connection (used by /ready).
func (c *RedisCache) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }

// Close closes the Redis connection.
func (c *RedisCache) Close() error { return c.client.Close() }

// TasksPrefix is the key prefix for cached task listings.
const TasksPrefix = "tasks:"

// TasksKey is the key for a cached task listing identified by a filter signature.
func TasksKey(signature string) string { return TasksPrefix + signature }

// TaskKey is the key for a single cached task.
func TaskKey(id string) string { return "task:" + id }

// UserKey is the key for a single cached user.
func UserKey(id string) string { return "user:" + id }
