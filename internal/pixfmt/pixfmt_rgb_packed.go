package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// Packed pixel formats use 16-bit integers to store RGB values
// with reduced precision to save memory

// Packed pixel format order types
type (
	RGB555Order struct{} // 5-5-5 bits RGB with 1 unused bit
	RGB565Order struct{} // 5-6-5 bits RGB
	BGR555Order struct{} // 5-5-5 bits BGR with 1 unused bit
	BGR565Order struct{} // 5-6-5 bits BGR
)

// Packed pixel utilities for RGB555 format (5-5-5 with 1 unused bit)
// Format: -RRRRR GGGGG BBBBB (bit 15 unused, typically set to 1)

// MakePixel555 packs 8-bit RGB values into RGB555 format
func MakePixel555(r, g, b basics.Int8u) basics.Int16u {
	return basics.Int16u(((uint16(r)&0xF8)<<7)|((uint16(g)&0xF8)<<2)|(uint16(b)>>3)) | 0x8000
}

// UnpackPixel555 extracts 8-bit RGB values from RGB555 format
// Uses the same format as AGG C++: 1RRRRRGGGGGBBBBB with gaps
func UnpackPixel555(pixel basics.Int16u) (r, g, b basics.Int8u) {
	r = basics.Int8u((pixel >> 7) & 0xF8)
	g = basics.Int8u((pixel >> 2) & 0xF8)
	b = basics.Int8u((pixel << 3) & 0xF8)
	return
}

// MakePixelBGR555 packs 8-bit RGB values into BGR555 format
func MakePixelBGR555(r, g, b basics.Int8u) basics.Int16u {
	return basics.Int16u(((uint16(b)&0xF8)<<7)|((uint16(g)&0xF8)<<2)|(uint16(r)>>3)) | 0x8000
}

// UnpackPixelBGR555 extracts 8-bit RGB values from BGR555 format
func UnpackPixelBGR555(pixel basics.Int16u) (r, g, b basics.Int8u) {
	b = basics.Int8u((pixel >> 7) & 0xF8)
	g = basics.Int8u((pixel >> 2) & 0xF8)
	r = basics.Int8u((pixel << 3) & 0xF8)
	return
}

// Packed pixel utilities for RGB565 format (5-6-5 bits)
// Format: RRRRR GGGGGG BBBBB

// MakePixel565 packs 8-bit RGB values into RGB565 format
func MakePixel565(r, g, b basics.Int8u) basics.Int16u {
	return basics.Int16u(((uint16(r) & 0xF8) << 8) | ((uint16(g) & 0xFC) << 3) | (uint16(b) >> 3))
}

// UnpackPixel565 extracts 8-bit RGB values from RGB565 format
func UnpackPixel565(pixel basics.Int16u) (r, g, b basics.Int8u) {
	r = basics.Int8u((pixel >> 8) & 0xF8)
	g = basics.Int8u((pixel >> 3) & 0xFC)
	b = basics.Int8u((pixel << 3) & 0xF8)
	return
}

// MakePixelBGR565 packs 8-bit RGB values into BGR565 format
func MakePixelBGR565(r, g, b basics.Int8u) basics.Int16u {
	return basics.Int16u(((uint16(b) & 0xF8) << 8) | ((uint16(g) & 0xFC) << 3) | (uint16(r) >> 3))
}

// UnpackPixelBGR565 extracts 8-bit RGB values from BGR565 format
func UnpackPixelBGR565(pixel basics.Int16u) (r, g, b basics.Int8u) {
	b = basics.Int8u((pixel >> 8) & 0xF8)
	g = basics.Int8u((pixel >> 3) & 0xFC)
	r = basics.Int8u((pixel << 3) & 0xF8)
	return
}

// PixFmtRGB555 represents a 16-bit RGB555 pixel format
type PixFmtRGB555[B blender.RGB16PackedBlender] struct {
	rbuf     *buffer.RenderingBufferU16
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtRGB555 creates a new RGB555 pixel format
func NewPixFmtRGB555[B blender.RGB16PackedBlender](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtRGB555[B] {
	return &PixFmtRGB555[B]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtRGB555[B]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtRGB555[B]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (2 for 16-bit)
func (pf *PixFmtRGB555[B]) PixWidth() int {
	return 2
}

// GetPixel returns the pixel at the given coordinates as RGB8
func (pf *PixFmtRGB555[B]) GetPixel(x, y int) color.RGB8[color.Linear] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[color.Linear]{}
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return color.RGB8[color.Linear]{}
	}

	r, g, b := UnpackPixel555(row[x])
	return color.RGB8[color.Linear]{R: r, G: g, B: b}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtRGB555[B]) CopyPixel(x, y int, c color.RGB8[color.Linear]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return
	}

	row[x] = MakePixel555(c.R, c.G, c.B)
}

// BlendPixel blends a pixel with alpha and coverage
func (pf *PixFmtRGB555[B]) BlendPixel(x, y int, c color.RGB8[color.Linear], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return
	}

	// Direct blending call - no type assertion needed with proper constraints
	pf.blender.BlendPix(&row[x], c.R, c.G, c.B, alpha, cover)
}

