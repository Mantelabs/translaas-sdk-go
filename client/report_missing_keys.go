package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Mantelabs/translaas-sdk-go/internal/httpx"
	"github.com/Mantelabs/translaas-sdk-go/models"
)

// ReportMissingKeys reports translation keys that could not be resolved at runtime.
// Requires a single-project API key; tenant-wide keys receive 401 from the server.
func (c *client) ReportMissingKeys(ctx context.Context, keys []models.ReportMissingKeyItem) error {
	if len(keys) == 0 {
		return nil
	}

	body, err := json.Marshal(models.ReportMissingKeysRequest{Keys: keys})
	if err != nil {
		return fmt.Errorf("marshal report-missing request: %w", err)
	}

	httpReq, err := c.buildPOSTRequest(ctx, sdkTranslationsPrefix+"/report-missing", body)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusAccepted {
		return nil
	}
	respBody, _ := readResponseBody(resp)
	return handleAPIError(resp.StatusCode, respBody)
}

func (c *client) buildPOSTRequest(ctx context.Context, endpoint string, body []byte) (*http.Request, error) {
	rawURL, err := httpx.BuildURL(c.baseURL, endpoint)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("X-Api-Key", c.apiKey)
	return httpReq, nil
}
