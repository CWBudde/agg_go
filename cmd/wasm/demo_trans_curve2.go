// Based on the original AGG examples: trans_curve2.cpp.
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
	transCurve2Points1   = [12]float64{60, 40, 170, 130, 230, 270, 370, 330, 430, 470, 550, 550}
	transCurve2Points2   = [12]float64{40, 60, 150, 170, 210, 290, 350, 350, 410, 490, 530, 570}
	transCurve2Selected  = -1
	transCurve2Animate   = false
	transCurve2DX1       [6]float64
	transCurve2DY1       [6]float64
	transCurve2DX2       [6]float64
	transCurve2DY2       [6]float64
)

func initTransCurve2Demo() {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}
	for i := 0; i < 6; i++ {
		transCurve2DX1[i] = (math.Mod(float64(i*1234+1), 10.0) - 5.0) * 0.5
		transCurve2DY1[i] = (math.Mod(float64(i*5678+2), 10.0) - 5.0) * 0.5
		transCurve2DX2[i] = (math.Mod(float64(i*1234+3), 10.0) - 5.0) * 0.5
		transCurve2DY2[i] = (math.Mod(float64(i*5678+4), 10.0) - 5.0) * 0.5
	}
}

type transDoubleAdapter struct {
	source *conv.ConvBSpline
}

func (a *transDoubleAdapter) Rewind(id uint) { a.source.Rewind(id) }
func (a *transDoubleAdapter) Vertex() (float64, float64, basics.PathCommand) {
	return a.source.Vertex()
}

func drawTransCurve2Demo() {
	initTransCurve2Demo()

	if transCurve2Animate {
		for i := 0; i < 6; i++ {
			moveTransCurve2Point(&transCurve2Points1[i*2], &transCurve2Points1[i*2+1], &transCurve2DX1[i], &transCurve2DY1[i])
			moveTransCurve2Point(&transCurve2Points2[i*2], &transCurve2Points2[i*2+1], &transCurve2DX2[i], &transCurve2DY2[i])
			// normalize distance
			d := math.Sqrt(math.Pow(transCurve2Points1[i*2]-transCurve2Points2[i*2], 2) + math.Pow(transCurve2Points1[i*2+1]-transCurve2Points2[i*2+1], 2))
			if d > 100 {
				transCurve2Points2[i*2] = transCurve2Points1[i*2] + (transCurve2Points2[i*2]-transCurve2Points1[i*2])*100/d
				transCurve2Points2[i*2+1] = transCurve2Points1[i*2+1] + (transCurve2Points2[i*2+1]-transCurve2Points1[i*2+1])*100/d
			}
		}
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Create guide paths
	ps1 := path.NewPathStorageStl()
	ps2 := path.NewPathStorageStl()
	ps1.MoveTo(transCurve2Points1[0], transCurve2Points1[1])
	ps2.MoveTo(transCurve2Points2[0], transCurve2Points2[1])
	for i := 1; i < 6; i++ {
		ps1.LineTo(transCurve2Points1[i*2], transCurve2Points1[i*2+1])
		ps2.LineTo(transCurve2Points2[i*2], transCurve2Points2[i*2+1])
	}

	bs1 := conv.NewConvBSpline(path.NewPathStorageStlVertexSourceAdapter(ps1))
	bs2 := conv.NewConvBSpline(path.NewPathStorageStlVertexSourceAdapter(ps2))
	bs1.SetInterpolationStep(1.0 / 40.0)
	bs2.SetInterpolationStep(1.0 / 40.0)

	// 2. Setup transformation
	tcurve := transform.NewTransDoublePath()
	tcurve.AddPaths(&transDoubleAdapter{bs1}, &transDoubleAdapter{bs2}, 0, 0)
	tcurve.SetBaseHeight(40.0)

	// 3. Transform the lion
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
	scaleX := tcurve.TotalLength1() / lionW * 0.8

	for _, lp := range lionPaths {
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 200))
		agg2d.NoLine()
		agg2d.ResetPath()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) { break }
			tx := (x - lx1) * scaleX
			ty := (y - (ly1+ly2)/2.0)
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

	// 4. Draw guide curves
	agg2d.NoFill()
	agg2d.LineWidth(1.0)
	agg2d.LineColor(agg.NewColor(170, 50, 20, 100))
	
	for _, bs := range []*conv.ConvBSpline{bs1, bs2} {
		agg2d.ResetPath()
		bs.Rewind(0)
		first := true
		for {
			vx, vy, cmd := bs.Vertex()
			if basics.IsStop(cmd) { break }
			if first { agg2d.MoveTo(vx, vy); first = false } else { agg2d.LineTo(vx, vy) }
		}
		agg2d.DrawPath(agg.StrokeOnly)
	}

	// 5. Draw handles
	for i := 0; i < 6; i++ {
		drawHandle(transCurve2Points1[i*2], transCurve2Points1[i*2+1])
		drawHandle(transCurve2Points2[i*2], transCurve2Points2[i*2+1])
	}
}

func moveTransCurve2Point(x, y, dx, dy *float64) {
	*x += *dx
	*y += *dy
	if *x < 0 || *x > float64(width) { *dx = -*dx }
	if *y < 0 || *y > float64(height) { *dy = -*dy }
}

func handleTransCurve2MouseDown(x, y float64) bool {
	transCurve2Selected = -1
	for i := 0; i < 6; i++ {
		if math.Sqrt(math.Pow(x-transCurve2Points1[i*2], 2)+math.Pow(y-transCurve2Points1[i*2+1], 2)) < 15 {
			transCurve2Selected = i
			return true
		}
		if math.Sqrt(math.Pow(x-transCurve2Points2[i*2], 2)+math.Pow(y-transCurve2Points2[i*2+1], 2)) < 15 {
			transCurve2Selected = i + 6
			return true
		}
	}
	return false
}

func handleTransCurve2MouseMove(x, y float64) bool {
	if transCurve2Selected != -1 {
		if transCurve2Selected < 6 {
			transCurve2Points1[transCurve2Selected*2] = x
			transCurve2Points1[transCurve2Selected*2+1] = y
		} else {
			idx := transCurve2Selected - 6
			transCurve2Points2[idx*2] = x
			transCurve2Points2[idx*2+1] = y
		}
		return true
	}
	return false
}

func handleTransCurve2MouseUp() {
	transCurve2Selected = -1
}

func toggleTransCurve2Animate() {
	transCurve2Animate = !transCurve2Animate
}
