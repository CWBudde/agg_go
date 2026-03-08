package quadwarp

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	imgacc "agg_go/internal/image"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	"agg_go/internal/scanline"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

type TransformMode int

const (
	TransformAffine TransformMode = iota
	TransformBilinear
	TransformPerspective
)

type InterpolatorMode int

const (
	InterpolatorLinear InterpolatorMode = iota
	InterpolatorLinearSubdiv
	InterpolatorTrans
	InterpolatorPerspectiveLerp
	InterpolatorPerspectiveExact
)

type SampleMode int

const (
	SampleNearest SampleMode = iota
	SampleFilter2x2
	SampleResample
)

type SourceMode int

const (
	SourceClone SourceMode = iota
	SourceWrapReflect
)

type Config struct {
	CanvasWidth  int
	CanvasHeight int
	Source       *agg.Image
	SourceRect   [4]float64 // x1,y1,x2,y2
	Quad         [4][2]float64

	Transform    TransformMode
	Interpolator InterpolatorMode
	Sampling     SampleMode
	SourceMode   SourceMode

	FilterKernel imgacc.FilterFunction
	Normalize    bool
	Blur         float64

	ForceParallelogram bool

	ShowQuadFill    bool
	ShowQuadOutline bool
	ShowHandles     bool
	QuadFillColor   agg.Color
	QuadLineColor   agg.Color
}

type pixFmtSrc struct {
	rbuf *buffer.RenderingBufferU8
}

func (p pixFmtSrc) Width() int    { return p.rbuf.Width() }
func (p pixFmtSrc) Height() int   { return p.rbuf.Height() }
func (p pixFmtSrc) PixWidth() int { return 4 }
func (p pixFmtSrc) PixPtr(x, y int) []basics.Int8u {
	return buffer.RowU8(p.rbuf, y)[x*4:]
}

type accessor interface {
	Span(x, y, length int) []basics.Int8u
	NextX() []basics.Int8u
	NextY() []basics.Int8u
}

type rgbaSource struct {
	acc accessor
	pf  *pixFmtSrc
}

func (s *rgbaSource) Width() int                  { return s.pf.Width() }
func (s *rgbaSource) Height() int                 { return s.pf.Height() }
func (s *rgbaSource) ColorType() string           { return "RGBA8" }
func (s *rgbaSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *rgbaSource) Span(x, y, length int) []basics.Int8u {
	return s.acc.Span(x, y, length)
}
func (s *rgbaSource) NextX() []basics.Int8u { return s.acc.NextX() }
func (s *rgbaSource) NextY() []basics.Int8u { return s.acc.NextY() }
func (s *rgbaSource) RowPtr(y int) []basics.Int8u {
	return s.pf.PixPtr(0, y)
}

type spanGen interface {
	Generate([]color.RGBA8[color.Linear], int, int)
}

type pathAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathAdapter) Rewind(pathID uint32) {
	a.ps.Rewind(uint(pathID))
}

