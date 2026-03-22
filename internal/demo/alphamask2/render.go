package alphamask2

import (
	"math"
	"sync"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/primitives"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/renderer/markers"
	outline "github.com/MeKo-Christian/agg_go/internal/renderer/outline"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

type Config struct {
	NumEllipses int
	Angle       float64
	Scale       float64
	SkewX       float64
	SkewY       float64
}

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	cmd := ev.e.Vertex(x, y)
	return uint32(cmd)
}

type outlineBlendBase interface {
	Width() int
	Height() int
	BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u)
	BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u)
}

type outlineBaseAdapter struct {
	renBase outlineBlendBase
}

func (a *outlineBaseAdapter) Width() int  { return a.renBase.Width() }
func (a *outlineBaseAdapter) Height() int { return a.renBase.Height() }

func (a *outlineBaseAdapter) BlendSolidHSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidHspan(x, y, length, c, convCovers)
}

func (a *outlineBaseAdapter) BlendSolidVSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidVspan(x, y, length, c, convCovers)
}

type outlineRenderer interface {
	AccurateJoinOnly() bool
	Color(c color.RGBA8[color.Linear])
	Line0(lp *primitives.LineParameters)
	Line1(lp *primitives.LineParameters, sx, sy int)
	Line2(lp *primitives.LineParameters, ex, ey int)
	Line3(lp *primitives.LineParameters, sx, sy, ex, ey int)
	Pie(x, y, x1, y1, x2, y2 int)
	Semidot(cmp func(int) bool, x, y, x1, y1 int)
}

type outlineAAAdapter struct {
	ren outlineRenderer
}

func (a *outlineAAAdapter) AccurateJoinOnly() bool             { return a.ren.AccurateJoinOnly() }
func (a *outlineAAAdapter) Color(c color.RGBA8[color.Linear])  { a.ren.Color(c) }
func (a *outlineAAAdapter) Line0(lp primitives.LineParameters) { a.ren.Line0(&lp) }
func (a *outlineAAAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}
func (a *outlineAAAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}
func (a *outlineAAAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}
func (a *outlineAAAdapter) Pie(x, y, x1, y1, x2, y2 int) { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *outlineAAAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRandSeed1() *clibcRand {
	return &clibcRand{
		state: [31]int32{
			-1726662223, 379960547, 1735697613, 1040273694, 1313901226,
			1627687941, -179304937, -2073333483, 1780058412, -1989503057,
			-615974602, 344556628, 939512070, -1249116260, 1507946756,
			-812545463, 154635395, 1388815473, -1926676823, 525320961,
			-1009028674, 968117788, -123449607, 1284210865, 435012392,
			-2017506339, -911064859, -370259173, 1132637927, 1398500161, -205601318,
		},
		fptr: 3,
		rptr: 0,
	}
}

func (r *clibcRand) next() int32 {
	r.state[r.fptr] += r.state[r.rptr]
	result := int32(uint32(r.state[r.fptr]) >> 1)
	r.fptr++
	if r.fptr >= 31 {
		r.fptr = 0
	}
	r.rptr++
	if r.rptr >= 31 {
		r.rptr = 0
	}
	return result
}

func (r *clibcRand) randN(n int) int      { return int(r.next()) % n }
func (r *clibcRand) randAnd(mask int) int { return int(r.next()) & mask }

var (
	lionOnce   sync.Once
	lionData   liondemo.LionData
	lionBaseDX float64
	lionBaseDY float64
)

func initLion() {
	lionData = liondemo.Parse()
	minX, minY := 1e9, 1e9
	maxX, maxY := -1e9, -1e9
	for idx := uint(0); idx < lionData.Path.TotalVertices(); idx++ {
		x, y, cmd := lionData.Path.Vertex(idx)
		pathCmd := basics.PathCommand(cmd)
		if basics.IsMoveTo(pathCmd) || basics.IsLineTo(pathCmd) {
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	lionBaseDX = (maxX - minX) / 2.0
	lionBaseDY = (maxY - minY) / 2.0
}

func rgbaRTL(rng *clibcRand, alphaBase int) color.RGBA8[color.Linear] {
	a := uint8(rng.randAnd(0x7F) + alphaBase)
	b := uint8(rng.randAnd(0x7F))
	g := uint8(rng.randAnd(0x7F))
	r := uint8(rng.randAnd(0x7F))
	return color.RGBA8[color.Linear]{R: r, G: g, B: b, A: a}
}

func RenderToBGR24(dst []uint8, width, height int, cfg Config) {
	lionOnce.Do(initLion)

	rbuf := buffer.NewRenderingBufferU8WithData(dst, width, height, width*3)
	mainPixf := pixfmt.NewPixFmtBGR24(rbuf)
	mainPixfAdaptor := pixfmt.NewPixFmtRGBARendererAdaptor(mainPixf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixfAdaptor)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	maskData := make([]uint8, width*height)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	maskPixf := pixfmt.NewPixFmtSGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineU8()
	rng := newClibcRandSeed1()

	numEllipses := cfg.NumEllipses
	if numEllipses <= 0 {
		numEllipses = 10
	}
	scale := cfg.Scale
	if scale == 0 {
		scale = 1
	}

	for i := 0; i < numEllipses; i++ {
		ry := float64(rng.randN(100) + 20)
		rx := float64(rng.randN(100) + 20)
		y := float64(rng.randN(height))
		x := float64(rng.randN(width))
		ell := shapes.NewEllipseWithParams(x, y, rx, ry, 100, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		a := uint8(rng.randAnd(127) + 128)
		v := uint8(rng.randAnd(127) + 128)
		renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.SRGB]{V: v, A: a})
	}

	mask := pixfmt.NewAMaskNoClipU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(mainPixfAdaptor, mask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-lionBaseDX, -lionBaseDY))
	mtx.Multiply(transform.NewTransAffineScaling(scale))
	mtx.Multiply(transform.NewTransAffineRotation(cfg.Angle + math.Pi))
	mtx.Multiply(transform.NewTransAffineSkewing(cfg.SkewX/1000.0, cfg.SkewY/1000.0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(width)/2, float64(height)/2))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rbAMask)
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, &lionData, &lionData, lionData.NPaths)

	renderMarkers(rbAMask, rng, width, height)
	renderOutlineLines(rbAMask, rng, width, height)
	renderGradientCircles(ras, sl, rbAMask, rng, width, height)
}

