// Package agg provides gradient functionality for the AGG2D high-level interface.
// This implements Phase 4: Gradient Support from the AGG 2.6 C++ library.
package agg2d

import (
	"math"

	"agg_go/internal/transform"
)

func buildProfileGradient(dst *[256]Color, c1, c2 Color, startGradient, endGradient int) {
	if endGradient <= startGradient {
		endGradient = startGradient + 1
	}
	k := 1.0 / float64(endGradient-startGradient)

	for i := 0; i < startGradient; i++ {
		dst[i] = c1
	}
	for i := startGradient; i < endGradient; i++ {
		dst[i] = c1.Gradient(c2, float64(i-startGradient)*k)
	}
	for i := endGradient; i < 256; i++ {
		dst[i] = c2
	}
}

func buildThreeColorGradient(dst *[256]Color, c1, c2, c3 Color) {
	for i := 0; i < 128; i++ {
		dst[i] = c1.Gradient(c2, float64(i)/127.0)
	}
	for i := 128; i < 256; i++ {
		dst[i] = c2.Gradient(c3, float64(i-128)/127.0)
	}
}

func setupRadialGradient(matrix *transform.TransAffine, x, y, r float64) (d1, d2 float64) {
	matrix.Reset()
	matrix.Translate(x, y)
	matrix.Invert()
	return 0.0, r
}

func (agg2d *Agg2D) setupWorldRadialGradient(matrix *transform.TransAffine, x, y, r float64) (d1, d2 float64) {
	screenRadius := agg2d.worldToScreen(r)
	screenX, screenY := x, y
	agg2d.worldToScreenPoint(&screenX, &screenY)
	return setupRadialGradient(matrix, screenX, screenY, screenRadius)
}

// FillLinearGradient sets up a linear gradient for fill operations.
// Parameters:
//   - x1, y1: Starting point of the gradient line
//   - x2, y2: Ending point of the gradient line
//   - c1: Color at the starting point
//   - c2: Color at the ending point
//   - profile: Gradient profile (0.0-1.0), controls transition sharpness
//   - 1.0 = linear transition (default)
//   - < 1.0 = sharper transition concentrated in the middle
//   - > 1.0 would create a softer transition (clamped to valid range)
//
// This matches the C++ Agg2D::fillLinearGradient method.
func (agg2d *Agg2D) FillLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	buildProfileGradient(&agg2d.fillGradient, c1, c2, 128-int(profile*127.0), 128+int(profile*127.0))

	// Calculate gradient angle and setup transformation matrix
	angle := math.Atan2(y2-y1, x2-x1)

	// Reset and setup the gradient transformation matrix
	agg2d.fillGradientMatrix.Reset()
	agg2d.fillGradientMatrix.Rotate(angle)
	agg2d.fillGradientMatrix.Translate(x1, y1)
	agg2d.fillGradientMatrix.Multiply(agg2d.transform)
	agg2d.fillGradientMatrix.Invert()

	// Set gradient parameters
	agg2d.fillGradientD1 = 0.0
	agg2d.fillGradientD2 = math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
	agg2d.fillGradientFlag = Linear

	// Set fill color to a placeholder (gradient takes precedence)
	agg2d.fillColor = NewColor(0, 0, 0, 255)
}

// LineLinearGradient sets up a linear gradient for line/stroke operations.
// Parameters are identical to FillLinearGradient but affect line rendering.
// This matches the C++ Agg2D::lineLinearGradient method.
func (agg2d *Agg2D) LineLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	buildProfileGradient(&agg2d.lineGradient, c1, c2, 128-int(profile*128.0), 128+int(profile*128.0))

	// Calculate gradient angle and setup transformation matrix
	angle := math.Atan2(y2-y1, x2-x1)

	// Reset and setup the gradient transformation matrix
	agg2d.lineGradientMatrix.Reset()
	agg2d.lineGradientMatrix.Rotate(angle)
	agg2d.lineGradientMatrix.Translate(x1, y1)
	agg2d.lineGradientMatrix.Multiply(agg2d.transform)
	agg2d.lineGradientMatrix.Invert()

	// Set gradient parameters
	agg2d.lineGradientD1 = 0.0
	agg2d.lineGradientD2 = math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
	agg2d.lineGradientFlag = Linear

	// Set line color to a placeholder (gradient takes precedence)
	agg2d.lineColor = NewColor(0, 0, 0, 255)
}

