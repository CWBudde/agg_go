// Package main ports AGG's gradient_focal.cpp demo.
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/gamma"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

const gradR = 100.0

type demo struct {
	gammaSlider *sliderctrl.SliderCtrl
	mouseX      float64
	mouseY      float64
	mouseValid  bool
	lastGamma   float64
	gradientLUT []color.RGBA8[color.Linear]
}

type (
	rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	renBase = renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]
)

type rasterVertexSourceAdapter struct {
	src ctrlbase.Ctrl[color.RGBA]
}

func (a *rasterVertexSourceAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *rasterVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

func newDemo() *demo {
	gammaSlider := sliderctrl.NewSliderCtrl(5.0, 5.0, 340.0, 12.0, false)
	gammaSlider.SetRange(0.5, 2.5)
	gammaSlider.SetValue(1.0)
	gammaSlider.SetLabel("Gamma = %.3f")

	return &demo{
		gammaSlider: gammaSlider,
		lastGamma:   math.NaN(),
	}
}

func buildGradientFocalLUT(g float64, size int) []color.RGBA8[color.Linear] {
	if size < 2 {
		size = 2
	}
	lut := make([]color.RGBA8[color.Linear], size)
	gammaLUT := gamma.NewGammaLUT8WithGamma(g)

	type stop struct {
		pos float64
		r   uint8
		g   uint8
		b   uint8
	}
	stops := []stop{
		{pos: 0.0, r: 0, g: 255, b: 0},
		{pos: 0.2, r: 120, g: 0, b: 0},
		{pos: 0.7, r: 120, g: 120, b: 0},
		{pos: 1.0, r: 0, g: 0, b: 255},
	}

	type stopGamma struct {
		pos float64
		r   float64
		g   float64
		b   float64
	}
	sg := make([]stopGamma, len(stops))
	for i, s := range stops {
		sg[i] = stopGamma{
			pos: s.pos,
			r:   float64(gammaLUT.Dir(basics.Int8u(s.r))),
			g:   float64(gammaLUT.Dir(basics.Int8u(s.g))),
			b:   float64(gammaLUT.Dir(basics.Int8u(s.b))),
		}
	}

	for i := 0; i < size; i++ {
		t := float64(i) / float64(size-1)
		j := 0
		for j < len(sg)-2 && t > sg[j+1].pos {
			j++
		}
		a := sg[j]
		b := sg[j+1]
		u := 0.0
		if den := b.pos - a.pos; den > 0 {
			u = (t - a.pos) / den
		}
		if u < 0 {
			u = 0
		}
		if u > 1 {
			u = 1
		}
		lut[i] = color.RGBA8[color.Linear]{
			R: uint8(a.r + (b.r-a.r)*u + 0.5),
			G: uint8(a.g + (b.g-a.g)*u + 0.5),
			B: uint8(a.b + (b.b-a.b)*u + 0.5),
			A: 255,
		}
	}

	return lut
}

func applyGammaInv(img *agg.Image, g float64) {
	if math.Abs(g-1.0) < 1e-9 {
		return
	}
	lut := gamma.NewGammaLUT8WithGamma(g)
	for i := 0; i+3 < len(img.Data); i += 4 {
		img.Data[i+0] = uint8(lut.Inv(basics.Int8u(img.Data[i+0])))
		img.Data[i+1] = uint8(lut.Inv(basics.Int8u(img.Data[i+1])))
		img.Data[i+2] = uint8(lut.Inv(basics.Int8u(img.Data[i+2])))
	}
}

func renderCtrl(ras *rasType, sl *scanline.ScanlineU8, rb *renBase, c ctrlbase.Ctrl[color.RGBA]) {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255.0 + 0.5)
	}
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(&rasterVertexSourceAdapter{src: c}, uint32(i))
		col := c.Color(i)
		renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{
			R: clamp(col.R),
			G: clamp(col.G),
			B: clamp(col.B),
			A: clamp(col.A),
		})
	}
}

func (d *demo) Render(img *agg.Image) {
	w := img.Width()
	h := img.Height()
	cx := float64(w) * 0.5
	cy := float64(h) * 0.5

	if !d.mouseValid {
		d.mouseX = cx
		d.mouseY = cy
		d.mouseValid = true
	}

	gammaVal := d.gammaSlider.Value()
	if d.gradientLUT == nil || math.Abs(gammaVal-d.lastGamma) > 1e-9 {
		d.gradientLUT = buildGradientFocalLUT(gammaVal, 1024)
		d.lastGamma = gammaVal
	}

	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)

	ras := newRasterizer()
	sl := scanline.NewScanlineU8()
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	gradientMtx := transform.NewTransAffine()
	gradientMtx.Translate(cx, cy)
	gradientMtx.Invert()

	fx := d.mouseX - cx
	fy := d.mouseY - cy

	interpolator := span.NewSpanInterpolatorLinearDefault(gradientMtx)
	gradientFunc := span.NewGradientRadialFocus(gradR, fx, fy)
	gradientReflect := span.NewGradientReflectAdaptor(gradientFunc)
	colorFn := span.NewGradientPrebuiltColorRGBA8[color.Linear](d.gradientLUT)
	spanGen := span.NewSpanGradient(interpolator, gradientReflect, colorFn, 0, gradR)

	ras.MoveToD(0, 0)
	ras.LineToD(float64(w), 0)
	ras.LineToD(float64(w), float64(h))
	ras.LineToD(0, float64(h))
	ras.ClosePolygon()
	renscan.RenderScanlinesAA(ras, sl, rb, alloc, spanGen)

	ctx.SetColor(agg.White)
	ctx.SetLineWidth(1.0)
	ctx.DrawCircle(cx, cy, gradR)

	renderCtrl(ras, sl, rb, d.gammaSlider)
	applyGammaInv(img, gammaVal)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	if d.gammaSlider.OnMouseButtonDown(fx, fy) {
		return true
	}
	d.mouseX = fx
	d.mouseY = fy
	d.mouseValid = true
	return true
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.gammaSlider.OnMouseMove(fx, fy, btn.Left) {
		return true
	}
	if !btn.Left {
		return false
	}
	d.mouseX = fx
	d.mouseY = fy
	d.mouseValid = true
	return true
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	_ = btn
	return d.gammaSlider.OnMouseButtonUp(float64(x), float64(y))
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Gradient Focal",
		Width:  600,
		Height: 400,
		FlipY:  true,
	}, newDemo())
}
