package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt/blender"
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
// 24-bit RGB (no stored alpha), blender does src-alpha coverage compositing.
type PixFmtAlphaBlendRGB[S color.Space, B blender.RGBBlender[S]] struct {
	rbuf    *buffer.RenderingBufferU8
	blender B
}

func NewPixFmtAlphaBlendRGB[S color.Space, B blender.RGBBlender[S]](rbuf *buffer.RenderingBufferU8, b B) *PixFmtAlphaBlendRGB[S, B] {
	return &PixFmtAlphaBlendRGB[S, B]{rbuf: rbuf, blender: b}
}

func (pf *PixFmtAlphaBlendRGB[S, B]) Width() int    { return pf.rbuf.Width() }
func (pf *PixFmtAlphaBlendRGB[S, B]) Height() int   { return pf.rbuf.Height() }
func (pf *PixFmtAlphaBlendRGB[S, B]) PixWidth() int { return 3 }

func (pf *PixFmtAlphaBlendRGB[S, B]) GetPixel(x, y int) color.RGB8[S] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[S]{}
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 3
	if off+2 >= len(row) {
		return color.RGB8[S]{}
	}

	r, g, b := pf.blender.GetPlain(row[off : off+3])
	return color.RGB8[S]{R: r, G: g, B: b}
}

// Pixel returns the pixel at the given coordinates (alias for GetPixel)
func (pf *PixFmtAlphaBlendRGB[S, B]) Pixel(x, y int) color.RGB8[S] {
	return pf.GetPixel(x, y)
}

func (pf *PixFmtAlphaBlendRGB[S, B]) CopyPixel(x, y int, c color.RGB8[S]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 3
	if off+2 >= len(row) {
		return
	}

	pf.blender.SetPlain(row[off:off+3], c.R, c.G, c.B)
}

// alpha = source alpha (plain), cover = coverage
func (pf *PixFmtAlphaBlendRGB[S, B]) BlendPixel(x, y int, c color.RGB8[S], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 3
	if off+2 >= len(row) {
		return
	}
	// Dest slice is the full 3-byte pixel; blender uses O to address channels.
	pf.blender.BlendPix(row[off:off+3], c.R, c.G, c.B, alpha, cover)
}

