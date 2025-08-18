// Package bezierarc provides Bezier arc generation and conversion functionality.
// This is a direct port of AGG's agg_bezier_arc.h/cpp which converts circular
// and elliptical arcs into cubic Bezier curves for high-quality rendering.
//
// The package provides two main types:
// - BezierArc: Basic arc approximation using cubic Bezier curves
// - BezierArcSVG: SVG-style arc generation from point to point
package bezierarc

import (
	"math"

	"agg_go/internal/basics"
)

// Constants for bezier arc calculations
const (
	// BezierArcAngleEpsilon prevents degenerate curves that converge to a single point.
	// The value isn't very critical - slight exceeding of pi/2 sweep angles is acceptable.
	BezierArcAngleEpsilon = 0.01
)

// ArcToBezier converts an arc to a cubic Bezier curve.
// This function takes arc parameters (center, radii, angles) and outputs
// 4 control points that define a cubic Bezier curve approximating the arc.
//
// Parameters:
//   - cx, cy: arc center coordinates
//   - rx, ry: arc radii in X and Y directions
//   - startAngle: starting angle in radians
//   - sweepAngle: sweep angle in radians (can be negative for clockwise)
//
// Returns a slice of 8 values: [x0, y0, x1, y1, x2, y2, x3, y3]
// representing the 4 control points of the cubic Bezier curve.
func ArcToBezier(cx, cy, rx, ry, startAngle, sweepAngle float64) []float64 {
	curve := make([]float64, 8)

	// Calculate the halfway point parameters
	x0 := math.Cos(sweepAngle / 2.0)
	y0 := math.Sin(sweepAngle / 2.0)

	// Calculate control point offsets using the magic number 4/3
	// This is the standard formula for converting circular arcs to Bezier curves
	tx := (1.0 - x0) * 4.0 / 3.0
	ty := y0 - tx*x0/y0

	// Define the 4 control points in local coordinates
	px := [4]float64{x0, x0 + tx, x0 + tx, x0}
	py := [4]float64{-y0, -ty, ty, y0}

	// Calculate rotation parameters for the final transformation
	sn := math.Sin(startAngle + sweepAngle/2.0)
	cs := math.Cos(startAngle + sweepAngle/2.0)

	// Transform control points to global coordinates
	for i := 0; i < 4; i++ {
		curve[i*2] = cx + rx*(px[i]*cs-py[i]*sn)
		curve[i*2+1] = cy + ry*(px[i]*sn+py[i]*cs)
	}

	return curve
}

// BezierArc generates vertices for an elliptical arc using cubic Bezier approximation.
// It can produce at most 4 consecutive cubic Bezier curves (up to 13 vertices).
// This is a direct port of AGG's bezier_arc class.
type BezierArc struct {
	vertex      uint               // Current vertex index during iteration
	numVertices uint               // Total number of vertices (always even)
	vertices    [26]float64        // Vertex array - pairs of x,y coordinates
	cmd         basics.PathCommand // Current path command
}

// NewBezierArc creates a new uninitialized bezier arc.
func NewBezierArc() *BezierArc {
	return &BezierArc{
		vertex: 26, // Start beyond array to indicate uninitialized
		cmd:    basics.PathCmdLineTo,
	}
}

// NewBezierArcWithParams creates and initializes a new bezier arc.
func NewBezierArcWithParams(x, y, rx, ry, startAngle, sweepAngle float64) *BezierArc {
	arc := NewBezierArc()
	arc.Init(x, y, rx, ry, startAngle, sweepAngle)
	return arc
}

