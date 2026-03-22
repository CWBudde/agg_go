// Port of AGG C++ gamma_tuner.cpp.
//
// This version keeps the original control layout and draw order:
// - R/G/B/Gamma sliders on the left
// - pattern radio box on the right
// - vertical gradient background
// - alpha-blended pattern square
// - vertical strips
//
// The example runs through the low-level runner so the platform layer can
// handle y-flip the same way AGG's platform_support does.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	pixgamma "github.com/MeKo-Christian/agg_go/internal/pixfmt/gamma"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	isl "github.com/MeKo-Christian/agg_go/internal/scanline"
)

const (
	canvasW = 500
	canvasH = 500

	squareSize = 400
	verStrips  = 5
)

type demo struct {
	rSlider     *sliderctrl.SliderCtrl
	gSlider     *sliderctrl.SliderCtrl
	bSlider     *sliderctrl.SliderCtrl
	gammaSlider *sliderctrl.SliderCtrl
	pattern     *rboxctrl.RboxCtrl[icol.RGBA]
}

type (
	rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	renBase = renderer.RendererBase[*pixgamma.PixFmtRGBA32GammaBlend, icol.RGBA8[icol.Linear]]
)

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

type rasterVertexSourceAdapter struct {
	src ctrlbase.Ctrl[icol.RGBA]
}

