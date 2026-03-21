// Port of AGG C++ aa_test.cpp – anti-aliasing quality test.
//
// Renders radial lines, ellipses at various sizes, gradient lines, and
// gradient triangles on a black background. Note: the C++ uses flip_y=false,
// so no y-flip is needed in the output.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
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
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	frameWidth  = 480
	frameHeight = 350
)

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters
// ---------------------------------------------------------------------------

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// Vertex-source adapters
// ---------------------------------------------------------------------------

type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	var vx, vy float64
	cmd := ev.e.Vertex(&vx, &vy)
	*x, *y = vx, vy
	return uint32(cmd)
}

type pathStlVS struct{ ps *path.PathStorageStl }

func (a *pathStlVS) Rewind(id uint) { a.ps.Rewind(id) }
func (a *pathStlVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

type convVS struct{ src conv.VertexSource }

func (a *convVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// dashedLine draws a stroked (optionally dashed) line.
// ---------------------------------------------------------------------------

func drawStrokedLine(
	ras *rasType,
	sl *scanline.ScanlineU8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	x1, y1, x2, y2, lineWidth float64,
	c color.RGBA8[color.Linear],
) {
	ps := path.NewPathStorageStl()
	ps.MoveTo(x1+0.5, y1+0.5)
	ps.LineTo(x2+0.5, y2+0.5)
	stroke := conv.NewConvStroke(&pathStlVS{ps: ps})
	stroke.SetWidth(lineWidth)
	stroke.SetLineCap(basics.RoundCap)
	ras.Reset()
	ras.AddPath(&convVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, c)
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	// Black background.
	mainRb.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineU8()

	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	whiteAlpha := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 51} // 0.2 * 255

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0
	radius := math.Min(cx, cy)

	// Radial line test: 180 lines from centre outward.
	for i := 180; i > 0; i-- {
		n := 2.0 * basics.Pi * float64(i) / 180.0
		drawStrokedLine(ras, sl, mainRb,
			cx+radius*math.Sin(n), cy+radius*math.Cos(n),
			cx, cy,
			1.0, whiteAlpha)
	}

	// Integral point sizes 1..20.
	for i := 1; i <= 20; i++ {
		ell := shapes.NewEllipseWithParams(
			20+float64(i*(i+1))+0.5,
			20.5,
			float64(i)/2.0, float64(i)/2.0,
			uint32(8+i), false,
		)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, white)

		// Fractional point sizes 0..2.
		ell2 := shapes.NewEllipseWithParams(
			18+float64(i*4)+0.5, 33+0.5,
			float64(i)/20.0, float64(i)/20.0,
			8, false,
		)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell2}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, white)

		// Fractional point positioning.
		ell3 := shapes.NewEllipseWithParams(
			18+float64(i*4)+float64(i-1)/10.0+0.5,
			27+float64(i-1)/10.0+0.5,
			0.5, 0.5, 8, false,
		)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell3}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, white)

		// Integral line widths 1..20 (solid white lines, no gradient).
		fi := float64(i)
		c := color.RGBA8[color.Linear]{
			R: uint8(float64(i%2) * 255),
			G: uint8(float64(i%3) * 0.5 * 255),
			B: uint8(float64(i%5) * 0.25 * 255),
			A: 255,
		}
		x1 := 20 + fi*(fi+1)
		y1 := 40.5
		x2 := 20 + fi*(fi+1) + (fi-1)*4
		y2 := 100.5
		drawStrokedLine(ras, sl, mainRb, x1, y1, x2, y2, fi, c)

		// Fractional line lengths H (white lines).
		x1 = 17.5 + fi*4
		y1 = 107
		x2 = 17.5 + fi*4 + fi/6.66666667
		y2 = 107
		drawStrokedLine(ras, sl, mainRb, x1, y1, x2, y2, 1.0, white)

		// Fractional line lengths V.
		x1 = 18 + fi*4
		y1 = 112.5
		x2 = 18 + fi*4
		y2 = 112.5 + fi/6.66666667
		drawStrokedLine(ras, sl, mainRb, x1, y1, x2, y2, 1.0, white)

		// Fractional line positioning (red).
		red := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}
		x1 = 21.5
		y1 = 120 + (fi-1)*3.1
		x2 = 52.5
		y2 = 120 + (fi-1)*3.1
		drawStrokedLine(ras, sl, mainRb, x1, y1, x2, y2, 1.0, red)

		// Fractional line width 2..0 (green).
		green := color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255}
		x1 = 52.5
		y1 = 118 + fi*3
		x2 = 83.5
		y2 = 118 + fi*3
		drawStrokedLine(ras, sl, mainRb, x1, y1, x2, y2, 2.0-(fi-1)/10.0, green)

		// Stippled fractional width 2..0 (blue) - simplified as solid.
		blue := color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255}
		x1 = 83.5
		y1 = 119 + fi*3
		x2 = 114.5
		y2 = 119 + fi*3
		drawStrokedLine(ras, sl, mainRb, x1, y1, x2, y2, 2.0-(fi-1)/10.0, blue)

		if i <= 10 {
			// Integral line width, horz aligned.
			drawStrokedLine(ras, sl, mainRb,
				125.5, 119.5+float64(i+2)*(fi/2.0),
				135.5, 119.5+float64(i+2)*(fi/2.0),
				fi, white)
		}

		// Fractional line width 0..2, 1 px H.
		drawStrokedLine(ras, sl, mainRb,
			17.5+fi*4, 192, 18.5+fi*4, 192,
			fi/10.0, white)

		// Fractional line positioning, 1 px H.
		drawStrokedLine(ras, sl, mainRb,
			17.5+fi*4+(fi-1)/10.0, 186,
			18.5+fi*4+(fi-1)/10.0, 186,
			1.0, white)
	}

	// Triangles.
	for i := 1; i <= 13; i++ {
		fi := float64(i)
		c := color.RGBA8[color.Linear]{
			R: uint8(float64(i%2) * 255),
			G: uint8(float64(i%3) * 0.5 * 255),
			B: uint8(float64(i%5) * 0.25 * 255),
			A: 255,
		}
		ras.Reset()
		ras.MoveToD(float64(w)-150, float64(h)-20-fi*(fi+1.5))
		ras.LineToD(float64(w)-20, float64(h)-20-fi*(fi+1))
		ras.LineToD(float64(w)-20, float64(h)-20-fi*(fi+2))
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, c)
	}

	// No y-flip: C++ aa_test uses flip_y=false.
	copy(img.Data, workBuf)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AA Test",
		Width:  frameWidth,
		Height: frameHeight,
	}, &demo{})
}
