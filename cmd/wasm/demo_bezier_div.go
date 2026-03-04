// Port of AGG C++ bezier_div.cpp – Bezier Curve Subdivision with accuracy metrics.
package main

import (
	"fmt"
	"math"
	"time"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/conv"
	"agg_go/internal/ctrl/bezier"
	"agg_go/internal/ctrl/checkbox"
	"agg_go/internal/ctrl/rbox"
	"agg_go/internal/ctrl/slider"
	"agg_go/internal/curves"
	"agg_go/internal/path"
)

// --- Enum mappings (matching C++ line_cap_e, line_join_e, inner_join_e) ---

var (
	bdLineCaps   = []basics.LineCap{basics.ButtCap, basics.SquareCap, basics.RoundCap}
	bdLineJoins  = []basics.LineJoin{basics.MiterJoin, basics.MiterJoinRevert, basics.RoundJoin, basics.BevelJoin, basics.MiterJoinRound}
	bdInnerJoins = []basics.InnerJoin{basics.InnerBevel, basics.InnerMiter, basics.InnerJag, basics.InnerRound}
)

// --- Controls ---

var (
	bdCurve1      *bezier.BezierCtrl[agg.Color]
	bdAngleTol    *slider.SliderCtrl
	bdApproxScale *slider.SliderCtrl
	bdCuspLimit   *slider.SliderCtrl
	bdWidth       *slider.SliderCtrl
	bdShowPoints  *checkbox.CheckboxCtrl[agg.Color]
	bdShowOutline *checkbox.CheckboxCtrl[agg.Color]
	bdCurveType   *rbox.RboxCtrl[agg.Color]
	bdCaseType    *rbox.RboxCtrl[agg.Color]
	bdInnerJoin   *rbox.RboxCtrl[agg.Color]
	bdLineJoin    *rbox.RboxCtrl[agg.Color]
	bdLineCap     *rbox.RboxCtrl[agg.Color]

	bdInitialized bool
	bdCurCaseType = -1
)

