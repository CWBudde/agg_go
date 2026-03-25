// Package main ports AGG's image_perspective.cpp demo.
package main

import (
	"fmt"
	"math"
	"time"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageperspective"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
)

const handleRadius = 5.0

type demo struct {
	quad      [4][2]float64
	dragIdx   int
	transType *rboxctrl.RboxCtrl[icol.RGBA]
}

type ctrlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func toAggColor(c icol.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255.0 + 0.5)
		}
	}

	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

func renderCtrl(
	a *agg.Agg2D,
	ras interface {
		Reset()
		AddPath(vs rasterizer.VertexSource, pathID uint32)
	},
	c ctrlbase.Ctrl[icol.RGBA],
) {
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
}

func newDemo() *demo {
	transType := rboxctrl.NewDefaultRboxCtrl(420, 5.0, 420+170.0, 70.0, false)
	transType.SetTextSize(8, 0)
	transType.SetTextThickness(1.0)
	transType.AddItem("Affine Parallelogram")
	transType.AddItem("Bilinear")
	transType.AddItem("Perspective")
	transType.SetCurItem(2)

	return &demo{
		quad: [4][2]float64{
			{100, 100},
			{500, 100},
			{500, 500},
			{100, 500},
		},
		dragIdx:   -1,
		transType: transType,
	}
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)

	start := time.Now()
	imageperspective.Draw(ctx, imageperspective.Config{
		Mode: d.transType.CurItem(),
		Quad: d.quad,
	})
	elapsed := time.Since(start)

	a := ctx.GetAgg2D()
	renderCtrl(a, a.GetInternalRasterizer(), d.transType)

	a.FontGSV(10)
	a.FlipText(false)
	a.FillColor(agg.Black)
	a.NoLine()
	a.Text(10, 10, fmt.Sprintf("%3.2f ms", float64(elapsed)/float64(time.Millisecond)), false, 0, 0)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}

	fx, fy := float64(x), float64(y)
	if d.transType.OnMouseButtonDown(fx, fy) {
		return true
	}
	for i, pt := range d.quad {
		dx := fx - pt[0]
		dy := fy - pt[1]
		if math.Sqrt(dx*dx+dy*dy) <= handleRadius {
			d.dragIdx = i
			return true
		}
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	_ = x
	_ = y

	redraw := d.transType.OnMouseButtonUp(float64(x), float64(y))
	if d.dragIdx >= 0 {
		d.dragIdx = -1
		redraw = true
	}
	return redraw
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	redraw := d.transType.OnMouseMove(float64(x), float64(y), btn.Left)
	if d.dragIdx < 0 || !btn.Left {
		return redraw
	}

	d.quad[d.dragIdx] = [2]float64{float64(x), float64(y)}
	return true
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image Perspective",
		Width:  600,
		Height: 600,
		FlipY:  true,
	}, newDemo())
}
