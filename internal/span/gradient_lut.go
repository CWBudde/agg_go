package span

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// ColorInterpolator defines the interface for color interpolation between two colors.
// This provides flexibility for different color types and interpolation strategies.
type ColorInterpolator[T any] interface {
	// Inc advances the interpolator to the next step
	Inc()
	// Color returns the current interpolated color
	Color() T
}

// ColorInterpolatorGeneric provides generic color interpolation using the color's Gradient method.
// This works with any color type that implements a Gradient method.
type ColorInterpolatorGeneric[T interface{ Gradient(T, float64) T }] struct {
	c1    T    // starting color
	c2    T    // ending color
	len   uint // total steps
	count uint // current step
}

// NewColorInterpolatorGeneric creates a new generic color interpolator.
func NewColorInterpolatorGeneric[T interface{ Gradient(T, float64) T }](c1, c2 T, length uint) *ColorInterpolatorGeneric[T] {
	if length == 0 {
		length = 1
	}
	return &ColorInterpolatorGeneric[T]{
		c1:    c1,
		c2:    c2,
		len:   length,
		count: 0,
	}
}

// Inc advances the interpolator to the next step.
func (ci *ColorInterpolatorGeneric[T]) Inc() {
	ci.count++
}

// Color returns the current interpolated color.
func (ci *ColorInterpolatorGeneric[T]) Color() T {
	if ci.count >= ci.len {
		return ci.c2
	}
	return ci.c1.Gradient(ci.c2, float64(ci.count)/float64(ci.len))
}

// ColorInterpolatorRGBA8 provides optimized interpolation for 8-bit RGBA colors
// using DDA line interpolators for each color channel.
type ColorInterpolatorRGBA8[CS any] struct {
	r, g, b, a *DDALineInterpolator
}

// NewColorInterpolatorRGBA8 creates a new optimized RGBA8 color interpolator.
func NewColorInterpolatorRGBA8[CS any](c1, c2 color.RGBA8[CS], length uint) *ColorInterpolatorRGBA8[CS] {
	if length == 0 {
		length = 1
	}

	// Use 14-bit fraction shift for high precision color interpolation
	const fractionShift = 14

	return &ColorInterpolatorRGBA8[CS]{
		r: NewDDALineInterpolator(int(c1.R), int(c2.R), length, fractionShift),
		g: NewDDALineInterpolator(int(c1.G), int(c2.G), length, fractionShift),
		b: NewDDALineInterpolator(int(c1.B), int(c2.B), length, fractionShift),
		a: NewDDALineInterpolator(int(c1.A), int(c2.A), length, fractionShift),
	}
}

// Inc advances the interpolator to the next step.
func (ci *ColorInterpolatorRGBA8[CS]) Inc() {
	ci.r.Inc()
	ci.g.Inc()
	ci.b.Inc()
	ci.a.Inc()
}

// Color returns the current interpolated RGBA8 color.
func (ci *ColorInterpolatorRGBA8[CS]) Color() color.RGBA8[CS] {
	return color.RGBA8[CS]{
		R: basics.Int8u(ci.r.Y()),
		G: basics.Int8u(ci.g.Y()),
		B: basics.Int8u(ci.b.Y()),
		A: basics.Int8u(ci.a.Y()),
	}
}

// ColorInterpolatorGray8 provides optimized interpolation for 8-bit grayscale colors
// using DDA line interpolators for value and alpha channels.
type ColorInterpolatorGray8[CS any] struct {
	v, a *DDALineInterpolator
}

// NewColorInterpolatorGray8 creates a new optimized Gray8 color interpolator.
func NewColorInterpolatorGray8[CS any](c1, c2 color.Gray8[CS], length uint) *ColorInterpolatorGray8[CS] {
	if length == 0 {
		length = 1
	}

	// Use 14-bit fraction shift for high precision color interpolation
	const fractionShift = 14

	return &ColorInterpolatorGray8[CS]{
		v: NewDDALineInterpolator(int(c1.V), int(c2.V), length, fractionShift),
		a: NewDDALineInterpolator(int(c1.A), int(c2.A), length, fractionShift),
	}
}

// Inc advances the interpolator to the next step.
func (ci *ColorInterpolatorGray8[CS]) Inc() {
	ci.v.Inc()
	ci.a.Inc()
}

// Color returns the current interpolated Gray8 color.
func (ci *ColorInterpolatorGray8[CS]) Color() color.Gray8[CS] {
	return color.Gray8[CS]{
		V: basics.Int8u(ci.v.Y()),
		A: basics.Int8u(ci.a.Y()),
	}
}

// ColorPoint represents a color stop in a gradient with an offset position.
type ColorPoint[T any] struct {
	Offset float64 // position in gradient [0.0, 1.0]
	Color  T       // color at this position
}

