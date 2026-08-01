package cachefile_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Mantelabs/translaas-sdk-go/cachefile"
	"github.com/Mantelabs/translaas-sdk-go/models"
)

func newTestProvider(t *testing.T) *cachefile.FileProvider {
	t.Helper()
	dir := t.TempDir()
	provider, err := cachefile.NewFileProvider(dir)
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}
	return provider
}

func sampleProject(t *testing.T) *models.TranslationProject {
	t.Helper()
	return &models.TranslationProject{
		Groups: map[string]json.RawMessage{
			"common": json.RawMessage(`{"hello":"Hello","bye":"Bye"}`),
		},
	}
}

func TestFileProviderRoundTripProject(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	project := sampleProject(t)

	if err := provider.SaveProject(ctx, "demo-project", "en", project); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	got, err := provider.GetProject(ctx, "demo-project", "en")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got == nil {
		t.Fatal("expected cached project")
	}

	group, err := got.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	value, ok := group.GetValue("hello")
	if !ok || value != "Hello" {
		t.Fatalf("got hello=%q ok=%v", value, ok)
	}
}

func TestFileProviderGetGroup(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	if err := provider.SaveProject(ctx, "demo-project", "en", sampleProject(t)); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	group, err := provider.GetGroup(ctx, "demo-project", "common", "en")
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if group == nil {
		t.Fatal("expected group")
	}

	missing, err := provider.GetGroup(ctx, "demo-project", "missing", "en")
	if err != nil {
		t.Fatalf("GetGroup missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil group, got %+v", missing)
	}
}

func TestFileProviderRoundTripLocales(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	locales := &models.ProjectLocales{
		Project: "demo-project",
		Locales: []string{"en", "de"},
	}

	if err := provider.SaveLocales(ctx, "demo-project", locales); err != nil {
		t.Fatalf("SaveLocales: %v", err)
	}

	got, err := provider.GetLocales(ctx, "demo-project")
	if err != nil {
		t.Fatalf("GetLocales: %v", err)
	}
	if got == nil {
		t.Fatal("expected locales")
	}
	if len(got.Locales) != 2 || got.Locales[0] != "en" || got.Locales[1] != "de" {
		t.Fatalf("unexpected locales: %+v", got.Locales)
	}
}

func TestFileProviderExpiredEntry(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	past := time.Now().UTC().Add(-time.Hour)

	if err := provider.SaveProject(ctx, "demo-project", "en", sampleProject(t), cachefile.WithExpiresAt(&past)); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	got, err := provider.GetProject(ctx, "demo-project", "en")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got != nil {
		t.Fatal("expected expired project to miss")
	}

	cached, err := provider.IsCached(ctx, "demo-project", "en")
	if err != nil {
		t.Fatalf("IsCached: %v", err)
	}
	if cached {
		t.Fatal("expected expired entry to not be cached")
	}
}

func TestFileProviderMissingFiles(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()

	got, err := provider.GetProject(ctx, "missing", "en")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got != nil {
		t.Fatal("expected miss")
	}
}

func TestFileProviderCorruptJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	provider, err := cachefile.NewFileProvider(dir)
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}

	projectDir := filepath.Join(dir, "demo-project", "en")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "project.json"), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err = provider.GetProject(context.Background(), "demo-project", "en")
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
	var cacheErr *models.OfflineCacheError
	if !errors.As(err, &cacheErr) {
		t.Fatalf("expected OfflineCacheError, got %T: %v", err, err)
	}
}

func TestFileProviderManifestAfterSave(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	if err := provider.SaveProject(ctx, "demo-project", "en", sampleProject(t)); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	manifest, err := provider.GetManifest(ctx)
	if err != nil {
		t.Fatalf("GetManifest: %v", err)
	}
	if manifest == nil {
		t.Fatal("expected manifest")
	}
	if manifest.Version != cachefile.ManifestVersion {
		t.Fatalf("version=%q", manifest.Version)
	}
	info, ok := manifest.Projects["demo-project"]
	if !ok {
		t.Fatalf("projects=%v", manifest.Projects)
	}
	if len(info.Languages) != 1 || info.Languages[0] != "en" {
		t.Fatalf("languages=%v", info.Languages)
	}
}

func TestFileProviderClear(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	if err := provider.SaveProject(ctx, "demo-project", "en", sampleProject(t)); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	if err := provider.Clear(ctx); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	got, err := provider.GetProject(ctx, "demo-project", "en")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got != nil {
		t.Fatal("expected miss after clear")
	}
}

