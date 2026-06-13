// Port of AGG's scanline_boolean2.cpp – interactive demo with four rbox controls.
//
// Left-click / drag moves the spiral / shape-B centre.
package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	icol "github.com/cwbudde/agg_go/internal/color"
	ctrl "github.com/cwbudde/agg_go/internal/ctrl"
	"github.com/cwbudde/agg_go/internal/ctrl/rbox"
	"github.com/cwbudde/agg_go/internal/demo/scanlineboolean2"
)

const (
	frameWidth  = 655
	frameHeight = 520
)

type demo struct {
	polygons   *rbox.RboxCtrl[icol.RGBA]
	fillRule   *rbox.RboxCtrl[icol.RGBA]
	scanlineTy *rbox.RboxCtrl[icol.RGBA]
	operation  *rbox.RboxCtrl[icol.RGBA]
	mx, my     float64
	dragging   bool
}

func newDemo() *demo {
	d := &demo{
		mx: float64(frameWidth) * 0.5,
		my: float64(frameHeight) * 0.5,
	}

	d.polygons = rbox.NewDefaultRboxCtrl(5, 5, 210, 110, false)
	d.polygons.AddItem("Two Simple Paths")
	d.polygons.AddItem("Closed Stroke")
	d.polygons.AddItem("Great Britain and Arrows")
	d.polygons.AddItem("Great Britain and Spiral")
	d.polygons.AddItem("Spiral and Glyph")
	d.polygons.SetCurItem(3)

	d.fillRule = rbox.NewDefaultRboxCtrl(200, 5, 305, 50, false)
	d.fillRule.AddItem("Even-Odd")
	d.fillRule.AddItem("Non Zero")
	d.fillRule.SetCurItem(1)

	d.scanlineTy = rbox.NewDefaultRboxCtrl(300, 5, 415, 70, false)
	d.scanlineTy.AddItem("scanline_p")
	d.scanlineTy.AddItem("scanline_u")
	d.scanlineTy.AddItem("scanline_bin")
	d.scanlineTy.SetCurItem(1)

	d.operation = rbox.NewDefaultRboxCtrl(535, 5, 650, 145, false)
	d.operation.AddItem("None")
	d.operation.AddItem("OR")
	d.operation.AddItem("AND")
	d.operation.AddItem("XOR Linear")
	d.operation.AddItem("XOR Saddle")
	d.operation.AddItem("A-B")
	d.operation.AddItem("B-A")
	d.operation.SetCurItem(2)

	return d
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	scanlineboolean2.Draw(ctx, scanlineboolean2.Config{
		Mode:         d.polygons.CurItem(),
		FillRule:     d.fillRule.CurItem(),
		ScanlineType: d.scanlineTy.CurItem(),
		Operation:    d.operation.CurItem(),
		CenterX:      d.mx,
		CenterY:      d.my,
	})

	a := ctx.GetAgg2D()
	renderCtrl(a, d.polygons)
	renderCtrl(a, d.fillRule)
	renderCtrl(a, d.scanlineTy)
	renderCtrl(a, d.operation)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.polygons.OnMouseButtonDown(fx, fy) ||
		d.fillRule.OnMouseButtonDown(fx, fy) ||
		d.scanlineTy.OnMouseButtonDown(fx, fy) ||
		d.operation.OnMouseButtonDown(fx, fy) {
		return true
	}
	if btn.Left {
		d.mx = fx
		d.my = fy
		d.dragging = true
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.polygons.OnMouseMove(fx, fy, false) ||
		d.fillRule.OnMouseMove(fx, fy, false) ||
		d.scanlineTy.OnMouseMove(fx, fy, false) ||
		d.operation.OnMouseMove(fx, fy, false) {
		return true
	}
	if d.dragging && btn.Left {
		d.mx = fx
		d.my = fy
		return true
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, _ lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	d.dragging = false
	return d.polygons.OnMouseButtonUp(fx, fy) ||
		d.fillRule.OnMouseButtonUp(fx, fy) ||
		d.scanlineTy.OnMouseButtonUp(fx, fy) ||
		d.operation.OnMouseButtonUp(fx, fy)
}

// renderCtrl draws a single AGG control widget via the Agg2D rasterizer.
func renderCtrl(a *agg.Agg2D, c ctrl.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	adapter := &ctrlVS{c: c}
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
}

// ctrlVS adapts ctrl.Ctrl to the rasterizer vertex-source interface.
type ctrlVS struct{ c ctrl.Ctrl[icol.RGBA] }

func (v *ctrlVS) Rewind(pathID uint32) { v.c.Rewind(uint(pathID)) }
func (v *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.c.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// toAggColor converts an internal float-RGBA to a public agg.Color.
func toAggColor(c icol.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

func main() {
	lowlevelrunner.Run(runnerConfig(), newDemo())
}

func runnerConfig() lowlevelrunner.Config {
	return lowlevelrunner.Config{
		Title:                 "Scanline Boolean 2",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}
}
