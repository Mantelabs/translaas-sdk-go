package client

import (
	"context"

	"github.com/acuencadev/translaas-sdk-go/models"
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
	result, err := handleJSONGETStatus(resp, cfg.requestContext, dest, func() any {
		return emptyTranslationProject()
	})
	if err != nil {
		return nil, err
	}
	return result.(*models.TranslationProject), nil
}
