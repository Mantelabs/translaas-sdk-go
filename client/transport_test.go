package client

import "testing"

func TestParseContentDisposition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		header string
		want   string
	}{
		{`attachment; filename="bundle.zip"`, "bundle.zip"},
		{`attachment; filename*=UTF-8''my%20bundle.zip`, "my bundle.zip"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := parseContentDisposition(tt.header); got != tt.want {
			t.Fatalf("parseContentDisposition(%q) = %q, want %q", tt.header, got, tt.want)
		}
	}
}