// Init initializes the bezier arc with the specified parameters.
// This generates the cubic Bezier curve approximation of the arc.
func (ba *BezierArc) Init(x, y, rx, ry, startAngle, sweepAngle float64) {
	// Normalize start angle to [0, 2π)
	startAngle = math.Mod(startAngle, 2.0*basics.Pi)

	// Clamp sweep angle to [-2π, 2π]
	if sweepAngle >= 2.0*basics.Pi {
		sweepAngle = 2.0 * basics.Pi
	}
	if sweepAngle <= -2.0*basics.Pi {
		sweepAngle = -2.0 * basics.Pi
	}

	// Handle degenerate case where sweep angle is essentially zero
	if math.Abs(sweepAngle) < 1e-10 {
		ba.numVertices = 4
		ba.cmd = basics.PathCmdLineTo
		ba.vertices[0] = x + rx*math.Cos(startAngle)
		ba.vertices[1] = y + ry*math.Sin(startAngle)
		ba.vertices[2] = x + rx*math.Cos(startAngle+sweepAngle)
		ba.vertices[3] = y + ry*math.Sin(startAngle+sweepAngle)
		return
	}

	// Generate cubic Bezier curves to approximate the arc
	// We break the arc into segments of at most π/2 radians each
	totalSweep := 0.0
	ba.numVertices = 2
	ba.cmd = basics.PathCmdCurve4

	for ba.numVertices < 26 {
		prevSweep := totalSweep
		var localSweep float64
		done := false

		if sweepAngle < 0.0 {
			// Clockwise direction
			localSweep = -basics.Pi * 0.5
			totalSweep -= basics.Pi * 0.5
			if totalSweep <= sweepAngle+BezierArcAngleEpsilon {
				localSweep = sweepAngle - prevSweep
				done = true
			}
		} else {
			// Counter-clockwise direction
			localSweep = basics.Pi * 0.5
			totalSweep += basics.Pi * 0.5
			if totalSweep >= sweepAngle-BezierArcAngleEpsilon {
				localSweep = sweepAngle - prevSweep
				done = true
			}
		}

		// Generate the Bezier curve for this segment
		curve := ArcToBezier(x, y, rx, ry, startAngle, localSweep)

		// Copy the curve points into our vertex array
		// Skip the first point if this isn't the first segment (to avoid duplication)
		startIdx := int(ba.numVertices - 2)
		for i := 0; i < 8; i++ {
			ba.vertices[startIdx+i] = curve[i]
		}

		ba.numVertices += 6
		startAngle += localSweep

		if done {
			break
		}
	}
}

// Rewind resets the arc to the beginning for vertex iteration.
func (ba *BezierArc) Rewind(pathId uint32) {
	ba.vertex = 0
}

// Vertex returns the next vertex in the arc.
// Returns PathCmdStop when all vertices have been consumed.
func (ba *BezierArc) Vertex(x, y *float64) basics.PathCommand {
	if ba.vertex >= ba.numVertices {
		return basics.PathCmdStop
	}

	*x = ba.vertices[ba.vertex]
	*y = ba.vertices[ba.vertex+1]
	ba.vertex += 2

	if ba.vertex == 2 {
		return basics.PathCmdMoveTo
	}
	return ba.cmd
}

// NumVertices returns the number of vertices (actually returns doubled count).
// For 1 vertex it returns 2, for 2 vertices it returns 4, etc.
func (ba *BezierArc) NumVertices() uint {
	return ba.numVertices
}

// Vertices returns a pointer to the internal vertex array.
func (ba *BezierArc) Vertices() []float64 {
	return ba.vertices[:ba.numVertices]
}

// BezierArcSVG computes SVG-style Bezier arcs.
// This handles the complex case of defining an arc from one point to another
// with specified radii, rotation, and arc flags, as used in SVG path commands.
type BezierArcSVG struct {
	arc     BezierArc // The underlying bezier arc
	radiiOk bool      // Whether the radii are valid for the given endpoints
}

// NewBezierArcSVG creates a new uninitialized SVG bezier arc.
func NewBezierArcSVG() *BezierArcSVG {
	return &BezierArcSVG{
		arc:     *NewBezierArc(),
		radiiOk: false,
	}
}

// NewBezierArcSVGWithParams creates and initializes a new SVG bezier arc.
func NewBezierArcSVGWithParams(x1, y1, rx, ry, angle float64, largeArcFlag, sweepFlag bool, x2, y2 float64) *BezierArcSVG {
	svg := NewBezierArcSVG()
	svg.Init(x1, y1, rx, ry, angle, largeArcFlag, sweepFlag, x2, y2)
	return svg
}

