// Package client implements the Translaas SDK HTTP client matching .NET TranslaasClient.
package client

import (
	"context"
	"time"

	"github.com/acuencadev/translaas-sdk-go/models"
)

const (
	// DefaultTimeout is applied when Options.Timeout is zero.
	DefaultTimeout = 30 * time.Second

	sdkTranslationsPrefix = "sdk/v1/translations"
)

// Client is the consumer-facing HTTP client boundary.
type Client interface {
	GetEntry(ctx context.Context, group, entry, lang string, opts ...GetEntryOption) (string, error)

	GetGroup(ctx context.Context, project, group, lang string, opts ...GetGroupOption) (*models.TranslationGroup, error)
	GetProject(ctx context.Context, project, lang string, opts ...GetProjectOption) (*models.TranslationProject, error)
	GetProjectLocales(ctx context.Context, project string, opts ...GetProjectLocalesOption) (*models.ProjectLocales, error)
	GetOfflineCache(ctx context.Context, project string, opts ...GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error)

	ReportMissingKeys(ctx context.Context, keys []models.ReportMissingKeyItem) error
	ValidateAPIKey(ctx context.Context) (*models.ValidateAPIKeyResponse, error)
}
