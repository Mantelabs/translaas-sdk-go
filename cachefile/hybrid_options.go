package cachefile

import "time"

const (
	defaultHybridMemoryExpiration = 30 * time.Minute
	defaultHybridMaxEntries       = 1000
)

// HybridOptions configures the in-memory L1 layer over a file-backed L2 Provider.
type HybridOptions struct {
	// Enabled turns the memory layer on. When false, HybridProvider delegates to L2 only.
	Enabled bool
	// MemoryExpiration is the TTL for L1 entries. Zero uses DefaultHybridMemoryExpiration.
	MemoryExpiration time.Duration
	// MaxEntries is the LRU capacity per L1 partition (projects, groups, locales). Zero uses defaultHybridMaxEntries.
	MaxEntries int
}

// DefaultHybridOptions returns HybridOptions with .NET-aligned defaults.
func DefaultHybridOptions() HybridOptions {
	return HybridOptions{
		Enabled:          true,
		MemoryExpiration: defaultHybridMemoryExpiration,
		MaxEntries:       defaultHybridMaxEntries,
	}
}

func normalizeHybridOptions(opts HybridOptions) HybridOptions {
	if opts.MemoryExpiration <= 0 {
		opts.MemoryExpiration = defaultHybridMemoryExpiration
	}
	if opts.MaxEntries <= 0 {
		opts.MaxEntries = defaultHybridMaxEntries
	}
	return opts
}
