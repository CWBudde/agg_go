// Port of AGG's blend_color.cpp demo.
//
// This example keeps the original interactive controls:
// - method rbox
// - blur radius slider
// - draggable shadow quadrilateral
//
// The C++ demo does not draw any timing text; this port follows that layout.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlpkg "github.com/MeKo-Christian/agg_go/internal/ctrl"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	blenddata "github.com/MeKo-Christian/agg_go/internal/demo/blendcolor"
	"github.com/MeKo-Christian/agg_go/internal/effects"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	frameWidth  = 440
	frameHeight = 330
)

type pathStorageAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathStorageAdapter) Rewind(pathID uint) {
	a.ps.Rewind(pathID)
}

func (a *pathStorageAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, rawCmd := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(rawCmd)
}

type rasPathAdapter struct {
	vs conv.VertexSource
}

func (a *rasPathAdapter) Rewind(pathID uint32) {
	a.vs.Rewind(uint(pathID))
}

func (a *rasPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type controlPathAdapter struct {
	ctrl ctrlpkg.Ctrl[color.RGBA]
}

func (a *controlPathAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *controlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type demo struct {
	glyphPath *path.PathStorageStl
	shapeBox  basics.Rect[float64]

	methodCtrl *rboxctrl.RboxCtrl[color.RGBA]
	radiusCtrl *slider.SliderCtrl
	shadowCtrl *polygon.PolygonCtrl[color.RGBA]

	colorLUT []color.RGBA8[color.SRGB]
}

func newDemo() *demo {
	glyphPath := blenddata.BuildGlyphPath()
	shape := conv.NewConvCurve(&pathStorageAdapter{ps: glyphPath})
	box, ok := basics.BoundingRectSingle[float64](shape, 0)
	if !ok {
		box = basics.Rect[float64]{X1: 0, Y1: 0, X2: 0, Y2: 0}
	}

	methodCtrl := rboxctrl.NewDefaultRboxCtrl(10, 10, 130, 55, false)
	methodCtrl.SetTextSize(8.0, 0)
	methodCtrl.AddItem("Single Color")
	methodCtrl.AddItem("Color LUT")
	methodCtrl.SetCurItem(1)

	radiusCtrl := slider.NewSliderCtrl(140, 14, 430, 22, false)
	radiusCtrl.SetRange(0.0, 40.0)
	radiusCtrl.SetValue(15.0)
	radiusCtrl.SetLabel("Blur Radius=%1.2f")

	shadowCtrl := polygon.NewDefaultPolygonCtrl(4, 5.0)
	shadowCtrl.SetClose(true)
	shadowCtrl.SetLineColor(color.NewRGBA(0.0, 0.3, 0.5, 0.3))
	shadowCtrl.SetXn(0, box.X1)
	shadowCtrl.SetYn(0, box.Y1)
	shadowCtrl.SetXn(1, box.X2)
	shadowCtrl.SetYn(1, box.Y1)
	shadowCtrl.SetXn(2, box.X2)
	shadowCtrl.SetYn(2, box.Y2)
	shadowCtrl.SetXn(3, box.X1)
	shadowCtrl.SetYn(3, box.Y2)

	return &demo{
		glyphPath:  glyphPath,
		shapeBox:   box,
		methodCtrl: methodCtrl,
		radiusCtrl: radiusCtrl,
		shadowCtrl: shadowCtrl,
		colorLUT:   blenddata.BuildColorLUT(),
	}
}

func (d *demo) currentQuad() [8]float64 {
	return [8]float64{
		d.shadowCtrl.Xn(0), d.shadowCtrl.Yn(0),
		d.shadowCtrl.Xn(1), d.shadowCtrl.Yn(1),
		d.shadowCtrl.Xn(2), d.shadowCtrl.Yn(2),
		d.shadowCtrl.Xn(3), d.shadowCtrl.Yn(3),
	}
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, w, h, img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Pre[color.SRGB](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtAlphaBlendRGBA[color.SRGB, blender.BlenderRGBA8Pre[color.SRGB, order.RGBA]], color.RGBA8[color.SRGB]](pf)
	renBase.Clear(color.RGBA8[color.SRGB]{R: 255, G: 249, B: 249, A: 255})

	grayBuf := make([]byte, w*h)
	grayRbuf := buffer.NewRenderingBufferU8WithData(grayBuf, w, h, w)
	grayPix := pixfmt.NewPixFmtSGray8(grayRbuf)
	grayRen := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtSGray8, color.Gray8[color.SRGB]](grayPix)
	grayRen.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	shape := conv.NewConvCurve(&pathStorageAdapter{ps: d.glyphPath})
	shadowPersp := transform.NewTransPerspectiveRectToQuad(
		d.shapeBox.X1, d.shapeBox.Y1, d.shapeBox.X2, d.shapeBox.Y2, d.currentQuad(),
	)
	shadowTrans := conv.NewConvTransform(shape, shadowPersp)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.ClipBox(0, 0, float64(w), float64(h))
	sl := scanline.NewScanlineP8()

	ras.Reset()
	ras.AddPath(&rasPathAdapter{vs: shadowTrans}, 0)
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		renscan.RenderScanlinesAASolid(ras, sl, grayRen, color.Gray8[color.SRGB]{V: 255, A: 255})
	}

	bbox, ok := basics.BoundingRectSingle[float64](shadowTrans, 0)
	if !ok {
		bbox = basics.Rect[float64]{X1: 0, Y1: 0, X2: float64(w), Y2: float64(h)}
	}
	bbox.X1 -= d.radiusCtrl.Value()
	bbox.Y1 -= d.radiusCtrl.Value()
	bbox.X2 += d.radiusCtrl.Value()
	bbox.Y2 += d.radiusCtrl.Value()
	if bbox.Clip(basics.Rect[float64]{X1: 0, Y1: 0, X2: float64(w), Y2: float64(h)}) {
		x1 := int(bbox.X1)
		y1 := int(bbox.Y1)
		x2 := int(bbox.X2)
		y2 := int(bbox.Y2)
		bw := x2 - x1
		bh := y2 - y1
		if bw > 0 && bh > 0 {
			shadowBuf := make([]byte, bw*bh)
			shadowRbuf := buffer.NewRenderingBufferU8WithData(shadowBuf, bw, bh, bw)
			shadowPix := pixfmt.NewPixFmtSGray8(shadowRbuf)

			for y := 0; y < bh; y++ {
				srcRow := grayPix.RowData(y1 + y)
				if srcRow == nil {
					continue
				}
				copy(shadowPix.RowData(y), srcRow[x1:x1+bw])
			}

			blurR := int(math.Round(d.radiusCtrl.Value()))
			if blurR > 0 {
				effects.StackBlurGray8(shadowPix, blurR, blurR)
			}

			if d.methodCtrl.CurItem() == 0 {
				renBase.BlendFromColor(shadowPix, color.RGBA8[color.SRGB]{R: 0, G: 100, B: 0, A: 255}, nil, x1, y1, 255)
			} else {
				renBase.BlendFromLUT(shadowPix, d.colorLUT, nil, x1, y1, 255)
			}
		}
	}

	rasCtrl := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	slCtrl := scanline.NewScanlineP8()
	d.renderControl(rasCtrl, slCtrl, renBase, d.methodCtrl)
	d.renderControl(rasCtrl, slCtrl, renBase, d.radiusCtrl)
	d.renderControl(rasCtrl, slCtrl, renBase, d.shadowCtrl)
}

