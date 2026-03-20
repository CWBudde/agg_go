package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icolor "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlpkg "github.com/MeKo-Christian/agg_go/internal/ctrl"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	ctrlpoly "github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	isl "github.com/MeKo-Christian/agg_go/internal/scanline"
)

const (
	frameWidth  = 600
	frameHeight = 600
)

type vertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type rasterVertexSourceAdapter struct {
	src vertexSource
}

func (a *rasterVertexSourceAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *rasterVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type demo struct {
	poly   *ctrlpoly.PolygonCtrl[icolor.RGBA]
	close  *checkbox.CheckboxCtrl[icolor.RGBA]
	points *slider.SliderCtrl
	flip   bool
}

func newDemo() *demo {
	d := &demo{
		poly:   ctrlpoly.NewPolygonCtrl[icolor.RGBA](6, 5.0, icolor.NewRGBA(0, 0.3, 0.5, 0.6)),
		close:  checkbox.NewDefaultCheckboxCtrl(350, 5, "Close", false),
		points: slider.NewSliderCtrl(5, 5, 340, 12, false),
	}

	d.poly.SetLineWidth(1.0)
	d.poly.SetClose(true)
	d.poly.SetInPolygonCheck(true)
	d.close.SetChecked(false)
	d.points.SetRange(1.0, 40.0)
	d.points.SetValue(20.0)
	d.points.SetLabel("Number of intermediate Points = %.3f")

	d.initPolygon()
	return d
}

func (d *demo) initPolygon() {
	if d.flip {
		d.poly.SetXn(0, 100)
		d.poly.SetYn(0, frameHeight-100)
		d.poly.SetXn(1, frameWidth-100)
		d.poly.SetYn(1, frameHeight-100)
		d.poly.SetXn(2, frameWidth-100)
		d.poly.SetYn(2, 100)
		d.poly.SetXn(3, 100)
		d.poly.SetYn(3, 100)
	} else {
		d.poly.SetXn(0, 100)
		d.poly.SetYn(0, 100)
		d.poly.SetXn(1, frameWidth-100)
		d.poly.SetYn(1, 100)
		d.poly.SetXn(2, frameWidth-100)
		d.poly.SetYn(2, frameHeight-100)
		d.poly.SetXn(3, 100)
		d.poly.SetYn(3, frameHeight-100)
	}

	d.poly.SetXn(4, frameWidth/2)
	d.poly.SetYn(4, frameHeight/2)
	d.poly.SetXn(5, frameWidth/2)
	d.poly.SetYn(5, frameHeight/3)
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer + y-flip (flip_y=true in C++).
	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32Linear(workRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icolor.Linear], icolor.RGBA8[icolor.Linear]](pf)
	renBase.Clear(icolor.RGBA8[icolor.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := isl.NewScanlineP8()
	rasWrap := &rasterizerAdapter{ras: ras}
	slWrap := &scanlineWrapperP8{sl: sl}

	renderCurve(ras, rasWrap, slWrap, renBase, d.poly, d.close, d.points)
	renderControl(ras, rasWrap, slWrap, renBase, d.poly)
	renderControl(ras, rasWrap, slWrap, renBase, d.close)
	renderControl(ras, rasWrap, slWrap, renBase, d.points)

	copyFlipY(workBuf, img.Data, w, h)
}

func renderCurve(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	rasWrap *rasterizerAdapter,
	slWrap *scanlineWrapperP8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[icolor.Linear], icolor.RGBA8[icolor.Linear]],
	poly *ctrlpoly.PolygonCtrl[icolor.RGBA],
	close *checkbox.CheckboxCtrl[icolor.RGBA],
	points *slider.SliderCtrl,
) {
	src := ctrlpoly.NewSimplePolygonVertexSource(poly.PolygonData(), poly.NumPoints(), false, close.IsChecked())
	bspline := conv.NewConvBSpline(src)
	bspline.SetInterpolationStep(1.0 / points.Value())

	stroke := conv.NewConvStroke(bspline)
	stroke.SetWidth(2.0)

	ras.Reset()
	ras.AddPath(&rasterVertexSourceAdapter{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(
		rasWrap,
		slWrap,
		renBase,
		icolor.RGBA8[icolor.Linear]{R: 0, G: 0, B: 0, A: 255},
	)
}

func renderControl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	rasWrap *rasterizerAdapter,
	slWrap *scanlineWrapperP8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[icolor.Linear], icolor.RGBA8[icolor.Linear]],
	ctrl ctrlpkg.Ctrl[icolor.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlPathAdapter{ctrl: ctrl}, uint32(pathID))
		renscan.RenderScanlinesAASolid(
			rasWrap,
			slWrap,
			renBase,
			toRGBA8(ctrl.Color(pathID)),
		)
	}
}

