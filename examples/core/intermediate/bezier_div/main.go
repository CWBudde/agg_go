// Port of AGG C++ bezier_div.cpp – Bezier curve subdivision accuracy demo.
//
// Shows a cubic Bezier curve rendered as a wide stroked shape together with
// the subdivision points. Default values from the WASM demo are used as
// constants; interactive sliders belong in the platform (SDL2/X11) variant.
//
// Default: subdivision mode, control points (170,424)(13,87)(488,423)(26,333),
// angle tolerance=15°, approx scale=1.0, stroke width=50.
package main

import (
	"fmt"
	"math"
	"time"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	bezierctrl "github.com/cwbudde/agg_go/internal/ctrl/bezier"
	checkboxctrl "github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	rboxctrl "github.com/cwbudde/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/curves"
	"github.com/cwbudde/agg_go/internal/demo/timing"
	"github.com/cwbudde/agg_go/internal/gsv"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
)

const (
	width  = 655
	height = 520
)

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters
// ---------------------------------------------------------------------------
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// Vertex-source adapters
// ---------------------------------------------------------------------------

// convVS adapts conv.VertexSource to rasterizer.VertexSource.
type convVS struct{ src conv.VertexSource }

func (a *convVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ellipseVS adapts shapes.Ellipse to rasterizer.VertexSource.
type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	var vx, vy float64
	cmd := ev.e.Vertex(&vx, &vy)
	*x, *y = vx, vy
	return uint32(cmd)
}

type curvePoint struct {
	x, y float64
	dist float64
}