func (pf *PixFmtAlphaBlendRGB[S, B]) BlendPixelRGBA(x, y int, c color.RGBA8[S], cover basics.Int8u) {
	pf.BlendPixel(x, y, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

func (pf *PixFmtAlphaBlendRGB[S, B]) CopyHline(x1, y, x2 int, c color.RGB8[S]) {
	if y < 0 || y >= pf.Height() {
		return
	}
	x1 = ClampX(x1, pf.Width())
	x2 = ClampX(x2, pf.Width())
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	row := buffer.RowU8(pf.rbuf, y)
	if ro, ok := any(pf.blender).(blender.RawRGBOrder); ok {
		// Fast path: direct index access
		ir, ig, ib := ro.IdxR(), ro.IdxG(), ro.IdxB()
		for x := x1; x <= x2; x++ {
			off := x * 3
			if off+2 >= len(row) {
				break
			}
			row[off+ir], row[off+ig], row[off+ib] = c.R, c.G, c.B
		}
	} else {
		// Safe path: use blender SetPlain
		for x := x1; x <= x2; x++ {
			off := x * 3
			if off+2 >= len(row) {
				break
			}
			pf.blender.SetPlain(row[off:off+3], c.R, c.G, c.B)
		}
	}
}

func (pf *PixFmtAlphaBlendRGB[S, B]) BlendHline(x1, y, x2 int, c color.RGB8[S], alpha, cover basics.Int8u) {
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
		off := x * 3
		if off+2 >= len(row) {
			break
		}
		pf.blender.BlendPix(row[off:off+3], c.R, c.G, c.B, alpha, cover)
	}
}

func (pf *PixFmtAlphaBlendRGB[S, B]) CopyVline(x, y1, y2 int, c color.RGB8[S]) {
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

func (pf *PixFmtAlphaBlendRGB[S, B]) BlendVline(x, y1, y2 int, c color.RGB8[S], alpha, cover basics.Int8u) {
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

func (pf *PixFmtAlphaBlendRGB[S, B]) CopyBar(x1, y1, x2, y2 int, c color.RGB8[S]) {
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

func (pf *PixFmtAlphaBlendRGB[S, B]) BlendBar(x1, y1, x2, y2 int, c color.RGB8[S], alpha, cover basics.Int8u) {
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

func (pf *PixFmtAlphaBlendRGB[S, B]) BlendSolidHspan(x, y, length int, c color.RGB8[S], alpha basics.Int8u, covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU8(pf.rbuf, y)
	if covers == nil {
		for i := 0; i < length; i++ {
			off := (x + i) * 3
			if off+2 >= len(row) {
				break
			}
			pf.blender.BlendPix(row[off:off+3], c.R, c.G, c.B, alpha, 255)
		}
		return
	}
	for i := 0; i < length && i < len(covers); i++ {
		if covers[i] == 0 {
			continue
		}
		off := (x + i) * 3
		if off+2 >= len(row) {
			break
		}
		pf.blender.BlendPix(row[off:off+3], c.R, c.G, c.B, alpha, covers[i])
	}
}

func (pf *PixFmtAlphaBlendRGB[S, B]) BlendSolidVspan(x, y, length int, c color.RGB8[S], alpha basics.Int8u, covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if covers == nil {
		for i := 0; i < length; i++ {
			pf.BlendPixel(x, y+i, c, alpha, 255)
		}
		return
	}
	for i := 0; i < length && i < len(covers); i++ {
		if covers[i] != 0 {
			pf.BlendPixel(x, y+i, c, alpha, covers[i])
		}
	}
}

// CopyColorHspan copies a horizontal span of colors
func (pf *PixFmtAlphaBlendRGB[S, B]) CopyColorHspan(x, y, length int, colors []color.RGB8[S]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		pf.CopyPixel(x+i, y, colors[colorIdx])
	}
}

// BlendColorHspan blends a horizontal span of colors
func (pf *PixFmtAlphaBlendRGB[S, B]) BlendColorHspan(x, y, length int, colors []color.RGB8[S], covers []basics.Int8u, alpha, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		c := colors[colorIdx]

		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr > 0 {
			pf.BlendPixel(x+i, y, c, alpha, cvr)
		}
	}
}

// CopyColorVspan copies a vertical span of colors
func (pf *PixFmtAlphaBlendRGB[S, B]) CopyColorVspan(x, y, length int, colors []color.RGB8[S]) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		pf.CopyPixel(x, y+i, colors[colorIdx])
	}
}

// BlendColorVspan blends a vertical span of colors
func (pf *PixFmtAlphaBlendRGB[S, B]) BlendColorVspan(x, y, length int, colors []color.RGB8[S], covers []basics.Int8u, alpha, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		c := colors[colorIdx]

		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr > 0 {
			pf.BlendPixel(x, y+i, c, alpha, cvr)
		}
	}
}

func (pf *PixFmtAlphaBlendRGB[S, B]) Clear(c color.RGB8[S]) {
	if ro, ok := any(pf.blender).(blender.RawRGBOrder); ok {
		// Fast path: direct index access
		ir, ig, ib := ro.IdxR(), ro.IdxG(), ro.IdxB()
		for y := 0; y < pf.Height(); y++ {
			row := buffer.RowU8(pf.rbuf, y)
			p := 0
			for x := 0; x < pf.Width(); x++ {
				if p+2 >= len(row) {
					break
				}
				row[p+ir], row[p+ig], row[p+ib] = c.R, c.G, c.B
				p += 3
			}
		}
	} else {
		// Safe path: use blender SetPlain
		for y := 0; y < pf.Height(); y++ {
			row := buffer.RowU8(pf.rbuf, y)
			for x := 0; x < pf.Width(); x++ {
				off := x * 3
				if off+2 >= len(row) {
					break
				}
				pf.blender.SetPlain(row[off:off+3], c.R, c.G, c.B)
			}
		}
	}
}

// Fill is an alias for Clear (fills entire buffer)
func (pf *PixFmtAlphaBlendRGB[CS, B]) Fill(c color.RGB8[CS]) {
	pf.Clear(c)
}

