package scanlineboolean2

import (
	"fmt"
	"math"
	"time"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/demo/aggshapes"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	isc "github.com/MeKo-Christian/agg_go/internal/scanline"
)

type Config struct {
	Mode         int
	FillRule     int
	ScanlineType int
	Operation    int
	CenterX      float64
	CenterY      float64
}

const (
	referenceWidth  = 655.0
	referenceHeight = 520.0
)

type pt struct {
	x float64
	y float64
}

type contour []pt

type colorDef struct {
	r float64
	g float64
	b float64
	a float64
}

type rasterVertexSource interface {
	Rewind(pathID uint32)
	Vertex(x, y *float64) uint32
}

type rasterPathAdapter struct {
	src interface {
		Rewind(pathID uint)
		Vertex() (x, y float64, cmd basics.PathCommand)
	}
}

func (a *rasterPathAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *rasterPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type rasterScanlineAdapter struct{ sl *isc.ScanlineU8 }

func (a *rasterScanlineAdapter) ResetSpans()                    { a.sl.ResetSpans() }
func (a *rasterScanlineAdapter) AddCell(x int, cover uint32)    { a.sl.AddCell(x, uint(cover)) }
func (a *rasterScanlineAdapter) AddSpan(x, l int, cover uint32) { a.sl.AddSpan(x, l, uint(cover)) }
func (a *rasterScanlineAdapter) Finalize(y int)                 { a.sl.Finalize(y) }
func (a *rasterScanlineAdapter) NumSpans() int                  { return a.sl.NumSpans() }

type rasterScanlineP8Adapter struct{ sl *isc.ScanlineP8 }

func (a *rasterScanlineP8Adapter) ResetSpans()                    { a.sl.ResetSpans() }
func (a *rasterScanlineP8Adapter) AddCell(x int, cover uint32)    { a.sl.AddCell(x, uint(cover)) }
func (a *rasterScanlineP8Adapter) AddSpan(x, l int, cover uint32) { a.sl.AddSpan(x, l, uint(cover)) }
func (a *rasterScanlineP8Adapter) Finalize(y int)                 { a.sl.Finalize(y) }
func (a *rasterScanlineP8Adapter) NumSpans() int                  { return a.sl.NumSpans() }

type rasterScanlineBinAdapter struct{ sl *isc.ScanlineBin }

func (a *rasterScanlineBinAdapter) ResetSpans()                    { a.sl.ResetSpans() }
func (a *rasterScanlineBinAdapter) AddCell(x int, cover uint32)    { a.sl.AddCell(x, uint(cover)) }
func (a *rasterScanlineBinAdapter) AddSpan(x, l int, cover uint32) { a.sl.AddSpan(x, l, uint(cover)) }
func (a *rasterScanlineBinAdapter) Finalize(y int)                 { a.sl.Finalize(y) }
func (a *rasterScanlineBinAdapter) NumSpans() int                  { return a.sl.NumSpans() }

type storageScanlineU8 struct {
	sl   *isc.ScanlineU8
	iter storageScanlineU8Iter
}

type storageScanlineU8Iter struct {
	spans []isc.Span
	idx   int
}

func (s *storageScanlineU8) Y() int        { return s.sl.Y() }
func (s *storageScanlineU8) NumSpans() int { return s.sl.NumSpans() }
func (s *storageScanlineU8) ResetSpans()   { s.sl.ResetSpans() }
func (s *storageScanlineU8) AddSpan(x, length int, cover basics.Int8u) {
	s.sl.AddSpan(x, length, uint(cover))
}
func (s *storageScanlineU8) AddCells(x, length int, covers []basics.Int8u) {
	s.sl.AddCells(x, length, covers)
}
func (s *storageScanlineU8) Finalize(y int) { s.sl.Finalize(y) }
func (s *storageScanlineU8) Begin() isc.ScanlineIterator {
	s.iter.spans = s.sl.Spans()
	s.iter.idx = 0
	return &s.iter
}
func (it *storageScanlineU8Iter) GetSpan() isc.SpanInfo {
	span := it.spans[it.idx]
	return isc.SpanInfo{X: int(span.X), Len: int(span.Len), Covers: span.Covers}
}
func (it *storageScanlineU8Iter) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

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

type boolScanlineU8 struct {
	sl   *isc.ScanlineU8
	iter boolScanlineU8Iter
}

type boolScanlineU8Iter struct {
	spans []isc.Span
	idx   int
}

func newBoolScanlineU8() *boolScanlineU8 { return &boolScanlineU8{sl: isc.NewScanlineU8()} }
func (s *boolScanlineU8) Y() int         { return s.sl.Y() }
func (s *boolScanlineU8) NumSpans() int  { return s.sl.NumSpans() }
func (s *boolScanlineU8) ResetSpans()    { s.sl.ResetSpans() }
func (s *boolScanlineU8) AddCell(x int, cover uint) {
	s.sl.AddCell(x, cover)
}
func (s *boolScanlineU8) AddCells(x, length int, covers []basics.Int8u) {
	s.sl.AddCells(x, length, covers)
}
func (s *boolScanlineU8) AddSpan(x, length int, cover basics.Int8u) {
	s.sl.AddSpan(x, length, uint(cover))
}
func (s *boolScanlineU8) Finalize(y int) { s.sl.Finalize(y) }
func (s *boolScanlineU8) Begin() renscan.ScanlineIterator {
	s.iter.spans = s.sl.Spans()
	s.iter.idx = 0
	return &s.iter
}
func (it *boolScanlineU8Iter) GetSpan() renscan.SpanData {
	span := it.spans[it.idx]
	return renscan.SpanData{X: int(span.X), Len: int(span.Len), Covers: span.Covers}
}
func (it *boolScanlineU8Iter) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

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
func (s *boolScanlineP8) Begin() renscan.ScanlineIterator {
	s.iter.spans = s.sl.Spans()
	s.iter.idx = 0
	return &s.iter
}
func (it *boolScanlineP8Iter) GetSpan() renscan.SpanData {
	span := it.spans[it.idx]
	return renscan.SpanData{X: int(span.X), Len: int(span.Len), Covers: span.Covers}
}
func (it *boolScanlineP8Iter) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

type boolScanlineBin struct {
	sl   *isc.ScanlineBin
	iter boolScanlineBinIter
}

type boolScanlineBinIter struct {
	spans []isc.SpanBin
	idx   int
}

func newBoolScanlineBin() *boolScanlineBin { return &boolScanlineBin{sl: isc.NewScanlineBin()} }
func (s *boolScanlineBin) Y() int          { return s.sl.Y() }
func (s *boolScanlineBin) NumSpans() int   { return s.sl.NumSpans() }
func (s *boolScanlineBin) ResetSpans()     { s.sl.ResetSpans() }
func (s *boolScanlineBin) AddCell(x int, cover uint) {
	s.sl.AddCell(x, cover)
}
func (s *boolScanlineBin) AddCells(x, length int, covers []basics.Int8u) {
	s.sl.AddCells(x, length, covers)
}
func (s *boolScanlineBin) AddSpan(x, length int, cover basics.Int8u) {
	s.sl.AddSpan(x, length, uint(cover))
}
func (s *boolScanlineBin) Finalize(y int) { s.sl.Finalize(y) }
func (s *boolScanlineBin) Begin() renscan.ScanlineIterator {
	s.iter.spans = s.sl.Spans()
	s.iter.idx = 0
	return &s.iter
}
func (it *boolScanlineBinIter) GetSpan() renscan.SpanData {
	span := it.spans[it.idx]
	return renscan.SpanData{X: int(span.X), Len: int(span.Len), Covers: []basics.Int8u{255}}
}
func (it *boolScanlineBinIter) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

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

type binStorageBoolRasterizer struct {
	storage *isc.ScanlineStorageBin
	embed   *isc.EmbeddedScanlineBin
}

func newBinStorageBoolRasterizer(storage *isc.ScanlineStorageBin) *binStorageBoolRasterizer {
	return &binStorageBoolRasterizer{storage: storage, embed: isc.NewEmbeddedScanlineBin(storage)}
}

func (r *binStorageBoolRasterizer) RewindScanlines() bool { return r.storage.RewindScanlines() }
func (r *binStorageBoolRasterizer) MinX() int             { return r.storage.MinX() }
func (r *binStorageBoolRasterizer) MinY() int             { return r.storage.MinY() }
func (r *binStorageBoolRasterizer) MaxX() int             { return r.storage.MaxX() }
func (r *binStorageBoolRasterizer) MaxY() int             { return r.storage.MaxY() }
func (r *binStorageBoolRasterizer) SweepScanline(sl isc.BooleanScanlineInterface) bool {
	if !r.storage.SweepEmbeddedScanline(r.embed) {
		return false
	}
	sl.ResetSpans()
	iter := r.embed.Begin()
	for i := 0; i < r.embed.NumSpans(); i++ {
		span := iter.Span()
		sl.AddSpan(int(span.X), int(span.Len), 255)
		if i < r.embed.NumSpans()-1 {
			iter.Next()
		}
	}
	sl.Finalize(r.embed.Y())
	return true
}

type boolRenderer struct {
	scanlines []boolRenderedScanline
}

type boolRenderedScanline struct {
	y     int
	spans []renscan.SpanData
}

func (r *boolRenderer) Prepare() {
	r.scanlines = r.scanlines[:0]
}

func (r *boolRenderer) Render(sl isc.BooleanScanlineInterface) {
	item := boolRenderedScanline{y: sl.Y(), spans: make([]renscan.SpanData, 0, sl.NumSpans())}
	iter := sl.Begin()
	for i := 0; i < sl.NumSpans(); i++ {
		span := iter.GetSpan()
		item.spans = append(item.spans, renscan.SpanData{
			X:      span.X,
			Len:    span.Len,
			Covers: append([]basics.Int8u(nil), span.Covers...),
		})
		if i < sl.NumSpans()-1 {
			iter.Next()
		}
	}
	r.scanlines = append(r.scanlines, item)
}

func Draw(ctx *agg.Context, cfg Config) {
	w := float64(ctx.GetImage().Width())
	h := float64(ctx.GetImage().Height())
	frameOffX := (w - referenceWidth) * 0.5
	frameOffY := (h - referenceHeight) * 0.5

	if math.IsNaN(cfg.CenterX) || math.IsNaN(cfg.CenterY) {
		cfg.CenterX = w * 0.5
		cfg.CenterY = h * 0.5
	}

	cfg.Mode = clampInt(cfg.Mode, 0, 4)
	cfg.FillRule = clampInt(cfg.FillRule, 0, 1)
	cfg.ScanlineType = clampInt(cfg.ScanlineType, 0, 2)
	cfg.Operation = clampInt(cfg.Operation, 0, 6)

	// The upstream demo runs with flip_y=true. Convert the dragged center from
	// screen coordinates into the original 655x520 reference frame.
	cfg.CenterX = cfg.CenterX - frameOffX
	cfg.CenterY = referenceHeight - (cfg.CenterY - frameOffY)

	a, b := buildShapes(cfg, referenceWidth, referenceHeight)
	a = transformContours(mirrorContoursY(a, referenceHeight), 0, 0, 1, 1, frameOffX, frameOffY)
	b = transformContours(mirrorContoursY(b, referenceHeight), 0, 0, 1, 1, frameOffX, frameOffY)

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)
	agg2d.FillEvenOdd(cfg.FillRule == 0)

	fillA, lineA, fillB := sceneColors(cfg.Mode)
	drawContours(agg2d, a, fillA, lineA)
	drawContours(agg2d, b, fillB, agg.Transparent)

	if cfg.Operation == 0 {
		return
	}

	combineMS, renderMS, numSpans := combineAndRender(ctx.GetImage(), cfg, a, b)
	drawOverlay(agg2d, frameOffX, frameOffY, combineMS, renderMS, numSpans)
}

