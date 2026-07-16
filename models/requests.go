package models

// GetTranslationRequest is the query model for GET /sdk/v1/translations/text.
type GetTranslationRequest struct {
	Group   string   `json:"group,omitempty"`
	Entry   string   `json:"entry,omitempty"`
	Lang    string   `json:"lang,omitempty"`
	Number  *float64 `json:"n,omitempty"`
	Project string   `json:"project,omitempty"`
	Channel string   `json:"channel,omitempty"`
	Version string   `json:"v,omitempty"`
}

// GetGroupTranslationsRequest is the query model for GET /sdk/v1/translations/group.
type GetGroupTranslationsRequest struct {
	Project        string `json:"project,omitempty"`
	Group          string `json:"group,omitempty"`
	Lang           string `json:"lang,omitempty"`
	Format         string `json:"format,omitempty"`
	Channel        string `json:"channel,omitempty"`
	Version        string `json:"v,omitempty"`
	IncludeContext *bool  `json:"includeContext,omitempty"`
}

// GetProjectTranslationsRequest is the query model for GET /sdk/v1/translations/project.
type GetProjectTranslationsRequest struct {
	Project        string `json:"project,omitempty"`
	Lang           string `json:"lang,omitempty"`
	Format         string `json:"format,omitempty"`
	Channel        string `json:"channel,omitempty"`
	Version        string `json:"v,omitempty"`
	IncludeContext *bool  `json:"includeContext,omitempty"`
}

// GetProjectLocalesRequest is the query model for GET /sdk/v1/translations/locales.
type GetProjectLocalesRequest struct {
	Project string `json:"project,omitempty"`
	Channel string `json:"channel,omitempty"`
	Version string `json:"v,omitempty"`
}

// GetOfflineCacheRequest is the query model for GET /sdk/v1/translations/offline-cache.
type GetOfflineCacheRequest struct {
	Project        string `json:"project,omitempty"`
	Channel        string `json:"channel,omitempty"`
	Version        string `json:"v,omitempty"`
	IncludeContext *bool  `json:"includeContext,omitempty"`
}

// ReportMissingKeyItem is one missing key for POST /sdk/v1/translations/report-missing.
type ReportMissingKeyItem struct {
	GroupKey        string `json:"groupKey"`
	EntryKey        string `json:"entryKey"`
	LanguageIsoCode string `json:"languageIsoCode"`
}

// ReportMissingKeysRequest is the POST body for reporting missing translation keys.
type ReportMissingKeysRequest struct {
	Keys []ReportMissingKeyItem `json:"keys"`
}
