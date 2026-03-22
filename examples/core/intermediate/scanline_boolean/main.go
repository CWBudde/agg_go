// Faithful port of AGG's scanline_boolean.cpp example.
//
// The demo renders two groups of circles generated along the edges of two
// quadrilaterals, then combines them using scanline boolean algebra (Union).
// For the static (headless) output the defaults are Union with opacity 1.0.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/gamma"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	isc "github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	imgWidth  = 800
	imgHeight = 600
)

// Concrete color and renderer types used throughout.
type colorType = color.RGBA8[color.Linear]
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

// srgba8 builds an sRGB color and converts to linear for the pixel format,
// matching C++ agg::srgba8(r,g,b,a).
func srgba8(r, g, b, a uint8) colorType {
	return color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: r, G: g, B: b, A: a})
}

// --- Interactive polygon vertex source (matches C++ interactive_polygon) ---
// Emits: conv_stroke of the closed polygon, then filled ellipses at each vertex.

type interactivePolygonVS struct {
	ps     *path.PathStorageStl
	stroke *conv.ConvStroke
	ell    *shapes.Ellipse
	quad   [8]float64
	radius float64
	status int // 0=stroke phase, 1..n=ellipse phases
}

func newInteractivePolygonVS(quad [8]float64, pointRadius float64) *interactivePolygonVS {
	ps := path.NewPathStorageStl()
	ps.MoveTo(quad[0], quad[1])
	ps.LineTo(quad[2], quad[3])
	ps.LineTo(quad[4], quad[5])
	ps.LineTo(quad[6], quad[7])
	ps.ClosePolygon(0)

	src := &pathToConvSource{ps: ps}
	stroke := conv.NewConvStroke(src)
	stroke.SetWidth(1.0)

	return &interactivePolygonVS{
		ps:     ps,
		stroke: stroke,
		ell:    shapes.NewEllipse(),
		quad:   quad,
		radius: pointRadius,
	}
}

// pathToConvSource adapts PathStorageStl to conv.VertexSource.
type pathToConvSource struct{ ps *path.PathStorageStl }

