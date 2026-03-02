// Based on the original AGG examples: aa_demo.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/path"
	renscan "agg_go/internal/renderer/scanline"
)

// enlargedPixel holds a single pixel to be drawn after the scanline sweep.
type enlargedPixel struct {
	x, y  float64
	color agg.Color
}

// EnlargedRenderer implements the AGG Scanline Renderer interface.
// It draws a "zoomed in" view where each physical pixel is rendered as a large square.
//
// Drawing via ctx.FillRectangle goes through the shared rasterizer, which would
// corrupt the ongoing sweep. So Render only collects pixels; Flush draws them.
type EnlargedRenderer struct {
	ctx       *agg.Context
	pixelSize float64
	color     agg.Color
	pixels    []enlargedPixel
}

func (r *EnlargedRenderer) Prepare() { r.pixels = r.pixels[:0] }

func (r *EnlargedRenderer) SetColor(c color.RGBA8[color.Linear]) {
	r.color = agg.NewColorRGBA8(c)
}

func (r *EnlargedRenderer) Render(sl renscan.ScanlineInterface) {
	y := sl.Y()
	numSpans := sl.NumSpans()
	it := sl.Begin()

	//nolint:intrange
	for i := 0; i < numSpans; i++ {
		span := it.GetSpan()
		x := span.X
		numPix := span.Len
		covers := span.Covers

		// Handle solid spans (negative len means solid with single cover value)
		if numPix < 0 {
			numPix = -numPix
			cover := covers[0]
			alpha := (uint16(cover) * uint16(r.color.A)) >> 8
			c := agg.NewColor(r.color.R, r.color.G, r.color.B, uint8(alpha))
			for j := 0; j < numPix; j++ {
				r.pixels = append(r.pixels, enlargedPixel{
					x:     float64(x+j) * r.pixelSize,
					y:     float64(y) * r.pixelSize,
					color: c,
				})
			}
		} else {
			for j := 0; j < numPix; j++ {
				cover := covers[j]
				alpha := (uint16(cover) * uint16(r.color.A)) >> 8
				r.pixels = append(r.pixels, enlargedPixel{
					x:     float64(x+j) * r.pixelSize,
					y:     float64(y) * r.pixelSize,
					color: agg.NewColor(r.color.R, r.color.G, r.color.B, uint8(alpha)),
				})
			}
		}

		if i < numSpans-1 {
			it.Next()
		}
	}
}

// Flush draws all collected pixels. Must be called after ScanlineRender returns.
func (r *EnlargedRenderer) Flush() {
	for _, p := range r.pixels {
		r.ctx.SetColor(p.color)
		r.ctx.FillRectangle(p.x, p.y, r.pixelSize, r.pixelSize)
	}
}

var (
	aaTriangleX = [3]float64{20, 500, 143}
	aaTriangleY = [3]float64{100, 50, 310}
	aaPixelSize = 32.0
	aaSelected  = -1
	aaDragDX    = 0.0
	aaDragDY    = 0.0
)

func drawAADemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Draw the enlarged pixel representation
	ps := path.NewPathStorageStl()
	ps.MoveTo(aaTriangleX[0]/aaPixelSize, aaTriangleY[0]/aaPixelSize)
	ps.LineTo(aaTriangleX[1]/aaPixelSize, aaTriangleY[1]/aaPixelSize)
	ps.LineTo(aaTriangleX[2]/aaPixelSize, aaTriangleY[2]/aaPixelSize)
	ps.ClosePolygon(basics.PathFlagsNone)

	enlargedRen := &EnlargedRenderer{
		ctx:       ctx,
		pixelSize: aaPixelSize,
		color:     agg.Black,
	}

	// Use adapter to fix Rewind(uint) vs Rewind(uint32) mismatch
	adapter := &pathSourceAdapter{ps: ps}

	// Set up the rasterizer
	ras := agg2d.GetInternalRasterizer()
	ras.Reset()
	ras.AddPath(adapter, 0)

	// Collect pixel data during sweep (drawing would reset the shared rasterizer)
	agg2d.ScanlineRender(ras, enlargedRen)
	// Now flush: draw all collected enlarged pixels
	enlargedRen.Flush()

	// 2. Draw the "real size" triangle outline for comparison
	ctx.SetColor(agg.NewColor(0, 150, 160, 200))
	ctx.SetLineWidth(2.0)
	ctx.BeginPath()
	ctx.MoveTo(aaTriangleX[0], aaTriangleY[0])
	ctx.LineTo(aaTriangleX[1], aaTriangleY[1])
	ctx.LineTo(aaTriangleX[2], aaTriangleY[2])
	ctx.ClosePath()
	ctx.Stroke()

	// 3. Draw interactive handles
	for i := 0; i < 3; i++ {
		ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
		ctx.FillCircle(aaTriangleX[i], aaTriangleY[i], 5)
		ctx.SetColor(agg.Black)
		ctx.DrawCircle(aaTriangleX[i], aaTriangleY[i], 5)
	}
}

func handleAAMouseDown(x, y float64) bool {
	aaSelected = -1
	for i := 0; i < 3; i++ {
		dist := math.Sqrt(math.Pow(x-aaTriangleX[i], 2) + math.Pow(y-aaTriangleY[i], 2))
		if dist < 10 {
			aaSelected = i
			aaDragDX = x - aaTriangleX[i]
			aaDragDY = y - aaTriangleY[i]
			return true
		}
	}
	return false
}

func handleAAMouseMove(x, y float64) bool {
	if aaSelected != -1 {
		aaTriangleX[aaSelected] = x - aaDragDX
		aaTriangleY[aaSelected] = y - aaDragDY
		return true
	}
	return false
}

func handleAAMouseUp() {
	aaSelected = -1
}
