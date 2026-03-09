// Based on the original AGG examples: simple_blur.cpp.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

var (
	simpleBlurCX = 400.0
	simpleBlurCY = 300.0
)

func drawSimpleBlurDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Clear background
	agg2d.ClearAll(agg.White)

	// 2. Draw Lion
	drawLionToAgg2D(agg2d, 1.0)

	// 3. Draw blurred ellipse
	rx, ry := 100.0, 100.0

	// We need the background before the ellipse outline for the blur source
	bgImg := agg.CreateImage(width, height)
	copy(bgImg.Data, ctx.GetImage().Data)

	// Draw ellipse outline
	agg2d.NoFill()
	agg2d.LineColor(agg.NewColor(0, 51, 0, 255)) // rgba(0, 0.2, 0)
	agg2d.LineWidth(2.0)
	agg2d.ResetPath()
	agg2d.AddEllipse(simpleBlurCX, simpleBlurCY, rx, ry, agg.CCW)
	agg2d.DrawPath(agg.StrokeOnly)

	// 4. Apply simple 3x3 blur inside the ellipse
	applySimpleBlurInsideEllipse(ctx.GetImage(), bgImg, simpleBlurCX, simpleBlurCY, rx, ry)
}

func applySimpleBlurInsideEllipse(dst, src *agg.Image, cx, cy, rx, ry float64) {
	w, h := dst.Width(), dst.Height()
	dstData := dst.Data
	srcData := src.Data
	stride := w * 4

	rx2 := rx * rx
	ry2 := ry * ry

	for y := 0; y < h; y++ {
		dy := float64(y) - cy
		dy2 := dy * dy
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			dx2 := dx * dx

			// Check if inside ellipse: (x-cx)^2/rx^2 + (y-cy)^2/ry^2 <= 1
			if dx2/rx2+dy2/ry2 <= 1.0 {
				// 3x3 box blur
				if x > 0 && x < w-1 && y > 0 && y < h-1 {
					var r, g, b, a uint32
					for iy := -1; iy <= 1; iy++ {
						rowOffset := (y + iy) * stride
						for ix := -1; ix <= 1; ix++ {
							idx := rowOffset + (x+ix)*4
							r += uint32(srcData[idx])
							g += uint32(srcData[idx+1])
							b += uint32(srcData[idx+2])
							a += uint32(srcData[idx+3])
						}
					}
					dstIdx := y*stride + x*4
					dstData[dstIdx] = uint8(r / 9)
					dstData[dstIdx+1] = uint8(g / 9)
					dstData[dstIdx+2] = uint8(b / 9)
					dstData[dstIdx+3] = uint8(a / 9)
				}
			}
		}
	}
}

func drawLionToAgg2D(agg2d *agg.Agg2D, scale float64) {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	agg2d.PushTransform()
	defer agg2d.PopTransform()

	// Boundary for the lion in the original parse_lion.cpp
	const (
		lionX1, lionY1 = 21, 9
		lionX2, lionY2 = 478, 442
	)

	baseDX := (lionX2 - lionX1) * 0.5
	baseDY := (lionY2 - lionY1) * 0.5

	agg2d.Translate(-baseDX, -baseDY)
	agg2d.Scale(scale, scale)
	agg2d.Translate(float64(width)*0.5, float64(height)*0.5)

	for _, lp := range lionPaths {
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		agg2d.NoLine()
		agg2d.ResetPath()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(x, y)
			} else {
				agg2d.LineTo(x, y)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}
}
