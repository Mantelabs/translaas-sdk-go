package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/acuencadev/translaas-sdk-go/cache"
)

func newCachedTestClient(
	t *testing.T,
	handler http.HandlerFunc,
	mode cache.Mode,
	provider cache.Provider,
	optFns ...Option,
) Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	opts := []Option{WithHTTPClient(srv.Client())}
	if provider != nil {
		opts = append(opts, WithCacheProvider(provider))
	}
	opts = append(opts, optFns...)

	cli, err := New(Options{
		APIKey:    "test-api-key",
		BaseURL:   srv.URL,
		CacheMode: mode,
	}, opts...)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return cli
}

func TestShouldCache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode cache.Mode
		op   string
		want bool
	}{
		{cache.ModeNone, "entry", false},
		{cache.ModeNone, "group", false},
		{cache.ModeNone, "project", false},
		{cache.ModeNone, "locales", false},
		{cache.ModeEntry, "entry", true},
		{cache.ModeEntry, "group", false},
		{cache.ModeEntry, "project", false},
		{cache.ModeEntry, "locales", true},
		{cache.ModeGroup, "entry", false},
		{cache.ModeGroup, "group", true},
		{cache.ModeGroup, "project", false},
		{cache.ModeGroup, "locales", true},
		{cache.ModeProject, "entry", false},
		{cache.ModeProject, "group", true},
		{cache.ModeProject, "project", true},
		{cache.ModeProject, "locales", true},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String()+"_"+tt.op, func(t *testing.T) {
			t.Parallel()
			if got := shouldCache(tt.mode, tt.op); got != tt.want {
				t.Fatalf("shouldCache(%v, %q) = %v, want %v", tt.mode, tt.op, got, tt.want)
			}
		})
	}
}

func TestNew_DefaultMemoryProviderWhenCacheEnabled(t *testing.T) {
	t.Parallel()
	cli, err := New(Options{
		APIKey:    "key",
		BaseURL:   "https://example.com",
		CacheMode: cache.ModeEntry,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	c := cli.(*client)
	if c.cacheProvider == nil {
		t.Fatal("expected default cache provider")
	}
}

func TestGetEntry_CacheHitMiss(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := cache.NewMemoryProvider()

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		_, _ = io.WriteString(w, "Hello, World!")
	}, cache.ModeEntry, provider)

	ctx := context.Background()
	got, err := cli.GetEntry(ctx, "ui", "greeting", "en")
	if err != nil || got != "Hello, World!" {
		t.Fatalf("first GetEntry() = (%q, %v)", got, err)
	}
	if requestCount.Load() != 1 {
		t.Fatalf("first request count = %d", requestCount.Load())
	}

	got, err = cli.GetEntry(ctx, "ui", "greeting", "en")
	if err != nil || got != "Hello, World!" {
		t.Fatalf("second GetEntry() = (%q, %v)", got, err)
	}
	if requestCount.Load() != 1 {
		t.Fatalf("second request count = %d, want cache hit", requestCount.Load())
	}
}

func TestGetEntry_NotCachedInGroupMode(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := cache.NewMemoryProvider()

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		_, _ = io.WriteString(w, "value")
	}, cache.ModeGroup, provider)

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		got, err := cli.GetEntry(ctx, "ui", "greeting", "en")
		if err != nil || got != "value" {
			t.Fatalf("GetEntry() = (%q, %v)", got, err)
		}
	}
	if requestCount.Load() != 2 {
		t.Fatalf("request count = %d, want passthrough without entry cache", requestCount.Load())
	}
}

