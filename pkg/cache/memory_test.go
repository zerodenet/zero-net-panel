package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	c, err := New(Config{Provider: "memory"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	type payload struct {
		Value string
	}

	if err := c.Set(ctx, "key", payload{Value: "hello"}, time.Second); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	var out payload
	if err := c.Get(ctx, "key", &out); err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if out.Value != "hello" {
		t.Fatalf("unexpected value: %v", out.Value)
	}

	time.Sleep(1100 * time.Millisecond)

	if err := c.Get(ctx, "key", &out); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
