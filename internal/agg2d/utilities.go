// Package agg2d utilities for AGG2D high-level interface.
// This file contains utility methods that match the C++ AGG2D interface.
package agg2d

import (
	"math"

	"agg_go/internal/color"
	"agg_go/internal/renderer"
)

// Note: Mathematical constants Pi, Deg2Rad, Rad2Deg functions are defined in constants.go

// Deg2Rad converts degrees to radians.
// This matches the C++ Agg2D::deg2Rad static method.
func Deg2Rad(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

// Rad2Deg converts radians to degrees.
// This matches the C++ Agg2D::rad2Deg static method.
func Rad2Deg(radians float64) float64 {
	return radians * 180.0 / math.Pi
}

// AlignPoint aligns a point to pixel boundaries for crisp rendering.
// This matches the C++ Agg2D::alignPoint method.
func (agg2d *Agg2D) AlignPoint(x, y *float64) {
	if x != nil {
		*x = math.Floor(*x) + 0.5
	}
	if y != nil {
		*y = math.Floor(*y) + 0.5
	}
}

// InBox checks if a world coordinate point is inside the current clipping box.
// This matches the C++ Agg2D::inBox method.
func (agg2d *Agg2D) InBox(worldX, worldY float64) bool {
	// Transform world coordinates to screen coordinates
	screenX, screenY := worldX, worldY
	agg2d.transform.Transform(&screenX, &screenY)

	// Check if the screen coordinates are within the clip box
	return screenX >= agg2d.clipBox.X1 && screenX <= agg2d.clipBox.X2 &&
		screenY >= agg2d.clipBox.Y1 && screenY <= agg2d.clipBox.Y2
}

// WorldToScreenScalar converts a world scalar value to screen coordinates.
// This matches the C++ Agg2D::worldToScreen(double scalar) method.
func (agg2d *Agg2D) WorldToScreenScalar(scalar float64) float64 {
	// Calculate the overall scaling factor from the transformation matrix
	scaleX := math.Sqrt(agg2d.transform.SX*agg2d.transform.SX + agg2d.transform.SHY*agg2d.transform.SHY)
	scaleY := math.Sqrt(agg2d.transform.SHX*agg2d.transform.SHX + agg2d.transform.SY*agg2d.transform.SY)
	scale := (scaleX + scaleY) / 2.0

	return scalar * scale
}

// ScreenToWorldScalar converts a screen scalar value to world coordinates.
// This matches the C++ Agg2D::screenToWorld(double scalar) method.
func (agg2d *Agg2D) ScreenToWorldScalar(scalar float64) float64 {
	// Calculate the overall scaling factor from the transformation matrix
	scaleX := math.Sqrt(agg2d.transform.SX*agg2d.transform.SX + agg2d.transform.SHY*agg2d.transform.SHY)
	scaleY := math.Sqrt(agg2d.transform.SHX*agg2d.transform.SHX + agg2d.transform.SY*agg2d.transform.SY)
	scale := (scaleX + scaleY) / 2.0

	if scale > 0.0 {
		return scalar / scale
	}
	return scalar
}

// NoFill disables fill rendering.
// This matches the C++ Agg2D::noFill method.
func (agg2d *Agg2D) NoFill() {
	agg2d.fillColor = Color{0, 0, 0, 0} // Transparent fill
	agg2d.fillGradientFlag = Solid
}

// NoLine disables line/stroke rendering.
// This matches the C++ Agg2D::noLine method.
func (agg2d *Agg2D) NoLine() {
	agg2d.lineColor = Color{0, 0, 0, 0} // Transparent line
	agg2d.lineGradientFlag = Solid
}

// ClearClipBox clears the current clipping box with the specified color.
// This matches the C++ Agg2D::clearClipBox(Color c) method.
func (agg2d *Agg2D) ClearClipBox(c Color) {
	agg2d.ClearClipBoxRGBA(c[0], c[1], c[2], c[3])
}

// ClearClipBoxRGBA clears the current clipping box with the specified RGBA values.
// This matches the C++ Agg2D::clearClipBox(unsigned r, g, b, a) method.
func (agg2d *Agg2D) ClearClipBoxRGBA(r, g, b, a uint8) {
	if agg2d.renBase == nil || agg2d.pixfmt == nil {
		return
	}

	// Convert the clip box coordinates to integers
	x1 := int(math.Floor(agg2d.clipBox.X1))
	y1 := int(math.Floor(agg2d.clipBox.Y1))
	x2 := int(math.Ceil(agg2d.clipBox.X2))
	y2 := int(math.Ceil(agg2d.clipBox.Y2))

	// Ensure coordinates are within buffer bounds
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > agg2d.renBase.Width() {
		x2 = agg2d.renBase.Width() - 1
	}
	if y2 > agg2d.renBase.Height() {
		y2 = agg2d.renBase.Height() - 1
	}

	// Skip if the clip box is empty or invalid
	if x1 > x2 || y1 > y2 {
		return
	}

	// Create a color for clearing
	clearColor := color.RGBA8[color.Linear]{R: r, G: g, B: b, A: a}

	// Create a temporary renderer base to handle the clearing with clipping
	rendererBase := renderer.NewRendererBaseWithPixfmt(agg2d.pixfmt)

	// Set the clipping box to match our desired clear area
	rendererBase.ClipBox(x1, y1, x2, y2)

	// Use CopyBar to clear the clipped area
	rendererBase.CopyBar(x1, y1, x2, y2, clearColor)
}

// Clamp clamps a value between min and max
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampByte clamps a float64 value to uint8 range [0, 255]
func ClampByte(value float64) uint8 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}

// LinearInterpolate performs linear interpolation between two values
func LinearInterpolate(a, b, t float64) float64 {
	return a + t*(b-a)
}

// ColorInterpolate performs linear interpolation between two colors
func ColorInterpolate(c1, c2 Color, t float64) Color {
	// Clamp interpolation factor
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	return Color{
		ClampByte(LinearInterpolate(float64(c1[0]), float64(c2[0]), t)),
		ClampByte(LinearInterpolate(float64(c1[1]), float64(c2[1]), t)),
		ClampByte(LinearInterpolate(float64(c1[2]), float64(c2[2]), t)),
		ClampByte(LinearInterpolate(float64(c1[3]), float64(c2[3]), t)),
	}
}

// Distance calculates the Euclidean distance between two points
func Distance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// Angle calculates the angle in radians from point (x1, y1) to point (x2, y2)
func Angle(x1, y1, x2, y2 float64) float64 {
	return math.Atan2(y2-y1, x2-x1)
}

// NormalizeAngle normalizes an angle to be within [0, 2π) range
func NormalizeAngle(angle float64) float64 {
	for angle < 0 {
		angle += 2 * math.Pi
	}
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}
	return angle
}

// AngleDifference calculates the shortest angular difference between two angles
func AngleDifference(angle1, angle2 float64) float64 {
	diff := angle2 - angle1
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}
	return diff
}

// IsZero checks if a floating-point value is effectively zero
func IsZero(value float64) bool {
	const epsilon = 1e-10
	return math.Abs(value) < epsilon
}

// IsEqual checks if two floating-point values are effectively equal
func IsEqual(a, b float64) bool {
	const epsilon = 1e-10
	return math.Abs(a-b) < epsilon
}

// Sign returns the sign of a value: -1, 0, or 1
func Sign(value float64) int {
	if value > 0 {
		return 1
	}
	if value < 0 {
		return -1
	}
	return 0
}
