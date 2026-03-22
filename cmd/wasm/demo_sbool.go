// Faithful port of AGG's scanline_boolean.cpp for the web demo.
//
// Renders two groups of circles generated along the edges of two interactive
// quadrilaterals, then combines them using scanline boolean algebra.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/gamma"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	isc "github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

// --- State ---

var (
	// Two quads stored as [x0,y0, x1,y1, x2,y2, x3,y3]
	sboolQuad1    [8]float64
	sboolQuad2    [8]float64
	sboolOp       = 0   // 0=Union … 6=B-A (maps to isc.BoolOp)
	sboolMul1     = 1.0 // Opacity for shape 1
	sboolMul2     = 1.0 // Opacity for shape 2
	sboolSelected = -1
	sboolQuadIdx  = 0 // 0 = quad1, 1 = quad2
	sboolDragDX   = 0.0
	sboolDragDY   = 0.0
	sboolInited   = false
)

func sboolInit() {
	w := float64(width)
	h := float64(height)
	sboolQuad1 = [8]float64{
		50, 200 - 20,
		w/2 - 25, 200,
		w/2 - 25, h - 50 - 20,
		50, h - 50,
	}
	sboolQuad2 = [8]float64{
		w/2 + 25, 200 - 20,
		w - 50, 200,
		w - 50, h - 50 - 20,
		w/2 + 25, h - 50,
	}
	sboolInited = true
}

// --- Concrete types ---

type (
	sboolColorType = color.RGBA8[color.Linear]
	sboolPfType    = renderer.PixelFormat[sboolColorType]
	sboolRbType    = *renderer.RendererBase[sboolPfType, sboolColorType]
	sboolRasType   = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
)

func sboolNewRas() *sboolRasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip())
}

// srgba8 converts sRGB values to linear for the pixel format.
func sboolSrgba8(r, g, b, a uint8) sboolColorType {
	return color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: r, G: g, B: b, A: a})
}

// --- Circle path vertex source (matches C++ generate_circles) ---

type sboolCirclePathVS struct {
	vx  []float64
	vy  []float64
	cmd []uint32
	idx int
}

func (p *sboolCirclePathVS) Rewind(_ uint32) { p.idx = 0 }
func (p *sboolCirclePathVS) Vertex(x, y *float64) uint32 {
	if p.idx >= len(p.cmd) {
		return uint32(basics.PathCmdStop)
	}
	*x = p.vx[p.idx]
	*y = p.vy[p.idx]
	c := p.cmd[p.idx]
	p.idx++
	return c
}

func sboolGenerateCircles(quad [8]float64, numCircles int, radius float64) *sboolCirclePathVS {
	ps := &sboolCirclePathVS{}
	ell := shapes.NewEllipse()
	for i := range 4 {
		n1 := i * 2
		n2 := ((i + 1) % 4) * 2
		for j := range numCircles {
			t := float64(j) / float64(numCircles)
			cx := quad[n1] + (quad[n2]-quad[n1])*t
			cy := quad[n1+1] + (quad[n2+1]-quad[n1+1])*t
			ell.Init(cx, cy, radius, radius, 100, false)
			ell.Rewind(0)
			for {
				var vx, vy float64
				cmd := ell.Vertex(&vx, &vy)
				if cmd == basics.PathCmdStop {
					break
				}
				ps.vx = append(ps.vx, vx)
				ps.vy = append(ps.vy, vy)
				ps.cmd = append(ps.cmd, uint32(cmd))
			}
		}
	}
	return ps
}

// --- Interactive polygon vertex source (matches C++ interactive_polygon) ---

type sboolInteractivePolygonVS struct {
	stroke *conv.ConvStroke
	ell    *shapes.Ellipse
	quad   [8]float64
	radius float64
	status int
}

func newSboolInteractivePolygonVS(quad [8]float64, pointRadius float64) *sboolInteractivePolygonVS {
	ps := path.NewPathStorageStl()
	ps.MoveTo(quad[0], quad[1])
	ps.LineTo(quad[2], quad[3])
	ps.LineTo(quad[4], quad[5])
	ps.LineTo(quad[6], quad[7])
	ps.ClosePolygon(0)

	src := &sboolPathToConvSource{ps: ps}
	stroke := conv.NewConvStroke(src)
	stroke.SetWidth(1.0)

	return &sboolInteractivePolygonVS{
		stroke: stroke,
		ell:    shapes.NewEllipse(),
		quad:   quad,
		radius: pointRadius,
	}
}

type sboolPathToConvSource struct{ ps *path.PathStorageStl }

