package pixfmt

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
)

// compositeRGBA128Blender is the minimal contract the composite float pixfmt
// needs from its embedded blender: composite one pixel and report its operator.
type compositeRGBA128Blender interface {
	BlendPix(dst []float32, r, g, b, a, cover float32)
	GetOp() blender.CompOp
}

// straightSpanBlenderF32 is the optional fast-path interface a float composite
// blender may implement to blend a whole solid span in one concrete call instead
// of one interface dispatch per pixel. Only the straight-alpha
// CompositeBlenderRGBA128Plain implements it; the premultiplied-source Pre
// blender falls back to per-pixel BlendPix. It is the float twin of
// straightSpanBlender and is bit-for-bit identical to the per-pixel path.
type straightSpanBlenderF32 interface {
	BlendSolidSpanStraight(dst []float32, r, g, b, a float32, covers []basics.Int8u, count int)
}

// PixFmtCompositeRGBA128 is the float (128-bit, 4 x float32) twin of
// PixFmtCompositeRGBA and the structural analogue of AGG's
// pixfmt_custom_blend_rgba instantiated with the float rgba32 color type.
//
// Like the 8-bit composite pixfmt it delegates every write to a Porter-Duff / SVG
// composite operator selected by the embedded blender. The buffer storage
// convention follows the blender: the default constructor uses the straight-alpha
// CompositeBlenderRGBA128Plain (premultiply on read, demultiply back to straight
// on write — the convention masks/ToGoImage/cross-backend all read), while
// NewPixFmtCompositeRGBA128Pre keeps genuinely premultiplied storage. Coverage
// arrives as basics.Int8u (0..255) per the renderer.PixelFormat contract and is
// normalised to [0,1] for the float blender.
type PixFmtCompositeRGBA128[CS color.Space, O order.RGBAOrder] struct {
	rbuf          *buffer.RenderingBufferF32
	blender       compositeRGBA128Blender
	premultiplied bool
}

// NewPixFmtCompositeRGBA128 creates a composite float pixfmt that expects
// straight-alpha source colors (like AGG's comp_op_adaptor_rgba).
func NewPixFmtCompositeRGBA128[CS color.Space, O order.RGBAOrder](rbuf *buffer.RenderingBufferF32, op blender.CompOp) *PixFmtCompositeRGBA128[CS, O] {
	return &PixFmtCompositeRGBA128[CS, O]{
		rbuf:    rbuf,
		blender: blender.NewCompositeBlenderRGBA128Plain[CS, O](op),
	}
}

// NewPixFmtCompositeRGBA128Pre creates a composite float pixfmt whose source
// input is already premultiplied (like AGG's comp_op_adaptor_rgba_pre).
func NewPixFmtCompositeRGBA128Pre[CS color.Space, O order.RGBAOrder](rbuf *buffer.RenderingBufferF32, op blender.CompOp) *PixFmtCompositeRGBA128[CS, O] {
	return &PixFmtCompositeRGBA128[CS, O]{
		rbuf:          rbuf,
		blender:       blender.NewCompositeBlenderRGBA128Pre[CS, O](op),
		premultiplied: true,
	}
}

// Basic properties

func (pf *PixFmtCompositeRGBA128[CS, O]) Width() int    { return pf.rbuf.Width() }
func (pf *PixFmtCompositeRGBA128[CS, O]) Height() int   { return pf.rbuf.Height() }
func (pf *PixFmtCompositeRGBA128[CS, O]) PixWidth() int { return 16 }
func (pf *PixFmtCompositeRGBA128[CS, O]) Stride() int   { return pf.rbuf.Stride() }

// RowPtr returns the float32 slice backing row y.
func (pf *PixFmtCompositeRGBA128[CS, O]) RowPtr(y int) []float32 {
	return buffer.RowF32(pf.rbuf, y)
}

func (pf *PixFmtCompositeRGBA128[CS, O]) pixelSlice(x, y int) []float32 {
	row := buffer.RowF32(pf.rbuf, y)
	off := x * 4
	if off < 0 || off+4 > len(row) {
		return nil
	}
	return row[off : off+4]
}

// GetPixel reads the stored pixel in channel order O (straight alpha for the
// default Plain blender; premultiplied for the Pre variant).
func (pf *PixFmtCompositeRGBA128[CS, O]) GetPixel(x, y int) color.RGBA32[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA32[CS]{}
	}
	px := pf.pixelSlice(x, y)
	if px == nil {
		return color.RGBA32[CS]{}
	}
	var o O
	return color.RGBA32[CS]{
		R: px[o.IdxR()],
		G: px[o.IdxG()],
		B: px[o.IdxB()],
		A: px[o.IdxA()],
	}
}

// Pixel is an alias for GetPixel (satisfies renderer.PixelFormat).
func (pf *PixFmtCompositeRGBA128[CS, O]) Pixel(x, y int) color.RGBA32[CS] {
	return pf.GetPixel(x, y)
}

