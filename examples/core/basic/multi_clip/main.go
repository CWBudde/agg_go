// Port of AGG C++ multi_clip.cpp – multi-clip region rendering.
//
// Renders the lion through a grid of N×N inset clip rectangles using
// RendererMClip. Default: N=3 (3×3 grid of clip boxes).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

const clipN = 3

// Scanline/rasterizer adapters.
type demo struct{}

func (d *demo) Render(img *agg.Image) {
	width := img.Width()
	height := img.Height()

	ctx := agg.NewContextForImage(img)
	agg2d := ctx.GetAgg2D()

	// White background.
	agg2d.ClearAll(agg.White)

	// Setup lion transform: centred, facing right.
	agg2d.ResetTransformations()
	agg2d.Translate(-250, -250)
	agg2d.Rotate(basics.Pi)
	agg2d.Translate(float64(width)/2, float64(height)/2)
	mtx := agg2d.GetTransformations()

	// Setup multi-clip renderer.
	mainBuf := buffer.NewRenderingBufferWithData[uint8](img.Data, width, height, width*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainBuf)
	mclip := renderer.NewRendererMClip(mainPixf)

	mclip.ResetClipping(false) // start with no visible regions
	n := clipN
	for xi := 0; xi < n; xi++ {
		for yi := 0; yi < n; yi++ {
			x1 := int(float64(width) * float64(xi) / float64(n))
			y1 := int(float64(height) * float64(yi) / float64(n))
			x2 := int(float64(width) * float64(xi+1) / float64(n))
			y2 := int(float64(height) * float64(yi+1) / float64(n))
			mclip.AddClipBox(x1+5, y1+5, x2-5, y2-5)
		}
	}

	ras := agg2d.GetInternalRasterizer()
	
	sl := scanline.NewScanlineU8()

	ld := liondemo.Parse()
	for i := 0; i < ld.NPaths; i++ {
		c := color.RGBA8[color.Linear]{R: ld.Colors[i].R, G: ld.Colors[i].G, B: ld.Colors[i].B, A: 255}
		ras.Reset()
		ld.Path.Rewind(ld.PathIdx[i])
		for {
			x, y, cmd := ld.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			tx, ty := mtx.Transform(x, y)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			}
		}
		renscan.RenderScanlinesAASolid(ras, sl, mclip, c)
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{Title: "Multi Clip", Width: 512, Height: 400}, &demo{})
}