func (a *pathToConvSource) Rewind(pathID uint) { a.ps.Rewind(pathID) }
func (a *pathToConvSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

// Rewind implements rasterizer.VertexSource.
func (ip *interactivePolygonVS) Rewind(_ uint32) {
	ip.status = 0
	ip.stroke.Rewind(0)
}

// Vertex implements rasterizer.VertexSource.
func (ip *interactivePolygonVS) Vertex(x, y *float64) uint32 {
	if ip.status == 0 {
		// Stroke phase
		vx, vy, cmd := ip.stroke.Vertex()
		if !basics.IsStop(cmd) {
			*x = vx
			*y = vy
			return uint32(cmd)
		}
		// Stroke done, start first ellipse
		ip.ell.Init(ip.quad[0], ip.quad[1], ip.radius, ip.radius, 32, false)
		ip.status = 1
	}
	// Ellipse phases
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

// --- Circle path vertex source ---

type circlePathVS struct {
	vx  []float64
	vy  []float64
	cmd []uint32
	idx int
}

func (p *circlePathVS) Rewind(_ uint32) { p.idx = 0 }
func (p *circlePathVS) Vertex(x, y *float64) uint32 {
	if p.idx >= len(p.cmd) {
		return uint32(basics.PathCmdStop)
	}
	*x = p.vx[p.idx]
	*y = p.vy[p.idx]
	c := p.cmd[p.idx]
	p.idx++
	return c
}

// generateCircles mirrors the C++ generate_circles function.
// It creates circles along the edges of a quad (4 corners -> 4 edges).
func generateCircles(quad [8]float64, numCircles int, radius float64) *circlePathVS {
	ps := &circlePathVS{}
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

// --- BooleanScanlineInterface adapter for ScanlineP8 ---

type boolScanlineP8 struct {
	sl   *isc.ScanlineP8
	iter boolScanlineP8Iter
}

type boolScanlineP8Iter struct {
	spans []isc.SpanP8
	idx   int
}

func newBoolScanlineP8() *boolScanlineP8 { return &boolScanlineP8{sl: isc.NewScanlineP8()} }
func (s *boolScanlineP8) Y() int         { return s.sl.Y() }
func (s *boolScanlineP8) NumSpans() int  { return s.sl.NumSpans() }
func (s *boolScanlineP8) ResetSpans()    { s.sl.ResetSpans() }
func (s *boolScanlineP8) AddCell(x int, cover uint) {
	s.sl.AddCell(x, cover)
}
func (s *boolScanlineP8) AddCells(x, length int, covers []basics.Int8u) {
	s.sl.AddCells(x, length, covers)
}
func (s *boolScanlineP8) AddSpan(x, length int, cover basics.Int8u) {
	s.sl.AddSpan(x, length, uint(cover))
}
func (s *boolScanlineP8) Finalize(y int) { s.sl.Finalize(y) }
func (s *boolScanlineP8) Begin() isc.ScanlineIterator {
	s.iter.spans = s.sl.Spans()
	s.iter.idx = 0
	return &s.iter
}

func (it *boolScanlineP8Iter) GetSpan() isc.SpanInfo {
	span := it.spans[it.idx]
	return isc.SpanInfo{X: int(span.X), Len: int(span.Len), Covers: span.Covers}
}
func (it *boolScanlineP8Iter) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

// --- Boolean renderer that renders directly using renderer_base ---

type boolRendererSolid struct {
	rb    *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], colorType]
	color colorType
}

func (r *boolRendererSolid) Prepare() {}
func (r *boolRendererSolid) Render(sl isc.BooleanScanlineInterface) {
	y := sl.Y()
	iter := sl.Begin()
	for i := 0; i < sl.NumSpans(); i++ {
		span := iter.GetSpan()
		x := span.X
		length := span.Len
		if length < 0 {
			// Solid span: single cover value for |length| pixels
			cover := basics.Int8u(0)
			if len(span.Covers) > 0 {
				cover = span.Covers[0]
			}
			r.rb.BlendHline(x, y, x-length-1, r.color, cover)
		} else {
			// Per-pixel coverage span
			r.rb.BlendSolidHspan(x, y, length, r.color, span.Covers)
		}
		if i < sl.NumSpans()-1 {
			iter.Next()
		}
	}
}

// --- Storage rasterizer adapter for CombineShapesAA ---

type aaStorageBoolRasterizer struct {
	storage *isc.ScanlineStorageAA[basics.Int8u]
	embed   *isc.EmbeddedScanline[basics.Int8u]
}

func newAAStorageBoolRasterizer(storage *isc.ScanlineStorageAA[basics.Int8u]) *aaStorageBoolRasterizer {
	return &aaStorageBoolRasterizer{storage: storage, embed: isc.NewEmbeddedScanline(storage)}
}

func (r *aaStorageBoolRasterizer) RewindScanlines() bool { return r.storage.RewindScanlines() }
func (r *aaStorageBoolRasterizer) MinX() int             { return r.storage.MinX() }
func (r *aaStorageBoolRasterizer) MinY() int             { return r.storage.MinY() }
func (r *aaStorageBoolRasterizer) MaxX() int             { return r.storage.MaxX() }
func (r *aaStorageBoolRasterizer) MaxY() int             { return r.storage.MaxY() }
func (r *aaStorageBoolRasterizer) SweepScanline(sl isc.BooleanScanlineInterface) bool {
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

// --- Storage scanline adapter for ScanlineStorageAA.Render ---

type storageScanlineP8 struct {
	sl   *isc.ScanlineP8
	iter storageScanlineP8Iter
}

type storageScanlineP8Iter struct {
	spans []isc.SpanP8
	idx   int
}

func (s *storageScanlineP8) Y() int        { return s.sl.Y() }
func (s *storageScanlineP8) NumSpans() int { return s.sl.NumSpans() }
func (s *storageScanlineP8) ResetSpans()   { s.sl.ResetSpans() }
func (s *storageScanlineP8) AddSpan(x, length int, cover basics.Int8u) {
	s.sl.AddSpan(x, length, uint(cover))
}
func (s *storageScanlineP8) AddCells(x, length int, covers []basics.Int8u) {
	s.sl.AddCells(x, length, covers)
}
func (s *storageScanlineP8) Finalize(y int) { s.sl.Finalize(y) }
func (s *storageScanlineP8) Begin() isc.ScanlineIterator {
	s.iter.spans = s.sl.Spans()
	s.iter.idx = 0
	return &s.iter
}

func (it *storageScanlineP8Iter) GetSpan() isc.SpanInfo {
	span := it.spans[it.idx]
	return isc.SpanInfo{X: int(span.X), Len: int(span.Len), Covers: span.Covers}
}
func (it *storageScanlineP8Iter) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

// --- Helpers ---

func newRas() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip())
}

func renderRasterizerToStorage(
	ras *rasType,
	sl *isc.ScanlineP8,
	storage *isc.ScanlineStorageAA[basics.Int8u],
) {
	storage.Prepare()
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	storageSL := &storageScanlineP8{sl: sl}
	for ras.SweepScanline(sl) {
		storage.Render(storageSL)
	}
}

