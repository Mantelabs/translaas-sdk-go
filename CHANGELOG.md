# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `cachefile.ParseOfflineZip`, `ResolveProjectKey`, and `FileProvider.ImportOfflineBundle` for offline ZIP bundle import (HTTP spec §7.6).
- `cachefile.SyncService.SyncFromOfflineZip` to download and persist an offline ZIP in one call.

## [0.4.0-beta] - 2026-07-20

### Documentation

- Expanded README with installation, quickstart, caching, offline, error handling, compatibility matrix, and links to `examples/go/` sample apps.
- Added runnable `basic` and `offline` console samples under `translaas-all/examples/go/` (not in this library repo).
- Updated umbrella KB article `.docs/kb/sdk-go.md` to document the official SDK (plain-text text endpoint; removed incorrect interim JSON client example).

### Added

- Live API integration tests under `tests/integration/` (build tag `integration`, env-gated via `TRANSLAAS_API_KEY`).
- `make test-integration` target and optional `integration.yml` GitHub Actions workflow for manual runs.
- `web` package: stdlib `net/http` middleware, request language provider, and context helpers for request-scoped `service.Service`.
- Optional framework modules: `web/gin`, `web/echo`, and `web/chi` with separate `go.mod` files.
- `service.Service.WithPrependedProviders` and `language.Resolver.PrependProviders` for per-request language resolution.
- `service.Service` and `T()`: convenience translation API with functional options (`WithLang`, `WithNumber`, `WithParameters`, `WithRequestContext`).
- `service/language`: provider chain resolver with `DefaultLanguageProvider`, `AcceptLanguageProvider`, and `ContextLanguageProvider`.
- `cachefile.SyncService`: offline cache synchronization with optional background sync and event callbacks.
- `cachefile.OfflineCacheOptions`: configuration for offline cache directory, projects, languages, and auto-sync interval.
- `cachefile.CachingClient`: offline decorator with CacheFirst, APIFirst, and CacheOnly fallback modes.
- `cachefile.HybridProvider`: expirable LRU memory L1 over file L2 with promotion, write-through, and warmup helpers.
- Repository foundation: module layout, CI, linting, Makefile, and contributor docs.
- `models` package: DTOs, typed errors, request context, translation payloads, and golden test fixtures.
- `internal/httpx` package: URL builder, query reflection, plural `N` injection, and merge helpers (.NET parity).
- `client` package: options validation, HTTP transport, and `GetEntry` for the text translation endpoint.
- Documented that runnable samples belong in [translaas-sdk-examples](https://github.com/acuencadev/translaas-sdk-examples) (`go/`), not in this library repo.

### Removed

- In-repo `examples/basic` placeholder (use translaas-sdk-examples instead).
