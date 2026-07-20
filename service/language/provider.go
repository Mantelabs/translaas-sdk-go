package language

import (
	"context"
	"strings"
)

// Provider resolves a language code from context (.NET ILanguageProvider).
type Provider interface {
	Language(ctx context.Context) (string, error)
}

type defaultLanguageProvider struct {
	lang string
}

// NewDefaultLanguageProvider returns a provider that always yields the configured language.
func NewDefaultLanguageProvider(lang string) Provider {
	normalized := NormalizeLanguageCode(lang)
	if normalized == "" {
		normalized = strings.TrimSpace(strings.ToLower(lang))
	}
	return defaultLanguageProvider{lang: normalized}
}

func (p defaultLanguageProvider) Language(context.Context) (string, error) {
	if p.lang == "" {
		return "", nil
	}
	return p.lang, nil
}

type acceptLanguageProvider struct{}

// NewAcceptLanguageProvider parses Accept-Language from context (see WithAcceptLanguage).
func NewAcceptLanguageProvider() Provider {
	return acceptLanguageProvider{}
}

func (acceptLanguageProvider) Language(ctx context.Context) (string, error) {
	header, ok := AcceptLanguageFromContext(ctx)
	if !ok {
		return "", nil
	}
	return ParseAcceptLanguage(header), nil
}

type contextLanguageProvider struct{}

// NewContextLanguageProvider reads an explicit language from context (see WithLanguage).
func NewContextLanguageProvider() Provider {
	return contextLanguageProvider{}
}

func (contextLanguageProvider) Language(ctx context.Context) (string, error) {
	lang, ok := LanguageFromContext(ctx)
	if !ok {
		return "", nil
	}
	normalized := NormalizeLanguageCode(lang)
	if normalized != "" {
		return normalized, nil
	}
	return strings.TrimSpace(strings.ToLower(lang)), nil
}
