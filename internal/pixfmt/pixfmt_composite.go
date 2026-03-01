package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt/blender"
)

type compositeRGBABlender[CS color.Space, O order.RGBAOrder] interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u)
	GetOp() blender.CompOp
}

// PixFmtCompositeRGBA represents an RGBA pixel format with composite blending
type PixFmtCompositeRGBA[CS color.Space, O order.RGBAOrder] struct {
	rbuf          *buffer.RenderingBufferU8
	blender       compositeRGBABlender[CS, O]
	premultiplied bool
}

// NewPixFmtCompositeRGBA creates a new composite RGBA pixel format
func NewPixFmtCompositeRGBA[CS color.Space, O order.RGBAOrder](rbuf *buffer.RenderingBufferU8, op blender.CompOp) *PixFmtCompositeRGBA[CS, O] {
	return &PixFmtCompositeRGBA[CS, O]{
		rbuf:    rbuf,
		blender: blender.NewCompositeBlender[CS, O](op),
	}
}

// NewPixFmtCompositeRGBAPre creates a new premultiplied-source composite RGBA pixel format.
func NewPixFmtCompositeRGBAPre[CS color.Space, O order.RGBAOrder](rbuf *buffer.RenderingBufferU8, op blender.CompOp) *PixFmtCompositeRGBA[CS, O] {
	return &PixFmtCompositeRGBA[CS, O]{
		rbuf:          rbuf,
		blender:       blender.NewCompositeBlenderPre[CS, O](op),
		premultiplied: true,
	}
}

// Width returns the buffer width
func (pf *PixFmtCompositeRGBA[CS, O]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtCompositeRGBA[CS, O]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (4 for RGBA)
func (pf *PixFmtCompositeRGBA[CS, O]) PixWidth() int {
	return 4
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtCompositeRGBA[CS, O]) GetPixel(x, y int) color.RGBA8[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA8[CS]{}
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return color.RGBA8[CS]{}
	}

	var order O
	return color.RGBA8[CS]{
		R: row[pixelOffset+order.IdxR()],
		G: row[pixelOffset+order.IdxG()],
		B: row[pixelOffset+order.IdxB()],
		A: row[pixelOffset+order.IdxA()],
	}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtCompositeRGBA[CS, O]) CopyPixel(x, y int, c color.RGBA8[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return
	}

	var order O
	row[pixelOffset+order.IdxR()] = c.R
	row[pixelOffset+order.IdxG()] = c.G
	row[pixelOffset+order.IdxB()] = c.B
	row[pixelOffset+order.IdxA()] = c.A
}

// BlendPixel blends a pixel using composite blending
func (pf *PixFmtCompositeRGBA[CS, O]) BlendPixel(x, y int, c color.RGBA8[CS], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return
	}

	// Use the composite blender to perform the blending operation
	pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
}

// BlendHline blends a horizontal line of pixels (interface method)
func (pf *PixFmtCompositeRGBA[CS, O]) BlendHline(x, y, length int, c color.RGBA8[CS], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() {
		return
	}

	startX := max(0, x)
	endX := min(x+length, pf.Width())

	if startX >= endX {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)

	for i := startX; i < endX; i++ {
		pixelOffset := i * 4
		if pixelOffset+3 < len(row) {
			pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
		}
	}
}

// BlendVline blends a vertical line of pixels (interface method)
func (pf *PixFmtCompositeRGBA[CS, O]) BlendVline(x, y, length int, c color.RGBA8[CS], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() {
		return
	}

	startY := max(0, y)
	endY := min(y+length, pf.Height())

	if startY >= endY {
		return
	}

	pixelOffset := x * 4

	for i := startY; i < endY; i++ {
		row := buffer.RowU8(pf.rbuf, i)
		if pixelOffset+3 < len(row) {
			pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
		}
	}
}

// BlendSolidHspan blends a horizontal span with solid color (interface method)
func (pf *PixFmtCompositeRGBA[CS, O]) BlendSolidHspan(x, y, length int, c color.RGBA8[CS], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() {
		return
	}

	startX := max(0, x)
	endX := min(x+length, pf.Width())

	if startX >= endX {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)

	for i := startX; i < endX; i++ {
		pixelOffset := i * 4
		coverIndex := i - x
		if pixelOffset+3 < len(row) && coverIndex < len(covers) {
			pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, covers[coverIndex])
		}
	}
}

// SetCompOp changes the composite operation
func (pf *PixFmtCompositeRGBA[CS, O]) SetCompOp(op blender.CompOp) {
	if pf.premultiplied {
		pf.blender = blender.NewCompositeBlenderPre[CS, O](op)
		return
	}
	pf.blender = blender.NewCompositeBlender[CS, O](op)
}

// GetCompOp returns the current composite operation
func (pf *PixFmtCompositeRGBA[CS, O]) GetCompOp() blender.CompOp {
	return pf.blender.GetOp()
}

// Pixel returns the pixel at the given coordinates (alias for GetPixel)
func (pf *PixFmtCompositeRGBA[CS, O]) Pixel(x, y int) color.RGBA8[CS] {
	return pf.GetPixel(x, y)
}

// CopyHline copies a horizontal line of pixels without blending
func (pf *PixFmtCompositeRGBA[CS, O]) CopyHline(x, y, length int, c color.RGBA8[CS]) {
	for i := 0; i < length; i++ {
		pf.CopyPixel(x+i, y, c)
	}
}

// CopyVline copies a vertical line of pixels without blending
func (pf *PixFmtCompositeRGBA[CS, O]) CopyVline(x, y, length int, c color.RGBA8[CS]) {
	for i := 0; i < length; i++ {
		pf.CopyPixel(x, y+i, c)
	}
}

