//go:build integration

package integration_test

import "os"

const (
	defaultBaseURL    = "https://sdk-api.translaas.local"
	defaultProject    = "test-project"
	fixtureGroup      = "ui"
	fixtureEntrySave  = "button.save"
	fixtureEntryCount = "items.count"
	fixtureLang       = "en"
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
