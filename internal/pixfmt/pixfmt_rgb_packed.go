package pixfmt

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
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

// ─── PixFmtRGB555 ────────────────────────────────────────────────────────────

// PixFmtRGB555 is the Go equivalent of AGG's pixfmt_alpha_blend_rgb_packed for
// RGB555 (15-bit) packed pixel format. The color type is RGBA8[Linear] to match
// C++ rgba8 (blender_rgb555::color_type): alpha is used for blending but not stored in the 16-bit pixel.
type PixFmtRGB555[B blender.RGB16PackedBlender] struct {
	rbuf    *buffer.RenderingBufferU16
	blender B
}

// NewPixFmtRGB555 creates a new RGB555 pixel format over the given rendering buffer.
func NewPixFmtRGB555[B blender.RGB16PackedBlender](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtRGB555[B] {
	return &PixFmtRGB555[B]{rbuf: rbuf, blender: blender}
}

func (pf *PixFmtRGB555[B]) Width() int    { return pf.rbuf.Width() }
func (pf *PixFmtRGB555[B]) Height() int   { return pf.rbuf.Height() }
func (pf *PixFmtRGB555[B]) PixWidth() int { return 2 }

// Pixel returns the color at (x, y). Alpha is always 255 (packed formats are opaque).
func (pf *PixFmtRGB555[B]) Pixel(x, y int) color.RGBA8[color.Linear] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA8[color.Linear]{}
	}
	row := buffer.RowU16(pf.rbuf, y)
	r, g, b := pf.blender.UnpackPix(row[x])
	return color.RGBA8[color.Linear]{R: r, G: g, B: b, A: 255}
}

// CopyPixel writes c directly, bypassing alpha compositing.
func (pf *PixFmtRGB555[B]) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU16(pf.rbuf, y)
	row[x] = pf.blender.MakePix(c.R, c.G, c.B)
}

// BlendPixel blends a pixel with coverage.
func (pf *PixFmtRGB555[B]) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.A == 0 {
		return
	}
	row := buffer.RowU16(pf.rbuf, y)
	pf.blender.BlendPix(&row[x], c.R, c.G, c.B, c.A, cover)
}

// CopyHline writes a solid horizontal run without blending.
func (pf *PixFmtRGB555[B]) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	packed := pf.blender.MakePix(c.R, c.G, c.B)
	row := buffer.RowU16(pf.rbuf, y)
	for i := range length {
		row[x+i] = packed
	}
}

// BlendHline blends a horizontal line with uniform coverage.
func (pf *PixFmtRGB555[B]) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	if c.A == basics.CoverFull && cover == basics.CoverFull {
		packed := pf.blender.MakePix(c.R, c.G, c.B)
		for i := range length {
			row[x+i] = packed
		}
	} else {
		for i := range length {
			pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cover)
		}
	}
}

// CopyVline writes a solid vertical run without blending.
func (pf *PixFmtRGB555[B]) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		pf.CopyPixel(x, y+i, c)
	}
}

// BlendVline blends a solid vertical run with uniform coverage.
func (pf *PixFmtRGB555[B]) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if c.A == basics.CoverFull && cover == basics.CoverFull {
		packed := pf.blender.MakePix(c.R, c.G, c.B)
		for i := range length {
			buffer.RowU16(pf.rbuf, y+i)[x] = packed
		}
	} else {
		for i := range length {
			row := buffer.RowU16(pf.rbuf, y+i)
			pf.blender.BlendPix(&row[x], c.R, c.G, c.B, c.A, cover)
		}
	}
}

// CopyBar fills a rectangle without blending.
func (pf *PixFmtRGB555[B]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear]) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

