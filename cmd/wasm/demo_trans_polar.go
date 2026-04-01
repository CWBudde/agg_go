// Based on the original AGG examples: trans_polar.cpp.
package main

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
)

// polarTransform implements AGG's trans_polar coordinate transformation.
type polarTransform struct {
	baseAngle      float64
	baseScale      float64
	baseX, baseY   float64
	transX, transY float64
	spiral         float64
}

func (p *polarTransform) Transform(x, y *float64) {
	x1 := (*x + p.baseX) * p.baseAngle
	y1 := (*y+p.baseY)*p.baseScale + (*x * p.spiral)
	*x = math.Cos(x1)*y1 + p.transX
	*y = math.Sin(x1)*y1 + p.transY
}

// Package-level slider controls for the trans_polar demo.
var (
	tpSlider1      *sliderctrl.SliderCtrl
	tpSliderSpiral *sliderctrl.SliderCtrl
	tpSliderBaseY  *sliderctrl.SliderCtrl
)

func initTransPolarSliders() {
	if tpSlider1 != nil {
		return
	}
	tpSlider1 = sliderctrl.NewSliderCtrl(10, 10, 590, 17, false)
	tpSlider1.SetRange(0.0, 100.0)
	tpSlider1.SetNumSteps(5)
	tpSlider1.SetValue(32.0)
	tpSlider1.SetLabel("Some Value=%1.0f")

	tpSliderSpiral = sliderctrl.NewSliderCtrl(10, 30, 590, 37, false)
	tpSliderSpiral.SetLabel("Spiral=%.3f")
	tpSliderSpiral.SetRange(-0.1, 0.1)
	tpSliderSpiral.SetValue(0.0)

	tpSliderBaseY = sliderctrl.NewSliderCtrl(10, 50, 590, 57, false)
	tpSliderBaseY.SetLabel("Base Y=%.3f")
	tpSliderBaseY.SetRange(50.0, 200.0)
	tpSliderBaseY.SetValue(120.0)
}

// tpSegmAdapter wraps ConvSegmentator to conv.VertexSource (Vertex returns uint32
// but the interface requires basics.PathCommand).
type tpSegmAdapter struct{ s *conv.ConvSegmentator }

func (a *tpSegmAdapter) Rewind(id uint) { a.s.Rewind(id) }
func (a *tpSegmAdapter) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.s.Vertex()
	return x, y, basics.PathCommand(cmd)
}

type (
	tpPixFmt     = pixfmt.PixFmtRGBA32[color.Linear]
	tpRenBase    = renderer.RendererBase[*tpPixFmt, color.RGBA8[color.Linear]]
	tpRasterizer = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
)

func newTPRasterizer() *tpRasterizer {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
}

// renderTPSlider renders all paths of a slider control into the scene.
func renderTPSlider(
	ras *tpRasterizer,
	sl *scanline.ScanlineU8,
	ren *tpRenBase,
	s *sliderctrl.SliderCtrl,
) {
	adapter := conv.NewRasterizerVertexSourceAdapter(s)
	for pathID := uint(0); pathID < s.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		c := s.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, ren, color.RGBA8[color.Linear]{
			R: uint8(c.R*255 + 0.5),
			G: uint8(c.G*255 + 0.5),
			B: uint8(c.B*255 + 0.5),
			A: uint8(c.A*255 + 0.5),
		})
	}
}

// renderTPSliderPolar renders slider1 paths through the polar transformation.
func renderTPSliderPolar(
	ras *tpRasterizer,
	sl *scanline.ScanlineU8,
	ren *tpRenBase,
	s *sliderctrl.SliderCtrl,
	trans *polarTransform,
) {
	segm := conv.NewConvSegmentator(s)
	pipeline := conv.NewConvTransform[conv.VertexSource, *polarTransform](&tpSegmAdapter{s: segm}, trans)
	adapter := conv.NewRasterizerVertexSourceAdapter(pipeline)
	for pathID := uint(0); pathID < s.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		c := s.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, ren, color.RGBA8[color.Linear]{
			R: uint8(c.R*255 + 0.5),
			G: uint8(c.G*255 + 0.5),
			B: uint8(c.B*255 + 0.5),
			A: uint8(c.A*255 + 0.5),
		})
	}
}

func drawTransPolarDemo() {
	initTransPolarSliders()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[*tpPixFmt, color.RGBA8[color.Linear]](pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newTPRasterizer()
	sl := scanline.NewScanlineU8()

	// Render the three slider controls.
	renderTPSlider(ras, sl, ren, tpSlider1)
	renderTPSlider(ras, sl, ren, tpSliderSpiral)
	renderTPSlider(ras, sl, ren, tpSliderBaseY)

	// Build polar transformer matching the C++ on_draw parameters.
	trans := &polarTransform{
		baseAngle: 2.0 * math.Pi / -600.0,
		baseScale: -1.0,
		baseX:     0.0,
		baseY:     tpSliderBaseY.Value(),
		transX:    float64(width) / 2.0,
		transY:    float64(height)/2.0 + 30.0,
		spiral:    -tpSliderSpiral.Value(),
	}

	// Render slider1 again, warped into polar/circular form.
	renderTPSliderPolar(ras, sl, ren, tpSlider1, trans)

	applyLinearToSRGB(img)
}

func handleTransPolarMouseDown(x, y float64) bool {
	initTransPolarSliders()
	return tpSlider1.OnMouseButtonDown(x, y) ||
		tpSliderSpiral.OnMouseButtonDown(x, y) ||
		tpSliderBaseY.OnMouseButtonDown(x, y)
}

func handleTransPolarMouseMove(x, y float64) bool {
	if tpSlider1 == nil {
		return false
	}
	return tpSlider1.OnMouseMove(x, y, true) ||
		tpSliderSpiral.OnMouseMove(x, y, true) ||
		tpSliderBaseY.OnMouseMove(x, y, true)
}

func handleTransPolarMouseUp() {
	if tpSlider1 == nil {
		return
	}
	tpSlider1.OnMouseButtonUp(0, 0)
	tpSliderSpiral.OnMouseButtonUp(0, 0)
	tpSliderBaseY.OnMouseButtonUp(0, 0)
}
