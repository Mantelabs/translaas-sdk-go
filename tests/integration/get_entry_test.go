//go:build integration

package integration_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/models"
	"github.com/stretchr/testify/require"
)

func TestGetEntry_ExistingEntry(t *testing.T) {
	c := newIntegrationClient(t)

	got, err := c.GetEntry(context.Background(), fixtureGroup, fixtureEntrySave, fixtureLang)
	require.NoError(t, err)
	require.NotEmpty(t, got)
}

func TestGetEntry_WithPluralization(t *testing.T) {
	c := newIntegrationClient(t)

	got, err := c.GetEntry(
		context.Background(),
		fixtureGroup,
		fixtureEntryCount,
		fixtureLang,
		client.WithNumber(5),
	)
	require.NoError(t, err)
	require.NotEmpty(t, got)
}

func TestGetEntry_NotFoundReturnsEntryKey(t *testing.T) {
	c := newIntegrationClient(t)

	const entry = "nonexistent.entry"
	got, err := c.GetEntry(context.Background(), "nonexistent", entry, fixtureLang)
	require.NoError(t, err)
	require.Equal(t, entry, got)
}

func TestGetEntry_InvalidAPIKey(t *testing.T) {
	cfg := requireIntegrationConfig(t)
	c := newClientWithOptions(t, cfg, client.Options{
		APIKey:  "invalid-api-key",
		BaseURL: cfg.BaseURL,
	})

	_, err := c.GetEntry(context.Background(), fixtureGroup, fixtureEntrySave, fixtureLang)
	require.Error(t, err)

	var apiErr *models.APIError
	require.True(t, errors.As(err, &apiErr))
	require.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, apiErr.StatusCode)
}
