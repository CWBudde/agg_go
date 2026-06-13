// Port of AGG C++ pattern_fill.cpp.
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
	checkbox "github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	imageacc "github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

const (
	canvasW = 640
	canvasH = 480

	defaultPolygonAngle = 0.0
	defaultPolygonScale = 1.0
	defaultPatternAngle = 0.0
	defaultPatternSize  = 30.0
	defaultPatternAlpha = 0.1
)

type patternPixFmt struct {
	data         []basics.Int8u
	w, h, stride int
}

func (p patternPixFmt) Width() int    { return p.w }
func (p patternPixFmt) Height() int   { return p.h }
func (p patternPixFmt) PixWidth() int { return 4 }

func (p patternPixFmt) PixPtr(x, y int) []basics.Int8u {
	if y < 0 || y >= p.h || x < 0 || x >= p.w {
		return p.data[:0]
	}
	return p.data[y*p.stride+x*4:]
}

type patternSource struct {
	accessor *imageacc.ImageAccessorWrap[patternPixFmt, *imageacc.WrapModeReflectAutoPow2, *imageacc.WrapModeReflectAutoPow2]
	pf       patternPixFmt
}

func (s *patternSource) Width() int                  { return s.pf.w }
func (s *patternSource) Height() int                 { return s.pf.h }
func (s *patternSource) ColorType() string           { return "RGBA8" }
func (s *patternSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *patternSource) Span(x, y, length int) []basics.Int8u {
	return s.accessor.Span(x, y, length)
}
func (s *patternSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *patternSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *patternSource) RowPtr(y int) []basics.Int8u {
	return s.pf.PixPtr(0, y)
}