func combineAndRender(img *agg.Image, cfg Config, a, b []contour) (float64, float64, int) {
	ras1 := newRasterizer(cfg.FillRule)
	ras2 := newRasterizer(cfg.FillRule)
	ras1.AddPath(contoursToRasterPath(a), 0)
	ras2.AddPath(contoursToRasterPath(b), 0)

	switch cfg.ScanlineType {
	case 0:
		return combineAndRenderP8(img, ras1, ras2, cfg.Operation)
	case 1:
		return combineAndRenderU8(img, ras1, ras2, cfg.Operation)
	default:
		return combineAndRenderBin(img, ras1, ras2, cfg.Operation)
	}
}

func combineAndRenderP8(
	img *agg.Image,
	ras1, ras2 *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	op int,
) (float64, float64, int) {
	storage1 := isc.NewScanlineStorageAA[basics.Int8u]()
	storage2 := isc.NewScanlineStorageAA[basics.Int8u]()
	slRaster := isc.NewScanlineP8()
	renderRasterizerToAAStorageP8(ras1, slRaster, storage1)
	renderRasterizerToAAStorageP8(ras2, slRaster, storage2)

	sg1 := newAAStorageBoolRasterizer(storage1)
	sg2 := newAAStorageBoolRasterizer(storage2)
	sl1 := newBoolScanlineP8()
	sl2 := newBoolScanlineP8()
	slOut := newBoolScanlineP8()
	ren := &boolRenderer{}

	start := time.Now()
	for i := 0; i < 10; i++ {
		isc.CombineShapesAA(mapOperation(op), sg1, sg2, sl1, sl2, slOut, ren)
	}
	combineMS := float64(time.Since(start).Microseconds()) / 10000.0

	start = time.Now()
	numSpans := renderCollectedScanlines(img, ren.scanlines)
	renderMS := float64(time.Since(start).Microseconds()) / 1000.0
	return combineMS, renderMS, numSpans
}

