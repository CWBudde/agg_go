# AGG Go Build Orchestration
# Run `just --list` to see all available commands

set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

# Default recipe - show available commands
default:
    @just --list

# Build commands

# Build everything (library + examples)
build: build-lib build-examples

# Build the library
build-lib:
    @echo "Building AGG Go library..."
    go build ./...

# Build all examples
build-examples:
    @echo "Building core examples..."
    @cd examples/core/basic/hello_world && go build -o /tmp/example-test . || echo "Failed: hello_world"
    @cd examples/core/basic/shapes && go build -o /tmp/example-test . || echo "Failed: shapes" 
    @cd examples/core/basic/lines && go build -o /tmp/example-test . || echo "Failed: lines"
    @cd examples/core/basic/rounded_rect && go build -o /tmp/example-test . || echo "Failed: rounded_rect"
    @cd examples/core/basic/colors_gray && go build -o /tmp/example-test . || echo "Failed: colors_gray"
    @cd examples/core/basic/colors_rgba && go build -o /tmp/example-test . || echo "Failed: colors_rgba"
    @cd examples/core/basic/embedded_fonts_hello && go build -o /tmp/example-test . || echo "Failed: embedded_fonts_hello"
    @cd examples/core/basic/basic_demo && go build -o /tmp/example-test . || echo "Failed: basic_demo"
    @cd examples/core/intermediate/gradients && go build -o /tmp/example-test . || echo "Failed: gradients"
    @cd examples/core/intermediate/text_rendering && go build -o /tmp/example-test . || echo "Failed: text_rendering"
    @cd examples/core/intermediate/controls/gamma_correction && go build -o /tmp/example-test . || echo "Failed: gamma_correction"
    @cd examples/core/intermediate/controls/slider_demo && go build -o /tmp/example-test . || echo "Failed: slider_demo"
    @cd examples/core/intermediate/controls/rbox_demo && go build -o /tmp/example-test . || echo "Failed: rbox_demo"
    @cd examples/core/intermediate/controls/spline_demo && go build -o /tmp/example-test . || echo "Failed: spline_demo"
    @cd examples/core/advanced/advanced_rendering && go build -o /tmp/example-test . || echo "Failed: advanced_rendering"
    @echo "Building platform examples (with build tags)..."
    @cd examples/platform/sdl2 && go build -tags sdl2 -o /tmp/example-test . || echo "SDL2 dependencies missing (optional)"
    @cd examples/platform/x11 && go build -tags x11 -o /tmp/example-test . || echo "X11 dependencies missing (optional)"
    @rm -f /tmp/example-test

# Build specific example
build-example EXAMPLE:
    @echo "Building example: {{EXAMPLE}}"
    go build -o /tmp/agg-example examples/{{EXAMPLE}}/main.go

# Test commands

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
    @echo "Running unit tests..."
    go test ./internal/...

# Run unit tests with coverage
test-coverage:
    @echo "Running tests with coverage..."
    go test -coverprofile=coverage.out ./internal/...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run integration tests
test-integration:
    @echo "Running integration tests..."
    go test ./tests/integration/...

# Run benchmark tests
test-bench:
    @echo "Running benchmark tests..."
    go test -bench=. ./tests/benchmark/...

# Run visual regression tests
test-visual:
    @echo "Running visual tests..."
    go test ./tests/visual/...

# Run freetype tests
test-freetype:
    @echo "Running freetype tests..."
    go test -v -tags=freetype ./...

# Development commands

# Format all Go code
fmt:
    @echo "Formatting Go code..."
    treefmt --allow-missing-formatter

# Run go vet
vet:
    @echo "Running go vet..."
    go vet ./...

# Run linters (golangci-lint and treefmt)
lint:
    @echo "Running golangci-lint..."
    golangci-lint run ./...
    @echo "Running treefmt..."
    treefmt --fail-on-change

# Fix linting issues automatically
lint-fix:
    @echo "Fixing golangci-lint issues..."
    golangci-lint run --fix ./...
    @echo "Formatting with treefmt..."
    treefmt

# Tidy dependencies
tidy:
    @echo "Tidying dependencies..."
    go mod tidy

# Run all checks (fmt, vet, lint, test)
check: fmt vet lint tidy test

# Clean commands

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    go clean ./...
    rm -f coverage.out coverage.html
    find . -name "*.test" -delete
    find . -name "*_test.exe" -delete

# Clean and rebuild everything
rebuild: clean build

# Run commands

# Run hello world example
run-hello:
    @echo "Running hello world example..."
    go run examples/core/basic/hello_world/main.go

# Run specific example
run-example EXAMPLE:
    @echo "Running example: {{EXAMPLE}}"
    go run examples/{{EXAMPLE}}/main.go

# Run basic example (alias for convenience)
run EXAMPLE:
    @echo "Running example: {{EXAMPLE}}"
    go run examples/core/basic/{{EXAMPLE}}/main.go

# Run all basic examples
run-examples-basic:
    @echo "Running all basic examples..."
    @GOCACHE=$PWD/.gocache sh -c 'root=$PWD; out=examples/core/basic/_out; \
        mkdir -p "$out"; \
        for d in $(find examples/core/basic -mindepth 1 -maxdepth 1 -type d | sort); do \
            if [ -f "$d/main.go" ]; then \
                name=$(basename "$d"); \
                echo "--- $name"; \
                ( cd "$out" && go run "$root/$d/main.go" ) || exit $?; \
            fi; \
        done'

# Run X11 demo (requires X11 headers and running X server)
run-x11-demo:
    @echo "Running X11 demo..."
    @echo "Note: Requires X11 headers and running X server"
    go run -tags x11 examples/platform/x11/main.go

