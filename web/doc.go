// Package web provides optional integrations with Go web frameworks.
//
// The root web package depends only on the standard library and service. Framework
// adapters live in separate modules under web/gin, web/echo, and web/chi so the core
// module does not pull third-party routers.
//
// # Security
//
// Translation strings returned by the SDK are not HTML-escaped. When rendering HTML,
// use html/template (which auto-escapes) rather than text/template or raw string
// concatenation. JSON and API responses must be encoded at the serializer layer.
package web
