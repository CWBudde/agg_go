// Package pixfmt provides pixel format implementations for AGG.
// This package handles the actual pixel-level rendering operations.
package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

// Re-export order types for convenience.
// This allows users to write pixfmt.RGBAOrder instead of importing order separately.
type (
	RGBAOrder = order.RGBA
	BGRAOrder = order.BGRA
	ARGBOrder = order.ARGB
	ABGROrder = order.ABGR
)

// Pixel format category tags for type safety
type (
	PixFmtGrayTag struct{}
	PixFmtRGBTag  struct{}
	PixFmtRGBATag struct{}
)

// BlenderBase provides the base interface for pixel blending operations
type BlenderBase[C any, O any] interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u)
	Get(p []basics.Int8u, cover basics.Int8u) color.RGBA
	GetRaw(p []basics.Int8u) (r, g, b, a basics.Int8u)
	Set(p []basics.Int8u, c color.RGBA)
	SetRaw(p []basics.Int8u, r, g, b, a basics.Int8u)
}

// PixelType represents a pixel with multiple components
type PixelType[T any] struct {
	Components []T
}

// Set sets all components to the same value
func (p *PixelType[T]) Set(value T) {
	for i := range p.Components {
		p.Components[i] = value
	}
}

// Get returns the component at the specified index
func (p *PixelType[T]) Get(index int) T {
	if index >= 0 && index < len(p.Components) {
		return p.Components[index]
	}
	var zero T
	return zero
}

// Set component at specified index
func (p *PixelType[T]) SetComponent(index int, value T) {
	if index >= 0 && index < len(p.Components) {
		p.Components[index] = value
	}
}

// Constants for cover operations
const (
	CoverFull = 255
	CoverNone = 0
)

// Utility functions for pixel format implementations

// ClampX clamps x coordinate to valid range
func ClampX(x, width int) int {
	if x < 0 {
		return 0
	}
	if x >= width {
		return width - 1
	}
	return x
}

// ClampY clamps y coordinate to valid range
func ClampY(y, height int) int {
	if y < 0 {
		return 0
	}
	if y >= height {
		return height - 1
	}
	return y
}

// InBounds checks if coordinates are within bounds
func InBounds(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
