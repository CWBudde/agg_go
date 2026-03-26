// Based on the original AGG examples: trans_curve2.cpp.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	transCurve2Points1  = [12]float64{60, 40, 170, 130, 230, 270, 370, 330, 430, 470, 550, 550}
	transCurve2Points2  = [12]float64{40, 60, 150, 170, 210, 290, 350, 350, 410, 490, 530, 570}
	transCurve2Selected = -1
	transCurve2Animate  = false
	transCurve2DX1      [6]float64
	transCurve2DY1      [6]float64
	transCurve2DX2      [6]float64
	transCurve2DY2      [6]float64
)

const (
	transCurve2RefW = 600.0
	transCurve2RefH = 600.0
)

func transCurve2FrameOffset() (float64, float64) {
	return (float64(width) - transCurve2RefW) * 0.5, (float64(height) - transCurve2RefH) * 0.5
}

func initTransCurve2Demo() {
	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}
	for i := 0; i < 6; i++ {
		transCurve2DX1[i] = (math.Mod(float64(i*1234+1), 10.0) - 5.0) * 0.5
		transCurve2DY1[i] = (math.Mod(float64(i*5678+2), 10.0) - 5.0) * 0.5
		transCurve2DX2[i] = (math.Mod(float64(i*1234+3), 10.0) - 5.0) * 0.5
		transCurve2DY2[i] = (math.Mod(float64(i*5678+4), 10.0) - 5.0) * 0.5
	}
}

type transDoubleAdapter struct {
	source *conv.ConvBSpline
}

func (a *transDoubleAdapter) Rewind(id uint) { a.source.Rewind(id) }
func (a *transDoubleAdapter) Vertex() (float64, float64, basics.PathCommand) {
	return a.source.Vertex()
}

func drawTransCurve2Demo() {
	initTransCurve2Demo()
	offX, offY := transCurve2FrameOffset()

	if transCurve2Animate {
		for i := 0; i < 6; i++ {
			moveTransCurve2Point(&transCurve2Points1[i*2], &transCurve2Points1[i*2+1], &transCurve2DX1[i], &transCurve2DY1[i])
			moveTransCurve2Point(&transCurve2Points2[i*2], &transCurve2Points2[i*2+1], &transCurve2DX2[i], &transCurve2DY2[i])
			// normalize distance
			d := math.Sqrt((transCurve2Points1[i*2]-transCurve2Points2[i*2])*(transCurve2Points1[i*2]-transCurve2Points2[i*2]) + (transCurve2Points1[i*2+1]-transCurve2Points2[i*2+1])*(transCurve2Points1[i*2+1]-transCurve2Points2[i*2+1]))
			if d > 100 {
				transCurve2Points2[i*2] = transCurve2Points1[i*2] + (transCurve2Points2[i*2]-transCurve2Points1[i*2])*100/d
				transCurve2Points2[i*2+1] = transCurve2Points1[i*2+1] + (transCurve2Points2[i*2+1]-transCurve2Points1[i*2+1])*100/d
			}
		}
	}

	// --- Low-level rendering setup ---
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	addPathToRas := func(src conv.VertexSource) {
		src.Rewind(0)
		for {
			vx, vy, cmd := src.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(vx, vy, uint32(cmd))
		}
	}

	renderSolid := func(c color.RGBA8[color.Linear]) {
		if ras.RewindScanlines() {
			sl.Reset(ras.MinX(), ras.MaxX())
			for ras.SweepScanline(sl) {
				y := sl.Y()
				for _, span := range sl.Spans() {
					if span.Len > 0 {
						ren.BlendSolidHspan(int(span.X), y, int(span.Len), c, span.Covers)
					}
				}
			}
		}
	}

	// 1. Create guide paths
	ps1 := path.NewPathStorageStl()
	ps2 := path.NewPathStorageStl()
	ps1.MoveTo(transCurve2Points1[0], transCurve2Points1[1])
	ps2.MoveTo(transCurve2Points2[0], transCurve2Points2[1])
	for i := 1; i < 6; i++ {
		ps1.LineTo(transCurve2Points1[i*2], transCurve2Points1[i*2+1])
		ps2.LineTo(transCurve2Points2[i*2], transCurve2Points2[i*2+1])
	}

	bs1 := conv.NewConvBSpline(path.NewPathStorageStlVertexSourceAdapter(ps1))
	bs2 := conv.NewConvBSpline(path.NewPathStorageStlVertexSourceAdapter(ps2))
	bs1.SetInterpolationStep(1.0 / 40.0)
	bs2.SetInterpolationStep(1.0 / 40.0)

	// 2. Setup transformation
	tcurve := transform.NewTransDoublePath()
	tcurve.AddPaths(&transDoubleAdapter{bs1}, &transDoubleAdapter{bs2}, 0, 0)
	tcurve.SetBaseHeight(40.0)

	// 3. Find lion bounding box
	lx1, ly1, lx2, ly2 := 1e9, 1e9, -1e9, -1e9
	for idx := uint(0); idx < lionData.Path.TotalVertices(); idx++ {
		x, y, cmd := lionData.Path.Vertex(idx)
		if !basics.IsVertex(basics.PathCommand(cmd)) {
			continue
		}
		if x < lx1 {
			lx1 = x
		}
		if x > lx2 {
			lx2 = x
		}
		if y < ly1 {
			ly1 = y
		}
		if y > ly2 {
			ly2 = y
		}
	}
	lionW := lx2 - lx1
	scaleX := tcurve.TotalLength1() / lionW * 0.8

	// 4. Render transformed lion paths
	for i := 0; i < lionData.NPaths; i++ {
		c := lionData.Colors[i]
		fillColor := color.RGBA8[color.Linear]{R: c.R, G: c.G, B: c.B, A: 200}

		lionData.Path.Rewind(lionData.PathIdx[i])
		ras.Reset()
		for {
			x, y, cmd := lionData.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			tx := (x - lx1) * scaleX
			ty := (y - (ly1+ly2)*0.5)
			tcurve.Transform(&tx, &ty)
			tx += offX
			ty += offY
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			}
		}
		ras.ClosePolygon()
		renderSolid(fillColor)
	}

	// 5. Draw guide curves as stroked paths
	guideColor := color.RGBA8[color.Linear]{R: 170, G: 50, B: 20, A: 100}

	for _, bs := range []*conv.ConvBSpline{bs1, bs2} {
		// Build an offset spline into a PathStorageStl, then stroke it
		psGuide := path.NewPathStorageStl()
		bs.Rewind(0)
		first := true
		for {
			vx, vy, cmd := bs.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if first {
				psGuide.MoveTo(vx+offX, vy+offY)
				first = false
			} else {
				psGuide.LineTo(vx+offX, vy+offY)
			}
		}
		guideAdapter := path.NewPathStorageStlVertexSourceAdapter(psGuide)
		guideStroke := conv.NewConvStroke(guideAdapter)
		guideStroke.SetWidth(1.0)
		ras.Reset()
		addPathToRas(guideStroke)
		renderSolid(guideColor)
	}

	// 6. Draw handles
	for i := 0; i < 6; i++ {
		drawHandle(transCurve2Points1[i*2]+offX, transCurve2Points1[i*2+1]+offY)
		drawHandle(transCurve2Points2[i*2]+offX, transCurve2Points2[i*2+1]+offY)
	}
}

