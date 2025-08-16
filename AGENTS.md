# Repository Guidelines

## Project Structure & Modules
- `agg.go`, `types.go`: Public API surface.
- `internal/…`: All implementation packages (e.g., `basics`, `color`, `buffer`, `pixfmt`, `scanline`, `rasterizer`, `renderer`).
- `examples/…`: Runnable examples (e.g., `examples/basic/hello_world`).
- `tests/{unit,integration,benchmark,visual}`: Test suites by type.
- `docs/…`: Architecture and status docs.

## Build, Test, and Dev Commands
- `just build`: Build library and examples.
- `just test` | `just test-unit` | `just test-integration`: Run tests.
- `just test-coverage`: Generate `coverage.html` report.
- `just run-example EXAMPLE`: Run `examples/EXAMPLE/main.go` (e.g., `basic/hello_world`).
- `just fmt` | `just vet` | `just lint` | `just tidy` | `just check`: Format, vet, lint, tidy, then run tests.
- `just init-hooks`: Install a pre-commit hook to run `just pre-commit`.

## Coding Style & Naming
- Formatting: enforced by `gofumpt` + `gci` via `treefmt`; run `just lint` or `treefmt --fail-on-change`.
- Linting: `golangci-lint` (includes `misspell`, `gocritic`); keep code idiomatic and simple.
- Packages: short, lowercase names (no underscores).
- Files: `snake_case.go`; tests as `*_test.go`.
- Identifiers: Exported `PascalCase`, unexported `camelCase`. Avoid stutter in public API.

## Testing Guidelines
- Unit tests live beside code under `internal/<pkg>` or under `tests/unit` when cross-package.
- Integration/visual/bench tests under `tests/{integration,visual,benchmark}`.
- Use standard `testing` (`TestXxx`, `BenchmarkXxx`). Run `just test-coverage` to inspect coverage.
- Prefer deterministic tests; isolate state and avoid global mutation.

## Commit & Pull Request Guidelines
- Commits: short, imperative subject (e.g., "rasterizer: fix cell merge"). Git history favors concise messages.
- PRs: focused scope; include description, linked issue/TASK.md items, and before/after images for visual changes.
- CI locally: ensure `just check` passes; update docs/examples when API changes.

## Security & Configuration Tips
- No manual memory management; avoid unsafe unless justified and reviewed.
- Run `just tidy` to keep `go.mod` clean. Use `build-*-platforms` only when needed.
