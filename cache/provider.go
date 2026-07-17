package cache

import "context"

// Provider is the in-memory cache boundary consumed by the HTTP client.
//
// Get copies a hit into dest when dest is a non-nil pointer to the stored type.
// It returns (true, nil) on hit, (false, nil) on miss, and an error only on
// invalid dest or internal failure.
type Provider interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl TTL) error
	Remove(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}
