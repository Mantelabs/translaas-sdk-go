package cachefile_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Mantelabs/translaas-sdk-go/cachefile"
	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/Mantelabs/translaas-sdk-go/models"
)

type syncMockClient struct {
	mu sync.Mutex

	getProjectCalls        int
	getProjectLocalesCalls int

	getProjectFn        func(ctx context.Context, project, lang string, opts ...client.GetProjectOption) (*models.TranslationProject, error)
	getProjectLocalesFn func(ctx context.Context, project string, opts ...client.GetProjectLocalesOption) (*models.ProjectLocales, error)
	getOfflineCacheFn   func(ctx context.Context, project string, opts ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error)
}

func (m *syncMockClient) GetEntry(context.Context, string, string, string, ...client.GetEntryOption) (string, error) {
	return "", errors.New("unexpected GetEntry")
}

func (m *syncMockClient) GetGroup(context.Context, string, string, string, ...client.GetGroupOption) (*models.TranslationGroup, error) {
	return nil, errors.New("unexpected GetGroup")
}

func (m *syncMockClient) GetProject(ctx context.Context, project, lang string, opts ...client.GetProjectOption) (*models.TranslationProject, error) {
	m.mu.Lock()
	m.getProjectCalls++
	fn := m.getProjectFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, project, lang, opts...)
	}
	return nil, errors.New("unexpected GetProject")
}

func (m *syncMockClient) GetProjectLocales(ctx context.Context, project string, opts ...client.GetProjectLocalesOption) (*models.ProjectLocales, error) {
	m.mu.Lock()
	m.getProjectLocalesCalls++
	fn := m.getProjectLocalesFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, project, opts...)
	}
	return nil, errors.New("unexpected GetProjectLocales")
}

func (m *syncMockClient) GetOfflineCache(ctx context.Context, project string, opts ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error) {
	m.mu.Lock()
	fn := m.getOfflineCacheFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, project, opts...)
	}
	return nil, errors.New("unexpected GetOfflineCache")
}

func (m *syncMockClient) ReportMissingKeys(context.Context, []models.ReportMissingKeyItem) error {
	return errors.New("unexpected ReportMissingKeys")
}

func (m *syncMockClient) ValidateAPIKey(context.Context) (*models.ValidateAPIKeyResponse, error) {
	return nil, errors.New("unexpected ValidateAPIKey")
}

func (m *syncMockClient) getProjectCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getProjectCalls
}

type syncMockCache struct {
	mu sync.Mutex

	saveProjectCalls int
	saveLocalesCalls int

	projects map[string]*models.TranslationProject
	locales  map[string]*models.ProjectLocales
}

func newSyncMockCache() *syncMockCache {
	return &syncMockCache{
		projects: make(map[string]*models.TranslationProject),
		locales:  make(map[string]*models.ProjectLocales),
	}
}

func (m *syncMockCache) GetProject(_ context.Context, project, lang string) (*models.TranslationProject, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.projects[project+":"+lang], nil
}

func (m *syncMockCache) SaveProject(_ context.Context, project, lang string, data *models.TranslationProject, _ ...cachefile.SaveOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveProjectCalls++
	m.projects[project+":"+lang] = data
	return nil
}

func (m *syncMockCache) GetGroup(context.Context, string, string, string) (*models.TranslationGroup, error) {
	return nil, nil
}

func (m *syncMockCache) GetLocales(_ context.Context, project string) (*models.ProjectLocales, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.locales[project], nil
}

func (m *syncMockCache) SaveLocales(_ context.Context, project string, data *models.ProjectLocales, _ ...cachefile.SaveOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveLocalesCalls++
	m.locales[project] = data
	return nil
}

func (m *syncMockCache) GetManifest(context.Context) (*cachefile.CacheManifest, error) {
	return nil, nil
}

func (m *syncMockCache) UpdateManifest(context.Context, func(*cachefile.CacheManifest) error) error {
	return nil
}

func (m *syncMockCache) IsCached(context.Context, string, string) (bool, error) {
	return false, nil
}

func (m *syncMockCache) Clear(context.Context) error {
	return nil
}

func (m *syncMockCache) saveProjectCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveProjectCalls
}

func (m *syncMockCache) saveLocalesCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveLocalesCalls
}

