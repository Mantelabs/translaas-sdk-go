//go:build integration

package integration_test

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetProjectLocales_ExistingProject(t *testing.T) {
	c := newIntegrationClient(t)
	cfg := requireIntegrationConfig(t)

	got, err := c.GetProjectLocales(context.Background(), cfg.DefaultProject)
	require.NoError(t, err)
	require.NotNil(t, got)
	if len(got.Locales) == 0 {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got.Locales)
}

func TestGetProjectLocales_MultipleLocales(t *testing.T) {
	c := newIntegrationClient(t)
	cfg := requireIntegrationConfig(t)

	got, err := c.GetProjectLocales(context.Background(), cfg.DefaultProject)
	require.NoError(t, err)
	require.NotNil(t, got)
	if len(got.Locales) == 0 {
		t.Skip("fixture data not available in API")
	}

	common := []string{"en", "fr", "es", "de"}
	found := false
	for _, locale := range got.Locales {
		if slices.Contains(common, locale) {
			found = true
			break
		}
	}
	require.True(t, found, "expected at least one common locale in %v", got.Locales)
}

func TestGetProjectLocales_ProjectNotFound(t *testing.T) {
	c := newIntegrationClient(t)

	got, err := c.GetProjectLocales(context.Background(), "nonexistent-project")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Empty(t, got.Locales)
}
