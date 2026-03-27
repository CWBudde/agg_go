// Package main ports AGG's gouraud_mesh.cpp demo as closely as the standalone
// runner allows.
package main

import (
	"fmt"
	"time"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/span"
)

const (
	frameWidth  = 400
	frameHeight = 400
	meshCols    = 20
	meshRows    = 20
	cellSize    = 17.0
	meshStartX  = 40.0
	meshStartY  = 40.0
)

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRandSeed1() *clibcRand {
	return &clibcRand{
		state: [31]int32{
			-1726662223, 379960547, 1735697613, 1040273694, 1313901226,
			1627687941, -179304937, -2073333483, 1780058412, -1989503057,
			-615974602, 344556628, 939512070, -1249116260, 1507946756,
			-812545463, 154635395, 1388815473, -1926676823, 525320961,
			-1009028674, 968117788, -123449607, 1284210865, 435012392,
			-2017506339, -911064859, -370259173, 1132637927, 1398500161, -205601318,
		},
		fptr: 3,
		rptr: 0,
	}
}

func (r *clibcRand) next() int32 {
	r.state[r.fptr] += r.state[r.rptr]
	result := int32(uint32(r.state[r.fptr]) >> 1)
	r.fptr++
	if r.fptr >= len(r.state) {
		r.fptr = 0
	}
	r.rptr++
	if r.rptr >= len(r.state) {
		r.rptr = 0
	}
	return result
}

func (r *clibcRand) randN(n int) int      { return int(r.next()) % n }
func (r *clibcRand) randAnd(mask int) int { return int(r.next()) & mask }

func cxxRandom(rng *clibcRand, v1, v2 float64) float64 {
	return (v2-v1)*float64(rng.randN(1000))/999.0 + v1
}

func srgbaRandRTL(rng *clibcRand) icol.RGBA8[icol.SRGB] {
	return icol.RGBA8[icol.SRGB]{
		A: 255,
		B: uint8(rng.randAnd(0xFF)),
		G: uint8(rng.randAnd(0xFF)),
		R: uint8(rng.randAnd(0xFF)),
	}
}

func srgbaDirRTL(rng *clibcRand) icol.RGBA8[icol.SRGB] {
	return icol.RGBA8[icol.SRGB]{
		B: uint8(rng.randAnd(1)),
		G: uint8(rng.randAnd(1)),
		R: uint8(rng.randAnd(1)),
	}
}

type meshPoint struct {
	x, y   float64
	dx, dy float64
	color  icol.RGBA8[icol.SRGB]
	dc     icol.RGBA8[icol.SRGB]
}

type meshTriangle struct {
	p1, p2, p3 int
}

type meshEdge struct {
	p1, p2 int
	tl, tr int
}

type meshCtrl struct {
	cols, rows int
	dragIdx    int
	dragDX     float64
	dragDY     float64
	cellW      float64
	cellH      float64
	startX     float64
	startY     float64
	vertices   []meshPoint
	triangles  []meshTriangle
	edges      []meshEdge
}

func newMeshCtrl() *meshCtrl {
	return &meshCtrl{dragIdx: -1}
}

