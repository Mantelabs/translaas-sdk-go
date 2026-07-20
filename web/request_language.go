package web

import (
	"context"
	"net/http"
	"strings"

	"github.com/acuencadev/translaas-sdk-go/service/language"
)

// LanguageSource identifies where RequestLanguageProvider reads a language code.
type LanguageSource int

const (
	SourceQuery LanguageSource = iota
	SourceHeader
	SourceCookie
	SourceRoute
	SourceAcceptLanguage
)

// RequestLanguageOptions configures request language extraction.
type RequestLanguageOptions struct {
	QueryParam string
	HeaderName string
	CookieName string
	RouteParam string
}

// RouteLanguageFunc supplies a route/path parameter language when SourceRoute is enabled.
type RouteLanguageFunc func(r *http.Request) string

// RequestLanguageProvider resolves language from an HTTP request (.NET RequestLanguageProvider).
type RequestLanguageProvider struct {
	request   *http.Request
	opts      RequestLanguageOptions
	sources   []LanguageSource
	routeFunc RouteLanguageFunc
}

// NewRequestLanguageProvider constructs a provider for the given request and source order.
func NewRequestLanguageProvider(
	r *http.Request,
	opts RequestLanguageOptions,
	sources []LanguageSource,
	routeFunc RouteLanguageFunc,
) *RequestLanguageProvider {
	normalized := normalizeRequestLanguageOptions(opts)
	copied := append([]LanguageSource(nil), sources...)
	return &RequestLanguageProvider{
		request:   r,
		opts:      normalized,
		sources:   copied,
		routeFunc: routeFunc,
	}
}

// Language returns the first non-empty language from the configured sources.
func (p *RequestLanguageProvider) Language(context.Context) (string, error) {
	if p == nil || p.request == nil {
		return "", nil
	}

	for _, source := range p.sources {
		if lang := p.languageFromSource(source); lang != "" {
			return lang, nil
		}
	}

	return "", nil
}

func (p *RequestLanguageProvider) languageFromSource(source LanguageSource) string {
	switch source {
	case SourceQuery:
		return normalizeLanguage(p.request.URL.Query().Get(p.opts.QueryParam))
	case SourceHeader:
		return normalizeLanguage(p.request.Header.Get(p.opts.HeaderName))
	case SourceCookie:
		if cookie, err := p.request.Cookie(p.opts.CookieName); err == nil {
			return normalizeLanguage(cookie.Value)
		}
		return ""
	case SourceRoute:
		if p.routeFunc != nil {
			return normalizeLanguage(p.routeFunc(p.request))
		}
		if p.opts.RouteParam != "" {
			return normalizeLanguage(p.request.PathValue(p.opts.RouteParam))
		}
		return ""
	case SourceAcceptLanguage:
		return language.ParseAcceptLanguage(p.request.Header.Get("Accept-Language"))
	default:
		return ""
	}
}

func normalizeRequestLanguageOptions(opts RequestLanguageOptions) RequestLanguageOptions {
	if opts.QueryParam == "" {
		opts.QueryParam = "lang"
	}
	if opts.HeaderName == "" {
		opts.HeaderName = "Accept-Language"
	}
	if opts.CookieName == "" {
		opts.CookieName = "language"
	}
	return opts
}

func normalizeLanguage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if normalized := language.NormalizeLanguageCode(value); normalized != "" {
		return normalized
	}
	return strings.ToLower(value)
}

func defaultLanguageSources() []LanguageSource {
	return []LanguageSource{
		SourceQuery,
		SourceAcceptLanguage,
		SourceCookie,
	}
}
