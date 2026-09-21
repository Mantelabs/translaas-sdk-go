// Package translaaschi integrates Translaas with the chi router.
//
// Install: go get github.com/Mantelabs/translaas-sdk-go/web/chi
//
// Translation strings are not HTML-escaped by the SDK. Use html/template when rendering HTML.
package translaaschi

import (
	"errors"
	"net/http"

	"github.com/Mantelabs/translaas-sdk-go/service"
	"github.com/Mantelabs/translaas-sdk-go/web"
	"github.com/go-chi/chi/v5"
)

// Middleware injects a request-scoped Translaas service into chi requests.
func Middleware(opts web.MiddlewareOptions) (func(http.Handler) http.Handler, error) {
	if opts.BaseService == nil {
		return nil, errors.New("translaaschi: BaseService is required")
	}

	effective := opts
	if effective.RouteLanguage == nil && effective.RequestLanguage.RouteParam != "" {
		routeParam := effective.RequestLanguage.RouteParam
		effective.RouteLanguage = func(r *http.Request) string {
			return chi.URLParam(r, routeParam)
		}
	}

	return web.Middleware(effective)
}

// T resolves a translation for the current chi request.
// Translation strings are not HTML-escaped; use html/template when rendering HTML.
func T(r *http.Request, group, entry string, opts ...service.TOption) (string, error) {
	svc, ok := web.ServiceFromContext(r.Context())
	if !ok {
		return "", errors.New("translaaschi: service not found in context; register Middleware")
	}
	return svc.T(r.Context(), group, entry, opts...)
}
