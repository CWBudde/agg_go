# AGG Go Documentation

This directory contains the user and developer documentation for the AGG Go
port.

## Start Here

- [guides/getting-started.md](./guides/getting-started.md) for the first runnable example
- [guides/core-types.md](./guides/core-types.md) for the public root-level types
- [guides/basic-shapes.md](./guides/basic-shapes.md) for the `Context` shape API
- [TEXT_RENDERING.md](./TEXT_RENDERING.md) for the current text-rendering workflow
- [guides/image-compositing.md](./guides/image-compositing.md) for image drawing and blending
- [guides/migrating-from-cpp-agg.md](./guides/migrating-from-cpp-agg.md) for C++ AGG to Go migration
- [guides/performance-optimization.md](./guides/performance-optimization.md) for practical performance advice
- [AGG2D_PARITY.md](./AGG2D_PARITY.md) for C++ `Agg2D` to Go API mapping

## Reference Material

- `architecture/` for package and pipeline overviews
- `concepts/` for translation and architectural notes
- `AGG_DELTAS.md` for intentional deviations from original AGG
- `SIMD_OPTIMIZATIONS.md` for performance and SIMD notes

## Project Tracking

- `PLAN.md` in the repository root is the authoritative phased plan
- `docs/TASKS.md` tracks file-level implementation inventory

The public API is centered on `Context` for the simplest workflow and `Agg2D`
for closer parity with the original C++ high-level interface.
