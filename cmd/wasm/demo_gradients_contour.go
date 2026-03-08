// Port of AGG C++ gradients_contour.cpp.
//
// Demonstrates three advanced gradient types:
//   - Contour gradient: color transitions follow the shape of an arbitrary path
//     using a Distance Transform (DT) computed from the stroked contour.
//   - Auto Contour: same as Contour but uses the drawn shape itself as the contour.
//   - Asymmetric Conic (Angle) gradient: full-circle color wheel.
//   - Flat Fill: simple solid fill for comparison.
//
// Controls (HTML/URL):
//
//	polygon  (0-3): Star, Great Britain, Spiral, Glyph
//	gradient (0-3): Contour, Auto Contour, Conic, Flat Fill
//	reflect  (bool): mirror the gradient past d2
//	c1, c2   (0-512): contour DT sampling range
//	d1, d2   (0-512): gradient mapping range
//	colors   (2-11): number of color stops in the LUT
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/demo/aggshapes"
	"agg_go/internal/path"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

// --- State ---

var (
	gradientsContourPolygon  = 0     // 0: Star, 1: GB, 2: Spiral, 3: Glyph
	gradientsContourGradient = 1     // 0: Contour, 1: Auto Contour, 2: Conic, 3: Flat
	gradientsContourReflect  = true  // mirror gradient past d2
	gradientsContourC1       = 0.0   // contour DT start  (0–512)
	gradientsContourC2       = 512.0 // contour DT end    (0–512)
	gradientsContourD1       = 0.0   // gradient range start
	gradientsContourD2       = 100.0 // gradient range end
	gradientsContourColors   = 2     // 2–11
)

// --- Setters called from JS / main_stub ---

func setGradientsContourPolygon(v int)  { gradientsContourPolygon = v }
func setGradientsContourGradient(v int) { gradientsContourGradient = v }
func setGradientsContourReflect(v bool) { gradientsContourReflect = v }
func setGradientsContourC1(v float64)   { gradientsContourC1 = v }
func setGradientsContourC2(v float64)   { gradientsContourC2 = v }
func setGradientsContourD1(v float64)   { gradientsContourD1 = v }
func setGradientsContourD2(v float64)   { gradientsContourD2 = v }
func setGradientsContourColors(v int)   { gradientsContourColors = v }

// --- Types ---

// contourConvVS adapts conv.VertexSource to the rasterizer's path interface.
type contourConvVS struct{ vs conv.VertexSource }