func renderMarkers(
	rbAMask *renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	rng *clibcRand,
	width, height int,
) {
	m := markers.NewRendererMarkers(rbAMask)
	for i := 0; i < 50; i++ {
		m.LineColor(rgbaRTL(rng, 0x7F))
		m.FillColor(rgbaRTL(rng, 0x7F))

		y2 := rng.randN(height)
		x2 := rng.randN(width)
		y1 := rng.randN(height)
		x1 := rng.randN(width)
		m.Line(m.Coord(float64(x1)), m.Coord(float64(y1)), m.Coord(float64(x2)), m.Coord(float64(y2)), false)

		markerType := markers.MarkerType(rng.randN(int(markers.EndOfMarkers)))
		radius := rng.randN(10) + 5
		y := rng.randN(height)
		x := rng.randN(width)
		m.Marker(x, y, radius, markerType)
	}
}

func renderOutlineLines(
	rbAMask *renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	rng *clibcRand,
	width, height int,
) {
	profile := outline.NewLineProfileAA()
	profile.Width(5.0)

	renOutline := outline.NewRendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]](
		&outlineBaseAdapter{renBase: rbAMask},
		profile,
	)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter, color.RGBA8[color.Linear]](
		&outlineAAAdapter{ren: renOutline},
	)
	rasOutline.SetRoundCap(true)

	for i := 0; i < 50; i++ {
		renOutline.Color(rgbaRTL(rng, 0x7F))
		y1 := rng.randN(height)
		x1 := rng.randN(width)
		rasOutline.MoveToD(float64(x1), float64(y1))
		y2 := rng.randN(height)
		x2 := rng.randN(width)
		rasOutline.LineToD(float64(x2), float64(y2))
		rasOutline.Render(false)
	}
}

func renderGradientCircles(
	ras *rasType,
	sl *scanline.ScanlineU8,
	rbAMask *renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	rng *clibcRand,
	width, height int,
) {
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	for i := 0; i < 50; i++ {
		x := rng.randN(width)
		y := rng.randN(height)
		r := float64(rng.randN(10) + 5)

		grm := transform.NewTransAffineScaling(r / 10.0)
		grm.Multiply(transform.NewTransAffineTranslation(float64(x), float64(y)))
		grm.Invert()

		inter := span.NewSpanInterpolatorLinearDefault(grm)
		colorFunc := span.NewGradientLinearColorRGBA8(
			color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 0},
			func() color.RGBA8[color.Linear] {
				b := uint8(rng.randAnd(0x7F))
				g := uint8(rng.randAnd(0x7F))
				rv := uint8(rng.randAnd(0x7F))
				return color.RGBA8[color.Linear]{R: rv, G: g, B: b, A: 255}
			}(),
			256,
		)
		spanGen := span.NewSpanGradient(inter, span.GradientRadial{}, colorFunc, 0, 10)

		ell := shapes.NewEllipseWithParams(float64(x), float64(y), r, r, 32, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		renscan.RenderScanlinesAA(ras, sl, rbAMask, alloc, spanGen)
	}
}
