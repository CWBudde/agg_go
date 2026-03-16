// Package pixfmt provides pixel format implementations for AGG.
// This package handles the actual pixel-level rendering operations.
package pixfmt

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/order"
)

// Re-export order types for convenience.
// This keeps public pixfmt declarations close to AGG's Order template parameter
// without forcing callers to import the order package explicitly.
type (
	RGBAOrder = order.RGBA
	BGRAOrder = order.BGRA
	ARGBOrder = order.ARGB
	ABGROrder = order.ABGR
)

// Pixel format category tags mirror AGG's pixfmt_gray_tag, pixfmt_rgb_tag, and
// pixfmt_rgba_tag. They let generic code constrain pixfmt families without
// introducing inheritance-style hierarchies.
type (
	PixFmtGrayTag struct{}
	PixFmtRGBTag  struct{}
	PixFmtRGBATag struct{}
)

// BlenderBase is the common per-pixel contract used by pixfmt implementations.
//
// In AGG this role is fulfilled by template blenders such as blender_rgba and
// blender_rgba_pre. The Go port keeps the same separation of concerns but
// expresses it as an interface so pixel formats can stay generic over channel
// order and blending math.
type BlenderBase[C any, O any] interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u)
	Get(p []basics.Int8u, cover basics.Int8u) color.RGBA
	GetRaw(p []basics.Int8u) (r, g, b, a basics.Int8u)
	Set(p []basics.Int8u, c color.RGBA)
	SetRaw(p []basics.Int8u, r, g, b, a basics.Int8u)
}

// PixelType is a small helper for formats that expose their components as a
// slice rather than fixed fields.
type PixelType[T any] struct {
	Components []T
}

// Set fills every component with the same value.
func (p *PixelType[T]) Set(value T) {
	for i := range p.Components {
		p.Components[i] = value
	}
}

// Get returns the component at index or the zero value when out of range.
func (p *PixelType[T]) Get(index int) T {
	if index >= 0 && index < len(p.Components) {
		return p.Components[index]
	}
	var zero T
	return zero
}

// SetComponent writes a single component when index is in range.
func (p *PixelType[T]) SetComponent(index int, value T) {
	if index >= 0 && index < len(p.Components) {
		p.Components[index] = value
	}
}

// Coverage constants match AGG's cover_none and cover_full values.
const (
	CoverFull = 255
	CoverNone = 0
)

// ClampX clamps x to a valid pixel column.
func ClampX(x, width int) int {
	if x < 0 {
		return 0
	}
	if x >= width {
		return width - 1
	}
	return x
}

// ClampY clamps y to a valid pixel row.
func ClampY(y, height int) int {
	if y < 0 {
		return 0
	}
	if y >= height {
		return height - 1
	}
	return y
}

// InBounds reports whether x,y addresses a valid pixel.
func InBounds(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
}

// Min returns the smaller integer.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger integer.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
