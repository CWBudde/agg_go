// Based on the original AGG examples: trans_curve1.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/conv"
	liondemo "agg_go/internal/demo/lion"
	"agg_go/internal/path"
	"agg_go/internal/transform"
)

var (
	transCurvePoints    = [12]float64{50, 50, 170, 130, 230, 270, 370, 330, 430, 470, 550, 550}
	transCurveSelected  = -1
	transCurveAnimate   = false
	transCurveDX        [6]float64
	transCurveDY        [6]float64
)

func initTransCurveDemo() {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}
	for i := 0; i < 6; i++ {
		transCurveDX[i] = (math.Mod(float64(i*1234), 10.0) - 5.0) * 0.5
		transCurveDY[i] = (math.Mod(float64(i*5678), 10.0) - 5.0) * 0.5
	}
}

type transSingleAdapter struct {
	source interface {
		Rewind(uint)
		NextVertex() (float64, float64, uint32)
	}
}

func (a *transSingleAdapter) Rewind(id uint) { a.source.Rewind(id) }
func (a *transSingleAdapter) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.source.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

func drawTransCurveDemo() {
	initTransCurveDemo()

	if transCurveAnimate {
		for i := 0; i < 6; i++ {
			transCurvePoints[i*2] += transCurveDX[i]
			transCurvePoints[i*2+1] += transCurveDY[i]
			if transCurvePoints[i*2] < 0 || transCurvePoints[i*2] > float64(width) { transCurveDX[i] = -transCurveDX[i] }
			if transCurvePoints[i*2+1] < 0 || transCurvePoints[i*2+1] > float64(height) { transCurveDY[i] = -transCurveDY[i] }
		}
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Create the base path (control polygon)
	ps := path.NewPathStorageStl()
	ps.MoveTo(transCurvePoints[0], transCurvePoints[1])
	for i := 1; i < 6; i++ {
		ps.LineTo(transCurvePoints[i*2], transCurvePoints[i*2+1])
	}

	// 2. Smooth it with B-Spline
	bspline := conv.NewConvBSpline(ps)
	bspline.SetInterpolationStep(1.0 / 40.0)

	// 3. Create the transformation
	tcurve := transform.NewTransSinglePath()
	tcurve.AddPath(&transSingleAdapter{bspline})

	// 4. Transform the lion along the curve
	lx1, ly1, lx2, ly2 := 1e9, 1e9, -1e9, -1e9
	for _, lp := range lionPaths {
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) { break }
			if x < lx1 { lx1 = x }
			if x > lx2 { lx2 = x }
			if y < ly1 { ly1 = y }
			if y > ly2 { ly2 = y }
		}
	}
	
	lionW := lx2 - lx1
	scaleX := tcurve.TotalLength() / lionW * 0.8
	scaleY := 0.5 // flatten it a bit

	for _, lp := range lionPaths {
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 200))
		agg2d.NoLine()

		agg2d.ResetPath()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) { break }

			tx := (x - lx1) * scaleX
			ty := (y - (ly1+ly2)/2.0) * scaleY
			tcurve.Transform(&tx, &ty)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(tx, ty)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}

	// 5. Draw the curve itself
	agg2d.LineColor(agg.NewColor(170, 50, 20, 100))
	agg2d.LineWidth(2.0)
	agg2d.NoFill()
	agg2d.ResetPath()
	bspline.Rewind(0)
	first := true
	for {
		vx, vy, cmd := bspline.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) { break }
		if first {
			agg2d.MoveTo(vx, vy)
			first = false
		} else {
			agg2d.LineTo(vx, vy)
		}
	}
	agg2d.DrawPath(agg.StrokeOnly)

	// 6. Draw handles
	for i := 0; i < 6; i++ {
		drawHandle(transCurvePoints[i*2], transCurvePoints[i*2+1])
	}
}

func handleTransCurveMouseDown(x, y float64) bool {
	transCurveSelected = -1
	for i := 0; i < 6; i++ {
		dist := math.Sqrt(math.Pow(x-transCurvePoints[i*2], 2) + math.Pow(y-transCurvePoints[i*2+1], 2))
		if dist < 15 {
			transCurveSelected = i
			return true
		}
	}
	return false
}

func handleTransCurveMouseMove(x, y float64) bool {
	if transCurveSelected != -1 {
		transCurvePoints[transCurveSelected*2] = x
		transCurvePoints[transCurveSelected*2+1] = y
		return true
	}
	return false
}

func handleTransCurveMouseUp() {
	transCurveSelected = -1
}

func toggleTransCurveAnimate() {
	transCurveAnimate = !transCurveAnimate
}
