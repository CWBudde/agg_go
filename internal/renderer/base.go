// Package renderer provides high-level rendering functionality for AGG.
package renderer

import (
	"agg_go/internal/basics"
)

// PixelFormat defines the required methods for a pixel format that can be
// used with RendererBase, parameterized by a concrete color type C.
type PixelFormat[C any] interface {
	// Basic properties
	Width() int
	Height() int
	PixWidth() int

	// Pixel operations
	CopyPixel(x, y int, c C)
	BlendPixel(x, y int, c C, cover basics.Int8u)
	Pixel(x, y int) C

	// Line operations
	CopyHline(x, y, length int, c C)
	BlendHline(x, y, length int, c C, cover basics.Int8u)
	CopyVline(x, y, length int, c C)
	BlendVline(x, y, length int, c C, cover basics.Int8u)

	// Rectangle operations
	CopyBar(x1, y1, x2, y2 int, c C)
	BlendBar(x1, y1, x2, y2 int, c C, cover basics.Int8u)

	// Span operations for anti-aliasing
	BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u)
	BlendSolidVspan(x, y, length int, c C, covers []basics.Int8u)

	// Color span operations
	CopyColorHspan(x, y, length int, colors []C)
	BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u)
	CopyColorVspan(x, y, length int, colors []C)
	BlendColorVspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u)

	// Clear operations
	Clear(c C)
	Fill(c C)
}

// RendererBase provides the foundational structure for all rendering operations.
// It is a generic type that works with any pixel format (PF) and color type (C).
//
// The RendererBase manages the core rendering state including the pixel format
// for drawing operations and the clipping boundaries that constrain where
// pixels can be rendered. This serves as the base for more specialized
// renderer implementations.
//
// Type parameters:
//   - PF: The pixel format type that must implement PixelFormat[C]
//   - C: The color type used by the pixel format
type RendererBase[PF PixelFormat[C], C any] struct {
	pixfmt  PF           // The pixel format
	clipBox basics.RectI // Current clipping rectangle
}

// NewRendererBase creates a new typed renderer with default (empty) clipping.
func NewRendererBase[PF PixelFormat[C], C any]() *RendererBase[PF, C] {
	return &RendererBase[PF, C]{
		clipBox: basics.RectI{X1: 1, Y1: 1, X2: 0, Y2: 0}, // Invalid box (empty)
	}
}

// NewRendererBaseWithPixfmt creates a new typed renderer with the given pixel format.
func NewRendererBaseWithPixfmt[PF PixelFormat[C], C any](pixfmt PF) *RendererBase[PF, C] {
	return &RendererBase[PF, C]{
		pixfmt:  pixfmt,
		clipBox: basics.RectI{X1: 0, Y1: 0, X2: pixfmt.Width() - 1, Y2: pixfmt.Height() - 1},
	}
}

// Attach attaches a pixel format to the typed renderer.
func (r *RendererBase[PF, C]) Attach(pixfmt PF) {
	r.pixfmt = pixfmt
	r.clipBox = basics.RectI{X1: 0, Y1: 0, X2: pixfmt.Width() - 1, Y2: pixfmt.Height() - 1}
}

// Ren returns a reference to the pixel format (const version)
func (r *RendererBase[PF, C]) Ren() PF { return r.pixfmt }

// RenMut returns a mutable reference to the pixel format
func (r *RendererBase[PF, C]) RenMut() *PF { return &r.pixfmt }

// Width returns the width of the rendering buffer
func (r *RendererBase[PF, C]) Width() int { return r.pixfmt.Width() }

// Height returns the height of the rendering buffer
func (r *RendererBase[PF, C]) Height() int { return r.pixfmt.Height() }

// ClipBox sets the clipping box with bounds checking
// Returns true if the clipping box intersects with the buffer bounds
func (r *RendererBase[PF, C]) ClipBox(x1, y1, x2, y2 int) bool {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	cb := basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
	bufferBounds := basics.RectI{X1: 0, Y1: 0, X2: r.Width() - 1, Y2: r.Height() - 1}
	clipped, hasIntersection := basics.IntersectRectangles(cb, bufferBounds)
	if hasIntersection {
		r.clipBox = clipped
		return true
	}
	r.clipBox = basics.RectI{X1: 1, Y1: 1, X2: 0, Y2: 0}
	return false
}

