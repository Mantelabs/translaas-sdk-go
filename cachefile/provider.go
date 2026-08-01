package cachefile

import (
	"context"
	"time"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

// Provider is the offline disk cache contract (L2). Implementations must be safe for concurrent use.
type Provider interface {
	GetProject(ctx context.Context, project, lang string) (*models.TranslationProject, error)
	SaveProject(ctx context.Context, project, lang string, data *models.TranslationProject, opts ...SaveOption) error

	GetGroup(ctx context.Context, project, group, lang string) (*models.TranslationGroup, error)
	GetLocales(ctx context.Context, project string) (*models.ProjectLocales, error)
	SaveLocales(ctx context.Context, project string, data *models.ProjectLocales, opts ...SaveOption) error

	GetManifest(ctx context.Context) (*CacheManifest, error)
	UpdateManifest(ctx context.Context, update func(*CacheManifest) error) error

	IsCached(ctx context.Context, project, lang string) (bool, error)
	Clear(ctx context.Context) error
}

type saveConfig struct {
	expiresAt *time.Time
	cachedAt  time.Time
}

// SaveOption configures SaveProject and SaveLocales.
type SaveOption func(*saveConfig)

// WithExpiresAt sets wrapper ExpiresAt (nil = no expiry).
func WithExpiresAt(t *time.Time) SaveOption {
	return func(cfg *saveConfig) {
		cfg.expiresAt = t
	}
}

func applySaveOptions(opts []SaveOption) saveConfig {
	cfg := saveConfig{cachedAt: time.Now().UTC()}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
