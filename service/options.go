package service

import "github.com/acuencadev/translaas-sdk-go/models"

type tConfig struct {
	lang           string
	langSet        bool
	number         *float64
	parameters     map[string]string
	requestContext *models.RequestContext
}

// TOption configures a single T call.
type TOption func(*tConfig)

// WithLang sets an explicit language. Non-empty values bypass the resolver; empty or
// whitespace-only values trigger automatic resolution.
func WithLang(lang string) TOption {
	return func(cfg *tConfig) {
		cfg.lang = lang
		cfg.langSet = true
	}
}

// WithNumber sets the plural/interpolation count forwarded to GetEntry.
func WithNumber(n float64) TOption {
	return func(cfg *tConfig) {
		cfg.number = &n
	}
}

// WithParameters adds interpolation query parameters forwarded to GetEntry.
func WithParameters(params map[string]string) TOption {
	return func(cfg *tConfig) {
		cfg.parameters = params
	}
}

// WithRequestContext supplies per-request channel, version, project, and conditional headers.
func WithRequestContext(rc *models.RequestContext) TOption {
	return func(cfg *tConfig) {
		cfg.requestContext = rc
	}
}

func applyTOptions(opts ...TOption) tConfig {
	cfg := tConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
