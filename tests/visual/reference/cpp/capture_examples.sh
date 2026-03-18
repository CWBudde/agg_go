#!/usr/bin/env bash

set -u
set -o pipefail

usage() {
    cat <<'EOF'
Usage:
  capture_examples.sh [--mode strict|fallback] [--agg-root DIR]
                      [--srcdir DIR] [--bindir DIR] [--outdir DIR]

Modes:
  strict    Use the direct screenshot path only.
  fallback  Kept for compatibility; behaves the same as strict.

Defaults:
  --mode fallback
  --agg-root ../agg-2.6 relative to this script's repository root
EOF
}

mode="fallback"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../../../../" && pwd)"
agg_root="${AGG_ROOT:-$repo_root/../agg-2.6}"

poll_ms="${POLL_MS:-250}"
ppm_timeout_ms="${PPM_TIMEOUT_MS:-10000}"

while [ "$#" -gt 0 ]; do
    case "$1" in
        --mode)
            mode="${2:-}"
            shift 2
            ;;
        --mode=*)
            mode="${1#*=}"
            shift
            ;;
        --agg-root)
            agg_root="${2:-}"
            shift 2
            ;;
        --agg-root=*)
            agg_root="${1#*=}"
            shift
            ;;
        --srcdir)
            srcdir="${2:-}"
            shift 2
            ;;
        --srcdir=*)
            srcdir="${1#*=}"
            shift
            ;;
        --bindir)
            bindir="${2:-}"
            shift 2
            ;;
        --bindir=*)
            bindir="${1#*=}"
            shift
            ;;
        --outdir)
            outdir="${2:-}"
            shift 2
            ;;
        --outdir=*)
            outdir="${1#*=}"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "unknown option: $1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

if [ "$mode" != "strict" ] && [ "$mode" != "fallback" ]; then
    echo "invalid mode: $mode" >&2
    usage >&2
    exit 1
fi

agg_root="$(cd "$agg_root" && pwd)"
srcdir="${SRC_DIR:-$agg_root/agg-src/examples}"
bindir="${BUILD_DIR:-$agg_root/build/examples}"
outdir="${OUT_DIR:-$repo_root/tests/visual/reference/cpp/examples}"

if [ ! -d "$srcdir" ]; then
    echo "source directory not found: $srcdir" >&2
    exit 1
fi
if [ ! -d "$bindir" ]; then
    echo "build directory not found: $bindir" >&2
    exit 1
fi
mkdir -p "$outdir"

wait_for_file() {
    local file="$1"
    local pid="$2"
    local elapsed_ms=0
    local sleep_s

    sleep_s="$(awk -v ms="$poll_ms" 'BEGIN { printf "%.3f", ms / 1000.0 }')"

    while [ "$elapsed_ms" -lt "$ppm_timeout_ms" ]; do
        if [ -f "$file" ] && ! kill -0 "$pid" 2>/dev/null; then
            return 0
        fi
        if ! kill -0 "$pid" 2>/dev/null; then
            [ -f "$file" ] && return 0
            return 1
        fi
        sleep "$sleep_s"
        elapsed_ms=$((elapsed_ms + poll_ms))
    done

    [ -f "$file" ]
}

stage_assets() {
    local dir=""
    local file=""
    for dir in "$srcdir/X11" "$srcdir/art" "$srcdir/SDL2"; do
        [ -d "$dir" ] || continue
        for file in "$dir"/*; do
            [ -f "$file" ] || continue
            case "$file" in
                *.ppm|*.bmp|*.txt)
                    cp -f "$file" "$bindir/"
                    ;;
            esac
        done
    done
}

capture_one() {
    local bin="$1"
    local out="$outdir/$bin.png"
    local log="/tmp/${bin}.log"
    local ppm="$bindir/screenshot.ppm"
    local pid=""
    local status=0

    if [ ! -x "$bindir/$bin" ]; then
        echo "skip $bin: binary missing"
        return 0
    fi

    rm -f "$ppm" "$log"
    (
        cd "$bindir" &&
        "./$bin" --screenshot
    ) >"$log" 2>&1 &
    pid=$!

    if wait_for_file "$ppm" "$pid"; then
        wait "$pid" 2>/dev/null || true

        if convert "$ppm" "$out"; then
            rm -f "$ppm"
        else
            echo "fail $bin: convert of screenshot.ppm failed"
            sed -n '1,120p' "$log" || true
            status=1
        fi
    else
        echo "fail $bin: missing screenshot.ppm"
        sed -n '1,120p' "$log" || true
        kill "$pid" 2>/dev/null || true
        wait "$pid" 2>/dev/null || true
        status=1
    fi

    if [ "$status" -eq 0 ]; then
        if [ "$(id -u)" -eq 0 ] && [ -n "${SUDO_USER:-}" ] && [ "$SUDO_USER" != "root" ]; then
            chown "$SUDO_USER":"$SUDO_USER" "$out" 2>/dev/null || true
        fi
        echo "ok $bin"
    fi

    return "$status"
}

failures=0
captured=0

stage_assets

for src in "$srcdir"/*.cpp; do
    [ -e "$src" ] || continue
    bin="$(basename "${src%.cpp}")"
    if capture_one "$bin"; then
        captured=$((captured + 1))
    else
        failures=$((failures + 1))
    fi
done

echo "done: $captured captured, $failures failed, mode=$mode"
if [ "$failures" -gt 0 ]; then
    exit 1
fi
