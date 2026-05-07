// Port of AGG C++ image_transforms.cpp.
//
// The standalone version now matches the original demo more closely:
// - the rendered frame uses flip_y=true
// - the seven original controls are present
// - the image center and polygon can be dragged with the mouse
// - the rotate checkboxes advance the angles during idle time
package main

import (
	"math"

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
	"github.com/cwbudde/agg_go/internal/demo/imageassets"
	"github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
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
	accessor *image.ImageAccessorClip[imagePixFmt]
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

type ctrlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }
func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type ellipseVertexSourceAdapter struct {
	ell *shapes.Ellipse
}

func (a *ellipseVertexSourceAdapter) Rewind(pathID uint32) {
	a.ell.Rewind(pathID)
}

func (a *ellipseVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	return uint32(a.ell.Vertex(x, y))
}

type ellipseConvAdapter struct {
	ell *shapes.Ellipse
}

func (a *ellipseConvAdapter) Rewind(pathID uint) {
	a.ell.Rewind(uint32(pathID))
}

func (a *ellipseConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.ell.Vertex(&x, &y)
	return x, y, cmd
}

type spanGenAdapter struct {
	sg *span.SpanImageFilterRGBABilinearClip[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]]
}

func (a *spanGenAdapter) Prepare() {}
func (a *spanGenAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	if length <= 0 {
		return
	}
	a.sg.Generate(colors[:length], x, y)
}

func clampU8(v float64) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 1:
		return 255
	default:
		return uint8(v*255.0 + 0.5)
	}
}

func renderCtrl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: ctrl}, uint32(pathID))
		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, color.RGBA8[color.Linear]{
			R: clampU8(c.R),
			G: clampU8(c.G),
			B: clampU8(c.B),
			A: clampU8(c.A),
		})
	}
}

func renderSolidPath(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	vs rasterizer.VertexSource,
	c color.RGBA8[color.Linear],
) {
	ras.Reset()
	ras.AddPath(vs, 0)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(renBase)
	renSolid.SetColor(c)
	renscan.RenderScanlines(ras, sl, renSolid)
}

func createStar(cx, cy, w, h float64) *path.PathStorageStl {
	r := w
	if h < r {
		r = h
	}
	r1 := r/3 - 8.0
	r2 := r1 / 1.45
	nr := 14

	ps := path.NewPathStorageStl()
	for i := 0; i < nr; i++ {
		a := 2.0*math.Pi*float64(i)/float64(nr) - math.Pi*0.5
		dx := math.Cos(a)
		dy := math.Sin(a)
		if i&1 != 0 {
			ps.LineTo(cx+dx*r1, cy+dy*r1)
			continue
		}
		if i == 0 {
			ps.MoveTo(cx+dx*r2, cy+dy*r2)
			continue
		}
		ps.LineTo(cx+dx*r2, cy+dy*r2)
	}
	ps.ClosePolygon(basics.PathFlagsNone)
	return ps
}

func buildPolygonTransform(cx, cy, angleDeg, scale float64) *transform.TransAffine {
	mtx := transform.NewTransAffine()
	mtx.Translate(-cx, -cy)
	mtx.Rotate(angleDeg * math.Pi / 180.0)
	mtx.Scale(scale)
	mtx.Translate(cx, cy)
	return mtx
}

func buildImageTransform(example int, polygonCX, polygonCY, imageCX, imageCY, imageCenterX, imageCenterY, polygonAngle, polygonScale, imageAngle, imageScale float64) *transform.TransAffine {
	mtx := transform.NewTransAffine()

	switch example {
	case 0:
		// Identity matrix.
	case 1:
		mtx.Translate(-imageCenterX, -imageCenterY)
		mtx.Rotate(polygonAngle * math.Pi / 180.0)
		mtx.Scale(polygonScale)
		mtx.Translate(polygonCX, polygonCY)
		mtx.Invert()
	case 2:
		mtx.Translate(-imageCenterX, -imageCenterY)
		mtx.Rotate(imageAngle * math.Pi / 180.0)
		mtx.Scale(imageScale)
		mtx.Translate(imageCX, imageCY)
		mtx.Invert()
	case 3:
		mtx.Translate(-imageCenterX, -imageCenterY)
		mtx.Rotate(imageAngle * math.Pi / 180.0)
		mtx.Scale(imageScale)
		mtx.Translate(polygonCX, polygonCY)
		mtx.Invert()
	case 4:
		mtx.Translate(-imageCX, -imageCY)
		mtx.Rotate(polygonAngle * math.Pi / 180.0)
		mtx.Scale(polygonScale)
		mtx.Translate(polygonCX, polygonCY)
		mtx.Invert()
	case 5:
		mtx.Translate(-imageCenterX, -imageCenterY)
		mtx.Rotate(imageAngle * math.Pi / 180.0)
		mtx.Rotate(polygonAngle * math.Pi / 180.0)
		mtx.Scale(imageScale)
		mtx.Scale(polygonScale)
		mtx.Translate(imageCX, imageCY)
		mtx.Invert()
	case 6:
		mtx.Translate(-imageCX, -imageCY)
		mtx.Rotate(imageAngle * math.Pi / 180.0)
		mtx.Scale(imageScale)
		mtx.Translate(imageCX, imageCY)
		mtx.Invert()
	}

	return mtx
}

