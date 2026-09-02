# Repository Guidelines

## Project Structure & Module Organization

Executable entry points live in `cmd/shortener` (server), `cmd/client`, and `cmd/staticlint` (custom multichecker). Application code is under `internal/app`: HTTP routing and middleware are in `server`, persistence implementations are in `store/{memory,file,database}`, and configuration, logging, and models have dedicated packages. Reusable helpers belong in `pkg`; test data belongs beside its package under `fixtures`. PostgreSQL migrations are timestamped SQL files in `internal/app/store/database/migrations`.

## Build, Test, and Development Commands

- `make build` builds `cmd/shortener/shortener`.
- `make run` starts the built server; `make run_with_file`, `make run_with_db`, and `make run_with_config` select storage/configuration modes.
- `go test ./...` runs the unit and example tests without requiring PostgreSQL.
- `make test_with_db` runs all Go tests with the local PostgreSQL DSN and coverage reporting.
- `make lint` runs the configured `golangci-lint` checks.
- `make build_checker && make check` builds and runs the repository's custom static analyzer.

Start PostgreSQL as documented in `README.md` before database tests. `make test` also requires the bundled autotest binary and a built server.

## Coding Style & Naming Conventions

Target Go 1.26 or later. Format changed files with `gofmt` or `goimports`. Use tabs, short lowercase package names, PascalCase for exported identifiers, and camelCase for unexported identifiers. Document public declarations and follow existing error-wrapping patterns. `.golangci.yaml` enables `errcheck`, `govet`, `staticcheck`, `unused`, and related checks.

## Testing Guidelines

Place tests beside production code in `*_test.go`; name functions `TestXxx` and examples `ExampleXxx`. Favor table-driven subtests and use `testify` where helpful. Cover every behavior change, including the relevant storage backend. Database tests use `DATABASE_DSN`; CI provides PostgreSQL and uploads coverage to Codecov. No fixed threshold is declared, so avoid reducing meaningful coverage.

## Commit & Pull Request Guidelines

Recent commits use short, imperative, sentence-style subjects such as `Updated README` and `Fix for bool flag`. Keep each commit focused. CI requires feature branches named `iter<number>` (for example, `iter18`); `main` runs the complete suite. Pull requests should explain the behavior change, note configuration or migration impacts, link the issue when applicable, and include commands/results used for verification. Include screenshots only for user-visible output or documentation rendering changes.

## Security & Configuration

Do not commit real DSNs, JWT secrets, generated storage files, or coverage artifacts. Copy values from `config.example.json` and override them through environment variables; configuration precedence is environment, flags, then config file.
