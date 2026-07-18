package cachefile_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/acuencadev/translaas-sdk-go/cachefile"
	"github.com/acuencadev/translaas-sdk-go/models"
)

type mockL2Provider struct {
	mu sync.Mutex

	getProjectCalls int
	getGroupCalls   int
	isCachedCalls   int

	projects map[string]*models.TranslationProject
	groups   map[string]*models.TranslationGroup
	locales  map[string]*models.ProjectLocales
	cached   map[string]bool
}

func newMockL2Provider() *mockL2Provider {
	return &mockL2Provider{
		projects: make(map[string]*models.TranslationProject),
		groups:   make(map[string]*models.TranslationGroup),
		locales:  make(map[string]*models.ProjectLocales),
		cached:   make(map[string]bool),
	}
}

func (m *mockL2Provider) GetProject(_ context.Context, project, lang string) (*models.TranslationProject, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getProjectCalls++
	key := project + ":" + lang
	return m.projects[key], nil
}

func (m *mockL2Provider) SaveProject(_ context.Context, project, lang string, data *models.TranslationProject, _ ...cachefile.SaveOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[project+":"+lang] = data
	m.cached[project+":"+lang] = true
	return nil
}

func (m *mockL2Provider) GetGroup(_ context.Context, project, group, lang string) (*models.TranslationGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getGroupCalls++
	key := project + ":" + group + ":" + lang
	return m.groups[key], nil
}

func (m *mockL2Provider) GetLocales(_ context.Context, project string) (*models.ProjectLocales, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.locales[project], nil
}

func (m *mockL2Provider) SaveLocales(_ context.Context, project string, data *models.ProjectLocales, _ ...cachefile.SaveOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.locales[project] = data
	return nil
}

func (m *mockL2Provider) GetManifest(context.Context) (*cachefile.CacheManifest, error) {
	return nil, nil
}

func (m *mockL2Provider) UpdateManifest(context.Context, func(*cachefile.CacheManifest) error) error {
	return nil
}

func (m *mockL2Provider) IsCached(_ context.Context, project, lang string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isCachedCalls++
	return m.cached[project+":"+lang], nil
}

func (m *mockL2Provider) Clear(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects = make(map[string]*models.TranslationProject)
	m.groups = make(map[string]*models.TranslationGroup)
	m.locales = make(map[string]*models.ProjectLocales)
	m.cached = make(map[string]bool)
	return nil
}

func (m *mockL2Provider) getProjectCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getProjectCalls
}

func (m *mockL2Provider) getGroupCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getGroupCalls
}

func (m *mockL2Provider) isCachedCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isCachedCalls
}

func testProject(label string) *models.TranslationProject {
	return &models.TranslationProject{
		Groups: map[string]json.RawMessage{
			"common": json.RawMessage(`{"` + label + `":"` + label + `"}`),
		},
	}
}

func newHybridWithMock(t *testing.T, opts cachefile.HybridOptions) (*cachefile.HybridProvider, *mockL2Provider) {
	t.Helper()
	l2 := newMockL2Provider()
	provider, err := cachefile.NewHybridProvider(l2, opts)
	if err != nil {
		t.Fatalf("NewHybridProvider: %v", err)
	}
	return provider, l2
}

func TestNewHybridProviderRequiresL2(t *testing.T) {
	t.Parallel()

	_, err := cachefile.NewHybridProvider(nil, cachefile.DefaultHybridOptions())
	if err == nil {
		t.Fatal("expected error for nil L2")
	}
}

func TestHybridProviderPromotesL2HitToL1(t *testing.T) {
	t.Parallel()

	l2 := newMockL2Provider()
	l2.projects["demo:en"] = testProject("Hello")

	provider, err := cachefile.NewHybridProvider(l2, cachefile.DefaultHybridOptions())
	if err != nil {
		t.Fatalf("NewHybridProvider: %v", err)
	}

	ctx := context.Background()
	if _, err := provider.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("first GetProject: %v", err)
	}
	if _, err := provider.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("second GetProject: %v", err)
	}

	if got := l2.getProjectCallCount(); got != 1 {
		t.Fatalf("GetProject L2 calls = %d, want 1", got)
	}
}

func TestHybridProviderSaveProjectPopulatesL1(t *testing.T) {
	t.Parallel()

	provider, l2 := newHybridWithMock(t, cachefile.DefaultHybridOptions())
	ctx := context.Background()
	project := testProject("saved")

	if err := provider.SaveProject(ctx, "demo", "en", project); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	if _, err := provider.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got := l2.getProjectCallCount(); got != 0 {
		t.Fatalf("GetProject L2 calls = %d, want 0", got)
	}

	if _, err := provider.GetGroup(ctx, "demo", "common", "en"); err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if got := l2.getGroupCallCount(); got != 0 {
		t.Fatalf("GetGroup L2 calls = %d, want 0", got)
	}
}

func TestHybridProviderL1ExpiresAfterTTL(t *testing.T) {
	t.Parallel()

	opts := cachefile.DefaultHybridOptions()
	opts.MemoryExpiration = 50 * time.Millisecond

	l2 := newMockL2Provider()
	l2.projects["demo:en"] = testProject("Hello")

	provider, err := cachefile.NewHybridProvider(l2, opts)
	if err != nil {
		t.Fatalf("NewHybridProvider: %v", err)
	}

	ctx := context.Background()
	if _, err := provider.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("first GetProject: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := provider.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("second GetProject: %v", err)
	}

	if got := l2.getProjectCallCount(); got != 2 {
		t.Fatalf("GetProject L2 calls = %d, want 2 after TTL expiry", got)
	}
}

