// Based on the original AGG example: blur.cpp (flip_y = true)
// Renders an "a" glyph shape, applies blur as shadow, then draws the shape on top.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
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

const (
	frameWidth  = 440
	frameHeight = 330
	blurRadius  = 15.0
	blurMethod  = 0 // 0: Stack blur, 1: Recursive blur
)

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters (shared lowlevelrunner pattern)
// ---------------------------------------------------------------------------
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// Vertex-source adapters
// ---------------------------------------------------------------------------

// pathStorageConvVS adapts PathStorageStl to conv.VertexSource.
// PathBase exposes NextVertex() (no-arg iterator style) rather than Vertex()
// (indexed), so we wrap it here.
type pathStorageConvVS struct {
	ps *path.PathStorageStl
}

func (v *pathStorageConvVS) Rewind(pathID uint) {
	v.ps.Rewind(pathID)
}

func (v *pathStorageConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := v.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

// convCurveRasVS bridges conv.ConvCurve to the rasterizer vertex source
// interface (Rewind(uint32) / Vertex(*x,*y) uint32).
type convCurveRasVS struct {
	src *conv.ConvCurve
}

func (v *convCurveRasVS) Rewind(pathID uint32) {
	v.src.Rewind(uint(pathID))
}

func (v *convCurveRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// "a" glyph path builder
// ---------------------------------------------------------------------------

// buildGlyphPath constructs the "a" glyph using PathStorageStl with Curve3
// entries. The transformation (scale + translate) is applied via ConvTransform
// so the rasterizer sees screen coordinates.
func buildGlyphPath() *path.PathStorageStl {
	ps := path.NewPathStorageStl()

	// Outer contour
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

	// Inner counter
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
// Rendering helper
// ---------------------------------------------------------------------------

func renderGlyph(
	ras *rasType,
	sl *scanline.ScanlineP8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	mtx *transform.TransAffine,
	fillColor color.RGBA8[color.Linear],
) {
	ps := buildGlyphPath()

	// Apply affine transform (scale 4,-4 + translate 150,230 matching C++)
	xformed := conv.NewConvTransform(&pathStorageConvVS{ps: ps}, mtx)

	// Approximate quadric curves to line segments
	curved := conv.NewConvCurve(xformed)

	ras.Reset()
	ras.AddPath(&convCurveRasVS{src: curved}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, renBase, fillColor)
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer – render into this, then y-flip into img.Data (flip_y=true).
	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineP8()

	// Transform matching C++ demo: scale(4, -4), translate(150, 230).
	mtx := transform.NewTransAffineScalingXY(4.0, -4.0)
	mtx.Multiply(transform.NewTransAffineTranslation(150, 230))

	// 1. Draw shadow (dark fill).
	renderGlyph(ras, sl, renBase, mtx,
		color.RGBA8[color.Linear]{R: 25, G: 25, B: 25, A: 255})

	// 2. Apply blur to the work buffer.
	blurImageData(workBuf, w, h, blurRadius, blurMethod)

	// 3. Draw the shape on top.
	renderGlyph(ras, sl, renBase, mtx,
		color.RGBA8[color.Linear]{R: 153, G: 230, B: 179, A: 204})

	// 4. Y-flip into output buffer (flip_y=true in C++).
	copyFlipY(workBuf, img.Data, w, h)
}

// ---------------------------------------------------------------------------
// Blur helpers
// ---------------------------------------------------------------------------

func blurImageData(data []uint8, w, h int, radius float64, method int) {
	if radius <= 0 {
		return
	}

	stride := w * 4

	pixels := make([][]color.RGBA8[color.Linear], h)
	for y := 0; y < h; y++ {
		pixels[y] = make([]color.RGBA8[color.Linear], w)
		for x := 0; x < w; x++ {
			idx := y*stride + x*4
			pixels[y][x] = color.RGBA8[color.Linear]{
				R: data[idx],
				G: data[idx+1],
				B: data[idx+2],
				A: data[idx+3],
			}
		}
	}

	if method == 0 {
		sb := effects.NewSimpleStackBlur()
		sb.Blur(pixels, int(radius))
	} else {
		rb := effects.NewSimpleRecursiveBlur()
		rb.BlurHorizontal(pixels, radius)
		pixels = transposePixels(pixels)
		rb.BlurHorizontal(pixels, radius)
		pixels = transposePixels(pixels)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*stride + x*4
			pix := pixels[y][x]
			data[idx] = uint8(pix.R)
			data[idx+1] = uint8(pix.G)
			data[idx+2] = uint8(pix.B)
			data[idx+3] = uint8(pix.A)
		}
	}
}

func transposePixels(pixels [][]color.RGBA8[color.Linear]) [][]color.RGBA8[color.Linear] {
	if len(pixels) == 0 {
		return pixels
	}
	h := len(pixels)
	w := len(pixels[0])
	newPixels := make([][]color.RGBA8[color.Linear], w)
	for x := 0; x < w; x++ {
		newPixels[x] = make([]color.RGBA8[color.Linear], h)
		for y := 0; y < h; y++ {
			newPixels[x][y] = pixels[y][x]
		}
	}
	return newPixels
}

func copyFlipY(src, dst []uint8, width, height int) {
	stride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Blur",
		Width:  frameWidth,
		Height: frameHeight,
	}, &demo{})
}
