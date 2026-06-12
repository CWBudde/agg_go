package scanline

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/rasterizer"
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

// LayeredCompoundRasterizerInterface is the sweep surface required by
// RenderScanlinesCompoundLayered. It matches the concrete method set of
// rasterizer.RasterizerCompoundAA.
type LayeredCompoundRasterizerInterface interface {
	RewindScanlines() bool
	MinX() int
	MaxX() int
	SweepStyles() uint32
	Style(styleIdx uint32) uint32
	ScanlineStart() int
	ScanlineLength() uint32
	AllocateCoverBuffer(length int) []basics.Int8u
	SweepScanline(sl rasterizer.CompoundScanlineInterface, styleIdx int) bool
}

// scanlineCoverAdapter adapts a renderer-side Scanline (uint covers) to the
// rasterizer.CompoundScanlineInterface (Int8u covers) consumed by the
// compound rasterizer's sweep methods.
type scanlineCoverAdapter struct{ sl ScanlineInterface }

func (a *scanlineCoverAdapter) ResetSpans()                      { a.sl.ResetSpans() }
func (a *scanlineCoverAdapter) AddCell(x int, c basics.Int8u)    { a.sl.AddCell(x, uint(c)) }
func (a *scanlineCoverAdapter) AddSpan(x, l int, c basics.Int8u) { a.sl.AddSpan(x, l, uint(c)) }
func (a *scanlineCoverAdapter) Finalize(y int)                   { a.sl.Finalize(y) }
func (a *scanlineCoverAdapter) NumSpans() int                    { return a.sl.NumSpans() }

// RenderScanlinesCompoundLayered is the Go port of AGG's
// render_scanlines_compound_layered (agg_renderer_scanline.h). Unlike
// RenderScanlinesCompound it accumulates per-pixel coverage in the
// rasterizer's cover buffer, so overlapping styles within one scanline
// saturate at full coverage instead of blending twice.
func RenderScanlinesCompoundLayered[C any, PC interface {
	*C
	AddWithCover(src C, cover basics.Int8u)
}](ras LayeredCompoundRasterizerInterface, slAA ScanlineInterface,
	ren BaseRendererInterface[C], alloc SpanAllocatorInterface[C],
	styleHandler StyleHandlerInterface[C],
) {
	if !ras.RewindScanlines() {
		return
	}

	minX := ras.MinX()
	length := ras.MaxX() - minX + 2
	slAA.Reset(minX, ras.MaxX())

	colorSpan := alloc.Allocate(length * 2)
	mixBuffer := colorSpan[length:]
	coverBuffer := ras.AllocateCoverBuffer(length)
	slRas := &scanlineCoverAdapter{sl: slAA}

	for {
		numStyles := ras.SweepStyles()
		if numStyles == 0 {
			break
		}

		if numStyles == 1 {
			// Optimization for a single style. Happens often.
			if ras.SweepScanline(slRas, 0) {
				style := int(ras.Style(0))
				if styleHandler.IsSolid(style) {
					RenderScanlineAASolid(slAA, ren, styleHandler.Color(style))
				} else {
					y := slAA.Y()
					iter := slAA.BeginIterator()
					numSpans := slAA.NumSpans()
					for s := 0; s < numSpans; s++ {
						span := iter.GetSpan()
						styleHandler.GenerateSpan(colorSpan[:span.Len], span.X, y, span.Len, style)
						ren.BlendColorHspan(span.X, y, span.Len,
							colorSpan[:span.Len], span.Covers, basics.CoverFull)
						if s < numSpans-1 {
							iter.Next()
						}
					}
				}
			}
			continue
		}

		slStart := ras.ScanlineStart()
		slLen := int(ras.ScanlineLength())
		if slLen == 0 {
			continue
		}

		clear(mixBuffer[slStart-minX : slStart-minX+slLen])
		clear(coverBuffer[slStart-minX : slStart-minX+slLen])

		// Mirrors C++: stays at the sentinel (and gets clipped away by the
		// renderer base) if no style sweeps a scanline.
		slY := math.MaxInt32

		for i := uint32(0); i < numStyles; i++ {
			style := int(ras.Style(i))
			solid := styleHandler.IsSolid(style)

			if !ras.SweepScanline(slRas, int(i)) {
				continue
			}
			slY = slAA.Y()

			iter := slAA.BeginIterator()
			numSpans := slAA.NumSpans()
			for s := 0; s < numSpans; s++ {
				span := iter.GetSpan()

				if solid {
					c := styleHandler.Color(style)
					accumulateCompoundSpan[C, PC](span, mixBuffer, coverBuffer, minX,
						func(int) C { return c })
				} else {
					spanColors := colorSpan[:span.Len]
					styleHandler.GenerateSpan(spanColors, span.X, slY, span.Len, style)
					accumulateCompoundSpan[C, PC](span, mixBuffer, coverBuffer, minX,
						func(j int) C { return spanColors[j] })
				}

				if s < numSpans-1 {
					iter.Next()
				}
			}
		}

		ren.BlendColorHspan(slStart, slY, slLen,
			mixBuffer[slStart-minX:slStart-minX+slLen], nil, basics.CoverFull)
	}
}

// accumulateCompoundSpan adds one coverage span into the mix buffer, clamping
// the accumulated coverage at CoverFull like AGG's layered compound renderer.
func accumulateCompoundSpan[C any, PC interface {
	*C
	AddWithCover(src C, cover basics.Int8u)
}](span SpanData, mixBuffer []C, coverBuffer []basics.Int8u, minX int, colorAt func(int) C) {
	for j := 0; j < span.Len; j++ {
		idx := span.X - minX + j
		cover := span.Covers[j]
		if int(coverBuffer[idx])+int(cover) > basics.CoverFull {
			cover = basics.Int8u(basics.CoverFull) - coverBuffer[idx]
		}
		if cover > 0 {
			PC(&mixBuffer[idx]).AddWithCover(colorAt(j), cover)
			coverBuffer[idx] += cover
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
