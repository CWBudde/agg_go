// Package agg2d adapters for AGG2D high-level interface.
// This file contains adapter types that bridge different interfaces.
package agg2d

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
)

// baseRendererAdapter adapts renderer base functionality.
// It caches a RendererBase instance, matching the C++ design where
// renderer_base is stored as a value member of Agg2D and reused across
// all rendering calls. The cached instance is rebuilt only when the
// pixel format or clip box changes.
type baseRendererAdapter[C any] struct {
	pf  renderer.PixelFormat[C]
	ren *renderer.RendererBase[renderer.PixelFormat[C], C]
}

func newBaseRendererAdapter[C any](pf renderer.PixelFormat[C]) *baseRendererAdapter[C] {
	b := &baseRendererAdapter[C]{pf: pf}
	b.ren = renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[C], C](pf)
	return b
}

func (b *baseRendererAdapter[C]) rendererBase() *renderer.RendererBase[renderer.PixelFormat[C], C] {
	return b.ren
}

func (b *baseRendererAdapter[C]) ClipBox(x1, y1, x2, y2 int) bool {
	return b.ren.ClipBox(x1, y1, x2, y2)
}

func (b *baseRendererAdapter[C]) Width() int  { return b.pf.Width() }
func (b *baseRendererAdapter[C]) Height() int { return b.pf.Height() }

func (b *baseRendererAdapter[C]) Clear(c C) {
	b.rendererBase().Clear(c)
}

func (b *baseRendererAdapter[C]) CopyPixel(x, y int, c C) {
	b.rendererBase().CopyPixel(x, y, c)
}

func (b *baseRendererAdapter[C]) BlendPixel(x, y int, c C, cover basics.Int8u) {
	b.rendererBase().BlendPixel(x, y, c, cover)
}

func (b *baseRendererAdapter[C]) BlendHline(x, y, x2 int, c C, cover basics.Int8u) {
	b.rendererBase().BlendHline(x, y, x2, c, cover)
}

func (b *baseRendererAdapter[C]) BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	b.rendererBase().BlendColorHspan(x, y, length, colors, covers, cover)
}

func (b *baseRendererAdapter[C]) BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u) {
	b.rendererBase().BlendSolidHspan(x, y, length, c, covers)
}

// BlendFrom blends from another pixel format using the rendering pipeline
func (b *baseRendererAdapter[C]) BlendFrom(src renderer.PixelFormat[C], rectSrcPtr *basics.RectI, dx, dy int, cover basics.Int8u) {
	b.rendererBase().BlendFrom(src, rectSrcPtr, dx, dy, cover)
}

// CopyFrom copies from another pixel format using the rendering pipeline
func (b *baseRendererAdapter[C]) CopyFrom(src renderer.PixelFormat[C], rectSrcPtr *basics.RectI, dx, dy int) {
	b.rendererBase().CopyFrom(src, rectSrcPtr, dx, dy)
}

// imagePixelFormat exposes an Image through the source-side accessors used by
// AGG-style span filters and pixfmt row-copy helpers.
type imagePixelFormat struct {
	img      *Image
	x, y, x0 int
	pixelBuf [4]basics.Int8u
}

func newImagePixelFormat(img *Image) *imagePixelFormat {
	return &imagePixelFormat{img: img}
}

type imagePixelFormatPre struct {
	img    *Image
	rowY   int
	rowBuf []basics.Int8u
}

func newImagePixelFormatPre(img *Image) *imagePixelFormatPre {
	return &imagePixelFormatPre{img: img, rowY: -1}
}

func (ipf *imagePixelFormat) Width() int {
	return ipf.img.Width()
}

func (ipf *imagePixelFormat) Height() int {
	return ipf.img.Height()
}

// ColorType returns the RGBA color type identifier used by span image filters.
func (ipf *imagePixelFormat) ColorType() string {
	return "RGBA8"
}

// OrderType returns the channel layout for source spans.
func (ipf *imagePixelFormat) OrderType() color.ColorOrder {
	return color.OrderRGBA
}

// Pixel returns a pixel as RGBA8 color
func (ipf *imagePixelFormat) Pixel(x, y int) color.RGBA8[color.Linear] {
	pixel := ipf.img.GetPixel(x, y)
	return color.NewRGBA8[color.Linear](pixel[0], pixel[1], pixel[2], pixel[3])
}

func (ipf *imagePixelFormatPre) Width() int {
	return ipf.img.Width()
}

