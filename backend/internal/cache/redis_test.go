package cache

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// These tests need a Redis server. Set REDIS_TEST_ADDR (default localhost:6379);
// they are skipped when Redis is not reachable.
func newTestCache(t *testing.T) *RedisCache {
	t.Helper()
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	c, err := NewRedisCache(addr, "", 15)
	if err != nil {
		t.Skipf("redis not available: %v", err)
	}
	t.Cleanup(func() { _ = c.client.FlushDB(context.Background()).Err(); _ = c.Close() })
	_ = c.client.FlushDB(context.Background()).Err()
	return c
}

func TestSetGetDelete(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	type item struct{ N int }
	if err := c.Set(ctx, "k", item{7}, time.Minute); err != nil {
		t.Fatal(err)
	}
	var got item
	if err := c.Get(ctx, "k", &got); err != nil || got.N != 7 {
		t.Fatalf("got %v err %v", got, err)
	}
	if err := c.Delete(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(ctx, "k", &got); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("want miss, got %v", err)
	}
	hits, misses, rate := c.Stats.Snapshot()
	if hits != 1 || misses != 1 || rate != 0.5 {
		t.Fatalf("stats = %d %d %f", hits, misses, rate)
	}
	if err := c.Ping(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDeletePattern(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	_ = c.Set(ctx, TasksKey("a"), 1, time.Minute)
	_ = c.Set(ctx, TasksKey("b"), 2, time.Minute)
	_ = c.Set(ctx, UserKey("u"), 3, time.Minute)
	if err := c.DeletePattern(ctx, TasksPrefix+"*"); err != nil {
		t.Fatal(err)
	}
	var v int
	if err := c.Get(ctx, TasksKey("a"), &v); !errors.Is(err, ErrCacheMiss) {
		t.Fatal("tasks key survived")
	}
	if err := c.Get(ctx, UserKey("u"), &v); err != nil || v != 3 {
		t.Fatal("user key was deleted")
	}
	// nothing left to delete is fine
	if err := c.DeletePattern(ctx, "nothing:*"); err != nil {
		t.Fatal(err)
	}
}

func TestErrors(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	if err := c.Set(ctx, "bad", make(chan int), time.Minute); err == nil {
		t.Fatal("unmarshalable value accepted")
	}
	_ = c.client.Set(ctx, "notjson", "{", time.Minute).Err()
	var v map[string]int
	if err := c.Get(ctx, "notjson", &v); err == nil || errors.Is(err, ErrCacheMiss) {
		t.Fatalf("want unmarshal error, got %v", err)
	}
	cctx, cancel := context.WithCancel(ctx)
	cancel()
	if err := c.Get(cctx, "k", &v); err == nil || errors.Is(err, ErrCacheMiss) {
		t.Fatalf("want context error, got %v", err)
	}
	if err := c.DeletePattern(cctx, "*"); err == nil {
		t.Fatal("want context error from DeletePattern")
	}
}

func TestConnectFailure(t *testing.T) {
	if _, err := NewRedisCache("127.0.0.1:1", "", 0); err == nil {
		t.Fatal("expected connection error")
	}
}

func TestStatsEmptyAndKeys(t *testing.T) {
	var s Stats
	if _, _, r := s.Snapshot(); r != 0 {
		t.Fatal("empty stats should have 0 rate")
	}
	if TaskKey("1") != "task:1" || UserKey("2") != "user:2" || TasksKey("x") != "tasks:x" {
		t.Fatal("key helpers changed")
	}
}