// BlendBar blends a rectangle with uniform coverage.
func (pf *PixFmtRGB555[B]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

// BlendSolidHspan blends a horizontal span with varying per-pixel coverage.
func (pf *PixFmtRGB555[B]) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	opaque := c.A == basics.CoverFull
	if covers == nil {
		if opaque {
			packed := pf.blender.MakePix(c.R, c.G, c.B)
			for i := range length {
				row[x+i] = packed
			}
		} else {
			for i := range length {
				pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, basics.CoverFull)
			}
		}
	} else {
		for i := range length {
			if i < len(covers) {
				cover := covers[i]
				if cover == 0 {
					continue
				}
				if opaque && cover == basics.CoverFull {
					row[x+i] = pf.blender.MakePix(c.R, c.G, c.B)
				} else {
					pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cover)
				}
			}
		}
	}
}

// BlendSolidVspan blends a vertical span with varying per-pixel coverage.
func (pf *PixFmtRGB555[B]) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if covers == nil {
		for i := range length {
			pf.BlendPixel(x, y+i, c, basics.CoverFull)
		}
	} else {
		for i := range length {
			if i < len(covers) && covers[i] > 0 {
				pf.BlendPixel(x, y+i, c, covers[i])
			}
		}
	}
}

// CopyColorHspan copies a horizontal span of colors without blending.
func (pf *PixFmtRGB555[B]) CopyColorHspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	for i := range length {
		pf.CopyPixel(x+i, y, colors[i%len(colors)])
	}
}

// BlendColorHspan blends a horizontal span of colors with per-pixel or uniform coverage.
func (pf *PixFmtRGB555[B]) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	if length > len(colors) {
		length = len(colors)
	}
	row := buffer.RowU16(pf.rbuf, y)
	for i := range length {
		c := colors[i]
		if c.A == 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr == 0 {
			continue
		}
		pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cvr)
	}
}

// CopyColorVspan copies a vertical span of colors without blending.
func (pf *PixFmtRGB555[B]) CopyColorVspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		pf.CopyPixel(x, y+i, colors[i%len(colors)])
	}
}

// BlendColorVspan blends a vertical span of colors with per-pixel or uniform coverage.
func (pf *PixFmtRGB555[B]) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		c := colors[i%len(colors)]
		if c.A == 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x, y+i, c, cvr)
	}
}

// Clear fills the entire buffer with c (alpha channel ignored; all pixels opaque).
func (pf *PixFmtRGB555[B]) Clear(c color.RGBA8[color.Linear]) {
	packed := pf.blender.MakePix(c.R, c.G, c.B)
	for y := range pf.Height() {
		row := buffer.RowU16(pf.rbuf, y)
		for i := range row {
			row[i] = packed
		}
	}
}

// Fill is an alias for Clear.
func (pf *PixFmtRGB555[B]) Fill(c color.RGBA8[color.Linear]) { pf.Clear(c) }

// ─── PixFmtRGB565 ────────────────────────────────────────────────────────────

// PixFmtRGB565 is the Go equivalent of AGG's pixfmt_alpha_blend_rgb_packed for
// RGB565 (16-bit) packed pixel format.
type PixFmtRGB565[B blender.RGB16PackedBlender] struct {
	rbuf    *buffer.RenderingBufferU16
	blender B
}

// NewPixFmtRGB565 creates a new RGB565 pixel format over the given rendering buffer.
func NewPixFmtRGB565[B blender.RGB16PackedBlender](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtRGB565[B] {
	return &PixFmtRGB565[B]{rbuf: rbuf, blender: blender}
}

func (pf *PixFmtRGB565[B]) Width() int    { return pf.rbuf.Width() }
func (pf *PixFmtRGB565[B]) Height() int   { return pf.rbuf.Height() }
func (pf *PixFmtRGB565[B]) PixWidth() int { return 2 }

func (pf *PixFmtRGB565[B]) Pixel(x, y int) color.RGBA8[color.Linear] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA8[color.Linear]{}
	}
	row := buffer.RowU16(pf.rbuf, y)
	r, g, b := pf.blender.UnpackPix(row[x])
	return color.RGBA8[color.Linear]{R: r, G: g, B: b, A: 255}
}

