# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `cachefile.HybridProvider`: expirable LRU memory L1 over file L2 with promotion, write-through, and warmup helpers.
- Repository foundation: module layout, CI, linting, Makefile, and contributor docs.
- `models` package: DTOs, typed errors, request context, translation payloads, and golden test fixtures.
- `internal/httpx` package: URL builder, query reflection, plural `N` injection, and merge helpers (.NET parity).
- `client` package: options validation, HTTP transport, and `GetEntry` for the text translation endpoint.
- Documented that runnable samples belong in [translaas-sdk-examples](https://github.com/acuencadev/translaas-sdk-examples) (`go/`), not in this library repo.

### Removed

- In-repo `examples/basic` placeholder (use translaas-sdk-examples instead).
