package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/demo/aggshapes"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
)

var (
	am3Operation = 0 // 0: AND, 1: SUB
	am3Polygon   = 3 // Default to GB and Spiral
	am3X, am3Y   float64
)

// ---------------------------------------------------------------------------
// Rasterizer type alias
// ---------------------------------------------------------------------------

type am3RasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func am3NewRasterizer() *am3RasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// conv.VertexSource → rasterizer.VertexSource adapter
// ---------------------------------------------------------------------------

type am3RasterVS struct{ src conv.VertexSource }

func (a *am3RasterVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *am3RasterVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// transformed path: PathStorageStl + affine matrix as conv.VertexSource
// ---------------------------------------------------------------------------

type am3TransformedPathVS struct {
	ps  *path.PathStorageStl
	mtx *transform.TransAffine
}

func (t *am3TransformedPathVS) Rewind(id uint) { t.ps.Rewind(id) }
func (t *am3TransformedPathVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := t.ps.NextVertex()
	t.mtx.Transform(&x, &y)
	return x, y, basics.PathCommand(cmd)
}

// ---------------------------------------------------------------------------
// pathStorageVS: PathStorageStl as conv.VertexSource
// ---------------------------------------------------------------------------

type am3PathStorageVS struct{ ps *path.PathStorageStl }

func (p *am3PathStorageVS) Rewind(id uint) { p.ps.Rewind(id) }
func (p *am3PathStorageVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := p.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

// ---------------------------------------------------------------------------
// Spiral vertex source – matches C++ spiral class from alpha_mask3.cpp
// ---------------------------------------------------------------------------

type am3Spiral struct {
	x, y         float64
	r1, r2       float64
	step         float64
	startAngle   float64
	angle, currR float64
	da, dr       float64
	start        bool
}

func am3NewSpiral(x, y, r1, r2, step, startAngle float64) *am3Spiral {
	return &am3Spiral{
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

func (s *am3Spiral) Rewind(_ uint) {
	s.angle = s.startAngle
	s.currR = s.r1
	s.start = true
}

func (s *am3Spiral) Vertex() (float64, float64, basics.PathCommand) {
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
// Alpha-mask generation helper
// ---------------------------------------------------------------------------

func am3GenerateAlphaMask(
	ras *am3RasType,
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
	ras.AddPath(&am3RasterVS{src: vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, maskRb, fillColor)

	mask := pixfmt.NewAMaskNoClipU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	return mask, maskBuf
}

// ---------------------------------------------------------------------------
// Masked rendering helper
// ---------------------------------------------------------------------------

func am3PerformRendering(
	ras *am3RasType,
	sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mask *pixfmt.AMaskNoClipU8,
	vs conv.VertexSource,
) {
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(mainPixf, mask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)
	ras.Reset()
	ras.AddPath(&am3RasterVS{src: vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rbAMask,
		color.RGBA8[color.Linear]{R: 127, G: 0, B: 0, A: 127})
}

// ---------------------------------------------------------------------------
// Main draw function
// ---------------------------------------------------------------------------

func drawAlphaMask3Demo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	if am3X == 0 && am3Y == 0 {
		am3X = float64(w) / 2
		am3Y = float64(h) / 2
	}

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := am3NewRasterizer()
	sl := scanline.NewScanlineP8()

	opAND := am3Operation == 0

	switch am3Polygon {
	case 0:
		am3RenderTwoSimplePaths(ras, sl, mainPixf, mainRb, opAND, w, h)
	case 1:
		am3RenderClosedStroke(ras, sl, mainPixf, mainRb, opAND, w, h)
	case 2:
		am3RenderGBAndArrows(ras, sl, mainPixf, mainRb, opAND, w, h)
	case 3:
		am3RenderGBAndSpiral(ras, sl, mainPixf, mainRb, opAND, w, h)
	case 4:
		am3RenderSpiralAndGlyph(ras, sl, mainPixf, mainRb, opAND, w, h)
	}

	applyLinearToSRGB(img)

	logStatus(fmt.Sprintf("Alpha Mask 3 Demo: Op=%d, Poly=%d", am3Operation, am3Polygon))
}

// ---------------------------------------------------------------------------
// Case 0: Two simple paths
// ---------------------------------------------------------------------------

func am3RenderTwoSimplePaths(
	ras *am3RasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool, w, h int,
) {
	x := am3X - float64(w)/2 + 100
	y := am3Y - float64(h)/2 + 100

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

	ps1vs := &am3PathStorageVS{ps: ps1}
	ps2vs := &am3PathStorageVS{ps: ps2}

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: ps1vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 25})

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: ps2vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 25})

	mask, _ := am3GenerateAlphaMask(ras, sl, ps1vs, opAND, w, h)
	am3PerformRendering(ras, sl, mainPixf, mask, ps2vs)
}

// ---------------------------------------------------------------------------
// Case 1: Closed stroke
// ---------------------------------------------------------------------------

func am3RenderClosedStroke(
	ras *am3RasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool, w, h int,
) {
	x := am3X - float64(w)/2 + 100
	y := am3Y - float64(h)/2 + 100

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

	ps1vs := &am3PathStorageVS{ps: ps1}
	ps2vs := &am3PathStorageVS{ps: ps2}
	stroke := conv.NewConvStroke(ps2vs)
	stroke.SetWidth(10.0)

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: ps1vs}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 25})

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 25})

	mask, _ := am3GenerateAlphaMask(ras, sl, ps1vs, opAND, w, h)
	am3PerformRendering(ras, sl, mainPixf, mask, stroke)
}

