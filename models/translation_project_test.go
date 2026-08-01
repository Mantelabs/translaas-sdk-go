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

func TestTranslationProject_APIMetadataAndGroups(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"Project": "translaas-sdk-samples",
		"Lang": "en",
		"Version": 245734752,
		"GeneratedAt": "2026-01-15T12:00:00Z",
		"groupEntryContext": { "common": { "welcome.message": { "note": "ctx" } } },
		"common": { "welcome.message": "Welcome" },
		"messages": { "item": "1 item" }
	}`)
	var project TranslationProject
	if err := json.Unmarshal(raw, &project); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if project.Project != "translaas-sdk-samples" || project.Lang != "en" {
		t.Fatalf("metadata: Project=%q Lang=%q", project.Project, project.Lang)
	}
	if len(project.Version) == 0 || project.GeneratedAt == nil {
		t.Fatal("expected Version and GeneratedAt metadata")
	}
	if len(project.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d: %v", len(project.Groups), project.Groups)
	}
	for _, key := range []string{"Project", "Lang", "Version", "GeneratedAt", "groupEntryContext"} {
		if _, ok := project.Groups[key]; ok {
			t.Fatalf("metadata key %q should not be in Groups", key)
		}
	}

	common, err := project.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup(common): %v", err)
	}
	value, ok := common.GetValue("welcome.message")
	if !ok || value != "Welcome" {
		t.Fatalf("GetValue(welcome.message) = (%q, %v)", value, ok)
	}

	versionGroup, err := project.GetGroup("Version")
	if err != nil {
		t.Fatalf("GetGroup(Version): %v", err)
	}
	if versionGroup != nil {
		t.Fatal("GetGroup(Version) should be nil for scalar metadata")
	}
}

func TestTranslationProject_GetGroupIgnoresNonObjectValues(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"Version": 123,
		"common": { "welcome.message": "Welcome" }
	}`)
	var project TranslationProject
	if err := json.Unmarshal(raw, &project); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := project.Groups["Version"]; ok {
		t.Fatal("Version should not be stored in Groups")
	}
	versionGroup, err := project.GetGroup("Version")
	if err != nil {
		t.Fatalf("GetGroup(Version): %v", err)
	}
	if versionGroup != nil {
		t.Fatal("GetGroup(Version) should be nil for scalar metadata")
	}

	common, err := project.GetGroup("common")
	if err != nil {
		t.Fatalf("GetGroup(common): %v", err)
	}
	value, ok := common.GetValue("welcome.message")
	if !ok || value != "Welcome" {
		t.Fatalf("GetValue(welcome.message) = (%q, %v)", value, ok)
	}
}

func TestTranslationProject_AllGroupsAccessibleWithMetadata(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"Project": "translaas-sdk-samples",
		"Lang": "en",
		"Version": 1,
		"common": { "welcome.message": "Welcome" },
		"messages": { "item.one": "1 item", "item.other": "{count} items" }
	}`)
	var project TranslationProject
	if err := json.Unmarshal(raw, &project); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for groupName := range project.Groups {
		group, err := project.GetGroup(groupName)
		if err != nil {
			t.Fatalf("GetGroup(%q): %v", groupName, err)
		}
		if group == nil {
			t.Fatalf("GetGroup(%q) returned nil", groupName)
		}
		if len(group.Entries) == 0 {
			t.Fatalf("group %q has no entries", groupName)
		}
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
