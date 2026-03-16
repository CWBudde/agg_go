package renderer

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/glyph"
)

// BaseRendererInterface is the base-renderer contract needed by solid raster-text renderers.
type BaseRendererInterface[C any] interface {
	// BlendSolidHspan blends a horizontal span with solid color
	BlendSolidHspan(x, y, len int, c C, covers []basics.CoverType)
	// BlendSolidVspan blends a vertical span with solid color
	BlendSolidVspan(x, y, len int, c C, covers []basics.CoverType)
}

// ScanlineRendererInterface is the minimal scanline-renderer contract used by raster text.
type ScanlineRendererInterface interface {
	// Prepare prepares the renderer for scanline rendering
	Prepare()
	// Render renders a scanline
	Render(scanline ScanlineInterface)
}

// ScanlineInterface is the subset of scanline behavior needed by text renderers.
type ScanlineInterface interface {
	Y() int
	NumSpans() int
	Begin() SpanIterator
}

// SpanIterator iterates over the spans contained in one scanline.
type SpanIterator interface {
	Next() *Span
	HasNext() bool
}

// Span stores one contiguous glyph-coverage run.
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

// NewRendererRasterHTextSolid creates the horizontal solid raster-text renderer.
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

// RendererRasterVTextSolid renders vertical glyph rasters using solid-color spans.
type RendererRasterVTextSolid[BR BaseRendererInterface[C], GG glyph.GlyphGenerator, C any] struct {
	ren   BR
	glyph GG
	color C
}

// NewRendererRasterVTextSolid creates the vertical solid raster-text renderer.
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

// ScanlineSingleSpan adapts a single glyph span to the scanline interface.
type ScanlineSingleSpan struct {
	y    int
	span Span
}

// NewScanlineSingleSpan creates the single-span scanline used by raster text.
func NewScanlineSingleSpan(x, y, length int, covers []basics.CoverType) *ScanlineSingleSpan {
	return &ScanlineSingleSpan{
		y: y,
		span: Span{
			X:      x,
			Len:    length,
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

// SingleSpanIterator iterates over the one span contained in ScanlineSingleSpan.
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

// RendererRasterHText renders horizontal glyph rasters through an abstract
// scanline renderer, which is useful for generated spans such as gradients.
type RendererRasterHText[SR ScanlineRendererInterface, GG glyph.GlyphGenerator] struct {
	ren   SR
	glyph GG
}

// NewRendererRasterHText creates the scanline-based horizontal raster-text renderer.
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
