// Port of AGG C++ lion_lens.cpp - lion with warp-magnifier lens effect.
//
// Renders the lion vector art with the original AGG control widgets and the
// warp-magnifier lens. The lens starts at the C++ default position
// (200, 150), while the sliders control magnification and radius.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	llWidth  = 500
	llHeight = 600

	// C++ defaults.
	defaultLensScale  = 3.0
	defaultLensRadius = 70.0
	defaultLensX      = 200.0
	defaultLensY      = 150.0
)

type demo struct {
	lion         *liondemo.LionData
	baseDX       float64
	baseDY       float64
	lensX        float64
	lensY        float64
	lightX       float64
	lightY       float64
	magnSlider   *sliderctrl.SliderCtrl
	radiusSlider *sliderctrl.SliderCtrl
}

func newDemo() *demo {
	magnSlider := sliderctrl.NewSliderCtrl(5, 5, 495, 12, false)
	magnSlider.SetRange(0.01, 4.0)
	magnSlider.SetValue(defaultLensScale)
	magnSlider.SetLabel("Scale=%3.2f")

	radiusSlider := sliderctrl.NewSliderCtrl(5, 20, 495, 27, false)
	radiusSlider.SetRange(0.0, 100.0)
	radiusSlider.SetValue(defaultLensRadius)
	radiusSlider.SetLabel("Radius=%3.2f")

	return &demo{
		lensX:        defaultLensX,
		lensY:        defaultLensY,
		lightX:       defaultLensX,
		lightY:       defaultLensY,
		magnSlider:   magnSlider,
		radiusSlider: radiusSlider,
	}
}

func (d *demo) OnInit() {
	d.ensureLion()
	d.lensX = defaultLensX
	d.lensY = defaultLensY
	d.lightX = defaultLensX
	d.lightY = defaultLensY
}

func (d *demo) Render(img *agg.Image) {
	d.ensureLion()

	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	lens := transform.NewTransWarpMagnifier()
	lens.SetCenter(d.lensX, d.lensY)
	lens.SetMagnification(d.magnSlider.Value())
	lens.SetRadius(d.radiusSlider.Value() / d.magnSlider.Value())

	mtx := transform.NewTransAffine()
	mtx.Translate(-d.baseDX, -d.baseDY)
	mtx.Rotate(math.Pi)
	mtx.Translate(float64(llWidth)*0.5, float64(llHeight)*0.5)

	for i := 0; i < d.lion.NPaths; i++ {
		a.FillColor(agg.NewColor(
			d.lion.Colors[i].R,
			d.lion.Colors[i].G,
			d.lion.Colors[i].B,
			255,
		))
		a.NoLine()
		a.ResetPath()

		d.lion.Path.Rewind(d.lion.PathIdx[i])
		for {
			x, y, cmd := d.lion.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			mtx.Transform(&x, &y)
			lens.Transform(&x, &y)

			switch {
			case basics.IsMoveTo(basics.PathCommand(cmd)):
				a.MoveTo(x, y)
			case basics.IsLineTo(basics.PathCommand(cmd)):
				a.LineTo(x, y)
			}
		}

		a.ClosePolygon()
		a.DrawPath(agg.FillOnly)
	}

	renderCtrl(a, d.magnSlider)
	renderCtrl(a, d.radiusSlider)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	if btn.Left {
		if d.magnSlider.OnMouseButtonDown(fx, fy) {
			return true
		}
		if d.radiusSlider.OnMouseButtonDown(fx, fy) {
			return true
		}

		d.lensX = fx
		d.lensY = fy
		return true
	}

	if btn.Right {
		d.lightX = fx
		d.lightY = fy
		return true
	}

	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	redraw := false
	if d.magnSlider.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}
	if d.radiusSlider.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}

	if btn.Left && !d.magnSlider.InRect(fx, fy) && !d.radiusSlider.InRect(fx, fy) {
		d.lensX = fx
		d.lensY = fy
		redraw = true
	}

	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	redraw := false
	if d.magnSlider.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	if d.radiusSlider.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	return redraw
}

func (d *demo) ensureLion() {
	if d.lion != nil {
		return
	}

	ld := liondemo.Parse()
	d.lion = &ld

	bx1, by1, bx2, by2 := 1e9, 1e9, -1e9, -1e9
	for idx := uint(0); idx < d.lion.Path.TotalVertices(); idx++ {
		x, y, cmd := d.lion.Path.Vertex(idx)
		if !basics.IsVertex(basics.PathCommand(cmd)) {
			continue
		}
		if x < bx1 {
			bx1 = x
		}
		if y < by1 {
			by1 = y
		}
		if x > bx2 {
			bx2 = x
		}
		if y > by2 {
			by2 = y
		}
	}

	d.baseDX = (bx2 - bx1) * 0.5
	d.baseDY = (by2 - by1) * 0.5
}

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
}

type ctrlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }

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

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Lion Lens",
		Width:  llWidth,
		Height: llHeight,
		FlipY:  true,
	}, newDemo())
}
