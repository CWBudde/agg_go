// Package primitives provides line anti-aliasing basic structures and functions.
// This implements a port of AGG's agg_line_aa_basics.h functionality.
package primitives

import (
	"math"

	"agg_go/internal/basics"
)

// Line subpixel scale constants
const (
	LineSubpixelShift = 8                             // line_subpixel_shift
	LineSubpixelScale = 1 << LineSubpixelShift        // line_subpixel_scale = 256
	LineSubpixelMask  = LineSubpixelScale - 1         // line_subpixel_mask = 255
	LineMaxCoord      = (1 << 28) - 1                 // line_max_coord
	LineMaxLength     = 1 << (LineSubpixelShift + 10) // line_max_length
)

// Medium resolution subpixel scale constants
const (
	LineMRSubpixelShift = 4                        // line_mr_subpixel_shift
	LineMRSubpixelScale = 1 << LineMRSubpixelShift // line_mr_subpixel_scale = 16
	LineMRSubpixelMask  = LineMRSubpixelScale - 1  // line_mr_subpixel_mask = 15
)

// LineMR reduces resolution for medium-resolution operations.
func LineMR(x int) int {
	return x >> (LineSubpixelShift - LineMRSubpixelShift)
}

// LineHR increases resolution for high-resolution operations.
func LineHR(x int) int {
	return x << (LineSubpixelShift - LineMRSubpixelShift)
}

// LineDblHR doubles the resolution.
func LineDblHR(x int) int {
	return x << LineSubpixelShift
}

// LineDBLHR is an alias for LineDblHR for compatibility.
func LineDBLHR(x int) int {
	return LineDblHR(x)
}

// LineCoordSatInstance provides a global instance for coordinate saturation.
var LineCoordSatInstance = LineCoordSat{}

// LineCoordSatConv function interface
func LineCoordSatConv(x float64) int {
	return LineCoordSatInstance.Conv(x)
}

// LineCoord provides coordinate conversion for standard precision.
type LineCoord struct{}

// Conv converts a floating-point coordinate to subpixel integer coordinates.
func (LineCoord) Conv(x float64) int {
	return basics.IRound(x * LineSubpixelScale)
}

// LineCoordSat provides coordinate conversion with saturation.
type LineCoordSat struct{}

// Conv converts a floating-point coordinate to subpixel integer coordinates with saturation.
func (LineCoordSat) Conv(x float64) int {
	return basics.SaturationIRound(x*LineSubpixelScale, LineMaxCoord)
}

// LineParameters represents line segment parameters for anti-aliased rendering.
// This is equivalent to AGG's line_parameters struct.
type LineParameters struct {
	X1, Y1, X2, Y2 int  // Start and end coordinates
	DX, DY         int  // Absolute differences
	SX, SY         int  // Step directions (-1 or +1)
	Vertical       bool // true if dy >= dx
	Inc            int  // Increment (vertical ? sy : sx)
	Len            int  // Length of the line
	Octant         int  // Octant number (0-7)
}

// Quadrant lookup tables
var (
	// orthogonalQuadrant maps octant to orthogonal quadrant [0-3]
	orthogonalQuadrant = [8]uint8{0, 0, 1, 1, 3, 3, 2, 2}

	// diagonalQuadrant maps octant to diagonal quadrant [0-3]
	diagonalQuadrant = [8]uint8{0, 1, 2, 1, 0, 3, 2, 3}
)

// NewLineParameters creates a new LineParameters struct with calculated values.
func NewLineParameters(x1, y1, x2, y2, len int) LineParameters {
	dx := basics.Abs(x2 - x1)
	dy := basics.Abs(y2 - y1)
	sx := 1
	if x2 <= x1 {
		sx = -1
	}
	sy := 1
	if y2 <= y1 {
		sy = -1
	}
	vertical := dy >= dx
	inc := sx
	if vertical {
		inc = sy
	}

	// Calculate octant based on step directions and vertical flag
	// bit 0 = vertical flag, bit 1 = sx < 0, bit 2 = sy < 0
	octant := 0
	if vertical {
		octant |= 1
	}
	if sx < 0 {
		octant |= 2
	}
	if sy < 0 {
		octant |= 4
	}

	return LineParameters{
		X1: x1, Y1: y1, X2: x2, Y2: y2,
		DX: dx, DY: dy,
		SX: sx, SY: sy,
		Vertical: vertical,
		Inc:      inc,
		Len:      len,
		Octant:   octant,
	}
}

// OrthogonalQuadrant returns the orthogonal quadrant for this line.
func (lp *LineParameters) OrthogonalQuadrant() uint8 {
	return orthogonalQuadrant[lp.Octant]
}