// PixFmtRGB565 represents a 16-bit RGB565 pixel format
type PixFmtRGB565[B blender.RGB16PackedBlender] struct {
	rbuf     *buffer.RenderingBufferU16
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtRGB565 creates a new RGB565 pixel format
func NewPixFmtRGB565[B blender.RGB16PackedBlender](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtRGB565[B] {
	return &PixFmtRGB565[B]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtRGB565[B]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtRGB565[B]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (2 for 16-bit)
func (pf *PixFmtRGB565[B]) PixWidth() int {
	return 2
}

// GetPixel returns the pixel at the given coordinates as RGB8
func (pf *PixFmtRGB565[B]) GetPixel(x, y int) color.RGB8[color.Linear] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[color.Linear]{}
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return color.RGB8[color.Linear]{}
	}

	r, g, b := UnpackPixel565(row[x])
	return color.RGB8[color.Linear]{R: r, G: g, B: b}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtRGB565[B]) CopyPixel(x, y int, c color.RGB8[color.Linear]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return
	}

	row[x] = MakePixel565(c.R, c.G, c.B)
}

// BlendPixel blends a pixel with alpha and coverage
func (pf *PixFmtRGB565[B]) BlendPixel(x, y int, c color.RGB8[color.Linear], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return
	}

	// Direct blending call - no type assertion needed with proper constraints
	pf.blender.BlendPix(&row[x], c.R, c.G, c.B, alpha, cover)
}

// PixFmtBGR555 represents a 16-bit BGR555 pixel format
type PixFmtBGR555[B blender.RGB16PackedBlender] struct {
	rbuf     *buffer.RenderingBufferU16
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtBGR555 creates a new BGR555 pixel format
func NewPixFmtBGR555[B blender.RGB16PackedBlender](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtBGR555[B] {
	return &PixFmtBGR555[B]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtBGR555[B]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtBGR555[B]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (2 for 16-bit)
func (pf *PixFmtBGR555[B]) PixWidth() int {
	return 2
}

// GetPixel returns the pixel at the given coordinates as RGB8
func (pf *PixFmtBGR555[B]) GetPixel(x, y int) color.RGB8[color.Linear] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[color.Linear]{}
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return color.RGB8[color.Linear]{}
	}

	r, g, b := UnpackPixelBGR555(row[x])
	return color.RGB8[color.Linear]{R: r, G: g, B: b}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtBGR555[B]) CopyPixel(x, y int, c color.RGB8[color.Linear]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return
	}

	row[x] = MakePixelBGR555(c.R, c.G, c.B)
}

// BlendPixel blends a pixel with alpha and coverage
func (pf *PixFmtBGR555[B]) BlendPixel(x, y int, c color.RGB8[color.Linear], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return
	}

	// Direct blending call - no type assertion needed with proper constraints
	pf.blender.BlendPix(&row[x], c.R, c.G, c.B, alpha, cover)
}

// PixFmtBGR565 represents a 16-bit BGR565 pixel format
type PixFmtBGR565[B blender.RGB16PackedBlender] struct {
	rbuf     *buffer.RenderingBufferU16
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtBGR565 creates a new BGR565 pixel format
func NewPixFmtBGR565[B blender.RGB16PackedBlender](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtBGR565[B] {
	return &PixFmtBGR565[B]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtBGR565[B]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtBGR565[B]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (2 for 16-bit)
func (pf *PixFmtBGR565[B]) PixWidth() int {
	return 2
}

// GetPixel returns the pixel at the given coordinates as RGB8
func (pf *PixFmtBGR565[B]) GetPixel(x, y int) color.RGB8[color.Linear] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[color.Linear]{}
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return color.RGB8[color.Linear]{}
	}

	r, g, b := UnpackPixelBGR565(row[x])
	return color.RGB8[color.Linear]{R: r, G: g, B: b}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtBGR565[B]) CopyPixel(x, y int, c color.RGB8[color.Linear]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return
	}

	row[x] = MakePixelBGR565(c.R, c.G, c.B)
}

// BlendPixel blends a pixel with alpha and coverage
func (pf *PixFmtBGR565[B]) BlendPixel(x, y int, c color.RGB8[color.Linear], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	if x >= len(row) {
		return
	}

	// Direct blending call - no type assertion needed with proper constraints
	pf.blender.BlendPix(&row[x], c.R, c.G, c.B, alpha, cover)
}

// NoBlender is a placeholder type for pixel formats without blending capability
type NoBlender struct{}

// BlendPix implements RGB16PackedBlender interface with no-op blending
func (nb NoBlender) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {
	// No-op: plain formats don't perform blending, just overwrite
	// This method should not be called for plain formats, but provided for interface compliance
}

// MakePix implements RGB16PackedBlender interface - should not be used for NoBlender
func (nb NoBlender) MakePix(r, g, b basics.Int8u) basics.Int16u {
	// This should not be called for NoBlender, return zero value
	return 0
}

// Convenience type aliases for common packed formats
type (
	// RGB555 formats
	PixFmtRGB555Plain = PixFmtRGB555[NoBlender]
	PixFmtBGR555Plain = PixFmtBGR555[NoBlender]

	// RGB565 formats
	PixFmtRGB565Plain = PixFmtRGB565[NoBlender]
	PixFmtBGR565Plain = PixFmtBGR565[NoBlender]
)