type curve4Source interface {
	SetApproximationScale(float64)
	SetAngleTolerance(float64)
	SetCuspLimit(float64)
	Init(x1, y1, x2, y2, x3, y3, x4, y4 float64)
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct {
	curve1         *bezierctrl.BezierCtrl[color.RGBA]
	angleTolerance *sliderctrl.SliderCtrl
	approxScale    *sliderctrl.SliderCtrl
	cuspLimit      *sliderctrl.SliderCtrl
	width          *sliderctrl.SliderCtrl
	showPoints     *checkboxctrl.CheckboxCtrl[color.RGBA]
	showOutline    *checkboxctrl.CheckboxCtrl[color.RGBA]
	curveType      *rboxctrl.RboxCtrl[color.RGBA]
	caseType       *rboxctrl.RboxCtrl[color.RGBA]
	innerJoin      *rboxctrl.RboxCtrl[color.RGBA]
	lineJoin       *rboxctrl.RboxCtrl[color.RGBA]
	lineCap        *rboxctrl.RboxCtrl[color.RGBA]
	allCtrls       []ctrlbase.Ctrl[color.RGBA]
}

func newDemo() *demo {
	curve1 := bezierctrl.NewDefaultBezierCtrl()
	curve1.SetLineColor(color.RGBA{R: 0, G: 0.3, B: 0.5, A: 0.8})
	curve1.SetCurve(170, 424, 13, 87, 488, 423, 26, 333)

	angleTolerance := sliderctrl.NewSliderCtrl(5.0, 5.0, 240.0, 12.0, false)
	angleTolerance.SetLabel("Angle Tolerance=%.0f deg")
	angleTolerance.SetRange(0, 90)
	angleTolerance.SetValue(15)

	approxScale := sliderctrl.NewSliderCtrl(5.0, 17+5.0, 240.0, 17+12.0, false)
	approxScale.SetLabel("Approximation Scale=%.3f")
	approxScale.SetRange(0.1, 5)
	approxScale.SetValue(1.0)

	cuspLimit := sliderctrl.NewSliderCtrl(5.0, 17+17+5.0, 240.0, 17+17+12.0, false)
	cuspLimit.SetLabel("Cusp Limit=%.0f deg")
	cuspLimit.SetRange(0, 90)
	cuspLimit.SetValue(0)

	widthCtrl := sliderctrl.NewSliderCtrl(245.0, 5.0, 495.0, 12.0, false)
	widthCtrl.SetLabel("Width=%.2f")
	widthCtrl.SetRange(-50, 100)
	widthCtrl.SetValue(50.0)

	showPoints := checkboxctrl.NewDefaultCheckboxCtrl(250.0, 15+5, "Show Points", false)
	showPoints.SetChecked(true)

	showOutline := checkboxctrl.NewDefaultCheckboxCtrl(250.0, 30+5, "Show Stroke Outline", false)
	showOutline.SetChecked(true)

	curveType := rboxctrl.NewDefaultRboxCtrl(535.0, 5.0, 535.0+115.0, 55.0, false)
	curveType.AddItem("Incremental")
	curveType.AddItem("Subdiv")
	curveType.SetCurItem(1)

	caseType := rboxctrl.NewDefaultRboxCtrl(535.0, 60.0, 535.0+115.0, 195.0, false)
	caseType.SetTextSize(7, 0)
	caseType.SetTextThickness(1.0)
	caseType.AddItem("Random")
	caseType.AddItem("13---24")
	caseType.AddItem("Smooth Cusp 1")
	caseType.AddItem("Smooth Cusp 2")
	caseType.AddItem("Real Cusp 1")
	caseType.AddItem("Real Cusp 2")
	caseType.AddItem("Fancy Stroke")
	caseType.AddItem("Jaw")
	caseType.AddItem("Ugly Jaw")

	innerJoin := rboxctrl.NewDefaultRboxCtrl(535.0, 200.0, 535.0+115.0, 290.0, false)
	innerJoin.SetTextSize(8, 0)
	innerJoin.AddItem("Inner Bevel")
	innerJoin.AddItem("Inner Miter")
	innerJoin.AddItem("Inner Jag")
	innerJoin.AddItem("Inner Round")
	innerJoin.SetCurItem(3)

	lineJoinCtrl := rboxctrl.NewDefaultRboxCtrl(535.0, 295.0, 535.0+115.0, 385.0, false)
	lineJoinCtrl.SetTextSize(8, 0)
	lineJoinCtrl.AddItem("Miter Join")
	lineJoinCtrl.AddItem("Miter Revert")
	lineJoinCtrl.AddItem("Round Join")
	lineJoinCtrl.AddItem("Bevel Join")
	lineJoinCtrl.AddItem("Miter Round")
	lineJoinCtrl.SetCurItem(1)

	lineCapCtrl := rboxctrl.NewDefaultRboxCtrl(535.0, 395.0, 535.0+115.0, 455.0, false)
	lineCapCtrl.SetTextSize(8, 0)
	lineCapCtrl.AddItem("Butt Cap")
	lineCapCtrl.AddItem("Square Cap")
	lineCapCtrl.AddItem("Round Cap")
	lineCapCtrl.SetCurItem(0)

	d := &demo{
		curve1:         curve1,
		angleTolerance: angleTolerance,
		approxScale:    approxScale,
		cuspLimit:      cuspLimit,
		width:          widthCtrl,
		showPoints:     showPoints,
		showOutline:    showOutline,
		curveType:      curveType,
		caseType:       caseType,
		innerJoin:      innerJoin,
		lineJoin:       lineJoinCtrl,
		lineCap:        lineCapCtrl,
	}
	d.allCtrls = []ctrlbase.Ctrl[color.RGBA]{
		curve1, angleTolerance, approxScale, cuspLimit, widthCtrl,
		showPoints, showOutline, curveType, caseType,
		innerJoin, lineJoinCtrl, lineCapCtrl,
	}
	return d
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)

	// Light cream background.
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineU8()

	angleTol := d.angleTolerance.Value() * math.Pi / 180.0
	cuspLimitVal := d.cuspLimit.Value() * math.Pi / 180.0
	incremental := d.curveType.CurItem() == 0

	curvePath := d.buildCurvePath(d.approxScale.Value(), angleTol, cuspLimitVal, incremental)

	// Wide stroke from the curve.
	curveAdapter := path.NewPathStorageStlVertexSourceAdapter(curvePath)
	stroke := conv.NewConvStroke(curveAdapter)
	stroke.SetWidth(d.width.Value())
	stroke.SetLineJoin(basics.LineJoin(d.lineJoin.CurItem()))
	stroke.SetLineCap(basics.LineCap(d.lineCap.CurItem()))
	stroke.SetInnerJoin(basics.InnerJoin(d.innerJoin.CurItem()))
	stroke.SetInnerMiterLimit(1.01)

	// Fill the wide stroke (semi-transparent green).
	ras.Reset()
	ras.AddPath(&convVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 0, G: 128, B: 0, A: 128})

	// Subdivision points as small dots.
	if d.showPoints.IsChecked() {
		curvePath.Rewind(0)
		for {
			x, y, cmd := curvePath.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if basics.IsVertex(basics.PathCommand(cmd)) {
				dot := shapes.NewEllipseWithParams(x, y, 1.5, 1.5, 8, false)
				ras.Reset()
				ras.AddPath(&ellipseVS{e: dot}, 0)
				renscan.RenderScanlinesAASolid(ras, sl, mainRb,
					color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 128})
			}
		}
	}

	// Outline of the wide stroke (stroke of a stroke).
	if d.showOutline.IsChecked() {
		stroke2 := conv.NewConvStroke(stroke)
		stroke2.SetWidth(1.5)
		ras.Reset()
		ras.AddPath(&convVS{src: stroke2}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb,
			color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 128})
	}

	statsText := d.statsText(curvePath, angleTol, cuspLimitVal, incremental)
	stats := gsv.NewGSVText()
	stats.SetSize(8.0, 0)
	stats.SetStartPoint(10.0, 85.0)
	stats.SetText(statsText)

	statsStroke := gsv.NewGSVTextOutline(stats)
	statsStroke.SetWidth(1.5)
	ras.Reset()
	ras.AddPath(&convVS{src: statsStroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// Render all controls.
	for _, c := range d.allCtrls {
		renderCtrl(ras, sl, mainRb, c)
	}

	// Copy with y-flip (C++ uses flip_y=true).
	copyFlipY(workBuf, img.Data, w, h)
}

func (d *demo) buildCurvePath(approxScale, angleTol, cuspLimitVal float64, incremental bool) *path.PathStorageStl {
	curvePath := path.NewPathStorageStl()
	curve := d.newCurve(approxScale, angleTol, cuspLimitVal, incremental)
	curve.Rewind(0)
	for {
		x, y, cmd := curve.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			curvePath.MoveTo(x, y)
		} else if basics.IsVertex(cmd) {
			curvePath.LineTo(x, y)
		}
	}
	return curvePath
}

