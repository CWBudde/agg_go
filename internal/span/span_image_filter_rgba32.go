// Package span: float (RGBA32, 4 x float32) twins of the RGBA image-filter span
// generators. These mirror span_image_filter_rgba.go's 8-bit generators but
// operate on straight float32 source channels and emit color.RGBA32 spans, as
// required by the float Agg2D image-transform path (PLAN.md §4.7).
//
// They are faithful ports of AGG's span_image_filter_rgba.h templates
// instantiated for agg::rgba32 (value_type=float, calc_type/long_type=double,
// downshift = divide, full_value = 1.0). One intentional deviation from the
// shared C++ template applies to the bilinear generator: see the note on
// SpanImageFilterRGBA32Bilinear.Generate.
package span

import (
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/transform"
)

// RGBA32SourceInterface is the float source contract used by the RGBA32 image
// filters. It mirrors RGBASourceInterface but exposes float32 channel slices.
type RGBA32SourceInterface interface {
	SourceInterface
	OrderType() color.ColorOrder
	Span(x, y, length int) []float32
	NextX() []float32
	NextY() []float32
	RowPtr(y int) []float32
}

// imageFilterScaleF and imageSubpixelScaleSqF are the float divisors equivalent
// to AGG's color_type::downshift for rgba32 (a / (1 << n)).
const (
	imageFilterScaleF     = float64(image.ImageFilterScale)
	imageSubpixelScaleSqF = float64(image.ImageSubpixelScale) * float64(image.ImageSubpixelScale)
)

