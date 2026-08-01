package cachefile

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

// Offline pluralization uses simplified rules (1 → One, else Other), not full CLDR.
// Live API evaluates server-side plural rules via the n query parameter.

var placeholderPattern = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

func resolveEntryFromGroup(
	group *models.TranslationGroup,
	entry string,
	number *float64,
	params map[string]string,
) (string, bool) {
	if group == nil {
		return "", false
	}

	if group.HasPluralForms(entry) {
		category := determinePluralCategory(number)
		form, ok := group.GetPluralForm(entry, category)
		if !ok && category != models.PluralOther {
			form, ok = group.GetPluralForm(entry, models.PluralOther)
		}
		if !ok {
			return "", false
		}
		return substituteParameters(form, number, params), true
	}

	value, ok := group.GetValue(entry)
	if !ok {
		return "", false
	}
	return substituteParameters(value, number, params), true
}

func determinePluralCategory(number *float64) models.PluralCategory {
	if number == nil {
		return models.PluralOther
	}
	if *number == 1 {
		return models.PluralOne
	}
	return models.PluralOther
}

func substituteParameters(template string, number *float64, params map[string]string) string {
	if template == "" {
		return ""
	}

	merged := mergeNumberIntoParameters(number, params)
	if len(merged) == 0 {
		return template
	}

	return placeholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		submatches := placeholderPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		if value, ok := paramLookup(merged, submatches[1]); ok {
			return value
		}
		return match
	})
}

func mergeNumberIntoParameters(number *float64, params map[string]string) map[string]string {
	if number == nil && len(params) == 0 {
		return nil
	}

	merged := make(map[string]string)
	for k, v := range params {
		merged[k] = v
	}
	if number != nil && !hasParamKey(merged, "N") {
		merged["N"] = strconv.FormatFloat(*number, 'f', -1, 64)
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func hasParamKey(params map[string]string, name string) bool {
	_, ok := paramLookup(params, name)
	return ok
}

func paramLookup(params map[string]string, name string) (string, bool) {
	for k, v := range params {
		if strings.EqualFold(k, name) {
			return v, true
		}
	}
	return "", false
}

func newLocalesOfflineCacheError(project string, cause error) *models.OfflineCacheError {
	msg := fmt.Sprintf("Project locales for '%s' not found in offline cache.", project)
	if cause != nil {
		msg = fmt.Sprintf("Project locales for '%s' not found in offline cache and API is unavailable.", project)
	}
	return &models.OfflineCacheError{
		Message: msg,
		Project: project,
		Cause:   cause,
	}
}
