package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

// PixFmtRGBA64 represents a 64-bit RGBA pixel format (16-bit per channel)
type PixFmtRGBA64[B any] struct {
	buf    *buffer.RenderingBufferU8
	stride int
}

// NewPixFmtRGBA64 creates a new 64-bit RGBA pixel format
func NewPixFmtRGBA64[B any](buf *buffer.RenderingBufferU8) *PixFmtRGBA64[B] {
	return &PixFmtRGBA64[B]{
		buf:    buf,
		stride: 8, // 4 channels * 2 bytes per channel
	}
}

// Width returns the width of the pixel format
func (pf *PixFmtRGBA64[B]) Width() int {
	return pf.buf.Width()
}

// Height returns the height of the pixel format
func (pf *PixFmtRGBA64[B]) Height() int {
	return pf.buf.Height()
}

// RowData returns a pointer to the row data at the given y coordinate
func (pf *PixFmtRGBA64[B]) RowData(y int) []basics.Int8u {
	return pf.buf.Row(y)
}

// MakePix creates a pixel pointer at the given coordinates
func (pf *PixFmtRGBA64[B]) MakePix(x, y int) []basics.Int16u {
	row := pf.buf.Row(y)
	// Convert to 16-bit slice
	ptr := x * pf.stride
	result := make([]basics.Int16u, 4)
	for i := 0; i < 4; i++ {
		result[i] = basics.Int16u(row[ptr+i*2]) | (basics.Int16u(row[ptr+i*2+1]) << 8)
	}
	return result
}

// Pixel returns the color at the given coordinates
func (pf *PixFmtRGBA64[B]) Pixel(x, y int) color.RGBA16[color.Linear] {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() {
		pix := pf.MakePix(x, y)
		return color.RGBA16[color.Linear]{
			R: pix[0], G: pix[1], B: pix[2], A: pix[3],
		}
	}
	return color.RGBA16[color.Linear]{}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtRGBA64[B]) CopyPixel(x, y int, c color.RGBA16[color.Linear]) {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() {
		row := pf.buf.Row(y)
		ptr := x * pf.stride
		// Write 16-bit values as little-endian bytes
		row[ptr+0] = basics.Int8u(c.R)
		row[ptr+1] = basics.Int8u(c.R >> 8)
		row[ptr+2] = basics.Int8u(c.G)
		row[ptr+3] = basics.Int8u(c.G >> 8)
		row[ptr+4] = basics.Int8u(c.B)
		row[ptr+5] = basics.Int8u(c.B >> 8)
		row[ptr+6] = basics.Int8u(c.A)
		row[ptr+7] = basics.Int8u(c.A >> 8)
	}
}

// BlendPixel blends a single pixel using the blender
func (pf *PixFmtRGBA64[B]) BlendPixel(x, y int, c color.RGBA16[color.Linear], cover basics.Int8u) {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() && !c.IsTransparent() {
		row := pf.buf.Row(y)
		ptr := x * pf.stride

		// Convert cover from 8-bit to 16-bit
		cover16 := basics.Int16u(cover) | (basics.Int16u(cover) << 8)

		var blender B
		switch any(blender).(type) {
		case BlenderRGBA16:
			BlendRGBA16Pixel(row[ptr:], c, cover16, BlenderRGBA16{})
		case BlenderRGBA16Pre:
			BlendRGBA16Pixel(row[ptr:], c, cover16, BlenderRGBA16Pre{})
		case BlenderRGBA16Plain:
			BlendRGBA16Pixel(row[ptr:], c, cover16, BlenderRGBA16Plain{})
		default:
			// Default to standard blending
			BlendRGBA16Pixel(row[ptr:], c, cover16, BlenderRGBA16{})
		}
	}
}

// CopyHline copies a horizontal line of pixels
func (pf *PixFmtRGBA64[B]) CopyHline(x, y, length int, c color.RGBA16[color.Linear]) {
	if y >= 0 && y < pf.Height() && length > 0 {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			for i := x1; i < x2; i++ {
				ptr := i * pf.stride
				row[ptr+0] = basics.Int8u(c.R)
				row[ptr+1] = basics.Int8u(c.R >> 8)
				row[ptr+2] = basics.Int8u(c.G)
				row[ptr+3] = basics.Int8u(c.G >> 8)
				row[ptr+4] = basics.Int8u(c.B)
				row[ptr+5] = basics.Int8u(c.B >> 8)
				row[ptr+6] = basics.Int8u(c.A)
				row[ptr+7] = basics.Int8u(c.A >> 8)
			}
		}
	}
}

