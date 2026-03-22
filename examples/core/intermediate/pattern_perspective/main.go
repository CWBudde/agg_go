// Package main ports AGG's pattern_perspective.cpp demo.
package main

import (
	"flag"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	polygonctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	"github.com/MeKo-Christian/agg_go/internal/demo/patternperspective"
)

const (
	frameWidth  = 600
	frameHeight = 600
)

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

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
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

type demo struct {
	mode      int
	quad      *polygonctrl.PolygonCtrl[icol.RGBA]
	transType *rboxctrl.RboxCtrl[icol.RGBA]
	controls  []ctrlbase.Ctrl[icol.RGBA]
}

func newDemo() *demo {
	quad := polygonctrl.NewDefaultPolygonCtrl(4, 5.0)
	quad.SetClose(true)
	quad.SetInPolygonCheck(true)
	quad.SetXn(0, 100)
	quad.SetYn(0, 100)
	quad.SetXn(1, 500)
	quad.SetYn(1, 100)
	quad.SetXn(2, 500)
	quad.SetYn(2, 500)
	quad.SetXn(3, 100)
	quad.SetYn(3, 500)

	transType := rboxctrl.NewDefaultRboxCtrl(460, 540.0, 590.0, 595.0, true)
	transType.SetTextSize(8, 0)
	transType.SetTextThickness(1.0)
	transType.AddItem("Affine")
	transType.AddItem("Bilinear")
	transType.AddItem("Perspective")
	transType.SetCurItem(2)

	return &demo{
		mode:      2,
		quad:      quad,
		transType: transType,
		controls:  []ctrlbase.Ctrl[icol.RGBA]{transType},
	}
}

func (d *demo) quadPoints() [4][2]float64 {
	return [4][2]float64{
		{d.quad.Xn(0), d.quad.Yn(0)},
		{d.quad.Xn(1), d.quad.Yn(1)},
		{d.quad.Xn(2), d.quad.Yn(2)},
		{d.quad.Xn(3), d.quad.Yn(3)},
	}
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	patternperspective.Draw(ctx, patternperspective.Config{
		Mode: d.transType.CurItem(),
		Quad: d.quadPoints(),
	})

	a := ctx.GetAgg2D()
	for _, ctrl := range d.controls {
		renderCtrl(a, ctrl)
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}

	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonDown(fx, fy) {
			d.mode = d.transType.CurItem()
			return true
		}
	}
	return d.quad.OnMouseButtonDown(fx, fy)
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false

	for _, ctrl := range d.controls {
		if ctrl.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}
	if d.quad.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}

	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	_ = btn
	fx, fy := float64(x), float64(y)
	redraw := false

	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonUp(fx, fy) {
			redraw = true
		}
	}
	if d.quad.OnMouseButtonUp(fx, fy) {
		redraw = true
	}

	return redraw
}

func main() {
	mode := flag.Int("mode", 2, "0=Affine, 1=Bilinear, 2=Perspective")
	flag.Parse()

	d := newDemo()
	d.mode = *mode
	d.transType.SetCurItem(*mode)

	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Pattern Perspective",
		Width:  frameWidth,
		Height: frameHeight,
	}, d)
}
