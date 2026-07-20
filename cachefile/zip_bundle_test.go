package cachefile_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/acuencadev/translaas-sdk-go/cachefile"
	"github.com/acuencadev/translaas-sdk-go/models"
)

func buildTestOfflineZIP(t *testing.T, mutate func(w *zip.Writer)) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	manifest := map[string]any{
		"version":    "1.0",
		"sdkVersion": "1.0.0",
		"createdAt":  "2026-01-01T00:00:00Z",
		"lastSyncAt": "2026-01-01T00:00:00Z",
		"projects": map[string]any{
			"demo-project": map[string]any{
				"languages":  []string{"en", "de"},
				"lastSyncAt": "2026-01-01T00:00:00Z",
				"status":     "synced",
			},
		},
	}
	writeZipJSON(t, w, "manifest.json", manifest)

	localesWrapper := map[string]any{
		"cachedAt": "2026-01-01T00:00:00Z",
		"data":     map[string]any{"locales": []string{"en", "de"}},
	}
	writeZipJSON(t, w, "demo-project/locales.json", localesWrapper)

	enProject := map[string]any{
		"cachedAt": "2026-01-01T00:00:00Z",
		"data":     map[string]any{"common": map[string]any{"hello": "Hello"}},
	}
	deProject := map[string]any{
		"cachedAt": "2026-01-01T00:00:00Z",
		"data":     map[string]any{"common": map[string]any{"hello": "Hallo"}},
	}
	writeZipJSON(t, w, "demo-project/en/project.json", enProject)
	writeZipJSON(t, w, "demo-project/de/project.json", deProject)

	if mutate != nil {
		mutate(w)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func writeZipJSON(t *testing.T, w *zip.Writer, name string, value any) {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %s: %v", name, err)
	}
	if err := writeZipEntry(w, name, payload); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func writeZipEntry(w *zip.Writer, name string, payload []byte) error {
	entry, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = entry.Write(payload)
	return err
}

func TestParseOfflineZip(t *testing.T) {
	t.Parallel()

	bundle, err := cachefile.ParseOfflineZip(buildTestOfflineZIP(t, nil))
	if err != nil {
		t.Fatalf("ParseOfflineZip: %v", err)
	}
	if bundle.Manifest.Version != "1.0" {
		t.Fatalf("manifest version = %q, want 1.0", bundle.Manifest.Version)
	}
	if len(bundle.LocalesByProject["demo-project"].Data.Locales) != 2 {
		t.Fatalf("locales = %+v", bundle.LocalesByProject["demo-project"].Data.Locales)
	}

	en := bundle.ProjectsByProjectLang["demo-project"]["en"].Data
	group, err := en.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	value, ok := group.GetValue("hello")
	if !ok || value != "Hello" {
		t.Fatalf("hello = %q ok=%v", value, ok)
	}

	de := bundle.ProjectsByProjectLang["demo-project"]["de"].Data
	group, err = de.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup de: %v", err)
	}
	value, ok = group.GetValue("hello")
	if !ok || value != "Hallo" {
		t.Fatalf("de hello = %q ok=%v", value, ok)
	}
}

func TestParseOfflineZipEmpty(t *testing.T) {
	t.Parallel()

	_, err := cachefile.ParseOfflineZip(nil)
	if err == nil {
		t.Fatal("expected error for empty ZIP content")
	}
	var cacheErr *models.OfflineCacheError
	if !errors.As(err, &cacheErr) {
		t.Fatalf("expected OfflineCacheError, got %T", err)
	}
}

func TestParseOfflineZipCorrupt(t *testing.T) {
	t.Parallel()

	_, err := cachefile.ParseOfflineZip([]byte("not-a-zip"))
	if err == nil {
		t.Fatal("expected error for corrupt ZIP")
	}
}

func TestParseOfflineZipZipSlipRejected(t *testing.T) {
	t.Parallel()

	content := buildTestOfflineZIP(t, func(w *zip.Writer) {
		if err := writeZipEntry(w, "../evil.json", []byte(`{}`)); err != nil {
			t.Fatalf("write evil entry: %v", err)
		}
	})

	_, err := cachefile.ParseOfflineZip(content)
	if err == nil {
		t.Fatal("expected zip-slip rejection")
	}
}

func TestParseOfflineZipUnknownManifestVersion(t *testing.T) {
	t.Parallel()

	content := buildTestOfflineZIP(t, func(w *zip.Writer) {
		writeZipJSON(t, w, "manifest.json", map[string]any{
			"version":    "9.9",
			"sdkVersion": "1.0.0",
			"createdAt":  "2026-01-01T00:00:00Z",
			"lastSyncAt": "2026-01-01T00:00:00Z",
			"projects":   map[string]any{},
		})
	})

	bundle, err := cachefile.ParseOfflineZip(content)
	if err != nil {
		t.Fatalf("ParseOfflineZip: %v", err)
	}
	if bundle.Manifest.Version != "9.9" {
		t.Fatalf("version = %q", bundle.Manifest.Version)
	}
}

func TestResolveProjectKey(t *testing.T) {
	t.Parallel()

	bundle, err := cachefile.ParseOfflineZip(buildTestOfflineZIP(t, nil))
	if err != nil {
		t.Fatalf("ParseOfflineZip: %v", err)
	}

	key, err := cachefile.ResolveProjectKey(bundle, "demo-project")
	if err != nil || key != "demo-project" {
		t.Fatalf("ResolveProjectKey = %q err=%v", key, err)
	}
}

