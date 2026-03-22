// Port of the AGG C++ example simple_blur.cpp.
//
// Renders the AGG lion, then applies a simple 3x3 box-blur inside an ellipse
// and draws the ellipse outline on top — demonstrating basic pixel-level
// post-processing on a rendered scene.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

type demo struct {
	cx, cy float64
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Draw the lion centered on the canvas.
	drawLion(agg2d, img.Width(), img.Height())

	rx, ry := 100.0, 100.0

	// Snapshot the background before the ellipse outline is drawn so the
	// blur samples the clean lion pixels.
	bgImg := agg.CreateImage(img.Width(), img.Height())
	copy(bgImg.Data, img.Data)

	// Draw the ellipse outline over the lion (double-stroked like C++).
	agg2d.NoFill()
	agg2d.LineColor(agg.NewColor(0, 51, 0, 255))
	agg2d.LineWidth(6.0)
	agg2d.ResetPath()
	agg2d.AddEllipse(d.cx, d.cy, rx, ry, agg.CCW)
	agg2d.DrawPath(agg.StrokeOnly)

	// Apply 3x3 box-blur inside the ellipse using the pre-outline snapshot.
	applyBlurInsideEllipse(img, bgImg, d.cx, d.cy, rx, ry)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if btn.Left {
		d.cx = float64(x)
		d.cy = float64(y)
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if btn.Left {
		d.cx = float64(x)
		d.cy = float64(y)
		return true
	}
	return false
}

func (d *demo) OnMouseUp(_, _ int, _ lowlevelrunner.Buttons) bool { return false }

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
	agg2d.Translate(float64(width)*0.25, float64(height)*0.5)

	ld := liondemo.Parse()
	for i := 0; i < ld.NPaths; i++ {
		agg2d.FillColor(agg.NewColor(ld.Colors[i].R, ld.Colors[i].G, ld.Colors[i].B, 255))
		agg2d.NoLine()
		agg2d.ResetPath()
		ld.Path.Rewind(ld.PathIdx[i])
		for {
			x, y, cmd := ld.Path.NextVertex()
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
	lowlevelrunner.Run(lowlevelrunner.Config{Title: "Simple Blur", Width: 512, Height: 400}, &demo{
		cx: 100,
		cy: 102,
	})
}
