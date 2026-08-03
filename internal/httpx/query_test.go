package httpx

import (
	"net/url"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

func parseTestURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse(%q): %v", raw, err)
	}
	return u
}

func assertQueryEqual(t *testing.T, got, want *url.URL) {
	t.Helper()
	gotValues := QueryValues(got)
	wantValues := QueryValues(want)
	for key, wantVals := range wantValues {
		gotVals, ok := gotValues[key]
		if !ok {
			t.Fatalf("missing query key %q in %q", key, got.RawQuery)
		}
		if len(gotVals) != len(wantVals) {
			t.Fatalf("key %q: got %v want %v", key, gotVals, wantVals)
		}
		for i := range wantVals {
			if gotVals[i] != wantVals[i] {
				t.Fatalf("key %q: got %q want %q", key, gotVals[i], wantVals[i])
			}
		}
	}
	if len(gotValues) != len(wantValues) {
		t.Fatalf("query mismatch: got %v want %v", gotValues, wantValues)
	}
	if got.Scheme != want.Scheme || got.Host != want.Host || got.Path != want.Path {
		t.Fatalf("url parts mismatch: got %q want %q", got.String(), want.String())
	}
}

func TestAppendQueryValues_GetTranslationRequest(t *testing.T) {
	t.Parallel()

	u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text")
	req := models.GetTranslationRequest{
		Group: "ui",
		Entry: "button.save",
		Lang:  "en",
	}
	if err := AppendQueryValues(u, req); err != nil {
		t.Fatalf("AppendQueryValues() error = %v", err)
	}

	want := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text?group=ui&entry=button.save&lang=en")
	assertQueryEqual(t, u, want)
}

func TestAppendQueryValues_GetTranslationRequestWithNumber(t *testing.T) {
	t.Parallel()

	n := 5.0
	u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text")
	req := models.GetTranslationRequest{
		Group:  "ui",
		Entry:  "button.save",
		Lang:   "en",
		Number: &n,
	}
	if err := AppendQueryValues(u, req); err != nil {
		t.Fatalf("AppendQueryValues() error = %v", err)
	}
	if got := QueryValues(u).Get("n"); got != "5" {
		t.Fatalf("n = %q, want 5", got)
	}
}

func TestAppendQueryValues_DecimalNumber(t *testing.T) {
	t.Parallel()

	n := 1.31
	u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text")
	req := models.GetTranslationRequest{Number: &n}
	if err := AppendQueryValues(u, req); err != nil {
		t.Fatalf("AppendQueryValues() error = %v", err)
	}
	if got := QueryValues(u).Get("n"); got != "1.31" {
		t.Fatalf("n = %q, want 1.31", got)
	}
}

func TestAppendQueryValues_GetGroupTranslationsRequest(t *testing.T) {
	t.Parallel()

	trueVal := true
	falseVal := false
	tests := []struct {
		name string
		req  models.GetGroupTranslationsRequest
		want map[string]string
	}{
		{
			name: "include context true",
			req: models.GetGroupTranslationsRequest{
				Project:        "proj",
				Group:          "ui",
				Lang:           "en",
				IncludeContext: &trueVal,
			},
			want: map[string]string{
				"project":        "proj",
				"group":          "ui",
				"lang":           "en",
				"includeContext": "true",
			},
		},
		{
			name: "include context false",
			req: models.GetGroupTranslationsRequest{
				IncludeContext: &falseVal,
			},
			want: map[string]string{"includeContext": "false"},
		},
		{
			name: "include context nil omitted",
			req: models.GetGroupTranslationsRequest{
				Project: "proj",
			},
			want: map[string]string{"project": "proj"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/group")
			if err := AppendQueryValues(u, tt.req); err != nil {
				t.Fatalf("AppendQueryValues() error = %v", err)
			}
			got := QueryValues(u)
			if len(got) != len(tt.want) {
				t.Fatalf("query = %v, want %v", got, tt.want)
			}
			for key, value := range tt.want {
				if got.Get(key) != value {
					t.Fatalf("%s = %q, want %q", key, got.Get(key), value)
				}
			}
		})
	}
}

func TestAppendQueryValues_OmitsZeroValues(t *testing.T) {
	t.Parallel()

	u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text")
	req := models.GetTranslationRequest{}
	if err := AppendQueryValues(u, req); err != nil {
		t.Fatalf("AppendQueryValues() error = %v", err)
	}
	if u.RawQuery != "" {
		t.Fatalf("expected empty query, got %q", u.RawQuery)
	}
}

