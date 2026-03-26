// Based on the original AGG examples: aa_demo.cpp.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

// aaRasType is the concrete rasterizer type used throughout this demo.
type aaRasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newAARasterizer() *aaRasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// aaRendererBase is the concrete renderer base type used throughout this demo.
type aaRendererBase = renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]

// rendererEnlargedAA draws each scanline pixel as a large square via its own
// rasterizer, matching the C++ renderer_enlarged exactly.
type rendererEnlargedAA struct {
	ras   *aaRasType
	sl    *scanline.ScanlineU8
	renRb *aaRendererBase
	size  float64
	col   color.RGBA8[color.Linear]
}

func newRendererEnlargedAA(renRb *aaRendererBase, size float64) *rendererEnlargedAA {
	return &rendererEnlargedAA{
		ras:   newAARasterizer(),
		sl:    scanline.NewScanlineU8(),
		renRb: renRb,
		size:  size,
	}
}

func (r *rendererEnlargedAA) Prepare() {}

func (r *rendererEnlargedAA) SetColor(c color.RGBA8[color.Linear]) { r.col = c }

func (r *rendererEnlargedAA) Render(sl renscan.ScanlineInterface) {
	y := sl.Y()
	it := sl.BeginIterator()
	for i, n := 0, sl.NumSpans(); i < n; i++ {
		span := it.GetSpan()
		x := span.X
		numPix := span.Len
		covers := span.Covers
		solid := numPix < 0
		if solid {
			numPix = -numPix
		}
		for j := 0; j < numPix; j++ {
			cover := covers[0]
			if !solid {
				cover = covers[j]
			}
			a := (uint16(cover) * uint16(r.col.A)) >> 8
			r.drawSquare(float64(x+j), float64(y),
				color.RGBA8[color.Linear]{R: r.col.R, G: r.col.G, B: r.col.B, A: uint8(a)})
		}
		if i < n-1 {
			it.Next()
		}
	}
}

func (r *rendererEnlargedAA) drawSquare(x, y float64, c color.RGBA8[color.Linear]) {
	r.ras.Reset()
	r.ras.MoveToD(x*r.size, y*r.size)
	r.ras.LineToD(x*r.size+r.size, y*r.size)
	r.ras.LineToD(x*r.size+r.size, y*r.size+r.size)
	r.ras.LineToD(x*r.size, y*r.size+r.size)
	renscan.RenderScanlinesAASolid(r.ras, r.sl, r.renRb, c)
}

// pathStlVSAA wraps PathStorageStl as a conv.VertexSource.
type pathStlVSAA struct{ ps *path.PathStorageStl }

func (a *pathStlVSAA) Rewind(id uint) { a.ps.Rewind(id) }
func (a *pathStlVSAA) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

// convVSAA adapts a conv.VertexSource to the rasterizer's VertexSource interface.
type convVSAA struct{ src conv.VertexSource }

func (a *convVSAA) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convVSAA) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

var (
	aaTriangleX = [3]float64{20, 728, 170}
	aaTriangleY = [3]float64{100, 75, 547}
	aaPixelSize = 32.0
	aaSelected  = -1
	aaDragDX    = 0.0
	aaDragDY    = 0.0
)

func drawAADemo() {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	sizeMul := aaPixelSize

	ras := newAARasterizer()
	sl := scanline.NewScanlineU8()

	// 1. Enlarged-pixel rendering.
	renEnlarged := newRendererEnlargedAA(mainRb, sizeMul)
	renEnlarged.SetColor(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras.Reset()
	ras.MoveToD(aaTriangleX[0]/sizeMul, aaTriangleY[0]/sizeMul)
	ras.LineToD(aaTriangleX[1]/sizeMul, aaTriangleY[1]/sizeMul)
	ras.LineToD(aaTriangleX[2]/sizeMul, aaTriangleY[2]/sizeMul)
	renscan.RenderScanlines(ras, sl, renEnlarged)

	// 2. Full-scale triangle outline in teal via conv_stroke.
	teal := color.RGBA8[color.Linear]{R: 0, G: 150, B: 160, A: 200}
	edges := [3][2]int{{0, 1}, {1, 2}, {2, 0}}
	for _, e := range edges {
		ps := path.NewPathStorageStl()
		ps.MoveTo(aaTriangleX[e[0]], aaTriangleY[e[0]])
		ps.LineTo(aaTriangleX[e[1]], aaTriangleY[e[1]])
		stroke := conv.NewConvStroke(&pathStlVSAA{ps: ps})
		stroke.SetWidth(2.0)
		ras.Reset()
		ras.AddPath(&convVSAA{src: stroke}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, teal)
	}

	// 3. Draw interactive handles.
	handleCol := color.RGBA8[color.Linear]{R: 204, G: 51, B: 26, A: 153}
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	for i := 0; i < 3; i++ {
		drawAAHandle(ras, sl, mainRb, aaTriangleX[i], aaTriangleY[i], handleCol, black)
	}

	applyLinearToSRGB(img)
}

// drawAAHandle renders a filled circle and its outline at (cx, cy).
func drawAAHandle(
	ras *aaRasType,
	sl *scanline.ScanlineU8,
	renRb *aaRendererBase,
	cx, cy float64,
	fillCol, outlineCol color.RGBA8[color.Linear],
) {
	const r = 5.0
	const steps = 32
	// Filled circle.
	ras.Reset()
	ras.MoveToD(cx+r, cy)
	for i := 1; i < steps; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(steps)
		ras.LineToD(cx+r*math.Cos(angle), cy+r*math.Sin(angle))
	}
	renscan.RenderScanlinesAASolid(ras, sl, renRb, fillCol)

	// Outline via conv_stroke.
	ps := path.NewPathStorageStl()
	ps.MoveTo(cx+r, cy)
	for i := 1; i < steps; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(steps)
		ps.LineTo(cx+r*math.Cos(angle), cy+r*math.Sin(angle))
	}
	ps.ClosePolygon(basics.PathFlagsNone)
	stroke := conv.NewConvStroke(&pathStlVSAA{ps: ps})
	stroke.SetWidth(1.0)
	ras.Reset()
	ras.AddPath(&convVSAA{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, renRb, outlineCol)
}

func handleAAMouseDown(x, y float64) bool {
	aaSelected = -1
	for i := 0; i < 3; i++ {
		dist := math.Sqrt((x-aaTriangleX[i])*(x-aaTriangleX[i]) + (y-aaTriangleY[i])*(y-aaTriangleY[i]))
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
