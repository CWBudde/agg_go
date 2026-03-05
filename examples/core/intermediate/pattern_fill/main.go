// Port of AGG C++ pattern_fill.cpp – tiled pattern fill.
//
// Renders a large star polygon filled with a tiling pattern of small stars.
// Default: pattern size=30, pattern angle=0°, polygon angle=0°, scale=1.0.
package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	imageacc "agg_go/internal/image"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	"agg_go/internal/scanline"
	"agg_go/internal/span"
)

const (
	canvasW      = 800
	canvasH      = 600
	patternSize  = 30
	patternAngle = 0.0 // degrees
	polygonAngle = 0.0 // degrees
)

// patPixFmt is a simple RGBA pixel format for the pattern tile.
type patPixFmt struct {
	data         []basics.Int8u
	w, h, stride int
}

func (p patPixFmt) Width() int    { return p.w }
func (p patPixFmt) Height() int   { return p.h }
func (p patPixFmt) PixWidth() int { return 4 }
func (p patPixFmt) PixPtr(x, y int) []basics.Int8u {
	if y < 0 || y >= p.h || x < 0 || x >= p.w {
		return p.data[0:]
	}
	return p.data[y*p.stride+x*4:]
}

// patSource wraps an ImageAccessorWrap over patPixFmt.
type patSource struct {
	accessor *imageacc.ImageAccessorWrap[patPixFmt, *imageacc.WrapModeRepeatAutoPow2, *imageacc.WrapModeRepeatAutoPow2]
	pf       patPixFmt
}

func (s *patSource) Width() int                  { return s.pf.w }
func (s *patSource) Height() int                 { return s.pf.h }
func (s *patSource) ColorType() string           { return "RGBA8" }
func (s *patSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *patSource) Span(x, y, l int) []basics.Int8u {
	return s.accessor.Span(x, y, l)
}
func (s *patSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *patSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *patSource) RowPtr(y int) []basics.Int8u {
	return s.pf.PixPtr(0, y)
}

// rasScanlineAdapter adapts ScanlineU8 to rasterizer.ScanlineInterface.
type rasScanlineAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rasScanlineAdapter) ResetSpans()                { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, c uint32)    { a.sl.AddCell(x, uint(c)) }
func (a *rasScanlineAdapter) AddSpan(x, l int, c uint32) { a.sl.AddSpan(x, l, uint(c)) }
func (a *rasScanlineAdapter) Finalize(y int)             { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int              { return a.sl.NumSpans() }

// pathSourceAdapter bridges PathStorageStl to rasterizer VertexSource.
type pathSourceAdapter struct{ ps *path.PathStorageStl }

func (a *pathSourceAdapter) Rewind(id uint32) { a.ps.Rewind(uint(id)) }
func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

// generatePatternTile creates a small star-on-dark-background pattern image.
func generatePatternTile(size int, angleDeg float64) patPixFmt {
	pf := patPixFmt{w: size, h: size, stride: size * 4, data: make([]basics.Int8u, size*size*4)}

	// Fill background dark.
	for i := 0; i < size*size*4; i += 4 {
		pf.data[i] = 102
		pf.data[i+1] = 0
		pf.data[i+2] = 26
		pf.data[i+3] = 200
	}

	// Draw the pattern via a temporary agg.Image.
	tmpImg := agg.CreateImage(size, size)
	tc := agg.NewContextForImage(tmpImg)
	tc.Clear(agg.RGBA(0.4, 0.0, 0.1, 0.8))

	cx, cy := float64(size)/2.0, float64(size)/2.0
	r1 := float64(size)/2.0 - 1
	r2 := r1 / 2.5
	n := 6
	startRad := angleDeg * math.Pi / 180.0

	// Star fill.
	tc.SetColor(agg.RGBA(0.43, 0.51, 0.20, 1.0))
	first := true
	for i := 0; i < n*2; i++ {
		a := math.Pi*2.0*float64(i)/float64(n*2) - math.Pi/2.0 + startRad
		r := r2
		if i%2 == 0 {
			r = r1
		}
		x := cx + math.Cos(a)*r
		y := cy + math.Sin(a)*r
		if first {
			tc.MoveTo(x, y)
			first = false
		} else {
			tc.LineTo(x, y)
		}
	}
	tc.ClosePath()
	tc.Fill()

	copy(pf.data, tmpImg.Data)
	return pf
}

// buildLargeStarPath builds a large star polygon centred on (cx,cy).
func buildLargeStarPath(cx, cy float64, w, h int, angleDeg float64) *path.PathStorageStl {
	r := float64(w)
	if float64(h) < r {
		r = float64(h)
	}
	r1 := r/3 - 8
	r2 := r1 / 1.45
	nr := 14
	startRad := angleDeg * math.Pi / 180.0

	ps := path.NewPathStorageStl()
	for i := 0; i < nr; i++ {
		a := math.Pi*2.0*float64(i)/float64(nr) - math.Pi/2.0 + startRad
		dx := math.Cos(a)
		dy := math.Sin(a)
		if i&1 != 0 {
			ps.LineTo(cx+dx*r1, cy+dy*r1)
		} else {
			if i == 0 {
				ps.MoveTo(cx+dx*r2, cy+dy*r2)
			} else {
				ps.LineTo(cx+dx*r2, cy+dy*r2)
			}
		}
	}
	ps.ClosePolygon(basics.PathFlagsNone)
	return ps
}

func main() {
	ctx := agg.NewContext(canvasW, canvasH)
	ctx.Clear(agg.RGBA(0.95, 0.95, 0.9, 1.0))

	// Generate pattern tile.
	pf := generatePatternTile(patternSize, patternAngle)
	wrapX := imageacc.NewWrapModeRepeatAutoPow2(basics.Int32u(pf.w))
	wrapY := imageacc.NewWrapModeRepeatAutoPow2(basics.Int32u(pf.h))
	accessor := imageacc.NewImageAccessorWrap[patPixFmt, *imageacc.WrapModeRepeatAutoPow2, *imageacc.WrapModeRepeatAutoPow2](&pf, wrapX, wrapY)
	src := &patSource{accessor: accessor, pf: pf}
	sg := span.NewSpanPatternRGBAWithParams[*patSource](src, 0, 0)

	// Destination pipeline.
	dstImg := ctx.GetImage()
	dstRbuf := buffer.NewRenderingBufferWithData[uint8](dstImg.Data, dstImg.Width(), dstImg.Height(), dstImg.Width()*4)
	dstPixf := pixfmt.NewPixFmtRGBA32Pre[color.Linear](dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	cx, cy := float64(canvasW)/2, float64(canvasH)/2
	ps := buildLargeStarPath(cx, cy, canvasW, canvasH, polygonAngle)

	ras.Reset()
	ras.ClipBox(0, 0, float64(canvasW), float64(canvasH))
	ras.AddPath(&pathSourceAdapter{ps: ps}, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					colors := alloc.Allocate(int(spanData.Len))
					sg.Generate(colors, int(spanData.X), y, uint(spanData.Len))
					renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}

	// Draw the star outline.
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.NoFill()
	a.LineColor(agg.NewColor(0, 60, 80, 200))
	a.LineWidth(2.0)
	ps2 := buildLargeStarPath(cx, cy, canvasW, canvasH, polygonAngle)
	a.ResetPath()
	ps2.Rewind(0)
	for {
		x, y, cmd := ps2.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		if basics.IsMoveTo(basics.PathCommand(cmd)) {
			a.MoveTo(x, y)
		} else if basics.IsVertex(basics.PathCommand(cmd)) {
			a.LineTo(x, y)
		}
	}
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)

	const filename = "pattern_fill.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
