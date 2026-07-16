package models

import (
	"encoding/json"
	"testing"
)

func TestTranslationGroup_FlatSimple(t *testing.T) {
	t.Parallel()
	var group TranslationGroup
	if err := json.Unmarshal(testdataPath(t, "translation_group_flat_simple.json"), &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	value, ok := group.GetValue("button.save")
	if !ok || value != "Save" {
		t.Fatalf("GetValue(button.save) = (%q, %v)", value, ok)
	}
	if group.HasPluralForms("button.save") {
		t.Fatal("button.save should not have plural forms")
	}
}

func TestTranslationGroup_Empty(t *testing.T) {
	t.Parallel()
	var group TranslationGroup
	if err := json.Unmarshal(testdataPath(t, "translation_group_empty.json"), &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(group.Entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(group.Entries))
	}
}

func TestTranslationGroup_PluralEN(t *testing.T) {
	t.Parallel()
	var group TranslationGroup
	if err := json.Unmarshal(testdataPath(t, "translation_group_plural_en.json"), &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !group.HasPluralForms("simple-count") {
		t.Fatal("expected plural forms for simple-count")
	}
	if _, ok := group.GetValue("simple-count"); ok {
		t.Fatal("GetValue should fail for plural entry")
	}

	forms, ok := group.GetPluralForms("simple-count")
	if !ok || forms[PluralOne] != "There is 1 record" || forms[PluralOther] != "There are {0} records" {
		t.Fatalf("unexpected plural forms: %+v, ok=%v", forms, ok)
	}

	one, ok := group.GetPluralForm("simple-count", PluralOne)
	if !ok || one != "There is 1 record" {
		t.Fatalf("GetPluralForm(one) = (%q, %v)", one, ok)
	}
}

func TestTranslationGroup_PluralAR(t *testing.T) {
	t.Parallel()
	var group TranslationGroup
	if err := json.Unmarshal(testdataPath(t, "translation_group_plural_ar.json"), &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	categories := []PluralCategory{PluralZero, PluralOne, PluralTwo, PluralFew, PluralMany, PluralOther}
	for _, cat := range categories {
		if _, ok := group.GetPluralForm("item", cat); !ok {
			t.Fatalf("missing plural category %v", cat)
		}
	}
}

func TestTranslationGroup_FullAPI(t *testing.T) {
	t.Parallel()
	var group TranslationGroup
	if err := json.Unmarshal(testdataPath(t, "translation_group_full_api.json"), &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if group.Project != "my-project" || group.Lang != "en" {
		t.Fatalf("unexpected metadata: project=%q lang=%q", group.Project, group.Lang)
	}
	if group.GeneratedAt == nil {
		t.Fatal("expected GeneratedAt")
	}
	if len(group.EntryContext) != 1 {
		t.Fatalf("expected entryContext, got %d", len(group.EntryContext))
	}

	value, ok := group.GetValue("welcome")
	if !ok || value != "Welcome" {
		t.Fatalf("GetValue(welcome) = (%q, %v)", value, ok)
	}

	roundTrip, err := json.Marshal(&group)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var again TranslationGroup
	if err := json.Unmarshal(roundTrip, &again); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if again.Project != group.Project || len(again.Entries) != len(group.Entries) {
		t.Fatalf("round-trip mismatch: %+v", again)
	}
}

func TestTranslationGroup_Mixed(t *testing.T) {
	t.Parallel()
	var group TranslationGroup
	if err := json.Unmarshal(testdataPath(t, "translation_group_mixed.json"), &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !group.HasPluralForms("simple-count") {
		t.Fatal("expected plural entry")
	}
	value, ok := group.GetValue("button.save")
	if !ok || value != "Save" {
		t.Fatalf("GetValue(button.save) = (%q, %v)", value, ok)
	}
}