// CopyFrom copies from another RGB pixel format
func (pf *PixFmtAlphaBlendRGB[CS, B]) CopyFrom(src *PixFmtAlphaBlendRGB[CS, B], srcX, srcY, dstX, dstY, width, height int) {
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

// ==============================================================================
// AGG Compatibility: Whole-Buffer Utilities
// ==============================================================================

// Premultiply converts the entire buffer to premultiplied alpha format.
// For RGB formats without stored alpha, this is a no-op since there's no alpha channel.
func (pf *PixFmtAlphaBlendRGB[S, B]) Premultiply() {
	// RGB formats don't store alpha, so premultiplying is not applicable
	// This is a no-op for RGB formats
}

// Demultiply converts the entire buffer from premultiplied alpha format.
// For RGB formats without stored alpha, this is a no-op since there's no alpha channel.
func (pf *PixFmtAlphaBlendRGB[S, B]) Demultiply() {
	// RGB formats don't store alpha, so demultiplying is not applicable
	// This is a no-op for RGB formats
}

// ApplyGammaDir applies gamma correction to the entire buffer using the forward direction.
func (pf *PixFmtAlphaBlendRGB[S, B]) ApplyGammaDir(gamma func(basics.Int8u) basics.Int8u) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			off := x * 3
			if off+2 >= len(row) {
				break
			}

			// Use fast path if available
			if ro, ok := any(pf.blender).(blender.RawRGBOrder); ok {
				ir, ig, ib := ro.IdxR(), ro.IdxG(), ro.IdxB()
				row[off+ir] = gamma(row[off+ir])
				row[off+ig] = gamma(row[off+ig])
				row[off+ib] = gamma(row[off+ib])
			} else {
				// Use blender interface
				r, g, b := pf.blender.GetPlain(row[off : off+3])
				r, g, b = gamma(r), gamma(g), gamma(b)
				pf.blender.SetPlain(row[off:off+3], r, g, b)
			}
		}
	}
}

// ApplyGammaInv applies inverse gamma correction to the entire buffer.
func (pf *PixFmtAlphaBlendRGB[S, B]) ApplyGammaInv(gamma func(basics.Int8u) basics.Int8u) {
	pf.ApplyGammaDir(gamma) // Same implementation for RGB formats
}

// Concrete RGB pixel format types for different color orders
type (
	PixFmtRGB24  = PixFmtAlphaBlendRGB[color.Linear, blender.BlenderRGB8[color.Linear, order.RGB]]
	PixFmtBGR24  = PixFmtAlphaBlendRGB[color.Linear, blender.BlenderRGB8[color.Linear, order.BGR]]
	PixFmtSRGB24 = PixFmtAlphaBlendRGB[color.SRGB, blender.BlenderRGB8[color.SRGB, order.RGB]]
	PixFmtSBGR24 = PixFmtAlphaBlendRGB[color.SRGB, blender.BlenderRGB8[color.SRGB, order.BGR]]
)

// Constructor functions for RGB24 pixel formats
func NewPixFmtRGB24(rbuf *buffer.RenderingBufferU8) *PixFmtRGB24 {
	return NewPixFmtAlphaBlendRGB[color.Linear](rbuf, blender.BlenderRGB8[color.Linear, order.RGB]{})
}

func NewPixFmtBGR24(rbuf *buffer.RenderingBufferU8) *PixFmtBGR24 {
	return NewPixFmtAlphaBlendRGB[color.Linear](rbuf, blender.BlenderRGB8[color.Linear, order.BGR]{})
}

func NewPixFmtSRGB24(rbuf *buffer.RenderingBufferU8) *PixFmtSRGB24 {
	return NewPixFmtAlphaBlendRGB[color.SRGB](rbuf, blender.BlenderRGB8[color.SRGB, order.RGB]{})
}

func NewPixFmtSBGR24(rbuf *buffer.RenderingBufferU8) *PixFmtSBGR24 {
	return NewPixFmtAlphaBlendRGB[color.SRGB](rbuf, blender.BlenderRGB8[color.SRGB, order.BGR]{})
}

// ==============================================================================
// Premultiplied RGB Pixel Format Types
// ==============================================================================

// 32-bit RGB with a padding byte. Only R/G/B stored/used; padding left untouched.
// ==============================================================================
// RGBX32 (32-bit RGB with padding byte) Pixel Formats
// ==============================================================================

