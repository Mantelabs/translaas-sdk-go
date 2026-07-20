package cachefile

import (
	"context"
	"fmt"
	"strings"

	"github.com/acuencadev/translaas-sdk-go/models"
)

// ImportOfflineBundle parses zipBytes and persists the matching project into this provider's cache directory.
// It uses SaveProject and SaveLocales so atomic writes and manifest updates stay consistent with API sync.
func (p *FileProvider) ImportOfflineBundle(ctx context.Context, project string, zipBytes []byte) error {
	project = strings.TrimSpace(project)
	if project == "" {
		return offlineCacheErr(p.dir, project, "", "project must not be empty", fmt.Errorf("empty project"))
	}
	if err := checkContext(ctx, p.dir, project, ""); err != nil {
		return err
	}

	bundle, err := ParseOfflineZip(zipBytes)
	if err != nil {
		if cacheErr, ok := err.(*models.OfflineCacheError); ok {
			cacheErr.CacheDirectory = p.dir
			cacheErr.Project = project
			return cacheErr
		}
		return offlineCacheErr(p.dir, project, "", "parse offline ZIP", err)
	}

	key, err := ResolveProjectKey(bundle, project)
	if err != nil {
		return offlineCacheErr(p.dir, project, "", err.Error(), err)
	}

	return applyOfflineBundle(ctx, p, project, key, bundle)
}

func applyOfflineBundle(ctx context.Context, cache Provider, project, key string, bundle *OfflineBundle) error {
	hasLocales := false
	if locales, ok := bundle.LocalesByProject[key]; ok {
		hasLocales = true
		data := locales.Data
		if err := cache.SaveLocales(ctx, project, &data, saveOptionsFromWrapper(locales.ExpiresAt)...); err != nil {
			return err
		}
	}

	projectsByLang := bundle.ProjectsByProjectLang[key]
	if !hasLocales && len(projectsByLang) == 0 {
		return fmt.Errorf("no offline data found for project key %q", key)
	}

	for lang, wrapped := range projectsByLang {
		data := wrapped.Data
		if err := cache.SaveProject(ctx, project, lang, &data, saveOptionsFromWrapper(wrapped.ExpiresAt)...); err != nil {
			return err
		}
	}

	return nil
}
