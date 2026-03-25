// Package main ports AGG's perspective.cpp demo.
//
// Demonstrates bilinear and perspective transformation of the lion artwork
// using a draggable quad control.  The demo is an interactive parallel to
// the C++ example: select "Bilinear" or "Perspective" with the radio box,
// drag the quad handles to warp the lion, and observe the bounded ellipse
// overlay being transformed alongside the artwork.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	polygonctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
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

const (
	frameWidth  = 600
	frameHeight = 600
)

// --- type aliases for readability ---

type (
	rasType      = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	colorType    = color.RGBA8[color.Linear]
	pixFmtType   = *pixfmt.PixFmtRGBA32[color.Linear]
	renBaseType  = *renderer.RendererBase[pixFmtType, colorType]
	renSolidType = *renscan.RendererScanlineAASolid[renBaseType, colorType]
)

// --- ellipseConvAdapter bridges shapes.Ellipse to conv.VertexSource ---

type ellipseConvAdapter struct {
	ell *shapes.Ellipse
}

func (a *ellipseConvAdapter) Rewind(pathID uint) {
	a.ell.Rewind(uint32(pathID))
}

func (a *ellipseConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.ell.Vertex(&x, &y)
	return x, y, cmd
}

// --- ctrlRasAdapter bridges ctrl.Ctrl to rasterizer.VertexSource ---

type ctrlRasAdapter struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlRasAdapter) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }

func (a *ctrlRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// --- demo state ---

type demo struct {
	lion      liondemo.LionData
	srcX1     float64
	srcY1     float64
	srcX2     float64
	srcY2     float64
	quad      *polygonctrl.PolygonCtrl[color.RGBA]
	transType *rboxctrl.RboxCtrl[color.RGBA]
}

func newDemo() *demo {
	ld := liondemo.Parse()

	// Compute bounding rect of the lion path data.
	pathVS := path.NewPathStorageStlVertexSourceAdapter(ld.Path)
	bounds, ok := basics.BoundingRect[float64](pathVS, basics.SliceGetID(ld.PathIdx), 0, uint(ld.NPaths))
	if !ok {
		panic("perspective: cannot compute lion bounding rect")
	}
	x1, y1, x2, y2 := bounds.X1, bounds.Y1, bounds.X2, bounds.Y2

	// Centre the initial quad on the 600×600 window (mirrors C++ on_init).
	dx := float64(frameWidth)/2.0 - (x2-x1)/2.0
	dy := float64(frameHeight)/2.0 - (y2-y1)/2.0

	quad := polygonctrl.NewDefaultPolygonCtrl(4, 5.0)
	quad.SetClose(true)
	quad.SetInPolygonCheck(true)
	quad.SetXn(0, x1+dx)
	quad.SetYn(0, y1+dy)
	quad.SetXn(1, x2+dx)
	quad.SetYn(1, y1+dy)
	quad.SetXn(2, x2+dx)
	quad.SetYn(2, y2+dy)
	quad.SetXn(3, x1+dx)
	quad.SetYn(3, y2+dy)

	// Radio-box control at top-right – matches C++ position.
	// flipY=false because render coords with FlipY=true already have y=0 at the
	// bottom, so y=5 is near the bottom of the buffer which the runner displays
	// near the bottom of the window (matching the C++ flip_y=true behaviour).
	transType := rboxctrl.NewDefaultRboxCtrl(420, 5.0, 420+130.0, 55.0, false)
	transType.AddItem("Bilinear")
	transType.AddItem("Perspective")
	transType.SetCurItem(0)

	return &demo{
		lion:      ld,
		srcX1:     x1,
		srcY1:     y1,
		srcX2:     x2,
		srcY2:     y2,
		quad:      quad,
		transType: transType,
	}
}

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

func clampU8(v float64) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 1:
		return 255
	default:
		return uint8(v*255.0 + 0.5)
	}
}

// renderCtrl renders all paths of a control widget in their respective colours.
func renderCtrl(ras *rasType, sl *scanline.ScanlineP8, rb renBaseType, ctrl ctrlbase.Ctrl[color.RGBA]) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlRasAdapter{ctrl: ctrl}, uint32(pathID))
		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, rb, colorType{
			R: clampU8(c.R), G: clampU8(c.G), B: clampU8(c.B), A: clampU8(c.A),
		})
	}
}

