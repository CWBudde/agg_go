package vcgen

import (
	"github.com/MeKo-Christian/agg_go/internal/array"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/curves"
)

// BSplineStatus matches AGG's vcgen_bspline state machine.
type BSplineStatus int

const (
	BSplineInitial BSplineStatus = iota
	BSplineReady
	BSplinePolygon
	BSplineEndPoly
	BSplineStop
)

// VCGenBSpline is a direct port of AGG's vcgen_bspline.
type VCGenBSpline struct {
	srcVertices       *array.PodBVector[basics.PointD]
	splineX           *curves.BSpline
	splineY           *curves.BSpline
	interpolationStep float64
	closed            uint32
	status            BSplineStatus
	srcVertex         uint
	curAbscissa       float64
	maxAbscissa       float64
}

// NewVCGenBSpline creates a new B-spline vertex generator.
func NewVCGenBSpline() *VCGenBSpline {
	return &VCGenBSpline{
		srcVertices:       array.NewPodBVector[basics.PointD](),
		splineX:           curves.NewBSpline(),
		splineY:           curves.NewBSpline(),
		interpolationStep: 1.0 / 50.0,
		closed:            0,
		status:            BSplineInitial,
		srcVertex:         0,
	}
}

// SetInterpolationStep sets the interpolation step size.
func (v *VCGenBSpline) SetInterpolationStep(step float64) {
	v.interpolationStep = step
}

// InterpolationStep returns the current interpolation step.
func (v *VCGenBSpline) InterpolationStep() float64 {
	return v.interpolationStep
}

// RemoveAll clears all vertices.
func (v *VCGenBSpline) RemoveAll() {
	v.srcVertices.RemoveAll()
	v.closed = 0
	v.status = BSplineInitial
	v.srcVertex = 0
}

// AddVertex adds a vertex to the generator.
func (v *VCGenBSpline) AddVertex(x, y float64, cmd basics.PathCommand) {
	v.status = BSplineInitial

	switch {
	case basics.IsMoveTo(cmd):
		v.srcVertices.ModifyLast(basics.PointD{X: x, Y: y})
	case basics.IsVertex(cmd):
		v.srcVertices.Add(basics.PointD{X: x, Y: y})
	default:
		v.closed = basics.GetCloseFlag(uint32(cmd))
	}
}

// PrepareSrc is part of the vcgen.VertexGenerator contract.
func (v *VCGenBSpline) PrepareSrc() {}

// Rewind prepares the generator for vertex iteration.
func (v *VCGenBSpline) Rewind(pathID uint) {
	v.curAbscissa = 0.0
	v.maxAbscissa = 0.0
	v.srcVertex = 0

	if v.status == BSplineInitial && v.srcVertices.Size() > 2 {
		if v.closed != 0 {
			v.splineX.Init(v.srcVertices.Size() + 8)
			v.splineY.Init(v.srcVertices.Size() + 8)

			v.splineX.AddPoint(0.0, v.srcPointPrev(v.srcVertices.Size()-3).X)
			v.splineY.AddPoint(0.0, v.srcPointPrev(v.srcVertices.Size()-3).Y)
			v.splineX.AddPoint(1.0, v.srcVertices.At(v.srcVertices.Size()-3).X)
			v.splineY.AddPoint(1.0, v.srcVertices.At(v.srcVertices.Size()-3).Y)
			v.splineX.AddPoint(2.0, v.srcVertices.At(v.srcVertices.Size()-2).X)
			v.splineY.AddPoint(2.0, v.srcVertices.At(v.srcVertices.Size()-2).Y)
			v.splineX.AddPoint(3.0, v.srcVertices.At(v.srcVertices.Size()-1).X)
			v.splineY.AddPoint(3.0, v.srcVertices.At(v.srcVertices.Size()-1).Y)
		} else {
			v.splineX.Init(v.srcVertices.Size())
			v.splineY.Init(v.srcVertices.Size())
		}

		for i := 0; i < v.srcVertices.Size(); i++ {
			x := float64(i)
			if v.closed != 0 {
				x += 4.0
			}
			pt := v.srcVertices.At(i)
			v.splineX.AddPoint(x, pt.X)
			v.splineY.AddPoint(x, pt.Y)
		}

		v.curAbscissa = 0.0
		v.maxAbscissa = float64(v.srcVertices.Size() - 1)
		if v.closed != 0 {
			v.curAbscissa = 4.0
			v.maxAbscissa += 5.0
			v.splineX.AddPoint(float64(v.srcVertices.Size()+4), v.srcVertices.At(0).X)
			v.splineY.AddPoint(float64(v.srcVertices.Size()+4), v.srcVertices.At(0).Y)
			v.splineX.AddPoint(float64(v.srcVertices.Size()+5), v.srcVertices.At(1).X)
			v.splineY.AddPoint(float64(v.srcVertices.Size()+5), v.srcVertices.At(1).Y)
			v.splineX.AddPoint(float64(v.srcVertices.Size()+6), v.srcVertices.At(2).X)
			v.splineY.AddPoint(float64(v.srcVertices.Size()+6), v.srcVertices.At(2).Y)
			v.splineX.AddPoint(float64(v.srcVertices.Size()+7), v.srcPointNext(2).X)
			v.splineY.AddPoint(float64(v.srcVertices.Size()+7), v.srcPointNext(2).Y)
		}

		v.splineX.Prepare()
		v.splineY.Prepare()
	}

	v.status = BSplineReady
}

