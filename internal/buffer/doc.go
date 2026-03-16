// Package buffer ports AGG's rendering-buffer accessors.
//
// It wraps pixel memory with width, height, stride, row, and cached-row access
// helpers that other packages use as the common attachment point for pixfmts
// and renderers.
package buffer