func combineAndRenderU8(
	img *agg.Image,
	ras1, ras2 *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	op int,
) (float64, float64, int) {
	storage1 := isc.NewScanlineStorageAA[basics.Int8u]()
	storage2 := isc.NewScanlineStorageAA[basics.Int8u]()
	slRaster := isc.NewScanlineU8()
	renderRasterizerToAAStorageU8(ras1, slRaster, storage1)
	renderRasterizerToAAStorageU8(ras2, slRaster, storage2)

	sg1 := newAAStorageBoolRasterizer(storage1)
	sg2 := newAAStorageBoolRasterizer(storage2)
	sl1 := newBoolScanlineU8()
	sl2 := newBoolScanlineU8()
	slOut := newBoolScanlineU8()
	ren := &boolRenderer{}

	start := time.Now()
	for i := 0; i < 10; i++ {
		isc.CombineShapesAA(mapOperation(op), sg1, sg2, sl1, sl2, slOut, ren)
	}
	combineMS := float64(time.Since(start).Microseconds()) / 10000.0

	start = time.Now()
	numSpans := renderCollectedScanlines(img, ren.scanlines)
	renderMS := float64(time.Since(start).Microseconds()) / 1000.0
	return combineMS, renderMS, numSpans
}

func combineAndRenderBin(
	img *agg.Image,
	ras1, ras2 *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	op int,
) (float64, float64, int) {
	storage1 := isc.NewScanlineStorageBin()
	storage2 := isc.NewScanlineStorageBin()
	slRaster := isc.NewScanlineBin()
	renderRasterizerToBinStorage(ras1, slRaster, &rasterScanlineBinAdapter{sl: slRaster}, storage1)
	renderRasterizerToBinStorage(ras2, slRaster, &rasterScanlineBinAdapter{sl: slRaster}, storage2)

	sg1 := newBinStorageBoolRasterizer(storage1)
	sg2 := newBinStorageBoolRasterizer(storage2)
	sl1 := newBoolScanlineBin()
	sl2 := newBoolScanlineBin()
	slOut := newBoolScanlineBin()
	ren := &boolRenderer{}

	start := time.Now()
	for i := 0; i < 10; i++ {
		isc.CombineShapesBin(mapOperation(op), sg1, sg2, sl1, sl2, slOut, ren)
	}
	combineMS := float64(time.Since(start).Microseconds()) / 10000.0

	start = time.Now()
	numSpans := renderCollectedScanlines(img, ren.scanlines)
	renderMS := float64(time.Since(start).Microseconds()) / 1000.0
	return combineMS, renderMS, numSpans
}

