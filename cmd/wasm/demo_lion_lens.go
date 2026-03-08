// Based on the original AGG examples: lion_lens.cpp.
package main

import (
	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/conv"
	liondemo "agg_go/internal/demo/lion"
	"agg_go/internal/transform"
)

var (
	lionLensScale        = 3.0
	lionLensRadius       = 70.0
	lionLensX, lionLensY float64
	lionLensInitialized  bool

	// Reusable pipeline components to reduce allocations
	lionLensPipelines []lionLensPipeline
)

type lionLensPipeline struct {
	pathSource *pathConvAdapter
	segm       *conv.ConvSegmentator
	segmSource *segmConvAdapter
	transMtx   *conv.ConvTransform[*segmConvAdapter, *transform.TransAffine]
	transMtxS  *transConvAdapter
	transLens  *conv.ConvTransform[*transConvAdapter, *transform.TransWarpMagnifier]
	final      *finalLensAdapter
}

func initLionLensDemo() {
	if lionLensInitialized {
		return
	}

	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	lionLensX = float64(width) * 0.5
	lionLensY = float64(height) * 0.5

	// Initialize pipelines for each path
	lionLensPipelines = make([]lionLensPipeline, len(lionPaths))
	mtx := transform.NewTransAffine()         // Dummy
	lens := transform.NewTransWarpMagnifier() // Dummy

	for i := range lionPaths {
		p := &lionLensPipelines[i]
		p.pathSource = &pathConvAdapter{ps: &lionPaths[i]}
		p.segm = conv.NewConvSegmentator(p.pathSource)
		p.segm.ApproximationScale(1.0)
		p.segmSource = &segmConvAdapter{segm: p.segm}
		p.transMtx = conv.NewConvTransform(p.segmSource, mtx)
		p.transMtxS = &transConvAdapter{ct: p.transMtx}
		p.transLens = conv.NewConvTransform(p.transMtxS, lens)
		p.final = &finalLensAdapter{ct: p.transLens}
	}

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
	lens.SetMagnification(lionLensScale)
	lens.SetRadius(lionLensRadius / lionLensScale)

	// Set up the base transformation for the lion
	g_x1, g_y1, g_x2, g_y2 := getLionBoundingRect(lionPaths)
	base_dx := (g_x2 - g_x1) * 0.5
	base_dy := (g_y2 - g_y1) * 0.5

	mtx := transform.NewTransAffine()
	mtx.Translate(-base_dx, -base_dy)
	// Go has no flip_y rendering; ScaleXY(-1,1) mirrors X to reproduce
	// the same visual as C++ rotate(Pi) + flip_y=true.
	mtx.ScaleXY(-1, 1)
	mtx.Translate(float64(width)*0.5, float64(height)*0.5)

	for i := range lionPaths {
		lp := &lionPaths[i]
		pipe := &lionLensPipelines[i]

		// Update shared transformers in existing pipeline
		pipe.transMtx.SetTransformer(mtx)
		pipe.transLens.SetTransformer(lens)

		agg2d.ResetPath()
		pipe.final.Rewind(0)
		for {
			x, y, cmd := pipe.final.ct.Vertex() // Direct access
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsMoveTo(cmd) {
				agg2d.MoveTo(x, y)
			} else if basics.IsLineTo(cmd) {
				agg2d.LineTo(x, y)
			}
		}

		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		agg2d.NoLine()
		agg2d.DrawPath(agg.FillOnly)
	}
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

func setLionLensScale(v float64)  { lionLensScale = v }
func setLionLensRadius(v float64) { lionLensRadius = v }

func handleLionLensMouseDown(x, y float64) bool {
	lionLensX = x
	lionLensY = y
	return true
}

func handleLionLensMouseMove(x, y float64) bool {
	lionLensX = x
	lionLensY = y
	return true
}

func handleLionLensMouseUp() {}
