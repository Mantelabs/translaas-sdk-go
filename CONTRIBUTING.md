# Contributing to Translaas SDK for Go

Thank you for contributing to the Translaas Go SDK. This repository implements the [Translaas SDK HTTP contract](https://github.com/acuencadev/translaas-all/blob/main/.docs/translaas-sdk-http-api-spec.md) with behavioral parity to the [.NET reference SDK](https://github.com/acuencadev/Translaas.SDK).

## Getting started

1. Fork and clone the repository.
2. Install **Go 1.22+** (CI uses 1.23.x).
3. Install [golangci-lint](https://golangci-lint.run/welcome/install/) for local linting.
4. Create a feature branch: `feature/short-description` or `fix/short-description`.

```bash
git checkout -b feature/my-change
make tidy test lint
```

## Repository layout

| Path | Purpose |
|------|---------|
| `models/` | DTOs, errors, request context |
| `cache/` | In-memory cache abstractions |
| `client/` | HTTP client |
| `cachefile/` | Offline file cache and decorator |
| `service/` | Convenience `T()` API and language resolution |
| `web/` | Optional framework integrations |
| `internal/httpx` | URL/query helpers (internal) |
| `internal/validate` | Options validation (internal) |
| `testdata/` | Golden JSON fixtures |

Runnable sample apps live in **[translaas-sdk-examples](https://github.com/acuencadev/translaas-sdk-examples)** (`go/`), not in this repository.

See the [implementation plan](https://github.com/acuencadev/translaas-all/blob/main/.docs/translaas-sdk-go-implementation.md) for the full roadmap.

## Development guidelines

### Test-driven development

- Write table-driven tests before or alongside implementation.
- Run `make test` (or `make test-race` on Linux/macOS with CGO enabled).
- Aim for high coverage on public APIs.

### Go conventions

- Accept `context.Context` as the first parameter on all I/O methods.
- Use functional options for client configuration.
- Export typed errors; support `errors.Is` / `errors.As`.
- Add godoc comments on every exported symbol.
- Prefer the standard library; justify new dependencies in PR descriptions.
- Run `make fmt vet lint` before pushing.

### Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation
- `test:` tests
- `chore:` tooling / maintenance
- `refactor:` code change without behavior change

Reference GitHub issues in the footer: `Closes #123`.

### Pull requests

- Keep PRs focused on a single issue or vertical slice.
- Link the tracking issue (`Closes #…`).
- Ensure CI passes (lint, test, build on Ubuntu and Windows).
- Update `CHANGELOG.md` under `[Unreleased]` for user-visible changes.

## Versioning

This module uses [Semantic Versioning](https://semver.org/). Consumers install via:

```bash
go get github.com/acuencadev/translaas-sdk-go@v0.x.x
```

Pre-release tags use `-alpha`, `-beta`, or `-rc` suffixes.
