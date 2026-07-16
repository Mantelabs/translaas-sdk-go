package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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
		return string(body), nil
	case http.StatusNoContent:
		assignResponseContext(resp, cfg.requestContext, false)
		return entry, nil
	case http.StatusNotModified:
		assignResponseContext(resp, cfg.requestContext, true)
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

func applyTextContext(req *models.GetTranslationRequest, ctx *models.RequestContext, defaultProjectID string) {
	if ctx != nil {
		if ctx.Project != "" {
			req.Project = ctx.Project
		}
		if ctx.Channel != "" {
			req.Channel = ctx.Channel
		}
		if ctx.Version != "" {
			req.Version = ctx.Version
		}
	}
	if req.Project == "" && defaultProjectID != "" {
		req.Project = defaultProjectID
	}
}

func assignResponseContext(resp *http.Response, ctx *models.RequestContext, notModified bool) {
	if ctx == nil {
		return
	}
	ctx.NotModified = notModified
	if etag := resp.Header.Get("ETag"); etag != "" {
		ctx.ResponseETag = etag
	}
}

func handleAPIError(statusCode int, body []byte) error {
	content := string(body)
	fallback := fmt.Sprintf("API request failed with status code %d.", statusCode)

	apiErr := &models.APIError{
		StatusCode:      statusCode,
		Message:         fallback,
		ResponseContent: content,
	}

	parsed, err := models.ParseTranslaasError(body)
	if err == nil && parsed != nil {
		apiErr.Message = parsed.FormatMessage(fallback)
		apiErr.Code = parsed.Code
	}
	return apiErr
}

func (c *client) mapTransportError(ctx context.Context, err error) error {
	if errors.Is(err, context.Canceled) {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.DeadlineExceeded) || isTimeoutError(err) {
		seconds := c.timeout.Seconds()
		return &models.APIError{
			StatusCode: http.StatusRequestTimeout,
			Message:    fmt.Sprintf("Request timed out after %g seconds.", seconds),
		}
	}
	return &models.APIError{
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf("Failed to retrieve translation: %s", err.Error()),
	}
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded")
}

func copyStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return make(map[string]string)
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