// Vertex returns the next generated vertex.
func (v *VCGenBSpline) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdLineTo

	for !basics.IsStop(cmd) {
		switch v.status {
		case BSplineInitial:
			v.Rewind(0)

		case BSplineReady:
			if v.srcVertices.Size() < 2 {
				cmd = basics.PathCmdStop
				break
			}

			if v.srcVertices.Size() == 2 {
				pt := v.srcVertices.At(int(v.srcVertex))
				x, y = pt.X, pt.Y
				v.srcVertex++
				if v.srcVertex == 1 {
					return x, y, basics.PathCmdMoveTo
				}
				if v.srcVertex == 2 {
					return x, y, basics.PathCmdLineTo
				}
				cmd = basics.PathCmdStop
				break
			}

			cmd = basics.PathCmdMoveTo
			v.status = BSplinePolygon
			v.srcVertex = 0

		case BSplinePolygon:
			if v.curAbscissa >= v.maxAbscissa {
				if v.closed != 0 {
					v.status = BSplineEndPoly
					break
				}

				pt := v.srcVertices.At(v.srcVertices.Size() - 1)
				x, y = pt.X, pt.Y
				v.status = BSplineEndPoly
				return x, y, basics.PathCmdLineTo
			}

			x = v.splineX.GetStateful(v.curAbscissa)
			y = v.splineY.GetStateful(v.curAbscissa)
			v.srcVertex++
			v.curAbscissa += v.interpolationStep
			if v.srcVertex == 1 {
				return x, y, basics.PathCmdMoveTo
			}
			return x, y, basics.PathCmdLineTo

		case BSplineEndPoly:
			v.status = BSplineStop
			if v.closed != 0 {
				return 0, 0, basics.PathCommand(uint32(basics.PathCmdEndPoly) | v.closed)
			}
			return 0, 0, basics.PathCmdEndPoly

		case BSplineStop:
			return 0, 0, basics.PathCmdStop
		}
	}

	return 0, 0, cmd
}

func (v *VCGenBSpline) srcPointPrev(idx int) basics.PointD {
	size := v.srcVertices.Size()
	if size == 0 {
		return basics.PointD{}
	}
	return v.srcVertices.At((idx + size - 1) % size)
}

func (v *VCGenBSpline) srcPointNext(idx int) basics.PointD {
	size := v.srcVertices.Size()
	if size == 0 {
		return basics.PointD{}
	}
	return v.srcVertices.At((idx + 1) % size)
}