# Run SDL2 demo (requires SDL2 dependencies)
run-sdl2-demo:
    @echo "Running SDL2 demo..."
    @echo "Note: Requires SDL2 development libraries"
    go run -tags sdl2 examples/platform/sdl2/main.go

# Run intermediate examples
run-examples-intermediate:
    @echo "Running intermediate examples..."
    @for dir in examples/core/intermediate/*/; do \
        if [ -f "$$dir/main.go" ]; then \
            echo "Running $$(basename $$dir)..."; \
            go run "$$dir/main.go" || echo "Failed: $$(basename $$dir)"; \
        fi \
    done

# Run advanced examples
run-examples-advanced:
    @echo "Running advanced examples..."
    @for dir in examples/core/advanced/*/; do \
        if [ -f "$$dir/main.go" ]; then \
            echo "Running $$(basename $$dir)..."; \
            go run "$$dir/main.go" || echo "Failed: $$(basename $$dir)"; \
        fi \
    done

# Run test examples (AGG 2.6 compatibility tests)
run-tests:
    @echo "Running test examples..."
    @echo "Note: Most test examples are not yet implemented"
    @for dir in examples/tests/*/; do \
        if [ -f "$$dir/main.go" ]; then \
            echo "Running test: $$(basename $$dir)..."; \
            go run "$$dir/main.go" || echo "Failed: $$(basename $$dir)"; \
        else \
            echo "Test not implemented: $$(basename $$dir)"; \
        fi \
    done

# Development workflow commands

# Start development mode (format, check, test on file changes)
dev:
    @echo "Starting development mode..."
    @echo "This requires 'watchexec' or similar file watcher"
    @echo "Install with: cargo install watchexec-cli"
    watchexec -w . -e go --ignore-paths target,vendor "just check"

# Quick development check (fast feedback)
quick: fmt vet
    @echo "Quick check complete"

# Documentation commands

# Generate Go documentation
docs:
    @echo "Generating documentation..."
    go doc -all . > docs/api/generated.md
    @echo "API documentation generated in docs/api/"

# Serve documentation locally
serve-docs:
    @echo "Serving documentation on http://localhost:6060"
    godoc -http=:6060

# Utility commands

# Show project statistics
stats:
    @echo "Project Statistics:"
    @echo "==================="
    @find . -name "*.go" -not -path "./vendor/*" | wc -l | xargs echo "Go files:"
    @find . -name "*.go" -not -path "./vendor/*" -exec wc -l {} + | tail -1 | awk '{print "Total lines: " $$1}'
    @echo ""
    @echo "Package breakdown:"
    @find internal -maxdepth 1 -type d | tail -n +2 | sort

# Check for TODO comments
todo:
    @echo "TODO items found:"
    @echo "================="
    @grep -r "TODO\|FIXME\|XXX\|HACK" --include="*.go" . || echo "No TODO items found"

# Update task status in TASKS.md
update-tasks:
    @echo "Checking implementation status..."
    @echo "This would scan code and update TASKS.md checkboxes"
    @echo "(Implementation needed)"

# Git helpers

# Pre-commit hook
pre-commit: check
    @echo "Pre-commit checks passed!"

# Initialize git hooks
init-hooks:
    @echo "Setting up git hooks..."
    echo '#!/bin/sh\njust pre-commit' > .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    @echo "Git hooks installed"

# Platform-specific commands

# Windows-specific build
build-windows:
    $env:GOOS='windows'; go build ./...

# Linux-specific build  
build-linux:
    GOOS=linux go build ./...

# Build with X11 support (Linux only)
build-x11:
    @echo "Building with X11 support..."
    go build -tags x11 ./...

# Build X11 examples
build-x11-examples:
    @echo "Building X11 examples..."
    go build -tags x11 -o /tmp/x11-demo examples/platform/x11/main.go

# Build SDL2 examples
build-sdl2-examples:
    @echo "Building SDL2 examples..."
    go build -tags sdl2 -o /tmp/sdl2-demo examples/platform/sdl2/main.go

# macOS-specific build
build-macos:
    GOOS=darwin go build ./...

# Cross-platform build all
build-all-platforms: build-windows build-linux build-macos

# Benchmarking and profiling

# Profile memory usage
profile-mem EXAMPLE="basic/hello_world":
    @echo "Profiling memory for {{EXAMPLE}}..."
    go run -memprofile=mem.prof examples/{{EXAMPLE}}/main.go
    go tool pprof mem.prof

# Profile CPU usage
profile-cpu EXAMPLE="basic/hello_world":
    @echo "Profiling CPU for {{EXAMPLE}}..."
    go run -cpuprofile=cpu.prof examples/{{EXAMPLE}}/main.go
    go tool pprof cpu.prof

# Continuous Integration commands

# CI build (strict mode)
ci-build:
    @echo "CI Build - strict mode"
    go build -race ./...

# CI test (with race detection)
ci-test:
    @echo "CI Test - with race detection"
    go test -race -timeout=10m ./...

# Full CI pipeline
ci: ci-build ci-test test-coverage
    @echo "CI pipeline completed"

# Release commands

# Prepare release (run all checks, update version)
release-prepare VERSION:
    @echo "Preparing release {{VERSION}}..."
    just check
    @echo "Release {{VERSION}} ready (manual git tag required)"

# Tag release
release-tag VERSION:
    @echo "Tagging release {{VERSION}}..."
    git tag -a v{{VERSION}} -m "Release v{{VERSION}}"
    @echo "Tagged v{{VERSION}} (push with: git push origin v{{VERSION}})"
