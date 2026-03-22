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

type outlineBlendBase[C any] interface {
	Width() int
	Height() int
	BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u)
	BlendSolidVspan(x, y, length int, c C, covers []basics.Int8u)
}

type outlineBaseAdapter[C any] struct {
	renBase outlineBlendBase[C]
}

func (a *outlineBaseAdapter[C]) Width() int  { return a.renBase.Width() }
func (a *outlineBaseAdapter[C]) Height() int { return a.renBase.Height() }

func (a *outlineBaseAdapter[C]) BlendSolidHSpan(x, y, length int, c C, covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidHspan(x, y, length, c, convCovers)
}

func (a *outlineBaseAdapter[C]) BlendSolidVSpan(x, y, length int, c C, covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidVspan(x, y, length, c, convCovers)
}

type outlineRenderer[C any] interface {
	AccurateJoinOnly() bool
	Color(c C)
	Line0(lp *primitives.LineParameters)
	Line1(lp *primitives.LineParameters, sx, sy int)
	Line2(lp *primitives.LineParameters, ex, ey int)
	Line3(lp *primitives.LineParameters, sx, sy, ex, ey int)
	Pie(x, y, x1, y1, x2, y2 int)
	Semidot(cmp func(int) bool, x, y, x1, y1 int)
}

type outlineAAAdapter[C any] struct {
	ren outlineRenderer[C]
}

func (a *outlineAAAdapter[C]) AccurateJoinOnly() bool { return a.ren.AccurateJoinOnly() }
func (a *outlineAAAdapter[C]) Color(c C)              { a.ren.Color(c) }
func (a *outlineAAAdapter[C]) Line0(lp primitives.LineParameters) {
	a.ren.Line0(&lp)
}

func (a *outlineAAAdapter[C]) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}

func (a *outlineAAAdapter[C]) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}

func (a *outlineAAAdapter[C]) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}
func (a *outlineAAAdapter[C]) Pie(x, y, x1, y1, x2, y2 int) { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *outlineAAAdapter[C]) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRand(seed int32) *clibcRand {
	if seed == 0 {
		seed = 1
	}

	// glibc rand()/random() state initialization:
	// 1. Park-Miller sequence for 31 values from srand(seed)
	// 2. copy first 3 values
	// 3. additive feedback warmup for 310 outputs
	//
	// The final state layout matches the in-memory representation used by the
	// existing seed=1/seed=1234 precomputed tables elsewhere in this repo:
	// state[0:3] hold the values immediately before the current rptr/fptr pair.
	const (
		mod  int64 = 2147483647
		mult int64 = 16807
	)

	var seq [344]int32
	seq[0] = seed
	for i := 1; i < 31; i++ {
		v := mult * int64(seq[i-1]) % mod
		if v < 0 {
			v += mod
		}
		seq[i] = int32(v)
	}
	for i := 31; i < 34; i++ {
		seq[i] = seq[i-31]
	}
	for i := 34; i < len(seq); i++ {
		seq[i] = seq[i-31] + seq[i-3]
	}

	rng := &clibcRand{
		fptr: 3,
		rptr: 0,
	}
	copy(rng.state[:3], seq[341:344])
	copy(rng.state[3:], seq[313:341])
	return rng
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

func srgbaRandRTL(rng *clibcRand, alphaBase int) color.RGBA8[color.SRGB] {
	// Keep the previous right-to-left overlay argument mapping for now.
	// The seed issue is resolved, but the exact C++ argument evaluation order for
	// nested calls in alpha_mask2.cpp is not locked down yet.
	return color.RGBA8[color.SRGB]{
		A: uint8(rng.randAnd(0x7F) + alphaBase),
		B: uint8(rng.randAnd(0x7F)),
		G: uint8(rng.randAnd(0x7F)),
		R: uint8(rng.randAnd(0x7F)),
	}
}

func RenderToBGR24(dst []uint8, width, height int, cfg Config) {
	lionOnce.Do(initLion)

	rbuf := buffer.NewRenderingBufferU8WithData(dst, width, height, width*3)
	mainPixfLinear := pixfmt.NewPixFmtBGR24(rbuf)
	mainPixfLinearAdaptor := pixfmt.NewPixFmtRGBARendererAdaptor(mainPixfLinear)
	mainRbLinear := renderer.NewRendererBaseWithPixfmt(mainPixfLinearAdaptor)
	mainRbLinear.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	maskData := make([]uint8, width*height)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	maskPixf := pixfmt.NewPixFmtSGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineU8()
	// C++ generate_alpha_mask() explicitly seeds rand() with srand(1432), and
	// the same RNG stream continues into the overlay rendering in on_draw().
	rng := newClibcRand(1432)

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
	amaskAdaptorLinear := pixfmt.NewPixFmtAMaskAdaptor(mainPixfLinearAdaptor, mask)
	rbAMaskLinear := renderer.NewRendererBaseWithPixfmt(amaskAdaptorLinear)

	mainPixfSRGB := pixfmt.NewPixFmtSBGR24(rbuf)
	mainPixfSRGBAdaptor := pixfmt.NewPixFmtRGBARendererAdaptor(mainPixfSRGB)
	amaskAdaptorSRGB := pixfmt.NewPixFmtAMaskAdaptor(mainPixfSRGBAdaptor, mask)
	rbAMaskSRGB := renderer.NewRendererBaseWithPixfmt(amaskAdaptorSRGB)

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-lionBaseDX, -lionBaseDY))
	mtx.Multiply(transform.NewTransAffineScaling(scale))
	mtx.Multiply(transform.NewTransAffineRotation(cfg.Angle + math.Pi))
	mtx.Multiply(transform.NewTransAffineSkewing(cfg.SkewX/1000.0, cfg.SkewY/1000.0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(width)/2, float64(height)/2))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rbAMaskLinear)
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, &lionData, &lionData, lionData.NPaths)

	renderMarkers(rbAMaskSRGB, rng, width, height)
	renderOutlineLines(rbAMaskSRGB, rng, width, height)
	renderGradientCircles(ras, sl, rbAMaskSRGB, rng, width, height)
}

