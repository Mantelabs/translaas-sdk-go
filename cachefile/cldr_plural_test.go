package cachefile

import "testing"

func TestLooksLikeBCP47(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{name: "en", tag: "en", want: true},
		{name: "pt-PT", tag: "pt-PT", want: true},
		{name: "zh-Hans-CN", tag: "zh-Hans-CN", want: true},
		{name: "underscore-raw", tag: "ar_EG", want: false},
		{name: "empty", tag: "", want: false},
		{name: "too-short", tag: "x", want: false},
		{name: "free-text", tag: "not a locale!!", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := looksLikeBCP47(tt.tag); got != tt.want {
				t.Fatalf("looksLikeBCP47(%q) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}
}