func TestHybridProviderLRUEvictsOldestEntry(t *testing.T) {
	t.Parallel()

	opts := cachefile.DefaultHybridOptions()
	opts.MaxEntries = 2
	opts.MemoryExpiration = time.Minute

	l2 := newMockL2Provider()
	l2.projects["a:en"] = testProject("A")
	l2.projects["b:en"] = testProject("B")
	l2.projects["c:en"] = testProject("C")

	provider, err := cachefile.NewHybridProvider(l2, opts)
	if err != nil {
		t.Fatalf("NewHybridProvider: %v", err)
	}

	ctx := context.Background()
	for _, id := range []string{"a", "b", "c"} {
		if _, err := provider.GetProject(ctx, id, "en"); err != nil {
			t.Fatalf("GetProject(%s): %v", id, err)
		}
	}

	if _, err := provider.GetProject(ctx, "a", "en"); err != nil {
		t.Fatalf("GetProject(a) after eviction: %v", err)
	}

	if got := l2.getProjectCallCount(); got != 4 {
		t.Fatalf("GetProject L2 calls = %d, want 4 (a evicted, then refetched)", got)
	}
}

func TestHybridProviderIsCachedUsesL1(t *testing.T) {
	t.Parallel()

	provider, l2 := newHybridWithMock(t, cachefile.DefaultHybridOptions())
	ctx := context.Background()

	if err := provider.SaveProject(ctx, "demo", "en", testProject("x")); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	cached, err := provider.IsCached(ctx, "demo", "en")
	if err != nil {
		t.Fatalf("IsCached: %v", err)
	}
	if !cached {
		t.Fatal("expected IsCached true from L1")
	}
	if got := l2.isCachedCallCount(); got != 0 {
		t.Fatalf("IsCached L2 calls = %d, want 0", got)
	}
}

func TestHybridProviderClearRemovesL1AndL2(t *testing.T) {
	t.Parallel()

	provider, _ := newHybridWithMock(t, cachefile.DefaultHybridOptions())
	ctx := context.Background()

	if err := provider.SaveProject(ctx, "demo", "en", testProject("x")); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	if err := provider.Clear(ctx); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	projects, groups, locales := provider.MemoryCacheStats()
	if projects != 0 || groups != 0 || locales != 0 {
		t.Fatalf("MemoryCacheStats after Clear = (%d,%d,%d), want zeros", projects, groups, locales)
	}
}

func TestHybridProviderWarmupPopulatesL1(t *testing.T) {
	t.Parallel()

	l2 := newMockL2Provider()
	l2.projects["demo:en"] = testProject("warm")

	provider, err := cachefile.NewHybridProvider(l2, cachefile.DefaultHybridOptions())
	if err != nil {
		t.Fatalf("NewHybridProvider: %v", err)
	}

	ctx := context.Background()
	ok, err := provider.Warmup(ctx, "demo", "en")
	if err != nil {
		t.Fatalf("Warmup: %v", err)
	}
	if !ok {
		t.Fatal("expected Warmup true")
	}

	if _, err := provider.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got := l2.getProjectCallCount(); got != 1 {
		t.Fatalf("GetProject L2 calls = %d, want 1 (warmup only)", got)
	}
}

func TestHybridProviderDisabledDelegatesToL2Only(t *testing.T) {
	t.Parallel()

	opts := cachefile.DefaultHybridOptions()
	opts.Enabled = false

	provider, l2 := newHybridWithMock(t, opts)
	ctx := context.Background()
	l2.projects["demo:en"] = testProject("delegated")

	if _, err := provider.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("first GetProject: %v", err)
	}
	if _, err := provider.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("second GetProject: %v", err)
	}

	if got := l2.getProjectCallCount(); got != 2 {
		t.Fatalf("GetProject L2 calls = %d, want 2 when L1 disabled", got)
	}
}

func TestHybridProviderIntegrationWithFileProvider(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fileProvider, err := cachefile.NewFileProvider(dir)
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}

	provider, err := cachefile.NewHybridProvider(fileProvider, cachefile.DefaultHybridOptions())
	if err != nil {
		t.Fatalf("NewHybridProvider: %v", err)
	}

	ctx := context.Background()
	project := testProject("integration")

	if err := provider.SaveProject(ctx, "demo", "en", project); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	got, err := provider.GetProject(ctx, "demo", "en")
	if err != nil || got == nil {
		t.Fatalf("GetProject: err=%v got=%v", err, got)
	}

	provider.ClearMemoryCache()

	got, err = provider.GetProject(ctx, "demo", "en")
	if err != nil || got == nil {
		t.Fatalf("GetProject after ClearMemoryCache: err=%v got=%v", err, got)
	}
}

func TestHybridProviderConcurrentAccessRaceSafe(t *testing.T) {
	t.Parallel()

	l2 := newMockL2Provider()
	l2.projects["demo:en"] = testProject("race")

	provider, err := cachefile.NewHybridProvider(l2, cachefile.DefaultHybridOptions())
	if err != nil {
		t.Fatalf("NewHybridProvider: %v", err)
	}

	ctx := context.Background()
	const workers = 16
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, _ = provider.GetProject(ctx, "demo", "en")
		}()
	}
	wg.Wait()
}
