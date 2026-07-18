package cachefile_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/cachefile"
	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/models"
)

const testProjectID = "demo-project"

type mockInnerClient struct {
	mu sync.Mutex

	getEntryCalls int
	getGroupCalls int

	getEntryFn        func(ctx context.Context, group, entry, lang string, opts ...client.GetEntryOption) (string, error)
	getGroupFn        func(ctx context.Context, project, group, lang string, opts ...client.GetGroupOption) (*models.TranslationGroup, error)
	validateAPIKeyFn  func(context.Context) (*models.ValidateAPIKeyResponse, error)
	reportMissingFn   func(context.Context, []models.ReportMissingKeyItem) error
	getOfflineCacheFn func(context.Context, string, ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error)
}

func (m *mockInnerClient) GetEntry(ctx context.Context, group, entry, lang string, opts ...client.GetEntryOption) (string, error) {
	m.mu.Lock()
	m.getEntryCalls++
	fn := m.getEntryFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, group, entry, lang, opts...)
	}
	return "", errors.New("unexpected GetEntry")
}

func (m *mockInnerClient) GetGroup(ctx context.Context, project, group, lang string, opts ...client.GetGroupOption) (*models.TranslationGroup, error) {
	m.mu.Lock()
	m.getGroupCalls++
	fn := m.getGroupFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, project, group, lang, opts...)
	}
	return nil, errors.New("unexpected GetGroup")
}

func (m *mockInnerClient) GetProject(context.Context, string, string, ...client.GetProjectOption) (*models.TranslationProject, error) {
	return nil, errors.New("unexpected GetProject")
}

func (m *mockInnerClient) GetProjectLocales(context.Context, string, ...client.GetProjectLocalesOption) (*models.ProjectLocales, error) {
	return nil, errors.New("unexpected GetProjectLocales")
}

func (m *mockInnerClient) GetOfflineCache(ctx context.Context, project string, opts ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error) {
	if m.getOfflineCacheFn != nil {
		return m.getOfflineCacheFn(ctx, project, opts...)
	}
	return nil, errors.New("unexpected GetOfflineCache")
}

func (m *mockInnerClient) ReportMissingKeys(ctx context.Context, keys []models.ReportMissingKeyItem) error {
	if m.reportMissingFn != nil {
		return m.reportMissingFn(ctx, keys)
	}
	return errors.New("unexpected ReportMissingKeys")
}

func (m *mockInnerClient) ValidateAPIKey(ctx context.Context) (*models.ValidateAPIKeyResponse, error) {
	if m.validateAPIKeyFn != nil {
		return m.validateAPIKeyFn(ctx)
	}
	return nil, errors.New("unexpected ValidateAPIKey")
}

func (m *mockInnerClient) getEntryCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getEntryCalls
}

type mockCacheProvider struct {
	mu sync.Mutex

	getGroupCalls int
	saveProject   int

	groups   map[string]*models.TranslationGroup
	projects map[string]*models.TranslationProject
	locales  map[string]*models.ProjectLocales
}

func newMockCacheProvider() *mockCacheProvider {
	return &mockCacheProvider{
		groups:   make(map[string]*models.TranslationGroup),
		projects: make(map[string]*models.TranslationProject),
		locales:  make(map[string]*models.ProjectLocales),
	}
}

func groupKey(project, group, lang string) string {
	return project + ":" + group + ":" + lang
}

func (m *mockCacheProvider) GetProject(_ context.Context, project, lang string) (*models.TranslationProject, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.projects[project+":"+lang], nil
}

func (m *mockCacheProvider) SaveProject(_ context.Context, project, lang string, data *models.TranslationProject, _ ...cachefile.SaveOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveProject++
	m.projects[project+":"+lang] = data
	return nil
}

func (m *mockCacheProvider) GetGroup(_ context.Context, project, group, lang string) (*models.TranslationGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getGroupCalls++
	return m.groups[groupKey(project, group, lang)], nil
}

func (m *mockCacheProvider) GetLocales(_ context.Context, project string) (*models.ProjectLocales, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.locales[project], nil
}

func (m *mockCacheProvider) SaveLocales(_ context.Context, project string, data *models.ProjectLocales, _ ...cachefile.SaveOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.locales[project] = data
	return nil
}

func (m *mockCacheProvider) GetManifest(context.Context) (*cachefile.CacheManifest, error) {
	return nil, nil
}

func (m *mockCacheProvider) UpdateManifest(context.Context, func(*cachefile.CacheManifest) error) error {
	return nil
}

