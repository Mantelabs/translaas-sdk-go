package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/models"
)

func loadTestdata(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read testdata %s: %v", name, err)
	}
	return data
}

func TestGetGroup_Success(t *testing.T) {
	t.Parallel()
	fixture := loadTestdata(t, "translation_group_full_api.json")
	var requestCount int

	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if !strings.HasSuffix(r.URL.Path, "/sdk/v1/translations/group") {
			t.Fatalf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("project") != "my-project" || q.Get("group") != "ui" || q.Get("lang") != "en" {
			t.Fatalf("query = %v", q)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("Accept = %q", r.Header.Get("Accept"))
		}
		w.Header().Set("ETag", `W/"grp-etag"`)
		_, _ = w.Write(fixture)
	})

	got, err := cli.GetGroup(context.Background(), "my-project", "ui", "en")
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}
	if got.Project != "my-project" || got.Lang != "en" {
		t.Fatalf("group metadata = %+v", got)
	}
	value, ok := got.GetValue("welcome")
	if !ok || value != "Welcome" {
		t.Fatalf("welcome = (%q, %v)", value, ok)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d", requestCount)
	}
}

func TestGetGroup_FlatFormat(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("format") != "flat-json" {
			t.Fatalf("format = %q", r.URL.Query().Get("format"))
		}
		_, _ = io.WriteString(w, `{"title":"Checkout"}`)
	})

	got, err := cli.GetGroup(context.Background(), "p", "g", "en", WithGroupFormat("flat-json"))
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}
	value, ok := got.GetValue("title")
	if !ok || value != "Checkout" {
		t.Fatalf("title = (%q, %v)", value, ok)
	}
}

func TestGetGroup_WithRequestContext(t *testing.T) {
	t.Parallel()
	include := true
	reqCtx := &models.RequestContext{
		Channel:        "canary",
		Version:        "42",
		IncludeContext: &include,
		IfNoneMatch:    `W/"old"`,
	}

	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("channel") != "canary" || q.Get("v") != "42" || q.Get("includeContext") != "true" {
			t.Fatalf("query = %v", q)
		}
		if r.Header.Get("If-None-Match") != `W/"old"` {
			t.Fatalf("If-None-Match = %q", r.Header.Get("If-None-Match"))
		}
		w.Header().Set("ETag", `W/"grp-new"`)
		_, _ = io.WriteString(w, `{"Entries":{"k":"v"}}`)
	})

	_, err := cli.GetGroup(context.Background(), "p", "g", "en", WithGroupRequestContext(reqCtx))
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}
	if reqCtx.ResponseETag != `W/"grp-new"` {
		t.Fatalf("ResponseETag = %q", reqCtx.ResponseETag)
	}
}

func TestGetGroup_NoContent(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	got, err := cli.GetGroup(context.Background(), "p", "g", "en")
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}
	if len(got.Entries) != 0 {
		t.Fatalf("entries = %v", got.Entries)
	}
}

func TestGetGroup_NotModified(t *testing.T) {
	t.Parallel()
	reqCtx := &models.RequestContext{IfNoneMatch: `W/"old"`}
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `W/"new"`)
		w.WriteHeader(http.StatusNotModified)
	})

	got, err := cli.GetGroup(context.Background(), "p", "g", "en", WithGroupRequestContext(reqCtx))
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}
	if !reqCtx.NotModified {
		t.Fatal("expected NotModified")
	}
	if reqCtx.ResponseETag != `W/"new"` {
		t.Fatalf("ResponseETag = %q", reqCtx.ResponseETag)
	}
	if len(got.Entries) != 0 {
		t.Fatalf("entries = %v", got.Entries)
	}
}

func TestGetGroup_NotFound(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"message":"not found","code":"NOT_FOUND"}`)
	})

	_, err := cli.GetGroup(context.Background(), "p", "g", "en")
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 APIError, got %v", err)
	}
}

func TestGetProject_Success(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sdk/v1/translations/project") {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"common":{"hello":"Hello"}}`)
	})

	got, err := cli.GetProject(context.Background(), "my-app", "en")
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}
	group, err := got.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}
	value, ok := group.GetValue("hello")
	if !ok || value != "Hello" {
		t.Fatalf("hello = (%q, %v)", value, ok)
	}
}

func TestGetProjectLocales_Success(t *testing.T) {
	t.Parallel()
	fixture := loadTestdata(t, "project_locales.json")
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sdk/v1/translations/locales") {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("project") != "my-app" {
			t.Fatalf("project = %q", r.URL.Query().Get("project"))
		}
		_, _ = w.Write(fixture)
	})

	got, err := cli.GetProjectLocales(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("GetProjectLocales() error = %v", err)
	}
	if len(got.Locales) != 4 {
		t.Fatalf("locales = %v", got.Locales)
	}
}

