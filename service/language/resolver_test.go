package language_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/models"
	"github.com/Mantelabs/translaas-sdk-go/service/language"
)

type stubProvider struct {
	lang string
	err  error
}

func (p stubProvider) Language(context.Context) (string, error) {
	return p.lang, p.err
}

func TestNewResolverRequiresProviders(t *testing.T) {
	t.Parallel()

	if _, err := language.NewResolver(); err == nil {
		t.Fatal("expected error for empty provider list")
	}
}

func TestResolverOrder(t *testing.T) {
	t.Parallel()

	ctx := language.WithLanguage(context.Background(), "fr")
	ctx = language.WithAcceptLanguage(ctx, "en-US,en;q=0.9")

	resolver, err := language.NewResolver(
		language.NewContextLanguageProvider(),
		language.NewAcceptLanguageProvider(),
		language.NewDefaultLanguageProvider("es"),
	)
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	got, err := resolver.Resolve(ctx)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "fr" {
		t.Fatalf("got %q, want fr", got)
	}
}

func TestResolverProviderErrorContinues(t *testing.T) {
	t.Parallel()

	resolver, err := language.NewResolver(
		stubProvider{err: errors.New("boom")},
		language.NewDefaultLanguageProvider("en"),
	)
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	got, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "en" {
		t.Fatalf("got %q, want en", got)
	}
}

func TestResolverNoLanguage(t *testing.T) {
	t.Parallel()

	resolver, err := language.NewResolver(
		stubProvider{lang: ""},
		stubProvider{lang: "   "},
	)
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	_, err = resolver.Resolve(context.Background())
	if !errors.Is(err, models.ErrNoLanguage) {
		t.Fatalf("expected ErrNoLanguage, got %v", err)
	}
}

func TestContextLanguageProvider(t *testing.T) {
	t.Parallel()

	provider := language.NewContextLanguageProvider()
	ctx := language.WithLanguage(context.Background(), "pt")

	got, err := provider.Language(ctx)
	if err != nil {
		t.Fatalf("Language: %v", err)
	}
	if got != "pt" {
		t.Fatalf("got %q, want pt", got)
	}
}

func TestAcceptLanguageProvider(t *testing.T) {
	t.Parallel()

	provider := language.NewAcceptLanguageProvider()
	ctx := language.WithAcceptLanguage(context.Background(), "en-US,en;q=0.9")

	got, err := provider.Language(ctx)
	if err != nil {
		t.Fatalf("Language: %v", err)
	}
	if got != "en" {
		t.Fatalf("got %q, want en", got)
	}
}

func TestDefaultLanguageProvider(t *testing.T) {
	t.Parallel()

	provider := language.NewDefaultLanguageProvider("de")

	got, err := provider.Language(context.Background())
	if err != nil {
		t.Fatalf("Language: %v", err)
	}
	if got != "de" {
		t.Fatalf("got %q, want de", got)
	}
}
