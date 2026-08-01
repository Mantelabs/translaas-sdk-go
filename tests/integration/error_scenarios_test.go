//go:build integration

package integration_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/models"
	"github.com/stretchr/testify/require"
)

func TestErrorScenarios_InvalidAPIKey(t *testing.T) {
	cfg := requireIntegrationConfig(t)
	c := newClientWithOptions(t, cfg, client.Options{
		APIKey:  "invalid-api-key-12345",
		BaseURL: cfg.BaseURL,
	})

	_, err := c.GetEntry(context.Background(), fixtureGroup, fixtureEntry, fixtureLang)
	require.Error(t, err)

	var apiErr *models.APIError
	require.True(t, errors.As(err, &apiErr))
	require.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, apiErr.StatusCode)
}

func TestErrorScenarios_InvalidBaseURL(t *testing.T) {
	cfg := requireIntegrationConfig(t)
	c := newClientWithOptions(t, cfg, client.Options{
		APIKey:  cfg.APIKey,
		BaseURL: "https://invalid-url-that-does-not-exist-12345.com/api",
	})

	_, err := c.GetEntry(context.Background(), fixtureGroup, fixtureEntry, fixtureLang)
	require.Error(t, err)
}

func TestErrorScenarios_RequestTimeout(t *testing.T) {
	cfg := requireIntegrationConfig(t)
	c := newClientWithOptions(t, cfg, client.Options{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Timeout: time.Millisecond,
	})

	_, err := c.GetEntry(context.Background(), fixtureGroup, fixtureEntry, fixtureLang)
	require.Error(t, err)

	var apiErr *models.APIError
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusRequestTimeout, apiErr.StatusCode)
	require.Contains(t, apiErr.Message, "timed out")
}

func TestErrorScenarios_EntryNotFoundReturnsKey(t *testing.T) {
	c := newIntegrationClient(t)

	const entry = "nonexistent-entry"
	got, err := c.GetEntry(context.Background(), "nonexistent-group", entry, "nonexistent-lang")
	if acceptSDKNotFound(t, err) {
		return
	}
	require.NoError(t, err)
	require.Equal(t, entry, got)
}
