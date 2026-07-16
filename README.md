# Translaas SDK for Go

Official [Translaas](https://github.com/acuencadev/translaas-all) client SDK for Go. Implementation in progress.

Part of the [translaas-all](https://github.com/acuencadev/translaas-all) umbrella workspace.

## Status

Early development. The `models` package (DTOs, errors, request context), `internal/httpx` (URL/query helpers), and `client.GetEntry` are implemented; remaining client methods are next. Track progress via [GitHub Issues](https://github.com/acuencadev/translaas-sdk-go/issues) and milestones **M0–M5**.

## Implementation plan

Phased roadmap aligned to the .NET reference SDK (`Translaas.SDK`):

- [translaas-sdk-go-implementation.md](https://github.com/acuencadev/translaas-all/blob/main/.docs/translaas-sdk-go-implementation.md)
- [translaas-sdk-dotnet-porting-reference.md](https://github.com/acuencadev/translaas-all/blob/main/.docs/translaas-sdk-dotnet-porting-reference.md)
- [translaas-sdk-http-api-spec.md](https://github.com/acuencadev/translaas-all/blob/main/.docs/translaas-sdk-http-api-spec.md)

## Requirements

- Go **1.22+** (module declares `toolchain go1.23.x`)

## Installation

```bash
go get github.com/acuencadev/translaas-sdk-go@latest
```

> No stable release yet. Pin to a commit or pre-release tag once published.

## Development

```bash
make help          # list targets
make tidy test lint
make coverage      # local coverage report
```

Runnable samples: **[translaas-sdk-examples](https://github.com/acuencadev/translaas-sdk-examples)** (`go/`).

### CI

GitHub Actions runs on every push/PR to `main`:

- `golangci-lint`
- `go vet`, `go test` (with `-race` on Linux)
- `go build` for all library packages
- Matrix: **Ubuntu** and **Windows**

## Package layout

```
models/       DTOs, errors, request context
cache/        In-memory caching
client/       HTTP client
cachefile/    Offline disk cache
service/      Convenience API
internal/     Non-exported helpers
testdata/     Golden JSON fixtures for unit tests
```

Sample apps: [translaas-sdk-examples/go](https://github.com/acuencadev/translaas-sdk-examples/tree/main/go).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). MIT licensed — see [LICENSE](./LICENSE).

## Related SDKs

- [.NET](https://github.com/acuencadev/Translaas.SDK)
- [Python](https://github.com/acuencadev/translaas-sdk-python)
- [JavaScript](https://github.com/acuencadev/translaas-sdk-js)