type demo struct {
	srcImg *agg.Image

	polygonAngle  *slider.SliderCtrl
	polygonScale  *slider.SliderCtrl
	imageAngle    *slider.SliderCtrl
	imageScale    *slider.SliderCtrl
	rotatePolygon *checkbox.CheckboxCtrl[color.RGBA]
	rotateImage   *checkbox.CheckboxCtrl[color.RGBA]
	example       *rbox.RboxCtrl[color.RGBA]

	polygonCX, polygonCY float64
	imageCX, imageCY     float64
	imageCenterX         float64
	imageCenterY         float64

	flag   int
	dragDX float64
	dragDY float64
}

func newDemo(srcImg *agg.Image) *demo {
	d := &demo{srcImg: linearizeSRGBImage(srcImg)}

	d.polygonAngle = slider.NewSliderCtrl(5, 5, 145, 11, false)
	d.polygonAngle.SetLabel("Polygon Angle=%3.2f")
	d.polygonAngle.SetRange(-180.0, 180.0)
	d.polygonAngle.SetValue(0.0)

	d.polygonScale = slider.NewSliderCtrl(5, 19, 145, 26, false)
	d.polygonScale.SetLabel("Polygon Scale=%3.2f")
	d.polygonScale.SetRange(0.1, 5.0)
	d.polygonScale.SetValue(1.0)

	d.imageAngle = slider.NewSliderCtrl(155, 5, 300, 12, false)
	d.imageAngle.SetLabel("Image Angle=%3.2f")
	d.imageAngle.SetRange(-180.0, 180.0)
	d.imageAngle.SetValue(0.0)

	d.imageScale = slider.NewSliderCtrl(155, 19, 300, 26, false)
	d.imageScale.SetLabel("Image Scale=%3.2f")
	d.imageScale.SetRange(0.1, 5.0)
	d.imageScale.SetValue(1.0)

	d.rotatePolygon = checkbox.NewDefaultCheckboxCtrl(5, 33, "Rotate Polygon", false)
	d.rotateImage = checkbox.NewDefaultCheckboxCtrl(5, 47, "Rotate Image", false)

	d.example = rbox.NewDefaultRboxCtrl(-3, 56, -3, 56, false)
	d.example.AddItem("0")
	d.example.AddItem("1")
	d.example.AddItem("2")
	d.example.AddItem("3")
	d.example.AddItem("4")
	d.example.AddItem("5")
	d.example.AddItem("6")
	d.example.SetCurItem(0)

	return d
}

func linearizeSRGBImage(src *agg.Image) *agg.Image {
	goImg := src.ToGoImage()
	if goImg == nil {
		return nil
	}

	w := goImg.Bounds().Dx()
	h := goImg.Bounds().Dy()
	buf := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		srcOff := y * goImg.Stride
		dstOff := y * w * 4
		for x := 0; x < w; x++ {
			s := srcOff + x*4
			d := dstOff + x*4
			c := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{
				R: goImg.Pix[s],
				G: goImg.Pix[s+1],
				B: goImg.Pix[s+2],
				A: goImg.Pix[s+3],
			})
			buf[d] = c.R
			buf[d+1] = c.G
			buf[d+2] = c.B
			buf[d+3] = c.A
		}
	}
	return agg.NewImage(buf, w, h, w*4)
}

func (d *demo) controls() []ctrlbase.Ctrl[color.RGBA] {
	return []ctrlbase.Ctrl[color.RGBA]{
		d.polygonAngle,
		d.polygonScale,
		d.imageAngle,
		d.imageScale,
		d.rotatePolygon,
		d.rotateImage,
		d.example,
	}
}

func (d *demo) OnInit() {
	if d.srcImg == nil {
		return
	}

	d.imageCenterX = float64(d.srcImg.Width()) * 0.5
	d.imageCenterY = float64(d.srcImg.Height()) * 0.5
	d.polygonCX = d.imageCenterX
	d.polygonCY = d.imageCenterY
	d.imageCX = d.imageCenterX
	d.imageCY = d.imageCenterY
}

func (d *demo) IsAnimated() bool {
	return d.rotatePolygon.IsChecked() || d.rotateImage.IsChecked()
}

func (d *demo) OnIdle() {
	if d.rotatePolygon.IsChecked() {
		v := d.polygonAngle.Value() + 0.5
		if v >= 180.0 {
			v -= 360.0
		}
		d.polygonAngle.SetValue(v)
	}
	if d.rotateImage.IsChecked() {
		v := d.imageAngle.Value() + 0.5
		if v >= 180.0 {
			v -= 360.0
		}
		d.imageAngle.SetValue(v)
	}
}