// CopyPixel writes c directly in order O, bypassing the composite operator.
func (pf *PixFmtCompositeRGBA128[CS, O]) CopyPixel(x, y int, c color.RGBA32[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	if px := pf.pixelSlice(x, y); px != nil {
		var o O
		px[o.IdxR()] = c.R
		px[o.IdxG()] = c.G
		px[o.IdxB()] = c.B
		px[o.IdxA()] = c.A
	}
}

// BlendPixel applies the active composite operator to a single pixel.
func (pf *PixFmtCompositeRGBA128[CS, O]) BlendPixel(x, y int, c color.RGBA32[CS], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	if px := pf.pixelSlice(x, y); px != nil {
		pf.blender.BlendPix(px, c.R, c.G, c.B, c.A, coverToF32(cover))
	}
}

// Line operations

// CopyHline writes a solid horizontal run without compositing.
func (pf *PixFmtCompositeRGBA128[CS, O]) CopyHline(x, y, length int, c color.RGBA32[CS]) {
	for i := range max(0, length) {
		pf.CopyPixel(x+i, y, c)
	}
}

// BlendHline composites a solid horizontal run with uniform coverage.
func (pf *PixFmtCompositeRGBA128[CS, O]) BlendHline(x, y, length int, c color.RGBA32[CS], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	startX := Max(0, x)
	endX := Min(x+length, pf.Width())
	if startX >= endX {
		return
	}
	row := pf.RowPtr(y)

	// Bit-exact span fast path (uniform full cover only — a partial uniform cover
	// is rare here and falls through to the per-pixel bridge), mirroring the 8-bit
	// PixFmtCompositeRGBA.BlendHline.
	if cover == basics.CoverFull {
		if sb, ok := pf.blender.(straightSpanBlenderF32); ok {
			sb.BlendSolidSpanStraight(row[startX*4:], c.R, c.G, c.B, c.A, nil, endX-startX)
			return
		}
	}

	fc := coverToF32(cover)
	for i := startX; i < endX; i++ {
		p := i * 4
		if p+4 <= len(row) {
			pf.blender.BlendPix(row[p:p+4], c.R, c.G, c.B, c.A, fc)
		}
	}
}

// CopyVline writes a solid vertical run without compositing.
func (pf *PixFmtCompositeRGBA128[CS, O]) CopyVline(x, y, length int, c color.RGBA32[CS]) {
	for i := range max(0, length) {
		pf.CopyPixel(x, y+i, c)
	}
}

// BlendVline composites a solid vertical run with uniform coverage.
func (pf *PixFmtCompositeRGBA128[CS, O]) BlendVline(x, y, length int, c color.RGBA32[CS], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}
	startY := Max(0, y)
	endY := Min(y+length, pf.Height())
	for i := startY; i < endY; i++ {
		if px := pf.pixelSlice(x, i); px != nil {
			pf.blender.BlendPix(px, c.R, c.G, c.B, c.A, coverToF32(cover))
		}
	}
}

// Rectangle operations

