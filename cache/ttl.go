package cache

import "time"

// TTL configures absolute and/or sliding expiration on Set.
// Zero durations disable that expiration type.
type TTL struct {
	Absolute time.Duration
	Sliding  time.Duration
}
