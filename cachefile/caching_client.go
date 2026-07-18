package cachefile

import (
	"context"
	"fmt"
	"strings"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/models"
)

var _ client.Client = (*CachingClient)(nil)

// CachingClient decorates client.Client with offline fallback semantics.
type CachingClient struct {
	inner client.Client
	cache Provider
	opts  Options
}

// NewCachingClient wraps inner with offline cache behavior.
func NewCachingClient(inner client.Client, cache Provider, opts Options) (*CachingClient, error) {
	if inner == nil {
		return nil, fmt.Errorf("cachefile: inner client must not be nil")
	}
	if cache == nil {
		return nil, fmt.Errorf("cachefile: cache provider must not be nil")
	}
	if strings.TrimSpace(opts.DefaultProjectID) == "" {
		return nil, fmt.Errorf("cachefile: DefaultProjectID must not be empty")
	}
	opts.DefaultProjectID = strings.TrimSpace(opts.DefaultProjectID)
	return &CachingClient{
		inner: inner,
		cache: cache,
		opts:  opts,
	}, nil
}

type getEntryConfig struct {
	number         *float64
	parameters     map[string]string
	requestContext *models.RequestContext
}

func parseGetEntryConfig(opts ...client.GetEntryOption) getEntryConfig {
	parsed := client.ParseGetEntryOptions(opts...)
	return getEntryConfig{
		number:         parsed.Number,
		parameters:     parsed.Parameters,
		requestContext: parsed.RequestContext,
	}
}

// GetEntry resolves a translation with offline fallback behavior.
func (c *CachingClient) GetEntry(ctx context.Context, group, entry, lang string, opts ...client.GetEntryOption) (string, error) {
	switch c.opts.FallbackMode {
	case FallbackAPIFirst:
		return c.getEntryAPIFirst(ctx, group, entry, lang, opts...)
	case FallbackCacheOnly:
		return c.getEntryCacheOnly(ctx, group, entry, lang, opts...)
	default:
		return c.getEntryCacheFirst(ctx, group, entry, lang, opts...)
	}
}

func (c *CachingClient) getEntryCacheFirst(
	ctx context.Context,
	group, entry, lang string,
	opts ...client.GetEntryOption,
) (string, error) {
	cfg := parseGetEntryConfig(opts...)
	if value, ok, err := c.getEntryFromCache(ctx, group, entry, lang, cfg); err != nil {
		return "", err
	} else if ok {
		return value, nil
	}

	result, err := c.inner.GetEntry(ctx, group, entry, lang, opts...)
	if err != nil {
		if isNetworkOrAPIError(err) {
			return "", models.NewOfflineCacheMissError(c.opts.DefaultProjectID, lang, group, entry)
		}
		return "", err
	}

	c.tryUpdateGroupCache(ctx, c.opts.DefaultProjectID, group, lang)
	return result, nil
}

func (c *CachingClient) getEntryAPIFirst(
	ctx context.Context,
	group, entry, lang string,
	opts ...client.GetEntryOption,
) (string, error) {
	cfg := parseGetEntryConfig(opts...)

	result, err := c.inner.GetEntry(ctx, group, entry, lang, opts...)
	if err == nil {
		c.tryUpdateGroupCache(ctx, c.opts.DefaultProjectID, group, lang)
		return result, nil
	}
	if !isNetworkOrAPIError(err) {
		return "", err
	}

	if value, ok, cacheErr := c.getEntryFromCache(ctx, group, entry, lang, cfg); cacheErr != nil {
		return "", cacheErr
	} else if ok {
		return value, nil
	}

	return "", models.NewOfflineCacheMissError(c.opts.DefaultProjectID, lang, group, entry)
}

func (c *CachingClient) getEntryCacheOnly(
	ctx context.Context,
	group, entry, lang string,
	opts ...client.GetEntryOption,
) (string, error) {
	cfg := parseGetEntryConfig(opts...)
	if value, ok, err := c.getEntryFromCache(ctx, group, entry, lang, cfg); err != nil {
		return "", err
	} else if ok {
		return value, nil
	}
	return "", models.NewOfflineCacheMissError(c.opts.DefaultProjectID, lang, group, entry)
}

