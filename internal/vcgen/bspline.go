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
	// TODO: Fix edge cases with very small interpolation steps
	// Very small steps can cause excessive vertex generation and potential infinite loops
	// Add reasonable minimum step size to prevent performance issues
	const minStep = 1e-4 // Minimum step to prevent excessive vertex generation
	if step < minStep {
		v.interpolationStep = minStep
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
		v.srcVertices.ModifyLast(basics.Point[float64]{X: x, Y: y})
	} else {
		if basics.IsVertex(cmd) {
			v.srcVertices.Add(basics.Point[float64]{X: x, Y: y})
		} else {
			v.closed = basics.IsClosed(uint32(cmd))
		}
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

	// TODO: Fix multiple rewinds state management
	// Issue: Multiple rewinds should produce same results, but spline is only prepared once
	// when status == BSplineInitial. On subsequent rewinds, we need to reset properly.

	// Check if we have sufficient vertices for B-spline generation
	if v.srcVertices.Size() <= 2 {
		v.status = BSplineStop
		return
	}

	if v.status == BSplineInitial && v.srcVertices.Size() > 2 {
		if v.closed {
			// For closed paths, we need extra points for continuity
			v.splineX = curves.NewBSplineWithCapacity(v.srcVertices.Size() + 8)
			v.splineY = curves.NewBSplineWithCapacity(v.srcVertices.Size() + 8)

			// Add wrap-around points for closed spline
			size := v.srcVertices.Size()
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
			v.maxAbscissa += 4.0
		}

		v.splineX.Prepare()
		v.splineY.Prepare()
		v.status = BSplineReady
	} else if v.srcVertices.Size() > 2 {
		// Reset for multiple rewinds - spline is already prepared
		v.curAbscissa = 0.0
		v.maxAbscissa = float64(v.srcVertices.Size() - 1)

		if v.closed {
			v.curAbscissa = 4.0
			v.maxAbscissa += 4.0
		}
	}

	v.status = BSplinePolygon
}

// Vertex returns the next vertex in the B-spline
func (v *VCGenBSpline) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdLineTo

	for {
		switch v.status {
		case BSplinePolygon:
			if v.srcVertex == 0 {
				v.srcVertex = 1
				cmd = basics.PathCmdMoveTo
				// TODO: Fix B-spline generator state management
				// Issue: When srcVertices.Size() <= 2, the spline is not prepared but we're still
				// trying to generate vertices. Should return Stop for insufficient points.
				if v.srcVertices.Size() > 2 {
					x = v.splineX.Get(v.curAbscissa)
					y = v.splineY.Get(v.curAbscissa)
					v.curAbscissa += v.interpolationStep
					return x, y, cmd
				} else {
					// Not enough points for B-spline, go to stop
					v.status = BSplineStop
					return 0, 0, basics.PathCmdStop
				}
			} else {
				if v.curAbscissa >= v.maxAbscissa {
					x = v.splineX.Get(v.maxAbscissa)
					y = v.splineY.Get(v.maxAbscissa)
					v.status = BSplineEndPoly
					return x, y, basics.PathCmdLineTo
				} else {
					x = v.splineX.Get(v.curAbscissa)
					y = v.splineY.Get(v.curAbscissa)
					v.curAbscissa += v.interpolationStep
					return x, y, basics.PathCmdLineTo
				}
			}

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
			// TODO: Handle cases where B-spline is not ready (insufficient points, RemoveAll, etc)
			return 0, 0, basics.PathCmdStop
		}
	}
}
