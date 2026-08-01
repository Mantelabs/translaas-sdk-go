//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/service"
	"github.com/acuencadev/translaas-sdk-go/service/language"
	"github.com/stretchr/testify/require"
)

func TestServiceT_ExplicitLanguage(t *testing.T) {
	cfg := requireIntegrationConfig(t)

	httpClient, err := client.New(client.Options{
		APIKey:           cfg.APIKey,
		BaseURL:          cfg.BaseURL,
		DefaultProjectID: cfg.DefaultProject,
	})
	require.NoError(t, err)

	resolver, err := language.NewResolver(language.NewDefaultLanguageProvider(fixtureLang))
	require.NoError(t, err)

	svc, err := service.New(httpClient, service.Options{Resolver: resolver})
	require.NoError(t, err)

	got, err := svc.T(
		context.Background(),
		fixtureGroup,
		fixtureEntry,
		service.WithLang(fixtureLang),
	)
	require.NoError(t, err)
	if got == fixtureEntry {
		t.Skip("fixture data not available in API")
	}
	require.NotEmpty(t, got)
}
