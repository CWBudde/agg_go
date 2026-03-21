// Port of AGG C++ component_rendering.cpp – component (channel) rendering.
//
// Three large circles are each rendered into an individual color channel
// (R, G, B) using grayscale rendering. The gray value 0 with alpha darkens
// just one channel, producing a subtractive CMY mixing effect.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	frameWidth  = 320
	frameHeight = 320
)

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters (same pattern as other lowlevel demos)
// ---------------------------------------------------------------------------
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ellipseVS adapts shapes.Ellipse to rasterizer.VertexSource.
type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	var vx, vy float64
	cmd := ev.e.Vertex(&vx, &vy)
	*x, *y = vx, vy
	return uint32(cmd)
}

// ctrlVS adapts a Ctrl to rasterizer.VertexSource.
type ctrlVS struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlVS) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Per-channel rendering via gray buffer overlay
// ---------------------------------------------------------------------------

// renderEllipseToChannel renders an anti-aliased ellipse and applies it as
// a per-channel darkening to the main RGBA buffer. This emulates the C++
// pixfmt_alpha_blend_gray<Step=3, Offset=N> technique.
//
// grayVal=0 with alpha means: dst_channel = dst_channel * (255 - alpha*cover/255) / 255
func renderEllipseToChannel(
	mainBuf []uint8, w, h int,
	cx, cy, rx, ry float64,
	grayVal uint8, alpha uint8,
	channelOffset int, // 0=R, 1=G, 2=B in RGBA layout
) {
	// Render the ellipse into a temporary gray8 buffer to get coverage.
	grayBuf := make([]uint8, w*h)
	grayRbuf := buffer.NewRenderingBufferU8WithData(grayBuf, w, h, w)
	grayPixf := pixfmt.NewPixFmtGray8(grayRbuf)
	grayRb := renderer.NewRendererBaseWithPixfmt(grayPixf)
	// Clear to white (255) — no effect.
	grayRb.Clear(color.Gray8[color.Linear]{V: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineP8()

	ell := shapes.NewEllipseWithParams(cx, cy, rx, ry, 100, false)
	ras.AddPath(&ellipseVS{e: ell}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, grayRb, color.Gray8[color.Linear]{V: grayVal, A: alpha})

	// Now apply: for each pixel, the gray buffer contains the blended result.
	// Gray started at 255, and rendering gray_type(0, alpha) blends toward 0.
	// The resulting gray value g means: channel = channel * g / 255.
	stride := w * 4
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			g := grayBuf[y*w+x]
			if g == 255 {
				continue // No effect.
			}
			idx := y*stride + x*4 + channelOffset
			mainBuf[idx] = uint8(uint16(mainBuf[idx]) * uint16(g) / 255)
		}
	}
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct {
	alphaSlider *sliderctrl.SliderCtrl
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work in RGBA32 with y=0 at bottom (C++ flip_y=true).
	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	alpha := uint8(d.alphaSlider.Value())
	fw := float64(w)
	fh := float64(h)

	// C++: ellipse at (w/2 - 0.87*50, h/2 - 0.5*50, 100, 100, 100)
	// Rendered into R channel (offset=0 in RGBA).
	renderEllipseToChannel(workBuf, w, h,
		fw/2-0.87*50, fh/2-0.5*50, 100, 100,
		0, alpha, 0)

	// Rendered into G channel (offset=1 in RGBA).
	renderEllipseToChannel(workBuf, w, h,
		fw/2+0.87*50, fh/2-0.5*50, 100, 100,
		0, alpha, 1)

	// Rendered into B channel (offset=2 in RGBA).
	renderEllipseToChannel(workBuf, w, h,
		fw/2, fh/2+50, 100, 100,
		0, alpha, 2)

	// Render the slider control on top.
	ras := newRasterizer()
	sl := scanline.NewScanlineP8()
	renderCtrl(ras, sl, mainRb, d.alphaSlider)

	// Copy with y-flip (C++ uses flip_y=true).
	copyFlipY(workBuf, img.Data, w, h)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(frameHeight-y)
	if btn.Left {
		return d.alphaSlider.OnMouseButtonDown(fx, fy)
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(frameHeight-y)
	return d.alphaSlider.OnMouseMove(fx, fy, btn.Left)
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(frameHeight-y)
	return d.alphaSlider.OnMouseButtonUp(fx, fy)
}

func renderCtrl(
	ras *rasType,
	sl *scanline.ScanlineP8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVS{ctrl: ctrl}, uint32(pathID))
		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, color.RGBA8[color.Linear]{
			R: clampU8(c.R),
			G: clampU8(c.G),
			B: clampU8(c.B),
			A: clampU8(c.A),
		})
	}
}

func clampU8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255.0 + 0.5)
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
	// C++: m_alpha(5, 5, 320-5, 10+5, !flip_y)  → flipY=false
	alphaSlider := sliderctrl.NewSliderCtrl(5, 5, frameWidth-5, 10+5, false)
	alphaSlider.SetLabel("Alpha=%.0f")
	alphaSlider.SetRange(0, 255)
	alphaSlider.SetValue(255)

	d := &demo{
		alphaSlider: alphaSlider,
	}
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Component Rendering",
		Width:  frameWidth,
		Height: frameHeight,
	}, d)
}
