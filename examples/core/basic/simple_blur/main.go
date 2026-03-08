// Port of the AGG C++ example simple_blur.cpp.
//
// Renders the AGG lion, then applies a simple 3x3 box-blur inside an ellipse
// and draws the ellipse outline on top — demonstrating basic pixel-level
// post-processing on a rendered scene.
package main

import (
	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/basics"
	liondemo "agg_go/internal/demo/lion"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Draw the lion centered on the canvas.
	drawLion(agg2d, ctx.Width(), ctx.Height())

	// Blur parameters — default values from the WASM demo.
	cx, cy := 400.0, 300.0
	rx, ry := 100.0, 100.0

	// Snapshot the background before the ellipse outline is drawn so the
	// blur samples the clean lion pixels.
	bgImg := agg.CreateImage(ctx.Width(), ctx.Height())
	copy(bgImg.Data, ctx.GetImage().Data)

	// Draw the ellipse outline over the lion.
	agg2d.NoFill()
	agg2d.LineColor(agg.NewColor(0, 51, 0, 255)) // dark green, ~rgba(0, 0.2, 0)
	agg2d.LineWidth(2.0)
	agg2d.ResetPath()
	agg2d.AddEllipse(cx, cy, rx, ry, agg.CCW)
	agg2d.DrawPath(agg.StrokeOnly)

	// Apply 3x3 box-blur inside the ellipse using the pre-outline snapshot.
	applyBlurInsideEllipse(ctx.GetImage(), bgImg, cx, cy, rx, ry)
}

// drawLion renders the AGG lion demo into agg2d, centered in the canvas.
func drawLion(agg2d *agg.Agg2D, width, height int) {
	const (
		lionX1, lionY1 = 21.0, 9.0
		lionX2, lionY2 = 478.0, 442.0
	)

	baseDX := (lionX2 - lionX1) * 0.5
	baseDY := (lionY2 - lionY1) * 0.5

	agg2d.PushTransform()
	defer agg2d.PopTransform()

	agg2d.Translate(-baseDX, -baseDY)
	agg2d.Scale(1.0, 1.0)
	agg2d.Translate(float64(width)*0.5, float64(height)*0.5)

	for _, lp := range liondemo.Parse() {
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

// applyBlurInsideEllipse performs a 3x3 box-blur on dst for all pixels inside
// the ellipse defined by (cx, cy, rx, ry), sampling from src.
func applyBlurInsideEllipse(dst, src *agg.Image, cx, cy, rx, ry float64) {
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
			if dx*dx/rx2+dy2/ry2 > 1.0 {
				continue // outside ellipse
			}
			if x == 0 || x == w-1 || y == 0 || y == h-1 {
				continue // skip border pixels
			}
			var r, g, b, a uint32
			for iy := -1; iy <= 1; iy++ {
				rowOff := (y + iy) * stride
				for ix := -1; ix <= 1; ix++ {
					idx := rowOff + (x+ix)*4
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

func main() {
	demorunner.Run(demorunner.Config{Title: "Simple Blur", Width: 800, Height: 600}, &demo{})
}
