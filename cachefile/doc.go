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
package cachefile
