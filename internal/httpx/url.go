package httpx

import (
	"fmt"
	"net/url"
	"strings"
)

// BuildURL joins baseURL and endpoint like .NET TranslaasClient.BuildEndpointUrl.
// Trims trailing slashes from baseURL and leading slashes from endpoint.
// Returns an error when baseURL is empty, not parseable, or lacks an http/https scheme.
func BuildURL(baseURL, endpoint string) (string, error) {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		return "", fmt.Errorf("httpx: base URL is required")
	}
	if strings.TrimSpace(endpoint) == "" {
		return "", fmt.Errorf("httpx: endpoint is required")
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("httpx: parse base URL: %w", err)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return "", fmt.Errorf("httpx: base URL must use http or https scheme")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("httpx: base URL must include a host")
	}

	trimmedBase := strings.TrimRight(base, "/")
	endpointPath := strings.TrimLeft(endpoint, "/")
	return trimmedBase + "/" + endpointPath, nil
}
