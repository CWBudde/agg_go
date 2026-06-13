// Package main is a Go port of the AGG raster_text.cpp example.
//
// It demonstrates all built-in embedded bitmap fonts by rendering a sample
// string in each font, then renders a gradient text line at the bottom using
// a sine-repeat circular gradient – matching the original C++ demo.
//
// Fidelity notes: the C++ reference renders into a linear RGBA framebuffer
// (color_type = rgba8, linear) and the platform encodes linear->sRGB at save
// time. We mirror that exactly: a linear PixFmtRGBA32 + renderer_base, the real
// span_gradient pipeline for the gradient text, and EncodeLinearRGBToSRGB on
// the runner. The gradient stop colors rgba(1,0,0)/rgba(0,0.5,0) are plain
// float->byte values (255,0,0)/(0,128,0) in linear space – they are NOT sRGB
// literals, so they are not decoded.
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/fonts"
	"github.com/cwbudde/agg_go/internal/glyph"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/renderer"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

type (
	pixfmtType = pixfmt.PixFmtRGBA32[color.Linear]
	renBase    = renderer.RendererBase[*pixfmtType, color.RGBA8[color.Linear]]

	// spanGenType is the concrete span_gradient instantiation used for the
	// gradient text: a sine-repeat circular gradient over a linear two-color
	// interpolator, driven by an identity linear interpolator.
	spanGenType = *span.SpanGradient[
		color.RGBA8[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		*gradientSineRepeatAdaptor,
		*span.GradientLinearColorRGBA8[color.Linear],
	]
)

// ---------------------------------------------------------------------------
// gradient_sine_repeat_adaptor<gradient_circle>
// ---------------------------------------------------------------------------

// gradientSineRepeatAdaptor ports the demo-local gradient_sine_repeat_adaptor
// template from raster_text.cpp, specialized on gradient_circle (identical to
// gradient_radial). It folds a radial distance into a repeating sine profile.
type gradientSineRepeatAdaptor struct {
	gradient span.GradientRadial // gradient_circle == gradient_radial
	periods  float64
}

func newGradientSineRepeatAdaptor() *gradientSineRepeatAdaptor {
	return &gradientSineRepeatAdaptor{periods: math.Pi * 2.0}
}

// SetPeriods mirrors gradient_sine_repeat_adaptor::periods.
func (g *gradientSineRepeatAdaptor) SetPeriods(p float64) {
	g.periods = p * math.Pi * 2.0
}

// Calculate matches the C++ adaptor:
//
//	int((1.0 + sin(m_gradient.calculate(x, y, d) * m_periods / d)) * d/2)
func (g *gradientSineRepeatAdaptor) Calculate(x, y, d int) int {
	r := float64(g.gradient.Calculate(x, y, d))
	return int((1.0 + math.Sin(r*g.periods/float64(d))) * float64(d) / 2.0)
}

// ---------------------------------------------------------------------------
// renderer_scanline_aa bridge for gradient raster text
// ---------------------------------------------------------------------------

// gradientTextRenderer plays the role of renderer_scanline_aa<base, alloc, sg>
// in the C++ demo: renderer_raster_htext feeds it single-span scanlines of
// glyph coverage, and it stamps gradient-generated colors into the base
// renderer. The span loop mirrors AGG's render_scanline_aa.
type gradientTextRenderer struct {
	rb    *renBase
	alloc *span.SpanAllocator[color.RGBA8[color.Linear]]
	sg    spanGenType
}

func (r *gradientTextRenderer) Prepare() { r.sg.Prepare() }

func (r *gradientTextRenderer) Render(sl renderer.ScanlineInterface) {
	y := sl.Y()
	it := sl.Begin()
	for it.HasNext() {
		s := it.Next()
		colors := r.alloc.Allocate(s.Len)
		r.sg.Generate(colors, s.X, y, s.Len)
		r.rb.BlendColorHspan(s.X, y, s.Len, colors, s.Covers, basics.CoverFull)
	}
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type fontEntry struct {
	data []byte
	name string
}

func fontList() []fontEntry {
	return []fontEntry{
		{fonts.GetGSE4x6(), "gse4x6"},
		{fonts.GetGSE4x8(), "gse4x8"},
		{fonts.GetGSE5x7(), "gse5x7"},
		{fonts.GetGSE5x9(), "gse5x9"},
		{fonts.GetGSE6x9(), "gse6x9"},
		{fonts.GetGSE6x12(), "gse6x12"},
		{fonts.GetGSE7x11(), "gse7x11"},
		{fonts.GetGSE7x11Bold(), "gse7x11_bold"},
		{fonts.GetGSE7x15(), "gse7x15"},
		{fonts.GetGSE7x15Bold(), "gse7x15_bold"},
		{fonts.GetGSE8x16(), "gse8x16"},
		{fonts.GetGSE8x16Bold(), "gse8x16_bold"},
		{fonts.GetMCS11Prop(), "mcs11_prop"},
		{fonts.GetMCS11PropCondensed(), "mcs11_prop_condensed"},
		{fonts.GetMCS12Prop(), "mcs12_prop"},
		{fonts.GetMCS13Prop(), "mcs13_prop"},
		{fonts.GetMCS5x10Mono(), "mcs5x10_mono"},
		{fonts.GetMCS5x11Mono(), "mcs5x11_mono"},
		{fonts.GetMCS6x10Mono(), "mcs6x10_mono"},
		{fonts.GetMCS6x11Mono(), "mcs6x11_mono"},
		{fonts.GetMCS7x12MonoHigh(), "mcs7x12_mono_high"},
		{fonts.GetMCS7x12MonoLow(), "mcs7x12_mono_low"},
		{fonts.GetVerdana12(), "verdana12"},
		{fonts.GetVerdana12Bold(), "verdana12_bold"},
		{fonts.GetVerdana13(), "verdana13"},
		{fonts.GetVerdana13Bold(), "verdana13_bold"},
		{fonts.GetVerdana14(), "verdana14"},
		{fonts.GetVerdana14Bold(), "verdana14_bold"},
		{fonts.GetVerdana16(), "verdana16"},
		{fonts.GetVerdana16Bold(), "verdana16_bold"},
		{fonts.GetVerdana17(), "verdana17"},
		{fonts.GetVerdana17Bold(), "verdana17_bold"},
		{fonts.GetVerdana18(), "verdana18"},
		{fonts.GetVerdana18Bold(), "verdana18_bold"},
	}
}

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	glyphGen := glyph.NewGlyphRasterBin(nil)

	// --- Solid black text in every embedded font -------------------------
	textRen := renderer.NewRendererRasterHTextSolid(rb, glyphGen)
	textRen.SetColor(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	y := 5.0
	for _, fe := range fontList() {
		if len(fe.data) == 0 {
			continue
		}
		glyphGen.SetFont(fe.data)
		text := "A quick brown fox jumps over the lazy dog 0123456789: " + fe.name
		textRen.RenderText(5, y, text, false)
		y += glyphGen.Height() + 1
	}

	// --- Gradient text via a custom span generator -----------------------
	// span_interpolator_linear with an identity matrix (mtx default-constructed
	// in C++); gradient_sine_repeat_adaptor<gradient_circle> with periods(5);
	// gradient_linear_color from rgba(1,0,0) to rgba(0,0.5,0); d1=0, d2=150.
	mtx := transform.NewTransAffine()
	interpolator := span.NewSpanInterpolatorLinearDefault(mtx)

	gradFunc := newGradientSineRepeatAdaptor()
	gradFunc.SetPeriods(5)

	colorFunc := span.NewGradientLinearColorRGBA8(
		color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}, // rgba(1.0, 0, 0)
		color.RGBA8[color.Linear]{R: 0, G: 128, B: 0, A: 255}, // rgba(0, 0.5, 0)
		256,
	)

	sg := span.NewSpanGradient(interpolator, gradFunc, colorFunc, 0, 150.0)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	gradRen := &gradientTextRenderer{rb: rb, alloc: alloc, sg: sg}

	glyphGen.SetFont(fonts.GetVerdana18Bold())
	gradTextRen := renderer.NewRendererRasterHText(gradRen, glyphGen)
	gradTextRen.RenderText(5, 465, "RADIAL REPEATING GRADIENT: A quick brown fox jumps over the lazy dog", false)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Raster Text",
		Width:                 640,
		Height:                480,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, &demo{})
}
