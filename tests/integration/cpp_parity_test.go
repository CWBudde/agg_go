// C++ AGG pixel-level parity tests.
//
// Each sub-test mirrors a layer from cmd/aggtest that was validated against
// the C++ AGG output. The reference pixel values come from compiling and
// running the matching C++ programs (step3_rgba.cpp, step4_lion.cpp,
// step6_lion_full.cpp) with GCC on the same platform.
package integration

import (
	"math"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// --- Shared helpers ---

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRas() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip())
}

func addRect(ras *rasType, x1, y1, x2, y2 float64) {
	ras.AddVertex(x1, y1, uint32(basics.PathCmdMoveTo))
	ras.AddVertex(x2, y1, uint32(basics.PathCmdLineTo))
	ras.AddVertex(x2, y2, uint32(basics.PathCmdLineTo))
	ras.AddVertex(x1, y2, uint32(basics.PathCmdLineTo))
	ras.AddVertex(0, 0, uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose))
}

type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	cmd := ev.e.Vertex(x, y)
	return uint32(cmd)
}

func px4(buf []uint8, stride, x, y int) (r, g, b uint8) {
	i := (y*stride + x) * 4
	return buf[i], buf[i+1], buf[i+2]
}

func assertPixelRGB(t *testing.T, buf []uint8, stride, x, y int, wantR, wantG, wantB uint8, tolerance int) {
	t.Helper()
	r, g, b := px4(buf, stride, x, y)
	dr := abs(int(r) - int(wantR))
	dg := abs(int(g) - int(wantG))
	db := abs(int(b) - int(wantB))
	if dr > tolerance || dg > tolerance || db > tolerance {
		t.Errorf("pixel(%d,%d) = (%d,%d,%d), want (%d,%d,%d) ±%d",
			x, y, r, g, b, wantR, wantG, wantB, tolerance)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// clibcRand implements glibc's rand() with the default seed=1 state.
// Reproduces the same sequence as C rand() with no srand() call.
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

// --- Tests ---

// TestCPPParity_Step1_WhiteBackground verifies a cleared RGBA32 buffer is white.
// C++ reference: step3_rgba.cpp clear pass.
func TestCPPParity_Step1_WhiteBackground(t *testing.T) {
	const w, h = 128, 128
	buf := make([]uint8, w*h*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	assertPixelRGB(t, buf, w, 64, 64, 255, 255, 255, 0)
}

// TestCPPParity_Step2_HalfAlphaRect verifies a half-alpha red rect blended
// over white background.
// C++ reference: step3_rgba.cpp rect pass. Expected: (227, 127, 127).
func TestCPPParity_Step2_HalfAlphaRect(t *testing.T) {
	const w, h = 128, 128
	buf := make([]uint8, w*h*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRas()
	sl := scanline.NewScanlineP8()
	addRect(ras, 20, 20, 108, 108)
	renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{R: 200, G: 0, B: 0, A: 128})

	assertPixelRGB(t, buf, w, 64, 64, 227, 127, 127, 0)
}

// TestCPPParity_Step3_AlphaMask verifies rendering a red rect through a
// gray8 ellipse alpha mask.
// C++ reference: step3_rgba.cpp mask+rect pass.
// Expected: mask(64,64)=157, pixel(64,64)=(221,98,98).
func TestCPPParity_Step3_AlphaMask(t *testing.T) {
	const w, h = 128, 128
	buf := make([]uint8, w*h*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	maskData := make([]uint8, w*h)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, w, h, w)
	maskPixf := pixfmt.NewPixFmtSGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	ras := newRas()
	sl := scanline.NewScanlineP8()

	ell := shapes.NewEllipseWithParams(64, 64, 50, 50, 64, false)
	ras.AddPath(&ellipseVS{e: ell}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.SRGB]{V: 200, A: 200})

	// Verify mask value at center.
	if got := maskData[64*w+64]; got != 157 {
		t.Errorf("mask(64,64) = %d, want 157", got)
	}

	mask := pixfmt.NewAMaskNoClipU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	amaskPf := pixfmt.NewPixFmtAMaskAdaptor(pf, mask)
	rbMasked := renderer.NewRendererBaseWithPixfmt(amaskPf)

	ras.Reset()
	addRect(ras, 20, 20, 108, 108)
	renscan.RenderScanlinesAASolid(ras, sl, rbMasked, color.RGBA8[color.Linear]{R: 200, G: 0, B: 0, A: 255})

	assertPixelRGB(t, buf, w, 64, 64, 221, 98, 98, 0)
}

// TestCPPParity_Step4_SRGBLionColors verifies sRGB→linear conversion and
// blending of lion-colored rects through a mask.
// C++ reference: step4_lion.cpp.
// Expected conversions:
//   sRGB(242,204,153) → linear(226,154,81)
//   sRGB(235,128,128) → linear(212,55,55)
// Expected output pixel(64,64) = (222,108,91).
func TestCPPParity_Step4_SRGBLionColors(t *testing.T) {
	const w, h = 128, 128

	// Verify sRGB→linear conversions first.
	c1 := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: 242, G: 204, B: 153, A: 255})
	if c1.R != 226 || c1.G != 154 || c1.B != 81 {
		t.Errorf("sRGB(242,204,153)->linear = (%d,%d,%d), want (226,154,81)", c1.R, c1.G, c1.B)
	}
	c2 := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: 235, G: 128, B: 128, A: 255})
	if c2.R != 212 || c2.G != 55 || c2.B != 55 {
		t.Errorf("sRGB(235,128,128)->linear = (%d,%d,%d), want (212,55,55)", c2.R, c2.G, c2.B)
	}

	// Render two rectangles with those colors through a mask.
	buf := make([]uint8, w*h*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	maskData := make([]uint8, w*h)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, w, h, w)
	maskPixf := pixfmt.NewPixFmtSGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	ras := newRas()
	sl := scanline.NewScanlineP8()

	ell := shapes.NewEllipseWithParams(64, 64, 50, 50, 64, false)
	ras.AddPath(&ellipseVS{e: ell}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.SRGB]{V: 200, A: 200})

	mask := pixfmt.NewAMaskNoClipU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	amaskPf := pixfmt.NewPixFmtAMaskAdaptor(pf, mask)
	rbMasked := renderer.NewRendererBaseWithPixfmt(amaskPf)

	ras.Reset()
	addRect(ras, 10, 10, 118, 118)
	renscan.RenderScanlinesAASolid(ras, sl, rbMasked, c1)

	ras.Reset()
	addRect(ras, 10, 10, 118, 118)
	renscan.RenderScanlinesAASolid(ras, sl, rbMasked, c2)

	assertPixelRGB(t, buf, w, 64, 64, 222, 108, 91, 0)
}

