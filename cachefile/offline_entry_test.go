package cachefile

import (
	"encoding/json"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/models"
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

func TestResolveEntryFromGroupPlural(t *testing.T) {
	t.Parallel()

	group := &models.TranslationGroup{
		Entries: map[string]json.RawMessage{
			"items": json.RawMessage(`{"One":"1 item","Other":"{N} items"}`),
		},
	}

	one, ok := resolveEntryFromGroup(group, "items", ptrFloat(1), nil)
	if !ok || one != "1 item" {
		t.Fatalf("n=1 got (%q, %v)", one, ok)
	}

	other, ok := resolveEntryFromGroup(group, "items", ptrFloat(2), nil)
	if !ok || other != "2 items" {
		t.Fatalf("n=2 got (%q, %v)", other, ok)
	}
}

func TestDeterminePluralCategory(t *testing.T) {
	t.Parallel()

	if determinePluralCategory(nil) != models.PluralOther {
		t.Fatal("nil number should map to Other")
	}
	if determinePluralCategory(ptrFloat(1)) != models.PluralOne {
		t.Fatal("1 should map to One")
	}
	if determinePluralCategory(ptrFloat(2)) != models.PluralOther {
		t.Fatal("2 should map to Other")
	}
}

func ptrFloat(v float64) *float64 {
	return &v
}
