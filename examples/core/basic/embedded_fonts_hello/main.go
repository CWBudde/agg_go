// Package main demonstrates rendering text using embedded bitmap fonts.
// It uses the raster text renderer with a tiny span renderer shim
// to blend spans directly into the Context image buffer, then saves a PNG.
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/fonts"
	"agg_go/internal/glyph"
	rtext "agg_go/internal/renderer"
)

// ColorRGBA8 defines the constraint for colors that can be converted to RGBA8.
type ColorRGBA8 interface {
	agg.Color | agg.RGBA8
}

// simpleSpanRenderer is a minimal bridge that implements
// BlendSolidHspan/BlendSolidVspan over an agg.Image buffer.
// It applies straight alpha blending modulated by per-pixel coverage.
type simpleSpanRenderer[C ColorRGBA8] struct {
	img    *agg.Image
	scaleX int
	scaleY int
}

// colorToRGBA8 converts supported agg color inputs to 8-bit RGBA.
func colorToRGBA8[C ColorRGBA8](c C) (r, g, b, a uint8) {
	switch v := any(c).(type) {
	case agg.Color:
		return v.R, v.G, v.B, v.A
	case agg.RGBA8:
		return v.R, v.G, v.B, v.A
	default:
		// Fallback: opaque black
		return 0, 0, 0, 255
	}
}

func (s *simpleSpanRenderer[C]) blendPixel(x, y int, sr, sg, sb, sa, cover uint8) {
	if x < 0 || y < 0 || x >= s.img.Width || y >= s.img.Height {
		return
	}
	// Modulate alpha by coverage
	alpha := uint32(sa) * uint32(cover) / 255
	inv := 255 - alpha

	off := (y*s.img.Width + x) * 4
	d := s.img.Data
	dr, dg, db, da := uint32(d[off+0]), uint32(d[off+1]), uint32(d[off+2]), uint32(d[off+3])

	d[off+0] = uint8((uint32(sr)*alpha + dr*inv) / 255)
	d[off+1] = uint8((uint32(sg)*alpha + dg*inv) / 255)
	d[off+2] = uint8((uint32(sb)*alpha + db*inv) / 255)
	d[off+3] = uint8((alpha + da*inv) / 255)
}

func (s *simpleSpanRenderer[C]) BlendSolidHspan(x, y, length int, c C, covers []basics.CoverType) {
	sr, sg, sb, sa := colorToRGBA8(c)
	if length <= 0 {
		return
	}
	if s.scaleX <= 0 {
		s.scaleX = 1
	}
	if s.scaleY <= 0 {
		s.scaleY = 1
	}

	// Scale coordinates
	xs := x * s.scaleX
	ys := y * s.scaleY

	// For each vertical replication
	for vy := 0; vy < s.scaleY; vy++ {
		yy := ys + vy
		if yy < 0 || yy >= s.img.Height {
			continue
		}
		// Expand horizontally per source cover
		for i := 0; i < length; i++ {
			cover := uint8(255)
			if covers != nil {
				cover = uint8(covers[i])
			}
			if cover == 0 {
				continue
			}
			baseX := xs + i*s.scaleX
			for vx := 0; vx < s.scaleX; vx++ {
				s.blendPixel(baseX+vx, yy, sr, sg, sb, sa, cover)
			}
		}
	}
}

func (s *simpleSpanRenderer[C]) BlendSolidVspan(x, y, length int, c C, covers []basics.CoverType) {
	sr, sg, sb, sa := colorToRGBA8(c)
	if length <= 0 {
		return
	}
	if s.scaleX <= 0 {
		s.scaleX = 1
	}
	if s.scaleY <= 0 {
		s.scaleY = 1
	}

	xs := x * s.scaleX
	ys := y * s.scaleY

	for i := 0; i < length; i++ {
		cover := uint8(255)
		if covers != nil {
			cover = uint8(covers[i])
		}
		if cover == 0 {
			continue
		}
		baseY := ys + i*s.scaleY
		for vy := 0; vy < s.scaleY; vy++ {
			for vx := 0; vx < s.scaleX; vx++ {
				s.blendPixel(xs+vx, baseY+vy, sr, sg, sb, sa, cover)
			}
		}
	}
}