func moveTransCurve2Point(x, y, dx, dy *float64) {
	*x += *dx
	*y += *dy
	if *x < 0 || *x > transCurve2RefW {
		*dx = -*dx
	}
	if *y < 0 || *y > transCurve2RefH {
		*dy = -*dy
	}
}

func handleTransCurve2MouseDown(x, y float64) bool {
	offX, offY := transCurve2FrameOffset()
	x -= offX
	y -= offY
	transCurve2Selected = -1
	for i := 0; i < 6; i++ {
		if math.Sqrt((x-transCurve2Points1[i*2])*(x-transCurve2Points1[i*2])+(y-transCurve2Points1[i*2+1])*(y-transCurve2Points1[i*2+1])) < 15 {
			transCurve2Selected = i
			return true
		}
		if math.Sqrt((x-transCurve2Points2[i*2])*(x-transCurve2Points2[i*2])+(y-transCurve2Points2[i*2+1])*(y-transCurve2Points2[i*2+1])) < 15 {
			transCurve2Selected = i + 6
			return true
		}
	}
	return false
}

func handleTransCurve2MouseMove(x, y float64) bool {
	offX, offY := transCurve2FrameOffset()
	x -= offX
	y -= offY
	if transCurve2Selected != -1 {
		if transCurve2Selected < 6 {
			transCurve2Points1[transCurve2Selected*2] = x
			transCurve2Points1[transCurve2Selected*2+1] = y
		} else {
			idx := transCurve2Selected - 6
			transCurve2Points2[idx*2] = x
			transCurve2Points2[idx*2+1] = y
		}
		return true
	}
	return false
}

func handleTransCurve2MouseUp() {
	transCurve2Selected = -1
}

func toggleTransCurve2Animate() {
	transCurve2Animate = !transCurve2Animate
}