func (pf *PixFmtRGB565[B]) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU16(pf.rbuf, y)
	row[x] = pf.blender.MakePix(c.R, c.G, c.B)
}

func (pf *PixFmtRGB565[B]) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.A == 0 {
		return
	}
	row := buffer.RowU16(pf.rbuf, y)
	pf.blender.BlendPix(&row[x], c.R, c.G, c.B, c.A, cover)
}

func (pf *PixFmtRGB565[B]) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	packed := pf.blender.MakePix(c.R, c.G, c.B)
	row := buffer.RowU16(pf.rbuf, y)
	for i := range length {
		row[x+i] = packed
	}
}

func (pf *PixFmtRGB565[B]) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	if c.A == basics.CoverFull && cover == basics.CoverFull {
		packed := pf.blender.MakePix(c.R, c.G, c.B)
		for i := range length {
			row[x+i] = packed
		}
	} else {
		for i := range length {
			pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cover)
		}
	}
}

func (pf *PixFmtRGB565[B]) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		pf.CopyPixel(x, y+i, c)
	}
}

func (pf *PixFmtRGB565[B]) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if c.A == basics.CoverFull && cover == basics.CoverFull {
		packed := pf.blender.MakePix(c.R, c.G, c.B)
		for i := range length {
			buffer.RowU16(pf.rbuf, y+i)[x] = packed
		}
	} else {
		for i := range length {
			row := buffer.RowU16(pf.rbuf, y+i)
			pf.blender.BlendPix(&row[x], c.R, c.G, c.B, c.A, cover)
		}
	}
}

func (pf *PixFmtRGB565[B]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear]) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

func (pf *PixFmtRGB565[B]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

func (pf *PixFmtRGB565[B]) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	opaque := c.A == basics.CoverFull
	if covers == nil {
		if opaque {
			packed := pf.blender.MakePix(c.R, c.G, c.B)
			for i := range length {
				row[x+i] = packed
			}
		} else {
			for i := range length {
				pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, basics.CoverFull)
			}
		}
	} else {
		for i := range length {
			if i < len(covers) {
				cover := covers[i]
				if cover == 0 {
					continue
				}
				if opaque && cover == basics.CoverFull {
					row[x+i] = pf.blender.MakePix(c.R, c.G, c.B)
				} else {
					pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cover)
				}
			}
		}
	}
}

func (pf *PixFmtRGB565[B]) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if covers == nil {
		for i := range length {
			pf.BlendPixel(x, y+i, c, basics.CoverFull)
		}
	} else {
		for i := range length {
			if i < len(covers) && covers[i] > 0 {
				pf.BlendPixel(x, y+i, c, covers[i])
			}
		}
	}
}

func (pf *PixFmtRGB565[B]) CopyColorHspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	for i := range length {
		pf.CopyPixel(x+i, y, colors[i%len(colors)])
	}
}

func (pf *PixFmtRGB565[B]) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	if length > len(colors) {
		length = len(colors)
	}
	row := buffer.RowU16(pf.rbuf, y)
	for i := range length {
		c := colors[i]
		if c.A == 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr == 0 {
			continue
		}
		pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cvr)
	}
}

func (pf *PixFmtRGB565[B]) CopyColorVspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		pf.CopyPixel(x, y+i, colors[i%len(colors)])
	}
}

func (pf *PixFmtRGB565[B]) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		c := colors[i%len(colors)]
		if c.A == 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x, y+i, c, cvr)
	}
}

func (pf *PixFmtRGB565[B]) Clear(c color.RGBA8[color.Linear]) {
	packed := pf.blender.MakePix(c.R, c.G, c.B)
	for y := range pf.Height() {
		row := buffer.RowU16(pf.rbuf, y)
		for i := range row {
			row[i] = packed
		}
	}
}

func (pf *PixFmtRGB565[B]) Fill(c color.RGBA8[color.Linear]) { pf.Clear(c) }

// ─── PixFmtBGR555 ────────────────────────────────────────────────────────────

