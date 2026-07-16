package models

import (
	"strings"
)

// ResolveDefaultProjectID returns the effective default project id when not configured explicitly.
func ResolveDefaultProjectID(configuredProjectID string, validate *ValidateAPIKeyResponse) (string, error) {
	if strings.TrimSpace(configuredProjectID) != "" {
		return strings.TrimSpace(configuredProjectID), nil
	}
	if validate == nil || len(validate.ProjectIDs) == 0 {
		return "", &ConfigurationError{
			Message: "Tenant-level API key requires DefaultProjectId in SDK configuration.",
		}
	}

	fromValidate := ReadJSONULID(validate.DefaultProjectID)
	if fromValidate == "" {
		fromValidate = ReadJSONULID(validate.ProjectID)
	}
	if fromValidate == "" && len(validate.ProjectIDs) > 0 {
		fromValidate = strings.TrimSpace(validate.ProjectIDs[0])
	}
	if fromValidate == "" {
		return "", &ConfigurationError{
			Message: "Could not resolve a default project from the validate API key response.",
		}
	}
	return fromValidate, nil
}
