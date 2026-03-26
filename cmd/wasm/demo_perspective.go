// Based on the original AGG examples: perspective.cpp.
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
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	perspectiveQuad                                                            = [8]float64{100, 100, 500, 100, 500, 500, 100, 500}
	perspectiveSelectedNode                                                    = -1
	perspectiveType                                                            = 0 // 0: Bilinear, 1: Perspective
	perspectiveLionX1, perspectiveLionY1, perspectiveLionX2, perspectiveLionY2 float64
	perspectiveInitialized                                                     = false
)

// --- concrete types for the perspective demo ---

type (
	perspColorType    = color.RGBA8[color.Linear]
	perspPixFmtType   = *pixfmt.PixFmtRGBA32[color.Linear]
	perspRenBaseType  = *renderer.RendererBase[perspPixFmtType, perspColorType]
	perspRenSolidType = *renscan.RendererScanlineAASolid[perspRenBaseType, perspColorType]
	perspRasType      = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
)

func perspNewRas() *perspRasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
}

func perspClampU8(v float64) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 1:
		return 255
	default:
		return uint8(v*255.0 + 0.5)
	}
}

// --- ellipse conv adapter ---

type perspEllipseConvAdapter struct {
	ell *shapes.Ellipse
}

func (a *perspEllipseConvAdapter) Rewind(pathID uint) {
	a.ell.Rewind(uint32(pathID))
}

func (a *perspEllipseConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.ell.Vertex(&x, &y)
	return x, y, cmd
}

func initPerspectiveDemo() {
	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	if perspectiveInitialized {
		return
	}

	// Find bounding box of the lion
	x1, y1, x2, y2 := 1e9, 1e9, -1e9, -1e9
	for idx := uint(0); idx < lionData.Path.TotalVertices(); idx++ {
		x, y, cmd := lionData.Path.Vertex(idx)
		if !basics.IsVertex(basics.PathCommand(cmd)) {
			continue
		}
		if x < x1 {
			x1 = x
		}
		if x > x2 {
			x2 = x
		}
		if y < y1 {
			y1 = y
		}
		if y > y2 {
			y2 = y
		}
	}
	perspectiveLionX1, perspectiveLionY1, perspectiveLionX2, perspectiveLionY2 = x1, y1, x2, y2

	// Initialize quad to center the lion
	cx, cy := float64(width)/2, float64(height)/2
	w, h := (x2 - x1), (y2 - y1)
	perspectiveQuad[0], perspectiveQuad[1] = cx-w/2, cy-h/2
	perspectiveQuad[2], perspectiveQuad[3] = cx+w/2, cy-h/2
	perspectiveQuad[4], perspectiveQuad[5] = cx+w/2, cy+h/2
	perspectiveQuad[6], perspectiveQuad[7] = cx-w/2, cy+h/2

	perspectiveInitialized = true
}

// perspRenderScene renders the lion and bounding ellipse through the given transformer.
func perspRenderScene(
	ras *perspRasType,
	sl *scanline.ScanlineP8,
	rb perspRenBaseType,
	renSolid perspRenSolidType,
	tr transform.Transformer,
	srcX1, srcY1, srcX2, srcY2 float64,
) {
	// ----- Lion -----
	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, tr)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, lionData, lionData, lionData.NPaths)

	// ----- Bounding ellipse (filled) -----
	cx := (srcX1 + srcX2) * 0.5
	cy := (srcY1 + srcY2) * 0.5
	rx := (srcX2 - srcX1) * 0.5
	ry := (srcY2 - srcY1) * 0.5

	ell := shapes.NewEllipseWithParams(cx, cy, rx, ry, 200, false)
	ellAdapter := &perspEllipseConvAdapter{ell: ell}
	transEll := conv.NewConvTransform(ellAdapter, tr)
	rasEll := conv.NewRasterizerVertexSourceAdapter(transEll)
	ras.Reset()
	ras.AddPath(rasEll, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, perspColorType{
		R: perspClampU8(0.5), G: perspClampU8(0.3), B: perspClampU8(0.0), A: perspClampU8(0.3),
	})

	// ----- Bounding ellipse (stroke) -----
	ellAdapter2 := &perspEllipseConvAdapter{ell: ell}
	ellStroke := conv.NewConvStroke(ellAdapter2)
	ellStroke.SetWidth(3.0)
	transEllStroke := conv.NewConvTransform(ellStroke, tr)
	rasEllStroke := conv.NewRasterizerVertexSourceAdapter(transEllStroke)
	ras.Reset()
	ras.AddPath(rasEllStroke, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, perspColorType{
		R: perspClampU8(0.0), G: perspClampU8(0.3), B: perspClampU8(0.2), A: perspClampU8(1.0),
	})
}

// perspRenderSolid sweeps a rasterizer and paints a single solid color.
func perspRenderSolid(ras *perspRasType, sl *scanline.ScanlineP8, rb perspRenBaseType, c perspColorType) {
	renscan.RenderScanlinesAASolid(ras, sl, rb, c)
}

