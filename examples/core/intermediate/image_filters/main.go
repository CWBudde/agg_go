// Port of AGG C++ image_filters.cpp.
//
// The standalone example follows the original layout:
// - transformed source image on the right
// - filter controls on the left
// - NSteps status text at the bottom
//
// The C++ demo runs with flip_y=true, so the standalone port mirrors the final
// frame before saving to match the original window orientation.
package main

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	"github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	"github.com/cwbudde/agg_go/internal/ctrl/rbox"
	"github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/demo/timing"
	"github.com/cwbudde/agg_go/internal/gsv"
	imgacc "github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

type imagePixFmt struct {
	rbuf *buffer.RenderingBufferU8
}

func (p imagePixFmt) Width() int    { return p.rbuf.Width() }
func (p imagePixFmt) Height() int   { return p.rbuf.Height() }
func (p imagePixFmt) PixWidth() int { return 4 }
func (p imagePixFmt) PixPtr(x, y int) []basics.Int8u {
	row := buffer.RowU8(p.rbuf, y)
	return row[x*4:]
}

type imageClipSource struct {
	accessor *imgacc.ImageAccessorClip[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *imageClipSource) Width() int                  { return s.ipf.Width() }
func (s *imageClipSource) Height() int                 { return s.ipf.Height() }
func (s *imageClipSource) ColorType() string           { return "RGBA8" }
func (s *imageClipSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *imageClipSource) Span(x, y, l int) []basics.Int8u {
	return s.accessor.Span(x, y, l)
}
func (s *imageClipSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *imageClipSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *imageClipSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

type spanImageGenerator interface {
	Generate(span []color.RGBA8[color.Linear], x, y int)
}

type spanGenAdapter struct {
	gen spanImageGenerator
}

func (a *spanGenAdapter) Prepare() {}

func (a *spanGenAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	if length <= 0 {
		return
	}
	a.gen.Generate(colors[:length], x, y)
}

type pathSourceAdapter struct{ ps *path.PathStorageStl }

func (a *pathSourceAdapter) Rewind(id uint32) { a.ps.Rewind(uint(id)) }
func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

type filterState struct {
	filterIdx  int
	radius     float64
	normalize  bool
	step       float64
	run        bool
	singleStep bool
	refresh    bool
	curAngle   float64
	curFilter  int
	numSteps   int
	numPix     float64
	time1      float64
	time2      float64
}

func defaultState() filterState {
	return filterState{
		filterIdx: 1,
		radius:    4.0,
		normalize: true,
		step:      5.0,
		curFilter: 1,
	}
}

func (s *filterState) clamp() {
	if s.radius < 2.0 {
		s.radius = 2.0
	}
	if s.radius > 8.0 {
		s.radius = 8.0
	}
	if s.filterIdx < 0 {
		s.filterIdx = 0
	}
	if s.filterIdx > 16 {
		s.filterIdx = 16
	}
	if s.step < 1.0 {
		s.step = 1.0
	}
	if s.step > 10.0 {
		s.step = 10.0
	}
}

type demo struct {
	srcImg *agg.Image
	state  filterState
}

func newDemo(srcImg *agg.Image) *demo {
	return &demo{srcImg: srcImg, state: defaultState()}
}

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

type renBaseType = renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]

func (d *demo) Render(img *agg.Image) {
	if d.srcImg == nil {
		return
	}
	d.state.clamp()

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)

	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	work := agg.CreateImage(d.srcImg.Width(), d.srcImg.Height())
	renderTransformedImage(work, d.srcImg, d.state)

	workRbuf := buffer.NewRenderingBufferU8WithData(work.Data, work.Width(), work.Height(), work.Width()*4)
	workPf := pixfmt.NewPixFmtRGBA32Linear(workRbuf)
	rb.CopyFrom(workPf, nil, 110, 35)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	drawStatusText(ras, sl, rb, &d.state)
	drawControls(ras, sl, rb, &d.state)
}

