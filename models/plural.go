package models

import "strings"

// PluralCategory represents CLDR plural categories used in translation payloads.
type PluralCategory int

const (
	// PluralZero represents the CLDR zero plural category.
	PluralZero PluralCategory = iota
	// PluralOne represents the CLDR one plural category.
	PluralOne
	// PluralTwo represents the CLDR two plural category.
	PluralTwo
	// PluralFew represents the CLDR few plural category.
	PluralFew
	// PluralMany represents the CLDR many plural category.
	PluralMany
	// PluralOther represents the CLDR other plural category.
	PluralOther
)

// String returns the enum name used in JSON plural maps.
func (c PluralCategory) String() string {
	switch c {
	case PluralZero:
		return "Zero"
	case PluralOne:
		return "One"
	case PluralTwo:
		return "Two"
	case PluralFew:
		return "Few"
	case PluralMany:
		return "Many"
	case PluralOther:
		return "Other"
	default:
		return "Other"
	}
}

func parsePluralCategory(name string) (PluralCategory, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "zero":
		return PluralZero, true
	case "one":
		return PluralOne, true
	case "two":
		return PluralTwo, true
	case "few":
		return PluralFew, true
	case "many":
		return PluralMany, true
	case "other":
		return PluralOther, true
	default:
		return PluralOther, false
	}
}
