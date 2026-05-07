// Port of AGG C++ rounded_rect.cpp – interactive rounded rectangle with controls.
//
// Renders a rounded rectangle defined by two draggable corner points, with
// adjustable radius and subpixel offset. Matches the C++ original's rendering
// pipeline: renderer_base + rasterizer_scanline_aa + scanline_p8 + conv_stroke.
// Default: corners at (100,100)-(500,350), radius=25, offset=0.
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	checkboxctrl "github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
)

const (
	demoWidth  = 600
	demoHeight = 400

	defaultRadius = 25.0
	defaultOffset = 0.5
)

type (
	colorType = color.RGBA8[color.Linear]
	pixFmt    = *pixfmt.PixFmtRGBA32Plain[color.Linear]
	renBaseT  = *renderer.RendererBase[pixFmt, colorType]
	rasType   = *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	slType    = *scanline.ScanlineP8
)

// rrVertexSource adapts shapes.RoundedRect to conv.VertexSource.
type rrVertexSource struct {
	rr *shapes.RoundedRect
}

func (s *rrVertexSource) Rewind(pathID uint) { s.rr.Rewind(uint32(pathID)) }

func (s *rrVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = s.rr.Vertex(&x, &y)
	return
}

// strokeVertexSource adapts conv.ConvStroke to the rasterizer VertexSource.
type strokeVertexSource struct {
	cs *conv.ConvStroke
}

func (s *strokeVertexSource) Rewind(pathID uint32) { s.cs.Rewind(uint(pathID)) }

func (s *strokeVertexSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := s.cs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// ellipseVertexSource adapts shapes.Ellipse to the rasterizer VertexSource.
type ellipseVertexSource struct {
	e *shapes.Ellipse
}

func (s *ellipseVertexSource) Rewind(pathID uint32) { s.e.Rewind(pathID) }

func (s *ellipseVertexSource) Vertex(x, y *float64) uint32 {
	cmd := s.e.Vertex(x, y)
	return uint32(cmd)
}

// Scanline/rasterizer adapters to bridge rasterizer ↔ renscan interfaces.
type demo struct {
	x      [2]float64
	y      [2]float64
	dx, dy float64
	idx    int

	radiusCtrl   *sliderctrl.SliderCtrl
	offsetCtrl   *sliderctrl.SliderCtrl
	whiteOnBlack *checkboxctrl.CheckboxCtrl[color.RGBA]
}

func newDemo() *demo {
	radius := sliderctrl.NewSliderCtrl(10, 10, demoWidth-10, 19, false)
	radius.SetLabel("radius=%4.3f")
	radius.SetRange(0.0, 50.0)
	radius.SetValue(defaultRadius)
	applyCPPSliderColors(radius)

	offset := sliderctrl.NewSliderCtrl(10, 10+20, demoWidth-10, 19+20, false)
	offset.SetLabel("subpixel offset=%4.3f")
	offset.SetRange(-2.0, 3.0)
	offset.SetValue(defaultOffset)
	applyCPPSliderColors(offset)

	ctrlGray := color.NewRGBAFromRGBA8(127, 127, 127, 255)
	whiteOnBlack := checkboxctrl.NewDefaultCheckboxCtrl(10, 10+40, "White on black", false)
	whiteOnBlack.SetTextColor(ctrlGray)
	whiteOnBlack.SetInactiveColor(ctrlGray)

	return &demo{
		x:            [2]float64{100, 500},
		y:            [2]float64{100, 350},
		idx:          -1,
		radiusCtrl:   radius,
		offsetCtrl:   offset,
		whiteOnBlack: whiteOnBlack,
	}
}

func (d *demo) Render(img *agg.Image) {
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Plain[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[pixFmt, colorType](pf)
	if d.whiteOnBlack.IsChecked() {
		rb.Clear(colorType{R: 0, G: 0, B: 0, A: 255})
	} else {
		rb.Clear(colorType{R: 255, G: 255, B: 255, A: 255})
	}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	// Render two "control" circles.
	gray := colorType{R: 127, G: 127, B: 127, A: 255}
	for i := 0; i < 2; i++ {
		e := shapes.NewEllipseWithParams(d.x[i], d.y[i], 3, 3, 16, false)
		ras.Reset()
		ras.AddPath(&ellipseVertexSource{e: e}, 0)
		renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](ras, sl, rb, gray)
	}

	// Create rounded rectangle.
	off := d.offsetCtrl.Value()
	rr := shapes.NewRoundedRect(d.x[0]+off, d.y[0]+off, d.x[1]+off, d.y[1]+off, d.radiusCtrl.Value())
	rr.NormalizeRadius()

	// Draw as outline.
	stroke := conv.NewConvStroke(&rrVertexSource{rr: rr})
	stroke.SetWidth(1.0)
	ras.Reset()
	ras.AddPath(&strokeVertexSource{cs: stroke}, 0)
	strokeColor := colorType{R: 0, G: 0, B: 0, A: 255}
	if d.whiteOnBlack.IsChecked() {
		strokeColor = colorType{R: 255, G: 255, B: 255, A: 255}
	}
	renscan.RenderScanlinesAASolid[colorType](ras, sl, rb, strokeColor)

	renderSlider(ras, sl, rb, d.radiusCtrl)
	renderSlider(ras, sl, rb, d.offsetCtrl)
	renderCheckbox(ras, sl, rb, d.whiteOnBlack)
}

func applyCPPSliderColors(s *sliderctrl.SliderCtrl) {
	s.SetBackgroundColor(linearRGBAForSRGBA8(color.NewRGBA(1.0, 0.9, 0.8, 1.0)))
	s.SetTriangleColor(linearRGBAForSRGBA8(color.NewRGBA(0.7, 0.6, 0.6, 1.0)))
	s.SetPointerPreviewColor(linearRGBAForSRGBA8(color.NewRGBA(0.6, 0.4, 0.4, 0.4)))
	s.SetPointerColor(linearRGBAForSRGBA8(color.NewRGBA(0.8, 0.0, 0.0, 0.6)))
}

func linearRGBAForSRGBA8(c color.RGBA) color.RGBA {
	return color.RGBA{
		R: color.ConvertToSRGB(c.R),
		G: color.ConvertToSRGB(c.G),
		B: color.ConvertToSRGB(c.B),
		A: c.A,
	}
}

func renderSlider(ras rasType, sl slType, rb renBaseT, s *sliderctrl.SliderCtrl) {
	renderControl(ras, sl, rb, s.NumPaths(), s.Rewind,
		func(x, y *float64) uint32 {
			vx, vy, cmd := s.Vertex()
			*x = vx
			*y = vy
			return uint32(cmd)
		},
		s.Color,
	)
}

func renderCheckbox(ras rasType, sl slType, rb renBaseT, cb *checkboxctrl.CheckboxCtrl[color.RGBA]) {
	renderControl(ras, sl, rb, cb.NumPaths(), cb.Rewind,
		func(x, y *float64) uint32 {
			vx, vy, cmd := cb.Vertex()
			*x = vx
			*y = vy
			return uint32(cmd)
		},
		cb.Color,
	)
}

func renderControl(
	ras rasType,
	sl slType,
	rb renBaseT,
	numPaths uint,
	rewindFn func(pathID uint),
	vertexFn func(x, y *float64) uint32,
	colorFn func(pathID uint) color.RGBA,
) {
	adapter := &ctrlPathAdapter{rewindFn: rewindFn, vertexFn: vertexFn}
	for pathID := uint(0); pathID < numPaths; pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		renscan.RenderScanlinesAASolid[colorType](ras, sl, rb, rgbaToRGBA8(colorFn(pathID)))
	}
}