type PixFmtAlphaBlendRGBX32[
	S color.Space,
	B blender.RGBXBlender[S],
] struct {
	rbuf    *buffer.RenderingBufferU8
	blender B
}

func NewPixFmtAlphaBlendRGBX32[
	S color.Space,
	B blender.RGBXBlender[S],
](rbuf *buffer.RenderingBufferU8, b B) *PixFmtAlphaBlendRGBX32[S, B] {
	return &PixFmtAlphaBlendRGBX32[S, B]{rbuf: rbuf, blender: b}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGBX32[S, B]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGBX32[S, B]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (4 for RGBX32)
func (pf *PixFmtAlphaBlendRGBX32[S, B]) PixWidth() int {
	return 4
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGBX32[S, B]) GetPixel(x, y int) color.RGB8[S] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[S]{}
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 4
	if off+3 >= len(row) {
		return color.RGB8[S]{}
	}

	// Use fast path if available, otherwise use blender interface
	if ro, ok := any(pf.blender).(blender.RawRGBXOrder); ok {
		ir, ig, ib := ro.IdxR(), ro.IdxG(), ro.IdxB()
		return color.RGB8[S]{R: row[off+ir], G: row[off+ig], B: row[off+ib]}
	} else {
		r, g, b := pf.blender.GetPlain(row[off : off+4])
		return color.RGB8[S]{R: r, G: g, B: b}
	}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGBX32[S, B]) CopyPixel(x, y int, c color.RGB8[S]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 4
	if off+3 >= len(row) {
		return
	}
	pf.blender.SetPlain(row[off:off+4], c.R, c.G, c.B)
}

// BlendPixel blends a pixel with the given alpha and coverage
func (pf *PixFmtAlphaBlendRGBX32[S, B]) BlendPixel(x, y int, c color.RGB8[S], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 4
	if off+3 >= len(row) {
		return
	}
	pf.blender.BlendPix(row[off:off+4], c.R, c.G, c.B, alpha, cover)
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGBX32[S, B]) Clear(c color.RGB8[S]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			off := x * 4
			if off+3 >= len(row) {
				break
			}
			pf.blender.SetPlain(row[off:off+4], c.R, c.G, c.B)
		}
	}
}

// Concrete RGBX32 pixel format types
type (
	PixFmtRGBX32[S color.Space] = PixFmtAlphaBlendRGBX32[S, blender.BlenderRGBX8[S, order.RGBX32]]
	PixFmtXRGB32[S color.Space] = PixFmtAlphaBlendRGBX32[S, blender.BlenderRGBX8[S, order.XRGB32]]
	PixFmtBGRX32[S color.Space] = PixFmtAlphaBlendRGBX32[S, blender.BlenderRGBX8[S, order.BGRX32]]
	PixFmtXBGR32[S color.Space] = PixFmtAlphaBlendRGBX32[S, blender.BlenderRGBX8[S, order.XBGR32]]
)

// Constructor functions for RGBX32 pixel formats
func NewPixFmtRGBX32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtRGBX32[S] {
	return NewPixFmtAlphaBlendRGBX32[S, blender.BlenderRGBX8[S, order.RGBX32]](rbuf, blender.BlenderRGBX8[S, order.RGBX32]{})
}

func NewPixFmtXRGB32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtXRGB32[S] {
	return NewPixFmtAlphaBlendRGBX32[S, blender.BlenderRGBX8[S, order.XRGB32]](rbuf, blender.BlenderRGBX8[S, order.XRGB32]{})
}

func NewPixFmtBGRX32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtBGRX32[S] {
	return NewPixFmtAlphaBlendRGBX32[S, blender.BlenderRGBX8[S, order.BGRX32]](rbuf, blender.BlenderRGBX8[S, order.BGRX32]{})
}

func NewPixFmtXBGR32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtXBGR32[S] {
	return NewPixFmtAlphaBlendRGBX32[S, blender.BlenderRGBX8[S, order.XBGR32]](rbuf, blender.BlenderRGBX8[S, order.XBGR32]{})
}

///

func NewPixFmtRGBX32Linear(r *buffer.RenderingBufferU8) *PixFmtRGBX32[color.Linear] {
	return NewPixFmtRGBX32[color.Linear](r)
}

