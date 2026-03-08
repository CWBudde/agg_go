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

		k = 1.0 - math.Pow(k*0.5, invG)

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
	curPattern := gammaTunerPattern
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
		k = 1.0 - math.Pow(k*0.5, invG)

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
}