type rasterizerVertexSource interface {
	Rewind(id uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type rasterizerAdapter struct {
	source rasterizerVertexSource
}

func (a *rasterizerAdapter) Rewind(id uint32) { a.source.Rewind(uint(id)) }
func (a *rasterizerAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.source.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func createStar(ps *path.PathStorageStl, xc, yc, r1, r2 float64, n uint, startAngleDeg float64) {
	ps.RemoveAll()
	startAngleRad := startAngleDeg * math.Pi / 180.0
	for i := uint(0); i < n; i++ {
		a := math.Pi*2.0*float64(i)/float64(n) - math.Pi/2.0
		dx := math.Cos(a + startAngleRad)
		dy := math.Sin(a + startAngleRad)

		if i&1 != 0 {
			ps.LineTo(xc+dx*r1, yc+dy*r1)
			continue
		}
		if i == 0 {
			ps.MoveTo(xc+dx*r2, yc+dy*r2)
			continue
		}
		ps.LineTo(xc+dx*r2, yc+dy*r2)
	}
	ps.ClosePolygon(basics.PathFlagsClose)
}

func generatePattern(size int, patternAngle, patternAlpha float64) patternPixFmt {
	ps := path.NewPathStorageStl()
	createStar(ps, float64(size)/2.0, float64(size)/2.0, float64(size)/2.5, float64(size)/6.0, 6, patternAngle)

	smooth := conv.NewConvSmoothPoly1Curve(path.NewPathStorageStlVertexSourceAdapter(ps))
	smooth.SetSmoothValue(1.0)
	smooth.SetApproximationScale(4.0)

	stroke := conv.NewConvStroke(smooth)
	stroke.SetWidth(float64(size) / 15.0)

	pf := patternPixFmt{
		data:   make([]basics.Int8u, size*size*4),
		w:      size,
		h:      size,
		stride: size * 4,
	}
	rbuf := buffer.NewRenderingBufferWithData[uint8](pf.data, size, size, pf.stride)
	pixf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pixf)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer[*renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](renBase)
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	renBase.Clear(color.NewRGBA8[color.Linear](102, 0, 26, basics.Int8u(patternAlpha*255.0)))

	ras.AddPath(&rasterizerAdapter{source: smooth}, 0)
	renSolid.SetColor(color.ConvertRGBA8SRGBToLinear(color.NewRGBA8[color.SRGB](110, 130, 50, 255)))
	renscan.RenderScanlinesAASolid(ras, sl, renSolid.BaseRenderer(), renSolid.Color())

	ras.AddPath(&rasterizerAdapter{source: stroke}, 0)
	renSolid.SetColor(color.ConvertRGBA8SRGBToLinear(color.NewRGBA8[color.SRGB](0, 50, 80, 255)))
	renscan.RenderScanlinesAASolid(ras, sl, renSolid.BaseRenderer(), renSolid.Color())

	return pf
}

type demo struct {
	polygonAngle  *sliderctrl.SliderCtrl
	polygonScale  *sliderctrl.SliderCtrl
	patternAngle  *sliderctrl.SliderCtrl
	patternSize   *sliderctrl.SliderCtrl
	patternAlpha  *sliderctrl.SliderCtrl
	rotatePolygon *checkbox.CheckboxCtrl[color.RGBA]
	rotatePattern *checkbox.CheckboxCtrl[color.RGBA]
	tiePattern    *checkbox.CheckboxCtrl[color.RGBA]
	polygonCX     float64
	polygonCY     float64
	dragDX        float64
	dragDY        float64
	dragging      bool
}

func newDemo() *demo {
	d := &demo{
		polygonCX: float64(canvasW) / 2.0,
		polygonCY: float64(canvasH) / 2.0,
	}

	d.polygonAngle = sliderctrl.NewSliderCtrl(5, 5, 145, 12, false)
	d.polygonAngle.SetLabel("Polygon Angle=%3.2f")
	d.polygonAngle.SetRange(-180.0, 180.0)
	d.polygonAngle.SetValue(defaultPolygonAngle)

	d.polygonScale = sliderctrl.NewSliderCtrl(5, 19, 145, 26, false)
	d.polygonScale.SetLabel("Polygon Scale=%3.2f")
	d.polygonScale.SetRange(0.1, 5.0)
	d.polygonScale.SetValue(defaultPolygonScale)

	d.patternAngle = sliderctrl.NewSliderCtrl(155, 5, 300, 12, false)
	d.patternAngle.SetLabel("Pattern Angle=%3.2f")
	d.patternAngle.SetRange(-180.0, 180.0)
	d.patternAngle.SetValue(defaultPatternAngle)

	d.patternSize = sliderctrl.NewSliderCtrl(155, 19, 300, 26, false)
	d.patternSize.SetLabel("Pattern Size=%3.2f")
	d.patternSize.SetRange(10.0, 40.0)
	d.patternSize.SetValue(defaultPatternSize)

	d.patternAlpha = sliderctrl.NewSliderCtrl(310, 5, 460, 12, false)
	d.patternAlpha.SetLabel("Background Alpha=%.2f")
	d.patternAlpha.SetRange(0.0, 1.0)
	d.patternAlpha.SetValue(defaultPatternAlpha)

	d.rotatePolygon = checkbox.NewDefaultCheckboxCtrl(5, 33, "Rotate Polygon", false)
	d.rotatePattern = checkbox.NewDefaultCheckboxCtrl(5, 47, "Rotate Pattern", false)
	d.tiePattern = checkbox.NewDefaultCheckboxCtrl(155, 33, "Tie pattern to polygon", false)

	return d
}

func (d *demo) Render(img *agg.Image) {
	dstRbuf := buffer.NewRenderingBufferWithData[uint8](img.Data, img.Width(), img.Height(), img.Stride())
	dstPixf := pixfmt.NewPixFmtRGBA32PreLinear(dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	// Controls render into the plain (non-premultiplied) base, matching C++
	// pattern_fill.cpp which uses renderer_base<pixfmt> rb for render_ctrl while
	// the pattern fill uses renderer_base<pixfmt_pre> rb_pre.
	dstPixfPlain := pixfmt.NewPixFmtRGBA32[color.Linear](dstRbuf)
	renBasePlain := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](dstPixfPlain)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	renBase.Clear(color.NewRGBA8[color.Linear](255, 255, 255, 255))

	size := int(d.patternSize.Value())
	pf := generatePattern(size, d.patternAngle.Value(), d.patternAlpha.Value())

	wrapX := imageacc.NewWrapModeReflectAutoPow2(basics.Int32u(pf.w))
	wrapY := imageacc.NewWrapModeReflectAutoPow2(basics.Int32u(pf.h))
	offsetX, offsetY := uint(0), uint(0)
	if d.tiePattern.IsChecked() {
		offsetX = uint(float64(img.Width()) - d.polygonCX)
		offsetY = uint(float64(img.Height()) - d.polygonCY)
	}
	imgSrc := imageacc.NewImageAccessorWrap[patternPixFmt, *imageacc.WrapModeReflectAutoPow2, *imageacc.WrapModeReflectAutoPow2](&pf, wrapX, wrapY)
	sg := span.NewSpanPatternRGBAWithParams[*patternSource](
		&patternSource{accessor: imgSrc, pf: pf},
		offsetX,
		offsetY,
	)

	ps := path.NewPathStorageStl()
	r := float64(canvasW)/3.0 - 8.0
	createStar(ps, d.polygonCX, d.polygonCY, r, r/1.45, 14, 0.0)

	polygonMtx := transform.NewTransAffine()
	polygonMtx.Multiply(transform.NewTransAffineTranslation(-d.polygonCX, -d.polygonCY))
	polygonMtx.Multiply(transform.NewTransAffineRotation(d.polygonAngle.Value() * math.Pi / 180.0))
	polygonMtx.Multiply(transform.NewTransAffineScaling(d.polygonScale.Value()))
	polygonMtx.Multiply(transform.NewTransAffineTranslation(d.polygonCX, d.polygonCY))

	tr := conv.NewConvTransform(path.NewPathStorageStlVertexSourceAdapter(ps), polygonMtx)
	ras.AddPath(&rasterizerAdapter{source: tr}, 0)
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		sg.Prepare()
		for ras.SweepScanline(sl) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				length := int(spanData.Len)
				if length < 0 {
					length = -length
				}
				if length == 0 {
					continue
				}
				colors := alloc.Allocate(length)
				sg.Generate(colors, int(spanData.X), y, uint(length))
				for i := 0; i < length; i++ {
					colors[i].Premultiply()
				}
				if spanData.Len < 0 {
					renBase.BlendColorHspan(int(spanData.X), y, length, colors, nil, spanData.Covers[0])
					continue
				}
				renBase.BlendColorHspan(int(spanData.X), y, length, colors, spanData.Covers, spanData.Covers[0])
			}
		}
	}

	renderCtrl(ras, sl, renBasePlain, d.polygonAngle)
	renderCtrl(ras, sl, renBasePlain, d.polygonScale)
	renderCtrl(ras, sl, renBasePlain, d.patternAngle)
	renderCtrl(ras, sl, renBasePlain, d.patternSize)
	renderCtrl(ras, sl, renBasePlain, d.patternAlpha)
	renderCtrl(ras, sl, renBasePlain, d.rotatePolygon)
	renderCtrl(ras, sl, renBasePlain, d.rotatePattern)
	renderCtrl(ras, sl, renBasePlain, d.tiePattern)
}

