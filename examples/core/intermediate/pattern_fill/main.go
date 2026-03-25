// Port of AGG C++ pattern_fill.cpp.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	imageacc "github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	canvasW = 640
	canvasH = 480

	defaultPolygonAngle = 0.0
	defaultPolygonScale = 1.0
	defaultPatternAngle = 0.0
	defaultPatternSize  = 30.0
	defaultPatternAlpha = 0.1
)

type patternPixFmt struct {
	data         []basics.Int8u
	w, h, stride int
}

func (p patternPixFmt) Width() int    { return p.w }
func (p patternPixFmt) Height() int   { return p.h }
func (p patternPixFmt) PixWidth() int { return 4 }

func (p patternPixFmt) PixPtr(x, y int) []basics.Int8u {
	if y < 0 || y >= p.h || x < 0 || x >= p.w {
		return p.data[:0]
	}
	return p.data[y*p.stride+x*4:]
}

type patternSource struct {
	accessor *imageacc.ImageAccessorWrap[patternPixFmt, *imageacc.WrapModeReflectAutoPow2, *imageacc.WrapModeReflectAutoPow2]
	pf       patternPixFmt
}

func (s *patternSource) Width() int                  { return s.pf.w }
func (s *patternSource) Height() int                 { return s.pf.h }
func (s *patternSource) ColorType() string           { return "RGBA8" }
func (s *patternSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *patternSource) Span(x, y, length int) []basics.Int8u {
	return s.accessor.Span(x, y, length)
}
func (s *patternSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *patternSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *patternSource) RowPtr(y int) []basics.Int8u {
	return s.pf.PixPtr(0, y)
}

type rasterizerVertexSource interface {
	Rewind(id uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type rasterizerAdapter struct {
	source rasterizerVertexSource
}

func (a *rasterizerAdapter) Rewind(id uint32) { a.source.Rewind(uint(id)) }
func (a *rasterizerAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.source.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func createStar(ps *path.PathStorageStl, xc, yc, r1, r2 float64, n uint, startAngleDeg float64) {
	ps.RemoveAll()
	startAngleRad := startAngleDeg * math.Pi / 180.0
	for i := uint(0); i < n; i++ {
		a := math.Pi*2.0*float64(i)/float64(n) - math.Pi/2.0
		dx := math.Cos(a + startAngleRad)
		dy := math.Sin(a + startAngleRad)

		if i&1 != 0 {
			ps.LineTo(xc+dx*r1, yc+dy*r1)
			continue
		}
		if i == 0 {
			ps.MoveTo(xc+dx*r2, yc+dy*r2)
			continue
		}
		ps.LineTo(xc+dx*r2, yc+dy*r2)
	}
	ps.ClosePolygon(basics.PathFlagsClose)
}

func generatePattern(size int, patternAngle, patternAlpha float64) patternPixFmt {
	ps := path.NewPathStorageStl()
	createStar(ps, float64(size)/2.0, float64(size)/2.0, float64(size)/2.5, float64(size)/6.0, 6, patternAngle)

	smooth := conv.NewConvSmoothPoly1Curve(path.NewPathStorageStlVertexSourceAdapter(ps))
	smooth.SetSmoothValue(1.0)
	smooth.SetApproximationScale(4.0)

	stroke := conv.NewConvStroke(smooth)
	stroke.SetWidth(float64(size) / 15.0)

	pf := patternPixFmt{
		data:   make([]basics.Int8u, size*size*4),
		w:      size,
		h:      size,
		stride: size * 4,
	}
	rbuf := buffer.NewRenderingBufferWithData[uint8](pf.data, size, size, pf.stride)
	pixf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pixf)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer[*renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](renBase)
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	renBase.Clear(color.NewRGBA8[color.Linear](102, 0, 26, basics.Int8u(patternAlpha*255.0)))

	ras.AddPath(&rasterizerAdapter{source: smooth}, 0)
	renSolid.SetColor(color.NewRGBA8[color.Linear](110, 130, 50, 255))
	renscan.RenderScanlinesAASolid(ras, sl, renSolid.BaseRenderer(), renSolid.Color())

	ras.AddPath(&rasterizerAdapter{source: stroke}, 0)
	renSolid.SetColor(color.NewRGBA8[color.Linear](0, 50, 80, 255))
	renscan.RenderScanlinesAASolid(ras, sl, renSolid.BaseRenderer(), renSolid.Color())

	return pf
}

type demo struct {
	patternSize  float64
	patternAngle float64
	patternAlpha float64
	polygonAngle float64
	polygonScale float64
	polygonCX    float64
	polygonCY    float64
}

func newDemo() *demo {
	return &demo{
		patternSize:  defaultPatternSize,
		patternAngle: defaultPatternAngle,
		patternAlpha: defaultPatternAlpha,
		polygonAngle: defaultPolygonAngle,
		polygonScale: defaultPolygonScale,
		polygonCX:    float64(canvasW) / 2.0,
		polygonCY:    float64(canvasH) / 2.0,
	}
}

func (d *demo) Render(img *agg.Image) {
	dstRbuf := buffer.NewRenderingBufferWithData[uint8](img.Data, img.Width(), img.Height(), img.Stride())
	dstPixf := pixfmt.NewPixFmtRGBA32PreLinear(dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	renBase.Clear(color.NewRGBA8[color.Linear](255, 255, 255, 255))

	size := int(d.patternSize)
	pf := generatePattern(size, d.patternAngle, d.patternAlpha)

	wrapX := imageacc.NewWrapModeReflectAutoPow2(basics.Int32u(pf.w))
	wrapY := imageacc.NewWrapModeReflectAutoPow2(basics.Int32u(pf.h))
	imgSrc := imageacc.NewImageAccessorWrap[patternPixFmt, *imageacc.WrapModeReflectAutoPow2, *imageacc.WrapModeReflectAutoPow2](&pf, wrapX, wrapY)
	sg := span.NewSpanPatternRGBAWithParams[*patternSource](
		&patternSource{accessor: imgSrc, pf: pf},
		0,
		0,
	)

	ps := path.NewPathStorageStl()
	r := float64(canvasW)/3.0 - 8.0
	createStar(ps, d.polygonCX, d.polygonCY, r, r/1.45, 14, 0.0)

	polygonMtx := transform.NewTransAffine()
	polygonMtx.Multiply(transform.NewTransAffineTranslation(-d.polygonCX, -d.polygonCY))
	polygonMtx.Multiply(transform.NewTransAffineRotation(d.polygonAngle * math.Pi / 180.0))
	polygonMtx.Multiply(transform.NewTransAffineScaling(d.polygonScale))
	polygonMtx.Multiply(transform.NewTransAffineTranslation(d.polygonCX, d.polygonCY))

	tr := conv.NewConvTransform(path.NewPathStorageStlVertexSourceAdapter(ps), polygonMtx)
	ras.AddPath(&rasterizerAdapter{source: tr}, 0)
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		sg.Prepare()
		for ras.SweepScanline(sl) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				length := int(spanData.Len)
				if length < 0 {
					length = -length
				}
				if length == 0 {
					continue
				}
				colors := alloc.Allocate(length)
				sg.Generate(colors, int(spanData.X), y, uint(length))
				if spanData.Len < 0 {
					renBase.BlendColorHspan(int(spanData.X), y, length, colors, nil, spanData.Covers[0])
					continue
				}
				renBase.BlendColorHspan(int(spanData.X), y, length, colors, spanData.Covers, spanData.Covers[0])
			}
		}
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Pattern Fill",
		Width:  canvasW,
		Height: canvasH,
		FlipY:  true,
	}, newDemo())
}
