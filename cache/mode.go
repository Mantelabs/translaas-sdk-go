package cache

import "strconv"

// Mode controls which client operations participate in in-memory caching.
//
// Wiring lives in the client package: ModeNone disables caching; ModeEntry caches
// GetEntry only; ModeGroup caches GetGroup (and group keys under ModeProject);
// ModeProject caches GetProject. GetProjectLocales is cached whenever Mode != ModeNone.
type Mode int

const (
	// ModeNone disables in-memory caching.
	ModeNone Mode = iota
	// ModeEntry caches individual GetEntry results.
	ModeEntry
	// ModeGroup caches GetGroup payloads.
	ModeGroup
	// ModeProject caches GetProject payloads.
	ModeProject
)

// String returns a stable diagnostic name for m.
func (m Mode) String() string {
	switch m {
	case ModeNone:
		return "None"
	case ModeEntry:
		return "Entry"
	case ModeGroup:
		return "Group"
	case ModeProject:
		return "Project"
	default:
		return "Mode(" + strconv.Itoa(int(m)) + ")"
	}
}
