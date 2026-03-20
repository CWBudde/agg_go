## Project Overview

This repository is a Go port of the Anti-Grain Geometry (AGG) 2.6 C++ library, a high-quality 2D graphics rendering library with sub-pixel accuracy and anti-aliasing.

The aim is to provide a close parity implementation in Go, preserving AGG's design and performance characteristics while adhering to Go idioms and best practices. The project includes a public API for users, internal packages for implementation details, examples, tests, and documentation.

## Architecture and Rendering Pipeline

The codebase is organized around a rendering pipeline similar to the original AGG design:

1. Path definition in `types.go` and related root-level API files.
2. Transformation in `internal/transform/`.
3. Path conversion in `internal/conv/`.
4. Rasterization in `internal/rasterizer/`.
5. Scanline generation in `internal/scanline/`.
6. Final pixel rendering in `internal/renderer/` and `internal/pixfmt/`.

### Public API Design

- Root-level files such as `agg.go`, `agg2d.go`, `types.go`, and `context.go` form the user-facing API.
- Internal complexity belongs under `internal/` and should stay hidden from consumers.
- Prefer keeping the public API clean and user-oriented even if internal implementation remains close to the C++ source.

## Repository Structure

- Public API: root-level files such as `agg.go`, `agg2d.go`, `types.go`, `context.go`.
- Internal implementation: `internal/<pkg>/`.
- Examples: `examples/<group>/<name>/`.
- Commands and demos: `cmd/` and `bin/`.
- Tests: `tests/{unit,integration,benchmark,visual}` and package-local `*_test.go` files.
- Documentation: `docs/`.
- Web and wasm support: `web/`, `wasm/`.

Important internal packages include `array`, `basics`, `buffer`, `color`, `conv`, `curves`, `font`, `path`, `pixfmt`, `rasterizer`, `renderer`, `scanline`, `span`, `transform`, `vcgen`, and `vpgen`.

## Development Commands

This repository uses `just` for common development workflows.

- `just --list`: show available commands.
- `just quick`: fast feedback, typically formatting and vetting. Run this often.
- `just check`: full validation including formatting, vet, lint, tidy, and tests.
- `just build`: build library and examples.
- `just build-lib`: build the library only.
- `just test`: run unit and integration tests.
- `just test-unit`: run unit tests only.
- `just test-integration`: run integration tests only.
- `just test-coverage`: generate coverage output.
- `just fmt`: format Go code.
- `just lint`: run `golangci-lint`.
- `just tidy`: clean up module dependencies.
- `just clean`: remove generated build artifacts.
- `just run-hello`: run the hello world example.
- `just run-example basic/hello_world`: run a specific example.
- `just dev`: watch files and run checks.
- `just ci`: run the fuller CI-style validation flow.
- `just docs`: generate API documentation.
- `just stats`: show project statistics.
- `just todo`: find TODO and FIXME markers.
- `just init-hooks`: install the pre-commit hook.

### Common Workflows

- Fast development loop: `just quick && just test-unit`.
- Pre-commit validation: `just check`.
- Coverage analysis: `just test-coverage`.
- Example verification: `just build` and `just run-hello` or `just run-example <group>/<name>`.

## Coding Style and Naming

- Use idiomatic Go and keep implementations simple.
- Prefer explicit interfaces over `interface{}` or `any`.
- Use constrained generics when they clearly preserve AGG template structure or remove duplication without obscuring hot paths.
- Prefer concrete types or concrete aliases for the public API, for common instantiations, and for performance-critical code unless generics are clearly justified.
- Packages should be short, lowercase, and use no underscores.
- Files should use `snake_case.go` naming.
- Exported identifiers use `PascalCase`; unexported identifiers use `camelCase`.
- Avoid API stutter in exported names.

### Formatting and Linting

- Formatting is enforced with `gofumpt` and `gci`, typically through `treefmt`.
- Use `just fmt`, `just lint`, or `treefmt --fail-on-change` as appropriate.
- Linting uses `golangci-lint` and includes checks such as `misspell` and `gocritic`.

### File Naming Scheme