// ResetClipping resets the clipping to the entire buffer or makes it empty
func (r *RendererBase[PF, C]) ResetClipping(visibility bool) {
	if visibility {
		r.clipBox = basics.RectI{X1: 0, Y1: 0, X2: r.Width() - 1, Y2: r.Height() - 1}
	} else {
		r.clipBox = basics.RectI{X1: 1, Y1: 1, X2: 0, Y2: 0}
	}
}

// ClipBoxNaked sets the clipping box without bounds checking
func (r *RendererBase[PF, C]) ClipBoxNaked(x1, y1, x2, y2 int) {
	r.clipBox = basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

// InBox tests if a point is inside the clipping box
func (r *RendererBase[PF, C]) InBox(x, y int) bool {
	return x >= r.clipBox.X1 && y >= r.clipBox.Y1 && x <= r.clipBox.X2 && y <= r.clipBox.Y2
}

// ClipBoxRect returns the current clipping box
func (r *RendererBase[PF, C]) ClipBoxRect() basics.RectI { return r.clipBox }

// Xmin returns the minimum x coordinate of the clipping box
func (r *RendererBase[PF, C]) Xmin() int { return r.clipBox.X1 }

// Ymin returns the minimum y coordinate of the clipping box
func (r *RendererBase[PF, C]) Ymin() int { return r.clipBox.Y1 }

// Xmax returns the maximum x coordinate of the clipping box
func (r *RendererBase[PF, C]) Xmax() int { return r.clipBox.X2 }

// Ymax returns the maximum y coordinate of the clipping box
func (r *RendererBase[PF, C]) Ymax() int { return r.clipBox.Y2 }

// BoundingClipBox returns the bounding clipping box (same as ClipBoxRect)
func (r *RendererBase[PF, C]) BoundingClipBox() basics.RectI { return r.clipBox }

// BoundingXmin returns the bounding minimum x coordinate
func (r *RendererBase[PF, C]) BoundingXmin() int { return r.clipBox.X1 }

// BoundingYmin returns the bounding minimum y coordinate
func (r *RendererBase[PF, C]) BoundingYmin() int { return r.clipBox.Y1 }

// BoundingXmax returns the bounding maximum x coordinate
func (r *RendererBase[PF, C]) BoundingXmax() int { return r.clipBox.X2 }

// BoundingYmax returns the bounding maximum y coordinate
func (r *RendererBase[PF, C]) BoundingYmax() int { return r.clipBox.Y2 }

// Clear clears the entire buffer with the given color (no blending)
func (r *RendererBase[PF, C]) Clear(c C) {
	if r.Width() > 0 {
		for y := 0; y < r.Height(); y++ {
			r.pixfmt.CopyHline(0, y, r.Width(), c)
		}
	}
}

// Fill fills the entire buffer with the given color using blending
func (r *RendererBase[PF, C]) Fill(c C) {
	if r.Width() > 0 {
		for y := 0; y < r.Height(); y++ {
			r.pixfmt.BlendHline(0, y, r.Width(), c, basics.CoverFull)
		}
	}
}

// CopyPixel copies a pixel at the given coordinates (respects clipping)
func (r *RendererBase[PF, C]) CopyPixel(x, y int, c C) {
	if r.InBox(x, y) {
		r.pixfmt.CopyPixel(x, y, c)
	}
}

// BlendPixel blends a pixel at the given coordinates with coverage (respects clipping)
func (r *RendererBase[PF, C]) BlendPixel(x, y int, c C, cover basics.Int8u) {
	if r.InBox(x, y) {
		r.pixfmt.BlendPixel(x, y, c, cover)
	}
}

// Pixel returns the pixel color at the given coordinates
// Returns the zero value of C if outside clipping box
func (r *RendererBase[PF, C]) Pixel(x, y int) C {
	if r.InBox(x, y) {
		return r.pixfmt.Pixel(x, y)
	}
	var zero C
	return zero
}

// CopyHline copies a horizontal line (respects clipping)
func (r *RendererBase[PF, C]) CopyHline(x1, y, x2 int, c C) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y > r.Ymax() || y < r.Ymin() {
		return
	}
	if x1 > r.Xmax() || x2 < r.Xmin() {
		return
	}
	if x1 < r.Xmin() {
		x1 = r.Xmin()
	}
	if x2 > r.Xmax() {
		x2 = r.Xmax()
	}
	r.pixfmt.CopyHline(x1, y, x2-x1+1, c)
}

