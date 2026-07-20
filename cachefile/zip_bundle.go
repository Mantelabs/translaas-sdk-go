package cachefile

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/acuencadev/translaas-sdk-go/models"
)

// OfflineBundle holds parsed offline ZIP contents keyed by path segments from the archive.
type OfflineBundle struct {
	Manifest              CacheManifest
	LocalesByProject      map[string]CachedLocales
	ProjectsByProjectLang map[string]map[string]CachedProject
}

// ParseOfflineZip reads an offline ZIP bundle (HTTP spec §7.6).
func ParseOfflineZip(content []byte) (*OfflineBundle, error) {
	if len(content) == 0 {
		return nil, zipParseErr("offline ZIP content is empty", fmt.Errorf("empty content"))
	}

	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, zipParseErr("invalid offline ZIP archive", err)
	}

	bundle := &OfflineBundle{
		LocalesByProject:      make(map[string]CachedLocales),
		ProjectsByProjectLang: make(map[string]map[string]CachedProject),
	}

	for _, file := range reader.File {
		if err := validateZipEntryName(file.Name); err != nil {
			return nil, zipParseErr(fmt.Sprintf("unsafe ZIP entry %q", file.Name), err)
		}
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		raw, err := readZipEntry(file)
		if err != nil {
			return nil, zipParseErr(fmt.Sprintf("read ZIP entry %q", file.Name), err)
		}

		if file.Name == "manifest.json" {
			var manifest CacheManifest
			if err := json.Unmarshal(raw, &manifest); err != nil {
				return nil, zipParseErr("decode manifest.json", err)
			}
			bundle.Manifest = manifest
			continue
		}

		parts := strings.Split(file.Name, "/")
		if len(parts) < 2 {
			continue
		}

		projectSegment := parts[0]
		fileName := parts[len(parts)-1]

		switch {
		case fileName == "locales.json" && len(parts) == 2:
			var wrapped CachedLocales
			if err := json.Unmarshal(raw, &wrapped); err != nil {
				return nil, zipParseErr(fmt.Sprintf("decode %q", file.Name), err)
			}
			bundle.LocalesByProject[projectSegment] = wrapped
		case fileName == "project.json" && len(parts) == 3:
			langSegment := parts[1]
			var wrapped CachedProject
			if err := json.Unmarshal(raw, &wrapped); err != nil {
				return nil, zipParseErr(fmt.Sprintf("decode %q", file.Name), err)
			}
			if bundle.ProjectsByProjectLang[projectSegment] == nil {
				bundle.ProjectsByProjectLang[projectSegment] = make(map[string]CachedProject)
			}
			bundle.ProjectsByProjectLang[projectSegment][langSegment] = wrapped
		}
	}

	return bundle, nil
}

// ResolveProjectKey maps a logical project id to the folder key used inside the bundle.
func ResolveProjectKey(bundle *OfflineBundle, project string) (string, error) {
	if bundle == nil {
		return "", fmt.Errorf("offline bundle must not be nil")
	}

	project = strings.TrimSpace(project)
	if project == "" {
		return "", fmt.Errorf("project must not be empty")
	}

	sanitized, err := SanitizePathSegment(project)
	if err != nil {
		return "", err
	}

	if bundle.hasProjectData(sanitized) {
		return sanitized, nil
	}
	if project != sanitized && bundle.hasProjectData(project) {
		return project, nil
	}

	if bundle.Manifest.Projects != nil {
		if _, ok := bundle.Manifest.Projects[sanitized]; ok {
			return sanitized, nil
		}
		if project != sanitized {
			if _, ok := bundle.Manifest.Projects[project]; ok {
				return project, nil
			}
		}
	}

	return "", fmt.Errorf("project %q not found in offline bundle", project)
}

func (b *OfflineBundle) hasProjectData(key string) bool {
	if _, ok := b.LocalesByProject[key]; ok {
		return true
	}
	langs, ok := b.ProjectsByProjectLang[key]
	return ok && len(langs) > 0
}

func validateZipEntryName(name string) error {
	if name == "" {
		return fmt.Errorf("empty entry name")
	}
	if strings.Contains(name, `\`) {
		return fmt.Errorf("backslash in entry name")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("parent traversal in entry name")
	}
	if strings.HasPrefix(name, "/") {
		return fmt.Errorf("absolute entry name")
	}

	cleaned := path.Clean(name)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("parent traversal in entry name")
	}
	return nil
}

func readZipEntry(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()

	return io.ReadAll(rc)
}

func zipParseErr(msg string, cause error) error {
	return &models.OfflineCacheError{
		Message: msg,
		Cause:   cause,
	}
}

func saveOptionsFromWrapper(expiresAt *time.Time) []SaveOption {
	if expiresAt == nil {
		return nil
	}
	expiry := *expiresAt
	return []SaveOption{WithExpiresAt(&expiry)}
}
