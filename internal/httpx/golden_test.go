package httpx

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

type goldenURLCase struct {
	Name        string            `json:"name"`
	Source      string            `json:"source"`
	BaseURL     string            `json:"baseURL"`
	Endpoint    string            `json:"endpoint"`
	RequestType string            `json:"requestType"`
	Request     json.RawMessage   `json:"request"`
	Extra       map[string]string `json:"extra"`
	InjectN     *float64          `json:"injectN"`
	WantPath    string            `json:"wantPath"`
	WantQuery   map[string]string `json:"wantQuery"`
}

func TestGoldenURLs(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(goldenTestdataPath("urls.json"))
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	var cases []goldenURLCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("unmarshal golden cases: %v", err)
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			rawURL, err := BuildURL(tc.BaseURL, tc.Endpoint)
			if err != nil {
				t.Fatalf("BuildURL() error = %v", err)
			}
			u, err := url.Parse(rawURL)
			if err != nil {
				t.Fatalf("url.Parse() error = %v", err)
			}

			if err := appendGoldenRequest(u, tc.RequestType, tc.Request); err != nil {
				t.Fatalf("append request: %v", err)
			}

			extra := tc.Extra
			if extra == nil && tc.InjectN != nil {
				extra = map[string]string{}
			}
			if tc.InjectN != nil {
				InjectPluralN(extra, tc.InjectN)
			}
			if len(extra) > 0 {
				MergeQueryParams(u, extra)
			}

			if u.Path != tc.WantPath {
				t.Fatalf("path = %q, want %q", u.Path, tc.WantPath)
			}

			got := QueryValues(u)
			if len(got) != len(tc.WantQuery) {
				t.Fatalf("query = %v, want %v", got, tc.WantQuery)
			}
			for key, want := range tc.WantQuery {
				if got.Get(key) != want {
					t.Fatalf("%s = %q, want %q", key, got.Get(key), want)
				}
			}
		})
	}
}

func appendGoldenRequest(u *url.URL, requestType string, raw json.RawMessage) error {
	switch requestType {
	case "none":
		return nil
	case "translation":
		var req models.GetTranslationRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return err
		}
		return AppendQueryValues(u, req)
	case "group":
		var req models.GetGroupTranslationsRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return err
		}
		return AppendQueryValues(u, req)
	default:
		return nil
	}
}

func goldenTestdataPath(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata", name)
}
