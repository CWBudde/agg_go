package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt/blender"
)

// Core 64-bit pixel format: depends only on the blender policy (B) and space (S).
// The blender owns channel order and storage (premul/plain).
type PixFmtAlphaBlendRGBA64[B blender.RGBABlender16[S, O], S color.Space, O order.RGBAOrder] struct {
	rbuf    *buffer.RenderingBufferU8
	blender B
}

// New (generic)
func NewPixFmtAlphaBlendRGBA64[B blender.RGBABlender16[S, O], S color.Space, O order.RGBAOrder](
	r *buffer.RenderingBufferU8, b B,
) *PixFmtAlphaBlendRGBA64[B, S, O] {
	return &PixFmtAlphaBlendRGBA64[B, S, O]{rbuf: r, blender: b}
}

func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) Width() int    { return pf.rbuf.Width() }
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) Height() int   { return pf.rbuf.Height() }
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) PixWidth() int { return 8 } // 4*16-bit

// GetPixel returns a *plain* RGBA16 (regardless of framebuffer storage).
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) GetPixel(x, y int) color.RGBA16[S] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA16[S]{}
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 8
	if off+7 >= len(row) {
		return color.RGBA16[S]{}
	}
	r, g, b, a := pf.blender.GetPlain(row[off : off+8])
	return color.RGBA16[S]{R: r, G: g, B: b, A: a}
}

func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) Pixel(x, y int) color.RGBA16[S] {
	return pf.GetPixel(x, y)
}

// CopyPixel writes a *plain* RGBA16 (blender converts as needed).
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) CopyPixel(x, y int, c color.RGBA16[S]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 8
	if off+7 >= len(row) {
		return
	}
	pf.blender.SetPlain(row[off:off+8], c.R, c.G, c.B, c.A)
}

// BlendPixel blends a *plain* source into the framebuffer with coverage (8-bit -> 16-bit).
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) BlendPixel(x, y int, c color.RGBA16[S], cover8 basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.IsTransparent() {
		return
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 8
	if off+7 >= len(row) {
		return
	}
	cover16 := toCover16(cover8)
	pf.blender.BlendPix(row[off:off+8], c.R, c.G, c.B, c.A, cover16)
}

// CopyHline copies a horizontal run.
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) CopyHline(x, y, length int, c color.RGBA16[S]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU8(pf.rbuf, y)
	p := x * 8
	for i := 0; i < length; i++ {
		if p+7 >= len(row) {
			break
		}
		pf.blender.SetPlain(row[p:p+8], c.R, c.G, c.B, c.A)
		p += 8
	}
}

// BlendHline blends a horizontal run with uniform coverage.
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) BlendHline(x, y, length int, c color.RGBA16[S], cover8 basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.IsTransparent() {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU8(pf.rbuf, y)
	cover16 := toCover16(cover8)
	p := x * 8
	for i := 0; i < length; i++ {
		if p+7 >= len(row) {
			break
		}
		pf.blender.BlendPix(row[p:p+8], c.R, c.G, c.B, c.A, cover16)
		p += 8
	}
}

// BlendSolidHspan blends a horizontal span with per-pixel coverage.
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) BlendSolidHspan(x, y, length int, c color.RGBA16[S], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.IsTransparent() {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU8(pf.rbuf, y)

	if covers == nil {
		// Full cover
		pf.BlendHline(x, y, length, c, 255)
		return
	}

	p := x * 8
	for i := 0; i < length && i < len(covers); i++ {
		if covers[i] != 0 && p+7 < len(row) {
			pf.blender.BlendPix(row[p:p+8], c.R, c.G, c.B, c.A, toCover16(covers[i]))
		}
		p += 8
	}
}

// Clear fills the whole buffer with a solid plain color.
func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) Clear(c color.RGBA16[S]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		p := 0
		for x := 0; x < pf.Width(); x++ {
			if p+7 >= len(row) {
				break
			}
			pf.blender.SetPlain(row[p:p+8], c.R, c.G, c.B, c.A)
			p += 8
		}
	}
}

func (pf *PixFmtAlphaBlendRGBA64[B, S, O]) Fill(c color.RGBA16[S]) { pf.Clear(c) }

// Helpers
func toCover16(c basics.Int8u) basics.Int16u { return basics.Int16u(c) * 257 }

// -----------------------------------------------------------------------------
// Friendly aliases/ctors (user chooses format by ctor, like AGG)
// -----------------------------------------------------------------------------

// Premultiplied framebuffer, plain-src blending (AGG blender_rgba16)
type (
	PixFmtRGBA64[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16[S, order.RGBA], S, order.RGBA]
	PixFmtBGRA64[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16[S, order.BGRA], S, order.BGRA]
	PixFmtARGB64[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16[S, order.ARGB], S, order.ARGB]
	PixFmtABGR64[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16[S, order.ABGR], S, order.ABGR]

	// Premultiplied framebuffer, premul-src style blending (AGG blender_rgba16_pre)
	PixFmtRGBA64Pre[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Pre[S, order.RGBA], S, order.RGBA]
	PixFmtBGRA64Pre[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Pre[S, order.BGRA], S, order.BGRA]
	PixFmtARGB64Pre[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Pre[S, order.ARGB], S, order.ARGB]
	PixFmtABGR64Pre[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Pre[S, order.ABGR], S, order.ABGR]

	// Plain framebuffer (AGG blender_rgba16_plain)
	PixFmtRGBA64Plain[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Plain[S, order.RGBA], S, order.RGBA]
	PixFmtBGRA64Plain[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Plain[S, order.BGRA], S, order.BGRA]
	PixFmtARGB64Plain[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Plain[S, order.ARGB], S, order.ARGB]
	PixFmtABGR64Plain[S color.Space] = PixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Plain[S, order.ABGR], S, order.ABGR]
)

