package syncutil_test

import (
	"sync"
	"testing"

	"github.com/flagmint/flagmint-go/internal/syncutil"
)

func TestRWValue_LoadStore(t *testing.T) {
	var v syncutil.RWValue[int]

	if got := v.Load(); got != 0 {
		t.Errorf("zero value: got %d, want 0", got)
	}

	v.Store(42)
	if got := v.Load(); got != 42 {
		t.Errorf("after store: got %d, want 42", got)
	}
}

func TestRWValue_Swap(t *testing.T) {
	var v syncutil.RWValue[string]
	v.Store("first")
	prev := v.Swap("second")
	if prev != "first" {
		t.Errorf("swap returned %q, want %q", prev, "first")
	}
	if got := v.Load(); got != "second" {
		t.Errorf("after swap load: got %q, want %q", got, "second")
	}
}

func TestRWValue_Concurrent(t *testing.T) {
	var v syncutil.RWValue[int]
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(n int) { defer wg.Done(); v.Store(n) }(i)
		go func() { defer wg.Done(); _ = v.Load() }()
	}
	wg.Wait()
}
