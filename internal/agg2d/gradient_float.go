// Package agg2d float gradient setup (L5). Float twin of gradient.go: gradient
// profiles are interpolated in float (RGBA32) space and stored directly in the
// float gradient array, then copied into the float gradient LUT at render time.
package agg2d

import (
	"math"

	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/transform"
)

// buildProfileGradient32 fills a float gradient array with a two-color profile.
func buildProfileGradient32(dst *[256]color.RGBA32[color.Linear], c1, c2 Color, startGradient, endGradient int) {
	if endGradient <= startGradient {
		endGradient = startGradient + 1
	}
	f1 := colorToRGBA32(c1)
	f2 := colorToRGBA32(c2)
	k := float32(1.0) / float32(endGradient-startGradient)

	for i := 0; i < startGradient && i < 256; i++ {
		dst[i] = f1
	}
	for i := startGradient; i < endGradient && i < 256; i++ {
		if i < 0 {
			continue
		}
		dst[i] = f1.Gradient(f2, float32(i-startGradient)*k)
	}
	for i := endGradient; i < 256; i++ {
		if i < 0 {
			continue
		}
		dst[i] = f2
	}
}

// buildThreeColorGradient32 fills a float gradient array with a three-color ramp.
func buildThreeColorGradient32(dst *[256]color.RGBA32[color.Linear], c1, c2, c3 Color) {
	f1, f2, f3 := colorToRGBA32(c1), colorToRGBA32(c2), colorToRGBA32(c3)
	for i := range 128 {
		dst[i] = f1.Gradient(f2, float32(i)/127.0)
	}
	for i := 128; i < 256; i++ {
		dst[i] = f2.Gradient(f3, float32(i-128)/127.0)
	}
}

func (a *Agg2DFloat) setupWorldRadialGradient(matrix *transform.TransAffine, x, y, r float64) (d1, d2 float64) {
	screenRadius := a.WorldToScreenScalar(r)
	screenX, screenY := x, y
	a.WorldToScreen(&screenX, &screenY)
	matrix.Reset()
	matrix.Translate(screenX, screenY)
	matrix.Invert()
	return 0.0, screenRadius
}

// FillLinearGradient sets up a linear gradient for fill operations.
func (a *Agg2DFloat) FillLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	buildProfileGradient32(&a.fillGradient, c1, c2, 128-int(profile*127.0), 128+int(profile*127.0))
	a.fillGradientLUTDirty = true

	angle := math.Atan2(y2-y1, x2-x1)
	a.fillGradientMatrix.Reset()
	a.fillGradientMatrix.Rotate(angle)
	a.fillGradientMatrix.Translate(x1, y1)
	a.fillGradientMatrix.Multiply(a.transform)
	a.fillGradientMatrix.Invert()

	a.fillGradientD1 = 0.0
	a.fillGradientD2 = math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
	a.fillGradientFlag = Linear
	a.fillColor = NewColor(0, 0, 0, 255)
}

// FillLinearGradientStops sets up a linear fill gradient from an arbitrary sorted
// slice of ColorStops (Position 0 = start point, Position 1 = end point). It is
// the float twin of Agg2D.FillLinearGradientStops, building the LUT from every
// stop instead of a two-color profile.
func (a *Agg2DFloat) FillLinearGradientStops(x1, y1, x2, y2 float64, stops []ColorStop) {
	buildNStopGradient32(&a.fillGradient, stops)
	a.fillGradientLUTDirty = true

	angle := math.Atan2(y2-y1, x2-x1)
	a.fillGradientMatrix.Reset()
	a.fillGradientMatrix.Rotate(angle)
	a.fillGradientMatrix.Translate(x1, y1)
	a.fillGradientMatrix.Multiply(a.transform)
	a.fillGradientMatrix.Invert()

	a.fillGradientD1 = 0.0
	a.fillGradientD2 = math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
	a.fillGradientFlag = Linear
	a.fillColor = NewColor(0, 0, 0, 255)
}

// LineLinearGradient sets up a linear gradient for stroke operations.
func (a *Agg2DFloat) LineLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	buildProfileGradient32(&a.lineGradient, c1, c2, 128-int(profile*128.0), 128+int(profile*128.0))
	a.lineGradientLUTDirty = true

	angle := math.Atan2(y2-y1, x2-x1)
	a.lineGradientMatrix.Reset()
	a.lineGradientMatrix.Rotate(angle)
	a.lineGradientMatrix.Translate(x1, y1)
	a.lineGradientMatrix.Multiply(a.transform)
	a.lineGradientMatrix.Invert()

	a.lineGradientD1 = 0.0
	a.lineGradientD2 = math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
	a.lineGradientFlag = Linear
	a.lineColor = NewColor(0, 0, 0, 255)
}

// FillRadialGradient sets up a two-color radial gradient for fill operations.
func (a *Agg2DFloat) FillRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	buildProfileGradient32(&a.fillGradient, c1, c2, 128-int(profile*127.0), 128+int(profile*127.0))
	a.fillGradientLUTDirty = true
	a.fillGradientD1, a.fillGradientD2 = a.setupWorldRadialGradient(a.fillGradientMatrix, x, y, r)
	a.fillGradientFlag = Radial
	a.fillColor = NewColor(0, 0, 0, 255)
}

