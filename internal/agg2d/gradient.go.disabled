// Package agg provides gradient functionality for the AGG2D high-level interface.
// This implements Phase 4: Gradient Support from the AGG 2.6 C++ library.
package agg2d

import (
	"math"
)

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
	// Calculate gradient profile bounds
	// Profile determines the distribution of the gradient:
	// The gradient array has 256 entries indexed 0-255
	// startGradient and endGradient define the range where interpolation occurs
	startGradient := 128 - int(profile*127.0)
	endGradient := 128 + int(profile*127.0)

	// Ensure we have at least one interpolation step
	if endGradient <= startGradient {
		endGradient = startGradient + 1
	}

	// Calculate interpolation step
	k := 1.0 / float64(endGradient-startGradient)

	// Fill the gradient array
	// Before startGradient: solid c1 color
	for i := 0; i < startGradient; i++ {
		agg2d.fillGradient[i] = c1
	}

	// Between startGradient and endGradient: interpolated colors
	for i := startGradient; i < endGradient; i++ {
		factor := float64(i-startGradient) * k
		agg2d.fillGradient[i] = c1.Gradient(c2, factor)
	}

	// After endGradient: solid c2 color
	for i := endGradient; i < 256; i++ {
		agg2d.fillGradient[i] = c2
	}

	// Calculate gradient angle and setup transformation matrix
	angle := math.Atan2(y2-y1, x2-x1)

	// Reset and setup the gradient transformation matrix
	agg2d.fillGradientMatrix.Reset()
	agg2d.fillGradientMatrix.Rotate(angle)
	agg2d.fillGradientMatrix.Translate(x1, y1)
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
	// Calculate gradient profile bounds
	startGradient := 128 - int(profile*127.0)
	endGradient := 128 + int(profile*127.0)

	// Ensure we have at least one interpolation step
	if endGradient <= startGradient {
		endGradient = startGradient + 1
	}

	// Calculate interpolation step
	k := 1.0 / float64(endGradient-startGradient)

	// Fill the gradient array
	// Before startGradient: solid c1 color
	for i := 0; i < startGradient; i++ {
		agg2d.lineGradient[i] = c1
	}

	// Between startGradient and endGradient: interpolated colors
	for i := startGradient; i < endGradient; i++ {
		factor := float64(i-startGradient) * k
		agg2d.lineGradient[i] = c1.Gradient(c2, factor)
	}

	// After endGradient: solid c2 color
	for i := endGradient; i < 256; i++ {
		agg2d.lineGradient[i] = c2
	}

	// Calculate gradient angle and setup transformation matrix
	angle := math.Atan2(y2-y1, x2-x1)

	// Reset and setup the gradient transformation matrix
	agg2d.lineGradientMatrix.Reset()
	agg2d.lineGradientMatrix.Rotate(angle)
	agg2d.lineGradientMatrix.Translate(x1, y1)
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
	// Calculate gradient profile bounds
	startGradient := 128 - int(profile*127.0)
	endGradient := 128 + int(profile*127.0)

	// Ensure we have at least one interpolation step
	if endGradient <= startGradient {
		endGradient = startGradient + 1
	}

	// Calculate interpolation step
	k := 1.0 / float64(endGradient-startGradient)

	// Fill the gradient array
	// Before startGradient: solid c1 color
	for i := 0; i < startGradient; i++ {
		agg2d.fillGradient[i] = c1
	}

	// Between startGradient and endGradient: interpolated colors
	for i := startGradient; i < endGradient; i++ {
		factor := float64(i-startGradient) * k
		agg2d.fillGradient[i] = c1.Gradient(c2, factor)
	}

	// After endGradient: solid c2 color
	for i := endGradient; i < 256; i++ {
		agg2d.fillGradient[i] = c2
	}

	// Setup radial gradient transformation
	// Convert to screen coordinates if needed
	screenRadius := agg2d.worldToScreen(r)
	screenX, screenY := x, y
	agg2d.worldToScreenPoint(&screenX, &screenY)

	// Setup transformation matrix for radial gradient
	agg2d.fillGradientMatrix.Reset()
	agg2d.fillGradientMatrix.Translate(screenX, screenY)
	agg2d.fillGradientMatrix.Invert()

	// Set gradient parameters
	agg2d.fillGradientD1 = 0.0
	agg2d.fillGradientD2 = screenRadius
	agg2d.fillGradientFlag = Radial

	// Set fill color to a placeholder (gradient takes precedence)
	agg2d.fillColor = NewColor(0, 0, 0, 255)
}

