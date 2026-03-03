// Based on the original AGG examples: lion_lens.cpp.
package main

import (
	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/conv"
	"agg_go/internal/ctrl/slider"
	liondemo "agg_go/internal/demo/lion"
	"agg_go/internal/transform"
)

var (
	lionLensMagnSlider   *slider.SliderCtrl
	lionLensRadiusSlider *slider.SliderCtrl
	lionLensX, lionLensY float64
	lionLensInitialized  bool
)

func initLionLensDemo() {
	if lionLensInitialized {
		return
	}

	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	lionLensMagnSlider = slider.NewSliderCtrl(5, 5, 495, 17, false)
	lionLensMagnSlider.SetRange(0.01, 4.0)
	lionLensMagnSlider.SetValue(3.0)
	lionLensMagnSlider.SetLabel("Scale=%3.2f")

	lionLensRadiusSlider = slider.NewSliderCtrl(5, 20, 495, 32, false)
	lionLensRadiusSlider.SetRange(0.0, 100.0)
	lionLensRadiusSlider.SetValue(70.0)
	lionLensRadiusSlider.SetLabel("Radius=%3.2f")

	lionLensX = 200
	lionLensY = 150
	lionLensInitialized = true
}

func drawLionLensDemo() {
	initLionLensDemo()

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	// Set up the lens
	lens := transform.NewTransWarpMagnifier()
	lens.SetCenter(lionLensX, lionLensY)
	lens.SetMagnification(lionLensMagnSlider.Value())
	lens.SetRadius(lionLensRadiusSlider.Value() / lionLensMagnSlider.Value())

	// Set up the base transformation for the lion
	g_x1, g_y1, g_x2, g_y2 := getLionBoundingRect(lionPaths)
	base_dx := (g_x2 - g_x1) / 2.0
	base_dy := (g_y2 - g_y1) / 2.0

	mtx := transform.NewTransAffine()
	mtx.Translate(-base_dx, -base_dy)
	mtx.Rotate(agg.Pi)
	mtx.Translate(float64(width)/2.0, float64(height)/2.0)

	ras := agg2d.GetInternalRasterizer()

	for i := range lionPaths {
		lp := &lionPaths[i]
		// Pipeline: Path -> Segmentator -> Transform (Affine) -> Transform (Lens)
		// We need to segment the path because the lens transform is non-linear
		
		// Use adapter for PathStorageStl to satisfy conv.VertexSource
		pathSource := &pathConvAdapter{ps: lp}
		segm := conv.NewConvSegmentator(pathSource)

		// Segm adapter to satisfy conv.VertexSource for ConvTransform
		segmSource := &segmConvAdapter{segm: segm}
		transMtx := conv.NewConvTransform(segmSource, mtx)

		// TransMtx adapter to satisfy conv.VertexSource for the second ConvTransform
		transMtxSource := &transConvAdapter{ct: transMtx}
		transLens := conv.NewConvTransform(transMtxSource, lens)

		ras.Reset()
		// Final adapter for the rasterizer
		adapter := &finalLensAdapter{ct: transLens}
		ras.AddPath(adapter, 0)
		
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		agg2d.NoLine()
		agg2d.DrawPath(agg.FillOnly)
	}

	// Render controls
	renderSlider(agg2d, lionLensMagnSlider)
	renderSlider(agg2d, lionLensRadiusSlider)
}

func getLionBoundingRect(paths []liondemo.Path) (x1, y1, x2, y2 float64) {
	x1, y1, x2, y2 = 1e100, 1e100, -1e100, -1e100
	first := true
	for _, lp := range paths {
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if basics.IsVertex(basics.PathCommand(cmd)) {
				if first {
					x1, y1, x2, y2 = x, y, x, y
					first = false
				} else {
					if x < x1 { x1 = x }
					if y < y1 { y1 = y }
					if x > x2 { x2 = x }
					if y > y2 { y2 = y }
				}
			}
		}
	}
	return
}

type pathConvAdapter struct {
	ps *liondemo.Path
}

func (a *pathConvAdapter) Rewind(pathID uint) {
	a.ps.Path.Rewind(pathID)
}

func (a *pathConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, vcmd := a.ps.Path.NextVertex()
	return vx, vy, basics.PathCommand(vcmd)
}

type segmConvAdapter struct {
	segm *conv.ConvSegmentator
}

func (a *segmConvAdapter) Rewind(pathID uint) {
	a.segm.Rewind(pathID)
}

func (a *segmConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, vcmd := a.segm.Vertex()
	return vx, vy, basics.PathCommand(vcmd)
}

type transConvAdapter struct {
	ct *conv.ConvTransform[*segmConvAdapter, *transform.TransAffine]
}

func (a *transConvAdapter) Rewind(pathID uint) {
	a.ct.Rewind(pathID)
}

func (a *transConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	return a.ct.Vertex()
}

type finalLensAdapter struct {
	ct *conv.ConvTransform[*transConvAdapter, *transform.TransWarpMagnifier]
}

func (a *finalLensAdapter) Rewind(pathID uint32) {
	a.ct.Rewind(uint(pathID))
}

func (a *finalLensAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, vcmd := a.ct.Vertex()
	*x = vx
	*y = vy
	return uint32(vcmd)
}

func handleLionLensMouseDown(x, y float64) bool {
	if !lionLensInitialized {
		return false
	}
	if lionLensMagnSlider.OnMouseButtonDown(x, y) { return true }
	if lionLensRadiusSlider.OnMouseButtonDown(x, y) { return true }
	
	lionLensX = x
	lionLensY = y
	return true
}

func handleLionLensMouseMove(x, y float64) bool {
	if !lionLensInitialized {
		return false
	}
	if lionLensMagnSlider.OnMouseMove(x, y, true) { return true }
	if lionLensRadiusSlider.OnMouseMove(x, y, true) { return true }
	
	lionLensX = x
	lionLensY = y
	return true
}

func handleLionLensMouseUp() {
	if !lionLensInitialized {
		return
	}
	lionLensMagnSlider.OnMouseButtonUp(0, 0)
	lionLensRadiusSlider.OnMouseButtonUp(0, 0)
}
