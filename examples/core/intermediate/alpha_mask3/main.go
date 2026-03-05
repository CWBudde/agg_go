// Port of AGG C++ alpha_mask3.cpp – polygon-masked rendering with spiral.
//
// Creates an alpha mask from a star/spiral polygon, then renders colored shapes
// through it. Default: AND operation (mask shows interior), polygon type 3
// (Great Britain + spiral shape).
package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
)

const (
	width  = 800
	height = 600
)

// Scanline/rasterizer bridge types.

type rasAdp3 struct {
	ras interface {
		RewindScanlines() bool
		SweepScanline(sl rasterizer.ScanlineInterface) bool
		MinX() int
		MaxX() int
	}
}

func (r *rasAdp3) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *rasAdp3) MinX() int             { return r.ras.MinX() }
func (r *rasAdp3) MaxX() int             { return r.ras.MaxX() }

type slAdpP8v3 struct{ sl *scanline.ScanlineP8 }

func (a *slAdpP8v3) ResetSpans()                 { a.sl.ResetSpans() }
func (a *slAdpP8v3) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *slAdpP8v3) AddSpan(x, l int, c uint32)  { a.sl.AddSpan(x, l, uint(c)) }
func (a *slAdpP8v3) Finalize(y int)              { a.sl.Finalize(y) }
func (a *slAdpP8v3) NumSpans() int               { return a.sl.NumSpans() }

func (r *rasAdp3) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*slWrapP8v3); ok {
		return r.ras.SweepScanline(&slAdpP8v3{sl: w.sl})
	}
	return false
}

type slWrapP8v3 struct{ sl *scanline.ScanlineP8 }

func (w *slWrapP8v3) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *slWrapP8v3) Y() int               { return w.sl.Y() }
func (w *slWrapP8v3) NumSpans() int        { return w.sl.NumSpans() }

type spanIterP8v3 struct {
	spans []scanline.SpanP8
	idx   int
}

func (it *spanIterP8v3) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIterP8v3) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *slWrapP8v3) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIterP8v3{nil, 0}
	}
	return &spanIterP8v3{spans, 0}
}

// buildSpiralPath creates a spiral polygon used as the mask shape.
func buildSpiralPath(cx, cy, r1, r2 float64, numSteps int, startAngle float64) *path.PathStorageStl {
	ps := path.NewPathStorageStl()
	step := 2 * math.Pi / float64(numSteps)
	for i := 0; i <= numSteps; i++ {
		a := startAngle + float64(i)*step
		t := float64(i) / float64(numSteps)
		r := r1 + (r2-r1)*t
		x := cx + r*math.Cos(a)
		y := cy + r*math.Sin(a)
		if i == 0 {
			ps.MoveTo(x, y)
		} else {
			ps.LineTo(x, y)
		}
	}
	// Close with a straight line back to centre.
	ps.LineTo(cx, cy)
	ps.ClosePolygon(basics.PathFlagsNone)
	return ps
}

// buildStarPath creates a star polygon used as the mask shape.
func buildStarPath(cx, cy, r1, r2 float64, numPoints int) *path.PathStorageStl {
	ps := path.NewPathStorageStl()
	for i := 0; i <= numPoints*2; i++ {
		a := -math.Pi/2 + math.Pi*float64(i)/float64(numPoints)
		r := r2
		if i%2 == 0 {
			r = r1
		}
		x := cx + r*math.Cos(a)
		y := cy + r*math.Sin(a)
		if i == 0 {
			ps.MoveTo(x, y)
		} else {
			ps.LineTo(x, y)
		}
	}
	ps.ClosePolygon(basics.PathFlagsNone)
	return ps
}

func main() {
	ctx := agg.NewContext(width, height)
	img := ctx.GetImage()
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	ras := agg2d.GetInternalRasterizer()
	sl := scanline.NewScanlineP8()
	slAdapter := &slWrapP8v3{sl: sl}
	rasAdapter := &rasAdp3{ras: ras}

	cx, cy := float64(width)/2, float64(height)/2

	// Build mask from spiral + star.
	maskData := make([]uint8, width*height)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	// Spiral mask.
	spiral := buildSpiralPath(cx, cy, 20, 280, 500, 0)
	ras.Reset()
	spiral.Rewind(0)
	for {
		x, y, cmd := spiral.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		ras.AddVertex(x, y, uint32(cmd))
	}
	renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, maskRb, color.Gray8[color.Linear]{V: 255, A: 255})

	mask := pixfmt.NewAlphaMaskU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})

	// Render colored concentric rings through mask.
	mainBuf := buffer.NewRenderingBufferWithData[uint8](img.Data, width, height, width*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainBuf)
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(mainPixf, mask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	colors := []color.RGBA8[color.Linear]{
		{R: 255, G: 0, B: 0, A: 200},
		{R: 0, G: 180, B: 0, A: 200},
		{R: 0, G: 0, B: 255, A: 200},
		{R: 255, G: 180, B: 0, A: 200},
		{R: 180, G: 0, B: 255, A: 200},
		{R: 0, G: 200, B: 200, A: 200},
	}

	numRings := 6
	for i := numRings; i >= 1; i-- {
		r := float64(i) / float64(numRings) * 280.0
		// Build a filled circle via star (full circle = many-pointed star).
		ring := buildStarPath(cx, cy, r*0.6, r, 60)
		ras.Reset()
		ring.Rewind(0)
		for {
			x, y, cmd := ring.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}
		c := colors[(numRings-i)%len(colors)]
		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rbAMask, c)
	}

	const filename = "alpha_mask3.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
