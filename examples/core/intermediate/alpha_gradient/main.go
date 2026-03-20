// Port of AGG C++ alpha_gradient.cpp – alpha channel gradient demo.
//
// A large ellipse is filled with a circular color gradient whose alpha is
// modulated by a separate XY-product alpha gradient mapped over a parallelogram.
// Background is random colourful ellipses.
package main

import (
	"math/rand"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
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
// Rasterizer / scanline adapters
// ---------------------------------------------------------------------------

type rasterizerAdaptor struct {
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	sl  rasScanlineAdaptor
}

func newRasterizer() *rasterizerAdaptor {
	return &rasterizerAdaptor{
		ras: rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
			rasterizer.RasConvInt{},
			rasterizer.NewRasterizerSlNoClip(),
		),
		sl: rasScanlineAdaptor{sl: scanline.NewScanlineP8()},
	}
}

func (r *rasterizerAdaptor) Reset()                { r.ras.Reset() }
func (r *rasterizerAdaptor) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *rasterizerAdaptor) MinX() int             { return r.ras.MinX() }
func (r *rasterizerAdaptor) MaxX() int             { return r.ras.MaxX() }

func (r *rasterizerAdaptor) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*scanlineWrapper); ok {
		r.sl.sl = w.sl
		return r.ras.SweepScanline(&r.sl)
	}
	return false
}

type rasScanlineAdaptor struct{ sl *scanline.ScanlineP8 }

func (a *rasScanlineAdaptor) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdaptor) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdaptor) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdaptor) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdaptor) NumSpans() int  { return a.sl.NumSpans() }

type scanlineWrapper struct{ sl *scanline.ScanlineP8 }

func (w *scanlineWrapper) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapper) Y() int               { return w.sl.Y() }
func (w *scanlineWrapper) NumSpans() int        { return w.sl.NumSpans() }

func (w *scanlineWrapper) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIter{nil, 0}
	}
	return &spanIter{spans, 0}
}

type spanIter struct {
	spans []scanline.SpanP8
	idx   int
}

func (it *spanIter) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIter) Next() bool { it.idx++; return it.idx < len(it.spans) }

// ---------------------------------------------------------------------------
// Vertex-source adapters
// ---------------------------------------------------------------------------

type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	var vx, vy float64
	cmd := ev.e.Vertex(&vx, &vy)
	*x, *y = vx, vy
	return uint32(cmd)
}

type vcgenStrokeVS struct {
	ps *path.PathStorageStl
}

func (v *vcgenStrokeVS) Rewind(id uint) { v.ps.Rewind(id) }
func (v *vcgenStrokeVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := v.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
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
// Demo
// ---------------------------------------------------------------------------

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRasterizer()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}

	cx := float64(w) / 2
	cy := float64(h) / 2

	// 1. Random background ellipses (seed 1234, matching C++ srand(1234)).
	rng := rand.New(rand.NewSource(1234))
	for i := 0; i < 100; i++ {
		ell := shapes.NewEllipseWithParams(
			float64(rng.Intn(w)), float64(rng.Intn(h)),
			float64(rng.Intn(60)+5), float64(rng.Intn(60)+5),
			50, false,
		)
		ras.Reset()
		ras.ras.AddPath(&ellipseVS{e: ell}, 0)
		c := color.RGBA8[color.Linear]{
			R: uint8(rng.Intn(256)),
			G: uint8(rng.Intn(256)),
			B: uint8(rng.Intn(256)),
			A: uint8(rng.Intn(128)),
		}
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, c)
	}

	// 2. Gradient matrix.
	gradMtx := transform.NewTransAffine()
	gradMtx.Multiply(transform.NewTransAffineScalingXY(0.75, 1.2))
	gradMtx.Multiply(transform.NewTransAffineRotation(-basics.Pi / 3.0))
	gradMtx.Multiply(transform.NewTransAffineTranslation(cx, cy))
	gradMtx.Invert()

	// 3. Control points defining the alpha parallelogram.
	pts := [3][2]float64{{257, 60}, {369, 170}, {143, 310}}
	parl := [6]float64{pts[0][0], pts[0][1], pts[1][0], pts[1][1], pts[2][0], pts[2][1]}
	alphaMtx := transform.NewTransAffineParlToRect(parl, -100, -100, 100, 100)

	// 4. Color LUT: dark teal → yellow-green → dark red.
	var colorArr gradColorArray
	fillColorArray(&colorArr,
		agg.RGBA(0, 0.19, 0.19, 1),
		agg.RGBA(0.7, 0.7, 0.19, 1),
		agg.RGBA(0.31, 0, 0, 1),
	)

	// 5. Alpha LUT: linear 0→255 (matches C++ default spline).
	var alphaArr [256]basics.Int8u
	for i := range alphaArr {
		alphaArr[i] = basics.Int8u(i)
	}

	// 6. Render gradient ellipse via span generator.
	spanGen := newAlphaGradSpanGen(gradMtx, alphaMtx, &colorArr, &alphaArr)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	ras.Reset()
	ell := shapes.NewEllipseWithParams(cx, cy, 150, 150, 100, false)
	ras.ras.AddPath(&ellipseVS{e: ell}, 0)
	renscan.RenderScanlinesAA(ras, sl, mainRb, alloc, spanGen)

	// 7. Control point dots.
	ctrlCol := color.RGBA8[color.Linear]{R: 0, G: 102, B: 102, A: 79}
	for i := 0; i < 3; i++ {
		dot := shapes.NewEllipseWithParams(pts[i][0], pts[i][1], 5, 5, 20, false)
		ras.Reset()
		ras.ras.AddPath(&ellipseVS{e: dot}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, ctrlCol)
	}

	// 8. Parallelogram outline.
	p3x := pts[0][0] + pts[2][0] - pts[1][0]
	p3y := pts[0][1] + pts[2][1] - pts[1][1]
	ps := path.NewPathStorageStl()
	ps.MoveTo(pts[0][0], pts[0][1])
	ps.LineTo(pts[1][0], pts[1][1])
	ps.LineTo(pts[2][0], pts[2][1])
	ps.LineTo(p3x, p3y)
	ps.ClosePolygon(basics.PathFlagsCW)

	// Render as thin stroke using vcgen_stroke approach.
	ras.Reset()
	ps.Rewind(0)
	for {
		x, y, cmd := ps.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		ras.ras.AddVertex(x, y, cmd)
	}
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// Copy with y-flip (C++ uses flip_y=true).
	copyFlipY(workBuf, img.Data, w, h)
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
		Title:  "Alpha Gradient",
		Width:  frameWidth,
		Height: frameHeight,
	}, &demo{})
}