func TestGetGroup_CacheHitMiss(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := cache.NewMemoryProvider()
	body := `{"Entries":{"welcome":"Welcome"}}`

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		_, _ = io.WriteString(w, body)
	}, cache.ModeGroup, provider)

	ctx := context.Background()
	got, err := cli.GetGroup(ctx, "p", "ui", "en")
	if err != nil {
		t.Fatalf("first GetGroup() error = %v", err)
	}
	value, ok := got.GetValue("welcome")
	if !ok || value != "Welcome" {
		t.Fatalf("welcome = (%q, %v)", value, ok)
	}

	got, err = cli.GetGroup(ctx, "p", "ui", "en")
	if err != nil {
		t.Fatalf("second GetGroup() error = %v", err)
	}
	value, ok = got.GetValue("welcome")
	if !ok || value != "Welcome" {
		t.Fatalf("cached welcome = (%q, %v)", value, ok)
	}
	if requestCount.Load() != 1 {
		t.Fatalf("request count = %d", requestCount.Load())
	}
}

func TestGetProject_CacheHitMiss(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := cache.NewMemoryProvider()
	body := `{"Groups":{"ui":{"title":"Checkout"}}}`

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		_, _ = io.WriteString(w, body)
	}, cache.ModeProject, provider)

	ctx := context.Background()
	if _, err := cli.GetProject(ctx, "p", "en"); err != nil {
		t.Fatalf("first GetProject() error = %v", err)
	}
	if _, err := cli.GetProject(ctx, "p", "en"); err != nil {
		t.Fatalf("second GetProject() error = %v", err)
	}
	if requestCount.Load() != 1 {
		t.Fatalf("request count = %d", requestCount.Load())
	}
}

func TestGetProjectLocales_CachedInEntryMode(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := cache.NewMemoryProvider()
	body := `{"locales":["en","fr"]}`

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		_, _ = io.WriteString(w, body)
	}, cache.ModeEntry, provider)

	ctx := context.Background()
	got, err := cli.GetProjectLocales(ctx, "p")
	if err != nil {
		t.Fatalf("first GetProjectLocales() error = %v", err)
	}
	if len(got.Locales) != 2 {
		t.Fatalf("locales = %v", got.Locales)
	}

	if _, err := cli.GetProjectLocales(ctx, "p"); err != nil {
		t.Fatalf("second GetProjectLocales() error = %v", err)
	}
	if requestCount.Load() != 1 {
		t.Fatalf("request count = %d", requestCount.Load())
	}
}

func TestGetEntry_304ReturnsCachedValue(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := &sequentialHitProvider{
		inner: cache.NewMemoryProvider(),
	}

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusNotModified)
	}, cache.ModeEntry, provider)

	ctx := context.Background()
	_ = provider.inner.Set(ctx, cache.EntryKey("ui", "greeting", "en", nil, nil, "", "", ""), "Cached greeting", cache.TTL{})

	got, err := cli.GetEntry(ctx, "ui", "greeting", "en")
	if err != nil || got != "Cached greeting" {
		t.Fatalf("304 fallback GetEntry() = (%q, %v)", got, err)
	}
	if requestCount.Load() != 1 {
		t.Fatalf("request count = %d", requestCount.Load())
	}
}

type sequentialHitProvider struct {
	inner    cache.Provider
	getCalls int
}

func (p *sequentialHitProvider) Get(ctx context.Context, key string, dest any) (bool, error) {
	p.getCalls++
	if p.getCalls == 1 {
		return false, nil
	}
	return p.inner.Get(ctx, key, dest)
}

func (p *sequentialHitProvider) Set(ctx context.Context, key string, value any, ttl cache.TTL) error {
	return p.inner.Set(ctx, key, value, ttl)
}

func (p *sequentialHitProvider) Remove(ctx context.Context, key string) error {
	return p.inner.Remove(ctx, key)
}

func (p *sequentialHitProvider) Clear(ctx context.Context) error {
	return p.inner.Clear(ctx)
}

func TestGetEntry_304WithoutCacheReturnsEmpty(t *testing.T) {
	t.Parallel()
	provider := cache.NewMemoryProvider()

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}, cache.ModeEntry, provider)

	got, err := cli.GetEntry(context.Background(), "ui", "greeting", "en")
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
	if got != "" {
		t.Fatalf("got %q, want empty string on 304 miss", got)
	}
}

