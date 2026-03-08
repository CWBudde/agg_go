// Based on the original AGG examples: gouraud_mesh.cpp.
package main

import (
	"math/rand"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	"agg_go/internal/scanline"
	"agg_go/internal/span"
)

type meshPoint struct {
	x, y   float64
	dx, dy float64
	color  color.RGBA8[color.Linear]
	dc     [3]int // direction of color change
}

type meshTriangle struct {
	p1, p2, p3 uint32
}

type meshEdge struct {
	p1, p2 uint32
	tl, tr int // styles left and right
}

var (
	meshVertices  []meshPoint
	meshTriangles []meshTriangle
	meshEdges     []meshEdge
	meshCols      = 10
	meshRows      = 10
	meshInited    = false
)

func setMeshSize(cols, rows int) {
	if cols < 2 {
		cols = 2
	}
	if rows < 2 {
		rows = 2
	}
	if cols == meshCols && rows == meshRows {
		return
	}
	meshCols = cols
	meshRows = rows
	meshInited = false
}

func initMesh() {
	if meshInited {
		return
	}
	rng := rand.New(rand.NewSource(1234))

	cellW := float64(width-80) / float64(meshCols-1)
	cellH := float64(height-80) / float64(meshRows-1)
	startX, startY := 40.0, 40.0

	meshVertices = nil
	for i := range meshRows {
		y := startY + float64(i)*cellH
		for j := 0; j < meshCols; j++ {
			x := startX + float64(j)*cellW
			meshVertices = append(meshVertices, meshPoint{
				x: x, y: y,
				dx: (rng.Float64() - 0.5) * 2.0,
				dy: (rng.Float64() - 0.5) * 2.0,
				color: color.RGBA8[color.Linear]{
					R: uint8(rng.Intn(256)),
					G: uint8(rng.Intn(256)),
					B: uint8(rng.Intn(256)),
					A: 255,
				},
				dc: [3]int{rng.Intn(2), rng.Intn(2), rng.Intn(2)},
			})
		}
	}

	meshTriangles = nil
	meshEdges = nil
	for i := 0; i < meshRows-1; i++ {
		for j := 0; j < meshCols-1; j++ {
			p1 := uint32(i*meshCols + j)
			p2 := p1 + 1
			p3 := p2 + uint32(meshCols)
			p4 := p1 + uint32(meshCols)

			meshTriangles = append(meshTriangles,
				meshTriangle{p1, p2, p3},
				meshTriangle{p3, p4, p1},
			)

			currCell := i*(meshCols-1) + j
			bottCell := -1
			if i > 0 {
				bottCell = currCell - (meshCols - 1)
			}
			leftCell := -1
			if j > 0 {
				leftCell = currCell - 1
			}

			currT1 := currCell * 2
			currT2 := currT1 + 1

			leftT1 := -1
			if leftCell >= 0 {
				leftT1 = leftCell * 2
			}

			bottT2 := -1
			if bottCell >= 0 {
				bottT2 = (bottCell * 2) + 1
			}

			meshEdges = append(meshEdges,
				meshEdge{p1, p2, currT1, bottT2},
				meshEdge{p1, p3, currT2, currT1},
				meshEdge{p1, p4, leftT1, currT2},
			)

			if j == meshCols-2 {
				meshEdges = append(meshEdges, meshEdge{p2, p3, currT1, -1})
			}
			if i == meshRows-2 {
				meshEdges = append(meshEdges, meshEdge{p3, p4, currT2, -1})
			}
		}
	}
	meshInited = true
}

type meshStyleHandler struct {
	triangles []*span.SpanGouraudRGBA
}

func (h *meshStyleHandler) IsSolid(style int) bool { return false }
func (h *meshStyleHandler) Color(style int) color.RGBA8[color.Linear] {
	return color.RGBA8[color.Linear]{}
}

func (h *meshStyleHandler) GenerateSpan(colors []color.RGBA8[color.Linear], x, y, length, style int) {
	if style >= 0 && style < len(h.triangles) {
		temp := make([]span.RGBAColor, length)
		h.triangles[style].Generate(temp, x, y, uint(length))
		for i := 0; i < length; i++ {
			colors[i] = color.RGBA8[color.Linear]{
				R: uint8(temp[i].R),
				G: uint8(temp[i].G),
				B: uint8(temp[i].B),
				A: uint8(temp[i].A),
			}
		}
	}
}

func drawGouraudMeshDemo() {
	initMesh()

	// Update mesh
	for i := range meshVertices {
		p := &meshVertices[i]
		p.x += p.dx
		p.y += p.dy

		if p.x < 0 || p.x > float64(width) {
			p.dx = -p.dx
		}
		if p.y < 0 || p.y > float64(height) {
			p.dy = -p.dy
		}

		c := &p.color
		updateChan := func(val *basics.Int8u, dir *int) {
			v := int(*val)
			if *dir != 0 {
				v += 2
			} else {
				v -= 2
			}
			if v < 0 {
				v = 0
				*dir = 1
			}
			if v > 255 {
				v = 255
				*dir = 0
			}
			*val = basics.Int8u(v)
		}
		updateChan(&c.R, &p.dc[0])
		updateChan(&c.G, &p.dc[1])
		updateChan(&c.B, &p.dc[2])
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	renBase.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}) // rgba(0, 0, 0)

	styles := &meshStyleHandler{}
	for _, t := range meshTriangles {
		p1 := meshVertices[t.p1]
		p2 := meshVertices[t.p2]
		p3 := meshVertices[t.p3]

		c1 := span.RGBAColor{R: int(p1.color.R), G: int(p1.color.G), B: int(p1.color.B), A: int(p1.color.A)}
		c2 := span.RGBAColor{R: int(p2.color.R), G: int(p2.color.G), B: int(p2.color.B), A: int(p2.color.A)}
		c3 := span.RGBAColor{R: int(p3.color.R), G: int(p3.color.G), B: int(p3.color.B), A: int(p3.color.A)}

		g := span.NewSpanGouraudRGBAWithTriangle(
			c1, c2, c3,
			p1.x, p1.y, p2.x, p2.y, p3.x, p3.y,
			0,
		)
		g.Prepare()
		styles.triangles = append(styles.triangles, g)
	}

	clipper := &compoundNoClip{}
	rasc := rasterizer.NewRasterizerCompoundAA(clipper)

	for _, e := range meshEdges {
		p1 := meshVertices[e.p1]
		p2 := meshVertices[e.p2]
		rasc.Styles(e.tl, e.tr)
		rasc.MoveToD(p1.x, p1.y)
		rasc.LineToD(p2.x, p2.y)
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
	adapterAA := &flashScanlineAdapter{sl: slAA}
	adapterBin := &flashScanlineAdapter{sl: slBin}

	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

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
				y := slAA.Y()
				for _, spanData := range slAA.Spans() {
					if spanData.Len > 0 {
						colors := alloc.Allocate(int(spanData.Len))
						styles.GenerateSpan(colors, int(spanData.X), y, int(spanData.Len), style)
						renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
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
						for _, spanData := range slAA.Spans() {
							colors := alloc.Allocate(int(spanData.Len))
							styles.GenerateSpan(colors, int(spanData.X), y, int(spanData.Len), style)
							for j := 0; j < int(spanData.Len); j++ {
								ptr := &mixBuffer[int(spanData.X)-minX+j]
								cover := spanData.Covers[j]
								ptr.AddWithCover(colors[j], cover)
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
