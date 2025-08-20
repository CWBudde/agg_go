# GEMINI.md

## Project Overview

This project is a Go port of the Anti-Grain Geometry (AGG) library, a high-quality 2D graphics library. The goal of this project is to provide a minimal, idiomatic Go implementation of AGG 2.6, maintaining the core functionality and API structure of the original C++ codebase.

The main package `agg` provides the primary API for creating and drawing on a rendering context. The core data structures include `Context`, `Image`, `Path`, and various `Color` types. The internal packages, such as `internal/basics` and `internal/buffer`, provide the low-level implementations for the public API.

## Building and Running

The project uses a `Justfile` for build orchestration.

### Building

-   **Build the library:** `just build-lib`
-   **Build all examples:** `just build-examples`
-   **Build a specific example:** `just build-example EXAMPLE=<example_name>`

### Running

-   **Run a specific example:** `just run-example EXAMPLE=<example_name>`
-   **Run the "hello world" example:** `just run-hello`

### Testing

-   **Run all tests:** `just test`
-   **Run unit tests:** `just test-unit`
-   **Run integration tests:** `just test-integration`
-   **Run benchmark tests:** `just test-bench`
-   **Run visual regression tests:** `just test-visual`
-   **Run tests with coverage:** `just test-coverage`

## Development Conventions

### Formatting

-   **Format all Go code:** `just fmt`

### Linting

-   **Run linters:** `just lint`
-   **Fix linting issues:** `just lint-fix`

### Dependencies

-   **Tidy dependencies:** `just tidy`

### All Checks

-   **Run all checks (fmt, vet, lint, test):** `just check`
