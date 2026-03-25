// Port of the AGG C++ example simple_blur.cpp.
//
// Renders the AGG lion, then applies a simple 3x3 box-blur inside an ellipse
// and draws the ellipse outline on top — demonstrating basic pixel-level
// post-processing on a rendered scene.
package main

import (
	"math"
	"sync"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	lionOnce sync.Once
	lionData *liondemo.LionData
)

type demo struct {
	cx, cy float64
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Draw the lion fill on the left, then the colored outline pass on the right.
	drawLionFill(agg2d, img.Width(), img.Height())
	drawLionOutline(agg2d, img.Width(), img.Height())

	rx, ry := 100.0, 100.0

	// Snapshot the background before the ellipse outline is drawn so the
	// blur samples the clean lion pixels.
	bgImg := agg.NewImage(make([]uint8, len(img.Data)), img.Width(), img.Height(), img.Stride())
	copy(bgImg.Data, img.Data)

	// Draw the ellipse outline over the lion.
	agg2d.NoFill()
	agg2d.LineColor(agg.NewColor(0, 51, 0, 255))
	agg2d.LineWidth(2.0)
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

func drawLionFill(agg2d *agg.Agg2D, width, height int) {
	ld := liondata()
	x1, y1, x2, y2 := getLionBoundingRect(ld)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	m := transform.NewTransAffine()
	m.Translate(-cx, -cy)
	m.Scale(1.0)
	m.Rotate(math.Pi)
	m.Translate(float64(width)*0.25, float64(height)*0.5)

	agg2d.NoLine()
	for i := 0; i < ld.NPaths; i++ {
		agg2d.FillColor(agg.NewColor(ld.Colors[i].R, ld.Colors[i].G, ld.Colors[i].B, 255))
		agg2d.ResetPath()
		ld.Path.Rewind(ld.PathIdx[i])
		for {
			x, y, cmd := ld.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			m.Transform(&x, &y)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(x, y)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(x, y)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}
}

func drawLionOutline(agg2d *agg.Agg2D, width, height int) {
	ld := liondata()
	x1, y1, x2, y2 := getLionBoundingRect(ld)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	m := transform.NewTransAffine()
	m.Translate(-cx, -cy)
	m.Scale(1.0)
	m.Rotate(math.Pi)
	m.Translate(float64(width)*0.75, float64(height)*0.5)

	agg2d.NoFill()
	agg2d.LineWidth(1.0)
	for i := 0; i < ld.NPaths; i++ {
		agg2d.LineColor(agg.NewColor(ld.Colors[i].R, ld.Colors[i].G, ld.Colors[i].B, 255))
		agg2d.ResetPath()
		ld.Path.Rewind(ld.PathIdx[i])
		for {
			x, y, cmd := ld.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			m.Transform(&x, &y)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(x, y)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(x, y)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.StrokeOnly)
	}
}

// applyBlurInsideEllipse performs a 3x3 box-blur on dst for all pixels inside
// the ellipse defined by (cx, cy, rx, ry), sampling from src.
func applyBlurInsideEllipse(dst, src *agg.Image, cx, cy, rx, ry float64) {
	w, h := dst.Width(), dst.Height()
	dstData := dst.Data
	srcData := src.Data
	dstStride := dst.Stride()
	srcStride := src.Stride()
	dstBase := 0
	srcBase := 0
	if dstStride < 0 {
		dstBase = (h - 1) * -dstStride
	}
	if srcStride < 0 {
		srcBase = (h - 1) * -srcStride
	}

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
				rowOff := srcBase + (y+iy)*srcStride
				for ix := -1; ix <= 1; ix++ {
					idx := rowOff + (x+ix)*4
					r += uint32(srcData[idx])
					g += uint32(srcData[idx+1])
					b += uint32(srcData[idx+2])
					a += uint32(srcData[idx+3])
				}
			}
			dstIdx := dstBase + y*dstStride + x*4
			dstData[dstIdx] = uint8(r / 9)
			dstData[dstIdx+1] = uint8(g / 9)
			dstData[dstIdx+2] = uint8(b / 9)
			dstData[dstIdx+3] = uint8(a / 9)
		}
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{Title: "Simple Blur", Width: 512, Height: 400, FlipY: true, DisableLinearRGBToSRGB: true}, &demo{
		cx: 100,
		cy: 102,
	})
}

func liondata() *liondemo.LionData {
	lionOnce.Do(func() {
		ld := liondemo.Parse()
		lionData = &ld
	})
	return lionData
}

func getLionBoundingRect(ld *liondemo.LionData) (x1, y1, x2, y2 float64) {
	x1, y1, x2, y2 = math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for idx := uint(0); idx < ld.Path.TotalVertices(); idx++ {
		x, y, cmd := ld.Path.Vertex(idx)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			if x < x1 {
				x1 = x
			}
			if y < y1 {
				y1 = y
			}
			if x > x2 {
				x2 = x
			}
			if y > y2 {
				y2 = y
			}
		}
	}
	return
}
