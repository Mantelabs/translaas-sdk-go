package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestErrorTypes_Messages(t *testing.T) {
	t.Parallel()

	apiErr := &APIError{StatusCode: 401, Message: "unauthorized"}
	if apiErr.Error() != "unauthorized" {
		t.Fatalf("APIError message = %q", apiErr.Error())
	}
	apiFallback := &APIError{StatusCode: 500}
	if apiFallback.Error() != "translaas API error: status 500" {
		t.Fatalf("APIError fallback = %q", apiFallback.Error())
	}

	cfg := &ConfigurationError{Message: "bad config"}
	if cfg.Error() != "bad config" {
		t.Fatalf("ConfigurationError = %q", cfg.Error())
	}

	root := errors.New("disk")
	offline := &OfflineCacheError{Message: "io", Cause: root}
	if offline.Error() != "io" || offline.Unwrap() != root {
		t.Fatalf("OfflineCacheError unwrap failed")
	}

	miss := &OfflineCacheMissError{Message: "miss", Cause: root}
	if miss.Error() != "miss" || miss.Unwrap() != root {
		t.Fatalf("OfflineCacheMissError unwrap failed")
	}
}

func TestParseTranslaasError_EmptyBody(t *testing.T) {
	t.Parallel()
	parsed, err := ParseTranslaasError(nil)
	if err != nil || parsed != nil {
		t.Fatalf("expected nil,nil got parsed=%v err=%v", parsed, err)
	}
}

