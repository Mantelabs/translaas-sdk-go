// Package language provides automatic language resolution for service.T.
//
// Chain providers with NewResolver; the first non-empty language wins. Typical order:
//
//	resolver, err := language.NewResolver(
//	    language.NewContextLanguageProvider(),
//	    language.NewAcceptLanguageProvider(),
//	    language.NewDefaultLanguageProvider("en"),
//	)
package language
