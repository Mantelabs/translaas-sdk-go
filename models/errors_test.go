package models

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testdataPath(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	path := filepath.Join(filepath.Dir(file), "..", "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read testdata %s: %v", name, err)
	}
	return data
}

func TestTranslaasError_FormatMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      TranslaasError
		fallback string
		want     string
	}{
		{
			name:     "code and message",
			err:      TranslaasError{Message: "Error message", Code: "ERROR_CODE"},
			fallback: "fallback",
			want:     "[ERROR_CODE] Error message",
		},
		{
			name:     "message only",
			err:      TranslaasError{Message: "Error message"},
			fallback: "fallback",
			want:     "Error message",
		},
		{
			name:     "fallback when message empty",
			err:      TranslaasError{Code: "ERROR_CODE"},
			fallback: "API request failed",
			want:     "[ERROR_CODE] API request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.FormatMessage(tt.fallback); got != tt.want {
				t.Fatalf("FormatMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTranslaasError_Golden(t *testing.T) {
	t.Parallel()
	data := testdataPath(t, "translaas_error_full.json")
	parsed, err := ParseTranslaasError(data)
	if err != nil {
		t.Fatalf("ParseTranslaasError: %v", err)
	}
	if parsed.Code != "ERROR_CODE" || parsed.Message != "Error message" {
		t.Fatalf("unexpected parsed error: %+v", parsed)
	}
}

func TestOfflineCacheMissError_Message(t *testing.T) {
	t.Parallel()

	tests := []struct {
		project, lang, group, entry, want string
	}{
		{
			project: "p1", lang: "en", group: "ui", entry: "save",
			want: "Translation entry 'save' in group 'ui' for project 'p1' and language 'en' was not found in the offline cache.",
		},
		{
			project: "p1", lang: "en", group: "ui",
			want: "Translation group 'ui' for project 'p1' and language 'en' was not found in the offline cache.",
		},
		{
			project: "p1", lang: "en",
			want: "Project 'p1' for language 'en' was not found in the offline cache.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want[:20], func(t *testing.T) {
			t.Parallel()
			err := NewOfflineCacheMissError(tt.project, tt.lang, tt.group, tt.entry)
			if err.Error() != tt.want {
				t.Fatalf("Error() = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRequestContext_Reset(t *testing.T) {
	t.Parallel()
	ctx := &RequestContext{ResponseETag: "etag", NotModified: true}
	ctx.Reset()
	if ctx.ResponseETag != "" || ctx.NotModified {
		t.Fatalf("Reset() did not clear response fields: %+v", ctx)
	}
}

func TestRequestContext_ResetNil(t *testing.T) {
	t.Parallel()
	var ctx *RequestContext
	ctx.Reset()
}