func newSyncService(
	t *testing.T,
	inner *syncMockClient,
	cache cachefile.Provider,
	opts cachefile.OfflineCacheOptions,
	callbacks cachefile.SyncCallbacks,
) *cachefile.SyncService {
	t.Helper()
	svc, err := cachefile.NewSyncService(inner, cache, opts, callbacks)
	if err != nil {
		t.Fatalf("NewSyncService: %v", err)
	}
	return svc
}

func TestNewSyncServiceValidation(t *testing.T) {
	inner := &syncMockClient{}
	cache := newSyncMockCache()
	opts := cachefile.DefaultOfflineCacheOptions()

	if _, err := cachefile.NewSyncService(nil, cache, opts, cachefile.SyncCallbacks{}); err == nil {
		t.Fatal("expected error for nil client")
	}
	if _, err := cachefile.NewSyncService(inner, nil, opts, cachefile.SyncCallbacks{}); err == nil {
		t.Fatal("expected error for nil cache")
	}
}

func TestSyncProjectFetchesAndCaches(t *testing.T) {
	ctx := context.Background()
	inner := &syncMockClient{
		getProjectFn: func(_ context.Context, project, lang string, _ ...client.GetProjectOption) (*models.TranslationProject, error) {
			if project != "demo" || lang != "en" {
				t.Fatalf("unexpected GetProject(%q, %q)", project, lang)
			}
			return testProject("en"), nil
		},
	}
	cache := newSyncMockCache()
	svc := newSyncService(t, inner, cache, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{})

	if err := svc.SyncProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("SyncProject: %v", err)
	}
	if cache.saveProjectCallCount() != 1 {
		t.Fatalf("SaveProject calls = %d, want 1", cache.saveProjectCallCount())
	}
	if inner.getProjectCallCount() != 1 {
		t.Fatalf("GetProject calls = %d, want 1", inner.getProjectCallCount())
	}
}

func TestSyncProjectRaisesCompletedCallback(t *testing.T) {
	ctx := context.Background()
	inner := &syncMockClient{
		getProjectFn: func(context.Context, string, string, ...client.GetProjectOption) (*models.TranslationProject, error) {
			return &models.TranslationProject{}, nil
		},
	}
	cache := newSyncMockCache()

	var completed cachefile.SyncCompletedEvent
	svc := newSyncService(t, inner, cache, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{
		OnSyncCompleted: func(event cachefile.SyncCompletedEvent) {
			completed = event
		},
	})

	if err := svc.SyncProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("SyncProject: %v", err)
	}
	if completed.Project != "demo" || completed.Language != "en" {
		t.Fatalf("completed event = %+v", completed)
	}
}

func TestSyncProjectRaisesFailedCallbackAndReturnsError(t *testing.T) {
	ctx := context.Background()
	syncErr := errors.New("api down")
	inner := &syncMockClient{
		getProjectFn: func(context.Context, string, string, ...client.GetProjectOption) (*models.TranslationProject, error) {
			return nil, syncErr
		},
	}
	cache := newSyncMockCache()

	var failed cachefile.SyncFailedEvent
	svc := newSyncService(t, inner, cache, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{
		OnSyncFailed: func(event cachefile.SyncFailedEvent) {
			failed = event
		},
	})

	if err := svc.SyncProject(ctx, "demo", "en"); !errors.Is(err, syncErr) {
		t.Fatalf("SyncProject err = %v, want %v", err, syncErr)
	}
	if failed.Project != "demo" || failed.Language != "en" || !errors.Is(failed.Err, syncErr) {
		t.Fatalf("failed event = %+v", failed)
	}
}

func TestSyncProjectAllLanguagesFiltersConfiguredLanguages(t *testing.T) {
	ctx := context.Background()
	inner := &syncMockClient{
		getProjectLocalesFn: func(_ context.Context, project string, _ ...client.GetProjectLocalesOption) (*models.ProjectLocales, error) {
			return &models.ProjectLocales{Project: project, Locales: []string{"en", "es", "fr"}}, nil
		},
		getProjectFn: func(_ context.Context, project, lang string, _ ...client.GetProjectOption) (*models.TranslationProject, error) {
			return testProject("en"), nil
		},
	}
	cache := newSyncMockCache()
	opts := cachefile.DefaultOfflineCacheOptions()
	opts.Languages = []string{"en", "es"}
	svc := newSyncService(t, inner, cache, opts, cachefile.SyncCallbacks{})

	if err := svc.SyncProjectAllLanguages(ctx, "demo"); err != nil {
		t.Fatalf("SyncProjectAllLanguages: %v", err)
	}
	if cache.saveLocalesCallCount() != 1 {
		t.Fatalf("SaveLocales calls = %d, want 1", cache.saveLocalesCallCount())
	}
	if inner.getProjectCallCount() != 2 {
		t.Fatalf("GetProject calls = %d, want 2", inner.getProjectCallCount())
	}
}

