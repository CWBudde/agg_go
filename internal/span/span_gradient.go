// Package span provides gradient span generation functionality for AGG.
// This implements a port of AGG's span_gradient classes and functions.
package span

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// Gradient subpixel precision constants
const (
	GradientSubpixelShift = 4                          // 4 bits of precision
	GradientSubpixelScale = 1 << GradientSubpixelShift // 16x precision
	GradientSubpixelMask  = GradientSubpixelScale - 1  // Masking value (15)
)

// GradientFunction defines the interface for gradient shape functions.
// These functions calculate the gradient distance for given coordinates.
type GradientFunction interface {
	// Calculate computes gradient distance at coordinates (x, y) with maximum distance d2
	Calculate(x, y, d2 int) int
}

// ColorFunction defines the interface for gradient color functions.
// These functions provide color lookup based on gradient position.
type ColorFunction[ColorT any] interface {
	// Size returns the number of colors in the gradient
	Size() int

	// ColorAt returns the color at the specified index
	ColorAt(index int) ColorT
}

// SpanGradient generates gradient-filled pixel spans.
// This is a port of AGG's span_gradient template class.
type SpanGradient[ColorT any, InterpolatorT SpanInterpolatorInterface, GradientT GradientFunction, ColorT2 ColorFunction[ColorT]] struct {
	interpolator     InterpolatorT
	gradientFunction GradientT
	colorFunction    ColorT2
	d1               int // Start distance (subpixel precision)
	d2               int // End distance (subpixel precision)
	downscaleShift   int // Calculated as interpolator.SubpixelShift - GradientSubpixelShift
}

// NewSpanGradient creates a new gradient span generator.
func NewSpanGradient[ColorT any, InterpolatorT SpanInterpolatorInterface, GradientT GradientFunction, ColorT2 ColorFunction[ColorT]](
	interpolator InterpolatorT,
	gradientFunction GradientT,
	colorFunction ColorT2,
	d1, d2 float64,
) *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2] {
	downscaleShift := interpolator.SubpixelShift() - GradientSubpixelShift
	if downscaleShift < 0 {
		downscaleShift = 0
	}

	return &SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]{
		interpolator:     interpolator,
		gradientFunction: gradientFunction,
		colorFunction:    colorFunction,
		d1:               basics.IRound(d1 * GradientSubpixelScale),
		d2:               basics.IRound(d2 * GradientSubpixelScale),
		downscaleShift:   downscaleShift,
	}
}

// Interpolator returns the current interpolator.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) Interpolator() InterpolatorT {
	return sg.interpolator
}

// GradientFunction returns the current gradient function.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) GradientFunction() GradientT {
	return sg.gradientFunction
}

// ColorFunction returns the current color function.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) ColorFunction() ColorT2 {
	return sg.colorFunction
}

// D1 returns the start distance as a float value.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) D1() float64 {
	return float64(sg.d1) / GradientSubpixelScale
}

// D2 returns the end distance as a float value.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) D2() float64 {
	return float64(sg.d2) / GradientSubpixelScale
}

// SetInterpolator sets a new interpolator.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) SetInterpolator(interpolator InterpolatorT) {
	sg.interpolator = interpolator
	sg.downscaleShift = interpolator.SubpixelShift() - GradientSubpixelShift
	if sg.downscaleShift < 0 {
		sg.downscaleShift = 0
	}
}

// SetGradientFunction sets a new gradient function.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) SetGradientFunction(gradientFunction GradientT) {
	sg.gradientFunction = gradientFunction
}

// SetColorFunction sets a new color function.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) SetColorFunction(colorFunction ColorT2) {
	sg.colorFunction = colorFunction
}

// SetD1 sets the start distance.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) SetD1(d1 float64) {
	sg.d1 = basics.IRound(d1 * GradientSubpixelScale)
}

// SetD2 sets the end distance.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) SetD2(d2 float64) {
	sg.d2 = basics.IRound(d2 * GradientSubpixelScale)
}

