//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/models"
	"github.com/stretchr/testify/require"
)

func TestValidateAPIKey_ValidKey(t *testing.T) {
	c := newIntegrationClient(t)

	got, err := c.ValidateAPIKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, got)
	require.True(t, got.IsValid)
}

func TestNewWithResolvedProject_SingleProjectKey(t *testing.T) {
	cfg := requireIntegrationConfig(t)

	c, err := client.NewWithResolvedProject(context.Background(), client.Options{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	validate, err := c.ValidateAPIKey(context.Background())
	require.NoError(t, err)

	if models.ReadJSONULID(validate.ProjectID) == "" {
		t.Skip("API key is not single-project scoped")
	}

	got, err := c.GetEntry(context.Background(), fixtureGroup, fixtureEntry, fixtureLang)
	require.NoError(t, err)
	if got == fixtureEntry {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got)
}
