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
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	bezierctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/bezier"
	checkboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/curves"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
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

	// Build the curve using the selected method.
	curve := curves.NewCurve4Div()
	curve.SetApproximationScale(d.approxScale.Value())
	curve.SetAngleTolerance(angleTol)
	curve.SetCuspLimit(cuspLimitVal)
	curve.Init(d.curve1.X1(), d.curve1.Y1(),
		d.curve1.X2(), d.curve1.Y2(),
		d.curve1.X3(), d.curve1.Y3(),
		d.curve1.X4(), d.curve1.Y4())

	// Collect subdivision points.
	curvePath := path.NewPathStorageStl()
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

	// Render all controls.
	for _, c := range d.allCtrls {
		renderCtrl(ras, sl, mainRb, c)
	}

	// Copy with y-flip (C++ uses flip_y=true).
	copyFlipY(workBuf, img.Data, w, h)
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
		Title:  "Bezier Div",
		Width:  width,
		Height: height,
	}, newDemo())
}