// BlendHline blends a horizontal line of pixels
func (pf *PixFmtRGBA64[B]) BlendHline(x, y, length int, c color.RGBA16[color.Linear], cover basics.Int8u) {
	if y >= 0 && y < pf.Height() && length > 0 && !c.IsTransparent() {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			cover16 := basics.Int16u(cover) | (basics.Int16u(cover) << 8)

			var blender B
			switch any(blender).(type) {
			case BlenderRGBA16:
				for i := x1; i < x2; i++ {
					BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, BlenderRGBA16{})
				}
			case BlenderRGBA16Pre:
				for i := x1; i < x2; i++ {
					BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, BlenderRGBA16Pre{})
				}
			case BlenderRGBA16Plain:
				for i := x1; i < x2; i++ {
					BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, BlenderRGBA16Plain{})
				}
			default:
				for i := x1; i < x2; i++ {
					BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, BlenderRGBA16{})
				}
			}
		}
	}
}

// BlendSolidHspan blends a horizontal span with varying coverage
func (pf *PixFmtRGBA64[B]) BlendSolidHspan(x, y, length int, c color.RGBA16[color.Linear], covers []basics.Int8u) {
	if y >= 0 && y < pf.Height() && length > 0 && !c.IsTransparent() {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			coverOffset := x1 - x

			var blender B
			switch any(blender).(type) {
			case BlenderRGBA16:
				for i := x1; i < x2; i++ {
					if covers[coverOffset] > 0 {
						cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
						BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, BlenderRGBA16{})
					}
					coverOffset++
				}
			case BlenderRGBA16Pre:
				for i := x1; i < x2; i++ {
					if covers[coverOffset] > 0 {
						cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
						BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, BlenderRGBA16Pre{})
					}
					coverOffset++
				}
			case BlenderRGBA16Plain:
				for i := x1; i < x2; i++ {
					if covers[coverOffset] > 0 {
						cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
						BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, BlenderRGBA16Plain{})
					}
					coverOffset++
				}
			default:
				for i := x1; i < x2; i++ {
					if covers[coverOffset] > 0 {
						cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
						BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, BlenderRGBA16{})
					}
					coverOffset++
				}
			}
		}
	}
}

// RGBA16 Blender definitions

// BlenderRGBA16 implements standard RGBA blending for 16-bit values
type BlenderRGBA16 struct{}

// BlendPix blends a 16-bit RGBA pixel
func (bl BlenderRGBA16) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	alpha := color.RGBA16MultCover(a, cover)
	if alpha > 0 {
		// Read destination as 16-bit values
		dr := basics.Int16u(dst[0]) | (basics.Int16u(dst[1]) << 8)
		dg := basics.Int16u(dst[2]) | (basics.Int16u(dst[3]) << 8)
		db := basics.Int16u(dst[4]) | (basics.Int16u(dst[5]) << 8)
		da := basics.Int16u(dst[6]) | (basics.Int16u(dst[7]) << 8)

		// Blend
		dr = color.RGBA16Lerp(dr, r, alpha)
		dg = color.RGBA16Lerp(dg, g, alpha)
		db = color.RGBA16Lerp(db, b, alpha)
		da = color.RGBA16Prelerp(da, alpha, alpha)

		// Write back as little-endian bytes
		dst[0] = basics.Int8u(dr)
		dst[1] = basics.Int8u(dr >> 8)
		dst[2] = basics.Int8u(dg)
		dst[3] = basics.Int8u(dg >> 8)
		dst[4] = basics.Int8u(db)
		dst[5] = basics.Int8u(db >> 8)
		dst[6] = basics.Int8u(da)
		dst[7] = basics.Int8u(da >> 8)
	}
}

// BlenderRGBA16Pre implements premultiplied RGBA blending for 16-bit values
type BlenderRGBA16Pre struct{}

