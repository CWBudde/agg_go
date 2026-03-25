// Package main ports AGG's gradients_contour.cpp demo.
//
// The scene uses a contour-based gradient, an optional bitmap-gradient stroke,
// four shape presets, and a draggable perspective frame. The layout mirrors
// the original platform_support example rather than the earlier simplified Go
// port.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	checkboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	polygonctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/demo/aggshapes"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	frameWidth  = 520
	frameHeight = 520
)

type spanInterpolator interface {
	Begin(x, y float64, length int)
	Next()
	Coordinates() (x, y int)
	SubpixelShift() int
}

type convVS struct{ vs conv.VertexSource }

func (a *convVS) Rewind(id uint32) { a.vs.Rewind(uint(id)) }
func (a *convVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vs.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type pathConvVS struct{ ps *path.PathStorage }

func (a *pathConvVS) Rewind(id uint) { a.ps.Rewind(id) }
func (a *pathConvVS) NextVertex() (x, y float64, cmd uint32) {
	return a.ps.NextVertex()
}

type convAsPath struct{ vs conv.VertexSource }

func (a *convAsPath) Rewind(id uint) { a.vs.Rewind(id) }
func (a *convAsPath) NextVertex() (x, y float64, cmd uint32) {
	vx, vy, c := a.vs.Vertex()
	return vx, vy, uint32(c)
}

type stlConvVS struct{ ps *path.PathStorageStl }

func (a *stlConvVS) Rewind(id uint) { a.ps.Rewind(id) }
func (a *stlConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

type pathAsConv struct{ ps *path.PathStorage }

func (a *pathAsConv) Rewind(id uint) { a.ps.Rewind(id) }
func (a *pathAsConv) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

type contourSpanGen struct {
	interp    spanInterpolator
	calcFunc  func(x, y, d2 int) int
	reflect   bool
	colors    []color.RGBA8[color.Linear]
	d1scaled  int
	d2scaled  int
	downscale int
}

func (g *contourSpanGen) Prepare() {}

func (g *contourSpanGen) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	g.interp.Begin(float64(x)+0.5, float64(y)+0.5, length)
	nColors := len(g.colors)
	dRange := g.d2scaled - g.d1scaled
	if dRange < 1 {
		dRange = 1
	}
	for i := 0; i < length; i++ {
		ix, iy := g.interp.Coordinates()
		d := g.calcFunc(ix>>g.downscale, iy>>g.downscale, g.d2scaled)
		if g.reflect {
			d2 := g.d2scaled * 2
			d %= d2
			if d < 0 {
				d += d2
			}
			if d >= g.d2scaled {
				d = d2 - d
			}
		}
		ci := ((d - g.d1scaled) * nColors) / dRange
		if ci < 0 {
			ci = 0
		} else if ci >= nColors {
			ci = nColors - 1
		}
		colors[i] = g.colors[ci]
		g.interp.Next()
	}
}

type gradSpiral struct {
	x, y, r1, r2, step, startAngle float64
	angle, currR, da, dr           float64
	started                        bool
}

func newGradSpiral(x, y, r1, r2, step, startAngle float64) *gradSpiral {
	return &gradSpiral{
		x:          x,
		y:          y,
		r1:         r1,
		r2:         r2,
		step:       step,
		startAngle: startAngle,
		da:         basics.Deg2RadF(4.0),
		dr:         step / 90.0,
	}
}

func (s *gradSpiral) Rewind(_ uint) {
	s.angle = s.startAngle
	s.currR = s.r1
	s.started = false
}

func (s *gradSpiral) Vertex() (x, y float64, cmd basics.PathCommand) {
	if s.currR > s.r2 {
		return 0, 0, basics.PathCmdStop
	}
	x = s.x + math.Cos(s.angle)*s.currR
	y = s.y + math.Sin(s.angle)*s.currR
	s.currR += s.dr
	s.angle += s.da
	if !s.started {
		s.started = true
		return x, y, basics.PathCmdMoveTo
	}
	return x, y, basics.PathCmdLineTo
}

type demo struct {
	polygons *rboxctrl.RboxCtrl[color.RGBA]
	gradient *rboxctrl.RboxCtrl[color.RGBA]
	stroke   *checkboxctrl.CheckboxCtrl[color.RGBA]
	refl     *checkboxctrl.CheckboxCtrl[color.RGBA]
	c1       *sliderctrl.SliderCtrl
	c2       *sliderctrl.SliderCtrl
	d1       *sliderctrl.SliderCtrl
	d2       *sliderctrl.SliderCtrl
	clrs     *sliderctrl.SliderCtrl
	persp    *polygonctrl.PolygonCtrl[color.RGBA]
	lastPoly int
	init     bool
}

func newDemo() *demo {
	polygons := rboxctrl.NewDefaultRboxCtrl(5.0, 5.0, 135.0, 90.0, false)
	polygons.SetTextSize(9.0, 0.0)
	polygons.SetTextThickness(1.0)
	polygons.AddItem("Simple Path")
	polygons.AddItem("Great Britain")
	polygons.AddItem("Spiral")
	polygons.AddItem("Glyph")
	polygons.SetCurItem(0)

	gradient := rboxctrl.NewDefaultRboxCtrl(145.0, 5.0, 305.0, 90.0, false)
	gradient.SetTextSize(9.0, 0.0)
	gradient.SetTextThickness(1.0)
	gradient.AddItem("Contour")
	gradient.AddItem("Auto Contour")
	gradient.AddItem("Asymmetric Conic")
	gradient.AddItem("Flat Fill")
	gradient.SetCurItem(1)

	stroke := checkboxctrl.NewDefaultCheckboxCtrl(305.0, 77.0, "Bitmap Gradient", false)
	refl := checkboxctrl.NewDefaultCheckboxCtrl(440.0, 77.0, "Reflect", false)
	refl.SetChecked(true)

	c1 := sliderctrl.NewSliderCtrl(310.0, 10.0, 400.0, 16.0, false)
	c1.SetLabel("C1=%1.0f")
	c1.SetRange(0.0, 512.0)
	c1.SetValue(0.0)

	c2 := sliderctrl.NewSliderCtrl(310.0, 30.0, 400.0, 36.0, false)
	c2.SetLabel("C2=%1.0f")
	c2.SetRange(0.0, 512.0)
	c2.SetValue(512.0)

	d1 := sliderctrl.NewSliderCtrl(410.0, 10.0, 500.0, 16.0, false)
	d1.SetLabel("D1=%1.0f")
	d1.SetRange(0.0, 512.0)
	d1.SetValue(0.0)

	d2 := sliderctrl.NewSliderCtrl(410.0, 30.0, 500.0, 36.0, false)
	d2.SetLabel("D2=%1.0f")
	d2.SetRange(0.1, 512.0)
	d2.SetValue(100.0)

	clrs := sliderctrl.NewSliderCtrl(310.0, 50.0, 500.0, 58.0, false)
	clrs.SetLabel("Colors=%1.0f")
	clrs.SetRange(2.0, 11.0)
	clrs.SetValue(2.0)
	clrs.SetNumSteps(9)

	persp := polygonctrl.NewDefaultPolygonCtrl(4, 5.0)
	persp.SetClose(true)
	persp.SetInPolygonCheck(true)
	persp.SetLineColor(color.NewRGBA(0.0, 0.3, 0.5, 0.3))

	return &demo{
		polygons: polygons,
		gradient: gradient,
		stroke:   stroke,
		refl:     refl,
		c1:       c1,
		c2:       c2,
		d1:       d1,
		d2:       d2,
		clrs:     clrs,
		persp:    persp,
		lastPoly: 0,
		init:     true,
	}
}

func (d *demo) mainVS() conv.VertexSource {
	switch d.polygons.CurItem() {
	case 1:
		gbPS := path.NewPathStorageStl()
		aggshapes.MakeGBPoly(gbPS)
		return &stlConvVS{ps: gbPS}
	case 2:
		sp := newGradSpiral(0, 0, 10, 150, 30, 0.0)
		stroke := conv.NewConvStroke(sp)
		stroke.SetWidth(22.0)
		return stroke
	case 3:
		glyph := path.NewPathStorage()
		buildGlyphPath(glyph)
		curve := conv.NewConvCurve(&pathAsConv{ps: glyph})
		curve.SetApproximationScale(10.0)
		return curve
	default:
		return &pathAsConv{ps: buildStarPath()}
	}
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	if d.polygons.CurItem() != d.lastPoly {
		d.lastPoly = d.polygons.CurItem()
		d.init = true
	}

	vs := d.mainVS()
	x1, y1, x2, y2, ok := boundingRect(vs)
	if !ok {
		return
	}

	margin := 120.0
	scaleX := (float64(frameWidth) - margin) / (x2 - x1)
	scaleY := (float64(frameHeight) - margin) / (y2 - y1)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	scaledW := scale * (x2 - x1)
	scaledH := scale * (y2 - y1)

	if d.init {
		d.persp.SetXn(0, 100.0)
		d.persp.SetYn(0, 105.0)
		d.persp.SetXn(1, 100.0+scaledW)
		d.persp.SetYn(1, 105.0)
		d.persp.SetXn(2, 100.0+scaledW)
		d.persp.SetYn(2, 105.0+scaledH)
		d.persp.SetXn(3, 100.0)
		d.persp.SetYn(3, 105.0+scaledH)
		d.init = false
	}

	affine := transform.NewTransAffine()
	affine.Multiply(transform.NewTransAffineTranslation(-x1, -y1))
	affine.Multiply(transform.NewTransAffineScaling(scale))

	quad := [8]float64{
		d.persp.Xn(0), d.persp.Yn(0),
		d.persp.Xn(1), d.persp.Yn(1),
		d.persp.Xn(2), d.persp.Yn(2),
		d.persp.Xn(3), d.persp.Yn(3),
	}

	scaledToQuad := transform.NewTransPerspectiveRectToQuad(0, 0, scaledW, scaledH, quad)
	scaledToQuad.IsValid(1e-12)

	colors := buildGradientContourLUT(int(math.Round(d.clrs.Value())))
	ras := a.GetInternalRasterizer()

	switch d.gradient.CurItem() {
	case 0, 1:
		gc := span.NewGradientContour()
		gc.SetFrame(0)
		gc.SetD1(d.c1.Value())
		gc.SetD2(d.c2.Value())

		contourPath := path.NewPathStorage()
		if d.gradient.CurItem() == 0 {
			contourPath.ConcatPath(&pathConvVS{ps: buildStarPath()}, 0)
		} else {
			shapeToScaled := conv.NewConvTransform(vs, affine)
			contourPath.ConcatPath(&convAsPath{vs: shapeToScaled}, 0)
		}
		if gc.ContourCreate(contourPath) == nil {
			return
		}

		interp := span.NewSpanInterpolatorPerspectiveLerpQuadToRect(quad, 0, 0, scaledW, scaledH, 8)
		downscale := interp.SubpixelShift() - span.GradientSubpixelShift
		if downscale < 0 {
			downscale = 0
		}

		spanGen := &contourSpanGen{
			interp:    interp,
			calcFunc:  func(x, y, d2 int) int { return gc.Calculate(x, y, d2) },
			reflect:   d.refl.IsChecked(),
			colors:    colors,
			d1scaled:  basics.IRound(d.d1.Value() * float64(span.GradientSubpixelScale)),
			d2scaled:  basics.IRound(d.d2.Value() * float64(span.GradientSubpixelScale)),
			downscale: downscale,
		}

		shapeToScaled := conv.NewConvTransform(vs, affine)
		shapeToQuad := conv.NewConvTransform(shapeToScaled, scaledToQuad)
		ras.Reset()
		ras.AddPath(&convVS{vs: shapeToQuad}, 0)
		a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	case 2:
		conicMtx := transform.NewTransAffineTranslation(-270.0, -300.0)
		interp := span.NewSpanInterpolatorLinearDefault(conicMtx)
		downscale := interp.SubpixelShift() - span.GradientSubpixelShift
		if downscale < 0 {
			downscale = 0
		}

		calcFunc := func(x, y, d2 int) int {
			res := math.Atan2(float64(y), float64(x))
			if res < 0 {
				v := math.Abs(1600 - math.Round(math.Abs(res)*float64(d2)/math.Pi/2))
				return int(math.Abs(v))
			}
			return basics.IRound(res * float64(d2) / math.Pi / 2)
		}

		spanGen := &contourSpanGen{
			interp:    interp,
			calcFunc:  calcFunc,
			reflect:   false,
			colors:    colors,
			d1scaled:  basics.IRound(d.d1.Value() * float64(span.GradientSubpixelScale)),
			d2scaled:  basics.IRound(d.d2.Value() * float64(span.GradientSubpixelScale)),
			downscale: downscale,
		}

		shapeToScaled := conv.NewConvTransform(vs, affine)
		shapeToQuad := conv.NewConvTransform(shapeToScaled, scaledToQuad)
		ras.Reset()
		ras.AddPath(&convVS{vs: shapeToQuad}, 0)
		a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	case 3:
		shapeToScaled := conv.NewConvTransform(vs, affine)
		shapeToQuad := conv.NewConvTransform(shapeToScaled, scaledToQuad)
		ras.Reset()
		ras.AddPath(&convVS{vs: shapeToQuad}, 0)
		a.RenderRasterizerWithColor(agg.RGBA(0, 0.6, 0, 0.5))
	}

	renderCtrl(a, d.persp)
	renderCtrl(a, d.polygons)
	renderCtrl(a, d.gradient)
	renderCtrl(a, d.stroke)
	renderCtrl(a, d.refl)
	renderCtrl(a, d.c1)
	renderCtrl(a, d.c2)
	renderCtrl(a, d.d1)
	renderCtrl(a, d.d2)
	renderCtrl(a, d.clrs)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	handled := false
	for _, ctrl := range d.controls() {
		if ctrl.OnMouseButtonDown(fx, fy) {
			handled = true
		}
	}
	if d.polygons.CurItem() != d.lastPoly {
		d.init = true
	}
	return handled
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false
	for _, ctrl := range d.controls() {
		if ctrl.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}
	if d.polygons.CurItem() != d.lastPoly {
		d.init = true
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
	if d.polygons.CurItem() != d.lastPoly {
		d.init = true
		redraw = true
	}
	return redraw
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

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[color.RGBA]) {
	ras := a.GetInternalRasterizer()
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
}

func toAggColor(c color.RGBA) agg.Color {
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
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

type gradContourStop struct {
	t       float64
	r, g, b uint8
}

func buildGradientContourLUT(numColors int) []color.RGBA8[color.Linear] {
	if numColors < 2 {
		numColors = 2
	}
	if numColors > 11 {
		numColors = 11
	}

	var stops []gradContourStop
	switch numColors {
	case 2:
		stops = []gradContourStop{
			{0.0, 178, 34, 34},
			{1.0, 255, 255, 0},
		}
	case 3:
		stops = []gradContourStop{
			{0.0, 245, 233, 131},
			{0.5, 146, 35, 219},
			{1.0, 255, 35, 0},
		}
	case 4:
		stops = []gradContourStop{
			{0.0, 0, 0, 255},
			{0.2, 120, 120, 0},
			{0.7, 120, 0, 0},
			{1.0, 0, 255, 0},
		}
	case 5:
		stops = []gradContourStop{
			{0.2, 230, 188, 106},
			{0.4, 207, 148, 31},
			{0.6, 69, 56, 30},
			{0.8, 43, 33, 13},
			{1.0, 227, 221, 209},
		}
	case 6:
		stops = []gradContourStop{
			{0.0, 125, 99, 255},
			{0.2, 118, 79, 210},
			{0.4, 105, 58, 81},
			{0.6, 217, 74, 102},
			{0.8, 242, 148, 90},
			{1.0, 242, 200, 102},
		}
	case 7:
		stops = []gradContourStop{
			{0.00, 216, 237, 232},
			{0.16, 196, 214, 226},
			{0.32, 175, 194, 217},
			{0.48, 155, 176, 210},
			{0.64, 140, 162, 202},
			{0.80, 130, 149, 193},
			{1.00, 72, 102, 165},
		}
	case 8:
		stops = []gradContourStop{
			{0.00, 255, 223, 168},
			{0.14, 255, 199, 162},
			{0.28, 255, 175, 156},
			{0.42, 255, 151, 151},
			{0.56, 255, 127, 145},
			{0.70, 255, 104, 140},
			{0.84, 255, 80, 133},
			{1.00, 255, 56, 128},
		}
	case 9:
		stops = []gradContourStop{
			{0.000, 255, 4, 163},
			{0.125, 255, 4, 109},
			{0.250, 255, 4, 46},
			{0.375, 255, 75, 75},
			{0.500, 255, 120, 83},
			{0.625, 255, 143, 83},
			{0.750, 255, 180, 83},
			{0.875, 255, 209, 83},
			{1.000, 255, 246, 83},
		}
	case 10:
		stops = []gradContourStop{
			{0.00, 255, 0, 0},
			{0.11, 255, 198, 198},
			{0.22, 255, 255, 0},
			{0.33, 255, 255, 226},
			{0.44, 85, 85, 255},
			{0.55, 226, 226, 255},
			{0.66, 28, 255, 28},
			{0.77, 226, 255, 226},
			{0.88, 255, 72, 255},
			{1.00, 255, 227, 255},
		}
	default:
		stops = buildSpectrumStops()
	}

	const lutSize = 1024
	lut := make([]color.RGBA8[color.Linear], lutSize)
	if len(stops) > 0 && stops[0].t > 0 {
		s := stops[0]
		s.t = 0
		stops = append([]gradContourStop{s}, stops...)
	}
	if len(stops) > 0 && stops[len(stops)-1].t < 1 {
		s := stops[len(stops)-1]
		s.t = 1
		stops = append(stops, s)
	}

	lerp := func(a, b uint8, t float64) uint8 {
		return uint8(float64(a)*(1-t) + float64(b)*t + 0.5)
	}

	for i := 0; i < lutSize; i++ {
		t := float64(i) / float64(lutSize-1)
		for j := 1; j < len(stops); j++ {
			if t > stops[j].t {
				continue
			}
			dt := stops[j].t - stops[j-1].t
			var lt float64
			if dt > 0 {
				lt = (t - stops[j-1].t) / dt
			}
			lut[i] = color.RGBA8[color.Linear]{
				R: lerp(stops[j-1].r, stops[j].r, lt),
				G: lerp(stops[j-1].g, stops[j].g, lt),
				B: lerp(stops[j-1].b, stops[j].b, lt),
				A: 255,
			}
			break
		}
	}

	return lut
}

func buildSpectrumStops() []gradContourStop {
	gamma := 1.8
	wavelengths := []float64{380, 420, 460, 500, 540, 580, 620, 660, 700, 740, 780}
	stops := make([]gradContourStop, len(wavelengths))
	for i, wl := range wavelengths {
		r, g, b := wavelengthToRGB(wl, gamma)
		stops[i] = gradContourStop{t: float64(i) / float64(len(wavelengths)-1), r: r, g: g, b: b}
	}
	return stops
}

func wavelengthToRGB(wl, gamma float64) (r, g, b uint8) {
	var fr, fg, fb float64
	switch {
	case wl >= 380 && wl <= 440:
		fr = -(wl - 440) / (440 - 380)
		fb = 1
	case wl >= 440 && wl <= 490:
		fg = (wl - 440) / (490 - 440)
		fb = 1
	case wl >= 490 && wl <= 510:
		fg = 1
		fb = -(wl - 510) / (510 - 490)
	case wl >= 510 && wl <= 580:
		fr = (wl - 510) / (580 - 510)
		fg = 1
	case wl >= 580 && wl <= 645:
		fr = 1
		fg = -(wl - 645) / (645 - 580)
	case wl >= 645 && wl <= 780:
		fr = 1
	}

	var factor float64
	switch {
	case wl >= 380 && wl <= 420:
		factor = 0.3 + 0.7*(wl-380)/(420-380)
	case wl >= 700 && wl <= 780:
		factor = 0.3 + 0.7*(780-wl)/(780-700)
	default:
		factor = 1.0
	}

	pow := func(v float64) float64 {
		if v <= 0 {
			return 0
		}
		return math.Pow(v*factor, gamma)
	}

	r = uint8(pow(fr) * 255)
	g = uint8(pow(fg) * 255)
	b = uint8(pow(fb) * 255)
	return
}

func buildStarPath() *path.PathStorage {
	ps := path.NewPathStorage()
	ps.MoveTo(12, 40)
	ps.LineTo(52, 40)
	ps.LineTo(72, 6)
	ps.LineTo(92, 40)
	ps.LineTo(132, 40)
	ps.LineTo(112, 76)
	ps.LineTo(132, 112)
	ps.LineTo(92, 112)
	ps.LineTo(72, 148)
	ps.LineTo(52, 112)
	ps.LineTo(12, 112)
	ps.LineTo(32, 76)
	ps.ClosePolygon(0)
	return ps
}

func buildGlyphPath(ps *path.PathStorage) {
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
}

func boundingRect(vs conv.VertexSource) (x1, y1, x2, y2 float64, ok bool) {
	vs.Rewind(0)
	first := true
	for {
		x, y, cmd := vs.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if basics.IsVertex(cmd) {
			if first {
				x1, y1, x2, y2 = x, y, x, y
				first = false
				continue
			}
			if x < x1 {
				x1 = x
			}
			if y < y1 {
				y1 = y
			}
			if x > x2 {
				x2 = x
			}
			if y > y2 {
				y2 = y
			}
		}
	}
	return x1, y1, x2, y2, !first
}

func (d *demo) controls() []ctrlbase.Ctrl[color.RGBA] {
	return []ctrlbase.Ctrl[color.RGBA]{
		d.persp,
		d.polygons,
		d.gradient,
		d.stroke,
		d.refl,
		d.c1,
		d.c2,
		d.d1,
		d.d2,
		d.clrs,
	}
}

func (d *demo) OnKey(_ rune) bool { return false }

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Gradients Contour (Distance Transform)",
		Width:  frameWidth,
		Height: frameHeight,
		FlipY:  true,
		// The AGG reference demo feeds display-space sRGBA bytes into the
		// gradient LUT and control colors directly. Re-encoding the finished
		// framebuffer as sRGB would apply a second transfer curve and wash the
		// image out, so keep the output buffer in the same byte space as the
		// original C++ screenshot.
		DisableLinearRGBToSRGB: true,
	}, newDemo())
}
