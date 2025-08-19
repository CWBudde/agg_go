// Package transform provides transformation interfaces and utilities for AGG.
package transform

// Transformer interface for applying coordinate transformations.
// This interface abstracts the transformation operation, allowing different
// transformation types (affine, perspective, etc.) to be used interchangeably.
type Transformer interface {
	// Transform applies the transformation to the given point coordinates.
	// The x and y coordinates are modified in place.
	Transform(x, y *float64)
}