// NewColorPoint creates a new color point with offset clamped to [0.0, 1.0].
func NewColorPoint[T any](offset float64, c T) ColorPoint[T] {
	if offset < 0.0 {
		offset = 0.0
	}
	if offset > 1.0 {
		offset = 1.0
	}
	return ColorPoint[T]{Offset: offset, Color: c}
}

// GradientLUT represents a gradient lookup table for efficient color interpolation.
// This is equivalent to AGG's gradient_lut<ColorInterpolator, ColorLutSize> template class.
type GradientLUT[T any, CI ColorInterpolator[T]] struct {
	lutSize      int                              // size of the lookup table
	colorProfile *array.PodBVector[ColorPoint[T]] // color stops/profile
	colorLUT     []T                              // precomputed lookup table
}

// NewGradientLUT creates a new gradient lookup table with the specified size.
// Common sizes are 256, 512, and 1024.
func NewGradientLUT[T any, CI ColorInterpolator[T]](lutSize int) *GradientLUT[T, CI] {
	return &GradientLUT[T, CI]{
		lutSize:      lutSize,
		colorProfile: array.NewPodBVector[ColorPoint[T]](),
		colorLUT:     make([]T, lutSize),
	}
}

// RemoveAll clears all color stops from the gradient.
func (gl *GradientLUT[T, CI]) RemoveAll() {
	gl.colorProfile.RemoveAll()
}

// AddColor adds a color stop to the gradient at the specified offset.
// The offset must be in the range [0.0, 1.0] and defines a color stop
// as described in the SVG specification for gradients.
func (gl *GradientLUT[T, CI]) AddColor(offset float64, c T) {
	gl.colorProfile.Add(NewColorPoint(offset, c))
}

// BuildLUT builds the lookup table from the color profile.
// This must be called after adding all color stops and before using the LUT.
func (gl *GradientLUT[T, CI]) BuildLUT(newInterpolator func(T, T, uint) CI) {
	// Sort color stops by offset
	array.QuickSort(gl.colorProfile, func(a, b ColorPoint[T]) bool {
		return a.Offset < b.Offset
	})

	// Remove duplicate offsets
	newSize := array.RemoveDuplicates(gl.colorProfile, func(a, b ColorPoint[T]) bool {
		return a.Offset == b.Offset
	})
	gl.colorProfile.CutAt(newSize)

	if gl.colorProfile.Size() >= 2 {
		// Fill LUT with interpolated colors
		start := int(basics.URound(gl.colorProfile.At(0).Offset * float64(gl.lutSize)))
		var end int

		// Fill initial segment with first color
		c := gl.colorProfile.At(0).Color
		for i := 0; i < start; i++ {
			gl.colorLUT[i] = c
		}

		// Interpolate between color stops
		for i := 1; i < gl.colorProfile.Size(); i++ {
			end = int(basics.URound(gl.colorProfile.At(i).Offset * float64(gl.lutSize)))

			// Clamp end to lutSize for boundary conditions
			if end > gl.lutSize {
				end = gl.lutSize
			}

			// If this is the last stop and it ends at or beyond lutSize,
			// we need to ensure the final LUT position gets the exact final color
			isLastStop := i == gl.colorProfile.Size()-1
			actualEnd := end
			if isLastStop && end == gl.lutSize {
				actualEnd = end - 1 // Stop one position before to leave room for exact final color
			}

			// Create interpolator for this segment
			ci := newInterpolator(
				gl.colorProfile.At(i-1).Color,
				gl.colorProfile.At(i).Color,
				uint(actualEnd-start+1),
			)

			// Fill segment with interpolated colors
			for start < actualEnd {
				gl.colorLUT[start] = ci.Color()
				ci.Inc()
				start++
			}

			// If this is the last stop ending at lutSize, set the final position exactly
			if isLastStop && end == gl.lutSize {
				gl.colorLUT[gl.lutSize-1] = gl.colorProfile.At(i).Color
				start = gl.lutSize
			}
		}

		// Fill remaining segment with last color
		c = gl.colorProfile.At(gl.colorProfile.Size() - 1).Color
		for end < gl.lutSize {
			gl.colorLUT[end] = c
			end++
		}
	}
}

// Size returns the size of the lookup table.
// This allows the GradientLUT to be used directly as a ColorF in span_gradient.
func (gl *GradientLUT[T, CI]) Size() int {
	return gl.lutSize
}

// At returns the color at the specified index in the lookup table.
// This allows the GradientLUT to be used directly as a ColorF in span_gradient.
func (gl *GradientLUT[T, CI]) At(i int) T {
	if i < 0 {
		return gl.colorLUT[0]
	}
	if i >= gl.lutSize {
		return gl.colorLUT[gl.lutSize-1]
	}
	return gl.colorLUT[i]
}