func (d *demo) Render(img *agg.Image) {
	if d.srcImg == nil {
		return
	}

	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	w, h := img.Width(), img.Height()

	dstRbuf := buffer.NewRenderingBufferU8()
	dstRbuf.Attach(img.Data, w, h, img.Stride())
	dstPixf := pixfmt.NewPixFmtRGBA32[color.Linear](dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	polygonAngle := d.polygonAngle.Value()
	polygonScale := d.polygonScale.Value()
	imageAngle := d.imageAngle.Value()
	imageScale := d.imageScale.Value()
	if !finitePositive(polygonScale) {
		polygonScale = 1.0
	}
	if !finitePositive(imageScale) {
		imageScale = 1.0
	}

	polygonMtx := buildPolygonTransform(d.polygonCX, d.polygonCY, polygonAngle, polygonScale)
	imageMtx := buildImageTransform(
		d.example.CurItem(),
		d.polygonCX, d.polygonCY,
		d.imageCX, d.imageCY,
		d.imageCenterX, d.imageCenterY,
		polygonAngle, polygonScale,
		imageAngle, imageScale,
	)

	imgRbuf := buffer.NewRenderingBufferU8()
	// Match the C++ image buffer orientation. The low-level runner already
	// flips the output frame, but the source texture itself needs to be sampled
	// from bottom-up row order here.
	imgRbuf.Attach(d.srcImg.Data, d.srcImg.Width(), d.srcImg.Height(), -d.srcImg.Stride())
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{255, 255, 255, 255})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imageMtx, 8)
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(src, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}, interp)
	sgAdp := &spanGenAdapter{sg: sg}

	star := createStar(d.polygonCX, d.polygonCY, float64(w), float64(h))
	starTr := conv.NewConvTransform(path.NewPathStorageStlVertexSourceAdapter(star), polygonMtx)
	ras.ClipBox(0, 0, float64(w), float64(h))
	ras.AddPath(conv.NewRasterizerVertexSourceAdapter(starTr), 0)
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

	// Image center handle, matching the C++ demo.
	e1 := shapes.NewEllipseWithParams(d.imageCX, d.imageCY, 5, 5, 20, false)
	e2 := shapes.NewEllipseWithParams(d.imageCX, d.imageCY, 2, 2, 20, false)
	c1 := conv.NewConvStroke(&ellipseConvAdapter{ell: e1})
	renderSolidPath(ras, sl, renBase, &ellipseVertexSourceAdapter{ell: e1}, color.RGBA8[color.Linear]{R: 179, G: 204, B: 0, A: 255})
	renderSolidPath(ras, sl, renBase, conv.NewRasterizerVertexSourceAdapter(c1), color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})
	renderSolidPath(ras, sl, renBase, &ellipseVertexSourceAdapter{ell: e2}, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// Render the controls last, like the reference example.
	for _, ctrl := range d.controls() {
		renderCtrl(ras, sl, renBase, ctrl)
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}

	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.controls() {
		if ctrl.OnMouseButtonDown(fx, fy) {
			return true
		}
	}

	if math.Hypot(fx-d.imageCX, fy-d.imageCY) < 5.0 {
		d.dragDX = fx - d.imageCX
		d.dragDY = fy - d.imageCY
		d.flag = 1
		return true
	}

	polygonMtx := buildPolygonTransform(d.polygonCX, d.polygonCY, d.polygonAngle.Value(), d.polygonScale.Value())
	star := createStar(d.polygonCX, d.polygonCY, float64(d.srcImg.Width()), float64(d.srcImg.Height()))
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AddPath(conv.NewRasterizerVertexSourceAdapter(conv.NewConvTransform(path.NewPathStorageStlVertexSourceAdapter(star), polygonMtx)), 0)
	if ras.HitTest(int(fx), int(fy)) {
		d.dragDX = fx - d.polygonCX
		d.dragDY = fy - d.polygonCY
		d.flag = 2
		return true
	}

	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false

	for _, ctrl := range d.controls() {
		if ctrl.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}

	if !btn.Left {
		d.flag = 0
		return redraw
	}

	switch d.flag {
	case 1:
		d.imageCX = fx - d.dragDX
		d.imageCY = fy - d.dragDY
		redraw = true
	case 2:
		d.polygonCX = fx - d.dragDX
		d.polygonCY = fy - d.dragDY
		redraw = true
	}

	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false

	for _, ctrl := range d.controls() {
		if ctrl.OnMouseButtonUp(fx, fy) {
			redraw = true
		}
	}

	if d.flag != 0 {
		d.flag = 0
		redraw = true
	}

	return redraw
}

func finitePositive(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > 0.0
}

func main() {
	srcImg, err := imageassets.Spheres()
	if err != nil {
		panic(err)
	}

	lowlevelrunner.Run(runnerConfig(srcImg.Width(), srcImg.Height()), newDemo(srcImg))
}

func runnerConfig(width, height int) lowlevelrunner.Config {
	return lowlevelrunner.Config{
		Title:                 "Image Transforms",
		Width:                 width,
		Height:                height,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}
}