func (a *pathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

type scanlineAdapter struct {
	sl *scanline.ScanlineU8
}

func (a *scanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *scanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *scanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *scanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *scanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

func Draw(ctx *agg.Context, cfg Config) {
	if ctx == nil || cfg.Source == nil {
		return
	}

	quad := cfg.Quad
	if cfg.ForceParallelogram {
		quad[3][0] = quad[0][0] + (quad[2][0] - quad[1][0])
		quad[3][1] = quad[0][1] + (quad[2][1] - quad[1][1])
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	if cfg.ShowQuadFill {
		agg2d.FillColor(cfg.QuadFillColor)
		agg2d.NoLine()
		agg2d.ResetPath()
		agg2d.MoveTo(quad[0][0], quad[0][1])
		agg2d.LineTo(quad[1][0], quad[1][1])
		agg2d.LineTo(quad[2][0], quad[2][1])
		agg2d.LineTo(quad[3][0], quad[3][1])
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}

	outImg := ctx.GetImage()
	outRbuf := buffer.NewRenderingBufferU8()
	outRbuf.Attach(outImg.Data, outImg.Width(), outImg.Height(), outImg.Width()*4)
	outPixFmt := pixfmt.NewPixFmtRGBA32PreLinear(outRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](outPixFmt)

	srcRbuf := buffer.NewRenderingBufferU8()
	srcRbuf.Attach(cfg.Source.Data, cfg.Source.Width(), cfg.Source.Height(), cfg.Source.Width()*4)
	srcPf := pixFmtSrc{rbuf: srcRbuf}
	var acc accessor
	switch cfg.SourceMode {
	case SourceWrapReflect:
		wx := imgacc.NewWrapModeReflectAutoPow2(basics.Int32u(cfg.Source.Width()))
		wy := imgacc.NewWrapModeReflectAutoPow2(basics.Int32u(cfg.Source.Height()))
		acc = imgacc.NewImageAccessorWrap[pixFmtSrc, *imgacc.WrapModeReflectAutoPow2, *imgacc.WrapModeReflectAutoPow2](&srcPf, wx, wy)
	default:
		acc = imgacc.NewImageAccessorClone(&srcPf)
	}
	source := &rgbaSource{acc: acc, pf: &srcPf}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	pth := path.NewPathStorageStl()
	pth.MoveTo(quad[0][0], quad[0][1])
	pth.LineTo(quad[1][0], quad[1][1])
	pth.LineTo(quad[2][0], quad[2][1])
	pth.LineTo(quad[3][0], quad[3][1])
	pth.ClosePolygon(basics.PathFlagsNone)

	ras.Reset()
	ras.ClipBox(0, 0, float64(cfg.CanvasWidth), float64(cfg.CanvasHeight))
	ras.AddPath(&pathAdapter{ps: pth}, 0)

	filter := imgacc.NewImageFilterLUTWithFilter(cfg.FilterKernel, cfg.Normalize)

	x1, y1 := cfg.SourceRect[0], cfg.SourceRect[1]
	x2, y2 := cfg.SourceRect[2], cfg.SourceRect[3]
	q8 := [8]float64{
		quad[0][0], quad[0][1],
		quad[1][0], quad[1][1],
		quad[2][0], quad[2][1],
		quad[3][0], quad[3][1],
	}

	var gen spanGen

	switch cfg.Transform {
	case TransformAffine:
		dstParl := [6]float64{quad[0][0], quad[0][1], quad[1][0], quad[1][1], quad[2][0], quad[2][1]}
		tr := transform.NewTransAffineParlToRect(dstParl, x1, y1, x2, y2)
		interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](tr, 8)
		switch cfg.Sampling {
		case SampleNearest:
			gen = span.NewSpanImageFilterRGBANNWithParams[*rgbaSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interp)
		case SampleResample:
			rg := span.NewSpanImageResampleRGBAAffineWithParams[*rgbaSource](source, interp, filter)
			rg.Blur(cfg.Blur)
			gen = rg
		default:
			gen = span.NewSpanImageFilterRGBA2x2WithParams[*rgbaSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interp, filter)
		}
	case TransformBilinear:
		tr := transform.NewTransBilinearQuadToRect(q8, x1, y1, x2, y2)
		if !tr.IsValid() {
			return
		}
		interp := span.NewSpanInterpolatorLinear[*transform.TransBilinear](tr, 8)
		switch cfg.Sampling {
		case SampleNearest:
			gen = span.NewSpanImageFilterRGBANNWithParams[*rgbaSource, *span.SpanInterpolatorLinear[*transform.TransBilinear]](source, interp)
		case SampleResample:
			// Closest available equivalent: general resample over perspective-lerp exactness is not needed for bilinear mode.
			rg := span.NewSpanImageResampleRGBAWithParams[*rgbaSource, *span.SpanInterpolatorLinear[*transform.TransBilinear]](source, interp, filter)
			gen = rg
		default:
			gen = span.NewSpanImageFilterRGBA2x2WithParams[*rgbaSource, *span.SpanInterpolatorLinear[*transform.TransBilinear]](source, interp, filter)
		}
	case TransformPerspective:
		switch cfg.Interpolator {
		case InterpolatorTrans:
			tr := transform.NewTransPerspectiveQuadToRect(q8, x1, y1, x2, y2)
			if !tr.IsValid(1e-10) {
				return
			}
			interp := span.NewSpanInterpolatorTrans[*transform.TransPerspective](tr)
			switch cfg.Sampling {
			case SampleResample:
				rg := span.NewSpanImageResampleRGBAWithParams[*rgbaSource, *span.SpanInterpolatorTrans[*transform.TransPerspective]](source, interp, filter)
				rg.Blur(cfg.Blur)
				gen = rg
			case SampleNearest:
				gen = span.NewSpanImageFilterRGBANNWithParams[*rgbaSource, *span.SpanInterpolatorTrans[*transform.TransPerspective]](source, interp)
			default:
				gen = span.NewSpanImageFilterRGBA2x2WithParams[*rgbaSource, *span.SpanInterpolatorTrans[*transform.TransPerspective]](source, interp, filter)
			}
		case InterpolatorPerspectiveLerp:
			interp := span.NewSpanInterpolatorPerspectiveLerpQuadToRect(q8, x1, y1, x2, y2, 8)
			if !interp.IsValid() {
				return
			}
			switch cfg.Sampling {
			case SampleResample:
				subdiv := span.NewSpanSubdivAdaptor(interp)
				rg := span.NewSpanImageResampleRGBAWithParams[*rgbaSource, *span.SpanSubdivAdaptor[*span.SpanInterpolatorPerspectiveLerp]](source, subdiv, filter)
				rg.Blur(cfg.Blur)
				gen = rg
			case SampleNearest:
				gen = span.NewSpanImageFilterRGBANNWithParams[*rgbaSource, *span.SpanInterpolatorPerspectiveLerp](source, interp)
			default:
				gen = span.NewSpanImageFilterRGBA2x2WithParams[*rgbaSource, *span.SpanInterpolatorPerspectiveLerp](source, interp, filter)
			}
		case InterpolatorPerspectiveExact:
			interp := span.NewSpanInterpolatorPerspectiveExactQuadToRect(q8, x1, y1, x2, y2, 8)
			if !interp.IsValid() {
				return
			}
			switch cfg.Sampling {
			case SampleResample:
				subdiv := span.NewSpanSubdivAdaptor(interp)
				rg := span.NewSpanImageResampleRGBAWithParams[*rgbaSource, *span.SpanSubdivAdaptor[*span.SpanInterpolatorPerspectiveExact]](source, subdiv, filter)
				rg.Blur(cfg.Blur)
				gen = rg
			case SampleNearest:
				gen = span.NewSpanImageFilterRGBANNWithParams[*rgbaSource, *span.SpanInterpolatorPerspectiveExact](source, interp)
			default:
				gen = span.NewSpanImageFilterRGBA2x2WithParams[*rgbaSource, *span.SpanInterpolatorPerspectiveExact](source, interp, filter)
			}
		default:
			tr := transform.NewTransPerspectiveQuadToRect(q8, x1, y1, x2, y2)
			if !tr.IsValid(1e-10) {
				return
			}
			// AGG uses span_interpolator_linear_subdiv<..., 8> where SubdivShift
			// stays at its default (4). Keep that parity here.
			interp := span.NewSpanInterpolatorLinearSubdiv[*transform.TransPerspective](tr, 8, 4)
			switch cfg.Sampling {
			case SampleResample:
				rg := span.NewSpanImageResampleRGBAWithParams[*rgbaSource, *span.SpanInterpolatorLinearSubdiv[*transform.TransPerspective]](source, interp, filter)
				rg.Blur(cfg.Blur)
				gen = rg
			case SampleNearest:
				gen = span.NewSpanImageFilterRGBANNWithParams[*rgbaSource, *span.SpanInterpolatorLinearSubdiv[*transform.TransPerspective]](source, interp)
			default:
				gen = span.NewSpanImageFilterRGBA2x2WithParams[*rgbaSource, *span.SpanInterpolatorLinearSubdiv[*transform.TransPerspective]](source, interp, filter)
			}
		}
	}

	if gen == nil {
		return
	}
	if prep, ok := gen.(interface{ Prepare() }); ok {
		prep.Prepare()
	}

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&scanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spn := range sl.Spans() {
				if spn.Len <= 0 {
					continue
				}
				count := int(spn.Len)
				colors := alloc.Allocate(count)
				gen.Generate(colors[:count], int(spn.X), y)
				renBase.BlendColorHspan(int(spn.X), y, count, colors, spn.Covers, basics.CoverFull)
			}
		}
	}

	if cfg.ShowQuadOutline {
		ctx.SetColor(cfg.QuadLineColor)
		ctx.SetLineWidth(1.5)
		ctx.BeginPath()
		ctx.MoveTo(quad[0][0], quad[0][1])
		ctx.LineTo(quad[1][0], quad[1][1])
		ctx.LineTo(quad[2][0], quad[2][1])
		ctx.LineTo(quad[3][0], quad[3][1])
		ctx.ClosePath()
		ctx.Stroke()
	}
	if cfg.ShowHandles {
		for i := 0; i < 4; i++ {
			x := quad[i][0]
			y := quad[i][1]
			ctx.SetColor(agg.RGBA(0.8, 0.1, 0.1, 0.75))
			ctx.FillCircle(x, y, 4)
			ctx.SetColor(agg.RGBA(0.05, 0.05, 0.05, 0.85))
			ctx.DrawCircle(x, y, 4)
		}
	}
}

