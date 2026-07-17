package client

import (
	"context"
	"encoding/json"

	"github.com/acuencadev/translaas-sdk-go/cache"
	"github.com/acuencadev/translaas-sdk-go/models"
)

func shouldCache(mode cache.Mode, op string) bool {
	switch op {
	case "entry":
		return mode == cache.ModeEntry
	case "group":
		return mode == cache.ModeGroup || mode == cache.ModeProject
	case "project":
		return mode == cache.ModeProject
	case "locales":
		return mode != cache.ModeNone
	default:
		return false
	}
}

func (c *client) cachingEnabled(op string) bool {
	return c.cacheProvider != nil && shouldCache(c.cacheMode, op)
}

func (c *client) tryCacheGetString(ctx context.Context, key string) (string, bool) {
	var value string
	ok, err := c.cacheProvider.Get(ctx, key, &value)
	if err != nil || !ok {
		return "", false
	}
	return value, true
}

func (c *client) tryCacheGetGroup(ctx context.Context, key string) (*models.TranslationGroup, bool) {
	var value models.TranslationGroup
	ok, err := c.cacheProvider.Get(ctx, key, &value)
	if err != nil || !ok {
		return nil, false
	}
	return &value, true
}

func (c *client) tryCacheGetProject(ctx context.Context, key string) (*models.TranslationProject, bool) {
	var value models.TranslationProject
	ok, err := c.cacheProvider.Get(ctx, key, &value)
	if err != nil || !ok {
		return nil, false
	}
	return &value, true
}

func (c *client) tryCacheGetLocales(ctx context.Context, key string) (*models.ProjectLocales, bool) {
	var value models.ProjectLocales
	ok, err := c.cacheProvider.Get(ctx, key, &value)
	if err != nil || !ok {
		return nil, false
	}
	return &value, true
}

func (c *client) cacheSetString(ctx context.Context, key, value string) {
	_ = c.cacheProvider.Set(ctx, key, value, c.cacheTTL)
}

func (c *client) cacheSetGroup(ctx context.Context, key string, value *models.TranslationGroup) {
	if cloned := cloneTranslationGroup(value); cloned != nil {
		_ = c.cacheProvider.Set(ctx, key, cloned, c.cacheTTL)
	}
}

func (c *client) cacheSetProject(ctx context.Context, key string, value *models.TranslationProject) {
	if cloned := cloneTranslationProject(value); cloned != nil {
		_ = c.cacheProvider.Set(ctx, key, cloned, c.cacheTTL)
	}
}

func (c *client) cacheSetLocales(ctx context.Context, key string, value *models.ProjectLocales) {
	if cloned := cloneProjectLocales(value); cloned != nil {
		_ = c.cacheProvider.Set(ctx, key, cloned, c.cacheTTL)
	}
}

func resolveEntryProject(reqCtx *models.RequestContext, defaultProjectID string) string {
	if reqCtx != nil && reqCtx.Project != "" {
		return reqCtx.Project
	}
	return defaultProjectID
}

func snapshotChannelVersion(reqCtx *models.RequestContext) (channel, version string) {
	if reqCtx == nil {
		return "", ""
	}
	return reqCtx.Channel, reqCtx.Version
}

func snapshotIncludeContext(reqCtx *models.RequestContext) *bool {
	if reqCtx == nil {
		return nil
	}
	return reqCtx.IncludeContext
}

func cloneTranslationGroup(src *models.TranslationGroup) *models.TranslationGroup {
	if src == nil {
		return nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var dst models.TranslationGroup
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil
	}
	return &dst
}

func cloneTranslationProject(src *models.TranslationProject) *models.TranslationProject {
	if src == nil {
		return nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var dst models.TranslationProject
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil
	}
	return &dst
}

func cloneProjectLocales(src *models.ProjectLocales) *models.ProjectLocales {
	if src == nil {
		return nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var dst models.ProjectLocales
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil
	}
	return &dst
}
