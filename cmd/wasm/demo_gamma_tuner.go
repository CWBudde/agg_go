// Based on the original AGG examples: gamma_tuner.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/ctrl/rbox"
	"agg_go/internal/ctrl/slider"
)

var (
	sliderR          *slider.SliderCtrl
	sliderG          *slider.SliderCtrl
	sliderB          *slider.SliderCtrl
	sliderGamma      *slider.SliderCtrl
	patternRBox      *rbox.RboxCtrl[agg.Color]
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

	g := sliderGamma.Value()
	r := sliderR.Value()
	gg := sliderG.Value()
	b := sliderB.Value()

	const (
		squareSize = 400
		verStrips  = 5
	)

	// Pre-calculate colors for the gradient
	invG := 1.0 / g

	// 1. Draw vertical gradient directly into canvasBuf
	for y := 0; y < height; y++ {
		k := (float64(y) - 80) / (squareSize - 1)
		if y < 80 {
			k = 0.0
		} else if y >= 80+squareSize {
			k = 1.0
		}

		k = 1.0 - math.Pow(k/2.0, invG)

		cr := uint8(r*255.0*(1.0-k) + 0.5)
		cg := uint8(gg*255.0*(1.0-k) + 0.5)
		cb := uint8(b*255.0*(1.0-k) + 0.5)

		rowOffset := y * width * 4
		for x := 0; x < width; x++ {
			idx := rowOffset + x*4
			canvasBuf[idx] = cr
			canvasBuf[idx+1] = cg
			canvasBuf[idx+2] = cb
			canvasBuf[idx+3] = 255
		}
	}

	// 2. Clear square area to black
	for y := 80; y < 80+squareSize; y++ {
		rowOffset := y * width * 4
		for x := 50; x < 50+squareSize; x++ {
			idx := rowOffset + x*4
			canvasBuf[idx] = 0
			canvasBuf[idx+1] = 0
			canvasBuf[idx+2] = 0
			canvasBuf[idx+3] = 255
		}
	}

	// 3. Draw pattern directly into canvasBuf
	curPattern := patternRBox.CurItem()
	for i := 0; i < squareSize; i += 2 {
		k := float64(i) / (squareSize - 1)
		k = 1.0 - math.Pow(k, invG)

		pcr := r * 255.0 * (1.0 - k)
		pcg := gg * 255.0 * (1.0 - k)
		pcb := b * 255.0 * (1.0 - k)

		y1 := 80 + i
		y2 := 80 + i + 1
		row1 := y1 * width * 4
		row2 := y2 * width * 4

		for j := 0; j < squareSize; j++ {
			var alpha1, alpha2 float64
			kj := float64(j) / (squareSize - 1)

			switch curPattern {
			case 0: // Horizontal
				alpha1 = kj
				alpha2 = 1.0 - kj
			case 1: // Vertical
				if j&1 != 0 {
					alpha1 = kj
					alpha2 = alpha1
				} else {
					alpha1 = 1.0 - kj
					alpha2 = alpha1
				}
			case 2: // Checkered
				if j&1 != 0 {
					alpha1 = kj
					alpha2 = 1.0 - kj
				} else {
					alpha2 = kj
					alpha1 = 1.0 - alpha2
				}
			}

			idx1 := row1 + (50+j)*4
			canvasBuf[idx1] = uint8(pcr*alpha1 + 0.5)
			canvasBuf[idx1+1] = uint8(pcg*alpha1 + 0.5)
			canvasBuf[idx1+2] = uint8(pcb*alpha1 + 0.5)

			idx2 := row2 + (50+j)*4
			canvasBuf[idx2] = uint8(pcr*alpha2 + 0.5)
			canvasBuf[idx2+1] = uint8(pcg*alpha2 + 0.5)
			canvasBuf[idx2+2] = uint8(pcb*alpha2 + 0.5)
		}
	}

	// 4. Draw vertical strips
	for i := 0; i < squareSize; i++ {
		k := float64(i) / (squareSize - 1)
		k = 1.0 - math.Pow(k/2.0, invG)

		cr := uint8(r*255.0*(1.0-k) + 0.5)
		cg := uint8(gg*255.0*(1.0-k) + 0.5)
		cb := uint8(b*255.0*(1.0-k) + 0.5)

		y := 80 + i
		rowOffset := y * width * 4
		for j := 0; j < verStrips; j++ {
			xc := squareSize * (j + 1) / (verStrips + 1)
			for dx := -10; dx <= 10; dx++ {
				x := 50 + xc + dx
				if x >= 0 && x < width {
					idx := rowOffset + x*4
					canvasBuf[idx] = cr
					canvasBuf[idx+1] = cg
					canvasBuf[idx+2] = cb
				}
			}
		}
	}

	// 5. Render controls
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
		adapter := &sliderAdapter{s: s}
		ras.AddPath(adapter, uint32(i))
		c := s.Color(i)
		agg2d.RenderRasterizerWithColor(agg.RGBA(c.R, c.G, c.B, c.A))
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
		adapter := &rboxAdapter{r: r}
		ras.AddPath(adapter, uint32(i))
		agg2d.RenderRasterizerWithColor(r.Color(i))
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
	if sliderR.OnMouseButtonDown(x, y) {
		return true
	}
	if sliderG.OnMouseButtonDown(x, y) {
		return true
	}
	if sliderB.OnMouseButtonDown(x, y) {
		return true
	}
	if sliderGamma.OnMouseButtonDown(x, y) {
		return true
	}
	if patternRBox.OnMouseButtonDown(x, y) {
		return true
	}
	return false
}

func handleGammaTunerMouseMove(x, y float64) bool {
	if !tunerInitialized {
		return false
	}
	if sliderR.OnMouseMove(x, y, true) {
		return true
	}
	if sliderG.OnMouseMove(x, y, true) {
		return true
	}
	if sliderB.OnMouseMove(x, y, true) {
		return true
	}
	if sliderGamma.OnMouseMove(x, y, true) {
		return true
	}
	if patternRBox.OnMouseMove(x, y, true) {
		return true
	}
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