// Init initializes the SVG arc with the specified parameters.
// This is the complex algorithm that converts SVG arc parameters into
// the center-based parameters needed for the underlying BezierArc.
func (bas *BezierArcSVG) Init(x0, y0, rx, ry, angle float64, largeArcFlag, sweepFlag bool, x2, y2 float64) {
	bas.radiiOk = true

	// Ensure radii are positive
	if rx < 0.0 {
		rx = -rx
	}
	if ry < 0.0 {
		ry = -ry
	}

	// Calculate the middle point between current and final points
	dx2 := (x0 - x2) / 2.0
	dy2 := (y0 - y2) / 2.0

	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	// Calculate (x1, y1) - the midpoint in rotated coordinates
	x1 := cosA*dx2 + sinA*dy2
	y1 := -sinA*dx2 + cosA*dy2

	// Ensure radii are large enough
	prx := rx * rx
	pry := ry * ry
	px1 := x1 * x1
	py1 := y1 * y1

	// Check that radii are large enough - if not, scale them up
	radiiCheck := px1/prx + py1/pry
	if radiiCheck > 1.0 {
		rx = math.Sqrt(radiiCheck) * rx
		ry = math.Sqrt(radiiCheck) * ry
		prx = rx * rx
		pry = ry * ry
		if radiiCheck > 10.0 {
			bas.radiiOk = false
		}
	}

	// Calculate (cx1, cy1) - the center in rotated coordinates
	var sign float64
	if largeArcFlag == sweepFlag {
		sign = -1.0
	} else {
		sign = 1.0
	}

	sq := (prx*pry - prx*py1 - pry*px1) / (prx*py1 + pry*px1)
	if sq < 0 {
		sq = 0
	}
	coef := sign * math.Sqrt(sq)
	cx1 := coef * ((rx * y1) / ry)
	cy1 := coef * -((ry * x1) / rx)

	// Calculate (cx, cy) - the center in original coordinates
	sx2 := (x0 + x2) / 2.0
	sy2 := (y0 + y2) / 2.0
	cx := sx2 + (cosA*cx1 - sinA*cy1)
	cy := sy2 + (sinA*cx1 + cosA*cy1)

	// Calculate the start_angle and sweep_angle
	ux := (x1 - cx1) / rx
	uy := (y1 - cy1) / ry
	vx := (-x1 - cx1) / rx
	vy := (-y1 - cy1) / ry

	// Calculate the start angle
	n := math.Sqrt(ux*ux + uy*uy)
	p := ux // (1 * ux) + (0 * uy)
	if uy < 0 {
		sign = -1.0
	} else {
		sign = 1.0
	}
	v := p / n
	if v < -1.0 {
		v = -1.0
	}
	if v > 1.0 {
		v = 1.0
	}
	startAngle := sign * math.Acos(v)

	// Calculate the sweep angle
	n = math.Sqrt((ux*ux + uy*uy) * (vx*vx + vy*vy))
	p = ux*vx + uy*vy
	if ux*vy-uy*vx < 0 {
		sign = -1.0
	} else {
		sign = 1.0
	}
	v = p / n
	if v < -1.0 {
		v = -1.0
	}
	if v > 1.0 {
		v = 1.0
	}
	sweepAngle := sign * math.Acos(v)

	if !sweepFlag && sweepAngle > 0 {
		sweepAngle -= basics.Pi * 2.0
	} else if sweepFlag && sweepAngle < 0 {
		sweepAngle += basics.Pi * 2.0
	}

	// Build and transform the resulting arc
	bas.arc.Init(0.0, 0.0, rx, ry, startAngle, sweepAngle)

	// Apply the transformation to all vertices except the first and last
	// (which we'll set manually for precision)
	for i := uint(2); i < bas.arc.numVertices-2; i += 2 {
		// Apply rotation then translation
		x := bas.arc.vertices[i]
		y := bas.arc.vertices[i+1]

		// Rotation
		newX := cosA*x - sinA*y
		newY := sinA*x + cosA*y

		// Translation
		bas.arc.vertices[i] = newX + cx
		bas.arc.vertices[i+1] = newY + cy
	}

	// Ensure exact endpoint matching
	bas.arc.vertices[0] = x0
	bas.arc.vertices[1] = y0
	if bas.arc.numVertices > 2 {
		bas.arc.vertices[bas.arc.numVertices-2] = x2
		bas.arc.vertices[bas.arc.numVertices-1] = y2
	}
}

// RadiiOk returns whether the arc could be constructed with the given radii.
// Returns false if the radii had to be scaled up significantly.
func (bas *BezierArcSVG) RadiiOk() bool {
	return bas.radiiOk
}

// Rewind resets the SVG arc to the beginning for vertex iteration.
func (bas *BezierArcSVG) Rewind(pathId uint32) {
	bas.arc.Rewind(pathId)
}

// Vertex returns the next vertex in the SVG arc.
func (bas *BezierArcSVG) Vertex(x, y *float64) basics.PathCommand {
	return bas.arc.Vertex(x, y)
}

// NumVertices returns the number of vertices in the SVG arc.
func (bas *BezierArcSVG) NumVertices() uint {
	return bas.arc.NumVertices()
}

// Vertices returns the vertex array of the SVG arc.
func (bas *BezierArcSVG) Vertices() []float64 {
	return bas.arc.Vertices()
}
