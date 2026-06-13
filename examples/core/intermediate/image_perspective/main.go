// Package main ports AGG's image_perspective.cpp demo.
//
// The original example is built on AGG_BGRA32 (linear color_type): the sRGB
// spheres image is decoded to linear on load, all blending happens in linear
// space, and the platform encodes linear->sRGB when saving. This port mirrors
// that with linear pixfmts plus EncodeLinearRGBToSRGB in the runner config.
package main

import (
	"fmt"
	"time"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	icol "github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	polygonctrl "github.com/cwbudde/agg_go/internal/ctrl/polygon"
	rboxctrl "github.com/cwbudde/agg_go/internal/ctrl/rbox"
	"github.com/cwbudde/agg_go/internal/demo/imageassets"
	"github.com/cwbudde/agg_go/internal/demo/timing"
	"github.com/cwbudde/agg_go/internal/gsv"
	imgacc "github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

const (
	frameWidth  = 600
	frameHeight = 600
)

// The C++ demo uses the default rasterizer_scanline_aa<> with its integer
// clipper (clip_box is set before the image pass).
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.IntConv, *rasterizer.RasterizerSlClip[int, rasterizer.IntConv]]

type renBase = renderer.RendererBase[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]]

type renBasePre = renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[icol.Linear], icol.RGBA8[icol.Linear]]

// ctrlVertexSource wraps a ctrl.Ctrl into the rasterizer interface.
type ctrlVertexSource struct{ ctrl ctrlbase.Ctrl[icol.RGBA] }

func (a *ctrlVertexSource) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlVertexSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type gsvStrokeVS struct{ s *conv.ConvStroke }

func (g *gsvStrokeVS) Rewind(id uint32) { g.s.Rewind(uint(id)) }
func (g *gsvStrokeVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := g.s.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ctrlColor converts a control's float color the way the C++ demo does:
// rgba -> rgba8 is a plain *255 quantization with no colorspace conversion.
func ctrlColor(c icol.RGBA) icol.RGBA8[icol.Linear] {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255.0 + 0.5)
		}
	}
	return icol.RGBA8[icol.Linear]{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

// renderCtrl is the Go equivalent of C++ agg::render_ctrl.
func renderCtrl(ras *rasType, sl *scanline.ScanlineU8, rb *renBase, c ctrlbase.Ctrl[icol.RGBA]) {
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSource{ctrl: c}, uint32(pathID))
		renscan.RenderScanlinesAASolid(ras, sl, rb, ctrlColor(c.Color(pathID)))
	}
}

