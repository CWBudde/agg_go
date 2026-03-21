// Port of AGG C++ alpha_mask3.cpp – alpha-mask as polygon clipper.
//
// Implements all five polygon/operation combos selectable via radio buttons,
// matching the C++ original layout: polygons rbox bottom-left, operation rbox
// bottom-right, and timing text in the lower-centre.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	"github.com/MeKo-Christian/agg_go/internal/demo/aggshapes"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	frameWidth  = 640
	frameHeight = 520
)

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters
// ---------------------------------------------------------------------------
type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// conv.VertexSource → rasterizer.VertexSource adapter
// ---------------------------------------------------------------------------

type rasterVS struct{ src conv.VertexSource }

func (a *rasterVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *rasterVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// transformed path: PathStorageStl + affine matrix as conv.VertexSource
// ---------------------------------------------------------------------------

type transformedPathVS struct {
	ps  *path.PathStorageStl
	mtx *transform.TransAffine
}

func (t *transformedPathVS) Rewind(id uint) { t.ps.Rewind(id) }
func (t *transformedPathVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := t.ps.NextVertex()
	t.mtx.Transform(&x, &y)
	return x, y, basics.PathCommand(cmd)
}

// ---------------------------------------------------------------------------
// Spiral vertex source – matches C++ spiral class from alpha_mask3.cpp
// ---------------------------------------------------------------------------

type spiral struct {
	x, y         float64
	r1, r2       float64
	step         float64
	startAngle   float64
	angle, currR float64
	da, dr       float64
	start        bool
}

func newSpiral(x, y, r1, r2, step, startAngle float64) *spiral {
	return &spiral{
		x:          x,
		y:          y,
		r1:         r1,
		r2:         r2,
		step:       step,
		startAngle: startAngle,
		da:         4.0 * basics.Deg2Rad,
		dr:         step / 90.0,
	}
}

func (s *spiral) Rewind(_ uint) {
	s.angle = s.startAngle
	s.currR = s.r1
	s.start = true
}

func (s *spiral) Vertex() (float64, float64, basics.PathCommand) {
	if s.currR > s.r2 {
		return 0, 0, basics.PathCmdStop
	}
	x := s.x + math.Cos(s.angle)*s.currR
	y := s.y + math.Sin(s.angle)*s.currR
	s.currR += s.dr
	s.angle += s.da
	if s.start {
		s.start = false
		return x, y, basics.PathCmdMoveTo
	}
	return x, y, basics.PathCmdLineTo
}

// ---------------------------------------------------------------------------
// Control rendering helper
// ---------------------------------------------------------------------------

type ctrlInterface interface {
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) color.RGBA
}

type ctrlRasAdapter struct {
	ctrl   ctrlInterface
	pathID uint
}

func (a *ctrlRasAdapter) Rewind(id uint32) {
	a.pathID = uint(id)
	a.ctrl.Rewind(uint(id))
}

func (a *ctrlRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func rgbaToRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA8[color.Linear]{
		R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A),
	}
}

func renderCtrl[Rb renscan.BaseRendererInterface[color.RGBA8[color.Linear]]](
	ras *rasType,
	sl *scanline.ScanlineP8,
	rb Rb,
	ctrl ctrlInterface,
) {
	adapter := &ctrlRasAdapter{ctrl: ctrl}
	for i := uint(0); i < ctrl.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(i))
		renscan.RenderScanlinesAASolid(ras, sl, rb, rgbaToRGBA8(ctrl.Color(i)))
	}
}

// ---------------------------------------------------------------------------
// Text rendering helper (gsv_text + conv_stroke, like C++ draw_text)
// ---------------------------------------------------------------------------

func drawText(
	ras *rasType,
	sl *scanline.ScanlineP8,
	rb renscan.BaseRendererInterface[color.RGBA8[color.Linear]],
	x, y float64,
	str string,
) {
	txt := gsv.NewGSVText()
	txtStroke := conv.NewConvStroke(txt)
	txtStroke.SetWidth(1.5)
	txtStroke.SetLineCap(basics.RoundCap)
	txt.SetSize(10.0, 0)
	txt.SetStartPoint(x, y)
	txt.SetText(str)

	ras.Reset()
	ras.AddPath(&rasterVS{src: txtStroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})
}

