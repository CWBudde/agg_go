// Package main ports AGG's gradient_focal.cpp demo.
//
// The original renders into a linear BGR24 buffer: the gradient LUT is built
// and interpolated in sRGB space (color_interpolator<srgba8>) and decoded to
// linear per span pixel; blending happens in linear space and the platform
// encodes linear->sRGB when saving (EncodeLinearRGBToSRGB here).
package main

import (
	"fmt"
	"math"
	"time"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/demo/timing"
	"github.com/cwbudde/agg_go/internal/gamma"
	"github.com/cwbudde/agg_go/internal/gsv"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
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

// convRasAdapter wraps a conv.VertexSource into the rasterizer interface.
type convRasAdapter struct{ src conv.VertexSource }

func (a *convRasAdapter) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ellipseVS adapts shapes.Ellipse to conv.VertexSource.
type ellipseVS struct{ e *shapes.Ellipse }

func (s *ellipseVS) Rewind(id uint) { s.e.Rewind(uint32(id)) }
func (s *ellipseVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = s.e.Vertex(&x, &y)
	return
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

// buildGradientFocalLUT ports the C++ demo's build_gradient_lut() with
// agg::gradient_lut<color_interpolator<srgba8>, 1024>.
//
// Stop colors: rgba8_gamma_dir(srgba8(...), gamma_lut) takes rgba8 (linear),
// so each srgba8 literal is decoded to linear, gamma.dir'ed, and re-encoded
// to sRGB when stored in the srgba8 LUT profile (a non-identity roundtrip).
// build_lut() interpolates between stops in sRGB space using the generic
// color_interpolator (srgba8 has no fast specialization), i.e.
// c1.gradient(c2, count/len) with ik = uround(k*255). Finally, span_gradient
// assigns each srgba8 LUT entry to the linear color_type, an sRGB->linear
// decode, which we bake into the returned table.
func buildGradientFocalLUT(g float64, size int) []color.RGBA8[color.Linear] {
	gammaLUT := gamma.NewGammaLUT8WithGamma(g)

	stop := func(r, gc, b uint8) color.RGBA8[color.SRGB] {
		lin := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: r, G: gc, B: b, A: 255})
		lin.R = uint8(gammaLUT.Dir(basics.Int8u(lin.R)))
		lin.G = uint8(gammaLUT.Dir(basics.Int8u(lin.G)))
		lin.B = uint8(gammaLUT.Dir(basics.Int8u(lin.B)))
		return color.ConvertRGBA8LinearToSRGB(lin)
	}

	type colorPoint struct {
		offset float64
		c      color.RGBA8[color.SRGB]
	}
	profile := []colorPoint{
		{0.0, stop(0, 255, 0)},
		{0.2, stop(120, 0, 0)},
		{0.7, stop(120, 120, 0)},
		{1.0, stop(0, 0, 255)},
	}

	lut := make([]color.RGBA8[color.SRGB], size)
	start := int(basics.URound(profile[0].offset * float64(size)))
	end := start
	for i := 0; i < start; i++ {
		lut[i] = profile[0].c
	}
	for i := 1; i < len(profile); i++ {
		end = int(basics.URound(profile[i].offset * float64(size)))
		c1 := profile[i-1].c
		c2 := profile[i].c
		length := end - start + 1
		count := 0
		for start < end {
			k := basics.Int8u(basics.URound(float64(count) / float64(length) * 255.0))
			lut[start] = c1.Gradient(c2, k)
			count++
			start++
		}
	}
	for ; end < size; end++ {
		lut[end] = profile[len(profile)-1].c
	}

	out := make([]color.RGBA8[color.Linear], size)
	for i, e := range lut {
		out[i] = color.ConvertRGBA8SRGBToLinear(e)
	}
	return out
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

func drawGSVText(ras *rasType, sl *scanline.ScanlineU8, rb *renBase, x, y float64, text string) {
	txt := gsv.NewGSVText()
	txt.SetStartPoint(x, y)
	txt.SetSize(10.0, 0)
	txt.SetText(text)

	// The C++ demo strokes gsv_text with a plain conv_stroke (butt caps,
	// miter joins), not gsv_text_outline (which forces round caps).
	stroke := conv.NewConvStroke(txt)
	stroke.SetWidth(1.5)

	ras.Reset()
	ras.AddPath(&convRasAdapter{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})
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

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

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
	startTime := time.Now()
	renscan.RenderScanlinesAA(ras, sl, rb, alloc, spanGen)
	elapsedMs := float64(time.Since(startTime)) / float64(time.Millisecond)

	// Transformed circle showing the gradient boundary: ellipse with
	// auto-calculated steps, conv_stroke with default width 1.0.
	ell := shapes.NewEllipseWithParams(cx, cy, gradR, gradR, 0, false)
	ellStroke := conv.NewConvStroke(&ellipseVS{e: ell})
	ras.Reset()
	ras.AddPath(&convRasAdapter{src: ellStroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	if timing.ShowText() {
		drawGSVText(ras, sl, rb, 10.0, 35.0, fmt.Sprintf("%3.2f ms", elapsedMs))
	}

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
		Title:                 "Gradient Focal",
		Width:                 600,
		Height:                400,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
