# Repository Guidelines

This document describes how to work in this repository efficiently and consistently.

## Project Structure & Module Organization

- Public API: `agg.go`, `types.go`.
- Implementation: `internal/<pkg>/` (e.g., `basics`, `color`, `buffer`, `pixfmt`, `scanline`, `rasterizer`, `renderer`).
- Examples: `examples/<group>/<name>/` (e.g., `examples/basic/hello_world`).
- Tests: `tests/{unit,integration,benchmark,visual}`.
- Docs: `docs/` for architecture and status notes.

## Build, Test, and Development Commands

- Build library and examples: `just build`.
- Run tests: `just test` | unit only: `just test-unit` | integration: `just test-integration`.
- Coverage report: `just test-coverage` (opens/produces `coverage.html`).
- Run an example: `just run-example basic/hello_world`.
- Quality checks: `just fmt` | `just vet` | `just lint` | `just tidy` | `just check`.
- Install pre-commit hook: `just init-hooks` (runs `just pre-commit` on commit).

## Coding Style & Naming Conventions

- Formatting: `gofumpt` + `gci` via `treefmt`. Enforce with `just lint` or `treefmt --fail-on-change`.
- Linting: `golangci-lint` (includes `misspell`, `gocritic`). Keep code idiomatic and simple.
- Packages: short, lowercase, no underscores.
- Files: `snake_case.go`; tests as `*_test.go`.
- Identifiers: Exported `PascalCase`, unexported `camelCase`. Avoid API stutter.

## Testing Guidelines

- Framework: standard `testing` (`TestXxx`, `BenchmarkXxx`).
- Location: unit tests beside code (`internal/<pkg>`) or cross-package in `tests/unit`.
- Other suites: `tests/{integration,visual,benchmark}`.
- Practices: deterministic tests; isolate state; avoid global mutation.
- Coverage: `just test-coverage` then open `coverage.html`.

## Commit & Pull Request Guidelines

- Commits: short, imperative subject (e.g., `rasterizer: fix cell merge`).
- PRs: focused scope; include description, linked issue (and `TASKS.md` items when relevant), and before/after images for visual changes.
- CI locally: ensure `just check` passes; update docs/examples when API changes.

## Security & Configuration Tips

- No manual memory management; avoid `unsafe` unless justified and reviewed.
- Keep `go.mod` tidy with `just tidy`.
- Use platform-specific builds only when necessary (`build-*-platforms`).
