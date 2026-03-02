// Based on the original AGG examples: flash_rasterizer.cpp.
package main

import (
	"math"
	"math/rand"

	agg "agg_go"
	"agg_go/internal/array"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
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

func (p *flashPath) Rewind(pathID uint32) {
	// Not used for our simple manual adding
}

func (p *flashPath) Vertex(x, y *float64) uint32 {
	// Not used
	return uint32(basics.PathCmdStop)
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

func initFlashDemo() {
	if flashShapes != nil {
		return
	}

	rand.Seed(1234)
	flashColors = make([]color.RGBA8[color.Linear], 20)
	for i := range flashColors {
		flashColors[i] = color.RGBA8[color.Linear]{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 200,
		}
	}

	// Create some overlapping shapes
	for i := 0; i < 15; i++ {
		cx := rand.Float64()*float64(width)
		cy := rand.Float64()*float64(height)
		r := 20.0 + rand.Float64()*80.0
		
		path := flashPath{
			leftFill:  rand.Intn(len(flashColors)),
			rightFill: -1,
		}
		
		// Create a star-like polygon
		numPoints := 5 + rand.Intn(5)
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

	// Use internal components for compound rendering
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	// Create pixel format and renderer base
	// Using RGBA32Pre to match Agg2D's internal rendering
	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	renBase.ClipBox(0, 0, img.Width(), img.Height())

	// Create compound rasterizer
	clipper := rasterizer.NewRasterizerSlClip[float64, rasterizer.DblConv](rasterizer.DblConv{})
	rasc := rasterizer.NewRasterizerCompoundAA(clipper)
	rasc.ClipBox(0, 0, float64(img.Width()), float64(img.Height()))
	
	for _, p := range flashShapes {
		rasc.Styles(p.leftFill, p.rightFill)
		for _, v := range p.vertices {
			rasc.AddVertex(v.x, v.y, v.cmd)
		}
	}

	// We'll use our own RenderScanlinesCompound-like loop because of interface mismatches
	rasc.Sort()
	if !rasc.RewindScanlines() {
		return
	}

	slAA := scanline.NewScanlineU8()
	slBin := scanline.NewScanlineU8()
	adapterAA := &flashScanlineAdapter{sl: slAA}
	adapterBin := &flashScanlineAdapter{sl: slBin}
	
	styleHandler := &flashStyleHandler{colors: flashColors}
	alloc := renscan.NewSpanAllocator[color.RGBA8[color.Linear]]()

	minX := rasc.MinX()
	maxX := rasc.MaxX()
	length := maxX - minX + 2
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
				
				// Replicate RenderScanlineAASolid
				iter := slAA.Spans()
				y := slAA.Y()
				for _, span := range iter {
					if span.Len > 0 {
						renBase.BlendSolidHspan(int(span.X), y, int(span.Len), c, span.Covers)
					}
				}
			}
		} else {
			if rasc.SweepScanline(adapterBin, -1) {
				// Replicate multiple styles rendering
				y := slBin.Y()
				
				// Clear mix buffer for spans in slBin
				for _, span := range slBin.Spans() {
					for j := 0; j < int(span.Len); j++ {
						mixBuffer[int(span.X)-minX+j] = color.RGBA8[color.Linear]{}
					}
				}

				for i := uint32(0); i < numStyles; i++ {
					style := int(rasc.Style(i))
					if rasc.SweepScanline(adapterAA, int(i)) {
						c := styleHandler.Color(style)
						for _, span := range slAA.Spans() {
							for j := 0; j < int(span.Len); j++ {
								ptr := &mixBuffer[int(span.X)-minX+j]
								cover := span.Covers[j]
								// AddWithCover logic from internal/color
								ptr.AddWithCover(c, cover)
							}
						}
					}
				}

				// Finally blend the mixBuffer to renBase
				for _, span := range slBin.Spans() {
					renBase.BlendColorHspan(int(span.X), y, int(span.Len), mixBuffer[int(span.X)-minX:], nil, basics.CoverFull)
				}
			}
		}
	}
}
