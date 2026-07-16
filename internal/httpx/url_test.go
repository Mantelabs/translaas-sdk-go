package httpx

import (
	"testing"
)

func TestBuildURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		baseURL   string
		endpoint  string
		want      string
		wantError bool
	}{
		{
			name:     "combine base and endpoint",
			baseURL:  "https://api.test.com",
			endpoint: "sdk/v1/translations/text",
			want:     "https://api.test.com/sdk/v1/translations/text",
		},
		{
			name:     "trailing slash on base",
			baseURL:  "https://api.test.com/",
			endpoint: "sdk/v1/translations/text",
			want:     "https://api.test.com/sdk/v1/translations/text",
		},
		{
			name:     "leading slash on endpoint",
			baseURL:  "https://api.test.com",
			endpoint: "/sdk/v1/translations/text",
			want:     "https://api.test.com/sdk/v1/translations/text",
		},
		{
			name:     "both slashes trimmed",
			baseURL:  "https://api.test.com/",
			endpoint: "/sdk/v1/translations/text",
			want:     "https://api.test.com/sdk/v1/translations/text",
		},
		{
			name:     "base with path prefix",
			baseURL:  "https://api.test.com/custom",
			endpoint: "sdk/v1/translations/text",
			want:     "https://api.test.com/custom/sdk/v1/translations/text",
		},
		{
			name:     "validate api key path",
			baseURL:  "https://api.test.com",
			endpoint: "api/v1/api-keys/validate",
			want:     "https://api.test.com/api/v1/api-keys/validate",
		},
		{
			name:      "empty base",
			baseURL:   "   ",
			endpoint:  "sdk/v1/translations/text",
			wantError: true,
		},
		{
			name:      "invalid base",
			baseURL:   "not-a-url",
			endpoint:  "sdk/v1/translations/text",
			wantError: true,
		},
		{
			name:      "ftp scheme rejected",
			baseURL:   "ftp://files.test.com",
			endpoint:  "sdk/v1/translations/text",
			wantError: true,
		},
		{
			name:      "empty endpoint",
			baseURL:   "https://api.test.com",
			endpoint:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := BuildURL(tt.baseURL, tt.endpoint)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("BuildURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildURL_TrimsWhitespaceBase(t *testing.T) {
	t.Parallel()
	got, err := BuildURL("  https://api.test.com  ", "sdk/v1/translations/text")
	if err != nil {
		t.Fatalf("BuildURL() error = %v", err)
	}
	if got != "https://api.test.com/sdk/v1/translations/text" {
		t.Fatalf("got %q", got)
	}
}
