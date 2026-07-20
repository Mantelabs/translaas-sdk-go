package language

import (
	"regexp"
	"strings"
)

var regionLanguagePattern = regexp.MustCompile(`^([a-z]{2})(?:[-_][a-z]{2,})?$`)

// ParseAcceptLanguage extracts the primary ISO 639-1 code from an Accept-Language header.
func ParseAcceptLanguage(acceptLanguage string) string {
	if acceptLanguage == "" {
		return ""
	}

	parts := strings.Split(acceptLanguage, ",")
	if len(parts) == 0 {
		return ""
	}

	first := strings.TrimSpace(parts[0])
	if first == "" {
		return ""
	}

	if idx := strings.Index(first, ";"); idx >= 0 {
		first = strings.TrimSpace(first[:idx])
	}

	return NormalizeLanguageCode(first)
}

// NormalizeLanguageCode converts a language tag to ISO 639-1 when possible.
func NormalizeLanguageCode(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return ""
	}

	if len(lang) == 2 && isAlpha2(lang) {
		return lang
	}

	match := regionLanguagePattern.FindStringSubmatch(lang)
	if len(match) == 2 && isAlpha2(match[1]) {
		return match[1]
	}

	return ""
}

func isAlpha2(code string) bool {
	if len(code) != 2 {
		return false
	}
	for _, r := range code {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}
