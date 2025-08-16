package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

// RGB pixel type for internal operations (24-bit, no alpha)
type RGBPixelType struct {
	R, G, B basics.Int8u
}

// Set sets all RGB components
func (p *RGBPixelType) Set(r, g, b basics.Int8u) {
	p.R, p.G, p.B = r, g, b
}

// SetColor sets from a color type
func (p *RGBPixelType) SetColor(c color.RGB8[color.Linear]) {
	p.R, p.G, p.B = c.R, c.G, c.B
}

// GetColor returns as color type
func (p *RGBPixelType) GetColor() color.RGB8[color.Linear] {
	return color.RGB8[color.Linear]{R: p.R, G: p.G, B: p.B}
}

// PixFmtAlphaBlendRGB represents the main RGB pixel format with alpha blending
// This is a 24-bit format (3 bytes per pixel) without alpha channel storage
type PixFmtAlphaBlendRGB[B any, CS any] struct {
	rbuf     *buffer.RenderingBufferU8
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGB creates a new RGB pixel format
func NewPixFmtAlphaBlendRGB[B any, CS any](rbuf *buffer.RenderingBufferU8, blender B) *PixFmtAlphaBlendRGB[B, CS] {
	return &PixFmtAlphaBlendRGB[B, CS]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGB[B, CS]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGB[B, CS]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (3 for RGB)
func (pf *PixFmtAlphaBlendRGB[B, CS]) PixWidth() int {
	return 3
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGB[B, CS]) GetPixel(x, y int) color.RGB8[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[CS]{}
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return color.RGB8[CS]{}
	}

	// Use fixed RGB order - RGB24 formats always use RGB order
	return color.RGB8[CS]{
		R: row[pixelOffset],
		G: row[pixelOffset+1],
		B: row[pixelOffset+2],
	}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGB[B, CS]) CopyPixel(x, y int, c color.RGB8[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	// Use fixed RGB order - RGB24 formats always use RGB order
	row[pixelOffset] = c.R
	row[pixelOffset+1] = c.G
	row[pixelOffset+2] = c.B
}

// BlendPixel blends a pixel with the given alpha and coverage
func (pf *PixFmtAlphaBlendRGB[B, CS]) BlendPixel(x, y int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	// Use interface assertion for blending
	if blender, ok := interface{}(pf.blender).(RGBBlender); ok {
		blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, cover)
	}
}

// BlendPixelRGBA blends an RGBA pixel (ignores alpha channel for storage)
func (pf *PixFmtAlphaBlendRGB[B, CS]) BlendPixelRGBA(x, y int, c color.RGBA8[CS], cover basics.Int8u) {
	pf.BlendPixel(x, y, color.RGB8[CS]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

// CopyHline copies a horizontal line
func (pf *PixFmtAlphaBlendRGB[B, CS]) CopyHline(x1, y, x2 int, c color.RGB8[CS]) {
	if y < 0 || y >= pf.Height() {
		return
	}

	x1 = ClampX(x1, pf.Width())
	x2 = ClampX(x2, pf.Width())
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	row := buffer.RowU8(pf.rbuf, y)
	for x := x1; x <= x2; x++ {
		pixelOffset := x * 3
		if pixelOffset+2 < len(row) {
			row[pixelOffset] = c.R
			row[pixelOffset+1] = c.G
			row[pixelOffset+2] = c.B
		}
	}
}

// BlendHline blends a horizontal line
func (pf *PixFmtAlphaBlendRGB[B, CS]) BlendHline(x1, y, x2 int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() {
		return
	}

	x1 = ClampX(x1, pf.Width())
	x2 = ClampX(x2, pf.Width())
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	row := buffer.RowU8(pf.rbuf, y)
	if blender, ok := interface{}(pf.blender).(RGBBlender); ok {
		for x := x1; x <= x2; x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, cover)
			}
		}
	}
}

// CopyVline copies a vertical line
func (pf *PixFmtAlphaBlendRGB[B, CS]) CopyVline(x, y1, y2 int, c color.RGB8[CS]) {
	if x < 0 || x >= pf.Width() {
		return
	}

	y1 = ClampY(y1, pf.Height())
	y2 = ClampY(y2, pf.Height())
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	for y := y1; y <= y2; y++ {
		pf.CopyPixel(x, y, c)
	}
}

// BlendVline blends a vertical line
func (pf *PixFmtAlphaBlendRGB[B, CS]) BlendVline(x, y1, y2 int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() {
		return
	}

	y1 = ClampY(y1, pf.Height())
	y2 = ClampY(y2, pf.Height())
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	for y := y1; y <= y2; y++ {
		pf.BlendPixel(x, y, c, alpha, cover)
	}
}

// CopyBar copies a filled rectangle
func (pf *PixFmtAlphaBlendRGB[B, CS]) CopyBar(x1, y1, x2, y2 int, c color.RGB8[CS]) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2, c)
	}
}

// BlendBar blends a filled rectangle
func (pf *PixFmtAlphaBlendRGB[B, CS]) BlendBar(x1, y1, x2, y2 int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2, c, alpha, cover)
	}
}

