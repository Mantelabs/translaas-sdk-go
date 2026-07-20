package cachefile

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SyncFromOfflineZip downloads the offline ZIP for project via the inner client and imports it.
// It is a no-op when GetOfflineCache returns NotModified or empty content.
func (s *SyncService) SyncFromOfflineZip(ctx context.Context, project string) error {
	project = strings.TrimSpace(project)
	if project == "" {
		return fmt.Errorf("cachefile: project must not be empty")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.client.GetOfflineCache(ctx, project)
	if err != nil {
		s.emitSyncFailed(SyncFailedEvent{Project: project, Err: err})
		return err
	}
	if result == nil || result.NotModified || len(result.Content) == 0 {
		return nil
	}

	if fp, ok := s.cache.(*FileProvider); ok {
		if err := fp.ImportOfflineBundle(ctx, project, result.Content); err != nil {
			s.emitSyncFailed(SyncFailedEvent{Project: project, Err: err})
			return err
		}
	} else {
		bundle, err := ParseOfflineZip(result.Content)
		if err != nil {
			s.emitSyncFailed(SyncFailedEvent{Project: project, Err: err})
			return err
		}
		key, err := ResolveProjectKey(bundle, project)
		if err != nil {
			s.emitSyncFailed(SyncFailedEvent{Project: project, Err: err})
			return err
		}
		if err := applyOfflineBundle(ctx, s.cache, project, key, bundle); err != nil {
			s.emitSyncFailed(SyncFailedEvent{Project: project, Err: err})
			return err
		}
	}

	s.emitSyncCompleted(SyncCompletedEvent{
		Project:  project,
		SyncedAt: time.Now().UTC(),
	})
	return nil
}
