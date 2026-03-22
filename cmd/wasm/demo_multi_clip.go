package main

import (
	"fmt"
	"math"
	"math/rand"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

var multiClipN = 3.0

func drawMultiClipDemo() {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()

	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Fill background with white
	agg2d.ClearAll(agg.White)

	// Setup transformation for the lion
	baseDX, baseDY := 0.0, 0.0
	// Get bounding box for lion
	if lionData.NPaths > 0 {
		x1, y1, x2, y2 := 20.0, 20.0, 480.0, 380.0
		baseDX = (x2 - x1) * 0.5
		baseDY = (y2 - y1) * 0.5
	}

	agg2d.ResetTransformations()
	agg2d.Translate(-baseDX, -baseDY)
	agg2d.Scale(amLionScale, amLionScale)
	agg2d.Rotate(amLionAngle + basics.Pi)
	agg2d.Skew(amLionSkewX/1000.0, amLionSkewY/1000.0)
	agg2d.Translate(float64(w)/2, float64(h)/2)

	// Use RendererMClip
	tempRbuf := buffer.NewRenderingBufferWithData[uint8](ctx.GetImage().Data, w, h, w*4)
	imgPixf := pixfmt.NewPixFmtRGBA32[color.Linear](tempRbuf)

	// Use the generic renderer
	mclip := renderer.NewRendererMClip(imgPixf)

	mclip.ResetClipping(false) // Visibility: false means "no visible regions"
	n := multiClipN
	for x := 0.0; x < n; x++ {
		for y := 0.0; y < n; y++ {
			x1 := int(float64(w) * x / n)
			y1 := int(float64(h) * y / n)
			x2 := int(float64(w) * (x + 1) / n)
			y2 := int(float64(h) * (y + 1) / n)
			mclip.AddClipBox(x1+5, y1+5, x2-5, y2-5)
		}
	}

	// Render the lion with multi-clip
	ras := agg2d.GetInternalRasterizer()
	
	sl := scanline.NewScanlineU8()

	for i := 0; i < lionData.NPaths; i++ {
		c := color.RGBA8[color.Linear]{R: lionData.Colors[i].R, G: lionData.Colors[i].G, B: lionData.Colors[i].B, A: 255}

		ras.Reset()
		lionData.Path.Rewind(lionData.PathIdx[i])
		for {
			x, y, cmd := lionData.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			tx, ty := x, y
			agg2d.GetTransformations().Transform(tx, ty)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			}
		}

		renscan.RenderScanlinesAASolid(ras, sl, mclip, c)
	}

	// Random lines
	for i := 0; i < 50; i++ {
		x1, y1 := float64(rand.Intn(w)), float64(rand.Intn(h))
		x2, y2 := float64(rand.Intn(w)), float64(rand.Intn(h))
		c := color.RGBA8[color.Linear]{
			R: uint8(rand.Intn(128)),
			G: uint8(rand.Intn(128)),
			B: uint8(rand.Intn(128)),
			A: uint8(rand.Intn(128) + 127),
		}

		ras.Reset()
		ras.AddVertex(x1, y1, uint32(basics.PathCmdMoveTo))
		ras.AddVertex(x2, y2, uint32(basics.PathCmdLineTo))
		renscan.RenderScanlinesAASolid(ras, sl, mclip, c)
	}

	logStatus(fmt.Sprintf("Multi-Clip Demo: N=%.0f", n))
}

func setMultiClipN(n float64) {
	multiClipN = n
}

func handleMultiClipMouseDown(x, y float64) bool {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	dx := x - float64(w)/2
	dy := y - float64(h)/2
	amLionAngle = math.Atan2(dy, dx)
	amLionScale = math.Sqrt(dy*dy+dx*dx) / 100.0
	return true
}
