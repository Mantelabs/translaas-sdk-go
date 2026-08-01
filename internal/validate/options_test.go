package validate

import (
	"errors"
	"testing"
	"time"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

func TestClient_Valid(t *testing.T) {
	t.Parallel()
	err := Client(ClientOptions{
		APIKey:  "key",
		BaseURL: "https://api.test.com",
		Timeout: 0,
	})
	if err != nil {
		t.Fatalf("Client() error = %v", err)
	}
}

func TestClient_MissingAPIKey(t *testing.T) {
	t.Parallel()
	err := Client(ClientOptions{BaseURL: "https://api.test.com"})
	var cfgErr *models.ConfigurationError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected ConfigurationError, got %T", err)
	}
}

func TestClient_InvalidBaseURL(t *testing.T) {
	t.Parallel()
	err := Client(ClientOptions{APIKey: "key", BaseURL: "not-a-url"})
	var cfgErr *models.ConfigurationError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected ConfigurationError, got %T", err)
	}
}

func TestClient_NegativeTimeout(t *testing.T) {
	t.Parallel()
	err := Client(ClientOptions{
		APIKey:  "key",
		BaseURL: "https://api.test.com",
		Timeout: -time.Second,
	})
	var cfgErr *models.ConfigurationError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected ConfigurationError, got %T", err)
	}
}