func drawGSVText(ras *rasType, sl *scanline.ScanlineU8, rb *renBase, x, y float64, text string) {
	txt := gsv.NewGSVText()
	txt.SetSize(10.0, 0)
	txt.SetStartPoint(x, y)
	txt.SetText(text)

	// The C++ demo strokes gsv_text with a plain conv_stroke (butt caps,
	// miter joins), not gsv_text_outline (which forces round caps).
	stroke := conv.NewConvStroke(txt)
	stroke.SetWidth(1.5)

	ras.Reset()
	ras.AddPath(&gsvStrokeVS{s: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, icol.RGBA8[icol.Linear]{R: 0, G: 0, B: 0, A: 255})
}

// imagePixFmt exposes the source image buffer to the image accessors.
type imagePixFmt struct {
	rbuf *buffer.RenderingBufferU8
}

func (p imagePixFmt) Width() int    { return p.rbuf.Width() }
func (p imagePixFmt) Height() int   { return p.rbuf.Height() }
func (p imagePixFmt) PixWidth() int { return 4 }
func (p imagePixFmt) PixPtr(x, y int) []basics.Int8u {
	return buffer.RowU8(p.rbuf, y)[x*4:]
}

// imageCloneSource adapts image_accessor_clone to the span RGBA source interface.
type imageCloneSource struct {
	accessor *imgacc.ImageAccessorClone[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *imageCloneSource) Width() int                 { return s.ipf.Width() }
func (s *imageCloneSource) Height() int                { return s.ipf.Height() }
func (s *imageCloneSource) ColorType() string          { return "RGBA8" }
func (s *imageCloneSource) OrderType() icol.ColorOrder { return icol.OrderRGBA }
func (s *imageCloneSource) Span(x, y, l int) []basics.Int8u {
	return s.accessor.Span(x, y, l)
}
func (s *imageCloneSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *imageCloneSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *imageCloneSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

type spanImageGenerator interface {
	Generate(span []icol.RGBA8[icol.Linear], x, y int)
}

// renderImageSpans is the equivalent of C++ agg::render_scanlines_aa for the
// image pass: generated spans are blended through the premultiplied base.
func renderImageSpans(ras *rasType, sl *scanline.ScanlineU8, rbPre *renBasePre, sg spanImageGenerator) {
	alloc := span.NewSpanAllocator[icol.RGBA8[icol.Linear]]()
	if !ras.RewindScanlines() {
		return
	}
	// C++ render_scanlines_aa always calls prepare(); it is a no-op for the
	// plain filter generators here.
	if p, ok := sg.(interface{ Prepare() }); ok {
		p.Prepare()
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(sl) {
		y := sl.Y()
		for _, sp := range sl.Spans() {
			if sp.Len <= 0 {
				continue
			}
			colors := alloc.Allocate(int(sp.Len))
			sg.Generate(colors[:sp.Len], int(sp.X), y)
			rbPre.BlendColorHspan(int(sp.X), y, int(sp.Len), colors, sp.Covers, basics.CoverFull)
		}
	}
}

type demo struct {
	srcImg    *agg.Image
	quad      *polygonctrl.PolygonCtrl[icol.RGBA]
	transType *rboxctrl.RboxCtrl[icol.RGBA]
}

func newDemo() *demo {
	srcImg := loadSourceImage()

	// C++ on_init: a fixed quad 100 pixels in from each window edge.
	quad := polygonctrl.NewDefaultPolygonCtrl(4, 5.0)
	quad.SetXn(0, 100)
	quad.SetYn(0, 100)
	quad.SetXn(1, frameWidth-100)
	quad.SetYn(1, 100)
	quad.SetXn(2, frameWidth-100)
	quad.SetYn(2, frameHeight-100)
	quad.SetXn(3, 100)
	quad.SetYn(3, frameHeight-100)

	transType := rboxctrl.NewDefaultRboxCtrl(420, 5.0, 420+170.0, 70.0, false)
	transType.AddItem("Affine Parallelogram")
	transType.AddItem("Bilinear")
	transType.AddItem("Perspective")
	transType.SetCurItem(2)

	return &demo{
		srcImg:    srcImg,
		quad:      quad,
		transType: transType,
	}
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, w, h, img.Stride())

	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]](pf)
	pfPre := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	rbPre := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[icol.Linear], icol.RGBA8[icol.Linear]](pfPre)

	rb.Clear(icol.RGBA8[icol.Linear]{R: 255, G: 255, B: 255, A: 255})

	if d.transType.CurItem() == 0 {
		// For the affine parallelogram transformations we calculate the
		// 4-th (implicit) point of the parallelogram.
		d.quad.SetXn(3, d.quad.Xn(0)+(d.quad.Xn(2)-d.quad.Xn(1)))
		d.quad.SetYn(3, d.quad.Yn(0)+(d.quad.Yn(2)-d.quad.Yn(1)))
	}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.IntConv, *rasterizer.RasterizerSlClip[int, rasterizer.IntConv]](
		rasterizer.IntConv{}, rasterizer.NewRasterizerSlClip[int, rasterizer.IntConv](rasterizer.IntConv{}),
	)
	sl := scanline.NewScanlineU8()

	// Render the "quad" tool (outline + handle circles).
	ras.Reset()
	ras.AddPath(&ctrlVertexSource{ctrl: d.quad}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, ctrlColor(icol.NewRGBA(0, 0.3, 0.5, 0.6)))

	// Prepare the polygon to rasterize: the destination (transformed) quad.
	ras.ClipBox(0, 0, float64(w), float64(h))
	ras.Reset()
	ras.MoveToD(d.quad.Xn(0), d.quad.Yn(0))
	ras.LineToD(d.quad.Xn(1), d.quad.Yn(1))
	ras.LineToD(d.quad.Xn(2), d.quad.Yn(2))
	ras.LineToD(d.quad.Xn(3), d.quad.Yn(3))

	// C++: agg::image_filter_lut filter(filter_kernel, false) - NOT normalized.
	filter := imgacc.NewImageFilterLUTWithFilter(imgacc.BilinearFilter{}, false)

	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(d.srcImg.Data, d.srcImg.Width(), d.srcImg.Height(), d.srcImg.Stride())
	ipf := imagePixFmt{rbuf: imgRbuf}
	source := &imageCloneSource{accessor: imgacc.NewImageAccessorClone(&ipf), ipf: &ipf}

	srcW := float64(d.srcImg.Width())
	srcH := float64(d.srcImg.Height())
	parl := [6]float64{
		d.quad.Xn(0), d.quad.Yn(0),
		d.quad.Xn(1), d.quad.Yn(1),
		d.quad.Xn(2), d.quad.Yn(2),
	}
	q8 := [8]float64{
		d.quad.Xn(0), d.quad.Yn(0),
		d.quad.Xn(1), d.quad.Yn(1),
		d.quad.Xn(2), d.quad.Yn(2),
		d.quad.Xn(3), d.quad.Yn(3),
	}

	startTime := time.Now()
	switch d.transType.CurItem() {
	case 0:
		tr := transform.NewTransAffineParlToRect(parl, 0, 0, srcW, srcH)
		interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](tr, 8)
		sg := span.NewSpanImageFilterRGBANNWithParams[*imageCloneSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source, interp,
		)
		renderImageSpans(ras, sl, rbPre, sg)

	case 1:
		tr := transform.NewTransBilinearQuadToRect(q8, 0, 0, srcW, srcH)
		if tr.IsValid() {
			interp := span.NewSpanInterpolatorLinear[*transform.TransBilinear](tr, 8)
			sg := span.NewSpanImageFilterRGBA2x2WithParams[*imageCloneSource, *span.SpanInterpolatorLinear[*transform.TransBilinear]](
				source, interp, filter,
			)
			renderImageSpans(ras, sl, rbPre, sg)
		}

	case 2:
		tr := transform.NewTransPerspectiveQuadToRect(q8, 0, 0, srcW, srcH)
		if tr.IsValid(transform.AffineEpsilon) {
			interp := span.NewSpanInterpolatorTrans[*transform.TransPerspective](tr)
			sg := span.NewSpanImageFilterRGBA2x2WithParams[*imageCloneSource, *span.SpanInterpolatorTrans[*transform.TransPerspective]](
				source, interp, filter,
			)
			renderImageSpans(ras, sl, rbPre, sg)
		}
	}
	elapsedMs := float64(time.Since(startTime)) / float64(time.Millisecond)

	if timing.ShowText() {
		drawGSVText(ras, sl, rb, 10.0, 10.0, fmt.Sprintf("%3.2f ms", elapsedMs))
	}

	renderCtrl(ras, sl, rb, d.transType)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	if d.transType.OnMouseButtonDown(fx, fy) {
		return true
	}
	return d.quad.OnMouseButtonDown(fx, fy)
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := d.transType.OnMouseMove(fx, fy, btn.Left)
	if d.quad.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := d.transType.OnMouseButtonUp(fx, fy)
	if d.quad.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	return redraw
}