// BlendPix blends a premultiplied 16-bit RGBA pixel
func (bl BlenderRGBA16Pre) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	cr := color.RGBA16MultCover(r, cover)
	cg := color.RGBA16MultCover(g, cover)
	cb := color.RGBA16MultCover(b, cover)
	ca := color.RGBA16MultCover(a, cover)

	// Read destination as 16-bit values
	dr := basics.Int16u(dst[0]) | (basics.Int16u(dst[1]) << 8)
	dg := basics.Int16u(dst[2]) | (basics.Int16u(dst[3]) << 8)
	db := basics.Int16u(dst[4]) | (basics.Int16u(dst[5]) << 8)
	da := basics.Int16u(dst[6]) | (basics.Int16u(dst[7]) << 8)

	// Blend
	dr = color.RGBA16Prelerp(dr, cr, ca)
	dg = color.RGBA16Prelerp(dg, cg, ca)
	db = color.RGBA16Prelerp(db, cb, ca)
	da = color.RGBA16Prelerp(da, ca, ca)

	// Write back as little-endian bytes
	dst[0] = basics.Int8u(dr)
	dst[1] = basics.Int8u(dr >> 8)
	dst[2] = basics.Int8u(dg)
	dst[3] = basics.Int8u(dg >> 8)
	dst[4] = basics.Int8u(db)
	dst[5] = basics.Int8u(db >> 8)
	dst[6] = basics.Int8u(da)
	dst[7] = basics.Int8u(da >> 8)
}

// BlenderRGBA16Plain implements plain RGBA blending for 16-bit values
type BlenderRGBA16Plain struct{}

// BlendPix blends a plain 16-bit RGBA pixel
func (bl BlenderRGBA16Plain) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	alpha := color.RGBA16MultCover(a, cover)
	if alpha > 0 {
		// Read destination as 16-bit values
		dr := basics.Int16u(dst[0]) | (basics.Int16u(dst[1]) << 8)
		dg := basics.Int16u(dst[2]) | (basics.Int16u(dst[3]) << 8)
		db := basics.Int16u(dst[4]) | (basics.Int16u(dst[5]) << 8)
		da := basics.Int16u(dst[6]) | (basics.Int16u(dst[7]) << 8)

		// Premultiply destination for calculation
		pdr := color.RGBA16Multiply(dr, da)
		pdg := color.RGBA16Multiply(dg, da)
		pdb := color.RGBA16Multiply(db, da)

		// Blend in premultiplied space
		pdr = color.RGBA16Lerp(pdr, r, alpha)
		pdg = color.RGBA16Lerp(pdg, g, alpha)
		pdb = color.RGBA16Lerp(pdb, b, alpha)
		da = color.RGBA16Prelerp(da, alpha, alpha)

		// Demultiply result back to plain space
		if da > 0 {
			dr = basics.Int16u((uint32(pdr)*65535 + uint32(da)/2) / uint32(da))
			dg = basics.Int16u((uint32(pdg)*65535 + uint32(da)/2) / uint32(da))
			db = basics.Int16u((uint32(pdb)*65535 + uint32(da)/2) / uint32(da))
		} else {
			dr, dg, db = 0, 0, 0
		}

		// Write back as little-endian bytes
		dst[0] = basics.Int8u(dr)
		dst[1] = basics.Int8u(dr >> 8)
		dst[2] = basics.Int8u(dg)
		dst[3] = basics.Int8u(dg >> 8)
		dst[4] = basics.Int8u(db)
		dst[5] = basics.Int8u(db >> 8)
		dst[6] = basics.Int8u(da)
		dst[7] = basics.Int8u(da >> 8)
	}
}

// Helper function to blend a single RGBA16 pixel
func BlendRGBA16Pixel[B interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u)
}](dst []basics.Int8u, src color.RGBA16[color.Linear], cover basics.Int16u, blender B) {
	if !src.IsTransparent() {
		blender.BlendPix(dst, src.R, src.G, src.B, src.A, cover)
	}
}

// Concrete RGBA64 pixel format types for different component orders
type (
	PixFmtRGBA64Linear = PixFmtRGBA64[BlenderRGBA16]
	PixFmtRGBA64Pre    = PixFmtRGBA64[BlenderRGBA16Pre]
	PixFmtRGBA64Plain  = PixFmtRGBA64[BlenderRGBA16Plain]

	// Different byte orders (would need separate implementations)
	PixFmtARGB64Linear = PixFmtRGBA64[BlenderRGBA16] // TODO: Implement ARGB order
	PixFmtABGR64Linear = PixFmtRGBA64[BlenderRGBA16] // TODO: Implement ABGR order
	PixFmtBGRA64Linear = PixFmtRGBA64[BlenderRGBA16] // TODO: Implement BGRA order
)

// Helper functions to create specific RGBA64 pixel formats
func NewPixFmtRGBA64Linear(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Linear {
	return NewPixFmtRGBA64[BlenderRGBA16](buf)
}

func NewPixFmtRGBA64Pre(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Pre {
	return NewPixFmtRGBA64[BlenderRGBA16Pre](buf)
}

func NewPixFmtRGBA64Plain(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Plain {
	return NewPixFmtRGBA64[BlenderRGBA16Plain](buf)
}

// Utility functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
