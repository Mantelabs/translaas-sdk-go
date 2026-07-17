package cache

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/acuencadev/translaas-sdk-go/models"
)

func newTestMemoryProvider(t *testing.T, opts ...MemoryOption) *memoryProvider {
	t.Helper()
	p := NewMemoryProvider(opts...).(*memoryProvider)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	current := base
	p.now = func() time.Time { return current }
	t.Cleanup(func() {
		p.now = time.Now
	})
	return p
}

func advanceProviderClock(p *memoryProvider, d time.Duration) {
	next := p.now().Add(d)
	p.now = func() time.Time { return next }
}

func TestMemoryProvider_SetAndGetString(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()

	if err := p.Set(ctx, "k1", "value1", TTL{}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	var got string
	ok, err := p.Get(ctx, "k1", &got)
	if err != nil || !ok || got != "value1" {
		t.Fatalf("Get() ok=%v got=%q err=%v", ok, got, err)
	}
}

func TestMemoryProvider_GetTranslationGroup(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()

	group := &models.TranslationGroup{
		Project: "p",
		Lang:    "en",
		Entries: map[string]json.RawMessage{"hello": json.RawMessage(`"Hi"`)},
	}
	if err := p.Set(ctx, "group", group, TTL{}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	var got models.TranslationGroup
	ok, err := p.Get(ctx, "group", &got)
	if err != nil || !ok {
		t.Fatalf("Get() ok=%v err=%v", ok, err)
	}
	if got.Project != "p" || got.Lang != "en" {
		t.Fatalf("got = %+v", got)
	}
}

func TestMemoryProvider_Miss(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	var got string
	ok, err := p.Get(context.Background(), "missing", &got)
	if err != nil || ok {
		t.Fatalf("Get() ok=%v err=%v", ok, err)
	}
}

func TestMemoryProvider_AbsoluteExpiration(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()

	if err := p.Set(ctx, "k1", "value1", TTL{Absolute: 100 * time.Millisecond}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	var got string
	ok, err := p.Get(ctx, "k1", &got)
	if err != nil || !ok || got != "value1" {
		t.Fatalf("Get() ok=%v got=%q err=%v", ok, got, err)
	}

	advanceProviderClock(p, 150*time.Millisecond)
	ok, err = p.Get(ctx, "k1", &got)
	if err != nil || ok {
		t.Fatalf("Get() after expiry ok=%v err=%v", ok, err)
	}
}

func TestMemoryProvider_SlidingExpiration(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()

	if err := p.Set(ctx, "k1", "value1", TTL{Sliding: 200 * time.Millisecond}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	var got string
	ok, err := p.Get(ctx, "k1", &got)
	if err != nil || !ok {
		t.Fatalf("initial Get() ok=%v err=%v", ok, err)
	}

	advanceProviderClock(p, 100*time.Millisecond)
	ok, err = p.Get(ctx, "k1", &got)
	if err != nil || !ok {
		t.Fatalf("refresh Get() ok=%v err=%v", ok, err)
	}

	advanceProviderClock(p, 250*time.Millisecond)
	ok, err = p.Get(ctx, "k1", &got)
	if err != nil || ok {
		t.Fatalf("expired Get() ok=%v err=%v", ok, err)
	}
}

func TestMemoryProvider_BothExpirations(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()

	if err := p.Set(ctx, "k1", "value1", TTL{
		Absolute: 100 * time.Millisecond,
		Sliding:  time.Second,
	}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	advanceProviderClock(p, 150*time.Millisecond)
	var got string
	ok, err := p.Get(ctx, "k1", &got)
	if err != nil || ok {
		t.Fatalf("Get() ok=%v err=%v", ok, err)
	}
}

func TestMemoryProvider_RemoveAndClear(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()

	_ = p.Set(ctx, "k1", "v1", TTL{})
	_ = p.Set(ctx, "k2", "v2", TTL{})

	if err := p.Remove(ctx, "k1"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	var got string
	ok, _ := p.Get(ctx, "k1", &got)
	if ok {
		t.Fatal("expected miss after Remove")
	}

	if err := p.Clear(ctx); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	ok, _ = p.Get(ctx, "k2", &got)
	if ok {
		t.Fatal("expected miss after Clear")
	}
}

func TestMemoryProvider_LRUEviction(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t, WithMaxSize(2))
	ctx := context.Background()

	_ = p.Set(ctx, "k1", "v1", TTL{})
	advanceProviderClock(p, time.Millisecond)
	_ = p.Set(ctx, "k2", "v2", TTL{})
	advanceProviderClock(p, time.Millisecond)

	var got string
	_, _ = p.Get(ctx, "k1", &got)

	_ = p.Set(ctx, "k3", "v3", TTL{})

	ok, _ := p.Get(ctx, "k2", &got)
	if ok {
		t.Fatal("expected k2 evicted")
	}
	ok, _ = p.Get(ctx, "k1", &got)
	if !ok || got != "v1" {
		t.Fatalf("k1 = ok=%v got=%q", ok, got)
	}
	ok, _ = p.Get(ctx, "k3", &got)
	if !ok || got != "v3" {
		t.Fatalf("k3 = ok=%v got=%q", ok, got)
	}
}

func TestMemoryProvider_Statistics(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t, WithStatistics())
	ctx := context.Background()

	_ = p.Set(ctx, "k1", "v1", TTL{})
	var got string
	_, _ = p.Get(ctx, "k1", &got)
	_, _ = p.Get(ctx, "missing", &got)

	stats := p.GetStatistics()
	if stats.Hits != 1 || stats.Misses != 1 || stats.Size != 1 {
		t.Fatalf("stats = %+v", stats)
	}

	_ = p.Clear(ctx)
	stats = p.GetStatistics()
	if stats.Hits != 0 || stats.Misses != 0 || stats.Size != 0 {
		t.Fatalf("stats after clear = %+v", stats)
	}
}

func TestMemoryProvider_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := "key"
			val := "value"
			_ = p.Set(ctx, key, val, TTL{})
			var got string
			_, _ = p.Get(ctx, key, &got)
		}()
	}
	wg.Wait()
}

func TestMemoryProvider_ContextCancellation(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var got string
	_, err := p.Get(ctx, "k", &got)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() error = %v", err)
	}
	if err := p.Set(ctx, "k", "v", TTL{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Set() error = %v", err)
	}
}

func TestMemoryProvider_TypeMismatch(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()
	_ = p.Set(ctx, "k", "string-value", TTL{})

	var got int
	ok, err := p.Get(ctx, "k", &got)
	if err == nil || ok {
		t.Fatalf("Get() ok=%v err=%v", ok, err)
	}
}

func TestMemoryProvider_SetOverwrite(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ctx := context.Background()
	_ = p.Set(ctx, "k", "v1", TTL{})
	_ = p.Set(ctx, "k", "v2", TTL{})

	var got string
	ok, err := p.Get(ctx, "k", &got)
	if err != nil || !ok || got != "v2" {
		t.Fatalf("Get() ok=%v got=%q err=%v", ok, got, err)
	}
}

func TestMemoryProvider_InvalidDest(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	ok, err := p.Get(context.Background(), "k", nil)
	if err == nil || ok {
		t.Fatalf("Get(nil dest) ok=%v err=%v", ok, err)
	}
}

func TestMemoryProvider_RemoveMissingKey(t *testing.T) {
	t.Parallel()
	p := newTestMemoryProvider(t)
	if err := p.Remove(context.Background(), "missing"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
}
