package cachefile

import (
	"time"

	"github.com/acuencadev/translaas-sdk-go/models"
)

const (
	// ManifestVersion is the root manifest schema version written by the SDK.
	ManifestVersion = "1.0"
	// DefaultSDKVersion is the offline cache format version recorded in manifest.json.
	DefaultSDKVersion = "1.0.0"
)

// CachedProject wraps a translation project payload with cache metadata.
type CachedProject struct {
	CachedAt  time.Time                 `json:"cachedAt"`
	ExpiresAt *time.Time                `json:"expiresAt,omitempty"`
	Data      models.TranslationProject `json:"data"`
}

// CachedLocales wraps supported locales with cache metadata.
type CachedLocales struct {
	CachedAt  time.Time             `json:"cachedAt"`
	ExpiresAt *time.Time            `json:"expiresAt,omitempty"`
	Data      models.ProjectLocales `json:"data"`
}

// CacheManifest is the root offline cache index (manifest.json).
type CacheManifest struct {
	Version    string                      `json:"version"`
	SDKVersion string                      `json:"sdkVersion"`
	CreatedAt  time.Time                   `json:"createdAt"`
	LastSyncAt time.Time                   `json:"lastSyncAt"`
	Projects   map[string]ProjectCacheInfo `json:"projects"`
}

// ProjectCacheInfo tracks cached languages for one project.
type ProjectCacheInfo struct {
	Languages  []string  `json:"languages"`
	LastSyncAt time.Time `json:"lastSyncAt"`
	Status     string    `json:"status"`
}
