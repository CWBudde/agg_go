// Based on the original AGG examples: gamma_tuner.cpp.
package main

import (
	"math"
)

var (
	gammaTunerR       = 1.0
	gammaTunerG       = 1.0
	gammaTunerB       = 1.0
	gammaTunerGamma   = 2.2
	gammaTunerPattern = 2 // 0=Horizontal, 1=Vertical, 2=Checkered
)

func drawGammaTunerDemo() {
	g := gammaTunerGamma
	r := gammaTunerR
	gg := gammaTunerG
	b := gammaTunerB

	const (
		verStrips = 5
	)

	padX := width / 10
	if padX < 24 {
		padX = 24
	}
	padY := height / 20
	if padY < 24 {
		padY = 24
	}
	squareSize := width - 2*padX
	if alt := height - 2*padY; alt < squareSize {
		squareSize = alt
	}
	if squareSize < 2 {
		squareSize = 2
	}
	squareX := (width - squareSize) / 2
	squareY := (height - squareSize) / 2
	span1 := make([]float64, squareSize)
	span2 := make([]float64, squareSize)

	// Pre-calculate colors for the gradient
	invG := 1.0 / g
	gradientKForScreenY := func(y int) float64 {
		srcY := float64(height-1-y) / float64(height-1)
		return 1.0 - math.Pow(srcY*0.5, invG)
	}

	// 1. Draw vertical gradient directly into canvasBuf.
	// The upstream demo uses flip_y=true, so evaluate the gradient from the
	// mirrored Y coordinate to match the original orientation.
	for y := 0; y < height; y++ {
		k := gradientKForScreenY(y)

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
	for y := squareY; y < squareY+squareSize; y++ {
		rowOffset := y * width * 4
		for x := squareX; x < squareX+squareSize; x++ {
			idx := rowOffset + x*4
			canvasBuf[idx] = 0
			canvasBuf[idx+1] = 0
			canvasBuf[idx+2] = 0
			canvasBuf[idx+3] = 255
		}
	}

	// 3. Build the original AGG horizontal alpha spans.
	switch gammaTunerPattern {
	case 0: // Horizontal
		for i := 0; i < squareSize; i++ {
			a := float64(i) / float64(squareSize)
			span1[i] = a
			span2[i] = 1.0 - a
		}
	case 1: // Vertical
		for i := 0; i < squareSize; i++ {
			a := float64(i) / float64(squareSize)
			if i&1 != 0 {
				span1[i] = a
				span2[i] = a
			} else {
				span1[i] = 1.0 - a
				span2[i] = 1.0 - a
			}
		}
	default: // Checkered
		for i := 0; i < squareSize; i++ {
			a := float64(i) / float64(squareSize)
			if i&1 != 0 {
				span1[i] = a
				span2[i] = 1.0 - a
			} else {
				span1[i] = 1.0 - a
				span2[i] = a
			}
		}
	}

	// 4. Draw pattern directly into canvasBuf
	for i := 0; i < squareSize; i += 2 {
		k := float64(squareSize-1-i) / float64(squareSize-1)
		k = 1.0 - math.Pow(k, invG)

		pcr := r * 255.0 * (1.0 - k)
		pcg := gg * 255.0 * (1.0 - k)
		pcb := b * 255.0 * (1.0 - k)

		y1 := squareY + i
		y2 := squareY + i + 1
		row1 := y1 * width * 4
		row2 := y2 * width * 4

		for j := 0; j < squareSize; j++ {
			idx1 := row1 + (squareX+j)*4
			canvasBuf[idx1] = uint8(pcr*span1[j] + 0.5)
			canvasBuf[idx1+1] = uint8(pcg*span1[j] + 0.5)
			canvasBuf[idx1+2] = uint8(pcb*span1[j] + 0.5)

			idx2 := row2 + (squareX+j)*4
			canvasBuf[idx2] = uint8(pcr*span2[j] + 0.5)
			canvasBuf[idx2+1] = uint8(pcg*span2[j] + 0.5)
			canvasBuf[idx2+2] = uint8(pcb*span2[j] + 0.5)
		}
	}

	// 5. Draw vertical strips
	for i := 0; i < squareSize; i++ {
		y := squareY + i
		k := gradientKForScreenY(y)
		cr := uint8(r*255.0*(1.0-k) + 0.5)
		cg := uint8(gg*255.0*(1.0-k) + 0.5)
		cb := uint8(b*255.0*(1.0-k) + 0.5)

		rowOffset := y * width * 4
		for j := 0; j < verStrips; j++ {
			xc := squareSize * (j + 1) / (verStrips + 1)
			for dx := -10; dx <= 10; dx++ {
				x := squareX + xc + dx
				if x >= 0 && x < width {
					idx := rowOffset + x*4
					canvasBuf[idx] = cr
					canvasBuf[idx+1] = cg
					canvasBuf[idx+2] = cb
				}
			}
		}
	}
}