// quadAsFloat8 returns the four quad corners as a flat [8]float64 array
// expected by the transform constructors.
func (d *demo) quadAsFloat8() [8]float64 {
	return [8]float64{
		d.quad.Xn(0), d.quad.Yn(0),
		d.quad.Xn(1), d.quad.Yn(1),
		d.quad.Xn(2), d.quad.Yn(2),
		d.quad.Xn(3), d.quad.Yn(3),
	}
}

// renderScene renders the lion, filled ellipse and ellipse stroke through the
// supplied transformer (bilinear or perspective).
func renderScene(
	ras *rasType,
	sl *scanline.ScanlineP8,
	rb renBaseType,
	renSolid renSolidType,
	lion *liondemo.LionData,
	srcX1, srcY1, srcX2, srcY2 float64,
	tr transform.Transformer,
) {
	// ----- Lion -----
	pathVS := path.NewPathStorageStlVertexSourceAdapter(lion.Path)
	transVS := conv.NewConvTransform(pathVS, tr)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, lion, lion, lion.NPaths)

	// ----- Bounding ellipse (filled) -----
	cx := (srcX1 + srcX2) * 0.5
	cy := (srcY1 + srcY2) * 0.5
	rx := (srcX2 - srcX1) * 0.5
	ry := (srcY2 - srcY1) * 0.5

	ell := shapes.NewEllipseWithParams(cx, cy, rx, ry, 200, false)
	ellAdapter := &ellipseConvAdapter{ell: ell}

	transEll := conv.NewConvTransform(ellAdapter, tr)
	rasEll := conv.NewRasterizerVertexSourceAdapter(transEll)
	ras.Reset()
	ras.AddPath(rasEll, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, colorType{
		R: clampU8(0.5), G: clampU8(0.3), B: clampU8(0.0), A: clampU8(0.3),
	})

	// ----- Bounding ellipse (stroke) -----
	ellAdapter2 := &ellipseConvAdapter{ell: ell}
	ellStroke := conv.NewConvStroke(ellAdapter2)
	ellStroke.SetWidth(3.0)
	transEllStroke := conv.NewConvTransform(ellStroke, tr)
	rasEllStroke := conv.NewRasterizerVertexSourceAdapter(transEllStroke)
	ras.Reset()
	ras.AddPath(rasEllStroke, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, colorType{
		R: clampU8(0.0), G: clampU8(0.3), B: clampU8(0.2), A: clampU8(1.0),
	})
}

func (d *demo) Render(img *agg.Image) {
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(colorType{R: 255, G: 255, B: 255, A: 255})

	ras := newRasterizer()
	ras.ClipBox(0, 0, float64(img.Width()), float64(img.Height()))
	sl := scanline.NewScanlineP8()
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rb)

	quad := d.quadAsFloat8()

	// Mirror the C++ parse_lion() flip_x+flip_y by swapping the source-rect
	// corners: passing (x2,y2,x1,y1) instead of (x1,y1,x2,y2) makes the
	// transform pre-apply both axis flips so the lion is right-side-up in
	// the FlipY=true render buffer, exactly matching the C++ behaviour.
	switch d.transType.CurItem() {
	case 0: // Bilinear
		tr := transform.NewTransBilinearRectToQuad(d.srcX2, d.srcY2, d.srcX1, d.srcY1, quad)
		if tr.IsValid() {
			renderScene(ras, sl, rb, renSolid, &d.lion, d.srcX1, d.srcY1, d.srcX2, d.srcY2, tr)
		}
	case 1: // Perspective
		tr := transform.NewTransPerspectiveRectToQuad(d.srcX2, d.srcY2, d.srcX1, d.srcY1, quad)
		if tr.IsValid(1e-14) {
			renderScene(ras, sl, rb, renSolid, &d.lion, d.srcX1, d.srcY1, d.srcX2, d.srcY2, tr)
		}
	}

	// ----- Quad control (filled polygon outline + handles) -----
	ras.Reset()
	ras.AddPath(&ctrlRasAdapter{ctrl: d.quad}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, colorType{
		R: clampU8(0.0), G: clampU8(0.3), B: clampU8(0.5), A: clampU8(0.6),
	})

	// ----- Radio-box control -----
	renderCtrl(ras, sl, rb, d.transType)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	if d.transType.OnMouseButtonDown(fx, fy) {
		return true
	}
	return d.quad.OnMouseButtonDown(fx, fy)
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := d.transType.OnMouseMove(fx, fy, btn.Left)
	if d.quad.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, _ lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := d.transType.OnMouseButtonUp(fx, fy)
	if d.quad.OnMouseButtonUp(fx, fy) {
		redraw = true
	}
	return redraw
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "AGG Example. Perspective Transformations",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
