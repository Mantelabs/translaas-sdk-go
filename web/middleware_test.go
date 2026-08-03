package web_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/Mantelabs/translaas-sdk-go/models"
	"github.com/Mantelabs/translaas-sdk-go/service"
	"github.com/Mantelabs/translaas-sdk-go/service/language"
	"github.com/Mantelabs/translaas-sdk-go/web"
)

type mockClient struct {
	mu sync.Mutex

	lastLang string
}

func (m *mockClient) GetEntry(_ context.Context, _, _, lang string, _ ...client.GetEntryOption) (string, error) {
	m.mu.Lock()
	m.lastLang = lang
	m.mu.Unlock()
	return "hello", nil
}

func (m *mockClient) GetGroup(context.Context, string, string, string, ...client.GetGroupOption) (*models.TranslationGroup, error) {
	return nil, errors.New("unexpected GetGroup")
}

func (m *mockClient) GetProject(context.Context, string, string, ...client.GetProjectOption) (*models.TranslationProject, error) {
	return nil, errors.New("unexpected GetProject")
}

func (m *mockClient) GetProjectLocales(context.Context, string, ...client.GetProjectLocalesOption) (*models.ProjectLocales, error) {
	return nil, errors.New("unexpected GetProjectLocales")
}

func (m *mockClient) GetOfflineCache(context.Context, string, ...client.GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error) {
	return nil, errors.New("unexpected GetOfflineCache")
}

func (m *mockClient) ReportMissingKeys(context.Context, []models.ReportMissingKeyItem) error {
	return errors.New("unexpected ReportMissingKeys")
}

func (m *mockClient) ValidateAPIKey(context.Context) (*models.ValidateAPIKeyResponse, error) {
	return nil, errors.New("unexpected ValidateAPIKey")
}

func newBaseService(t *testing.T, inner client.Client) *service.Service {
	t.Helper()

	resolver, err := language.NewResolver(language.NewDefaultLanguageProvider("en"))
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	svc, err := service.New(inner, service.Options{Resolver: resolver})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc
}

func TestServiceFromContextMissing(t *testing.T) {
	t.Parallel()

	if _, ok := web.ServiceFromContext(context.Background()); ok {
		t.Fatal("expected missing service")
	}
}

func TestMiddlewareInjectsService(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	base := newBaseService(t, inner)
	mw, err := web.Middleware(web.DefaultMiddlewareOptions(base))
	if err != nil {
		t.Fatalf("Middleware: %v", err)
	}

	var got bool
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := web.ServiceFromContext(r.Context()); !ok {
			t.Fatal("expected service in context")
		}
		got = true
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?lang=de", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !got {
		t.Fatal("handler was not called")
	}
}

func TestMiddlewareResolvesQueryLanguage(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	base := newBaseService(t, inner)
	mw, err := web.Middleware(web.DefaultMiddlewareOptions(base))
	if err != nil {
		t.Fatalf("Middleware: %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svc, ok := web.ServiceFromContext(r.Context())
		if !ok {
			t.Fatal("expected service in context")
		}
		if _, err := svc.T(r.Context(), "ui", "welcome"); err != nil {
			t.Fatalf("T: %v", err)
		}
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?lang=de", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if inner.lastLang != "de" {
		t.Fatalf("lastLang = %q, want de", inner.lastLang)
	}
}

func TestMiddlewareAcceptLanguageFallback(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	base := newBaseService(t, inner)
	mw, err := web.Middleware(web.DefaultMiddlewareOptions(base))
	if err != nil {
		t.Fatalf("Middleware: %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svc, ok := web.ServiceFromContext(r.Context())
		if !ok {
			t.Fatal("expected service in context")
		}
		if _, err := svc.T(r.Context(), "ui", "welcome"); err != nil {
			t.Fatalf("T: %v", err)
		}
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if inner.lastLang != "fr" {
		t.Fatalf("lastLang = %q, want fr", inner.lastLang)
	}
}

func TestMiddlewareRequiresBaseService(t *testing.T) {
	t.Parallel()

	if _, err := web.Middleware(web.MiddlewareOptions{}); err == nil {
		t.Fatal("expected error for missing BaseService")
	}
}

func TestRequestLanguageProviderRouteFunc(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	provider := web.NewRequestLanguageProvider(
		req,
		web.RequestLanguageOptions{},
		[]web.LanguageSource{web.SourceRoute},
		func(*http.Request) string { return "pt" },
	)

	lang, err := provider.Language(context.Background())
	if err != nil {
		t.Fatalf("Language: %v", err)
	}
	if lang != "pt" {
		t.Fatalf("lang = %q, want pt", lang)
	}
}

func TestRequestLanguageProviderCookieSource(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "language", Value: "es"})

	provider := web.NewRequestLanguageProvider(
		req,
		web.RequestLanguageOptions{},
		[]web.LanguageSource{web.SourceCookie},
		nil,
	)

	lang, err := provider.Language(context.Background())
	if err != nil {
		t.Fatalf("Language: %v", err)
	}
	if lang != "es" {
		t.Fatalf("lang = %q, want es", lang)
	}
}
