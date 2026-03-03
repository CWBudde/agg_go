// Based on the original AGG examples: gamma_tuner.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/ctrl/rbox"
	"agg_go/internal/ctrl/slider"
)

var (
	sliderR       *slider.SliderCtrl
	sliderG       *slider.SliderCtrl
	sliderB       *slider.SliderCtrl
	sliderGamma   *slider.SliderCtrl
	patternRBox   *rbox.RboxCtrl[agg.Color]
	tunerInitialized bool
)

func initGammaTunerDemo() {
	if tunerInitialized {
		return
	}

	sliderR = slider.NewSliderCtrl(5, 5, 345, 16, false)
	sliderR.SetLabel("R=%.2f")
	sliderR.SetValue(1.0)

	sliderG = slider.NewSliderCtrl(5, 20, 345, 31, false)
	sliderG.SetLabel("G=%.2f")
	sliderG.SetValue(1.0)

	sliderB = slider.NewSliderCtrl(5, 35, 345, 46, false)
	sliderB.SetLabel("B=%.2f")
	sliderB.SetValue(1.0)

	sliderGamma = slider.NewSliderCtrl(5, 50, 345, 61, false)
	sliderGamma.SetLabel("Gamma=%.2f")
	sliderGamma.SetRange(0.5, 4.0)
	sliderGamma.SetValue(2.2)

	patternRBox = rbox.NewRboxCtrl[agg.Color](355, 1, 495, 60, false,
		agg.White, agg.Black, agg.Black, agg.Gray, agg.Red)
	patternRBox.AddItem("Horizontal")
	patternRBox.AddItem("Vertical")
	patternRBox.AddItem("Checkered")
	patternRBox.SetCurItem(2)

	tunerInitialized = true
}

func drawGammaTunerDemo() {
	initGammaTunerDemo()

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	g := sliderGamma.Value()
	r := sliderR.Value()
	gg := sliderG.Value()
	b := sliderB.Value()

	const (
		squareSize = 400
		verStrips  = 5
	)

	// Draw vertical gradient
	w, h := float64(width), float64(height)
	for i := 0.0; i < h; i++ {
		k := (i - 80) / (squareSize - 1)
		if i < 80 {
			k = 0.0
		}
		if i >= 80+squareSize {
			k = 1.0
		}

		k = 1.0 - math.Pow(k/2.0, 1.0/g)
		
		cr := uint8(math.Round(r * 255.0 * (1.0 - k)))
		cg := uint8(math.Round(gg * 255.0 * (1.0 - k)))
		cb := uint8(math.Round(b * 255.0 * (1.0 - k)))

		agg2d.LineColor(agg.NewColor(cr, cg, cb, 255))
		agg2d.Line(0, i, w-1, i)
	}

	// Clear the area
	agg2d.FillColor(agg.Black)
	agg2d.NoLine()
	agg2d.Rectangle(50, 80, 50+squareSize-1, 80+squareSize-1)

	// Draw the pattern
	curPattern := patternRBox.CurItem()
	for i := 0; i < squareSize; i += 2 {
		k := float64(i) / (squareSize - 1)
		k = 1.0 - math.Pow(k, 1.0/g)

		cr := r * 255.0 * (1.0 - k)
		cg := gg * 255.0 * (1.0 - k)
		cb := b * 255.0 * (1.0 - k)

		for j := 0; j < squareSize; j++ {
			var alpha1, alpha2 float64
			kj := float64(j) / (squareSize - 1)

			switch curPattern {
			case 0: // Horizontal
				alpha1 = kj * 255.0
				alpha2 = 255.0 - alpha1
			case 1: // Vertical
				if j&1 != 0 {
					alpha1 = kj * 255.0
					alpha2 = alpha1
				} else {
					alpha1 = 255.0 - kj*255.0
					alpha2 = alpha1
				}
			case 2: // Checkered
				if j&1 != 0 {
					alpha1 = kj * 255.0
					alpha2 = 255.0 - alpha1
				} else {
					alpha2 = kj * 255.0
					alpha1 = 255.0 - alpha2
				}
			}

			agg2d.FillColor(agg.NewColor(uint8(math.Round(cr)), uint8(math.Round(cg)), uint8(math.Round(cb)), uint8(math.Round(alpha1))))
			agg2d.Rectangle(50+float64(j), 80+float64(i), 51+float64(j), 81+float64(i))
			
			agg2d.FillColor(agg.NewColor(uint8(math.Round(cr)), uint8(math.Round(cg)), uint8(math.Round(cb)), uint8(math.Round(alpha2))))
			agg2d.Rectangle(50+float64(j), 80+float64(i+1), 51+float64(j), 81+float64(i+1))
		}
	}

	// Draw vertical strips
	for i := 0.0; i < squareSize; i++ {
		k := i / (squareSize - 1)
		k = 1.0 - math.Pow(k/2.0, 1.0/g)
		
		cr := uint8(math.Round(r * 255.0 * (1.0 - k)))
		cg := uint8(math.Round(gg * 255.0 * (1.0 - k)))
		cb := uint8(math.Round(b * 255.0 * (1.0 - k)))
		agg2d.LineColor(agg.NewColor(cr, cg, cb, 255))

		for j := 0; j < verStrips; j++ {
			xc := float64(squareSize * (j + 1) / (verStrips + 1))
			agg2d.Line(50+xc-10, i+80, 50+xc+10, i+80)
		}
	}

	// Render controls
	renderSlider(agg2d, sliderR)
	renderSlider(agg2d, sliderG)
	renderSlider(agg2d, sliderB)
	renderSlider(agg2d, sliderGamma)
	renderRBox(agg2d, patternRBox)
}

