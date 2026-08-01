package cachefile

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

func isInvalidFilenameRune(r rune) bool {
	if r < 32 {
		return true
	}
	switch r {
	case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
		return true
	default:
		return false
	}
}

func validateSegmentInput(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("path segment must not be empty")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("path segment must not contain '..'")
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("path segment must not be absolute")
	}
	return nil
}

// SanitizePathSegment replaces invalid filename characters with '_' for use in cache directory names.
func SanitizePathSegment(name string) (string, error) {
	if err := validateSegmentInput(name); err != nil {
		return "", err
	}

	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if isInvalidFilenameRune(r) {
			b.WriteRune('_')
			continue
		}
		b.WriteRune(r)
	}

	out := strings.TrimSpace(b.String())
	if out == "" {
		return "", fmt.Errorf("path segment is empty after sanitization")
	}
	return out, nil
}

func (p *FileProvider) sanitizedProjectDir(project string) (string, string, error) {
	safe, err := SanitizePathSegment(project)
	if err != nil {
		return "", "", offlineCacheErr(p.dir, project, "", fmt.Sprintf("invalid project id: %s", project), err)
	}
	return filepath.Join(p.dir, safe), safe, nil
}

func (p *FileProvider) projectFile(projectDir, lang string) (string, error) {
	safeLang, err := SanitizePathSegment(lang)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectDir, safeLang, "project.json"), nil
}

func (p *FileProvider) localesFile(projectDir string) string {
	return filepath.Join(projectDir, "locales.json")
}

func (p *FileProvider) manifestFile() string {
	return filepath.Join(p.dir, "manifest.json")
}

func containsLanguage(languages []string, lang string) bool {
	for _, existing := range languages {
		if existing == lang {
			return true
		}
	}
	return false
}

func appendLanguage(languages []string, lang string) []string {
	if containsLanguage(languages, lang) {
		return languages
	}
	out := make([]string, len(languages)+1)
	copy(out, languages)
	out[len(languages)] = lang
	return out
}

func normalizeLocales(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if !containsLanguage(out, value) {
			out = append(out, value)
		}
	}
	return out
}

func isExpired(expiresAt *time.Time, now time.Time) bool {
	return expiresAt != nil && expiresAt.Before(now)
}

func offlineCacheErr(dir, project, lang, msg string, cause error) *models.OfflineCacheError {
	if msg == "" && cause != nil {
		msg = cause.Error()
	}
	return &models.OfflineCacheError{
		Message:        msg,
		CacheDirectory: dir,
		Project:        project,
		Language:       lang,
		Cause:          cause,
	}
}

func checkContext(ctx context.Context, dir, project, lang string) error {
	if err := ctx.Err(); err != nil {
		return offlineCacheErr(dir, project, lang, "operation cancelled", err)
	}
	return nil
}