// loadSourceImage returns the spheres art image decoded to linear RGBA.
// The C++ platform loads the sRGB PPM into the linear window format, and the
// bundled PPM decodes top-down while the demo frame is bottom-up, so flip it.
func loadSourceImage() *agg.Image {
	img, err := imageassets.Spheres()
	if err != nil {
		panic("image_perspective: failed to load spheres image: " + err.Error())
	}
	w, h := img.Width(), img.Height()
	for p := 0; p < w*h; p++ {
		c := icol.ConvertRGB8SRGBToLinear(icol.RGB8[icol.SRGB]{
			R: img.Data[p*4+0],
			G: img.Data[p*4+1],
			B: img.Data[p*4+2],
		})
		img.Data[p*4+0] = c.R
		img.Data[p*4+1] = c.G
		img.Data[p*4+2] = c.B
	}
	flipImageY(img)
	return img
}

func flipImageY(img *agg.Image) {
	w, h := img.Width(), img.Height()
	if w == 0 || h == 0 {
		return
	}
	stride := w * 4
	row := make([]byte, stride)
	for y := 0; y < h/2; y++ {
		top := y * stride
		bottom := (h - 1 - y) * stride
		copy(row, img.Data[top:top+stride])
		copy(img.Data[top:top+stride], img.Data[bottom:bottom+stride])
		copy(img.Data[bottom:bottom+stride], row)
	}
}

func runnerConfig() lowlevelrunner.Config {
	return lowlevelrunner.Config{
		Title:                 "Image Perspective",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}
}

func main() {
	lowlevelrunner.Run(runnerConfig(), newDemo())
}