func renderSlider(agg2d *agg.Agg2D, s *slider.SliderCtrl) {
	ras := agg2d.GetInternalRasterizer()
	numPaths := s.NumPaths()
	for i := uint(0); i < numPaths; i++ {
		ras.Reset()
		s.Rewind(i)
		adapter := &sliderAdapter{s: s}
		ras.AddPath(adapter, 0)
		c := s.Color(i)
		agg2d.FillColor(agg.RGBA(c.R, c.G, c.B, c.A))
		agg2d.DrawPath(agg.FillOnly)
	}
}

type sliderAdapter struct {
	s *slider.SliderCtrl
}

func (a *sliderAdapter) Rewind(pathID uint32) {
	a.s.Rewind(uint(pathID))
}

func (a *sliderAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.s.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func renderRBox(agg2d *agg.Agg2D, r *rbox.RboxCtrl[agg.Color]) {
	ras := agg2d.GetInternalRasterizer()
	numPaths := r.NumPaths()
	for i := uint(0); i < numPaths; i++ {
		ras.Reset()
		r.Rewind(i)
		adapter := &rboxAdapter{r: r}
		ras.AddPath(adapter, 0)
		agg2d.FillColor(r.Color(i))
		agg2d.DrawPath(agg.FillOnly)
	}
}

type rboxAdapter struct {
	r *rbox.RboxCtrl[agg.Color]
}

func (a *rboxAdapter) Rewind(pathID uint32) {
	a.r.Rewind(uint(pathID))
}

func (a *rboxAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.r.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func handleGammaTunerMouseDown(x, y float64) bool {
	if !tunerInitialized {
		return false
	}
	if sliderR.OnMouseButtonDown(x, y) { return true }
	if sliderG.OnMouseButtonDown(x, y) { return true }
	if sliderB.OnMouseButtonDown(x, y) { return true }
	if sliderGamma.OnMouseButtonDown(x, y) { return true }
	if patternRBox.OnMouseButtonDown(x, y) { return true }
	return false
}

func handleGammaTunerMouseMove(x, y float64) bool {
	if !tunerInitialized {
		return false
	}
	if sliderR.OnMouseMove(x, y, true) { return true }
	if sliderG.OnMouseMove(x, y, true) { return true }
	if sliderB.OnMouseMove(x, y, true) { return true }
	if sliderGamma.OnMouseMove(x, y, true) { return true }
	if patternRBox.OnMouseMove(x, y, true) { return true }
	return false
}

func handleGammaTunerMouseUp() {
	if !tunerInitialized {
		return
	}
	sliderR.OnMouseButtonUp(0, 0)
	sliderG.OnMouseButtonUp(0, 0)
	sliderB.OnMouseButtonUp(0, 0)
	sliderGamma.OnMouseButtonUp(0, 0)
	patternRBox.OnMouseButtonUp(0, 0)
}