func (m *mockCacheProvider) IsCached(context.Context, string, string) (bool, error) {
	return false, nil
}

func (m *mockCacheProvider) Clear(context.Context) error {
	return nil
}

func (m *mockCacheProvider) saveProjectCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveProject
}

func testGroup(entry, value string) *models.TranslationGroup {
	return &models.TranslationGroup{
		Entries: map[string]json.RawMessage{
			entry: json.RawMessage(`"` + value + `"`),
		},
	}
}

func newCachingClient(t *testing.T, inner *mockInnerClient, cache *mockCacheProvider, mode cachefile.FallbackMode) *cachefile.CachingClient {
	t.Helper()
	client, err := cachefile.NewCachingClient(inner, cache, cachefile.Options{
		FallbackMode:     mode,
		DefaultProjectID: testProjectID,
	})
	if err != nil {
		t.Fatalf("NewCachingClient: %v", err)
	}
	return client
}

func TestNewCachingClientValidation(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{}
	cache := newMockCacheProvider()

	if _, err := cachefile.NewCachingClient(nil, cache, cachefile.DefaultOptions()); err == nil {
		t.Fatal("expected error for nil inner")
	}
	if _, err := cachefile.NewCachingClient(inner, nil, cachefile.DefaultOptions()); err == nil {
		t.Fatal("expected error for nil cache")
	}
	if _, err := cachefile.NewCachingClient(inner, cache, cachefile.DefaultOptions()); err == nil {
		t.Fatal("expected error for empty DefaultProjectID")
	}
}

func TestGetEntryCacheFirstHit(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{}
	cache := newMockCacheProvider()
	cache.groups[groupKey(testProjectID, "common", "en")] = testGroup("hello", "Hello World")

	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheFirst)
	got, err := c.GetEntry(context.Background(), "common", "hello", "en")
	if err != nil || got != "Hello World" {
		t.Fatalf("GetEntry: got=%q err=%v", got, err)
	}
	if inner.getEntryCallCount() != 0 {
		t.Fatalf("inner GetEntry calls = %d, want 0", inner.getEntryCallCount())
	}
}

func TestGetEntryCacheFirstMissCallsAPI(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{
		getEntryFn: func(context.Context, string, string, string, ...client.GetEntryOption) (string, error) {
			return "Hello from API", nil
		},
		getGroupFn: func(context.Context, string, string, string, ...client.GetGroupOption) (*models.TranslationGroup, error) {
			return testGroup("hello", "Hello from API"), nil
		},
	}
	cache := newMockCacheProvider()
	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheFirst)

	got, err := c.GetEntry(context.Background(), "common", "hello", "en")
	if err != nil || got != "Hello from API" {
		t.Fatalf("GetEntry: got=%q err=%v", got, err)
	}
	if inner.getEntryCallCount() != 1 {
		t.Fatalf("inner GetEntry calls = %d, want 1", inner.getEntryCallCount())
	}
	if cache.saveProjectCount() != 1 {
		t.Fatalf("SaveProject calls = %d, want 1", cache.saveProjectCount())
	}
}

func TestGetEntryCacheFirstAPIFailureReturnsMiss(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{
		getEntryFn: func(context.Context, string, string, string, ...client.GetEntryOption) (string, error) {
			return "", &models.APIError{StatusCode: http.StatusBadGateway, Message: "network"}
		},
	}
	cache := newMockCacheProvider()
	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheFirst)

	_, err := c.GetEntry(context.Background(), "common", "hello", "en")
	var miss *models.OfflineCacheMissError
	if !errors.As(err, &miss) {
		t.Fatalf("expected OfflineCacheMissError, got %v", err)
	}
}

func TestGetEntryAPIFirstFallsBackToCache(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{
		getEntryFn: func(context.Context, string, string, string, ...client.GetEntryOption) (string, error) {
			return "", &models.APIError{StatusCode: http.StatusBadGateway, Message: "network"}
		},
	}
	cache := newMockCacheProvider()
	cache.groups[groupKey(testProjectID, "common", "en")] = testGroup("hello", "Hello from Cache")

	c := newCachingClient(t, inner, cache, cachefile.FallbackAPIFirst)
	got, err := c.GetEntry(context.Background(), "common", "hello", "en")
	if err != nil || got != "Hello from Cache" {
		t.Fatalf("GetEntry: got=%q err=%v", got, err)
	}
}

