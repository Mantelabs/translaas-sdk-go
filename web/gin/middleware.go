// Package translaasgin integrates Translaas with the Gin web framework.
//
// Install: go get github.com/acuencadev/translaas-sdk-go/web/gin
//
// Translation strings are not HTML-escaped by the SDK. Use html/template when rendering HTML.
package translaasgin

import (
	"errors"
	"net/http"

	"github.com/acuencadev/translaas-sdk-go/service"
	"github.com/acuencadev/translaas-sdk-go/web"
	"github.com/gin-gonic/gin"
)

// Middleware injects a request-scoped Translaas service into Gin requests.
func Middleware(opts web.MiddlewareOptions) (gin.HandlerFunc, error) {
	if opts.BaseService == nil {
		return nil, errors.New("translaasgin: BaseService is required")
	}

	return func(c *gin.Context) {
		effective := opts
		if effective.RouteLanguage == nil && effective.RequestLanguage.RouteParam != "" {
			routeParam := effective.RequestLanguage.RouteParam
			effective.RouteLanguage = func(*http.Request) string {
				return c.Param(routeParam)
			}
		}

		stdMw, err := web.Middleware(effective)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		stdMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}, nil
}

// T resolves a translation for the current Gin request.
func T(c *gin.Context, group, entry string, opts ...service.TOption) (string, error) {
	svc, ok := web.ServiceFromContext(c.Request.Context())
	if !ok {
		return "", errors.New("translaasgin: service not found in context; register Middleware")
	}
	return svc.T(c.Request.Context(), group, entry, opts...)
}
