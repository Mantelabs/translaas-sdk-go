//go:build integration

package integration_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/Mantelabs/translaas-sdk-go/models"
	"github.com/stretchr/testify/require"
)

func TestGetEntry_ExistingEntry(t *testing.T) {
	c := newIntegrationClient(t)

	got, err := c.GetEntry(context.Background(), fixtureGroup, fixtureEntry, fixtureLang)
	requireNoErrorOrSkipNotFound(t, err)
	if got == "" || got == fixtureEntry {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got)
}

func TestGetEntry_WithPluralization(t *testing.T) {
	c := newIntegrationClient(t)

	got, err := c.GetEntry(
		context.Background(),
		fixturePluralGroup,
		fixturePluralEntry,
		fixtureLang,
		client.WithNumber(5),
	)
	requireNoErrorOrSkipNotFound(t, err)
	if got == "" || got == fixturePluralEntry {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got)
}

func TestGetEntry_NotFoundReturnsEntryKey(t *testing.T) {
	c := newIntegrationClient(t)

	const entry = "nonexistent.entry"
	got, err := c.GetEntry(context.Background(), "nonexistent", entry, fixtureLang)
	if acceptSDKNotFound(t, err) {
		return
	}
	require.NoError(t, err)
	require.Equal(t, entry, got)
}

func TestGetEntry_InvalidAPIKey(t *testing.T) {
	cfg := requireIntegrationConfig(t)
	c := newClientWithOptions(t, cfg, client.Options{
		APIKey:  "invalid-api-key",
		BaseURL: cfg.BaseURL,
	})

	_, err := c.GetEntry(context.Background(), fixtureGroup, fixtureEntry, fixtureLang)
	require.Error(t, err)

	var apiErr *models.APIError
	require.True(t, errors.As(err, &apiErr))
	require.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, apiErr.StatusCode)
}