// Prepare is called before rendering begins (no-op for basic gradients).
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) Prepare() {
	// No preparation needed for basic gradients
}

// Generate fills a span with gradient colors.
// This is the core method that produces gradient-filled spans.
func (sg *SpanGradient[ColorT, InterpolatorT, GradientT, ColorT2]) Generate(span []ColorT, x, y, length int) {
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

		// Map distance to color index
		colorIndex := ((d - sg.d1) * sg.colorFunction.Size()) / dd

		// Clamp color index to valid range
		if colorIndex < 0 {
			colorIndex = 0
		}
		if colorIndex >= sg.colorFunction.Size() {
			colorIndex = sg.colorFunction.Size() - 1
		}

		// Set the color in the span
		span[i] = sg.colorFunction.ColorAt(colorIndex)

		// Advance to next pixel
		sg.interpolator.Next()
	}
}

// Gradient shape functions

// GradientLinearX implements a horizontal linear gradient.
type GradientLinearX struct{}

func (g GradientLinearX) Calculate(x, y, d2 int) int {
	return x
}

// GradientLinearY implements a vertical linear gradient.
type GradientLinearY struct{}

func (g GradientLinearY) Calculate(x, y, d2 int) int {
	return y
}

// GradientRadial implements a circular/radial gradient.
type GradientRadial struct{}

func (g GradientRadial) Calculate(x, y, d2 int) int {
	return int(basics.FastSqrt(uint32(x*x + y*y)))
}

// GradientRadialDouble implements a radial gradient with double precision.
type GradientRadialDouble struct{}

func (g GradientRadialDouble) Calculate(x, y, d2 int) int {
	return int(basics.URound(math.Sqrt(float64(x)*float64(x) + float64(y)*float64(y))))
}

// GradientRadialFocus implements a radial gradient with a focal point.
type GradientRadialFocus struct {
	r   int     // Radius (subpixel precision)
	fx  int     // Focus X (subpixel precision)
	fy  int     // Focus Y (subpixel precision)
	r2  float64 // r squared (cache)
	fx2 float64 // fx squared (cache)
	fy2 float64 // fy squared (cache)
	mul float64 // multiplier (cache)
}

// NewGradientRadialFocus creates a new radial focus gradient.
func NewGradientRadialFocus(r, fx, fy float64) *GradientRadialFocus {
	g := &GradientRadialFocus{}
	g.Init(r, fx, fy)
	return g
}

// Init initializes the radial focus gradient parameters.
func (g *GradientRadialFocus) Init(r, fx, fy float64) {
	g.r = basics.IRound(r * GradientSubpixelScale)
	g.fx = basics.IRound(fx * GradientSubpixelScale)
	g.fy = basics.IRound(fy * GradientSubpixelScale)
	g.updateValues()
}

func (g *GradientRadialFocus) updateValues() {
	// Calculate invariant values
	g.r2 = float64(g.r) * float64(g.r)
	g.fx2 = float64(g.fx) * float64(g.fx)
	g.fy2 = float64(g.fy) * float64(g.fy)
	d := g.r2 - (g.fx2 + g.fy2)

	// Avoid degenerate case where focal point is on the circle
	if d == 0 {
		if g.fx != 0 {
			if g.fx < 0 {
				g.fx++
			} else {
				g.fx--
			}
		}
		if g.fy != 0 {
			if g.fy < 0 {
				g.fy++
			} else {
				g.fy--
			}
		}
		g.fx2 = float64(g.fx) * float64(g.fx)
		g.fy2 = float64(g.fy) * float64(g.fy)
		d = g.r2 - (g.fx2 + g.fy2)
	}
	g.mul = float64(g.r) / d
}

func (g *GradientRadialFocus) Calculate(x, y, d2 int) int {
	dx := float64(x - g.fx)
	dy := float64(y - g.fy)
	d2Calc := dx*float64(g.fy) - dy*float64(g.fx)
	d3 := g.r2*(dx*dx+dy*dy) - d2Calc*d2Calc
	return basics.IRound((dx*float64(g.fx) + dy*float64(g.fy) + math.Sqrt(math.Abs(d3))) * g.mul)
}