func initBezierDivDemo() {
	if bdInitialized {
		return
	}

	ctrlColor := agg.RGBA(0, 0.3, 0.5, 0.8)

	bdCurve1 = bezier.NewBezierCtrl[agg.Color](ctrlColor)
	bdCurve1.SetCurve(170, 424, 13, 87, 488, 423, 26, 333)

	bdAngleTol = slider.NewSliderCtrl(5, 5, 240, 12, false)
	bdAngleTol.SetRange(0, 90)
	bdAngleTol.SetValue(15)
	bdAngleTol.SetLabel("Angle Tolerance=%.0f deg")

	bdApproxScale = slider.NewSliderCtrl(5, 22, 240, 29, false)
	bdApproxScale.SetRange(0.1, 5)
	bdApproxScale.SetValue(1.0)
	bdApproxScale.SetLabel("Approximation Scale=%.3f")

	bdCuspLimit = slider.NewSliderCtrl(5, 39, 240, 46, false)
	bdCuspLimit.SetRange(0, 90)
	bdCuspLimit.SetValue(0)
	bdCuspLimit.SetLabel("Cusp Limit=%.0f deg")

	bdWidth = slider.NewSliderCtrl(245, 5, 495, 12, false)
	bdWidth.SetRange(-50, 100)
	bdWidth.SetValue(50)
	bdWidth.SetLabel("Width=%.2f")

	inact := agg.NewColor(0, 0, 0, 255)
	txtC := agg.NewColor(0, 0, 0, 255)
	act := agg.NewColor(102, 0, 0, 255)

	bdShowPoints = checkbox.NewCheckboxCtrl[agg.Color](250, 20, "Show Points", false, inact, txtC, act)
	bdShowPoints.SetChecked(true)

	bdShowOutline = checkbox.NewCheckboxCtrl[agg.Color](250, 35, "Show Stroke Outline", false, inact, txtC, act)
	bdShowOutline.SetChecked(true)

	bg := agg.RGBA(1, 1, 1, 0.5)
	border := agg.NewColor(0, 0, 0, 255)
	txtRb := agg.NewColor(0, 0, 0, 255)
	inactRb := agg.NewColor(0, 0, 0, 255)
	actRb := agg.NewColor(102, 0, 0, 255)

	bdCurveType = rbox.NewRboxCtrl[agg.Color](535, 5, 650, 55, false, bg, border, txtRb, inactRb, actRb)
	bdCurveType.AddItem("Incremental")
	bdCurveType.AddItem("Subdiv")
	bdCurveType.SetCurItem(1)

	bdCaseType = rbox.NewRboxCtrl[agg.Color](535, 60, 650, 195, false, bg, border, txtRb, inactRb, actRb)
	bdCaseType.SetTextSize(7, 0)
	bdCaseType.SetTextThickness(1.0)
	bdCaseType.AddItem("Random")
	bdCaseType.AddItem("13---24")
	bdCaseType.AddItem("Smooth Cusp 1")
	bdCaseType.AddItem("Smooth Cusp 2")
	bdCaseType.AddItem("Real Cusp 1")
	bdCaseType.AddItem("Real Cusp 2")
	bdCaseType.AddItem("Fancy Stroke")
	bdCaseType.AddItem("Jaw")
	bdCaseType.AddItem("Ugly Jaw")
	bdCaseType.SetCurItem(0)

	bdInnerJoin = rbox.NewRboxCtrl[agg.Color](535, 200, 650, 290, false, bg, border, txtRb, inactRb, actRb)
	bdInnerJoin.SetTextSize(8, 0)
	bdInnerJoin.AddItem("Inner Bevel")
	bdInnerJoin.AddItem("Inner Miter")
	bdInnerJoin.AddItem("Inner Jag")
	bdInnerJoin.AddItem("Inner Round")
	bdInnerJoin.SetCurItem(3)

	bdLineJoin = rbox.NewRboxCtrl[agg.Color](535, 295, 650, 385, false, bg, border, txtRb, inactRb, actRb)
	bdLineJoin.SetTextSize(8, 0)
	bdLineJoin.AddItem("Miter Join")
	bdLineJoin.AddItem("Miter Revert")
	bdLineJoin.AddItem("Round Join")
	bdLineJoin.AddItem("Bevel Join")
	bdLineJoin.AddItem("Miter Round")
	bdLineJoin.SetCurItem(1)

	bdLineCap = rbox.NewRboxCtrl[agg.Color](535, 395, 650, 455, false, bg, border, txtRb, inactRb, actRb)
	bdLineCap.SetTextSize(8, 0)
	bdLineCap.AddItem("Butt Cap")
	bdLineCap.AddItem("Square Cap")
	bdLineCap.AddItem("Round Cap")
	bdLineCap.SetCurItem(0)

	bdInitialized = true
}

// bdHandleCaseTypeChange updates the curve to a preset when the case type changes.
func bdHandleCaseTypeChange() {
	item := bdCaseType.CurItem()
	if item == bdCurCaseType {
		return
	}
	switch item {
	case 0: // Random – leave current curve
	case 1:
		bdCurve1.SetCurve(150, 150, 350, 150, 150, 150, 350, 150)
	case 2:
		bdCurve1.SetCurve(50, 142, 483, 251, 496, 62, 26, 333)
	case 3:
		bdCurve1.SetCurve(50, 142, 484, 251, 496, 62, 26, 333)
	case 4:
		bdCurve1.SetCurve(100, 100, 300, 200, 200, 200, 200, 100)
	case 5:
		bdCurve1.SetCurve(475, 157, 200, 100, 453, 100, 222, 157)
	case 6:
		bdCurve1.SetCurve(129, 233, 32, 283, 258, 285, 159, 232)
		bdWidth.SetValue(100)
	case 7:
		bdCurve1.SetCurve(100, 100, 300, 200, 264, 286, 264, 284)
	case 8:
		bdCurve1.SetCurve(100, 100, 413, 304, 264, 286, 264, 284)
	}
	bdCurCaseType = item
}