func TestParseTranslaasError_InvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := ParseTranslaasError([]byte("not-json"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveDefaultProjectID_Unresolvable(t *testing.T) {
	t.Parallel()
	validate := &ValidateAPIKeyResponse{
		IsValid:    true,
		ProjectIDs: []string{"   "},
	}
	_, err := ResolveDefaultProjectID("", validate)
	var cfg *ConfigurationError
	if !errors.As(err, &cfg) {
		t.Fatalf("expected ConfigurationError, got %v", err)
	}
}

func TestTranslationProject_GetGroupMissing(t *testing.T) {
	t.Parallel()
	project := TranslationProject{Groups: map[string]json.RawMessage{}}
	group, err := project.GetGroup("missing")
	if err != nil || group != nil {
		t.Fatalf("GetGroup missing = (%v, %v)", group, err)
	}
}

func TestTranslationProject_GetGroupInvalidJSON(t *testing.T) {
	t.Parallel()
	project := TranslationProject{
		Groups: map[string]json.RawMessage{
			"bad": json.RawMessage(`not-json`),
		},
	}
	_, err := project.GetGroup("bad")
	if err == nil {
		t.Fatal("expected error for invalid group json")
	}
}

func TestTranslationGroup_InvalidJSON(t *testing.T) {
	t.Parallel()
	var group TranslationGroup
	if err := json.Unmarshal([]byte(`"nope"`), &group); err == nil {
		t.Fatal("expected error for non-object json")
	}
}

func TestTranslationGroup_GetValueMissing(t *testing.T) {
	t.Parallel()
	group := TranslationGroup{Entries: map[string]json.RawMessage{}}
	if _, ok := group.GetValue("missing"); ok {
		t.Fatal("expected missing value")
	}
	if forms, ok := group.GetPluralForms("missing"); ok || forms != nil {
		t.Fatal("expected no plural forms")
	}
}

func TestPluralCategory_String(t *testing.T) {
	t.Parallel()
	if PluralOne.String() != "One" || PluralOther.String() != "Other" {
		t.Fatalf("unexpected string values")
	}
	if PluralCategory(99).String() != "Other" {
		t.Fatal("default String() should be Other")
	}
}

func TestGetGroupTranslationsRequest_IncludeContext(t *testing.T) {
	t.Parallel()
	include := true
	req := GetGroupTranslationsRequest{IncludeContext: &include}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !jsonContains(t, data, "includeContext") {
		t.Fatalf("missing includeContext in %s", data)
	}
}

func jsonContains(t *testing.T, data []byte, key string) bool {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	_, ok := m[key]
	return ok
}

func TestOfflineCacheDownloadResult_Fields(t *testing.T) {
	t.Parallel()
	result := OfflineCacheDownloadResult{
		NotModified: true, ETag: "W/\"abc\"", SuggestedFileName: "bundle.zip",
	}
	if !result.NotModified || result.ETag == "" || result.Content != nil {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestValidateAPIKeyResponse_Unmarshal(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"isValid": true,
		"projectId": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		"integrationName": "my-app",
		"authenticatedAt": "2026-07-16T08:00:00Z"
	}`)
	var resp ValidateAPIKeyResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.IsValid || ReadJSONULID(resp.ProjectID) == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestTranslationProject_UnmarshalInvalidGroupEntryContext(t *testing.T) {
	t.Parallel()
	var project TranslationProject
	err := json.Unmarshal([]byte(`{"groupEntryContext":"bad"}`), &project)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestTranslationGroup_EntryContextPascalCase(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"welcome": "Hi",
		"EntryContext": { "welcome": { "note": "x" } }
	}`)
	var group TranslationGroup
	if err := json.Unmarshal(raw, &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(group.EntryContext) != 1 {
		t.Fatalf("expected EntryContext, got %d", len(group.EntryContext))
	}
}

func TestGetPluralForms_InvalidObject(t *testing.T) {
	t.Parallel()
	group := TranslationGroup{
		Entries: map[string]json.RawMessage{
			"broken": json.RawMessage(`{"one":1}`),
		},
	}
	if forms, ok := group.GetPluralForms("broken"); ok {
		t.Fatalf("expected false for non-string plural values, got %+v", forms)
	}
}

func TestTranslationProject_GetGroupEmptyRaw(t *testing.T) {
	t.Parallel()
	project := TranslationProject{
		Groups: map[string]json.RawMessage{
			"empty": json.RawMessage(`{}`),
		},
	}
	group, err := project.GetGroup("empty")
	if err != nil || group == nil || len(group.Entries) != 0 {
		t.Fatalf("GetGroup(empty) = (%+v, %v)", group, err)
	}
}

func TestTranslationProject_GetGroupNonObjectFallback(t *testing.T) {
	t.Parallel()
	project := TranslationProject{
		Groups: map[string]json.RawMessage{
			"str": json.RawMessage(`"plain"`),
		},
	}
	_, err := project.GetGroup("str")
	if err == nil {
		t.Fatal("expected error unmarshaling string group payload")
	}
}

func TestTranslationGroup_FullAPIInvalidEntries(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Entries":"not-an-object"}`)
	var group TranslationGroup
	if err := json.Unmarshal(raw, &group); err == nil {
		t.Fatal("expected error for invalid Entries payload")
	}
}

func TestTranslationGroup_GeneratedAtInvalidIgnored(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Entries":{},"GeneratedAt":"not-a-date"}`)
	var group TranslationGroup
	if err := json.Unmarshal(raw, &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if group.GeneratedAt != nil {
		t.Fatal("invalid GeneratedAt should be ignored")
	}
}

func TestTranslationGroup_MarshalMinimal(t *testing.T) {
	t.Parallel()
	group := TranslationGroup{Entries: map[string]json.RawMessage{
		"k": json.RawMessage(`"v"`),
	}}
	data, err := json.Marshal(group)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !jsonContains(t, data, "Entries") {
		t.Fatalf("missing Entries in %s", data)
	}
}

func TestGetPluralForms_UnknownKeysSkipped(t *testing.T) {
	t.Parallel()
	group := TranslationGroup{
		Entries: map[string]json.RawMessage{
			"mix": json.RawMessage(`{"one":"1","invalid":"x"}`),
		},
	}
	forms, ok := group.GetPluralForms("mix")
	if !ok || forms[PluralOne] != "1" || len(forms) != 1 {
		t.Fatalf("unexpected forms: %+v ok=%v", forms, ok)
	}
}