func (d *demo) newCurve(approxScale, angleTol, cuspLimitVal float64, incremental bool) curve4Source {
	var curve curve4Source
	if incremental {
		curve = curves.NewCurve4Inc()
	} else {
		curve = curves.NewCurve4Div()
	}
	curve.SetApproximationScale(approxScale)
	curve.SetAngleTolerance(angleTol)
	curve.SetCuspLimit(cuspLimitVal)
	d.initCurve(curve)
	return curve
}

func (d *demo) initCurve(curve curve4Source) {
	curve.Init(d.curve1.X1(), d.curve1.Y1(),
		d.curve1.X2(), d.curve1.Y2(),
		d.curve1.X3(), d.curve1.Y3(),
		d.curve1.X4(), d.curve1.Y4())
}

func (d *demo) statsText(curvePath *path.PathStorageStl, angleTol, cuspLimitVal float64, incremental bool) string {
	maxError01, maxAngleError01 := d.calcMaxError(0.01, angleTol, cuspLimitVal, incremental)
	maxError1, maxAngleError1 := d.calcMaxError(0.1, angleTol, cuspLimitVal, incremental)
	maxError10, maxAngleError10 := d.calcMaxError(1, angleTol, cuspLimitVal, incremental)
	maxError100, maxAngleError100 := d.calcMaxError(10, angleTol, cuspLimitVal, incremental)
	maxError1000, maxAngleError1000 := d.calcMaxError(100, angleTol, cuspLimitVal, incremental)

	header := fmt.Sprintf("Num Points=%d", countPathVertices(curvePath))
	if timing.ShowText() {
		header += fmt.Sprintf(" Time=%.2fmks", d.measureCurveTime(angleTol, cuspLimitVal, incremental))
	}

	return fmt.Sprintf(
		"%s\n\n"+
			" Dist Error: x0.01=%.5f x0.1=%.5f x1=%.5f x10=%.5f x100=%.5f\n\n"+
			"Angle Error: x0.01=%.1f x0.1=%.1f x1=%.1f x10=%.1f x100=%.1f",
		header,
		maxError01, maxError1, maxError10, maxError100, maxError1000,
		maxAngleError01, maxAngleError1, maxAngleError10, maxAngleError100, maxAngleError1000,
	)
}

func (d *demo) measureCurveTime(angleTol, cuspLimitVal float64, incremental bool) float64 {
	curve := d.newCurve(d.approxScale.Value(), angleTol, cuspLimitVal, incremental)
	start := time.Now()
	for i := 0; i < 100; i++ {
		d.initCurve(curve)
		curve.Rewind(0)
		for {
			_, _, cmd := curve.Vertex()
			if basics.IsStop(cmd) {
				break
			}
		}
	}
	return time.Since(start).Seconds() * 1_000_000.0 / 100.0
}

