package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type goldenKeyCase struct {
	Name   string        `json:"name"`
	Method string        `json:"method"`
	Args   goldenKeyArgs `json:"args"`
	Want   string        `json:"want"`
}

type goldenKeyArgs struct {
	Group          string            `json:"group"`
	Entry          string            `json:"entry"`
	Lang           string            `json:"lang"`
	Number         *float64          `json:"number"`
	Parameters     map[string]string `json:"parameters"`
	Project        string            `json:"project"`
	Channel        string            `json:"channel"`
	Version        string            `json:"version"`
	Format         string            `json:"format"`
	IncludeContext *bool             `json:"includeContext"`
}

func TestKeyBuilder_GoldenVectors(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("testdata", "cache_keys.json"))
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	var cases []goldenKeyCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("unmarshal golden file: %v", err)
	}

	builder := KeyBuilder{}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			got := buildGoldenKey(t, builder, tc)
			if got != tc.Want {
				t.Fatalf("key = %q, want %q", got, tc.Want)
			}
		})
	}
}

func buildGoldenKey(t *testing.T, builder KeyBuilder, tc goldenKeyCase) string {
	t.Helper()
	switch tc.Method {
	case "entry":
		return builder.EntryKey(
			tc.Args.Group,
			tc.Args.Entry,
			tc.Args.Lang,
			tc.Args.Number,
			tc.Args.Parameters,
			tc.Args.Project,
			tc.Args.Channel,
			tc.Args.Version,
		)
	case "group":
		return builder.GroupKey(
			tc.Args.Project,
			tc.Args.Group,
			tc.Args.Lang,
			tc.Args.Format,
			tc.Args.Channel,
			tc.Args.Version,
			tc.Args.IncludeContext,
		)
	case "project":
		return builder.ProjectKey(
			tc.Args.Project,
			tc.Args.Lang,
			tc.Args.Format,
			tc.Args.Channel,
			tc.Args.Version,
			tc.Args.IncludeContext,
		)
	case "locales":
		return builder.LocalesKey(tc.Args.Project, tc.Args.Channel, tc.Args.Version)
	case "offline":
		return builder.OfflineKey(tc.Args.Project, tc.Args.Channel, tc.Args.Version, tc.Args.IncludeContext)
	default:
		t.Fatalf("unknown method %q", tc.Method)
		return ""
	}
}

func TestPackageLevelKeyHelpers(t *testing.T) {
	t.Parallel()
	if got := EntryKey("ui", "save", "en", nil, nil, "", "", ""); got != "entry:ui:save:en" {
		t.Fatalf("EntryKey = %q", got)
	}
}

func TestMode_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeNone, "None"},
		{ModeEntry, "Entry"},
		{ModeGroup, "Group"},
		{ModeProject, "Project"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Fatalf("Mode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestFormatCacheNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		n    float64
		want string
	}{
		{1, "1"},
		{1.0, "1"},
		{1.5, "1.5"},
		{0.0000001, "1e-07"},
		{1e15, "1e+15"},
	}
	for _, tt := range tests {
		if got := formatCacheNumber(tt.n); got != tt.want {
			t.Fatalf("formatCacheNumber(%v) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