// FillRadialGradient sets up a radial gradient for fill operations.
// Parameters:
//   - x, y: Center point of the radial gradient
//   - r: Radius of the gradient
//   - c1: Color at the center
//   - c2: Color at the edge
//   - profile: Gradient profile (0.0-1.0), controls transition sharpness
//
// This matches the C++ Agg2D::fillRadialGradient method.
func (agg2d *Agg2D) FillRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	buildProfileGradient(&agg2d.fillGradient, c1, c2, 128-int(profile*127.0), 128+int(profile*127.0))
	agg2d.fillGradientD1, agg2d.fillGradientD2 = agg2d.setupWorldRadialGradient(agg2d.fillGradientMatrix, x, y, r)
	agg2d.fillGradientFlag = Radial
	agg2d.fillColor = NewColor(0, 0, 0, 255)
}

// LineRadialGradient sets up a radial gradient for line/stroke operations.
// Parameters are identical to FillRadialGradient but affect line rendering.
// This matches the C++ Agg2D::lineRadialGradient method.
func (agg2d *Agg2D) LineRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	buildProfileGradient(&agg2d.lineGradient, c1, c2, 128-int(profile*128.0), 128+int(profile*128.0))
	agg2d.lineGradientD1, agg2d.lineGradientD2 = agg2d.setupWorldRadialGradient(agg2d.lineGradientMatrix, x, y, r)
	agg2d.lineGradientFlag = Radial
	agg2d.lineColor = NewColor(0, 0, 0, 255)
}

// FillRadialGradientMultiStop sets up a radial gradient with three colors.
// This creates a gradient from c1 (center) to c2 (middle) to c3 (edge).
// The transition points are fixed at 50% intervals.
// This matches the C++ Agg2D::fillRadialGradient(x, y, r, c1, c2, c3) method.
func (agg2d *Agg2D) FillRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	buildThreeColorGradient(&agg2d.fillGradient, c1, c2, c3)
	agg2d.fillGradientD1, agg2d.fillGradientD2 = agg2d.setupWorldRadialGradient(agg2d.fillGradientMatrix, x, y, r)
	agg2d.fillGradientFlag = Radial
	agg2d.fillColor = NewColor(0, 0, 0, 255)
}

// LineRadialGradientMultiStop sets up a radial gradient with three colors for line operations.
// This matches the C++ Agg2D::lineRadialGradient(x, y, r, c1, c2, c3) method.
func (agg2d *Agg2D) LineRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	buildThreeColorGradient(&agg2d.lineGradient, c1, c2, c3)
	agg2d.lineGradientD1, agg2d.lineGradientD2 = agg2d.setupWorldRadialGradient(agg2d.lineGradientMatrix, x, y, r)
	agg2d.lineGradientFlag = Radial
	agg2d.lineColor = NewColor(0, 0, 0, 255)
}

// FillRadialGradientPos sets up the position and radius for fill radial gradient
// without changing the colors. This matches the C++ Agg2D::fillRadialGradient(x, y, r) method.
func (agg2d *Agg2D) FillRadialGradientPos(x, y, r float64) {
	agg2d.fillGradientD1, agg2d.fillGradientD2 = agg2d.setupWorldRadialGradient(agg2d.fillGradientMatrix, x, y, r)
}

// LineRadialGradientPos sets up the position and radius for line radial gradient
// without changing the colors. This matches the C++ Agg2D::lineRadialGradient(x, y, r) method.
func (agg2d *Agg2D) LineRadialGradientPos(x, y, r float64) {
	agg2d.lineGradientD1, agg2d.lineGradientD2 = agg2d.setupWorldRadialGradient(agg2d.lineGradientMatrix, x, y, r)
}

// Helper methods for coordinate transformation

// worldToScreen converts a world coordinate distance to screen coordinates
func (agg2d *Agg2D) worldToScreen(distance float64) float64 {
	return agg2d.worldToScreenScalar(distance)
}

// worldToScreenPoint converts world coordinates to screen coordinates
func (agg2d *Agg2D) worldToScreenPoint(x, y *float64) {
	agg2d.WorldToScreen(x, y)
}

// Accessor methods for gradient parameters

// FillGradientD1 returns the fill gradient start distance
func (agg2d *Agg2D) FillGradientD1() float64 {
	return agg2d.fillGradientD1
}

// FillGradientD2 returns the fill gradient end distance
func (agg2d *Agg2D) FillGradientD2() float64 {
	return agg2d.fillGradientD2
}

// LineGradientD1 returns the line gradient start distance
func (agg2d *Agg2D) LineGradientD1() float64 {
	return agg2d.lineGradientD1
}

// LineGradientD2 returns the line gradient end distance
func (agg2d *Agg2D) LineGradientD2() float64 {
	return agg2d.lineGradientD2
}

// FillGradientFlag returns the current fill gradient type
func (agg2d *Agg2D) FillGradientFlag() Gradient {
	return agg2d.fillGradientFlag
}

// LineGradientFlag returns the current line gradient type
func (agg2d *Agg2D) LineGradientFlag() Gradient {
	return agg2d.lineGradientFlag
}
