package cachefile

import (
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/Mantelabs/translaas-sdk-go/models"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
)

const (
	englishLocale   = "en"
	maxSafeInteger  = 1 << 53
	operandModLimit = 10_000_000
)

var languageTagCache sync.Map // map[string]language.Tag

func normalizeLocaleTag(lang string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(lang, "_", "-"))
	if trimmed == "" {
		return englishLocale
	}
	return trimmed
}

func baseLanguage(locale string) string {
	if i := strings.IndexByte(locale, '-'); i > 0 {
		return locale[:i]
	}
	return locale
}

// looksLikeBCP47 accepts tags such as en, pt-PT, zh-Hans-CN.
// It rejects free text so language.Parse is not asked to produce und (always Other).
func looksLikeBCP47(locale string) bool {
	if len(locale) < 2 {
		return false
	}

	firstSegment := true
	segmentLength := 0
	for i := 0; i < len(locale); i++ {
		c := locale[i]
		if c == '-' {
			if segmentLength == 0 || (firstSegment && segmentLength < 2) {
				return false
			}
			firstSegment = false
			segmentLength = 0
			continue
		}

		isLetter := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		isDigit := c >= '0' && c <= '9'
		if firstSegment {
			if !isLetter {
				return false
			}
		} else if !isLetter && !isDigit {
			return false
		}

		segmentLength++
		if segmentLength > 8 {
			return false
		}
	}

	if firstSegment {
		return segmentLength >= 2
	}
	return segmentLength >= 1
}

func tryParseLanguageTag(tag string) (language.Tag, bool) {
	if !looksLikeBCP47(tag) {
		return language.Und, false
	}
	parsed, err := language.Parse(tag)
	if err != nil || parsed.IsRoot() {
		return language.Und, false
	}
	return parsed, true
}

func resolveLanguageTag(normalized string) language.Tag {
	if tag, ok := tryParseLanguageTag(normalized); ok {
		return tag
	}
	base := baseLanguage(normalized)
	if !strings.EqualFold(base, normalized) {
		if tag, ok := tryParseLanguageTag(base); ok {
			return tag
		}
	}
	return language.English
}

func languageTagForPlural(lang string) language.Tag {
	normalized := normalizeLocaleTag(lang)
	key := strings.ToLower(normalized)
	if cached, ok := languageTagCache.Load(key); ok {
		return cached.(language.Tag)
	}
	tag := resolveLanguageTag(normalized)
	languageTagCache.Store(key, tag)
	return tag
}

func operandsFromFloat64(n float64) (i, v, w, f, t int, ok bool) {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, 0, 0, 0, 0, false
	}

	abs := math.Abs(n)
	if abs <= maxSafeInteger && abs == math.Trunc(abs) {
		return clampOperand(int(abs)), 0, 0, 0, 0, true
	}

	formatted := strconv.FormatFloat(abs, 'f', -1, 64)
	intPart, fracPart, hasFrac := strings.Cut(formatted, ".")
	intVal, err := strconv.Atoi(intPart)
	if err != nil {
		return 0, 0, 0, 0, 0, false
	}
	i = clampOperand(intVal)
	if !hasFrac || fracPart == "" {
		return i, 0, 0, 0, 0, true
	}

	v = len(fracPart)
	trimmed := strings.TrimRight(fracPart, "0")
	w = len(trimmed)
	if frac, err := strconv.Atoi(fracPart); err == nil {
		f = clampOperand(frac)
	}
	if trimmed == "" {
		t = 0
	} else if frac, err := strconv.Atoi(trimmed); err == nil {
		t = clampOperand(frac)
	}
	return i, v, w, f, t, true
}

func clampOperand(n int) int {
	if n < 0 {
		n = -n
	}
	if n >= operandModLimit {
		return n % operandModLimit
	}
	return n
}

func mapPluralForm(form plural.Form) models.PluralCategory {
	switch form {
	case plural.Zero:
		return models.PluralZero
	case plural.One:
		return models.PluralOne
	case plural.Two:
		return models.PluralTwo
	case plural.Few:
		return models.PluralFew
	case plural.Many:
		return models.PluralMany
	default:
		return models.PluralOther
	}
}

func matchCardinalCategory(tag language.Tag, n float64) models.PluralCategory {
	i, v, w, f, t, ok := operandsFromFloat64(n)
	if !ok {
		return models.PluralOther
	}
	return mapPluralForm(plural.Cardinal.MatchPlural(tag, i, v, w, f, t))
}
