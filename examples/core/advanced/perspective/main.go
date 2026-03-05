// Port of AGG C++ perspective.cpp – perspective / bilinear transformation.
//
// Renders the lion vector art through a bilinear (default) or perspective
// transform defined by a 4-corner quadrilateral. The quad defaults to a
// slight perspective distortion centred on the 800×600 canvas.
package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
	"agg_go/internal/basics"
	liondemo "agg_go/internal/demo/lion"
	"agg_go/internal/transform"
)

func main() {
	const width, height = 800, 600

	ctx := agg.NewContext(width, height)
	ctx.Clear(agg.RGBA(0.95, 0.95, 0.85, 1.0))

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	lionPaths := liondemo.Parse()

	// Find bounding box of the lion.
	x1, y1, x2, y2 := 1e9, 1e9, -1e9, -1e9
	for _, lp := range lionPaths {
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if x < x1 {
				x1 = x
			}
			if x > x2 {
				x2 = x
			}
			if y < y1 {
				y1 = y
			}
			if y > y2 {
				y2 = y
			}
		}
	}

	// Define destination quadrilateral (slight perspective effect).
	cx, cy := float64(width)/2, float64(height)/2
	w, h := (x2-x1)*0.8, (y2-y1)*0.8
	quad := [8]float64{
		cx - w/2 + 30, cy - h/2, // top-left (shifted in)
		cx + w/2, cy - h/2, // top-right
		cx + w/2 - 30, cy + h/2, // bottom-right (shifted in)
		cx - w/2, cy + h/2, // bottom-left
	}

	// Bilinear transform from lion bbox to quad.
	tr := transform.NewTransBilinearRectToQuad(x1, y1, x2, y2, quad)
	if !tr.IsValid() {
		panic("bilinear transform is not valid")
	}

	for _, lp := range lionPaths {
		a.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		a.NoLine()
		a.ResetPath()

		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			tx, ty := x, y
			tr.Transform(&tx, &ty)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				a.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				a.LineTo(tx, ty)
			}
		}
		a.ClosePolygon()
		a.DrawPath(agg.FillOnly)
	}

	// Draw the quad outline.
	a.NoFill()
	a.LineColor(agg.NewColor(0, 0, 80, 180))
	a.LineWidth(1.5)
	a.ResetPath()
	a.MoveTo(quad[0], quad[1])
	a.LineTo(quad[2], quad[3])
	a.LineTo(quad[4], quad[5])
	a.LineTo(quad[6], quad[7])
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)

	const filename = "perspective.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
