// Package gamma provides gamma correction control widgets for AGG.
// This is a port of AGG's gamma control functionality from agg_gamma_spline.h/cpp.
package gamma

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/curves"
)

// GammaSpline implements a cubic spline-based gamma curve editor.
// This is a Go port of AGG's gamma_spline class from agg_gamma_spline.h/cpp.
//
// A very simple class for Bi-cubic Spline interpolation used for gamma correction.
// The gamma curve is defined by 4 control values (kx1, ky1, kx2, ky2) that determine
// the shape of the curve. Each value can be in range [0...2]. Value 1.0 means one
// quarter of the bounding rectangle.
//
// The class supports:
// - Gamma correction curve visualization and editing
// - Multi-point spline-based curve definition
// - Real-time gamma preview functionality
// - Integration with color management pipeline
// - Vertex source interface for rendering
type GammaSpline struct {
	// Gamma lookup table for fast correction
	gamma [256]uint8

	// Control points for the spline
	x [4]float64
	y [4]float64

	// B-spline for curve interpolation
	spline *curves.BSpline

	// Rendering bounds (set by Box method)
	x1, y1, x2, y2 float64

	// Current position for vertex generation
	curX float64
}

// NewGammaSpline creates a new gamma spline with default values.
// Default curve is identity (no gamma correction).
func NewGammaSpline() *GammaSpline {
	gs := &GammaSpline{
		spline: curves.NewBSpline(),
		x1:     0,
		y1:     0,
		x2:     10,
		y2:     10,
		curX:   0.0,
	}
	// Initialize with identity gamma (no correction)
	gs.Values(1.0, 1.0, 1.0, 1.0)
	return gs
}

// Values sets the gamma curve control points.
// kx1, ky1: first control point (typically lower-left influence)
// kx2, ky2: second control point (typically upper-right influence)
// Each value should be in range [0.001, 1.999] for stable curves.
func (gs *GammaSpline) Values(kx1, ky1, kx2, ky2 float64) {
	// Clamp values to safe range to prevent numerical instability
	if kx1 < 0.001 {
		kx1 = 0.001
	}
	if kx1 > 1.999 {
		kx1 = 1.999
	}
	if ky1 < 0.001 {
		ky1 = 0.001
	}
	if ky1 > 1.999 {
		ky1 = 1.999
	}
	if kx2 < 0.001 {
		kx2 = 0.001
	}
	if kx2 > 1.999 {
		kx2 = 1.999
	}
	if ky2 < 0.001 {
		ky2 = 0.001
	}
	if ky2 > 1.999 {
		ky2 = 1.999
	}

	// Set up the 4 control points for the cubic spline
	// Start point: (0, 0)
	gs.x[0] = 0.0
	gs.y[0] = 0.0

	// First control point: quarter of the input values
	gs.x[1] = kx1 * 0.25
	gs.y[1] = ky1 * 0.25

	// Second control point: complement of quarter values
	gs.x[2] = 1.0 - kx2*0.25
	gs.y[2] = 1.0 - ky2*0.25

	// End point: (1, 1)
	gs.x[3] = 1.0
	gs.y[3] = 1.0

	// Initialize the spline with the control points
	gs.spline.InitFromPoints(gs.x[:], gs.y[:])

	// Generate the gamma lookup table
	for i := 0; i < 256; i++ {
		// Convert index to normalized input [0, 1]
		input := float64(i) / 255.0
		// Get spline output and convert to 8-bit value
		output := gs.Y(input) * 255.0
		gs.gamma[i] = uint8(output)
	}
}

// GetValues returns the current gamma curve control points.
// kx1, ky1: first control point
// kx2, ky2: second control point
func (gs *GammaSpline) GetValues() (kx1, ky1, kx2, ky2 float64) {
	kx1 = gs.x[1] * 4.0
	ky1 = gs.y[1] * 4.0
	kx2 = (1.0 - gs.x[2]) * 4.0
	ky2 = (1.0 - gs.y[2]) * 4.0
	return
}

