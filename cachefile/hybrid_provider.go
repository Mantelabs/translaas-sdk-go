package cachefile

import (
	"context"
	"fmt"
	"time"

	"github.com/acuencadev/translaas-sdk-go/models"
	lru "github.com/hashicorp/golang-lru/v2/expirable"
)

var _ Provider = (*HybridProvider)(nil)

// HybridProvider combines an expirable LRU memory cache (L1) with a disk Provider (L2).
type HybridProvider struct {
	l2   Provider
	opts HybridOptions
	l1   *hybridL1
}

type hybridL1 struct {
	projects *lru.LRU[string, *models.TranslationProject]
	groups   *lru.LRU[string, *models.TranslationGroup]
	locales  *lru.LRU[string, *models.ProjectLocales]
}

// NewHybridProvider wraps l2 with an optional in-memory L1 cache.
func NewHybridProvider(l2 Provider, opts HybridOptions) (*HybridProvider, error) {
	if l2 == nil {
		return nil, fmt.Errorf("cachefile: L2 provider must not be nil")
	}

	opts = normalizeHybridOptions(opts)
	p := &HybridProvider{
		l2:   l2,
		opts: opts,
	}
	if opts.Enabled {
		p.l1 = newHybridL1(opts.MaxEntries, opts.MemoryExpiration)
	}
	return p, nil
}

func newHybridL1(maxEntries int, ttl time.Duration) *hybridL1 {
	return &hybridL1{
		projects: lru.NewLRU[string, *models.TranslationProject](maxEntries, nil, ttl),
		groups:   lru.NewLRU[string, *models.TranslationGroup](maxEntries, nil, ttl),
		locales:  lru.NewLRU[string, *models.ProjectLocales](maxEntries, nil, ttl),
	}
}

// ClearMemoryCache removes all L1 entries without touching L2.
func (p *HybridProvider) ClearMemoryCache() {
	if p.l1 == nil {
		return
	}
	p.l1.projects.Purge()
	p.l1.groups.Purge()
	p.l1.locales.Purge()
}

// MemoryCacheStats returns current L1 entry counts by partition.
func (p *HybridProvider) MemoryCacheStats() (projects, groups, locales int) {
	if p.l1 == nil {
		return 0, 0, 0
	}
	return p.l1.projects.Len(), p.l1.groups.Len(), p.l1.locales.Len()
}

// Warmup loads project data from L2 into L1 for the given project and language.
func (p *HybridProvider) Warmup(ctx context.Context, project, lang string) (bool, error) {
	if p.l1 == nil {
		return false, nil
	}

	data, err := p.l2.GetProject(ctx, project, lang)
	if err != nil {
		return false, err
	}
	if data == nil {
		return false, nil
	}

	p.l1.projects.Add(projectCacheKey(project, lang), data)
	p.cacheProjectGroupsL1(project, lang, data)
	return true, nil
}

// GetProject checks L1, then L2, promoting L2 hits into L1.
func (p *HybridProvider) GetProject(ctx context.Context, project, lang string) (*models.TranslationProject, error) {
	if p.l1 != nil {
		if cached, ok := p.l1.projects.Get(projectCacheKey(project, lang)); ok {
			return cached, nil
		}
	}

	result, err := p.l2.GetProject(ctx, project, lang)
	if err != nil || result == nil || p.l1 == nil {
		return result, err
	}

	p.l1.projects.Add(projectCacheKey(project, lang), result)
	return result, nil
}

// SaveProject updates L1 and persists to L2.
func (p *HybridProvider) SaveProject(
	ctx context.Context,
	project, lang string,
	data *models.TranslationProject,
	opts ...SaveOption,
) error {
	if p.l1 != nil {
		p.l1.projects.Add(projectCacheKey(project, lang), data)
		p.cacheProjectGroupsL1(project, lang, data)
	}
	return p.l2.SaveProject(ctx, project, lang, data, opts...)
}

// GetGroup checks L1, then L2, promoting L2 hits into L1.
func (p *HybridProvider) GetGroup(ctx context.Context, project, group, lang string) (*models.TranslationGroup, error) {
	if p.l1 != nil {
		if cached, ok := p.l1.groups.Get(groupCacheKey(project, group, lang)); ok {
			return cached, nil
		}
	}

	result, err := p.l2.GetGroup(ctx, project, group, lang)
	if err != nil || result == nil || p.l1 == nil {
		return result, err
	}

	p.l1.groups.Add(groupCacheKey(project, group, lang), result)
	return result, nil
}

// GetLocales checks L1, then L2, promoting L2 hits into L1.
func (p *HybridProvider) GetLocales(ctx context.Context, project string) (*models.ProjectLocales, error) {
	if p.l1 != nil {
		if cached, ok := p.l1.locales.Get(localesCacheKey(project)); ok {
			return cached, nil
		}
	}

	result, err := p.l2.GetLocales(ctx, project)
	if err != nil || result == nil || p.l1 == nil {
		return result, err
	}

	p.l1.locales.Add(localesCacheKey(project), result)
	return result, nil
}

// SaveLocales updates L1 and persists to L2.
func (p *HybridProvider) SaveLocales(
	ctx context.Context,
	project string,
	data *models.ProjectLocales,
	opts ...SaveOption,
) error {
	if p.l1 != nil {
		p.l1.locales.Add(localesCacheKey(project), data)
	}
	return p.l2.SaveLocales(ctx, project, data, opts...)
}

// GetManifest delegates to L2 (manifest is not cached in L1).
func (p *HybridProvider) GetManifest(ctx context.Context) (*CacheManifest, error) {
	return p.l2.GetManifest(ctx)
}

// UpdateManifest delegates to L2 (manifest is not cached in L1).
func (p *HybridProvider) UpdateManifest(ctx context.Context, update func(*CacheManifest) error) error {
	return p.l2.UpdateManifest(ctx, update)
}

// IsCached reports whether data exists in L1 or L2.
func (p *HybridProvider) IsCached(ctx context.Context, project, lang string) (bool, error) {
	if p.l1 != nil {
		if _, ok := p.l1.projects.Get(projectCacheKey(project, lang)); ok {
			return true, nil
		}
	}
	return p.l2.IsCached(ctx, project, lang)
}

// Clear removes all L1 and L2 cache data.
func (p *HybridProvider) Clear(ctx context.Context) error {
	if p.l1 != nil {
		p.ClearMemoryCache()
	}
	return p.l2.Clear(ctx)
}

func (p *HybridProvider) cacheProjectGroupsL1(project, lang string, data *models.TranslationProject) {
	if p.l1 == nil || data == nil {
		return
	}
	for name := range data.Groups {
		group, err := data.GetGroup(name)
		if err != nil || group == nil {
			continue
		}
		p.l1.groups.Add(groupCacheKey(project, name, lang), group)
	}
}

func projectCacheKey(project, lang string) string {
	return "project:" + project + ":" + lang
}

func groupCacheKey(project, group, lang string) string {
	return "group:" + project + ":" + group + ":" + lang
}

func localesCacheKey(project string) string {
	return "locales:" + project
}