func (c *CachingClient) getEntryFromCache(
	ctx context.Context,
	group, entry, lang string,
	cfg getEntryConfig,
) (string, bool, error) {
	cachedGroup, err := c.cache.GetGroup(ctx, c.opts.DefaultProjectID, group, lang)
	if err != nil {
		return "", false, err
	}
	if cachedGroup == nil {
		return "", false, nil
	}
	value, ok := resolveEntryFromGroup(cachedGroup, entry, cfg.number, cfg.parameters)
	return value, ok, nil
}

// GetGroup retrieves a group with offline fallback behavior.
func (c *CachingClient) GetGroup(
	ctx context.Context,
	project, group, lang string,
	opts ...client.GetGroupOption,
) (*models.TranslationGroup, error) {
	switch c.opts.FallbackMode {
	case FallbackAPIFirst:
		return c.getGroupAPIFirst(ctx, project, group, lang, opts...)
	case FallbackCacheOnly:
		return c.getGroupCacheOnly(ctx, project, group, lang)
	default:
		return c.getGroupCacheFirst(ctx, project, group, lang, opts...)
	}
}

func (c *CachingClient) getGroupCacheFirst(
	ctx context.Context,
	project, group, lang string,
	opts ...client.GetGroupOption,
) (*models.TranslationGroup, error) {
	cached, err := c.cache.GetGroup(ctx, project, group, lang)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	result, err := c.inner.GetGroup(ctx, project, group, lang, opts...)
	if err != nil {
		if isNetworkOrAPIError(err) {
			return nil, models.NewOfflineCacheMissError(project, lang, group, "")
		}
		return nil, err
	}

	c.tryUpdateGroupCache(ctx, project, group, lang)
	return result, nil
}

func (c *CachingClient) getGroupAPIFirst(
	ctx context.Context,
	project, group, lang string,
	opts ...client.GetGroupOption,
) (*models.TranslationGroup, error) {
	result, err := c.inner.GetGroup(ctx, project, group, lang, opts...)
	if err == nil {
		c.tryUpdateGroupCache(ctx, project, group, lang)
		return result, nil
	}
	if !isNetworkOrAPIError(err) {
		return nil, err
	}

	cached, cacheErr := c.cache.GetGroup(ctx, project, group, lang)
	if cacheErr != nil {
		return nil, cacheErr
	}
	if cached != nil {
		return cached, nil
	}

	return nil, models.NewOfflineCacheMissError(project, lang, group, "")
}

func (c *CachingClient) getGroupCacheOnly(ctx context.Context, project, group, lang string) (*models.TranslationGroup, error) {
	cached, err := c.cache.GetGroup(ctx, project, group, lang)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}
	return nil, models.NewOfflineCacheMissError(project, lang, group, "")
}

// GetProject retrieves a project with offline fallback behavior.
func (c *CachingClient) GetProject(
	ctx context.Context,
	project, lang string,
	opts ...client.GetProjectOption,
) (*models.TranslationProject, error) {
	switch c.opts.FallbackMode {
	case FallbackAPIFirst:
		return c.getProjectAPIFirst(ctx, project, lang, opts...)
	case FallbackCacheOnly:
		return c.getProjectCacheOnly(ctx, project, lang)
	default:
		return c.getProjectCacheFirst(ctx, project, lang, opts...)
	}
}

