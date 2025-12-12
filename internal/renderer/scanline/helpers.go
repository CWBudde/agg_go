// Package scanline provides helper rendering functions for AGG scanline rendering.
package scanline

import (
	"agg_go/internal/basics"
)

// RenderScanlines is a generic scanline rendering function that works with any renderer.
// This corresponds to AGG's render_scanlines function.
func RenderScanlines[C any](ras RasterizerInterface, sl ScanlineInterface, renderer RendererInterface[C]) {
	if !ras.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	if resetScanline, ok := sl.(ResettableScanline); ok {
		resetScanline.Reset(ras.MinX(), ras.MaxX())
	}

	// Prepare the renderer
	renderer.Prepare()

	// Sweep through all scanlines
	for ras.SweepScanline(sl) {
		renderer.Render(sl)
	}
}

// PathColorStorage represents a storage interface for path colors.
// This is used by RenderAllPaths to access colors by index.
type PathColorStorage[C any] interface {
	// GetColor returns the color at the specified index
	GetColor(index int) C
}

// PathIdStorage represents a storage interface for path IDs.
// This is used by RenderAllPaths to access path IDs by index.
type PathIdStorage interface {
	// GetPathId returns the path ID at the specified index
	GetPathId(index int) int
}

// VertexSourceInterface represents a vertex source that can be used for path rendering.
type VertexSourceInterface interface {
	// This would typically include methods for path traversal
	// For now, we'll keep it minimal as a placeholder
}

// RenderAllPaths renders multiple paths with different colors.
// This corresponds to AGG's render_all_paths function.
func RenderAllPaths[C any](ras RasterizerInterface, sl ScanlineInterface, renderer RendererInterface[C],
	vertexSource VertexSourceInterface, colorStorage PathColorStorage[C],
	pathIdStorage PathIdStorage, numPaths int,
) {
	// This is a simplified version - in a full implementation, we'd need
	// to define more complete interfaces for the rasterizer's path handling
	for i := 0; i < numPaths; i++ {
		// Reset the rasterizer for this path
		if resettable, ok := ras.(Resettable); ok {
			resettable.Reset()
		}

		// Add the path to the rasterizer
		pathId := pathIdStorage.GetPathId(i)
		if addPathInterface, ok := ras.(interface {
			AddPath(vs VertexSourceInterface, pathId int)
		}); ok {
			addPathInterface.AddPath(vertexSource, pathId)
		}

		// Set the color on the renderer
		color := colorStorage.GetColor(i)
		if colorSetter, ok := renderer.(ColorSetter[C]); ok {
			colorSetter.SetColor(color)
		}

		// Render the scanlines
		RenderScanlines(ras, sl, renderer)
	}
}

// RenderScanlinesCompound renders scanlines using compound rasterizer with multiple styles.
// This corresponds to AGG's render_scanlines_compound function.
// PC is the pointer type constraint that ensures *C has AddWithCover method for color blending.
func RenderScanlinesCompound[C any, PC interface {
	*C
	AddWithCover(src C, cover basics.Int8u)
}](ras CompoundRasterizerInterface, slAA ScanlineInterface,
	slBin ScanlineInterface, ren BaseRendererInterface[C], alloc SpanAllocatorInterface[C],
	styleHandler StyleHandlerInterface[C],
) {
	if !ras.RewindScanlines() {
		return
	}

	minX := ras.MinX()
	maxX := ras.MaxX()
	length := maxX - minX + 2

	// Reset scanlines
	if resetScanline, ok := slAA.(ResettableScanline); ok {
		resetScanline.Reset(minX, maxX)
	}
	if resetScanline, ok := slBin.(ResettableScanline); ok {
		resetScanline.Reset(minX, maxX)
	}

	// Allocate buffers for compound rendering
	colorSpan := alloc.Allocate(length * 2)
	mixBuffer := colorSpan[length:] // Second half of the allocation

	var numStyles int
	for {
		numStyles = ras.SweepStyles()
		if numStyles <= 0 {
			break
		}

		if numStyles == 1 {
			// Optimization for single style - common case
			if ras.SweepScanlineWithStyle(slAA, 0) {
				style := ras.Style(0)
				if styleHandler.IsSolid(style) {
					// Just solid fill
					color := styleHandler.Color(style)
					RenderScanlineAASolid(slAA, ren, color)
				} else {
					// Arbitrary span generator
					renderCompoundSpanGenerated(slAA, ren, alloc, styleHandler, style)
				}
			}
		} else {
			// Multiple styles - use compound rendering
			if ras.SweepScanlineWithStyle(slBin, -1) {
				renderCompoundMultipleStyles[C, PC](ras, slAA, slBin, ren, alloc,
					styleHandler, colorSpan, mixBuffer, minX, numStyles)
			}
		}
	}
}

// renderCompoundSpanGenerated renders a scanline with span generation for compound rendering.
func renderCompoundSpanGenerated[C any](sl ScanlineInterface, ren BaseRendererInterface[C],
	alloc SpanAllocatorInterface[C], styleHandler StyleHandlerInterface[C], style int,
) {
	iter := sl.Begin()
	numSpans := sl.NumSpans()
	y := sl.Y()

	for i := 0; i < numSpans; i++ {
		span := iter.GetSpan()

		colors := alloc.Allocate(span.Len)
		styleHandler.GenerateSpan(colors, span.X, y, span.Len, style)

		ren.BlendColorHspan(span.X, y, span.Len, colors, span.Covers, basics.CoverFull)

		if i < numSpans-1 {
			iter.Next()
		}
	}
}

