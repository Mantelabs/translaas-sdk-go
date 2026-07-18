package cachefile

// FallbackMode selects cache vs API ordering for intercepted reads.
type FallbackMode int

const (
	// FallbackCacheFirst reads disk first, then API on miss.
	FallbackCacheFirst FallbackMode = iota
	// FallbackAPIFirst reads API first, then disk on network/API errors.
	FallbackAPIFirst
	// FallbackCacheOnly reads disk only.
	FallbackCacheOnly
)

// Options configures the offline CachingClient decorator.
type Options struct {
	FallbackMode     FallbackMode
	DefaultProjectID string
}

// DefaultOptions returns Options with .NET-aligned defaults.
func DefaultOptions() Options {
	return Options{
		FallbackMode: FallbackCacheFirst,
	}
}