func (c *CachingClient) getProjectCacheFirst(
	ctx context.Context,
	project, lang string,
	opts ...client.GetProjectOption,
) (*models.TranslationProject, error) {
	cached, err := c.cache.GetProject(ctx, project, lang)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	result, err := c.inner.GetProject(ctx, project, lang, opts...)
	if err != nil {
		if isNetworkOrAPIError(err) {
			return nil, models.NewOfflineCacheMissError(project, lang, "", "")
		}
		return nil, err
	}

	if err := c.cache.SaveProject(ctx, project, lang, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CachingClient) getProjectAPIFirst(
	ctx context.Context,
	project, lang string,
	opts ...client.GetProjectOption,
) (*models.TranslationProject, error) {
	result, err := c.inner.GetProject(ctx, project, lang, opts...)
	if err == nil {
		if saveErr := c.cache.SaveProject(ctx, project, lang, result); saveErr != nil {
			return nil, saveErr
		}
		return result, nil
	}
	if !isNetworkOrAPIError(err) {
		return nil, err
	}

	cached, cacheErr := c.cache.GetProject(ctx, project, lang)
	if cacheErr != nil {
		return nil, cacheErr
	}
	if cached != nil {
		return cached, nil
	}

	return nil, models.NewOfflineCacheMissError(project, lang, "", "")
}

func (c *CachingClient) getProjectCacheOnly(ctx context.Context, project, lang string) (*models.TranslationProject, error) {
	cached, err := c.cache.GetProject(ctx, project, lang)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}
	return nil, models.NewOfflineCacheMissError(project, lang, "", "")
}

// GetProjectLocales retrieves locales with offline fallback behavior.
func (c *CachingClient) GetProjectLocales(
	ctx context.Context,
	project string,
	opts ...client.GetProjectLocalesOption,
) (*models.ProjectLocales, error) {
	switch c.opts.FallbackMode {
	case FallbackAPIFirst:
		return c.getProjectLocalesAPIFirst(ctx, project, opts...)
	case FallbackCacheOnly:
		return c.getProjectLocalesCacheOnly(ctx, project)
	default:
		return c.getProjectLocalesCacheFirst(ctx, project, opts...)
	}
}

func (c *CachingClient) getProjectLocalesCacheFirst(
	ctx context.Context,
	project string,
	opts ...client.GetProjectLocalesOption,
) (*models.ProjectLocales, error) {
	cached, err := c.cache.GetLocales(ctx, project)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	result, err := c.inner.GetProjectLocales(ctx, project, opts...)
	if err != nil {
		if isNetworkOrAPIError(err) {
			return nil, newLocalesOfflineCacheError(project, err)
		}
		return nil, err
	}

	if err := c.cache.SaveLocales(ctx, project, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CachingClient) getProjectLocalesAPIFirst(
	ctx context.Context,
	project string,
	opts ...client.GetProjectLocalesOption,
) (*models.ProjectLocales, error) {
	result, err := c.inner.GetProjectLocales(ctx, project, opts...)
	if err == nil {
		if saveErr := c.cache.SaveLocales(ctx, project, result); saveErr != nil {
			return nil, saveErr
		}
		return result, nil
	}
	if !isNetworkOrAPIError(err) {
		return nil, err
	}

	cached, cacheErr := c.cache.GetLocales(ctx, project)
	if cacheErr != nil {
		return nil, cacheErr
	}
	if cached != nil {
		return cached, nil
	}

	return nil, newLocalesOfflineCacheError(project, err)
}

func (c *CachingClient) getProjectLocalesCacheOnly(ctx context.Context, project string) (*models.ProjectLocales, error) {
	cached, err := c.cache.GetLocales(ctx, project)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}
	return nil, newLocalesOfflineCacheError(project, nil)
}

// GetOfflineCache delegates to the inner client.
func (c *CachingClient) GetOfflineCache(
	ctx context.Context,
	project string,
	opts ...client.GetOfflineCacheOption,
) (*models.OfflineCacheDownloadResult, error) {
	return c.inner.GetOfflineCache(ctx, project, opts...)
}

// ReportMissingKeys delegates to the inner client.
func (c *CachingClient) ReportMissingKeys(ctx context.Context, keys []models.ReportMissingKeyItem) error {
	return c.inner.ReportMissingKeys(ctx, keys)
}

// ValidateAPIKey delegates to the inner client.
func (c *CachingClient) ValidateAPIKey(ctx context.Context) (*models.ValidateAPIKeyResponse, error) {
	return c.inner.ValidateAPIKey(ctx)
}
