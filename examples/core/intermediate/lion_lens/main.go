// Port of AGG C++ lion_lens.cpp - lion with warp-magnifier lens effect.
//
// Renders the lion vector art with the original AGG control widgets and the
// warp-magnifier lens. The lens starts at the C++ default position
// (200, 150), while the sliders control magnification and radius.
//
// This mirrors the C++ rendering pipeline faithfully:
//
//	conv_segmentator(g_path) -> conv_transform(mtx) -> conv_transform(lens)
//	render_all_paths(g_rasterizer, g_scanline, r, trans_lens, ...)
//
// The conv_segmentator is essential: it subdivides every line segment into
// ~1px pieces (in source space) BEFORE the non-linear warp magnifier runs, so
// straight edges curve smoothly through the lens instead of only having their
// endpoints displaced.
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	liondemo "github.com/cwbudde/agg_go/internal/demo/lion"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
)

const (
	llWidth  = 500
	llHeight = 600

	// C++ defaults.
	defaultLensScale  = 3.0
	defaultLensRadius = 70.0
	defaultLensX      = 200.0
	defaultLensY      = 150.0
)

type rasterizerType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

type rendererBaseType = renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]

type demo struct {
	lion         liondemo.LionData
	baseDX       float64
	baseDY       float64
	lensX        float64
	lensY        float64
	lightX       float64
	lightY       float64
	magnSlider   *sliderctrl.SliderCtrl
	radiusSlider *sliderctrl.SliderCtrl
}

func newDemo() *demo {
	ld := liondemo.Parse()

	pathVS := path.NewPathStorageStlVertexSourceAdapter(ld.Path)
	bounds, ok := basics.BoundingRect[float64](pathVS, basics.SliceGetID(ld.PathIdx), 0, uint(ld.NPaths))
	if !ok {
		panic("lion_lens: bounding rect not found")
	}

	magnSlider := sliderctrl.NewSliderCtrl(5, 5, 495, 12, false)
	magnSlider.SetRange(0.01, 4.0)
	magnSlider.SetValue(defaultLensScale)
	magnSlider.SetLabel("Scale=%3.2f")

	radiusSlider := sliderctrl.NewSliderCtrl(5, 20, 495, 27, false)
	radiusSlider.SetRange(0.0, 100.0)
	radiusSlider.SetValue(defaultLensRadius)
	radiusSlider.SetLabel("Radius=%3.2f")

	return &demo{
		lion:         ld,
		baseDX:       (bounds.X2 - bounds.X1) * 0.5,
		baseDY:       (bounds.Y2 - bounds.Y1) * 0.5,
		lensX:        defaultLensX,
		lensY:        defaultLensY,
		lightX:       defaultLensX,
		lightY:       defaultLensY,
		magnSlider:   magnSlider,
		radiusSlider: radiusSlider,
	}
}

func (d *demo) OnInit() {
	d.lensX = defaultLensX
	d.lensY = defaultLensY
	d.lightX = defaultLensX
	d.lightY = defaultLensY
}

func newRasterizer() *rasterizerType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

func (d *demo) Render(img *agg.Image) {
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineP8()

	// Warp magnifier lens (non-linear transform).
	lens := transform.NewTransWarpMagnifier()
	lens.SetCenter(d.lensX, d.lensY)
	lens.SetMagnification(d.magnSlider.Value())
	lens.SetRadius(d.radiusSlider.Value() / d.magnSlider.Value())

	// Affine: center the lion in the window, rotated by pi (matches C++).
	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-d.baseDX, -d.baseDY))
	mtx.Multiply(transform.NewTransAffineRotation(math.Pi))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(img.Width())/2, float64(img.Height())/2))

	// conv_segmentator(g_path) -> conv_transform(mtx) -> conv_transform(lens).
	pathVS := path.NewPathStorageStlVertexSourceAdapter(d.lion.Path)
	segm := conv.NewConvSegmentator(pathVS)
	transMtx := conv.NewConvTransform[conv.VertexSource, *transform.TransAffine](&segmAdapter{segm}, mtx)
	transLens := conv.NewConvTransform[conv.VertexSource, *transform.TransWarpMagnifier](transMtx, lens)

	rasVS := conv.NewRasterizerVertexSourceAdapter(transLens)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rb)

	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, &d.lion, &d.lion, d.lion.NPaths)

	renderCtrl(ras, sl, rb, d.magnSlider)
	renderCtrl(ras, sl, rb, d.radiusSlider)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	if btn.Left {
		if d.magnSlider.OnMouseButtonDown(fx, fy) {
			return true
		}
		if d.radiusSlider.OnMouseButtonDown(fx, fy) {
			return true
		}

		d.lensX = fx
		d.lensY = fy
		return true
	}

	if btn.Right {
		d.lightX = fx
		d.lightY = fy
		return true
	}

	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	redraw := d.magnSlider.OnMouseMove(fx, fy, btn.Left)

	if d.radiusSlider.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}

	if btn.Left && !d.magnSlider.InRect(fx, fy) && !d.radiusSlider.InRect(fx, fy) {
		d.lensX = fx
		d.lensY = fy
		redraw = true
	}

	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	redraw := d.magnSlider.OnMouseButtonUp(fx, fy)

	if d.radiusSlider.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	return redraw
}

// segmAdapter bridges conv.ConvSegmentator (uint32 cmd) to the conv.VertexSource
// contract (basics.PathCommand cmd) so it can feed conv.ConvTransform.
type segmAdapter struct{ s *conv.ConvSegmentator }

func (a *segmAdapter) Rewind(id uint) { a.s.Rewind(id) }

func (a *segmAdapter) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.s.Vertex()
	return x, y, basics.PathCommand(cmd)
}

func renderCtrl(
	ras *rasterizerType,
	sl *scanline.ScanlineP8,
	rb *rendererBaseType,
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
		Title:                 "Lion Lens",
		Width:                 llWidth,
		Height:                llHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
