package cachefile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GetManifest reads root manifest.json or returns (nil, nil) when absent.
func (p *FileProvider) GetManifest(ctx context.Context) (*CacheManifest, error) {
	if err := checkContext(ctx, p.dir, "", ""); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getManifestLocked()
}

func (p *FileProvider) getManifestLocked() (*CacheManifest, error) {
	manifest, err := parseJSONFile[CacheManifest](p.manifestFile())
	if err != nil {
		return nil, offlineCacheErr(p.dir, "", "", fmt.Sprintf("read manifest: %s", err.Error()), err)
	}
	return manifest, nil
}

// UpdateManifest read-modify-writes manifest.json atomically.
func (p *FileProvider) UpdateManifest(ctx context.Context, update func(*CacheManifest) error) error {
	if err := checkContext(ctx, p.dir, "", ""); err != nil {
		return err
	}
	if update == nil {
		return offlineCacheErr(p.dir, "", "", "manifest update function is nil", fmt.Errorf("nil update"))
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	return p.updateManifestLocked(ctx, update)
}

func (p *FileProvider) updateManifestLocked(ctx context.Context, update func(*CacheManifest) error) error {
	manifest, err := p.getManifestLocked()
	if err != nil {
		return err
	}
	if manifest == nil {
		now := time.Now().UTC()
		manifest = &CacheManifest{
			Version:    ManifestVersion,
			SDKVersion: DefaultSDKVersion,
			CreatedAt:  now,
			LastSyncAt: now,
			Projects:   make(map[string]ProjectCacheInfo),
		}
	}
	if manifest.Projects == nil {
		manifest.Projects = make(map[string]ProjectCacheInfo)
	}

	if err := update(manifest); err != nil {
		return offlineCacheErr(p.dir, "", "", "update manifest", err)
	}

	manifest.LastSyncAt = time.Now().UTC()
	if err := writeJSONAtomic(ctx, p.manifestFile(), manifest); err != nil {
		return offlineCacheErr(p.dir, "", "", fmt.Sprintf("write manifest: %s", err.Error()), err)
	}
	return nil
}

func (p *FileProvider) recordProjectLanguageLocked(ctx context.Context, sanitizedProject, lang string) error {
	return p.updateManifestLocked(ctx, func(m *CacheManifest) error {
		info := m.Projects[sanitizedProject]
		info.Languages = appendLanguage(info.Languages, lang)
		info.LastSyncAt = time.Now().UTC()
		info.Status = "synced"
		m.Projects[sanitizedProject] = info
		return nil
	})
}

func (p *FileProvider) recordProjectLocalesLocked(ctx context.Context, sanitizedProject string, locales []string) error {
	return p.updateManifestLocked(ctx, func(m *CacheManifest) error {
		info := m.Projects[sanitizedProject]
		for _, lang := range normalizeLocales(locales) {
			info.Languages = appendLanguage(info.Languages, lang)
		}
		info.LastSyncAt = time.Now().UTC()
		info.Status = "synced"
		m.Projects[sanitizedProject] = info
		return nil
	})
}

func localesFromManifest(manifest *CacheManifest, sanitizedProject string) []string {
	if manifest == nil || manifest.Projects == nil {
		return nil
	}
	info, ok := manifest.Projects[sanitizedProject]
	if !ok {
		return nil
	}
	return normalizeLocales(info.Languages)
}

func (p *FileProvider) scanCachedLocaleDirectories(projectDir string) ([]string, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	locales := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectPath := filepath.Join(projectDir, entry.Name(), "project.json")
		if _, err := os.Stat(projectPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		locales = append(locales, entry.Name())
	}
	if len(locales) == 0 {
		return nil, nil
	}
	return locales, nil
}