// PixFmtBGR555 is the Go equivalent of AGG's pixfmt_alpha_blend_rgb_packed for
// BGR555 (15-bit) packed pixel format.
type PixFmtBGR555[B blender.RGB16PackedBlender] struct {
	rbuf    *buffer.RenderingBufferU16
	blender B
}

// NewPixFmtBGR555 creates a new BGR555 pixel format over the given rendering buffer.
func NewPixFmtBGR555[B blender.RGB16PackedBlender](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtBGR555[B] {
	return &PixFmtBGR555[B]{rbuf: rbuf, blender: blender}
}

func (pf *PixFmtBGR555[B]) Width() int    { return pf.rbuf.Width() }
func (pf *PixFmtBGR555[B]) Height() int   { return pf.rbuf.Height() }
func (pf *PixFmtBGR555[B]) PixWidth() int { return 2 }

func (pf *PixFmtBGR555[B]) Pixel(x, y int) color.RGBA8[color.Linear] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA8[color.Linear]{}
	}
	row := buffer.RowU16(pf.rbuf, y)
	r, g, b := pf.blender.UnpackPix(row[x])
	return color.RGBA8[color.Linear]{R: r, G: g, B: b, A: 255}
}

func (pf *PixFmtBGR555[B]) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU16(pf.rbuf, y)
	row[x] = pf.blender.MakePix(c.R, c.G, c.B)
}

func (pf *PixFmtBGR555[B]) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.A == 0 {
		return
	}
	row := buffer.RowU16(pf.rbuf, y)
	pf.blender.BlendPix(&row[x], c.R, c.G, c.B, c.A, cover)
}

func (pf *PixFmtBGR555[B]) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	packed := pf.blender.MakePix(c.R, c.G, c.B)
	row := buffer.RowU16(pf.rbuf, y)
	for i := range length {
		row[x+i] = packed
	}
}

func (pf *PixFmtBGR555[B]) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	if c.A == basics.CoverFull && cover == basics.CoverFull {
		packed := pf.blender.MakePix(c.R, c.G, c.B)
		for i := range length {
			row[x+i] = packed
		}
	} else {
		for i := range length {
			pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cover)
		}
	}
}

func (pf *PixFmtBGR555[B]) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		pf.CopyPixel(x, y+i, c)
	}
}

func (pf *PixFmtBGR555[B]) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if c.A == basics.CoverFull && cover == basics.CoverFull {
		packed := pf.blender.MakePix(c.R, c.G, c.B)
		for i := range length {
			buffer.RowU16(pf.rbuf, y+i)[x] = packed
		}
	} else {
		for i := range length {
			row := buffer.RowU16(pf.rbuf, y+i)
			pf.blender.BlendPix(&row[x], c.R, c.G, c.B, c.A, cover)
		}
	}
}

func (pf *PixFmtBGR555[B]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear]) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

func (pf *PixFmtBGR555[B]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

func (pf *PixFmtBGR555[B]) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	opaque := c.A == basics.CoverFull
	if covers == nil {
		if opaque {
			packed := pf.blender.MakePix(c.R, c.G, c.B)
			for i := range length {
				row[x+i] = packed
			}
		} else {
			for i := range length {
				pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, basics.CoverFull)
			}
		}
	} else {
		for i := range length {
			if i < len(covers) {
				cover := covers[i]
				if cover == 0 {
					continue
				}
				if opaque && cover == basics.CoverFull {
					row[x+i] = pf.blender.MakePix(c.R, c.G, c.B)
				} else {
					pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cover)
				}
			}
		}
	}
}

func (pf *PixFmtBGR555[B]) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if covers == nil {
		for i := range length {
			pf.BlendPixel(x, y+i, c, basics.CoverFull)
		}
	} else {
		for i := range length {
			if i < len(covers) && covers[i] > 0 {
				pf.BlendPixel(x, y+i, c, covers[i])
			}
		}
	}
}

