# Translaas SDK for Go

Official [Translaas](https://github.com/acuencadev/translaas-all) client SDK for Go. Delivers translations from the Translaas SDK Delivery API with in-memory caching, file-backed offline mode, a convenience `T()` API, and optional web framework integrations.

Part of the [translaas-all](https://github.com/acuencadev/translaas-all) umbrella workspace.

## Status

MVP complete (milestones **M1–M4**): HTTP client, in-memory cache, offline file cache, sync service, `service.T()`, and optional `web` middleware. Current release: **`v0.5.0`**. Track work via [GitHub Issues](https://github.com/Mantelabs/translaas-sdk-go/issues) and semver tags.

Runnable sample apps live in the meta-repo under [`examples/go/`](https://github.com/acuencadev/translaas-all/tree/main/examples/go) — not in this library repository.

## Requirements

- Go **1.26+** (module declares `go 1.26.0`)

## Installation

Pin to a semver tag (recommended for production):

```bash
go get github.com/Mantelabs/translaas-sdk-go@v0.5.0
```

Track commits on `main` after the latest tag:

```bash
go get github.com/Mantelabs/translaas-sdk-go@latest
```

The Go module path is `github.com/Mantelabs/translaas-sdk-go`. When developing from a `translaas-all` checkout before submodule registration ([#18](https://github.com/Mantelabs/translaas-sdk-go/issues/18)), use a local `replace` in your `go.mod`.

### Packages

Consumers import subpackages — there is no single root package:

| Package | Import path | Role |
|---------|-------------|------|
| `models` | `github.com/Mantelabs/translaas-sdk-go/models` | DTOs, typed errors, request context |
| `cache` | `github.com/Mantelabs/translaas-sdk-go/cache` | `Mode`, key builder, memory provider |
| `client` | `github.com/Mantelabs/translaas-sdk-go/client` | HTTP client |
| `cachefile` | `github.com/Mantelabs/translaas-sdk-go/cachefile` | Disk cache, hybrid L1, decorator, sync |
| `service` | `github.com/Mantelabs/translaas-sdk-go/service` | `T()` convenience API |
| `web` | `github.com/Mantelabs/translaas-sdk-go/web` | stdlib middleware (optional: `web/gin`, `web/echo`, `web/chi`) |

## Quick start

### Option A — `service.Service` (recommended)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Mantelabs/translaas-sdk-go/cache"
	"github.com/Mantelabs/translaas-sdk-go/client"
	"github.com/Mantelabs/translaas-sdk-go/service"
)

func main() {
	c, err := client.New(client.Options{
		APIKey:             os.Getenv("TRANSLAAS_API_KEY"),
		BaseURL:            envOr("TRANSLAAS_BASE_URL", "https://api.translaas.local"),
		DefaultProjectID:   envOr("TRANSLAAS_DEFAULT_PROJECT", "my-project"),
		DefaultLanguage:    "en",
		CacheMode:          cache.ModeGroup,
		InsecureSkipVerify: true, // DEV-ONLY — local Traefik certs
	})
	if err != nil {
		log.Fatal(err)
	}

	svc, err := service.New(c)
	if err != nil {
		log.Fatal(err)
	}

	text, err := svc.T(context.Background(), "common", "welcome.message", service.WithLang("en"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

`DefaultProjectID` must be the project **slug** (from Admin → project details or the API-key validate response), not the display name. Local Docker dogfood: slug `translaassdksamples`, display name `translaas-sdk-samples`.

Omit `InsecureSkipVerify` (or set it `false`) against hosted APIs with publicly trusted certificates. For web apps, pass `service.Options{Resolver: …}` so Accept-Language middleware can resolve language without `WithLang`.

See also: [`examples/go/basic`](https://github.com/acuencadev/translaas-all/tree/main/examples/go/basic).

### Option B — `client.Client` (direct API)

```go
text, err := c.GetEntry(ctx, "ui", "button.save", "en")
// 200 → plain text body; 204 → returns entry key unchanged

group, err := c.GetGroup(ctx, "my-project", "ui", "en")
// JSON payload; 204 → empty group (not an error)
```

Additional options: `client.WithNumber`, `client.WithParameters`, `client.WithRequestContext`. When `DefaultProjectID` is empty on a tenant-scoped key, use `client.NewWithResolvedProject`.

The text endpoint returns **plain text** (`Accept: text/plain`), not a JSON wrapper.

## Configuration

| Field / env | Notes |
|-------------|-------|
| `Options.APIKey` / `TRANSLAAS_API_KEY` | Required for live API (except `FallbackCacheOnly` after cache is seeded) |
| `Options.BaseURL` / `TRANSLAAS_BASE_URL` | **Origin only** — do not append `/api` or `/sdk`. Local Docker: `https://api.translaas.local` |
| `Options.DefaultProjectID` | Project **slug** for text endpoint and offline entry lookups |
| `Options.DefaultLanguage` | Optional fallback for `service.T` when `WithLang` is omitted and no resolver yields a language |
| `Options.InsecureSkipVerify` | **Dev-only.** Skip TLS verify on the built-in HTTP client (local Traefik). Ignored when `WithHTTPClient` is set. Default `false` |
| `Options.Timeout` | Default 30s; deadline/transport timeout maps to `*models.APIError` with status **408** |
| `Options.CacheMode` | `cache.ModeNone` … `ModeProject`; **recommend `ModeGroup`** |

See [API Keys & Authentication](https://github.com/acuencadev/translaas-all/blob/main/.docs/kb/auth-api-keys.md).

## In-memory caching

| `cache.Mode` | Behavior |
|--------------|----------|
| `ModeNone` | Every call hits the API |
| `ModeEntry` | Cache individual `GetEntry` results |
| `ModeGroup` | Cache group payloads (**recommended**) |
| `ModeProject` | Cache full project payloads |

`GetProjectLocales` is cached whenever mode ≠ `ModeNone`.

Details: [SDK caching KB](https://github.com/acuencadev/translaas-all/blob/main/.docs/kb/sdk-caching-resiliency.md).

## Offline / file cache

Wrap the HTTP client with `cachefile.CachingClient`:

```go
fileProvider, err := cachefile.NewFileProvider(".translaas-cache")
hybrid, err := cachefile.NewHybridProvider(fileProvider, cachefile.DefaultHybridOptions())
inner, _ := client.New(/* … */)
cached, err := cachefile.NewCachingClient(inner, hybrid, cachefile.Options{
	FallbackMode:     cachefile.FallbackCacheFirst,
	DefaultProjectID: "my-project",
})
```

| `FallbackMode` | Order |
|----------------|-------|
| `FallbackCacheFirst` | Disk → API on miss |
| `FallbackAPIFirst` | API → disk on network/API errors |
| `FallbackCacheOnly` | Disk only |

Use `cachefile.NewSyncService` to populate the on-disk cache. Optional background sync: `StartBackgroundSync`. To seed the cache from a downloaded offline ZIP (CDN, air-gapped bundle, or `GetOfflineCache`):

```go
result, err := client.GetOfflineCache(ctx, "my-project")
fileProvider, err := cachefile.NewFileProvider(".translaas-cache")
err = fileProvider.ImportOfflineBundle(ctx, "my-project", result.Content)

// Or combine download + import:
syncSvc, err := cachefile.NewSyncService(innerClient, fileProvider, cachefile.DefaultOfflineCacheOptions(), cachefile.SyncCallbacks{})
err = syncSvc.SyncFromOfflineZip(ctx, "my-project")
```

**Offline pluralization:** cache reads select CLDR cardinal forms (`zero` / `one` / `two` / `few` / `many` / `other`) for the request `lang` via `golang.org/x/text`. The locale tag matters (`pt` vs `pt-PT`). Live `GetEntry` still sends `n` and lets the server select.

Runnable sample: [`examples/go/offline`](https://github.com/acuencadev/translaas-all/tree/main/examples/go/offline).

## Error handling

Use `errors.As` for typed errors:

| Type | When |
|------|------|
| `*models.APIError` | HTTP 4xx/5xx from the API; timeout → status **408** |
| `*models.TransportError` | Connect / TLS / DNS failures (not an HTTP status; `Unwrap` for the cause) |
| `*models.ConfigurationError` | Invalid `client.New` options |
| `*models.OfflineCacheError` | Corrupt or unreadable disk cache |
| `*models.OfflineCacheMissError` | `FallbackCacheOnly` miss |
| `models.ErrNoLanguage` | No `WithLang`, resolver, or `DefaultLanguage` |

```go
var apiErr *models.APIError
if errors.As(err, &apiErr) {
	// apiErr.StatusCode, apiErr.Code, apiErr.Message
}
```

## Web frameworks

Optional middleware and helpers for `net/http`, Gin, Echo, and chi. Translation strings are **not** HTML-escaped — escape in templates.

Samples: [`examples/go/nethttp`](https://github.com/acuencadev/translaas-all/tree/main/examples/go/nethttp), [`gin`](https://github.com/acuencadev/translaas-all/tree/main/examples/go/gin), [`echo`](https://github.com/acuencadev/translaas-all/tree/main/examples/go/echo), [`chi`](https://github.com/acuencadev/translaas-all/tree/main/examples/go/chi).

## Compatibility

| Go SDK | .NET SDK | Delivery API | Notes |
|--------|----------|--------------|-------|
| `v0.5.0` (current) | `0.5.0` | `/sdk/v1` + `/api/v1/validate` | CLDR offline plurals, transport errors, convenience API |
| `v0.4.0-beta` | `v0.4.1-beta` | `/sdk/v1` + `/api/v1/validate` | M4 parity: client, cache, offline, `T()`, web |
| `v0.3.0-beta` | — | same | Offline + sync |
| `v0.2.0-beta` | — | same | In-memory `CacheMode` |
| `v0.1.0-alpha` | — | same | Read-only client |

**Known divergences:** no built-in retry policy in Go v1; text endpoint returns plain text (not JSON).

## Development

```bash
make help              # list targets
make tidy test lint
make test-integration  # live API (requires TRANSLAAS_API_KEY)
make coverage
```

Integration tests: [`tests/integration/README.md`](./tests/integration/README.md).

### CI

GitHub Actions on every push/PR to `main` (`.github/workflows/ci.yml`):

- `golangci-lint`
- `go vet`, `go test` (with `-race`)
- `web/gin`, `web/echo`, `web/chi` module tests
- `go build` for all library packages
- **Ubuntu only** on push/PR (cross-SDK policy to limit Actions spend)

Optional manual workflows:

- **Actions → CI → Run workflow** — Windows and macOS vet/test/build
- **Actions → Integration Tests → Run workflow** — live API smoke tests (see `.github/workflows/integration.yml`)

### Releasing (maintainers)

Tag-driven releases use `.github/workflows/release.yml` — the same quality bar as CI, plus a GitHub Release whose body is extracted from `CHANGELOG.md`.

1. Merge PRs with user-visible notes under `[Unreleased]`.
2. Promote `[Unreleased]` into `## [x.y.z] - YYYY-MM-DD` in `CHANGELOG.md`.
3. Merge to `main` and confirm CI is green.
4. Optionally run live integration tests via **Actions → Integration Tests → Run workflow**.
5. Create and push the tag (triggers the release workflow):

   ```bash
   bash scripts/create-release-tag.sh 0.5.0
   # or: bash scripts/create-release-tag.sh --dry-run
   ```

6. Verify the [GitHub Release](https://github.com/Mantelabs/translaas-sdk-go/releases) notes and module proxy indexing (`go get github.com/Mantelabs/translaas-sdk-go@v0.5.0`).

See [CONTRIBUTING.md](./CONTRIBUTING.md#releasing) for the full checklist.

## Documentation

- [Go SDK getting started](https://github.com/Mantelabs/translaas-all/blob/main/.docs/sdk/go.md)
- [Implementation plan](https://github.com/acuencadev/translaas-all/blob/main/.docs/translaas-sdk-go-implementation.md)
- [HTTP API spec](https://github.com/acuencadev/translaas-all/blob/main/.docs/translaas-sdk-http-api-spec.md)
- [Porting reference](https://github.com/acuencadev/translaas-all/blob/main/.docs/translaas-sdk-dotnet-porting-reference.md)

## Package layout

```
models/       DTOs, errors, request context
cache/        In-memory caching
client/       HTTP client
cachefile/    Offline disk cache
service/      Convenience API
web/          Optional framework integrations
internal/     Non-exported helpers
testdata/     Golden JSON fixtures
tests/        Integration tests (build tag: integration)
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). MIT licensed — see [LICENSE](./LICENSE).

## Related SDKs

- [.NET](https://github.com/acuencadev/Translaas.SDK)
- [Python](https://github.com/acuencadev/translaas-sdk-python)
- [JavaScript](https://github.com/acuencadev/translaas-sdk-js)
