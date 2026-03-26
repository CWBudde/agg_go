// Port of AGG C++ gamma_correction.cpp – "Thin red ellipse / gamma correction".
//
// Shows how the anti-aliasing gamma affects the visual quality of thin
// colored lines rendered over a split dark/light background.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/gamma"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

// --- State ---

var (
	gammaValue    = 1.0
	gammaThick    = 1.0 // line thickness
	gammaContrast = 1.0 // 0=no contrast, 1=full

	// Ellipse radii – updated by mouse drag.
	gammaRX = float64(width) / 3.0
	gammaRY = float64(height) / 3.0
)

// --- Vertex source adapters (local to this demo) ---

// gcConvVS adapts a conv.VertexSource to the rasterizer VertexSource interface.
type gcConvVS struct {
	src conv.VertexSource
}

func (a *gcConvVS) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *gcConvVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// gcEllipseSrc adapts shapes.Ellipse to conv.VertexSource.
type gcEllipseSrc struct {
	ell *shapes.Ellipse
}

func (a *gcEllipseSrc) Rewind(pathID uint) { a.ell.Rewind(uint32(pathID)) }
func (a *gcEllipseSrc) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.ell.Vertex(&x, &y)
	return x, y, cmd
}

// --- Drawing ---

func drawGammaCorrectionDemo() {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)

	w := img.Width()
	h := img.Height()
	cx := w / 2
	cy := h / 2

	dark := gammaContrast
	f := func(v float64) uint8 { return uint8(v*255 + 0.5) }

	// Background: matching C++ copy_bar calls.
	// Left half: dark gray
	ren.CopyBar(0, 0, cx, h, color.RGBA8[color.Linear]{R: f(1.0 - dark), G: f(1.0 - dark), B: f(1.0 - dark), A: 255})
	// Right half: light gray
	ren.CopyBar(cx+1, 0, w, h, color.RGBA8[color.Linear]{R: f(dark), G: f(dark), B: f(dark), A: 255})
	// Bottom half: reddish (overwrites both sides as in C++)
	ren.CopyBar(0, cy+1, w, h, color.RGBA8[color.Linear]{R: 255, G: f(1.0 - dark), B: f(1.0 - dark), A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// Apply gamma to the rasterizer's AA coverage mapping.
	gp := gamma.NewGammaPower(gammaValue)
	ras.SetGamma(gp.Apply)

	// Gamma power curve as a green polyline.
	drawGammaCurveLowLevel(ras, sl, ren, float64(w)/2-128, 50, gammaValue)

	// 5 concentric stroked ellipses: Red, Green, Blue, Black, White.
	type ellipseSpec struct {
		dr  float64
		col color.RGBA8[color.Linear]
	}
	specs := []ellipseSpec{
		{0, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}},
		{5, color.RGBA8[color.Linear]{R: 0, G: 200, B: 0, A: 255}},
		{10, color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255}},
		{15, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}},
		{20, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}},
	}

	fcx := float64(w) / 2
	fcy := float64(h) / 2

	for _, s := range specs {
		ell := shapes.NewEllipseWithParams(fcx, fcy, gammaRX-s.dr, gammaRY-s.dr, 150, false)
		stroke := conv.NewConvStroke(&gcEllipseSrc{ell: ell})
		stroke.SetWidth(gammaThick)

		ras.Reset()
		ras.AddPath(&gcConvVS{src: stroke}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, ren, s.col)
	}

	// No sRGB conversion for this demo.
}

// drawGammaCurveLowLevel draws the gamma power curve as a thin green polyline.
func drawGammaCurveLowLevel(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	ren renscan.BaseRendererInterface[color.RGBA8[color.Linear]],
	startX, startY, g float64,
) {
	const npts = 256
	ps := path.NewPathStorageStl()
	for i := 0; i < npts; i++ {
		v := float64(i) / float64(npts-1)
		gv := math.Pow(v, g)
		px := startX + float64(i)
		py := startY + gv*255.0
		if i == 0 {
			ps.MoveTo(px, py)
		} else {
			ps.LineTo(px, py)
		}
	}

	psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)
	stroke := conv.NewConvStroke(psAdapter)
	stroke.SetWidth(2.0)

	ras.Reset()
	ras.AddPath(&gcConvVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, ren, color.RGBA8[color.Linear]{R: 80, G: 160, B: 80, A: 255})
}

// --- Mouse handlers ---

func handleGammaCorrectionMouseDown(x, y float64) bool {
	handleGammaCorrectionMouseMove(x, y)
	return true
}

func handleGammaCorrectionMouseMove(x, y float64) bool {
	gammaRX = math.Abs(float64(width)/2 - x)
	gammaRY = math.Abs(float64(height)/2 - y)
	if gammaRX < 5 {
		gammaRX = 5
	}
	if gammaRY < 5 {
		gammaRY = 5
	}
	return true
}
