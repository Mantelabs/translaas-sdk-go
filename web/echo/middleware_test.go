package translaasecho_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/Mantelabs/translaas-sdk-go/models"
	"github.com/Mantelabs/translaas-sdk-go/service"
	"github.com/Mantelabs/translaas-sdk-go/service/language"
	"github.com/Mantelabs/translaas-sdk-go/web"
	translaasecho "github.com/Mantelabs/translaas-sdk-go/web/echo"
	"github.com/labstack/echo/v4"
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

func TestMiddlewareAndTemplateFunc(t *testing.T) {
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

	e := echo.New()
	mw, err := translaasecho.Middleware(web.DefaultMiddlewareOptions(base))
	if err != nil {
		t.Fatalf("Middleware: %v", err)
	}
	e.Use(mw)
	e.GET("/", func(c echo.Context) error {
		fn := translaasecho.TemplateFunc(c)
		text, err := fn("ui", "welcome")
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, text)
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?lang=de", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if inner.lastLang != "de" {
		t.Fatalf("lastLang = %q, want de", inner.lastLang)
	}
}

func TestMiddlewareAndT(t *testing.T) {
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

	e := echo.New()
	mw, err := translaasecho.Middleware(web.DefaultMiddlewareOptions(base))
	if err != nil {
		t.Fatalf("Middleware: %v", err)
	}
	e.Use(mw)
	e.GET("/", func(c echo.Context) error {
		text, err := translaasecho.T(c, "ui", "welcome")
		if err != nil {
			return err
		}
		if text != "hello" {
			t.Fatalf("text = %q, want hello", text)
		}
		return c.String(http.StatusOK, text)
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?lang=de", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if inner.lastLang != "de" {
		t.Fatalf("lastLang = %q, want de", inner.lastLang)
	}
}

func TestTWithoutMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	text, err := translaasecho.T(c, "ui", "welcome")
	if text != "" {
		t.Fatalf("text = %q, want empty", text)
	}
	if err == nil {
		t.Fatal("expected error when middleware is missing")
	}
	if !strings.Contains(err.Error(), "service not found") {
		t.Fatalf("err = %v, want service not found", err)
	}
}
