// Port of AGG C++ perspective.cpp – perspective / bilinear transformation.
//
// Renders the lion vector art through a bilinear (default) or perspective
// transform defined by a 4-corner quadrilateral. The quad defaults to a
// slight perspective distortion centred on the 600×600 canvas.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const handleRadius = 8.0

type demo struct {
	quad    [8]float64
	dragIdx int
	// lion bounding box (computed once)
	lx1, ly1, lx2, ly2 float64
}

func newDemo() *demo {
	const width, height = 600, 600

	ld := liondemo.Parse()

	// Find bounding box of the lion.
	lx1, ly1, lx2, ly2 := 1e9, 1e9, -1e9, -1e9
	for idx := uint(0); idx < ld.Path.TotalVertices(); idx++ {
		x, y, cmd := ld.Path.Vertex(idx)
		if !basics.IsVertex(basics.PathCommand(cmd)) {
			continue
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

	// Define destination quadrilateral (slight perspective effect).
	cx, cy := float64(width)/2, float64(height)/2
	w, h := (lx2-lx1)*0.8, (ly2-ly1)*0.8
	quad := [8]float64{
		cx - w/2 + 30, cy - h/2, // top-left (shifted in)
		cx + w/2, cy - h/2, // top-right
		cx + w/2 - 30, cy + h/2, // bottom-right (shifted in)
		cx - w/2, cy + h/2, // bottom-left
	}

	return &demo{
		quad:    quad,
		dragIdx: -1,
		lx1:     lx1,
		ly1:     ly1,
		lx2:     lx2,
		ly2:     ly2,
	}
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.RGBA(0.95, 0.95, 0.85, 1.0))

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	ld := liondemo.Parse()

	// Bilinear transform from lion bbox to quad.
	tr := transform.NewTransBilinearRectToQuad(d.lx1, d.ly1, d.lx2, d.ly2, d.quad)
	if !tr.IsValid() {
		return
	}

	for i := 0; i < ld.NPaths; i++ {
		a.FillColor(agg.NewColor(ld.Colors[i].R, ld.Colors[i].G, ld.Colors[i].B, 255))
		a.NoLine()
		a.ResetPath()

		ld.Path.Rewind(ld.PathIdx[i])
		for {
			x, y, cmd := ld.Path.NextVertex()
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
	a.MoveTo(d.quad[0], d.quad[1])
	a.LineTo(d.quad[2], d.quad[3])
	a.LineTo(d.quad[4], d.quad[5])
	a.LineTo(d.quad[6], d.quad[7])
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)

	// Draw drag handles at each corner.
	for i := 0; i < 4; i++ {
		hx, hy := d.quad[i*2], d.quad[i*2+1]
		a.FillColor(agg.NewColor(204, 51, 26, 153))
		a.NoLine()
		a.ResetPath()
		a.AddEllipse(hx, hy, handleRadius, handleRadius, agg.CCW)
		a.DrawPath(agg.FillOnly)

		a.NoFill()
		a.LineColor(agg.NewColor(0, 0, 0, 255))
		a.LineWidth(1.0)
		a.ResetPath()
		a.AddEllipse(hx, hy, handleRadius, handleRadius, agg.CCW)
		a.DrawPath(agg.StrokeOnly)
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	for i := 0; i < 4; i++ {
		dx := fx - d.quad[i*2]
		dy := fy - d.quad[i*2+1]
		if math.Sqrt(dx*dx+dy*dy) <= handleRadius {
			d.dragIdx = i
			return true
		}
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	if d.dragIdx >= 0 {
		d.dragIdx = -1
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if d.dragIdx < 0 || !btn.Left {
		return false
	}
	d.quad[d.dragIdx*2] = float64(x)
	d.quad[d.dragIdx*2+1] = float64(y)
	return true
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Perspective",
		Width:  600,
		Height: 600,
	}, newDemo())
}
