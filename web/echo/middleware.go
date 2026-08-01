// Package translaasecho integrates Translaas with the Echo web framework.
//
// Install: go get github.com/Mantelabs/translaas-sdk-go/web/echo
//
// Translation strings are not HTML-escaped by the SDK. Use html/template when rendering HTML.
package translaasecho

import (
	"errors"
	"net/http"

	"github.com/Mantelabs/translaas-sdk-go/service"
	"github.com/Mantelabs/translaas-sdk-go/web"
	"github.com/labstack/echo/v4"
)

// Middleware injects a request-scoped Translaas service into Echo requests.
func Middleware(opts web.MiddlewareOptions) (echo.MiddlewareFunc, error) {
	if opts.BaseService == nil {
		return nil, errors.New("translaasecho: BaseService is required")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			effective := opts
			if effective.RouteLanguage == nil && effective.RequestLanguage.RouteParam != "" {
				routeParam := effective.RequestLanguage.RouteParam
				effective.RouteLanguage = func(*http.Request) string {
					return c.Param(routeParam)
				}
			}

			stdMw, err := web.Middleware(effective)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "translaas middleware misconfigured")
			}

			var handlerErr error
			stdMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.SetRequest(r)
				handlerErr = next(c)
			})).ServeHTTP(c.Response(), c.Request())

			return handlerErr
		}
	}, nil
}

// TemplateFunc returns a template callable bound to the Echo request context.
func TemplateFunc(c echo.Context) func(group, entry string, opts ...service.TOption) (string, error) {
	return func(group, entry string, opts ...service.TOption) (string, error) {
		svc, ok := web.ServiceFromContext(c.Request().Context())
		if !ok {
			return "", errors.New("translaasecho: service not found in context; register Middleware")
		}
		return svc.T(c.Request().Context(), group, entry, opts...)
	}
}
