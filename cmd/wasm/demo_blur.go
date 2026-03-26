// Based on the original AGG examples: blur.cpp.
//
// Renders an "a" glyph shape, applies stack/recursive blur as a shadow effect,
// then draws the shape on top.
//
// State variables blurRadius and blurMethod are preserved and exposed to JS.
//go:build js && wasm

package main

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/effects"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	blurRadius = 15.0
	blurMethod = 0 // 0: Stack blur, 1: Recursive blur
)

// ---------------------------------------------------------------------------
// Path building
// ---------------------------------------------------------------------------

func buildBlurGlyphPath() *path.PathStorageStl {
	ps := path.NewPathStorageStl()

	ps.MoveTo(28.47, 6.45)
	ps.Curve3(21.58, 1.12, 19.82, 0.29)
	ps.Curve3(17.19, -0.93, 14.21, -0.93)
	ps.Curve3(9.57, -0.93, 6.57, 2.25)
	ps.Curve3(3.56, 5.42, 3.56, 10.60)
	ps.Curve3(3.56, 13.87, 5.03, 16.26)
	ps.Curve3(7.03, 19.58, 11.99, 22.51)
	ps.Curve3(16.94, 25.44, 28.47, 29.64)
	ps.LineTo(28.47, 31.40)
	ps.Curve3(28.47, 38.09, 26.34, 40.58)
	ps.Curve3(24.22, 43.07, 20.17, 43.07)
	ps.Curve3(17.09, 43.07, 15.28, 41.41)
	ps.Curve3(13.43, 39.75, 13.43, 37.60)
	ps.LineTo(13.53, 34.77)
	ps.Curve3(13.53, 32.52, 12.38, 31.30)
	ps.Curve3(11.23, 30.08, 9.38, 30.08)
	ps.Curve3(7.57, 30.08, 6.42, 31.35)
	ps.Curve3(5.27, 32.62, 5.27, 34.81)
	ps.Curve3(5.27, 39.01, 9.57, 42.53)
	ps.Curve3(13.87, 46.04, 21.63, 46.04)
	ps.Curve3(27.59, 46.04, 31.40, 44.04)
	ps.Curve3(34.28, 42.53, 35.64, 39.31)
	ps.Curve3(36.52, 37.21, 36.52, 30.71)
	ps.LineTo(36.52, 15.53)
	ps.Curve3(36.52, 9.13, 36.77, 7.69)
	ps.Curve3(37.01, 6.25, 37.57, 5.76)
	ps.Curve3(38.13, 5.27, 38.87, 5.27)
	ps.Curve3(39.65, 5.27, 40.23, 5.62)
	ps.Curve3(41.26, 6.25, 44.19, 9.18)
	ps.LineTo(44.19, 6.45)
	ps.Curve3(38.72, -0.88, 33.74, -0.88)
	ps.Curve3(31.35, -0.88, 29.93, 0.78)
	ps.Curve3(28.52, 2.44, 28.47, 6.45)
	ps.ClosePolygon(basics.PathFlagsCW)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(basics.PathFlagsCW)

	return ps
}

// ---------------------------------------------------------------------------
// Vertex-source adapters
// ---------------------------------------------------------------------------

type blurPathStlVS struct{ ps *path.PathStorageStl }

func (v *blurPathStlVS) Rewind(id uint) { v.ps.Rewind(id) }
func (v *blurPathStlVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := v.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

type blurConvSourceVS struct{ src conv.VertexSource }

func (v *blurConvSourceVS) Rewind(id uint32) { v.src.Rewind(uint(id)) }
func (v *blurConvSourceVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Rasterizer type
// ---------------------------------------------------------------------------

type blurRasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newBlurRasterizer() *blurRasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// Blur helpers
// ---------------------------------------------------------------------------

func blurClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// blurRegion blurs the rectangle [x0,y0,x1,y1) in buf (row-major, RGBA).
func blurRegion(buf []uint8, w, h int, x0, y0, x1, y1 int, radius float64, method int) {
	if radius <= 0 {
		return
	}
	x0 = blurClampInt(x0, 0, w)
	y0 = blurClampInt(y0, 0, h)
	x1 = blurClampInt(x1, 0, w)
	y1 = blurClampInt(y1, 0, h)
	if x0 >= x1 || y0 >= y1 {
		return
	}

	stride := w * 4
	rw := x1 - x0
	rh := y1 - y0

	pixels := make([][]color.RGBA8[color.Linear], rh)
	for row := range rh {
		pixels[row] = make([]color.RGBA8[color.Linear], rw)
		for col := range rw {
			idx := (y0+row)*stride + (x0+col)*4
			pixels[row][col] = color.RGBA8[color.Linear]{
				R: buf[idx],
				G: buf[idx+1],
				B: buf[idx+2],
				A: buf[idx+3],
			}
		}
	}

	r := int(radius)
	if method == 0 {
		sb := effects.NewSimpleStackBlur()
		sb.Blur(pixels, r)
	} else {
		rb := effects.NewSimpleRecursiveBlur()
		rb.BlurHorizontal(pixels, radius)
		pixels = blurTranspose(pixels)
		rb.BlurHorizontal(pixels, radius)
		pixels = blurTranspose(pixels)
	}

	for row := range rh {
		for col := range rw {
			idx := (y0+row)*stride + (x0+col)*4
			pix := pixels[row][col]
			buf[idx] = pix.R
			buf[idx+1] = pix.G
			buf[idx+2] = pix.B
			buf[idx+3] = pix.A
		}
	}
}

func blurTranspose(pixels [][]color.RGBA8[color.Linear]) [][]color.RGBA8[color.Linear] {
	if len(pixels) == 0 {
		return pixels
	}
	h := len(pixels)
	w := len(pixels[0])
	out := make([][]color.RGBA8[color.Linear], w)
	for col := range w {
		out[col] = make([]color.RGBA8[color.Linear], h)
		for row := range h {
			out[col][row] = pixels[row][col]
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Main render function
// ---------------------------------------------------------------------------

func drawBlurDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newBlurRasterizer()
	sl := scanline.NewScanlineP8()

	// Shape transform: scale(4, -4) + translate to canvas centre.
	// The original Agg2D demo used Scale(4, -4) + Translate(150, 400)
	// which maps the glyph (y-up) into y-down screen space.
	shapeMtx := transform.NewTransAffineScalingXY(4.0, -4.0)
	shapeMtx.Multiply(transform.NewTransAffineTranslation(150, 400))

	renderShape := func(fillColor color.RGBA8[color.Linear]) {
		ps := buildBlurGlyphPath()
		xformed := conv.NewConvTransform(&blurPathStlVS{ps: ps}, shapeMtx)
		curved := conv.NewConvCurve(xformed)
		ras.Reset()
		ras.AddPath(&blurConvSourceVS{src: curved}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, fillColor)
	}

	// 1. Render shadow shape.
	renderShape(color.RGBA8[color.Linear]{R: 25, G: 25, B: 25, A: 255})

	// 2. Blur the entire canvas (simple full-canvas blur as in the original wasm demo).
	blurRegion(img.Data, w, h, 0, 0, w, h, blurRadius, blurMethod)

	// 3. Draw the shape itself on top.
	renderShape(color.RGBA8[color.Linear]{R: 153, G: 230, B: 179, A: 204})

	// 4. Apply linear→sRGB encoding.
	applyLinearToSRGB(img)
}
