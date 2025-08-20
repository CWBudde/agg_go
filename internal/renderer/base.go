// Package renderer provides high-level rendering functionality for AGG.
// This package implements the base renderer that provides clipping,
// pixel operations, and drawing primitives on top of pixel formats.
package renderer

import (
	"agg_go/internal/basics"
)

// PixelFormatInterface defines the required methods for a pixel format
// that can be used with RendererBase
type PixelFormatInterface interface {
	// Basic properties
	Width() int
	Height() int
	PixWidth() int

	// Pixel operations
	CopyPixel(x, y int, c interface{})
	BlendPixel(x, y int, c interface{}, cover basics.Int8u)
	Pixel(x, y int) interface{}

	// Line operations
	CopyHline(x, y, len int, c interface{})
	BlendHline(x, y, len int, c interface{}, cover basics.Int8u)
	CopyVline(x, y, len int, c interface{})
	BlendVline(x, y, len int, c interface{}, cover basics.Int8u)

	// Span operations for anti-aliasing
	BlendSolidHspan(x, y, len int, c interface{}, covers []basics.Int8u)
	BlendSolidVspan(x, y, len int, c interface{}, covers []basics.Int8u)

	// Color span operations
	CopyColorHspan(x, y, len int, colors []interface{})
	BlendColorHspan(x, y, len int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u)
	CopyColorVspan(x, y, len int, colors []interface{})
	BlendColorVspan(x, y, len int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u)

	// Copy from another buffer
	CopyFrom(src interface{}, dstX, dstY, srcX, srcY, len int)
}

// ColorTypeInterface defines the required methods for color types
type ColorTypeInterface interface {
	// NoColor returns a transparent/empty color
	NoColor() interface{}
}

// RendererBase provides the base renderer template class functionality.
// This is the Go equivalent of AGG's renderer_base<PixelFormat> template class.
type RendererBase[PF PixelFormatInterface, CT ColorTypeInterface] struct {
	pixfmt  PF           // The pixel format
	clipBox basics.RectI // Current clipping rectangle
}

// NewRendererBase creates a new renderer with default (empty) clipping
func NewRendererBase[PF PixelFormatInterface, CT ColorTypeInterface]() *RendererBase[PF, CT] {
	return &RendererBase[PF, CT]{
		clipBox: basics.RectI{X1: 1, Y1: 1, X2: 0, Y2: 0}, // Invalid box (empty)
	}
}

// NewRendererBaseWithPixfmt creates a new renderer with the given pixel format
func NewRendererBaseWithPixfmt[PF PixelFormatInterface, CT ColorTypeInterface](pixfmt PF) *RendererBase[PF, CT] {
	return &RendererBase[PF, CT]{
		pixfmt:  pixfmt,
		clipBox: basics.RectI{X1: 0, Y1: 0, X2: pixfmt.Width() - 1, Y2: pixfmt.Height() - 1},
	}
}

// Attach attaches a pixel format to the renderer
func (r *RendererBase[PF, CT]) Attach(pixfmt PF) {
	r.pixfmt = pixfmt
	r.clipBox = basics.RectI{X1: 0, Y1: 0, X2: pixfmt.Width() - 1, Y2: pixfmt.Height() - 1}
}

// Ren returns a reference to the pixel format (const version)
func (r *RendererBase[PF, CT]) Ren() PF {
	return r.pixfmt
}

// RenMut returns a mutable reference to the pixel format
func (r *RendererBase[PF, CT]) RenMut() *PF {
	return &r.pixfmt
}

// Width returns the width of the rendering buffer
func (r *RendererBase[PF, CT]) Width() int {
	return r.pixfmt.Width()
}

// Height returns the height of the rendering buffer
func (r *RendererBase[PF, CT]) Height() int {
	return r.pixfmt.Height()
}

// ClipBox sets the clipping box with bounds checking
// Returns true if the clipping box intersects with the buffer bounds
func (r *RendererBase[PF, CT]) ClipBox(x1, y1, x2, y2 int) bool {
	// Normalize the rectangle (ensure x1 <= x2, y1 <= y2)
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Create rectangle and clip against buffer bounds
	cb := basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
	bufferBounds := basics.RectI{X1: 0, Y1: 0, X2: r.Width() - 1, Y2: r.Height() - 1}

	// Intersect with buffer bounds
	clipped, hasIntersection := basics.IntersectRectangles(cb, bufferBounds)

	if hasIntersection {
		r.clipBox = clipped
		return true
	}

	// Set invalid clip box (empty region)
	r.clipBox = basics.RectI{X1: 1, Y1: 1, X2: 0, Y2: 0}
	return false
}

