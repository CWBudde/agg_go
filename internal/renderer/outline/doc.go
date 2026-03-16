// Package outline implements AGG's outline-oriented rendering path.
//
// Unlike scanline rendering, which fills spans produced from polygon coverage,
// outline rendering walks stroke geometry directly with distance and line
// interpolators. It is used for anti-aliased vector outlines and patterned line
// rendering where the original AGG library uses dedicated outline renderers.
package outline
