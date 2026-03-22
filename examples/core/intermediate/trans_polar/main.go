package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

type transPolar struct {
	baseAngle      float64
	baseScale      float64
	baseX, baseY   float64
	transX, transY float64
	spiral         float64
}

func (p *transPolar) Transform(x, y *float64) {
	x1 := (*x + p.baseX) * p.baseAngle
	y1 := (*y+p.baseY)*p.baseScale + (*x * p.spiral)
	*x = math.Cos(x1)*y1 + p.transX
	*y = math.Sin(x1)*y1 + p.transY
}

type segmAdapter struct{ s *conv.ConvSegmentator }

func (a *segmAdapter) Rewind(id uint) { a.s.Rewind(id) }
func (a *segmAdapter) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.s.Vertex()
	return x, y, basics.PathCommand(cmd)
}

type TransformedControl struct {
	ctrl     ctrlbase.Ctrl[color.RGBA]
	pipeline conv.VertexSource
}

func (tc *TransformedControl) NumPaths() uint {
	return tc.ctrl.NumPaths()
}

func (tc *TransformedControl) Rewind(pathID uint) {
	tc.pipeline.Rewind(pathID)
}

func (tc *TransformedControl) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := tc.pipeline.Vertex()
	return x, y, cmd
}

func (tc *TransformedControl) Color(i uint) color.RGBA {
	return tc.ctrl.Color(i)
}

type demo struct {
	slider1      *sliderctrl.SliderCtrl
	sliderSpiral *sliderctrl.SliderCtrl
	sliderBaseY  *sliderctrl.SliderCtrl
}

func newDemo() *demo {
	slider1 := sliderctrl.NewSliderCtrl(10, 10, 600-10, 17, false)
	slider1.SetRange(0.0, 100.0)
	slider1.SetNumSteps(5)
	slider1.SetValue(32.0)
	slider1.SetLabel("Some Value=%1.0f")

	sliderSpiral := sliderctrl.NewSliderCtrl(10, 10+20, 600-10, 17+20, false)
	sliderSpiral.SetLabel("Spiral=%.3f")
	sliderSpiral.SetRange(-0.1, 0.1)
	sliderSpiral.SetValue(0.0)

	sliderBaseY := sliderctrl.NewSliderCtrl(10, 10+40, 600-10, 17+40, false)
	sliderBaseY.SetLabel("Base Y=%.3f")
	sliderBaseY.SetRange(50.0, 200.0)
	sliderBaseY.SetValue(120.0)

	return &demo{
		slider1:      slider1,
		sliderSpiral: sliderSpiral,
		sliderBaseY:  sliderBaseY,
	}
}

type ctrlVS struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlVS) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type ctrlVSRender struct {
	ctrl interface {
		NumPaths() uint
		Rewind(uint)
		Vertex() (float64, float64, basics.PathCommand)
		Color(uint) color.RGBA
	}
}

func (a *ctrlVSRender) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlVSRender) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
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

func renderCtrlStandard(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
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

func renderCtrlTransformed(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl *TransformedControl,
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVSRender{ctrl: ctrl}, uint32(pathID))
		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, color.RGBA8[color.Linear]{
			R: clampU8(c.R),
			G: clampU8(c.G),
			B: clampU8(c.B),
			A: clampU8(c.A),
		})
	}
}

func (d *demo) Render(img *agg.Image) {
	w := img.Width()
	h := img.Height()

	mainBuf := buffer.NewRenderingBufferU8WithData(img.Data, w, h, img.Stride())
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainBuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](mainPixf)

	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	renderCtrlStandard(ras, sl, rb, d.slider1)
	renderCtrlStandard(ras, sl, rb, d.sliderSpiral)
	renderCtrlStandard(ras, sl, rb, d.sliderBaseY)

	trans := &transPolar{
		baseAngle: 2.0 * math.Pi / -600.0,
		baseScale: -1.0,
		baseX:     0.0,
		baseY:     d.sliderBaseY.Value(),
		transX:    float64(w) / 2.0,
		transY:    float64(h)/2.0 + 30.0,
		spiral:    -d.sliderSpiral.Value(),
	}

	segm := conv.NewConvSegmentator(d.slider1)
	pipeline := conv.NewConvTransform[conv.VertexSource, *transPolar](&segmAdapter{segm}, trans)

	ctrl := &TransformedControl{
		ctrl:     d.slider1,
		pipeline: pipeline,
	}

	renderCtrlTransformed(ras, sl, rb, ctrl)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if btn.Left {
		if d.slider1.OnMouseButtonDown(fx, fy) ||
			d.sliderSpiral.OnMouseButtonDown(fx, fy) ||
			d.sliderBaseY.OnMouseButtonDown(fx, fy) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if btn.Left {
		if d.slider1.OnMouseMove(fx, fy, true) ||
			d.sliderSpiral.OnMouseMove(fx, fy, true) ||
			d.sliderBaseY.OnMouseMove(fx, fy, true) {
			return true
		}
	} else {
		if d.slider1.OnMouseMove(fx, fy, false) ||
			d.sliderSpiral.OnMouseMove(fx, fy, false) ||
			d.sliderBaseY.OnMouseMove(fx, fy, false) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.slider1.OnMouseButtonUp(fx, fy) ||
		d.sliderSpiral.OnMouseButtonUp(fx, fy) ||
		d.sliderBaseY.OnMouseButtonUp(fx, fy) {
		return true
	}
	return false
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AGG Example. Polar Transformer",
		Width:  600,
		Height: 400,
		FlipY:  true,
	}, newDemo())
}
