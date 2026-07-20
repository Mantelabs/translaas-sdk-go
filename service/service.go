package service

import (
	"context"
	"errors"
	"strings"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/models"
	"github.com/acuencadev/translaas-sdk-go/service/language"
)

// Service is the convenience translation API (.NET ITranslaasService).
type Service struct {
	client   client.Client
	resolver *language.Resolver
}

// Options configures Service construction.
type Options struct {
	Resolver *language.Resolver
}

// New constructs a Service wrapping any client.Client implementation.
func New(c client.Client, opts Options) (*Service, error) {
	if c == nil {
		return nil, errors.New("service: client is required")
	}
	return &Service{
		client:   c,
		resolver: opts.Resolver,
	}, nil
}

// T retrieves a single translation with optional automatic language resolution.
func (s *Service) T(ctx context.Context, group, entry string, opts ...TOption) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	cfg := applyTOptions(opts...)
	lang, err := s.resolveLanguage(ctx, cfg)
	if err != nil {
		return "", err
	}

	getOpts := make([]client.GetEntryOption, 0, 3)
	if cfg.number != nil {
		getOpts = append(getOpts, client.WithNumber(*cfg.number))
	}
	if cfg.parameters != nil {
		getOpts = append(getOpts, client.WithParameters(cfg.parameters))
	}
	if cfg.requestContext != nil {
		getOpts = append(getOpts, client.WithRequestContext(cfg.requestContext))
	}

	return s.client.GetEntry(ctx, group, entry, lang, getOpts...)
}

func (s *Service) resolveLanguage(ctx context.Context, cfg tConfig) (string, error) {
	if cfg.langSet && strings.TrimSpace(cfg.lang) != "" {
		return cfg.lang, nil
	}

	if s.resolver == nil {
		return "", models.ErrNoLanguage
	}

	return s.resolver.Resolve(ctx)
}