// ApplyGammaInv applies inverse gamma to RGB channels in-place.
func ApplyGammaInv(img *agg.Image, gamma float64) {
	if img == nil || gamma <= 0 || math.Abs(gamma-1.0) < 1e-9 {
		return
	}
	inv := 1.0 / gamma
	for i := 0; i+3 < len(img.Data); i += 4 {
		r := math.Pow(float64(img.Data[i])/255.0, inv)
		g := math.Pow(float64(img.Data[i+1])/255.0, inv)
		b := math.Pow(float64(img.Data[i+2])/255.0, inv)
		img.Data[i] = uint8(r*255.0 + 0.5)
		img.Data[i+1] = uint8(g*255.0 + 0.5)
		img.Data[i+2] = uint8(b*255.0 + 0.5)
	}
}

// CopyWithGammaDir returns a copy with direct gamma applied to RGB channels.
func CopyWithGammaDir(src *agg.Image, gamma float64) *agg.Image {
	if src == nil {
		return nil
	}
	out := agg.NewImage(append([]byte(nil), src.Data...), src.Width(), src.Height(), src.Width()*4)
	if gamma <= 0 || math.Abs(gamma-1.0) < 1e-9 {
		return out
	}
	for i := 0; i+3 < len(out.Data); i += 4 {
		r := math.Pow(float64(out.Data[i])/255.0, gamma)
		g := math.Pow(float64(out.Data[i+1])/255.0, gamma)
		b := math.Pow(float64(out.Data[i+2])/255.0, gamma)
		out.Data[i] = uint8(r*255.0 + 0.5)
		out.Data[i+1] = uint8(g*255.0 + 0.5)
		out.Data[i+2] = uint8(b*255.0 + 0.5)
	}
	return out
}
