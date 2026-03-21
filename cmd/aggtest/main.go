// Layered comparison tests against C++ AGG output.
// Mirrors step3_rgba.cpp / step4_lion.cpp.
// Run with: go run ./cmd/aggtest/
package main

import (
	"fmt"
	"image"
	gocolor "image/color"
	"image/png"
	"math"
	"os"

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

const W, H = 128, 128

func savePNG4(fname string, buf []uint8, w, h int) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			i := (y*w + x) * 4
			img.SetRGBA(x, y, gocolor.RGBA{R: buf[i], G: buf[i+1], B: buf[i+2], A: 255})
		}
	}
	f, _ := os.Create(fname)
	defer f.Close()
	png.Encode(f, img)
}

func px4(buf []uint8, x, y int) (r, g, b uint8) {
	i := (y*W + x) * 4
	return buf[i], buf[i+1], buf[i+2]
}

// ---------------------------------------------------------------------------
// Rasterizer/scanline adapters — same pattern as alpha_mask2/main.go
// ---------------------------------------------------------------------------

type rasterizerAdaptor struct {
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	sl  rasScanlineAdaptor
}

func newRas() *rasterizerAdaptor {
	return &rasterizerAdaptor{
		ras: rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
			rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip()),
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

func (r *rasterizerAdaptor) addRect(x1, y1, x2, y2 float64) {
	r.ras.AddVertex(x1, y1, uint32(basics.PathCmdMoveTo))
	r.ras.AddVertex(x2, y1, uint32(basics.PathCmdLineTo))
	r.ras.AddVertex(x2, y2, uint32(basics.PathCmdLineTo))
	r.ras.AddVertex(x1, y2, uint32(basics.PathCmdLineTo))
	r.ras.AddVertex(0, 0, uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose))
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

type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	cmd := ev.e.Vertex(x, y)
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Step 1: plain white background
// ---------------------------------------------------------------------------
func step1() {
	buf := make([]uint8, W*H*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, W, H, W*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})
	savePNG4("/tmp/aggtest/step1_bg_go.png", buf, W, H)
	r, g, b := px4(buf, 64, 64)
	fmt.Printf("  pixel(64,64) = %d %d %d  [C++ expects 255 255 255]\n", r, g, b)
}

// ---------------------------------------------------------------------------
// Step 2: white background + red half-alpha rectangle
// ---------------------------------------------------------------------------
func step2() {
	buf := make([]uint8, W*H*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, W, H, W*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRas()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}
	ras.addRect(20, 20, 108, 108)
	renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{R: 200, G: 0, B: 0, A: 128})

	savePNG4("/tmp/aggtest/step2_rect_go.png", buf, W, H)
	r, g, b := px4(buf, 64, 64)
	fmt.Printf("  pixel(64,64) = %d %d %d  [C++ expects 227 127 127]\n", r, g, b)
}

// ---------------------------------------------------------------------------
// Step 3: white + gray8 mask (ellipse) + red rect through mask
// Mirrors step3_rgba.cpp exactly.
// ---------------------------------------------------------------------------
func step3() {
	buf := make([]uint8, W*H*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, W, H, W*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	maskData := make([]uint8, W*H)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, W, H, W)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	ras := newRas()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}

	ell := shapes.NewEllipseWithParams(64, 64, 50, 50, 64, false)
	ras.ras.AddPath(&ellipseVS{e: ell}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.Linear]{V: 200, A: 200})
	fmt.Printf("  mask(64,64) = %d  [C++ expects 157]\n", maskData[64*W+64])

	mask := pixfmt.NewAlphaMaskU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	amaskPf := pixfmt.NewPixFmtAMaskAdaptor(pf, mask)
	rbMasked := renderer.NewRendererBaseWithPixfmt(amaskPf)

	ras.Reset()
	ras.addRect(20, 20, 108, 108)
	renscan.RenderScanlinesAASolid(ras, sl, rbMasked, color.RGBA8[color.Linear]{R: 200, G: 0, B: 0, A: 255})

	savePNG4("/tmp/aggtest/step3_amask_go.png", buf, W, H)
	r, g, b := px4(buf, 64, 64)
	fmt.Printf("  out(64,64) = %d %d %d  [C++ expects 221 98 98]\n", r, g, b)
}

