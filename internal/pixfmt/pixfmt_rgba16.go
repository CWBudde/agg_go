package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt/blender"
)

// pixWidth16 is bytes per pixel for RGBA16 (4 channels × 2 bytes).
const pixWidth16 = 8

// cover8to16 scales an 8-bit coverage value (0–255) to 16-bit (0–65535).
// 255 × 257 = 65535 exactly.
func cover8to16(c basics.Int8u) basics.Int16u {
	return basics.Int16u(c) * 257
}

// PixFmtAlphaBlendRGBA16 is the 16-bit RGBA pixel format with alpha blending.
// Each pixel occupies 8 bytes (4 channels × 2 bytes, little-endian), stored in
// a standard RenderingBufferU8 with stride = width*8.
type PixFmtAlphaBlendRGBA16[S color.Space, B blender.RGBABlender16[S]] struct {
	rbuf    *buffer.RenderingBufferU8
	blender B
}

// NewPixFmtAlphaBlendRGBA16 creates a new RGBA16 pixel format.
func NewPixFmtAlphaBlendRGBA16[S color.Space, B blender.RGBABlender16[S]](rbuf *buffer.RenderingBufferU8, b B) *PixFmtAlphaBlendRGBA16[S, B] {
	return &PixFmtAlphaBlendRGBA16[S, B]{rbuf: rbuf, blender: b}
}

func (pf *PixFmtAlphaBlendRGBA16[S, B]) Width() int    { return pf.rbuf.Width() }
func (pf *PixFmtAlphaBlendRGBA16[S, B]) Height() int   { return pf.rbuf.Height() }
func (pf *PixFmtAlphaBlendRGBA16[S, B]) PixWidth() int { return pixWidth16 }
func (pf *PixFmtAlphaBlendRGBA16[S, B]) Stride() int   { return pf.rbuf.Stride() }

// RowData returns the raw row bytes for transfer operations.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) RowData(y int) []basics.Int8u {
	if y < 0 || y >= pf.Height() {
		return nil
	}
	return buffer.RowU8(pf.rbuf, y)
}

// Pixel returns the (demultiplied) color at (x, y).
func (pf *PixFmtAlphaBlendRGBA16[S, B]) Pixel(x, y int) color.RGBA16[S] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA16[S]{}
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * pixWidth16
	if off+pixWidth16 > len(row) {
		return color.RGBA16[S]{}
	}
	r, g, b, a := pf.blender.GetPlain(row[off : off+pixWidth16])
	return color.RGBA16[S]{R: r, G: g, B: b, A: a}
}

// CopyPixel writes a pixel without blending (uses blender's SetPlain).
func (pf *PixFmtAlphaBlendRGBA16[S, B]) CopyPixel(x, y int, c color.RGBA16[S]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * pixWidth16
	if off+pixWidth16 > len(row) {
		return
	}
	pf.blender.SetPlain(row[off:off+pixWidth16], c.R, c.G, c.B, c.A)
}

// BlendPixel blends a pixel with 8-bit coverage (converted to 16-bit internally).
func (pf *PixFmtAlphaBlendRGBA16[S, B]) BlendPixel(x, y int, c color.RGBA16[S], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.IsTransparent() {
		return
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * pixWidth16
	if off+pixWidth16 > len(row) {
		return
	}
	pf.blender.BlendPix(row[off:off+pixWidth16], c.R, c.G, c.B, c.A, cover8to16(cover))
}

// CopyHline copies a horizontal line without blending.
// Uses RawRGBA16Order fast path when the blender supports it.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) CopyHline(x, y, length int, c color.RGBA16[S]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)
	off := x * pixWidth16
	for i := 0; i < length; i++ {
		pf.blender.SetPlain(row[off:off+pixWidth16], c.R, c.G, c.B, c.A)
		off += pixWidth16
	}
}