func (a *contourConvVS) Rewind(pathID uint32) { a.vs.Rewind(uint(pathID)) }
func (a *contourConvVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vs.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// contourPathVS adapts conv.VertexSource to path.VertexSource for ConcatPath.
type contourPathVS struct{ vs conv.VertexSource }

func (a *contourPathVS) Rewind(pathID uint) { a.vs.Rewind(pathID) }
func (a *contourPathVS) NextVertex() (x, y float64, cmd uint32) {
	vx, vy, c := a.vs.Vertex()
	return vx, vy, uint32(c)
}

// pathStorageVS adapts *path.PathStorage to the rasterizer's path interface.
type pathStorageVS struct{ ps *path.PathStorage }

func (a *pathStorageVS) Rewind(pathID uint32) { a.ps.Rewind(uint(pathID)) }
func (a *pathStorageVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x, *y = vx, vy
	return cmd
}

// gradContourSpanGen is a unified span generator for contour & conic gradients.
type gradContourSpanGen struct {
	interp    *span.SpanInterpolatorLinear[*transform.TransAffine]
	calcFunc  func(x, y, d2 int) int // gradient function; x,y in grad-subpixel units
	reflect   bool
	colors    []color.RGBA8[color.Linear]
	d1scaled  int
	d2scaled  int
	downscale int
}

func (g *gradContourSpanGen) Prepare() {}

func (g *gradContourSpanGen) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
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
			d = d % d2
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

// --- Spiral vertex source ---

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

// --- Color palette builder (matches C++ demo exactly) ---

type gradContourStop struct {
	t       float64
	r, g, b uint8
}

func buildGradientContourLUT(numColors int) []color.RGBA8[color.Linear] {
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
	default: // 11 – spectrum
		stops = buildSpectrumStops()
	}

	const lutSize = 1024
	lut := make([]color.RGBA8[color.Linear], lutSize)

	// Extend stops to cover [0,1] if they don't already.
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
		// find segment
		for j := 1; j < len(stops); j++ {
			if t <= stops[j].t {
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

// wavelengthToRGB converts a wavelength (nm) to an approximate RGB colour.
// This matches the C++ srgba8::from_wavelength implementation.
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
	// Intensity factor at spectrum ends
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

// --- Star path ---

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

// --- Main draw function ---

func drawGradientsContourDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	canvasW := float64(width)
	canvasH := float64(height)

	// Build the main vertex source based on polygon selector.
	// All shapes live in their natural coordinate space; we scale them to fit.
	var mainVS conv.VertexSource  // shape to draw
	contourSrc := buildStarPath() // used as contour source for gradient mode 0

	starPS := buildStarPath() // also used for rendering in polygon mode 0

	switch gradientsContourPolygon {
	case 0: // Star
		mainVS = &psConvVS{ps: starPS}
	case 1: // Great Britain
		gbPS := path.NewPathStorageStl()
		aggshapes.MakeGBPoly(gbPS)
		mainVS = &psStlConvVS{ps: gbPS}
	case 2: // Spiral (stroked)
		sp := newGradSpiral(0, 0, 10, 150, 30, 0)
		stroke := conv.NewConvStroke(sp)
		stroke.SetWidth(22.0)
		mainVS = stroke
	case 3: // Glyph with curves
		glyph := path.NewPathStorage()
		buildGlyphPath(glyph)
		curve := conv.NewConvCurve(&psConvVS{ps: glyph})
		curve.SetApproximationScale(10)
		mainVS = curve
	}

	// Get bounding box of the main shape.
	x1, y1, x2, y2, ok := boundingRectSingleConv(mainVS)
	if !ok {
		return
	}

	// Scale to fit canvas with margins matching C++ (uses width-120, height-120 margin).
	margin := 120.0
	scaleX := (canvasW - margin) / (x2 - x1)
	scaleY := (canvasH - margin) / (y2 - y1)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	scaledW := scale * (x2 - x1)
	scaledH := scale * (y2 - y1)

	// Centre the scaled shape in the canvas.
	offsetX := (canvasW - scaledW) / 2
	offsetY := (canvasH - scaledH) / 2

	// Build the shape-to-screen affine: translate(-x1,-y1) * scale * translate(offsetX, offsetY).
	shapeToScreen := transform.NewTransAffine()
	shapeToScreen.Multiply(transform.NewTransAffineTranslation(-x1, -y1))
	shapeToScreen.Multiply(transform.NewTransAffineScaling(scale))
	shapeToScreen.Multiply(transform.NewTransAffineTranslation(offsetX, offsetY))

	// The gradient interpolator maps screen → gradient space.
	// Gradient space = scaled shape space (after scale * translate(-x1,-y1)),
	// which equals screen space minus the final (offsetX, offsetY) offset.
	// So: gradMtx = translate(-offsetX, -offsetY)
	gradMtx := transform.NewTransAffineTranslation(-offsetX, -offsetY)

	// Build the shape pipeline for rendering.
	mainVS.Rewind(0) // reset after bounding-rect scan
	shapeT := conv.NewConvTransform(mainVS, shapeToScreen)

	// Build the color LUT.
	colors := buildGradientContourLUT(gradientsContourColors)

	ras := a.GetInternalRasterizer()
	interp := span.NewSpanInterpolatorLinearDefault(gradMtx)
	downscale := interp.SubpixelShift() - span.GradientSubpixelShift
	if downscale < 0 {
		downscale = 0
	}

	switch gradientsContourGradient {
	case 0, 1: // Contour / Auto Contour
		gc := span.NewGradientContour()
		gc.SetFrame(0)
		gc.SetD1(gradientsContourC1)
		gc.SetD2(gradientsContourC2)

		// Build the contour path.
		contourPath := path.NewPathStorage()
		if gradientsContourGradient == 0 {
			// Mode 0: contour from the star (unscaled).
			contourPath.ConcatPath(contourSrc, 0)
		} else {
			// Mode 1: auto-contour from the current shape in scaled space.
			// Rebuild main VS without the screen offset (only scale).
			shapeToScaled := transform.NewTransAffine()
			shapeToScaled.Multiply(transform.NewTransAffineTranslation(-x1, -y1))
			shapeToScaled.Multiply(transform.NewTransAffineScaling(scale))
			mainVS.Rewind(0)
			scaledT := conv.NewConvTransform(mainVS, shapeToScaled)
			contourPath.ConcatPath(&contourPathVS{vs: scaledT}, 0)
		}

		gc.ContourCreate(contourPath)

		calcFunc := func(x, y, d2 int) int {
			return gc.Calculate(x, y, d2)
		}

		spanGen := &gradContourSpanGen{
			interp:    interp,
			calcFunc:  calcFunc,
			reflect:   gradientsContourReflect,
			colors:    colors,
			d1scaled:  basics.IRound(gradientsContourD1 * float64(span.GradientSubpixelScale)),
			d2scaled:  basics.IRound(gradientsContourD2 * float64(span.GradientSubpixelScale)),
			downscale: downscale,
		}

		ras.Reset()
		ras.AddPath(&contourConvVS{vs: shapeT}, 0)
		a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	case 2: // Asymmetric Conic (angle) gradient
		// Centre of the conic gradient: centre of the canvas.
		conicCX := canvasW / 2
		conicCY := canvasH / 2
		conicMtx := transform.NewTransAffineTranslation(-conicCX, -conicCY)
		conicInterp := span.NewSpanInterpolatorLinearDefault(conicMtx)
		conicDownscale := conicInterp.SubpixelShift() - span.GradientSubpixelShift
		if conicDownscale < 0 {
			conicDownscale = 0
		}

		conicCalc := func(x, y, d2 int) int {
			res := math.Atan2(float64(y), float64(x))
			if res < 0 {
				v := math.Abs(1600 - math.Round(math.Abs(res)*float64(d2)/math.Pi/2))
				return int(math.Abs(v))
			}
			return basics.IRound(res * float64(d2) / math.Pi / 2)
		}

		spanGen := &gradContourSpanGen{
			interp:    conicInterp,
			calcFunc:  conicCalc,
			reflect:   false,
			colors:    colors,
			d1scaled:  basics.IRound(gradientsContourD1 * float64(span.GradientSubpixelScale)),
			d2scaled:  basics.IRound(gradientsContourD2 * float64(span.GradientSubpixelScale)),
			downscale: conicDownscale,
		}

		shapeT.Rewind(0) // reset shape vertex source for rendering
		mainVS.Rewind(0)
		shapeT2 := conv.NewConvTransform(mainVS, shapeToScreen)

		ras.Reset()
		ras.AddPath(&contourConvVS{vs: shapeT2}, 0)
		a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	case 3: // Flat fill
		shapeT.Rewind(0)
		mainVS.Rewind(0)
		shapeT3 := conv.NewConvTransform(mainVS, shapeToScreen)

		ras.Reset()
		ras.AddPath(&contourConvVS{vs: shapeT3}, 0)
		a.RenderRasterizerWithColor(agg.RGBA(0, 0.6, 0, 0.5))
	}
}

// --- Helper: bounding rect for conv.VertexSource ---

func boundingRectSingleConv(vs conv.VertexSource) (x1, y1, x2, y2 float64, ok bool) {
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
			} else {
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
	}
	return x1, y1, x2, y2, !first
}

// --- Path.PathStorage as conv.VertexSource ---

type psConvVS struct {
	ps *path.PathStorage
}

func (a *psConvVS) Rewind(pathID uint) { a.ps.Rewind(pathID) }
func (a *psConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

// --- path.PathStorageStl as conv.VertexSource ---

type psStlConvVS struct {
	ps *path.PathStorageStl
}

func (a *psStlConvVS) Rewind(pathID uint) { a.ps.Rewind(pathID) }
func (a *psStlConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

// --- Glyph path (matches C++ demo) ---

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
