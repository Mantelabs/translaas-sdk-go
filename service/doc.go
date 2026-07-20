// Package service provides the convenience translation API and language resolution.
//
// Use Service.T for high-level lookups with optional automatic language resolution:
//
//	resolver, err := language.NewResolver(
//	    language.NewContextLanguageProvider(),
//	    language.NewAcceptLanguageProvider(),
//	    language.NewDefaultLanguageProvider("en"),
//	)
//	svc, err := service.New(cachingClient, service.Options{Resolver: resolver})
//	text, err := svc.T(ctx, "common", "welcome")
//
// Pass service.WithLang to bypass the resolver. The inner client may be a plain
// client.Client or cachefile.CachingClient — Service only delegates to GetEntry.
package service
