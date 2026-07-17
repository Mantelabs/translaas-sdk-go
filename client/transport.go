package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/acuencadev/translaas-sdk-go/internal/httpx"
	"github.com/acuencadev/translaas-sdk-go/models"
)

func (c *client) buildGETRequest(
	ctx context.Context,
	endpoint string,
	queryModel any,
	accept string,
	reqCtx *models.RequestContext,
) (*http.Request, error) {
	rawURL, err := httpx.BuildURL(c.baseURL, endpoint)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse request url: %w", err)
	}
	if queryModel != nil {
		if err := httpx.AppendQueryValues(u, queryModel); err != nil {
			return nil, err
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("X-Api-Key", c.apiKey)
	if accept != "" {
		httpReq.Header.Set("Accept", accept)
	}
	if reqCtx != nil && reqCtx.IfNoneMatch != "" {
		httpReq.Header.Set("If-None-Match", reqCtx.IfNoneMatch)
	}
	return httpReq, nil
}

func (c *client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.mapTransportError(ctx, err)
	}
	return resp, nil
}

func decodeJSONBody(body []byte, dest any) error {
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return body, nil
}

func handleJSONGETStatus(
	resp *http.Response,
	reqCtx *models.RequestContext,
	dest any,
	empty func() any,
) (any, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		assignResponseContext(resp, reqCtx, false)
		body, err := readResponseBody(resp)
		if err != nil {
			return nil, err
		}
		if err := decodeJSONBody(body, dest); err != nil {
			return nil, err
		}
		return dest, nil
	case http.StatusNoContent:
		assignResponseContext(resp, reqCtx, false)
		return empty(), nil
	case http.StatusNotModified:
		assignResponseContext(resp, reqCtx, true)
		return empty(), nil
	default:
		body, _ := readResponseBody(resp)
		return nil, handleAPIError(resp.StatusCode, body)
	}
}

func applySnapshotContext(channel, version *string, includeContext **bool, ctx *models.RequestContext) {
	if ctx == nil {
		return
	}
	applyChannelVersion(channel, version, ctx)
	if ctx.IncludeContext != nil {
		*includeContext = ctx.IncludeContext
	}
}

func applyChannelVersion(channel, version *string, ctx *models.RequestContext) {
	if ctx == nil {
		return
	}
	if ctx.Channel != "" {
		*channel = ctx.Channel
	}
	if ctx.Version != "" {
		*version = ctx.Version
	}
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

func emptyTranslationGroup() *models.TranslationGroup {
	return &models.TranslationGroup{Entries: map[string]json.RawMessage{}}
}

func emptyTranslationProject() *models.TranslationProject {
	return &models.TranslationProject{Groups: map[string]json.RawMessage{}}
}

func emptyProjectLocales() *models.ProjectLocales {
	return &models.ProjectLocales{Locales: []string{}}
}

func parseContentDisposition(header string) string {
	if header == "" {
		return ""
	}
	if _, params, err := mime.ParseMediaType(header); err == nil {
		if filename, ok := params["filename"]; ok && filename != "" {
			return filename
		}
	}
	return parseFilenameStar(header)
}

func parseFilenameStar(header string) string {
	const prefix = "filename*="
	lower := strings.ToLower(header)
	idx := strings.Index(lower, prefix)
	if idx < 0 {
		return ""
	}
	value := strings.TrimSpace(header[idx+len(prefix):])
	if semi := strings.Index(value, ";"); semi >= 0 {
		value = value[:semi]
	}
	value = strings.Trim(value, `"`)
	parts := strings.SplitN(value, "''", 2)
	if len(parts) != 2 {
		return value
	}
	decoded, err := url.PathUnescape(parts[1])
	if err != nil {
		return parts[1]
	}
	return decoded
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

func requireNonEmpty(field, name string) error {
	if strings.TrimSpace(field) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