// ---------------------------------------------------------------------------
// Alpha-mask generation and masked rendering helpers
// ---------------------------------------------------------------------------

func generateAlphaMask(
	ras *rasType,
	sl *scanline.ScanlineP8,
	vs conv.VertexSource,
	opAND bool,
	w, h int,
) (*pixfmt.AMaskNoClipU8, *buffer.RenderingBufferU8) {
	maskData := make([]uint8, w*h)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, w, h, w)
	maskPixf := pixfmt.NewPixFmtSGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)

	var clearColor, fillColor color.Gray8[color.SRGB]
	if opAND {
		clearColor = color.Gray8[color.SRGB]{V: 0, A: 255}
		fillColor = color.Gray8[color.SRGB]{V: 255, A: 255}
	} else {
		clearColor = color.Gray8[color.SRGB]{V: 255, A: 255}
		fillColor = color.Gray8[color.SRGB]{V: 0, A: 255}
	}
	maskRb.Clear(clearColor)
	ras.Reset()
	ras.AddPath(&rasterVS{src: vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, maskRb, fillColor)

	mask := pixfmt.NewAMaskNoClipU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	return mask, maskBuf
}

func performRendering(
	ras *rasType,
	sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mask *pixfmt.AMaskNoClipU8,
	vs conv.VertexSource,
) {
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(mainPixf, mask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)
	ras.Reset()
	ras.AddPath(&rasterVS{src: vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rbAMask,
		color.RGBA8[color.Linear]{R: 127, G: 0, B: 0, A: 127})
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct {
	mx, my float64

	polygons  *rbox.RboxCtrl[color.RGBA]
	operation *rbox.RboxCtrl[color.RGBA]
}

func newDemo() *demo {
	d := &demo{
		mx: float64(frameWidth) / 2,
		my: float64(frameHeight) / 2,
	}

	// C++: m_polygons(5.0, 5.0, 5.0+205.0, 110.0, !flip_y)  flip_y=true → !flip_y=false
	d.polygons = rbox.NewDefaultRboxCtrl(5, 5, 210, 110, false)
	_ = d.polygons.AddItem("Two Simple Paths")
	_ = d.polygons.AddItem("Closed Stroke")
	_ = d.polygons.AddItem("Great Britain and Arrows")
	_ = d.polygons.AddItem("Great Britain and Spiral")
	_ = d.polygons.AddItem("Spiral and Glyph")
	d.polygons.SetCurItem(3)

	// C++: m_operation(555.0, 5.0, 555.0+80.0, 55.0, !flip_y)
	d.operation = rbox.NewDefaultRboxCtrl(555, 5, 635, 55, false)
	_ = d.operation.AddItem("SUB")
	_ = d.operation.AddItem("AND")
	d.operation.SetCurItem(1) // default AND

	return d
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineP8()

	opAND := d.operation.CurItem() == 1

	switch d.polygons.CurItem() {
	case 0:
		d.renderTwoSimplePaths(ras, sl, mainPixf, mainRb, opAND)
	case 1:
		d.renderClosedStroke(ras, sl, mainPixf, mainRb, opAND)
	case 2:
		d.renderGBAndArrows(ras, sl, mainPixf, mainRb, opAND)
	case 3:
		d.renderGBAndSpiral(ras, sl, mainPixf, mainRb, opAND)
	case 4:
		d.renderSpiralAndGlyph(ras, sl, mainPixf, mainRb, opAND)
	}

	// Timing text (static labels, no actual timing)
	// C++ draw_text(250, 20, ...) and draw_text(250, 5, ...) in flipped coords.
	drawText(ras, sl, mainRb, 250, 20, "Generate AlphaMask: -")
	drawText(ras, sl, mainRb, 250, 5, "Render with AlphaMask: -")

	// Control widgets
	renderCtrl(ras, sl, mainRb, d.polygons)
	renderCtrl(ras, sl, mainRb, d.operation)

	copyFlipY(workBuf, img.Data, w, h)
}

// ---------------------------------------------------------------------------
// Case 0: Two simple paths
// ---------------------------------------------------------------------------

func (d *demo) renderTwoSimplePaths(
	ras *rasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool,
) {
	x := d.mx - float64(frameWidth)/2 + 100
	y := d.my - float64(frameHeight)/2 + 100

	ps1 := path.NewPathStorageStl()
	ps1.MoveTo(x+140, y+145)
	ps1.LineTo(x+225, y+44)
	ps1.LineTo(x+296, y+219)
	ps1.ClosePolygon(basics.PathFlagsNone)
	ps1.LineTo(x+226, y+289)
	ps1.LineTo(x+82, y+292)
	ps1.MoveTo(x+220, y+222)
	ps1.LineTo(x+363, y+249)
	ps1.LineTo(x+265, y+331)
	ps1.MoveTo(x+242, y+243)
	ps1.LineTo(x+268, y+309)
	ps1.LineTo(x+325, y+261)
	ps1.MoveTo(x+259, y+259)
	ps1.LineTo(x+273, y+288)
	ps1.LineTo(x+298, y+266)

	ps2 := path.NewPathStorageStl()
	ps2.MoveTo(100+32, 100+77)
	ps2.LineTo(100+473, 100+263)
	ps2.LineTo(100+351, 100+290)
	ps2.LineTo(100+354, 100+374)

	ps1vs := &pathStorageVS{ps: ps1}
	ps2vs := &pathStorageVS{ps: ps2}

	ras.Reset()
	ras.AddPath(&rasterVS{src: ps1vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 25})

	ras.Reset()
	ras.AddPath(&rasterVS{src: ps2vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 25})

	mask, _ := generateAlphaMask(ras, sl, ps1vs, opAND, frameWidth, frameHeight)
	performRendering(ras, sl, mainPixf, mask, ps2vs)
}

// ---------------------------------------------------------------------------
// Case 1: Closed stroke
// ---------------------------------------------------------------------------

func (d *demo) renderClosedStroke(
	ras *rasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool,
) {
	x := d.mx - float64(frameWidth)/2 + 100
	y := d.my - float64(frameHeight)/2 + 100

	ps1 := path.NewPathStorageStl()
	ps1.MoveTo(x+140, y+145)
	ps1.LineTo(x+225, y+44)
	ps1.LineTo(x+296, y+219)
	ps1.ClosePolygon(basics.PathFlagsNone)
	ps1.LineTo(x+226, y+289)
	ps1.LineTo(x+82, y+292)
	ps1.MoveTo(x+220-50, y+222)
	ps1.LineTo(x+265-50, y+331)
	ps1.LineTo(x+363-50, y+249)
	ps1.ClosePolygon(basics.PathFlagsCCW)

	ps2 := path.NewPathStorageStl()
	ps2.MoveTo(100+32, 100+77)
	ps2.LineTo(100+473, 100+263)
	ps2.LineTo(100+351, 100+290)
	ps2.LineTo(100+354, 100+374)
	ps2.ClosePolygon(basics.PathFlagsNone)

	ps1vs := &pathStorageVS{ps: ps1}
	ps2vs := &pathStorageVS{ps: ps2}
	stroke := conv.NewConvStroke(ps2vs)
	stroke.SetWidth(10.0)

	ras.Reset()
	ras.AddPath(&rasterVS{src: ps1vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 25})

	ras.Reset()
	ras.AddPath(&rasterVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 25})

	mask, _ := generateAlphaMask(ras, sl, ps1vs, opAND, frameWidth, frameHeight)
	performRendering(ras, sl, mainPixf, mask, stroke)
}

// ---------------------------------------------------------------------------
// Case 2: Great Britain and Arrows
// ---------------------------------------------------------------------------

func (d *demo) renderGBAndArrows(
	ras *rasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool,
) {
	gbPoly := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(gbPoly)
	arrows := path.NewPathStorageStl()
	aggshapes.MakeArrows(arrows)

	mtx1 := transform.NewTransAffine()
	mtx1.Translate(-1150, -1150)
	mtx1.Scale(2.0)

	mtx2 := *mtx1
	mtx2.Translate(d.mx-float64(frameWidth)/2, d.my-float64(frameHeight)/2)

	transGB := &transformedPathVS{ps: gbPoly, mtx: mtx1}
	transArrows := &transformedPathVS{ps: arrows, mtx: &mtx2}

	ras.Reset()
	ras.AddPath(&rasterVS{src: transGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 127, G: 127, B: 0, A: 25})

	strokeGB := conv.NewConvStroke(transGB)
	strokeGB.SetWidth(0.1)
	ras.Reset()
	ras.AddPath(&rasterVS{src: strokeGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras.Reset()
	ras.AddPath(&rasterVS{src: transArrows}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 127, B: 127, A: 25})

	mask, _ := generateAlphaMask(ras, sl, transGB, opAND, frameWidth, frameHeight)
	performRendering(ras, sl, mainPixf, mask, transArrows)
}

// ---------------------------------------------------------------------------
// Case 3: Great Britain and Spiral
// ---------------------------------------------------------------------------

func (d *demo) renderGBAndSpiral(
	ras *rasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool,
) {
	sp := newSpiral(d.mx, d.my, 10, 150, 30, 0.0)
	stroke := conv.NewConvStroke(sp)
	stroke.SetWidth(15.0)

	gbPoly := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(gbPoly)

	mtx := transform.NewTransAffine()
	mtx.Translate(-1150, -1150)
	mtx.Scale(2.0)

	transGB := &transformedPathVS{ps: gbPoly, mtx: mtx}

	ras.Reset()
	ras.AddPath(&rasterVS{src: transGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 127, G: 127, B: 0, A: 25})

	strokeGB := conv.NewConvStroke(transGB)
	strokeGB.SetWidth(0.1)
	ras.Reset()
	ras.AddPath(&rasterVS{src: strokeGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras.Reset()
	ras.AddPath(&rasterVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 127, B: 127, A: 25})

	mask, _ := generateAlphaMask(ras, sl, transGB, opAND, frameWidth, frameHeight)
	performRendering(ras, sl, mainPixf, mask, stroke)
}

// ---------------------------------------------------------------------------
// Case 4: Spiral and Glyph
// ---------------------------------------------------------------------------

func (d *demo) renderSpiralAndGlyph(
	ras *rasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool,
) {
	sp := newSpiral(d.mx, d.my, 10, 150, 30, 0.0)
	stroke := conv.NewConvStroke(sp)
	stroke.SetWidth(15.0)

	glyph := path.NewPathStorageStl()
	glyph.MoveTo(28.47, 6.45)
	glyph.Curve3(21.58, 1.12, 19.82, 0.29)
	glyph.Curve3(17.19, -0.93, 14.21, -0.93)
	glyph.Curve3(9.57, -0.93, 6.57, 2.25)
	glyph.Curve3(3.56, 5.42, 3.56, 10.60)
	glyph.Curve3(3.56, 13.87, 5.03, 16.26)
	glyph.Curve3(7.03, 19.58, 11.99, 22.51)
	glyph.Curve3(16.94, 25.44, 28.47, 29.64)
	glyph.LineTo(28.47, 31.40)
	glyph.Curve3(28.47, 38.09, 26.34, 40.58)
	glyph.Curve3(24.22, 43.07, 20.17, 43.07)
	glyph.Curve3(17.09, 43.07, 15.28, 41.41)
	glyph.Curve3(13.43, 39.75, 13.43, 37.60)
	glyph.LineTo(13.53, 34.77)
	glyph.Curve3(13.53, 32.52, 12.38, 31.30)
	glyph.Curve3(11.23, 30.08, 9.38, 30.08)
	glyph.Curve3(7.57, 30.08, 6.42, 31.35)
	glyph.Curve3(5.27, 32.62, 5.27, 34.81)
	glyph.Curve3(5.27, 39.01, 9.57, 42.53)
	glyph.Curve3(13.87, 46.04, 21.63, 46.04)
	glyph.Curve3(27.59, 46.04, 31.40, 44.04)
	glyph.Curve3(34.28, 42.53, 35.64, 39.31)
	glyph.Curve3(36.52, 37.21, 36.52, 30.71)
	glyph.LineTo(36.52, 15.53)
	glyph.Curve3(36.52, 9.13, 36.77, 7.69)
	glyph.Curve3(37.01, 6.25, 37.57, 5.76)
	glyph.Curve3(38.13, 5.27, 38.87, 5.27)
	glyph.Curve3(39.65, 5.27, 40.23, 5.62)
	glyph.Curve3(41.26, 6.25, 44.19, 9.18)
	glyph.LineTo(44.19, 6.45)
	glyph.Curve3(38.72, -0.88, 33.74, -0.88)
	glyph.Curve3(31.35, -0.88, 29.93, 0.78)
	glyph.Curve3(28.52, 2.44, 28.47, 6.45)
	glyph.ClosePolygon(basics.PathFlagsNone)
	glyph.MoveTo(28.47, 9.62)
	glyph.LineTo(28.47, 26.66)
	glyph.Curve3(21.09, 23.73, 18.95, 22.51)
	glyph.Curve3(15.09, 20.36, 13.43, 18.02)
	glyph.Curve3(11.77, 15.67, 11.77, 12.89)
	glyph.Curve3(11.77, 9.38, 13.87, 7.06)
	glyph.Curve3(15.97, 4.74, 18.70, 4.74)
	glyph.Curve3(22.41, 4.74, 28.47, 9.62)
	glyph.ClosePolygon(basics.PathFlagsNone)

	glyphMtx := transform.NewTransAffine()
	glyphMtx.Scale(4.0)
	glyphMtx.Translate(220, 200)

	transGlyph := &transformedPathVS{ps: glyph, mtx: glyphMtx}
	curveGlyph := conv.NewConvCurve(transGlyph)

	ras.Reset()
	ras.AddPath(&rasterVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 25})

	ras.Reset()
	ras.AddPath(&rasterVS{src: curveGlyph}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 25})

	mask, _ := generateAlphaMask(ras, sl, stroke, opAND, frameWidth, frameHeight)
	performRendering(ras, sl, mainPixf, mask, curveGlyph)
}

// ---------------------------------------------------------------------------
// pathStorageVS: PathStorageStl as conv.VertexSource
// ---------------------------------------------------------------------------

type pathStorageVS struct{ ps *path.PathStorageStl }

func (p *pathStorageVS) Rewind(id uint) { p.ps.Rewind(id) }
func (p *pathStorageVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := p.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

// ---------------------------------------------------------------------------
// Copy with y-flip
// ---------------------------------------------------------------------------

func copyFlipY(src, dst []uint8, width, height int) {
	stride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

// ---------------------------------------------------------------------------
// Mouse interaction
// ---------------------------------------------------------------------------

func (d *demo) OnMouseButtonDown(x, y float64) bool {
	// Remap y for flip_y=false controls (controls are in window coords).
	if d.polygons.OnMouseButtonDown(x, y) || d.operation.OnMouseButtonDown(x, y) {
		return true
	}
	d.mx = x
	d.my = float64(frameHeight) - y
	return true
}

func (d *demo) OnMouseMove(x, y float64, buttonDown bool) bool {
	if !buttonDown {
		return false
	}
	if d.polygons.OnMouseMove(x, y, buttonDown) || d.operation.OnMouseMove(x, y, buttonDown) {
		return true
	}
	d.mx = x
	d.my = float64(frameHeight) - y
	return true
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	d := newDemo()

	// Suppress unused import warning – fmt used for potential future timing
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AGG Example. Alpha-Mask as a Polygon Clipper",
		Width:  frameWidth,
		Height: frameHeight,
	}, d)
}
