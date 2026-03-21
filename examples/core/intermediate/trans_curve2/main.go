// Port of AGG C++ trans_curve2.cpp – double-path (two-rail) curve transform.
//
// Renders the lion through a "double path" transformer: two B-spline curves
// define the top and bottom rails and the lion is warped to fit between them.
// Default control points match the WASM demo initial state.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// Default spline control points (6 points each rail).
var (
	points1 = [12]float64{60, 40, 170, 130, 230, 270, 370, 330, 430, 470, 550, 550}
	points2 = [12]float64{40, 60, 150, 170, 210, 290, 350, 350, 410, 490, 530, 570}
)

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.RGBA(1.0, 1.0, 0.95, 1.0))

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	lionPaths := liondemo.Parse()

	// Find lion bounding box.
	lx1, ly1, lx2, ly2 := 1e9, 1e9, -1e9, -1e9
	for _, lp := range lionPaths {
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if x < lx1 {
				lx1 = x
			}
			if x > lx2 {
				lx2 = x
			}
			if y < ly1 {
				ly1 = y
			}
			if y > ly2 {
				ly2 = y
			}
		}
	}

	// Build the two path storages from the control points.
	buildSplinePath := func(pts [12]float64) *path.PathStorageStl {
		ps := path.NewPathStorageStl()
		ps.MoveTo(pts[0], pts[1])
		for i := 2; i < 12; i += 2 {
			ps.LineTo(pts[i], pts[i+1])
		}
		return ps
	}

	ps1 := buildSplinePath(points1)
	ps2 := buildSplinePath(points2)

	p1Adapter := path.NewPathStorageStlVertexSourceAdapter(ps1)
	p2Adapter := path.NewPathStorageStlVertexSourceAdapter(ps2)

	spline1 := conv.NewConvBSpline(p1Adapter)
	spline1.SetInterpolationStep(1.0 / 50.0)

	spline2 := conv.NewConvBSpline(p2Adapter)
	spline2.SetInterpolationStep(1.0 / 50.0)

	type splineAdapter struct{ s *conv.ConvBSpline }
	sp1 := &splineAdapter{s: spline1}
	sp2 := &splineAdapter{s: spline2}
	_ = sp1
	_ = sp2

	// Use double-path transformer.
	dp := transform.NewTransDoublePath()
	dp.AddPaths(spline1, spline2, 0, 0)
	dp.SetBaseLength(math.Abs(lx2-lx1) + 20)
	dp.SetBaseHeight(math.Abs(ly2 - ly1))

	for _, lp := range lionPaths {
		a.FillColor(agg.NewColor(lp.Color.R, lp.Color.G, lp.Color.B, 220))
		a.NoLine()
		a.ResetPath()

		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			// Shift lion to start at (0,0).
			tx, ty := x-lx1, y-ly1
			dp.Transform(&tx, &ty)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				a.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				a.LineTo(tx, ty)
			}
		}
		a.ClosePolygon()
		a.DrawPath(agg.FillOnly)
	}

	// Draw the control point splines.
	a.NoFill()
	a.LineColor(agg.NewColor(0, 80, 180, 200))
	a.LineWidth(1.5)
	for _, pts := range [][12]float64{points1, points2} {
		a.ResetPath()
		a.MoveTo(pts[0], pts[1])
		for i := 2; i < 12; i += 2 {
			a.LineTo(pts[i], pts[i+1])
		}
		a.DrawPath(agg.StrokeOnly)
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Trans Curve 2",
		Width:  600,
		Height: 600,
	}, &demo{})
}
