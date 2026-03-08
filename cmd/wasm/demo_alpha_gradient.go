// Port of AGG C++ alpha_gradient.cpp.
//
// A large ellipse is filled with a circular color gradient (dark teal →
// yellow-green → dark red) whose alpha channel is modulated by a separate
// XY-product alpha gradient mapped over a draggable parallelogram.  The
// combined effect reveals how the two gradients interact: colours show through
// fully only where both gradients are non-zero.
//
// Three draggable control points define the parallelogram; dragging inside the
// triangle moves all three together.
//
// A spline control at the bottom-left lets the user adjust the alpha curve
// (matching the C++ m_alpha spline_ctrl widget).
package main

import (
	"math"
	"math/rand"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	ctrlspline "agg_go/internal/ctrl/spline"
	"agg_go/internal/shapes"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

// --- State ---

var (
	// Three draggable control points that define the alpha-gradient parallelogram.
	alphaGradPts = [3][2]float64{
		{257, 60},
		{369, 170},
		{143, 310},
	}
	alphaGradSelected = -1
	alphaGradDragDX   = 0.0
	alphaGradDragDY   = 0.0
	alphaGradDragAll  = false

	// Spline control for the alpha curve (bottom-left widget).
	alphaGradSplineCtrl *ctrlspline.SplineCtrl[color.RGBA]
)

func getAlphaGradSplineCtrl() *ctrlspline.SplineCtrl[color.RGBA] {
	if alphaGradSplineCtrl == nil {
		// Match C++: m_alpha(2, 2, 200, 30, 6, !flip_y)
		// In C++ with flip_y=true, y=2 appears near the BOTTOM of the window.
		// We replicate that position at the bottom-left of the 800×600 canvas.
		y1 := float64(height) - 35
		y2 := float64(height) - 5
		ctrl := ctrlspline.NewSplineCtrl[color.RGBA](2, y1, 200, y2, 6, false)

		// C++ initial points: a straight diagonal from (0,0) to (1,1).
		ctrl.SetPoint(0, 0.0, 0.0)
		ctrl.SetPoint(1, 1.0/5.0, 1.0-4.0/5.0) // = (0.2, 0.2)
		ctrl.SetPoint(2, 2.0/5.0, 1.0-3.0/5.0) // = (0.4, 0.4)
		ctrl.SetPoint(3, 3.0/5.0, 1.0-2.0/5.0) // = (0.6, 0.6)
		ctrl.SetPoint(4, 4.0/5.0, 1.0-1.0/5.0) // = (0.8, 0.8)
		ctrl.SetPoint(5, 1.0, 1.0)

		// Default colors matching AGG's spline_ctrl defaults.
		ctrl.SetBackgroundColor(color.NewRGBA(1.0, 1.0, 0.9, 1.0))
		ctrl.SetBorderColor(color.NewRGBA(0.0, 0.0, 0.0, 1.0))
		ctrl.SetCurveColor(color.NewRGBA(0.0, 0.0, 0.0, 1.0))
		ctrl.SetInactivePointColor(color.NewRGBA(0.0, 0.0, 0.0, 1.0))
		ctrl.SetActivePointColor(color.NewRGBA(1.0, 0.0, 0.0, 1.0))

		alphaGradSplineCtrl = ctrl
	}
	return alphaGradSplineCtrl
}

// --- Color-array type (implements span.ColorFunction) ---

type gradColorArray struct {
	data [256]color.RGBA8[color.Linear]
}

func (a *gradColorArray) Size() int { return 256 }

func (a *gradColorArray) ColorAt(i int) color.RGBA8[color.Linear] { return a.data[i] }

// fillColorArray populates a 256-entry LUT with a 3-stop gradient:
// indices 0–127 interpolate begin→middle, indices 128–255 middle→end.
func fillColorArray(arr *gradColorArray, begin, middle, end agg.Color) {
	lerp := func(a, b uint8, t float64) basics.Int8u {
		return basics.Int8u(float64(a)*(1-t) + float64(b)*t)
	}
	for i := 0; i < 128; i++ {
		t := float64(i) / 128.0
		arr.data[i] = color.RGBA8[color.Linear]{
			R: lerp(begin.R, middle.R, t),
			G: lerp(begin.G, middle.G, t),
			B: lerp(begin.B, middle.B, t),
			A: 255,
		}
	}
	for i := 128; i < 256; i++ {
		t := float64(i-128) / 128.0
		arr.data[i] = color.RGBA8[color.Linear]{
			R: lerp(middle.R, end.R, t),
			G: lerp(middle.G, end.G, t),
			B: lerp(middle.B, end.B, t),
			A: 255,
		}
	}
}

// --- Combined span generator ---

// alphaGradSpanGen combines a circular color gradient with an XY alpha gradient
// in a single Generate pass, avoiding the SpanConverter type-parameter complexity.
type alphaGradSpanGen struct {
	gradInterp  *span.SpanInterpolatorLinear[*transform.TransAffine]
	alphaInterp *span.SpanInterpolatorLinear[*transform.TransAffine]
	colorArray  gradColorArray
	alphaArray  [256]basics.Int8u
	// All distances stored in gradient-subpixel units (×GradientSubpixelScale).
	d1c, d2c  int
	d1a, d2a  int
	downscale int // = interpolator.SubpixelShift() - GradientSubpixelShift
}

func newAlphaGradSpanGen(
	gradMtx, alphaMtx *transform.TransAffine,
	colorArr gradColorArray,
	alphaArr [256]basics.Int8u,
) *alphaGradSpanGen {
	gi := span.NewSpanInterpolatorLinearDefault(gradMtx)
	ai := span.NewSpanInterpolatorLinearDefault(alphaMtx)

	ds := gi.SubpixelShift() - span.GradientSubpixelShift
	if ds < 0 {
		ds = 0
	}

	return &alphaGradSpanGen{
		gradInterp:  gi,
		alphaInterp: ai,
		colorArray:  colorArr,
		alphaArray:  alphaArr,
		d1c:         0,
		d2c:         basics.IRound(150 * span.GradientSubpixelScale),
		d1a:         0,
		d2a:         basics.IRound(100 * span.GradientSubpixelScale),
		downscale:   ds,
	}
}

func (g *alphaGradSpanGen) Prepare() {}

func (g *alphaGradSpanGen) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	colorGrad := span.GradientRadial{}
	alphaGrad := span.GradientXY{}

	ddc := g.d2c - g.d1c
	if ddc < 1 {
		ddc = 1
	}
	dda := g.d2a - g.d1a
	if dda < 1 {
		dda = 1
	}

	g.gradInterp.Begin(float64(x)+0.5, float64(y)+0.5, length)
	g.alphaInterp.Begin(float64(x)+0.5, float64(y)+0.5, length)

	for i := 0; i < length; i++ {
		// ---- color gradient (radial, GradientRadial ignores d2) ----
		cx, cy := g.gradInterp.Coordinates()
		d := colorGrad.Calculate(cx>>g.downscale, cy>>g.downscale, g.d2c)
		ci := ((d - g.d1c) * 256) / ddc
		if ci < 0 {
			ci = 0
		} else if ci >= 256 {
			ci = 255
		}
		colors[i] = g.colorArray.data[ci]

		// ---- alpha gradient (XY product, d2 used as divisor) ----
		ax, ay := g.alphaInterp.Coordinates()
		ad := alphaGrad.Calculate(ax>>g.downscale, ay>>g.downscale, g.d2a)
		ai := ((ad - g.d1a) * 256) / dda
		if ai < 0 {
			ai = 0
		} else if ai >= 256 {
			ai = 255
		}
		colors[i].A = g.alphaArray[ai]

		g.gradInterp.Next()
		g.alphaInterp.Next()
	}
}

