// Based on the original AGG examples: aa_test.cpp.
//
//go:build js && wasm

package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
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

// Native size of the aa_test demo; centered in the 800×600 wasm canvas.
const (
	aaTestNativeWidth  = 480
	aaTestNativeHeight = 350
	aaTestOffsetX      = (800 - aaTestNativeWidth) / 2  // 160
	aaTestOffsetY      = (600 - aaTestNativeHeight) / 2 // 125
)

// ---------------------------------------------------------------------------
// Local type aliases (mirrors standalone)
// ---------------------------------------------------------------------------

type (
	aaTestRasType     = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	aaTestRenBaseType = *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]
)

func aaTestNewRasterizer() *aaTestRasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// Vertex-source adapters
// ---------------------------------------------------------------------------

type aaTestEllipseVS struct{ e *shapes.Ellipse }

func (ev *aaTestEllipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *aaTestEllipseVS) Vertex(x, y *float64) uint32 {
	var vx, vy float64
	cmd := ev.e.Vertex(&vx, &vy)
	*x, *y = vx, vy
	return uint32(cmd)
}

type aaTestPathStlVS struct{ ps *path.PathStorageStl }

func (a *aaTestPathStlVS) Rewind(id uint) { a.ps.Rewind(id) }
func (a *aaTestPathStlVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

type aaTestConvVS struct{ src conv.VertexSource }

func (a *aaTestConvVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *aaTestConvVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Gradient support
// ---------------------------------------------------------------------------

// aaTestMutableLUT is a 256-entry sRGB colour lookup table.
type aaTestMutableLUT struct {
	lut [256]color.RGBA8[color.SRGB]
}

func (m *aaTestMutableLUT) Size() int                             { return 256 }
func (m *aaTestMutableLUT) ColorAt(i int) color.RGBA8[color.SRGB] { return m.lut[i] }

func (m *aaTestMutableLUT) fillColors(c1, c2 color.RGBA8[color.Linear]) {
	s1 := color.ConvertRGBA8LinearToSRGB(c1)
	s2 := color.ConvertRGBA8LinearToSRGB(c2)
	for i := range m.lut {
		m.lut[i] = s1.Gradient(s2, basics.Int8u(i))
	}
}

// aaTestCalcLinearGradientTransform matches AGG's calc_linear_gradient_transform.
func aaTestCalcLinearGradientTransform(gradMtx *transform.TransAffine, x1, y1, x2, y2 float64) {
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

type (
	aaTestGradInterpType  = *span.SpanInterpolatorLinear[*transform.TransAffine]
	aaTestGradAdapterType = span.SRGBColorAdapter[*aaTestMutableLUT]
	aaTestGradSpanGenType = *span.SpanGradient[color.RGBA8[color.Linear], aaTestGradInterpType, span.GradientLinearX, aaTestGradAdapterType]
	aaTestGradAllocType   = *span.SpanAllocator[color.RGBA8[color.Linear]]
)

// ---------------------------------------------------------------------------
// Line drawing helpers
// ---------------------------------------------------------------------------

func aaTestRasAddLine(ras *aaTestRasType, x1, y1, x2, y2, lineWidth, dashLength float64) {
	ps := path.NewPathStorageStl()
	ps.MoveTo(x1+0.5, y1+0.5)
	ps.LineTo(x2+0.5, y2+0.5)
	src := &aaTestPathStlVS{ps: ps}

	ras.Reset()
	if dashLength > 0.0 {
		dash := conv.NewConvDash(src)
		dash.RemoveAllDashes()
		dash.AddDash(dashLength, dashLength)
		stroke := conv.NewConvStroke(dash)
		stroke.SetWidth(lineWidth)
		stroke.SetLineCap(basics.RoundCap)
		ras.AddPath(&aaTestConvVS{src: stroke}, 0)
	} else {
		stroke := conv.NewConvStroke(src)
		stroke.SetWidth(lineWidth)
		stroke.SetLineCap(basics.RoundCap)
		ras.AddPath(&aaTestConvVS{src: stroke}, 0)
	}
}

func aaTestDrawSolidLine(
	ras *aaTestRasType,
	sl *scanline.ScanlineU8,
	rb aaTestRenBaseType,
	x1, y1, x2, y2, lineWidth, dashLength float64,
	c color.RGBA8[color.Linear],
) {
	aaTestRasAddLine(ras, x1, y1, x2, y2, lineWidth, dashLength)
	renscan.RenderScanlinesAASolid(ras, sl, rb, c)
}

func aaTestDrawGradientLine(
	ras *aaTestRasType,
	sl *scanline.ScanlineU8,
	rb aaTestRenBaseType,
	alloc aaTestGradAllocType,
	gradMtx *transform.TransAffine,
	sg aaTestGradSpanGenType,
	x1, y1, x2, y2, lineWidth, dashLength float64,
) {
	aaTestCalcLinearGradientTransform(gradMtx, x1, y1, x2, y2)
	aaTestRasAddLine(ras, x1, y1, x2, y2, lineWidth, dashLength)
	renscan.RenderScanlinesAA(ras, sl, rb, alloc, sg)
}

// ---------------------------------------------------------------------------
// Demo entry point
// ---------------------------------------------------------------------------

func drawAATestDemo() {
	img := ctx.GetImage()

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras := aaTestNewRasterizer()
	sl := scanline.NewScanlineU8()

	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	whiteAlpha := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 51} // ~0.2 * 255

	// Shared gradient infrastructure.
	gradMtx := transform.NewTransAffine()
	lut := &aaTestMutableLUT{}
	adapter := span.NewSRGBColorAdapter(lut)
	interpolator := span.NewSpanInterpolatorLinearDefault(gradMtx)
	spanGen := span.NewSpanGradient(interpolator, span.GradientLinearX{}, adapter, 0.0, 100.0)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	// Canvas centre in the native 480×350 coordinate frame.
	ox := float64(aaTestOffsetX) // 160
	oy := float64(aaTestOffsetY) // 125

	nw := float64(aaTestNativeWidth)
	nh := float64(aaTestNativeHeight)
	cx := nw / 2.0
	cy := nh / 2.0
	radius := math.Min(cx, cy)

	// Radial line test: 180 lines from centre outward.
	for i := 180; i > 0; i-- {
		n := 2.0 * basics.Pi * float64(i) / 180.0
		dashLen := 0.0
		if i < 90 {
			dashLen = float64(i)
		}
		aaTestDrawSolidLine(ras, sl, rb,
			ox+cx+radius*math.Sin(n), oy+cy+radius*math.Cos(n),
			ox+cx, oy+cy,
			1.0, dashLen, whiteAlpha)
	}

	for i := 1; i <= 20; i++ {
		fi := float64(i)

		// Integral point sizes 1..20.
		ell := shapes.NewEllipseWithParams(
			ox+20+fi*(fi+1)+0.5, oy+20.5,
			fi/2.0, fi/2.0,
			uint32(8+i), false,
		)
		ras.Reset()
		ras.AddPath(&aaTestEllipseVS{e: ell}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, rb, white)

		// Fractional point sizes 0..2.
		ell2 := shapes.NewEllipseWithParams(
			ox+18+fi*4+0.5, oy+33+0.5,
			fi/20.0, fi/20.0,
			8, false,
		)
		ras.Reset()
		ras.AddPath(&aaTestEllipseVS{e: ell2}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, rb, white)

		// Fractional point positioning.
		ell3 := shapes.NewEllipseWithParams(
			ox+18+fi*4+(fi-1)/10.0+0.5,
			oy+27+(fi-1)/10.0+0.5,
			0.5, 0.5, 8, false,
		)
		ras.Reset()
		ras.AddPath(&aaTestEllipseVS{e: ell3}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, rb, white)

		// Integral line widths 1..20 — gradient white → end colour.
		endC := color.RGBA8[color.Linear]{
			R: uint8(float64(i%2) * 255),
			G: uint8(float64(i%3) * 0.5 * 255),
			B: uint8(float64(i%5) * 0.25 * 255),
			A: 255,
		}
		lut.fillColors(white, endC)
		x1 := ox + 20 + fi*(fi+1)
		y1 := oy + 40.5
		x2 := ox + 20 + fi*(fi+1) + (fi-1)*4
		y2 := oy + 100.5
		aaTestDrawGradientLine(ras, sl, rb, alloc, gradMtx, spanGen, x1, y1, x2, y2, fi, 0)

		// Fractional line lengths H — gradient red → blue.
		lut.fillColors(
			color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255},
			color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255},
		)
		x1 = ox + 17.5 + fi*4
		y1 = oy + 107
		x2 = ox + 17.5 + fi*4 + fi/6.66666667
		y2 = oy + 107
		aaTestDrawGradientLine(ras, sl, rb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 1.0, 0)

		// Fractional line lengths V — gradient red → blue.
		x1 = ox + 18 + fi*4
		y1 = oy + 112.5
		x2 = ox + 18 + fi*4
		y2 = oy + 112.5 + fi/6.66666667
		aaTestDrawGradientLine(ras, sl, rb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 1.0, 0)

		// Fractional line positioning — gradient red → white.
		lut.fillColors(
			color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255},
			white,
		)
		x1 = ox + 21.5
		y1 = oy + 120 + (fi-1)*3.1
		x2 = ox + 52.5
		y2 = oy + 120 + (fi-1)*3.1
		aaTestDrawGradientLine(ras, sl, rb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 1.0, 0)

		// Fractional line width 2..0 — gradient green → white.
		lut.fillColors(
			color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255},
			white,
		)
		x1 = ox + 52.5
		y1 = oy + 118 + fi*3
		x2 = ox + 83.5
		y2 = oy + 118 + fi*3
		aaTestDrawGradientLine(ras, sl, rb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 2.0-(fi-1)/10.0, 0)

		// Stippled fractional width 2..0 — gradient blue → white, dashed 3 px.
		lut.fillColors(
			color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255},
			white,
		)
		x1 = ox + 83.5
		y1 = oy + 119 + fi*3
		x2 = ox + 114.5
		y2 = oy + 119 + fi*3
		aaTestDrawGradientLine(ras, sl, rb, alloc, gradMtx, spanGen, x1, y1, x2, y2, 2.0-(fi-1)/10.0, 3.0)

		// Integral line width, horz aligned (solid white, mipmap test).
		if i <= 10 {
			aaTestDrawSolidLine(ras, sl, rb,
				ox+125.5, oy+119.5+float64(i+2)*(fi/2.0),
				ox+135.5, oy+119.5+float64(i+2)*(fi/2.0),
				fi, 0, white)
		}

		// Fractional line width 0..2, 1 px H (solid white).
		aaTestDrawSolidLine(ras, sl, rb,
			ox+17.5+fi*4, oy+192, ox+18.5+fi*4, oy+192,
			fi/10.0, 0, white)

		// Fractional line positioning, 1 px H (solid white).
		aaTestDrawSolidLine(ras, sl, rb,
			ox+17.5+fi*4+(fi-1)/10.0, oy+186,
			ox+18.5+fi*4+(fi-1)/10.0, oy+186,
			1.0, 0, white)
	}

	// Triangles — gradient white → end colour.
	for i := 1; i <= 13; i++ {
		fi := float64(i)
		endC := color.RGBA8[color.Linear]{
			R: uint8(float64(i%2) * 255),
			G: uint8(float64(i%3) * 0.5 * 255),
			B: uint8(float64(i%5) * 0.25 * 255),
			A: 255,
		}
		lut.fillColors(white, endC)
		x1 := ox + nw - 150
		y1 := oy + nh - 20 - fi*(fi+1.5)
		x2 := ox + nw - 20
		y2 := oy + nh - 20 - fi*(fi+1)
		aaTestCalcLinearGradientTransform(gradMtx, x1, y1, x2, y2)
		ras.Reset()
		ras.MoveToD(x1, y1)
		ras.LineToD(x2, y2)
		ras.LineToD(ox+nw-20, oy+nh-20-fi*(fi+2))
		renscan.RenderScanlinesAA(ras, sl, rb, alloc, spanGen)
	}

	// No sRGB conversion: standalone uses EncodeLinearRGBToSRGB: false.
	_ = img
}
