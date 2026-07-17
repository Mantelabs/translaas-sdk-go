package cache

// MemoryOption configures NewMemoryProvider.
type MemoryOption func(*memoryConfig)

type memoryConfig struct {
	maxSize          int
	enableStatistics bool
}

// WithMaxSize enables LRU eviction when the store reaches maxSize entries.
func WithMaxSize(maxSize int) MemoryOption {
	return func(cfg *memoryConfig) {
		if maxSize > 0 {
			cfg.maxSize = maxSize
		}
	}
}

// WithStatistics enables hit/miss counters on the provider.
func WithStatistics() MemoryOption {
	return func(cfg *memoryConfig) {
		cfg.enableStatistics = true
	}
}

// Statistics holds optional cache counters when WithStatistics is enabled.
type Statistics struct {
	Hits   int64
	Misses int64
	Size   int
}

// GetStatistics returns a snapshot of cache counters.
func (p *memoryProvider) GetStatistics() Statistics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return Statistics{
		Hits:   p.hits,
		Misses: p.misses,
		Size:   len(p.entries),
	}
}