func (pf *PixFmtBGR555[B]) CopyColorHspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	for i := range length {
		pf.CopyPixel(x+i, y, colors[i%len(colors)])
	}
}

func (pf *PixFmtBGR555[B]) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	if length > len(colors) {
		length = len(colors)
	}
	row := buffer.RowU16(pf.rbuf, y)
	for i := range length {
		c := colors[i]
		if c.A == 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr == 0 {
			continue
		}
		pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cvr)
	}
}

func (pf *PixFmtBGR555[B]) CopyColorVspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		pf.CopyPixel(x, y+i, colors[i%len(colors)])
	}
}

func (pf *PixFmtBGR555[B]) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		c := colors[i%len(colors)]
		if c.A == 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x, y+i, c, cvr)
	}
}

func (pf *PixFmtBGR555[B]) Clear(c color.RGBA8[color.Linear]) {
	packed := pf.blender.MakePix(c.R, c.G, c.B)
	for y := range pf.Height() {
		row := buffer.RowU16(pf.rbuf, y)
		for i := range row {
			row[i] = packed
		}
	}
}

func (pf *PixFmtBGR555[B]) Fill(c color.RGBA8[color.Linear]) { pf.Clear(c) }

// ─── PixFmtBGR565 ────────────────────────────────────────────────────────────

// PixFmtBGR565 is the Go equivalent of AGG's pixfmt_alpha_blend_rgb_packed for
// BGR565 (16-bit) packed pixel format.
type PixFmtBGR565[B blender.RGB16PackedBlender] struct {
	rbuf    *buffer.RenderingBufferU16
	blender B
}

// NewPixFmtBGR565 creates a new BGR565 pixel format over the given rendering buffer.
func NewPixFmtBGR565[B blender.RGB16PackedBlender](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtBGR565[B] {
	return &PixFmtBGR565[B]{rbuf: rbuf, blender: blender}
}

func (pf *PixFmtBGR565[B]) Width() int    { return pf.rbuf.Width() }
func (pf *PixFmtBGR565[B]) Height() int   { return pf.rbuf.Height() }
func (pf *PixFmtBGR565[B]) PixWidth() int { return 2 }

func (pf *PixFmtBGR565[B]) Pixel(x, y int) color.RGBA8[color.Linear] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA8[color.Linear]{}
	}
	row := buffer.RowU16(pf.rbuf, y)
	r, g, b := pf.blender.UnpackPix(row[x])
	return color.RGBA8[color.Linear]{R: r, G: g, B: b, A: 255}
}

func (pf *PixFmtBGR565[B]) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	row := buffer.RowU16(pf.rbuf, y)
	row[x] = pf.blender.MakePix(c.R, c.G, c.B)
}

func (pf *PixFmtBGR565[B]) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.A == 0 {
		return
	}
	row := buffer.RowU16(pf.rbuf, y)
	pf.blender.BlendPix(&row[x], c.R, c.G, c.B, c.A, cover)
}

func (pf *PixFmtBGR565[B]) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	packed := pf.blender.MakePix(c.R, c.G, c.B)
	row := buffer.RowU16(pf.rbuf, y)
	for i := range length {
		row[x+i] = packed
	}
}

func (pf *PixFmtBGR565[B]) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	if c.A == basics.CoverFull && cover == basics.CoverFull {
		packed := pf.blender.MakePix(c.R, c.G, c.B)
		for i := range length {
			row[x+i] = packed
		}
	} else {
		for i := range length {
			pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cover)
		}
	}
}

func (pf *PixFmtBGR565[B]) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		pf.CopyPixel(x, y+i, c)
	}
}

func (pf *PixFmtBGR565[B]) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if c.A == basics.CoverFull && cover == basics.CoverFull {
		packed := pf.blender.MakePix(c.R, c.G, c.B)
		for i := range length {
			buffer.RowU16(pf.rbuf, y+i)[x] = packed
		}
	} else {
		for i := range length {
			row := buffer.RowU16(pf.rbuf, y+i)
			pf.blender.BlendPix(&row[x], c.R, c.G, c.B, c.A, cover)
		}
	}
}