// bdBuildCurvePath generates path vertices from a cubic Bezier curve into a PathStorageStl.
func bdBuildCurvePath(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit float64, incremental bool) *path.PathStorageStl {
	ps := path.NewPathStorageStl()

	if incremental {
		c := curves.NewCurve4Inc()
		c.SetApproximationScale(approxScale)
		c.Init(x1, y1, x2, y2, x3, y3, x4, y4)
		c.Rewind(0)
		first := true
		for {
			x, y, cmd := c.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if first || basics.IsMoveTo(cmd) {
				ps.MoveTo(x, y)
				first = false
			} else {
				ps.LineTo(x, y)
			}
		}
	} else {
		c := curves.NewCurve4Div()
		c.SetApproximationScale(approxScale)
		c.SetAngleTolerance(angleTol)
		c.SetCuspLimit(cuspLimit)
		c.Init(x1, y1, x2, y2, x3, y3, x4, y4)
		c.Rewind(0)
		for {
			x, y, cmd := c.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsMoveTo(cmd) {
				ps.MoveTo(x, y)
			} else {
				ps.LineTo(x, y)
			}
		}
	}
	return ps
}

// bdMeasureTime measures the time to generate the curve 100 times (in microseconds).
func bdMeasureTime(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit float64, incremental bool) float64 {
	start := time.Now()
	for i := 0; i < 100; i++ {
		if incremental {
			c := curves.NewCurve4Inc()
			c.SetApproximationScale(approxScale)
			c.Init(x1, y1, x2, y2, x3, y3, x4, y4)
			c.Rewind(0)
			for {
				_, _, cmd := c.Vertex()
				if basics.IsStop(cmd) {
					break
				}
			}
		} else {
			c := curves.NewCurve4Div()
			c.SetApproximationScale(approxScale)
			c.SetAngleTolerance(angleTol)
			c.SetCuspLimit(cuspLimit)
			c.Init(x1, y1, x2, y2, x3, y3, x4, y4)
			c.Rewind(0)
			for {
				_, _, cmd := c.Vertex()
				if basics.IsStop(cmd) {
					break
				}
			}
		}
	}
	return float64(time.Since(start).Microseconds()) / 100.0
}

// bdCalcDistance computes Euclidean distance between two points.
func bdCalcDistance(x1, y1, x2, y2 float64) float64 {
	dx, dy := x2-x1, y2-y1
	return math.Sqrt(dx*dx + dy*dy)
}

// bdBezier4Point evaluates the cubic Bezier curve at parameter mu in [0,1].
func bdBezier4Point(x1, y1, x2, y2, x3, y3, x4, y4, mu float64) (float64, float64) {
	mum1 := 1 - mu
	mum13 := mum1 * mum1 * mum1
	mu3 := mu * mu * mu
	x := mum13*x1 + 3*mu*mum1*mum1*x2 + 3*mu*mu*mum1*x3 + mu3*x4
	y := mum13*y1 + 3*mu*mum1*mum1*y2 + 3*mu*mu*mum1*y3 + mu3*y4
	return x, y
}

// bdCalcLinePointDist computes signed distance from a point to a line.
func bdCalcLinePointDist(x1, y1, x2, y2, x, y float64) float64 {
	dx, dy := x2-x1, y2-y1
	d := math.Sqrt(dx*dx + dy*dy)
	if d < 1e-10 {
		return bdCalcDistance(x1, y1, x, y)
	}
	return ((x-x1)*dy - (y-y1)*dx) / d
}

type bdCurvePoint struct{ x, y, dist float64 }

