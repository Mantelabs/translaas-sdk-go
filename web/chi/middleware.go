// Package translaaschi integrates Translaas with the chi router.
//
// Install: go get github.com/acuencadev/translaas-sdk-go/web/chi
//
// Translation strings are not HTML-escaped by the SDK. Use html/template when rendering HTML.
package translaaschi

import (
	"errors"
	"net/http"

	"github.com/acuencadev/translaas-sdk-go/web"
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