// CopyVline copies a vertical line (respects clipping)
func (r *RendererBase[PF, C]) CopyVline(x, y1, y2 int, c C) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x > r.Xmax() || x < r.Xmin() {
		return
	}
	if y1 > r.Ymax() || y2 < r.Ymin() {
		return
	}
	if y1 < r.Ymin() {
		y1 = r.Ymin()
	}
	if y2 > r.Ymax() {
		y2 = r.Ymax()
	}
	r.pixfmt.CopyVline(x, y1, y2-y1+1, c)
}

// BlendHline blends a horizontal line with coverage (respects clipping)
func (r *RendererBase[PF, C]) BlendHline(x1, y, x2 int, c C, cover basics.Int8u) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y > r.Ymax() || y < r.Ymin() {
		return
	}
	if x1 > r.Xmax() || x2 < r.Xmin() {
		return
	}
	if x1 < r.Xmin() {
		x1 = r.Xmin()
	}
	if x2 > r.Xmax() {
		x2 = r.Xmax()
	}
	r.pixfmt.BlendHline(x1, y, x2-x1+1, c, cover)
}

// BlendVline blends a vertical line with coverage (respects clipping)
func (r *RendererBase[PF, C]) BlendVline(x, y1, y2 int, c C, cover basics.Int8u) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x > r.Xmax() || x < r.Xmin() {
		return
	}
	if y1 > r.Ymax() || y2 < r.Ymin() {
		return
	}
	if y1 < r.Ymin() {
		y1 = r.Ymin()
	}
	if y2 > r.Ymax() {
		y2 = r.Ymax()
	}
	r.pixfmt.BlendVline(x, y1, y2-y1+1, c, cover)
}

// CopyBar copies a rectangular bar (respects clipping)
func (r *RendererBase[PF, C]) CopyBar(x1, y1, x2, y2 int, c C) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	rc := basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
	clipped, hasIntersection := basics.IntersectRectangles(rc, r.clipBox)
	if hasIntersection {
		for y := clipped.Y1; y <= clipped.Y2; y++ {
			r.pixfmt.CopyHline(clipped.X1, y, clipped.X2-clipped.X1+1, c)
		}
	}
}

// BlendBar blends a rectangular bar with coverage (respects clipping)
func (r *RendererBase[PF, C]) BlendBar(x1, y1, x2, y2 int, c C, cover basics.Int8u) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	rc := basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
	clipped, hasIntersection := basics.IntersectRectangles(rc, r.clipBox)
	if hasIntersection {
		for y := clipped.Y1; y <= clipped.Y2; y++ {
			r.pixfmt.BlendHline(clipped.X1, y, clipped.X2-clipped.X1+1, c, cover)
		}
	}
}

// BlendSolidHspan blends a horizontal span with solid color and coverage array
func (r *RendererBase[PF, C]) BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u) {
	if y > r.Ymax() || y < r.Ymin() {
		return
	}
	if x < r.Xmin() {
		length -= r.Xmin() - x
		if length <= 0 {
			return
		}
		covers = covers[r.Xmin()-x:]
		x = r.Xmin()
	}
	if x+length > r.Xmax() {
		length = r.Xmax() - x + 1
		if length <= 0 {
			return
		}
	}
	r.pixfmt.BlendSolidHspan(x, y, length, c, covers)
}

