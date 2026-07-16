package models

// RequestContext carries per-request options and response metadata for SDK calls.
type RequestContext struct {
	// Request fields mapped to query params or headers.
	Channel        string
	Version        string
	Project        string
	IncludeContext *bool
	IfNoneMatch    string

	// Response fields populated by the HTTP client after each request.
	ResponseETag string
	NotModified  bool
}

// Reset clears response fields before a new request.
func (c *RequestContext) Reset() {
	if c == nil {
		return
	}
	c.ResponseETag = ""
	c.NotModified = false
}
