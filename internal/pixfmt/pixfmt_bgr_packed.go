package pixfmt

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
)

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
	var ok bool
	x, length, _, ok = clipSpan(x, length, pf.Width())
	if !ok {
		return
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
	var ok bool
	x, length, _, ok = clipSpan(x, length, pf.Width())
	if !ok {
		return
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
	var ok bool
	y, length, _, ok = clipSpan(y, length, pf.Height())
	if !ok {
		return
	}
	for i := range length {
		pf.CopyPixel(x, y+i, c)
	}
}

func (pf *PixFmtBGR555[B]) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	var ok bool
	y, length, _, ok = clipSpan(y, length, pf.Height())
	if !ok {
		return
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
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok {
		return
	}
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
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
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok {
		return
	}
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
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
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		pf.CopyPixel(x+i, y, colors[i])
	}
}

func (pf *PixFmtBGR555[B]) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
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
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		pf.CopyPixel(x, y+i, colors[i])
	}
}

func (pf *PixFmtBGR555[B]) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
	}
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		c := colors[i]
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
	var ok bool
	x, length, _, ok = clipSpan(x, length, pf.Width())
	if !ok {
		return
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
	var ok bool
	x, length, _, ok = clipSpan(x, length, pf.Width())
	if !ok {
		return
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
	var ok bool
	y, length, _, ok = clipSpan(y, length, pf.Height())
	if !ok {
		return
	}
	for i := range length {
		pf.CopyPixel(x, y+i, c)
	}
}

func (pf *PixFmtBGR565[B]) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}
	var ok bool
	y, length, _, ok = clipSpan(y, length, pf.Height())
	if !ok {
		return
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
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok {
		return
	}
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
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
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok {
		return
	}
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
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
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		pf.CopyPixel(x+i, y, colors[i])
	}
}

func (pf *PixFmtBGR565[B]) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
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
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		pf.CopyPixel(x, y+i, colors[i])
	}
}

func (pf *PixFmtBGR565[B]) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
	}
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		c := colors[i]
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