// bdCalcMaxError computes approximation accuracy (dist error, angle error) at a given scale.
func bdCalcMaxError(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit, scale float64, incremental bool) (float64, float64) {
	scaledApprox := approxScale * scale

	var cps []bdCurvePoint
	if incremental {
		c := curves.NewCurve4Inc()
		c.SetApproximationScale(scaledApprox)
		c.Init(x1, y1, x2, y2, x3, y3, x4, y4)
		c.Rewind(0)
		for {
			x, y, cmd := c.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsVertex(cmd) {
				cps = append(cps, bdCurvePoint{x: x, y: y})
			}
		}
	} else {
		c := curves.NewCurve4Div()
		c.SetApproximationScale(scaledApprox)
		c.SetAngleTolerance(angleTol)
		c.SetCuspLimit(cuspLimit)
		c.Init(x1, y1, x2, y2, x3, y3, x4, y4)
		c.Rewind(0)
		for {
			x, y, cmd := c.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsVertex(cmd) {
				cps = append(cps, bdCurvePoint{x: x, y: y})
			}
		}
	}

	if len(cps) < 2 {
		return 0, 0
	}

	// Compute cumulative arc length for curve points
	curveDist := 0.0
	for i := 1; i < len(cps); i++ {
		cps[i-1].dist = curveDist
		curveDist += bdCalcDistance(cps[i-1].x, cps[i-1].y, cps[i].x, cps[i].y)
	}
	cps[len(cps)-1].dist = curveDist

	// Generate 4096 reference points on the true Bezier curve
	const nRef = 4096
	refs := make([]bdCurvePoint, nRef)
	for i := 0; i < nRef; i++ {
		mu := float64(i) / float64(nRef-1)
		refs[i].x, refs[i].y = bdBezier4Point(x1, y1, x2, y2, x3, y3, x4, y4, mu)
	}
	refDist := 0.0
	for i := 1; i < nRef; i++ {
		refs[i-1].dist = refDist
		refDist += bdCalcDistance(refs[i-1].x, refs[i-1].y, refs[i].x, refs[i].y)
	}
	refs[nRef-1].dist = refDist

	// For each reference point, binary-search the nearest segment and measure distance
	maxErr := 0.0
	for _, ref := range refs {
		lo, hi := 0, len(cps)-1
		for hi-lo > 1 {
			k := (lo + hi) >> 1
			if ref.dist < cps[k].dist {
				hi = k
			} else {
				lo = k
			}
		}
		if lo >= hi || hi >= len(cps) {
			continue
		}
		err := math.Abs(bdCalcLinePointDist(cps[lo].x, cps[lo].y, cps[hi].x, cps[hi].y, ref.x, ref.y))
		if err > maxErr {
			maxErr = err
		}
	}

	// Angle error: max angle between consecutive segments
	maxAngle := 0.0
	for i := 2; i < len(cps); i++ {
		a1 := math.Atan2(cps[i-1].y-cps[i-2].y, cps[i-1].x-cps[i-2].x)
		a2 := math.Atan2(cps[i].y-cps[i-1].y, cps[i].x-cps[i-1].x)
		da := math.Abs(a1 - a2)
		if da >= math.Pi {
			da = 2*math.Pi - da
		}
		if da > maxAngle {
			maxAngle = da
		}
	}

	return maxErr * scale, maxAngle * 180.0 / math.Pi
}

// --- Rendering helpers ---

// bdIterPath feeds a VertexSource into the agg2d path builder.
func bdIterPath(a *agg.Agg2D, src conv.VertexSource) {
	for {
		x, y, cmd := src.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		switch {
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
		case basics.IsEndPoly(cmd):
			if basics.IsClose(uint32(cmd)) {
				a.ClosePolygon()
			}
		case basics.IsVertex(cmd):
			a.LineTo(x, y)
		}
	}
}

// renderCheckbox renders a CheckboxCtrl via the internal rasterizer.
func renderCheckbox(agg2d *agg.Agg2D, c *checkbox.CheckboxCtrl[agg.Color]) {
	ras := agg2d.GetInternalRasterizer()
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		adapter := &checkboxAdapter{c: c}
		ras.AddPath(adapter, uint32(i))
		agg2d.RenderRasterizerWithColor(c.Color(i))
	}
}

type checkboxAdapter struct {
	c *checkbox.CheckboxCtrl[agg.Color]
}