func (m *meshCtrl) generate(cols, rows int, cellW, cellH, startX, startY float64, rng *clibcRand) {
	m.cols = cols
	m.rows = rows
	m.cellW = cellW
	m.cellH = cellH
	m.startX = startX
	m.startY = startY
	m.dragIdx = -1

	m.vertices = m.vertices[:0]
	y := startY
	for i := 0; i < m.rows; i++ {
		x := startX
		for j := 0; j < m.cols; j++ {
			m.vertices = append(m.vertices, meshPoint{
				x:     x,
				y:     y,
				dx:    cxxRandom(rng, -0.5, 0.5),
				dy:    cxxRandom(rng, -0.5, 0.5),
				color: srgbaRandRTL(rng),
				dc:    srgbaDirRTL(rng),
			})
			x += cellW
		}
		y += cellH
	}

	m.triangles = m.triangles[:0]
	m.edges = m.edges[:0]
	for i := 0; i < m.rows-1; i++ {
		for j := 0; j < m.cols-1; j++ {
			p1 := i*m.cols + j
			p2 := p1 + 1
			p3 := p2 + m.cols
			p4 := p1 + m.cols

			m.triangles = append(m.triangles,
				meshTriangle{p1: p1, p2: p2, p3: p3},
				meshTriangle{p1: p3, p2: p4, p3: p1},
			)

			currCell := i*(m.cols-1) + j
			leftCell := -1
			if j != 0 {
				leftCell = currCell - 1
			}
			bottomCell := -1
			if i != 0 {
				bottomCell = currCell - (m.cols - 1)
			}

			currT1 := currCell * 2
			currT2 := currT1 + 1
			leftT1 := -1
			if leftCell >= 0 {
				leftT1 = leftCell * 2
			}
			bottomT2 := -1
			if bottomCell >= 0 {
				bottomT2 = bottomCell*2 + 1
			}

			m.edges = append(m.edges,
				meshEdge{p1: p1, p2: p2, tl: currT1, tr: bottomT2},
				meshEdge{p1: p1, p2: p3, tl: currT2, tr: currT1},
				meshEdge{p1: p1, p2: p4, tl: leftT1, tr: currT2},
			)

			if j == m.cols-2 {
				m.edges = append(m.edges, meshEdge{p1: p2, p2: p3, tl: currT1, tr: -1})
			}
			if i == m.rows-2 {
				m.edges = append(m.edges, meshEdge{p1: p3, p2: p4, tl: currT2, tr: -1})
			}
		}
	}
}

func (m *meshCtrl) vertex(i int) *meshPoint {
	return &m.vertices[i]
}

func (m *meshCtrl) vertexAt(x, y int) *meshPoint {
	return &m.vertices[y*m.cols+x]
}

func (m *meshCtrl) randomizePoints() {
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			xc := float64(j)*m.cellW + m.startX
			yc := float64(i)*m.cellH + m.startY
			x1 := xc - m.cellW/4
			y1 := yc - m.cellH/4
			x2 := xc + m.cellW/4
			y2 := yc + m.cellH/4

			p := m.vertexAt(j, i)
			p.x += p.dx
			p.y += p.dy

			if p.x < x1 {
				p.x = x1
				p.dx = -p.dx
			}
			if p.y < y1 {
				p.y = y1
				p.dy = -p.dy
			}
			if p.x > x2 {
				p.x = x2
				p.dx = -p.dx
			}
			if p.y > y2 {
				p.y = y2
				p.dy = -p.dy
			}
		}
	}
}

func (m *meshCtrl) rotateColors() {
	for i := 1; i < len(m.vertices); i++ {
		p := &m.vertices[i]
		rotateColorChannel(&p.color.R, &p.dc.R)
		rotateColorChannel(&p.color.G, &p.dc.G)
		rotateColorChannel(&p.color.B, &p.dc.B)
	}
}

func rotateColorChannel(value, dir *uint8) {
	v := int(*value)
	if *dir != 0 {
		v += 5
	} else {
		v -= 5
	}
	if v < 0 {
		v = 0
		*dir ^= 1
	}
	if v > 255 {
		v = 255
		*dir ^= 1
	}
	*value = uint8(v)
}

func sqr(v float64) float64 { return v * v }

func (m *meshCtrl) onMouseDown(x, y float64, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	for i := range m.vertices {
		p := &m.vertices[i]
		if sqr(x-p.x)+sqr(y-p.y) < 25 {
			m.dragIdx = i
			m.dragDX = x - p.x
			m.dragDY = y - p.y
			return true
		}
	}
	return false
}

func (m *meshCtrl) onMouseMove(x, y float64, btn lowlevelrunner.Buttons) bool {
	if btn.Left {
		if m.dragIdx >= 0 {
			p := m.vertex(m.dragIdx)
			p.x = x - m.dragDX
			p.y = y - m.dragDY
			return true
		}
		return false
	}
	return m.onMouseUp(x, y, btn)
}

func (m *meshCtrl) onMouseUp(_, _ float64, _ lowlevelrunner.Buttons) bool {
	wasDragging := m.dragIdx >= 0
	m.dragIdx = -1
	return wasDragging
}

type meshStyleHandler struct {
	triangles []*span.SpanGouraudRGBA
}

