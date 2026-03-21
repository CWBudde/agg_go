// Port of AGG C++ alpha_mask.cpp – alpha-masked lion rendering.
//
// Generates a grayscale alpha mask from random ellipses, then renders the
// lion through it so only the mask's bright regions show the lion colours.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
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

const (
	frameWidth  = 512
	frameHeight = 400
)

// ---------------------------------------------------------------------------
// glibc rand() with default seed (no srand call = seed 1).
// State pre-computed from glibc srand(1) initialization + 310 warmup cycles.
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

// randN returns rand() % n, matching C++ rand() % n.
func (r *clibcRand) randN(n int) int { return int(r.next()) % n }

// randAnd returns rand() & mask, matching C++ rand() & mask.
func (r *clibcRand) randAnd(mask int) int { return int(r.next()) & mask }

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters (bridge internal → renderer/scanline iface)
// ---------------------------------------------------------------------------
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// Vertex-source adapter for shapes.Ellipse
// ---------------------------------------------------------------------------

type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	var vx, vy float64
	cmd := ev.e.Vertex(&vx, &vy)
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct {
	angle, scale float64
	skewX, skewY float64
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer: render with positive stride (y-down).
	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)

	ras := newRasterizer()
	sl := scanline.NewScanlineP8()

	// --- Generate grayscale alpha mask (10 random ellipses) ---
	maskData := make([]uint8, w*h)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, w, h, w)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	// C++ uses no srand call (= seed 1).
	rng := newClibcRandSeed1()
	// C++ argument evaluation order (GCC x86, right-to-left):
	// ell.init(rand()%cx, rand()%cy, rand()%100+20, rand()%100+20, 100) → ry, rx, y, x
	// r.color(sgray8(rand()&0xFF, rand()&0xFF)) → a, v  (alpha arg evaluated first)
	for i := 0; i < 10; i++ {
		ry := float64(rng.randN(100) + 20)
		rx := float64(rng.randN(100) + 20)
		y := float64(rng.randN(h))
		x := float64(rng.randN(w))
		ell := shapes.NewEllipseWithParams(x, y, rx, ry, 100, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		a := uint8(rng.randAnd(0xFF))
		v := uint8(rng.randAnd(0xFF))
		renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.Linear]{V: v, A: a})
	}

	mask := pixfmt.NewAlphaMaskU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})

	// --- White background ---
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// --- Render lion through alpha mask ---
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(mainPixf, mask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	lionPaths := liondemo.Parse()

	// Compute lion bounding rect.
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

	// C++ transform: translate(-baseDX,-baseDY) * scale(s,s) * rotate(angle+pi) * translate(w/2,h/2)
	// With flip_y=true. In Go (y-down), replace rotate(angle+pi) with Scale(-1,1) at default angle=0.
	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-baseDX, -baseDY))
	mtx.Multiply(transform.NewTransAffineScaling(d.scale))
	mtx.Multiply(transform.NewTransAffineRotation(d.angle + math.Pi))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(w)/2, float64(h)/2))

	for _, lp := range lionPaths {
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
			}
		}
		renscan.RenderScanlinesAASolid(ras, sl, rbAMask, c)
	}

	// Copy work buffer to output with y-flip (match C++ flip_y=true).
	copyFlipY(workBuf, img.Data, w, h)
}

// copyFlipY copies RGBA pixels from src to dst with vertical flip.
func copyFlipY(src, dst []uint8, width, height int) {
	stride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

func main() {
	d := &demo{scale: 1.0}
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Alpha Mask",
		Width:  frameWidth,
		Height: frameHeight,
	}, d)
}
