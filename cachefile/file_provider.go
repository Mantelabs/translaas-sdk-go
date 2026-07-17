package cachefile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/acuencadev/translaas-sdk-go/models"
)

var _ Provider = (*FileProvider)(nil)

// FileProvider persists offline translation payloads as JSON on disk.
type FileProvider struct {
	dir string
	mu  sync.RWMutex
	now func() time.Time
}

// NewFileProvider creates a file-backed offline cache at cacheDirectory.
// Relative paths resolve against the process working directory.
func NewFileProvider(cacheDirectory string) (*FileProvider, error) {
	if cacheDirectory == "" {
		return nil, fmt.Errorf("cache directory must not be empty")
	}

	abs, err := filepath.Abs(cacheDirectory)
	if err != nil {
		return nil, fmt.Errorf("resolve cache directory: %w", err)
	}

	return &FileProvider{
		dir: abs,
		now: time.Now,
	}, nil
}

// CacheDirectory returns the absolute cache root path.
func (p *FileProvider) CacheDirectory() string {
	return p.dir
}

// GetProject returns cached project data or (nil, nil) on miss or expiry.
func (p *FileProvider) GetProject(ctx context.Context, project, lang string) (*models.TranslationProject, error) {
	if err := checkContext(ctx, p.dir, project, lang); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getProjectLocked(project, lang)
}

func (p *FileProvider) getProjectLocked(project, lang string) (*models.TranslationProject, error) {
	projectDir, _, err := p.sanitizedProjectDir(project)
	if err != nil {
		return nil, err
	}

	projectPath, err := p.projectFile(projectDir, lang)
	if err != nil {
		return nil, offlineCacheErr(p.dir, project, lang, fmt.Sprintf("invalid language: %s", lang), err)
	}

	wrapped, err := parseJSONFile[CachedProject](projectPath)
	if err != nil {
		return nil, offlineCacheErr(p.dir, project, lang, fmt.Sprintf("read project cache: %s", err.Error()), err)
	}
	if wrapped == nil || isExpired(wrapped.ExpiresAt, p.now()) {
		return nil, nil
	}

	return &wrapped.Data, nil
}

// SaveProject writes project data to disk and updates the root manifest.
func (p *FileProvider) SaveProject(
	ctx context.Context,
	project, lang string,
	data *models.TranslationProject,
	opts ...SaveOption,
) error {
	if err := checkContext(ctx, p.dir, project, lang); err != nil {
		return err
	}
	if data == nil {
		return offlineCacheErr(p.dir, project, lang, "project data must not be nil", fmt.Errorf("nil data"))
	}

	cfg := applySaveOptions(opts)

	p.mu.Lock()
	defer p.mu.Unlock()

	projectDir, sanitizedProject, err := p.sanitizedProjectDir(project)
	if err != nil {
		return err
	}

	projectPath, err := p.projectFile(projectDir, lang)
	if err != nil {
		return offlineCacheErr(p.dir, project, lang, fmt.Sprintf("invalid language: %s", lang), err)
	}

	wrapped := CachedProject{
		CachedAt:  cfg.cachedAt,
		ExpiresAt: cfg.expiresAt,
		Data:      *data,
	}
	if err := writeJSONAtomic(ctx, projectPath, wrapped); err != nil {
		return offlineCacheErr(p.dir, project, lang, fmt.Sprintf("write project cache: %s", err.Error()), err)
	}

	if err := p.recordProjectLanguageLocked(ctx, sanitizedProject, lang); err != nil {
		return err
	}
	return nil
}

// GetGroup returns a group extracted from the cached project payload.
func (p *FileProvider) GetGroup(ctx context.Context, project, group, lang string) (*models.TranslationGroup, error) {
	projectData, err := p.GetProject(ctx, project, lang)
	if err != nil || projectData == nil {
		return nil, err
	}

	groupData, err := projectData.GetGroup(group)
	if err != nil {
		return nil, offlineCacheErr(p.dir, project, lang, fmt.Sprintf("read group %s: %s", group, err.Error()), err)
	}
	return groupData, nil
}