// Radius returns the gradient radius.
func (g *GradientRadialFocus) Radius() float64 {
	return float64(g.r) / GradientSubpixelScale
}

// FocusX returns the focus X coordinate.
func (g *GradientRadialFocus) FocusX() float64 {
	return float64(g.fx) / GradientSubpixelScale
}

// FocusY returns the focus Y coordinate.
func (g *GradientRadialFocus) FocusY() float64 {
	return float64(g.fy) / GradientSubpixelScale
}

// GradientDiamond implements a diamond-shaped gradient.
type GradientDiamond struct{}

func (g GradientDiamond) Calculate(x, y, d2 int) int {
	ax := basics.Abs(x)
	ay := basics.Abs(y)
	if ax > ay {
		return ax
	}
	return ay
}

// GradientXY implements an XY product gradient.
type GradientXY struct{}

func (g GradientXY) Calculate(x, y, d2 int) int {
	return basics.Abs(x) * basics.Abs(y) / d2
}

// GradientSqrtXY implements a square root XY gradient.
type GradientSqrtXY struct{}

func (g GradientSqrtXY) Calculate(x, y, d2 int) int {
	return int(basics.FastSqrt(uint32(basics.Abs(x) * basics.Abs(y))))
}

// GradientConic implements a conic (angular) gradient.
type GradientConic struct{}

func (g GradientConic) Calculate(x, y, d2 int) int {
	return int(basics.URound(math.Abs(math.Atan2(float64(y), float64(x))) * float64(d2) / math.Pi))
}

// Gradient wrapper adaptors for repeat and reflect modes

// GradientRepeatAdaptor wraps a gradient function to repeat beyond the gradient range.
type GradientRepeatAdaptor[GT GradientFunction] struct {
	gradient GT
}

// NewGradientRepeatAdaptor creates a new repeat adaptor.
func NewGradientRepeatAdaptor[GT GradientFunction](gradient GT) *GradientRepeatAdaptor[GT] {
	return &GradientRepeatAdaptor[GT]{gradient: gradient}
}

func (g *GradientRepeatAdaptor[GT]) Calculate(x, y, d int) int {
	ret := g.gradient.Calculate(x, y, d) % d
	if ret < 0 {
		ret += d
	}
	return ret
}

// GradientReflectAdaptor wraps a gradient function to reflect beyond the gradient range.
type GradientReflectAdaptor[GT GradientFunction] struct {
	gradient GT
}

// NewGradientReflectAdaptor creates a new reflect adaptor.
func NewGradientReflectAdaptor[GT GradientFunction](gradient GT) *GradientReflectAdaptor[GT] {
	return &GradientReflectAdaptor[GT]{gradient: gradient}
}

func (g *GradientReflectAdaptor[GT]) Calculate(x, y, d int) int {
	d2 := d << 1
	ret := g.gradient.Calculate(x, y, d) % d2
	if ret < 0 {
		ret += d2
	}
	if ret >= d {
		ret = d2 - ret
	}
	return ret
}

// Color functions

// GradientLinearColorRGBA implements linear color interpolation for RGBA colors.
type GradientLinearColorRGBA struct {
	c1   color.RGBA // Start color
	c2   color.RGBA // End color
	size int        // Number of color steps
	mult float64    // Multiplier for optimization
}

// NewGradientLinearColorRGBA creates a new linear RGBA color gradient.
func NewGradientLinearColorRGBA(c1, c2 color.RGBA, size int) *GradientLinearColorRGBA {
	if size <= 0 {
		size = 256
	}

	return &GradientLinearColorRGBA{
		c1:   c1,
		c2:   c2,
		size: size,
		mult: 1.0 / float64(size-1),
	}
}

// Size returns the number of colors in the gradient.
func (g *GradientLinearColorRGBA) Size() int {
	return g.size
}