func drawPerspectiveDemo() {
	initPerspectiveDemo()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(perspColorType{R: 255, G: 255, B: 255, A: 255})

	ras := perspNewRas()
	ras.ClipBox(0, 0, float64(img.Width()), float64(img.Height()))
	sl := scanline.NewScanlineP8()
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rb)

	// The WASM canvas uses Y-down coordinates (Y=0 at top, matching the lion's
	// own data coordinate system).  Swap only the X corners to mirror the C++
	// flip_x behaviour while keeping Y correctly oriented.  Do NOT also swap the
	// Y corners as the standalone FlipY=true example does: that was needed to
	// compensate for the Y-up rendering buffer and would invert the lion here.
	switch perspectiveType {
	case 0: // Bilinear
		tr := transform.NewTransBilinearRectToQuad(
			perspectiveLionX2, perspectiveLionY1, perspectiveLionX1, perspectiveLionY2,
			perspectiveQuad,
		)
		if tr.IsValid() {
			perspRenderScene(ras, sl, rb, renSolid, tr,
				perspectiveLionX1, perspectiveLionY1, perspectiveLionX2, perspectiveLionY2)
		}
	case 1: // Perspective
		tr := transform.NewTransPerspectiveRectToQuad(
			perspectiveLionX2, perspectiveLionY1, perspectiveLionX1, perspectiveLionY2,
			perspectiveQuad,
		)
		if tr.IsValid(1e-14) {
			perspRenderScene(ras, sl, rb, renSolid, tr,
				perspectiveLionX1, perspectiveLionY1, perspectiveLionX2, perspectiveLionY2)
		}
	}

	// ----- Quad outline -----
	quadColor := perspColorType{
		R: perspClampU8(0.0), G: perspClampU8(0.3), B: perspClampU8(0.5), A: perspClampU8(0.6),
	}
	quadPath := path.NewPathStorageStl()
	quadPath.MoveTo(perspectiveQuad[0], perspectiveQuad[1])
	quadPath.LineTo(perspectiveQuad[2], perspectiveQuad[3])
	quadPath.LineTo(perspectiveQuad[4], perspectiveQuad[5])
	quadPath.LineTo(perspectiveQuad[6], perspectiveQuad[7])
	quadPath.ClosePolygon(0)
	quadSrc := &perspPathConvSource{ps: quadPath}
	quadStroke := conv.NewConvStroke(quadSrc)
	quadStroke.SetWidth(2.0)
	quadRasVS := conv.NewRasterizerVertexSourceAdapter(quadStroke)
	ras.Reset()
	ras.AddPath(quadRasVS, 0)
	perspRenderSolid(ras, sl, rb, quadColor)

	// ----- Handle circles -----
	handleFill := perspColorType{
		R: perspClampU8(0.8), G: perspClampU8(0.2), B: perspClampU8(0.1), A: perspClampU8(0.6),
	}
	handleOutline := perspColorType{R: 0, G: 0, B: 0, A: 255}
	for i := range 4 {
		hx := perspectiveQuad[i*2]
		hy := perspectiveQuad[i*2+1]

		// Filled circle
		ell := shapes.NewEllipseWithParams(hx, hy, 5, 5, 32, false)
		ellRas := &perspEllipseRasAdapter{ell: ell}
		ras.Reset()
		ras.AddPath(ellRas, 0)
		perspRenderSolid(ras, sl, rb, handleFill)

		// Outline circle
		ell2 := shapes.NewEllipseWithParams(hx, hy, 5, 5, 32, false)
		ellAdp2 := &perspEllipseConvAdapter{ell: ell2}
		ellStroke := conv.NewConvStroke(ellAdp2)
		ellStroke.SetWidth(1.0)
		ellStrokeRas := conv.NewRasterizerVertexSourceAdapter(ellStroke)
		ras.Reset()
		ras.AddPath(ellStrokeRas, 0)
		perspRenderSolid(ras, sl, rb, handleOutline)
	}

	applyLinearToSRGB(img)
}

// perspPathConvSource adapts PathStorageStl to the conv.VertexSource interface.
type perspPathConvSource struct {
	ps *path.PathStorageStl
}

func (a *perspPathConvSource) Rewind(pathID uint) { a.ps.Rewind(pathID) }

func (a *perspPathConvSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

// perspEllipseRasAdapter adapts shapes.Ellipse directly to the rasterizer vertex source.
type perspEllipseRasAdapter struct {
	ell *shapes.Ellipse
}

func (a *perspEllipseRasAdapter) Rewind(pathID uint32) { a.ell.Rewind(pathID) }

func (a *perspEllipseRasAdapter) Vertex(x, y *float64) uint32 {
	return uint32(a.ell.Vertex(x, y))
}

func handlePerspectiveMouseDown(x, y float64) bool {
	perspectiveSelectedNode = -1
	for i := 0; i < 4; i++ {
		dist := math.Sqrt((x-perspectiveQuad[i*2])*(x-perspectiveQuad[i*2]) + (y-perspectiveQuad[i*2+1])*(y-perspectiveQuad[i*2+1]))
		if dist < 10 {
			perspectiveSelectedNode = i
			return true
		}
	}
	return false
}

func handlePerspectiveMouseMove(x, y float64) bool {
	if perspectiveSelectedNode != -1 {
		perspectiveQuad[perspectiveSelectedNode*2] = x
		perspectiveQuad[perspectiveSelectedNode*2+1] = y
		return true
	}
	return false
}

func handlePerspectiveMouseUp() {
	perspectiveSelectedNode = -1
}

func setPerspectiveType(t int) {
	perspectiveType = t
}