func TestFileProviderContextCancelled(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := provider.SaveProject(ctx, "demo-project", "en", sampleProject(t))
	if err == nil {
		t.Fatal("expected error")
	}
	var cacheErr *models.OfflineCacheError
	if !errors.As(err, &cacheErr) {
		t.Fatalf("expected OfflineCacheError, got %T", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestFileProviderAtomicWriteReplacesExisting(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	if err := provider.SaveProject(ctx, "demo-project", "en", sampleProject(t)); err != nil {
		t.Fatalf("first SaveProject: %v", err)
	}

	updated := &models.TranslationProject{
		Groups: map[string]json.RawMessage{
			"common": json.RawMessage(`{"hello":"Updated"}`),
		},
	}
	if err := provider.SaveProject(ctx, "demo-project", "en", updated); err != nil {
		t.Fatalf("second SaveProject: %v", err)
	}

	got, err := provider.GetProject(ctx, "demo-project", "en")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	group, err := got.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	value, ok := group.GetValue("hello")
	if !ok || value != "Updated" {
		t.Fatalf("got hello=%q ok=%v", value, ok)
	}

	raw, err := os.ReadFile(filepath.Join(provider.CacheDirectory(), "demo-project", "en", "project.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !json.Valid(raw) {
		t.Fatalf("invalid JSON on disk: %s", raw)
	}
}

func TestFileProviderLocalesFallbackManifest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	provider, err := cachefile.NewFileProvider(dir)
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}

	manifest := cachefile.CacheManifest{
		Version:    cachefile.ManifestVersion,
		SDKVersion: cachefile.DefaultSDKVersion,
		CreatedAt:  time.Now().UTC(),
		LastSyncAt: time.Now().UTC(),
		Projects: map[string]cachefile.ProjectCacheInfo{
			"demo-project": {
				Languages:  []string{"en", "fr"},
				LastSyncAt: time.Now().UTC(),
				Status:     "synced",
			},
		},
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := provider.GetLocales(context.Background(), "demo-project")
	if err != nil {
		t.Fatalf("GetLocales: %v", err)
	}
	if got == nil || len(got.Locales) != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestFileProviderLocalesFallbackDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	provider, err := cachefile.NewFileProvider(dir)
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}

	projectDir := filepath.Join(dir, "demo-project", "de")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	wrapped := cachefile.CachedProject{
		CachedAt: time.Now().UTC(),
		Data:     *sampleProject(t),
	}
	raw, err := json.MarshalIndent(wrapped, "", "  ")
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "project.json"), raw, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := provider.GetLocales(context.Background(), "demo-project")
	if err != nil {
		t.Fatalf("GetLocales: %v", err)
	}
	if got == nil || len(got.Locales) != 1 || got.Locales[0] != "de" {
		t.Fatalf("got %+v", got)
	}
}

func TestFileProviderConcurrentAccess(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	var wg sync.WaitGroup

	for _, lang := range []string{"en", "de", "fr"} {
		lang := lang
		wg.Add(1)
		go func() {
			defer wg.Done()
			project := &models.TranslationProject{
				Groups: map[string]json.RawMessage{
					"common": json.RawMessage(`{"hello":"` + lang + `"}`),
				},
			}
			if err := provider.SaveProject(ctx, "demo-project", lang, project); err != nil {
				t.Errorf("SaveProject(%s): %v", lang, err)
			}
		}()
	}
	wg.Wait()

	for _, lang := range []string{"en", "de", "fr"} {
		got, err := provider.GetProject(ctx, "demo-project", lang)
		if err != nil {
			t.Fatalf("GetProject(%s): %v", lang, err)
		}
		if got == nil {
			t.Fatalf("miss for %s", lang)
		}
	}
}

func TestNewFileProviderResolvesAbsolutePath(t *testing.T) {
	t.Parallel()

	rel := "relative-cache-dir"
	provider, err := cachefile.NewFileProvider(rel)
	if err != nil {
		t.Fatalf("NewFileProvider: %v", err)
	}
	if !filepath.IsAbs(provider.CacheDirectory()) {
		t.Fatalf("expected absolute path, got %q", provider.CacheDirectory())
	}
	_ = os.RemoveAll(provider.CacheDirectory())
}
