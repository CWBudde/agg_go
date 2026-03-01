package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/path"
	renscan "agg_go/internal/renderer/scanline"
)

// EnlargedRenderer implements the AGG Scanline Renderer interface.
// It draws a "zoomed in" view where each physical pixel is rendered as a large square.
type EnlargedRenderer struct {
	ctx       *agg.Context
	pixelSize float64
	color     agg.Color
}

func (r *EnlargedRenderer) Prepare() {}

func (r *EnlargedRenderer) SetColor(c color.RGBA8[color.Linear]) {
	r.color = agg.NewColorRGBA8(c)
}

func (r *EnlargedRenderer) Render(sl renscan.ScanlineInterface) {
	y := sl.Y()
	it := sl.Begin()
	for {
		span := it.GetSpan()
		x := span.X
		numPix := span.Len
		covers := span.Covers
		
		for i := 0; i < numPix; i++ {
			cover := covers[i]
			alpha := (uint16(cover) * uint16(r.color.A)) >> 8
			
			// Draw the "enlarged pixel" square
			r.ctx.SetColor(agg.NewColor(r.color.R, r.color.G, r.color.B, uint8(alpha)))
			r.ctx.FillRectangle(
				float64(x+i)*r.pixelSize, 
				float64(y)*r.pixelSize, 
				r.pixelSize, 
				r.pixelSize,
			)
		}

		if !it.Next() {
			break
		}
	}
}

var (
	aaTriangleX = [3]float64{57, 369, 143}
	aaTriangleY = [3]float64{100, 170, 310}
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
	adapter := &AAAdapter{ps: ps}
	
	// Set up the rasterizer
	ras := agg2d.GetInternalRasterizer()
	ras.Reset()
	ras.AddPath(adapter, 0)
	
	// Use our custom renderer to "zoom in" on the pixels
	agg2d.ScanlineRender(ras, enlargedRen)

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
