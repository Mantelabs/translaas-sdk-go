package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestResolveDefaultProjectID_Configured(t *testing.T) {
	t.Parallel()
	got, err := ResolveDefaultProjectID("  my-project ", &ValidateAPIKeyResponse{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my-project" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveDefaultProjectID_FromValidate(t *testing.T) {
	t.Parallel()
	var validate ValidateAPIKeyResponse
	if err := json.Unmarshal(testdataPath(t, "validate_api_key_tenant.json"), &validate); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	got, err := ResolveDefaultProjectID("", &validate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "proj-a" {
		t.Fatalf("got %q, want proj-a", got)
	}
}

func TestResolveDefaultProjectID_MissingProjectIDs(t *testing.T) {
	t.Parallel()
	_, err := ResolveDefaultProjectID("", &ValidateAPIKeyResponse{IsValid: true})
	var cfg *ConfigurationError
	if !errors.As(err, &cfg) {
		t.Fatalf("expected ConfigurationError, got %T: %v", err, err)
	}
}

func TestReadJSONULID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		raw  json.RawMessage
		want string
	}{
		{json.RawMessage(`"01ARZ3NDEKTSV4RRFFQ69G5FAV"`), "01ARZ3NDEKTSV4RRFFQ69G5FAV"},
		{json.RawMessage(`null`), ""},
		{json.RawMessage(nil), ""},
	}
	for _, tt := range tests {
		if got := ReadJSONULID(tt.raw); got != tt.want {
			t.Fatalf("ReadJSONULID(%s) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestRequests_JSONTags(t *testing.T) {
	t.Parallel()
	n := 5.0
	req := GetTranslationRequest{
		Group: "ui", Entry: "button.save", Lang: "en", Number: &n,
		Project: "my-project", Channel: "stable", Version: "42",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal map: %v", err)
	}
	for _, key := range []string{"group", "entry", "lang", "n", "project", "channel", "v"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing key %q in %v", key, decoded)
		}
	}
}

func TestReportMissingKeysRequest_JSON(t *testing.T) {
	t.Parallel()
	req := ReportMissingKeysRequest{
		Keys: []ReportMissingKeyItem{{
			GroupKey: "ui", EntryKey: "missing.key", LanguageIsoCode: "en",
		}},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded ReportMissingKeysRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Keys) != 1 || decoded.Keys[0].GroupKey != "ui" {
		t.Fatalf("unexpected decoded request: %+v", decoded)
	}
}
