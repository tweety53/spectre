# Agent instructions

## Go skills — apply on every change touching Go code

Invoke and apply the applicable skills: `go-code-review`, `go-coding-standards`,
`go-error-handling`, `go-documentation`, `go-test-quality`, `go-test-table-driven`,
`go-refactoring`, `go-architecture-review`, `go-data-structures`, `go-interface-design`,
`go-semantic-tools`, `go-security-audit`, `go-dependency-audit`, `go-project-layout`, `go-cli`,
`go-performance-review`, `go-troubleshooting`, `go-modernize`, `modern-go-guidelines:use-modern-go`.

## Skills that do not apply here, and why

- `go-concurrency-review`, `go-context` — this codebase has no goroutines and no `context.Context`.
- `go-api-design`, `go-grpc`, `go-database`, `go-observability`, `go-dependency-injection`,
  `go-design-patterns` — no HTTP, RPC, database, logging or DI surface.
- `go-ci` — no CI configuration exists in this repository.

A skill is skipped by a stated judgement, never silently. When the codebase later grows one of these
surfaces, delete the corresponding line rather than ignoring the rule.

## Project facts

- Go version: this module targets the version pinned in `go.mod` (`go 1.26.5`).
- Standard library only — no external dependencies (`go.mod` has no `require` block).
- Checks, all of which must be clean before a change is done:

  ```bash
  gofmt -l .
  go vet ./...
  go test ./...
  ```

- Exit codes are a contract: `0` success, `1` findings or a content refusal, `2` usage or IO error.
