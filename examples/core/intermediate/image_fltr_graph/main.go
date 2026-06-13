// Port of AGG C++ image_fltr_graph.cpp demo.
//
// It compares interpolation filter shapes by plotting:
//   - raw filter function (red),
//   - unnormalized discrete sum response (green),
//   - normalized LUT weights (blue).
//
// Checkbox controls select which filters to display; a radius slider
// controls the variable-radius filters (sinc, lanczos, blackman).
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	icol "github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	checkboxctrl "github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/path"
)

const (
	frameWidth  = 780
	frameHeight = 300
	numFilters  = 16
)

// filterInfo describes a named filter and whether it has a variable radius.
type filterInfo struct {
	name     string
	variable bool
}

var filterNames = [numFilters]filterInfo{
	{"bilinear", false},
	{"bicubic ", false},
	{"spline16", false},
	{"spline36", false},
	{"hanning ", false},
	{"hamming ", false},
	{"hermite ", false},
	{"kaiser  ", false},
	{"quadric ", false},
	{"catrom  ", false},
	{"gaussian", false},
	{"bessel  ", false},
	{"mitchell", false},
	{"sinc    ", true},
	{"lanczos ", true},
	{"blackman", true},
}

// -- filter evaluation adaptor (mirrors C++ filter_base / adaptor) -----------

type filterEval interface {
	image.FilterFunction
	SetRadius(r float64)
}

type constFilter struct{ fn image.FilterFunction }

func (f *constFilter) Radius() float64              { return f.fn.Radius() }
func (f *constFilter) SetRadius(float64)            {}
func (f *constFilter) CalcWeight(x float64) float64 { return f.fn.CalcWeight(math.Abs(x)) }

type varFilter struct {
	factory func(r float64) image.FilterFunction
	fn      image.FilterFunction
}

func newVarFilter(factory func(r float64) image.FilterFunction) *varFilter {
	f := &varFilter{factory: factory}
	f.SetRadius(2.0)
	return f
}

func (f *varFilter) Radius() float64              { return f.fn.Radius() }
func (f *varFilter) SetRadius(r float64)          { f.fn = f.factory(r) }
func (f *varFilter) CalcWeight(x float64) float64 { return f.fn.CalcWeight(math.Abs(x)) }

func newFilter(index int, radius float64) filterEval {
	switch index {
	case 0:
		return &constFilter{fn: image.BilinearFilter{}}
	case 1:
		return &constFilter{fn: image.BicubicFilter{}}
	case 2:
		return &constFilter{fn: image.Spline16Filter{}}
	case 3:
		return &constFilter{fn: image.Spline36Filter{}}
	case 4:
		return &constFilter{fn: image.HanningFilter{}}
	case 5:
		return &constFilter{fn: image.HammingFilter{}}
	case 6:
		return &constFilter{fn: image.HermiteFilter{}}
	case 7:
		return &constFilter{fn: image.NewKaiserFilter(6.33)}
	case 8:
		return &constFilter{fn: image.QuadricFilter{}}
	case 9:
		return &constFilter{fn: image.CatromFilter{}}
	case 10:
		return &constFilter{fn: image.GaussianFilter{}}
	case 11:
		return &constFilter{fn: image.BesselFilter{}}
	case 12:
		return &constFilter{fn: image.NewMitchellFilter(1.0/3.0, 1.0/3.0)}
	case 13:
		f := newVarFilter(func(r float64) image.FilterFunction { return image.NewSincFilter(r) })
		f.SetRadius(radius)
		return f
	case 14:
		f := newVarFilter(func(r float64) image.FilterFunction { return image.NewLanczosFilter(r) })
		f.SetRadius(radius)
		return f
	case 15:
		f := newVarFilter(func(r float64) image.FilterFunction { return image.NewBlackmanFilter(r) })
		f.SetRadius(radius)
		return f
	default:
		return &constFilter{fn: image.BilinearFilter{}}
	}
}

// -- demo struct --------------------------------------------------------------

type demo struct {
	checkboxes [numFilters]*checkboxctrl.CheckboxCtrl[icol.RGBA]
	radius     *sliderctrl.SliderCtrl
	controls   []ctrlbase.Ctrl[icol.RGBA]
}

func newDemo() *demo {
	d := &demo{}

	// Create checkbox controls — matching C++ positions exactly.
	// C++ uses !flip_y → false for control flipY.
	for i := 0; i < numFilters; i++ {
		cb := checkboxctrl.NewDefaultCheckboxCtrl(
			8.0, 30.0+15.0*float64(i),
			filterNames[i].name, false,
		)
		d.checkboxes[i] = cb
	}

	// Radius slider — matches C++ m_radius(5.0, 5.0, 780-5, 10.0, !flip_y).
	d.radius = sliderctrl.NewSliderCtrl(5.0, 5.0, frameWidth-5, 10.0, false)
	d.radius.SetRange(2.0, 8.0)
	d.radius.SetValue(4.0)
	d.radius.SetLabel("Radius=%.3f")

	// Collect all controls for event dispatching and rendering.
	d.controls = make([]ctrlbase.Ctrl[icol.RGBA], 0, numFilters+1)
	for i := 0; i < numFilters; i++ {
		d.controls = append(d.controls, d.checkboxes[i])
	}
	d.controls = append(d.controls, d.radius)

	return d
}

// -- vertex-source adapter for controls → rasterizer -------------------------

type ctrlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }
func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func toAggColor(c icol.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255.0 + 0.5)
	}
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
}

