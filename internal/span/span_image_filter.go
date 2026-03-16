package span

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// SourceInterface is the minimal source contract used by the base image-filter
// generators before format-specific accessors add sampling methods.
type SourceInterface interface {
	Width() int
	Height() int
}

// SpanImageFilter is the Go equivalent of AGG's span_image_filter base class.
// It stores the source, coordinate interpolator, filter LUT, and the
// half-pixel sampling offsets shared by the concrete image-filter generators.
type SpanImageFilter[Source SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	source       Source
	interpolator Interpolator
	filter       *image.ImageFilterLUT
	dxDbl        float64
	dyDbl        float64
	dxInt        int
	dyInt        int
}

// NewSpanImageFilter creates an unattached image-filter base with AGG's default
// half-pixel filter offsets.
func NewSpanImageFilter[Source SourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilter[Source, Interpolator] {
	return &SpanImageFilter[Source, Interpolator]{
		dxDbl: 0.5,
		dyDbl: 0.5,
		dxInt: image.ImageSubpixelScale / 2,
		dyInt: image.ImageSubpixelScale / 2,
	}
}

// NewSpanImageFilterWithParams creates an attached image-filter base.
func NewSpanImageFilterWithParams[Source SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilter[Source, Interpolator] {
	return &SpanImageFilter[Source, Interpolator]{
		source:       src,
		interpolator: interpolator,
		filter:       filter,
		dxDbl:        0.5,
		dyDbl:        0.5,
		dxInt:        image.ImageSubpixelScale / 2,
		dyInt:        image.ImageSubpixelScale / 2,
	}
}

// Attach replaces the source image.
func (sif *SpanImageFilter[Source, Interpolator]) Attach(src Source) {
	sif.source = src
}

// Source returns the attached source image.
func (sif *SpanImageFilter[Source, Interpolator]) Source() Source {
	return sif.source
}

// Filter returns the attached filter LUT.
func (sif *SpanImageFilter[Source, Interpolator]) Filter() *image.ImageFilterLUT {
	return sif.filter
}

// FilterDxInt returns the X offset in image-subpixel units.
func (sif *SpanImageFilter[Source, Interpolator]) FilterDxInt() int {
	return sif.dxInt
}

// FilterDyInt returns the Y offset in image-subpixel units.
func (sif *SpanImageFilter[Source, Interpolator]) FilterDyInt() int {
	return sif.dyInt
}

// FilterDxDbl returns the X offset in user-space units.
func (sif *SpanImageFilter[Source, Interpolator]) FilterDxDbl() float64 {
	return sif.dxDbl
}

// FilterDyDbl returns the Y offset in user-space units.
func (sif *SpanImageFilter[Source, Interpolator]) FilterDyDbl() float64 {
	return sif.dyDbl
}

// SetInterpolator replaces the coordinate interpolator.
func (sif *SpanImageFilter[Source, Interpolator]) SetInterpolator(interpolator Interpolator) {
	sif.interpolator = interpolator
}

// SetFilter replaces the filter LUT.
func (sif *SpanImageFilter[Source, Interpolator]) SetFilter(filter *image.ImageFilterLUT) {
	sif.filter = filter
}

// FilterOffset sets the shared sampling offset used by the concrete generators.
func (sif *SpanImageFilter[Source, Interpolator]) FilterOffset(dx, dy float64) {
	sif.dxDbl = dx
	sif.dyDbl = dy
	sif.dxInt = basics.IRound(dx * float64(image.ImageSubpixelScale))
	sif.dyInt = basics.IRound(dy * float64(image.ImageSubpixelScale))
}

// FilterOffsetUniform sets the same sampling offset on both axes.
func (sif *SpanImageFilter[Source, Interpolator]) FilterOffsetUniform(d float64) {
	sif.FilterOffset(d, d)
}

// Interpolator returns the attached coordinate interpolator.
func (sif *SpanImageFilter[Source, Interpolator]) Interpolator() Interpolator {
	return sif.interpolator
}

// Prepare is a no-op on the base type.
func (sif *SpanImageFilter[Source, Interpolator]) Prepare() {
}

// SpanImageResampleAffine is the Go equivalent of AGG's
// span_image_resample_affine. It inspects the affine transform to derive kernel
// radii and inverse scale factors for image resampling.
type SpanImageResampleAffine[Source SourceInterface] struct {
	base       *SpanImageFilter[Source, *SpanInterpolatorLinear[*transform.TransAffine]]
	scaleLimit float64
	blurX      float64
	blurY      float64
	rx         int
	ry         int
	rxInv      int
	ryInv      int
}

// NewSpanImageResampleAffine creates an unattached affine resampler.
func NewSpanImageResampleAffine[Source SourceInterface]() *SpanImageResampleAffine[Source] {
	baseFilter := NewSpanImageFilter[Source, *SpanInterpolatorLinear[*transform.TransAffine]]()
	return &SpanImageResampleAffine[Source]{
		base:       baseFilter,
		scaleLimit: 200.0,
		blurX:      1.0,
		blurY:      1.0,
	}
}

// NewSpanImageResampleAffineWithParams creates an attached affine resampler.
func NewSpanImageResampleAffineWithParams[Source SourceInterface](
	src Source,
	interpolator *SpanInterpolatorLinear[*transform.TransAffine],
	filter *image.ImageFilterLUT,
) *SpanImageResampleAffine[Source] {
	baseFilter := &SpanImageFilter[Source, *SpanInterpolatorLinear[*transform.TransAffine]]{
		source:       src,
		interpolator: interpolator,
		filter:       filter,
		dxDbl:        0.5,
		dyDbl:        0.5,
		dxInt:        image.ImageSubpixelScale / 2,
		dyInt:        image.ImageSubpixelScale / 2,
	}
	return &SpanImageResampleAffine[Source]{
		base:       baseFilter,
		scaleLimit: 200.0,
		blurX:      1.0,
		blurY:      1.0,
	}
}

// ScaleLimit returns the current clamp applied to derived scale factors.
func (sira *SpanImageResampleAffine[Source]) ScaleLimit() int {
	return int(basics.URound(sira.scaleLimit))
}

// SetScaleLimit updates the scale clamp.
func (sira *SpanImageResampleAffine[Source]) SetScaleLimit(v int) {
	sira.scaleLimit = float64(v)
}

// BlurX returns the X blur factor.
func (sira *SpanImageResampleAffine[Source]) BlurX() float64 {
	return sira.blurX
}

// BlurY returns the Y blur factor.
func (sira *SpanImageResampleAffine[Source]) BlurY() float64 {
	return sira.blurY
}

// SetBlurX sets the X blur factor.
func (sira *SpanImageResampleAffine[Source]) SetBlurX(v float64) {
	sira.blurX = v
}

// SetBlurY sets the Y blur factor.
func (sira *SpanImageResampleAffine[Source]) SetBlurY(v float64) {
	sira.blurY = v
}

// Blur sets the same blur factor on both axes.
func (sira *SpanImageResampleAffine[Source]) Blur(v float64) {
	sira.blurX = v
	sira.blurY = v
}

// SetInterpolator replaces the affine interpolator.
func (sira *SpanImageResampleAffine[Source]) SetInterpolator(interpolator *SpanInterpolatorLinear[*transform.TransAffine]) {
	sira.base.SetInterpolator(interpolator)
}

// Source returns the current source image.
func (sira *SpanImageResampleAffine[Source]) Source() Source {
	return sira.base.Source()
}

// Filter returns the current filter LUT.
func (sira *SpanImageResampleAffine[Source]) Filter() *image.ImageFilterLUT {
	return sira.base.Filter()
}

// Interpolator returns the affine coordinate interpolator.
func (sira *SpanImageResampleAffine[Source]) Interpolator() *SpanInterpolatorLinear[*transform.TransAffine] {
	return sira.base.Interpolator()
}

// Prepare mirrors AGG's span_image_resample_affine::prepare by extracting the
// absolute affine scale, clamping it, applying blur, and caching fixed-point
// radii and inverse radii for the concrete resamplers.
func (sira *SpanImageResampleAffine[Source]) Prepare() {
	if sira.base.interpolator == nil {
		return
	}

	// Get the transformer from the interpolator
	transformer := sira.base.interpolator.Transformer()
	if transformer == nil {
		return
	}

	// Extract scaling factors from the affine transformation
	scaleX, scaleY := transformer.GetScalingAbs()

	// Limit the combined scale to prevent excessive memory usage
	scaleXY := scaleX * scaleY
	if scaleXY > sira.scaleLimit {
		scaleX = scaleX * sira.scaleLimit / scaleXY
		scaleY = scaleY * sira.scaleLimit / scaleXY
	}

	// Ensure minimum scale of 1
	if scaleX < 1 {
		scaleX = 1
	}
	if scaleY < 1 {
		scaleY = 1
	}

	// Apply individual scale limits
	if scaleX > sira.scaleLimit {
		scaleX = sira.scaleLimit
	}
	if scaleY > sira.scaleLimit {
		scaleY = sira.scaleLimit
	}

	// Apply blur factors
	scaleX *= sira.blurX
	scaleY *= sira.blurY

	// Ensure minimum scale after blur
	if scaleX < 1 {
		scaleX = 1
	}
	if scaleY < 1 {
		scaleY = 1
	}

	// Calculate integer scaling factors in subpixel precision
	sira.rx = int(basics.URound(scaleX * float64(image.ImageSubpixelScale)))
	sira.rxInv = int(basics.URound(1.0 / scaleX * float64(image.ImageSubpixelScale)))

	sira.ry = int(basics.URound(scaleY * float64(image.ImageSubpixelScale)))
	sira.ryInv = int(basics.URound(1.0 / scaleY * float64(image.ImageSubpixelScale)))
}

// RX returns the X scaling factor in subpixel precision.
func (sira *SpanImageResampleAffine[Source]) RX() int {
	return sira.rx
}

// RY returns the Y scaling factor in subpixel precision.
func (sira *SpanImageResampleAffine[Source]) RY() int {
	return sira.ry
}

// RXInv returns the inverse X scaling factor in subpixel precision.
func (sira *SpanImageResampleAffine[Source]) RXInv() int {
	return sira.rxInv
}

// RYInv returns the inverse Y scaling factor in subpixel precision.
func (sira *SpanImageResampleAffine[Source]) RYInv() int {
	return sira.ryInv
}

// SpanImageResample provides general image resampling with configurable interpolation.
// This is a port of AGG's span_image_resample template class.
type SpanImageResample[Source SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base       *SpanImageFilter[Source, Interpolator]
	scaleLimit int
	blurX      int
	blurY      int
}

// NewSpanImageResample creates a new general resampling span filter.
func NewSpanImageResample[Source SourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageResample[Source, Interpolator] {
	baseFilter := NewSpanImageFilter[Source, Interpolator]()
	return &SpanImageResample[Source, Interpolator]{
		base:       baseFilter,
		scaleLimit: 20,
		blurX:      image.ImageSubpixelScale,
		blurY:      image.ImageSubpixelScale,
	}
}

// NewSpanImageResampleWithParams creates a new general resampling filter with parameters.
func NewSpanImageResampleWithParams[Source SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageResample[Source, Interpolator] {
	baseFilter := &SpanImageFilter[Source, Interpolator]{
		source:       src,
		interpolator: interpolator,
		filter:       filter,
		dxDbl:        0.5,
		dyDbl:        0.5,
		dxInt:        image.ImageSubpixelScale / 2,
		dyInt:        image.ImageSubpixelScale / 2,
	}
	return &SpanImageResample[Source, Interpolator]{
		base:       baseFilter,
		scaleLimit: 20,
		blurX:      image.ImageSubpixelScale,
		blurY:      image.ImageSubpixelScale,
	}
}

// ScaleLimit returns the current scale limit.
func (sir *SpanImageResample[Source, Interpolator]) ScaleLimit() int {
	return sir.scaleLimit
}

// SetScaleLimit sets the scale limit.
func (sir *SpanImageResample[Source, Interpolator]) SetScaleLimit(v int) {
	sir.scaleLimit = v
}

// BlurX returns the X blur factor.
func (sir *SpanImageResample[Source, Interpolator]) BlurX() float64 {
	return float64(sir.blurX) / float64(image.ImageSubpixelScale)
}

// BlurY returns the Y blur factor.
func (sir *SpanImageResample[Source, Interpolator]) BlurY() float64 {
	return float64(sir.blurY) / float64(image.ImageSubpixelScale)
}

// SetBlurX sets the X blur factor.
func (sir *SpanImageResample[Source, Interpolator]) SetBlurX(v float64) {
	sir.blurX = int(basics.URound(v * float64(image.ImageSubpixelScale)))
}

// SetBlurY sets the Y blur factor.
func (sir *SpanImageResample[Source, Interpolator]) SetBlurY(v float64) {
	sir.blurY = int(basics.URound(v * float64(image.ImageSubpixelScale)))
}

// Blur sets both X and Y blur factors to the same value.
func (sir *SpanImageResample[Source, Interpolator]) Blur(v float64) {
	blur := int(basics.URound(v * float64(image.ImageSubpixelScale)))
	sir.blurX = blur
	sir.blurY = blur
}

// Source returns the source from the base filter.
func (sir *SpanImageResample[Source, Interpolator]) Source() Source {
	return sir.base.Source()
}

// Filter returns the filter from the base filter.
func (sir *SpanImageResample[Source, Interpolator]) Filter() *image.ImageFilterLUT {
	return sir.base.Filter()
}

// Interpolator returns the interpolator from the base filter.
func (sir *SpanImageResample[Source, Interpolator]) Interpolator() Interpolator {
	return sir.base.Interpolator()
}

// AdjustScale adjusts scaling factors according to scale limits and blur factors.
// This is equivalent to AGG's adjust_scale method.
func (sir *SpanImageResample[Source, Interpolator]) AdjustScale(rx, ry *int) {
	// Ensure minimum scale
	if *rx < image.ImageSubpixelScale {
		*rx = image.ImageSubpixelScale
	}
	if *ry < image.ImageSubpixelScale {
		*ry = image.ImageSubpixelScale
	}

	// Apply scale limits
	if *rx > image.ImageSubpixelScale*sir.scaleLimit {
		*rx = image.ImageSubpixelScale * sir.scaleLimit
	}
	if *ry > image.ImageSubpixelScale*sir.scaleLimit {
		*ry = image.ImageSubpixelScale * sir.scaleLimit
	}

	// Apply blur factors
	*rx = (*rx * sir.blurX) >> image.ImageSubpixelShift
	*ry = (*ry * sir.blurY) >> image.ImageSubpixelShift

	// Ensure minimum scale after blur
	if *rx < image.ImageSubpixelScale {
		*rx = image.ImageSubpixelScale
	}
	if *ry < image.ImageSubpixelScale {
		*ry = image.ImageSubpixelScale
	}
}
