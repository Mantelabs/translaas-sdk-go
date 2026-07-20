package language

import "context"

type contextKey struct {
	name string
}

var (
	languageKey       = contextKey{name: "translaas-language"}
	acceptLanguageKey = contextKey{name: "translaas-accept-language"}
)

// WithLanguage stores an explicit language code on ctx for ContextLanguageProvider.
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, languageKey, lang)
}

// LanguageFromContext returns the explicit language stored by WithLanguage.
func LanguageFromContext(ctx context.Context) (string, bool) {
	lang, ok := ctx.Value(languageKey).(string)
	if !ok || lang == "" {
		return "", false
	}
	return lang, true
}

// WithAcceptLanguage stores a raw Accept-Language header value on ctx.
func WithAcceptLanguage(ctx context.Context, acceptLanguageHeader string) context.Context {
	return context.WithValue(ctx, acceptLanguageKey, acceptLanguageHeader)
}

// AcceptLanguageFromContext returns the Accept-Language header stored by WithAcceptLanguage.
func AcceptLanguageFromContext(ctx context.Context) (string, bool) {
	header, ok := ctx.Value(acceptLanguageKey).(string)
	if !ok || header == "" {
		return "", false
	}
	return header, true
}