type ctrlPathAdapter struct {
	rewindFn func(pathID uint)
	vertexFn func(x, y *float64) uint32
}

func (a *ctrlPathAdapter) Rewind(pathID uint32) { a.rewindFn(uint(pathID)) }

func (a *ctrlPathAdapter) Vertex(x, y *float64) uint32 { return a.vertexFn(x, y) }

func rgbaToRGBA8(c color.RGBA) colorType {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return colorType{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	if d.radiusCtrl.OnMouseButtonDown(fx, fy) {
		return true
	}
	if d.offsetCtrl.OnMouseButtonDown(fx, fy) {
		return true
	}
	if d.whiteOnBlack.OnMouseButtonDown(fx, fy) {
		return true
	}
	for i := 0; i < 2; i++ {
		if math.Sqrt((fx-d.x[i])*(fx-d.x[i])+(fy-d.y[i])*(fy-d.y[i])) < 5.0 {
			d.dx = fx - d.x[i]
			d.dy = fy - d.y[i]
			d.idx = i
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := d.radiusCtrl.OnMouseMove(fx, fy, btn.Left)
	if d.offsetCtrl.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}
	if btn.Left && d.idx >= 0 {
		d.x[d.idx] = float64(x) - d.dx
		d.y[d.idx] = float64(y) - d.dy
		return true
	}
	if !btn.Left {
		d.idx = -1
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, _ lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := d.radiusCtrl.OnMouseButtonUp(fx, fy)
	if d.offsetCtrl.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	if d.whiteOnBlack.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	d.idx = -1
	return redraw
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Rounded Rectangle",
		Width:  demoWidth,
		Height: demoHeight,
		FlipY:  true,
	}, newDemo())
}
