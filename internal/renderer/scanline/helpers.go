// Package scanline provides helper rendering functions for AGG scanline rendering.
package scanline

import (
	"agg_go/internal/basics"
)

// RenderScanlines is a generic scanline rendering function that works with any renderer.
// This corresponds to AGG's render_scanlines function.
func RenderScanlines(ras RasterizerInterface, sl ScanlineInterface, renderer RendererInterface) {
	if !ras.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	if resetInterface, ok := sl.(interface{ Reset(minX, maxX int) }); ok {
		resetInterface.Reset(ras.MinX(), ras.MaxX())
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
type PathColorStorage interface {
	// GetColor returns the color at the specified index
	GetColor(index int) interface{}
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
func RenderAllPaths(ras RasterizerInterface, sl ScanlineInterface, renderer interface{},
	vertexSource VertexSourceInterface, colorStorage PathColorStorage,
	pathIdStorage PathIdStorage, numPaths int,
) {
	// This is a simplified version - in a full implementation, we'd need
	// to define more complete interfaces for the rasterizer's path handling
	for i := 0; i < numPaths; i++ {
		// Reset the rasterizer for this path
		if resetInterface, ok := ras.(interface{ Reset() }); ok {
			resetInterface.Reset()
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
		if colorInterface, ok := renderer.(interface{ SetColor(color interface{}) }); ok {
			colorInterface.SetColor(color)
		}

		// Render the scanlines
		if rendererInterface, ok := renderer.(RendererInterface); ok {
			RenderScanlines(ras, sl, rendererInterface)
		}
	}
}

// RenderScanlinesCompound renders scanlines using compound rasterizer with multiple styles.
// This corresponds to AGG's render_scanlines_compound function.
func RenderScanlinesCompound(ras CompoundRasterizerInterface, slAA ScanlineInterface,
	slBin ScanlineInterface, ren BaseRendererInterface, alloc SpanAllocatorInterface,
	styleHandler StyleHandlerInterface,
) {
	if !ras.RewindScanlines() {
		return
	}

	minX := ras.MinX()
	maxX := ras.MaxX()
	length := maxX - minX + 2

	// Reset scanlines
	if resetInterface, ok := slAA.(interface{ Reset(minX, maxX int) }); ok {
		resetInterface.Reset(minX, maxX)
	}
	if resetInterface, ok := slBin.(interface{ Reset(minX, maxX int) }); ok {
		resetInterface.Reset(minX, maxX)
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
				renderCompoundMultipleStyles(ras, slAA, slBin, ren, alloc,
					styleHandler, colorSpan, mixBuffer, minX, numStyles)
			}
		}
	}
}

// renderCompoundSpanGenerated renders a scanline with span generation for compound rendering.
func renderCompoundSpanGenerated(sl ScanlineInterface, ren BaseRendererInterface,
	alloc SpanAllocatorInterface, styleHandler StyleHandlerInterface, style int,
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
func renderCompoundMultipleStyles(ras CompoundRasterizerInterface, slAA ScanlineInterface,
	slBin ScanlineInterface, ren BaseRendererInterface, alloc SpanAllocatorInterface,
	styleHandler StyleHandlerInterface, colorSpan []interface{}, mixBuffer []interface{},
	minX int, numStyles int,
) {
	// Clear the mix buffer spans
	iterBin := slBin.Begin()
	numSpansBin := slBin.NumSpans()

	for i := 0; i < numSpansBin; i++ {
		span := iterBin.GetSpan()

		// Clear mix buffer section for this span
		for j := 0; j < span.Len; j++ {
			mixBuffer[span.X-minX+j] = nil
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
					renderCompoundSolidStyle(span, styleHandler, style, mixBuffer, minX)
				} else {
					// Span generator processing
					renderCompoundGeneratedStyle(span, slAA, styleHandler, style,
						colorSpan, mixBuffer, minX, alloc)
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

		ren.BlendColorHspan(span.X, y, span.Len, mixBuffer[span.X-minX:],
			nil, basics.CoverFull)

		if i < numSpansBin-1 {
			iterBin.Next()
		}
	}
}

// renderCompoundSolidStyle renders a span with solid color for compound rendering.
func renderCompoundSolidStyle(span SpanData, styleHandler StyleHandlerInterface,
	style int, mixBuffer []interface{}, minX int,
) {
	color := styleHandler.Color(style)

	for i := 0; i < span.Len; i++ {
		cover := span.Covers[i]
		bufferIndex := span.X - minX + i

		if cover == basics.CoverFull {
			mixBuffer[bufferIndex] = color
		} else {
			// Blend with existing color in mix buffer
			// This is a simplified version - real implementation would need
			// proper color blending based on the color type
			if mixBuffer[bufferIndex] == nil {
				mixBuffer[bufferIndex] = color
			}
		}
	}
}

// renderCompoundGeneratedStyle renders a span with generated colors for compound rendering.
func renderCompoundGeneratedStyle(span SpanData, sl ScanlineInterface,
	styleHandler StyleHandlerInterface, style int, colorSpan []interface{},
	mixBuffer []interface{}, minX int, alloc SpanAllocatorInterface,
) {
	colors := alloc.Allocate(span.Len)
	styleHandler.GenerateSpan(colors, span.X, sl.Y(), span.Len, style)

	for i := 0; i < span.Len; i++ {
		cover := span.Covers[i]
		bufferIndex := span.X - minX + i

		if cover == basics.CoverFull {
			mixBuffer[bufferIndex] = colors[i]
		} else {
			// Blend with existing color in mix buffer
			// This is a simplified version - real implementation would need
			// proper color blending based on the color type
			if mixBuffer[bufferIndex] == nil {
				mixBuffer[bufferIndex] = colors[i]
			}
		}
	}
}
