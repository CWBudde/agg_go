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

# Build all examples for current platform with organized output
build-all-examples: build-core-examples build-platform-examples
    @echo "All examples built for {{platform}} platform"

# Build core examples for multiple platforms (cross-compile)
build-all-examples-multi:
    @echo "Building core examples for multiple platforms..."
    @mkdir -p bin/linux bin/darwin bin/windows
    @echo "Cross-compiling for Linux..."
    @GOOS=linux go build -o bin/linux/hello_world examples/core/basic/hello_world/main.go && echo "  ✓ linux/hello_world" || echo "  ✗ linux/hello_world"
    @GOOS=linux go build -o bin/linux/shapes examples/core/basic/shapes/main.go && echo "  ✓ linux/shapes" || echo "  ✗ linux/shapes"
    @echo "Cross-compiling for macOS..."
    @GOOS=darwin go build -o bin/darwin/hello_world examples/core/basic/hello_world/main.go && echo "  ✓ darwin/hello_world" || echo "  ✗ darwin/hello_world"
    @GOOS=darwin go build -o bin/darwin/shapes examples/core/basic/shapes/main.go && echo "  ✓ darwin/shapes" || echo "  ✗ darwin/shapes"
    @echo "Cross-compiling for Windows..."
    @GOOS=windows go build -o bin/windows/hello_world.exe examples/core/basic/hello_world/main.go && echo "  ✓ windows/hello_world.exe" || echo "  ✗ windows/hello_world.exe"
    @GOOS=windows go build -o bin/windows/shapes.exe examples/core/basic/shapes/main.go && echo "  ✓ windows/shapes.exe" || echo "  ✗ windows/shapes.exe"
    @echo "Cross-compilation complete (selected examples)"

# Build only core examples (no platform dependencies)
build-core-examples:
    @echo "Building core examples only..."
    @mkdir -p bin/core
    @echo "Building hello_world..."
    @go build -o bin/core/hello_world{{bin_ext}} examples/core/basic/hello_world/main.go && echo "  ✓ hello_world" || echo "  ✗ hello_world"
    @echo "Building shapes..."
    @go build -o bin/core/shapes{{bin_ext}} examples/core/basic/shapes/main.go && echo "  ✓ shapes" || echo "  ✗ shapes"
    @echo "Building lines..."
    @go build -o bin/core/lines{{bin_ext}} examples/core/basic/lines/main.go && echo "  ✓ lines" || echo "  ✗ lines"
    @echo "Building rounded_rect..."
    @go build -o bin/core/rounded_rect{{bin_ext}} examples/core/basic/rounded_rect/main.go && echo "  ✓ rounded_rect" || echo "  ✗ rounded_rect"
    @echo "Building colors_gray..."
    @go build -o bin/core/colors_gray{{bin_ext}} examples/core/basic/colors_gray/main.go && echo "  ✓ colors_gray" || echo "  ✗ colors_gray"
    @echo "Building colors_rgba..."
    @go build -o bin/core/colors_rgba{{bin_ext}} examples/core/basic/colors_rgba/main.go && echo "  ✓ colors_rgba" || echo "  ✗ colors_rgba"
    @echo "Building embedded_fonts_hello..."
    @go build -o bin/core/embedded_fonts_hello{{bin_ext}} examples/core/basic/embedded_fonts_hello/main.go && echo "  ✓ embedded_fonts_hello" || echo "  ✗ embedded_fonts_hello"
    @echo "Building basic_demo..."
    @go build -o bin/core/basic_demo{{bin_ext}} examples/core/basic/basic_demo/main.go && echo "  ✓ basic_demo" || echo "  ✗ basic_demo"
    @echo "Building gradients..."
    @go build -o bin/core/gradients{{bin_ext}} examples/core/intermediate/gradients/main.go && echo "  ✓ gradients" || echo "  ✗ gradients"
    @echo "Building text_rendering..."
    @go build -o bin/core/text_rendering{{bin_ext}} examples/core/intermediate/text_rendering/main.go && echo "  ✓ text_rendering" || echo "  ✗ text_rendering"
    @echo "Building advanced_rendering..."
    @go build -o bin/core/advanced_rendering{{bin_ext}} examples/core/advanced/advanced_rendering/main.go && echo "  ✓ advanced_rendering" || echo "  ✗ advanced_rendering"

# Build only platform-specific examples
build-platform-examples:
    @echo "Building platform examples for {{platform}}..."
    @mkdir -p bin/{{platform}}
    @echo "Building X11 demo..."
    @go build -tags x11 -o bin/{{platform}}/x11_demo{{bin_ext}} examples/platform/x11/main.go && echo "  ✓ x11_demo" || echo "  ✗ x11_demo (X11 dependencies missing)"
    @echo "Building SDL2 demo..."
    @go build -tags sdl2 -o bin/{{platform}}/sdl2_demo{{bin_ext}} examples/platform/sdl2/main.go && echo "  ✓ sdl2_demo" || echo "  ✗ sdl2_demo (SDL2 dependencies missing)"

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