func renderTransformedImage(dst, src *agg.Image, st filterState) {
	dstRbuf := buffer.NewRenderingBufferWithData[uint8](dst.Data, dst.Width(), dst.Height(), dst.Width()*4)

	// C++ transform_image clears through the plain renderer base before
	// blending the filtered spans through the premultiplied one.
	clearPixf := pixfmt.NewPixFmtRGBA32Linear(dstRbuf)
	clearBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](clearPixf)
	clearBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	dstPixf := pixfmt.NewPixFmtRGBA32Pre[color.Linear](dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// Match the C++ example: rotate the source around its own center.
	width := float64(src.Width())
	height := float64(src.Height())
	srcMtx := transform.NewTransAffine()
	srcMtx.Translate(-width/2.0, -height/2.0)
	srcMtx.Rotate(st.curAngle * math.Pi / 180.0)
	srcMtx.Translate(width/2.0, height/2.0)

	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-width/2.0, -height/2.0)
	imgMtx.Rotate(st.curAngle * math.Pi / 180.0)
	imgMtx.Translate(width/2.0, height/2.0)
	imgMtx.Invert()

	// Build the ellipse path in source space, then transform it.
	r := width
	if height < r {
		r = height
	}
	r = r*0.5 - 4.0
	clipPath := buildEllipsePath(width*0.5, height*0.5, r, srcMtx)

	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(src.Data, src.Width(), src.Height(), src.Width()*4)
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := imgacc.NewImageAccessorClip(&ipf, []basics.Int8u{0, 0, 0, 0})
	source := &imageClipSource{accessor: accessor, ipf: &ipf}

	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
	sg := buildSpanGenerator(source, interp, st)
	sgAdp := &spanGenAdapter{gen: sg}

	ras.Reset()
	ras.ClipBox(0, 0, float64(dst.Width()), float64(dst.Height()))
	ras.AddPath(&pathSourceAdapter{ps: clipPath}, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(sl) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len <= 0 {
					continue
				}
				colors := alloc.Allocate(int(spanData.Len))
				sgAdp.Generate(colors, int(spanData.X), y, int(spanData.Len))
				renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
			}
		}
	}
}

func buildEllipsePath(cx, cy, r float64, mtx *transform.TransAffine) *path.PathStorageStl {
	ps := path.NewPathStorageStl()
	const steps = 200
	for i := 0; i <= steps; i++ {
		a := 2 * math.Pi * float64(i) / float64(steps)
		x := cx + r*math.Cos(a)
		y := cy + r*math.Sin(a)
		if mtx != nil {
			mtx.Transform(&x, &y)
		}
		if i == 0 {
			ps.MoveTo(x, y)
		} else {
			ps.LineTo(x, y)
		}
	}
	ps.ClosePolygon(basics.PathFlagsNone)
	return ps
}

func buildSpanGenerator(
	source *imageClipSource,
	interp *span.SpanInterpolatorLinear[*transform.TransAffine],
	st filterState,
) spanImageGenerator {
	switch st.filterIdx {
	case 0:
		return span.NewSpanImageFilterRGBANNWithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interp)
	case 1:
		return span.NewSpanImageFilterRGBABilinearClipWithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			color.RGBA8[color.Linear]{},
			interp,
		)
	case 5:
		return span.NewSpanImageFilterRGBA2x2WithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			interp,
			imgacc.NewImageFilterLUTWithFilter(imgacc.HanningFilter{}, st.normalize),
		)
	case 6:
		return span.NewSpanImageFilterRGBA2x2WithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			interp,
			imgacc.NewImageFilterLUTWithFilter(imgacc.HammingFilter{}, st.normalize),
		)
	case 7:
		return span.NewSpanImageFilterRGBA2x2WithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			interp,
			imgacc.NewImageFilterLUTWithFilter(imgacc.HermiteFilter{}, st.normalize),
		)
	default:
		return span.NewSpanImageFilterRGBAWithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			interp,
			imgacc.NewImageFilterLUTWithFilter(newFilter(st.filterIdx, st.radius), st.normalize),
		)
	}
}

func drawGSVText(ras *rasType, sl *scanline.ScanlineU8, rb *renBaseType, x, y float64, text string) {
	txt := gsv.NewGSVText()
	txt.SetStartPoint(x, y)
	txt.SetSize(10.0, 0)
	txt.SetText(text)

	// The C++ demo strokes gsv_text with a plain conv_stroke (butt caps,
	// miter joins), not gsv_text_outline (which forces round caps).
	stroke := conv.NewConvStroke(txt)
	stroke.SetWidth(1.5)

	ras.Reset()
	ras.AddPath(&gsvStrokeVS{s: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})
}

func drawStatusText(ras *rasType, sl *scanline.ScanlineU8, rb *renBaseType, st *filterState) {
	drawGSVText(ras, sl, rb, 10.0, 295.0, "NSteps="+itoa(st.numSteps))

	if timing.ShowText() && st.time1 != st.time2 && st.numPix > 0 {
		kpix := st.numPix / (st.time2 - st.time1)
		drawGSVText(ras, sl, rb, 10.0, 310.0, ftoa2(kpix)+" Kpix/sec")
	}
}

