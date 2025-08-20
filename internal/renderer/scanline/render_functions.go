// Package scanline provides core rendering functions for AGG scanline rendering.
package scanline

import (
	"agg_go/internal/basics"
)

// RenderScanlineAASolid renders a single anti-aliased scanline with solid color.
// This corresponds to AGG's render_scanline_aa_solid function.
func RenderScanlineAASolid(sl ScanlineInterface, ren BaseRendererInterface, color interface{}) {
	y := sl.Y()
	numSpans := sl.NumSpans()
	iter := sl.Begin()

	for i := 0; i < numSpans; i++ {
		span := iter.GetSpan()
		x := span.X

		if span.Len > 0 {
			// Positive length: anti-aliased span with coverage array
			ren.BlendSolidHspan(x, y, span.Len, color, span.Covers)
		} else {
			// Negative length: solid span with single coverage value
			// Calculate end coordinate: x - span.Len - 1
			endX := x - span.Len - 1
			cover := span.Covers[0] // First (and only) coverage value
			ren.BlendHline(x, y, endX, color, cover)
		}

		// Move to next span
		if i < numSpans-1 {
			iter.Next()
		}
	}
}

// RenderScanlinesAASolid renders all anti-aliased scanlines from a rasterizer with solid color.
// This corresponds to AGG's render_scanlines_aa_solid function.
func RenderScanlinesAASolid(ras RasterizerInterface, sl ScanlineInterface, ren BaseRendererInterface, color interface{}) {
	if !ras.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	if resetScanline, ok := sl.(ResettableScanline); ok {
		resetScanline.Reset(ras.MinX(), ras.MaxX())
	}

	// Store color reference to avoid repeated interface{} passing
	// This corresponds to AGG's: typename BaseRenderer::color_type ren_color = color;
	// In Go, we can't pre-convert the type, but we avoid repeated parameter passing
	renderColor := color

	// Sweep through all scanlines
	for ras.SweepScanline(sl) {
		// Inline the render_scanline_aa_solid logic for performance
		// This is equivalent to calling RenderScanlineAASolid but avoids function call overhead
		y := sl.Y()
		numSpans := sl.NumSpans()
		iter := sl.Begin()

		for i := 0; i < numSpans; i++ {
			span := iter.GetSpan()
			x := span.X

			if span.Len > 0 {
				// Positive length: anti-aliased span with coverage array
				ren.BlendSolidHspan(x, y, span.Len, renderColor, span.Covers)
			} else {
				// Negative length: solid span with single coverage value
				endX := x - span.Len - 1
				cover := span.Covers[0]
				ren.BlendHline(x, y, endX, renderColor, cover)
			}

			// Move to next span
			if i < numSpans-1 {
				iter.Next()
			}
		}
	}
}

// RenderScanlineAA renders a single anti-aliased scanline with span generation.
// This corresponds to AGG's render_scanline_aa function.
func RenderScanlineAA(sl ScanlineInterface, ren BaseRendererInterface, alloc SpanAllocatorInterface, spanGen SpanGeneratorInterface) {
	y := sl.Y()
	numSpans := sl.NumSpans()
	iter := sl.Begin()

	for i := 0; i < numSpans; i++ {
		span := iter.GetSpan()
		x := span.X
		length := span.Len
		covers := span.Covers

		// Handle negative length (convert to positive)
		if length < 0 {
			length = -length
		}

		// Allocate colors for this span
		colors := alloc.Allocate(length)

		// Generate colors for the span
		spanGen.Generate(colors, x, y, length)

		// Blend the span
		if span.Len < 0 {
			// Solid span: use single coverage value
			ren.BlendColorHspan(x, y, length, colors, nil, covers[0])
		} else {
			// Anti-aliased span: use coverage array
			ren.BlendColorHspan(x, y, length, colors, covers, covers[0])
		}

		// Move to next span
		if i < numSpans-1 {
			iter.Next()
		}
	}
}

// RenderScanlinesAA renders all anti-aliased scanlines from a rasterizer with span generation.
// This corresponds to AGG's render_scanlines_aa function.
func RenderScanlinesAA(ras RasterizerInterface, sl ScanlineInterface, ren BaseRendererInterface, alloc SpanAllocatorInterface, spanGen SpanGeneratorInterface) {
	if !ras.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	if resetScanline, ok := sl.(ResettableScanline); ok {
		resetScanline.Reset(ras.MinX(), ras.MaxX())
	}

	// Prepare the span generator
	spanGen.Prepare()

	// Sweep through all scanlines
	for ras.SweepScanline(sl) {
		RenderScanlineAA(sl, ren, alloc, spanGen)
	}
}

