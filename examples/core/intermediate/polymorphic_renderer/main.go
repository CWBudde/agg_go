// Package main ports AGG's polymorphic_renderer.cpp demo.
//
// In the original C++ demo, a virtual base class (polymorphic_renderer_solid_rgba8_base)
// and a template adaptor let one rendering routine work with any pixel-format backend.
// Go interfaces provide this naturally: the same draw call works through any
// implementation of the SolidRenderer interface below, without virtual keyword,
// explicit factory, or heap allocation of C++ base classes.
//
// Visual: a filled triangle on a white background.
// Drag the three vertex handles to reshape it.
//
// The C++ original uses pix_format_rgb555 (15-bit packed pixels); this port matches
// that by rendering into a uint16 scratch buffer and converting to RGBA8 for display.
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	rendsl "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
)

// SolidRenderer is the Go equivalent of C++'s polymorphic_renderer_solid_rgba8_base.
// Any pixel-format backend that implements these methods is a valid renderer.
type SolidRenderer interface {
	Clear(c color.RGBA8[color.SRGB])
	SetColor(c color.RGBA8[color.SRGB])
	Prepare()
	Render(sl rendsl.ScanlineInterface)
}

// rgb555Renderer is backed by PixFmtRGB555, matching C++'s pix_format_rgb555.
// Internally it uses linear RGBA8 colors (matching C++ blender_rgb555::color_type = rgba8).
// The SolidRenderer interface accepts sRGB colors (matching C++ polymorphic_renderer_solid_rgba8_base),
// and this adaptor converts sRGB→linear before passing to the renderer, just as the C++ adaptor
// implicitly converts srgba8→rgba8 via the rgba8T<linear>(rgba8T<sRGB>) constructor.
type rgb555Renderer struct {
	pf      *pixfmt.PixFmtRGB555[blender.BlenderRGB555]
	renBase *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]
	ren     *rendsl.RendererScanlineAASolid[*renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]
}

func newRGB555Renderer(w, h int) (*rgb555Renderer, []basics.Int16u) {
	buf16 := make([]basics.Int16u, w*h)
	rbuf16 := buffer.NewRenderingBufferU16WithData(buf16, w, h, -w*2) // negative stride = flip_y
	pf := pixfmt.NewPixFmtRGB555(rbuf16, blender.BlenderRGB555{})
	rb := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)
	ren := rendsl.NewRendererScanlineAASolidWithRenderer(rb)
	return &rgb555Renderer{pf: pf, renBase: rb, ren: ren}, buf16
}

// Clear converts sRGB→linear before clearing, matching the C++ implicit srgba8→rgba8 conversion.
func (r *rgb555Renderer) Clear(c color.RGBA8[color.SRGB]) { r.renBase.Clear(color.ConvertToLinear(c)) }

// SetColor converts sRGB→linear before setting the color, matching C++ adaptor behavior.
func (r *rgb555Renderer) SetColor(c color.RGBA8[color.SRGB]) {
	r.ren.SetColor(color.ConvertToLinear(c))
}
func (r *rgb555Renderer) Prepare()                           { r.ren.Prepare() }
func (r *rgb555Renderer) Render(sl rendsl.ScanlineInterface) { r.ren.Render(sl) }

// --- Demo ---

type demo struct {
	x, y     [3]float64
	selected int
	dragDX   float64
	dragDY   float64
}

func newDemo() *demo {
	return &demo{
		x:        [3]float64{100, 369, 143},
		y:        [3]float64{60, 170, 310},
		selected: -1,
	}
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Render into a uint16 RGB555 scratch buffer (flip_y via negative stride).
	ren, buf16 := newRGB555Renderer(w, h)
	var sr SolidRenderer = ren

	// Build the triangle path.
	ps := path.NewPathStorageStl()
	ps.MoveTo(d.x[0], d.y[0])
	ps.LineTo(d.x[1], d.y[1])
	ps.LineTo(d.x[2], d.y[2])
	ps.ClosePolygon(basics.PathFlagsNone)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AddPath(&psAdapter{ps: ps}, 0)

	sl := scanline.NewScanlineP8()

	// Polymorphic dispatch: same code works with any SolidRenderer,
	// just as the C++ version works with any PixFmt.
	sr.Clear(color.RGBA8[color.SRGB]{R: 255, G: 255, B: 255, A: 255})
	sr.SetColor(color.RGBA8[color.SRGB]{R: 80, G: 30, B: 20, A: 255})
	rendsl.RenderScanlines(ras, sl, sr)

	// Convert the RGB555 uint16 buffer to RGBA8 for display.
	// buf16 has negative stride so raw index 0 corresponds to screen bottom-left,
	// matching img.Data when the runner uses FlipY:true with -w*4 stride.
	for i, pix := range buf16 {
		r, g, b := pixfmt.UnpackPixel555(pix)
		img.Data[i*4+0] = r
		img.Data[i*4+1] = g
		img.Data[i*4+2] = b
		img.Data[i*4+3] = 255
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	d.selected = -1
	for i := range 3 {
		dx := float64(x) - d.x[i]
		dy := float64(y) - d.y[i]
		if math.Sqrt(dx*dx+dy*dy) < 10 {
			d.selected = i
			d.dragDX = dx
			d.dragDY = dy
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if d.selected < 0 {
		return false
	}
	d.x[d.selected] = float64(x) - d.dragDX
	d.y[d.selected] = float64(y) - d.dragDY
	return true
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	d.selected = -1
	return false
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                  "Polymorphic Renderer",
		Width:                  400,
		Height:                 330,
		FlipY:                  true,
		DisableLinearRGBToSRGB: true,
	}, newDemo())
}

// --- Minimal adapters ---

type psAdapter struct{ ps *path.PathStorageStl }

func (a *psAdapter) Rewind(id uint32) { a.ps.Rewind(uint(id)) }
func (a *psAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x, *y = vx, vy
	return cmd
}