func TestGetEntryCacheOnlyNoAPI(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{
		getEntryFn: func(context.Context, string, string, string, ...client.GetEntryOption) (string, error) {
			t.Fatal("API should not be called in CacheOnly mode")
			return "", nil
		},
	}
	cache := newMockCacheProvider()
	cache.groups[groupKey(testProjectID, "common", "en")] = testGroup("hello", "Cached")

	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheOnly)
	got, err := c.GetEntry(context.Background(), "common", "hello", "en")
	if err != nil || got != "Cached" {
		t.Fatalf("GetEntry: got=%q err=%v", got, err)
	}
}

func TestGetEntryCacheOnlyMiss(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{}
	cache := newMockCacheProvider()
	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheOnly)

	_, err := c.GetEntry(context.Background(), "common", "hello", "en")
	var miss *models.OfflineCacheMissError
	if !errors.As(err, &miss) {
		t.Fatalf("expected OfflineCacheMissError, got %v", err)
	}
}

func TestGetEntryCacheOnlyParameterSubstitution(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{}
	cache := newMockCacheProvider()
	cache.groups[groupKey(testProjectID, "messages", "en")] = testGroup("greeting", "Hello {userName}, you have {N} items")

	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheOnly)
	got, err := c.GetEntry(
		context.Background(),
		"messages",
		"greeting",
		"en",
		client.WithNumber(5),
		client.WithParameters(map[string]string{"userName": "John"}),
	)
	if err != nil || got != "Hello John, you have 5 items" {
		t.Fatalf("GetEntry: got=%q err=%v", got, err)
	}
}

func TestGetGroupCacheFirstHit(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{}
	cache := newMockCacheProvider()
	cache.groups[groupKey(testProjectID, "common", "en")] = testGroup("hello", "Hi")

	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheFirst)
	group, err := c.GetGroup(context.Background(), testProjectID, "common", "en")
	if err != nil || group == nil {
		t.Fatalf("GetGroup: err=%v group=%v", err, group)
	}
	if inner.getGroupCalls != 0 {
		t.Fatalf("inner GetGroup calls = %d, want 0", inner.getGroupCalls)
	}
}

func TestGetProjectLocalesCacheOnlyUsesOfflineCacheError(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{}
	cache := newMockCacheProvider()
	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheOnly)

	_, err := c.GetProjectLocales(context.Background(), testProjectID)
	var offlineErr *models.OfflineCacheError
	if !errors.As(err, &offlineErr) {
		t.Fatalf("expected OfflineCacheError, got %v", err)
	}
}

func TestPassthroughValidateAPIKey(t *testing.T) {
	t.Parallel()

	inner := &mockInnerClient{
		validateAPIKeyFn: func(context.Context) (*models.ValidateAPIKeyResponse, error) {
			return &models.ValidateAPIKeyResponse{IsValid: true}, nil
		},
	}
	cache := newMockCacheProvider()
	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheFirst)

	resp, err := c.ValidateAPIKey(context.Background())
	if err != nil || resp == nil || !resp.IsValid {
		t.Fatalf("ValidateAPIKey: resp=%v err=%v", resp, err)
	}
}

func TestCachingClientIntegrationWithFileProvider(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fileProvider, err := cachefile.NewFileProvider(dir)
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}

	project := &models.TranslationProject{
		Groups: map[string]json.RawMessage{
			"common": json.RawMessage(`{"hello":"Offline Hello"}`),
		},
	}
	ctx := context.Background()
	if err := fileProvider.SaveProject(ctx, testProjectID, "en", project); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	inner := &mockInnerClient{
		getEntryFn: func(context.Context, string, string, string, ...client.GetEntryOption) (string, error) {
			t.Fatal("API should not be called")
			return "", nil
		},
	}
	c, err := cachefile.NewCachingClient(inner, fileProvider, cachefile.Options{
		FallbackMode:     cachefile.FallbackCacheOnly,
		DefaultProjectID: testProjectID,
	})
	if err != nil {
		t.Fatalf("NewCachingClient: %v", err)
	}

	got, err := c.GetEntry(ctx, "common", "hello", "en")
	if err != nil || got != "Offline Hello" {
		t.Fatalf("GetEntry: got=%q err=%v", got, err)
	}
}

func TestCachingClientConcurrentGetEntry(t *testing.T) {
	t.Parallel()

	cache := newMockCacheProvider()
	cache.groups[groupKey(testProjectID, "common", "en")] = testGroup("hello", "Hi")
	inner := &mockInnerClient{}
	c := newCachingClient(t, inner, cache, cachefile.FallbackCacheFirst)

	const workers = 16
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, _ = c.GetEntry(context.Background(), "common", "hello", "en")
		}()
	}
	wg.Wait()
}
