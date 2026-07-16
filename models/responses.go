package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TranslationGroup represents a translation group with one or more entries.
type TranslationGroup struct {
	Project      string                     `json:"Project,omitempty"`
	Lang         string                     `json:"Lang,omitempty"`
	Version      json.RawMessage            `json:"Version,omitempty"`
	GeneratedAt  *time.Time                 `json:"GeneratedAt,omitempty"`
	Entries      map[string]json.RawMessage `json:"Entries"`
	EntryContext map[string]json.RawMessage `json:"entryContext,omitempty"`
}

// UnmarshalJSON supports full API shape (with Entries) and flat offline/cache shape.
func (g *TranslationGroup) UnmarshalJSON(data []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}

	*g = TranslationGroup{Entries: make(map[string]json.RawMessage)}

	if entriesRaw, ok := root["Entries"]; ok {
		g.Project = rawString(root["Project"])
		g.Lang = rawString(root["Lang"])
		if v, ok := root["Version"]; ok {
			g.Version = append(json.RawMessage(nil), v...)
		}
		if ga, ok := root["GeneratedAt"]; ok {
			var t time.Time
			if err := json.Unmarshal(ga, &t); err == nil {
				g.GeneratedAt = &t
			}
		}
		if err := json.Unmarshal(entriesRaw, &g.Entries); err != nil {
			return err
		}
		readEntryContext(root, g)
		return nil
	}

	g.Entries = root
	readEntryContext(root, g)
	return nil
}

// MarshalJSON writes the group using the .NET SDK write shape.
func (g TranslationGroup) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte('{')

	first := true
	write := func(key, value string) {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		keyJSON, _ := json.Marshal(key)
		buf.Write(keyJSON)
		buf.WriteByte(':')
		buf.WriteString(value)
	}

	if g.Project != "" {
		v, _ := json.Marshal(g.Project)
		write("Project", string(v))
	}
	if g.Lang != "" {
		v, _ := json.Marshal(g.Lang)
		write("Lang", string(v))
	}
	if len(g.Version) > 0 {
		write("Version", string(g.Version))
	}
	if g.GeneratedAt != nil {
		v, _ := json.Marshal(g.GeneratedAt)
		write("GeneratedAt", string(v))
	}
	if len(g.EntryContext) > 0 {
		v, err := json.Marshal(g.EntryContext)
		if err != nil {
			return nil, err
		}
		write("entryContext", string(v))
	}

	entriesJSON, err := json.Marshal(g.Entries)
	if err != nil {
		return nil, err
	}
	if !first {
		buf.WriteByte(',')
	}
	buf.WriteString(`"Entries":`)
	buf.Write(entriesJSON)
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func readEntryContext(root map[string]json.RawMessage, g *TranslationGroup) {
	for _, name := range []string{"entryContext", "EntryContext"} {
		if raw, ok := root[name]; ok {
			var ctx map[string]json.RawMessage
			if err := json.Unmarshal(raw, &ctx); err == nil {
				g.EntryContext = ctx
				return
			}
		}
	}
}

func rawString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return strings.Trim(string(raw), `"`)
	}
	return s
}

// GetValue returns a simple string entry value, or ("", false) when missing or plural.
func (g *TranslationGroup) GetValue(key string) (string, bool) {
	raw, ok := g.Entries[key]
	if !ok {
		return "", false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, true
	}
	return "", false
}

// HasPluralForms reports whether an entry is a plural object map.
func (g *TranslationGroup) HasPluralForms(key string) bool {
	raw, ok := g.Entries[key]
	if !ok {
		return false
	}
	return len(raw) > 0 && raw[0] == '{'
}

