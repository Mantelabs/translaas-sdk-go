//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/stretchr/testify/require"
)

func TestGetProject_ExistingProject(t *testing.T) {
	c := newIntegrationClient(t)
	cfg := requireIntegrationConfig(t)

	got, err := c.GetProject(context.Background(), cfg.DefaultProject, fixtureLang)
	requireNoErrorOrSkipNotFound(t, err)
	require.NotNil(t, got)
	if len(got.Groups) == 0 {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got.Groups)
}

func TestGetProject_WithFormat(t *testing.T) {
	c := newIntegrationClient(t)
	cfg := requireIntegrationConfig(t)

	got, err := c.GetProject(
		context.Background(),
		cfg.DefaultProject,
		fixtureLang,
		client.WithProjectFormat("json"),
	)
	requireNoErrorOrSkipNotFound(t, err)
	require.NotNil(t, got)
	if len(got.Groups) == 0 {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got.Groups)
}

func TestGetProject_ProjectNotFound(t *testing.T) {
	c := newIntegrationClient(t)

	got, err := c.GetProject(context.Background(), "nonexistent-project", fixtureLang)
	if acceptSDKNotFound(t, err) {
		return
	}
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Empty(t, got.Groups)
}

func TestGetProject_MultipleGroups(t *testing.T) {
	c := newIntegrationClient(t)
	cfg := requireIntegrationConfig(t)

	got, err := c.GetProject(context.Background(), cfg.DefaultProject, fixtureLang)
	requireNoErrorOrSkipNotFound(t, err)
	require.NotNil(t, got)
	if len(got.Groups) == 0 {
		t.Skip("fixture data not available in API")
	}

	walked := 0
	for groupName := range got.Groups {
		group, groupErr := got.GetGroup(groupName)
		require.NoError(t, groupErr)
		if group == nil || len(group.Entries) == 0 {
			continue
		}
		walked++
	}
	if walked == 0 {
		t.Skip("fixture data not available in API")
	}
}