// ---------------------------------------------------------------------------
// Step 4: two lion-colored rects through mask, sRGB->linear conversion applied
// Mirrors step4_lion.cpp.
// ---------------------------------------------------------------------------
func step4() {
	buf := make([]uint8, W*H*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, W, H, W*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	maskData := make([]uint8, W*H)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, W, H, W)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	ras := newRas()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}

	ell := shapes.NewEllipseWithParams(64, 64, 50, 50, 64, false)
	ras.ras.AddPath(&ellipseVS{e: ell}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.Linear]{V: 200, A: 200})

	mask := pixfmt.NewAlphaMaskU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	amaskPf := pixfmt.NewPixFmtAMaskAdaptor(pf, mask)
	rbMasked := renderer.NewRendererBaseWithPixfmt(amaskPf)

	// sRGB(242,204,153) -> linear
	c1 := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: 242, G: 204, B: 153, A: 255})
	fmt.Printf("  sRGB(242,204,153)->linear(%d,%d,%d)  [C++ expects 226,154,81]\n", c1.R, c1.G, c1.B)
	ras.Reset()
	ras.addRect(10, 10, 118, 118)
	renscan.RenderScanlinesAASolid(ras, sl, rbMasked, c1)

	// sRGB(235,128,128) -> linear
	c2 := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: 235, G: 128, B: 128, A: 255})
	fmt.Printf("  sRGB(235,128,128)->linear(%d,%d,%d)  [C++ expects 212,55,55]\n", c2.R, c2.G, c2.B)
	ras.Reset()
	ras.addRect(10, 10, 118, 118)
	renscan.RenderScanlinesAASolid(ras, sl, rbMasked, c2)

	savePNG4("/tmp/aggtest/step4_lion_go.png", buf, W, H)
	r, g, b := px4(buf, 64, 64)
	fmt.Printf("  out(64,64) = %d %d %d  [C++ expects 222 108 91]\n", r, g, b)
}

func main() {
	fmt.Println("=== Step 1: plain background ===")
	step1()
	fmt.Println("=== Step 2: background + half-alpha rect ===")
	step2()
	fmt.Println("=== Step 3: background + mask + rect through amask ===")
	step3()
	fmt.Println("=== Step 4: two lion-colored rects through mask ===")
	step4()
	fmt.Println("=== Step 5: full alpha_mask2 mask generation ===")
	step5()
	fmt.Println("=== Step 6: full lion through amask ===")
	step6()
}