// BlendSolidVspan blends a vertical span with solid color and coverage array
func (r *RendererBase[PF, C]) BlendSolidVspan(x, y, length int, c C, covers []basics.Int8u) {
	if x > r.Xmax() || x < r.Xmin() {
		return
	}
	if y < r.Ymin() {
		length -= r.Ymin() - y
		if length <= 0 {
			return
		}
		covers = covers[r.Ymin()-y:]
		y = r.Ymin()
	}
	if y+length > r.Ymax() {
		length = r.Ymax() - y + 1
		if length <= 0 {
			return
		}
	}
	r.pixfmt.BlendSolidVspan(x, y, length, c, covers)
}

// CopyColorHspan copies a horizontal span with color array
func (r *RendererBase[PF, C]) CopyColorHspan(x, y, length int, colors []C) {
	if y > r.Ymax() || y < r.Ymin() {
		return
	}
	if x < r.Xmin() {
		d := r.Xmin() - x
		length -= d
		if length <= 0 {
			return
		}
		colors = colors[d:]
		x = r.Xmin()
	}
	if x+length > r.Xmax() {
		length = r.Xmax() - x + 1
		if length <= 0 {
			return
		}
	}
	r.pixfmt.CopyColorHspan(x, y, length, colors)
}

// CopyColorVspan copies a vertical span with color array
func (r *RendererBase[PF, C]) CopyColorVspan(x, y, length int, colors []C) {
	if x > r.Xmax() || x < r.Xmin() {
		return
	}
	if y < r.Ymin() {
		d := r.Ymin() - y
		length -= d
		if length <= 0 {
			return
		}
		colors = colors[d:]
		y = r.Ymin()
	}
	if y+length > r.Ymax() {
		length = r.Ymax() - y + 1
		if length <= 0 {
			return
		}
	}
	r.pixfmt.CopyColorVspan(x, y, length, colors)
}

// BlendColorHspan blends a horizontal span with color and coverage arrays
func (r *RendererBase[PF, C]) BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	if y > r.Ymax() || y < r.Ymin() {
		return
	}
	if x < r.Xmin() {
		d := r.Xmin() - x
		length -= d
		if length <= 0 {
			return
		}
		if covers != nil {
			covers = covers[d:]
		}
		colors = colors[d:]
		x = r.Xmin()
	}
	if x+length > r.Xmax() {
		length = r.Xmax() - x + 1
		if length <= 0 {
			return
		}
	}
	r.pixfmt.BlendColorHspan(x, y, length, colors, covers, cover)
}

// BlendColorVspan blends a vertical span with color and coverage arrays
func (r *RendererBase[PF, C]) BlendColorVspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	if x > r.Xmax() || x < r.Xmin() {
		return
	}
	if y < r.Ymin() {
		d := r.Ymin() - y
		length -= d
		if length <= 0 {
			return
		}
		if covers != nil {
			covers = covers[d:]
		}
		colors = colors[d:]
		y = r.Ymin()
	}
	if y+length > r.Ymax() {
		length = r.Ymax() - y + 1
		if length <= 0 {
			return
		}
	}
	r.pixfmt.BlendColorVspan(x, y, length, colors, covers, cover)
}

// ClipRectArea clips rectangles for copying/blending operations (typed version).
// Returns the actual area (width/height) that can be processed in rc.X2/rc.Y2.
func (r *RendererBase[PF, C]) ClipRectArea(dst *basics.RectI, src *basics.RectI, wsrc, hsrc int) basics.RectI {
	rc := basics.RectI{X1: 0, Y1: 0, X2: 0, Y2: 0}
	cb := r.clipBox
	cb.X2++
	cb.Y2++

	if src.X1 < 0 {
		dst.X1 -= src.X1
		src.X1 = 0
	}
	if src.Y1 < 0 {
		dst.Y1 -= src.Y1
		src.Y1 = 0
	}

	if src.X2 > wsrc-1 {
		src.X2 = wsrc - 1
	}
	if src.Y2 > hsrc-1 {
		src.Y2 = hsrc - 1
	}

	if dst.X1 < cb.X1 {
		src.X1 += cb.X1 - dst.X1
		dst.X1 = cb.X1
	}
	if dst.Y1 < cb.Y1 {
		src.Y1 += cb.Y1 - dst.Y1
		dst.Y1 = cb.Y1
	}

	if dst.X2+1 > cb.X2 { // cb uses X2/Y2 as inclusive before ++ above
		dst.X2 = cb.X2 - 1
	}
	if dst.Y2+1 > cb.Y2 {
		dst.Y2 = cb.Y2 - 1
	}

	rc.X2 = dst.X2 - dst.X1 + 1
	rc.Y2 = dst.Y2 - dst.Y1 + 1

	if rc.X2 > (src.X2 - src.X1 + 1) {
		rc.X2 = (src.X2 - src.X1 + 1)
	}
	if rc.Y2 > (src.Y2 - src.Y1 + 1) {
		rc.Y2 = (src.Y2 - src.Y1 + 1)
	}
	if rc.X2 < 0 {
		rc.X2 = 0
	}
	if rc.Y2 < 0 {
		rc.Y2 = 0
	}
	return rc
}

