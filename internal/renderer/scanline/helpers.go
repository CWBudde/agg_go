package scanline

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
)

// RenderScanlines is the canonical AGG-style helper that sweeps a rasterizer
// and feeds every produced scanline to a renderer.
func RenderScanlines[C any](ras RasterizerInterface, sl ScanlineInterface, renderer RendererInterface[C]) {
	if !ras.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	sl.Reset(ras.MinX(), ras.MaxX())

	// Prepare the renderer
	renderer.Prepare()

	// Sweep through all scanlines
	for ras.SweepScanline(sl) {
		renderer.Render(sl)
	}
}

// PathColorStorage provides per-path colors to RenderAllPaths.
type PathColorStorage[C any] interface {
	// GetColor returns the color at the specified index
	GetColor(index int) C
}

// PathIDStorage provides the path IDs paired with colors in RenderAllPaths.
type PathIDStorage interface {
	// GetPathID returns the path ID at the specified index.
	GetPathID(index int) uint32
}

// MultiPathRasterizerInterface extends RasterizerInterface with path ingestion.
type MultiPathRasterizerInterface interface {
	RasterizerInterface
	Reset()
	AddPath(vs rasterizer.VertexSource, pathID uint32)
}

// RenderAllPaths renders multiple paths by repeatedly resetting the rasterizer,
// adding one path, setting its color, and invoking RenderScanlines.
func RenderAllPaths[C any](ras MultiPathRasterizerInterface, sl ScanlineInterface, renderer RendererInterface[C],
	vertexSource rasterizer.VertexSource, colorStorage PathColorStorage[C],
	pathIDStorage PathIDStorage, numPaths int,
) {
	for i := 0; i < numPaths; i++ {
		ras.Reset()
		ras.AddPath(vertexSource, pathIDStorage.GetPathID(i))
		renderer.SetColor(colorStorage.GetColor(i))
		RenderScanlines(ras, sl, renderer)
	}
}

// RenderScanlinesCompound sweeps a style-aware rasterizer and resolves each
// style either as a solid fill or as generated span data before blending the
// composed result.
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
	slAA.Reset(minX, maxX)
	slBin.Reset(minX, maxX)

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
	iter := sl.BeginIterator()
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
	// Clear only the mix buffer spans, matching AGG's render_scanlines_compound.
	iterBin := slBin.BeginIterator()
	numSpansBin := slBin.NumSpans()

	for i := 0; i < numSpansBin; i++ {
		span := iterBin.GetSpan()

		for j := 0; j < span.Len; j++ {
			var zero C
			mixBuffer[span.X-minX+j] = zero
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
			iter := slAA.BeginIterator()
			numSpans := slAA.NumSpans()

			for i := 0; i < numSpans; i++ {
				span := iter.GetSpan()

				if solid {
					renderCompoundSolidStyle[C, PC](span, styleHandler, style, mixBuffer, minX)
				} else {
					renderCompoundGeneratedStyle[C, PC](span, slAA, styleHandler, style,
						colorSpan, mixBuffer, minX, alloc)
				}

				if i < numSpans-1 {
					iter.Next()
				}
			}
		}
	}

	// Emit the blended result
	iterBin = slBin.BeginIterator()
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
	style int, mixBuffer []C, minX int,
) {
	sourceColor := styleHandler.Color(style)

	for i := 0; i < span.Len; i++ {
		cover := span.Covers[i]
		bufferIndex := span.X - minX + i

		if cover == basics.CoverFull {
			mixBuffer[bufferIndex] = sourceColor
		} else if cover > 0 {
			PC(&mixBuffer[bufferIndex]).AddWithCover(sourceColor, cover)
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
	mixBuffer []C, minX int, alloc SpanAllocatorInterface[C],
) {
	colors := alloc.Allocate(span.Len)
	styleHandler.GenerateSpan(colors, span.X, sl.Y(), span.Len, style)

	for i := 0; i < span.Len; i++ {
		cover := span.Covers[i]
		bufferIndex := span.X - minX + i

		if cover == basics.CoverFull {
			mixBuffer[bufferIndex] = colors[i]
		} else if cover > 0 {
			PC(&mixBuffer[bufferIndex]).AddWithCover(colors[i], cover)
		}
	}
}
