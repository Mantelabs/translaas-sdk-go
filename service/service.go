package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/Mantelabs/translaas-sdk-go/models"
	"github.com/Mantelabs/translaas-sdk-go/service/language"
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

// WithPrependedProviders returns a new Service sharing the client with request-scoped
// providers tried before the existing resolver chain.
func (s *Service) WithPrependedProviders(providers ...language.Provider) (*Service, error) {
	if s == nil {
		return nil, errors.New("service: nil receiver")
	}

	var resolver *language.Resolver
	var err error
	if s.resolver != nil {
		resolver, err = s.resolver.PrependProviders(providers...)
	} else {
		resolver, err = language.NewResolver(providers...)
	}
	if err != nil {
		return nil, err
	}

	return &Service{
		client:   s.client,
		resolver: resolver,
	}, nil
}

// New constructs a Service wrapping any client.Client implementation.
// Resolver is optional: pass nothing for service.New(c), or Options{Resolver}
// for web/request-scoped language providers. Language resolution for T:
// non-empty WithLang, then Resolver, then the client's DefaultLanguage.
func New(c client.Client, opts ...Options) (*Service, error) {
	if c == nil {
		return nil, errors.New("service: client is required")
	}
	var merged Options
	for _, o := range opts {
		if o.Resolver != nil {
			merged.Resolver = o.Resolver
		}
	}
	return &Service{
		client:   c,
		resolver: merged.Resolver,
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

type defaultLanguageClient interface {
	DefaultLanguage() string
}

func defaultLanguageFromClient(c client.Client) string {
	if dl, ok := c.(defaultLanguageClient); ok {
		return strings.TrimSpace(dl.DefaultLanguage())
	}
	return ""
}

func (s *Service) resolveLanguage(ctx context.Context, cfg tConfig) (string, error) {
	if cfg.langSet && strings.TrimSpace(cfg.lang) != "" {
		return cfg.lang, nil
	}

	if s.resolver != nil {
		lang, err := s.resolver.Resolve(ctx)
		if err == nil && strings.TrimSpace(lang) != "" {
			return lang, nil
		}
		if err != nil && !errors.Is(err, models.ErrNoLanguage) {
			return "", err
		}
	}

	if lang := defaultLanguageFromClient(s.client); lang != "" {
		return lang, nil
	}

	return "", models.ErrNoLanguage
}
