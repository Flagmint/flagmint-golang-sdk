package cache_test

import (
	"testing"

	"github.com/flagmint/flagmint-go/cache"
)

func TestMemoryCache(t *testing.T) {
	c := cache.NewMemoryCache()

	// Miss before insertion.
	if _, ok := c.Get("k1"); ok {
		t.Fatal("expected cache miss")
	}

	// Hit after insertion.
	c.Set("k1", "value1")
	v, ok := c.Get("k1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v != "value1" {
		t.Errorf("got %v, want value1", v)
	}

	// Delete removes the entry.
	c.Delete("k1")
	if _, ok := c.Get("k1"); ok {
		t.Fatal("expected cache miss after delete")
	}

	// Flush removes all entries.
	c.Set("a", 1)
	c.Set("b", 2)
	c.Flush()
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected cache miss after flush")
	}
}
