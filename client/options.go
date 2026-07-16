package client

import (
	"net/http"
	"strings"
	"time"

	"github.com/acuencadev/translaas-sdk-go/internal/validate"
)

// Options configures the Translaas HTTP client.
type Options struct {
	APIKey           string
	BaseURL          string
	Timeout          time.Duration
	DefaultProjectID string
}

type clientConfig struct {
	httpClient *http.Client
}

// Option configures client construction.
type Option func(*clientConfig)

// WithHTTPClient supplies a custom HTTP client (primarily for tests).
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *clientConfig) {
		cfg.httpClient = httpClient
	}
}

type client struct {
	apiKey           string
	baseURL          string
	timeout          time.Duration
	defaultProjectID string
	httpClient       *http.Client
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
	}, nil
}