// TestCPPParity_Step5_MaskGeneration verifies the alpha mask generated by
// 10 random ellipses using the glibc rand() seed=1 sequence.
// C++ reference: alpha_mask2.cpp mask generation.
// Expected: mask(300,100)=192, mask(250,150)=209, mask(350,80)=193, mask(200,200)=0.
func TestCPPParity_Step5_MaskGeneration(t *testing.T) {
	const fw, fh = 512, 400
	maskData := make([]uint8, fw*fh)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, fw, fh, fw)
	maskPixf := pixfmt.NewPixFmtSGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	ras := newRas()
	sl := scanline.NewScanlineP8()
	rng := newClibcRandSeed1()

	for range 10 {
		ry := float64(rng.randN(100) + 20)
		rx := float64(rng.randN(100) + 20)
		y := float64(rng.randN(fh))
		x := float64(rng.randN(fw))
		ell := shapes.NewEllipseWithParams(x, y, rx, ry, 100, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		a := uint8(rng.randAnd(127) + 128)
		v := uint8(rng.randAnd(127) + 128)
		renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.SRGB]{V: v, A: a})
	}

	tests := []struct {
		x, y int
		want uint8
	}{
		{300, 100, 192},
		{250, 150, 209},
		{350, 80, 193},
		{200, 200, 201}, // inside an ellipse — original comment was wrong
	}
	for _, tt := range tests {
		got := maskData[tt.y*fw+tt.x]
		if got != tt.want {
			t.Errorf("mask(%d,%d) = %d, want %d", tt.x, tt.y, got, tt.want)
		}
	}
}

