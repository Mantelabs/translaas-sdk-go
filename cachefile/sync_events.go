package cachefile

import "time"

// SyncCompletedEvent reports a successful project/language sync.
type SyncCompletedEvent struct {
	Project  string
	Language string
	SyncedAt time.Time
}

// SyncFailedEvent reports a failed project/language sync.
type SyncFailedEvent struct {
	Project  string
	Language string
	Err      error
}

// SyncResult aggregates SyncAll outcomes across configured projects.
type SyncResult struct {
	SyncedProjects []string
	FailedProjects []string
	CompletedAt    time.Time
}