// Y calculates the gamma-corrected value for a given input.
// x: input value in range [0, 1]
// Returns: gamma-corrected output value in range [0, 1]
func (gs *GammaSpline) Y(x float64) float64 {
	// Clamp input to [0, 1] range
	if x < 0.0 {
		x = 0.0
	}
	if x > 1.0 {
		x = 1.0
	}

	// Get the spline value
	val := gs.spline.Get(x)

	// Clamp output to [0, 1] range
	if val < 0.0 {
		val = 0.0
	}
	if val > 1.0 {
		val = 1.0
	}

	return val
}

// Gamma returns the 256-entry gamma lookup table.
// This table can be used for fast gamma correction of 8-bit values.
func (gs *GammaSpline) Gamma() []uint8 {
	return gs.gamma[:]
}

// Box sets the bounding box for curve rendering.
// x1, y1: top-left corner
// x2, y2: bottom-right corner
func (gs *GammaSpline) Box(x1, y1, x2, y2 float64) {
	gs.x1 = x1
	gs.y1 = y1
	gs.x2 = x2
	gs.y2 = y2
}

// Rewind resets the vertex iterator to the beginning of the curve.
// pathID: ignored (for interface compatibility)
func (gs *GammaSpline) Rewind(pathID uint) {
	gs.curX = 0.0
}

// Vertex generates the next vertex along the gamma curve.
// Returns the coordinates and path command for rendering the curve.
func (gs *GammaSpline) Vertex() (x, y float64, cmd basics.PathCommand) {
	if gs.curX == 0.0 {
		// First vertex - move to starting point
		x = gs.x1
		y = gs.y1
		gs.curX += 1.0 / (gs.x2 - gs.x1)
		return x, y, basics.PathCmdMoveTo
	}

	if gs.curX > 1.0 {
		// End of curve
		return 0, 0, basics.PathCmdStop
	}

	// Calculate current position along the curve
	x = gs.x1 + gs.curX*(gs.x2-gs.x1)
	y = gs.y1 + gs.Y(gs.curX)*(gs.y2-gs.y1)

	gs.curX += 1.0 / (gs.x2 - gs.x1)
	return x, y, basics.PathCmdLineTo
}

// NormalizedY is a convenience function that applies gamma correction to a
// normalized input value and returns a normalized output value.
func (gs *GammaSpline) NormalizedY(x float64) float64 {
	return gs.Y(x)
}

// ApplyGamma applies gamma correction to an 8-bit value using the lookup table.
// input: input value [0, 255]
// Returns: gamma-corrected output value [0, 255]
func (gs *GammaSpline) ApplyGamma(input uint8) uint8 {
	return gs.gamma[input]
}

// ApplyGammaFloat applies gamma correction to a floating-point value.
// input: input value [0.0, 1.0]
// Returns: gamma-corrected output value [0.0, 1.0]
func (gs *GammaSpline) ApplyGammaFloat(input float64) float64 {
	return gs.Y(input)
}

// GetCurvePoints returns points along the gamma curve for debugging/inspection.
// numPoints: number of points to sample along the curve
// Returns: slices of X and Y coordinates
func (gs *GammaSpline) GetCurvePoints(numPoints int) ([]float64, []float64) {
	if numPoints <= 0 {
		return nil, nil
	}

	xPoints := make([]float64, numPoints)
	yPoints := make([]float64, numPoints)

	for i := 0; i < numPoints; i++ {
		x := float64(i) / float64(numPoints-1)
		xPoints[i] = x
		yPoints[i] = gs.Y(x)
	}

	return xPoints, yPoints
}

// IsIdentity checks if the current gamma curve is effectively an identity function.
// tolerance: maximum allowed deviation from identity
// Returns: true if the curve is close to identity (no gamma correction)
func (gs *GammaSpline) IsIdentity(tolerance float64) bool {
	// Sample the curve at several points
	const numSamples = 16
	for i := 0; i < numSamples; i++ {
		x := float64(i) / float64(numSamples-1)
		y := gs.Y(x)
		if math.Abs(y-x) > tolerance {
			return false
		}
	}
	return true
}