func TestGetGroup_304DoesNotPoisonCache(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := cache.NewMemoryProvider()
	body := `{"Entries":{"title":"Original"}}`

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		switch count {
		case 1:
			_, _ = io.WriteString(w, body)
		default:
			w.WriteHeader(http.StatusNotModified)
		}
	}, cache.ModeGroup, provider)

	ctx := context.Background()
	got, err := cli.GetGroup(ctx, "p", "ui", "en")
	if err != nil {
		t.Fatalf("warm GetGroup() error = %v", err)
	}
	value, ok := got.GetValue("title")
	if !ok || value != "Original" {
		t.Fatalf("warm title = (%q, %v)", value, ok)
	}

	for i := 0; i < 2; i++ {
		got, err = cli.GetGroup(ctx, "p", "ui", "en")
		if err != nil {
			t.Fatalf("GetGroup() error = %v", err)
		}
		value, ok = got.GetValue("title")
		if !ok || value != "Original" {
			t.Fatalf("cached title after 304 = (%q, %v)", value, ok)
		}
	}
}

func TestGetEntry_CacheAbsoluteExpiry(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := cache.NewMemoryProvider()

	cli := newCachedTestClient(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			_, _ = io.WriteString(w, "value")
		},
		cache.ModeEntry,
		provider,
		WithCacheTTL(cache.TTL{Absolute: 20 * time.Millisecond}),
	)

	ctx := context.Background()
	if _, err := cli.GetEntry(ctx, "ui", "greeting", "en"); err != nil {
		t.Fatalf("first GetEntry() error = %v", err)
	}
	time.Sleep(30 * time.Millisecond)
	if _, err := cli.GetEntry(ctx, "ui", "greeting", "en"); err != nil {
		t.Fatalf("second GetEntry() error = %v", err)
	}
	if requestCount.Load() != 2 {
		t.Fatalf("request count = %d, want expiry miss", requestCount.Load())
	}
}

func TestValidateAPIKey_NotCached(t *testing.T) {
	t.Parallel()
	var requestCount atomic.Int32
	provider := cache.NewMemoryProvider()
	body := `{"valid":true,"projectId":"01ARZ3NDEKTSV4RRFFQ69G5FAV"}`

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		_, _ = io.WriteString(w, body)
	}, cache.ModeProject, provider)

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if _, err := cli.ValidateAPIKey(ctx); err != nil {
			t.Fatalf("ValidateAPIKey() error = %v", err)
		}
	}
	if requestCount.Load() != 2 {
		t.Fatalf("request count = %d, want passthrough", requestCount.Load())
	}
}

func TestGetEntry_CachedValueIsolatedFromMutation(t *testing.T) {
	t.Parallel()
	provider := cache.NewMemoryProvider()

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "original")
	}, cache.ModeEntry, provider)

	ctx := context.Background()
	_, err := cli.GetEntry(ctx, "ui", "greeting", "en")
	if err != nil {
		t.Fatalf("first GetEntry() error = %v", err)
	}

	second, err := cli.GetEntry(ctx, "ui", "greeting", "en")
	if err != nil {
		t.Fatalf("second GetEntry() error = %v", err)
	}
	if second != "original" {
		t.Fatalf("cached value = %q, want original", second)
	}
}

func TestGetGroup_CachedValueIsolatedFromMutation(t *testing.T) {
	t.Parallel()
	provider := cache.NewMemoryProvider()
	body := `{"Entries":{"title":"Original"}}`

	cli := newCachedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, body)
	}, cache.ModeGroup, provider)

	ctx := context.Background()
	first, err := cli.GetGroup(ctx, "p", "ui", "en")
	if err != nil {
		t.Fatalf("first GetGroup() error = %v", err)
	}
	if first.Entries == nil {
		first.Entries = make(map[string]json.RawMessage)
	}
	first.Entries["title"] = json.RawMessage(`"Changed"`)

	second, err := cli.GetGroup(ctx, "p", "ui", "en")
	if err != nil {
		t.Fatalf("second GetGroup() error = %v", err)
	}
	value, ok := second.GetValue("title")
	if !ok || value != "Original" {
		t.Fatalf("cached title = (%q, %v)", value, ok)
	}
}
