package httpx

import (
	"net/url"
	"testing"
)

func TestAppendQueryValues_NilURL(t *testing.T) {
	t.Parallel()
	err := AppendQueryValues(nil, struct{}{})
	if err == nil {
		t.Fatal("expected error for nil url")
	}
}

func TestAppendQueryValues_NilParams(t *testing.T) {
	t.Parallel()
	u := &url.URL{Scheme: "https", Host: "api.test.com"}
	if err := AppendQueryValues(u, nil); err == nil {
		t.Fatal("expected error for nil params")
	}
	var nilReq *struct{}
	if err := AppendQueryValues(u, nilReq); err == nil {
		t.Fatal("expected error for nil pointer params")
	}
}

func TestAppendQueryValues_PointerToStruct(t *testing.T) {
	t.Parallel()
	type req struct {
		Name string `json:"name"`
	}
	u := &url.URL{Scheme: "https", Host: "api.test.com", Path: "/"}
	r := req{Name: "test"}
	if err := AppendQueryValues(u, &r); err != nil {
		t.Fatalf("AppendQueryValues() error = %v", err)
	}
	if QueryValues(u).Get("name") != "test" {
		t.Fatalf("name = %q", QueryValues(u).Get("name"))
	}
}

func TestAppendQueryValues_NumericAndSkippedFields(t *testing.T) {
	t.Parallel()
	type req struct {
		SkipNoTag string   `json:"-"`
		SkipEmpty string   `json:",omitempty"`
		Count     int      `json:"count"`
		Port      uint16   `json:"port"`
		Ratio     float32  `json:"ratio"`
		Unused    []string `json:"unused"`
	}
	u := &url.URL{Scheme: "https", Host: "api.test.com", Path: "/"}
	r := req{Count: 3, Port: 8080, Ratio: 1.5}
	if err := AppendQueryValues(u, r); err != nil {
		t.Fatalf("AppendQueryValues() error = %v", err)
	}
	values := QueryValues(u)
	if values.Get("count") != "3" {
		t.Fatalf("count = %q", values.Get("count"))
	}
	if values.Get("port") != "8080" {
		t.Fatalf("port = %q", values.Get("port"))
	}
	if values.Get("ratio") != "1.5" {
		t.Fatalf("ratio = %q", values.Get("ratio"))
	}
	if values.Get("unused") != "" {
		t.Fatal("unsupported field kind should be skipped")
	}
}

func TestMergeQueryParams_NilURL(t *testing.T) {
	t.Parallel()
	MergeQueryParams(nil, map[string]string{"a": "b"}) // should not panic
}

func TestQueryValues_NilURL(t *testing.T) {
	t.Parallel()
	if len(QueryValues(nil)) != 0 {
		t.Fatal("expected empty values for nil url")
	}
}

func TestQueryValues_InvalidQuery(t *testing.T) {
	t.Parallel()
	u := &url.URL{RawQuery: "%"}
	if len(QueryValues(u)) != 0 {
		t.Fatal("expected empty values for invalid query")
	}
}

func TestParseQueryParams_InvalidQuery(t *testing.T) {
	t.Parallel()
	params := parseQueryParams("%")
	if params != nil {
		t.Fatalf("expected nil params, got %v", params)
	}
}

func TestBuildURL_ParseError(t *testing.T) {
	t.Parallel()
	// url.Parse rarely fails; whitespace-trimmed invalid host without scheme hits scheme check first.
	_, err := BuildURL("://missing-scheme", "endpoint")
	if err == nil {
		t.Fatal("expected error")
	}
}