func TestAppendQueryValues_RejectsNonStruct(t *testing.T) {
	t.Parallel()

	u := parseTestURL(t, "https://api.test.com/")
	for _, params := range []any{"bad", 123, []string{"a"}} {
		if err := AppendQueryValues(u, params); err == nil {
			t.Fatalf("AppendQueryValues(%T) expected error", params)
		}
	}
}

func TestMergeQueryParams_CaseInsensitiveReplace(t *testing.T) {
	t.Parallel()

	u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text?group=ui&n=5")
	MergeQueryParams(u, map[string]string{"N": "5"})

	values := QueryValues(u)
	if values.Get("N") != "5" {
		t.Fatalf("N = %q", values.Get("N"))
	}
	if values.Get("n") != "" {
		t.Fatalf("lowercase n should be replaced, got %q", values.Get("n"))
	}
	if values.Get("group") != "ui" {
		t.Fatalf("group = %q", values.Get("group"))
	}
}

func TestMergeQueryParams_ParametersWin(t *testing.T) {
	t.Parallel()

	u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text?n=5")
	MergeQueryParams(u, map[string]string{"N": "10"})
	if got := QueryValues(u).Get("N"); got != "10" {
		t.Fatalf("N = %q, want 10", got)
	}
}

func TestMergeQueryParams_URLEncoding(t *testing.T) {
	t.Parallel()

	u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text")
	MergeQueryParams(u, map[string]string{
		"userName": "John Doe",
		"message":  "Hello & Welcome",
	})
	if got := u.RawQuery; got == "" {
		t.Fatal("expected encoded query")
	}
	decoded := QueryValues(u)
	if decoded.Get("userName") != "John Doe" {
		t.Fatalf("userName = %q", decoded.Get("userName"))
	}
	if decoded.Get("message") != "Hello & Welcome" {
		t.Fatalf("message = %q", decoded.Get("message"))
	}
}

func TestMergeQueryParams_SkipsEmptyValues(t *testing.T) {
	t.Parallel()

	u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text?group=ui")
	MergeQueryParams(u, map[string]string{"userName": "", "N": "5"})
	if QueryValues(u).Get("userName") != "" {
		t.Fatal("empty extra value should be skipped")
	}
	if QueryValues(u).Get("N") != "5" {
		t.Fatalf("N = %q", QueryValues(u).Get("N"))
	}
}

func TestInjectPluralN(t *testing.T) {
	t.Parallel()

	n := 5.0
	extra := map[string]string{}
	InjectPluralN(extra, &n)
	if extra["N"] != "5" {
		t.Fatalf("N = %q, want 5", extra["N"])
	}

	decimal := 1.31
	InjectPluralN(map[string]string{}, &decimal)

	existing := map[string]string{"n": "10"}
	InjectPluralN(existing, &n)
	if existing["N"] != "" {
		t.Fatal("should not inject when n key exists")
	}

	InjectPluralN(nil, &n)
	InjectPluralN(extra, nil)
}

func TestGetEntryFlow_NInjectionAndMerge(t *testing.T) {
	t.Parallel()

	rawURL, err := BuildURL("https://api.test.com", "sdk/v1/translations/text")
	if err != nil {
		t.Fatalf("BuildURL() error = %v", err)
	}
	u := parseTestURL(t, rawURL)

	n := 5.0
	req := models.GetTranslationRequest{
		Group:  "ui",
		Entry:  "button.save",
		Lang:   "en",
		Number: &n,
	}
	if err := AppendQueryValues(u, req); err != nil {
		t.Fatalf("AppendQueryValues() error = %v", err)
	}

	extra := map[string]string{"userName": "John"}
	InjectPluralN(extra, &n)
	MergeQueryParams(u, extra)

	values := QueryValues(u)
	if values.Get("N") != "5" {
		t.Fatalf("N = %q, want 5", values.Get("N"))
	}
	if values.Get("n") != "" {
		t.Fatalf("lowercase n should be replaced by N")
	}
	if values.Get("userName") != "John" {
		t.Fatalf("userName = %q", values.Get("userName"))
	}
}

func TestQueryOrderIndependence(t *testing.T) {
	t.Parallel()

	build := func() *url.URL {
		u := parseTestURL(t, "https://api.test.com/sdk/v1/translations/text")
		req := models.GetTranslationRequest{Group: "ui", Entry: "button.save", Lang: "en"}
		_ = AppendQueryValues(u, req)
		MergeQueryParams(u, map[string]string{"userName": "John", "N": "5"})
		return u
	}

	first := build()
	second := build()
	assertQueryEqual(t, first, second)
}