func saveAsPNG(img *agg.Image, filename string) error {
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	copy(goImg.Pix, img.Data)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, goImg)
}

func main() {
	const W, H = 640, 220
	ctx := agg.NewContext(W, H)
	// Background
	ctx.Clear(agg.RGB(0.97, 0.97, 1.0))

	// Prepare text rendering pipeline
	// Use a compact readable embedded font (5x7 or 4x6)
	// Choose different fonts to demonstrate variety
	fontSmall := fonts.GetGSE4x6()
	fontMedium := fonts.GetGSE5x7()
	fontMono := fonts.GetMCS5x10Mono()
	fontVerdana := fonts.GetVerdana12()

	gSmall := glyph.NewGlyphRasterBin(fontSmall)
	gMedium := glyph.NewGlyphRasterBin(fontMedium)
	gMono := glyph.NewGlyphRasterBin(fontMono)
	gVerdana := glyph.NewGlyphRasterBin(fontVerdana)

	// Bridge spans into the context image
	// Unscaled renderer (1x)
	ren1x := &simpleSpanRenderer[agg.RGBA8]{img: ctx.GetImage(), scaleX: 1, scaleY: 1}
	text1x := rtext.NewRendererRasterHTextSolid[*simpleSpanRenderer[agg.RGBA8], *glyph.GlyphRasterBin](ren1x, gSmall)

	// Scaled renderer (2x)
	ren2x := &simpleSpanRenderer[agg.RGBA8]{img: ctx.GetImage(), scaleX: 2, scaleY: 2}
	text2x := rtext.NewRendererRasterHTextSolid[*simpleSpanRenderer[agg.RGBA8], *glyph.GlyphRasterBin](ren2x, gMedium)

	// Scaled renderer (3x) for larger display
	ren3x := &simpleSpanRenderer[agg.RGBA8]{img: ctx.GetImage(), scaleX: 3, scaleY: 3}
	text3x := rtext.NewRendererRasterHTextSolid[*simpleSpanRenderer[agg.RGBA8], *glyph.GlyphRasterBin](ren3x, gVerdana)

	// Draw a baseline guide (optional)
	ctx.SetColor(agg.RGBA(0.85, 0.9, 1.0, 1))
	ctx.DrawLine(20, 70, float64(W-20), 70)

	// Render several lines showing fonts and sizes
	text1x.SetColor(agg.NewRGBA8(20, 30, 40, 255))
	text1x.RenderText(20, 60, "GSE4x6 @1x: Hello World!", false)

	text2x.SetColor(agg.NewRGBA8(200, 60, 40, 255))
	text2x.RenderText(20, 100, "GSE5x7 @2x: Embedded Fonts", false)

	// Monospace at 2x as well
	ren2x.scaleX, ren2x.scaleY = 2, 2
	text2x.Attach(ren2x)
	text2x = rtext.NewRendererRasterHTextSolid[*simpleSpanRenderer[agg.RGBA8], *glyph.GlyphRasterBin](ren2x, gMono)
	text2x.SetColor(agg.NewRGBA8(40, 120, 200, 255))
	text2x.RenderText(20, 135, "MCS5x10Mono @2x", false)

	// Verdana12 at 3x
	text3x.SetColor(agg.NewRGBA8(30, 80, 60, 255))
	text3x.RenderText(20, 185, "Verdana12 @3x", false)

	// Save PNG (when using `just run-examples-basic`, this lands in examples/basic/_out)
	out := "embedded_fonts_hello.png"
	if err := saveAsPNG(ctx.GetImage(), out); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}
	fmt.Printf("Wrote %s (%dx%d)\n", out, W, H)
}
