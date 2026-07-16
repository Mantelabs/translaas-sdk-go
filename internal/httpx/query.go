package httpx

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type queryParam struct {
	key   string
	value string
}

// AppendQueryValues adds query parameters by reflecting exported fields on params.
// params must be a struct or pointer to struct. JSON struct tag names are used;
// fields with tag "-" or no tag are skipped. Nil pointers, empty strings, zero
// numeric values, and false bools are omitted. Non-nil *bool values are always
// included as "true" or "false".
func AppendQueryValues(u *url.URL, params any) error {
	if u == nil {
		return fmt.Errorf("httpx: url is nil")
	}
	if params == nil {
		return fmt.Errorf("httpx: params is nil")
	}

	value := reflect.ValueOf(params)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return fmt.Errorf("httpx: params is nil")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("httpx: params must be a struct or pointer to struct")
	}

	current := parseQueryParams(u.RawQuery)
	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name, ok := jsonQueryName(field)
		if !ok {
			continue
		}
		formatted, ok := formatQueryValue(value.Field(i))
		if !ok {
			continue
		}
		current = upsertQueryParam(current, name, formatted)
	}
	u.RawQuery = encodeQueryParams(current)
	return nil
}

// MergeQueryParams merges extra into u's query string. On case-insensitive key
// collision the existing key is removed and replaced with extra's key casing.
// Empty values in extra are skipped.
func MergeQueryParams(u *url.URL, extra map[string]string) {
	if u == nil || len(extra) == 0 {
		return
	}

	current := parseQueryParams(u.RawQuery)
	for key, value := range extra {
		if strings.TrimSpace(value) == "" {
			continue
		}
		current = upsertQueryParam(current, key, value)
	}
	u.RawQuery = encodeQueryParams(current)
}

// InjectPluralN adds key "N" with invariant-formatted n when no case-insensitive
// "N" key exists in extra. No-op when n or extra is nil.
func InjectPluralN(extra map[string]string, n *float64) {
	if extra == nil || n == nil {
		return
	}
	for key := range extra {
		if strings.EqualFold(key, "N") {
			return
		}
	}
	extra["N"] = strconv.FormatFloat(*n, 'f', -1, 64)
}

func jsonQueryName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return "", false
	}
	name, _, _ := strings.Cut(tag, ",")
	name = strings.TrimSpace(name)
	if name == "" || name == "-" {
		return "", false
	}
	return name, true
}

func formatQueryValue(v reflect.Value) (string, bool) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return "", false
		}
		elem := v.Elem()
		switch elem.Kind() {
		case reflect.Bool:
			return strconv.FormatBool(elem.Bool()), true
		case reflect.Float32, reflect.Float64:
			return strconv.FormatFloat(elem.Float(), 'f', -1, 64), true
		default:
			return formatQueryValue(elem)
		}
	case reflect.String:
		s := v.String()
		if s == "" {
			return "", false
		}
		return s, true
	case reflect.Bool:
		if !v.Bool() {
			return "", false
		}
		return strconv.FormatBool(v.Bool()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := v.Int()
		if i == 0 {
			return "", false
		}
		return strconv.FormatInt(i, 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i := v.Uint()
		if i == 0 {
			return "", false
		}
		return strconv.FormatUint(i, 10), true
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if f == 0 {
			return "", false
		}
		return strconv.FormatFloat(f, 'f', -1, 64), true
	default:
		return "", false
	}
}

func parseQueryParams(rawQuery string) []queryParam {
	if rawQuery == "" {
		return nil
	}
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return nil
	}
	params := make([]queryParam, 0, len(values))
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}
		params = append(params, queryParam{key: key, value: vals[0]})
	}
	return params
}

func upsertQueryParam(params []queryParam, key, value string) []queryParam {
	out := make([]queryParam, 0, len(params)+1)
	for _, p := range params {
		if strings.EqualFold(p.key, key) {
			continue
		}
		out = append(out, p)
	}
	return append(out, queryParam{key: key, value: value})
}

func encodeQueryParams(params []queryParam) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = url.QueryEscape(p.key) + "=" + url.QueryEscape(p.value)
	}
	return strings.Join(parts, "&")
}

// QueryValues returns a case-sensitive map of decoded query parameters from u.
// Useful for order-independent test assertions.
func QueryValues(u *url.URL) url.Values {
	if u == nil {
		return url.Values{}
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return url.Values{}
	}
	return values
}