func TestSyncProjectAllLanguagesSyncsAllWhenLanguagesEmpty(t *testing.T) {
	ctx := context.Background()
	inner := &syncMockClient{
		getProjectLocalesFn: func(_ context.Context, project string, _ ...client.GetProjectLocalesOption) (*models.ProjectLocales, error) {
			return &models.ProjectLocales{Project: project, Locales: []string{"en", "es", "fr"}}, nil
		},
		getProjectFn: func(_ context.Context, project, lang string, _ ...client.GetProjectOption) (*models.TranslationProject, error) {
			return testProject("en"), nil
		},
	}
	cache := newSyncMockCache()
	svc := newSyncService(t, inner, cache, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{})

	if err := svc.SyncProjectAllLanguages(ctx, "demo"); err != nil {
		t.Fatalf("SyncProjectAllLanguages: %v", err)
	}
	if inner.getProjectCallCount() != 3 {
		t.Fatalf("GetProject calls = %d, want 3", inner.getProjectCallCount())
	}
}

func TestSyncProjectAllLanguagesContinuesOnLanguageFailure(t *testing.T) {
	ctx := context.Background()
	inner := &syncMockClient{
		getProjectLocalesFn: func(_ context.Context, project string, _ ...client.GetProjectLocalesOption) (*models.ProjectLocales, error) {
			return &models.ProjectLocales{Project: project, Locales: []string{"en", "es"}}, nil
		},
		getProjectFn: func(_ context.Context, _, lang string, _ ...client.GetProjectOption) (*models.TranslationProject, error) {
			if lang == "en" {
				return nil, errors.New("en failed")
			}
			return testProject(lang), nil
		},
	}
	cache := newSyncMockCache()

	var failedLangs []string
	svc := newSyncService(t, inner, cache, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{
		OnSyncFailed: func(event cachefile.SyncFailedEvent) {
			failedLangs = append(failedLangs, event.Language)
		},
	})

	if err := svc.SyncProjectAllLanguages(ctx, "demo"); err != nil {
		t.Fatalf("SyncProjectAllLanguages: %v", err)
	}
	if len(failedLangs) != 1 || failedLangs[0] != "en" {
		t.Fatalf("failedLangs = %v, want [en]", failedLangs)
	}
	if cache.saveProjectCallCount() != 1 {
		t.Fatalf("SaveProject calls = %d, want 1", cache.saveProjectCallCount())
	}
}

func TestSyncAllAggregatesPartialFailures(t *testing.T) {
	ctx := context.Background()
	inner := &syncMockClient{
		getProjectLocalesFn: func(_ context.Context, project string, _ ...client.GetProjectLocalesOption) (*models.ProjectLocales, error) {
			if project == "bad" {
				return nil, errors.New("locales failed")
			}
			return &models.ProjectLocales{Project: project, Locales: []string{"en"}}, nil
		},
		getProjectFn: func(_ context.Context, project, lang string, _ ...client.GetProjectOption) (*models.TranslationProject, error) {
			return testProject("en"), nil
		},
	}
	cache := newSyncMockCache()
	opts := cachefile.DefaultOfflineCacheOptions()
	opts.Projects = []string{"bad", "good"}

	var allCompleted cachefile.SyncResult
	svc := newSyncService(t, inner, cache, opts, cachefile.SyncCallbacks{
		OnSyncAllCompleted: func(result cachefile.SyncResult) {
			allCompleted = result
		},
	})

	result, err := svc.SyncAll(ctx)
	if err != nil {
		t.Fatalf("SyncAll: %v", err)
	}
	if len(result.SyncedProjects) != 1 || result.SyncedProjects[0] != "good" {
		t.Fatalf("SyncedProjects = %v, want [good]", result.SyncedProjects)
	}
	if len(result.FailedProjects) != 1 || result.FailedProjects[0] != "bad" {
		t.Fatalf("FailedProjects = %v, want [bad]", result.FailedProjects)
	}
	if len(allCompleted.SyncedProjects) != 1 || len(allCompleted.FailedProjects) != 1 {
		t.Fatalf("OnSyncAllCompleted = %+v", allCompleted)
	}
}

