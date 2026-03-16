// Package effects ports AGG's image post-processing helpers.
//
// The package is centered on the blur family from agg_blur.h: stack blur with
// the same precomputed mul/shift tables AGG uses for fast 8-bit kernels, plus
// smaller convenience filters such as SlightBlur for very small-radius cleanup.
//
// These helpers are intentionally separate from the main rasterizer/renderer
// pipeline so callers can apply them explicitly as a post-process step.
package effects
