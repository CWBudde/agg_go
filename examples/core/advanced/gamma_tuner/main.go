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
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	pixgamma "github.com/MeKo-Christian/agg_go/internal/pixfmt/gamma"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
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

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())
	pixf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]](pixf)

	rVal := d.rSlider.Value()
	gVal := d.gSlider.Value()
	bVal := d.bSlider.Value()
	gammaVal := d.gammaSlider.Value()
	lut := pixgamma.NewSimpleGammaLut(gammaVal)

	baseColor := func(a float64) icol.RGBA8[icol.Linear] {
		return icol.RGBA8[icol.Linear]{
			R: gammaByte(lut, rVal*a),
			G: gammaByte(lut, gVal*a),
			B: gammaByte(lut, bVal*a),
			A: 255,
		}
	}

	// Vertical gradient background.
	for y := 0; y < h; y++ {
		k := float64(y-80) / float64(squareSize-1)
		if y < 80 {
			k = 0
		}
		if y >= 80+squareSize {
			k = 1
		}

		k = 1 - math.Pow(k/2, 1/gammaVal)
		rb.CopyHline(0, y, w-1, baseColor(k))
	}

	// Pattern setup.
	span1 := make([]icol.RGBA8[icol.Linear], squareSize)
	span2 := make([]icol.RGBA8[icol.Linear], squareSize)
	buildSpans := func() {
		for i := 0; i < squareSize; i++ {
			a1, a2 := patternAlpha(i, squareSize, d.pattern.CurItem())
			span1[i] = icol.RGBA8[icol.Linear]{R: 0, G: 0, B: 0, A: a1}
			span2[i] = icol.RGBA8[icol.Linear]{R: 0, G: 0, B: 0, A: a2}
		}
	}
	buildSpans()

	// Clear the square.
	rb.CopyBar(50, 80, 50+squareSize-1, 80+squareSize-1, icol.RGBA8[icol.Linear]{R: 0, G: 0, B: 0, A: 255})

	// Draw the pattern.
	for i := 0; i < squareSize; i += 2 {
		k := float64(i) / float64(squareSize-1)
		k = 1 - math.Pow(k, 1/gammaVal)
		c := baseColor(k)

		for j := 0; j < squareSize; j++ {
			span1[j].R, span1[j].G, span1[j].B = c.R, c.G, c.B
			span2[j].R, span2[j].G, span2[j].B = c.R, c.G, c.B
		}

		rb.BlendColorHspan(50, i+80, squareSize, span1, nil, basics.CoverFull)
		rb.BlendColorHspan(50, i+80+1, squareSize, span2, nil, basics.CoverFull)
	}

	// Draw vertical strips.
	for i := 0; i < squareSize; i++ {
		k := float64(i) / float64(squareSize-1)
		k = 1 - math.Pow(k/2, 1/gammaVal)
		c := baseColor(k)
		y := i + 80
		for j := 0; j < verStrips; j++ {
			xc := squareSize * (j + 1) / (verStrips + 1)
			rb.CopyHline(50+xc-10, y, 50+xc+10, c)
		}
	}

	// Border around the square.
	border := icol.RGBA8[icol.Linear]{R: 100, G: 100, B: 100, A: 150}
	rb.CopyHline(50, 80, 50+squareSize-1, border)
	rb.CopyHline(50, 80+squareSize-1, 50+squareSize-1, border)
	rb.CopyVline(50, 80, 80+squareSize-1, border)
	rb.CopyVline(50+squareSize-1, 80, 80+squareSize-1, border)

	// Controls on top.
	ctx := agg.NewContextForImage(img)
	a := ctx.GetAgg2D()
	renderCtrl(a, d.gammaSlider)
	renderCtrl(a, d.rSlider)
	renderCtrl(a, d.gSlider)
	renderCtrl(a, d.bSlider)
	renderCtrl(a, d.pattern)
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

func gammaByte(lut *pixgamma.SimpleGammaLut, v float64) uint8 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return lut.Dir(uint8(v*255.0 + 0.5))
}

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(&rasterVertexSourceAdapter{src: c}, uint32(i))
		a.RenderRasterizerWithColor(toAggColor(c.Color(i)))
	}
}

func toAggColor(c icol.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255.0 + 0.5)
		}
	}
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

func patternAlpha(j, size, pattern int) (uint8, uint8) {
	alpha := uint8(j * 255 / size)
	invAlpha := 255 - alpha

	switch pattern {
	case 0:
		return alpha, invAlpha
	case 1:
		if j&1 != 0 {
			return alpha, alpha
		}
		return invAlpha, invAlpha
	default:
		if j&1 != 0 {
			return alpha, invAlpha
		}
		return invAlpha, alpha
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