func (a *checkboxAdapter) Rewind(pathID uint32) { a.c.Rewind(uint(pathID)) }
func (a *checkboxAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.c.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// renderBezierCtrl renders a BezierCtrl via the internal rasterizer.
func renderBezierCtrl(agg2d *agg.Agg2D, b *bezier.BezierCtrl[agg.Color]) {
	ras := agg2d.GetInternalRasterizer()
	for i := uint(0); i < b.NumPaths(); i++ {
		ras.Reset()
		adapter := &bezierCtrlAdapter{b: b}
		ras.AddPath(adapter, uint32(i))
		agg2d.RenderRasterizerWithColor(b.Color(i))
	}
}

type bezierCtrlAdapter struct{ b *bezier.BezierCtrl[agg.Color] }

func (a *bezierCtrlAdapter) Rewind(pathID uint32) { a.b.Rewind(uint(pathID)) }
func (a *bezierCtrlAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.b.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// --- Main demo ---

func drawBezierDivDemo() {
	initBezierDivDemo()
	bdHandleCaseTypeChange()

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Light cream background (rgba(1.0, 1.0, 0.95) from original)
	a.FillColor(agg.NewColor(255, 255, 242, 255))
	a.NoLine()
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(float64(width), 0)
	a.LineTo(float64(width), float64(height))
	a.LineTo(0, float64(height))
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Get control point coordinates from BezierCtrl
	x1, y1 := bdCurve1.X1(), bdCurve1.Y1()
	x2, y2 := bdCurve1.X2(), bdCurve1.Y2()
	x3, y3 := bdCurve1.X3(), bdCurve1.Y3()
	x4, y4 := bdCurve1.X4(), bdCurve1.Y4()

	approxScale := bdApproxScale.Value()
	angleTol := bdAngleTol.Value() * math.Pi / 180.0
	cuspLimit := bdCuspLimit.Value() * math.Pi / 180.0
	strokeWidth := bdWidth.Value()
	incremental := bdCurveType.CurItem() == 0

	// Build curve path
	curvePath := bdBuildCurvePath(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit, incremental)

	// Count vertices
	numPoints := 0
	curvePath.Rewind(0)
	for {
		_, _, cmd := curvePath.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		if basics.IsVertex(basics.PathCommand(cmd)) {
			numPoints++
		}
	}

	// Create stroke converter
	curveAdapter := path.NewPathStorageStlVertexSourceAdapter(curvePath)
	stroke := conv.NewConvStroke(curveAdapter)
	stroke.SetWidth(strokeWidth)
	stroke.SetLineJoin(bdLineJoins[bdLineJoin.CurItem()])
	stroke.SetLineCap(bdLineCaps[bdLineCap.CurItem()])
	stroke.SetInnerJoin(bdInnerJoins[bdInnerJoin.CurItem()])
	stroke.SetInnerMiterLimit(1.01)

	// Draw wide filled stroke (rgba(0, 0.5, 0, 0.5) = green semi-transparent)
	a.ResetPath()
	stroke.Rewind(0)
	bdIterPath(a, stroke)
	a.FillColor(agg.RGBA(0, 0.5, 0, 0.5))
	a.NoLine()
	a.DrawPath(agg.FillOnly)

	// Show subdivision points as small dots (r=1.5)
	if bdShowPoints.IsChecked() {
		a.FillColor(agg.RGBA(0, 0, 0, 0.5))
		a.NoLine()
		curvePath.Rewind(0)
		for {
			x, y, cmd := curvePath.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if basics.IsVertex(basics.PathCommand(cmd)) {
				a.FillCircle(x, y, 1.5)
			}
		}
	}

	// Show stroke outline (stroke of stroke, thin black)
	if bdShowOutline.IsChecked() {
		stroke2 := conv.NewConvStroke(stroke)
		a.ResetPath()
		stroke2.Rewind(0)
		bdIterPath(a, stroke2)
		a.FillColor(agg.RGBA(0, 0, 0, 0.5))
		a.NoLine()
		a.DrawPath(agg.FillOnly)
	}

	// Measure performance and accuracy
	curveTime := bdMeasureTime(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit, incremental)

	e01, ae01 := bdCalcMaxError(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit, 0.01, incremental)
	e1, ae1 := bdCalcMaxError(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit, 0.1, incremental)
	e10, ae10 := bdCalcMaxError(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit, 1, incremental)
	e100, ae100 := bdCalcMaxError(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit, 10, incremental)
	e1000, ae1000 := bdCalcMaxError(x1, y1, x2, y2, x3, y3, x4, y4, approxScale, angleTol, cuspLimit, 100, incremental)

	statsText := fmt.Sprintf(
		"Num Points=%d Time=%.2fmks\n\n"+
			" Dist Error: x0.01=%.5f x0.1=%.5f x1=%.5f x10=%.5f x100=%.5f\n\n"+
			"Angle Error: x0.01=%.1f x0.1=%.1f x1=%.1f x10=%.1f x100=%.1f",
		numPoints, curveTime,
		e01, e1, e10, e100, e1000,
		ae01, ae1, ae10, ae100, ae1000,
	)

	a.FillColor(agg.Black)
	a.NoLine()
	a.FontGSV(10)
	a.Text(10, 445, statsText, false, 0, 0)

	// Render all controls
	renderBezierCtrl(a, bdCurve1)
	renderSlider(a, bdAngleTol)
	renderSlider(a, bdApproxScale)
	renderSlider(a, bdCuspLimit)
	renderSlider(a, bdWidth)
	renderCheckbox(a, bdShowPoints)
	renderCheckbox(a, bdShowOutline)
	renderRBox(a, bdCurveType)
	renderRBox(a, bdCaseType)
	renderRBox(a, bdInnerJoin)
	renderRBox(a, bdLineJoin)
	renderRBox(a, bdLineCap)
}

// --- Mouse handlers ---

func handleBezierDivMouseDown(x, y float64) bool {
	if !bdInitialized {
		return false
	}
	if bdCurve1.OnMouseButtonDown(x, y) {
		return true
	}
	if bdAngleTol.OnMouseButtonDown(x, y) {
		return true
	}
	if bdApproxScale.OnMouseButtonDown(x, y) {
		return true
	}
	if bdCuspLimit.OnMouseButtonDown(x, y) {
		return true
	}
	if bdWidth.OnMouseButtonDown(x, y) {
		return true
	}
	if bdShowPoints.OnMouseButtonDown(x, y) {
		return true
	}
	if bdShowOutline.OnMouseButtonDown(x, y) {
		return true
	}
	if bdCurveType.OnMouseButtonDown(x, y) {
		return true
	}
	if bdCaseType.OnMouseButtonDown(x, y) {
		bdHandleCaseTypeChange()
		return true
	}
	if bdInnerJoin.OnMouseButtonDown(x, y) {
		return true
	}
	if bdLineJoin.OnMouseButtonDown(x, y) {
		return true
	}
	if bdLineCap.OnMouseButtonDown(x, y) {
		return true
	}
	return false
}

func handleBezierDivMouseMove(x, y float64) bool {
	if !bdInitialized {
		return false
	}
	if bdCurve1.OnMouseMove(x, y, true) {
		return true
	}
	if bdAngleTol.OnMouseMove(x, y, true) {
		return true
	}
	if bdApproxScale.OnMouseMove(x, y, true) {
		return true
	}
	if bdCuspLimit.OnMouseMove(x, y, true) {
		return true
	}
	if bdWidth.OnMouseMove(x, y, true) {
		return true
	}
	return false
}

func handleBezierDivMouseUp() {
	if !bdInitialized {
		return
	}
	bdCurve1.OnMouseButtonUp(0, 0)
	bdAngleTol.OnMouseButtonUp(0, 0)
	bdApproxScale.OnMouseButtonUp(0, 0)
	bdCuspLimit.OnMouseButtonUp(0, 0)
	bdWidth.OnMouseButtonUp(0, 0)
}
