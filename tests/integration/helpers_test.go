//go:build integration

package integration_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/acuencadev/translaas-sdk-go/client"
	"github.com/acuencadev/translaas-sdk-go/models"
	"github.com/stretchr/testify/require"
)

const sdkNotFoundSkipMessage = "SDK resource not found (HTTP 404) — set TRANSLAAS_DEFAULT_PROJECT to an existing project id (default: translaas-sdk-samples)"

func isSDKNotFound(err error) bool {
	var apiErr *models.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func requireNoErrorOrSkipNotFound(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	if isSDKNotFound(err) {
		t.Skip(sdkNotFoundSkipMessage)
	}
	require.NoError(t, err)
}

// acceptSDKNotFound reports whether err is Mantelabs HTTP 404 (missing SDK resource).
// Returns false when err is nil; otherwise skips via require on unexpected errors.
func acceptSDKNotFound(t *testing.T, err error) bool {
	t.Helper()
	if err == nil {
		return false
	}
	if isSDKNotFound(err) {
		return true
	}
	require.NoError(t, err)
	return false
}

var (
	reachabilityOnce sync.Once
	apiReachable     bool
)

func probeAPIReachable(cfg Config) bool {
	reachabilityOnce.Do(func() {
		c, err := client.New(client.Options{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Timeout: 5 * time.Second,
		})
		if err != nil {
			return
		}
		_, err = c.ValidateAPIKey(context.Background())
		if err == nil {
			apiReachable = true
			return
		}
		var apiErr *models.APIError
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 401 || apiErr.StatusCode == 403) {
			apiReachable = true
			return
		}
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "no such host") ||
			strings.Contains(msg, "connection refused") ||
			strings.Contains(msg, "actively refused") {
			return
		}
		apiReachable = true
	})
	return apiReachable
}

func requireIntegrationConfig(t *testing.T) Config {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	cfg := loadConfig()
	if !cfg.Enabled {
		t.Skip("integration tests disabled: set TRANSLAAS_API_KEY")
	}
	if !probeAPIReachable(cfg) {
		t.Skip("integration API not reachable at " + cfg.BaseURL)
	}
	return cfg
}

func newIntegrationClient(t *testing.T) client.Client {
	t.Helper()
	cfg := requireIntegrationConfig(t)
	return newClientWithOptions(t, cfg, client.Options{
		APIKey:           cfg.APIKey,
		BaseURL:          cfg.BaseURL,
		DefaultProjectID: cfg.DefaultProject,
		Timeout:          30 * time.Second,
	})
}

func newClientWithOptions(t *testing.T, cfg Config, opts client.Options) client.Client {
	t.Helper()
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.BaseURL == "" {
		opts.BaseURL = cfg.BaseURL
	}
	if opts.APIKey == "" {
		opts.APIKey = cfg.APIKey
	}
	c, err := client.New(opts)
	require.NoError(t, err)
	return c
}