// LineRadialGradient sets up a two-color radial gradient for stroke operations.
func (a *Agg2DFloat) LineRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	buildProfileGradient32(&a.lineGradient, c1, c2, 128-int(profile*128.0), 128+int(profile*128.0))
	a.lineGradientLUTDirty = true
	a.lineGradientD1, a.lineGradientD2 = a.setupWorldRadialGradient(a.lineGradientMatrix, x, y, r)
	a.lineGradientFlag = Radial
	a.lineColor = NewColor(0, 0, 0, 255)
}

// buildNStopGradient32 fills a float gradient array by interpolating between an
// arbitrary number of color stops. The stops slice must be sorted by ascending
// Position; out-of-range positions are clamped to the nearest stop. Float twin
// of buildNStopGradient.
func buildNStopGradient32(dst *[256]color.RGBA32[color.Linear], stops []ColorStop) {
	if len(stops) == 0 {
		return
	}
	if len(stops) == 1 {
		c := colorToRGBA32(stops[0].Color)
		for i := range dst {
			dst[i] = c
		}
		return
	}
	first := colorToRGBA32(stops[0].Color)
	last := colorToRGBA32(stops[len(stops)-1].Color)
	for i := range 256 {
		pos := float64(i) / 255.0
		if pos <= stops[0].Position {
			dst[i] = first
			continue
		}
		if pos >= stops[len(stops)-1].Position {
			dst[i] = last
			continue
		}
		// Binary search for the bracketing stops.
		lo, hi := 0, len(stops)-1
		for hi-lo > 1 {
			mid := (lo + hi) / 2
			if stops[mid].Position <= pos {
				lo = mid
			} else {
				hi = mid
			}
		}
		span := stops[hi].Position - stops[lo].Position
		var t float32
		if span > 0 {
			t = float32((pos - stops[lo].Position) / span)
		}
		dst[i] = colorToRGBA32(stops[lo].Color).Gradient(colorToRGBA32(stops[hi].Color), t)
	}
}

// FillRadialGradientStops sets up a radial fill gradient from an arbitrary sorted
// slice of ColorStops (position 0 = centre, position 1 = edge).
func (a *Agg2DFloat) FillRadialGradientStops(x, y, r float64, stops []ColorStop) {
	buildNStopGradient32(&a.fillGradient, stops)
	a.fillGradientLUTDirty = true
	a.fillGradientD1, a.fillGradientD2 = a.setupWorldRadialGradient(a.fillGradientMatrix, x, y, r)
	a.fillGradientFlag = Radial
	a.fillColor = NewColor(0, 0, 0, 255)
}

// FillRadialGradientMultiStop sets up a three-color radial gradient for fill.
func (a *Agg2DFloat) FillRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	buildThreeColorGradient32(&a.fillGradient, c1, c2, c3)
	a.fillGradientLUTDirty = true
	a.fillGradientD1, a.fillGradientD2 = a.setupWorldRadialGradient(a.fillGradientMatrix, x, y, r)
	a.fillGradientFlag = Radial
	a.fillColor = NewColor(0, 0, 0, 255)
}

// LineRadialGradientMultiStop sets up a three-color radial gradient for strokes.
func (a *Agg2DFloat) LineRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	buildThreeColorGradient32(&a.lineGradient, c1, c2, c3)
	a.lineGradientLUTDirty = true
	a.lineGradientD1, a.lineGradientD2 = a.setupWorldRadialGradient(a.lineGradientMatrix, x, y, r)
	a.lineGradientFlag = Radial
	a.lineColor = NewColor(0, 0, 0, 255)
}

// FillRadialGradientPos repositions the fill radial gradient (centre and radius)
// without changing the color ramp.
func (a *Agg2DFloat) FillRadialGradientPos(x, y, r float64) {
	a.fillGradientD1, a.fillGradientD2 = a.setupWorldRadialGradient(a.fillGradientMatrix, x, y, r)
}

// LineRadialGradientPos repositions the line radial gradient (centre and radius)
// without changing the color ramp.
func (a *Agg2DFloat) LineRadialGradientPos(x, y, r float64) {
	a.lineGradientD1, a.lineGradientD2 = a.setupWorldRadialGradient(a.lineGradientMatrix, x, y, r)
}

// FillGradientD1 returns the fill gradient start distance.
func (a *Agg2DFloat) FillGradientD1() float64 { return a.fillGradientD1 }

// FillGradientD2 returns the fill gradient end distance.
func (a *Agg2DFloat) FillGradientD2() float64 { return a.fillGradientD2 }

// LineGradientD1 returns the line gradient start distance.
func (a *Agg2DFloat) LineGradientD1() float64 { return a.lineGradientD1 }

// LineGradientD2 returns the line gradient end distance.
func (a *Agg2DFloat) LineGradientD2() float64 { return a.lineGradientD2 }

// FillGradientFlag returns the current fill gradient type.
func (a *Agg2DFloat) FillGradientFlag() Gradient { return a.fillGradientFlag }

// LineGradientFlag returns the current line gradient type.
func (a *Agg2DFloat) LineGradientFlag() Gradient { return a.lineGradientFlag }