// RenderScanlineBinSolid renders a single binary scanline with solid color.
// This corresponds to AGG's render_scanline_bin_solid function.
func RenderScanlineBinSolid(sl ScanlineInterface, ren BaseRendererInterface, color interface{}) {
	numSpans := sl.NumSpans()
	iter := sl.Begin()
	y := sl.Y()

	for i := 0; i < numSpans; i++ {
		span := iter.GetSpan()

		// For binary scanlines, calculate the end coordinate
		// This matches AGG's formula: span->x - 1 + ((span->len < 0) ? -span->len : span->len)
		var endX int
		if span.Len < 0 {
			endX = span.X - span.Len - 1
		} else {
			endX = span.X + span.Len - 1
		}

		// Render with full coverage
		ren.BlendHline(span.X, y, endX, color, basics.CoverFull)

		// Move to next span
		if i < numSpans-1 {
			iter.Next()
		}
	}
}

// RenderScanlinesBinSolid renders all binary scanlines from a rasterizer with solid color.
// This corresponds to AGG's render_scanlines_bin_solid function.
func RenderScanlinesBinSolid(ras RasterizerInterface, sl ScanlineInterface, ren BaseRendererInterface, color interface{}) {
	if !ras.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	if resetScanline, ok := sl.(ResettableScanline); ok {
		resetScanline.Reset(ras.MinX(), ras.MaxX())
	}

	// Store color reference to avoid repeated interface{} passing
	// This corresponds to AGG's: typename BaseRenderer::color_type ren_color(color);
	renderColor := color

	// Sweep through all scanlines
	for ras.SweepScanline(sl) {
		// Inline the render_scanline_bin_solid logic for performance
		numSpans := sl.NumSpans()
		iter := sl.Begin()
		y := sl.Y()

		for i := 0; i < numSpans; i++ {
			span := iter.GetSpan()

			// Calculate end coordinate for binary rendering
			// This matches AGG's formula: span->x - 1 + ((span->len < 0) ? -span->len : span->len)
			var endX int
			if span.Len < 0 {
				endX = span.X - span.Len - 1
			} else {
				endX = span.X + span.Len - 1
			}

			// Render with full coverage
			ren.BlendHline(span.X, y, endX, renderColor, basics.CoverFull)

			// Move to next span
			if i < numSpans-1 {
				iter.Next()
			}
		}
	}
}

// RenderScanlineBin renders a single binary scanline with span generation.
// This corresponds to AGG's render_scanline_bin function.
func RenderScanlineBin(sl ScanlineInterface, ren BaseRendererInterface, alloc SpanAllocatorInterface, spanGen SpanGeneratorInterface) {
	y := sl.Y()
	numSpans := sl.NumSpans()
	iter := sl.Begin()

	for i := 0; i < numSpans; i++ {
		span := iter.GetSpan()
		x := span.X
		length := span.Len

		// Handle negative length (convert to positive)
		if length < 0 {
			length = -length
		}

		// Allocate colors for this span
		colors := alloc.Allocate(length)

		// Generate colors for the span
		spanGen.Generate(colors, x, y, length)

		// Blend the span with full coverage (binary rendering)
		ren.BlendColorHspan(x, y, length, colors, nil, basics.CoverFull)

		// Move to next span
		if i < numSpans-1 {
			iter.Next()
		}
	}
}

// RenderScanlinesBin renders all binary scanlines from a rasterizer with span generation.
// This corresponds to AGG's render_scanlines_bin function.
func RenderScanlinesBin(ras RasterizerInterface, sl ScanlineInterface, ren BaseRendererInterface, alloc SpanAllocatorInterface, spanGen SpanGeneratorInterface) {
	if !ras.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	if resetScanline, ok := sl.(ResettableScanline); ok {
		resetScanline.Reset(ras.MinX(), ras.MaxX())
	}

	// Prepare the span generator
	spanGen.Prepare()

	// Sweep through all scanlines
	for ras.SweepScanline(sl) {
		RenderScanlineBin(sl, ren, alloc, spanGen)
	}
}
