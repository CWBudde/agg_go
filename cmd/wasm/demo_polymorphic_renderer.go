// Port of AGG's polymorphic_renderer.cpp.
//
// Demonstrates the Go equivalent of C++ virtual-dispatch polymorphism:
// the same rendering code operates uniformly on different pixel-format
// backends through a Go interface, without virtual keyword or base class.
//
// Visual: a single filled triangle on a white background. Drag the three
// corner handles to reshape it.
package main

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
)

// --- State ---

var (
	polyRenX        = [3]float64{100, 369, 143}
	polyRenY        = [3]float64{60, 170, 310}
	polyRenSelected = -1
	polyRenDragDX   = 0.0
	polyRenDragDY   = 0.0
)

// --- Polymorphic renderer types ---

// polyRenSolidRenderer is the Go equivalent of C++'s polymorphic_renderer_solid_rgba8_base.
// Any pixel-format backend that implements these methods is a valid renderer.
type polyRenSolidRenderer interface {
	Clear(c color.RGBA8[color.SRGB])
	SetColor(c color.RGBA8[color.SRGB])
	Prepare()
	Render(sl renscan.ScanlineInterface)
}

// polyRenRGB555Renderer is backed by PixFmtRGB555, matching C++'s pix_format_rgb555.
// Internally it uses linear RGBA8 colors. The interface accepts sRGB colors and
// this adaptor converts sRGB→linear before passing to the renderer.
type polyRenRGB555Renderer struct {
	renBase *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]
	ren     *renscan.RendererScanlineAASolid[*renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]
}

func newPolyRenRGB555Renderer(w, h int) (*polyRenRGB555Renderer, []basics.Int16u) {
	buf16 := make([]basics.Int16u, w*h)
	rbuf16 := buffer.NewRenderingBufferU16WithData(buf16, w, h, w*2) // positive stride = y-down (no flip)
	pf := pixfmt.NewPixFmtRGB555(rbuf16, blender.BlenderRGB555{})
	rb := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)
	ren := renscan.NewRendererScanlineAASolidWithRenderer(rb)
	return &polyRenRGB555Renderer{renBase: rb, ren: ren}, buf16
}

func (r *polyRenRGB555Renderer) Clear(c color.RGBA8[color.SRGB]) {
	r.renBase.Clear(color.ConvertToLinear(c))
}

func (r *polyRenRGB555Renderer) SetColor(c color.RGBA8[color.SRGB]) {
	r.ren.SetColor(color.ConvertToLinear(c))
}

func (r *polyRenRGB555Renderer) Prepare()                            { r.ren.Prepare() }
func (r *polyRenRGB555Renderer) Render(sl renscan.ScanlineInterface) { r.ren.Render(sl) }

// --- Rendering ---

func drawPolymorphicRendererDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	// Render into a uint16 RGB555 scratch buffer (positive stride = y-down, matching canvas).
	ren, buf16 := newPolyRenRGB555Renderer(w, h)
	var sr polyRenSolidRenderer = ren

	// Build the triangle path.
	ps := path.NewPathStorageStl()
	ps.MoveTo(polyRenX[0], polyRenY[0])
	ps.LineTo(polyRenX[1], polyRenY[1])
	ps.LineTo(polyRenX[2], polyRenY[2])
	ps.ClosePolygon(basics.PathFlagsNone)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AddPath(&pathSourceAdapter{ps: ps}, 0)

	sl := scanline.NewScanlineP8()

	// Polymorphic dispatch: same code works with any polyRenSolidRenderer,
	// just as the C++ version works with any PixFmt.
	sr.Clear(color.RGBA8[color.SRGB]{R: 255, G: 255, B: 255, A: 255})
	sr.SetColor(color.RGBA8[color.SRGB]{R: 80, G: 30, B: 20, A: 255})
	renscan.RenderScanlines(ras, sl, sr)

	// Convert the RGB555 uint16 buffer to RGBA8 for display.
	// No sRGB conversion needed (DisableLinearRGBToSRGB: true in standalone).
	for i, pix := range buf16 {
		r, g, b := pixfmt.UnpackPixel555(pix)
		img.Data[i*4+0] = r
		img.Data[i*4+1] = g
		img.Data[i*4+2] = b
		img.Data[i*4+3] = 255
	}

	// Draw interactive vertex handles.
	for i := 0; i < 3; i++ {
		drawHandle(polyRenX[i], polyRenY[i])
	}
}

// --- Mouse handlers ---

func handlePolyRenMouseDown(x, y float64) bool {
	polyRenSelected = -1
	for i := 0; i < 3; i++ {
		dx := x - polyRenX[i]
		dy := y - polyRenY[i]
		if math.Sqrt(dx*dx+dy*dy) < 10 {
			polyRenSelected = i
			polyRenDragDX = dx
			polyRenDragDY = dy
			return true
		}
	}
	return false
}

func handlePolyRenMouseMove(x, y float64) bool {
	if polyRenSelected < 0 {
		return false
	}
	polyRenX[polyRenSelected] = x - polyRenDragDX
	polyRenY[polyRenSelected] = y - polyRenDragDY
	return true
}

func handlePolyRenMouseUp() {
	polyRenSelected = -1
}
