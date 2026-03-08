// Port of AGG C++ flash_rasterizer.cpp – compound rasterizer with styled fills.
//
// Renders overlapping star polygons using the compound AA rasterizer, which
// correctly handles left/right fill styles for non-zero winding rules.
// Default: 15 random star shapes with 20 random fill colours.
package main

import (
	"math"
	"math/rand"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	"agg_go/internal/scanline"
)

const (
	width  = 800
	height = 600
)

type flashVertex struct {
	x, y float64
	cmd  uint32
}

type flashPath struct {
	vertices  []flashVertex
	leftFill  int
	rightFill int
}

type flashStyleHandler struct {
	colors []color.RGBA8[color.Linear]
}

func (h *flashStyleHandler) IsSolid(style int) bool { return true }
func (h *flashStyleHandler) Color(style int) color.RGBA8[color.Linear] {
	if style < 0 || style >= len(h.colors) {
		return color.RGBA8[color.Linear]{}
	}
	return h.colors[style]
}

func (h *flashStyleHandler) GenerateSpan(colors []color.RGBA8[color.Linear], x, y, l, style int) {
}

type flashSLAdapter struct{ sl *scanline.ScanlineU8 }

func (a *flashSLAdapter) ResetSpans()                      { a.sl.ResetSpans() }
func (a *flashSLAdapter) AddCell(x int, c basics.Int8u)    { a.sl.AddCell(x, uint(c)) }
func (a *flashSLAdapter) AddSpan(x, l int, c basics.Int8u) { a.sl.AddSpan(x, l, uint(c)) }
func (a *flashSLAdapter) Finalize(y int)                   { a.sl.Finalize(y) }
func (a *flashSLAdapter) NumSpans() int                    { return a.sl.NumSpans() }

type compoundNoClip struct{ x1, y1 float64 }

func (c *compoundNoClip) ResetClipping()                 {}
func (c *compoundNoClip) ClipBox(x1, y1, x2, y2 float64) {}
func (c *compoundNoClip) MoveTo(x, y float64)            { c.x1, c.y1 = x, y }
func (c *compoundNoClip) LineTo(outline *rasterizer.RasterizerCellsAAStyled, x, y float64) {
	outline.Line(
		basics.IRound(c.x1*basics.PolySubpixelScale), basics.IRound(c.y1*basics.PolySubpixelScale),
		basics.IRound(x*basics.PolySubpixelScale), basics.IRound(y*basics.PolySubpixelScale),
	)
	c.x1, c.y1 = x, y
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	rng := rand.New(rand.NewSource(1234))

	// Random fill colours.
	colors := make([]color.RGBA8[color.Linear], 20)
	for i := range colors {
		colors[i] = color.RGBA8[color.Linear]{
			R: uint8(rng.Intn(256)),
			G: uint8(rng.Intn(256)),
			B: uint8(rng.Intn(256)),
			A: 200,
		}
	}

	// Build star shapes.
	var shapes []flashPath
	for i := 0; i < 15; i++ {
		cx := rng.Float64() * float64(width)
		cy := rng.Float64() * float64(height)
		r := 20.0 + rng.Float64()*80.0
		numPts := 5 + rng.Intn(5)

		fp := flashPath{
			leftFill:  rng.Intn(len(colors)),
			rightFill: -1,
		}
		for j := 0; j < numPts*2; j++ {
			a := 2.0 * math.Pi * float64(j) / float64(numPts*2)
			dist := r
			if j%2 == 1 {
				dist /= 2.0
			}
			x := cx + dist*math.Cos(a)
			y := cy + dist*math.Sin(a)
			cmd := uint32(basics.PathCmdLineTo)
			if j == 0 {
				cmd = uint32(basics.PathCmdMoveTo)
			}
			fp.vertices = append(fp.vertices, flashVertex{x: x, y: y, cmd: cmd})
		}
		fp.vertices = append(fp.vertices, flashVertex{
			cmd: uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose),
		})
		shapes = append(shapes, fp)
	}

	// Setup rendering pipeline.
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](pixFmt)
	renBase.ClipBox(0, 0, width, height)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	clipper := &compoundNoClip{}
	rasc := rasterizer.NewRasterizerCompoundAA(clipper)

	for _, p := range shapes {
		rasc.Styles(p.leftFill, p.rightFill)
		for _, v := range p.vertices {
			rasc.AddVertex(v.x, v.y, v.cmd)
		}
	}

	rasc.Sort()
	if !rasc.RewindScanlines() {
		return
	}

	minX := rasc.MinX()
	maxX := rasc.MaxX()

	slAA := scanline.NewScanlineU8()
	slBin := scanline.NewScanlineU8()
	slAA.Reset(minX, maxX)
	slBin.Reset(minX, maxX)
	adAA := &flashSLAdapter{sl: slAA}
	adBin := &flashSLAdapter{sl: slBin}

	styleHandler := &flashStyleHandler{colors: colors}
	length := maxX - minX + 2
	if length < 0 {
		length = 0
	}
	colorSpan := make([]color.RGBA8[color.Linear], length*2)
	mixBuffer := colorSpan[length:]

	for {
		numStyles := rasc.SweepStyles()
		if numStyles == 0 {
			break
		}
		if numStyles == 1 {
			if rasc.SweepScanline(adAA, 0) {
				style := int(rasc.Style(0))
				c := styleHandler.Color(style)
				y := slAA.Y()
				for _, sp := range slAA.Spans() {
					if sp.Len > 0 {
						renBase.BlendSolidHspan(int(sp.X), y, int(sp.Len), c, sp.Covers)
					}
				}
			}
		} else {
			if rasc.SweepScanline(adBin, -1) {
				y := slBin.Y()
				for _, sp := range slBin.Spans() {
					for j := 0; j < int(sp.Len); j++ {
						mixBuffer[int(sp.X)-minX+j] = color.RGBA8[color.Linear]{}
					}
				}
				for i := uint32(0); i < numStyles; i++ {
					style := int(rasc.Style(i))
					if rasc.SweepScanline(adAA, int(i)) {
						for _, sp := range slAA.Spans() {
							c := styleHandler.Color(style)
							for j := 0; j < int(sp.Len); j++ {
								ptr := &mixBuffer[int(sp.X)-minX+j]
								cover := sp.Covers[j]
								ptr.AddWithCover(c, cover)
							}
						}
					}
				}
				for _, sp := range slBin.Spans() {
					renBase.BlendColorHspan(int(sp.X), y, int(sp.Len), mixBuffer[int(sp.X)-minX:], nil, basics.CoverFull)
				}
			}
		}
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Flash Rasterizer",
		Width:  width,
		Height: height,
	}, &demo{})
}
