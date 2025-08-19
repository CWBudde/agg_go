// Package span provides gradient alpha span generation functionality for AGG.
// This implements a port of AGG's span_gradient_alpha classes and functions.
package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// AlphaFunction defines the interface for gradient alpha functions.
// These functions provide alpha lookup based on gradient position.
type AlphaFunction interface {
	// Size returns the number of alpha values in the gradient
	Size() int

	// AlphaAt returns the alpha value at the specified index
	AlphaAt(index int) basics.Int8u
}

// AlphaType defines the interface for colors that have alpha channels.
type AlphaType interface {
	SetAlpha(alpha basics.Int8u)
	GetAlpha() basics.Int8u
}

// SpanGradientAlpha generates alpha-only gradient spans.
// This is a port of AGG's span_gradient_alpha template class.
type SpanGradientAlpha[ColorT AlphaType, InterpolatorT SpanInterpolatorInterface, GradientT GradientFunction, AlphaT AlphaFunction] struct {
	interpolator     InterpolatorT
	gradientFunction GradientT
	alphaFunction    AlphaT
	d1               int // Start distance (subpixel precision)
	d2               int // End distance (subpixel precision)
	downscaleShift   int // Calculated as interpolator.SubpixelShift - GradientSubpixelShift
}

// NewSpanGradientAlpha creates a new alpha gradient span generator.
func NewSpanGradientAlpha[ColorT AlphaType, InterpolatorT SpanInterpolatorInterface, GradientT GradientFunction, AlphaT AlphaFunction](
	interpolator InterpolatorT,
	gradientFunction GradientT,
	alphaFunction AlphaT,
	d1, d2 float64,
) *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT] {
	downscaleShift := interpolator.SubpixelShift() - GradientSubpixelShift
	if downscaleShift < 0 {
		downscaleShift = 0
	}

	return &SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]{
		interpolator:     interpolator,
		gradientFunction: gradientFunction,
		alphaFunction:    alphaFunction,
		d1:               basics.IRound(d1 * GradientSubpixelScale),
		d2:               basics.IRound(d2 * GradientSubpixelScale),
		downscaleShift:   downscaleShift,
	}
}

// Interpolator returns the current interpolator.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) Interpolator() InterpolatorT {
	return sg.interpolator
}

// GradientFunction returns the current gradient function.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) GradientFunction() GradientT {
	return sg.gradientFunction
}

// AlphaFunction returns the current alpha function.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) AlphaFunction() AlphaT {
	return sg.alphaFunction
}

// D1 returns the start distance as a float value.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) D1() float64 {
	return float64(sg.d1) / GradientSubpixelScale
}

// D2 returns the end distance as a float value.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) D2() float64 {
	return float64(sg.d2) / GradientSubpixelScale
}

// SetInterpolator sets a new interpolator.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) SetInterpolator(interpolator InterpolatorT) {
	sg.interpolator = interpolator
	sg.downscaleShift = interpolator.SubpixelShift() - GradientSubpixelShift
	if sg.downscaleShift < 0 {
		sg.downscaleShift = 0
	}
}

// SetGradientFunction sets a new gradient function.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) SetGradientFunction(gradientFunction GradientT) {
	sg.gradientFunction = gradientFunction
}

// SetAlphaFunction sets a new alpha function.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) SetAlphaFunction(alphaFunction AlphaT) {
	sg.alphaFunction = alphaFunction
}

// SetD1 sets the start distance.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) SetD1(d1 float64) {
	sg.d1 = basics.IRound(d1 * GradientSubpixelScale)
}

// SetD2 sets the end distance.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) SetD2(d2 float64) {
	sg.d2 = basics.IRound(d2 * GradientSubpixelScale)
}

// Prepare is called before rendering begins (no-op for basic alpha gradients).
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) Prepare() {
	// No preparation needed for basic alpha gradients
}