func (d *demo) renderControl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineP8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.SRGB, blender.BlenderRGBA8Pre[color.SRGB, order.RGBA]], color.RGBA8[color.SRGB]],
	ctrl ctrlpkg.Ctrl[color.RGBA],
) {
	adapter := &controlPathAdapter{ctrl: ctrl}
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		renscan.RenderScanlinesAASolid(ras, sl, renBase, rgbaToRGBA8(ctrl.Color(pathID)))
	}
}

func rgbaToRGBA8(c color.RGBA) color.RGBA8[color.SRGB] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA8[color.SRGB]{
		R: clamp(c.R),
		G: clamp(c.G),
		B: clamp(c.B),
		A: clamp(c.A),
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	handled := d.methodCtrl.OnMouseButtonDown(fx, fy)
	handled = d.radiusCtrl.OnMouseButtonDown(fx, fy) || handled
	handled = d.shadowCtrl.OnMouseButtonDown(fx, fy) || handled
	return handled
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	pressed := btn.Left || btn.Right || btn.Middle
	handled := d.methodCtrl.OnMouseMove(fx, fy, pressed)
	handled = d.radiusCtrl.OnMouseMove(fx, fy, pressed) || handled
	handled = d.shadowCtrl.OnMouseMove(fx, fy, pressed) || handled
	return handled
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	handled := d.methodCtrl.OnMouseButtonUp(fx, fy)
	handled = d.radiusCtrl.OnMouseButtonUp(fx, fy) || handled
	handled = d.shadowCtrl.OnMouseButtonUp(fx, fy) || handled
	return handled
}

func (d *demo) OnKey(key rune) bool {
	switch key {
	case 'm', 'M':
		cur := d.methodCtrl.CurItem()
		if cur <= 0 {
			d.methodCtrl.SetCurItem(1)
		} else {
			d.methodCtrl.SetCurItem(0)
		}
		return true
	case '+', '=':
		v := d.radiusCtrl.Value() + 0.5
		if v > 40.0 {
			v = 40.0
		}
		d.radiusCtrl.SetValue(v)
		return true
	case '-', '_':
		v := d.radiusCtrl.Value() - 0.5
		if v < 0.0 {
			v = 0.0
		}
		d.radiusCtrl.SetValue(v)
		return true
	}
	return false
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Blend Color",
		Width:  frameWidth,
		Height: frameHeight,
		FlipY:  true,
	}, newDemo())
}