// BlendHline blends a horizontal line with uniform 8-bit coverage.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) BlendHline(x, y, length int, c color.RGBA16[S], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.IsTransparent() {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	// Full coverage + opaque → direct copy.
	if cover == basics.CoverFull && c.IsOpaque() {
		pf.CopyHline(x, y, length, c)
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	cover16 := cover8to16(cover)
	off := x * pixWidth16
	for i := 0; i < length; i++ {
		pf.blender.BlendPix(row[off:off+pixWidth16], c.R, c.G, c.B, c.A, cover16)
		off += pixWidth16
	}
}

// BlendSolidHspan blends a horizontal span with per-pixel 8-bit coverage.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) BlendSolidHspan(x, y, length int, c color.RGBA16[S], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.IsTransparent() {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)
	off := x * pixWidth16
	if covers == nil {
		for i := 0; i < length; i++ {
			pf.blender.BlendPix(row[off:off+pixWidth16], c.R, c.G, c.B, c.A, 0xFFFF)
			off += pixWidth16
		}
	} else {
		for i := 0; i < length && i < len(covers); i++ {
			if covers[i] > 0 {
				pf.blender.BlendPix(row[off:off+pixWidth16], c.R, c.G, c.B, c.A, cover8to16(covers[i]))
			}
			off += pixWidth16
		}
	}
}

// BlendColorHspan blends a horizontal span of per-pixel colors.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) BlendColorHspan(x, y, length int, colors []color.RGBA16[S], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	for i := 0; i < length; i++ {
		c := colors[i%len(colors)]
		if c.IsTransparent() {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x+i, y, c, cvr)
	}
}

// CopyColorHspan copies a horizontal span of per-pixel colors without blending.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) CopyColorHspan(x, y, length int, colors []color.RGBA16[S]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	for i := 0; i < length; i++ {
		pf.CopyPixel(x+i, y, colors[i%len(colors)])
	}
}

// Clear fills the entire buffer with c.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) Clear(c color.RGBA16[S]) {
	w, h := pf.Width(), pf.Height()
	if w <= 0 || h <= 0 {
		return
	}
	for y := 0; y < h; y++ {
		pf.CopyHline(0, y, w, c)
	}
}

// Fill is an alias for Clear.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) Fill(c color.RGBA16[S]) {
	pf.Clear(c)
}

// ApplyGammaDir applies a 16-bit gamma function to RGB channels of the entire buffer.
func (pf *PixFmtAlphaBlendRGBA16[S, B]) ApplyGammaDir(gamma func(basics.Int16u) basics.Int16u) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			off := x * pixWidth16
			if off+pixWidth16 > len(row) {
				break
			}
			r, g, b, a := pf.blender.GetPlain(row[off : off+pixWidth16])
			pf.blender.SetPlain(row[off:off+pixWidth16], gamma(r), gamma(g), gamma(b), a)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// Type aliases
////////////////////////////////////////////////////////////////////////////////

type (
	// Plain→premul blending (AGG blender_rgba)
	PixFmtRGBA64[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16[S, order.RGBA]]
	PixFmtBGRA64[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16[S, order.BGRA]]
	PixFmtARGB64[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16[S, order.ARGB]]
	PixFmtABGR64[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16[S, order.ABGR]]

	// Premul→premul blending (AGG blender_rgba_pre)
	PixFmtRGBA64Pre[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16Pre[S, order.RGBA]]
	PixFmtBGRA64Pre[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16Pre[S, order.BGRA]]
	PixFmtARGB64Pre[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16Pre[S, order.ARGB]]
	PixFmtABGR64Pre[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16Pre[S, order.ABGR]]

	// Plain→plain blending (AGG blender_rgba_plain)
	PixFmtRGBA64Plain[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16Plain[S, order.RGBA]]
	PixFmtBGRA64Plain[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16Plain[S, order.BGRA]]
	PixFmtARGB64Plain[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16Plain[S, order.ARGB]]
	PixFmtABGR64Plain[S color.Space] = PixFmtAlphaBlendRGBA16[S, blender.BlenderRGBA16Plain[S, order.ABGR]]
)

////////////////////////////////////////////////////////////////////////////////
// Generic constructors
////////////////////////////////////////////////////////////////////////////////

func NewPixFmtRGBA64[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtRGBA64[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16[S, order.RGBA]{})
}

func NewPixFmtBGRA64[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtBGRA64[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16[S, order.BGRA]{})
}

