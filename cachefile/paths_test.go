package cachefile_test

import (
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/cachefile"
)

func TestSanitizePathSegment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "simple", input: "my-project", want: "my-project"},
		{name: "replace slash", input: "my/project", want: "my_project"},
		{name: "replace colon", input: "my:project", want: "my_project"},
		{name: "empty", input: "   ", wantErr: true},
		{name: "traversal", input: "..", wantErr: true},
		{name: "embedded traversal", input: "foo..bar", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := cachefile.SanitizePathSegment(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
