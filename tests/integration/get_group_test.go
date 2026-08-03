//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/stretchr/testify/require"
)

func TestGetGroup_ExistingGroup(t *testing.T) {
	c := newIntegrationClient(t)
	cfg := requireIntegrationConfig(t)

	got, err := c.GetGroup(context.Background(), cfg.DefaultProject, fixtureGroup, fixtureLang)
	requireNoErrorOrSkipNotFound(t, err)
	require.NotNil(t, got)
	if len(got.Entries) == 0 {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got.Entries)
}

func TestGetGroup_WithFormat(t *testing.T) {
	c := newIntegrationClient(t)
	cfg := requireIntegrationConfig(t)

	got, err := c.GetGroup(
		context.Background(),
		cfg.DefaultProject,
		fixtureGroup,
		fixtureLang,
		client.WithGroupFormat("json"),
	)
	requireNoErrorOrSkipNotFound(t, err)
	require.NotNil(t, got)
	if len(got.Entries) == 0 {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got.Entries)
}

func TestGetGroup_GroupNotFound(t *testing.T) {
	c := newIntegrationClient(t)
	cfg := requireIntegrationConfig(t)

	got, err := c.GetGroup(context.Background(), cfg.DefaultProject, "nonexistent-group", fixtureLang)
	if acceptSDKNotFound(t, err) {
		return
	}
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Empty(t, got.Entries)
}

func TestGetGroup_ProjectNotFound(t *testing.T) {
	c := newIntegrationClient(t)

	got, err := c.GetGroup(context.Background(), "nonexistent-project", fixtureGroup, fixtureLang)
	if acceptSDKNotFound(t, err) {
		return
	}
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Empty(t, got.Entries)
}