func newMeshStyleHandler(mesh *meshCtrl) *meshStyleHandler {
	h := &meshStyleHandler{triangles: make([]*span.SpanGouraudRGBA, 0, len(mesh.triangles))}
	for _, t := range mesh.triangles {
		p1 := mesh.vertex(t.p1)
		p2 := mesh.vertex(t.p2)
		p3 := mesh.vertex(t.p3)

		g := span.NewSpanGouraudRGBAWithTriangle(
			span.RGBAColor{R: int(p1.color.R), G: int(p1.color.G), B: int(p1.color.B), A: int(p1.color.A)},
			span.RGBAColor{R: int(p2.color.R), G: int(p2.color.G), B: int(p2.color.B), A: int(p2.color.A)},
			span.RGBAColor{R: int(p3.color.R), G: int(p3.color.G), B: int(p3.color.B), A: int(p3.color.A)},
			p1.x, p1.y,
			p2.x, p2.y,
			p3.x, p3.y,
			0,
		)
		g.Prepare()
		h.triangles = append(h.triangles, g)
	}
	return h
}

func (h *meshStyleHandler) IsSolid(int) bool { return false }

func (h *meshStyleHandler) Color(int) icol.RGBA8[icol.SRGB] {
	return icol.RGBA8[icol.SRGB]{}
}

func (h *meshStyleHandler) GenerateSpan(colors []icol.RGBA8[icol.SRGB], x, y, length, style int) {
	if style < 0 || style >= len(h.triangles) {
		return
	}
	tmp := make([]span.RGBAColor, length)
	h.triangles[style].Generate(tmp, x, y, uint(length))
	for i := 0; i < length; i++ {
		colors[i] = icol.RGBA8[icol.SRGB]{
			R: uint8(tmp[i].R),
			G: uint8(tmp[i].G),
			B: uint8(tmp[i].B),
			A: uint8(tmp[i].A),
		}
	}
}

type compoundNoClip struct{ x1, y1 float64 }

func (c *compoundNoClip) ResetClipping()                             {}
func (c *compoundNoClip) ClipBox(float64, float64, float64, float64) {}
func (c *compoundNoClip) MoveTo(x, y float64)                        { c.x1, c.y1 = x, y }
func (c *compoundNoClip) LineTo(outline *rasterizer.RasterizerCellsAAStyled, x, y float64) {
	outline.Line(
		basics.IRound(c.x1*basics.PolySubpixelScale),
		basics.IRound(c.y1*basics.PolySubpixelScale),
		basics.IRound(x*basics.PolySubpixelScale),
		basics.IRound(y*basics.PolySubpixelScale),
	)
	c.x1, c.y1 = x, y
}

type convVertexSourceRasVS struct{ src conv.VertexSource }

