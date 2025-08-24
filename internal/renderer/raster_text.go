// Package renderer provides raster text rendering functionality for AGG.
// This package implements text renderers that can render text using
// glyph rasterizers with solid colors or scanline renderers.
package renderer

import (
	"agg_go/internal/basics"
	"agg_go/internal/glyph"
)

// BaseRendererInterface defines the methods required by solid text renderers
type BaseRendererInterface[C any] interface {
	// BlendSolidHspan blends a horizontal span with solid color
	BlendSolidHspan(x, y, len int, c C, covers []basics.CoverType)
	// BlendSolidVspan blends a vertical span with solid color
	BlendSolidVspan(x, y, len int, c C, covers []basics.CoverType)
}

// ScanlineRendererInterface defines the methods required by scanline text renderers
type ScanlineRendererInterface interface {
	// Prepare prepares the renderer for scanline rendering
	Prepare()
	// Render renders a scanline
	Render(scanline ScanlineInterface)
}

// ScanlineInterface defines the interface for scanlines
type ScanlineInterface interface {
	Y() int
	NumSpans() int
	Begin() SpanIterator
}

// SpanIterator defines the interface for span iteration
type SpanIterator interface {
	Next() *Span
	HasNext() bool
}

// Span represents a single span with coverage data
type Span struct {
	X      int
	Len    int
	Covers []basics.CoverType
}

// RendererRasterHTextSolid renders horizontal text with solid colors.
// This is the Go equivalent of AGG's renderer_raster_htext_solid template class.
type RendererRasterHTextSolid[BR BaseRendererInterface[C], GG glyph.GlyphGenerator, C any] struct {
	ren   BR
	glyph GG
	color C
}

// NewRendererRasterHTextSolid creates a new horizontal solid text renderer
func NewRendererRasterHTextSolid[BR BaseRendererInterface[C], GG glyph.GlyphGenerator, C any](ren BR, glyphGen GG) *RendererRasterHTextSolid[BR, GG, C] {
	return &RendererRasterHTextSolid[BR, GG, C]{
		ren:   ren,
		glyph: glyphGen,
	}
}

// Attach attaches a new base renderer
func (r *RendererRasterHTextSolid[BR, GG, C]) Attach(ren BR) {
	r.ren = ren
}

// SetColor sets the text color
func (r *RendererRasterHTextSolid[BR, GG, C]) SetColor(c C) {
	r.color = c
}

// Color returns the current text color
func (r *RendererRasterHTextSolid[BR, GG, C]) Color() C {
	return r.color
}

// RenderText renders the given text at the specified position
func (r *RendererRasterHTextSolid[BR, GG, C]) RenderText(x, y float64, str string, flip bool) {
	var rect glyph.GlyphRect

	for _, ch := range str {
		r.glyph.Prepare(&rect, x, y, ch, flip)
		if rect.X2 >= rect.X1 {
			var i int
			// Glyph.Span internally flips the row index. For non-flipped text,
			// pass (rect.Y2 - i) so the internal flip restores the original order.
			// For flipped text, pass (i - rect.Y1) to intentionally mirror.
			if !flip {
				for i = rect.Y1; i <= rect.Y2; i++ {
					covers := r.glyph.Span(rect.Y2 - i)
					if len(covers) > 0 {
						r.ren.BlendSolidHspan(rect.X1, i, rect.X2-rect.X1+1, r.color, covers)
					}
				}
			} else {
				for i = rect.Y1; i <= rect.Y2; i++ {
					covers := r.glyph.Span(i - rect.Y1)
					if len(covers) > 0 {
						r.ren.BlendSolidHspan(rect.X1, i, rect.X2-rect.X1+1, r.color, covers)
					}
				}
			}
		}
		x += rect.DX
		y += rect.DY
	}
}

// RendererRasterVTextSolid renders vertical text with solid colors.
// This is the Go equivalent of AGG's renderer_raster_vtext_solid template class.
type RendererRasterVTextSolid[BR BaseRendererInterface[C], GG glyph.GlyphGenerator, C any] struct {
	ren   BR
	glyph GG
	color C
}

// NewRendererRasterVTextSolid creates a new vertical solid text renderer
func NewRendererRasterVTextSolid[BR BaseRendererInterface[C], GG glyph.GlyphGenerator, C any](ren BR, glyphGen GG) *RendererRasterVTextSolid[BR, GG, C] {
	return &RendererRasterVTextSolid[BR, GG, C]{
		ren:   ren,
		glyph: glyphGen,
	}
}