// --- Ellipse VertexSource adapter (Ellipse.Vertex returns PathCommand, not uint32) ---

type ellipseVS struct{ ell *shapes.Ellipse }

func (a *ellipseVS) Rewind(pathID uint32) { a.ell.Rewind(pathID) }

func (a *ellipseVS) Vertex(x, y *float64) uint32 { return uint32(a.ell.Vertex(x, y)) }

// --- Spline control rendering ---

// renderAlphaSplineCtrl renders the spline control widget using Agg2D's rasterizer.
func renderAlphaSplineCtrl(a *agg.Agg2D, ctrl *ctrlspline.SplineCtrl[color.RGBA]) {
	ras := a.GetInternalRasterizer()
	numPaths := ctrl.NumPaths()
	for i := uint(0); i < numPaths; i++ {
		ras.Reset()
		ctrl.Rewind(i)
		for {
			x, y, cmd := ctrl.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}
		c := ctrl.Color(i)
		a.RenderRasterizerWithColor(agg.RGBA(c.R, c.G, c.B, c.A))
	}
}

// --- Drawing ---

func drawAlphaGradientDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	alphaCtrl := getAlphaGradSplineCtrl()

	cx := float64(width) / 2
	cy := float64(height) / 2

	// 1. Random colourful background ellipses (seed 1234, matches C++ srand(1234)).
	rng := rand.New(rand.NewSource(1234))
	a.NoLine()
	for i := 0; i < 100; i++ {
		ex := float64(rng.Intn(width))
		ey := float64(rng.Intn(height))
		rx := float64(rng.Intn(60)) + 5
		ry := float64(rng.Intn(60)) + 5
		r := uint8(rng.Intn(256))
		g := uint8(rng.Intn(256))
		b := uint8(rng.Intn(256))
		al := uint8(rng.Intn(128))
		a.FillColor(agg.NewColor(r, g, b, al))
		a.ResetPath() // C++ AGG2D resets path before each shape call.
		a.AddEllipse(ex, ey, rx, ry, agg.CCW)
		a.DrawPath(agg.FillOnly)
	}

	// 2. Gradient matrix: scale(0.75, 1.2) × rotate(-π/3) × translate(cx, cy), inverted.
	gradMtx := transform.NewTransAffine()
	gradMtx.Multiply(transform.NewTransAffineScalingXY(0.75, 1.2))
	gradMtx.Multiply(transform.NewTransAffineRotation(-math.Pi / 3.0))
	gradMtx.Multiply(transform.NewTransAffineTranslation(cx, cy))
	gradMtx.Invert()

	// 3. Alpha matrix: parallelogram → rectangle (-100,-100, 100,100).
	parl := [6]float64{
		alphaGradPts[0][0], alphaGradPts[0][1],
		alphaGradPts[1][0], alphaGradPts[1][1],
		alphaGradPts[2][0], alphaGradPts[2][1],
	}
	alphaMtx := transform.NewTransAffineParlToRect(parl, -100, -100, 100, 100)

	// 4. Color LUT: dark teal → yellow-green → dark red.
	var colorArr gradColorArray
	fillColorArray(&colorArr,
		agg.RGBA(0, 0.19, 0.19, 1),
		agg.RGBA(0.7, 0.7, 0.19, 1),
		agg.RGBA(0.31, 0, 0, 1),
	)

	// 5. Alpha LUT from spline control.
	var alphaArr [256]basics.Int8u
	for i := range alphaArr {
		alphaArr[i] = basics.Int8u(alphaCtrl.Value(float64(i)/255.0) * 255)
	}

	// 6. Render the 150-px circle with the combined span generator.
	spanGen := newAlphaGradSpanGen(gradMtx, alphaMtx, colorArr, alphaArr)
	ras := a.GetInternalRasterizer()
	ras.Reset()
	ell := shapes.NewEllipseWithParams(cx, cy, 150, 150, 100, false)
	ras.AddPath(&ellipseVS{ell}, 0)
	a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	// 7. Control points.
	a.NoLine()
	a.FillColor(agg.NewColor(0, 102, 102, 79)) // (0, 0.4, 0.4, 0.31)*255
	for i := 0; i < 3; i++ {
		a.FillCircle(alphaGradPts[i][0], alphaGradPts[i][1], 5)
	}

	// 8. Parallelogram outline (4th point = p0 + p2 − p1).
	p3x := alphaGradPts[0][0] + alphaGradPts[2][0] - alphaGradPts[1][0]
	p3y := alphaGradPts[0][1] + alphaGradPts[2][1] - alphaGradPts[1][1]

	a.LineColor(agg.Black)
	a.LineWidth(1.0)
	a.NoFill()
	a.ResetPath()
	a.MoveTo(alphaGradPts[0][0], alphaGradPts[0][1])
	a.LineTo(alphaGradPts[1][0], alphaGradPts[1][1])
	a.LineTo(alphaGradPts[2][0], alphaGradPts[2][1])
	a.LineTo(p3x, p3y)
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)

	// 9. Spline control widget (bottom-left, matching C++ render_ctrl call).
	renderAlphaSplineCtrl(a, alphaCtrl)
}

