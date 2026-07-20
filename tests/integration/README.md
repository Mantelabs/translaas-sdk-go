# Translaas Go SDK — integration tests

Live API integration tests for `github.com/acuencadev/translaas-sdk-go`. They mirror [.NET `Translaas.Client.IntegrationTests`](https://github.com/acuencadev/Translaas.SDK/tree/main/tests/Translaas.Client.IntegrationTests).

## Prerequisites

- A running Translaas delivery API (development environment)
- Valid API key with access to fixture project data

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TRANSLAAS_API_KEY` | **Yes** to run | — | Raw `X-Api-Key` value |
| `TRANSLAAS_BASE_URL` | No | `https://sdk-api.translaas.local` | API origin only (no `/api` or `/sdk` suffix) |
| `TRANSLAAS_DEFAULT_PROJECT` | No | `test-project` | Default project for `GetEntry` and scoped reads |

When `TRANSLAAS_API_KEY` is unset, tests are **skipped** (not failed).

## Fixture data

Tests expect the following data in your development API (same as the .NET suite):

| Field | Value |
|-------|-------|
| Project | `test-project` |
| Group | `ui` |
| Entries | `button.save`, `button.cancel`, `items.count` |
| Language | `en` (optional: `fr`, `es`, `de`) |

Tests that require populated payloads **soft-skip** when the API returns empty containers (204 semantics).

### API behavior

| Endpoint | Missing resource | Client behavior |
|----------|------------------|-----------------|
| `GetEntry` | 204 | Returns the **entry key** as fallback |
| `GetGroup` / `GetProject` / `GetProjectLocales` | 204 | Returns an **empty container** (not an error) |
| Invalid API key | 401/403 | `*models.APIError` |

## Running locally

### Linux / macOS

```bash
export TRANSLAAS_API_KEY="your-api-key"
export TRANSLAAS_BASE_URL="https://api-dev.example.com"   # optional
make test-integration
```

### Windows (PowerShell)

```powershell
$env:TRANSLAAS_API_KEY = "your-api-key"
$env:TRANSLAAS_BASE_URL = "https://api-dev.example.com"   # optional
make test-integration
```

### Direct `go test`

```bash
go test -tags=integration -count=1 -timeout 5m ./tests/integration/...
```

Skip even when credentials are set (short mode):

```bash
go test -tags=integration -short ./tests/integration/...
```

Default `make test` and CI **do not** pass `-tags=integration`.

## CI (optional)

The repository includes `.github/workflows/integration.yml` for manual runs via **workflow_dispatch**. Configure these secrets on the repo:

- `TRANSLAAS_API_KEY`
- `TRANSLAAS_BASE_URL` (optional)

PR CI does **not** require integration secrets.

## Security

Never commit API keys. Use environment variables or CI secrets only.
