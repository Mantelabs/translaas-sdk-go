// Package service provides the convenience translation API and language resolution.
//
// Hello-world uses client.Options.DefaultLanguage and service.New(c) — no resolver:
//
//	c, err := client.New(client.Options{
//	    APIKey: os.Getenv("TRANSLAAS_API_KEY"),
//	    BaseURL: "https://api.translaas.local",
//	    DefaultProjectID: "my-project",
//	    DefaultLanguage: "en",
//	})
//	svc, err := service.New(c)
//	text, err := svc.T(ctx, "common", "welcome.message", service.WithLang("en"))
//
// For web apps, chain language providers and pass Options{Resolver}:
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