func (a *rasterVertexSourceAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *rasterVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func newDemo() *demo {
	d := &demo{
		rSlider:     sliderctrl.NewSliderCtrl(5, 5, 350-5, 11, false),
		gSlider:     sliderctrl.NewSliderCtrl(5, 20, 350-5, 26, false),
		bSlider:     sliderctrl.NewSliderCtrl(5, 35, 350-5, 41, false),
		gammaSlider: sliderctrl.NewSliderCtrl(5, 50, 350-5, 56, false),
		pattern:     rboxctrl.NewDefaultRboxCtrl(355, 1, 495, 60, false),
	}

	d.rSlider.SetValue(1.0)
	d.rSlider.SetLabel("R=%.2f")

	d.gSlider.SetValue(1.0)
	d.gSlider.SetLabel("G=%.2f")

	d.bSlider.SetValue(1.0)
	d.bSlider.SetLabel("B=%.2f")

	d.gammaSlider.SetRange(0.5, 4.0)
	d.gammaSlider.SetValue(2.2)
	d.gammaSlider.SetLabel("Gamma=%.2f")

	d.pattern.SetTextSize(8.0, 0.0)
	d.pattern.AddItem("Horizontal")
	d.pattern.AddItem("Vertical")
	d.pattern.AddItem("Checkered")
	d.pattern.SetCurItem(2)

	return d
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	rVal := d.rSlider.Value()
	gVal := d.gSlider.Value()
	bVal := d.bSlider.Value()
	gammaVal := d.gammaSlider.Value()

	// Use gamma-correct blending to match C++ pixfmt_sbgr24_gamma / blender_rgb_gamma.
	// Copy operations (background, strips) remain raw sRGB — gamma only applies to blends.
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())
	pixf := pixgamma.NewPixFmtRGBA32GammaBlend(rbuf, pixgamma.NewSimpleGammaLut(gammaVal))
	rb := renderer.NewRendererBaseWithPixfmt[*pixgamma.PixFmtRGBA32GammaBlend, icol.RGBA8[icol.Linear]](pixf)

	// rawColor computes raw sRGB matching C++ color.gradient(black, k):
	//   k=0 → full color, k=1 → black (brightness = 1-k)
	// C++ copy_hline writes raw sRGB directly (no gamma applied).
	rawColor := func(k float64) icol.RGBA8[icol.Linear] {
		a := 1 - k
		clamp := func(v float64) uint8 {
			if v <= 0 {
				return 0
			}
			if v >= 1 {
				return 255
			}
			return uint8(v*255.0 + 0.5)
		}
		return icol.RGBA8[icol.Linear]{
			R: clamp(rVal * a),
			G: clamp(gVal * a),
			B: clamp(bVal * a),
			A: 255,
		}
	}

	// Vertical gradient background (C++: copy_hline — raw sRGB, no gamma).
	for y := 0; y < h; y++ {
		k := float64(y-80) / float64(squareSize-1)
		if y < 80 {
			k = 0
		}
		if y >= 80+squareSize {
			k = 1
		}
		k = 1 - math.Pow(k/2, 1/gammaVal)
		rb.CopyHline(0, y, w-1, rawColor(k))
	}

	span1 := make([]icol.RGBA8[icol.Linear], squareSize)
	span2 := make([]icol.RGBA8[icol.Linear], squareSize)

	// Pre-compute span alpha values, matching C++ span setup before the draw loop.
	for j := 0; j < squareSize; j++ {
		alpha := uint8(j * 255 / squareSize)
		inv := 255 - alpha
		switch d.pattern.CurItem() {
		case 0: // Horizontal
			span1[j].A = alpha
			span2[j].A = inv
		case 1: // Vertical
			if j&1 != 0 {
				span1[j].A = alpha
				span2[j].A = alpha
			} else {
				span1[j].A = inv
				span2[j].A = inv
			}
		default: // Checkered
			if j&1 != 0 {
				span1[j].A = alpha
				span2[j].A = inv
			} else {
				span2[j].A = alpha
				span1[j].A = inv
			}
		}
	}

	// Clear the square.
	rb.CopyBar(50, 80, 50+squareSize-1, 80+squareSize-1, icol.RGBA8[icol.Linear]{R: 0, G: 0, B: 0, A: 255})

	// Draw the pattern — only update RGB per row; alpha stays fixed from pre-computation.
	for i := 0; i < squareSize; i += 2 {
		k := float64(i) / float64(squareSize-1)
		k = 1 - math.Pow(k, 1/gammaVal)
		c := rawColor(k)
		for j := 0; j < squareSize; j++ {
			span1[j].R, span1[j].G, span1[j].B = c.R, c.G, c.B
			span2[j].R, span2[j].G, span2[j].B = c.R, c.G, c.B
		}
		rb.BlendColorHspan(50, i+80, squareSize, span1, nil, basics.CoverFull)
		rb.BlendColorHspan(50, i+80+1, squareSize, span2, nil, basics.CoverFull)
	}

	// Draw vertical strips (C++: copy_hline — raw sRGB, no gamma).
	for i := 0; i < squareSize; i++ {
		k := float64(i) / float64(squareSize-1)
		k = 1 - math.Pow(k/2, 1/gammaVal)
		c := rawColor(k)
		y := i + 80
		for j := 0; j < verStrips; j++ {
			xc := squareSize * (j + 1) / (verStrips + 1)
			rb.CopyHline(50+xc-10, y, 50+xc+10, c)
		}
	}

	// Controls on top — use the same gamma-correct renderer so anti-aliasing matches C++.
	ras := newRasterizer()
	sl := isl.NewScanlineU8()
	renderCtrl(ras, sl, rb, d.gammaSlider)
	renderCtrl(ras, sl, rb, d.rSlider)
	renderCtrl(ras, sl, rb, d.gSlider)
	renderCtrl(ras, sl, rb, d.bSlider)
	renderCtrl(ras, sl, rb, d.pattern)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	changed := d.gammaSlider.OnMouseButtonDown(fx, fy)
	if d.rSlider.OnMouseButtonDown(fx, fy) {
		changed = true
	}
	if d.gSlider.OnMouseButtonDown(fx, fy) {
		changed = true
	}
	if d.bSlider.OnMouseButtonDown(fx, fy) {
		changed = true
	}
	if d.pattern.OnMouseButtonDown(fx, fy) {
		changed = true
	}
	return changed
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	changed := d.gammaSlider.OnMouseMove(fx, fy, btn.Left)
	if d.rSlider.OnMouseMove(fx, fy, btn.Left) {
		changed = true
	}
	if d.gSlider.OnMouseMove(fx, fy, btn.Left) {
		changed = true
	}
	if d.bSlider.OnMouseMove(fx, fy, btn.Left) {
		changed = true
	}
	if d.pattern.OnMouseMove(fx, fy, btn.Left) {
		changed = true
	}
	return changed
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	changed := d.gammaSlider.OnMouseButtonUp(fx, fy)
	if d.rSlider.OnMouseButtonUp(fx, fy) {
		changed = true
	}
	if d.gSlider.OnMouseButtonUp(fx, fy) {
		changed = true
	}
	if d.bSlider.OnMouseButtonUp(fx, fy) {
		changed = true
	}
	if d.pattern.OnMouseButtonUp(fx, fy) {
		changed = true
	}
	return changed
}

func renderCtrl(ras *rasType, sl *isl.ScanlineU8, rb *renBase, c ctrlbase.Ctrl[icol.RGBA]) {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255.0 + 0.5)
	}
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(&rasterVertexSourceAdapter{src: c}, uint32(i))
		col := c.Color(i)
		renscan.RenderScanlinesAASolid(ras, sl, rb,
			icol.RGBA8[icol.Linear]{R: clamp(col.R), G: clamp(col.G), B: clamp(col.B), A: clamp(col.A)})
	}
}


func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AGG Example. Gamma Tuner",
		Width:  canvasW,
		Height: canvasH,
		FlipY:  true,
	}, newDemo())
}
