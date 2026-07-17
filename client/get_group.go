package client

import (
	"context"
	"net/http"

	"github.com/acuencadev/translaas-sdk-go/cache"
	"github.com/acuencadev/translaas-sdk-go/models"
)

type getGroupConfig struct {
	format         string
	requestContext *models.RequestContext
}

// GetGroupOption configures a single GetGroup call.
type GetGroupOption func(*getGroupConfig)

// WithGroupFormat sets the response format query parameter (for example "flat-json").
func WithGroupFormat(format string) GetGroupOption {
	return func(cfg *getGroupConfig) {
		cfg.format = format
	}
}

// WithGroupRequestContext supplies per-request channel, version, includeContext, and conditional headers.
func WithGroupRequestContext(rc *models.RequestContext) GetGroupOption {
	return func(cfg *getGroupConfig) {
		cfg.requestContext = rc
	}
}

// GetGroup retrieves one translation group for a project and language.
func (c *client) GetGroup(ctx context.Context, project, group, lang string, opts ...GetGroupOption) (*models.TranslationGroup, error) {
	if err := requireNonEmpty(project, "project"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(group, "group"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(lang, "lang"); err != nil {
		return nil, err
	}

	cfg := getGroupConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.requestContext != nil {
		cfg.requestContext.Reset()
	}

	reqModel := models.GetGroupTranslationsRequest{
		Project: project,
		Group:   group,
		Lang:    lang,
		Format:  cfg.format,
	}
	applySnapshotContext(&reqModel.Channel, &reqModel.Version, &reqModel.IncludeContext, cfg.requestContext)

	channel, version := snapshotChannelVersion(cfg.requestContext)
	includeContext := snapshotIncludeContext(cfg.requestContext)
	cacheKey := cache.GroupKey(project, group, lang, cfg.format, channel, version, includeContext)

	if c.cachingEnabled("group") {
		if cached, ok := c.tryCacheGetGroup(ctx, cacheKey); ok {
			return cached, nil
		}
	}

	httpReq, err := c.buildGETRequest(ctx, sdkTranslationsPrefix+"/group", reqModel, "application/json", cfg.requestContext)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	dest := &models.TranslationGroup{}
	switch resp.StatusCode {
	case http.StatusOK:
		assignResponseContext(resp, cfg.requestContext, false)
		body, err := readResponseBody(resp)
		if err != nil {
			return nil, err
		}
		if err := decodeJSONBody(body, dest); err != nil {
			return nil, err
		}
		if c.cachingEnabled("group") {
			c.cacheSetGroup(ctx, cacheKey, dest)
		}
		return dest, nil
	case http.StatusNoContent:
		assignResponseContext(resp, cfg.requestContext, false)
		return emptyTranslationGroup(), nil
	case http.StatusNotModified:
		assignResponseContext(resp, cfg.requestContext, true)
		if c.cacheProvider != nil {
			if cached, ok := c.tryCacheGetGroup(ctx, cacheKey); ok {
				return cached, nil
			}
		}
		return emptyTranslationGroup(), nil
	default:
		body, _ := readResponseBody(resp)
		return nil, handleAPIError(resp.StatusCode, body)
	}
}