// BlendSolidHspan blends a horizontal span with varying coverage
func (pf *PixFmtAlphaBlendRGB[B, CS]) BlendSolidHspan(x, y, length int, c color.RGB8[CS], alpha basics.Int8u, covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)
	if blender, ok := interface{}(pf.blender).(RGBBlender); ok {
		if covers == nil {
			// Uniform coverage
			for i := 0; i < length; i++ {
				pixelOffset := (x + i) * 3
				if pixelOffset+2 < len(row) {
					blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, 255)
				}
			}
		} else {
			// Varying coverage
			for i := 0; i < length && i < len(covers); i++ {
				if covers[i] > 0 {
					pixelOffset := (x + i) * 3
					if pixelOffset+2 < len(row) {
						blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, covers[i])
					}
				}
			}
		}
	}
}

// BlendSolidVspan blends a vertical span with varying coverage
func (pf *PixFmtAlphaBlendRGB[B, CS]) BlendSolidVspan(x, y, length int, c color.RGB8[CS], alpha basics.Int8u, covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	if covers == nil {
		// Uniform coverage
		for i := 0; i < length; i++ {
			pf.BlendPixel(x, y+i, c, alpha, 255)
		}
	} else {
		// Varying coverage
		for i := 0; i < length && i < len(covers); i++ {
			if covers[i] > 0 {
				pf.BlendPixel(x, y+i, c, alpha, covers[i])
			}
		}
	}
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGB[B, CS]) Clear(c color.RGB8[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				row[pixelOffset] = c.R
				row[pixelOffset+1] = c.G
				row[pixelOffset+2] = c.B
			}
		}
	}
}

// Fill is an alias for Clear (fills entire buffer)
func (pf *PixFmtAlphaBlendRGB[B, CS]) Fill(c color.RGB8[CS]) {
	pf.Clear(c)
}

// CopyFrom copies from another RGB pixel format
func (pf *PixFmtAlphaBlendRGB[B, CS]) CopyFrom(src *PixFmtAlphaBlendRGB[B, CS], srcX, srcY, dstX, dstY, width, height int) {
	// Clamp source and destination rectangles
	if srcX < 0 {
		width += srcX
		dstX -= srcX
		srcX = 0
	}
	if srcY < 0 {
		height += srcY
		dstY -= srcY
		srcY = 0
	}
	if dstX < 0 {
		width += dstX
		srcX -= dstX
		dstX = 0
	}
	if dstY < 0 {
		height += dstY
		srcY -= dstY
		dstY = 0
	}

	if srcX+width > src.Width() {
		width = src.Width() - srcX
	}
	if srcY+height > src.Height() {
		height = src.Height() - srcY
	}
	if dstX+width > pf.Width() {
		width = pf.Width() - dstX
	}
	if dstY+height > pf.Height() {
		height = pf.Height() - dstY
	}

	if width <= 0 || height <= 0 {
		return
	}

	// Copy pixel by pixel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := src.GetPixel(srcX+x, srcY+y)
			pf.CopyPixel(dstX+x, dstY+y, pixel)
		}
	}
}

// Concrete RGB pixel format types for different color orders
type (
	PixFmtRGB24  = PixFmtAlphaBlendRGB[BlenderRGB24, color.Linear]
	PixFmtBGR24  = PixFmtAlphaBlendRGB[BlenderBGR24, color.Linear]
	PixFmtSRGB24 = PixFmtAlphaBlendRGB[BlenderRGB24SRGB, color.SRGB]
	PixFmtSBGR24 = PixFmtAlphaBlendRGB[BlenderBGR24SRGB, color.SRGB]
)

// Constructor functions for RGB24 pixel formats
func NewPixFmtRGB24(rbuf *buffer.RenderingBufferU8) *PixFmtRGB24 {
	return NewPixFmtAlphaBlendRGB[BlenderRGB24, color.Linear](rbuf, BlenderRGB24{})
}

func NewPixFmtBGR24(rbuf *buffer.RenderingBufferU8) *PixFmtBGR24 {
	return NewPixFmtAlphaBlendRGB[BlenderBGR24, color.Linear](rbuf, BlenderBGR24{})
}

func NewPixFmtSRGB24(rbuf *buffer.RenderingBufferU8) *PixFmtSRGB24 {
	return NewPixFmtAlphaBlendRGB[BlenderRGB24SRGB, color.SRGB](rbuf, BlenderRGB24SRGB{})
}

func NewPixFmtSBGR24(rbuf *buffer.RenderingBufferU8) *PixFmtSBGR24 {
	return NewPixFmtAlphaBlendRGB[BlenderBGR24SRGB, color.SRGB](rbuf, BlenderBGR24SRGB{})
}

// RGB48 types (16-bit per channel)
type (
	PixFmtRGB48 = PixFmtAlphaBlendRGB[BlenderRGB24, color.Linear] // TODO: Create 16-bit blenders
	PixFmtBGR48 = PixFmtAlphaBlendRGB[BlenderBGR24, color.Linear]
)

// TODO: Implement RGB48 (16-bit per channel) formats
// These would require 16-bit blenders and different pixel layouts
