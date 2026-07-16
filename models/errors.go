package models

import (
	"encoding/json"
	"errors"
	"fmt"
)

// TranslaasError is the JSON error envelope returned by the Translaas API.
type TranslaasError struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// FormatMessage returns a display message matching .NET client formatting:
// "[code] message" when code is present, otherwise message or fallback.
func (e TranslaasError) FormatMessage(fallback string) string {
	msg := e.Message
	if msg == "" {
		msg = fallback
	}
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s", e.Code, msg)
	}
	return msg
}

// ParseTranslaasError unmarshals an API error body. Returns nil when body is empty.
func ParseTranslaasError(body []byte) (*TranslaasError, error) {
	if len(body) == 0 {
		return nil, nil
	}
	var err TranslaasError
	if e := json.Unmarshal(body, &err); e != nil {
		return nil, e
	}
	return &err, nil
}

// APIError represents an HTTP failure from the Translaas API.
type APIError struct {
	StatusCode      int
	Code            string
	Message         string
	ResponseContent string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("translaas API error: status %d", e.StatusCode)
}

// ConfigurationError indicates invalid SDK configuration.
type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string {
	return e.Message
}

// OfflineCacheError indicates offline cache I/O or deserialization failures.
type OfflineCacheError struct {
	Message        string
	CacheDirectory string
	Project        string
	Language       string
	Cause          error
}

func (e *OfflineCacheError) Error() string {
	return e.Message
}

func (e *OfflineCacheError) Unwrap() error {
	return e.Cause
}

// OfflineCacheMissError indicates expected data was not found in the offline cache.
type OfflineCacheMissError struct {
	Message        string
	CacheDirectory string
	Project        string
	Language       string
	Group          string
	Entry          string
	Cause          error
}

func (e *OfflineCacheMissError) Error() string {
	return e.Message
}

func (e *OfflineCacheMissError) Unwrap() error {
	return e.Cause
}

// NewOfflineCacheMissError builds a miss error with .NET-compatible messaging.
func NewOfflineCacheMissError(project, language, group, entry string) *OfflineCacheMissError {
	return &OfflineCacheMissError{
		Message:  buildOfflineCacheMissMessage(project, language, group, entry),
		Project:  project,
		Language: language,
		Group:    group,
		Entry:    entry,
	}
}

func buildOfflineCacheMissMessage(project, language, group, entry string) string {
	switch {
	case entry != "" && group != "":
		return fmt.Sprintf(
			"Translation entry '%s' in group '%s' for project '%s' and language '%s' was not found in the offline cache.",
			entry, group, project, language,
		)
	case group != "":
		return fmt.Sprintf(
			"Translation group '%s' for project '%s' and language '%s' was not found in the offline cache.",
			group, project, language,
		)
	default:
		return fmt.Sprintf(
			"Project '%s' for language '%s' was not found in the offline cache.",
			project, language,
		)
	}
}

// ErrNoLanguage is returned when language resolution yields no language.
var ErrNoLanguage = errors.New("no language could be resolved")