// --- Demo ---

type demo struct {
	quad1 [8]float64 // 4 points as (x0,y0, x1,y1, x2,y2, x3,y3)
	quad2 [8]float64
}

func (d *demo) OnInit() {
	w := float64(imgWidth)
	h := float64(imgHeight)

	// C++ on_init() positions
	d.quad1 = [8]float64{
		50, 200 - 20,
		w/2 - 25, 200,
		w/2 - 25, h - 50 - 20,
		50, h - 50,
	}
	d.quad2 = [8]float64{
		w/2 + 25, 200 - 20,
		w - 50, 200,
		w - 50, h - 50 - 20,
		w/2 + 25, h - 50,
	}
}

func (d *demo) Render(img *agg.Image) {
	w := img.Width()
	h := img.Height()

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(colorType{R: 255, G: 255, B: 255, A: 255})

	sl := isc.NewScanlineP8()
	ras := newRas()
	ras1 := newRas()
	ras2 := newRas()

	// Default: Union operation, opacity 1.0
	op := isc.BoolOr

	// Apply gamma (opacity) to rasterizers
	gammaFn1 := gamma.NewGammaMultiply(1.0)
	gammaFn2 := gamma.NewGammaMultiply(1.0)
	ras1.SetGamma(gammaFn1.Apply)
	ras2.SetGamma(gammaFn2.Apply)

	ras.ClipBox(0, 0, float64(w), float64(h))

	// Generate circles along quad edges
	ps1 := generateCircles(d.quad1, 5, 20)
	ps2 := generateCircles(d.quad2, 5, 20)

	// ras1 uses even-odd filling rule (matches C++)
	ras1.FillingRule(basics.FillEvenOdd)

	// Render shape 1 semi-transparently: srgba8(240, 255, 200, 100)
	ras1.AddPath(ps1, 0)
	renscan.RenderScanlinesAASolid(ras1, sl, rb, srgba8(240, 255, 200, 100))

	// Render shape 2 semi-transparently: srgba8(255, 240, 240, 100)
	ras2.AddPath(ps2, 0)
	renscan.RenderScanlinesAASolid(ras2, sl, rb, srgba8(255, 240, 240, 100))

	// --- Scanline boolean combine ---
	storage1 := isc.NewScanlineStorageAA[basics.Int8u]()
	storage2 := isc.NewScanlineStorageAA[basics.Int8u]()
	slRaster := isc.NewScanlineP8()

	// Re-rasterize into storage (same settings as above)
	ras1.Reset()
	ras1.FillingRule(basics.FillEvenOdd)
	ras1.SetGamma(gammaFn1.Apply)
	ras1.AddPath(ps1, 0)
	renderRasterizerToStorage(ras1, slRaster, storage1)

	ras2.Reset()
	ras2.SetGamma(gammaFn2.Apply)
	ras2.AddPath(ps2, 0)
	renderRasterizerToStorage(ras2, slRaster, storage2)

	sg1 := newAAStorageBoolRasterizer(storage1)
	sg2 := newAAStorageBoolRasterizer(storage2)
	sl1 := newBoolScanlineP8()
	sl2 := newBoolScanlineP8()
	slResult := newBoolScanlineP8()

	// Prepare scanlines with combined bounds
	minX := min(sg1.MinX(), sg2.MinX())
	maxX := max(sg1.MaxX(), sg2.MaxX())
	sl1.sl.Reset(minX, maxX)
	sl2.sl.Reset(minX, maxX)
	slResult.sl.Reset(minX, maxX)

	// Boolean result rendered in black: srgba8(0, 0, 0)
	sren := &boolRendererSolid{
		rb:    rb,
		color: srgba8(0, 0, 0, 255),
	}
	isc.CombineShapesAA(op, sg1, sg2, sl1, sl2, slResult, sren)

	// --- Render quad outlines + vertex dots (matches C++ interactive_polygon) ---
	// C++: rgba(0, 0.3, 0.5, 0.6)
	quadColor := colorType{R: 0, G: 77, B: 128, A: 153}

	ras.Reset()
	q1vs := newInteractivePolygonVS(d.quad1, 5.0)
	ras.AddPath(q1vs, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, quadColor)

	ras.Reset()
	q2vs := newInteractivePolygonVS(d.quad2, 5.0)
	ras.AddPath(q2vs, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, quadColor)
}

func main() {
	d := &demo{}
	d.OnInit()
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Scanline Boolean",
		Width:  imgWidth,
		Height: imgHeight,
	}, d)
}