// ResetClipping resets the clipping to the entire buffer or makes it empty
func (r *RendererBase[PF, CT]) ResetClipping(visibility bool) {
	if visibility {
		r.clipBox = basics.RectI{X1: 0, Y1: 0, X2: r.Width() - 1, Y2: r.Height() - 1}
	} else {
		r.clipBox = basics.RectI{X1: 1, Y1: 1, X2: 0, Y2: 0}
	}
}

// ClipBoxNaked sets the clipping box without bounds checking
func (r *RendererBase[PF, CT]) ClipBoxNaked(x1, y1, x2, y2 int) {
	r.clipBox = basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

// InBox tests if a point is inside the clipping box
func (r *RendererBase[PF, CT]) InBox(x, y int) bool {
	return x >= r.clipBox.X1 && y >= r.clipBox.Y1 &&
		x <= r.clipBox.X2 && y <= r.clipBox.Y2
}

// ClipBoxRect returns the current clipping box
func (r *RendererBase[PF, CT]) ClipBoxRect() basics.RectI {
	return r.clipBox
}

// Xmin returns the minimum x coordinate of the clipping box
func (r *RendererBase[PF, CT]) Xmin() int {
	return r.clipBox.X1
}

// Ymin returns the minimum y coordinate of the clipping box
func (r *RendererBase[PF, CT]) Ymin() int {
	return r.clipBox.Y1
}

// Xmax returns the maximum x coordinate of the clipping box
func (r *RendererBase[PF, CT]) Xmax() int {
	return r.clipBox.X2
}

// Ymax returns the maximum y coordinate of the clipping box
func (r *RendererBase[PF, CT]) Ymax() int {
	return r.clipBox.Y2
}

// BoundingClipBox returns the bounding clipping box (same as ClipBoxRect)
func (r *RendererBase[PF, CT]) BoundingClipBox() basics.RectI {
	return r.clipBox
}

// BoundingXmin returns the bounding minimum x coordinate
func (r *RendererBase[PF, CT]) BoundingXmin() int {
	return r.clipBox.X1
}

// BoundingYmin returns the bounding minimum y coordinate
func (r *RendererBase[PF, CT]) BoundingYmin() int {
	return r.clipBox.Y1
}

// BoundingXmax returns the bounding maximum x coordinate
func (r *RendererBase[PF, CT]) BoundingXmax() int {
	return r.clipBox.X2
}

// BoundingYmax returns the bounding maximum y coordinate
func (r *RendererBase[PF, CT]) BoundingYmax() int {
	return r.clipBox.Y2
}

// Clear clears the entire buffer with the given color (no blending)
func (r *RendererBase[PF, CT]) Clear(c interface{}) {
	if r.Width() > 0 {
		for y := 0; y < r.Height(); y++ {
			r.pixfmt.CopyHline(0, y, r.Width(), c)
		}
	}
}

// Fill fills the entire buffer with the given color using blending
func (r *RendererBase[PF, CT]) Fill(c interface{}) {
	if r.Width() > 0 {
		for y := 0; y < r.Height(); y++ {
			r.pixfmt.BlendHline(0, y, r.Width(), c, basics.CoverFull)
		}
	}
}

// CopyPixel copies a pixel at the given coordinates (respects clipping)
func (r *RendererBase[PF, CT]) CopyPixel(x, y int, c interface{}) {
	if r.InBox(x, y) {
		r.pixfmt.CopyPixel(x, y, c)
	}
}

// BlendPixel blends a pixel at the given coordinates with coverage (respects clipping)
func (r *RendererBase[PF, CT]) BlendPixel(x, y int, c interface{}, cover basics.Int8u) {
	if r.InBox(x, y) {
		r.pixfmt.BlendPixel(x, y, c, cover)
	}
}

// Pixel returns the pixel color at the given coordinates
// Returns NoColor if outside clipping box
func (r *RendererBase[PF, CT]) Pixel(x, y int) interface{} {
	if r.InBox(x, y) {
		return r.pixfmt.Pixel(x, y)
	}
	var ct CT
	return ct.NoColor()
}

// CopyHline copies a horizontal line (respects clipping)
func (r *RendererBase[PF, CT]) CopyHline(x1, y, x2 int, c interface{}) {
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
func (r *RendererBase[PF, CT]) CopyVline(x, y1, y2 int, c interface{}) {
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
func (r *RendererBase[PF, CT]) BlendHline(x1, y, x2 int, c interface{}, cover basics.Int8u) {
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
func (r *RendererBase[PF, CT]) BlendVline(x, y1, y2 int, c interface{}, cover basics.Int8u) {
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
func (r *RendererBase[PF, CT]) CopyBar(x1, y1, x2, y2 int, c interface{}) {
	// Normalize rectangle
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Create rectangle and clip it
	rc := basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
	clipped, hasIntersection := basics.IntersectRectangles(rc, r.clipBox)

	if hasIntersection {
		for y := clipped.Y1; y <= clipped.Y2; y++ {
			r.pixfmt.CopyHline(clipped.X1, y, clipped.X2-clipped.X1+1, c)
		}
	}
}

// BlendBar blends a rectangular bar with coverage (respects clipping)
func (r *RendererBase[PF, CT]) BlendBar(x1, y1, x2, y2 int, c interface{}, cover basics.Int8u) {
	// Normalize rectangle
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Create rectangle and clip it
	rc := basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
	clipped, hasIntersection := basics.IntersectRectangles(rc, r.clipBox)

	if hasIntersection {
		for y := clipped.Y1; y <= clipped.Y2; y++ {
			r.pixfmt.BlendHline(clipped.X1, y, clipped.X2-clipped.X1+1, c, cover)
		}
	}
}

// BlendSolidHspan blends a horizontal span with solid color and coverage array
func (r *RendererBase[PF, CT]) BlendSolidHspan(x, y, length int, c interface{}, covers []basics.Int8u) {
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
func (r *RendererBase[PF, CT]) BlendSolidVspan(x, y, length int, c interface{}, covers []basics.Int8u) {
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
func (r *RendererBase[PF, CT]) CopyColorHspan(x, y, length int, colors []interface{}) {
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
func (r *RendererBase[PF, CT]) CopyColorVspan(x, y, length int, colors []interface{}) {
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
func (r *RendererBase[PF, CT]) BlendColorHspan(x, y, length int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u) {
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
func (r *RendererBase[PF, CT]) BlendColorVspan(x, y, length int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u) {
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

// ClipRectArea clips rectangles for copying operations
// Returns the actual area that can be copied
func (r *RendererBase[PF, CT]) ClipRectArea(dst *basics.RectI, src *basics.RectI, wsrc, hsrc int) basics.RectI {
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

	if src.X2 > wsrc {
		src.X2 = wsrc
	}
	if src.Y2 > hsrc {
		src.Y2 = hsrc
	}

	if dst.X1 < cb.X1 {
		src.X1 += cb.X1 - dst.X1
		dst.X1 = cb.X1
	}
	if dst.Y1 < cb.Y1 {
		src.Y1 += cb.Y1 - dst.Y1
		dst.Y1 = cb.Y1
	}

	if dst.X2 > cb.X2 {
		dst.X2 = cb.X2
	}
	if dst.Y2 > cb.Y2 {
		dst.Y2 = cb.Y2
	}

	rc.X2 = dst.X2 - dst.X1
	rc.Y2 = dst.Y2 - dst.Y1

	if rc.X2 > src.X2-src.X1 {
		rc.X2 = src.X2 - src.X1
	}
	if rc.Y2 > src.Y2-src.Y1 {
		rc.Y2 = src.Y2 - src.Y1
	}
	return rc
}

// CopyFrom copies from another rendering buffer
// This is a generic method that can copy from any source buffer
func (r *RendererBase[PF, CT]) CopyFrom(src interface{}, rectSrcPtr *basics.RectI, dx, dy int) {
	// We need to access the source buffer's width and height
	// This is a simplified version - in a full implementation, we'd need
	// to define an interface for source buffers or use reflection

	// For now, we'll assume the source has Width() and Height() methods
	// and can be accessed via the pixel format's CopyFrom method

	// This would need to be implemented based on the specific source type
	// The C++ version uses templates to handle different source types

	// Placeholder implementation:
	r.pixfmt.CopyFrom(src, dx, dy, 0, 0, 0)
}