// ColorAt returns the color at the specified index.
func (g *GradientLinearColorRGBA) ColorAt(index int) color.RGBA {
	return g.c1.Gradient(g.c2, float64(index)*g.mult)
}

// GradientLinearColorRGBA8 implements linear color interpolation for RGBA8 colors.
type GradientLinearColorRGBA8[CS ColorSpace] struct {
	c1   color.RGBA8[CS] // Start color
	c2   color.RGBA8[CS] // End color
	size int             // Number of color steps
	mult float64         // Multiplier for optimization
}

// NewGradientLinearColorRGBA8 creates a new linear RGBA8 color gradient.
func NewGradientLinearColorRGBA8[CS ColorSpace](c1, c2 color.RGBA8[CS], size int) *GradientLinearColorRGBA8[CS] {
	if size <= 0 {
		size = 256
	}

	return &GradientLinearColorRGBA8[CS]{
		c1:   c1,
		c2:   c2,
		size: size,
		mult: 1.0 / float64(size-1),
	}
}

// Size returns the number of colors in the gradient.
func (g *GradientLinearColorRGBA8[CS]) Size() int {
	return g.size
}

// ColorAt returns the color at the specified index.
func (g *GradientLinearColorRGBA8[CS]) ColorAt(index int) color.RGBA8[CS] {
	k := basics.Int8u(float64(index) * g.mult * 255.0)
	return g.c1.Gradient(g.c2, k)
}

// SetColors updates the RGBA gradient colors and size.
func (g *GradientLinearColorRGBA) SetColors(c1, c2 color.RGBA, size int) {
	if size <= 0 {
		size = 256
	}
	g.c1 = c1
	g.c2 = c2
	g.size = size
	g.mult = 1.0 / float64(size-1)
}

// SetColors updates the RGBA8 gradient colors and size.
func (g *GradientLinearColorRGBA8[CS]) SetColors(c1, c2 color.RGBA8[CS], size int) {
	if size <= 0 {
		size = 256
	}
	g.c1 = c1
	g.c2 = c2
	g.size = size
	g.mult = 1.0 / float64(size-1)
}

// Helper functions for creating common gradient configurations

// NewLinearGradientRGBA8 creates a linear RGBA8 gradient span generator.
func NewLinearGradientRGBA8[InterpolatorT SpanInterpolatorInterface](
	interpolator InterpolatorT,
	startColor, endColor color.RGBA8[color.Linear],
	d1, d2 float64,
	size int,
) *SpanGradient[color.RGBA8[color.Linear], InterpolatorT, GradientLinearX, *GradientLinearColorRGBA8[color.Linear]] {
	if size <= 0 {
		size = 256
	}

	gradientFunc := GradientLinearX{}
	colorFunc := NewGradientLinearColorRGBA8(startColor, endColor, size)

	return NewSpanGradient[color.RGBA8[color.Linear], InterpolatorT, GradientLinearX, *GradientLinearColorRGBA8[color.Linear]](
		interpolator, gradientFunc, colorFunc, d1, d2)
}

// NewRadialGradientRGBA8 creates a radial RGBA8 gradient span generator.
func NewRadialGradientRGBA8[InterpolatorT SpanInterpolatorInterface](
	interpolator InterpolatorT,
	startColor, endColor color.RGBA8[color.Linear],
	d1, d2 float64,
	size int,
) *SpanGradient[color.RGBA8[color.Linear], InterpolatorT, GradientRadial, *GradientLinearColorRGBA8[color.Linear]] {
	if size <= 0 {
		size = 256
	}

	gradientFunc := GradientRadial{}
	colorFunc := NewGradientLinearColorRGBA8(startColor, endColor, size)

	return NewSpanGradient[color.RGBA8[color.Linear], InterpolatorT, GradientRadial, *GradientLinearColorRGBA8[color.Linear]](
		interpolator, gradientFunc, colorFunc, d1, d2)
}
