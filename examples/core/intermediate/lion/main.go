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
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	frameWidth  = 512
	frameHeight = 400
)

type rasterizerType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

type demo struct {
	lion      liondemo.LionData
	baseDX    float64
	baseDY    float64
	angle     float64
	scale     float64
	skewX     float64
	skewY     float64
	width     int
	height    int
	alphaCtrl *sliderctrl.SliderCtrl

	controlDrag bool
}

type lionColorView struct {
	data  *liondemo.LionData
	alpha uint8
}

func (v lionColorView) GetColor(index int) color.RGBA8[color.Linear] {
	c := v.data.Colors[index]
	c.A = v.alpha
	return c
}

func newDemo() *demo {
	ld := liondemo.Parse()
	pathVS := path.NewPathStorageStlVertexSourceAdapter(ld.Path)
	bounds, ok := basics.BoundingRect[float64](pathVS, basics.SliceGetID(ld.PathIdx), 0, uint(ld.NPaths))
	if !ok {
		panic("lion: bounding rect not found")
	}

	alphaCtrl := sliderctrl.NewSliderCtrl(5, 5, frameWidth-5, 12, false)
	alphaCtrl.SetLabel("Alpha%3.3f")
	alphaCtrl.SetValue(0.1)

	return &demo{
		lion:      ld,
		baseDX:    (bounds.X2 - bounds.X1) * 0.5,
		baseDY:    (bounds.Y2 - bounds.Y1) * 0.5,
		scale:     1.0,
		width:     frameWidth,
		height:    frameHeight,
		alphaCtrl: alphaCtrl,
	}
}

func newRasterizer() *rasterizerType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

func (d *demo) Render(img *agg.Image) {
	d.width = img.Width()
	d.height = img.Height()

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineP8()

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-d.baseDX, -d.baseDY))
	mtx.Multiply(transform.NewTransAffineScaling(d.scale))
	mtx.Multiply(transform.NewTransAffineRotation(d.angle + math.Pi))
	mtx.Multiply(transform.NewTransAffineSkewing(d.skewX/1000.0, d.skewY/1000.0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(img.Width())/2, float64(img.Height())/2))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(d.lion.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rb)
	colors := lionColorView{
		data:  &d.lion,
		alpha: uint8(d.alphaCtrl.Value()*255.0 + 0.5),
	}

	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, colors, &d.lion, d.lion.NPaths)
	renderCtrl(ras, sl, rb, d.alphaCtrl)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	d.controlDrag = false

	if btn.Left && d.alphaCtrl.OnMouseButtonDown(fx, fy) {
		d.controlDrag = true
		return true
	}

	if btn.Left {
		d.transform(d.width, d.height, fx, fy)
		return true
	}

	if btn.Right {
		d.skewX = fx
		d.skewY = fy
		return true
	}

	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	if d.controlDrag {
		return d.alphaCtrl.OnMouseMove(fx, fy, btn.Left)
	}

	if btn.Left {
		d.transform(d.width, d.height, fx, fy)
		return true
	}

	if btn.Right {
		d.skewX = fx
		d.skewY = fy
		return true
	}

	return false
}

func (d *demo) OnMouseUp(x, y int, _ lowlevelrunner.Buttons) bool {
	redraw := false
	if d.controlDrag {
		redraw = d.alphaCtrl.OnMouseButtonUp(float64(x), float64(y))
		d.controlDrag = false
	}
	return redraw
}

func (d *demo) transform(width, height int, x, y float64) {
	x -= float64(width) / 2
	y -= float64(height) / 2
	d.angle = math.Atan2(y, x)
	d.scale = math.Sqrt(x*x+y*y) / 100.0
}

func renderCtrl(
	ras *rasterizerType,
	sl *scanline.ScanlineP8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ctrlVS := ctrlVertexSource{ctrl: ctrl}
		ras.AddPath(&ctrlVS, uint32(pathID))

		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{
			R: clampU8(c.R),
			G: clampU8(c.G),
			B: clampU8(c.B),
			A: clampU8(c.A),
		})
	}
}

type ctrlVertexSource struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (c *ctrlVertexSource) Rewind(pathID uint32) { c.ctrl.Rewind(uint(pathID)) }

func (c *ctrlVertexSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := c.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func clampU8(v float64) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 1:
		return 255
	default:
		return uint8(v*255.0 + 0.5)
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Lion",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