// ---------------------------------------------------------------------------
// Case 2: Great Britain and Arrows
// ---------------------------------------------------------------------------

func am3RenderGBAndArrows(
	ras *am3RasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool, w, h int,
) {
	gbPoly := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(gbPoly)
	arrows := path.NewPathStorageStl()
	aggshapes.MakeArrows(arrows)

	mtx1 := transform.NewTransAffine()
	mtx1.Translate(-1150, -1150)
	mtx1.Scale(2.0)

	mtx2 := *mtx1
	mtx2.Translate(am3X-float64(w)/2, am3Y-float64(h)/2)

	transGB := &am3TransformedPathVS{ps: gbPoly, mtx: mtx1}
	transArrows := &am3TransformedPathVS{ps: arrows, mtx: &mtx2}

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: transGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 127, G: 127, B: 0, A: 25})

	strokeGB := conv.NewConvStroke(transGB)
	strokeGB.SetWidth(0.1)
	ras.Reset()
	ras.AddPath(&am3RasterVS{src: strokeGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: transArrows}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 127, B: 127, A: 25})

	mask, _ := am3GenerateAlphaMask(ras, sl, transGB, opAND, w, h)
	am3PerformRendering(ras, sl, mainPixf, mask, transArrows)
}

// ---------------------------------------------------------------------------
// Case 3: Great Britain and Spiral
// ---------------------------------------------------------------------------

func am3RenderGBAndSpiral(
	ras *am3RasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool, w, h int,
) {
	sp := am3NewSpiral(am3X, am3Y, 10, 150, 30, 0.0)
	stroke := conv.NewConvStroke(sp)
	stroke.SetWidth(15.0)

	gbPoly := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(gbPoly)

	mtx := transform.NewTransAffine()
	mtx.Translate(-1150, -1150)
	mtx.Scale(2.0)

	transGB := &am3TransformedPathVS{ps: gbPoly, mtx: mtx}

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: transGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 127, G: 127, B: 0, A: 25})

	strokeGB := conv.NewConvStroke(transGB)
	strokeGB.SetWidth(0.1)
	ras.Reset()
	ras.AddPath(&am3RasterVS{src: strokeGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 127, B: 127, A: 25})

	mask, _ := am3GenerateAlphaMask(ras, sl, transGB, opAND, w, h)
	am3PerformRendering(ras, sl, mainPixf, mask, stroke)
}

// ---------------------------------------------------------------------------
// Case 4: Spiral and Glyph
// ---------------------------------------------------------------------------

func am3RenderSpiralAndGlyph(
	ras *am3RasType, sl *scanline.ScanlineP8,
	mainPixf *pixfmt.PixFmtRGBA32[color.Linear],
	mainRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	opAND bool, w, h int,
) {
	sp := am3NewSpiral(am3X, am3Y, 10, 150, 30, 0.0)
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

	transGlyph := &am3TransformedPathVS{ps: glyph, mtx: glyphMtx}
	curveGlyph := conv.NewConvCurve(transGlyph)

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 25})

	ras.Reset()
	ras.AddPath(&am3RasterVS{src: curveGlyph}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb, color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 25})

	mask, _ := am3GenerateAlphaMask(ras, sl, stroke, opAND, w, h)
	am3PerformRendering(ras, sl, mainPixf, mask, curveGlyph)
}
