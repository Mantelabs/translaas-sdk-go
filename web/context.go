package web

import (
	"context"

	"github.com/acuencadev/translaas-sdk-go/service"
)

type contextKey struct{}

// WithService stores a request-scoped service on ctx.
func WithService(ctx context.Context, svc *service.Service) context.Context {
	return context.WithValue(ctx, contextKey{}, svc)
}

// ServiceFromContext returns the service injected by Middleware.
func ServiceFromContext(ctx context.Context) (*service.Service, bool) {
	svc, ok := ctx.Value(contextKey{}).(*service.Service)
	if !ok || svc == nil {
		return nil, false
	}
	return svc, true
}
