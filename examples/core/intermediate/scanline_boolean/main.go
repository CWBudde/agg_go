// Faithful port of AGG's scanline_boolean.cpp example.
//
// Renders two groups of circles generated along the edges of two
// quadrilaterals, then combines them using scanline boolean algebra (Union).
// Controls (radio box, sliders, reset checkbox) are rendered matching C++.
//
// The image is rendered in a flipped work buffer and copied with y-flip,
// matching the C++ original's flip_y=true coordinate system.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	checkboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
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

// Concrete types used throughout.
type (
	colorType = color.RGBA8[color.Linear]
	rasType   = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	rbType    = *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], colorType]
)

// srgba8 converts sRGB to linear, matching C++ agg::srgba8(r,g,b,a).
func srgba8(r, g, b, a uint8) colorType {
	return color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: r, G: g, B: b, A: a})
}

// toRGBA8 converts float RGBA to RGBA8[Linear].
func toRGBA8(c color.RGBA) colorType {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255.0 + 0.5)
	}
	return colorType{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

// --- Interactive polygon vertex source (matches C++ interactive_polygon) ---

type interactivePolygonVS struct {
	stroke *conv.ConvStroke
	ell    *shapes.Ellipse
	quad   [8]float64
	radius float64
	status int
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
		stroke: stroke,
		ell:    shapes.NewEllipse(),
		quad:   quad,
		radius: pointRadius,
	}
}

type pathToConvSource struct{ ps *path.PathStorageStl }

func (a *pathToConvSource) Rewind(pathID uint) { a.ps.Rewind(pathID) }
func (a *pathToConvSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

func (ip *interactivePolygonVS) Rewind(_ uint32) {
	ip.status = 0
	ip.stroke.Rewind(0)
}

func (ip *interactivePolygonVS) Vertex(x, y *float64) uint32 {
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

// --- Circle path vertex source (matches C++ generate_circles) ---

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

// --- Boolean renderer that writes to renderer_base ---

type boolRendererSolid struct {
	rb    rbType
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
			cover := basics.Int8u(0)
			if len(span.Covers) > 0 {
				cover = span.Covers[0]
			}
			r.rb.BlendHline(x, y, x-length-1, r.color, cover)
		} else {
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

// copyFlipY copies src to dst with vertical flip (y=0 at bottom -> y=0 at top).
func copyFlipY(src, dst []uint8, w, h int) {
	stride := w * 4
	for y := range h {
		srcOff := (h - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

// --- Control rendering (adapted from distortions example) ---

type ctrlVertexSourceAdapter struct {
	src interface {
		Rewind(pathID uint)
		Vertex() (x, y float64, cmd basics.PathCommand)
	}
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func renderCtrl(ras *rasType, sl *isc.ScanlineP8, rb rbType, c ctrlbase.Ctrl[color.RGBA]) {
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{src: c}, uint32(i))
		renscan.RenderScanlinesAASolid(ras, sl, rb, toRGBA8(c.Color(i)))
	}
}

// --- Demo ---

type demo struct {
	quad1 [8]float64
	quad2 [8]float64
}

func (d *demo) OnInit() {
	w := float64(imgWidth)
	h := float64(imgHeight)

	// C++ on_init() positions — coordinates are in flip_y=true space (y=0 at bottom)
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

	// Work buffer: y=0 at bottom (flip_y=true convention).
	workBuf := make([]uint8, w*h*4)
	for i := 0; i < len(workBuf); i += 4 {
		workBuf[i] = 255
		workBuf[i+1] = 255
		workBuf[i+2] = 255
		workBuf[i+3] = 255
	}

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(workBuf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)

	sl := isc.NewScanlineP8()
	ras := newRas()
	ras1 := newRas()
	ras2 := newRas()

	// Default: Union operation, opacity 1.0
	op := isc.BoolOr
	mul1 := 1.0
	mul2 := 1.0

	// Apply gamma (opacity) to rasterizers
	gammaFn1 := gamma.NewGammaMultiply(mul1)
	gammaFn2 := gamma.NewGammaMultiply(mul2)
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

	minX := min(sg1.MinX(), sg2.MinX())
	maxX := max(sg1.MaxX(), sg2.MaxX())
	sl1.sl.Reset(minX, maxX)
	sl2.sl.Reset(minX, maxX)
	slResult.sl.Reset(minX, maxX)

	sren := &boolRendererSolid{rb: rb, color: srgba8(0, 0, 0, 255)}
	isc.CombineShapesAA(op, sg1, sg2, sl1, sl2, slResult, sren)

	// --- Render quad outlines + vertex dots (matches C++ interactive_polygon) ---
	// C++: rgba(0, 0.3, 0.5, 0.6)
	quadColor := colorType{R: 0, G: 77, B: 128, A: 153}

	ras.Reset()
	ras.AddPath(newInteractivePolygonVS(d.quad1, 5.0), 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, quadColor)

	ras.Reset()
	ras.AddPath(newInteractivePolygonVS(d.quad2, 5.0), 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, quadColor)

	// --- Render controls ---
	// C++ constructor positions (using !flip_y = false):
	//   m_trans_type(420, 5.0, 420+130.0, 145.0, !flip_y)
	//   m_reset     (350, 5.0,  "Reset", !flip_y)
	//   m_mul1      (5.0,  5.0, 340.0, 12.0, !flip_y)
	//   m_mul2      (5.0, 20.0, 340.0, 27.0, !flip_y)
	ctrlTransType := rboxctrl.NewDefaultRboxCtrl(420, 5.0, 420+130.0, 145.0, false)
	ctrlTransType.AddItem("Union")
	ctrlTransType.AddItem("Intersection")
	ctrlTransType.AddItem("Linear XOR")
	ctrlTransType.AddItem("Saddle XOR")
	ctrlTransType.AddItem("Abs Diff XOR")
	ctrlTransType.AddItem("A-B")
	ctrlTransType.AddItem("B-A")
	ctrlTransType.SetCurItem(0)

	ctrlReset := checkboxctrl.NewDefaultCheckboxCtrl(350, 5.0, "Reset", false)

	ctrlMul1 := sliderctrl.NewSliderCtrl(5.0, 5.0, 340.0, 12.0, false)
	ctrlMul1.SetValue(mul1)
	ctrlMul1.SetLabel("Opacity1=%.3f")

	ctrlMul2 := sliderctrl.NewSliderCtrl(5.0, 20.0, 340.0, 27.0, false)
	ctrlMul2.SetValue(mul2)
	ctrlMul2.SetLabel("Opacity2=%.3f")

	renderCtrl(ras, sl, rb, ctrlTransType)
	renderCtrl(ras, sl, rb, ctrlReset)
	renderCtrl(ras, sl, rb, ctrlMul1)
	renderCtrl(ras, sl, rb, ctrlMul2)

	// Flip work buffer to output
	copyFlipY(workBuf, img.Data, w, h)
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
