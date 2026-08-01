package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/Mantelabs/translaas-sdk-go/models"
	"github.com/Mantelabs/translaas-sdk-go/service"
	"github.com/Mantelabs/translaas-sdk-go/service/language"
)

type mockClient struct {
	mu sync.Mutex

	getEntryCalls int
	lastLang      string
	lastOpts      client.GetEntryCallOptions

	getEntryFn func(ctx context.Context, group, entry, lang string, opts ...client.GetEntryOption) (string, error)
}

func (m *mockClient) GetEntry(ctx context.Context, group, entry, lang string, opts ...client.GetEntryOption) (string, error) {
	m.mu.Lock()
	m.getEntryCalls++
	m.lastLang = lang
	m.lastOpts = client.ParseGetEntryOptions(opts...)
	fn := m.getEntryFn
	m.mu.Unlock()

	if fn != nil {
		return fn(ctx, group, entry, lang, opts...)
	}
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

func TestNewRequiresClient(t *testing.T) {
	t.Parallel()

	if _, err := service.New(nil, service.Options{}); err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestTExplicitLangBypassesResolver(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	resolver, err := language.NewResolver(language.NewDefaultLanguageProvider("en"))
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	svc, err := service.New(inner, service.Options{Resolver: resolver})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := svc.T(context.Background(), "common", "welcome", service.WithLang("de"))
	if err != nil {
		t.Fatalf("T: %v", err)
	}
	if got != "hello" {
		t.Fatalf("got %q, want hello", got)
	}
	if inner.lastLang != "de" {
		t.Fatalf("lastLang = %q, want de", inner.lastLang)
	}
}

func TestTEmptyLangUsesResolver(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	resolver, err := language.NewResolver(language.NewDefaultLanguageProvider("es"))
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	svc, err := service.New(inner, service.Options{Resolver: resolver})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = svc.T(context.Background(), "common", "welcome", service.WithLang("   "))
	if err != nil {
		t.Fatalf("T: %v", err)
	}
	if inner.lastLang != "es" {
		t.Fatalf("lastLang = %q, want es", inner.lastLang)
	}
}

func TestTNoResolverNoLang(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	svc, err := service.New(inner, service.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = svc.T(context.Background(), "common", "welcome")
	if !errors.Is(err, models.ErrNoLanguage) {
		t.Fatalf("expected ErrNoLanguage, got %v", err)
	}
}

func TestTForwardsGetEntryOptions(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	svc, err := service.New(inner, service.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	number := 3.0
	params := map[string]string{"name": "Ada"}
	reqCtx := &models.RequestContext{Project: "demo"}

	_, err = svc.T(
		context.Background(),
		"common",
		"items",
		service.WithLang("en"),
		service.WithNumber(number),
		service.WithParameters(params),
		service.WithRequestContext(reqCtx),
	)
	if err != nil {
		t.Fatalf("T: %v", err)
	}

	if inner.lastOpts.Number == nil || *inner.lastOpts.Number != number {
		t.Fatalf("number = %v, want %v", inner.lastOpts.Number, number)
	}
	if inner.lastOpts.Parameters["name"] != "Ada" {
		t.Fatalf("parameters = %v", inner.lastOpts.Parameters)
	}
	if inner.lastOpts.RequestContext != reqCtx {
		t.Fatalf("request context not forwarded")
	}
}

func TestTResolverFromContext(t *testing.T) {
	t.Parallel()

	inner := &mockClient{}
	resolver, err := language.NewResolver(
		language.NewContextLanguageProvider(),
		language.NewDefaultLanguageProvider("en"),
	)
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	svc, err := service.New(inner, service.Options{Resolver: resolver})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := language.WithLanguage(context.Background(), "pt")
	_, err = svc.T(ctx, "common", "welcome")
	if err != nil {
		t.Fatalf("T: %v", err)
	}
	if inner.lastLang != "pt" {
		t.Fatalf("lastLang = %q, want pt", inner.lastLang)
	}
}

func TestWithPrependedProviders(t *testing.T) {
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

	reqSvc, err := base.WithPrependedProviders(language.NewDefaultLanguageProvider("de"))
	if err != nil {
		t.Fatalf("WithPrependedProviders: %v", err)
	}

	_, err = reqSvc.T(context.Background(), "common", "welcome")
	if err != nil {
		t.Fatalf("T: %v", err)
	}
	if inner.lastLang != "de" {
		t.Fatalf("lastLang = %q, want de", inner.lastLang)
	}
}

func TestTRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	inner := &mockClient{
		getEntryFn: func(context.Context, string, string, string, ...client.GetEntryOption) (string, error) {
			return "", errors.New("should not be called")
		},
	}
	resolver, err := language.NewResolver(language.NewDefaultLanguageProvider("en"))
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	svc, err := service.New(inner, service.Options{Resolver: resolver})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = svc.T(ctx, "common", "welcome")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
