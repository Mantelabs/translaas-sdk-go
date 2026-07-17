package cache

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type cacheEntry struct {
	value           any
	absoluteExpiry  *time.Time
	slidingDuration time.Duration
	lastAccess      time.Time
}

type memoryProvider struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	maxSize int
	hits    int64
	misses  int64
	stats   bool
	now     func() time.Time
}

// NewMemoryProvider returns a thread-safe in-memory Provider.
func NewMemoryProvider(opts ...MemoryOption) Provider {
	cfg := memoryConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &memoryProvider{
		entries: make(map[string]*cacheEntry),
		maxSize: cfg.maxSize,
		stats:   cfg.enableStatistics,
		now:     time.Now,
	}
}

func (p *memoryProvider) Get(ctx context.Context, key string, dest any) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if dest == nil {
		return false, fmt.Errorf("cache: dest must be non-nil pointer")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.entries[key]
	if !ok {
		p.recordMissLocked()
		return false, nil
	}

	now := p.now()
	if entry.expired(now) {
		delete(p.entries, key)
		p.recordMissLocked()
		return false, nil
	}

	entry.lastAccess = now
	if err := assignValue(dest, entry.value); err != nil {
		return false, err
	}
	p.recordHitLocked()
	return true, nil
}

func (p *memoryProvider) Set(ctx context.Context, key string, value any, ttl TTL) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	now := p.now()
	entry := &cacheEntry{
		value:           value,
		slidingDuration: ttl.Sliding,
		lastAccess:      now,
	}
	if ttl.Absolute > 0 {
		expiry := now.Add(ttl.Absolute)
		entry.absoluteExpiry = &expiry
	}

	if existing, ok := p.entries[key]; ok {
		p.entries[key] = entry
		_ = existing
		return nil
	}

	if p.maxSize > 0 && len(p.entries) >= p.maxSize {
		p.evictLRULocked(now)
	}
	p.entries[key] = entry
	return nil
}

func (p *memoryProvider) Remove(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.entries, key)
	return nil
}

func (p *memoryProvider) Clear(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries = make(map[string]*cacheEntry)
	if p.stats {
		p.hits = 0
		p.misses = 0
	}
	return nil
}

func (e *cacheEntry) expired(now time.Time) bool {
	if e.absoluteExpiry != nil && !now.Before(*e.absoluteExpiry) {
		return true
	}
	if e.slidingDuration > 0 && now.Sub(e.lastAccess) >= e.slidingDuration {
		return true
	}
	return false
}

func (p *memoryProvider) evictLRULocked(now time.Time) {
	if len(p.entries) == 0 {
		return
	}
	var oldestKey string
	var oldestAccess time.Time
	first := true
	for key, entry := range p.entries {
		if first || entry.lastAccess.Before(oldestAccess) {
			oldestKey = key
			oldestAccess = entry.lastAccess
			first = false
		}
	}
	delete(p.entries, oldestKey)
	_ = now
}

func (p *memoryProvider) recordHitLocked() {
	if p.stats {
		p.hits++
	}
}

func (p *memoryProvider) recordMissLocked() {
	if p.stats {
		p.misses++
	}
}

func assignValue(dest, src any) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Pointer || destVal.IsNil() {
		return fmt.Errorf("cache: dest must be non-nil pointer")
	}
	destElem := destVal.Elem()
	srcVal := reflect.ValueOf(src)

	if srcVal.Type().AssignableTo(destElem.Type()) {
		destElem.Set(srcVal)
		return nil
	}
	if srcVal.Kind() == reflect.Pointer && !srcVal.IsNil() && srcVal.Elem().Type().AssignableTo(destElem.Type()) {
		destElem.Set(srcVal.Elem())
		return nil
	}
	if destElem.Kind() == reflect.Pointer && srcVal.Type().AssignableTo(destElem.Type().Elem()) {
		ptr := reflect.New(destElem.Type().Elem())
		ptr.Elem().Set(srcVal)
		destElem.Set(ptr)
		return nil
	}
	if destElem.Kind() == reflect.Pointer && srcVal.Kind() == reflect.Pointer && srcVal.Type().AssignableTo(destElem.Type()) {
		destElem.Set(srcVal)
		return nil
	}
	return fmt.Errorf("cache: cannot assign %T to %T", src, dest)
}
