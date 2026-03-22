// Port of AGG C++ gouraud.cpp - Gouraud shading demo with draggable nodes
// and control widgets for dilation, gamma, and opacity.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/gamma"
)

const (
	frameWidth  = 400
	frameHeight = 320
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

type demo struct {
	x [3]float64
	y [3]float64

	dilation *sliderctrl.SliderCtrl
	gamma    *sliderctrl.SliderCtrl
	opacity  *sliderctrl.SliderCtrl
	controls []ctrlbase.Ctrl[icol.RGBA]

	selected int
	dragDX   float64
	dragDY   float64
	height   int
}

func newDemo() *demo {
	dilation := sliderctrl.NewSliderCtrl(5, 5, frameWidth-5, 11, false)
	dilation.SetRange(0, 1)
	dilation.SetLabel("Dilation=%3.2f")
	dilation.SetValue(0.175)

	gamma := sliderctrl.NewSliderCtrl(5, 20, frameWidth-5, 26, false)
	gamma.SetRange(0, 1)
	gamma.SetLabel("Linear gamma=%3.2f")
	gamma.SetValue(0.809)

	opacity := sliderctrl.NewSliderCtrl(5, 35, frameWidth-5, 41, false)
	opacity.SetRange(0, 1)
	opacity.SetLabel("Opacity=%3.2f")
	opacity.SetValue(1.0)

	d := &demo{
		x:        [3]float64{57, 369, 143},
		y:        [3]float64{60, 170, 310},
		dilation: dilation,
		gamma:    gamma,
		opacity:  opacity,
		controls: []ctrlbase.Ctrl[icol.RGBA]{dilation, gamma, opacity},
		selected: -1,
		height:   frameHeight,
	}
	return d
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()
	d.height = h

	workBuf := make([]uint8, w*h*4)
	workImg := agg.NewImage(workBuf, w, h, w*4)
	ctx := agg.NewContextForImage(workImg)
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()
	ras := a.GetInternalRasterizer()
	ras.SetGamma(gamma.NewGammaLinear(0.0, d.gamma.Value()).Apply)

	alpha := clampU8(d.opacity.Value())

	xc := (d.x[0] + d.x[1] + d.x[2]) / 3.0
	yc := (d.y[0] + d.y[1] + d.y[2]) / 3.0
	x1 := (d.x[1]+d.x[0])*0.5 - (xc - (d.x[1]+d.x[0])*0.5)
	y1 := (d.y[1]+d.y[0])*0.5 - (yc - (d.y[1]+d.y[0])*0.5)
	x2 := (d.x[2]+d.x[1])*0.5 - (xc - (d.x[2]+d.x[1])*0.5)
	y2 := (d.y[2]+d.y[1])*0.5 - (yc - (d.y[2]+d.y[1])*0.5)
	x3 := (d.x[0]+d.x[2])*0.5 - (xc - (d.x[0]+d.x[2])*0.5)
	y3 := (d.y[0]+d.y[2])*0.5 - (yc - (d.y[0]+d.y[2])*0.5)

	cRed := agg.NewColor(255, 0, 0, alpha)
	cGreen := agg.NewColor(0, 255, 0, alpha)
	cBlue := agg.NewColor(0, 0, 255, alpha)
	cWhite := agg.NewColor(255, 255, 255, alpha)
	cBlack := agg.NewColor(0, 0, 0, alpha)

	a.GouraudTriangle(d.x[0], d.y[0], d.x[1], d.y[1], xc, yc, cRed, cGreen, cWhite, d.dilation.Value())
	a.GouraudTriangle(d.x[1], d.y[1], d.x[2], d.y[2], xc, yc, cGreen, cBlue, cWhite, d.dilation.Value())
	a.GouraudTriangle(d.x[2], d.y[2], d.x[0], d.y[0], xc, yc, cBlue, cRed, cWhite, d.dilation.Value())
	a.GouraudTriangle(d.x[0], d.y[0], d.x[1], d.y[1], x1, y1, cRed, cGreen, cBlack, d.dilation.Value())
	a.GouraudTriangle(d.x[1], d.y[1], d.x[2], d.y[2], x2, y2, cGreen, cBlue, cBlack, d.dilation.Value())
	a.GouraudTriangle(d.x[2], d.y[2], d.x[0], d.y[0], x3, y3, cBlue, cRed, cBlack, d.dilation.Value())

	for i := range d.x {
		a.FillColor(agg.NewColor(200, 50, 20, 150))
		a.NoLine()
		a.FillCircle(d.x[i], d.y[i], 8)
		a.LineColor(agg.Black)
		a.LineWidth(1.0)
		a.DrawCircle(d.x[i], d.y[i], 8)
	}

	ras.SetGamma(func(x float64) float64 { return x })
	for _, c := range d.controls {
		renderCtrl(a, c)
	}

	copyFlipY(workBuf, img.Data, w, h)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}

	fx, fy := float64(x), float64(d.height-y)
	d.selected = -1

	for _, c := range d.controls {
		if c.OnMouseButtonDown(fx, fy) {
			return true
		}
	}

	const nodeRadius2 = 100.0
	for i := 0; i < len(d.x); i++ {
		dx := fx - d.x[i]
		dy := fy - d.y[i]
		if dx*dx+dy*dy < nodeRadius2 {
			d.selected = i
			d.dragDX = fx - d.x[i]
			d.dragDY = fy - d.y[i]
			return true
		}
	}

	if pointInTriangle(d.x[0], d.y[0], d.x[1], d.y[1], d.x[2], d.y[2], fx, fy) {
		d.selected = 3
		d.dragDX = fx - d.x[0]
		d.dragDY = fy - d.y[0]
		return true
	}

	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(d.height-y)
	redraw := false

	for _, c := range d.controls {
		if c.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}

	if d.selected == -1 || !btn.Left {
		return redraw
	}

	if d.selected == 3 {
		newX := fx - d.dragDX
		newY := fy - d.dragDY
		shiftX := newX - d.x[0]
		shiftY := newY - d.y[0]
		for i := range d.x {
			d.x[i] += shiftX
			d.y[i] += shiftY
		}
		return true
	}

	d.x[d.selected] = fx - d.dragDX
	d.y[d.selected] = fy - d.dragDY
	return true
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(d.height-y)
	redraw := false

	for _, c := range d.controls {
		if c.OnMouseButtonUp(fx, fy) {
			redraw = true
		}
	}

	if d.selected != -1 {
		d.selected = -1
		redraw = true
	}

	return redraw
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
	return agg.NewColor(clampU8(c.R), clampU8(c.G), clampU8(c.B), clampU8(c.A))
}

func clampU8(v float64) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 1:
		return 255
	default:
		return uint8(v*255.0 + 0.5)
	}
}

func copyFlipY(src, dst []uint8, width, height int) {
	stride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

func pointInTriangle(x1, y1, x2, y2, x3, y3, px, py float64) bool {
	sign := func(ax, ay, bx, by, px, py float64) float64 {
		return (px-bx)*(ay-by) - (ax-bx)*(py-by)
	}
	d1 := sign(x1, y1, x2, y2, px, py)
	d2 := sign(x2, y2, x3, y3, px, py)
	d3 := sign(x3, y3, x1, y1, px, py)
	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0
	return !(hasNeg && hasPos)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Gouraud",
		Width:  frameWidth,
		Height: frameHeight,
	}, newDemo())
}
