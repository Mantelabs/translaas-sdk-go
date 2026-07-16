package validate

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/acuencadev/translaas-sdk-go/models"
)

var urlSchemePattern = regexp.MustCompile(`(?i)^https?://`)

// ClientOptions holds fields validated for client construction.
type ClientOptions struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// Client validates live client configuration, mirroring TranslaasClientOptions.Validate.
func Client(opts ClientOptions) error {
	if strings.TrimSpace(opts.APIKey) == "" {
		return &models.ConfigurationError{
			Message: "ApiKey is required and cannot be null or empty.",
		}
	}
	baseURL := strings.TrimSpace(opts.BaseURL)
	if baseURL == "" {
		return &models.ConfigurationError{
			Message: "BaseUrl is required and cannot be null or empty.",
		}
	}
	if !urlSchemePattern.MatchString(baseURL) {
		return &models.ConfigurationError{
			Message: fmt.Sprintf("BaseUrl must be a valid HTTP or HTTPS URL. Provided value: %s", opts.BaseURL),
		}
	}
	if opts.Timeout < 0 {
		return &models.ConfigurationError{
			Message: fmt.Sprintf("Timeout must be greater than zero. Provided value: %s", opts.Timeout),
		}
	}
	return nil
}
