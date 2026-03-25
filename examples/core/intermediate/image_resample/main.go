// Package main ports AGG's image_resample.cpp demo.
package main

import (
	"fmt"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	polygonctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageresample"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
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

type gsvVertexSourceAdapter struct {
	src *gsv.GSVTextOutline
}

func (a *gsvVertexSourceAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *gsvVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type demo struct {
	quad      *polygonctrl.PolygonCtrl[icol.RGBA]
	transType *rboxctrl.RboxCtrl[icol.RGBA]
	blur      *sliderctrl.SliderCtrl
	controls  []ctrlbase.Ctrl[icol.RGBA]
}

func newDemo() *demo {
	srcW, srcH, ok := imageresample.SourceSize()
	if !ok {
		srcW, srcH = 320, 300
	}
	x1 := float64(frameWidth-srcW) / 2.0
	y1 := float64(frameHeight-srcH) / 2.0

	quad := polygonctrl.NewDefaultPolygonCtrl(4, 5.0)
	quad.SetClose(true)
	quad.SetInPolygonCheck(true)
	quad.SetXn(0, x1)
	quad.SetYn(0, y1)
	quad.SetXn(1, x1+float64(srcW))
	quad.SetYn(1, y1)
	quad.SetXn(2, x1+float64(srcW))
	quad.SetYn(2, y1+float64(srcH))
	quad.SetXn(3, x1)
	quad.SetYn(3, y1+float64(srcH))

	transType := rboxctrl.NewDefaultRboxCtrl(400, 5.0, 600.0, 100.0, false)
	transType.SetTextSize(7, 0)
	transType.AddItem("Affine No Resample")
	transType.AddItem("Affine Resample")
	transType.AddItem("Perspective No Resample LERP")
	transType.AddItem("Perspective No Resample Exact")
	transType.AddItem("Perspective Resample LERP")
	transType.AddItem("Perspective Resample Exact")
	transType.SetCurItem(4)

	blur := sliderctrl.NewSliderCtrl(5.0, 5.0, 400-5.0, 10.0, false)
	blur.SetRange(0.5, 5.0)
	blur.SetValue(1.0)
	blur.SetLabel("Blur=%.3f")

	return &demo{
		quad:      quad,
		transType: transType,
		blur:      blur,
		controls:  []ctrlbase.Ctrl[icol.RGBA]{transType, blur},
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
	elapsed := imageresample.DrawTimed(ctx, imageresample.Config{
		Mode: d.transType.CurItem(),
		Blur: d.blur.Value(),
		Quad: d.quadPoints(),
	})

	a := ctx.GetAgg2D()
	renderTimingText(a, float64(elapsed)/1e6)

	for _, ctrl := range d.controls {
		renderCtrl(a, ctrl)
	}
}

func renderTimingText(a *agg.Agg2D, ms float64) {
	text := fmt.Sprintf("%3.2f ms", ms)
	gsvText := gsv.NewGSVText()
	gsvText.SetSize(10.0, 0)
	gsvText.SetStartPoint(10.0, 70.0)
	gsvText.SetText(text)

	outline := gsv.NewGSVTextOutline(gsvText)
	outline.SetWidth(1.5)

	ras := a.GetInternalRasterizer()
	ras.Reset()
	ras.AddPath(&gsvVertexSourceAdapter{src: outline}, 0)
	a.RenderRasterizerWithColor(agg.Black)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}

	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonDown(fx, fy) {
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

func (d *demo) OnKey(key rune) bool {
	if key != ' ' {
		return false
	}

	quad := d.quadPoints()
	imageresample.RotateQuad90(&quad)
	for i := range quad {
		d.quad.SetXn(uint(i), quad[i][0])
		d.quad.SetYn(uint(i), quad[i][1])
	}
	return true
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image Resample",
		Width:  frameWidth,
		Height: frameHeight,
		FlipY:  true,
	}, newDemo())
}
