// Package agg2d ports AGG's high-level Agg2D facade.
//
// The package owns the same top-level rendering state machine as the original
// C++ Agg2D class in agg2d.h / agg2d.cpp: attached buffer and clip state,
// current path storage, transform stack, fill and stroke styles, text state,
// image filtering and resampling state, and the rasterizer/scanline/span
// pipeline used to render those settings.
//
// Public wrappers in the repository root expose an idiomatic Go surface over
// this package, but the internal contracts here remain intentionally close to
// the original C++ layout so upstream AGG documentation and implementation
// details stay usable as a reference.
package agg2d