func renderRasterizerToAAStorageU8(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *isc.ScanlineU8,
	storage *isc.ScanlineStorageAA[basics.Int8u],
) {
	storage.Prepare()
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	rasSL := &rasterScanlineAdapter{sl: sl}
	storageSL := &storageScanlineU8{sl: sl}
	for ras.SweepScanline(rasSL) {
		storage.Render(storageSL)
	}
}

func renderRasterizerToAAStorageP8(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *isc.ScanlineP8,
	storage *isc.ScanlineStorageAA[basics.Int8u],
) {
	storage.Prepare()
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	rasSL := &rasterScanlineP8Adapter{sl: sl}
	storageSL := &storageScanlineP8{sl: sl}
	for ras.SweepScanline(rasSL) {
		storage.Render(storageSL)
	}
}

func renderRasterizerToBinStorage(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *isc.ScanlineBin,
	rasSL rasterizer.ScanlineInterface,
	storage *isc.ScanlineStorageBin,
) {
	storage.Prepare()
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(rasSL) {
		storage.RenderBinScanline(sl)
	}
}

func renderCollectedScanlines(img *agg.Image, scanlines []boolRenderedScanline) int {
	numSpans := 0
	for _, scanline := range scanlines {
		numSpans += len(scanline.spans)
		for _, span := range scanline.spans {
			length := span.Len
			if length < 0 {
				length = -length
			}
			if length == 0 {
				continue
			}
			if len(span.Covers) == 1 {
				for i := 0; i < length; i++ {
					blendPixel(img, span.X+i, scanline.y, resultColor(), span.Covers[0])
				}
				continue
			}
			for i := 0; i < length && i < len(span.Covers); i++ {
				blendPixel(img, span.X+i, scanline.y, resultColor(), span.Covers[i])
			}
		}
	}
	return numSpans
}