func NewPixFmtRGBX32SRGB(r *buffer.RenderingBufferU8) *PixFmtRGBX32[color.SRGB] {
	return NewPixFmtRGBX32[color.SRGB](r)
}

// ==============================================================================
// Premultiplied RGB Pixel Format Types
// ==============================================================================

// RGB24 premultiplied variants
type (
	PixFmtRGB24Pre  = PixFmtAlphaBlendRGB[color.Linear, blender.BlenderRGBPre[color.Linear, order.RGB]]
	PixFmtBGR24Pre  = PixFmtAlphaBlendRGB[color.Linear, blender.BlenderRGBPre[color.Linear, order.BGR]]
	PixFmtSRGB24Pre = PixFmtAlphaBlendRGB[color.SRGB, blender.BlenderRGBPre[color.SRGB, order.RGB]]
	PixFmtSBGR24Pre = PixFmtAlphaBlendRGB[color.SRGB, blender.BlenderRGBPre[color.SRGB, order.BGR]]
)

// RGBX32 premultiplied variants
type (
	PixFmtRGBX32Pre[S color.Space] = PixFmtAlphaBlendRGBX32[S, blender.BlenderRGBXPre[S, order.RGBX32]]
	PixFmtXRGB32Pre[S color.Space] = PixFmtAlphaBlendRGBX32[S, blender.BlenderRGBXPre[S, order.XRGB32]]
	PixFmtBGRX32Pre[S color.Space] = PixFmtAlphaBlendRGBX32[S, blender.BlenderRGBXPre[S, order.BGRX32]]
	PixFmtXBGR32Pre[S color.Space] = PixFmtAlphaBlendRGBX32[S, blender.BlenderRGBXPre[S, order.XBGR32]]
)

// Constructor functions for premultiplied RGB24 formats
func NewPixFmtRGB24Pre(rbuf *buffer.RenderingBufferU8) *PixFmtRGB24Pre {
	return NewPixFmtAlphaBlendRGB[color.Linear](rbuf, blender.BlenderRGBPre[color.Linear, order.RGB]{})
}

func NewPixFmtBGR24Pre(rbuf *buffer.RenderingBufferU8) *PixFmtBGR24Pre {
	return NewPixFmtAlphaBlendRGB[color.Linear](rbuf, blender.BlenderRGBPre[color.Linear, order.BGR]{})
}

func NewPixFmtSRGB24Pre(rbuf *buffer.RenderingBufferU8) *PixFmtSRGB24Pre {
	return NewPixFmtAlphaBlendRGB[color.SRGB](rbuf, blender.BlenderRGBPre[color.SRGB, order.RGB]{})
}

func NewPixFmtSBGR24Pre(rbuf *buffer.RenderingBufferU8) *PixFmtSBGR24Pre {
	return NewPixFmtAlphaBlendRGB[color.SRGB](rbuf, blender.BlenderRGBPre[color.SRGB, order.BGR]{})
}

// Constructor functions for premultiplied RGBX32 formats
func NewPixFmtRGBX32Pre[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtRGBX32Pre[S] {
	return NewPixFmtAlphaBlendRGBX32[S, blender.BlenderRGBXPre[S, order.RGBX32]](rbuf, blender.BlenderRGBXPre[S, order.RGBX32]{})
}

func NewPixFmtXRGB32Pre[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtXRGB32Pre[S] {
	return NewPixFmtAlphaBlendRGBX32[S, blender.BlenderRGBXPre[S, order.XRGB32]](rbuf, blender.BlenderRGBXPre[S, order.XRGB32]{})
}

func NewPixFmtBGRX32Pre[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtBGRX32Pre[S] {
	return NewPixFmtAlphaBlendRGBX32[S, blender.BlenderRGBXPre[S, order.BGRX32]](rbuf, blender.BlenderRGBXPre[S, order.BGRX32]{})
}

func NewPixFmtXBGR32Pre[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtXBGR32Pre[S] {
	return NewPixFmtAlphaBlendRGBX32[S, blender.BlenderRGBXPre[S, order.XBGR32]](rbuf, blender.BlenderRGBXPre[S, order.XBGR32]{})
}