func TestGetProjectLocales_NotModified(t *testing.T) {
	t.Parallel()
	reqCtx := &models.RequestContext{}
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	})

	got, err := cli.GetProjectLocales(context.Background(), "my-app", WithLocalesRequestContext(reqCtx))
	if err != nil {
		t.Fatalf("GetProjectLocales() error = %v", err)
	}
	if !reqCtx.NotModified {
		t.Fatal("expected NotModified")
	}
	if got.Locales == nil || len(got.Locales) != 0 {
		t.Fatalf("locales = %v", got.Locales)
	}
}

func TestGetOfflineCache_Success(t *testing.T) {
	t.Parallel()
	zipBytes := []byte("PK\x03\x04fake-zip")
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sdk/v1/translations/offline-cache") {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/zip" {
			t.Fatalf("Accept = %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Disposition", `attachment; filename="bundle.zip"`)
		w.Header().Set("ETag", `W/"offline-etag"`)
		_, _ = w.Write(zipBytes)
	})

	got, err := cli.GetOfflineCache(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("GetOfflineCache() error = %v", err)
	}
	if string(got.Content) != string(zipBytes) {
		t.Fatalf("content = %q", got.Content)
	}
	if got.SuggestedFileName != "bundle.zip" {
		t.Fatalf("filename = %q", got.SuggestedFileName)
	}
	if got.ETag != `W/"offline-etag"` {
		t.Fatalf("etag = %q", got.ETag)
	}
}

func TestGetOfflineCache_FilenameStar(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''my%20bundle.zip`)
		_, _ = io.WriteString(w, "zip")
	})

	got, err := cli.GetOfflineCache(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("GetOfflineCache() error = %v", err)
	}
	if got.SuggestedFileName != "my bundle.zip" {
		t.Fatalf("filename = %q", got.SuggestedFileName)
	}
}

func TestGetOfflineCache_NotModified(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `W/"offline"`)
		w.WriteHeader(http.StatusNotModified)
	})

	got, err := cli.GetOfflineCache(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("GetOfflineCache() error = %v", err)
	}
	if !got.NotModified || got.Content != nil {
		t.Fatalf("result = %+v", got)
	}
}

func TestReportMissingKeys_EmptyNoOp(t *testing.T) {
	t.Parallel()
	called := false
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	err := cli.ReportMissingKeys(context.Background(), nil)
	if err != nil {
		t.Fatalf("ReportMissingKeys() error = %v", err)
	}
	if called {
		t.Fatal("expected no HTTP request for empty keys")
	}
}

func TestReportMissingKeys_Success(t *testing.T) {
	t.Parallel()
	var method, body string
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.WriteHeader(http.StatusAccepted)
	})

	err := cli.ReportMissingKeys(context.Background(), []models.ReportMissingKeyItem{{
		GroupKey: "g", EntryKey: "k", LanguageIsoCode: "en",
	}})
	if err != nil {
		t.Fatalf("ReportMissingKeys() error = %v", err)
	}
	if method != http.MethodPost {
		t.Fatalf("method = %q", method)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	keys, ok := decoded["keys"].([]any)
	if !ok || len(keys) != 1 {
		t.Fatalf("body = %s", body)
	}
}

func TestReportMissingKeys_ValidationError(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	err := cli.ReportMissingKeys(context.Background(), []models.ReportMissingKeyItem{{
		GroupKey: "g", EntryKey: "k", LanguageIsoCode: "en",
	}})
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 APIError, got %v", err)
	}
}

func TestValidateAPIKey_Success(t *testing.T) {
	t.Parallel()
	fixture := loadTestdata(t, "validate_api_key_tenant.json")
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/api/v1/api-keys/validate") {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_, _ = w.Write(fixture)
	})

	got, err := cli.ValidateAPIKey(context.Background())
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}
	if !got.IsValid || len(got.ProjectIDs) != 2 {
		t.Fatalf("response = %+v", got)
	}
}

func TestValidateAPIKey_Unauthorized(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := cli.ValidateAPIKey(context.Background())
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 APIError, got %v", err)
	}
}

func TestNewWithResolvedProject_SingleProjectKey(t *testing.T) {
	t.Parallel()
	requests := 0
	srv := httptestNewServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch {
		case strings.HasSuffix(r.URL.Path, "/api/v1/api-keys/validate"):
			_, _ = io.WriteString(w, `{"isValid":true,"projectId":"01PROJECTULID123456789012"}`)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	})

	cli, err := NewWithResolvedProject(context.Background(), Options{
		APIKey:  "test-api-key",
		BaseURL: srv.URL,
	}, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("NewWithResolvedProject() error = %v", err)
	}
	impl := cli.(*client)
	if impl.defaultProjectID != "01PROJECTULID123456789012" {
		t.Fatalf("defaultProjectID = %q", impl.defaultProjectID)
	}
}

func TestNewWithResolvedProject_TenantKey(t *testing.T) {
	t.Parallel()
	srv := httptestNewServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"isValid":true,"tenantId":"01TENANT"}`)
	})

	cli, err := NewWithResolvedProject(context.Background(), Options{
		APIKey:  "test-api-key",
		BaseURL: srv.URL,
	}, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("NewWithResolvedProject() error = %v", err)
	}
	impl := cli.(*client)
	if impl.defaultProjectID != "" {
		t.Fatalf("defaultProjectID = %q, want empty", impl.defaultProjectID)
	}
}