// Generate fills a span with alpha gradient values.
// This modifies only the alpha channel of the provided colors.
func (sg *SpanGradientAlpha[ColorT, InterpolatorT, GradientT, AlphaT]) Generate(span []ColorT, x, y, length int) {
	// Calculate distance range
	dd := sg.d2 - sg.d1
	if dd < 1 {
		dd = 1
	}

	// Begin interpolation for this span
	sg.interpolator.Begin(float64(x)+0.5, float64(y)+0.5, length)

	// Generate each pixel in the span
	for i := 0; i < length; i++ {
		// Get transformed coordinates
		ix, iy := sg.interpolator.Coordinates()

		// Calculate gradient distance using the shape function
		d := sg.gradientFunction.Calculate(ix>>sg.downscaleShift, iy>>sg.downscaleShift, sg.d2)

		// Map distance to alpha index
		alphaIndex := ((d - sg.d1) * sg.alphaFunction.Size()) / dd

		// Clamp alpha index to valid range
		if alphaIndex < 0 {
			alphaIndex = 0
		}
		if alphaIndex >= sg.alphaFunction.Size() {
			alphaIndex = sg.alphaFunction.Size() - 1
		}

		// Set only the alpha channel in the span
		span[i].SetAlpha(sg.alphaFunction.AlphaAt(alphaIndex))

		// Advance to next pixel
		sg.interpolator.Next()
	}
}

// Alpha functions

// GradientAlphaLinear implements linear alpha interpolation between two alpha values.
type GradientAlphaLinear struct {
	start basics.Int8u // Start alpha value
	end   basics.Int8u // End alpha value
	size  int          // Number of alpha steps
	mult  float64      // Multiplier for optimization
}

// NewGradientAlphaLinear creates a new linear alpha gradient.
func NewGradientAlphaLinear(start, end basics.Int8u, size int) *GradientAlphaLinear {
	if size <= 0 {
		size = 256
	}

	return &GradientAlphaLinear{
		start: start,
		end:   end,
		size:  size,
		mult:  1.0 / float64(size-1),
	}
}

// Size returns the number of alpha values in the gradient.
func (g *GradientAlphaLinear) Size() int {
	return g.size
}

// AlphaAt returns the alpha value at the specified index.
func (g *GradientAlphaLinear) AlphaAt(index int) basics.Int8u {
	if index <= 0 {
		return g.start
	}
	if index >= g.size-1 {
		return g.end
	}

	t := float64(index) * g.mult
	return basics.Int8u(float64(g.start)*(1-t) + float64(g.end)*t)
}

// SetValues updates the alpha gradient values and size.
func (g *GradientAlphaLinear) SetValues(start, end basics.Int8u, size int) {
	if size <= 0 {
		size = 256
	}
	g.start = start
	g.end = end
	g.size = size
	g.mult = 1.0 / float64(size-1)
}

// GradientAlphaX implements identity alpha function (pass-through).
type GradientAlphaX struct {
	size int
}

// NewGradientAlphaX creates a new identity alpha function.
func NewGradientAlphaX(size int) *GradientAlphaX {
	if size <= 0 {
		size = 256
	}
	return &GradientAlphaX{size: size}
}

// Size returns the number of alpha values.
func (g *GradientAlphaX) Size() int {
	return g.size
}

// AlphaAt returns the index value as alpha (identity function).
func (g *GradientAlphaX) AlphaAt(index int) basics.Int8u {
	if index < 0 {
		return 0
	}
	if index >= g.size {
		return basics.Int8u(g.size - 1)
	}
	return basics.Int8u(index)
}

// GradientAlphaOneMinusX implements inverse alpha function (255-x for 8-bit).
type GradientAlphaOneMinusX struct {
	size int
}

// NewGradientAlphaOneMinusX creates a new inverse alpha function.
func NewGradientAlphaOneMinusX(size int) *GradientAlphaOneMinusX {
	if size <= 0 {
		size = 256
	}
	return &GradientAlphaOneMinusX{size: size}
}

// Size returns the number of alpha values.
func (g *GradientAlphaOneMinusX) Size() int {
	return g.size
}

// AlphaAt returns the inverse alpha value (255-index).
func (g *GradientAlphaOneMinusX) AlphaAt(index int) basics.Int8u {
	if index < 0 {
		return 255
	}
	if index >= g.size {
		return basics.Int8u(255 - (g.size - 1))
	}
	return basics.Int8u(255 - index)
}

// GradientAlphaLUT implements a lookup table for alpha values.
type GradientAlphaLUT struct {
	values []basics.Int8u
}

// NewGradientAlphaLUT creates a new alpha LUT from a slice of values.
func NewGradientAlphaLUT(values []basics.Int8u) *GradientAlphaLUT {
	if len(values) == 0 {
		values = make([]basics.Int8u, 256)
		for i := range values {
			values[i] = basics.Int8u(i)
		}
	}

	// Make a copy to avoid external mutations
	alphaCopy := make([]basics.Int8u, len(values))
	copy(alphaCopy, values)

	return &GradientAlphaLUT{values: alphaCopy}
}