// GetPluralForms returns plural category values for an entry.
func (g *TranslationGroup) GetPluralForms(key string) (map[PluralCategory]string, bool) {
	raw, ok := g.Entries[key]
	if !ok || len(raw) == 0 || raw[0] != '{' {
		return nil, false
	}

	var obj map[string]string
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, false
	}

	out := make(map[PluralCategory]string)
	for name, value := range obj {
		category, ok := parsePluralCategory(name)
		if !ok {
			continue
		}
		out[category] = value
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

// GetPluralForm returns one plural category value for an entry.
func (g *TranslationGroup) GetPluralForm(key string, category PluralCategory) (string, bool) {
	forms, ok := g.GetPluralForms(key)
	if !ok {
		return "", false
	}
	value, ok := forms[category]
	return value, ok
}

// TranslationProject represents a project payload with dynamic group keys.
type TranslationProject struct {
	GroupEntryContext map[string]json.RawMessage `json:"groupEntryContext,omitempty"`
	Groups            map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON separates groupEntryContext from extension group keys.
func (p *TranslationProject) UnmarshalJSON(data []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}

	p.Groups = make(map[string]json.RawMessage)
	for key, value := range root {
		if key == "groupEntryContext" {
			var ctx map[string]json.RawMessage
			if err := json.Unmarshal(value, &ctx); err != nil {
				return fmt.Errorf("decode groupEntryContext: %w", err)
			}
			p.GroupEntryContext = ctx
			continue
		}
		p.Groups[key] = append(json.RawMessage(nil), value...)
	}
	return nil
}

// MarshalJSON writes group keys and optional groupEntryContext.
func (p TranslationProject) MarshalJSON() ([]byte, error) {
	root := make(map[string]json.RawMessage, len(p.Groups)+1)
	for key, value := range p.Groups {
		root[key] = value
	}
	if len(p.GroupEntryContext) > 0 {
		raw, err := json.Marshal(p.GroupEntryContext)
		if err != nil {
			return nil, err
		}
		root["groupEntryContext"] = raw
	}
	return json.Marshal(root)
}

// GetGroup returns a group by name, supporting API and flat offline shapes.
func (p *TranslationProject) GetGroup(groupName string) (*TranslationGroup, error) {
	raw, ok := p.Groups[groupName]
	if !ok {
		return nil, nil
	}

	if len(raw) == 0 {
		return &TranslationGroup{Entries: map[string]json.RawMessage{}}, nil
	}

	if raw[0] != '{' {
		var group TranslationGroup
		if err := json.Unmarshal(raw, &group); err != nil {
			return nil, err
		}
		if group.Entries == nil {
			group.Entries = map[string]json.RawMessage{}
		}
		return &group, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	if _, hasEntries := obj["Entries"]; hasEntries {
		var group TranslationGroup
		if err := json.Unmarshal(raw, &group); err != nil {
			return nil, err
		}
		if group.Entries == nil {
			group.Entries = map[string]json.RawMessage{}
		}
		return &group, nil
	}

	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, err
	}
	return &TranslationGroup{Entries: entries}, nil
}

// ProjectLocales is the response for GET /sdk/v1/translations/locales.
type ProjectLocales struct {
	Project         string     `json:"project,omitempty"`
	Locales         []string   `json:"locales"`
	LastModifiedUTC *time.Time `json:"lastModifiedUtc,omitempty"`
}

// ValidateAPIKeyResponse is the response for GET /api/v1/api-keys/validate.
type ValidateAPIKeyResponse struct {
	IsValid          bool            `json:"isValid"`
	TenantID         json.RawMessage `json:"tenantId,omitempty"`
	ProjectID        json.RawMessage `json:"projectId,omitempty"`
	ProjectIDs       []string        `json:"projectIds,omitempty"`
	DefaultProjectID json.RawMessage `json:"defaultProjectId,omitempty"`
	IntegrationName  string          `json:"integrationName,omitempty"`
	AuthenticatedAt  *time.Time      `json:"authenticatedAt,omitempty"`
}

// OfflineCacheDownloadResult holds offline ZIP download metadata and bytes.
type OfflineCacheDownloadResult struct {
	NotModified       bool
	ETag              string
	SuggestedFileName string
	Content           []byte
}

// ReadJSONULID extracts a string ULID/id from flexible JSON element shapes.
func ReadJSONULID(raw json.RawMessage) string {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(string(raw))
}