// DiagonalQuadrant returns the diagonal quadrant for this line.
func (lp *LineParameters) DiagonalQuadrant() uint8 {
	return diagonalQuadrant[lp.Octant]
}

// SameOrthogonalQuadrant checks if this line is in the same orthogonal quadrant as another.
func (lp *LineParameters) SameOrthogonalQuadrant(other *LineParameters) bool {
	return orthogonalQuadrant[lp.Octant] == orthogonalQuadrant[other.Octant]
}

// SameDiagonalQuadrant checks if this line is in the same diagonal quadrant as another.
func (lp *LineParameters) SameDiagonalQuadrant(other *LineParameters) bool {
	return diagonalQuadrant[lp.Octant] == diagonalQuadrant[other.Octant]
}

// Divide splits this line into two halves.
func (lp *LineParameters) Divide() (LineParameters, LineParameters) {
	xmid := (lp.X1 + lp.X2) >> 1
	ymid := (lp.Y1 + lp.Y2) >> 1
	len2 := lp.Len >> 1

	lp1 := *lp
	lp1.X2 = xmid
	lp1.Y2 = ymid
	lp1.Len = len2
	lp1.DX = basics.Abs(lp1.X2 - lp1.X1)
	lp1.DY = basics.Abs(lp1.Y2 - lp1.Y1)

	lp2 := *lp
	lp2.X1 = xmid
	lp2.Y1 = ymid
	lp2.Len = len2
	lp2.DX = basics.Abs(lp2.X2 - lp2.X1)
	lp2.DY = basics.Abs(lp2.Y2 - lp2.Y1)

	return lp1, lp2
}

// Bisectrix calculates the bisector point between two line segments.
// This is equivalent to AGG's bisectrix function.
func Bisectrix(l1, l2 *LineParameters) (int, int) {
	k := float64(l2.Len) / float64(l1.Len)
	tx := float64(l2.X2) - float64(l2.X1-l1.X1)*k
	ty := float64(l2.Y2) - float64(l2.Y1-l1.Y1)*k

	// All bisectrices must be on the right of the line
	// If the next point is on the left (l1 => l2.2)
	// then the bisectrix should be rotated by 180 degrees.
	if float64(l2.X2-l2.X1)*float64(l2.Y1-l1.Y1) <
		float64(l2.Y2-l2.Y1)*float64(l2.X1-l1.X1)+100.0 {
		tx -= (tx - float64(l2.X1)) * 2.0
		ty -= (ty - float64(l2.Y1)) * 2.0
	}

	// Check if the bisectrix is too short
	dx := tx - float64(l2.X1)
	dy := ty - float64(l2.Y1)
	if int(math.Sqrt(dx*dx+dy*dy)) < LineSubpixelScale {
		x := (l2.X1 + l2.X1 + (l2.Y1 - l1.Y1) + (l2.Y2 - l2.Y1)) >> 1
		y := (l2.Y1 + l2.Y1 - (l2.X1 - l1.X1) - (l2.X2 - l2.X1)) >> 1
		return x, y
	}

	return basics.IRound(tx), basics.IRound(ty)
}

// FixDegenerateBisectrixStart fixes degenerate bisectrix at line start.
func FixDegenerateBisectrixStart(lp *LineParameters, x, y *int) {
	d := basics.IRound((float64(*x-lp.X2)*float64(lp.Y2-lp.Y1) -
		float64(*y-lp.Y2)*float64(lp.X2-lp.X1)) / float64(lp.Len))
	if d < LineSubpixelScale/2 {
		*x = lp.X1 + (lp.Y2 - lp.Y1)
		*y = lp.Y1 - (lp.X2 - lp.X1)
	}
}

// FixDegenerateBisectrixEnd fixes degenerate bisectrix at line end.
func FixDegenerateBisectrixEnd(lp *LineParameters, x, y *int) {
	d := basics.IRound((float64(*x-lp.X2)*float64(lp.Y2-lp.Y1) -
		float64(*y-lp.Y2)*float64(lp.X2-lp.X1)) / float64(lp.Len))
	if d < LineSubpixelScale/2 {
		*x = lp.X2 + (lp.Y2 - lp.Y1)
		*y = lp.Y2 - (lp.X2 - lp.X1)
	}
}

// CmpDistStart compares distance from start (used for caps).
func CmpDistStart(d int) bool {
	return d > 0
}

// CmpDistEnd compares distance from end (used for caps).
func CmpDistEnd(d int) bool {
	return d <= 0
}
