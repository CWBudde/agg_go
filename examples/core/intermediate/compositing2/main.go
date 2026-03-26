// Port of AGG C++ compositing2.cpp – compositing modes with radial gradient ramps.
//
// Renders one large radial gradient (ramp1, difference blend) and four smaller
// radial gradients (ramp2, selected comp-op) on a white background.
// Default: alpha_dst=1.0, alpha_src=1.0, comp_op=src-over (item 3).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	frameWidth  = 600
	frameHeight = 400
)

// generateColorRamp fills a [256]RGBA8 array with a 4-stop piecewise-linear ramp
// exactly as compositing2.cpp does.
func generateColorRamp(alpha float64) [256]color.RGBA8[color.Linear] {
	lerp := func(a, b color.RGBA8[color.Linear], t float64) color.RGBA8[color.Linear] {
		lerp1 := func(x, y uint8, t float64) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + 0.5) }
		return color.RGBA8[color.Linear]{
			R: lerp1(a.R, b.R, t),
			G: lerp1(a.G, b.G, t),
			B: lerp1(a.B, b.B, t),
			A: lerp1(a.A, b.A, t),
		}
	}
	a8 := uint8(alpha * 255)
	c1 := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: a8}
	c2 := color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: a8}
	c3 := color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: a8}
	c4 := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 0}

	var ramp [256]color.RGBA8[color.Linear]
	for i := 0; i < 85; i++ {
		ramp[i] = lerp(c1, c2, float64(i)/85.0)
	}
	for i := 85; i < 170; i++ {
		ramp[i] = lerp(c2, c3, float64(i-85)/85.0)
	}
	for i := 170; i < 256; i++ {
		ramp[i] = lerp(c3, c4, float64(i-170)/85.0)
	}
	return ramp
}

// arrayColorFunc wraps a [256]RGBA8 array so it satisfies span.ColorFunction.
type arrayColorFunc struct {
	data [256]color.RGBA8[color.Linear]
}

func (a *arrayColorFunc) Size() int { return 256 }
func (a *arrayColorFunc) ColorAt(i int) color.RGBA8[color.Linear] {
	if i < 0 {
		return a.data[0]
	}
	if i >= 256 {
		return a.data[255]
	}
	return a.data[i]
}

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters
// ---------------------------------------------------------------------------
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ellipseVS wraps shapes.Ellipse as rasterizer.VertexSource.
type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	cmd := ev.e.Vertex(x, y)
	return uint32(cmd)
}

// ctrlVS adapts a ctrl widget to the rasterizer VertexSource interface.
type ctrlVS struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (v *ctrlVS) Rewind(id uint32) { v.ctrl.Rewind(uint(id)) }
func (v *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------

type demo struct{}

// radialShape draws a radial gradient ellipse into rb using the given color ramp.
func radialShape(
	compPixf *pixfmt.PixFmtCompositeRGBA32,
	ras *rasType,
	sl *scanline.ScanlineU8,
	ramp *arrayColorFunc,
	x1, y1, x2, y2 float64,
) {
	cx := (x1 + x2) / 2.0
	cy := (y1 + y2) / 2.0
	dx := x2 - x1
	dy := y2 - y1
	r := 0.5 * min(dx, dy)

	gradMtx := transform.NewTransAffineScaling(r / 100.0)
	gradMtx.Multiply(transform.NewTransAffineTranslation(cx, cy))
	gradMtx.Invert()

	interp := span.NewSpanInterpolatorLinearDefault(gradMtx)
	gradFunc := span.GradientRadial{}
	spanGen := span.NewSpanGradient[color.RGBA8[color.Linear]](interp, gradFunc, ramp, 0, 100)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	rb := renderer.NewRendererBaseWithPixfmt(compPixf)

	ell := shapes.NewEllipseWithParams(cx, cy, r, r, 100, false)
	ras.Reset()
	ras.AddPath(&ellipseVS{e: ell}, 0)
	renscan.RenderScanlinesAA(ras, sl, rb, alloc, spanGen)
}

// renderCtrl renders a control widget using the primary (non-composite) renderer.
func renderCtrl(
	ras *rasType,
	sl *scanline.ScanlineU8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255.0 + 0.5)
	}
	toLinear := func(c color.RGBA) color.RGBA8[color.Linear] {
		return color.RGBA8[color.Linear]{
			R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A),
		}
	}

	vs := &ctrlVS{ctrl: ctrl}
	for i := range ctrl.NumPaths() {
		ras.Reset()
		ras.AddPath(vs, uint32(i))
		renscan.RenderScanlinesAASolid(ras, sl, rb, toLinear(ctrl.Color(i)))
	}
}

