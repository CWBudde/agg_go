// Port of AGG C++ alpha_gradient.cpp – alpha channel gradient demo.
//
// A large ellipse is filled with a circular color gradient whose alpha is
// modulated by a separate XY-product alpha gradient mapped over a parallelogram.
// Background is random colourful ellipses.
// A spline control at the bottom-left lets the user adjust the alpha curve.
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
	splinectrl "github.com/MeKo-Christian/agg_go/internal/ctrl/spline"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	frameWidth  = 400
	frameHeight = 320
)

// ---------------------------------------------------------------------------
// glibc rand() with srand(1234) — reproduces C++ srand(1234)/rand() sequence.
// State computed by simulating glibc's srand(1234) initialization + 310 warmup cycles.
// ---------------------------------------------------------------------------

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRandSeed1234() *clibcRand {
	return &clibcRand{
		state: [31]int32{
			997300753, 1787873760, -240326740, -39015925, -856741081,
			-2132388246, 1157487307, 1514271441, 112649172, -76012625,
			1994128572, 2062673662, 2076976597, 516503355, 1318736635,
			993161121, 888449716, 1552615853, -98235190, -1188648751,
			-140598225, -882898109, 538146087, 808667150, -1994922881,
			-1790577642, 716409717, -245274338, -1236232937, -501689970, 1965975355,
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

// randN returns rand() % n, matching C++ rand() % n.
func (r *clibcRand) randN(n int) int {
	return int(r.next()) % n
}

// randDouble returns rand()/double(RAND_MAX), matching C++ rand()/double(RAND_MAX).
func (r *clibcRand) randDouble() float64 {
	return float64(r.next()) / 2147483647.0
}

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters (shared with circles, gamma_correction, etc.)
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

// ellipseVS adapts shapes.Ellipse to rasterizer.VertexSource.
type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	var vx, vy float64
	cmd := ev.e.Vertex(&vx, &vy)
	*x, *y = vx, vy
	return uint32(cmd)
}

// convVS adapts conv.VertexSource (Rewind(uint), Vertex()->(x,y,cmd)) to
// rasterizer.VertexSource (Rewind(uint32), Vertex(*x,*y) uint32).
type convVS struct {
	src interface {
		Rewind(uint)
		Vertex() (float64, float64, basics.PathCommand)
	}
}

func (v *convVS) Rewind(id uint32) { v.src.Rewind(uint(id)) }
func (v *convVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ctrlVS adapts a ctrl.Ctrl (Rewind(uint), Vertex()) to rasterizer.VertexSource.
type ctrlVS struct {
	src interface {
		Rewind(uint)
		Vertex() (float64, float64, basics.PathCommand)
	}
}

func (v *ctrlVS) Rewind(id uint32) { v.src.Rewind(uint(id)) }
func (v *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// BGR24 renderer adapter (matches the C++ pixel format)
// ---------------------------------------------------------------------------

type bgr24Renderer struct {
	pf *pixfmt.PixFmtBGR24
}

func newBGR24Renderer(rbuf *buffer.RenderingBufferU8) *bgr24Renderer {
	pf := pixfmt.NewPixFmtBGR24(rbuf)
	return &bgr24Renderer{pf: pf}
}

func (r *bgr24Renderer) Clear(c color.RGBA8[color.Linear]) {
	r.pf.Clear(color.RGB8[color.Linear]{R: c.R, G: c.G, B: c.B})
}

func (r *bgr24Renderer) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	r.pf.BlendSolidHspan(x, y, length,
		color.RGB8[color.Linear]{R: c.R, G: c.G, B: c.B}, c.A, covers)
}

func (r *bgr24Renderer) BlendHline(x, y, x2 int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	r.pf.BlendHline(x, y, x2,
		color.RGB8[color.Linear]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

func (r *bgr24Renderer) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr == 0 {
			continue
		}
		r.pf.BlendPixel(x+i, y,
			color.RGB8[color.Linear]{R: colors[i].R, G: colors[i].G, B: colors[i].B},
			colors[i].A, cvr)
	}
}

// ---------------------------------------------------------------------------
// Color array for gradient
// ---------------------------------------------------------------------------

type gradColorArray struct {
	data [256]color.RGBA8[color.Linear]
}

func (a *gradColorArray) Size() int                               { return 256 }
func (a *gradColorArray) ColorAt(i int) color.RGBA8[color.Linear] { return a.data[i] }

func fillColorArray(arr *gradColorArray, begin, middle, end agg.Color) {
	lerp := func(a, b uint8, t float64) uint8 {
		return uint8(float64(a)*(1-t) + float64(b)*t)
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

// ---------------------------------------------------------------------------
// Combined color+alpha span generator
// ---------------------------------------------------------------------------

type alphaGradSpanGen struct {
	gradInterp  *span.SpanInterpolatorLinear[*transform.TransAffine]
	alphaInterp *span.SpanInterpolatorLinear[*transform.TransAffine]
	colorArray  gradColorArray
	alphaArray  [256]basics.Int8u
	d1c, d2c    int
	d1a, d2a    int
	downscale   int
}

func newAlphaGradSpanGen(
	gradMtx, alphaMtx *transform.TransAffine,
	colorArr *gradColorArray,
	alphaArr *[256]basics.Int8u,
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
		colorArray:  *colorArr,
		alphaArray:  *alphaArr,
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
		cx, cy := g.gradInterp.Coordinates()
		d := colorGrad.Calculate(cx>>g.downscale, cy>>g.downscale, g.d2c)
		ci := ((d - g.d1c) * 256) / ddc
		if ci < 0 {
			ci = 0
		} else if ci >= 256 {
			ci = 255
		}
		colors[i] = g.colorArray.data[ci]

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

// ---------------------------------------------------------------------------
// renderCtrl renders all paths of an AGG control widget.
// ---------------------------------------------------------------------------

func renderCtrl(
	ras *rasType,
	sl *scanline.ScanlineP8,
	rb *bgr24Renderer,
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVS{src: ctrl}, uint32(pathID))
		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{
			R: uint8(math.Round(c.R * 255)),
			G: uint8(math.Round(c.G * 255)),
			B: uint8(math.Round(c.B * 255)),
			A: uint8(math.Round(c.A * 255)),
		})
	}
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct {
	// Parallelogram control points
	mx [3]float64
	my [3]float64
	// Mouse drag state
	dx, dy float64
	idx    int
	// Spline alpha control
	alpha *splinectrl.SplineCtrl[color.RGBA]
}

func newDemo() *demo {
	d := &demo{idx: -1}
	d.mx[0] = 257
	d.my[0] = 60
	d.mx[1] = 369
	d.my[1] = 170
	d.mx[2] = 143
	d.my[2] = 310

	// C++: m_alpha(2, 2, 200, 30, 6, !flip_y)
	// flip_y=true, so !flip_y=false
	d.alpha = splinectrl.NewSplineCtrlRGBA(2, 2, 200, 30, 6, false)

	// Match C++ control point initialization:
	// m_alpha.point(0, 0.0,     0.0);
	// m_alpha.point(1, 1.0/5.0, 1.0 - 4.0/5.0);
	// m_alpha.point(2, 2.0/5.0, 1.0 - 3.0/5.0);
	// m_alpha.point(3, 3.0/5.0, 1.0 - 2.0/5.0);
	// m_alpha.point(4, 4.0/5.0, 1.0 - 1.0/5.0);
	// m_alpha.point(5, 1.0,     1.0);
	d.alpha.SetPoint(0, 0.0, 0.0)
	d.alpha.SetPoint(1, 1.0/5.0, 1.0-4.0/5.0)
	d.alpha.SetPoint(2, 2.0/5.0, 1.0-3.0/5.0)
	d.alpha.SetPoint(3, 3.0/5.0, 1.0-2.0/5.0)
	d.alpha.SetPoint(4, 4.0/5.0, 1.0-1.0/5.0)
	d.alpha.SetPoint(5, 1.0, 1.0)
	// updateSpline is called automatically by SetPoint; no explicit call needed.

	return d
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)

	if d.alpha.OnMouseButtonDown(fx, fy) {
		return true
	}

	for i := 0; i < 3; i++ {
		dx := fx - d.mx[i]
		dy := fy - d.my[i]
		if math.Sqrt(dx*dx+dy*dy) < 10.0 {
			d.dx = fx - d.mx[i]
			d.dy = fy - d.my[i]
			d.idx = i
			return true
		}
	}

	// Check if click is inside the triangle formed by the 3 control points.
	if pointInTriangle(d.mx[0], d.my[0], d.mx[1], d.my[1], d.mx[2], d.my[2], fx, fy) {
		d.dx = fx - d.mx[0]
		d.dy = fy - d.my[0]
		d.idx = 3
		return true
	}

	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	if d.alpha.OnMouseMove(fx, fy, btn.Left) {
		return true
	}

	if !btn.Left {
		d.idx = -1
		return false
	}

	if d.idx == 3 {
		dx := fx - d.dx
		dy := fy - d.dy
		d.mx[1] -= d.mx[0] - dx
		d.my[1] -= d.my[0] - dy
		d.mx[2] -= d.mx[0] - dx
		d.my[2] -= d.my[0] - dy
		d.mx[0] = dx
		d.my[0] = dy
		return true
	}
	if d.idx >= 0 {
		d.mx[d.idx] = fx - d.dx
		d.my[d.idx] = fy - d.dy
		return true
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	_ = btn
	d.alpha.OnMouseButtonUp(fx, fy)
	d.idx = -1
	return false
}

// pointInTriangle reports whether (px, py) is inside the triangle (x1,y1)-(x2,y2)-(x3,y3).
// This matches AGG's point_in_triangle logic.
func pointInTriangle(x1, y1, x2, y2, x3, y3, px, py float64) bool {
	sign := func(x1, y1, x2, y2, px, py float64) float64 {
		return (px-x2)*(y1-y2) - (x1-x2)*(py-y2)
	}
	d1 := sign(x1, y1, x2, y2, px, py)
	d2 := sign(x2, y2, x3, y3, px, py)
	d3 := sign(x3, y3, x1, y1, px, py)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(hasNeg && hasPos)
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Use BGR24 to match C++ (AGG_BGR24).
	workBuf := make([]uint8, w*h*3)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*3)
	rb := newBGR24Renderer(workRbuf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineP8()

	cx := float64(w) / 2
	cy := float64(h) / 2

	// 1. Random background ellipses (seed 1234, matching C++ srand(1234)/rand()).
	// C++ argument evaluation order (GCC x86, right-to-left):
	//   ell.init(rand()%w, rand()%h, rand()%60+5, rand()%60+5, 50)
	//   evaluates as: ry, rx, y, x  (fully right-to-left)
	//   rgba(rand()/RAND_MAX, rand()/RAND_MAX, rand()/RAND_MAX, rand()/RAND_MAX/2)
	//   evaluates as: a, b, g, r    (fully right-to-left)
	// Verified empirically via LD_PRELOAD rand() tracing of the compiled binary.
	rng := newClibcRandSeed1234()
	for i := 0; i < 100; i++ {
		ry := float64(rng.randN(60) + 5)
		rx := float64(rng.randN(60) + 5)
		y := float64(rng.randN(h))
		x := float64(rng.randN(w))
		ell := shapes.NewEllipseWithParams(x, y, rx, ry, 50, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		a := uint8(rng.randDouble()/2.0*255 + 0.5)
		b := uint8(rng.randDouble()*255 + 0.5)
		g := uint8(rng.randDouble()*255 + 0.5)
		r := uint8(rng.randDouble()*255 + 0.5)
		c := color.RGBA8[color.Linear]{R: r, G: g, B: b, A: a}
		renscan.RenderScanlinesAASolid(ras, sl, rb, c)
	}

	// 2. Gradient matrix.
	gradMtx := transform.NewTransAffine()
	gradMtx.Multiply(transform.NewTransAffineScalingXY(0.75, 1.2))
	gradMtx.Multiply(transform.NewTransAffineRotation(-basics.Pi / 3.0))
	gradMtx.Multiply(transform.NewTransAffineTranslation(cx, cy))
	gradMtx.Invert()

	// 3. Control points defining the alpha parallelogram.
	parl := [6]float64{d.mx[0], d.my[0], d.mx[1], d.my[1], d.mx[2], d.my[2]}
	alphaMtx := transform.NewTransAffineParlToRect(parl, -100, -100, 100, 100)

	// 4. Color LUT: dark teal → yellow-green → dark red.
	var colorArr gradColorArray
	fillColorArray(&colorArr,
		agg.RGBA(0, 0.19, 0.19, 1),
		agg.RGBA(0.7, 0.7, 0.19, 1),
		agg.RGBA(0.31, 0, 0, 1),
	)

	// 5. Alpha LUT from spline control: alpha_array[i] = from_double(m_alpha.value(i/255.0))
	var alphaArr [256]basics.Int8u
	for i := range alphaArr {
		v := d.alpha.Value(float64(i) / 255.0)
		if v < 0 {
			v = 0
		} else if v > 1 {
			v = 1
		}
		alphaArr[i] = basics.Int8u(math.Round(v * 255))
	}

	// 6. Render gradient ellipse via span generator.
	spanGen := newAlphaGradSpanGen(gradMtx, alphaMtx, &colorArr, &alphaArr)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	ras.Reset()
	ell := shapes.NewEllipseWithParams(cx, cy, 150, 150, 100, false)
	ras.AddPath(&ellipseVS{e: ell}, 0)
	renscan.RenderScanlinesAA(ras, sl, rb, alloc, spanGen)

	// 7. Control point dots.
	ctrlCol := color.RGBA8[color.Linear]{R: 0, G: 102, B: 102, A: 79}
	for i := 0; i < 3; i++ {
		dot := shapes.NewEllipseWithParams(d.mx[i], d.my[i], 5, 5, 20, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: dot}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, rb, ctrlCol)
	}

	// 8. Parallelogram outline via vcgen_stroke (ConvStroke).
	p3x := d.mx[0] + d.mx[2] - d.mx[1]
	p3y := d.my[0] + d.my[2] - d.my[1]
	ps := path.NewPathStorage()
	ps.MoveTo(d.mx[0], d.my[0])
	ps.LineTo(d.mx[1], d.my[1])
	ps.LineTo(d.mx[2], d.my[2])
	ps.LineTo(p3x, p3y)
	ps.ClosePolygon(basics.PathFlagsNone)

	stroke := conv.NewConvStroke(path.NewPathStorageVertexSourceAdapter(ps))
	ras.Reset()
	ras.AddPath(&convVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// 9. Render spline control widget.
	renderCtrl(ras, sl, rb, d.alpha)

	// Copy BGR24 work buffer → RGBA32 output with y-flip (C++ flip_y=true).
	copyBGR24FlipY(workBuf, img.Data, w, h)
}

// copyBGR24FlipY copies a BGR24 buffer (flipped) into an RGBA32 output buffer.
func copyBGR24FlipY(src, dst []uint8, width, height int) {
	srcStride := width * 3
	dstStride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * srcStride
		dstOff := y * dstStride
		for x := 0; x < width; x++ {
			s := srcOff + x*3
			d := dstOff + x*4
			// BGR → RGBA
			dst[d+0] = src[s+2] // R
			dst[d+1] = src[s+1] // G
			dst[d+2] = src[s+0] // B
			dst[d+3] = 255      // A
		}
	}
}

func main() {
	d := newDemo()
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Alpha Gradient",
		Width:  frameWidth,
		Height: frameHeight,
	}, d)
}
