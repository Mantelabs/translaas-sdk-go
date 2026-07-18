package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/acuencadev/translaas-sdk-go/cache"
	"github.com/acuencadev/translaas-sdk-go/internal/httpx"
	"github.com/acuencadev/translaas-sdk-go/models"
)

type getEntryConfig struct {
	number         *float64
	parameters     map[string]string
	requestContext *models.RequestContext
}

// GetEntryOption configures a single GetEntry call.
type GetEntryOption func(*getEntryConfig)

// WithNumber sets the plural/interpolation count (merged as query key N).
func WithNumber(n float64) GetEntryOption {
	return func(cfg *getEntryConfig) {
		cfg.number = &n
	}
}

// WithParameters adds interpolation query parameters merged after the request DTO.
func WithParameters(params map[string]string) GetEntryOption {
	return func(cfg *getEntryConfig) {
		cfg.parameters = params
	}
}

// WithRequestContext supplies per-request channel, version, project, and conditional headers.
func WithRequestContext(rc *models.RequestContext) GetEntryOption {
	return func(cfg *getEntryConfig) {
		cfg.requestContext = rc
	}
}

// GetEntryCallOptions holds parsed GetEntry call options for decorators.
type GetEntryCallOptions struct {
	Number         *float64
	Parameters     map[string]string
	RequestContext *models.RequestContext
}

// ParseGetEntryOptions parses GetEntry options for use by decorators.
func ParseGetEntryOptions(opts ...GetEntryOption) GetEntryCallOptions {
	cfg := getEntryConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return GetEntryCallOptions{
		Number:         cfg.number,
		Parameters:     cfg.parameters,
		RequestContext: cfg.requestContext,
	}
}

// GetEntry retrieves a single rendered translation string.
func (c *client) GetEntry(ctx context.Context, group, entry, lang string, opts ...GetEntryOption) (string, error) {
	if strings.TrimSpace(group) == "" {
		return "", fmt.Errorf("group is required")
	}
	if strings.TrimSpace(entry) == "" {
		return "", fmt.Errorf("entry is required")
	}
	if strings.TrimSpace(lang) == "" {
		return "", fmt.Errorf("lang is required")
	}

	cfg := getEntryConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.requestContext != nil {
		cfg.requestContext.Reset()
	}

	reqModel := models.GetTranslationRequest{
		Group:  group,
		Entry:  entry,
		Lang:   lang,
		Number: cfg.number,
	}
	applyTextContext(&reqModel, cfg.requestContext, c.defaultProjectID)

	extra := copyStringMap(cfg.parameters)
	httpx.InjectPluralN(extra, cfg.number)

	channel, version := snapshotChannelVersion(cfg.requestContext)
	project := resolveEntryProject(cfg.requestContext, c.defaultProjectID)
	cacheKey := cache.EntryKey(group, entry, lang, cfg.number, cfg.parameters, project, channel, version)

	if c.cachingEnabled("entry") {
		if cached, ok := c.tryCacheGetString(ctx, cacheKey); ok {
			return cached, nil
		}
	}

	httpReq, err := c.buildTextRequest(ctx, reqModel, extra, cfg.requestContext)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", c.mapTransportError(ctx, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		assignResponseContext(resp, cfg.requestContext, false)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read response body: %w", err)
		}
		value := string(body)
		if c.cachingEnabled("entry") {
			c.cacheSetString(ctx, cacheKey, value)
		}
		return value, nil
	case http.StatusNoContent:
		assignResponseContext(resp, cfg.requestContext, false)
		return entry, nil
	case http.StatusNotModified:
		assignResponseContext(resp, cfg.requestContext, true)
		if c.cacheProvider != nil {
			if cached, ok := c.tryCacheGetString(ctx, cacheKey); ok {
				return cached, nil
			}
		}
		return "", nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return "", handleAPIError(resp.StatusCode, body)
	}
}

func (c *client) buildTextRequest(
	ctx context.Context,
	reqModel models.GetTranslationRequest,
	extra map[string]string,
	reqCtx *models.RequestContext,
) (*http.Request, error) {
	rawURL, err := httpx.BuildURL(c.baseURL, sdkTranslationsPrefix+"/text")
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse request url: %w", err)
	}
	if err := httpx.AppendQueryValues(u, reqModel); err != nil {
		return nil, err
	}
	httpx.MergeQueryParams(u, extra)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("X-Api-Key", c.apiKey)
	httpReq.Header.Set("Accept", "text/plain")
	if reqCtx != nil && reqCtx.IfNoneMatch != "" {
		httpReq.Header.Set("If-None-Match", reqCtx.IfNoneMatch)
	}
	return httpReq, nil
}