// compOpItems maps radio button index to the corresponding CompOp value.
// Matches the C++ m_comp_op items (minus is commented out).
var compOpItems = []blender.CompOp{
	blender.CompOpClear,      // 0
	blender.CompOpSrc,        // 1
	blender.CompOpDst,        // 2
	blender.CompOpSrcOver,    // 3 (default)
	blender.CompOpDstOver,    // 4
	blender.CompOpSrcIn,      // 5
	blender.CompOpDstIn,      // 6
	blender.CompOpSrcOut,     // 7
	blender.CompOpDstOut,     // 8
	blender.CompOpSrcAtop,    // 9
	blender.CompOpDstAtop,    // 10
	blender.CompOpXor,        // 11
	blender.CompOpPlus,       // 12
	blender.CompOpMultiply,   // 13
	blender.CompOpScreen,     // 14
	blender.CompOpOverlay,    // 15
	blender.CompOpDarken,     // 16
	blender.CompOpLighten,    // 17
	blender.CompOpColorDodge, // 18
	blender.CompOpColorBurn,  // 19
	blender.CompOpHardLight,  // 20
	blender.CompOpSoftLight,  // 21
	blender.CompOpDifference, // 22
	blender.CompOpExclusion,  // 23
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer (y-down, flip at end since flip_y=true in C++)
	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)

	// Primary pixfmt for clearing white and rendering controls
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// Composite pixfmt reusing same buffer
	compPixf := pixfmt.NewPixFmtCompositeRGBA32(workRbuf, blender.CompOpDifference)

	ras := newRasterizer()
	sl := scanline.NewScanlineU8()

	// ---------------------------------------------------------------------------
	// Controls – C++: m_alpha_dst(5,5,400,11,!flip_y), m_alpha_src(5,20,400,26,!flip_y),
	//                 m_comp_op(420,5,590,340,!flip_y)  where flip_y=true → flipY=false
	// ---------------------------------------------------------------------------
	alphaDst := sliderctrl.NewSliderCtrl(5, 5, 400, 11, false)
	alphaDst.SetRange(0.0, 1.0)
	alphaDst.SetValue(1.0)
	alphaDst.SetLabel("Dst Alpha=%.2f")

	alphaSrc := sliderctrl.NewSliderCtrl(5, 20, 400, 26, false)
	alphaSrc.SetRange(0.0, 1.0)
	alphaSrc.SetValue(1.0)
	alphaSrc.SetLabel("Src Alpha=%.2f")

	compOpCtrl := rboxctrl.NewDefaultRboxCtrl(420, 5, 590, 340, false)
	compOpCtrl.SetTextSize(6.8, 0)
	compOpCtrl.AddItem("clear")
	compOpCtrl.AddItem("src")
	compOpCtrl.AddItem("dst")
	compOpCtrl.AddItem("src-over")
	compOpCtrl.AddItem("dst-over")
	compOpCtrl.AddItem("src-in")
	compOpCtrl.AddItem("dst-in")
	compOpCtrl.AddItem("src-out")
	compOpCtrl.AddItem("dst-out")
	compOpCtrl.AddItem("src-atop")
	compOpCtrl.AddItem("dst-atop")
	compOpCtrl.AddItem("xor")
	compOpCtrl.AddItem("plus")
	// "minus" is commented out in C++ original
	compOpCtrl.AddItem("multiply")
	compOpCtrl.AddItem("screen")
	compOpCtrl.AddItem("overlay")
	compOpCtrl.AddItem("darken")
	compOpCtrl.AddItem("lighten")
	compOpCtrl.AddItem("color-dodge")
	compOpCtrl.AddItem("color-burn")
	compOpCtrl.AddItem("hard-light")
	compOpCtrl.AddItem("soft-light")
	compOpCtrl.AddItem("difference")
	compOpCtrl.AddItem("exclusion")
	compOpCtrl.SetCurItem(3) // src-over

	// Resolve comp op from radio button selection
	selectedCompOp := compOpItems[compOpCtrl.CurItem()]

	ramp1 := &arrayColorFunc{data: generateColorRamp(alphaDst.Value())}
	ramp2 := &arrayColorFunc{data: generateColorRamp(alphaSrc.Value())}

	// Large background shape with difference blend
	compPixf.SetCompOp(blender.CompOpDifference)
	radialShape(compPixf, ras, sl, ramp1, 50, 50, 50+320, 50+320)

	// Four small shapes with selected comp-op
	compPixf.SetCompOp(selectedCompOp)
	cx, cy := 50.0, 50.0
	radialShape(compPixf, ras, sl, ramp2, cx+120-70, cy+120-70, cx+120+70, cy+120+70)
	radialShape(compPixf, ras, sl, ramp2, cx+200-70, cy+120-70, cx+200+70, cy+120+70)
	radialShape(compPixf, ras, sl, ramp2, cx+120-70, cy+200-70, cx+120+70, cy+200+70)
	radialShape(compPixf, ras, sl, ramp2, cx+200-70, cy+200-70, cx+200+70, cy+200+70)

	// Render controls on top of the scene
	renderCtrl(ras, sl, mainRb, alphaDst)
	renderCtrl(ras, sl, mainRb, alphaSrc)
	renderCtrl(ras, sl, mainRb, compOpCtrl)

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
		Title:                 "AGG Example. Compositing Modes",
		Width:                 frameWidth,
		Height:                frameHeight,
		EncodeLinearRGBToSRGB: true,
	}, &demo{})
}
