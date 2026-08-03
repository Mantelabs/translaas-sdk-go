package web

import (
	"errors"
	"net/http"

	"github.com/Mantelabs/translaas-sdk-go/service"
	"github.com/Mantelabs/translaas-sdk-go/service/language"
)

// MiddlewareOptions configures request-scoped service injection.
type MiddlewareOptions struct {
	BaseService *service.Service

	LanguageSources []LanguageSource
	RequestLanguage RequestLanguageOptions
	RouteLanguage   RouteLanguageFunc
}

// DefaultMiddlewareOptions returns options with common language source defaults.
func DefaultMiddlewareOptions(base *service.Service) MiddlewareOptions {
	return MiddlewareOptions{
		BaseService:     base,
		LanguageSources: defaultLanguageSources(),
		RequestLanguage: RequestLanguageOptions{},
	}
}

// Middleware injects a request-scoped service and language context into each request.
func Middleware(opts MiddlewareOptions) (func(http.Handler) http.Handler, error) {
	if opts.BaseService == nil {
		return nil, errors.New("web: BaseService is required")
	}

	sources := opts.LanguageSources
	if len(sources) == 0 {
		sources = defaultLanguageSources()
	}
	languageOpts := normalizeRequestLanguageOptions(opts.RequestLanguage)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqProvider := NewRequestLanguageProvider(r, languageOpts, sources, opts.RouteLanguage)
			reqSvc, err := opts.BaseService.WithPrependedProviders(reqProvider)
			if err != nil {
				http.Error(w, "translaas middleware misconfigured", http.StatusInternalServerError)
				return
			}

			ctx := r.Context()
			ctx = language.WithAcceptLanguage(ctx, r.Header.Get("Accept-Language"))
			if lang, err := reqProvider.Language(ctx); err == nil && lang != "" {
				ctx = language.WithLanguage(ctx, lang)
			}
			ctx = WithService(ctx, reqSvc)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}