// BlendVline blends a vertical line of pixels (already implemented above)
// CopyBar copies a rectangular area with solid color
func (pf *PixFmtCompositeRGBA[CS, O]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[CS]) {
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

// BlendBar blends a rectangular area with solid color
func (pf *PixFmtCompositeRGBA[CS, O]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[CS], cover basics.Int8u) {
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

// BlendSolidVspan blends a vertical span with solid color
func (pf *PixFmtCompositeRGBA[CS, O]) BlendSolidVspan(x, y, length int, c color.RGBA8[CS], covers []basics.Int8u) {
	for i := 0; i < length && i < len(covers); i++ {
		pf.BlendPixel(x, y+i, c, covers[i])
	}
}

// CopyColorHspan copies a horizontal span with varying colors
func (pf *PixFmtCompositeRGBA[CS, O]) CopyColorHspan(x, y, length int, colors []color.RGBA8[CS]) {
	for i := 0; i < length && i < len(colors); i++ {
		pf.CopyPixel(x+i, y, colors[i])
	}
}

// BlendColorHspan blends a horizontal span with varying colors
func (pf *PixFmtCompositeRGBA[CS, O]) BlendColorHspan(x, y, length int, colors []color.RGBA8[CS], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		actualCover := cover
		if i < len(covers) {
			actualCover = covers[i]
		}
		pf.BlendPixel(x+i, y, colors[i], actualCover)
	}
}

// BlendFrom blends a single scanline from another RGBA surface.
func (pf *PixFmtCompositeRGBA[CS, O]) BlendFrom(src interface {
	GetPixel(x, y int) color.RGBA8[CS]
	Width() int
	Height() int
}, xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u,
) {
	if ydst < 0 || ydst >= pf.Height() || ysrc < 0 || ysrc >= src.Height() || length <= 0 {
		return
	}

	if xsrc < 0 {
		length += xsrc
		xdst -= xsrc
		xsrc = 0
	}
	if xdst < 0 {
		length += xdst
		xsrc -= xdst
		xdst = 0
	}
	if xsrc+length > src.Width() {
		length = src.Width() - xsrc
	}
	if xdst+length > pf.Width() {
		length = pf.Width() - xdst
	}
	if length <= 0 {
		return
	}

	start := 0
	end := length
	step := 1
	if xdst > xsrc {
		start = length - 1
		end = -1
		step = -1
	}

	for i := start; i != end; i += step {
		pf.BlendPixel(xdst+i, ydst, src.GetPixel(xsrc+i, ysrc), cover)
	}
}

// CopyColorVspan copies a vertical span with varying colors
func (pf *PixFmtCompositeRGBA[CS, O]) CopyColorVspan(x, y, length int, colors []color.RGBA8[CS]) {
	for i := 0; i < length && i < len(colors); i++ {
		pf.CopyPixel(x, y+i, colors[i])
	}
}

// BlendColorVspan blends a vertical span with varying colors
func (pf *PixFmtCompositeRGBA[CS, O]) BlendColorVspan(x, y, length int, colors []color.RGBA8[CS], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		actualCover := cover
		if i < len(covers) {
			actualCover = covers[i]
		}
		pf.BlendPixel(x, y+i, colors[i], actualCover)
	}
}

// Clear fills the entire buffer with the specified color
func (pf *PixFmtCompositeRGBA[CS, O]) Clear(c color.RGBA8[CS]) {
	// Fill entire buffer
	width := pf.Width()
	height := pf.Height()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pf.CopyPixel(x, y, c)
		}
	}
}

// Fill fills the entire buffer with the specified color (alias for Clear)
func (pf *PixFmtCompositeRGBA[CS, O]) Fill(c color.RGBA8[CS]) {
	pf.Clear(c)
}

// Note: max and min functions are already defined in pixfmt_rgba64.go, so we don't redefine them here

// Concrete composite pixel format types for convenience
type (
	PixFmtCompositeRGBA32     = PixFmtCompositeRGBA[color.Linear, order.RGBA]
	PixFmtCompositeARGB32     = PixFmtCompositeRGBA[color.Linear, order.ARGB]
	PixFmtCompositeBGRA32     = PixFmtCompositeRGBA[color.Linear, order.BGRA]
	PixFmtCompositeABGR32     = PixFmtCompositeRGBA[color.Linear, order.ABGR]
	PixFmtCompositeRGBA32Pre  = PixFmtCompositeRGBA[color.Linear, order.RGBA]
	PixFmtCompositeRGBA32SRGB = PixFmtCompositeRGBA[color.SRGB, order.RGBA]
	PixFmtCompositeARGB32SRGB = PixFmtCompositeRGBA[color.SRGB, order.ARGB]
	PixFmtCompositeBGRA32SRGB = PixFmtCompositeRGBA[color.SRGB, order.BGRA]
	PixFmtCompositeABGR32SRGB = PixFmtCompositeRGBA[color.SRGB, order.ABGR]
)

// NewPixFmtCompositeRGBA32 creates a new composite RGBA32 pixel format
func NewPixFmtCompositeRGBA32(rbuf *buffer.RenderingBufferU8, op blender.CompOp) *PixFmtCompositeRGBA32 {
	return NewPixFmtCompositeRGBA[color.Linear, order.RGBA](rbuf, op)
}

func NewPixFmtCompositeRGBA32Pre(rbuf *buffer.RenderingBufferU8, op blender.CompOp) *PixFmtCompositeRGBA32Pre {
	return NewPixFmtCompositeRGBAPre[color.Linear, order.RGBA](rbuf, op)
}
