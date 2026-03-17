// Package idea ports AGG's idea.cpp demo.
package idea

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	icolor "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/gamma"
)

const (
	baseWidth  = 250.0
	baseHeight = 280.0
)

type State struct {
	Angle      float64
	Rotate     bool
	EvenOdd    bool
	Draft      bool
	Roundoff   bool
	AngleDelta float64
}

type pathAttributes struct {
	fill        agg.Color
	stroke      agg.Color
	strokeWidth float64
	contours    []contour
}

type contour []point

type point struct {
	x, y float64
}

func DefaultState() State {
	return State{
		AngleDelta: 0.01,
	}
}

func (s *State) Clamp() {
	if s.AngleDelta < 0 {
		s.AngleDelta = 0
	}
	if s.AngleDelta > 10 {
		s.AngleDelta = 10
	}
	for s.Angle >= 360.0 {
		s.Angle -= 360.0
	}
	for s.Angle < 0 {
		s.Angle += 360.0
	}
}

func (s *State) Advance() {
	if !s.Rotate {
		return
	}
	s.Angle += s.AngleDelta
	s.Clamp()
}

// ctrlIface is the minimal vertex-source interface shared by checkbox and slider controls.
type ctrlIface interface {
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) icolor.RGBA
}

// ctrlPathAdapter bridges ctrlIface to the rasterizer's vertex-source interface.
type ctrlPathAdapter struct {
	ctrl ctrlIface
}

func (a *ctrlPathAdapter) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }
func (a *ctrlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// linearToSRGB converts a linear float color component to an sRGB-encoded uint8,
// matching the conversion C++ AGG applies when constructing srgba8 from rgba.
func linearToSRGB(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	var s float64
	if v <= 0.0031308 {
		s = 12.92 * v
	} else {
		s = 1.055*math.Pow(v, 1.0/2.4) - 0.055
	}
	return uint8(s*255 + 0.5)
}

func renderCtrl(ag *agg.Agg2D, ctrl ctrlIface) {
	ras := ag.GetInternalRasterizer()
	adapter := &ctrlPathAdapter{ctrl: ctrl}
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		c := ctrl.Color(pathID)
		// Alpha stays linear (sRGB does not gamma-encode alpha).
		var a uint8
		if c.A <= 0 {
			a = 0
		} else if c.A >= 1 {
			a = 255
		} else {
			a = uint8(c.A*255 + 0.5)
		}
		ag.RenderRasterizerWithColor(agg.NewColor(linearToSRGB(c.R), linearToSRGB(c.G), linearToSRGB(c.B), a))
	}
}

func Draw(ctx *agg.Context, st State) {
	st.Clamp()
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()

	// Render UI controls (matching C++ on_draw initial render_ctrl calls).
	rotateCb := checkbox.NewDefaultCheckboxCtrl(10, 3, "Rotate", false)
	rotateCb.SetChecked(st.Rotate)
	evenOddCb := checkbox.NewDefaultCheckboxCtrl(60, 3, "Even-Odd", false)
	evenOddCb.SetChecked(st.EvenOdd)
	draftCb := checkbox.NewDefaultCheckboxCtrl(130, 3, "Draft", false)
	draftCb.SetChecked(st.Draft)
	roundoffCb := checkbox.NewDefaultCheckboxCtrl(175, 3, "Roundoff", false)
	roundoffCb.SetChecked(st.Roundoff)
	angleSlider := slider.NewSliderCtrl(10, 21, 240, 27, false)
	angleSlider.SetLabel("Step=%4.3f degree")
	angleSlider.SetValue(st.AngleDelta)
	for _, c := range []ctrlIface{rotateCb, evenOddCb, draftCb, roundoffCb, angleSlider} {
		renderCtrl(a, c)
	}

	scale, offX, offY := fitFrame(ctx.Width(), ctx.Height())
	angle := st.Angle * math.Pi / 180.0
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	centerX := offX + (baseWidth*0.5)*scale
	centerY := offY + (baseHeight*0.5+10.0)*scale

	a.FillEvenOdd(st.EvenOdd)

	ras := a.GetInternalRasterizer()
	if st.Draft {
		ras.SetGamma(gamma.NewGammaThreshold(0.4).Apply)
	} else {
		ras.SetGamma(func(x float64) float64 { return x })
	}

	for _, path := range paths {
		a.FillColor(path.fill)
		a.NoLine()
		a.ResetPath()
		for _, c := range path.contours {
			addContour(a, c, cosA, sinA, centerX, centerY, scale, st.Roundoff)
		}
		a.DrawPath(agg.FillOnly)

		if path.strokeWidth <= 0 {
			continue
		}
		a.NoFill()
		a.LineColor(path.stroke)
		a.LineWidth(path.strokeWidth * scale)
		a.ResetPath()
		for _, c := range path.contours {
			addContour(a, c, cosA, sinA, centerX, centerY, scale, st.Roundoff)
		}
		a.DrawPath(agg.StrokeOnly)
	}

	ras.SetGamma(func(x float64) float64 { return x })
	a.FillEvenOdd(false)
}

func addContour(a *agg.Agg2D, c contour, cosA, sinA, centerX, centerY, scale float64, roundoff bool) {
	if len(c) == 0 {
		return
	}
	x, y := transformPoint(c[0], cosA, sinA, centerX, centerY, scale, roundoff)
	a.MoveTo(x, y)
	for i := 1; i < len(c); i++ {
		x, y = transformPoint(c[i], cosA, sinA, centerX, centerY, scale, roundoff)
		a.LineTo(x, y)
	}
	a.ClosePolygon()
}

