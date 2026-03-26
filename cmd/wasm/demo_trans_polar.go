// Based on the original AGG examples: trans_polar.cpp.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

type transPolar struct {
	baseAngle      float64
	baseScale      float64
	baseX, baseY   float64
	transX, transY float64
	spiral         float64
}

func (p *transPolar) transform(x, y *float64) {
	x1 := (*x + p.baseX) * p.baseAngle
	y1 := (*y+p.baseY)*p.baseScale + (*x * p.spiral)
	*x = math.Cos(x1)*y1 + p.transX
	*y = math.Sin(x1)*y1 + p.transY
}

func (p *transPolar) Transform(x, y *float64) {
	p.transform(x, y)
}

var (
	polarBaseY  = 120.0
	polarSpiral = 0.0
)

func drawTransPolarDemo() {
	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pixFmt := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pixFmt)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// Setup polar transformer
	trans := &transPolar{
		baseAngle: 2.0 * math.Pi / 600.0, // spread 600 units over 2PI
		baseScale: 1.0,
		baseX:     0.0,
		baseY:     polarBaseY,
		transX:    float64(width) * 0.5,
		transY:    float64(height) * 0.5,
		spiral:    polarSpiral,
	}

	// Find bounding box of the lion
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
	// Scale lion to fit the "circle"
	scaleX := 600.0 / lionW

	for i := 0; i < lionData.NPaths; i++ {
		ps := path.NewPathStorageStl()

		lionData.Path.Rewind(lionData.PathIdx[i])

		for {
			x, y, cmd := lionData.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			// Normalize and scale lion
			tx := (x-lx1)*scaleX - 300.0 // Center it horizontally
			ty := y - (ly1+ly2)*0.5

			// Transform to polar
			trans.Transform(&tx, &ty)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ps.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ps.LineTo(tx, ty)
			}
		}
		ps.ClosePolygon(basics.PathFlagsNone)

		c := lionData.Colors[i]

		ras.Reset()
		ras.AddPath(&pathSourceAdapter{ps: ps}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, c)
	}

	applyLinearToSRGB(img)
}

func handleTransPolarMouseDown(x, y float64) bool {
	polarBaseY = y - float64(height)*0.5 + 120.0
	return true
}

func handleTransPolarMouseMove(x, y float64) bool {
	polarBaseY = y - float64(height)*0.5 + 120.0
	polarSpiral = (x - float64(width)*0.5) / 1000.0
	return true
}

func handleTransPolarMouseUp() {}