func (d *demo) calcMaxError(scale, angleTol, cuspLimitVal float64, incremental bool) (float64, float64) {
	curve := d.newCurve(d.approxScale.Value()*scale, angleTol, cuspLimitVal, incremental)

	var curvePoints []curvePoint
	curve.Rewind(0)
	for {
		x, y, cmd := curve.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsVertex(cmd) {
			curvePoints = append(curvePoints, curvePoint{x: x, y: y})
		}
	}
	if len(curvePoints) < 2 {
		return 0, 0
	}

	curveDist := 0.0
	for i := 1; i < len(curvePoints); i++ {
		curvePoints[i-1].dist = curveDist
		curveDist += basics.CalcDistance(
			curvePoints[i-1].x, curvePoints[i-1].y,
			curvePoints[i].x, curvePoints[i].y,
		)
	}
	curvePoints[len(curvePoints)-1].dist = curveDist

	const referenceCount = 4096
	referencePoints := make([]curvePoint, referenceCount)
	for i := range referencePoints {
		mu := float64(i) / float64(referenceCount-1)
		referencePoints[i].x, referencePoints[i].y = bezier4Point(
			d.curve1.X1(), d.curve1.Y1(),
			d.curve1.X2(), d.curve1.Y2(),
			d.curve1.X3(), d.curve1.Y3(),
			d.curve1.X4(), d.curve1.Y4(),
			mu,
		)
	}

	referenceDist := 0.0
	for i := 1; i < len(referencePoints); i++ {
		referencePoints[i-1].dist = referenceDist
		referenceDist += basics.CalcDistance(
			referencePoints[i-1].x, referencePoints[i-1].y,
			referencePoints[i].x, referencePoints[i].y,
		)
	}
	referencePoints[len(referencePoints)-1].dist = referenceDist

	maxError := 0.0
	for _, ref := range referencePoints {
		idx1, idx2 := findPoint(curvePoints, ref.dist)
		err := basics.CalcLinePointDistance(
			curvePoints[idx1].x, curvePoints[idx1].y,
			curvePoints[idx2].x, curvePoints[idx2].y,
			ref.x, ref.y,
		)
		if err > maxError {
			maxError = err
		}
	}

	maxAngleError := 0.0
	for i := 2; i < len(curvePoints); i++ {
		a1 := math.Atan2(curvePoints[i-1].y-curvePoints[i-2].y, curvePoints[i-1].x-curvePoints[i-2].x)
		a2 := math.Atan2(curvePoints[i].y-curvePoints[i-1].y, curvePoints[i].x-curvePoints[i-1].x)
		da := math.Abs(a1 - a2)
		if da >= math.Pi {
			da = 2*math.Pi - da
		}
		if da > maxAngleError {
			maxAngleError = da
		}
	}

	return maxError * scale, maxAngleError * 180.0 / math.Pi
}

func countPathVertices(p *path.PathStorageStl) int {
	count := 0
	p.Rewind(0)
	for {
		_, _, cmd := p.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		count++
	}
	return count
}

func findPoint(points []curvePoint, dist float64) (int, int) {
	i := 0
	j := len(points) - 1
	for j-i > 1 {
		k := (i + j) >> 1
		if dist < points[k].dist {
			j = k
		} else {
			i = k
		}
	}
	return i, j
}

func bezier4Point(x1, y1, x2, y2, x3, y3, x4, y4, mu float64) (float64, float64) {
	mum1 := 1 - mu
	mum13 := mum1 * mum1 * mum1
	mu3 := mu * mu * mu
	x := mum13*x1 + 3*mu*mum1*mum1*x2 + 3*mu*mu*mum1*x3 + mu3*x4
	y := mum13*y1 + 3*mu*mum1*mum1*y2 + 3*mu*mu*mum1*y3 + mu3*y4
	return x, y
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(height-y)
	if btn.Left {
		for _, c := range d.allCtrls {
			if c.OnMouseButtonDown(fx, fy) {
				return true
			}
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(height-y)
	for _, c := range d.allCtrls {
		if c.OnMouseMove(fx, fy, btn.Left) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(height-y)
	for _, c := range d.allCtrls {
		if c.OnMouseButtonUp(fx, fy) {
			return true
		}
	}
	return false
}

func renderCtrl(
	ras *rasType,
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVS{ctrl: ctrl}, uint32(pathID))
		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, color.RGBA8[color.Linear]{
			R: clampU8(c.R),
			G: clampU8(c.G),
			B: clampU8(c.B),
			A: clampU8(c.A),
		})
	}
}

type ctrlVS struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlVS) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func clampU8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255.0 + 0.5)
}

func copyFlipY(src, dst []uint8, width, height int) {
	stride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Bezier Div",
		Width:                 width,
		Height:                height,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
