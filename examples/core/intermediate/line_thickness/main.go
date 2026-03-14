package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/demo/linethickness"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

const (
	frameWidth  = 640
	frameHeight = 480
)

type control interface {
	InRect(x, y float64) bool
	OnMouseButtonDown(x, y float64) bool
	OnMouseButtonUp(x, y float64) bool
	OnMouseMove(x, y float64, buttonPressed bool) bool
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) color.RGBA
}

type controlPathAdapter struct {
	rewindFn func(pathID uint)
	vertexFn func() (x, y float64, cmd basics.PathCommand)
}

func (a *controlPathAdapter) Rewind(pathID uint32) { a.rewindFn(uint(pathID)) }
func (a *controlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vertexFn()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type rasScanlineAdapter struct {
	sl *scanline.ScanlineU8
}

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

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
	return color.RGBA8[color.Linear]{
		R: clamp(c.R),
		G: clamp(c.G),
		B: clamp(c.B),
		A: clamp(c.A),
	}
}

func renderControl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]],
	ctrl control,
) {
	adapter := &controlPathAdapter{
		rewindFn: ctrl.Rewind,
		vertexFn: ctrl.Vertex,
	}
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		col := rgbaToRGBA8(ctrl.Color(pathID))
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), col, spanData.Covers)
				}
			}
		}
	}
}

type demo struct {
	state     linethickness.State
	thickness *slider.SliderCtrl
	blur      *slider.SliderCtrl
	mono      *checkbox.CheckboxCtrl[color.RGBA]
	invert    *checkbox.CheckboxCtrl[color.RGBA]
	controls  []control
}

func newDemo() *demo {
	d := &demo{
		state: linethickness.DefaultState(),
	}

	d.thickness = slider.NewSliderCtrl(10, 480-19, 640-10, 480-10, false)
	d.thickness.SetRange(0.0, 5.0)
	d.thickness.SetValue(d.state.Thickness)
	d.thickness.SetLabel("Line thickness=%1.2f")

	d.blur = slider.NewSliderCtrl(10, 480-39, 640-10, 480-30, false)
	d.blur.SetRange(0.0, 2.0)
	d.blur.SetValue(d.state.Blur)
	d.blur.SetLabel("Blur radius=%1.2f")

	d.mono = checkbox.NewDefaultCheckboxCtrl(10, 480-64, "Monochrome", false)
	d.mono.SetChecked(d.state.Mono)

	d.invert = checkbox.NewDefaultCheckboxCtrl(10, 480-84, "Invert", false)
	d.invert.SetChecked(d.state.Invert)

	d.controls = []control{d.thickness, d.blur, d.mono, d.invert}
	return d
}

func (d *demo) syncState() {
	d.state.Thickness = d.thickness.Value()
	d.state.Blur = d.blur.Value()
	d.state.Mono = d.mono.IsChecked()
	d.state.Invert = d.invert.IsChecked()
	d.state.Clamp()
}

func (d *demo) Render(ctx *agg.Context) {
	d.syncState()
	linethickness.Draw(ctx, d.state)

	imgData := ctx.GetImage().Data
	rbuf := buffer.NewRenderingBufferU8WithData(imgData, frameWidth, frameHeight, frameWidth*4)
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]](pf)
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	for _, ctrl := range d.controls {
		renderControl(ras, sl, renBase, ctrl)
	}
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonDown(float64(x), float64(y)) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	redraw := false
	for _, ctrl := range d.controls {
		if ctrl.OnMouseMove(float64(x), float64(y), btn.Left) {
			redraw = true
		}
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool {
	_ = btn
	redraw := false
	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonUp(float64(x), float64(y)) {
			redraw = true
		}
	}
	return redraw
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Line Thickness",
		Width:  frameWidth,
		Height: frameHeight,
	}, newDemo())
}