//////////////////////////////////////////////////////////////////////////////////////
// Constructors
//////////////////////////////////////////////////////////////////////////////////////

// Constructors for RGBA pixel formats
func NewPixFmtRGBA64[S color.Space](r *buffer.RenderingBufferU8) *PixFmtRGBA64[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16[S, order.RGBA], S, order.RGBA](r, blender.BlenderRGBA16[S, order.RGBA]{})
}

func NewPixFmtBGRA64[S color.Space](r *buffer.RenderingBufferU8) *PixFmtBGRA64[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16[S, order.BGRA], S, order.BGRA](r, blender.BlenderRGBA16[S, order.BGRA]{})
}

func NewPixFmtARGB64[S color.Space](r *buffer.RenderingBufferU8) *PixFmtARGB64[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16[S, order.ARGB], S, order.ARGB](r, blender.BlenderRGBA16[S, order.ARGB]{})
}

func NewPixFmtABGR64[S color.Space](r *buffer.RenderingBufferU8) *PixFmtABGR64[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16[S, order.ABGR], S, order.ABGR](r, blender.BlenderRGBA16[S, order.ABGR]{})
}

// Constructors for RGBA pixel formats (premultiplied)
func NewPixFmtRGBA64Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtRGBA64Pre[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Pre[S, order.RGBA], S, order.RGBA](r, blender.BlenderRGBA16Pre[S, order.RGBA]{})
}

func NewPixFmtBGRA64Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtBGRA64Pre[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Pre[S, order.BGRA], S, order.BGRA](r, blender.BlenderRGBA16Pre[S, order.BGRA]{})
}

func NewPixFmtARGB64Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtARGB64Pre[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Pre[S, order.ARGB], S, order.ARGB](r, blender.BlenderRGBA16Pre[S, order.ARGB]{})
}

func NewPixFmtABGR64Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtABGR64Pre[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Pre[S, order.ABGR], S, order.ABGR](r, blender.BlenderRGBA16Pre[S, order.ABGR]{})
}

// Constructors for RGBA pixel formats (plain)
func NewPixFmtRGBA64Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtRGBA64Plain[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Plain[S, order.RGBA], S, order.RGBA](r, blender.BlenderRGBA16Plain[S, order.RGBA]{})
}

func NewPixFmtBGRA64Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtBGRA64Plain[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Plain[S, order.BGRA], S, order.BGRA](r, blender.BlenderRGBA16Plain[S, order.BGRA]{})
}

func NewPixFmtARGB64Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtARGB64Plain[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Plain[S, order.ARGB], S, order.ARGB](r, blender.BlenderRGBA16Plain[S, order.ARGB]{})
}

func NewPixFmtABGR64Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtABGR64Plain[S] {
	return NewPixFmtAlphaBlendRGBA64[blender.BlenderRGBA16Plain[S, order.ABGR], S, order.ABGR](r, blender.BlenderRGBA16Plain[S, order.ABGR]{})
}

// Constructors for RGBA pixel formats (linear)

func NewPixFmtRGBA64Linear(r *buffer.RenderingBufferU8) *PixFmtRGBA64[color.Linear] {
	return NewPixFmtRGBA64[color.Linear](r)
}

func NewPixFmtBGRA64Linear(r *buffer.RenderingBufferU8) *PixFmtBGRA64[color.Linear] {
	return NewPixFmtBGRA64[color.Linear](r)
}

func NewPixFmtARGB64Linear(r *buffer.RenderingBufferU8) *PixFmtARGB64[color.Linear] {
	return NewPixFmtARGB64[color.Linear](r)
}

func NewPixFmtABGR64Linear(r *buffer.RenderingBufferU8) *PixFmtABGR64[color.Linear] {
	return NewPixFmtABGR64[color.Linear](r)
}

// Constructors for RGBA pixel formats (linear, premultiplied)

func NewPixFmtRGBA64PreLinear(r *buffer.RenderingBufferU8) *PixFmtRGBA64Pre[color.Linear] {
	return NewPixFmtRGBA64Pre[color.Linear](r)
}

func NewPixFmtBGRA64PreLinear(r *buffer.RenderingBufferU8) *PixFmtBGRA64Pre[color.Linear] {
	return NewPixFmtBGRA64Pre[color.Linear](r)
}

func NewPixFmtARGB64PreLinear(r *buffer.RenderingBufferU8) *PixFmtARGB64Pre[color.Linear] {
	return NewPixFmtARGB64Pre[color.Linear](r)
}

func NewPixFmtABGR64PreLinear(r *buffer.RenderingBufferU8) *PixFmtABGR64Pre[color.Linear] {
	return NewPixFmtABGR64Pre[color.Linear](r)
}

// Constructors for RGBA pixel formats (linear, plain)

func NewPixFmtRGBA64PlainLinear(r *buffer.RenderingBufferU8) *PixFmtRGBA64Plain[color.Linear] {
	return NewPixFmtRGBA64Plain[color.Linear](r)
}

func NewPixFmtBGRA64PlainLinear(r *buffer.RenderingBufferU8) *PixFmtBGRA64Plain[color.Linear] {
	return NewPixFmtBGRA64Plain[color.Linear](r)
}

func NewPixFmtARGB64PlainLinear(r *buffer.RenderingBufferU8) *PixFmtARGB64Plain[color.Linear] {
	return NewPixFmtARGB64Plain[color.Linear](r)
}

func NewPixFmtABGR64PlainLinear(r *buffer.RenderingBufferU8) *PixFmtABGR64Plain[color.Linear] {
	return NewPixFmtABGR64Plain[color.Linear](r)
}