func drawControls(ras *rasType, sl *scanline.ScanlineU8, rb *renBaseType, st *filterState) {
	step := slider.NewSliderCtrl(115, 5, 400, 11, false)
	step.SetLabel("Step=%3.2f")
	step.SetRange(1.0, 10.0)
	step.SetValue(st.step)

	radius := slider.NewSliderCtrl(115, 20, 400, 26, false)
	radius.SetLabel("Filter Radius=%.3f")
	radius.SetRange(2.0, 8.0)
	radius.SetValue(st.radius)

	filters := rbox.NewDefaultRboxCtrl(0, 0, 110, 210, false)
	filters.SetBorderWidth(0, 0)
	filters.SetBackgroundColor(color.NewRGBA(0.0, 0.0, 0.0, 0.1))
	filters.SetTextSize(6.0, 0)
	filters.SetTextThickness(0.85)
	filters.AddItem("simple (NN)")
	filters.AddItem("bilinear")
	filters.AddItem("bicubic")
	filters.AddItem("spline16")
	filters.AddItem("spline36")
	filters.AddItem("hanning")
	filters.AddItem("hamming")
	filters.AddItem("hermite")
	filters.AddItem("kaiser")
	filters.AddItem("quadric")
	filters.AddItem("catrom")
	filters.AddItem("gaussian")
	filters.AddItem("bessel")
	filters.AddItem("mitchell")
	filters.AddItem("sinc")
	filters.AddItem("lanczos")
	filters.AddItem("blackman")
	filters.SetCurItem(st.filterIdx)

	run := checkbox.NewDefaultCheckboxCtrl(8, 245, "RUN Test!", false)
	run.SetTextSize(7.5, 0)
	run.SetChecked(st.run)

	singleStep := checkbox.NewDefaultCheckboxCtrl(8, 230, "Single Step", false)
	singleStep.SetTextSize(7.5, 0)
	singleStep.SetChecked(st.singleStep)

	normalize := checkbox.NewDefaultCheckboxCtrl(8, 215, "Normalize Filter", false)
	normalize.SetTextSize(7.5, 0)
	normalize.SetChecked(st.normalize)

	refresh := checkbox.NewDefaultCheckboxCtrl(8, 265, "Refresh", false)
	refresh.SetTextSize(7.5, 0)
	refresh.SetChecked(st.refresh)

	if st.filterIdx >= 14 {
		renderCtrl(ras, sl, rb, radius)
	}
	renderCtrl(ras, sl, rb, step)
	renderCtrl(ras, sl, rb, filters)
	renderCtrl(ras, sl, rb, run)
	renderCtrl(ras, sl, rb, normalize)
	renderCtrl(ras, sl, rb, singleStep)
	renderCtrl(ras, sl, rb, refresh)
}

// ctrlColor converts a control's float color like the C++ render_ctrl with a
// linear color_type: a plain *255 quantization, no colorspace conversion.
func ctrlColor(c color.RGBA) color.RGBA8[color.Linear] {
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
	return color.RGBA8[color.Linear]{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

// renderCtrl is the Go equivalent of C++ agg::render_ctrl.
func renderCtrl(ras *rasType, sl *scanline.ScanlineU8, rb *renBaseType, c ctrlbase.Ctrl[color.RGBA]) {
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSource{ctrl: c}, uint32(pathID))
		renscan.RenderScanlinesAASolid(ras, sl, rb, ctrlColor(c.Color(pathID)))
	}
}

type gsvStrokeVS struct{ s *conv.ConvStroke }

func (g *gsvStrokeVS) Rewind(pathID uint32) { g.s.Rewind(uint(pathID)) }
func (g *gsvStrokeVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := g.s.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type ctrlVertexSource struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlVertexSource) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlVertexSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type absFilter struct {
	base imgacc.FilterFunction
}

func (f absFilter) Radius() float64 {
	return f.base.Radius()
}

func (f absFilter) CalcWeight(x float64) float64 {
	return f.base.CalcWeight(math.Abs(x))
}