// CopyFrom copies from another typed pixel format into this renderer.
// If rectSrcPtr is nil, copies the full source.
func (r *RendererBase[PF, C]) CopyFrom(src PixelFormat[C], rectSrcPtr *basics.RectI, dx, dy int) {
	wsrc, hsrc := src.Width(), src.Height()
	if wsrc <= 0 || hsrc <= 0 || r.Width() <= 0 || r.Height() <= 0 {
		return
	}

	var srcRect basics.RectI
	if rectSrcPtr == nil {
		srcRect = basics.RectI{X1: 0, Y1: 0, X2: wsrc - 1, Y2: hsrc - 1}
	} else {
		srcRect = *rectSrcPtr
	}

	dstRect := basics.RectI{
		X1: dx,
		Y1: dy,
		X2: dx + (srcRect.X2 - srcRect.X1),
		Y2: dy + (srcRect.Y2 - srcRect.Y1),
	}

	rc := r.ClipRectArea(&dstRect, &srcRect, wsrc, hsrc)
	if rc.X2 <= 0 || rc.Y2 <= 0 {
		return
	}

	row := make([]C, rc.X2)
	for rowOfs := 0; rowOfs < rc.Y2; rowOfs++ {
		sy := srcRect.Y1 + rowOfs
		dy2 := dstRect.Y1 + rowOfs

		for i := 0; i < rc.X2; i++ {
			sx := srcRect.X1 + i
			row[i] = src.Pixel(sx, sy)
		}
		r.pixfmt.CopyColorHspan(dstRect.X1, dy2, rc.X2, row)
	}
}

// BlendFrom blends from another typed pixel format into this renderer with uniform coverage.
func (r *RendererBase[PF, C]) BlendFrom(src PixelFormat[C], rectSrcPtr *basics.RectI, dx, dy int, cover basics.Int8u) {
	wsrc, hsrc := src.Width(), src.Height()
	if wsrc <= 0 || hsrc <= 0 || r.Width() <= 0 || r.Height() <= 0 {
		return
	}

	var srcRect basics.RectI
	if rectSrcPtr == nil {
		srcRect = basics.RectI{X1: 0, Y1: 0, X2: wsrc - 1, Y2: hsrc - 1}
	} else {
		srcRect = *rectSrcPtr
	}

	dstRect := basics.RectI{
		X1: dx,
		Y1: dy,
		X2: dx + (srcRect.X2 - srcRect.X1),
		Y2: dy + (srcRect.Y2 - srcRect.Y1),
	}

	rc := r.ClipRectArea(&dstRect, &srcRect, wsrc, hsrc)
	if rc.X2 <= 0 || rc.Y2 <= 0 {
		return
	}

	row := make([]C, rc.X2)
	for rowOfs := 0; rowOfs < rc.Y2; rowOfs++ {
		sy := srcRect.Y1 + rowOfs
		dy2 := dstRect.Y1 + rowOfs

		for i := 0; i < rc.X2; i++ {
			sx := srcRect.X1 + i
			row[i] = src.Pixel(sx, sy)
		}
		// No per-pixel covers (nil); use uniform "cover".
		r.pixfmt.BlendColorHspan(dstRect.X1, dy2, rc.X2, row, nil, cover)
	}
}