type ctrlPathAdapter struct {
	ctrl ctrlpkg.Ctrl[icolor.RGBA]
}

func (a *ctrlPathAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type scanlineWrapperP8 struct {
	sl   *isl.ScanlineP8
	iter spanIterP8
}

func (w *scanlineWrapperP8) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapperP8) Y() int               { return w.sl.Y() }
func (w *scanlineWrapperP8) NumSpans() int        { return w.sl.NumSpans() }

func (w *scanlineWrapperP8) Begin() renscan.ScanlineIterator {
	w.iter.spans = w.sl.Spans()
	w.iter.idx = 0
	return &w.iter
}

type spanIterP8 struct {
	spans []isl.SpanP8
	idx   int
}

func (it *spanIterP8) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	covers := make([]uint8, len(s.Covers))
	for i := range s.Covers {
		covers[i] = uint8(s.Covers[i])
	}
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: covers}
}

func (it *spanIterP8) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

type rasterizerAdapter struct {
	ras interface {
		RewindScanlines() bool
		SweepScanline(sl rasterizer.ScanlineInterface) bool
		MinX() int
		MaxX() int
	}
}

func (r *rasterizerAdapter) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *rasterizerAdapter) MinX() int             { return r.ras.MinX() }
func (r *rasterizerAdapter) MaxX() int             { return r.ras.MaxX() }

func (r *rasterizerAdapter) SweepScanline(sl renscan.ScanlineInterface) bool {
	w, ok := sl.(*scanlineWrapperP8)
	if !ok {
		return false
	}
	return r.ras.SweepScanline(&rasScanlineAdapter{sl: w.sl})
}

type rasScanlineAdapter struct {
	sl *isl.ScanlineP8
}

func (a *rasScanlineAdapter) ResetSpans()             { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, c uint32) { a.sl.AddCell(x, uint(c)) }
func (a *rasScanlineAdapter) AddSpan(x, l int, c uint32) {
	a.sl.AddSpan(x, l, uint(c))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

func toRGBA8(c icolor.RGBA) icolor.RGBA8[icolor.Linear] {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255 + 0.5)
		}
	}

	return icolor.RGBA8[icolor.Linear]{
		R: clamp(c.R),
		G: clamp(c.G),
		B: clamp(c.B),
		A: clamp(c.A),
	}
}

func copyFlipY(src, dst []uint8, width, height int) {
	stride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}

	changed := d.close.OnMouseButtonDown(float64(x), float64(y))
	if d.points.OnMouseButtonDown(float64(x), float64(y)) {
		changed = true
	}
	if d.poly.OnMouseButtonDown(float64(x), float64(y)) {
		changed = true
	}
	return changed
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	changed := d.close.OnMouseMove(float64(x), float64(y), btn.Left)
	if d.points.OnMouseMove(float64(x), float64(y), btn.Left) {
		changed = true
	}

	if btn.Left {
		if d.poly.OnMouseMove(float64(x), float64(y), true) {
			changed = true
		}
	} else if d.poly.OnMouseButtonUp(float64(x), float64(y)) {
		changed = true
	}

	return changed
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	changed := d.close.OnMouseButtonUp(float64(x), float64(y))
	if d.points.OnMouseButtonUp(float64(x), float64(y)) {
		changed = true
	}
	if d.poly.OnMouseButtonUp(float64(x), float64(y)) {
		changed = true
	}
	return changed
}

func (d *demo) OnKey(key rune) bool {
	if key != ' ' {
		return false
	}

	d.flip = !d.flip
	d.initPolygon()
	return true
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AGG Example. BSpline Interpolator",
		Width:  frameWidth,
		Height: frameHeight,
	}, newDemo())
}
