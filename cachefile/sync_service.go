package cachefile

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Mantelabs/translaas-sdk-go/client"
)

// SyncCallbacks holds optional hooks for sync lifecycle events.
type SyncCallbacks struct {
	OnSyncCompleted    func(SyncCompletedEvent)
	OnSyncFailed       func(SyncFailedEvent)
	OnSyncAllCompleted func(SyncResult)
}

// SyncService synchronizes offline cache files with the Translaas API.
// It uses the inner HTTP client directly, not a CachingClient decorator.
type SyncService struct {
	client    client.Client
	cache     Provider
	options   OfflineCacheOptions
	callbacks SyncCallbacks

	mu sync.Mutex

	bgMu     sync.Mutex
	bgCancel context.CancelFunc
	bgDone   chan struct{}
}

// NewSyncService constructs a SyncService.
func NewSyncService(
	inner client.Client,
	cache Provider,
	options OfflineCacheOptions,
	callbacks SyncCallbacks,
) (*SyncService, error) {
	if inner == nil {
		return nil, fmt.Errorf("cachefile: sync client must not be nil")
	}
	if cache == nil {
		return nil, fmt.Errorf("cachefile: sync cache provider must not be nil")
	}
	return &SyncService{
		client:    inner,
		cache:     cache,
		options:   options,
		callbacks: callbacks,
	}, nil
}

// IsBackgroundSyncRunning reports whether StartBackgroundSync started a loop that has not stopped.
func (s *SyncService) IsBackgroundSyncRunning() bool {
	s.bgMu.Lock()
	defer s.bgMu.Unlock()
	return s.bgDone != nil
}

// SyncProject fetches one project language from the API and persists it to disk.
func (s *SyncService) SyncProject(ctx context.Context, project, lang string) error {
	project = strings.TrimSpace(project)
	if project == "" {
		return fmt.Errorf("cachefile: project must not be empty")
	}
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return fmt.Errorf("cachefile: language must not be empty")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	projectData, err := s.client.GetProject(ctx, project, lang)
	if err != nil {
		s.emitSyncFailed(SyncFailedEvent{Project: project, Language: lang, Err: err})
		return err
	}

	if err := s.cache.SaveProject(ctx, project, lang, projectData); err != nil {
		s.emitSyncFailed(SyncFailedEvent{Project: project, Language: lang, Err: err})
		return err
	}

	s.emitSyncCompleted(SyncCompletedEvent{
		Project:  project,
		Language: lang,
		SyncedAt: time.Now().UTC(),
	})
	return nil
}

// SyncProjectAllLanguages fetches locales and syncs each configured language for a project.
func (s *SyncService) SyncProjectAllLanguages(ctx context.Context, project string) error {
	project = strings.TrimSpace(project)
	if project == "" {
		return fmt.Errorf("cachefile: project must not be empty")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	locales, err := s.client.GetProjectLocales(ctx, project)
	if err != nil {
		return err
	}

	if err := s.cache.SaveLocales(ctx, project, locales); err != nil {
		return err
	}

	languages := filterSyncLanguages(locales.Locales, s.options.Languages)
	for _, lang := range languages {
		if err := ctx.Err(); err != nil {
			return err
		}

		projectData, err := s.client.GetProject(ctx, project, lang)
		if err != nil {
			s.emitSyncFailed(SyncFailedEvent{Project: project, Language: lang, Err: err})
			continue
		}

		if err := s.cache.SaveProject(ctx, project, lang, projectData); err != nil {
			s.emitSyncFailed(SyncFailedEvent{Project: project, Language: lang, Err: err})
			continue
		}

		s.emitSyncCompleted(SyncCompletedEvent{
			Project:  project,
			Language: lang,
			SyncedAt: time.Now().UTC(),
		})
	}

	return nil
}

// SyncAll synchronizes every project listed in OfflineCacheOptions.Projects.
func (s *SyncService) SyncAll(ctx context.Context) (*SyncResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &SyncResult{
		SyncedProjects: make([]string, 0, len(s.options.Projects)),
		FailedProjects: make([]string, 0),
		CompletedAt:    time.Now().UTC(),
	}

	for _, project := range s.options.Projects {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		if err := s.SyncProjectAllLanguages(ctx, project); err != nil {
			result.FailedProjects = append(result.FailedProjects, project)
			continue
		}
		result.SyncedProjects = append(result.SyncedProjects, project)
	}

	s.emitSyncAllCompleted(*result)
	return result, nil
}

// StartBackgroundSync runs an initial SyncAll, then repeats on AutoSyncInterval until ctx is canceled.
func (s *SyncService) StartBackgroundSync(ctx context.Context) {
	if !s.options.AutoSync || s.options.AutoSyncInterval == nil {
		return
	}

	s.bgMu.Lock()
	if s.bgDone != nil {
		s.bgMu.Unlock()
		return
	}

	bgCtx, cancel := context.WithCancel(ctx)
	s.bgCancel = cancel
	s.bgDone = make(chan struct{})
	s.bgMu.Unlock()

	go s.runBackgroundSync(bgCtx, *s.options.AutoSyncInterval)
}

// StopBackgroundSync cancels the background loop and waits for it to exit.
func (s *SyncService) StopBackgroundSync() {
	s.bgMu.Lock()
	cancel := s.bgCancel
	done := s.bgDone
	s.bgCancel = nil
	s.bgDone = nil
	s.bgMu.Unlock()

	if cancel == nil {
		return
	}

	cancel()
	<-done
}

func (s *SyncService) runBackgroundSync(ctx context.Context, interval time.Duration) {
	defer close(s.bgDone)

	run := func() {
		_, err := s.SyncAll(ctx)
		if err != nil && ctx.Err() != nil {
			return
		}
	}

	run()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func filterSyncLanguages(available, requested []string) []string {
	if len(requested) == 0 {
		out := make([]string, len(available))
		copy(out, available)
		return out
	}

	availableSet := make(map[string]struct{}, len(available))
	for _, lang := range available {
		availableSet[lang] = struct{}{}
	}

	out := make([]string, 0, len(requested))
	for _, lang := range requested {
		if _, ok := availableSet[lang]; ok {
			out = append(out, lang)
		}
	}
	return out
}

func (s *SyncService) emitSyncCompleted(event SyncCompletedEvent) {
	if s.callbacks.OnSyncCompleted != nil {
		s.callbacks.OnSyncCompleted(event)
	}
}

func (s *SyncService) emitSyncFailed(event SyncFailedEvent) {
	if s.callbacks.OnSyncFailed != nil {
		s.callbacks.OnSyncFailed(event)
	}
}

func (s *SyncService) emitSyncAllCompleted(result SyncResult) {
	if s.callbacks.OnSyncAllCompleted != nil {
		s.callbacks.OnSyncAllCompleted(result)
	}
}
