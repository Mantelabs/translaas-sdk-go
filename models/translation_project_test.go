package models

import (
	"encoding/json"
	"testing"
)

func TestTranslationProject_FlatGroups(t *testing.T) {
	t.Parallel()
	var project TranslationProject
	if err := json.Unmarshal(testdataPath(t, "translation_project_flat.json"), &project); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ui, err := project.GetGroup("ui")
	if err != nil {
		t.Fatalf("GetGroup(ui): %v", err)
	}
	if ui == nil {
		t.Fatal("expected ui group")
	}
	value, ok := ui.GetValue("button.save")
	if !ok || value != "Save" {
		t.Fatalf("GetValue(button.save) = (%q, %v)", value, ok)
	}

	common, err := project.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup(common): %v", err)
	}
	value, ok = common.GetValue("welcome")
	if !ok || value != "Welcome" {
		t.Fatalf("GetValue(welcome) = (%q, %v)", value, ok)
	}
}

func TestTranslationProject_APIGroupShape(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"ui": {
			"Project": "p",
			"Lang": "en",
			"Entries": { "save": "Save" }
		}
	}`)
	var project TranslationProject
	if err := json.Unmarshal(raw, &project); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	group, err := project.GetGroup("ui")
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	value, ok := group.GetValue("save")
	if !ok || value != "Save" {
		t.Fatalf("GetValue(save) = (%q, %v)", value, ok)
	}
}

func TestTranslationProject_GroupEntryContext(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"groupEntryContext": { "ui": { "note": "ctx" } },
		"ui": { "hello": "Hello" }
	}`)
	var project TranslationProject
	if err := json.Unmarshal(raw, &project); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(project.GroupEntryContext) != 1 {
		t.Fatalf("expected groupEntryContext, got %d", len(project.GroupEntryContext))
	}
	if len(project.Groups) != 1 {
		t.Fatalf("expected one group, got %d", len(project.Groups))
	}

	encoded, err := json.Marshal(&project)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var again TranslationProject
	if err := json.Unmarshal(encoded, &again); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if len(again.GroupEntryContext) != 1 || len(again.Groups) != 1 {
		t.Fatalf("round-trip mismatch: %+v", again)
	}
}

func TestProjectLocales_Unmarshal(t *testing.T) {
	t.Parallel()
	var locales ProjectLocales
	if err := json.Unmarshal(testdataPath(t, "project_locales.json"), &locales); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(locales.Locales) != 4 || locales.Locales[0] != "en" {
		t.Fatalf("unexpected locales: %+v", locales)
	}
}
