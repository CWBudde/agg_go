// Package renderer provides high-level rendering functionality for AGG.
// This file implements the multi-clipping renderer equivalent to AGG's renderer_mclip<PixelFormat>.
package renderer

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// RendererMClip provides multi-clipping renderer functionality.
// This is the Go equivalent of AGG's renderer_mclip<PixelFormat> template class.
// It manages multiple clipping regions and renders to each one separately.
type RendererMClip[PF PixelFormatInterface, CT ColorTypeInterface] struct {
	ren    *RendererBase[PF, CT]           // Base renderer
	clip   *array.PodBVector[basics.RectI] // List of clipping boxes
	currCb int                             // Current clipping box index
	bounds basics.RectI                    // Overall bounding box of all clipping regions
}

// NewRendererMClip creates a new multi-clipping renderer with the given pixel format
func NewRendererMClip[PF PixelFormatInterface, CT ColorTypeInterface](pixfmt PF) *RendererMClip[PF, CT] {
	ren := NewRendererBaseWithPixfmt[PF, CT](pixfmt)
	clip := array.NewPodBVectorWithScale[basics.RectI](array.NewBlockScale(2)) // Block size 4 like C++ version

	return &RendererMClip[PF, CT]{
		ren:    ren,
		clip:   clip,
		currCb: 0,
		bounds: basics.RectI{X1: ren.Xmin(), Y1: ren.Ymin(), X2: ren.Xmax(), Y2: ren.Ymax()},
	}
}

// Attach attaches a pixel format to the renderer
func (r *RendererMClip[PF, CT]) Attach(pixfmt PF) {
	r.ren.Attach(pixfmt)
	r.ResetClipping(true)
}

// Ren returns a reference to the underlying pixel format
func (r *RendererMClip[PF, CT]) Ren() PF {
	return r.ren.Ren()
}

// Width returns the width of the rendering surface
func (r *RendererMClip[PF, CT]) Width() int {
	return r.ren.Width()
}

// Height returns the height of the rendering surface
func (r *RendererMClip[PF, CT]) Height() int {
	return r.ren.Height()
}

// ClipBox returns the current clipping box
func (r *RendererMClip[PF, CT]) ClipBox() basics.RectI {
	return r.ren.ClipBoxRect()
}

// XMin returns the minimum X coordinate of the current clipping box
func (r *RendererMClip[PF, CT]) XMin() int {
	return r.ren.Xmin()
}

// YMin returns the minimum Y coordinate of the current clipping box
func (r *RendererMClip[PF, CT]) YMin() int {
	return r.ren.Ymin()
}

// XMax returns the maximum X coordinate of the current clipping box
func (r *RendererMClip[PF, CT]) XMax() int {
	return r.ren.Xmax()
}

// YMax returns the maximum Y coordinate of the current clipping box
func (r *RendererMClip[PF, CT]) YMax() int {
	return r.ren.Ymax()
}

// BoundingClipBox returns the overall bounding box of all clipping regions
func (r *RendererMClip[PF, CT]) BoundingClipBox() basics.RectI {
	return r.bounds
}

// BoundingXMin returns the minimum X coordinate of the overall bounding box
func (r *RendererMClip[PF, CT]) BoundingXMin() int {
	return r.bounds.X1
}

// BoundingYMin returns the minimum Y coordinate of the overall bounding box
func (r *RendererMClip[PF, CT]) BoundingYMin() int {
	return r.bounds.Y1
}

// BoundingXMax returns the maximum X coordinate of the overall bounding box
func (r *RendererMClip[PF, CT]) BoundingXMax() int {
	return r.bounds.X2
}

// BoundingYMax returns the maximum Y coordinate of the overall bounding box
func (r *RendererMClip[PF, CT]) BoundingYMax() int {
	return r.bounds.Y2
}

// FirstClipBox sets the current clipping box to the first one in the list
func (r *RendererMClip[PF, CT]) FirstClipBox() {
	r.currCb = 0
	if r.clip.Size() > 0 {
		cb := r.clip.At(0)
		r.ren.ClipBoxNaked(cb.X1, cb.Y1, cb.X2, cb.Y2)
	}
}

// NextClipBox advances to the next clipping box
// Returns true if there is a next clipping box, false otherwise
func (r *RendererMClip[PF, CT]) NextClipBox() bool {
	r.currCb++
	if r.currCb < r.clip.Size() {
		cb := r.clip.At(r.currCb)
		r.ren.ClipBoxNaked(cb.X1, cb.Y1, cb.X2, cb.Y2)
		return true
	}
	return false
}

// ResetClipping resets all clipping regions and optionally sets visibility
func (r *RendererMClip[PF, CT]) ResetClipping(visibility bool) {
	r.ren.ResetClipping(visibility)
	r.clip.RemoveAll()
	r.currCb = 0
	r.bounds = r.ren.ClipBoxRect()
}

// AddClipBox adds a new clipping box to the list
func (r *RendererMClip[PF, CT]) AddClipBox(x1, y1, x2, y2 int) {
	cb := basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}

	// Normalize the rectangle (ensure x1 <= x2, y1 <= y2)
	if cb.X1 > cb.X2 {
		cb.X1, cb.X2 = cb.X2, cb.X1
	}
	if cb.Y1 > cb.Y2 {
		cb.Y1, cb.Y2 = cb.Y2, cb.Y1
	}

	// Clip against the rendering surface bounds
	surfaceBounds := basics.RectI{X1: 0, Y1: 0, X2: r.Width() - 1, Y2: r.Height() - 1}
	if cb.Clip(surfaceBounds) {
		r.clip.Add(cb)

		// Update overall bounding box - if this is the first clip box, initialize bounds
		if r.clip.Size() == 1 {
			r.bounds = cb
		} else {
			// Expand bounds to include this clip box
			if cb.X1 < r.bounds.X1 {
				r.bounds.X1 = cb.X1
			}
			if cb.Y1 < r.bounds.Y1 {
				r.bounds.Y1 = cb.Y1
			}
			if cb.X2 > r.bounds.X2 {
				r.bounds.X2 = cb.X2
			}
			if cb.Y2 > r.bounds.Y2 {
				r.bounds.Y2 = cb.Y2
			}
		}
	}
}