// LineRadialGradient sets up a radial gradient for line/stroke operations.
// Parameters are identical to FillRadialGradient but affect line rendering.
// This matches the C++ Agg2D::lineRadialGradient method.
func (agg2d *Agg2D) LineRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	// Calculate gradient profile bounds
	startGradient := 128 - int(profile*127.0)
	endGradient := 128 + int(profile*127.0)

	// Ensure we have at least one interpolation step
	if endGradient <= startGradient {
		endGradient = startGradient + 1
	}

	// Calculate interpolation step
	k := 1.0 / float64(endGradient-startGradient)

	// Fill the gradient array
	// Before startGradient: solid c1 color
	for i := 0; i < startGradient; i++ {
		agg2d.lineGradient[i] = c1
	}

	// Between startGradient and endGradient: interpolated colors
	for i := startGradient; i < endGradient; i++ {
		factor := float64(i-startGradient) * k
		agg2d.lineGradient[i] = c1.Gradient(c2, factor)
	}

	// After endGradient: solid c2 color
	for i := endGradient; i < 256; i++ {
		agg2d.lineGradient[i] = c2
	}

	// Setup radial gradient transformation
	// Convert to screen coordinates if needed
	screenRadius := agg2d.worldToScreen(r)
	screenX, screenY := x, y
	agg2d.worldToScreenPoint(&screenX, &screenY)

	// Setup transformation matrix for radial gradient
	agg2d.lineGradientMatrix.Reset()
	agg2d.lineGradientMatrix.Translate(screenX, screenY)
	agg2d.lineGradientMatrix.Invert()

	// Set gradient parameters
	agg2d.lineGradientD1 = 0.0
	agg2d.lineGradientD2 = screenRadius
	agg2d.lineGradientFlag = Radial

	// Set line color to a placeholder (gradient takes precedence)
	agg2d.lineColor = NewColor(0, 0, 0, 255)
}

// FillRadialGradientMultiStop sets up a radial gradient with three colors.
// This creates a gradient from c1 (center) to c2 (middle) to c3 (edge).
// The transition points are fixed at 50% intervals.
// This matches the C++ Agg2D::fillRadialGradient(x, y, r, c1, c2, c3) method.
func (agg2d *Agg2D) FillRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	// Build gradient with three-color stops
	// First half: c1 to c2 (indices 0-127)
	for i := 0; i < 128; i++ {
		factor := float64(i) / 127.0
		agg2d.fillGradient[i] = c1.Gradient(c2, factor)
	}

	// Second half: c2 to c3 (indices 128-255)
	for i := 128; i < 256; i++ {
		factor := float64(i-128) / 127.0
		agg2d.fillGradient[i] = c2.Gradient(c3, factor)
	}

	// Setup radial gradient transformation
	// Convert to screen coordinates if needed
	screenRadius := agg2d.worldToScreen(r)
	screenX, screenY := x, y
	agg2d.worldToScreenPoint(&screenX, &screenY)

	// Setup transformation matrix for radial gradient
	agg2d.fillGradientMatrix.Reset()
	agg2d.fillGradientMatrix.Translate(screenX, screenY)
	agg2d.fillGradientMatrix.Invert()

	// Set gradient parameters
	agg2d.fillGradientD1 = 0.0
	agg2d.fillGradientD2 = screenRadius
	agg2d.fillGradientFlag = Radial

	// Set fill color to a placeholder (gradient takes precedence)
	agg2d.fillColor = NewColor(0, 0, 0, 255)
}

