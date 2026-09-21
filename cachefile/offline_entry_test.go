package cachefile

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

func TestSubstituteParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		number   *float64
		params   map[string]string
		want     string
	}{
		{
			name:     "simple substitution",
			template: "Hello {userName}",
			params:   map[string]string{"userName": "John"},
			want:     "Hello John",
		},
		{
			name:     "number injection",
			template: "You have {N} items",
			number:   ptrFloat(5),
			want:     "You have 5 items",
		},
		{
			name:     "number and params",
			template: "Hello {userName}, you have {N} items and {pending} pending",
			number:   ptrFloat(5),
			params:   map[string]string{"userName": "John", "pending": "3"},
			want:     "Hello John, you have 5 items and 3 pending",
		},
		{
			name:     "unknown placeholder preserved",
			template: "Hello {unknown}",
			params:   map[string]string{"userName": "John"},
			want:     "Hello {unknown}",
		},
		{
			name:     "case insensitive lookup",
			template: "Hello {UserName}",
			params:   map[string]string{"username": "Jane"},
			want:     "Hello Jane",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := substituteParameters(tt.template, tt.number, tt.params)
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestDeterminePluralCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		lang string
		n    *float64
		want models.PluralCategory
	}{
		{id: "CLDR-CAT-01", lang: "ar", n: ptrFloat(0), want: models.PluralZero},
		{id: "CLDR-CAT-02", lang: "ar", n: ptrFloat(2), want: models.PluralTwo},
		{id: "CLDR-CAT-03", lang: "pl", n: ptrFloat(2), want: models.PluralFew},
		{id: "CLDR-CAT-04", lang: "fr", n: ptrFloat(0), want: models.PluralOne},
		{id: "CLDR-CAT-05", lang: "en", n: ptrFloat(0), want: models.PluralOther},
		{id: "CLDR-CAT-06", lang: "he", n: ptrFloat(2), want: models.PluralTwo},
		{id: "CLDR-CAT-07", lang: "ja", n: ptrFloat(1), want: models.PluralOther},
		{id: "CLDR-CAT-08", lang: "pt", n: ptrFloat(0), want: models.PluralOne},
		{id: "CLDR-CAT-09", lang: "pt-PT", n: ptrFloat(0), want: models.PluralOther},
		{id: "CLDR-CAT-10", lang: "fr", n: nil, want: models.PluralOther},
		{id: "CLDR-CAT-11", lang: "ar", n: nil, want: models.PluralOther},
		{id: "CLDR-CAT-12", lang: "en", n: ptrFloat(math.NaN()), want: models.PluralOther},
		{id: "CLDR-CAT-13", lang: "en", n: ptrFloat(math.Inf(1)), want: models.PluralOther},
		{id: "CLDR-CAT-14", lang: "en", n: ptrFloat(1), want: models.PluralOne},
		{id: "CLDR-CAT-15", lang: "", n: ptrFloat(1), want: models.PluralOne},
		{id: "CLDR-CAT-16", lang: "not a locale!!", n: ptrFloat(0), want: models.PluralOther},
		{id: "CLDR-CAT-17", lang: "", n: ptrFloat(0), want: models.PluralOther},
		{id: "en-n-2", lang: "en", n: ptrFloat(2), want: models.PluralOther},
		{id: "en-1.0", lang: "en", n: ptrFloat(1.0), want: models.PluralOne},
		{id: "en-1.5", lang: "en", n: ptrFloat(1.5), want: models.PluralOther},
		{id: "fr-1.5", lang: "fr", n: ptrFloat(1.5), want: models.PluralOne},
		{id: "blank-spaces", lang: "   ", n: ptrFloat(1), want: models.PluralOne},
		{id: "invalid-n-1", lang: "not a locale!!", n: ptrFloat(1), want: models.PluralOne},
		{id: "en_US", lang: "en_US", n: ptrFloat(1), want: models.PluralOne},
		{id: "ar_EG", lang: "ar_EG", n: ptrFloat(0), want: models.PluralZero},
		{id: "bg-2", lang: "bg", n: ptrFloat(2), want: models.PluralOther},
		{id: "fr-CA-0", lang: "fr-CA", n: ptrFloat(0), want: models.PluralOne},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			got := determinePluralCategory(tt.n, tt.lang)
			if got != tt.want {
				t.Fatalf("lang=%q n=%v got %s want %s", tt.lang, formatPtrFloat(tt.n), got, tt.want)
			}
		})
	}
}

func TestResolveEntryFromGroupPlural(t *testing.T) {
	t.Parallel()

	t.Run("CLDR-RES-01", func(t *testing.T) {
		t.Parallel()
		group := pluralGroup(`{"One":"1 item","Other":"{N} items"}`)

		one, ok := resolveEntryFromGroup(group, "items", ptrFloat(1), "en", nil)
		if !ok || one != "1 item" {
			t.Fatalf("n=1 got (%q, %v)", one, ok)
		}

		other, ok := resolveEntryFromGroup(group, "items", ptrFloat(2), "en", nil)
		if !ok || other != "2 items" {
			t.Fatalf("n=2 got (%q, %v)", other, ok)
		}
	})

	t.Run("CLDR-RES-02", func(t *testing.T) {
		t.Parallel()
		group := pluralGroup(`{"One":"zero/one fr","Other":"other fr"}`)
		got, ok := resolveEntryFromGroup(group, "items", ptrFloat(0), "fr", nil)
		if !ok || got != "zero/one fr" {
			t.Fatalf("fr n=0 got (%q, %v)", got, ok)
		}
	})

	t.Run("CLDR-RES-03", func(t *testing.T) {
		t.Parallel()
		group := pluralGroup(`{"One":"one en","Other":"other en"}`)
		got, ok := resolveEntryFromGroup(group, "items", ptrFloat(0), "en", nil)
		if !ok || got != "other en" {
			t.Fatalf("en n=0 got (%q, %v)", got, ok)
		}
	})

	t.Run("CLDR-RES-04", func(t *testing.T) {
		t.Parallel()
		group := pluralGroup(`{"Other":"fallback only"}`)
		got, ok := resolveEntryFromGroup(group, "items", ptrFloat(0), "ar", nil)
		if !ok || got != "fallback only" {
			t.Fatalf("missing Zero fallback got (%q, %v)", got, ok)
		}
	})

	t.Run("CLDR-RES-05", func(t *testing.T) {
		t.Parallel()
		group := pluralGroup(`{"One":"one","Other":"other"}`)
		got, ok := resolveEntryFromGroup(group, "items", ptrFloat(0), "", nil)
		if !ok || got != "other" {
			t.Fatalf("invalid lang n=0 got (%q, %v)", got, ok)
		}
	})
}

func pluralGroup(formsJSON string) *models.TranslationGroup {
	return &models.TranslationGroup{
		Entries: map[string]json.RawMessage{
			"items": json.RawMessage(formsJSON),
		},
	}
}

func ptrFloat(v float64) *float64 {
	return &v
}

func formatPtrFloat(n *float64) string {
	if n == nil {
		return "<nil>"
	}
	return strconv.FormatFloat(*n, 'g', -1, 64)
}