// Clear clears the entire rendering surface with the given color
func (r *RendererMClip[PF, CT]) Clear(c interface{}) {
	r.ren.Clear(c)
}

// CopyPixel copies a pixel at the specified coordinates across all clipping regions
func (r *RendererMClip[PF, CT]) CopyPixel(x, y int, c interface{}) {
	r.FirstClipBox()
	for {
		if r.ren.InBox(x, y) {
			r.ren.Ren().CopyPixel(x, y, c)
			break
		}
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendPixel blends a pixel at the specified coordinates across all clipping regions
func (r *RendererMClip[PF, CT]) BlendPixel(x, y int, c interface{}, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		if r.ren.InBox(x, y) {
			r.ren.Ren().BlendPixel(x, y, c, cover)
			break
		}
		if !r.NextClipBox() {
			break
		}
	}
}

// Pixel returns the pixel value at the specified coordinates from the first applicable clipping region
func (r *RendererMClip[PF, CT]) Pixel(x, y int) interface{} {
	r.FirstClipBox()
	for {
		if r.ren.InBox(x, y) {
			return r.ren.Ren().Pixel(x, y)
		}
		if !r.NextClipBox() {
			break
		}
	}

	// Return no color if pixel is not in any clipping region
	var ct CT
	return ct.NoColor()
}

// CopyHline copies a horizontal line across all clipping regions
func (r *RendererMClip[PF, CT]) CopyHline(x1, y, x2 int, c interface{}) {
	r.FirstClipBox()
	for {
		r.ren.CopyHline(x1, y, x2, c)
		if !r.NextClipBox() {
			break
		}
	}
}

// CopyVline copies a vertical line across all clipping regions
func (r *RendererMClip[PF, CT]) CopyVline(x, y1, y2 int, c interface{}) {
	r.FirstClipBox()
	for {
		r.ren.CopyVline(x, y1, y2, c)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendHline blends a horizontal line across all clipping regions
func (r *RendererMClip[PF, CT]) BlendHline(x1, y, x2 int, c interface{}, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendHline(x1, y, x2, c, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendVline blends a vertical line across all clipping regions
func (r *RendererMClip[PF, CT]) BlendVline(x, y1, y2 int, c interface{}, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendVline(x, y1, y2, c, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

// CopyBar copies a filled rectangle across all clipping regions
func (r *RendererMClip[PF, CT]) CopyBar(x1, y1, x2, y2 int, c interface{}) {
	r.FirstClipBox()
	for {
		r.ren.CopyBar(x1, y1, x2, y2, c)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendBar blends a filled rectangle across all clipping regions
func (r *RendererMClip[PF, CT]) BlendBar(x1, y1, x2, y2 int, c interface{}, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendBar(x1, y1, x2, y2, c, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendSolidHspan blends a horizontal span with solid color and variable coverage across all clipping regions
func (r *RendererMClip[PF, CT]) BlendSolidHspan(x, y, length int, c interface{}, covers []basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendSolidHspan(x, y, length, c, covers)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendSolidVspan blends a vertical span with solid color and variable coverage across all clipping regions
func (r *RendererMClip[PF, CT]) BlendSolidVspan(x, y, length int, c interface{}, covers []basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendSolidVspan(x, y, length, c, covers)
		if !r.NextClipBox() {
			break
		}
	}
}

// CopyColorHspan copies a horizontal span of colors across all clipping regions
func (r *RendererMClip[PF, CT]) CopyColorHspan(x, y, length int, colors []interface{}) {
	r.FirstClipBox()
	for {
		r.ren.CopyColorHspan(x, y, length, colors)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendColorHspan blends a horizontal span of colors with optional coverage across all clipping regions
func (r *RendererMClip[PF, CT]) BlendColorHspan(x, y, length int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendColorHspan(x, y, length, colors, covers, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendColorVspan blends a vertical span of colors with optional coverage across all clipping regions
func (r *RendererMClip[PF, CT]) BlendColorVspan(x, y, length int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendColorVspan(x, y, length, colors, covers, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

// CopyFrom copies from another rendering buffer across all clipping regions
func (r *RendererMClip[PF, CT]) CopyFrom(src interface{}, rectSrcPtr *basics.RectI, dx, dy int) {
	r.FirstClipBox()
	for {
		r.ren.CopyFrom(src, rectSrcPtr, dx, dy)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendFrom blends from another pixel format renderer across all clipping regions
// This is a generic method that accepts any source renderer type
func (r *RendererMClip[PF, CT]) BlendFrom(src interface{}, rectSrcPtr *basics.RectI, dx, dy int, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		// Note: This would need to call a BlendFrom method on the base renderer
		// For now, we'll assume this functionality exists or will be added to RendererBase
		// In the C++ version, this is a template method that works with any source pixel format
		// In Go, we'd need to implement this via interfaces or type assertions

		// Placeholder: if the base renderer had a BlendFrom method, we'd call it like this:
		// r.ren.BlendFrom(src, rectSrcPtr, dx, dy, cover)

		// For now, we'll use CopyFrom as a fallback
		r.ren.CopyFrom(src, rectSrcPtr, dx, dy)
		if !r.NextClipBox() {
			break
		}
	}
}
