// Based on the original AGG examples: flash_rasterizer.cpp.
package main

import (
	"math"
	"math/rand"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	"agg_go/internal/scanline"
)

var (
	flashShapes []flashPath
	flashColors []color.RGBA8[color.Linear]
	flashScale  = 1.0
)

type flashPath struct {
	vertices  []flashVertex
	leftFill  int
	rightFill int
}

type flashVertex struct {
	x, y float64
	cmd  uint32
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
func (h *flashStyleHandler) GenerateSpan(colors []color.RGBA8[color.Linear], x, y, len, style int) {}

// flashScanlineAdapter adapts ScanlineU8 to CompoundScanlineInterface
type flashScanlineAdapter struct {
	sl *scanline.ScanlineU8
}

func (a *flashScanlineAdapter) ResetSpans() { a.sl.ResetSpans() }
func (a *flashScanlineAdapter) AddCell(x int, cover basics.Int8u) {
	a.sl.AddCell(x, uint(cover))
}

func (a *flashScanlineAdapter) AddSpan(x, len int, cover basics.Int8u) {
	a.sl.AddSpan(x, len, uint(cover))
}
func (a *flashScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *flashScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

type compoundNoClip struct {
	x1, y1 float64
}

func (c *compoundNoClip) ResetClipping()                 {}
func (c *compoundNoClip) ClipBox(x1, y1, x2, y2 float64) {}
func (c *compoundNoClip) MoveTo(x, y float64) {
	c.x1, c.y1 = x, y
}

func (c *compoundNoClip) LineTo(outline *rasterizer.RasterizerCellsAAStyled, x, y float64) {
	outline.Line(basics.IRound(c.x1*basics.PolySubpixelScale), basics.IRound(c.y1*basics.PolySubpixelScale),
		basics.IRound(x*basics.PolySubpixelScale), basics.IRound(y*basics.PolySubpixelScale))
	c.x1, c.y1 = x, y
}

func initFlashDemo() {
	if flashShapes != nil {
		return
	}

	rng := rand.New(rand.NewSource(1234))
	flashColors = make([]color.RGBA8[color.Linear], 20)
	for i := range flashColors {
		flashColors[i] = color.RGBA8[color.Linear]{
			R: uint8(rng.Intn(256)),
			G: uint8(rng.Intn(256)),
			B: uint8(rng.Intn(256)),
			A: 200,
		}
	}

	// Create some overlapping shapes
	for i := 0; i < 15; i++ {
		cx := rng.Float64() * float64(width)
		cy := rng.Float64() * float64(height)
		r := 20.0 + rng.Float64()*80.0

		path := flashPath{
			leftFill:  rng.Intn(len(flashColors)),
			rightFill: -1,
		}

		// Create a star-like polygon
		numPoints := 5 + rng.Intn(5)
		for j := 0; j < numPoints*2; j++ {
			angle := 2.0 * math.Pi * float64(j) / float64(numPoints*2)
			dist := r
			if j%2 == 1 {
				dist /= 2.0
			}
			x := cx + dist*math.Cos(angle)
			y := cy + dist*math.Sin(angle)
			cmd := uint32(basics.PathCmdLineTo)
			if j == 0 {
				cmd = uint32(basics.PathCmdMoveTo)
			}
			path.vertices = append(path.vertices, flashVertex{x: x, y: y, cmd: cmd})
		}
		path.vertices = append(path.vertices, flashVertex{cmd: uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)})
		flashShapes = append(flashShapes, path)
	}
}

func drawFlashRasterizerDemo() {
	initFlashDemo()

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	renBase.ClipBox(0, 0, img.Width(), img.Height())
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255}) // rgba(1.0, 1.0, 0.95)

	// Create compound rasterizer
	clipper := &compoundNoClip{}
	rasc := rasterizer.NewRasterizerCompoundAA(clipper)

	for _, p := range flashShapes {
		rasc.Styles(p.leftFill, p.rightFill)
		for _, v := range p.vertices {
			rasc.AddVertex(v.x, v.y, v.cmd)
		}
	}

	// We'll use our own RenderScanlinesCompound-like loop
	rasc.Sort()
	if !rasc.RewindScanlines() {
		return
	}

	slAA := scanline.NewScanlineU8()
	slBin := scanline.NewScanlineU8()
	adapterAA := &flashScanlineAdapter{sl: slAA}
	adapterBin := &flashScanlineAdapter{sl: slBin}

	styleHandler := &flashStyleHandler{colors: flashColors}

	minX := rasc.MinX()
	maxX := rasc.MaxX()
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
			if rasc.SweepScanline(adapterAA, 0) {
				style := int(rasc.Style(0))
				c := styleHandler.Color(style)

				y := slAA.Y()
				for _, spanData := range slAA.Spans() {
					if spanData.Len > 0 {
						renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), c, spanData.Covers)
					}
				}
			}
		} else {
			if rasc.SweepScanline(adapterBin, -1) {
				y := slBin.Y()

				for _, spanData := range slBin.Spans() {
					for j := 0; j < int(spanData.Len); j++ {
						mixBuffer[int(spanData.X)-minX+j] = color.RGBA8[color.Linear]{}
					}
				}

				for i := uint32(0); i < numStyles; i++ {
					style := int(rasc.Style(i))
					if rasc.SweepScanline(adapterAA, int(i)) {
						c := styleHandler.Color(style)
						for _, spanData := range slAA.Spans() {
							for j := 0; j < int(spanData.Len); j++ {
								ptr := &mixBuffer[int(spanData.X)-minX+j]
								cover := spanData.Covers[j]
								ptr.AddWithCover(c, cover)
							}
						}
					}
				}

				for _, spanData := range slBin.Spans() {
					renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), mixBuffer[int(spanData.X)-minX:], nil, basics.CoverFull)
				}
			}
		}
	}
}
