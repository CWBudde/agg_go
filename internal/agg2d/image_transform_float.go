// Package agg2d: float (RGBA32) affine/perspective image transform. This is the
// float twin of the TransformImage* path in image.go (PLAN.md Phase 4). It mirrors
// the 8-bit renderImage/newImageFilterGenerator/renderImagePerspective pipeline,
// swapping the RGBA8 span image filters for their RGBA32 counterparts and the
// straight Image source for an ImageFloat source.
//
// Composite image blend modes are honored: when a blend mode other than
// BlendAlpha is set, image transfers route through the premultiplied composite
// base renderer (renBaseCompPre). The image blend color's alpha is still applied
// after sampling, identical to the 8-bit span generator.
package agg2d

import (
	"errors"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

// imageSampleGeneratorFloat is the float sample-filter contract: it fills a span
// of RGBA32 colors for the destination pixel run starting at (x,y).
type imageSampleGeneratorFloat interface {
	Generate(span []color.RGBA32[color.Linear], x, y int)
}

// imageSpanGeneratorFloat adapts a float sample filter to the scanline renderer's
// SpanGeneratorInterface, applying the image blend color's alpha after sampling.
type imageSpanGeneratorFloat struct {
	sample     imageSampleGeneratorFloat
	blendColor Color
}

func newImageSpanGeneratorFloat(sample imageSampleGeneratorFloat, blendColor Color) *imageSpanGeneratorFloat {
	return &imageSpanGeneratorFloat{sample: sample, blendColor: blendColor}
}

func (sg *imageSpanGeneratorFloat) Prepare() {
	if preparer, ok := sg.sample.(interface{ Prepare() }); ok {
		preparer.Prepare()
	}
}

func (sg *imageSpanGeneratorFloat) Generate(colors []color.RGBA32[color.Linear], x, y, length int) {
	if sg.sample == nil || length <= 0 || len(colors) == 0 {
		return
	}
	if length < len(colors) {
		colors = colors[:length]
	}

	sg.sample.Generate(colors, x, y)

	// AGG applies the image blend color's alpha after sampling. The composite
	// operator itself is applied downstream by the composite base renderer, so
	// only the blend-color alpha scaling is honored here.
	if sg.blendColor[3] != 255 {
		alpha := float32(sg.blendColor[3]) / 255
		for i := range colors {
			colors[i].R *= alpha
			colors[i].G *= alpha
			colors[i].B *= alpha
			colors[i].A *= alpha
		}
	}
}

// currentImageRenderer returns the base renderer used for image transfers. As in
// the 8-bit path, images blend through the premultiplied base renderer, using
// the composite variant when a Porter-Duff / SVG blend mode is active.
func (a *Agg2DFloat) currentImageRenderer() *baseRendererAdapter[color.RGBA32[color.Linear]] {
	if a.blendMode != BlendAlpha && a.renBaseCompPre != nil {
		return a.renBaseCompPre
	}
	return a.renBasePre
}

// newImageFilterGeneratorFloat selects the float RGBA32 image filter that matches
// the current filter/resample state, mirroring the 8-bit newImageFilterGenerator.
func (a *Agg2DFloat) newImageFilterGeneratorFloat(
	source *imagePixelFormatFloat,
	interpolator *span.SpanInterpolatorLinear[*transform.TransAffine],
) imageSampleGeneratorFloat {
	if a.imageFilter == NoFilter {
		return span.NewSpanImageFilterRGBA32NNWithParams[*imagePixelFormatFloat, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interpolator)
	}

	resample := a.imageResample == ResampleAlways
	if a.imageResample == ResampleOnZoomOut && interpolator != nil {
		if tr := interpolator.Transformer(); tr != nil {
			sx, sy := tr.GetScalingAbs()
			if sx > 1.125 || sy > 1.125 {
				resample = true
			}
		}
	}
	if !resample && a.affineImageResamplePolicy == AffineImageResamplePreferFiltered && a.imageFilter != NoFilter {
		resample = true
	}

	if resample {
		return span.NewSpanImageResampleRGBA32AffineWithParams[*imagePixelFormatFloat](
			source,
			interpolator,
			a.imageFilterLUT,
		)
	}

	if a.imageFilter == Bilinear {
		return span.NewSpanImageFilterRGBA32BilinearWithParams[*imagePixelFormatFloat, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interpolator)
	}

	if a.imageFilterLUT == nil {
		return span.NewSpanImageFilterRGBA32BilinearWithParams[*imagePixelFormatFloat, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interpolator)
	}

	if a.imageFilterLUT.Diameter() == 2 {
		return span.NewSpanImageFilterRGBA32_2x2WithParams[*imagePixelFormatFloat, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			interpolator,
			a.imageFilterLUT,
		)
	}

	return span.NewSpanImageFilterRGBA32WithParams[*imagePixelFormatFloat, *span.SpanInterpolatorLinear[*transform.TransAffine]](
		source,
		interpolator,
		a.imageFilterLUT,
	)
}

// setImagePathRect builds an axis-aligned rectangle path in world coordinates.
func (a *Agg2DFloat) setImagePathRect(x1, y1, x2, y2 float64) {
	a.ResetPath()
	a.MoveTo(x1, y1)
	a.LineTo(x2, y1)
	a.LineTo(x2, y2)
	a.LineTo(x1, y2)
	a.ClosePolygon()
}

// setImagePathParallelogram builds a parallelogram path from 3 corners.
func (a *Agg2DFloat) setImagePathParallelogram(parallelogram []float64) {
	a.ResetPath()
	a.MoveTo(parallelogram[0], parallelogram[1])
	a.LineTo(parallelogram[2], parallelogram[3])
	a.LineTo(parallelogram[4], parallelogram[5])
	a.LineTo(
		parallelogram[0]+parallelogram[4]-parallelogram[2],
		parallelogram[1]+parallelogram[5]-parallelogram[3],
	)
	a.ClosePolygon()
}

// addImagePathToRasterizer feeds the current (curve-flattened, world-transformed)
// path into the rasterizer.
func (a *Agg2DFloat) addImagePathToRasterizer() {
	if a.path == nil || a.path.TotalVertices() == 0 || a.rasterizer == nil {
		return
	}
	transformedPath := conv.NewConvTransform(a.convCurve, a.transform)
	transformedPath.Rewind(0)
	for {
		x, y, cmd := transformedPath.Vertex()
		if cmd == basics.PathCmdStop {
			return
		}
		a.rasterizer.AddVertex(x, y, uint32(cmd))
	}
}

func (a *Agg2DFloat) imageFillingRule() basics.FillingRule {
	if a.evenOddFlag {
		return basics.FillEvenOdd
	}
	return basics.FillNonZero
}

// renderImageFloat renders the current path using AGG-style affine image span
// interpolation. parallelogram holds the 3 destination corners (6 floats).
func (a *Agg2DFloat) renderImageFloat(img *ImageFloat, x1, y1, x2, y2 int, parallelogram []float64) error {
	if img == nil || img.renBuf == nil {
		return errors.New("image or image buffer is nil")
	}
	if len(parallelogram) != 6 {
		return errors.New("parallelogram must have exactly 6 elements")
	}
	if a.rasterizer == nil || a.scanline == nil || a.spanAllocator == nil {
		return errors.New("render pipeline is not initialized")
	}

	src := [6]float64{
		float64(x1), float64(y1),
		float64(x2), float64(y1),
		float64(x2), float64(y2),
	}
	dst := [6]float64{
		parallelogram[0], parallelogram[1],
		parallelogram[2], parallelogram[3],
		parallelogram[4], parallelogram[5],
	}

	mtx := transform.NewTransAffineParlToParl(src, dst)
	if a.transform != nil {
		mtx.Multiply(a.transform)
	}
	mtx.Invert()

	a.rasterizer.Reset()
	a.rasterizer.FillingRule(a.imageFillingRule())
	a.addImagePathToRasterizer()

	interpolator := span.NewSpanInterpolatorLinearDefault(mtx)
	imageSource := newImagePixelFormatFloat(img)
	sampleGenerator := a.newImageFilterGeneratorFloat(imageSource, interpolator)
	spanGenerator := newImageSpanGeneratorFloat(sampleGenerator, a.imageBlendColor)

	renderer := a.currentImageRenderer()
	if renderer == nil {
		return nil
	}

	renscan.RenderScanlinesAA[color.RGBA32[color.Linear]](
		a.rasterizer, a.scanline, renderer, a.spanAllocator, spanGenerator,
	)
	return nil
}

// renderImagePerspectiveFloat renders img through a perspective (homographic)
// transform mapping the source rectangle to the destination quadrangle quad
// (8 floats: TL, TR, BR, BL). Bilinear filtering is always used.
func (a *Agg2DFloat) renderImagePerspectiveFloat(img *ImageFloat, x1, y1, x2, y2 int, quad [8]float64) error {
	if img == nil || img.renBuf == nil {
		return errors.New("image or image buffer is nil")
	}
	if a.rasterizer == nil || a.scanline == nil || a.spanAllocator == nil {
		return errors.New("render pipeline is not initialized")
	}

	interpolator := span.NewSpanInterpolatorPerspectiveLerp(0)
	interpolator.QuadToRect(quad, float64(x1), float64(y1), float64(x2), float64(y2))
	if !interpolator.IsValid() {
		return errors.New("degenerate perspective transform")
	}

	a.ResetPath()
	a.MoveTo(quad[0], quad[1])
	a.LineTo(quad[2], quad[3])
	a.LineTo(quad[4], quad[5])
	a.LineTo(quad[6], quad[7])
	a.ClosePolygon()

	a.rasterizer.Reset()
	a.rasterizer.FillingRule(a.imageFillingRule())
	a.addImagePathToRasterizer()

	imageSource := newImagePixelFormatFloat(img)
	sampleGenerator := span.NewSpanImageFilterRGBA32BilinearWithParams[*imagePixelFormatFloat, *span.SpanInterpolatorPerspectiveLerp](imageSource, interpolator)
	spanGenerator := newImageSpanGeneratorFloat(sampleGenerator, a.imageBlendColor)

	renderer := a.currentImageRenderer()
	if renderer == nil {
		return nil
	}
	renscan.RenderScanlinesAA[color.RGBA32[color.Linear]](
		a.rasterizer, a.scanline, renderer, a.spanAllocator, spanGenerator,
	)
	return nil
}

//----------------------------------------------------------------------------
// Public float TransformImage* surface (mirrors the 8-bit Agg2D methods).
//----------------------------------------------------------------------------

// TransformImageFloat transforms a source rectangle of img to a destination
// rectangle. This is the most general affine form; other overloads delegate here.
func (a *Agg2DFloat) TransformImageFloat(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	if imgX1 < 0 || imgY1 < 0 || imgX2 > img.Width() || imgY2 > img.Height() {
		return errors.New("invalid source rectangle bounds")
	}

	a.setImagePathRect(dstX1, dstY1, dstX2, dstY2)
	parallelogram := []float64{dstX1, dstY1, dstX2, dstY1, dstX2, dstY2}
	return a.renderImageFloat(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImageFloatSimple transforms the whole image to a destination rectangle.
func (a *Agg2DFloat) TransformImageFloatSimple(img *ImageFloat, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	return a.TransformImageFloat(img, 0, 0, img.Width(), img.Height(), dstX1, dstY1, dstX2, dstY2)
}

// TransformImageFloatParallelogram transforms a source rectangle to a destination
// parallelogram (6 floats: 3 corners).
func (a *Agg2DFloat) TransformImageFloatParallelogram(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	if len(parallelogram) != 6 {
		return errors.New("parallelogram must have exactly 6 elements (x1, y1, x2, y2, x3, y3)")
	}
	if imgX1 < 0 || imgY1 < 0 || imgX2 > img.Width() || imgY2 > img.Height() {
		return errors.New("invalid source rectangle bounds")
	}

	a.setImagePathParallelogram(parallelogram)
	return a.renderImageFloat(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImageFloatParallelogramSimple transforms the whole image to a
// destination parallelogram.
func (a *Agg2DFloat) TransformImageFloatParallelogramSimple(img *ImageFloat, parallelogram []float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	return a.TransformImageFloatParallelogram(img, 0, 0, img.Width(), img.Height(), parallelogram)
}

// TransformImageFloatPath transforms a source rectangle along the current path,
// clipping the image to the path shape (matching AGG's transformImagePath).
func (a *Agg2DFloat) TransformImageFloatPath(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	parallelogram := []float64{dstX1, dstY1, dstX2, dstY1, dstX2, dstY2}
	return a.renderImageFloat(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImageFloatPathSimple transforms the whole image along the current path.
func (a *Agg2DFloat) TransformImageFloatPathSimple(img *ImageFloat, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	return a.TransformImageFloatPath(img, 0, 0, img.Width(), img.Height(), dstX1, dstY1, dstX2, dstY2)
}

// TransformImageFloatPathParallelogram transforms a source rectangle along the
// current path to a destination parallelogram.
func (a *Agg2DFloat) TransformImageFloatPathParallelogram(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	if len(parallelogram) < 6 {
		return errors.New("parallelogram requires 6 coordinates (3 points)")
	}
	return a.renderImageFloat(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImageFloatPathParallelogramSimple transforms the whole image along the
// current path to a destination parallelogram.
func (a *Agg2DFloat) TransformImageFloatPathParallelogramSimple(img *ImageFloat, parallelogram []float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	return a.TransformImageFloatPathParallelogram(img, 0, 0, img.Width(), img.Height(), parallelogram)
}

// TransformImageFloatQuad transforms a source rectangle to an arbitrary
// destination quadrangle using perspective interpolation. quad holds the eight
// destination coordinates [x0,y0, x1,y1, x2,y2, x3,y3] for TL, TR, BR, BL.
func (a *Agg2DFloat) TransformImageFloatQuad(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, quad [8]float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	if imgX1 < 0 || imgY1 < 0 || imgX2 > img.Width() || imgY2 > img.Height() {
		return errors.New("invalid source rectangle bounds")
	}
	return a.renderImagePerspectiveFloat(img, imgX1, imgY1, imgX2, imgY2, quad)
}

// TransformImageFloatQuadSimple transforms the whole image to an arbitrary
// destination quadrangle using perspective interpolation.
func (a *Agg2DFloat) TransformImageFloatQuadSimple(img *ImageFloat, quad [8]float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	return a.renderImagePerspectiveFloat(img, 0, 0, img.Width(), img.Height(), quad)
}
