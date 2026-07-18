package cachefile

import "time"

const defaultCacheDirectory = ".translaas-cache"

// OfflineCacheOptions configures file-backed offline caching and background sync.
type OfflineCacheOptions struct {
	// Enabled turns offline file caching on. When false, sync helpers no-op at the app layer.
	Enabled bool

	// CacheDirectory is the root path for on-disk cache files (absolute or relative to CWD).
	CacheDirectory string

	// FallbackMode selects cache vs API ordering for CachingClient reads.
	FallbackMode FallbackMode

	// AutoSync enables periodic background synchronization when StartBackgroundSync is used.
	AutoSync bool

	// AutoSyncInterval controls the delay between background sync runs.
	// Nil disables interval-based sync (StartBackgroundSync becomes a no-op).
	AutoSyncInterval *time.Duration

	// Projects lists project IDs to sync in SyncAll and background sync.
	Projects []string

	// Languages limits pre-cache to these locale codes. Empty means all project locales.
	Languages []string

	// DefaultProjectID is required for offline GetEntry lookups via CachingClient.
	DefaultProjectID string
}

// DefaultOfflineCacheOptions returns .NET-aligned defaults.
func DefaultOfflineCacheOptions() OfflineCacheOptions {
	interval := time.Hour
	return OfflineCacheOptions{
		CacheDirectory:   defaultCacheDirectory,
		FallbackMode:     FallbackCacheFirst,
		AutoSync:         true,
		AutoSyncInterval: &interval,
	}
}
