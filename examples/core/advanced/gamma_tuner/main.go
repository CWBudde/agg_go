// Port of AGG C++ gamma_tuner.cpp – per-channel gamma tuning with patterns.
//
// Renders a vertical color gradient with alpha-blended pattern overlays
// using per-channel gamma correction. The C++ original uses gamma_lut on the
// pixel format and blend_color_hspan for the pattern; this Go port computes
// the equivalent pixel values directly.
// Default: R=1.0, G=1.0, B=1.0, Gamma=2.2, pattern=Checkered (index 2).
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
)

const (
	canvasW = 500
	canvasH = 500

	defaultR       = 1.0
	defaultG       = 1.0
	defaultB       = 1.0
	defaultGamma   = 2.2
	defaultPattern = 2 // 0=Horizontal, 1=Vertical, 2=Checkered
	squareSize     = 400
	verStrips      = 5
)

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()
	data := img.Data
	stride := w * 4

	r0 := defaultR
	g0 := defaultG
	b0 := defaultB
	gamma := defaultGamma

	// Step 1: Draw vertical gradient background (full canvas height).
	// Matches C++ code: k = (i-80) / (squareSize-1), clamped, then
	// color = userColor.gradient(black, 1 - pow(k/2, 1/gamma))
	for y := 0; y < h; y++ {
		k := float64(y-80) / float64(squareSize-1)
		if k < 0 {
			k = 0
		}
		if k > 1 {
			k = 1
		}
		blend := 1 - math.Pow(k/2, 1/gamma)
		cr := uint8(clampF(r0*blend) * 255)
		cg := uint8(clampF(g0*blend) * 255)
		cb := uint8(clampF(b0*blend) * 255)
		for x := 0; x < w; x++ {
			idx := y*stride + x*4
			data[idx] = cr
			data[idx+1] = cg
			data[idx+2] = cb
			data[idx+3] = 255
		}
	}

	// Step 2: Clear the square area to black.
	for y := 80; y < 80+squareSize && y < h; y++ {
		for x := 50; x < 50+squareSize && x < w; x++ {
			idx := y*stride + x*4
			data[idx] = 0
			data[idx+1] = 0
			data[idx+2] = 0
			data[idx+3] = 255
		}
	}

	// Step 3: Draw the pattern (pairs of scanlines).
	// For each pair of rows (i, i+1), compute color from gradient,
	// then blend with alpha spans.
	for i := 0; i < squareSize; i += 2 {
		k := float64(i) / float64(squareSize-1)
		blend := 1 - math.Pow(k, 1/gamma)
		cr := uint8(clampF(r0*blend) * 255)
		cg := uint8(clampF(g0*blend) * 255)
		cb := uint8(clampF(b0*blend) * 255)

		for j := 0; j < squareSize; j++ {
			a1, a2 := computeAlpha(j, squareSize, defaultPattern)

			y1 := i + 80
			y2 := i + 80 + 1
			x := 50 + j
			if x >= w || y1 >= h {
				continue
			}

			// Blend span1 color onto row y1.
			blendPixel(data, y1*stride+x*4, cr, cg, cb, a1)

			// Blend span2 color onto row y2.
			if y2 < h {
				blendPixel(data, y2*stride+x*4, cr, cg, cb, a2)
			}
		}
	}

	// Step 4: Draw vertical strips.
	for i := 0; i < squareSize; i++ {
		k := float64(i) / float64(squareSize-1)
		blend := 1 - math.Pow(k/2, 1/gamma)
		cr := uint8(clampF(r0*blend) * 255)
		cg := uint8(clampF(g0*blend) * 255)
		cb := uint8(clampF(b0*blend) * 255)
		y := i + 80
		if y >= h {
			break
		}
		for j := 0; j < verStrips; j++ {
			xc := squareSize * (j + 1) / (verStrips + 1)
			for dx := -10; dx <= 10; dx++ {
				x := 50 + xc + dx
				if x >= 0 && x < w {
					idx := y*stride + x*4
					data[idx] = cr
					data[idx+1] = cg
					data[idx+2] = cb
					data[idx+3] = 255
				}
			}
		}
	}

	// Step 5: Draw border around the square.
	ctx := agg.NewContextForImage(img)
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.NoFill()
	a.LineColor(agg.NewColor(100, 100, 100, 150))
	a.LineWidth(1.0)
	a.ResetPath()
	a.MoveTo(50, 80)
	a.LineTo(float64(50+squareSize), 80)
	a.LineTo(float64(50+squareSize), float64(80+squareSize))
	a.LineTo(50, float64(80+squareSize))
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)
}

// computeAlpha returns alpha values for the two interleaved scanlines
// at column j within the square, based on pattern type.
func computeAlpha(j, size, pattern int) (a1, a2 uint8) {
	alpha := uint8(j * 255 / size)
	invAlpha := 255 - alpha

	switch pattern {
	case 0: // Horizontal - alternating alpha/invAlpha spans
		a1 = alpha
		a2 = invAlpha
	case 1: // Vertical - both use same alpha, alternating odd/even
		if j&1 != 0 {
			a1 = alpha
		} else {
			a1 = invAlpha
		}
		a2 = a1
	default: // Checkered
		if j&1 != 0 {
			a1 = alpha
			a2 = invAlpha
		} else {
			a2 = alpha
			a1 = invAlpha
		}
	}
	return a1, a2
}

// blendPixel alpha-blends (r, g, b, a) onto data[idx..idx+3].
func blendPixel(data []uint8, idx int, r, g, b, a uint8) {
	if a == 0 {
		return
	}
	if a == 255 {
		data[idx] = r
		data[idx+1] = g
		data[idx+2] = b
		data[idx+3] = 255
		return
	}
	alpha := uint32(a)
	invAlpha := 255 - alpha
	data[idx] = uint8((uint32(data[idx])*invAlpha + uint32(r)*alpha) / 255)
	data[idx+1] = uint8((uint32(data[idx+1])*invAlpha + uint32(g)*alpha) / 255)
	data[idx+2] = uint8((uint32(data[idx+2])*invAlpha + uint32(b)*alpha) / 255)
	data[idx+3] = 255
}

func clampF(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Gamma Tuner",
		Width:  canvasW,
		Height: canvasH,
	}, &demo{})
}
