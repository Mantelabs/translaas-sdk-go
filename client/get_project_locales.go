package client

import (
	"context"
	"net/http"

	"github.com/acuencadev/translaas-sdk-go/cache"
	"github.com/acuencadev/translaas-sdk-go/models"
)

type getProjectLocalesConfig struct {
	requestContext *models.RequestContext
}

// GetProjectLocalesOption configures a single GetProjectLocales call.
type GetProjectLocalesOption func(*getProjectLocalesConfig)

// WithLocalesRequestContext supplies per-request channel, version, and conditional headers.
func WithLocalesRequestContext(rc *models.RequestContext) GetProjectLocalesOption {
	return func(cfg *getProjectLocalesConfig) {
		cfg.requestContext = rc
	}
}

// GetProjectLocales lists locales available for a project.
func (c *client) GetProjectLocales(ctx context.Context, project string, opts ...GetProjectLocalesOption) (*models.ProjectLocales, error) {
	if err := requireNonEmpty(project, "project"); err != nil {
		return nil, err
	}

	cfg := getProjectLocalesConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.requestContext != nil {
		cfg.requestContext.Reset()
	}

	reqModel := models.GetProjectLocalesRequest{Project: project}
	applyChannelVersion(&reqModel.Channel, &reqModel.Version, cfg.requestContext)

	channel, version := snapshotChannelVersion(cfg.requestContext)
	cacheKey := cache.LocalesKey(project, channel, version)

	if c.cachingEnabled("locales") {
		if cached, ok := c.tryCacheGetLocales(ctx, cacheKey); ok {
			return cached, nil
		}
	}

	httpReq, err := c.buildGETRequest(ctx, sdkTranslationsPrefix+"/locales", reqModel, "application/json", cfg.requestContext)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	dest := &models.ProjectLocales{}
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
		if c.cachingEnabled("locales") {
			c.cacheSetLocales(ctx, cacheKey, dest)
		}
		return dest, nil
	case http.StatusNoContent:
		assignResponseContext(resp, cfg.requestContext, false)
		return emptyProjectLocales(), nil
	case http.StatusNotModified:
		assignResponseContext(resp, cfg.requestContext, true)
		if c.cacheProvider != nil {
			if cached, ok := c.tryCacheGetLocales(ctx, cacheKey); ok {
				return cached, nil
			}
		}
		return emptyProjectLocales(), nil
	default:
		body, _ := readResponseBody(resp)
		return nil, handleAPIError(resp.StatusCode, body)
	}
}