func blendPixel(img *agg.Image, x, y int, c colorDef, cover uint8) {
	if x < 0 || y < 0 || x >= img.Width() || y >= img.Height() {
		return
	}
	alpha := int(math.Round(c.a * float64(cover)))
	if alpha <= 0 {
		return
	}
	if alpha > 255 {
		alpha = 255
	}

	stride := 0
	if img.Height() > 0 {
		stride = len(img.Data) / img.Height()
	}
	if stride <= 0 {
		return
	}

	idx := y*stride + x*4
	if idx+3 >= len(img.Data) {
		return
	}

	srcR := int(math.Round(c.r * 255.0))
	srcG := int(math.Round(c.g * 255.0))
	srcB := int(math.Round(c.b * 255.0))
	dstR := int(img.Data[idx])
	dstG := int(img.Data[idx+1])
	dstB := int(img.Data[idx+2])
	dstA := int(img.Data[idx+3])
	invA := 255 - alpha

	outA := alpha + (dstA*invA+127)/255
	outR := (srcR*alpha + dstR*invA + 127) / 255
	outG := (srcG*alpha + dstG*invA + 127) / 255
	outB := (srcB*alpha + dstB*invA + 127) / 255

	img.Data[idx] = uint8(clampInt(outR, 0, 255))
	img.Data[idx+1] = uint8(clampInt(outG, 0, 255))
	img.Data[idx+2] = uint8(clampInt(outB, 0, 255))
	img.Data[idx+3] = uint8(clampInt(outA, 0, 255))
}

func newRasterizer(fillRule int) *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip] {
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	if fillRule == 0 {
		ras.FillingRule(basics.FillEvenOdd)
	} else {
		ras.FillingRule(basics.FillNonZero)
	}
	return ras
}

func resultColor() colorDef {
	return colorDef{r: 0.5, g: 0.0, b: 0.0, a: 0.5}
}