// Size returns the number of alpha values in the LUT.
func (g *GradientAlphaLUT) Size() int {
	return len(g.values)
}

// AlphaAt returns the alpha value at the specified index.
func (g *GradientAlphaLUT) AlphaAt(index int) basics.Int8u {
	if index < 0 {
		return g.values[0]
	}
	if index >= len(g.values) {
		return g.values[len(g.values)-1]
	}
	return g.values[index]
}

// SetAlphaAt sets the alpha value at the specified index.
func (g *GradientAlphaLUT) SetAlphaAt(index int, alpha basics.Int8u) {
	if index >= 0 && index < len(g.values) {
		g.values[index] = alpha
	}
}

// Values returns a copy of the alpha values.
func (g *GradientAlphaLUT) Values() []basics.Int8u {
	result := make([]basics.Int8u, len(g.values))
	copy(result, g.values)
	return result
}

// AlphaType implementations for existing color types

// RGBA8AlphaWrapper wraps RGBA8 to implement AlphaType interface
type RGBA8AlphaWrapper[CS any] struct {
	Color *color.RGBA8[CS]
}

func (w RGBA8AlphaWrapper[CS]) SetAlpha(alpha basics.Int8u) {
	w.Color.A = alpha
}

func (w RGBA8AlphaWrapper[CS]) GetAlpha() basics.Int8u {
	return w.Color.A
}

// Gray8AlphaWrapper wraps Gray8 to implement AlphaType interface
type Gray8AlphaWrapper[CS any] struct {
	Color *color.Gray8[CS]
}

func (w Gray8AlphaWrapper[CS]) SetAlpha(alpha basics.Int8u) {
	w.Color.A = alpha
}

func (w Gray8AlphaWrapper[CS]) GetAlpha() basics.Int8u {
	return w.Color.A
}

// Helper functions to create wrappers
func NewRGBA8AlphaWrapper[CS any](c *color.RGBA8[CS]) RGBA8AlphaWrapper[CS] {
	return RGBA8AlphaWrapper[CS]{Color: c}
}

func NewGray8AlphaWrapper[CS any](c *color.Gray8[CS]) Gray8AlphaWrapper[CS] {
	return Gray8AlphaWrapper[CS]{Color: c}
}

// Helper functions for creating common alpha gradient configurations

// NewLinearAlphaGradientRGBA8 creates a linear alpha gradient for RGBA8 colors.
func NewLinearAlphaGradientRGBA8[InterpolatorT SpanInterpolatorInterface, CS any](
	interpolator InterpolatorT,
	startAlpha, endAlpha basics.Int8u,
	d1, d2 float64,
	size int,
) *SpanGradientAlpha[RGBA8AlphaWrapper[CS], InterpolatorT, GradientLinearX, *GradientAlphaLinear] {
	if size <= 0 {
		size = 256
	}

	gradientFunc := GradientLinearX{}
	alphaFunc := NewGradientAlphaLinear(startAlpha, endAlpha, size)

	return NewSpanGradientAlpha[RGBA8AlphaWrapper[CS], InterpolatorT, GradientLinearX, *GradientAlphaLinear](
		interpolator, gradientFunc, alphaFunc, d1, d2)
}

// NewRadialAlphaGradientRGBA8 creates a radial alpha gradient for RGBA8 colors.
func NewRadialAlphaGradientRGBA8[InterpolatorT SpanInterpolatorInterface, CS any](
	interpolator InterpolatorT,
	startAlpha, endAlpha basics.Int8u,
	d1, d2 float64,
	size int,
) *SpanGradientAlpha[RGBA8AlphaWrapper[CS], InterpolatorT, GradientRadial, *GradientAlphaLinear] {
	if size <= 0 {
		size = 256
	}

	gradientFunc := GradientRadial{}
	alphaFunc := NewGradientAlphaLinear(startAlpha, endAlpha, size)

	return NewSpanGradientAlpha[RGBA8AlphaWrapper[CS], InterpolatorT, GradientRadial, *GradientAlphaLinear](
		interpolator, gradientFunc, alphaFunc, d1, d2)
}
