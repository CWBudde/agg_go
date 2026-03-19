// Port of AGG C++ alpha_mask.cpp – alpha-masked lion rendering.
//
// Generates a grayscale alpha mask from random ellipses, then renders the
// lion through it so only the mask's bright regions show the lion colours.
package main

import (
	"math"
	"math/rand"

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
// Rasterizer / scanline adapters (bridge internal → renderer/scanline iface)
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

func (r *rasterizerAdaptor) Reset()              { r.ras.Reset() }
func (r *rasterizerAdaptor) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *rasterizerAdaptor) MinX() int           { return r.ras.MinX() }
func (r *rasterizerAdaptor) MaxX() int           { return r.ras.MaxX() }

func (r *rasterizerAdaptor) AddPath(vs rasterizer.VertexSource, pathID uint32) {
	r.ras.AddPath(vs, pathID)
}

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
func (w *scanlineWrapper) NumSpans() int         { return w.sl.NumSpans() }

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
	angle, scale   float64
	skewX, skewY   float64
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer: render with positive stride (y-down).
	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)

	ras := newRasterizer()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}

	// --- Generate grayscale alpha mask (10 random ellipses) ---
	maskData := make([]uint8, w*h)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, w, h, w)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	// C++ uses srand() default seed = implicit 1 from libc.
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 10; i++ {
		ell := shapes.NewEllipseWithParams(
			float64(rng.Intn(w)),
			float64(rng.Intn(h)),
			float64(rng.Intn(100)+20),
			float64(rng.Intn(100)+20),
			100, false,
		)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		v := uint8(rng.Intn(256))
		a := uint8(rng.Intn(256))
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
		c := color.RGBA8[color.Linear]{R: lp.Color[0], G: lp.Color[1], B: lp.Color[2], A: 255}
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
