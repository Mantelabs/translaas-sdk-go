package cachefile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/cachefile"
	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/Mantelabs/translaas-sdk-go/models"
)

func TestSyncFromOfflineZipImportsBundle(t *testing.T) {
	ctx := context.Background()
	zipBytes := buildTestOfflineZIP(t, nil)

	inner := &syncMockClient{
		getOfflineCacheFn: func(_ context.Context, project string, _ ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error) {
			if project != "demo-project" {
				t.Fatalf("unexpected project %q", project)
			}
			return &models.OfflineCacheDownloadResult{Content: zipBytes}, nil
		},
	}

	provider := newTestProvider(t)
	svc := newSyncService(t, inner, provider, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{})

	if err := svc.SyncFromOfflineZip(ctx, "demo-project"); err != nil {
		t.Fatalf("SyncFromOfflineZip: %v", err)
	}

	got, err := provider.GetProject(ctx, "demo-project", "en")
	if err != nil || got == nil {
		t.Fatalf("GetProject: err=%v got=%v", err, got)
	}
}

func TestSyncFromOfflineZipNotModifiedNoOp(t *testing.T) {
	ctx := context.Background()
	called := false
	inner := &syncMockClient{
		getOfflineCacheFn: func(context.Context, string, ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error) {
			called = true
			return &models.OfflineCacheDownloadResult{NotModified: true}, nil
		},
	}

	provider := newTestProvider(t)
	svc := newSyncService(t, inner, provider, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{})

	if err := svc.SyncFromOfflineZip(ctx, "demo-project"); err != nil {
		t.Fatalf("SyncFromOfflineZip: %v", err)
	}
	if !called {
		t.Fatal("expected GetOfflineCache to be called")
	}

	got, err := provider.GetProject(ctx, "demo-project", "en")
	if err != nil || got != nil {
		t.Fatalf("expected empty cache, err=%v got=%v", err, got)
	}
}

func TestSyncFromOfflineZipRaisesFailedCallback(t *testing.T) {
	ctx := context.Background()
	downloadErr := errors.New("download failed")
	inner := &syncMockClient{
		getOfflineCacheFn: func(context.Context, string, ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error) {
			return nil, downloadErr
		},
	}

	provider := newTestProvider(t)
	var failed cachefile.SyncFailedEvent
	svc := newSyncService(t, inner, provider, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{
		OnSyncFailed: func(event cachefile.SyncFailedEvent) {
			failed = event
		},
	})

	if err := svc.SyncFromOfflineZip(ctx, "demo-project"); !errors.Is(err, downloadErr) {
		t.Fatalf("SyncFromOfflineZip err = %v, want %v", err, downloadErr)
	}
	if failed.Project != "demo-project" || !errors.Is(failed.Err, downloadErr) {
		t.Fatalf("failed event = %+v", failed)
	}
}

func TestSyncFromOfflineZipGenericProviderFallback(t *testing.T) {
	ctx := context.Background()
	zipBytes := buildTestOfflineZIP(t, nil)

	inner := &syncMockClient{
		getOfflineCacheFn: func(context.Context, string, ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error) {
			return &models.OfflineCacheDownloadResult{Content: zipBytes}, nil
		},
	}
	cache := newSyncMockCache()
	svc := newSyncService(t, inner, cache, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{})

	if err := svc.SyncFromOfflineZip(ctx, "demo-project"); err != nil {
		t.Fatalf("SyncFromOfflineZip: %v", err)
	}
	if cache.saveProjectCallCount() != 2 {
		t.Fatalf("SaveProject calls = %d, want 2", cache.saveProjectCallCount())
	}
	if cache.saveLocalesCallCount() != 1 {
		t.Fatalf("SaveLocales calls = %d, want 1", cache.saveLocalesCallCount())
	}
}
