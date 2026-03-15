package linepatterns

import (
	"sync"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/curves"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/primitives"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	outline "github.com/MeKo-Christian/agg_go/internal/renderer/outline"
)

type curveSourceAdapter struct {
	cv *curves.Curve4
}

func (a *curveSourceAdapter) Rewind(pathID uint32) {
	a.cv.Rewind(uint(pathID))
}

func (a *curveSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.cv.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type imagePatternSource struct {
	img PatternImage
}

func (s *imagePatternSource) Width() float64  { return float64(s.img.Width) }
func (s *imagePatternSource) Height() float64 { return float64(s.img.Height) }

func (s *imagePatternSource) Pixel(x, y int) color.RGBA {
	if x < 0 || y < 0 || x >= s.img.Width || y >= s.img.Height {
		return color.NewRGBA(0, 0, 0, 0)
	}
	p := s.img.Pixels[y*s.img.Width+x]
	r := uint8((p >> 16) & 0xFF)
	g := uint8((p >> 8) & 0xFF)
	b := uint8(p & 0xFF)
	a := BrightnessToAlpha(int(r) + int(g) + int(b))
	return color.NewRGBAFromRGBA8(r, g, b, a)
}

type lineImageBaseAdapter struct {
	pf *pixfmt.PixFmtBGR24
}

func rgbaToRGB8(c color.RGBA) color.RGB8[color.Linear] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGB8[color.Linear]{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B)}
}

func (a *lineImageBaseAdapter) BlendColorHSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	if a.pf == nil {
		return
	}
	for i := 0; i < length && i < len(colors); i++ {
		c := colors[i]
		a.pf.BlendPixel(x+i, y, rgbaToRGB8(c), uint8(c.A*255+0.5), basics.CoverFull)
	}
}

func (a *lineImageBaseAdapter) BlendColorVSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	if a.pf == nil {
		return
	}
	for i := 0; i < length && i < len(colors); i++ {
		c := colors[i]
		a.pf.BlendPixel(x, y+i, rgbaToRGB8(c), uint8(c.A*255+0.5), basics.CoverFull)
	}
}

type lineOutlineImageAdapter struct {
	ren *outline.RendererOutlineImage
}

func (a *lineOutlineImageAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *lineOutlineImageAdapter) Color(c color.RGBA8[color.Linear]) {}
func (a *lineOutlineImageAdapter) Pie(x, y, x1, y1, x2, y2 int)      { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *lineOutlineImageAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}
func (a *lineOutlineImageAdapter) Line0(lp primitives.LineParameters) { a.ren.Line0(&lp) }
func (a *lineOutlineImageAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}
func (a *lineOutlineImageAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}
func (a *lineOutlineImageAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

type curveDef struct {
	x1, y1 float64
	x2, y2 float64
	x3, y3 float64
	x4, y4 float64
}

type Curve struct {
	X1, Y1 float64
	X2, Y2 float64
	X3, Y3 float64
	X4, Y4 float64
}

var linePatternCurves = []curveDef{
	{64, 19, 14, 126, 118, 266, 19, 265},
	{112, 113, 178, 32, 200, 132, 125, 438},
	{401, 24, 326, 149, 285, 11, 177, 77},
	{188, 427, 129, 295, 19, 283, 25, 410},
	{451, 346, 302, 218, 265, 441, 459, 400},
	{454, 198, 14, 13, 220, 291, 483, 283},
	{301, 398, 355, 231, 209, 211, 170, 353},
	{484, 101, 222, 33, 486, 435, 487, 138},
	{143, 147, 11, 45, 83, 427, 132, 197},
}

func DefaultCurves() []Curve {
	curves := make([]Curve, len(linePatternCurves))
	for i, c := range linePatternCurves {
		curves[i] = Curve{
			X1: c.x1, Y1: c.y1,
			X2: c.x2, Y2: c.y2,
			X3: c.x3, Y3: c.y3,
			X4: c.x4, Y4: c.y4,
		}
	}
	return curves
}

var (
	preparedLinePatternOnce     sync.Once
	preparedLinePatternPatterns []outline.Pattern
)

func prepareLinePatternResources() {
	preparedLinePatternPatterns = make([]outline.Pattern, len(Images))
	for i := range Images {
		filter := outline.NewPatternFilterRGBAAdapter()
		pattern := outline.NewLineImagePattern(filter)
		pattern.Create(&imagePatternSource{img: Images[i]})
		preparedLinePatternPatterns[i] = pattern
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func Draw(img *agg.Image, scaleX, startX float64) {
	preparedLinePatternOnce.Do(prepareLinePatternResources)
	curves := DefaultCurves()
	drawCurves(img, scaleX, startX, curves)
}

func DrawCurves(img *agg.Image, scaleX, startX float64, curves []Curve) {
	preparedLinePatternOnce.Do(prepareLinePatternResources)

	if len(curves) == 0 {
		drawCurves(img, scaleX, startX, DefaultCurves())
		return
	}
	drawCurves(img, scaleX, startX, curves)
}

func drawCurves(img *agg.Image, scaleX, startX float64, curvesData []Curve) {
	preparedLinePatternOnce.Do(prepareLinePatternResources)

	rgbData := make([]uint8, img.Width()*img.Height()*3)
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(rgbData, img.Width(), img.Height(), img.Width()*3)
	pf := pixfmt.NewPixFmtBGR24(rbuf)
	pf.Clear(color.RGB8[color.Linear]{R: 255, G: 255, B: 242})

	baseAdapter := &lineImageBaseAdapter{pf: pf}
	renImg := outline.NewRendererOutlineImage(baseAdapter, preparedLinePatternPatterns[0])
	rasImg := rasterizer.NewRasterizerOutlineAA[*lineOutlineImageAdapter, color.RGBA8[color.Linear]](&lineOutlineImageAdapter{ren: renImg})

	clampedScaleX := clamp(scaleX, 0.2, 3.0)
	clampedStartX := clamp(startX, 0.0, 10.0)
	for i, c := range curvesData {
		cv := curves.NewCurve4()
		cv.SetApproximationScale(1.0)
		cv.Init(c.X1, c.Y1, c.X2, c.Y2, c.X3, c.Y3, c.X4, c.Y4)
		renImg.SetPattern(preparedLinePatternPatterns[i%len(preparedLinePatternPatterns)])
		renImg.SetScaleX(clampedScaleX)
		renImg.SetStartX(clampedStartX)
		rasImg.AddPath(&curveSourceAdapter{cv: cv}, 0)
	}

	for y := 0; y < img.Height(); y++ {
		srcOff := y * img.Width() * 3
		dstOff := y * img.Width() * 4
		for x := 0; x < img.Width(); x++ {
			b := rgbData[srcOff+x*3+0]
			g := rgbData[srcOff+x*3+1]
			r := rgbData[srcOff+x*3+2]
			img.Data[dstOff+x*4+0] = r
			img.Data[dstOff+x*4+1] = g
			img.Data[dstOff+x*4+2] = b
			img.Data[dstOff+x*4+3] = 255
		}
	}
}