func sceneColors(mode int) (agg.Color, agg.Color, agg.Color) {
	if mode == 2 || mode == 3 {
		return agg.RGBA(0.5, 0.5, 0.0, 0.1), agg.Black, agg.RGBA(0.0, 0.5, 0.5, 0.1)
	}
	return agg.RGBA(0.0, 0.0, 0.0, 0.1), agg.Transparent, agg.RGBA(0.0, 0.6, 0.0, 0.1)
}

func drawOverlay(a *agg.Agg2D, offX, offY, combineMS, renderMS float64, numSpans int) {
	a.FontGSV(8)
	a.FillColor(agg.Black)
	a.NoLine()
	a.Text(420+offX, 40+offY, fmt.Sprintf("Combine=%.3fms", combineMS), false, 0, 0)
	a.Text(420+offX, 58+offY, fmt.Sprintf("Render=%.3fms", renderMS), false, 0, 0)
	a.Text(420+offX, 76+offY, fmt.Sprintf("num_spans=%d", numSpans), false, 0, 0)
}

func mapOperation(op int) isc.BoolOp {
	switch op {
	case 1:
		return isc.BoolOr
	case 2:
		return isc.BoolAnd
	case 3:
		return isc.BoolXor
	case 4:
		return isc.BoolXorSaddle
	case 5:
		return isc.BoolAMinusB
	case 6:
		return isc.BoolBMinusA
	default:
		return isc.BoolAnd
	}
}

func buildShapes(cfg Config, w, h float64) ([]contour, []contour) {
	switch cfg.Mode {
	case 0:
		return modeSimple(cfg, w, h)
	case 1:
		return modeClosedStroke(cfg, w, h)
	case 2:
		return modeGBArrows(cfg, w, h)
	case 3:
		return modeGBSpiral(cfg, w, h)
	case 4:
		return modeSpiralGlyph(cfg)
	default:
		return modeSimple(cfg, w, h)
	}
}

func modeSimple(cfg Config, w, h float64) ([]contour, []contour) {
	dx := cfg.CenterX - w/2 + 100
	dy := cfg.CenterY - h/2 + 100
	a := []contour{
		{{dx + 140, dy + 145}, {dx + 225, dy + 44}, {dx + 296, dy + 219}, {dx + 226, dy + 289}, {dx + 82, dy + 292}},
		{{dx + 220, dy + 222}, {dx + 363, dy + 249}, {dx + 265, dy + 331}},
		{{dx + 242, dy + 243}, {dx + 325, dy + 261}, {dx + 268, dy + 309}},
		{{dx + 259, dy + 259}, {dx + 273, dy + 288}, {dx + 298, dy + 266}},
	}
	b := []contour{{{132, 177}, {573, 363}, {451, 390}, {454, 474}}}
	return a, b
}

func modeClosedStroke(cfg Config, w, h float64) ([]contour, []contour) {
	dx := cfg.CenterX - w/2 + 100
	dy := cfg.CenterY - h/2 + 100

	ps1 := path.NewPathStorageStl()
	ps1.MoveTo(dx+140, dy+145)
	ps1.LineTo(dx+225, dy+44)
	ps1.LineTo(dx+296, dy+219)
	ps1.ClosePolygon(0)
	ps1.LineTo(dx+226, dy+289)
	ps1.LineTo(dx+82, dy+292)

	ps1.MoveTo(dx+170, dy+222)
	ps1.LineTo(dx+313, dy+249)
	ps1.LineTo(dx+215, dy+331)
	ps1.ClosePolygon(0)

	ps2 := path.NewPathStorageStl()
	ps2.MoveTo(132, 177)
	ps2.LineTo(573, 363)
	ps2.LineTo(451, 390)
	ps2.LineTo(454, 474)
	ps2.ClosePolygon(0)

	stroke := conv.NewConvStroke(pathSource(ps2))
	stroke.SetWidth(15.0)
	return pathToContours(ps1), vertexSourceToContours(stroke)
}

