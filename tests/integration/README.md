# Translaas Go SDK — integration tests

Live API integration tests for `github.com/Mantelabs/translaas-sdk-go`. They mirror [.NET `Translaas.Client.IntegrationTests`](https://github.com/acuencadev/Translaas.SDK/tree/main/tests/Translaas.Client.IntegrationTests).

## Prerequisites

- A running Translaas delivery API (development environment)
- Valid API key with access to fixture project data

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TRANSLAAS_API_KEY` | **Yes** to run | — | Raw `X-Api-Key` value |
| `TRANSLAAS_BASE_URL` | No | `https://api.translaas.local` | API origin only (no `/api` or `/sdk` suffix) |
| `TRANSLAAS_DEFAULT_PROJECT` | No | `translaas-sdk-samples` | Default project for `GetEntry` and scoped reads |

When `TRANSLAAS_API_KEY` is unset, tests are **skipped** (not failed).

## Local Docker (`platform/translaas`)

Local Compose exposes one API origin for Admin (`/api/v1/...`) and SDK (`/sdk/v1/...`) routes. Use **`https://api.translaas.local`** (same as `TRANSLAAS_BASE_URL` in platform `.env.example`).

```powershell
# After: docker compose --profile core up -d
$env:TRANSLAAS_API_KEY = "<your-sdk-api-key>"
make test-integration
```

## Fixture data

Canonical strings live in [translaas-sdk-examples `translaas_sdk_samples_strings.csv`](https://github.com/Mantelabs/translaas-sdk-examples/blob/main/dotnet/docs/translaas_sdk_samples_strings.csv). Live tests default to:

| Field | Value |
|-------|-------|
| Project | `translaas-sdk-samples` |
| Group (simple entry) | `common` |
| Entry (simple) | `welcome.message` |
| Group (plural) | `messages` |
| Entry (plural) | `item` |
| Language | `en` (optional: `fr`, `es`) |

Example SDK URL:

`GET /sdk/v1/translations/text?project=translaas-sdk-samples&group=common&lang=en&entry=welcome.message`

Constants are centralized in `config_test.go`. Override the project with `TRANSLAAS_DEFAULT_PROJECT` when your API uses a different project id.

Tests that require populated payloads **soft-skip** when the API returns empty containers, when the entry key is returned as fallback (missing fixture data), or when the Mantelabs platform returns **HTTP 404** for a missing project (wrong `TRANSLAAS_DEFAULT_PROJECT`).

### API behavior

| Endpoint | Missing resource | Go/.NET fixture API | Mantelabs platform | Integration test |
|----------|------------------|---------------------|--------------------|--------------------|
| `GetEntry` | entry/group | **204** → entry key fallback | **404** ProblemDetails | Accept fallback **or** 404 |
| `GetGroup` / `GetProject` / `GetProjectLocales` | project/group | **204** → empty container | **404** ProblemDetails | Accept empty **or** 404 |
| Invalid API key | — | **401/403** | **401/403** | `*models.APIError` |

Happy-path tests **soft-skip** on HTTP 404 when the configured project is missing, with a hint to set `TRANSLAAS_DEFAULT_PROJECT` (default: `translaas-sdk-samples`).

## Running locally

### Linux / macOS

```bash
export TRANSLAAS_API_KEY="your-api-key"
export TRANSLAAS_BASE_URL="https://api.translaas.local"   # optional
export TRANSLAAS_DEFAULT_PROJECT="translaas-sdk-samples"  # optional
make test-integration
```

### Windows (PowerShell)

```powershell
$env:TRANSLAAS_API_KEY = "your-api-key"
$env:TRANSLAAS_BASE_URL = "https://api.translaas.local"   # optional
$env:TRANSLAAS_DEFAULT_PROJECT = "translaas-sdk-samples"  # optional
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