// ---------------------------------------------------------------------------
// Step 5: reproduce the full alpha_mask2 mask (10 random ellipses, clibcRand seed1)
// and report mask values at key pixels from the real alpha_mask2 output.
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Step 6: full lion rendered through amask — mirrors step6_lion_full.cpp.
// Reports pixel values for comparison with C++ (expects 245,217,177 at 300,100).
// ---------------------------------------------------------------------------
func step6() {
	const fw, fh = 512, 400

	// Main RGBA32 buffer
	mainBuf := make([]uint8, fw*fh*4)
	mainRbuf := buffer.NewRenderingBufferU8WithData(mainBuf, fw, fh, fw*4)
	mainPf := pixfmt.NewPixFmtRGBA32[color.Linear](mainRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// Mask buffer
	maskData := make([]uint8, fw*fh)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, fw, fh, fw)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	ras := newRas()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}
	rng := newClibcRandSeed1()

	// Generate mask
	for range 10 {
		ry := float64(rng.randN(100) + 20)
		rx := float64(rng.randN(100) + 20)
		y := float64(rng.randN(fh))
		x := float64(rng.randN(fw))
		ell := shapes.NewEllipseWithParams(x, y, rx, ry, 100, false)
		ras.ras.Reset()
		ras.ras.AddPath(&ellipseVS{e: ell}, 0)
		a := uint8(rng.randAnd(127) + 128)
		v := uint8(rng.randAnd(127) + 128)
		renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.Linear]{V: v, A: a})
	}
	fmt.Printf("  mask(300,100) = %d  [C++ expects 192]\n", maskData[100*fw+300])

	// Setup amask adaptor
	mask := pixfmt.NewAlphaMaskU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	amaskPf := pixfmt.NewPixFmtAMaskAdaptor(mainPf, mask)
	amaskRb := renderer.NewRendererBaseWithPixfmt(amaskPf)

	// Parse lion
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
	}
	baseDX := (maxX - minX) / 2.0
	baseDY := (maxY - minY) / 2.0
	fmt.Printf("  npaths=%d bbox=(%.1f,%.1f)-(%.1f,%.1f) base_d=(%.1f,%.1f)\n",
		len(lionPaths), minX, minY, maxX, maxY, baseDX, baseDY)

	// Print first 5 colors (hex values are linear)
	for i := 0; i < 5 && i < len(lionPaths); i++ {
		c := lionPaths[i].Color
		fmt.Printf("  path %d: linear(%d,%d,%d,%d)\n", i, c.R, c.G, c.B, c.A)
	}

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-baseDX, -baseDY))
	mtx.Multiply(transform.NewTransAffineScaling(1.0))
	mtx.Multiply(transform.NewTransAffineRotation(math.Pi))
	mtx.Multiply(transform.NewTransAffineSkewing(0, 0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(fw)/2, float64(fh)/2))

	// Render each lion path through amask
	for _, lp := range lionPaths {
		// Lion hex colors are LINEAR values (C++ rgb8_packed returns rgba8/linear).
		// parse_lion stores into srgba8 array, which roundtrips linear→sRGB→linear,
		// but the net result is the original linear values (within ±1 rounding).
		// So use the hex values directly as linear — no sRGB conversion needed.
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
				ras.ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(pathCmd) {
				ras.ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			} else if basics.IsEndPoly(pathCmd) {
				ras.ras.AddVertex(0, 0, cmd)
			}
		}
		renscan.RenderScanlinesAASolid(ras, sl, amaskRb, c)
	}

	p := func(x, y int) (uint8, uint8, uint8) {
		i := (y*fw + x) * 4
		return mainBuf[i], mainBuf[i+1], mainBuf[i+2]
	}
	r1, g1, b1 := p(300, 100)
	r2, g2, b2 := p(250, 150)
	r3, g3, b3 := p(200, 200)
	fmt.Printf("  out(300,100) = %d %d %d  [C++ expects 245 217 177]\n", r1, g1, b1)
	fmt.Printf("  out(250,150) = %d %d %d  [C++ expects 244 213 171]\n", r2, g2, b2)
	fmt.Printf("  out(200,200) = %d %d %d  [C++ expects 255 255 255]\n", r3, g3, b3)

	savePNG4("/tmp/aggtest/step6_lion_go.png", mainBuf, fw, fh)
}

func step5() {
	const fw, fh = 512, 400
	maskData := make([]uint8, fw*fh)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, fw, fh, fw)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	ras := newRas()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}
	rng := newClibcRandSeed1()

	for range 10 {
		ry := float64(rng.randN(100) + 20)
		rx := float64(rng.randN(100) + 20)
		y := float64(rng.randN(fh))
		x := float64(rng.randN(fw))
		ell := shapes.NewEllipseWithParams(x, y, rx, ry, 100, false)
		ras.ras.Reset()
		ras.ras.AddPath(&ellipseVS{e: ell}, 0)
		a := uint8(rng.randAnd(127) + 128)
		v := uint8(rng.randAnd(127) + 128)
		renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.Linear]{V: v, A: a})
	}

	// Report mask values at pixels that differ between Go and C++ in alpha_mask2
	fmt.Printf("  mask(300,100) = %d\n", maskData[100*fw+300])
	fmt.Printf("  mask(250,150) = %d\n", maskData[150*fw+250])
	fmt.Printf("  mask(350, 80) = %d\n", maskData[80*fw+350])
	fmt.Printf("  mask(200,200) = %d  [white pixel, expect 0]\n", maskData[200*fw+200])
}
