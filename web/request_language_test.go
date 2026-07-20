package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/web"
)

func TestRequestLanguageProviderSourceOrder(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/?lang=de", nil)
	req.Header.Set("Accept-Language", "fr-FR")
	req.AddCookie(&http.Cookie{Name: "language", Value: "es"})

	provider := web.NewRequestLanguageProvider(
		req,
		web.RequestLanguageOptions{},
		[]web.LanguageSource{web.SourceQuery, web.SourceAcceptLanguage, web.SourceCookie},
		nil,
	)

	lang, err := provider.Language(req.Context())
	if err != nil {
		t.Fatalf("Language: %v", err)
	}
	if lang != "de" {
		t.Fatalf("lang = %q, want de", lang)
	}
}

func TestRequestLanguageProviderAcceptLanguageParsing(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	provider := web.NewRequestLanguageProvider(
		req,
		web.RequestLanguageOptions{},
		[]web.LanguageSource{web.SourceAcceptLanguage},
		nil,
	)

	lang, err := provider.Language(req.Context())
	if err != nil {
		t.Fatalf("Language: %v", err)
	}
	if lang != "en" {
		t.Fatalf("lang = %q, want en", lang)
	}
}
