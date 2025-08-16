# agg_go

A clean and direct port of agg-2.6 to Go.

This repository contains a minimal Go implementation of the Anti-Grain Geometry (AGG) library version 2.6, focusing on maintaining the core functionality and API structure of the original C++ codebase.

## Project Status

See [TASK.md](TASK.md) for a detailed implementation plan outlining all C++ files that need to be ported to Go.

## Potential Issues

- Unsigned Saturation.IRound: The generic `Saturation[T].IRound` in `internal/basics/constants.go` currently applies a lower-bound clamp using `-s.limit`. For unsigned types (e.g., `uint32`), applying unary minus wraps the value, which leads to negative inputs producing wrapped results instead of clamping to `0`. Example from tests: `Saturation[uint32](255).IRound(-1.2)` returns `4294967295` rather than `0`. This behavior is under review to confirm whether it matches the original AGG semantics. See `internal/basics/constants_extra_test.go::TestSaturationUnsignedIRound`.