// LineRadialGradientMultiStop sets up a radial gradient with three colors for line operations.
// This matches the C++ Agg2D::lineRadialGradient(x, y, r, c1, c2, c3) method.
func (agg2d *Agg2D) LineRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	// Build gradient with three-color stops
	// First half: c1 to c2 (indices 0-127)
	for i := 0; i < 128; i++ {
		factor := float64(i) / 127.0
		agg2d.lineGradient[i] = c1.Gradient(c2, factor)
	}

	// Second half: c2 to c3 (indices 128-255)
	for i := 128; i < 256; i++ {
		factor := float64(i-128) / 127.0
		agg2d.lineGradient[i] = c2.Gradient(c3, factor)
	}

	// Setup radial gradient transformation
	// Convert to screen coordinates if needed
	screenRadius := agg2d.worldToScreen(r)
	screenX, screenY := x, y
	agg2d.worldToScreenPoint(&screenX, &screenY)

	// Setup transformation matrix for radial gradient
	agg2d.lineGradientMatrix.Reset()
	agg2d.lineGradientMatrix.Translate(screenX, screenY)
	agg2d.lineGradientMatrix.Invert()

	// Set gradient parameters
	agg2d.lineGradientD1 = 0.0
	agg2d.lineGradientD2 = screenRadius
	agg2d.lineGradientFlag = Radial

	// Set line color to a placeholder (gradient takes precedence)
	agg2d.lineColor = NewColor(0, 0, 0, 255)
}

// FillRadialGradientPos sets up the position and radius for fill radial gradient
// without changing the colors. This matches the C++ Agg2D::fillRadialGradient(x, y, r) method.
func (agg2d *Agg2D) FillRadialGradientPos(x, y, r float64) {
	// Setup radial gradient transformation
	// Convert to screen coordinates if needed
	screenRadius := agg2d.worldToScreen(r)
	screenX, screenY := x, y
	agg2d.worldToScreenPoint(&screenX, &screenY)

	// Setup transformation matrix for radial gradient
	agg2d.fillGradientMatrix.Reset()
	agg2d.fillGradientMatrix.Translate(screenX, screenY)
	agg2d.fillGradientMatrix.Invert()

	// Set gradient parameters
	agg2d.fillGradientD1 = 0.0
	agg2d.fillGradientD2 = screenRadius
}

// LineRadialGradientPos sets up the position and radius for line radial gradient
// without changing the colors. This matches the C++ Agg2D::lineRadialGradient(x, y, r) method.
func (agg2d *Agg2D) LineRadialGradientPos(x, y, r float64) {
	// Setup radial gradient transformation
	// Convert to screen coordinates if needed
	screenRadius := agg2d.worldToScreen(r)
	screenX, screenY := x, y
	agg2d.worldToScreenPoint(&screenX, &screenY)

	// Setup transformation matrix for radial gradient
	agg2d.lineGradientMatrix.Reset()
	agg2d.lineGradientMatrix.Translate(screenX, screenY)
	agg2d.lineGradientMatrix.Invert()

	// Set gradient parameters
	agg2d.lineGradientD1 = 0.0
	agg2d.lineGradientD2 = screenRadius
}

// Helper methods for coordinate transformation
// These will need to be implemented based on the existing AGG2D transformation system

// worldToScreen converts a world coordinate distance to screen coordinates
func (agg2d *Agg2D) worldToScreen(distance float64) float64 {
	// For now, assume 1:1 mapping - this should be implemented based on
	// the current world-to-screen transformation matrix
	return distance
}

// worldToScreenPoint converts world coordinates to screen coordinates
func (agg2d *Agg2D) worldToScreenPoint(x, y *float64) {
	// For now, assume 1:1 mapping - this should be implemented based on
	// the current world-to-screen transformation matrix
	// The actual implementation would apply the current transformation matrix
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