func newFilter(idx int, radius float64) imgacc.FilterFunction {
	switch idx {
	case 1:
		return absFilter{base: imgacc.BilinearFilter{}}
	case 2:
		return absFilter{base: imgacc.BicubicFilter{}}
	case 3:
		return absFilter{base: imgacc.Spline16Filter{}}
	case 4:
		return absFilter{base: imgacc.Spline36Filter{}}
	case 5:
		return absFilter{base: imgacc.HanningFilter{}}
	case 6:
		return absFilter{base: imgacc.HammingFilter{}}
	case 7:
		return absFilter{base: imgacc.HermiteFilter{}}
	case 8:
		return absFilter{base: imgacc.NewKaiserFilter(0)}
	case 9:
		return absFilter{base: imgacc.QuadricFilter{}}
	case 10:
		return absFilter{base: imgacc.CatromFilter{}}
	case 11:
		return absFilter{base: imgacc.GaussianFilter{}}
	case 12:
		return absFilter{base: imgacc.BesselFilter{}}
	case 13:
		return absFilter{base: imgacc.NewMitchellFilter(1.0/3.0, 1.0/3.0)}
	case 14:
		return absFilter{base: imgacc.NewSincFilter(radius)}
	case 15:
		return absFilter{base: imgacc.NewLanczosFilter(radius)}
	case 16:
		return absFilter{base: imgacc.NewBlackmanFilter(radius)}
	default:
		return absFilter{base: imgacc.BilinearFilter{}}
	}
}

func flipImageY(img *agg.Image) {
	if img == nil {
		return
	}
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

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func ftoa2(v float64) string {
	if v < 0 {
		return "-" + ftoa2(-v)
	}
	whole := int(v)
	frac := int(math.Round((v - float64(whole)) * 100.0))
	if frac == 100 {
		whole++
		frac = 0
	}
	return itoa(whole) + "." + twoDigits(frac)
}

func twoDigits(v int) string {
	if v < 10 {
		return "0" + itoa(v)
	}
	return itoa(v)
}

func loadSourceImage() *agg.Image {
	paths := []string{
		filepath.Join("examples", "shared", "art", "spheres.ppm"),
		filepath.Join("..", "..", "..", "..", "examples", "shared", "art", "spheres.ppm"),
	}
	for _, p := range paths {
		img, err := loadPPMImage(p)
		if err == nil {
			// The AGG demo uses a BMP source that is stored bottom-up.
			// The bundled PPM copy decodes top-down, so flip it once here to
			// match the original image orientation.
			flipImageY(img)
			return img
		}
	}
	panic("image_filters: failed to load examples/shared/art/spheres.ppm")
}

func loadPPMImage(filename string) (*agg.Image, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(data) < 2 || data[0] != 'P' || data[1] != '6' {
		return nil, errors.New("unsupported ppm format: expected P6")
	}

	i := 2
	readToken := func() (string, error) {
		for i < len(data) {
			b := data[i]
			if b == '#' {
				for i < len(data) && data[i] != '\n' {
					i++
				}
				continue
			}
			if bytes.IndexByte([]byte{' ', '\t', '\n', '\r'}, b) >= 0 {
				i++
				continue
			}
			break
		}
		if i >= len(data) {
			return "", errors.New("unexpected end of ppm header")
		}
		start := i
		for i < len(data) {
			b := data[i]
			if bytes.IndexByte([]byte{' ', '\t', '\n', '\r', '#'}, b) >= 0 {
				break
			}
			i++
		}
		return string(data[start:i]), nil
	}

	wTok, err := readToken()
	if err != nil {
		return nil, err
	}
	hTok, err := readToken()
	if err != nil {
		return nil, err
	}
	maxTok, err := readToken()
	if err != nil {
		return nil, err
	}
	w, err := strconv.Atoi(wTok)
	if err != nil || w <= 0 {
		return nil, errors.New("invalid ppm width")
	}
	h, err := strconv.Atoi(hTok)
	if err != nil || h <= 0 {
		return nil, errors.New("invalid ppm height")
	}
	maxV, err := strconv.Atoi(maxTok)
	if err != nil || maxV != 255 {
		return nil, errors.New("unsupported ppm max value")
	}

	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	rgb := data[i:]
	if len(rgb) < w*h*3 {
		return nil, errors.New("ppm pixel data too short")
	}

	// The C++ platform loads the sRGB PPM into the linear window format
	// (convert<pixfmt_bgr24, pixfmt_srgb24>), so decode each pixel here.
	buf := make([]uint8, w*h*4)
	for p := 0; p < w*h; p++ {
		c := color.ConvertRGB8SRGBToLinear(color.RGB8[color.SRGB]{
			R: rgb[p*3+0],
			G: rgb[p*3+1],
			B: rgb[p*3+2],
		})
		buf[p*4+0] = c.R
		buf[p*4+1] = c.G
		buf[p*4+2] = c.B
		buf[p*4+3] = 255
	}

	return agg.NewImage(buf, w, h, w*4), nil
}

func main() {
	srcImg := loadSourceImage()
	w := srcImg.Width() + 110
	h := srcImg.Height() + 40
	if w < 305 {
		w = 305
	}
	if h < 325 {
		h = 325
	}

	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Image Filters",
		Width:                 w,
		Height:                h,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo(srcImg))
}
