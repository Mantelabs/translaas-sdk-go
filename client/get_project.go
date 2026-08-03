package client

import (
	"context"
	"net/http"

	"github.com/Mantelabs/translaas-sdk-go/cache"
	"github.com/Mantelabs/translaas-sdk-go/models"
)

type getProjectConfig struct {
	format         string
	requestContext *models.RequestContext
}

// GetProjectOption configures a single GetProject call.
type GetProjectOption func(*getProjectConfig)

// WithProjectFormat sets the response format query parameter (for example "flat-json").
func WithProjectFormat(format string) GetProjectOption {
	return func(cfg *getProjectConfig) {
		cfg.format = format
	}
}

// WithProjectRequestContext supplies per-request channel, version, includeContext, and conditional headers.
func WithProjectRequestContext(rc *models.RequestContext) GetProjectOption {
	return func(cfg *getProjectConfig) {
		cfg.requestContext = rc
	}
}

// GetProject retrieves the full project translation payload for one language.
func (c *client) GetProject(ctx context.Context, project, lang string, opts ...GetProjectOption) (*models.TranslationProject, error) {
	if err := requireNonEmpty(project, "project"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(lang, "lang"); err != nil {
		return nil, err
	}

	cfg := getProjectConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.requestContext != nil {
		cfg.requestContext.Reset()
	}

	reqModel := models.GetProjectTranslationsRequest{
		Project: project,
		Lang:    lang,
		Format:  cfg.format,
	}
	applySnapshotContext(&reqModel.Channel, &reqModel.Version, &reqModel.IncludeContext, cfg.requestContext)

	channel, version := snapshotChannelVersion(cfg.requestContext)
	includeContext := snapshotIncludeContext(cfg.requestContext)
	cacheKey := cache.ProjectKey(project, lang, cfg.format, channel, version, includeContext)

	if c.cachingEnabled("project") {
		if cached, ok := c.tryCacheGetProject(ctx, cacheKey); ok {
			return cached, nil
		}
	}

	httpReq, err := c.buildGETRequest(ctx, sdkTranslationsPrefix+"/project", reqModel, "application/json", cfg.requestContext)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	dest := &models.TranslationProject{}
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
		if c.cachingEnabled("project") {
			c.cacheSetProject(ctx, cacheKey, dest)
		}
		return dest, nil
	case http.StatusNoContent:
		assignResponseContext(resp, cfg.requestContext, false)
		return emptyTranslationProject(), nil
	case http.StatusNotModified:
		assignResponseContext(resp, cfg.requestContext, true)
		if c.cacheProvider != nil {
			if cached, ok := c.tryCacheGetProject(ctx, cacheKey); ok {
				return cached, nil
			}
		}
		return emptyTranslationProject(), nil
	default:
		body, _ := readResponseBody(resp)
		return nil, handleAPIError(resp.StatusCode, body)
	}
}