// Attach attaches a new base renderer
func (r *RendererRasterVTextSolid[BR, GG, C]) Attach(ren BR) {
	r.ren = ren
}

// SetColor sets the text color
func (r *RendererRasterVTextSolid[BR, GG, C]) SetColor(c C) {
	r.color = c
}

// Color returns the current text color
func (r *RendererRasterVTextSolid[BR, GG, C]) Color() C {
	return r.color
}

// RenderText renders the given text vertically at the specified position
func (r *RendererRasterVTextSolid[BR, GG, C]) RenderText(x, y float64, str string, flip bool) {
	var rect glyph.GlyphRect

	for _, ch := range str {
		r.glyph.Prepare(&rect, x, y, ch, !flip)
		if rect.X2 >= rect.X1 {
			var i int
			if flip {
				for i = rect.Y1; i <= rect.Y2; i++ {
					covers := r.glyph.Span(i - rect.Y1)
					if len(covers) > 0 {
						r.ren.BlendSolidVspan(i, rect.X1, rect.X2-rect.X1+1, r.color, covers)
					}
				}
			} else {
				for i = rect.Y1; i <= rect.Y2; i++ {
					covers := r.glyph.Span(rect.Y2 - i)
					if len(covers) > 0 {
						r.ren.BlendSolidVspan(i, rect.X1, rect.X2-rect.X1+1, r.color, covers)
					}
				}
			}
		}
		x += rect.DX
		y += rect.DY
	}
}

// ScanlineSingleSpan implements a scanline with a single span for text rendering
type ScanlineSingleSpan struct {
	y    int
	span Span
}

// NewScanlineSingleSpan creates a new single-span scanline
func NewScanlineSingleSpan(x, y int, len int, covers []basics.CoverType) *ScanlineSingleSpan {
	return &ScanlineSingleSpan{
		y: y,
		span: Span{
			X:      x,
			Len:    len,
			Covers: covers,
		},
	}
}

// Y returns the scanline's y coordinate
func (s *ScanlineSingleSpan) Y() int {
	return s.y
}

// NumSpans returns the number of spans (always 1)
func (s *ScanlineSingleSpan) NumSpans() int {
	return 1
}

// Begin returns an iterator for the spans
func (s *ScanlineSingleSpan) Begin() SpanIterator {
	return &SingleSpanIterator{span: &s.span, hasNext: true}
}

// SingleSpanIterator implements SpanIterator for a single span
type SingleSpanIterator struct {
	span    *Span
	hasNext bool
}

// Next returns the next span
func (it *SingleSpanIterator) Next() *Span {
	if it.hasNext {
		it.hasNext = false
		return it.span
	}
	return nil
}

// HasNext returns whether there are more spans
func (it *SingleSpanIterator) HasNext() bool {
	return it.hasNext
}

// RendererRasterHText renders horizontal text with scanline renderers (for gradients/patterns).
// This is the Go equivalent of AGG's renderer_raster_htext template class.
type RendererRasterHText[SR ScanlineRendererInterface, GG glyph.GlyphGenerator] struct {
	ren   SR
	glyph GG
}

// NewRendererRasterHText creates a new horizontal scanline text renderer
func NewRendererRasterHText[SR ScanlineRendererInterface, GG glyph.GlyphGenerator](ren SR, glyphGen GG) *RendererRasterHText[SR, GG] {
	return &RendererRasterHText[SR, GG]{
		ren:   ren,
		glyph: glyphGen,
	}
}

// RenderText renders the given text using scanline rendering
func (r *RendererRasterHText[SR, GG]) RenderText(x, y float64, str string, flip bool) {
	var rect glyph.GlyphRect

	for _, ch := range str {
		r.glyph.Prepare(&rect, x, y, ch, flip)
		if rect.X2 >= rect.X1 {
			r.ren.Prepare()
			var i int
			if !flip {
				for i = rect.Y1; i <= rect.Y2; i++ {
					covers := r.glyph.Span(rect.Y2 - i)
					if len(covers) > 0 {
						scanline := NewScanlineSingleSpan(rect.X1, i, rect.X2-rect.X1+1, covers)
						r.ren.Render(scanline)
					}
				}
			} else {
				for i = rect.Y1; i <= rect.Y2; i++ {
					covers := r.glyph.Span(i - rect.Y1)
					if len(covers) > 0 {
						scanline := NewScanlineSingleSpan(rect.X1, i, rect.X2-rect.X1+1, covers)
						r.ren.Render(scanline)
					}
				}
			}
		}
		x += rect.DX
		y += rect.DY
	}
}
