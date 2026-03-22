// Port of AGG C++ alpha_mask2.cpp – alpha-masked lion with affine-transformed mask.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	alphamask2demo "github.com/MeKo-Christian/agg_go/internal/demo/alphamask2"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

const (
	frameWidth  = 512
	frameHeight = 400
)

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
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

type demo struct {
	angle, scale float64
	skewX, skewY float64
	numCtrl      *sliderctrl.SliderCtrl
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*3)
	alphamask2demo.RenderToBGR24(workBuf, w, h, alphamask2demo.Config{
		NumEllipses: int(d.numCtrl.Value()),
		Angle:       d.angle,
		Scale:       d.scale,
		SkewX:       d.skewX,
		SkewY:       d.skewY,
	})

	// Render controls into the same pre-flip BGR24 work buffer as the scene so
	// they land at the same final screen position as in the C++ flip_y=true path.
	rbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*3)
	pixf := pixfmt.NewPixFmtBGR24(rbuf)
	pixfAdaptor := pixfmt.NewPixFmtRGBARendererAdaptor(pixf)
	rb := renderer.NewRendererBaseWithPixfmt(pixfAdaptor)
	ras := newRasterizer()
	sl := scanline.NewScanlineP8()
	renderCtrl(rb, ras, sl, d.numCtrl)

	// Convert BGR24 work buffer to RGBA output with y-flip.
	copyBGR24FlipY(workBuf, img.Data, w, h)
}

func copyBGR24FlipY(src, dst []uint8, width, height int) {
	srcStride := width * 3
	dstStride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * srcStride
		dstOff := y * dstStride
		for x := 0; x < width; x++ {
			s := srcOff + x*3
			d := dstOff + x*4
			dst[d+0] = src[s+2]
			dst[d+1] = src[s+1]
			dst[d+2] = src[s+0]
			dst[d+3] = 255
		}
	}
}

func renderCtrl(
	renBase renscan.BaseRendererInterface[color.RGBA8[color.Linear]],
	ras *rasType,
	sl *scanline.ScanlineP8,
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

func transformForPoint(width, height int, x, y float64) (angle, scale float64) {
	x -= float64(width) / 2
	y -= float64(height) / 2
	return math.Atan2(y, x), math.Sqrt(x*x+y*y) / 100.0
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	ctrlY := float64(frameHeight - y)
	if btn.Left && d.numCtrl.OnMouseButtonDown(float64(x), ctrlY) {
		return true
	}

	if btn.Left {
		d.angle, d.scale = transformForPoint(frameWidth, frameHeight, float64(x), float64(y))
		return true
	}

	if btn.Right {
		d.skewX = float64(x)
		d.skewY = float64(y)
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	ctrlY := float64(frameHeight - y)
	if d.numCtrl.OnMouseMove(float64(x), ctrlY, btn.Left) {
		return true
	}
	return d.OnMouseDown(x, y, btn)
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	return d.numCtrl.OnMouseButtonUp(float64(x), float64(frameHeight-y))
}

func main() {
	numCtrl := sliderctrl.NewSliderCtrl(5, 5, 150, 12, false)
	numCtrl.SetRange(5, 100)
	numCtrl.SetValue(10)
	numCtrl.SetLabel("N=%.2f")

	d := &demo{
		scale:   1.0,
		numCtrl: numCtrl,
	}
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Alpha Mask2",
		Width:  frameWidth,
		Height: frameHeight,
		FlipY:  true,
	}, d)
}
