package client

import (
	"context"

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
	result, err := handleJSONGETStatus(resp, cfg.requestContext, dest, func() any {
		return emptyTranslationGroup()
	})
	if err != nil {
		return nil, err
	}
	return result.(*models.TranslationGroup), nil
}
