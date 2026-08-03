package client

import (
	"net/http"
	"strings"
	"time"

	"github.com/Mantelabs/translaas-sdk-go/cache"
	"github.com/Mantelabs/translaas-sdk-go/internal/validate"
)

// Options configures the Translaas HTTP client.
type Options struct {
	APIKey           string
	BaseURL          string
	Timeout          time.Duration
	DefaultProjectID string
	CacheMode        cache.Mode
	CacheTTL         cache.TTL
}

type clientConfig struct {
	httpClient    *http.Client
	cacheProvider cache.Provider
	cacheMode     *cache.Mode
	cacheTTL      *cache.TTL
}

// Option configures client construction.
type Option func(*clientConfig)

// WithHTTPClient supplies a custom HTTP client (primarily for tests).
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *clientConfig) {
		cfg.httpClient = httpClient
	}
}

// WithCacheMode overrides Options.CacheMode.
func WithCacheMode(mode cache.Mode) Option {
	return func(cfg *clientConfig) {
		m := mode
		cfg.cacheMode = &m
	}
}

// WithCacheProvider supplies a custom in-memory cache provider (primarily for tests).
func WithCacheProvider(provider cache.Provider) Option {
	return func(cfg *clientConfig) {
		cfg.cacheProvider = provider
	}
}

// WithCacheTTL overrides Options.CacheTTL.
func WithCacheTTL(ttl cache.TTL) Option {
	return func(cfg *clientConfig) {
		t := ttl
		cfg.cacheTTL = &t
	}
}

type client struct {
	apiKey           string
	baseURL          string
	timeout          time.Duration
	defaultProjectID string
	httpClient       *http.Client
	cacheMode        cache.Mode
	cacheTTL         cache.TTL
	cacheProvider    cache.Provider
}

// New constructs a Client with validated options.
func New(opts Options, optFns ...Option) (Client, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if err := validate.Client(validate.ClientOptions{
		APIKey:  opts.APIKey,
		BaseURL: opts.BaseURL,
		Timeout: opts.Timeout,
	}); err != nil {
		return nil, err
	}

	cfg := clientConfig{}
	for _, optFn := range optFns {
		optFn(&cfg)
	}

	cacheMode := opts.CacheMode
	if cfg.cacheMode != nil {
		cacheMode = *cfg.cacheMode
	}
	cacheTTL := opts.CacheTTL
	if cfg.cacheTTL != nil {
		cacheTTL = *cfg.cacheTTL
	}

	var cacheProvider cache.Provider
	if cacheMode != cache.ModeNone {
		if cfg.cacheProvider != nil {
			cacheProvider = cfg.cacheProvider
		} else {
			cacheProvider = cache.NewMemoryProvider()
		}
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &client{
		apiKey:           strings.TrimSpace(opts.APIKey),
		baseURL:          strings.TrimSpace(opts.BaseURL),
		timeout:          timeout,
		defaultProjectID: strings.TrimSpace(opts.DefaultProjectID),
		httpClient:       httpClient,
		cacheMode:        cacheMode,
		cacheTTL:         cacheTTL,
		cacheProvider:    cacheProvider,
	}, nil
}