func transformPoint(p point, cosA, sinA, centerX, centerY, scale float64, roundoff bool) (float64, float64) {
	x := p.x*cosA - p.y*sinA
	y := p.x*sinA + p.y*cosA
	x = centerX + x*scale
	y = centerY + y*scale
	if roundoff {
		x = math.Floor(x + 0.5)
		y = math.Floor(y + 0.5)
	}
	return x, y
}

func fitFrame(w, h int) (scale, offX, offY float64) {
	sx := float64(w) / baseWidth
	sy := float64(h) / baseHeight
	scale = math.Min(sx, sy)
	if scale > 1.0 {
		scale = 1.0
	}
	if scale <= 0 {
		scale = 1.0
	}
	offX = (float64(w) - baseWidth*scale) * 0.5
	offY = (float64(h) - baseHeight*scale) * 0.5
	return scale, offX, offY
}

var paths = []pathAttributes{
	{
		fill:        agg.NewColor(255, 255, 0, 255),
		stroke:      agg.Black,
		strokeWidth: 1.0,
		contours: []contour{
			pointsFromCoords(bulbCoords),
		},
	},
	{
		fill:        agg.NewColor(255, 255, 200, 255),
		stroke:      agg.NewColor(90, 0, 0, 255),
		strokeWidth: 0.7,
		contours: []contour{
			pointsFromCoords(beam1Coords),
			pointsFromCoords(beam2Coords),
			pointsFromCoords(beam3Coords),
			pointsFromCoords(beam4Coords),
		},
	},
	{
		fill: agg.Black,
		contours: []contour{
			pointsFromCoords(fig1Coords),
			pointsFromCoords(fig2Coords),
			pointsFromCoords(fig3Coords),
			pointsFromCoords(fig4Coords),
			pointsFromCoords(fig5Coords),
			pointsFromCoords(fig6Coords),
		},
	},
}

func pointsFromCoords(coords []float64) contour {
	out := make(contour, 0, len(coords)/2)
	for i := 0; i+1 < len(coords); i += 2 {
		out = append(out, point{x: coords[i], y: coords[i+1]})
	}
	return out
}

var bulbCoords = []float64{
	-6, -67, -6, -71, -7, -74, -8, -76, -10, -79,
	-10, -82, -9, -84, -6, -86, -4, -87, -2, -86,
	-1, -86, 1, -84, 2, -82, 2, -79, 0, -77,
	-2, -73, -2, -71, -2, -69, -3, -67, -4, -65,
}

var beam1Coords = []float64{
	-14, -84, -22, -85, -23, -87, -22, -88, -21, -88,
}

var beam2Coords = []float64{
	-10, -92, -14, -96, -14, -98, -12, -99, -11, -97,
}

var beam3Coords = []float64{
	-1, -92, -2, -98, 0, -100, 2, -100, 1, -98,
}

var beam4Coords = []float64{
	5, -89, 11, -94, 13, -93, 13, -92, 12, -91,
}

var fig1Coords = []float64{
	1, -48, -3, -54, -7, -58, -12, -58, -17, -55, -20, -52, -21, -47,
	-20, -40, -17, -33, -11, -28, -6, -26, -2, -25, 2, -26, 4, -28, 5,
	-33, 5, -39, 3, -44, 12, -48, 12, -50, 12, -51, 3, -46,
}

var fig2Coords = []float64{
	11, -27, 6, -23, 4, -22, 3, -19, 5,
	-16, 6, -15, 11, -17, 19, -23, 25, -30, 32, -38, 32, -41, 32, -50, 30, -64, 32, -72,
	32, -75, 31, -77, 28, -78, 26, -80, 28, -87, 27, -89, 25, -88, 24, -79, 24, -76, 23,
	-75, 20, -76, 17, -76, 17, -74, 19, -73, 22, -73, 24, -71, 26, -69, 27, -64, 28, -55,
	28, -47, 28, -40, 26, -38, 20, -33, 14, -30,
}

var fig3Coords = []float64{
	-6, -20, -9, -21, -15, -21, -20, -17,
	-28, -8, -32, -1, -32, 1, -30, 6, -26, 8, -20, 10, -16, 12, -14, 14, -15, 16, -18, 20,
	-22, 20, -25, 19, -27, 20, -26, 22, -23, 23, -18, 23, -14, 22, -11, 20, -10, 17, -9, 14,
	-11, 11, -16, 9, -22, 8, -26, 5, -28, 2, -27, -2, -23, -8, -19, -11, -12, -14, -6, -15,
	-6, -18,
}

var fig4Coords = []float64{
	11, -6, 8, -16, 5, -21, -1, -23, -7,
	-22, -10, -17, -9, -10, -8, 0, -8, 10, -10, 18, -11, 22, -10, 26, -7, 28, -3, 30, 0, 31,
	5, 31, 10, 27, 14, 18, 14, 11, 11, 2,
}

var fig5Coords = []float64{
	0, 22, -5, 21, -8, 22, -9, 26, -8, 49,
	-8, 54, -10, 64, -10, 75, -9, 81, -10, 84, -16, 89, -18, 95, -18, 97, -13, 100, -12, 99,
	-12, 95, -10, 90, -8, 87, -6, 86, -4, 83, -3, 82, -5, 80, -6, 79, -7, 74, -6, 63, -3, 52,
	0, 42, 1, 31,
}

var fig6Coords = []float64{
	12, 31, 12, 24, 8, 21, 3, 21, 2, 24, 3,
	30, 5, 40, 8, 47, 10, 56, 11, 64, 11, 71, 10, 76, 8, 77, 8, 79, 10, 81, 13, 82, 17, 82, 26,
	84, 28, 87, 32, 86, 33, 81, 32, 80, 25, 79, 17, 79, 14, 79, 13, 76, 14, 72, 14, 64, 13, 55,
	12, 44, 12, 34,
}