func modeGBArrows(cfg Config, w, h float64) ([]contour, []contour) {
	psGB := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(psGB)
	psAr := path.NewPathStorageStl()
	aggshapes.MakeArrows(psAr)

	a := transformContours(pathToContours(psGB), -1150, -1150, 2.0, 2.0, 0, 0)
	tx := cfg.CenterX - w/2
	ty := cfg.CenterY - h/2
	b := transformContours(pathToContours(psAr), -1150, -1150, 2.0, 2.0, tx, ty)
	return a, b
}

func modeGBSpiral(cfg Config, w, h float64) ([]contour, []contour) {
	psGB := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(psGB)
	a := transformContours(pathToContours(psGB), -1150, -1150, 2.0, 2.0, 0, 0)

	spiralPath := buildSpiralPath(cfg.CenterX, cfg.CenterY, 10, 150, 30, 0.0)
	stroke := conv.NewConvStroke(pathSource(spiralPath))
	stroke.SetWidth(15.0)
	b := vertexSourceToContours(stroke)
	return a, b
}

func modeSpiralGlyph(cfg Config) ([]contour, []contour) {
	spiralPath := buildSpiralPath(cfg.CenterX, cfg.CenterY, 10, 150, 30, 0.0)
	stroke := conv.NewConvStroke(pathSource(spiralPath))
	stroke.SetWidth(15.0)
	a := vertexSourceToContours(stroke)

	glyph := buildGlyphPath()
	curve := conv.NewConvCurve(pathSource(glyph))
	b := transformContours(vertexSourceToContours(curve), 0, 0, 4.0, 4.0, 220, 200)
	return a, b
}

func buildSpiralPath(cx, cy, r1, r2, step, startAngle float64) *path.PathStorageStl {
	ps := path.NewPathStorageStl()
	angle := startAngle
	radius := r1
	deltaAngle := 4.0 * math.Pi / 180.0
	deltaRadius := step / 90.0
	first := true
	for radius <= r2 {
		x := cx + math.Cos(angle)*radius
		y := cy + math.Sin(angle)*radius
		if first {
			ps.MoveTo(x, y)
			first = false
		} else {
			ps.LineTo(x, y)
		}
		radius += deltaRadius
		angle += deltaAngle
	}
	return ps
}

func buildGlyphPath() *path.PathStorageStl {
	ps := path.NewPathStorageStl()
	ps.MoveTo(28.47, 6.45)
	ps.Curve3(21.58, 1.12, 19.82, 0.29)
	ps.Curve3(17.19, -0.93, 14.21, -0.93)
	ps.Curve3(9.57, -0.93, 6.57, 2.25)
	ps.Curve3(3.56, 5.42, 3.56, 10.60)
	ps.Curve3(3.56, 13.87, 5.03, 16.26)
	ps.Curve3(7.03, 19.58, 11.99, 22.51)
	ps.Curve3(16.94, 25.44, 28.47, 29.64)
	ps.LineTo(28.47, 31.40)
	ps.Curve3(28.47, 38.09, 26.34, 40.58)
	ps.Curve3(24.22, 43.07, 20.17, 43.07)
	ps.Curve3(17.09, 43.07, 15.28, 41.41)
	ps.Curve3(13.43, 39.75, 13.43, 37.60)
	ps.LineTo(13.53, 34.77)
	ps.Curve3(13.53, 32.52, 12.38, 31.30)
	ps.Curve3(11.23, 30.08, 9.38, 30.08)
	ps.Curve3(7.57, 30.08, 6.42, 31.35)
	ps.Curve3(5.27, 32.62, 5.27, 34.81)
	ps.Curve3(5.27, 39.01, 9.57, 42.53)
	ps.Curve3(13.87, 46.04, 21.63, 46.04)
	ps.Curve3(27.59, 46.04, 31.40, 44.04)
	ps.Curve3(34.28, 42.53, 35.64, 39.31)
	ps.Curve3(36.52, 37.21, 36.52, 30.71)
	ps.LineTo(36.52, 15.53)
	ps.Curve3(36.52, 9.13, 36.77, 7.69)
	ps.Curve3(37.01, 6.25, 37.57, 5.76)
	ps.Curve3(38.13, 5.27, 38.87, 5.27)
	ps.Curve3(39.65, 5.27, 40.23, 5.62)
	ps.Curve3(41.26, 6.25, 44.19, 9.18)
	ps.LineTo(44.19, 6.45)
	ps.Curve3(38.72, -0.88, 33.74, -0.88)
	ps.Curve3(31.35, -0.88, 29.93, 0.78)
	ps.Curve3(28.52, 2.44, 28.47, 6.45)
	ps.ClosePolygon(0)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(0)
	return ps
}

