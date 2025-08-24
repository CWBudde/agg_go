// Package renderer provides high-level rendering functionality for AGG.
// Typed multi-clip renderer equivalent of AGG's renderer_mclip<PixelFormat>.
package renderer

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// RendererMClip provides multi-clipping renderer functionality for typed pixel formats.
type RendererMClip[PF PixelFormat[C], C any] struct {
	ren    *RendererBase[PF, C]            // Base renderer (typed)
	clip   *array.PodBVector[basics.RectI] // List of clipping boxes
	currCb int                             // Current clipping box index
	bounds basics.RectI                    // Overall bounding box of all clipping regions
}

// NewRendererMClip creates a new typed multi-clipping renderer with the given pixel format.
func NewRendererMClip[PF PixelFormat[C], C any](pixfmt PF) *RendererMClip[PF, C] {
	ren := NewRendererBaseWithPixfmt[PF, C](pixfmt)
	clip := array.NewPodBVectorWithScale[basics.RectI](array.NewBlockScale(2)) // same block scaling as C++
	return &RendererMClip[PF, C]{
		ren:    ren,
		clip:   clip,
		currCb: 0,
		bounds: basics.RectI{X1: ren.Xmin(), Y1: ren.Ymin(), X2: ren.Xmax(), Y2: ren.Ymax()},
	}
}

// Attach attaches a pixel format to the renderer.
func (r *RendererMClip[PF, C]) Attach(pixfmt PF) {
	r.ren.Attach(pixfmt)
	r.ResetClipping(true)
}

// Ren returns a reference to the underlying pixel format.
func (r *RendererMClip[PF, C]) Ren() PF               { return r.ren.Ren() }
func (r *RendererMClip[PF, C]) Width() int            { return r.ren.Width() }
func (r *RendererMClip[PF, C]) Height() int           { return r.ren.Height() }
func (r *RendererMClip[PF, C]) ClipBox() basics.RectI { return r.ren.ClipBoxRect() }
func (r *RendererMClip[PF, C]) XMin() int             { return r.ren.Xmin() }
func (r *RendererMClip[PF, C]) YMin() int             { return r.ren.Ymin() }
func (r *RendererMClip[PF, C]) XMax() int             { return r.ren.Xmax() }
func (r *RendererMClip[PF, C]) YMax() int             { return r.ren.Ymax() }

func (r *RendererMClip[PF, C]) BoundingClipBox() basics.RectI { return r.bounds }
func (r *RendererMClip[PF, C]) BoundingXMin() int             { return r.bounds.X1 }
func (r *RendererMClip[PF, C]) BoundingYMin() int             { return r.bounds.Y1 }
func (r *RendererMClip[PF, C]) BoundingXMax() int             { return r.bounds.X2 }
func (r *RendererMClip[PF, C]) BoundingYMax() int             { return r.bounds.Y2 }

// Iteration over clip boxes
func (r *RendererMClip[PF, C]) FirstClipBox() {
	r.currCb = 0
	if r.clip.Size() > 0 {
		cb := r.clip.At(0)
		r.ren.ClipBoxNaked(cb.X1, cb.Y1, cb.X2, cb.Y2)
	}
}

func (r *RendererMClip[PF, C]) NextClipBox() bool {
	r.currCb++
	if r.currCb < r.clip.Size() {
		cb := r.clip.At(r.currCb)
		r.ren.ClipBoxNaked(cb.X1, cb.Y1, cb.X2, cb.Y2)
		return true
	}
	return false
}

// Clipping management
func (r *RendererMClip[PF, C]) ResetClipping(visibility bool) {
	r.ren.ResetClipping(visibility)
	r.clip.RemoveAll()
	r.currCb = 0
	r.bounds = r.ren.ClipBoxRect()
}

func (r *RendererMClip[PF, C]) AddClipBox(x1, y1, x2, y2 int) {
	cb := basics.RectI{X1: x1, Y1: y1, X2: x2, Y2: y2}
	if cb.X1 > cb.X2 {
		cb.X1, cb.X2 = cb.X2, cb.X1
	}
	if cb.Y1 > cb.Y2 {
		cb.Y1, cb.Y2 = cb.Y2, cb.Y1
	}
	surface := basics.RectI{X1: 0, Y1: 0, X2: r.Width() - 1, Y2: r.Height() - 1}
	if cb.Clip(surface) {
		r.clip.Add(cb)
		if r.clip.Size() == 1 {
			r.bounds = cb
		} else {
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

// Drawing ops (typed C)
func (r *RendererMClip[PF, C]) Clear(c C) { r.ren.Clear(c) }

func (r *RendererMClip[PF, C]) CopyPixel(x, y int, c C) {
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

func (r *RendererMClip[PF, C]) BlendPixel(x, y int, c C, cover basics.Int8u) {
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

func (r *RendererMClip[PF, C]) Pixel(x, y int) C {
	r.FirstClipBox()
	for {
		if r.ren.InBox(x, y) {
			return r.ren.Ren().Pixel(x, y)
		}
		if !r.NextClipBox() {
			break
		}
	}
	var zero C
	return zero
}

func (r *RendererMClip[PF, C]) CopyHline(x1, y, x2 int, c C) {
	r.FirstClipBox()
	for {
		r.ren.CopyHline(x1, y, x2, c)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) CopyVline(x, y1, y2 int, c C) {
	r.FirstClipBox()
	for {
		r.ren.CopyVline(x, y1, y2, c)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) BlendHline(x1, y, x2 int, c C, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendHline(x1, y, x2, c, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) BlendVline(x, y1, y2 int, c C, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendVline(x, y1, y2, c, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) CopyBar(x1, y1, x2, y2 int, c C) {
	r.FirstClipBox()
	for {
		r.ren.CopyBar(x1, y1, x2, y2, c)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) BlendBar(x1, y1, x2, y2 int, c C, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendBar(x1, y1, x2, y2, c, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendSolidHspan(x, y, length, c, covers)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) BlendSolidVspan(x, y, length int, c C, covers []basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendSolidVspan(x, y, length, c, covers)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) CopyColorHspan(x, y, length int, colors []C) {
	r.FirstClipBox()
	for {
		r.ren.CopyColorHspan(x, y, length, colors)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) CopyColorVspan(x, y, length int, colors []C) {
	r.FirstClipBox()
	for {
		r.ren.CopyColorVspan(x, y, length, colors)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendColorHspan(x, y, length, colors, covers, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

func (r *RendererMClip[PF, C]) BlendColorVspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendColorVspan(x, y, length, colors, covers, cover)
		if !r.NextClipBox() {
			break
		}
	}
}

// CopyFrom copies from a typed source across all clipping regions.
func (r *RendererMClip[PF, C]) CopyFrom(src PixelFormat[C], rectSrcPtr *basics.RectI, dx, dy int) {
	r.FirstClipBox()
	for {
		r.ren.CopyFrom(src, rectSrcPtr, dx, dy)
		if !r.NextClipBox() {
			break
		}
	}
}

// BlendFrom blends from a typed source across all clipping regions.
func (r *RendererMClip[PF, C]) BlendFrom(src PixelFormat[C], rectSrcPtr *basics.RectI, dx, dy int, cover basics.Int8u) {
	r.FirstClipBox()
	for {
		r.ren.BlendFrom(src, rectSrcPtr, dx, dy, cover)
		if !r.NextClipBox() {
			break
		}
	}
}