func (pf *PixFmtBGR565[B]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear]) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

func (pf *PixFmtBGR565[B]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

func (pf *PixFmtBGR565[B]) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	opaque := c.A == basics.CoverFull
	if covers == nil {
		if opaque {
			packed := pf.blender.MakePix(c.R, c.G, c.B)
			for i := range length {
				row[x+i] = packed
			}
		} else {
			for i := range length {
				pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, basics.CoverFull)
			}
		}
	} else {
		for i := range length {
			if i < len(covers) {
				cover := covers[i]
				if cover == 0 {
					continue
				}
				if opaque && cover == basics.CoverFull {
					row[x+i] = pf.blender.MakePix(c.R, c.G, c.B)
				} else {
					pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cover)
				}
			}
		}
	}
}

func (pf *PixFmtBGR565[B]) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if covers == nil {
		for i := range length {
			pf.BlendPixel(x, y+i, c, basics.CoverFull)
		}
	} else {
		for i := range length {
			if i < len(covers) && covers[i] > 0 {
				pf.BlendPixel(x, y+i, c, covers[i])
			}
		}
	}
}

func (pf *PixFmtBGR565[B]) CopyColorHspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	for i := range length {
		pf.CopyPixel(x+i, y, colors[i%len(colors)])
	}
}

func (pf *PixFmtBGR565[B]) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	if length > len(colors) {
		length = len(colors)
	}
	row := buffer.RowU16(pf.rbuf, y)
	for i := range length {
		c := colors[i]
		if c.A == 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr == 0 {
			continue
		}
		pf.blender.BlendPix(&row[x+i], c.R, c.G, c.B, c.A, cvr)
	}
}

func (pf *PixFmtBGR565[B]) CopyColorVspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		pf.CopyPixel(x, y+i, colors[i%len(colors)])
	}
}

func (pf *PixFmtBGR565[B]) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	for i := range length {
		c := colors[i%len(colors)]
		if c.A == 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x, y+i, c, cvr)
	}
}

func (pf *PixFmtBGR565[B]) Clear(c color.RGBA8[color.Linear]) {
	packed := pf.blender.MakePix(c.R, c.G, c.B)
	for y := range pf.Height() {
		row := buffer.RowU16(pf.rbuf, y)
		for i := range row {
			row[i] = packed
		}
	}
}

func (pf *PixFmtBGR565[B]) Fill(c color.RGBA8[color.Linear]) { pf.Clear(c) }

// ─── NoBlender ───────────────────────────────────────────────────────────────

// NoBlender is a placeholder type for pixel formats without blending capability.
type NoBlender struct{}

func (nb NoBlender) BlendPix(pixel *basics.Int16u, r, g, b, alpha, cover basics.Int8u) {}
func (nb NoBlender) MakePix(r, g, b basics.Int8u) basics.Int16u                        { return 0 }
func (nb NoBlender) UnpackPix(pixel basics.Int16u) (r, g, b basics.Int8u)              { return }

// ─── Convenience type aliases ─────────────────────────────────────────────────

// Plain formats (no blending)
type (
	PixFmtRGB555Plain = PixFmtRGB555[NoBlender]
	PixFmtBGR555Plain = PixFmtBGR555[NoBlender]
	PixFmtRGB565Plain = PixFmtRGB565[NoBlender]
	PixFmtBGR565Plain = PixFmtBGR565[NoBlender]
)

// Anti-aliased formats with standard blending (matches C++ pixfmt_rgb555 etc.)
type (
	PixFmtRGB555AA = PixFmtRGB555[blender.BlenderRGB555]
	PixFmtRGB565AA = PixFmtRGB565[blender.BlenderRGB565]
	PixFmtBGR555AA = PixFmtBGR555[blender.BlenderBGR555]
	PixFmtBGR565AA = PixFmtBGR565[blender.BlenderBGR565]
)
