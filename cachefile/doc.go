// Package cachefile provides offline file-backed caching for the Translaas SDK.
//
// FileProvider implements the L2 disk cache with JSON wrappers, a root manifest,
// atomic writes, and expiration-as-miss semantics aligned with the .NET SDK.
//
// HybridProvider adds an expirable LRU memory layer (L1) over any Provider (L2),
// promoting disk hits into memory and writing through to both tiers on save.
//
// CachingClient decorates client.Client with offline fallback modes (CacheFirst,
// APIFirst, CacheOnly), offline entry resolution, and cache warming after API reads.
//
// SyncService pulls translations from the API into a Provider using the inner client
// (not CachingClient) and supports optional background sync on a ticker.
//
// ParseOfflineZip and FileProvider.ImportOfflineBundle import offline ZIP bundles
// (HTTP spec §7.6) into the same on-disk layout. SyncFromOfflineZip combines download
// and import in one call.
package cachefile
