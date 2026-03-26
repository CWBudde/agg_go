// Port of AGG C++ line_thickness.cpp (flip_y = true).
//
// Renders a row of straight lines of increasing thickness and a wheel of
// fine lines, then applies a slight blur. Interactive sliders control the line
// thickness and blur radius; checkboxes select monochrome vs. colour and
// invert the foreground/background.
//
// Rendering is done in a work buffer (y=0 at bottom, C++ coordinate frame)
// and copied with a y-flip into the output image, matching the
// flip_y=true platform_support convention. Mouse y-coordinates are flipped
// in the handlers before being forwarded to the controls.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	checkboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/demo/linethickness"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

const (
	frameWidth  = linethickness.Width
	frameHeight = linethickness.Height
)

// rasType is the concrete rasterizer type used throughout this demo.
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ctrlVS adapts a ctrl.Ctrl to the rasterizer.VertexSource interface.
type ctrlVS struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (v *ctrlVS) Rewind(id uint32) { v.ctrl.Rewind(uint(id)) }
func (v *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// rgbaToRGBA8 converts the float-based color.RGBA to a clamped RGBA8.
func rgbaToRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA8[color.Linear]{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

// renderCtrl renders all paths of a control widget into the work buffer.
func renderCtrl(
	ras *rasType,
	sl *scanline.ScanlineU8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	vs := &ctrlVS{ctrl: ctrl}
	for i := range ctrl.NumPaths() {
		ras.Reset()
		ras.AddPath(vs, uint32(i))
		renscan.RenderScanlinesAASolid(ras, sl, rb, rgbaToRGBA8(ctrl.Color(i)))
	}
}

// copyFlipY copies src to dst with a vertical flip.
func copyFlipY(src, dst []uint8, w, h int) {
	stride := w * 4
	for y := range h {
		srcOff := (h - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

type demo struct {
	state     linethickness.State
	thickness *sliderctrl.SliderCtrl
	blur      *sliderctrl.SliderCtrl
	mono      *checkboxctrl.CheckboxCtrl[color.RGBA]
	invert    *checkboxctrl.CheckboxCtrl[color.RGBA]
	ctrls     []ctrlbase.Ctrl[color.RGBA]
}

func newDemo() *demo {
	d := &demo{
		state: linethickness.DefaultState(),
	}

	// C++: m_slider1(10, 10,    640-10, 19,    !flip_y=false)  line thickness
	//      m_slider2(10, 10+20, 640-10, 19+20, !flip_y=false)  blur radius
	//      m_cbox1  (10, 10+40, "Monochrome",  !flip_y=false)
	//      m_cbox2  (10, 10+60, "Invert",      !flip_y=false)
	d.thickness = sliderctrl.NewSliderCtrl(10, 10, 630, 19, false)
	d.thickness.SetRange(0.0, 5.0)
	d.thickness.SetValue(d.state.Thickness)
	d.thickness.SetLabel("Line thickness=%1.2f")

	d.blur = sliderctrl.NewSliderCtrl(10, 30, 630, 39, false)
	d.blur.SetRange(0.0, 2.0)
	d.blur.SetValue(d.state.Blur)
	d.blur.SetLabel("Blur radius=%1.2f")

	d.mono = checkboxctrl.NewDefaultCheckboxCtrl(10, 50, "Monochrome", false)
	d.mono.SetChecked(d.state.Mono)

	d.invert = checkboxctrl.NewDefaultCheckboxCtrl(10, 70, "Invert", false)
	d.invert.SetChecked(d.state.Invert)

	d.ctrls = []ctrlbase.Ctrl[color.RGBA]{d.thickness, d.blur, d.mono, d.invert}
	return d
}

func (d *demo) syncState() {
	d.state.Thickness = d.thickness.Value()
	d.state.Blur = d.blur.Value()
	d.state.Mono = d.mono.IsChecked()
	d.state.Invert = d.invert.IsChecked()
	d.state.Clamp()
}

func (d *demo) Render(img *agg.Image) {
	w, h := frameWidth, frameHeight
	d.syncState()

	// Work buffer: positive stride, y=0 at bottom (C++ y-up frame).
	workBuf := make([]uint8, w*h*4)

	// Render scene content (straight lines + wheel + blur).
	linethickness.Draw(workBuf, w, h, d.state)

	// Render controls on top of the blurred scene.
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(workBuf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	ras := newRasterizer()
	sl := scanline.NewScanlineU8()

	for _, ctrl := range d.ctrls {
		renderCtrl(ras, sl, rb, ctrl)
	}

	// Copy work buffer to output image with y-flip (flip_y=true convention).
	copyFlipY(workBuf, img.Data, w, h)
}

// flipY converts screen y (y=0 at top) to work-buffer y (y=0 at bottom).
func flipY(y int) int { return frameHeight - 1 - y }

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	wy := flipY(y)
	for _, ctrl := range d.ctrls {
		if ctrl.OnMouseButtonDown(float64(x), float64(wy)) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	wy := flipY(y)
	redraw := false
	for _, ctrl := range d.ctrls {
		if ctrl.OnMouseMove(float64(x), float64(wy), btn.Left) {
			redraw = true
		}
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	_ = btn
	wy := flipY(y)
	redraw := false
	for _, ctrl := range d.ctrls {
		if ctrl.OnMouseButtonUp(float64(x), float64(wy)) {
			redraw = true
		}
	}
	return redraw
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Anti-aliased lines with blurring",
		Width:                 frameWidth,
		Height:                frameHeight,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