func NewPixFmtARGB64[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtARGB64[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16[S, order.ARGB]{})
}

func NewPixFmtABGR64[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtABGR64[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16[S, order.ABGR]{})
}

func NewPixFmtRGBA64Pre[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtRGBA64Pre[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16Pre[S, order.RGBA]{})
}

func NewPixFmtBGRA64Pre[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtBGRA64Pre[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16Pre[S, order.BGRA]{})
}

func NewPixFmtARGB64Pre[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtARGB64Pre[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16Pre[S, order.ARGB]{})
}

func NewPixFmtABGR64Pre[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtABGR64Pre[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16Pre[S, order.ABGR]{})
}

func NewPixFmtRGBA64Plain[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtRGBA64Plain[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16Plain[S, order.RGBA]{})
}

func NewPixFmtBGRA64Plain[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtBGRA64Plain[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16Plain[S, order.BGRA]{})
}

func NewPixFmtARGB64Plain[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtARGB64Plain[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16Plain[S, order.ARGB]{})
}

func NewPixFmtABGR64Plain[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtABGR64Plain[S] {
	return NewPixFmtAlphaBlendRGBA16[S](rbuf, blender.BlenderRGBA16Plain[S, order.ABGR]{})
}

////////////////////////////////////////////////////////////////////////////////
// Concrete (linear) constructors
////////////////////////////////////////////////////////////////////////////////

func NewPixFmtRGBA64Linear(rbuf *buffer.RenderingBufferU8) *PixFmtRGBA64[color.Linear] {
	return NewPixFmtRGBA64[color.Linear](rbuf)
}

func NewPixFmtBGRA64Linear(rbuf *buffer.RenderingBufferU8) *PixFmtBGRA64[color.Linear] {
	return NewPixFmtBGRA64[color.Linear](rbuf)
}

func NewPixFmtARGB64Linear(rbuf *buffer.RenderingBufferU8) *PixFmtARGB64[color.Linear] {
	return NewPixFmtARGB64[color.Linear](rbuf)
}

func NewPixFmtABGR64Linear(rbuf *buffer.RenderingBufferU8) *PixFmtABGR64[color.Linear] {
	return NewPixFmtABGR64[color.Linear](rbuf)
}

func NewPixFmtRGBA64PreLinear(rbuf *buffer.RenderingBufferU8) *PixFmtRGBA64Pre[color.Linear] {
	return NewPixFmtRGBA64Pre[color.Linear](rbuf)
}

func NewPixFmtBGRA64PreLinear(rbuf *buffer.RenderingBufferU8) *PixFmtBGRA64Pre[color.Linear] {
	return NewPixFmtBGRA64Pre[color.Linear](rbuf)
}

func NewPixFmtARGB64PreLinear(rbuf *buffer.RenderingBufferU8) *PixFmtARGB64Pre[color.Linear] {
	return NewPixFmtARGB64Pre[color.Linear](rbuf)
}

func NewPixFmtABGR64PreLinear(rbuf *buffer.RenderingBufferU8) *PixFmtABGR64Pre[color.Linear] {
	return NewPixFmtABGR64Pre[color.Linear](rbuf)
}

func NewPixFmtRGBA64PlainLinear(rbuf *buffer.RenderingBufferU8) *PixFmtRGBA64Plain[color.Linear] {
	return NewPixFmtRGBA64Plain[color.Linear](rbuf)
}

func NewPixFmtBGRA64PlainLinear(rbuf *buffer.RenderingBufferU8) *PixFmtBGRA64Plain[color.Linear] {
	return NewPixFmtBGRA64Plain[color.Linear](rbuf)
}

func NewPixFmtARGB64PlainLinear(rbuf *buffer.RenderingBufferU8) *PixFmtARGB64Plain[color.Linear] {
	return NewPixFmtARGB64Plain[color.Linear](rbuf)
}

func NewPixFmtABGR64PlainLinear(rbuf *buffer.RenderingBufferU8) *PixFmtABGR64Plain[color.Linear] {
	return NewPixFmtABGR64Plain[color.Linear](rbuf)
}