func renderMarkers(
	rbAMask *renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]],
	rng *clibcRand,
	width, height int,
) {
	m := markers.NewRendererMarkers(rbAMask)
	for i := 0; i < 50; i++ {
		m.LineColor(srgbaRandRTL(rng, 0x7F))
		m.FillColor(srgbaRandRTL(rng, 0x7F))

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
	rbAMask *renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]],
	rng *clibcRand,
	width, height int,
) {
	profile := outline.NewLineProfileAA()
	profile.Width(5.0)

	renOutline := outline.NewRendererOutlineAA[*outlineBaseAdapter[color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]](
		&outlineBaseAdapter[color.RGBA8[color.SRGB]]{renBase: rbAMask},
		profile,
	)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter[color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]](
		&outlineAAAdapter[color.RGBA8[color.SRGB]]{ren: renOutline},
	)
	rasOutline.SetRoundCap(true)

	for i := 0; i < 50; i++ {
		renOutline.Color(srgbaRandRTL(rng, 0x7F))
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
	rbAMask *renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]],
	rng *clibcRand,
	width, height int,
) {
	alloc := span.NewSpanAllocator[color.RGBA8[color.SRGB]]()
	for i := 0; i < 50; i++ {
		x := rng.randN(width)
		y := rng.randN(height)
		r := float64(rng.randN(10) + 5)

		grm := transform.NewTransAffineScaling(r / 10.0)
		grm.Multiply(transform.NewTransAffineTranslation(float64(x), float64(y)))
		grm.Invert()

		inter := span.NewSpanInterpolatorLinearDefault(grm)
		colorFunc := span.NewGradientLinearColorRGBA8(
			color.RGBA8[color.SRGB]{R: 255, G: 255, B: 255, A: 0},
			color.RGBA8[color.SRGB]{
				R: uint8(rng.randAnd(0x7F)),
				G: uint8(rng.randAnd(0x7F)),
				B: uint8(rng.randAnd(0x7F)),
				A: 255,
			},
			256,
		)
		spanGen := span.NewSpanGradient(inter, span.GradientRadial{}, colorFunc, 0, 10)

		ell := shapes.NewEllipseWithParams(float64(x), float64(y), r, r, 32, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		renscan.RenderScanlinesAA(ras, sl, rbAMask, alloc, spanGen)
	}
}