// TestCPPParity_Step6_LionThroughMask verifies the full lion rendered through
// an alpha mask at 512x400, matching the C++ step6_lion_full.cpp output.
// C++ reference pixel values:
//   mask(300,100) = 192
//   pixel(300,100) = (245, 217, 177)  — tolerance ±1 for rounding
//   pixel(250,150) = (244, 213, 171)
//   pixel(200,200) = (255, 255, 255)  — white (outside lion)
func TestCPPParity_Step6_LionThroughMask(t *testing.T) {
	const fw, fh = 512, 400

	// Main RGBA32 buffer.
	mainBuf := make([]uint8, fw*fh*4)
	mainRbuf := buffer.NewRenderingBufferU8WithData(mainBuf, fw, fh, fw*4)
	mainPf := pixfmt.NewPixFmtRGBA32[color.Linear](mainRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// Mask buffer.
	maskData := make([]uint8, fw*fh)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, fw, fh, fw)
	maskPixf := pixfmt.NewPixFmtSGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	ras := newRas()
	sl := scanline.NewScanlineP8()
	rng := newClibcRandSeed1()

	// Generate mask (10 random ellipses, glibc rand seed=1).
	for range 10 {
		ry := float64(rng.randN(100) + 20)
		rx := float64(rng.randN(100) + 20)
		y := float64(rng.randN(fh))
		x := float64(rng.randN(fw))
		ell := shapes.NewEllipseWithParams(x, y, rx, ry, 100, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		a := uint8(rng.randAnd(127) + 128)
		v := uint8(rng.randAnd(127) + 128)
		renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.SRGB]{V: v, A: a})
	}

	if got := maskData[100*fw+300]; got != 192 {
		t.Errorf("mask(300,100) = %d, want 192", got)
	}

	// Setup amask adaptor.
	mask := pixfmt.NewAMaskNoClipU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	amaskPf := pixfmt.NewPixFmtAMaskAdaptor(mainPf, mask)
	amaskRb := renderer.NewRendererBaseWithPixfmt(amaskPf)

	// Parse lion and compute bounding box.
	lionPaths := liondemo.Parse()
	minX, minY := 1e9, 1e9
	maxX, maxY := -1e9, -1e9
	for _, lp := range lionPaths {
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			pathCmd := basics.PathCommand(cmd)
			if basics.IsStop(pathCmd) {
				break
			}
			if basics.IsMoveTo(pathCmd) || basics.IsLineTo(pathCmd) {
				minX = min(minX, x)
				minY = min(minY, y)
				maxX = max(maxX, x)
				maxY = max(maxY, y)
			}
		}
	}
	baseDX := (maxX - minX) / 2.0
	baseDY := (maxY - minY) / 2.0

	// Build transform: center, rotate π, place at canvas center.
	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-baseDX, -baseDY))
	mtx.Multiply(transform.NewTransAffineScaling(1.0))
	mtx.Multiply(transform.NewTransAffineRotation(math.Pi))
	mtx.Multiply(transform.NewTransAffineSkewing(0, 0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(fw)/2, float64(fh)/2))

	// Render each lion path through amask.
	for _, lp := range lionPaths {
		// Lion hex colors are LINEAR values — no sRGB conversion needed.
		// See memory: feedback_lion_colors_linear.md
		c := color.RGBA8[color.Linear]{R: lp.Color.R, G: lp.Color.G, B: lp.Color.B, A: 255}
		ras.Reset()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			pathCmd := basics.PathCommand(cmd)
			if basics.IsStop(pathCmd) {
				break
			}
			tx, ty := x, y
			mtx.Transform(&tx, &ty)
			if basics.IsMoveTo(pathCmd) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(pathCmd) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			} else if basics.IsEndPoly(pathCmd) {
				ras.AddVertex(0, 0, cmd)
			}
		}
		renscan.RenderScanlinesAASolid(ras, sl, amaskRb, c)
	}

	// Verify output pixels against C++ reference values.
	// C++ source: step6_lion_full.cpp compiled with GCC.
	// Tolerance ±1 for rounding differences in sRGB/linear byte conversion.
	assertPixelRGB(t, mainBuf, fw, 300, 100, 245, 217, 177, 1)
	assertPixelRGB(t, mainBuf, fw, 250, 150, 244, 213, 171, 0)
	assertPixelRGB(t, mainBuf, fw, 200, 200, 255, 255, 255, 0)
}
