package vcgen

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
	"agg_go/internal/curves"
)

// BSplineStatus represents the state of B-spline generation
type BSplineStatus int

const (
	BSplineInitial BSplineStatus = iota
	BSplineReady
	BSplinePolygon
	BSplineEndPoly
	BSplineStop
)

// VCGenBSpline generates vertices for B-spline interpolation
type VCGenBSpline struct {
	srcVertices       *array.PodBVector[basics.Point[float64]]
	splineX           *curves.BSpline
	splineY           *curves.BSpline
	interpolationStep float64
	closed            bool
	status            BSplineStatus
	srcVertex         int
	curAbscissa       float64
	maxAbscissa       float64
}

// NewVCGenBSpline creates a new B-spline vertex generator
func NewVCGenBSpline() *VCGenBSpline {
	return &VCGenBSpline{
		srcVertices:       array.NewPodBVector[basics.Point[float64]](),
		splineX:           curves.NewBSpline(),
		splineY:           curves.NewBSpline(),
		interpolationStep: 1.0 / 50.0, // Default step
		closed:            false,
		status:            BSplineInitial,
		srcVertex:         0,
		curAbscissa:       0.0,
		maxAbscissa:       0.0,
	}
}

// SetInterpolationStep sets the interpolation step size
func (v *VCGenBSpline) SetInterpolationStep(step float64) {
	// Prevent edge cases with very small interpolation steps that can cause
	// excessive vertex generation and potential infinite loops
	const minStep = 1e-3 // Minimum step to prevent excessive vertex generation
	const maxStep = 1.0  // Maximum step for reasonable interpolation

	if step < minStep {
		v.interpolationStep = minStep
	} else if step > maxStep {
		v.interpolationStep = maxStep
	} else {
		v.interpolationStep = step
	}
}

// InterpolationStep returns the current interpolation step
func (v *VCGenBSpline) InterpolationStep() float64 {
	return v.interpolationStep
}

// RemoveAll clears all vertices
func (v *VCGenBSpline) RemoveAll() {
	v.srcVertices.RemoveAll()
	v.closed = false
	v.status = BSplineInitial
	v.srcVertex = 0
}

// AddVertex adds a vertex to the B-spline path
func (v *VCGenBSpline) AddVertex(x, y float64, cmd basics.PathCommand) {
	v.status = BSplineInitial

	if basics.IsMoveTo(cmd) {
		// In C++, MoveTo calls modify_last which replaces the last point
		// Only modify if we have points, otherwise add as first point
		if v.srcVertices.Size() > 0 {
			v.srcVertices.ModifyLast(basics.Point[float64]{X: x, Y: y})
		} else {
			v.srcVertices.Add(basics.Point[float64]{X: x, Y: y})
		}
	} else if basics.IsVertex(cmd) {
		v.srcVertices.Add(basics.Point[float64]{X: x, Y: y})
	} else if basics.IsEndPoly(cmd) {
		v.closed = basics.IsClosed(uint32(cmd))
	}
}

// PrepareSrc prepares the B-spline for vertex generation
func (v *VCGenBSpline) PrepareSrc() {
	// This method is called by conv_adaptor_vcgen
}

