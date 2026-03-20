// Package main is a Go port of the AGG raster_text.cpp example.
// It demonstrates all built-in embedded bitmap fonts by rendering a sample
// string in each font, then renders a gradient text line at the bottom using
// a sine-repeat circular gradient – matching the original C++ demo.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/fonts"
	"github.com/MeKo-Christian/agg_go/internal/glyph"
	rtext "github.com/MeKo-Christian/agg_go/internal/renderer"
)

// ---------------------------------------------------------------------------
// Solid-color span renderer
// ---------------------------------------------------------------------------

// spanRenderer bridges RendererRasterHTextSolid into an agg.Image buffer.
type spanRenderer struct {
	img *agg.Image
}

func (s *spanRenderer) blendPixel(x, y int, sr, sg, sb, sa, cover uint8) {
	h := s.img.Height()
	fy := h - 1 - y
	if x < 0 || fy < 0 || x >= s.img.Width() || fy >= h {
		return
	}
	alpha := uint32(sa) * uint32(cover) / 255
	inv := 255 - alpha
	off := (fy*s.img.Width() + x) * 4
	d := s.img.Data
	d[off+0] = uint8((uint32(sr)*alpha + uint32(d[off+0])*inv) / 255)
	d[off+1] = uint8((uint32(sg)*alpha + uint32(d[off+1])*inv) / 255)
	d[off+2] = uint8((uint32(sb)*alpha + uint32(d[off+2])*inv) / 255)
	d[off+3] = uint8(alpha + uint32(d[off+3])*inv/255)
}

func (s *spanRenderer) BlendSolidHspan(x, y, length int, c agg.RGBA8, covers []basics.CoverType) {
	for i := 0; i < length; i++ {
		cover := uint8(255)
		if covers != nil && i < len(covers) {
			cover = uint8(covers[i])
		}
		s.blendPixel(x+i, y, c.R, c.G, c.B, c.A, cover)
	}
}

func (s *spanRenderer) BlendSolidVspan(x, y, length int, c agg.RGBA8, covers []basics.CoverType) {
	for i := 0; i < length; i++ {
		cover := uint8(255)
		if covers != nil && i < len(covers) {
			cover = uint8(covers[i])
		}
		s.blendPixel(x, y+i, c.R, c.G, c.B, c.A, cover)
	}
}

// ---------------------------------------------------------------------------
// Gradient span renderer
// ---------------------------------------------------------------------------

type gradientRenderer struct {
	img     *agg.Image
	periods float64
	d       float64
}

func newGradientRenderer(img *agg.Image) *gradientRenderer {
	return &gradientRenderer{img: img, periods: 5.0 * math.Pi * 2.0, d: 150.0}
}

func (g *gradientRenderer) Prepare() {}

func (g *gradientRenderer) gradientColor(x, y int) (r, gr, b uint8) {
	dist := math.Sqrt(float64(x*x + y*y))
	t := (1.0 + math.Sin(dist*g.periods/g.d)) * 0.5
	rf := 1.0 - t
	gf := 0.5 * t
	return uint8(rf * 255), uint8(gf * 255), 0
}

func (g *gradientRenderer) blendPixel(x, y int, cover uint8) {
	h := g.img.Height()
	fy := h - 1 - y
	if x < 0 || fy < 0 || x >= g.img.Width() || fy >= h {
		return
	}
	cr, cg, _ := g.gradientColor(x, y)
	alpha := uint32(cover)
	inv := 255 - alpha
	off := (fy*g.img.Width() + x) * 4
	d := g.img.Data
	d[off+0] = uint8((uint32(cr)*alpha + uint32(d[off+0])*inv) / 255)
	d[off+1] = uint8((uint32(cg)*alpha + uint32(d[off+1])*inv) / 255)
	d[off+2] = uint8((0*alpha + uint32(d[off+2])*inv) / 255)
	d[off+3] = uint8(alpha + uint32(d[off+3])*inv/255)
}

func (g *gradientRenderer) Render(sl rtext.ScanlineInterface) {
	y := sl.Y()
	it := sl.Begin()
	for it.HasNext() {
		span := it.Next()
		for i := 0; i < span.Len; i++ {
			cover := uint8(255)
			if span.Covers != nil && i < len(span.Covers) {
				cover = uint8(span.Covers[i])
			}
			g.blendPixel(span.X+i, y, cover)
		}
	}
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.RGB(1, 1, 1))

	type fontEntry struct {
		data []byte
		name string
	}
	fontList := []fontEntry{
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

	ren := &spanRenderer{img: img}
	g := glyph.NewGlyphRasterBin(fontList[0].data)
	textRen := rtext.NewRendererRasterHTextSolid[*spanRenderer, *glyph.GlyphRasterBin, agg.RGBA8](ren, g)
	textRen.SetColor(agg.NewRGBA8(0, 0, 0, 255))

	y := 5.0
	for _, fe := range fontList {
		if len(fe.data) == 0 {
			continue
		}
		g.SetFont(fe.data)
		text := "A quick brown fox jumps over the lazy dog 0123456789: " + fe.name
		textRen.RenderText(5, y, text, false)
		y += g.Height() + 1
	}

	gradRen := newGradientRenderer(img)
	g.SetFont(fonts.GetVerdana12())
	gradTextRen := rtext.NewRendererRasterHText[*gradientRenderer, *glyph.GlyphRasterBin](gradRen, g)
	gradTextRen.RenderText(5, 465, "RADIAL REPEATING GRADIENT: A quick brown fox jumps over the lazy dog", false)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Raster Text",
		Width:  640,
		Height: 480,
	}, &demo{})
}
