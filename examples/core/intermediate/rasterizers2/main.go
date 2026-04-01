// Port of AGG C++ rasterizers2.cpp example.
//
// Demonstrates five different rasterization methods for spiral paths:
//   - Bresenham with pixel-rounded accuracy
//   - Bresenham with subpixel accuracy
//   - Anti-aliased outline renderer
//   - Scanline rasterizer with conv_stroke
//   - Image-pattern outline renderer
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/gsv"
	"github.com/cwbudde/agg_go/internal/order"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
	"github.com/cwbudde/agg_go/internal/primitives"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	outline "github.com/cwbudde/agg_go/internal/renderer/outline"
	rprimitives "github.com/cwbudde/agg_go/internal/renderer/primitives"
	"github.com/cwbudde/agg_go/internal/scanline"
)

const (
	frameWidth  = 500
	frameHeight = 450
)

var pixmapChain = []uint32{
	16, 7,
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff975252, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xa4c7a2a2, 0x3affff00, 0x3affff00, 0xff975151, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff954f4f, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
}

// --- type aliases for readability ---

type (
	pixFmt    = *pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]]
	colorType = color.RGBA8[color.Linear]
	renBaseT  = *renderer.RendererBase[pixFmt, colorType]
	rasAAType = *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	slType    = *scanline.ScanlineP8
	renPrimT  = *rprimitives.RendererPrimitives[renBaseT, colorType]
)

// --- interface adapters between rasterizer.VertexSource and conv.VertexSource ---

// convToRasAdapter wraps a conv.VertexSource for use with rasterizer.AddPath.
type convToRasAdapter struct {
	src interface {
		Rewind(pathID uint)
		Vertex() (x, y float64, cmd basics.PathCommand)
	}
}

func (a *convToRasAdapter) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }

func (a *convToRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// rasToConvAdapter wraps a rasterizer.VertexSource for use with conv.NewConvStroke.
type rasToConvAdapter struct {
	src rasterizer.VertexSource
}

func (a *rasToConvAdapter) Rewind(pathID uint) { a.src.Rewind(uint32(pathID)) }

func (a *rasToConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	c := a.src.Vertex(&x, &y)
	cmd = basics.PathCommand(c)
	return
}

// --- spiral vertex source (implements rasterizer.VertexSource) ---

type spiral struct {
	x, y                 float64
	r1, r2               float64
	step, startAngle     float64
	angle, currR, da, dr float64
	start                bool
}

func newSpiral(x, y, r1, r2, step, startAngle float64) *spiral {
	return &spiral{
		x: x, y: y, r1: r1, r2: r2,
		step: step, startAngle: startAngle,
		da: basics.Deg2RadF(8.0),
		dr: step / 45.0,
	}
}

func (s *spiral) Rewind(uint32) {
	s.angle = s.startAngle
	s.currR = s.r1
	s.start = true
}

func (s *spiral) Vertex(x, y *float64) uint32 {
	if s.currR > s.r2 {
		return uint32(basics.PathCmdStop)
	}
	*x = s.x + math.Cos(s.angle)*s.currR
	*y = s.y + math.Sin(s.angle)*s.currR
	s.currR += s.dr
	s.angle += s.da
	if s.start {
		s.start = false
		return uint32(basics.PathCmdMoveTo)
	}
	return uint32(basics.PathCmdLineTo)
}

// --- roundoff source (implements rasterizer.VertexSource) ---

type roundoffSource struct {
	src *spiral
}

func (r *roundoffSource) Rewind(pathID uint32) { r.src.Rewind(pathID) }

func (r *roundoffSource) Vertex(x, y *float64) uint32 {
	cmd := r.src.Vertex(x, y)
	if basics.IsVertex(basics.PathCommand(cmd)) {
		*x = math.Floor(*x)
		*y = math.Floor(*y)
	}
	return cmd
}

// --- pattern source ---

type chainPatternSource struct {
	data []uint32
}

func (s *chainPatternSource) Width() float64  { return float64(s.data[0]) }
func (s *chainPatternSource) Height() float64 { return float64(s.data[1]) }

func (s *chainPatternSource) Pixel(x, y int) color.RGBA {
	w := int(s.data[0])
	idx := y*w + x + 2
	if idx < 2 || idx >= len(s.data) {
		return color.NewRGBA(0, 0, 0, 0)
	}
	p := s.data[idx]
	c := color.NewRGBAFromRGBA8(
		uint8((p>>16)&0xFF),
		uint8((p>>8)&0xFF),
		uint8(p&0xFF),
		uint8((p>>24)&0xFF),
	)
	c.Premultiply()
	return c
}

// --- outline renderer adapters ---

type outlineBaseAdapter struct {
	renBase renBaseT
}

func (a *outlineBaseAdapter) Width() int  { return a.renBase.Width() }
func (a *outlineBaseAdapter) Height() int { return a.renBase.Height() }

func (a *outlineBaseAdapter) BlendSolidHSpan(x, y, length int, c colorType, covers []basics.CoverType) {
	conv := make([]basics.Int8u, len(covers))
	for i := range covers {
		conv[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidHspan(x, y, length, c, conv)
}

func (a *outlineBaseAdapter) BlendSolidVSpan(x, y, length int, c colorType, covers []basics.CoverType) {
	conv := make([]basics.Int8u, len(covers))
	for i := range covers {
		conv[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidVspan(x, y, length, c, conv)
}

type outlineAAAdapter struct {
	ren *outline.RendererOutlineAA[*outlineBaseAdapter, colorType]
}

func (a *outlineAAAdapter) AccurateJoinOnly() bool { return a.ren.AccurateJoinOnly() }
func (a *outlineAAAdapter) Color(c colorType)      { a.ren.Color(c) }

func (a *outlineAAAdapter) Line0(lp primitives.LineParameters)             { a.ren.Line0(&lp) }         //nolint:gocritic
func (a *outlineAAAdapter) Line1(lp primitives.LineParameters, sx, sy int) { a.ren.Line1(&lp, sx, sy) } //nolint:gocritic
func (a *outlineAAAdapter) Line2(lp primitives.LineParameters, ex, ey int) { a.ren.Line2(&lp, ex, ey) } //nolint:gocritic
func (a *outlineAAAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) { //nolint:gocritic
	a.ren.Line3(&lp, sx, sy, ex, ey)
}
func (a *outlineAAAdapter) Pie(x, y, x1, y1, x2, y2 int) { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *outlineAAAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}

// --- image outline adapters ---

type imageBaseAdapter struct {
	renBase renBaseT
}

func rgbaToRGBA8(c color.RGBA) colorType {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return colorType{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

func (a *imageBaseAdapter) BlendColorHSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]colorType, len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorHspan(x, y, length, buf, nil, basics.CoverFull)
}

func (a *imageBaseAdapter) BlendColorVSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]colorType, len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorVspan(x, y, length, buf, nil, basics.CoverFull)
}

type outlineImageAdapter struct {
	ren *outline.RendererOutlineImage
}

func (a *outlineImageAdapter) AccurateJoinOnly() bool                     { return a.ren.AccurateJoinOnly() }
func (a *outlineImageAdapter) Color(colorType)                            {}
func (a *outlineImageAdapter) Line0(primitives.LineParameters)            {} //nolint:gocritic
func (a *outlineImageAdapter) Line1(primitives.LineParameters, int, int)  {} //nolint:gocritic
func (a *outlineImageAdapter) Line2(primitives.LineParameters, int, int)  {} //nolint:gocritic
func (a *outlineImageAdapter) Pie(int, int, int, int, int, int)           {}
func (a *outlineImageAdapter) Semidot(func(int) bool, int, int, int, int) {}
func (a *outlineImageAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) { //nolint:gocritic
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

// --- helpers ---

func renderScanlines(ras rasAAType, sl slType, renBase renBaseT, col colorType) {
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(sl) {
		y := sl.Y()
		for _, sp := range sl.Spans() {
			x := int(sp.X)
			if sp.Len > 0 {
				renBase.BlendSolidHspan(x, y, int(sp.Len), col, sp.Covers)
			} else {
				// Solid span: single cover value repeated for |Len| pixels.
				// C++: blend_hline(x, y, x - len - 1, color, *covers)
				renBase.BlendHline(x, y, x-int(sp.Len)-1, col, sp.Covers[0])
			}
		}
	}
}

func drawText(ras rasAAType, sl slType, renBase renBaseT, x, y float64, txt string) {
	t := gsv.NewGSVText()
	t.SetSize(8, 0)
	t.SetText(txt)
	t.SetStartPoint(x, y)

	stroke := conv.NewConvStroke(t)
	stroke.SetWidth(0.7)

	ras.Reset()
	ras.AddPath(&convToRasAdapter{src: stroke}, 0)
	renderScanlines(ras, sl, renBase, colorType{R: 0, G: 0, B: 0, A: 255})
}

func renderControl(
	ras rasAAType, sl slType, renBase renBaseT,
	numPaths uint,
	rewindFn func(pathID uint),
	vertexFn func(x, y *float64) uint32,
	colorFn func(pathID uint) color.RGBA,
) {
	adapter := &ctrlPathAdapter{rewindFn: rewindFn, vertexFn: vertexFn}
	for pathID := uint(0); pathID < numPaths; pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		col := rgbaToRGBA8(colorFn(pathID))
		renderScanlines(ras, sl, renBase, col)
	}
}

type ctrlPathAdapter struct {
	rewindFn func(pathID uint)
	vertexFn func(x, y *float64) uint32
}

func (a *ctrlPathAdapter) Rewind(pathID uint32) { a.rewindFn(uint(pathID)) }

func (a *ctrlPathAdapter) Vertex(x, y *float64) uint32 { return a.vertexFn(x, y) }

// --- checkbox rendering adapter ---

func renderCheckbox(ras rasAAType, sl slType, renBase renBaseT, cb *checkbox.CheckboxCtrl[color.RGBA]) {
	renderControl(ras, sl, renBase, cb.NumPaths(), cb.Rewind,
		func(x, y *float64) uint32 {
			vx, vy, cmd := cb.Vertex()
			*x = vx
			*y = vy
			return uint32(cmd)
		},
		cb.Color,
	)
}

func renderSlider(ras rasAAType, sl slType, renBase renBaseT, s *sliderctrl.SliderCtrl) {
	renderControl(ras, sl, renBase, s.NumPaths(), s.Rewind,
		func(x, y *float64) uint32 {
			vx, vy, cmd := s.Vertex()
			*x = vx
			*y = vy
			return uint32(cmd)
		},
		s.Color,
	)
}

// --- demo ---

type demo struct {
	stepSlider    *sliderctrl.SliderCtrl
	widthSlider   *sliderctrl.SliderCtrl
	testPerf      *checkbox.CheckboxCtrl[color.RGBA]
	rotate        *checkbox.CheckboxCtrl[color.RGBA]
	accurateJoins *checkbox.CheckboxCtrl[color.RGBA]
	scalePattern  *checkbox.CheckboxCtrl[color.RGBA]
	startAngle    float64
}

func newDemo() *demo {
	// C++: m_step(10.0, 10.0 + 4.0, 150.0, 10.0 + 8.0 + 4.0, !flip_y)
	step := sliderctrl.NewSliderCtrl(10.0, 14.0, 150.0, 22.0, false)
	step.SetRange(0.0, 2.0)
	step.SetValue(0.1)
	step.SetLabel("Step=%1.2f")

	// C++: m_width(150.0 + 10.0, 10.0 + 4.0, 400 - 10.0, 10.0 + 8.0 + 4.0, !flip_y)
	width := sliderctrl.NewSliderCtrl(160.0, 14.0, 390.0, 22.0, false)
	width.SetRange(0.0, 14.0)
	width.SetValue(3.0)
	width.SetLabel("Width=%1.2f")

	// C++: checkboxes at y = 10.0 + 4.0 + 16.0 = 30.0, !flip_y
	testPerf := checkbox.NewDefaultCheckboxCtrl(10.0, 30.0, "Test Performance", false)
	testPerf.SetTextSize(9.0, 7.0)

	rotate := checkbox.NewDefaultCheckboxCtrl(140.0, 30.0, "Rotate", false)
	rotate.SetTextSize(9.0, 7.0)

	accurateJoins := checkbox.NewDefaultCheckboxCtrl(210.0, 30.0, "Accurate Joins", false)
	accurateJoins.SetTextSize(9.0, 7.0)

	scalePattern := checkbox.NewDefaultCheckboxCtrl(320.0, 30.0, "Scale Pattern", false)
	scalePattern.SetTextSize(9.0, 7.0)
	scalePattern.SetChecked(true)

	return &demo{
		stepSlider:    step,
		widthSlider:   width,
		testPerf:      testPerf,
		rotate:        rotate,
		accurateJoins: accurateJoins,
		scalePattern:  scalePattern,
	}
}

func (d *demo) Render(img *agg.Image) {
	w := img.Width()
	h := img.Height()
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, w, h, img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[pixFmt, colorType](pf)
	renBase.Clear(colorType{R: 255, G: 255, B: 242, A: 255})

	rasAA := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	renPrim := rprimitives.NewRendererPrimitives[renBaseT, colorType](renBase)
	rasAliased := rasterizer.NewRasterizerOutline[renPrimT, colorType](renPrim)

	lineWidth := d.widthSlider.Value()

	profile := outline.NewLineProfileAA()
	profile.Width(lineWidth)
	renOutlineAA := outline.NewRendererOutlineAA[*outlineBaseAdapter, colorType](
		&outlineBaseAdapter{renBase: renBase}, profile,
	)
	rasOutlineAA := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter, colorType](
		&outlineAAAdapter{ren: renOutlineAA},
	)
	// C++: line_join depends on accurate_joins checkbox
	if d.accurateJoins.IsChecked() {
		rasOutlineAA.SetLineJoin(rasterizer.OutlineMiterAccurateJoin)
	} else {
		rasOutlineAA.SetLineJoin(rasterizer.OutlineRoundJoin)
	}
	rasOutlineAA.SetRoundCap(true)

	patternSource := &chainPatternSource{data: pixmapChain}
	filter := outline.NewPatternFilterRGBAAdapter()
	pattern := outline.NewLineImagePatternPow2(filter)
	if d.scalePattern.IsChecked() {
		scaledSrc := outline.NewLineImageScale(patternSource, lineWidth)
		pattern.Create(scaledSrc)
	} else {
		pattern.Create(patternSource)
	}
	renImg := outline.NewRendererOutlineImage(&imageBaseAdapter{renBase: renBase}, pattern)
	if d.scalePattern.IsChecked() {
		renImg.SetScaleX(lineWidth / patternSource.Height())
	}
	rasImg := rasterizer.NewRasterizerOutlineAA[*outlineImageAdapter, colorType](
		&outlineImageAdapter{ren: renImg},
	)

	brown := colorType{R: 102, G: 77, B: 26, A: 255}
	fw := float64(w)
	fh := float64(h)

	// (1) Bresenham lines, pixel-rounded accuracy
	s1 := newSpiral(fw/5, fh/4+50, 5, 70, 16, d.startAngle)
	renPrim.LineColor(brown)
	rasAliased.AddPath(&roundoffSource{src: s1}, 0)

	// (2) Bresenham lines, subpixel accuracy
	s2 := newSpiral(fw/2, fh/4+50, 5, 70, 16, d.startAngle)
	renPrim.LineColor(brown)
	rasAliased.AddPath(s2, 0)

	// (3) Anti-aliased outline
	s3 := newSpiral(fw/5, fh-fh/4+20, 5, 70, 16, d.startAngle)
	renOutlineAA.Color(brown)
	rasOutlineAA.AddPath(s3, 0)

	// (4) Scanline rasterizer with conv_stroke
	s4 := newSpiral(fw/2, fh-fh/4+20, 5, 70, 16, d.startAngle)
	stroke := conv.NewConvStroke(&rasToConvAdapter{src: s4})
	stroke.SetWidth(lineWidth)
	stroke.SetLineCap(basics.RoundCap)
	rasAA.Reset()
	rasAA.AddPath(&convToRasAdapter{src: stroke}, 0)
	renderScanlines(rasAA, sl, renBase, brown)

	// (5) Arbitrary image pattern
	s5 := newSpiral(fw-fw/5, fh-fh/4+20, 5, 70, 16, d.startAngle)
	rasImg.AddPath(s5, 0)

	// Labels
	drawText(rasAA, sl, renBase, 50, 80, "Bresenham lines,\n\nregular accuracy")
	drawText(rasAA, sl, renBase, fw/2-50, 80, "Bresenham lines,\n\nsubpixel accuracy")
	drawText(rasAA, sl, renBase, 50, fh/2+50, "Anti-aliased lines")
	drawText(rasAA, sl, renBase, fw/2-50, fh/2+50, "Scanline rasterizer")
	drawText(rasAA, sl, renBase, fw-fw/5-50, fh/2+50, "Arbitrary Image Pattern")

	// Controls
	renderSlider(rasAA, sl, renBase, d.stepSlider)
	renderSlider(rasAA, sl, renBase, d.widthSlider)
	renderCheckbox(rasAA, sl, renBase, d.testPerf)
	renderCheckbox(rasAA, sl, renBase, d.rotate)
	renderCheckbox(rasAA, sl, renBase, d.accurateJoins)
	renderCheckbox(rasAA, sl, renBase, d.scalePattern)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	if d.stepSlider.OnMouseButtonDown(fx, fy) {
		return true
	}
	if d.widthSlider.OnMouseButtonDown(fx, fy) {
		return true
	}
	if d.testPerf.OnMouseButtonDown(fx, fy) {
		return true
	}
	if d.rotate.OnMouseButtonDown(fx, fy) {
		return true
	}
	if d.accurateJoins.OnMouseButtonDown(fx, fy) {
		return true
	}
	if d.scalePattern.OnMouseButtonDown(fx, fy) {
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := d.stepSlider.OnMouseMove(fx, fy, btn.Left)

	if d.widthSlider.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := d.stepSlider.OnMouseButtonUp(fx, fy)

	if d.widthSlider.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	if d.testPerf.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	if d.rotate.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	if d.accurateJoins.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	if d.scalePattern.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	return redraw
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Rasterizers 2",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