func TestNewWithResolvedProject_Preconfigured(t *testing.T) {
	t.Parallel()
	called := false
	srv := httptestNewServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cli, err := NewWithResolvedProject(context.Background(), Options{
		APIKey:           "test-api-key",
		BaseURL:          srv.URL,
		DefaultProjectID: "preset",
	}, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("NewWithResolvedProject() error = %v", err)
	}
	impl := cli.(*client)
	if impl.defaultProjectID != "preset" {
		t.Fatalf("defaultProjectID = %q", impl.defaultProjectID)
	}
	if called {
		t.Fatal("expected no validate call when DefaultProjectID preset")
	}
}

func TestGetGroup_RequiredFields(t *testing.T) {
	t.Parallel()
	cli, err := New(Options{APIKey: "key", BaseURL: "https://api.test.com"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := cli.GetGroup(context.Background(), "", "g", "en"); err == nil {
		t.Fatal("expected project error")
	}
	if _, err := cli.GetGroup(context.Background(), "p", "", "en"); err == nil {
		t.Fatal("expected group error")
	}
}

func TestGetProject_WithOptions(t *testing.T) {
	t.Parallel()
	include := true
	reqCtx := &models.RequestContext{Channel: "stable", Version: "1", IncludeContext: &include}
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("format") != "flat-json" || q.Get("channel") != "stable" {
			t.Fatalf("query = %v", q)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	got, err := cli.GetProject(context.Background(), "p", "en",
		WithProjectFormat("flat-json"),
		WithProjectRequestContext(reqCtx),
	)
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}
	if len(got.Groups) != 0 {
		t.Fatalf("groups = %v", got.Groups)
	}
}

func TestGetProject_NotModified(t *testing.T) {
	t.Parallel()
	reqCtx := &models.RequestContext{}
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	})

	got, err := cli.GetProject(context.Background(), "p", "en", WithProjectRequestContext(reqCtx))
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}
	if !reqCtx.NotModified || len(got.Groups) != 0 {
		t.Fatalf("result = %+v, NotModified=%v", got, reqCtx.NotModified)
	}
}

func TestGetOfflineCache_WithRequestContext(t *testing.T) {
	t.Parallel()
	include := false
	reqCtx := &models.RequestContext{Channel: "canary", IncludeContext: &include, IfNoneMatch: `W/"x"`}
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("channel") != "canary" || q.Get("includeContext") != "false" {
			t.Fatalf("query = %v", q)
		}
		if r.Header.Get("If-None-Match") != `W/"x"` {
			t.Fatalf("If-None-Match = %q", r.Header.Get("If-None-Match"))
		}
		_, _ = io.WriteString(w, "zip")
	})

	_, err := cli.GetOfflineCache(context.Background(), "p", WithOfflineCacheRequestContext(reqCtx))
	if err != nil {
		t.Fatalf("GetOfflineCache() error = %v", err)
	}
}

func TestGetOfflineCache_NotFound(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := cli.GetOfflineCache(context.Background(), "missing")
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 APIError, got %v", err)
	}
}

func TestGetProjectLocales_NoContent(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	got, err := cli.GetProjectLocales(context.Background(), "p")
	if err != nil {
		t.Fatalf("GetProjectLocales() error = %v", err)
	}
	if len(got.Locales) != 0 {
		t.Fatalf("locales = %v", got.Locales)
	}
}

func TestParseFilenameStar_Fallback(t *testing.T) {
	t.Parallel()
	got := parseFilenameStar(`attachment; filename*=plain-name.zip`)
	if got != "plain-name.zip" {
		t.Fatalf("got %q", got)
	}
}

func httptestNewServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}
