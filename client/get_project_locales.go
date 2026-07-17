package client

import (
	"context"

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
	result, err := handleJSONGETStatus(resp, cfg.requestContext, dest, func() any {
		return emptyProjectLocales()
	})
	if err != nil {
		return nil, err
	}
	return result.(*models.ProjectLocales), nil
}
