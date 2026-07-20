package language

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/acuencadev/translaas-sdk-go/models"
)

// Resolver chains language providers; the first non-empty language wins.
type Resolver struct {
	providers []Provider
}

// NewResolver builds a resolver from providers evaluated in registration order.
func NewResolver(providers ...Provider) (*Resolver, error) {
	if len(providers) == 0 {
		return nil, errors.New("language: at least one provider is required")
	}
	copied := make([]Provider, len(providers))
	copy(copied, providers)
	return &Resolver{providers: copied}, nil
}

// Resolve returns the first non-empty language from the provider chain.
func (r *Resolver) Resolve(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	for _, provider := range r.providers {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		lang, err := provider.Language(ctx)
		if err != nil {
			slog.WarnContext(ctx, "language provider failed", "error", err)
			continue
		}
		if strings.TrimSpace(lang) != "" {
			return lang, nil
		}
	}

	return "", fmt.Errorf("%w", models.ErrNoLanguage)
}