// --- Mouse handlers ---

func handleAlphaGradMouseDown(x, y float64) bool {
	alphaCtrl := getAlphaGradSplineCtrl()

	// Let the spline control handle the event first.
	if alphaCtrl.OnMouseButtonDown(x, y) {
		return true
	}

	alphaGradSelected = -1
	alphaGradDragAll = false

	// Check proximity to each control point (hit radius 10 px).
	for i := 0; i < 3; i++ {
		dx := x - alphaGradPts[i][0]
		dy := y - alphaGradPts[i][1]
		if dx*dx+dy*dy < 100 {
			alphaGradSelected = i
			alphaGradDragDX = dx
			alphaGradDragDY = dy
			return true
		}
	}

	// Click inside the triangle → move all three points.
	if pointInTriangle(
		alphaGradPts[0][0], alphaGradPts[0][1],
		alphaGradPts[1][0], alphaGradPts[1][1],
		alphaGradPts[2][0], alphaGradPts[2][1],
		x, y,
	) {
		alphaGradDragAll = true
		alphaGradDragDX = x - alphaGradPts[0][0]
		alphaGradDragDY = y - alphaGradPts[0][1]
		return true
	}

	return false
}

func handleAlphaGradMouseMove(x, y float64) bool {
	alphaCtrl := getAlphaGradSplineCtrl()
	if alphaCtrl.OnMouseMove(x, y, true) {
		return true
	}

	if alphaGradDragAll {
		dx := x - alphaGradDragDX
		dy := y - alphaGradDragDY
		alphaGradPts[1][0] -= alphaGradPts[0][0] - dx
		alphaGradPts[1][1] -= alphaGradPts[0][1] - dy
		alphaGradPts[2][0] -= alphaGradPts[0][0] - dx
		alphaGradPts[2][1] -= alphaGradPts[0][1] - dy
		alphaGradPts[0][0] = dx
		alphaGradPts[0][1] = dy
		return true
	}
	if alphaGradSelected >= 0 {
		alphaGradPts[alphaGradSelected][0] = x - alphaGradDragDX
		alphaGradPts[alphaGradSelected][1] = y - alphaGradDragDY
		return true
	}
	return false
}

func handleAlphaGradMouseUp() {
	alphaCtrl := getAlphaGradSplineCtrl()
	alphaCtrl.OnMouseButtonUp(0, 0)
	alphaGradSelected = -1
	alphaGradDragAll = false
}

// pointInTriangle reports whether (px,py) lies inside the triangle (x1,y1)–(x2,y2)–(x3,y3).
func pointInTriangle(x1, y1, x2, y2, x3, y3, px, py float64) bool {
	sign := func(ax, ay, bx, by, cx, cy float64) float64 {
		return (ax-cx)*(by-cy) - (bx-cx)*(ay-cy)
	}
	d1 := sign(px, py, x1, y1, x2, y2)
	d2 := sign(px, py, x2, y2, x3, y3)
	d3 := sign(px, py, x3, y3, x1, y1)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !hasNeg || !hasPos
}