Follow the original AGG source naming where it helps traceability, but adapt to Go package conventions:

- C++ headers map naturally into package and file names, for example `rasterizer_scanline_aa.h` to `internal/rasterizer/scanline_aa.go`.
- Do not repeat package names in file names unnecessarily.
- Keep filenames descriptive, but rely on the package directory to provide namespace context.

## C++ to Go Translation Strategy

When implementing or fixing algorithms, prefer parity with the original AGG source located under `../agg-2.6/agg-src/`.

- C++ templates become Go generics or concrete types as appropriate.
- C++ inheritance and virtual dispatch become Go interfaces.
- C++ enums become typed Go constants.
- Manual memory management becomes Go slice- and GC-based ownership.

Choose between generics and concrete types pragmatically:

- Prefer generics for internal containers, algorithms, and compile-time type markers that directly model AGG templates.
- Prefer concrete types or aliases at the root API boundary and for the most common internal instantiations.
- Do not keep a generic design by default in a measured hotspot; simplify to concrete types when profiling shows that helps code generation, readability, or maintenance.
- Avoid premature de-genericization when there is no demonstrated performance problem.

Translation should stay as close as practical to the original implementation unless there is a clear Go-specific improvement or repository convention that requires a different approach.

### Template Translation Examples

- `pod_array<T>` to `PodArray[T]`.
- `rgba8T<Colorspace>` to a generic internal type such as `RGBA8[CS]`, with concrete aliases where usage is repetitive or performance-sensitive.
- `point_base<double>` or `rect_base<int>` to `PointD`, `RectI`, or the underlying generic `Point[T]` and `Rect[T]` as appropriate.
- AGG enum types to typed Go constants with methods where appropriate.

## Development Patterns

### Adding or Modifying Internal Components

1. Keep new implementation under the correct `internal/<pkg>/` package.
2. Start with the simplest faithful shape: use generics where they preserve template structure, but prefer concrete types or aliases when a type has only a few real instantiations or sits on a hot path.
3. Follow Go naming conventions instead of mechanically copying C++ identifiers.
4. Keep public API exposure minimal unless the change is explicitly part of the user-facing surface.

### Memory and Performance

- Use slices instead of raw-pointer style structures.
- Reuse buffers and slices where practical, especially for scanlines, spans, and rendering buffers.
- Avoid `unsafe` unless there is a strong justification and the code is reviewed carefully.
- Benchmark before replacing generics with concrete types in the name of performance; keep the simpler design unless a real hotspot is demonstrated.
- Preserve AGG's performance-sensitive behavior without sacrificing Go safety unless required.

### Error Handling

- Return `error` for invalid input, I/O issues, and recoverable failures.
- Use `panic` for programmer errors or invalid internal state when that matches existing project conventions.
- Handle edge cases gracefully in rendering code when possible.

## Testing Guidelines

- Use the standard Go `testing` package.
- Keep unit tests close to the code in `*_test.go` where appropriate.
- Use `tests/unit`, `tests/integration`, `tests/visual`, and `tests/benchmark` for broader suites.
- Prefer deterministic tests with isolated state.
- Run `just test-coverage` when investigating gaps or validating larger changes.

## Commit and Pull Request Guidance

- Use short, imperative commit subjects, for example `rasterizer: fix cell merge`.
- Keep pull requests focused.
- Include linked issues or task references where relevant.
- Include before-and-after images for visual rendering changes when useful.
- Run `just check` before opening or merging substantial changes.
- Update documentation and examples when public behavior changes.

## Task Tracking and Documentation

The repository follows the porting plan documented in `docs/TASKS.md`.

- Mark completed tasks as done with `[x]` when work is finished.
- Keep docs aligned with implementation changes.
- Prefer documenting architectural or parity-relevant deviations when behavior intentionally differs from AGG.

## Practical Guidance for Agents and Contributors

- Build context from existing code before making structural changes.
- Prefer code navigation and symbol-aware tooling when available, especially in larger refactors.
- When in doubt about algorithmic behavior, consult the original C++ implementation first.
- Fix root causes rather than layering superficial patches over rendering bugs.
