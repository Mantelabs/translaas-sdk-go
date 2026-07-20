package translaaschi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/models"
	"github.com/acuencadev/translaas-sdk-go/service"
	"github.com/acuencadev/translaas-sdk-go/service/language"
	translaaschi "github.com/acuencadev/translaas-sdk-go/web/chi"
	"github.com/acuencadev/translaas-sdk-go/web"
	"github.com/go-chi/chi/v5"
)

type mockClient struct {
	mu       sync.Mutex
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

func TestMiddlewareInjectsService(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	resolver, err := language.NewResolver(language.NewDefaultLanguageProvider("en"))
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}
	base, err := service.New(inner, service.Options{Resolver: resolver})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	mw, err := translaaschi.Middleware(web.DefaultMiddlewareOptions(base))
	if err != nil {
		t.Fatalf("Middleware: %v", err)
	}

	router := chi.NewRouter()
	router.Use(mw)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		svc, ok := web.ServiceFromContext(r.Context())
		if !ok {
			t.Fatal("expected service in context")
		}
		if _, err := svc.T(r.Context(), "ui", "welcome"); err != nil {
			t.Fatalf("T: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=de", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if inner.lastLang != "de" {
		t.Fatalf("lastLang = %q, want de", inner.lastLang)
	}
}