func TestSyncServiceUsesInnerClientDirectly(t *testing.T) {
	ctx := context.Background()
	inner := &syncMockClient{
		getProjectFn: func(context.Context, string, string, ...client.GetProjectOption) (*models.TranslationProject, error) {
			return testProject("sync"), nil
		},
	}
	cache := newSyncMockCache()

	svc, err := cachefile.NewSyncService(inner, cache, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{})
	if err != nil {
		t.Fatalf("NewSyncService: %v", err)
	}

	if err := svc.SyncProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("SyncProject: %v", err)
	}
	if inner.getProjectCallCount() != 1 {
		t.Fatalf("inner GetProject calls = %d, want 1", inner.getProjectCallCount())
	}

	decorated, err := cachefile.NewCachingClient(inner, cache, cachefile.Options{
		FallbackMode:     cachefile.FallbackCacheFirst,
		DefaultProjectID: "demo",
	})
	if err != nil {
		t.Fatalf("NewCachingClient: %v", err)
	}

	// SyncService must be wired with inner; a cache hit for "en" should not increment inner again.
	if _, err := decorated.GetProject(ctx, "demo", "en"); err != nil {
		t.Fatalf("decorated GetProject cached lang: %v", err)
	}
	if inner.getProjectCallCount() != 1 {
		t.Fatalf("inner GetProject calls after cached read = %d, want 1", inner.getProjectCallCount())
	}

	// Uncached language still goes through the decorator to inner.
	if _, err := decorated.GetProject(ctx, "demo", "fr"); err != nil {
		t.Fatalf("decorated GetProject uncached lang: %v", err)
	}
	if inner.getProjectCallCount() != 2 {
		t.Fatalf("inner GetProject calls after uncached read = %d, want 2", inner.getProjectCallCount())
	}
}

func TestStartBackgroundSyncNoOpWhenAutoSyncDisabled(t *testing.T) {
	inner := &syncMockClient{}
	cache := newSyncMockCache()
	opts := cachefile.DefaultOfflineCacheOptions()
	opts.AutoSync = false
	svc := newSyncService(t, inner, cache, opts, cachefile.SyncCallbacks{})

	svc.StartBackgroundSync(context.Background())
	if svc.IsBackgroundSyncRunning() {
		t.Fatal("expected background sync to stay stopped when AutoSync is false")
	}
}

func TestBackgroundSyncRespectsContextCancellation(t *testing.T) {
	inner := &syncMockClient{
		getProjectLocalesFn: func(_ context.Context, project string, _ ...client.GetProjectLocalesOption) (*models.ProjectLocales, error) {
			return &models.ProjectLocales{Project: project, Locales: []string{"en"}}, nil
		},
		getProjectFn: func(_ context.Context, project, lang string, _ ...client.GetProjectOption) (*models.TranslationProject, error) {
			return testProject("en"), nil
		},
	}
	cache := newSyncMockCache()
	opts := cachefile.DefaultOfflineCacheOptions()
	opts.Projects = []string{"demo"}
	interval := 20 * time.Millisecond
	opts.AutoSyncInterval = &interval
	svc := newSyncService(t, inner, cache, opts, cachefile.SyncCallbacks{})

	ctx, cancel := context.WithCancel(context.Background())
	svc.StartBackgroundSync(ctx)
	if !svc.IsBackgroundSyncRunning() {
		t.Fatal("expected background sync to be running")
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	svc.StopBackgroundSync()

	if svc.IsBackgroundSyncRunning() {
		t.Fatal("expected background sync to stop after context cancellation")
	}
}

func TestSyncProjectValidation(t *testing.T) {
	ctx := context.Background()
	inner := &syncMockClient{}
	cache := newSyncMockCache()
	svc := newSyncService(t, inner, cache, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{})

	if err := svc.SyncProject(ctx, "", "en"); err == nil {
		t.Fatal("expected error for empty project")
	}
	if err := svc.SyncProject(ctx, "demo", ""); err == nil {
		t.Fatal("expected error for empty language")
	}
}