// renderCompoundMultipleStyles renders scanlines with multiple styles using compound rendering.
// PC is the pointer type constraint that ensures *C has AddWithCover method for color blending.
func renderCompoundMultipleStyles[C any, PC interface {
	*C
	AddWithCover(src C, cover basics.Int8u)
}](ras CompoundRasterizerInterface, slAA ScanlineInterface,
	slBin ScanlineInterface, ren BaseRendererInterface[C], alloc SpanAllocatorInterface[C],
	styleHandler StyleHandlerInterface[C], colorSpan []C, mixBuffer []C,
	minX int, numStyles int,
) {
	// Allocate coverage buffer for tracking accumulated coverage per pixel
	length := len(mixBuffer)
	coverBuffer := ras.AllocateCoverBuffer(length)

	// Clear the mix buffer and cover buffer spans
	iterBin := slBin.Begin()
	numSpansBin := slBin.NumSpans()

	for i := 0; i < numSpansBin; i++ {
		span := iterBin.GetSpan()

		// Clear mix buffer and cover buffer sections for this span
		for j := 0; j < span.Len; j++ {
			var zero C
			mixBuffer[span.X-minX+j] = zero
			coverBuffer[span.X-minX+j] = 0
		}

		if i < numSpansBin-1 {
			iterBin.Next()
		}
	}

	// Process each style
	for styleIndex := 0; styleIndex < numStyles; styleIndex++ {
		style := ras.Style(styleIndex)
		solid := styleHandler.IsSolid(style)

		if ras.SweepScanlineWithStyle(slAA, styleIndex) {
			iter := slAA.Begin()
			numSpans := slAA.NumSpans()

			for i := 0; i < numSpans; i++ {
				span := iter.GetSpan()

				if solid {
					// Solid color processing
					renderCompoundSolidStyle[C, PC](span, styleHandler, style, mixBuffer, coverBuffer, minX)
				} else {
					// Span generator processing
					renderCompoundGeneratedStyle[C, PC](span, slAA, styleHandler, style,
						colorSpan, mixBuffer, coverBuffer, minX, alloc)
				}

				if i < numSpans-1 {
					iter.Next()
				}
			}
		}
	}

	// Emit the blended result
	iterBin = slBin.Begin()
	numSpansBin = slBin.NumSpans()
	y := slBin.Y()

	for i := 0; i < numSpansBin; i++ {
		span := iterBin.GetSpan()

		ren.BlendColorHspan(span.X, y, span.Len, mixBuffer[span.X-minX:span.X-minX+span.Len],
			nil, basics.CoverFull)

		if i < numSpansBin-1 {
			iterBin.Next()
		}
	}
}

// renderCompoundSolidStyle renders a span with solid color for compound rendering.
// PC is the pointer type constraint that ensures *C has AddWithCover method.
func renderCompoundSolidStyle[C any, PC interface {
	*C
	AddWithCover(src C, cover basics.Int8u)
}](span SpanData, styleHandler StyleHandlerInterface[C],
	style int, mixBuffer []C, coverBuffer []basics.Int8u, minX int,
) {
	sourceColor := styleHandler.Color(style)

	for i := 0; i < span.Len; i++ {
		cover := span.Covers[i]
		bufferIndex := span.X - minX + i

		// Check if accumulated coverage would exceed CoverFull
		if uint32(coverBuffer[bufferIndex])+uint32(cover) > uint32(basics.CoverFull) {
			cover = basics.CoverFull - coverBuffer[bufferIndex]
		}

		if cover > 0 {
			PC(&mixBuffer[bufferIndex]).AddWithCover(sourceColor, cover)
			coverBuffer[bufferIndex] += cover
		}
	}
}

// renderCompoundGeneratedStyle renders a span with generated colors for compound rendering.
// PC is the pointer type constraint that ensures *C has AddWithCover method.
func renderCompoundGeneratedStyle[C any, PC interface {
	*C
	AddWithCover(src C, cover basics.Int8u)
}](span SpanData, sl ScanlineInterface,
	styleHandler StyleHandlerInterface[C], style int, colorSpan []C,
	mixBuffer []C, coverBuffer []basics.Int8u, minX int, alloc SpanAllocatorInterface[C],
) {
	colors := alloc.Allocate(span.Len)
	styleHandler.GenerateSpan(colors, span.X, sl.Y(), span.Len, style)

	for i := 0; i < span.Len; i++ {
		cover := span.Covers[i]
		bufferIndex := span.X - minX + i

		// Check if accumulated coverage would exceed CoverFull
		if uint32(coverBuffer[bufferIndex])+uint32(cover) > uint32(basics.CoverFull) {
			cover = basics.CoverFull - coverBuffer[bufferIndex]
		}

		if cover > 0 {
			PC(&mixBuffer[bufferIndex]).AddWithCover(colors[i], cover)
			coverBuffer[bufferIndex] += cover
		}
	}
}
