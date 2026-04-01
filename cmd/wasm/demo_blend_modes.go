// Blend mode gallery demo (separate from compositing.cpp direct port).
package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
)

func drawBlendModesDemo() {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrcOver)
	ren := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtCompositeRGBA32, color.RGBA8[color.Linear]](pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 229, G: 229, B: 229, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	modes := []struct {
		name string
		mode agg.BlendMode
	}{
		{"Alpha", agg.BlendAlpha},
		{"Multiply", agg.BlendMultiply},
		{"Screen", agg.BlendScreen},
		{"Overlay", agg.BlendOverlay},
		{"Darken", agg.BlendDarken},
		{"Lighten", agg.BlendLighten},
		{"ColorDodge", agg.BlendColorDodge},
		{"ColorBurn", agg.BlendColorBurn},
		{"HardLight", agg.BlendHardLight},
		{"SoftLight", agg.BlendSoftLight},
		{"Difference", agg.BlendDifference},
		{"Exclusion", agg.BlendExclusion},
	}

	const cols = 4
	rows := (len(modes) + cols - 1) / cols
	canvasW := float64(img.Width())
	canvasH := float64(img.Height())
	const (
		gapX = 8.0
		gapY = 8.0
	)
	cellW := (canvasW - float64(cols-1)*gapX) / float64(cols)
	cellH := (canvasH - float64(rows-1)*gapY) / float64(rows)
	const (
		r   = 40.0
		c1x = 70.0
		c1y = 60.0
		c2x = 110.0
		c2y = 60.0
		c3x = 90.0
		c3y = 100.0
	)
	// Center the three-circle cluster within each grid cell.
	groupMinX := min3(c1x-r, c2x-r, c3x-r)
	groupMaxX := max3(c1x+r, c2x+r, c3x+r)
	groupMinY := min3(c1y-r, c2y-r, c3y-r)
	groupMaxY := max3(c1y+r, c2y+r, c3y+r)
	shiftX := (cellW-(groupMaxX-groupMinX))*0.5 - groupMinX
	const labelBandH = 20.0
	drawAreaH := cellH - labelBandH
	shiftY := (drawAreaH-(groupMaxY-groupMinY))*0.5 - groupMinY

	_ = rows // used in cell layout

	for i, m := range modes {
		x := float64(i%cols) * (cellW + gapX)
		y := float64(i/cols) * (cellH + gapY)

		// Cell background with normal blending.
		pf.SetCompOp(blender.CompOpSrcOver)
		fillRect(ras, sl, ren,
			x+4, y+4, x+cellW-4, y+cellH-4,
			color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 153})

		// Red circle — normal blending.
		fillCircle(ras, sl, ren,
			x+shiftX+c1x, y+shiftY+c1y, r,
			color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 180})

		// Green circle — normal blending.
		fillCircle(ras, sl, ren,
			x+shiftX+c2x, y+shiftY+c2y, r,
			color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 180})

		// Blue circle — per-mode blending.
		pf.SetCompOp(blendModeToCompOpBM(m.mode))
		fillCircle(ras, sl, ren,
			x+shiftX+c3x, y+shiftY+c3y, r,
			color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 200})
	}

	// Restore default.
	pf.SetCompOp(blender.CompOpSrcOver)
}

// blendModeToCompOpBM maps the public agg.BlendMode to a blender.CompOp.
func blendModeToCompOpBM(mode agg.BlendMode) blender.CompOp {
	switch mode {
	case agg.BlendAlpha, agg.BlendSrcOver:
		return blender.CompOpSrcOver
	case agg.BlendClear:
		return blender.CompOpClear
	case agg.BlendSrc:
		return blender.CompOpSrc
	case agg.BlendDst:
		return blender.CompOpDst
	case agg.BlendDstOver:
		return blender.CompOpDstOver
	case agg.BlendSrcIn:
		return blender.CompOpSrcIn
	case agg.BlendDstIn:
		return blender.CompOpDstIn
	case agg.BlendSrcOut:
		return blender.CompOpSrcOut
	case agg.BlendDstOut:
		return blender.CompOpDstOut
	case agg.BlendSrcAtop:
		return blender.CompOpSrcAtop
	case agg.BlendDstAtop:
		return blender.CompOpDstAtop
	case agg.BlendXor:
		return blender.CompOpXor
	case agg.BlendAdd:
		return blender.CompOpPlus
	case agg.BlendMultiply:
		return blender.CompOpMultiply
	case agg.BlendScreen:
		return blender.CompOpScreen
	case agg.BlendOverlay:
		return blender.CompOpOverlay
	case agg.BlendDarken:
		return blender.CompOpDarken
	case agg.BlendLighten:
		return blender.CompOpLighten
	case agg.BlendColorDodge:
		return blender.CompOpColorDodge
	case agg.BlendColorBurn:
		return blender.CompOpColorBurn
	case agg.BlendHardLight:
		return blender.CompOpHardLight
	case agg.BlendSoftLight:
		return blender.CompOpSoftLight
	case agg.BlendDifference:
		return blender.CompOpDifference
	case agg.BlendExclusion:
		return blender.CompOpExclusion
	default:
		return blender.CompOpSrcOver
	}
}

// fillCircle rasterizes a filled circle at (cx,cy) with radius r.
func fillCircle(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	ren renscan.BaseRendererInterface[color.RGBA8[color.Linear]],
	cx, cy, radius float64,
	c color.RGBA8[color.Linear],
) {
	ell := shapes.NewEllipseWithParams(cx, cy, radius, radius, 0, false)
	ras.Reset()
	ell.Rewind(0)
	for {
		var x, y float64
		cmd := ell.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		ras.AddVertex(x, y, uint32(cmd))
	}
	renscan.RenderScanlinesAASolid(ras, sl, ren, c)
}

// fillRect rasterizes a filled axis-aligned rectangle.
func fillRect(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	ren renscan.BaseRendererInterface[color.RGBA8[color.Linear]],
	x1, y1, x2, y2 float64,
	c color.RGBA8[color.Linear],
) {
	ras.Reset()
	ras.AddVertex(x1, y1, uint32(basics.PathCmdMoveTo))
	ras.AddVertex(x2, y1, uint32(basics.PathCmdLineTo))
	ras.AddVertex(x2, y2, uint32(basics.PathCmdLineTo))
	ras.AddVertex(x1, y2, uint32(basics.PathCmdLineTo))
	renscan.RenderScanlinesAASolid(ras, sl, ren, c)
}

func min3(a, b, c float64) float64 {
	if a > b {
		a = b
	}
	if a > c {
		a = c
	}
	return a
}

func max3(a, b, c float64) float64 {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	return a
}
