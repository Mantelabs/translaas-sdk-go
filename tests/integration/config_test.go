//go:build integration

package integration_test

import "os"

const (
	defaultBaseURL       = "https://api.translaas.local"
	defaultProject       = "translaas-sdk-samples"
	fixtureGroup         = "common"
	fixtureEntry         = "welcome.message"
	fixturePluralGroup   = "messages"
	fixturePluralEntry   = "item"
	fixtureLang          = "en"
)

// Config holds integration test environment settings.
type Config struct {
	APIKey         string
	BaseURL        string
	DefaultProject string
	Enabled        bool
}

func loadConfig() Config {
	apiKey := os.Getenv("TRANSLAAS_API_KEY")
	baseURL := os.Getenv("TRANSLAAS_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	project := os.Getenv("TRANSLAAS_DEFAULT_PROJECT")
	if project == "" {
		project = defaultProject
	}
	return Config{
		APIKey:         apiKey,
		BaseURL:        baseURL,
		DefaultProject: project,
		Enabled:        apiKey != "",
	}
}