func (a *convVertexSourceRasVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convVertexSourceRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type compoundSLAdapter struct{ sl *scanline.ScanlineU8 }

func (a *compoundSLAdapter) ResetSpans()                      { a.sl.ResetSpans() }
func (a *compoundSLAdapter) AddCell(x int, c basics.Int8u)    { a.sl.AddCell(x, uint(c)) }
func (a *compoundSLAdapter) AddSpan(x, l int, c basics.Int8u) { a.sl.AddSpan(x, l, uint(c)) }
func (a *compoundSLAdapter) Finalize(y int)                   { a.sl.Finalize(y) }
func (a *compoundSLAdapter) NumSpans() int                    { return a.sl.NumSpans() }

type demo struct {
	mesh        *meshCtrl
	initialized bool
}

func newDemo() *demo {
	return &demo{mesh: newMeshCtrl()}
}

func (d *demo) OnInit() {
	rng := newClibcRandSeed1()
	d.mesh.generate(meshCols, meshRows, cellSize, cellSize, meshStartX, meshStartY, rng)
	d.initialized = true
}

func (d *demo) IsAnimated() bool { return true }

func (d *demo) OnIdle() {
	if !d.initialized {
		return
	}
	d.mesh.randomizePoints()
	d.mesh.rotateColors()
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	return d.mesh.onMouseDown(float64(x), float64(y), btn)
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	return d.mesh.onMouseMove(float64(x), float64(y), btn)
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	return d.mesh.onMouseUp(float64(x), float64(y), btn)
}

func (d *demo) Render(img *agg.Image) {
	if !d.initialized {
		d.OnInit()
	}

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Pre[icol.SRGB](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[icol.SRGB], icol.RGBA8[icol.SRGB]](pf)
	renBase.Clear(icol.RGBA8[icol.SRGB]{A: 255})

	styles := newMeshStyleHandler(d.mesh)
	rasc := rasterizer.NewRasterizerCompoundAA(&compoundNoClip{})
	slAA := scanline.NewScanlineU8()
	slBin := scanline.NewScanlineU8()
	adAA := &compoundSLAdapter{sl: slAA}
	adBin := &compoundSLAdapter{sl: slBin}
	alloc := span.NewSpanAllocator[icol.RGBA8[icol.SRGB]]()

	start := time.Now()
	rasc.Reset()
	for _, e := range d.mesh.edges {
		p1 := d.mesh.vertex(e.p1)
		p2 := d.mesh.vertex(e.p2)
		rasc.Styles(e.tl, e.tr)
		rasc.MoveToD(p1.x, p1.y)
		rasc.LineToD(p2.x, p2.y)
	}
	if rasc.RewindScanlines() {
		minX := rasc.MinX()
		maxX := rasc.MaxX()
		slAA.Reset(minX, maxX)
		slBin.Reset(minX, maxX)

		length := maxX - minX + 2
		if length < 0 {
			length = 0
		}
		colorSpan := make([]icol.RGBA8[icol.SRGB], length*2)
		mixBuffer := colorSpan[length:]

		for {
			numStyles := rasc.SweepStyles()
			if numStyles == 0 {
				break
			}

			if numStyles == 1 {
				if rasc.SweepScanline(adAA, 0) {
					style := int(rasc.Style(0))
					y := slAA.Y()
					for _, sp := range slAA.Spans() {
						colors := alloc.Allocate(int(sp.Len))
						styles.GenerateSpan(colors, int(sp.X), y, int(sp.Len), style)
						renBase.BlendColorHspan(int(sp.X), y, int(sp.Len), colors, sp.Covers, basics.CoverFull)
					}
				}
				continue
			}

			if !rasc.SweepScanline(adBin, -1) {
				continue
			}

			y := slBin.Y()
			for _, sp := range slBin.Spans() {
				for j := 0; j < int(sp.Len); j++ {
					mixBuffer[int(sp.X)-minX+j] = icol.RGBA8[icol.SRGB]{}
				}
			}

			for i := uint32(0); i < numStyles; i++ {
				style := int(rasc.Style(i))
				if !rasc.SweepScanline(adAA, int(i)) {
					continue
				}
				for _, sp := range slAA.Spans() {
					colors := alloc.Allocate(int(sp.Len))
					styles.GenerateSpan(colors, int(sp.X), y, int(sp.Len), style)
					for j := 0; j < int(sp.Len); j++ {
						ptr := &mixBuffer[int(sp.X)-minX+j]
						ptr.AddWithCover(colors[j], sp.Covers[j])
					}
				}
			}

			for _, sp := range slBin.Spans() {
				start := int(sp.X) - minX
				renBase.BlendColorHspan(int(sp.X), y, int(sp.Len), mixBuffer[start:start+int(sp.Len)], nil, basics.CoverFull)
			}
		}
	}
	elapsedMs := float64(time.Since(start).Microseconds()) / 1000.0

	text := fmt.Sprintf("%3.2f ms, %d triangles, %.0f tri/sec",
		elapsedMs, len(d.mesh.triangles), float64(len(d.mesh.triangles))/elapsedMs*1000.0)
	drawText(renBase, text)
}

func drawText(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[icol.SRGB], icol.RGBA8[icol.SRGB]], text string) {
	txt := gsv.NewGSVText()
	txt.SetSize(10.0, 0)
	txt.SetStartPoint(10.0, 10.0)
	txt.SetText(text)

	txtStroke := conv.NewConvStroke(txt)
	txtStroke.SetWidth(1.5)
	txtStroke.SetLineCap(basics.RoundCap)
	txtStroke.SetLineJoin(basics.RoundJoin)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()
	ras.Reset()
	ras.AddPath(&convVertexSourceRasVS{src: txtStroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, renBase, icol.RGBA8[icol.SRGB]{R: 255, G: 255, B: 255, A: 255})
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                  "Gouraud Mesh",
		Width:                  frameWidth,
		Height:                 frameHeight,
		FlipY:                  true,
		DisableLinearRGBToSRGB: true,
	}, newDemo())
}