func (a *sboolPathToConvSource) Rewind(pathID uint) { a.ps.Rewind(pathID) }
func (a *sboolPathToConvSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

func (ip *sboolInteractivePolygonVS) Rewind(_ uint32) {
	ip.status = 0
	ip.stroke.Rewind(0)
}

func (ip *sboolInteractivePolygonVS) Vertex(x, y *float64) uint32 {
	if ip.status == 0 {
		vx, vy, cmd := ip.stroke.Vertex()
		if !basics.IsStop(cmd) {
			*x = vx
			*y = vy
			return uint32(cmd)
		}
		ip.ell.Init(ip.quad[0], ip.quad[1], ip.radius, ip.radius, 32, false)
		ip.status = 1
	}
	cmd := ip.ell.Vertex(x, y)
	if !basics.IsStop(basics.PathCommand(cmd)) {
		return uint32(cmd)
	}
	if ip.status >= 4 {
		return uint32(basics.PathCmdStop)
	}
	idx := ip.status * 2
	ip.ell.Init(ip.quad[idx], ip.quad[idx+1], ip.radius, ip.radius, 32, false)
	ip.status++
	return uint32(ip.ell.Vertex(x, y))
}

// --- BooleanScanlineInterface adapter for ScanlineP8 ---

type sboolScanlineP8 struct {
	sl   *isc.ScanlineP8
	iter sboolScanlineP8Iter
}

type sboolScanlineP8Iter struct {
	spans []isc.SpanP8
	idx   int
}

func newSboolScanlineP8() *sboolScanlineP8 { return &sboolScanlineP8{sl: isc.NewScanlineP8()} }
func (s *sboolScanlineP8) Y() int          { return s.sl.Y() }
func (s *sboolScanlineP8) NumSpans() int   { return s.sl.NumSpans() }
func (s *sboolScanlineP8) ResetSpans()     { s.sl.ResetSpans() }
func (s *sboolScanlineP8) AddCell(x int, cover uint) {
	s.sl.AddCell(x, cover)
}

func (s *sboolScanlineP8) AddCells(x, length int, covers []basics.Int8u) {
	s.sl.AddCells(x, length, covers)
}

func (s *sboolScanlineP8) AddSpan(x, length int, cover basics.Int8u) {
	s.sl.AddSpan(x, length, uint(cover))
}
func (s *sboolScanlineP8) Finalize(y int) { s.sl.Finalize(y) }
func (s *sboolScanlineP8) Begin() isc.ScanlineIterator {
	s.iter.spans = s.sl.Spans()
	s.iter.idx = 0
	return &s.iter
}

func (it *sboolScanlineP8Iter) GetSpan() isc.SpanInfo {
	span := it.spans[it.idx]
	return isc.SpanInfo{X: int(span.X), Len: int(span.Len), Covers: span.Covers}
}

func (it *sboolScanlineP8Iter) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

// --- Storage scanline adapter for rendering rasterizer → ScanlineStorageAA ---

type sboolStorageScanlineP8 struct {
	sl   *isc.ScanlineP8
	iter sboolStorageScanlineP8Iter
}

type sboolStorageScanlineP8Iter struct {
	spans []isc.SpanP8
	idx   int
}

func (s *sboolStorageScanlineP8) Y() int        { return s.sl.Y() }
func (s *sboolStorageScanlineP8) NumSpans() int { return s.sl.NumSpans() }
func (s *sboolStorageScanlineP8) ResetSpans()   { s.sl.ResetSpans() }
func (s *sboolStorageScanlineP8) AddSpan(x, length int, cover basics.Int8u) {
	s.sl.AddSpan(x, length, uint(cover))
}

func (s *sboolStorageScanlineP8) AddCells(x, length int, covers []basics.Int8u) {
	s.sl.AddCells(x, length, covers)
}
func (s *sboolStorageScanlineP8) Finalize(y int) { s.sl.Finalize(y) }
func (s *sboolStorageScanlineP8) Begin() isc.ScanlineIterator {
	s.iter.spans = s.sl.Spans()
	s.iter.idx = 0
	return &s.iter
}

func (it *sboolStorageScanlineP8Iter) GetSpan() isc.SpanInfo {
	span := it.spans[it.idx]
	return isc.SpanInfo{X: int(span.X), Len: int(span.Len), Covers: span.Covers}
}

func (it *sboolStorageScanlineP8Iter) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

// --- Storage rasterizer adapter for CombineShapesAA ---

type sboolAAStorageRasterizer struct {
	storage *isc.ScanlineStorageAA[basics.Int8u]
	embed   *isc.EmbeddedScanline[basics.Int8u]
}

func newSboolAAStorageRasterizer(storage *isc.ScanlineStorageAA[basics.Int8u]) *sboolAAStorageRasterizer {
	return &sboolAAStorageRasterizer{storage: storage, embed: isc.NewEmbeddedScanline(storage)}
}

func (r *sboolAAStorageRasterizer) RewindScanlines() bool { return r.storage.RewindScanlines() }
func (r *sboolAAStorageRasterizer) MinX() int             { return r.storage.MinX() }
func (r *sboolAAStorageRasterizer) MinY() int             { return r.storage.MinY() }
func (r *sboolAAStorageRasterizer) MaxX() int             { return r.storage.MaxX() }
func (r *sboolAAStorageRasterizer) MaxY() int             { return r.storage.MaxY() }
func (r *sboolAAStorageRasterizer) SweepScanline(sl isc.BooleanScanlineInterface) bool {
	if !r.storage.SweepEmbeddedScanline(r.embed) {
		return false
	}
	sl.ResetSpans()
	iter := r.embed.Begin()
	for i := 0; i < r.embed.NumSpans(); i++ {
		span := iter.GetSpan()
		if span.Len < 0 {
			cover := basics.Int8u(0)
			if len(span.Covers) > 0 {
				cover = span.Covers[0]
			}
			sl.AddSpan(int(span.X), int(-span.Len), cover)
		} else {
			sl.AddCells(int(span.X), int(span.Len), span.Covers)
		}
		if i < r.embed.NumSpans()-1 {
			iter.Next()
		}
	}
	sl.Finalize(r.embed.Y())
	return true
}

// --- Boolean renderer that writes to renderer_base ---

type sboolRendererSolid struct {
	rb sboolRbType
	c  sboolColorType
}

func (r *sboolRendererSolid) Prepare() {}
func (r *sboolRendererSolid) Render(sl isc.BooleanScanlineInterface) {
	y := sl.Y()
	iter := sl.Begin()
	for i := 0; i < sl.NumSpans(); i++ {
		span := iter.GetSpan()
		x := span.X
		length := span.Len
		if length < 0 {
			cover := basics.Int8u(0)
			if len(span.Covers) > 0 {
				cover = span.Covers[0]
			}
			r.rb.BlendHline(x, y, x-length-1, r.c, cover)
		} else {
			r.rb.BlendSolidHspan(x, y, length, r.c, span.Covers)
		}
		if i < sl.NumSpans()-1 {
			iter.Next()
		}
	}
}

// --- Helpers ---

func sboolRenderRasterizerToStorage(
	ras *sboolRasType,
	sl *isc.ScanlineP8,
	storage *isc.ScanlineStorageAA[basics.Int8u],
) {
	storage.Prepare()
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	storageSL := &sboolStorageScanlineP8{sl: sl}
	for ras.SweepScanline(sl) {
		storage.Render(storageSL)
	}
}

func sboolMapOp(op int) isc.BoolOp {
	switch op {
	case 0:
		return isc.BoolOr // Union
	case 1:
		return isc.BoolAnd // Intersection
	case 2:
		return isc.BoolXor // Linear XOR
	case 3:
		return isc.BoolXorSaddle // Saddle XOR
	case 4:
		return isc.BoolXorAbsDiff // Abs Diff XOR
	case 5:
		return isc.BoolAMinusB // A-B
	case 6:
		return isc.BoolBMinusA // B-A
	default:
		return isc.BoolOr
	}
}

// manualRenderSolid sweeps a rasterizer and renders with a single color
// using a manual loop (avoids generic interface mismatches in WASM context).
func sboolManualRenderSolid(
	ras *sboolRasType,
	rb sboolRbType,
	c sboolColorType,
) {
	sl := isc.NewScanlineP8()
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(sl) {
		y := sl.Y()
		for _, span := range sl.Spans() {
			if span.Len > 0 {
				rb.BlendSolidHspan(int(span.X), y, int(span.Len), c, span.Covers)
			} else if span.Len < 0 {
				cover := basics.Int8u(0)
				if len(span.Covers) > 0 {
					cover = span.Covers[0]
				}
				rb.BlendHline(int(span.X), y, int(span.X)-int(span.Len)-1, c, cover)
			}
		}
	}
}

// --- Main draw ---

func drawSBoolDemo() {
	if !sboolInited {
		sboolInit()
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[sboolColorType], sboolColorType](pixFmt)

	ras := sboolNewRas()
	ras1 := sboolNewRas()
	ras2 := sboolNewRas()

	// Apply gamma (opacity) to rasterizers
	gammaFn1 := gamma.NewGammaMultiply(sboolMul1)
	gammaFn2 := gamma.NewGammaMultiply(sboolMul2)
	ras1.SetGamma(gammaFn1.Apply)
	ras2.SetGamma(gammaFn2.Apply)

	ras.ClipBox(0, 0, float64(img.Width()), float64(img.Height()))

	// Generate circles along quad edges
	ps1 := sboolGenerateCircles(sboolQuad1, 5, 20)
	ps2 := sboolGenerateCircles(sboolQuad2, 5, 20)

	// ras1 uses even-odd filling rule (matches C++)
	ras1.FillingRule(basics.FillEvenOdd)

	// Render shape 1 semi-transparently: srgba8(240, 255, 200, 100)
	ras1.AddPath(ps1, 0)
	sboolManualRenderSolid(ras1, rb, sboolSrgba8(240, 255, 200, 100))

	// Render shape 2 semi-transparently: srgba8(255, 240, 240, 100)
	ras2.AddPath(ps2, 0)
	sboolManualRenderSolid(ras2, rb, sboolSrgba8(255, 240, 240, 100))

	// --- Scanline boolean combine ---
	storage1 := isc.NewScanlineStorageAA[basics.Int8u]()
	storage2 := isc.NewScanlineStorageAA[basics.Int8u]()
	slRaster := isc.NewScanlineP8()

	ras1.Reset()
	ras1.FillingRule(basics.FillEvenOdd)
	ras1.SetGamma(gammaFn1.Apply)
	ras1.AddPath(ps1, 0)
	sboolRenderRasterizerToStorage(ras1, slRaster, storage1)

	ras2.Reset()
	ras2.SetGamma(gammaFn2.Apply)
	ras2.AddPath(ps2, 0)
	sboolRenderRasterizerToStorage(ras2, slRaster, storage2)

	sg1 := newSboolAAStorageRasterizer(storage1)
	sg2 := newSboolAAStorageRasterizer(storage2)
	sl1 := newSboolScanlineP8()
	sl2 := newSboolScanlineP8()
	slResult := newSboolScanlineP8()

	minX := min(sg1.MinX(), sg2.MinX())
	maxX := max(sg1.MaxX(), sg2.MaxX())
	sl1.sl.Reset(minX, maxX)
	sl2.sl.Reset(minX, maxX)
	slResult.sl.Reset(minX, maxX)

	sren := &sboolRendererSolid{rb: rb, c: sboolSrgba8(0, 0, 0, 255)}
	isc.CombineShapesAA(sboolMapOp(sboolOp), sg1, sg2, sl1, sl2, slResult, sren)

	// --- Render quad outlines ---
	quadColor := sboolColorType{
		R: 0,
		G: 77,  // 0.3 * 255
		B: 128, // 0.5 * 255
		A: 153, // 0.6 * 255
	}

	ras.Reset()
	q1 := newSboolInteractivePolygonVS(sboolQuad1, 5.0)
	ras.AddPath(q1, 0)
	sboolManualRenderSolid(ras, rb, quadColor)

	ras.Reset()
	q2 := newSboolInteractivePolygonVS(sboolQuad2, 5.0)
	ras.AddPath(q2, 0)
	sboolManualRenderSolid(ras, rb, quadColor)
}

// --- Mouse interaction ---

func handleSBoolMouseDown(x, y float64) bool {
	sboolSelected = -1
	// Check quad1 corners
	for i := range 4 {
		qx, qy := sboolQuad1[i*2], sboolQuad1[i*2+1]
		if math.Sqrt((x-qx)*(x-qx)+(y-qy)*(y-qy)) < 15 {
			sboolSelected = i
			sboolQuadIdx = 0
			sboolDragDX = x - qx
			sboolDragDY = y - qy
			return true
		}
	}
	// Check quad2 corners
	for i := range 4 {
		qx, qy := sboolQuad2[i*2], sboolQuad2[i*2+1]
		if math.Sqrt((x-qx)*(x-qx)+(y-qy)*(y-qy)) < 15 {
			sboolSelected = i
			sboolQuadIdx = 1
			sboolDragDX = x - qx
			sboolDragDY = y - qy
			return true
		}
	}
	return false
}

func handleSBoolMouseMove(x, y float64) bool {
	if sboolSelected < 0 {
		return false
	}
	idx := sboolSelected * 2
	if sboolQuadIdx == 0 {
		sboolQuad1[idx] = x - sboolDragDX
		sboolQuad1[idx+1] = y - sboolDragDY
	} else {
		sboolQuad2[idx] = x - sboolDragDX
		sboolQuad2[idx+1] = y - sboolDragDY
	}
	return true
}

func handleSBoolMouseUp() {
	sboolSelected = -1
}

// --- Node getters/setters for URL persistence ---

func getSBoolQuad1() [8]float64 { return sboolQuad1 }
func getSBoolQuad2() [8]float64 { return sboolQuad2 }
