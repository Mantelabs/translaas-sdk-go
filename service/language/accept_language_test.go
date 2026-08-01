package language_test

import (
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/service/language"
)

func TestParseAcceptLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "primary with region", header: "en-US,en;q=0.9", want: "en"},
		{name: "quality suffix only", header: "fr;q=0.8", want: "fr"},
		{name: "simple code", header: "de", want: "de"},
		{name: "empty", header: "", want: ""},
		{name: "invalid", header: "invalid", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := language.ParseAcceptLanguage(tt.header); got != tt.want {
				t.Fatalf("ParseAcceptLanguage(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestNormalizeLanguageCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		lang string
		want string
	}{
		{lang: "en", want: "en"},
		{lang: "EN-US", want: "en"},
		{lang: "fr_FR", want: "fr"},
		{lang: "  pt  ", want: "pt"},
		{lang: "invalid", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			t.Parallel()
			if got := language.NormalizeLanguageCode(tt.lang); got != tt.want {
				t.Fatalf("NormalizeLanguageCode(%q) = %q, want %q", tt.lang, got, tt.want)
			}
		})
	}
}
