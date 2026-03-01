// Package agg2d utilities for AGG2D high-level interface.
// This file contains utility methods that match the C++ AGG2D interface.
package agg2d

import (
	"math"

	"agg_go/internal/color"
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
	if x == nil || y == nil {
		return
	}

	agg2d.WorldToScreen(x, y)
	*x = math.Floor(*x) + 0.5
	*y = math.Floor(*y) + 0.5
	agg2d.ScreenToWorld(x, y)
}

// InBox checks if a world coordinate point is inside the current clipping box.
// This matches the C++ Agg2D::inBox method.
func (agg2d *Agg2D) InBox(worldX, worldY float64) bool {
	agg2d.WorldToScreen(&worldX, &worldY)
	if agg2d.renBase != nil {
		return agg2d.renBase.rendererBase().InBox(int(worldX), int(worldY))
	}

	return int(worldX) >= int(agg2d.clipBox.X1) && int(worldX) <= int(agg2d.clipBox.X2) &&
		int(worldY) >= int(agg2d.clipBox.Y1) && int(worldY) <= int(agg2d.clipBox.Y2)
}

// WorldToScreenScalar converts a world scalar value to screen coordinates.
// This matches the C++ Agg2D::worldToScreen(double scalar) method.
func (agg2d *Agg2D) WorldToScreenScalar(scalar float64) float64 {
	x1, y1 := 0.0, 0.0
	x2, y2 := scalar, scalar
	agg2d.WorldToScreen(&x1, &y1)
	agg2d.WorldToScreen(&x2, &y2)
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx+dy*dy) / math.Sqrt(2.0)
}

// ScreenToWorldScalar converts a screen scalar value to world coordinates.
// This matches the C++ Agg2D::screenToWorld(double scalar) method.
func (agg2d *Agg2D) ScreenToWorldScalar(scalar float64) float64 {
	x1, y1 := 0.0, 0.0
	x2, y2 := scalar, scalar
	agg2d.ScreenToWorld(&x1, &y1)
	agg2d.ScreenToWorld(&x2, &y2)
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx+dy*dy) / math.Sqrt(2.0)
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
	if agg2d.renBase == nil {
		return
	}

	clearColor := color.RGBA8[color.Linear]{R: r, G: g, B: b, A: a}
	agg2d.renBase.rendererBase().CopyBar(0, 0, agg2d.renBase.Width(), agg2d.renBase.Height(), clearColor)
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
