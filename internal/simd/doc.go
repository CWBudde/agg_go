// Package simd provides runtime-dispatched bulk pixel kernels.
//
// The package keeps one public API and selects an implementation at runtime
// based on detected CPU features plus build tags. Scalar Go code is always the
// correctness baseline; amd64 and arm64 assembly paths are optional
// accelerators that must remain bit-identical to the generic implementation.
//
// Assembly entry points are declared in the arch-specific Go files and are
// intentionally hidden behind the same exported operations used by pixfmt and
// alpha-mask call sites.
package simd
