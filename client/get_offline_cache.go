package client

import (
	"context"
	"net/http"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

type getOfflineCacheConfig struct {
	requestContext *models.RequestContext
}

// GetOfflineCacheOption configures a single GetOfflineCache call.
type GetOfflineCacheOption func(*getOfflineCacheConfig)

// WithOfflineCacheRequestContext supplies per-request channel, version, includeContext, and conditional headers.
func WithOfflineCacheRequestContext(rc *models.RequestContext) GetOfflineCacheOption {
	return func(cfg *getOfflineCacheConfig) {
		cfg.requestContext = rc
	}
}

// GetOfflineCache downloads the offline translation bundle as a ZIP archive.
func (c *client) GetOfflineCache(ctx context.Context, project string, opts ...GetOfflineCacheOption) (*models.OfflineCacheDownloadResult, error) {
	if err := requireNonEmpty(project, "project"); err != nil {
		return nil, err
	}

	cfg := getOfflineCacheConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.requestContext != nil {
		cfg.requestContext.Reset()
	}

	reqModel := models.GetOfflineCacheRequest{Project: project}
	applySnapshotContext(&reqModel.Channel, &reqModel.Version, &reqModel.IncludeContext, cfg.requestContext)

	httpReq, err := c.buildGETRequest(ctx, sdkTranslationsPrefix+"/offline-cache", reqModel, "application/zip", cfg.requestContext)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	result := &models.OfflineCacheDownloadResult{}
	switch resp.StatusCode {
	case http.StatusOK:
		assignResponseContext(resp, cfg.requestContext, false)
		body, err := readResponseBody(resp)
		if err != nil {
			return nil, err
		}
		result.Content = body
		result.SuggestedFileName = parseContentDisposition(resp.Header.Get("Content-Disposition"))
		if etag := resp.Header.Get("ETag"); etag != "" {
			result.ETag = etag
		}
		return result, nil
	case http.StatusNotModified:
		assignResponseContext(resp, cfg.requestContext, true)
		result.NotModified = true
		if etag := resp.Header.Get("ETag"); etag != "" {
			result.ETag = etag
		}
		return result, nil
	default:
		body, _ := readResponseBody(resp)
		return nil, handleAPIError(resp.StatusCode, body)
	}
}
