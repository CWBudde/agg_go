package linepatterns

import (
	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/curves"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/primitives"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	outline "github.com/MeKo-Christian/agg_go/internal/renderer/outline"
)

type pathSourceAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathSourceAdapter) Rewind(pathID uint32) {
	a.ps.Rewind(uint(pathID))
}

func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x, *y = vx, vy
	return cmd
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
	brightness := (int(r) + int(g) + int(b)) / 3
	a := 255 - brightness
	if a < 0 {
		a = 0
	}
	c := color.NewRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
	c.Premultiply()
	return c
}

type lineImageBaseAdapter struct {
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]]
}

func rgbaToRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA8[color.Linear]{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

func (a *lineImageBaseAdapter) BlendColorHSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]color.RGBA8[color.Linear], len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorHspan(x, y, length, buf, nil, basics.CoverFull)
}

func (a *lineImageBaseAdapter) BlendColorVSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]color.RGBA8[color.Linear], len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorVspan(x, y, length, buf, nil, basics.CoverFull)
}

type lineOutlineImageAdapter struct {
	ren *outline.RendererOutlineImage
}

func (a *lineOutlineImageAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *lineOutlineImageAdapter) Color(c color.RGBA8[color.Linear]) {}

func lineNormals(lp primitives.LineParameters) (sx, sy, ex, ey int) {
	sx = lp.X1 + (lp.Y2 - lp.Y1)
	sy = lp.Y1 - (lp.X2 - lp.X1)
	ex = lp.X2 + (lp.Y2 - lp.Y1)
	ey = lp.Y2 - (lp.X2 - lp.X1)
	return
}

func (a *lineOutlineImageAdapter) Line0(lp primitives.LineParameters) {
	sx, sy, ex, ey := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *lineOutlineImageAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	_, _, ex, ey := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *lineOutlineImageAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	sx, sy, _, _ := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}
func (a *lineOutlineImageAdapter) Pie(x, y, x1, y1, x2, y2 int)                 {}
func (a *lineOutlineImageAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {}
func (a *lineOutlineImageAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

type curveDef struct {
	x1, y1 float64
	x2, y2 float64
	x3, y3 float64
	x4, y4 float64
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

func bezierPolyline(c curveDef) *path.PathStorageStl {
	cv := curves.NewCurve4Div()
	cv.Init(c.x1, c.y1, c.x2, c.y2, c.x3, c.y3, c.x4, c.y4)
	ps := path.NewPathStorageStl()
	for {
		x, y, cmd := cv.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			ps.MoveTo(x, y)
		} else {
			ps.LineTo(x, y)
		}
	}
	return ps
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
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]](pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	filter := outline.NewPatternFilterRGBAAdapter()
	pattern := outline.NewLineImagePattern(filter)
	renImg := outline.NewRendererOutlineImage(&lineImageBaseAdapter{renBase: renBase}, pattern)
	renImg.SetScaleX(clamp(scaleX, 0.2, 3.0))
	renImg.SetStartX(clamp(startX, 0.0, 10.0))
	rasImg := rasterizer.NewRasterizerOutlineAA[*lineOutlineImageAdapter, color.RGBA8[color.Linear]](&lineOutlineImageAdapter{ren: renImg})

	for i, curve := range linePatternCurves {
		src := &imagePatternSource{img: Images[i%len(Images)]}
		pattern.Create(src)
		rasImg.AddPath(&pathSourceAdapter{ps: bezierPolyline(curve)}, 0)
	}
}