func TestResolveProjectKeySanitizedFolder(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	writeZipJSON(t, w, "manifest.json", map[string]any{
		"version": "1.0", "projects": map[string]any{},
	})
	writeZipJSON(t, w, "my_project/locales.json", map[string]any{
		"cachedAt": "2026-01-01T00:00:00Z",
		"data":     map[string]any{"locales": []string{"en"}},
	})
	writeZipJSON(t, w, "my_project/en/project.json", map[string]any{
		"cachedAt": "2026-01-01T00:00:00Z",
		"data":     map[string]any{"common": map[string]any{"hello": "Hi"}},
	})
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	bundle, err := cachefile.ParseOfflineZip(buf.Bytes())
	if err != nil {
		t.Fatalf("ParseOfflineZip: %v", err)
	}

	key, err := cachefile.ResolveProjectKey(bundle, "my/project")
	if err != nil || key != "my_project" {
		t.Fatalf("ResolveProjectKey = %q err=%v", key, err)
	}
}

func TestImportOfflineBundle(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx := context.Background()
	content := buildTestOfflineZIP(t, nil)

	if err := provider.ImportOfflineBundle(ctx, "demo-project", content); err != nil {
		t.Fatalf("ImportOfflineBundle: %v", err)
	}

	got, err := provider.GetProject(ctx, "demo-project", "en")
	if err != nil || got == nil {
		t.Fatalf("GetProject en: err=%v got=%v", err, got)
	}
	group, err := got.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	value, ok := group.GetValue("hello")
	if !ok || value != "Hello" {
		t.Fatalf("hello = %q ok=%v", value, ok)
	}

	locales, err := provider.GetLocales(ctx, "demo-project")
	if err != nil || locales == nil {
		t.Fatalf("GetLocales: err=%v got=%v", err, locales)
	}
	if len(locales.Locales) != 2 {
		t.Fatalf("locales = %+v", locales.Locales)
	}

	group, err = provider.GetGroup(ctx, "demo-project", "common", "de")
	if err != nil || group == nil {
		t.Fatalf("GetGroup de: err=%v group=%v", err, group)
	}

	manifest, err := provider.GetManifest(ctx)
	if err != nil || manifest == nil {
		t.Fatalf("GetManifest: err=%v manifest=%v", err, manifest)
	}
	info, ok := manifest.Projects["demo-project"]
	if !ok {
		t.Fatalf("manifest projects = %+v", manifest.Projects)
	}
	if info.Status != "synced" || len(info.Languages) != 2 {
		t.Fatalf("project info = %+v", info)
	}
}

func TestImportOfflineBundlePreservesExpiry(t *testing.T) {
	t.Parallel()

	past := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	content := buildTestOfflineZIP(t, func(w *zip.Writer) {
		writeZipJSON(t, w, "demo-project/en/project.json", map[string]any{
			"cachedAt":  "2026-01-01T00:00:00Z",
			"expiresAt": past,
			"data":      map[string]any{"common": map[string]any{"hello": "Hello"}},
		})
	})

	provider := newTestProvider(t)
	ctx := context.Background()
	if err := provider.ImportOfflineBundle(ctx, "demo-project", content); err != nil {
		t.Fatalf("ImportOfflineBundle: %v", err)
	}

	got, err := provider.GetProject(ctx, "demo-project", "en")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got != nil {
		t.Fatal("expected expired import to miss on read")
	}
}

func TestImportOfflineBundleMultiProjectIsolation(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	writeZipJSON(t, w, "manifest.json", map[string]any{"version": "1.0", "projects": map[string]any{}})
	writeZipJSON(t, w, "project-a/en/project.json", map[string]any{
		"cachedAt": "2026-01-01T00:00:00Z",
		"data":     map[string]any{"common": map[string]any{"hello": "A"}},
	})
	writeZipJSON(t, w, "project-b/en/project.json", map[string]any{
		"cachedAt": "2026-01-01T00:00:00Z",
		"data":     map[string]any{"common": map[string]any{"hello": "B"}},
	})
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	provider := newTestProvider(t)
	ctx := context.Background()
	if err := provider.ImportOfflineBundle(ctx, "project-a", buf.Bytes()); err != nil {
		t.Fatalf("ImportOfflineBundle: %v", err)
	}

	gotA, err := provider.GetProject(ctx, "project-a", "en")
	if err != nil || gotA == nil {
		t.Fatalf("GetProject project-a: err=%v got=%v", err, gotA)
	}
	gotB, err := provider.GetProject(ctx, "project-b", "en")
	if err != nil {
		t.Fatalf("GetProject project-b: %v", err)
	}
	if gotB != nil {
		t.Fatal("expected project-b to remain uncached")
	}
}

func TestImportOfflineBundleContextCancelled(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := provider.ImportOfflineBundle(ctx, "demo-project", buildTestOfflineZIP(t, nil))
	if err == nil {
		t.Fatal("expected cancelled context error")
	}
	var cacheErr *models.OfflineCacheError
	if !errors.As(err, &cacheErr) {
		t.Fatalf("expected OfflineCacheError, got %T", err)
	}
	if !errors.Is(cacheErr, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", cacheErr.Cause)
	}
}