// clampUnitF clamps the lower bound to 0 and the alpha-derived upper bounds the
// way AGG's integer generators do, but against full_value()=1.0 for rgba32.
func clampChannelLow(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

// SpanImageFilterRGBA32NN ----------------------------------------------------

// SpanImageFilterRGBA32NN implements nearest-neighbor float RGBA filtering.
// Port of agg::span_image_filter_rgba_nn for rgba32.
type SpanImageFilterRGBA32NN[Source RGBA32SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBA32NNWithParams creates a nearest-neighbor float filter.
func NewSpanImageFilterRGBA32NNWithParams[Source RGBA32SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
) *SpanImageFilterRGBA32NN[Source, Interpolator] {
	return &SpanImageFilterRGBA32NN[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, nil),
	}
}

// Generate fills span with nearest-neighbor samples.
func (sif *SpanImageFilterRGBA32NN[Source, Interpolator]) Generate(span []color.RGBA32[color.Linear], x, y int) {
	length := len(span)
	if length == 0 {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)
	order := sif.base.source.OrderType()

	for i := 0; i < length; i++ {
		xHr, yHr := sif.base.interpolator.Coordinates()
		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		fgPtr := sif.base.source.Span(xLr, yLr, 1)
		if len(fgPtr) >= 4 {
			span[i] = color.RGBA32[color.Linear]{
				R: fgPtr[order.R],
				G: fgPtr[order.G],
				B: fgPtr[order.B],
				A: fgPtr[order.A],
			}
		} else {
			span[i] = color.RGBA32[color.Linear]{}
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGBA32Bilinear ----------------------------------------------

// SpanImageFilterRGBA32Bilinear implements bilinear float RGBA filtering.
// Port of agg::span_image_filter_rgba_bilinear for rgba32.
type SpanImageFilterRGBA32Bilinear[Source RGBA32SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBA32BilinearWithParams creates a bilinear float filter.
func NewSpanImageFilterRGBA32BilinearWithParams[Source RGBA32SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
) *SpanImageFilterRGBA32Bilinear[Source, Interpolator] {
	return &SpanImageFilterRGBA32Bilinear[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, nil),
	}
}

// Generate fills span with bilinear samples.
//
// Deviation from AGG's shared template: the C++ bilinear generator seeds the
// accumulator with image_subpixel_scale^2/2 as an integer rounding bias before
// the final downshift. For 8-bit channels that bias is ~0.5 of one 0..255 unit
// (negligible rounding). For rgba32 the channels are floats in [0,1], so the
// same bias would add a full +0.5 to every channel and corrupt the output.
// Float color does not quantize, so the correct float equivalent of "round to
// nearest" is no bias at all: this is a plain weighted average. The result then
// matches the 8-bit path within quantization tolerance, which is the parity goal.
func (sif *SpanImageFilterRGBA32Bilinear[Source, Interpolator]) Generate(span []color.RGBA32[color.Linear], x, y int) {
	length := len(span)
	if length == 0 {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)

	for i := 0; i < length; i++ {
		xHr, yHr := sif.base.interpolator.Coordinates()

		xHr -= sif.base.FilterDxInt()
		yHr -= sif.base.FilterDyInt()

		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		var fg [4]float64

		xHr &= image.ImageSubpixelMask
		yHr &= image.ImageSubpixelMask

		// Top-left sample.
		fgPtr := sif.base.source.Span(xLr, yLr, 2)
		weight := float64((image.ImageSubpixelScale - xHr) * (image.ImageSubpixelScale - yHr))
		accumulateF(&fg, fgPtr, weight)

		// Top-right sample.
		fgPtr = sif.base.source.NextX()
		weight = float64(xHr * (image.ImageSubpixelScale - yHr))
		accumulateF(&fg, fgPtr, weight)

		// Bottom-left sample.
		fgPtr = sif.base.source.NextY()
		weight = float64((image.ImageSubpixelScale - xHr) * yHr)
		accumulateF(&fg, fgPtr, weight)

		// Bottom-right sample.
		fgPtr = sif.base.source.NextX()
		weight = float64(xHr * yHr)
		accumulateF(&fg, fgPtr, weight)

		span[i] = color.RGBA32[color.Linear]{
			R: float32(fg[0] / imageSubpixelScaleSqF),
			G: float32(fg[1] / imageSubpixelScaleSqF),
			B: float32(fg[2] / imageSubpixelScaleSqF),
			A: float32(fg[3] / imageSubpixelScaleSqF),
		}

		sif.base.interpolator.Next()
	}
}

// accumulateF adds weight*channel for each of the 4 channels if the sample is valid.
func accumulateF(fg *[4]float64, fgPtr []float32, weight float64) {
	if len(fgPtr) >= 4 {
		fg[0] += weight * float64(fgPtr[0])
		fg[1] += weight * float64(fgPtr[1])
		fg[2] += weight * float64(fgPtr[2])
		fg[3] += weight * float64(fgPtr[3])
	}
}

// finalizeFilteredF applies AGG's per-variant clamping for the LUT-based
// generators: clamp negatives to 0, clamp alpha to full_value()=1, then clamp
// each RGB channel to alpha (premultiplied invariant).
func finalizeFilteredF(fg [4]float64) color.RGBA32[color.Linear] {
	fg[0] = clampChannelLow(fg[0])
	fg[1] = clampChannelLow(fg[1])
	fg[2] = clampChannelLow(fg[2])
	fg[3] = clampChannelLow(fg[3])

	if fg[3] > 1 {
		fg[3] = 1
	}
	if fg[0] > fg[3] {
		fg[0] = fg[3]
	}
	if fg[1] > fg[3] {
		fg[1] = fg[3]
	}
	if fg[2] > fg[3] {
		fg[2] = fg[3]
	}

	return color.RGBA32[color.Linear]{
		R: float32(fg[0]),
		G: float32(fg[1]),
		B: float32(fg[2]),
		A: float32(fg[3]),
	}
}

// SpanImageFilterRGBA32_2x2 --------------------------------------------------

// SpanImageFilterRGBA32_2x2 implements 2x2 float RGBA filtering with a LUT.
// Port of agg::span_image_filter_rgba_2x2 for rgba32.
type SpanImageFilterRGBA32_2x2[Source RGBA32SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBA32_2x2WithParams creates a 2x2 float filter.
func NewSpanImageFilterRGBA32_2x2WithParams[Source RGBA32SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilterRGBA32_2x2[Source, Interpolator] {
	return &SpanImageFilterRGBA32_2x2[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, filter),
	}
}

// Generate fills span using a 2x2 filter kernel.
func (sif *SpanImageFilterRGBA32_2x2[Source, Interpolator]) Generate(span []color.RGBA32[color.Linear], x, y int) {
	length := len(span)
	if length == 0 || sif.base.filter == nil {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)

	weightArray := sif.base.filter.WeightArray()
	if weightArray == nil {
		return
	}

	offset := (sif.base.filter.Diameter()/2 - 1) << image.ImageSubpixelShift

	for i := 0; i < length; i++ {
		xHr, yHr := sif.base.interpolator.Coordinates()

		xHr -= sif.base.FilterDxInt()
		yHr -= sif.base.FilterDyInt()

		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		var fg [4]float64

		xHr &= image.ImageSubpixelMask
		yHr &= image.ImageSubpixelMask

		// Sample 1 (top-left).
		fgPtr := sif.base.source.Span(xLr, yLr, 2)
		weight := float64((int(weightArray[xHr+image.ImageSubpixelScale+offset])*
			int(weightArray[yHr+image.ImageSubpixelScale+offset]) +
			image.ImageFilterScale/2) >> image.ImageFilterShift)
		accumulateF(&fg, fgPtr, weight)

		// Sample 2 (top-right).
		fgPtr = sif.base.source.NextX()
		weight = float64((int(weightArray[xHr+offset])*
			int(weightArray[yHr+image.ImageSubpixelScale+offset]) +
			image.ImageFilterScale/2) >> image.ImageFilterShift)
		accumulateF(&fg, fgPtr, weight)

		// Sample 3 (bottom-left).
		fgPtr = sif.base.source.NextY()
		weight = float64((int(weightArray[xHr+image.ImageSubpixelScale+offset])*
			int(weightArray[yHr+offset]) +
			image.ImageFilterScale/2) >> image.ImageFilterShift)
		accumulateF(&fg, fgPtr, weight)

		// Sample 4 (bottom-right).
		fgPtr = sif.base.source.NextX()
		weight = float64((int(weightArray[xHr+offset])*
			int(weightArray[yHr+offset]) +
			image.ImageFilterScale/2) >> image.ImageFilterShift)
		accumulateF(&fg, fgPtr, weight)

		fg[0] /= imageFilterScaleF
		fg[1] /= imageFilterScaleF
		fg[2] /= imageFilterScaleF
		fg[3] /= imageFilterScaleF

		// The 2x2 generator clamps alpha and the RGB-vs-alpha invariant only.
		if fg[3] > 1 {
			fg[3] = 1
		}
		if fg[0] > fg[3] {
			fg[0] = fg[3]
		}
		if fg[1] > fg[3] {
			fg[1] = fg[3]
		}
		if fg[2] > fg[3] {
			fg[2] = fg[3]
		}

		span[i] = color.RGBA32[color.Linear]{
			R: float32(fg[0]),
			G: float32(fg[1]),
			B: float32(fg[2]),
			A: float32(fg[3]),
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGBA32 ------------------------------------------------------

// SpanImageFilterRGBA32 implements general float RGBA filtering with a LUT of
// arbitrary diameter. Port of agg::span_image_filter_rgba for rgba32.
type SpanImageFilterRGBA32[Source RGBA32SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBA32WithParams creates a general float filter.
func NewSpanImageFilterRGBA32WithParams[Source RGBA32SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilterRGBA32[Source, Interpolator] {
	return &SpanImageFilterRGBA32[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, filter),
	}
}

// Generate fills span using a configurable filter kernel.
func (sif *SpanImageFilterRGBA32[Source, Interpolator]) Generate(span []color.RGBA32[color.Linear], x, y int) {
	length := len(span)
	if length == 0 || sif.base.filter == nil {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)

	diameter := sif.base.filter.Diameter()
	start := sif.base.filter.Start()
	weightArray := sif.base.filter.WeightArray()
	if weightArray == nil {
		return
	}

	for i := 0; i < length; i++ {
		xHr, yHr := sif.base.interpolator.Coordinates()

		xHr -= sif.base.FilterDxInt()
		yHr -= sif.base.FilterDyInt()

		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		var fg [4]float64

		xFract := xHr & image.ImageSubpixelMask
		yCount := diameter

		yHr = image.ImageSubpixelMask - (yHr & image.ImageSubpixelMask)

		fgPtr := sif.base.source.Span(xLr+start, yLr+start, diameter)

		for yCount > 0 {
			xCount := diameter
			weightY := weightArray[yHr]
			xHr = image.ImageSubpixelMask - xFract

			for xCount > 0 {
				weight := float64((int(weightY)*int(weightArray[xHr]) + image.ImageFilterScale/2) >> image.ImageFilterShift)
				accumulateF(&fg, fgPtr, weight)

				xCount--
				if xCount == 0 {
					break
				}
				xHr += image.ImageSubpixelScale
				fgPtr = sif.base.source.NextX()
			}

			yCount--
			if yCount == 0 {
				break
			}
			yHr += image.ImageSubpixelScale
			fgPtr = sif.base.source.NextY()
		}

		fg[0] /= imageFilterScaleF
		fg[1] /= imageFilterScaleF
		fg[2] /= imageFilterScaleF
		fg[3] /= imageFilterScaleF

		span[i] = finalizeFilteredF(fg)

		sif.base.interpolator.Next()
	}
}

// SpanImageResampleRGBA32Affine ----------------------------------------------

// SpanImageResampleRGBA32Affine provides affine resampling with automatic scale
// detection for float RGBA. Port of agg::span_image_resample_rgba_affine for rgba32.
type SpanImageResampleRGBA32Affine[Source RGBA32SourceInterface] struct {
	base *SpanImageResampleAffine[Source]
}

// NewSpanImageResampleRGBA32AffineWithParams creates an affine float resampler.
func NewSpanImageResampleRGBA32AffineWithParams[Source RGBA32SourceInterface](
	src Source,
	interpolator *SpanInterpolatorLinear[*transform.TransAffine],
	filter *image.ImageFilterLUT,
) *SpanImageResampleRGBA32Affine[Source] {
	return &SpanImageResampleRGBA32Affine[Source]{
		base: NewSpanImageResampleAffineWithParams(src, interpolator, filter),
	}
}

// Prepare derives the affine kernel radii from the transform.
func (sir *SpanImageResampleRGBA32Affine[Source]) Prepare() {
	sir.base.Prepare()
}

// Blur sets isotropic blur for the resampler.
func (sir *SpanImageResampleRGBA32Affine[Source]) Blur(v float64) {
	sir.base.Blur(v)
}

// Generate fills span using affine resampling.
func (sir *SpanImageResampleRGBA32Affine[Source]) Generate(span []color.RGBA32[color.Linear], x, y int) {
	length := len(span)
	if length == 0 {
		return
	}

	baseFilter := sir.base.base
	baseFilter.interpolator.Begin(float64(x)+baseFilter.FilterDxDbl(), float64(y)+baseFilter.FilterDyDbl(), length)

	filter := baseFilter.Filter()
	if filter == nil {
		bilinear := NewSpanImageFilterRGBA32BilinearWithParams(baseFilter.Source(), baseFilter.Interpolator())
		bilinear.Generate(span, x, y)
		return
	}

	diameter := filter.Diameter()
	filterScale := diameter << image.ImageSubpixelShift
	radiusX := (diameter * sir.base.RX()) >> 1
	radiusY := (diameter * sir.base.RY()) >> 1
	lenXLr := (diameter*sir.base.RX() + image.ImageSubpixelMask) >> image.ImageSubpixelShift

	weightArray := filter.WeightArray()
	order := baseFilter.source.OrderType()

	for i := 0; i < length; i++ {
		sx, sy := baseFilter.interpolator.Coordinates()

		sx += baseFilter.FilterDxInt() - radiusX
		sy += baseFilter.FilterDyInt() - radiusY

		var fg [4]float64

		yLr := sy >> image.ImageSubpixelShift
		yHr := ((image.ImageSubpixelMask - (sy & image.ImageSubpixelMask)) * sir.base.RYInv()) >> image.ImageSubpixelShift
		totalWeight := 0
		xLr := sx >> image.ImageSubpixelShift
		xHr := ((image.ImageSubpixelMask - (sx & image.ImageSubpixelMask)) * sir.base.RXInv()) >> image.ImageSubpixelShift

		xHr2 := xHr
		fgPtr := sir.base.Source().Span(xLr, yLr, lenXLr)

		for yHr < len(weightArray) {
			weightY := int(weightArray[yHr])
			xHr = xHr2

			for xHr < len(weightArray) {
				weight := (weightY*int(weightArray[xHr]) + image.ImageFilterScale/2) >> image.ImageFilterShift

				if len(fgPtr) >= 4 {
					fg[0] += float64(fgPtr[order.R]) * float64(weight)
					fg[1] += float64(fgPtr[order.G]) * float64(weight)
					fg[2] += float64(fgPtr[order.B]) * float64(weight)
					fg[3] += float64(fgPtr[order.A]) * float64(weight)
				}
				totalWeight += weight
				xHr += sir.base.RXInv()

				if xHr >= filterScale {
					break
				}
				fgPtr = sir.base.Source().NextX()
			}

			yHr += sir.base.RYInv()
			if yHr >= filterScale {
				break
			}
			fgPtr = sir.base.Source().NextY()
		}

		if totalWeight > 0 {
			tw := float64(totalWeight)
			fg[0] /= tw
			fg[1] /= tw
			fg[2] /= tw
			fg[3] /= tw
		}

		span[i] = finalizeFilteredF(fg)

		baseFilter.interpolator.Next()
	}
}
