// Port of AGG C++ aa_test.cpp – anti-aliasing quality test.
//
// Renders radial lines, ellipses at various sizes, gradient lines, and
// gradient triangles on a black background. Note: the C++ uses flip_y=false,
// so no y-flip is needed in the output.
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

const (
	frameWidth  = 480
	frameHeight = 350
)

// ---------------------------------------------------------------------------
// Type aliases
// ---------------------------------------------------------------------------

type (
	rasType     = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	renBaseType = *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]
)

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

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

type pathStlVS struct{ ps *path.PathStorageStl }

func (a *pathStlVS) Rewind(id uint) { a.ps.Rewind(id) }
func (a *pathStlVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

type convVS struct{ src conv.VertexSource }

func (a *convVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Gradient support
// ---------------------------------------------------------------------------

// mutableLUT is a 256-entry sRGB colour lookup table (= C++'s
// pod_auto_array<srgba8,256>) whose contents can be updated in-place between
// draw calls.  It implements ColorFunction[RGBA8[SRGB]].
//
// Pair it with span.SRGBColorAdapter when feeding a linear-pixfmt span
// generator; the adapter decodes each entry to linear on lookup, matching
// C++'s implicit rgba8(srgba8) conversion inside span_gradient::generate.
type mutableLUT struct {
	lut [256]color.RGBA8[color.SRGB]
}

func (m *mutableLUT) Size() int                             { return 256 }
func (m *mutableLUT) ColorAt(i int) color.RGBA8[color.SRGB] { return m.lut[i] }

// fillColors builds a ramp from c1 (index 0) to c2 (index 255) in sRGB byte
// space, matching C++'s fill_color_array with a pod_auto_array<srgba8,256>:
//   - ConvertRGBA8LinearToSRGB  ≡  srgba8(rgba) constructor
//   - RGBA8[SRGB].Gradient      ≡  srgba8::gradient  (lerp in sRGB space)
func (m *mutableLUT) fillColors(c1, c2 color.RGBA8[color.Linear]) {
	s1 := color.ConvertRGBA8LinearToSRGB(c1)
	s2 := color.ConvertRGBA8LinearToSRGB(c2)
	for i := range m.lut {
		m.lut[i] = s1.Gradient(s2, basics.Int8u(i))
	}
}

// calcLinearGradientTransform matches AGG's calc_linear_gradient_transform.
// It sets gradMtx so that a gradient_d2=100 range spans (x1,y1)→(x2,y2).
// The matrix is inverted (screen→gradient space) as in the C++ original.
func calcLinearGradientTransform(gradMtx *transform.TransAffine, x1, y1, x2, y2 float64) {
	dx := x2 - x1
	dy := y2 - y1
	dist := math.Sqrt(dx*dx + dy*dy)
	gradMtx.Reset()
	if dist > 1e-10 {
		gradMtx.Multiply(transform.NewTransAffineScaling(dist / 100.0))
		gradMtx.Multiply(transform.NewTransAffineRotation(math.Atan2(dy, dx)))
	}
	gradMtx.Multiply(transform.NewTransAffineTranslation(x1+0.5, y1+0.5))
	gradMtx.Invert()
}

// Gradient span-generator types.
// gradAdapterType wraps *mutableLUT (ColorFunction[RGBA8[SRGB]]) so that the
// span generator receives ColorFunction[RGBA8[Linear]] — the Go equivalent of
// C++'s implicit rgba8(srgba8) conversion inside span_gradient::generate.
type (
	gradInterpType  = *span.SpanInterpolatorLinear[*transform.TransAffine]
	gradAdapterType = span.SRGBColorAdapter[*mutableLUT]
	gradSpanGenType = *span.SpanGradient[color.RGBA8[color.Linear], gradInterpType, span.GradientLinearX, gradAdapterType]
	gradAllocType   = *span.SpanAllocator[color.RGBA8[color.Linear]]
)

// ---------------------------------------------------------------------------
// Line drawing helpers
// ---------------------------------------------------------------------------

// rasAddLine fills ras with a stroked line path (optionally dashed).
// Mirrors the draw() logic in the C++ dashed_line helper template.
func rasAddLine(ras *rasType, x1, y1, x2, y2, lineWidth, dashLength float64) {
	ps := path.NewPathStorageStl()
	ps.MoveTo(x1+0.5, y1+0.5)
	ps.LineTo(x2+0.5, y2+0.5)
	src := &pathStlVS{ps: ps}

	ras.Reset()
	if dashLength > 0.0 {
		dash := conv.NewConvDash(src)
		dash.RemoveAllDashes()
		dash.AddDash(dashLength, dashLength)
		stroke := conv.NewConvStroke(dash)
		stroke.SetWidth(lineWidth)
		stroke.SetLineCap(basics.RoundCap)
		ras.AddPath(&convVS{src: stroke}, 0)
	} else {
		stroke := conv.NewConvStroke(src)
		stroke.SetWidth(lineWidth)
		stroke.SetLineCap(basics.RoundCap)
		ras.AddPath(&convVS{src: stroke}, 0)
	}
}

// drawSolidLine strokes a line in a single solid colour (optionally dashed).
func drawSolidLine(
	ras *rasType,
	sl *scanline.ScanlineU8,
	rb renBaseType,
	x1, y1, x2, y2, lineWidth, dashLength float64,
	c color.RGBA8[color.Linear],
) {
	rasAddLine(ras, x1, y1, x2, y2, lineWidth, dashLength)
	renscan.RenderScanlinesAASolid(ras, sl, rb, c)
}

// drawGradientLine strokes a line using the shared gradient span generator.
// The caller must have filled the shared LUT before calling this function.
func drawGradientLine(
	ras *rasType,
	sl *scanline.ScanlineU8,
	rb renBaseType,
	alloc gradAllocType,
	gradMtx *transform.TransAffine,
	sg gradSpanGenType,
	x1, y1, x2, y2, lineWidth, dashLength float64,
) {
	calcLinearGradientTransform(gradMtx, x1, y1, x2, y2)
	rasAddLine(ras, x1, y1, x2, y2, lineWidth, dashLength)
	renscan.RenderScanlinesAA(ras, sl, rb, alloc, sg)
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
	mainRb.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineU8()

	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	whiteAlpha := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 51} // 0.2 * 255

	// Shared gradient infrastructure — reused across all gradient draw calls.
	gradMtx := transform.NewTransAffine()
	lut := &mutableLUT{}
	adapter := span.NewSRGBColorAdapter(lut)
	interpolator := span.NewSpanInterpolatorLinearDefault(gradMtx)
	spanGen := span.NewSpanGradient(interpolator, span.GradientLinearX{}, adapter, 0.0, 100.0)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0
	radius := math.Min(cx, cy)

	// Radial line test: 180 lines from centre outward.
	// Lines with i < 90 are dashed (dash_length = i), matching C++ original.
	for i := 180; i > 0; i-- {
		n := 2.0 * basics.Pi * float64(i) / 180.0
		dashLen := 0.0
		if i < 90 {
			dashLen = float64(i)
		}
		drawSolidLine(ras, sl, mainRb,
			cx+radius*math.Sin(n), cy+radius*math.Cos(n),
			cx, cy,
			1.0, dashLen, whiteAlpha)
	}

	for i := 1; i <= 20; i++ {
		fi := float64(i)

		// Integral point sizes 1..20.
		ell := shapes.NewEllipseWithParams(
			20+fi*(fi+1)+0.5, 20.5,
			fi/2.0, fi/2.0,
			uint32(8+i), false,
		)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, white)

		// Fractional point sizes 0..2.
		ell2 := shapes.NewEllipseWithParams(
			18+fi*4+0.5, 33+0.5,
			fi/20.0, fi/20.0,
			8, false,
		)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell2}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, white)

		// Fractional point positioning.
		ell3 := shapes.NewEllipseWithParams(
			18+fi*4+(fi-1)/10.0+0.5,
			27+(fi-1)/10.0+0.5,
			0.5, 0.5, 8, false,
		)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell3}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, white)

		// Integral line widths 1..20 — gradient white → end colour.
		endC := color.RGBA8[color.Linear]{
			R: uint8(float64(i%2) * 255),
			G: uint8(float64(i%3) * 0.5 * 255),
			B: uint8(float64(i%5) * 0.25 * 255),
			A: 255,
		}
		lut.fillColors(white, endC)
		x1 := 20 + fi*(fi+1)
		y1 := 40.5
		x2 := 20 + fi*(fi+1) + (fi-1)*4
		y2 := 100.5
		drawGradientLine(ras, sl, mainRb, alloc, gradMtx, spanGen, x1, y1, x2, y2, fi, 0)

		// Fractional line lengths H — gradient red → blue.
		lut.fillColors(
			color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255},
			color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255},
		)
		x1 = 17.5 + fi*4
		y1 = 107
		x2 = 17.5 + fi*4 + fi/6.66666667
		y2 = 107
		drawGradientLine(ras, sl, mainRb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 1.0, 0)

		// Fractional line lengths V — gradient red → blue.
		x1 = 18 + fi*4
		y1 = 112.5
		x2 = 18 + fi*4
		y2 = 112.5 + fi/6.66666667
		drawGradientLine(ras, sl, mainRb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 1.0, 0)

		// Fractional line positioning — gradient red → white.
		lut.fillColors(
			color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255},
			white,
		)
		x1 = 21.5
		y1 = 120 + (fi-1)*3.1
		x2 = 52.5
		y2 = 120 + (fi-1)*3.1
		drawGradientLine(ras, sl, mainRb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 1.0, 0)

		// Fractional line width 2..0 — gradient green → white.
		lut.fillColors(
			color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255},
			white,
		)
		x1 = 52.5
		y1 = 118 + fi*3
		x2 = 83.5
		y2 = 118 + fi*3
		drawGradientLine(ras, sl, mainRb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 2.0-(fi-1)/10.0, 0)

		// Stippled fractional width 2..0 — gradient blue → white, dashed 3 px.
		lut.fillColors(
			color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255},
			white,
		)
		x1 = 83.5
		y1 = 119 + fi*3
		x2 = 114.5
		y2 = 119 + fi*3
		drawGradientLine(ras, sl, mainRb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 2.0-(fi-1)/10.0, 3.0)

		// Integral line width, horz aligned (solid white, mipmap test).
		if i <= 10 {
			drawSolidLine(ras, sl, mainRb,
				125.5, 119.5+float64(i+2)*(fi/2.0),
				135.5, 119.5+float64(i+2)*(fi/2.0),
				fi, 0, white)
		}

		// Fractional line width 0..2, 1 px H (solid white).
		drawSolidLine(ras, sl, mainRb,
			17.5+fi*4, 192, 18.5+fi*4, 192,
			fi/10.0, 0, white)

		// Fractional line positioning, 1 px H (solid white).
		drawSolidLine(ras, sl, mainRb,
			17.5+fi*4+(fi-1)/10.0, 186,
			18.5+fi*4+(fi-1)/10.0, 186,
			1.0, 0, white)
	}

	// Triangles — gradient white → end colour, matching C++ original.
	for i := 1; i <= 13; i++ {
		fi := float64(i)
		endC := color.RGBA8[color.Linear]{
			R: uint8(float64(i%2) * 255),
			G: uint8(float64(i%3) * 0.5 * 255),
			B: uint8(float64(i%5) * 0.25 * 255),
			A: 255,
		}
		lut.fillColors(white, endC)
		x1 := float64(w) - 150
		y1 := float64(h) - 20 - fi*(fi+1.5)
		x2 := float64(w) - 20
		y2 := float64(h) - 20 - fi*(fi+1)
		calcLinearGradientTransform(gradMtx, x1, y1, x2, y2)
		ras.Reset()
		ras.MoveToD(x1, y1)
		ras.LineToD(x2, y2)
		ras.LineToD(float64(w)-20, float64(h)-20-fi*(fi+2))
		renscan.RenderScanlinesAA(ras, sl, mainRb, alloc, spanGen)
	}

	// No y-flip: C++ aa_test uses flip_y=false.
	copy(img.Data, workBuf)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "AA Test",
		Width:                 frameWidth,
		Height:                frameHeight,
		EncodeLinearRGBToSRGB: false,
	}, &demo{})
}
