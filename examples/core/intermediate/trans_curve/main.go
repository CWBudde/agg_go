// Based on the original AGG example: trans_curve1.cpp
// Transforms the lion paths along a B-Spline curve (single-path transform).
package main

import (
	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/basics"
	"agg_go/internal/conv"
	liondemo "agg_go/internal/demo/lion"
	"agg_go/internal/path"
	"agg_go/internal/transform"
)

// transSingleAdapter adapts ConvBSpline to transform.VertexSource
type transSingleAdapter struct {
	source *conv.ConvBSpline
}

func (a *transSingleAdapter) Rewind(id uint) { a.source.Rewind(id) }
func (a *transSingleAdapter) Vertex() (float64, float64, basics.PathCommand) {
	return a.source.Vertex()
}

func drawHandle(ctx *agg.Context, x, y float64) {
	ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
	ctx.FillCircle(x, y, 5)
	ctx.SetColor(agg.Black)
	ctx.DrawCircle(x, y, 5)
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	// Default control polygon (from WASM demo defaults)
	points := [12]float64{50, 50, 170, 130, 230, 270, 370, 330, 430, 470, 550, 550}

	ctx.Clear(agg.White)
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Create the base path (control polygon)
	ps := path.NewPathStorageStl()
	ps.MoveTo(points[0], points[1])
	for i := 1; i < 6; i++ {
		ps.LineTo(points[i*2], points[i*2+1])
	}

	// 2. Smooth it with B-Spline
	psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)
	bspline := conv.NewConvBSpline(psAdapter)
	bspline.SetInterpolationStep(1.0 / 40.0)

	// 3. Create the transformation
	tcurve := transform.NewTransSinglePath()
	tcurve.AddPath(&transSingleAdapter{bspline}, 0)

	// 4. Find bounding box of the lion
	lionPaths := liondemo.Parse()
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

	lionW := lx2 - lx1
	scaleX := tcurve.TotalLength() / lionW * 0.8
	scaleY := 0.5 // flatten it a bit

	// 5. Transform and render the lion
	for _, lp := range lionPaths {
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 200))
		agg2d.NoLine()

		agg2d.ResetPath()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			tx := (x - lx1) * scaleX
			ty := (y - (ly1+ly2)*0.5) * scaleY
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

	// 6. Draw the curve itself
	agg2d.LineColor(agg.NewColor(170, 50, 20, 100))
	agg2d.LineWidth(2.0)
	agg2d.NoFill()
	agg2d.ResetPath()
	bspline.Rewind(0)
	first := true
	for {
		vx, vy, cmd := bspline.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if first {
			agg2d.MoveTo(vx, vy)
			first = false
		} else {
			agg2d.LineTo(vx, vy)
		}
	}
	agg2d.DrawPath(agg.StrokeOnly)

	// 7. Draw control-point handles
	for i := 0; i < 6; i++ {
		drawHandle(ctx, points[i*2], points[i*2+1])
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Trans Curve",
		Width:  800,
		Height: 600,
	}, &demo{})
}