func (d *demo) OnIdle() {
	if d.rotatePolygon.IsChecked() {
		value := d.polygonAngle.Value() + 0.5
		if value >= 180.0 {
			value -= 360.0
		}
		d.polygonAngle.SetValue(value)
	}
	if d.rotatePattern.IsChecked() {
		value := d.patternAngle.Value() - 0.5
		if value <= -180.0 {
			value += 360.0
		}
		d.patternAngle.SetValue(value)
	}
}

func (d *demo) IsAnimated() bool {
	return d.rotatePolygon.IsChecked() || d.rotatePattern.IsChecked()
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	if d.handleCtrlMouseDown(float64(x), float64(y)) {
		return true
	}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ps := path.NewPathStorageStl()
	r := float64(canvasW)/3.0 - 8.0
	createStar(ps, d.polygonCX, d.polygonCY, r, r/1.45, 14, 0.0)
	polygonMtx := transform.NewTransAffine()
	polygonMtx.Multiply(transform.NewTransAffineTranslation(-d.polygonCX, -d.polygonCY))
	polygonMtx.Multiply(transform.NewTransAffineRotation(d.polygonAngle.Value() * math.Pi / 180.0))
	polygonMtx.Multiply(transform.NewTransAffineScaling(d.polygonScale.Value()))
	polygonMtx.Multiply(transform.NewTransAffineTranslation(d.polygonCX, d.polygonCY))
	tr := conv.NewConvTransform(path.NewPathStorageStlVertexSourceAdapter(ps), polygonMtx)
	ras.AddPath(&rasterizerAdapter{source: tr}, 0)
	if ras.HitTest(x, y) {
		d.dragDX = float64(x) - d.polygonCX
		d.dragDY = float64(y) - d.polygonCY
		d.dragging = true
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if d.handleCtrlMouseMove(float64(x), float64(y), btn.Left) {
		return true
	}
	if btn.Left && d.dragging {
		d.polygonCX = float64(x) - d.dragDX
		d.polygonCY = float64(y) - d.dragDY
		return true
	}
	if !btn.Left {
		d.dragging = false
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	changed := d.handleCtrlMouseUp(float64(x), float64(y))
	if d.dragging {
		d.dragging = false
		return true
	}
	return changed || btn.Left
}

func (d *demo) handleCtrlMouseDown(x, y float64) bool {
	for _, c := range d.controls() {
		if c.OnMouseButtonDown(x, y) {
			return true
		}
	}
	return false
}

func (d *demo) handleCtrlMouseMove(x, y float64, pressed bool) bool {
	changed := false
	for _, c := range d.controls() {
		if c.OnMouseMove(x, y, pressed) {
			changed = true
		}
	}
	return changed
}

func (d *demo) handleCtrlMouseUp(x, y float64) bool {
	changed := false
	for _, c := range d.controls() {
		if c.OnMouseButtonUp(x, y) {
			changed = true
		}
	}
	return changed
}

func (d *demo) controls() []ctrlbase.Ctrl[color.RGBA] {
	return []ctrlbase.Ctrl[color.RGBA]{
		d.polygonAngle,
		d.polygonScale,
		d.patternAngle,
		d.patternSize,
		d.patternAlpha,
		d.rotatePolygon,
		d.rotatePattern,
		d.tiePattern,
	}
}

func toRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	return color.ConvertFromRGBA[color.Linear](c)
}

func renderCtrl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineP8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	c ctrlbase.Ctrl[color.RGBA],
) {
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(&rasterizerAdapter{source: &ctrlPathAdapter{ctrl: c, pathID: i}}, uint32(i))
		renscan.RenderScanlinesAASolid(ras, sl, rb, toRGBA8(c.Color(i)))
	}
}

type ctrlPathAdapter struct {
	ctrl   ctrlbase.Ctrl[color.RGBA]
	pathID uint
}

func (a *ctrlPathAdapter) Rewind(id uint) {
	a.pathID = id
	a.ctrl.Rewind(id)
}

func (a *ctrlPathAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	return a.ctrl.Vertex()
}

func demoConfig() lowlevelrunner.Config {
	return lowlevelrunner.Config{
		Title:                 "Pattern Fill",
		Width:                 canvasW,
		Height:                canvasH,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}
}

func main() {
	lowlevelrunner.Run(demoConfig(), newDemo())
}