func pathSource(ps *path.PathStorageStl) interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
} {
	return &pathToConvSource{ps: ps}
}

type pathToConvSource struct{ ps *path.PathStorageStl }

func (a *pathToConvSource) Rewind(pathID uint) { a.ps.Rewind(pathID) }
func (a *pathToConvSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

func contoursToRasterPath(cs []contour) rasterVertexSource {
	ps := path.NewPathStorageStl()
	for _, c := range cs {
		if len(c) < 3 {
			continue
		}
		ps.MoveTo(c[0].x, c[0].y)
		for i := 1; i < len(c); i++ {
			ps.LineTo(c[i].x, c[i].y)
		}
		ps.ClosePolygon(0)
	}
	return &rasterPathAdapter{src: pathSource(ps)}
}

func mirrorContoursY(cs []contour, h float64) []contour {
	out := make([]contour, len(cs))
	for i := range cs {
		out[i] = make(contour, len(cs[i]))
		for j := range cs[i] {
			out[i][j] = pt{x: cs[i][j].x, y: h - cs[i][j].y}
		}
	}
	return out
}

func drawContours(a *agg.Agg2D, cs []contour, fill, line agg.Color) {
	for _, c := range cs {
		if len(c) < 3 {
			continue
		}
		a.ResetPath()
		a.MoveTo(c[0].x, c[0].y)
		for i := 1; i < len(c); i++ {
			a.LineTo(c[i].x, c[i].y)
		}
		a.ClosePolygon()
		a.FillColor(fill)
		if line.A > 0 {
			a.LineColor(line)
			a.LineWidth(1)
			a.DrawPath(agg.FillAndStroke)
		} else {
			a.NoLine()
			a.DrawPath(agg.FillOnly)
		}
	}
}

func vertexSourceToContours(vs interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}) []contour {
	var out []contour
	var cur contour
	vs.Rewind(0)
	for {
		x, y, cmd := vs.Vertex()
		switch {
		case basics.IsStop(cmd):
			if len(cur) >= 3 {
				out = append(out, closeIfNeeded(cur))
			}
			return out
		case basics.IsMoveTo(cmd):
			if len(cur) >= 3 {
				out = append(out, closeIfNeeded(cur))
			}
			cur = contour{{x: x, y: y}}
		case basics.IsVertex(cmd):
			cur = append(cur, pt{x: x, y: y})
		case basics.IsEndPoly(cmd):
			if len(cur) >= 3 {
				out = append(out, closeIfNeeded(cur))
			}
			cur = nil
		}
	}
}

func pathToContours(ps *path.PathStorageStl) []contour {
	return vertexSourceToContours(pathSource(ps))
}

func closeIfNeeded(c contour) contour {
	if len(c) < 2 {
		return c
	}
	first, last := c[0], c[len(c)-1]
	if math.Abs(first.x-last.x) > 1e-9 || math.Abs(first.y-last.y) > 1e-9 {
		c = append(c, first)
	}
	return c
}

func transformContours(cs []contour, tx1, ty1, sx, sy, tx2, ty2 float64) []contour {
	out := make([]contour, 0, len(cs))
	for _, c := range cs {
		cc := make(contour, 0, len(c))
		for _, p := range c {
			x := (p.x + tx1) * sx
			y := (p.y + ty1) * sy
			cc = append(cc, pt{x: x + tx2, y: y + ty2})
		}
		out = append(out, cc)
	}
	return out
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