func (ipf *imagePixelFormatPre) Height() int {
	return ipf.img.Height()
}

func (ipf *imagePixelFormatPre) Pixel(x, y int) color.RGBA8[color.Linear] {
	pixel := ipf.img.GetPixel(x, y)
	a := pixel[3]
	return color.NewRGBA8[color.Linear](
		color.RGBA8Multiply(pixel[0], a),
		color.RGBA8Multiply(pixel[1], a),
		color.RGBA8Multiply(pixel[2], a),
		a,
	)
}

func (ipf *imagePixelFormatPre) GetPixel(x, y int) color.RGBA8[color.Linear] {
	return ipf.Pixel(x, y)
}

func (ipf *imagePixelFormatPre) RowData(y int) []basics.Int8u {
	if ipf.img == nil || ipf.img.Data == nil || y < 0 || y >= ipf.img.height {
		return nil
	}
	if ipf.rowY == y && len(ipf.rowBuf) == ipf.img.width*4 {
		return ipf.rowBuf
	}

	src := ipf.img.PixelFormat().RowData(y)
	if src == nil {
		return nil
	}

	rowLen := ipf.img.width * 4
	if cap(ipf.rowBuf) < rowLen {
		ipf.rowBuf = make([]basics.Int8u, rowLen)
	} else {
		ipf.rowBuf = ipf.rowBuf[:rowLen]
	}

	for x := 0; x < ipf.img.width; x++ {
		off := x * 4
		a := src[off+3]
		ipf.rowBuf[off+0] = color.RGBA8Multiply(src[off+0], a)
		ipf.rowBuf[off+1] = color.RGBA8Multiply(src[off+1], a)
		ipf.rowBuf[off+2] = color.RGBA8Multiply(src[off+2], a)
		ipf.rowBuf[off+3] = a
	}
	ipf.rowY = y
	return ipf.rowBuf
}

func (ipf *imagePixelFormat) pixelSliceClamped(x, y int) []basics.Int8u {
	if ipf.img == nil || ipf.img.Data == nil || ipf.img.width <= 0 || ipf.img.height <= 0 {
		ipf.pixelBuf = [4]basics.Int8u{0, 0, 0, 0}
		return ipf.pixelBuf[:]
	}

	if x < 0 {
		x = 0
	} else if x >= ipf.img.width {
		x = ipf.img.width - 1
	}
	if y < 0 {
		y = 0
	} else if y >= ipf.img.height {
		y = ipf.img.height - 1
	}

	stride := ipf.img.Stride()
	offset := y*stride + x*4
	if offset+3 >= len(ipf.img.Data) {
		ipf.pixelBuf = [4]basics.Int8u{0, 0, 0, 0}
		return ipf.pixelBuf[:]
	}

	ipf.pixelBuf = [4]basics.Int8u{
		ipf.img.Data[offset],
		ipf.img.Data[offset+1],
		ipf.img.Data[offset+2],
		ipf.img.Data[offset+3],
	}
	return ipf.pixelBuf[:]
}

// Span initializes source sampling at (x,y) and returns the first pixel.
func (ipf *imagePixelFormat) Span(x, y, length int) []basics.Int8u {
	_ = length
	ipf.x = x
	ipf.x0 = x
	ipf.y = y
	return ipf.pixelSliceClamped(x, y)
}

// NextX advances sampling by one pixel in x direction.
func (ipf *imagePixelFormat) NextX() []basics.Int8u {
	ipf.x++
	return ipf.pixelSliceClamped(ipf.x, ipf.y)
}

// NextY advances sampling by one row at the original x position.
func (ipf *imagePixelFormat) NextY() []basics.Int8u {
	ipf.y++
	ipf.x = ipf.x0
	return ipf.pixelSliceClamped(ipf.x, ipf.y)
}

// RowPtr returns row bytes for scanline-based image filters.
func (ipf *imagePixelFormat) RowPtr(y int) []basics.Int8u {
	if ipf.img == nil || ipf.img.Data == nil || y < 0 || y >= ipf.img.height {
		return nil
	}

	stride := ipf.img.Stride()
	rowOffset := y * stride
	rowEnd := rowOffset + ipf.img.width*4
	if rowOffset < 0 || rowEnd > len(ipf.img.Data) {
		return nil
	}
	return ipf.img.Data[rowOffset:rowEnd]
}

func (ipf *imagePixelFormat) RowData(y int) []basics.Int8u {
	return ipf.RowPtr(y)
}
