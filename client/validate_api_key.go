package client

import (
	"context"
	"net/http"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

const validateAPIKeyPath = "api/v1/api-keys/validate"

// ValidateAPIKey validates the configured API key and returns tenant/project scope metadata.
func (c *client) ValidateAPIKey(ctx context.Context) (*models.ValidateAPIKeyResponse, error) {
	httpReq, err := c.buildGETRequest(ctx, validateAPIKeyPath, nil, "application/json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := readResponseBody(resp)
		return nil, handleAPIError(resp.StatusCode, body)
	}

	body, err := readResponseBody(resp)
	if err != nil {
		return nil, err
	}
	dest := &models.ValidateAPIKeyResponse{}
	if err := decodeJSONBody(body, dest); err != nil {
		return nil, err
	}
	return dest, nil
}