// Rewind rewinds the B-spline generator
func (v *VCGenBSpline) Rewind(pathID uint) {
	v.curAbscissa = 0.0
	v.maxAbscissa = 0.0
	v.srcVertex = 0

	// Multiple rewinds state management: prepare spline only on initial status
	// After first rewind, status becomes BSplineReady and we only reset parameters
	if v.status == BSplineInitial && v.srcVertices.Size() > 2 {
		if v.closed && v.srcVertices.Size() >= 3 {
			// For closed paths, we need extra points for continuity
			v.splineX = curves.NewBSplineWithCapacity(v.srcVertices.Size() + 8)
			v.splineY = curves.NewBSplineWithCapacity(v.srcVertices.Size() + 8)

			// Add wrap-around points for closed spline (matching C++ behavior)
			size := v.srcVertices.Size()
			// Add points before the start for continuity
			v.splineX.AddPoint(0.0, v.srcVertices.At(size-3).X)
			v.splineY.AddPoint(0.0, v.srcVertices.At(size-3).Y)
			v.splineX.AddPoint(1.0, v.srcVertices.At(size-3).X)
			v.splineY.AddPoint(1.0, v.srcVertices.At(size-3).Y)
			v.splineX.AddPoint(2.0, v.srcVertices.At(size-2).X)
			v.splineY.AddPoint(2.0, v.srcVertices.At(size-2).Y)
			v.splineX.AddPoint(3.0, v.srcVertices.At(size-1).X)
			v.splineY.AddPoint(3.0, v.srcVertices.At(size-1).Y)
		} else {
			v.splineX = curves.NewBSplineWithCapacity(v.srcVertices.Size())
			v.splineY = curves.NewBSplineWithCapacity(v.srcVertices.Size())
			v.closed = false // Disable closed mode if insufficient points
		}

		// Add all source vertices to the spline
		for i := 0; i < v.srcVertices.Size(); i++ {
			var x float64
			if v.closed {
				x = float64(i + 4)
			} else {
				x = float64(i)
			}
			point := v.srcVertices.At(i)
			v.splineX.AddPoint(x, point.X)
			v.splineY.AddPoint(x, point.Y)
		}

		v.curAbscissa = 0.0
		v.maxAbscissa = float64(v.srcVertices.Size() - 1)

		if v.closed {
			v.curAbscissa = 4.0
			v.maxAbscissa += 5.0 // In C++: += 5.0, not += 4.0
			// Add points after the end for continuity
			v.splineX.AddPoint(float64(v.srcVertices.Size()+4), v.srcVertices.At(0).X)
			v.splineY.AddPoint(float64(v.srcVertices.Size()+4), v.srcVertices.At(0).Y)
			v.splineX.AddPoint(float64(v.srcVertices.Size()+5), v.srcVertices.At(1).X)
			v.splineY.AddPoint(float64(v.srcVertices.Size()+5), v.srcVertices.At(1).Y)
			v.splineX.AddPoint(float64(v.srcVertices.Size()+6), v.srcVertices.At(2).X)
			v.splineY.AddPoint(float64(v.srcVertices.Size()+6), v.srcVertices.At(2).Y)
			// Add one more point if available
			if v.srcVertices.Size() > 3 {
				v.splineX.AddPoint(float64(v.srcVertices.Size()+7), v.srcVertices.At(3).X)
				v.splineY.AddPoint(float64(v.srcVertices.Size()+7), v.srcVertices.At(3).Y)
			} else {
				v.splineX.AddPoint(float64(v.srcVertices.Size()+7), v.srcVertices.At(0).X)
				v.splineY.AddPoint(float64(v.srcVertices.Size()+7), v.srcVertices.At(0).Y)
			}
		}

		v.splineX.Prepare()
		v.splineY.Prepare()
		v.status = BSplineReady
	} else {
		// For subsequent rewinds, just reset the abscissa values
		v.curAbscissa = 0.0
		v.maxAbscissa = float64(v.srcVertices.Size() - 1)

		if v.closed && v.srcVertices.Size() > 2 {
			v.curAbscissa = 4.0
			v.maxAbscissa += 5.0
		}
		v.status = BSplineReady
	}
}

// Vertex returns the next vertex in the B-spline
func (v *VCGenBSpline) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdLineTo

	for !basics.IsStop(cmd) {
		switch v.status {
		case BSplineInitial:
			// This should not happen in normal flow, but handle it
			v.Rewind(0)
			continue

		case BSplineReady:
			// Handle insufficient points case like C++
			if v.srcVertices.Size() < 2 {
				cmd = basics.PathCmdStop
				break
			}

			// Special case for exactly 2 points: output them directly (like C++)
			if v.srcVertices.Size() == 2 {
				if v.srcVertex == 0 {
					point := v.srcVertices.At(0)
					x, y = point.X, point.Y
					v.srcVertex++
					return x, y, basics.PathCmdMoveTo
				}
				if v.srcVertex == 1 {
					point := v.srcVertices.At(1)
					x, y = point.X, point.Y
					v.srcVertex++
					return x, y, basics.PathCmdLineTo
				}
				cmd = basics.PathCmdStop
				break
			}

			// For 3+ points, start spline generation
			cmd = basics.PathCmdMoveTo
			v.status = BSplinePolygon
			v.srcVertex = 0

		case BSplinePolygon:
			if v.curAbscissa >= v.maxAbscissa {
				if v.closed {
					v.status = BSplineEndPoly
					break
				} else {
					// For open splines, output the final vertex directly
					point := v.srcVertices.At(v.srcVertices.Size() - 1)
					x, y = point.X, point.Y
					v.status = BSplineEndPoly
					return x, y, basics.PathCmdLineTo
				}
			}

			// Use stateful interpolation for better performance (like C++)
			x = v.splineX.GetStateful(v.curAbscissa)
			y = v.splineY.GetStateful(v.curAbscissa)
			v.srcVertex++
			v.curAbscissa += v.interpolationStep

			// First vertex is MoveTo, rest are LineTo
			return x, y, func() basics.PathCommand {
				if v.srcVertex == 1 {
					return basics.PathCmdMoveTo
				}
				return basics.PathCmdLineTo
			}()

		case BSplineEndPoly:
			v.status = BSplineStop
			if v.closed {
				return 0, 0, basics.PathCmdEndPoly | basics.PathFlagClose
			} else {
				return 0, 0, basics.PathCmdEndPoly
			}

		case BSplineStop:
			return 0, 0, basics.PathCmdStop

		default:
			// Handle any unhandled state
			return 0, 0, basics.PathCmdStop
		}
	}
	return x, y, cmd
}