// GetLocales returns cached locales, falling back to manifest or locale directories.
func (p *FileProvider) GetLocales(ctx context.Context, project string) (*models.ProjectLocales, error) {
	if err := checkContext(ctx, p.dir, project, ""); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getLocalesLocked(project)
}

func (p *FileProvider) getLocalesLocked(project string) (*models.ProjectLocales, error) {
	projectDir, sanitizedProject, err := p.sanitizedProjectDir(project)
	if err != nil {
		return nil, err
	}

	localesPath := p.localesFile(projectDir)
	wrapped, err := parseJSONFile[CachedLocales](localesPath)
	if err != nil {
		return nil, offlineCacheErr(p.dir, project, "", fmt.Sprintf("read locales cache: %s", err.Error()), err)
	}
	if wrapped != nil && !isExpired(wrapped.ExpiresAt, p.now()) {
		out := wrapped.Data
		if out.Project == "" {
			out.Project = project
		}
		return &out, nil
	}

	manifest, err := p.getManifestLocked()
	if err != nil {
		return nil, err
	}
	if locales := localesFromManifest(manifest, sanitizedProject); len(locales) > 0 {
		return &models.ProjectLocales{
			Project: project,
			Locales: locales,
		}, nil
	}

	scanned, err := p.scanCachedLocaleDirectories(projectDir)
	if err != nil {
		return nil, offlineCacheErr(p.dir, project, "", fmt.Sprintf("scan locale directories: %s", err.Error()), err)
	}
	if len(scanned) == 0 {
		return nil, nil
	}
	return &models.ProjectLocales{
		Project: project,
		Locales: scanned,
	}, nil
}

// SaveLocales writes locales.json and updates the root manifest.
func (p *FileProvider) SaveLocales(
	ctx context.Context,
	project string,
	data *models.ProjectLocales,
	opts ...SaveOption,
) error {
	if err := checkContext(ctx, p.dir, project, ""); err != nil {
		return err
	}
	if data == nil {
		return offlineCacheErr(p.dir, project, "", "locales data must not be nil", fmt.Errorf("nil data"))
	}

	cfg := applySaveOptions(opts)

	p.mu.Lock()
	defer p.mu.Unlock()

	projectDir, sanitizedProject, err := p.sanitizedProjectDir(project)
	if err != nil {
		return err
	}

	payload := *data
	if payload.Project == "" {
		payload.Project = project
	}

	wrapped := CachedLocales{
		CachedAt:  cfg.cachedAt,
		ExpiresAt: cfg.expiresAt,
		Data:      payload,
	}
	if err := writeJSONAtomic(ctx, p.localesFile(projectDir), wrapped); err != nil {
		return offlineCacheErr(p.dir, project, "", fmt.Sprintf("write locales cache: %s", err.Error()), err)
	}

	if err := p.recordProjectLocalesLocked(ctx, sanitizedProject, payload.Locales); err != nil {
		return err
	}
	return nil
}

// IsCached reports whether a non-expired project/language payload exists on disk.
func (p *FileProvider) IsCached(ctx context.Context, project, lang string) (bool, error) {
	if err := checkContext(ctx, p.dir, project, lang); err != nil {
		return false, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	projectDir, _, err := p.sanitizedProjectDir(project)
	if err != nil {
		return false, err
	}

	projectPath, err := p.projectFile(projectDir, lang)
	if err != nil {
		return false, offlineCacheErr(p.dir, project, lang, fmt.Sprintf("invalid language: %s", lang), err)
	}

	wrapped, err := parseJSONFile[CachedProject](projectPath)
	if err != nil {
		return false, offlineCacheErr(p.dir, project, lang, fmt.Sprintf("read project cache: %s", err.Error()), err)
	}
	if wrapped == nil || isExpired(wrapped.ExpiresAt, p.now()) {
		return false, nil
	}
	return true, nil
}

// Clear removes the entire cache directory tree.
func (p *FileProvider) Clear(ctx context.Context) error {
	if err := checkContext(ctx, p.dir, "", ""); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := os.RemoveAll(p.dir); err != nil && !os.IsNotExist(err) {
		return offlineCacheErr(p.dir, "", "", fmt.Sprintf("clear cache directory: %s", err.Error()), err)
	}
	return nil
}
