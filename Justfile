# AGG Go Build Orchestration
# Run `just --list` to see all available commands

set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

# Platform detection variables
os := `uname -s`
arch := `uname -m`
platform := if os == "Linux" { "linux" } else if os == "Darwin" { "darwin" } else if os =~ "MINGW|MSYS|CYGWIN" { "windows" } else { "unknown" }

# Platform-specific binary extensions
bin_ext := if platform == "windows" { ".exe" } else { "" }

# Default recipe - show available commands
default:
    @just --list

# Build commands

# Build everything (library + demos)
build: build-lib build-demos

# Build the library
build-lib:
    @echo "Building AGG Go library..."
    go build ./...

# Build WASM demo
build-wasm:
    @echo "Building AGG Go WASM demo..."
    ./web/build-wasm.sh

# Serve the web demo locally
serve-web: build-wasm
    @echo "Serving AGG Go web demo on http://localhost:8080"
    @go run -e "package main; import ('net/http'; 'log'); func main() { log.Println('Serving on :8080'); log.Fatal(http.ListenAndServe(':8080', http.FileServer(http.Dir('web')))) }"

# Build all demo executables into ./bin/ (auto-discovers all examples/core/**/main.go)
build-demos:
    #!/usr/bin/env bash
    set -e
    mkdir -p bin
    ok=0; fail=0
    while IFS= read -r dir; do
        # Path relative to examples/core/, e.g. "basic/hello_world"
        relpath="${dir#examples/core/}"
        # Strip the tier prefix (basic/ intermediate/ advanced/)
        name="${relpath#basic/}"
        name="${name#intermediate/}"
        name="${name#advanced/}"
        # Replace remaining slashes (e.g. controls/gamma_correction -> controls_gamma_correction)
        name="${name//\//_}"
        errfile=$(mktemp)
        if go build -o "bin/${name}{{bin_ext}}" "./${dir}" 2>"$errfile"; then
            echo "  ✓ ${name}"
            ok=$((ok+1))
        else
            echo "  ✗ ${name}: $(head -1 "$errfile")"
            fail=$((fail+1))
        fi
        rm -f "$errfile"
    done < <(find examples/core -mindepth 2 -name "main.go" -exec dirname {} \; | sort)
    echo ""
    echo "${ok} built, ${fail} failed — executables in bin/"

# Alias for backward compatibility
build-examples: build-demos

# Build specific example by path relative to examples/ (e.g. core/basic/shapes)
build-example EXAMPLE:
    @echo "Building example: {{EXAMPLE}}"
    go build -o bin/$(basename {{EXAMPLE}}){{bin_ext}} ./examples/{{EXAMPLE}}

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

# Run SIMD-focused tests
test-simd:
    @echo "Running SIMD dispatch tests..."
    go test ./internal/simd -count=1
    go build ./internal/pixfmt

# Benchmark SIMD fill dispatch paths
bench-simd:
    @echo "Benchmarking SIMD fill paths..."
    go test ./internal/simd -bench 'Benchmark(FillRGBA|BlendSolidHspanRGBA)' -run '^$' -count=1

# Run tests on ARM64 using QEMU (requires qemu-user-static)
test-arm64:
    #!/usr/bin/env bash
    if ! command -v qemu-aarch64-static &> /dev/null; then
        echo "Error: qemu-aarch64-static not found"
        echo "Install with: sudo apt-get install qemu-user-static binfmt-support"
        exit 1
    fi
    GOOS=linux GOARCH=arm64 go test -exec="qemu-aarch64-static" -count=1 ./internal/simd
    GOOS=linux GOARCH=arm64 go build ./internal/pixfmt

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
    golangci-lint run --new ./...
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

# Clean everything (build artifacts + compiled examples)
clean-all: clean clean-examples

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

# Run a named core demo interactively (SDL2 window). Usage: just run-demo hello_world
# The EXAMPLE name is the binary name from bin/ (same as build-demos output).
# Finds the matching examples/core/**/EXAMPLE/ directory automatically.
run-demo EXAMPLE:
    #!/usr/bin/env bash
    set -e
    dir=$(find examples/core -mindepth 2 -type d -name "{{EXAMPLE}}" | head -1)
    if [ -z "$dir" ]; then
        echo "Demo '{{EXAMPLE}}' not found. Run 'just list-demos' to see available demos."
        exit 1
    fi
    echo "Running {{EXAMPLE}} (SDL2 interactive) from $dir"
    go run -tags sdl2 "./$dir"

# Same as run-demo but using X11 backend
run-demo-x11 EXAMPLE:
    #!/usr/bin/env bash
    set -e
    dir=$(find examples/core -mindepth 2 -type d -name "{{EXAMPLE}}" | head -1)
    if [ -z "$dir" ]; then
        echo "Demo '{{EXAMPLE}}' not found. Run 'just list-demos' to see available demos."
        exit 1
    fi
    echo "Running {{EXAMPLE}} (X11 interactive) from $dir"
    go run -tags x11 "./$dir"

# List all available demos
list-demos:
    @find examples/core -mindepth 2 -name "main.go" -exec dirname {} \; | sort | \
        sed 's|examples/core/[a-z]*/||; s|examples/core/[a-z]*/controls/|controls_|' | sort

# Run the interactive platform showcase demo (SDL2)
run-x11-demo:
    go run -tags x11 examples/platform/x11/main.go

# Run SDL2 demo (requires SDL2 dependencies)
run-sdl2-demo:
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

# List all available examples
list-examples:
    @echo "Available Examples:"
    @echo "=================="
    @echo ""
    @echo "Core Examples (platform independent):"
    @find examples/core -name "main.go" -type f | sed 's|examples/core/||; s|/main.go||' | sort | sed 's/^/  /'
    @echo ""
    @echo "Platform Examples:"
    @find examples/platform -name "main.go" -type f | sed 's|examples/platform/||; s|/main.go||' | sort | sed 's/^/  /'
    @if find examples/tests -name "main.go" -type f >/dev/null 2>&1; then \
        echo ""; \
        echo "Test Examples:"; \
        find examples/tests -name "main.go" -type f | sed 's|examples/tests/||; s|/main.go||' | sort | sed 's/^/  /'; \
    fi
    @echo ""
    @echo "Usage:"
    @echo "  just build-example-to-bin <name>    # Build specific example"
    @echo "  just run-example <path>             # Run example directly"

# Clean compiled examples from bin/
clean-examples:
    @echo "Cleaning compiled examples..."
    @if [ -d bin/ ]; then \
        rm -rf bin/; \
        echo "Removed bin/ directory"; \
    else \
        echo "bin/ directory doesn't exist"; \
    fi

# Show platform information
show-platform:
    @echo "Platform Information:"
    @echo "===================="
    @echo "OS: {{os}}"
    @echo "Architecture: {{arch}}"
    @echo "Platform: {{platform}}"
    @echo "Binary extension: {{bin_ext}}"
    @echo ""
    @echo "Available GUI libraries:"
    @if [ "{{platform}}" = "linux" ] && command -v pkg-config >/dev/null 2>&1; then \
        if pkg-config --exists x11; then \
            echo "  ✓ X11 available"; \
        else \
            echo "  ✗ X11 not available (install libx11-dev)"; \
        fi; \
        if pkg-config --exists sdl2; then \
            echo "  ✓ SDL2 available"; \
        else \
            echo "  ✗ SDL2 not available (install libsdl2-dev)"; \
        fi; \
    else \
        echo "  Platform detection needed for other OSes"; \
    fi

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