// -- rendering ----------------------------------------------------------------

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()

	xStart := 125.0
	xEnd := float64(frameWidth) - 15.0
	yStart := 10.0
	yEnd := float64(frameHeight) - 10.0
	xCenter := (xStart + xEnd) / 2.0

	// -- vertical reference grid + centre axis (same as C++) --
	for i := 0; i <= 16; i++ {
		x := xStart + (xEnd-xStart)*float64(i)/16.0
		alpha := uint8(100)
		if i == 8 {
			alpha = 255
		}
		strokeLine(a, 1.0, agg.NewColor(0, 0, 0, alpha), x+0.5, yStart, x+0.5, yEnd)
	}

	ys := yStart + (yEnd-yStart)/6.0
	strokeLine(a, 1.0, agg.NewColor(0, 0, 0, 255), xStart, ys, xEnd, ys)

	// -- per-filter graph curves --
	for i := 0; i < numFilters; i++ {
		if !d.checkboxes[i].IsChecked() {
			continue
		}

		filter := newFilter(i, d.radius.Value())
		radius := filter.Radius()
		n := int(radius * 256 * 2)
		if n < 2 {
			n = 2
		}
		dy := yEnd - ys
		xs := (xEnd+xStart)/2.0 - (radius * (xEnd - xStart) / 16.0)
		dx := (xEnd - xStart) * radius / 8.0

		// (1) Raw continuous filter function — red.
		{
			ps := path.NewPathStorageStl()
			ps.MoveTo(xs+0.5, ys+dy*filter.CalcWeight(-radius))
			for j := 1; j < n; j++ {
				px := xs + dx*float64(j)/float64(n) + 0.5
				py := ys + dy*filter.CalcWeight(float64(j)/256.0-radius)
				ps.LineTo(px, py)
			}
			strokePath(a, 1.5, agg.NewColor(128, 0, 0, 255), ps)
		}

		// (2) Unnormalized discrete sum response — green.
		// The || condition is intentionally kept for parity with C++.
		{
			ir := int(math.Ceil(radius) + 0.1)
			ps := path.NewPathStorageStl()
			for xint := 0; xint < 256; xint++ {
				sum := 0.0
				for xfract := -ir; xfract < ir; xfract++ {
					xf := float64(xint)/256.0 + float64(xfract)
					if xf >= -radius || xf <= radius {
						sum += filter.CalcWeight(xf)
					}
				}
				px := xCenter + ((-128.0+float64(xint))/128.0)*radius*(xEnd-xStart)/16.0
				py := ys + sum*256.0 - 256.0
				if xint == 0 {
					ps.MoveTo(px, py)
				} else {
					ps.LineTo(px, py)
				}
			}
			strokePath(a, 1.5, agg.NewColor(0, 128, 0, 255), ps)
		}

		// (3) Normalized LUT weights — blue.
		{
			lut := image.NewImageFilterLUTWithFilter(filter, true)
			weights := lut.WeightArray()
			xsLUT := (xEnd+xStart)/2.0 - (float64(lut.Diameter()) * (xEnd - xStart) / 32.0)
			nn := lut.Diameter() * 256
			ps := path.NewPathStorageStl()
			ps.MoveTo(xsLUT+0.5, ys+dy*float64(weights[0])/float64(image.ImageFilterScale))
			for j := 1; j < nn; j++ {
				px := xsLUT + dx*float64(j)/float64(n) + 0.5
				py := ys + dy*float64(weights[j])/float64(image.ImageFilterScale)
				ps.LineTo(px, py)
			}
			strokePath(a, 1.5, agg.NewColor(0, 0, 128, 255), ps)
		}
	}

	// -- render controls --
	for i := 0; i < numFilters; i++ {
		renderCtrl(a, d.checkboxes[i])
	}
	// Only show the radius slider when a variable-radius filter is active.
	if d.checkboxes[13].IsChecked() || d.checkboxes[14].IsChecked() || d.checkboxes[15].IsChecked() {
		renderCtrl(a, d.radius)
	}
}

// strokeLine draws a thin stroked line via the low-level pipeline.
//
// C++ draws the grid/axis lines with a raw agg::conv_stroke, whose default cap
// is butt. Agg2D defaults to a round cap, which would deposit AA coverage one
// row past each butt endpoint; force butt to match the C++ output exactly.
func strokeLine(a *agg.Agg2D, width float64, clr agg.Color, x1, y1, x2, y2 float64) {
	a.ResetPath()
	a.MoveTo(x1, y1)
	a.LineTo(x2, y2)
	a.NoFill()
	a.LineColor(clr)
	a.LineWidth(width)
	a.LineCap(agg.CapButt)
	a.DrawPath(agg.StrokeOnly)
}

// strokePath strokes a PathStorageStl through the rasterizer.
func strokePath(a *agg.Agg2D, width float64, clr agg.Color, ps *path.PathStorageStl) {
	adapter := path.NewPathStorageStlVertexSourceAdapter(ps)
	stroke := conv.NewConvStroke(adapter)
	stroke.SetWidth(width)

	ras := a.GetInternalRasterizer()
	ras.Reset()
	stroke.Rewind(0)
	for {
		x, y, cmd := stroke.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		ras.AddVertex(x, y, uint32(cmd))
	}
	a.RenderRasterizerWithColor(clr)
}

// -- mouse handling -----------------------------------------------------------

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	for _, c := range d.controls {
		if c.OnMouseButtonDown(fx, fy) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false
	for _, c := range d.controls {
		if c.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false
	for _, c := range d.controls {
		if c.OnMouseButtonUp(fx, fy) {
			redraw = true
		}
	}
	return redraw
}

// -- main ---------------------------------------------------------------------

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Image filters' shape comparison",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