// CopyBar writes a solid rectangle without compositing.
func (pf *PixFmtCompositeRGBA128[CS, O]) CopyBar(x1, y1, x2, y2 int, c color.RGBA32[CS]) {
	x1, y1, x2, y2 = normalizeBar(x1, y1, x2, y2)
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

// BlendBar composites a solid rectangle with uniform coverage.
func (pf *PixFmtCompositeRGBA128[CS, O]) BlendBar(x1, y1, x2, y2 int, c color.RGBA32[CS], cover basics.Int8u) {
	x1, y1, x2, y2 = normalizeBar(x1, y1, x2, y2)
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

// Span operations

// BlendSolidHspan composites a solid color along a horizontal span with
// per-pixel coverage.
func (pf *PixFmtCompositeRGBA128[CS, O]) BlendSolidHspan(x, y, length int, c color.RGBA32[CS], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	startX := Max(0, x)
	endX := Min(x+length, pf.Width())
	if startX >= endX {
		return
	}
	row := pf.RowPtr(y)

	// Bit-exact span fast path. covers is aligned to the first blended pixel
	// (startX), matching the per-pixel loop's covers[i-x] indexing. Require a
	// full-length cover slice so the kernel never reads past it (the per-pixel loop
	// tolerates a short slice; keep that edge case on the scalar path). Mirrors the
	// 8-bit PixFmtCompositeRGBA.BlendSolidHspan.
	if covers == nil || len(covers) >= length {
		if sb, ok := pf.blender.(straightSpanBlenderF32); ok {
			var cv []basics.Int8u
			if covers != nil {
				cv = covers[startX-x:]
			}
			sb.BlendSolidSpanStraight(row[startX*4:], c.R, c.G, c.B, c.A, cv, endX-startX)
			return
		}
	}

	for i := startX; i < endX; i++ {
		p := i * 4
		if p+4 > len(row) {
			continue
		}
		if covers == nil {
			pf.blender.BlendPix(row[p:p+4], c.R, c.G, c.B, c.A, 1.0)
			continue
		}
		ci := i - x
		if ci < len(covers) && covers[ci] > 0 {
			pf.blender.BlendPix(row[p:p+4], c.R, c.G, c.B, c.A, coverToF32(covers[ci]))
		}
	}
}

// BlendSolidVspan composites a solid color along a vertical span with per-pixel
// coverage.
func (pf *PixFmtCompositeRGBA128[CS, O]) BlendSolidVspan(x, y, length int, c color.RGBA32[CS], covers []basics.Int8u) {
	if covers == nil {
		for i := range max(0, length) {
			pf.BlendPixel(x, y+i, c, basics.CoverFull)
		}
		return
	}
	for i := 0; i < length && i < len(covers); i++ {
		if covers[i] > 0 {
			pf.BlendPixel(x, y+i, c, covers[i])
		}
	}
}

// CopyColorHspan writes a horizontal span of already-expanded colors.
func (pf *PixFmtCompositeRGBA128[CS, O]) CopyColorHspan(x, y, length int, colors []color.RGBA32[CS]) {
	for i := 0; i < length && i < len(colors); i++ {
		pf.CopyPixel(x+i, y, colors[i])
	}
}

// BlendColorHspan composites a horizontal span of per-pixel colors.
func (pf *PixFmtCompositeRGBA128[CS, O]) BlendColorHspan(x, y, length int, colors []color.RGBA32[CS], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		actualCover := cover
		if i < len(covers) {
			actualCover = covers[i]
		}
		pf.BlendPixel(x+i, y, colors[i], actualCover)
	}
}

// CopyColorVspan writes a vertical span of already-expanded colors.
func (pf *PixFmtCompositeRGBA128[CS, O]) CopyColorVspan(x, y, length int, colors []color.RGBA32[CS]) {
	for i := 0; i < length && i < len(colors); i++ {
		pf.CopyPixel(x, y+i, colors[i])
	}
}

// BlendColorVspan composites a vertical span of per-pixel colors.
func (pf *PixFmtCompositeRGBA128[CS, O]) BlendColorVspan(x, y, length int, colors []color.RGBA32[CS], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		actualCover := cover
		if i < len(covers) {
			actualCover = covers[i]
		}
		pf.BlendPixel(x, y+i, colors[i], actualCover)
	}
}

// Clear fills the entire buffer with c without compositing.
func (pf *PixFmtCompositeRGBA128[CS, O]) Clear(c color.RGBA32[CS]) {
	for y := 0; y < pf.Height(); y++ {
		pf.CopyHline(0, y, pf.Width(), c)
	}
}

// Fill is an alias for Clear.
func (pf *PixFmtCompositeRGBA128[CS, O]) Fill(c color.RGBA32[CS]) {
	pf.Clear(c)
}

// SetCompOp switches the active composite operator while preserving the source
// alpha convention chosen at construction time.
func (pf *PixFmtCompositeRGBA128[CS, O]) SetCompOp(op blender.CompOp) {
	if pf.premultiplied {
		pf.blender = blender.NewCompositeBlenderRGBA128Pre[CS, O](op)
		return
	}
	pf.blender = blender.NewCompositeBlenderRGBA128Plain[CS, O](op)
}

// GetCompOp returns the current composite operator.
func (pf *PixFmtCompositeRGBA128[CS, O]) GetCompOp() blender.CompOp {
	return pf.blender.GetOp()
}

// Concrete composite float pixel format aliases (Linear and sRGB, RGBA order).
type (
	PixFmtCompositeRGBA128Linear    = PixFmtCompositeRGBA128[color.Linear, order.RGBA]
	PixFmtCompositeRGBA128PreLinear = PixFmtCompositeRGBA128[color.Linear, order.RGBA]
	PixFmtCompositeSRGBA128         = PixFmtCompositeRGBA128[color.SRGB, order.RGBA]
)

// NewPixFmtCompositeRGBA128Linear creates a Linear/RGBA composite float pixfmt
// with straight-alpha source input.
func NewPixFmtCompositeRGBA128Linear(rbuf *buffer.RenderingBufferF32, op blender.CompOp) *PixFmtCompositeRGBA128Linear {
	return NewPixFmtCompositeRGBA128[color.Linear, order.RGBA](rbuf, op)
}

// NewPixFmtCompositeRGBA128PreLinear creates a Linear/RGBA composite float pixfmt
// whose source input is already premultiplied.
func NewPixFmtCompositeRGBA128PreLinear(rbuf *buffer.RenderingBufferF32, op blender.CompOp) *PixFmtCompositeRGBA128PreLinear {
	return NewPixFmtCompositeRGBA128Pre[color.Linear, order.RGBA](rbuf, op)
}
