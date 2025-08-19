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
    @echo "Building examples..."
    find examples -name "*.go" -path "*/main.go" -exec dirname {} \; | sort -u | xargs -I {} go build -o /dev/null {}

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

# Development commands

# Format all Go code
fmt:
    @echo "Formatting Go code..."
    go fmt ./...

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
    go run examples/basic/hello_world/main.go

# Run specific example
run-example EXAMPLE:
    @echo "Running example: {{EXAMPLE}}"
    go run examples/{{EXAMPLE}}/main.go

# Run hello_world example (alias for convenience)
run EXAMPLE:
    @echo "Running example: {{EXAMPLE}}"
    go run examples/basic/{{EXAMPLE}}/main.go

# Run X11 demo (requires X11 headers and running X server)
run-x11-demo:
    @echo "Running X11 demo..."
    @echo "Note: Requires X11 headers and running X server"
    go run -tags x11 examples/x11_demo/main.go

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
    go build -tags x11 -o /tmp/x11-demo examples/x11_demo/main.go

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