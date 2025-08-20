// Package span provides grayscale image filtering span generation functionality for AGG.
// This implements a port of AGG's agg_span_image_filter_gray.h functionality.
package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/image"
	"agg_go/internal/transform"
)

// GraySourceInterface defines the interface for grayscale image sources.
// This extends the basic SourceInterface with grayscale-specific methods.
type GraySourceInterface interface {
	SourceInterface
	// ColorType returns the grayscale color type
	ColorType() string
	// Span returns pixel data starting at (x, y) with given length
	Span(x, y, length int) []basics.Int8u
	// NextX advances to the next pixel in current span
	NextX() []basics.Int8u
	// NextY advances to the next row at original x position
	NextY() []basics.Int8u
	// RowPtr returns pointer to row data
	RowPtr(y int) []basics.Int8u
}

// SpanImageFilterGrayNN implements nearest neighbor filtering for grayscale images.
// This is a port of AGG's span_image_filter_gray_nn template class.
type SpanImageFilterGrayNN[Source GraySourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterGrayNN creates a new grayscale nearest neighbor filter.
func NewSpanImageFilterGrayNN[Source GraySourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterGrayNN[Source, Interpolator] {
	return &SpanImageFilterGrayNN[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterGrayNNWithParams creates a new grayscale nearest neighbor filter with parameters.
func NewSpanImageFilterGrayNNWithParams[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
) *SpanImageFilterGrayNN[Source, Interpolator] {
	return &SpanImageFilterGrayNN[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, nil),
	}
}

// Generate generates a span of grayscale pixels using nearest neighbor filtering.
func (sifnn *SpanImageFilterGrayNN[Source, Interpolator]) Generate(span []color.Gray8[color.Linear], x, y, length int) {
	sifnn.base.interpolator.Begin(float64(x)+sifnn.base.FilterDxDbl(), float64(y)+sifnn.base.FilterDyDbl(), length)

	for i := 0; i < length; i++ {
		sx, sy := sifnn.base.interpolator.Coordinates()

		// Convert to image coordinates (remove subpixel precision)
		imgX := sx >> image.ImageSubpixelShift
		imgY := sy >> image.ImageSubpixelShift

		// Get the pixel value from source
		pixelData := sifnn.base.source.Span(imgX, imgY, 1)
		if len(pixelData) > 0 {
			span[i] = color.NewGray8WithAlpha[color.Linear](pixelData[0], color.Gray8FullValue())
		} else {
			span[i] = color.NewGray8[color.Linear](0) // Default to black if out of bounds
		}

		sifnn.base.interpolator.Next()
	}
}

// SpanImageFilterGrayBilinear implements bilinear interpolation for grayscale images.
// This is a port of AGG's span_image_filter_gray_bilinear template class.
type SpanImageFilterGrayBilinear[Source GraySourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterGrayBilinear creates a new grayscale bilinear filter.
func NewSpanImageFilterGrayBilinear[Source GraySourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterGrayBilinear[Source, Interpolator] {
	return &SpanImageFilterGrayBilinear[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterGrayBilinearWithParams creates a new grayscale bilinear filter with parameters.
func NewSpanImageFilterGrayBilinearWithParams[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
) *SpanImageFilterGrayBilinear[Source, Interpolator] {
	return &SpanImageFilterGrayBilinear[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, nil),
	}
}

// Generate generates a span of grayscale pixels using bilinear interpolation.
func (sifb *SpanImageFilterGrayBilinear[Source, Interpolator]) Generate(span []color.Gray8[color.Linear], x, y, length int) {
	sifb.base.interpolator.Begin(float64(x)+sifb.base.FilterDxDbl(), float64(y)+sifb.base.FilterDyDbl(), length)

	for i := 0; i < length; i++ {
		xHr, yHr := sifb.base.interpolator.Coordinates()

		// Apply filter offset
		xHr -= sifb.base.FilterDxInt()
		yHr -= sifb.base.FilterDyInt()

		// Extract integer and fractional parts
		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		xHr &= image.ImageSubpixelMask
		yHr &= image.ImageSubpixelMask

		// Get 2x2 pixel neighborhood
		fgPtr := sifb.base.source.Span(xLr, yLr, 2)

		var fg int32 = 0

		if len(fgPtr) >= 1 {
			// Top-left pixel
			fg += int32(fgPtr[0]) * int32(image.ImageSubpixelScale-xHr) * int32(image.ImageSubpixelScale-yHr)

			// Top-right pixel
			fgPtr = sifb.base.source.NextX()
			if len(fgPtr) >= 1 {
				fg += int32(fgPtr[0]) * int32(xHr) * int32(image.ImageSubpixelScale-yHr)
			}

			// Bottom-left pixel
			fgPtr = sifb.base.source.NextY()
			if len(fgPtr) >= 1 {
				fg += int32(fgPtr[0]) * int32(image.ImageSubpixelScale-xHr) * int32(yHr)
			}

			// Bottom-right pixel
			fgPtr = sifb.base.source.NextX()
			if len(fgPtr) >= 1 {
				fg += int32(fgPtr[0]) * int32(xHr) * int32(yHr)
			}
		}

		// Downshift to get final value
		fg >>= image.ImageSubpixelShift * 2
		if fg < 0 {
			fg = 0
		}
		if fg > int32(color.Gray8FullValue()) {
			fg = int32(color.Gray8FullValue())
		}

		span[i] = color.NewGray8WithAlpha[color.Linear](basics.Int8u(fg), color.Gray8FullValue())
		sifb.base.interpolator.Next()
	}
}

// SpanImageFilterGrayBilinearClip implements bilinear interpolation with background color clipping.
// This is a port of AGG's span_image_filter_gray_bilinear_clip template class.
type SpanImageFilterGrayBilinearClip[Source GraySourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base      *SpanImageFilter[Source, Interpolator]
	backColor color.Gray8[color.Linear]
}

// NewSpanImageFilterGrayBilinearClip creates a new grayscale bilinear filter with clipping.
func NewSpanImageFilterGrayBilinearClip[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	backColor color.Gray8[color.Linear],
) *SpanImageFilterGrayBilinearClip[Source, Interpolator] {
	return &SpanImageFilterGrayBilinearClip[Source, Interpolator]{
		base:      NewSpanImageFilter[Source, Interpolator](),
		backColor: backColor,
	}
}

// NewSpanImageFilterGrayBilinearClipWithParams creates a new grayscale bilinear filter with clipping and parameters.
func NewSpanImageFilterGrayBilinearClipWithParams[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	backColor color.Gray8[color.Linear],
	interpolator Interpolator,
) *SpanImageFilterGrayBilinearClip[Source, Interpolator] {
	return &SpanImageFilterGrayBilinearClip[Source, Interpolator]{
		base:      NewSpanImageFilterWithParams(src, interpolator, nil),
		backColor: backColor,
	}
}

// BackgroundColor returns the background color.
func (sifbc *SpanImageFilterGrayBilinearClip[Source, Interpolator]) BackgroundColor() color.Gray8[color.Linear] {
	return sifbc.backColor
}

// SetBackgroundColor sets the background color.
func (sifbc *SpanImageFilterGrayBilinearClip[Source, Interpolator]) SetBackgroundColor(c color.Gray8[color.Linear]) {
	sifbc.backColor = c
}

// Generate generates a span of grayscale pixels using bilinear interpolation with clipping.
func (sifbc *SpanImageFilterGrayBilinearClip[Source, Interpolator]) Generate(span []color.Gray8[color.Linear], x, y, length int) {
	sifbc.base.interpolator.Begin(float64(x)+sifbc.base.FilterDxDbl(), float64(y)+sifbc.base.FilterDyDbl(), length)

	backV := sifbc.backColor.V
	backA := sifbc.backColor.A

	maxx := sifbc.base.source.Width() - 1
	maxy := sifbc.base.source.Height() - 1

	for i := 0; i < length; i++ {
		xHr, yHr := sifbc.base.interpolator.Coordinates()

		// Apply filter offset
		xHr -= sifbc.base.FilterDxInt()
		yHr -= sifbc.base.FilterDyInt()

		// Extract integer and fractional parts
		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		var fg int32 = 0
		var srcAlpha int32 = 0

		if xLr >= 0 && yLr >= 0 && xLr < maxx && yLr < maxy {
			// Fast path - all pixels are within bounds
			xHr &= image.ImageSubpixelMask
			yHr &= image.ImageSubpixelMask

			fgPtr := sifbc.base.source.RowPtr(yLr)
			if xLr < len(fgPtr) {
				fg += int32(fgPtr[xLr]) * int32(image.ImageSubpixelScale-xHr) * int32(image.ImageSubpixelScale-yHr)
				if xLr+1 < len(fgPtr) {
					fg += int32(fgPtr[xLr+1]) * int32(image.ImageSubpixelScale-yHr) * int32(xHr)
				}
			}

			yLr++
			fgPtr = sifbc.base.source.RowPtr(yLr)
			if xLr < len(fgPtr) {
				fg += int32(fgPtr[xLr]) * int32(image.ImageSubpixelScale-xHr) * int32(yHr)
				if xLr+1 < len(fgPtr) {
					fg += int32(fgPtr[xLr+1]) * int32(xHr) * int32(yHr)
				}
			}

			fg >>= image.ImageSubpixelShift * 2
			srcAlpha = int32(color.Gray8FullValue())
		} else {
			// Slow path - handle boundary conditions
			if xLr < -1 || yLr < -1 || xLr > maxx || yLr > maxy {
				// Completely outside - use background
				fg = int32(backV)
				srcAlpha = int32(backA)
			} else {
				// Partially outside - blend with background
				fg = 0
				srcAlpha = 0

				xHr &= image.ImageSubpixelMask
				yHr &= image.ImageSubpixelMask

				// Sample four corners with boundary checking
				weight := int32(image.ImageSubpixelScale-xHr) * int32(image.ImageSubpixelScale-yHr)
				if xLr >= 0 && yLr >= 0 && xLr <= maxx && yLr <= maxy {
					fgPtr := sifbc.base.source.RowPtr(yLr)
					if xLr < len(fgPtr) {
						fg += weight * int32(fgPtr[xLr])
						srcAlpha += weight * int32(color.Gray8FullValue())
					}
				} else {
					fg += int32(backV) * weight
					srcAlpha += int32(backA) * weight
				}

				xLr++
				weight = int32(xHr) * int32(image.ImageSubpixelScale-yHr)
				if xLr >= 0 && yLr >= 0 && xLr <= maxx && yLr <= maxy {
					fgPtr := sifbc.base.source.RowPtr(yLr)
					if xLr < len(fgPtr) {
						fg += weight * int32(fgPtr[xLr])
						srcAlpha += weight * int32(color.Gray8FullValue())
					}
				} else {
					fg += int32(backV) * weight
					srcAlpha += int32(backA) * weight
				}

				xLr--
				yLr++
				weight = int32(image.ImageSubpixelScale-xHr) * int32(yHr)
				if xLr >= 0 && yLr >= 0 && xLr <= maxx && yLr <= maxy {
					fgPtr := sifbc.base.source.RowPtr(yLr)
					if xLr < len(fgPtr) {
						fg += weight * int32(fgPtr[xLr])
						srcAlpha += weight * int32(color.Gray8FullValue())
					}
				} else {
					fg += int32(backV) * weight
					srcAlpha += int32(backA) * weight
				}

				xLr++
				weight = int32(xHr) * int32(yHr)
				if xLr >= 0 && yLr >= 0 && xLr <= maxx && yLr <= maxy {
					fgPtr := sifbc.base.source.RowPtr(yLr)
					if xLr < len(fgPtr) {
						fg += weight * int32(fgPtr[xLr])
						srcAlpha += weight * int32(color.Gray8FullValue())
					}
				} else {
					fg += int32(backV) * weight
					srcAlpha += int32(backA) * weight
				}

				fg >>= image.ImageSubpixelShift * 2
				srcAlpha >>= image.ImageSubpixelShift * 2
			}
		}

		if fg < 0 {
			fg = 0
		}
		if fg > int32(color.Gray8FullValue()) {
			fg = int32(color.Gray8FullValue())
		}
		if srcAlpha < 0 {
			srcAlpha = 0
		}
		if srcAlpha > int32(color.Gray8FullValue()) {
			srcAlpha = int32(color.Gray8FullValue())
		}

		span[i] = color.NewGray8WithAlpha[color.Linear](basics.Int8u(fg), basics.Int8u(srcAlpha))
		sifbc.base.interpolator.Next()
	}
}

// SpanImageFilterGray2x2 implements 2x2 filter with lookup table for grayscale images.
// This is a port of AGG's span_image_filter_gray_2x2 template class.
type SpanImageFilterGray2x2[Source GraySourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterGray2x2 creates a new grayscale 2x2 filter.
func NewSpanImageFilterGray2x2[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	filter *image.ImageFilterLUT,
) *SpanImageFilterGray2x2[Source, Interpolator] {
	base := NewSpanImageFilter[Source, Interpolator]()
	base.SetFilter(filter)
	return &SpanImageFilterGray2x2[Source, Interpolator]{
		base: base,
	}
}

// NewSpanImageFilterGray2x2WithParams creates a new grayscale 2x2 filter with parameters.
func NewSpanImageFilterGray2x2WithParams[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilterGray2x2[Source, Interpolator] {
	return &SpanImageFilterGray2x2[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, filter),
	}
}

// Generate generates a span of grayscale pixels using 2x2 filtering.
func (sif2x2 *SpanImageFilterGray2x2[Source, Interpolator]) Generate(span []color.Gray8[color.Linear], x, y, length int) {
	sif2x2.base.interpolator.Begin(float64(x)+sif2x2.base.FilterDxDbl(), float64(y)+sif2x2.base.FilterDyDbl(), length)

	filter := sif2x2.base.Filter()
	if filter == nil {
		// Fallback to bilinear if no filter
		bilinear := NewSpanImageFilterGrayBilinearWithParams(sif2x2.base.Source(), sif2x2.base.Interpolator())
		bilinear.Generate(span, x, y, length)
		return
	}

	weightArray := filter.WeightArray()
	weightOffset := (filter.Diameter()/2 - 1) << image.ImageSubpixelShift

	for i := 0; i < length; i++ {
		xHr, yHr := sif2x2.base.interpolator.Coordinates()

		// Apply filter offset
		xHr -= sif2x2.base.FilterDxInt()
		yHr -= sif2x2.base.FilterDyInt()

		// Extract integer and fractional parts
		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		xHr &= image.ImageSubpixelMask
		yHr &= image.ImageSubpixelMask

		var fg int32 = 0

		// Get 2x2 pixel neighborhood
		fgPtr := sif2x2.base.source.Span(xLr, yLr, 2)

		if len(fgPtr) >= 1 && len(weightArray) > weightOffset+xHr+image.ImageSubpixelScale {
			// Top-left pixel
			weight := (int32(weightArray[weightOffset+xHr+image.ImageSubpixelScale])*
				int32(weightArray[weightOffset+yHr+image.ImageSubpixelScale]) +
				image.ImageFilterScale/2) >> image.ImageFilterShift
			fg += weight * int32(fgPtr[0])

			// Top-right pixel
			fgPtr = sif2x2.base.source.NextX()
			if len(fgPtr) >= 1 {
				weight = (int32(weightArray[weightOffset+xHr])*
					int32(weightArray[weightOffset+yHr+image.ImageSubpixelScale]) +
					image.ImageFilterScale/2) >> image.ImageFilterShift
				fg += weight * int32(fgPtr[0])
			}

			// Bottom-left pixel
			fgPtr = sif2x2.base.source.NextY()
			if len(fgPtr) >= 1 {
				weight = (int32(weightArray[weightOffset+xHr+image.ImageSubpixelScale])*
					int32(weightArray[weightOffset+yHr]) +
					image.ImageFilterScale/2) >> image.ImageFilterShift
				fg += weight * int32(fgPtr[0])
			}

			// Bottom-right pixel
			fgPtr = sif2x2.base.source.NextX()
			if len(fgPtr) >= 1 {
				weight = (int32(weightArray[weightOffset+xHr])*
					int32(weightArray[weightOffset+yHr]) +
					image.ImageFilterScale/2) >> image.ImageFilterShift
				fg += weight * int32(fgPtr[0])
			}
		}

		fg >>= image.ImageFilterShift
		if fg < 0 {
			fg = 0
		}
		if fg > int32(color.Gray8FullValue()) {
			fg = int32(color.Gray8FullValue())
		}

		span[i] = color.NewGray8WithAlpha[color.Linear](basics.Int8u(fg), color.Gray8FullValue())
		sif2x2.base.interpolator.Next()
	}
}

// SpanImageFilterGray implements general grayscale filtering with arbitrary kernel size.
// This is a port of AGG's span_image_filter_gray template class.
type SpanImageFilterGray[Source GraySourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterGray creates a new general grayscale filter.
func NewSpanImageFilterGray[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	filter *image.ImageFilterLUT,
) *SpanImageFilterGray[Source, Interpolator] {
	base := NewSpanImageFilter[Source, Interpolator]()
	base.SetFilter(filter)
	return &SpanImageFilterGray[Source, Interpolator]{
		base: base,
	}
}

// NewSpanImageFilterGrayWithParams creates a new general grayscale filter with parameters.
func NewSpanImageFilterGrayWithParams[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilterGray[Source, Interpolator] {
	return &SpanImageFilterGray[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, filter),
	}
}

// Generate generates a span of grayscale pixels using general filtering.
func (sifg *SpanImageFilterGray[Source, Interpolator]) Generate(span []color.Gray8[color.Linear], x, y, length int) {
	sifg.base.interpolator.Begin(float64(x)+sifg.base.FilterDxDbl(), float64(y)+sifg.base.FilterDyDbl(), length)

	filter := sifg.base.Filter()
	if filter == nil {
		// Fallback to bilinear if no filter
		bilinear := NewSpanImageFilterGrayBilinearWithParams(sifg.base.Source(), sifg.base.Interpolator())
		bilinear.Generate(span, x, y, length)
		return
	}

	diameter := filter.Diameter()
	start := filter.Start()
	weightArray := filter.WeightArray()

	for i := 0; i < length; i++ {
		sx, sy := sifg.base.interpolator.Coordinates()

		// Apply filter offset
		sx -= sifg.base.FilterDxInt()
		sy -= sifg.base.FilterDyInt()

		xHr := sx
		yHr := sy

		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		var fg int32 = 0

		xFract := xHr & image.ImageSubpixelMask
		yCount := diameter

		yHr = image.ImageSubpixelMask - (yHr & image.ImageSubpixelMask)
		fgPtr := sifg.base.source.Span(xLr+start, yLr+start, diameter)

		for yIndex := 0; yIndex < yCount; yIndex++ {
			xCount := diameter
			weightY := int32(0)
			if yHr < len(weightArray) {
				weightY = int32(weightArray[yHr])
			}
			xHr = image.ImageSubpixelMask - xFract

			for xIndex := 0; xIndex < xCount; xIndex++ {
				if len(fgPtr) > 0 {
					weightX := int32(0)
					if xHr < len(weightArray) {
						weightX = int32(weightArray[xHr])
					}
					fg += int32(fgPtr[0]) * ((weightY*weightX + image.ImageFilterScale/2) >> image.ImageFilterShift)
				}

				if xIndex < xCount-1 {
					xHr += image.ImageSubpixelScale
					fgPtr = sifg.base.source.NextX()
				}
			}

			if yIndex < yCount-1 {
				yHr += image.ImageSubpixelScale
				fgPtr = sifg.base.source.NextY()
			}
		}

		fg >>= image.ImageFilterShift
		if fg < 0 {
			fg = 0
		}
		if fg > int32(color.Gray8FullValue()) {
			fg = int32(color.Gray8FullValue())
		}

		span[i] = color.NewGray8WithAlpha[color.Linear](basics.Int8u(fg), color.Gray8FullValue())
		sifg.base.interpolator.Next()
	}
}

// SpanImageResampleGrayAffine implements affine resampling for grayscale images.
// This is a port of AGG's span_image_resample_gray_affine template class.
type SpanImageResampleGrayAffine[Source GraySourceInterface] struct {
	base *SpanImageResampleAffine[Source]
}

// NewSpanImageResampleGrayAffine creates a new grayscale affine resampling filter.
func NewSpanImageResampleGrayAffine[Source GraySourceInterface]() *SpanImageResampleGrayAffine[Source] {
	return &SpanImageResampleGrayAffine[Source]{
		base: NewSpanImageResampleAffine[Source](),
	}
}

// NewSpanImageResampleGrayAffineWithParams creates a new grayscale affine resampling filter with parameters.
func NewSpanImageResampleGrayAffineWithParams[Source GraySourceInterface](
	src Source,
	interpolator *SpanInterpolatorLinear[*transform.TransAffine],
	filter *image.ImageFilterLUT,
) *SpanImageResampleGrayAffine[Source] {
	return &SpanImageResampleGrayAffine[Source]{
		base: NewSpanImageResampleAffineWithParams(src, interpolator, filter),
	}
}

// Generate generates a span of grayscale pixels using affine resampling.
func (sirga *SpanImageResampleGrayAffine[Source]) Generate(span []color.Gray8[color.Linear], x, y, length int) {
	// Get base filter from the resample affine structure
	baseFilter := sirga.base.base
	baseFilter.interpolator.Begin(float64(x)+baseFilter.FilterDxDbl(), float64(y)+baseFilter.FilterDyDbl(), length)

	filter := baseFilter.Filter()
	if filter == nil {
		// Fallback to bilinear if no filter
		bilinear := NewSpanImageFilterGrayBilinearWithParams(baseFilter.Source(), baseFilter.Interpolator())
		bilinear.Generate(span, x, y, length)
		return
	}

	diameter := filter.Diameter()
	filterScale := diameter << image.ImageSubpixelShift
	radiusX := (diameter * sirga.base.RX()) >> 1
	radiusY := (diameter * sirga.base.RY()) >> 1
	lenXLr := (diameter*sirga.base.RX() + image.ImageSubpixelMask) >> image.ImageSubpixelShift

	weightArray := filter.WeightArray()

	for i := 0; i < length; i++ {
		sx, sy := baseFilter.interpolator.Coordinates()

		sx += baseFilter.FilterDxInt() - radiusX
		sy += baseFilter.FilterDyInt() - radiusY

		var fg int32 = 0

		yLr := sy >> image.ImageSubpixelShift
		yHr := ((image.ImageSubpixelMask - (sy & image.ImageSubpixelMask)) * sirga.base.RYInv()) >> image.ImageSubpixelShift
		totalWeight := 0
		xLr := sx >> image.ImageSubpixelShift
		xHr := ((image.ImageSubpixelMask - (sx & image.ImageSubpixelMask)) * sirga.base.RXInv()) >> image.ImageSubpixelShift

		xHr2 := xHr
		fgPtr := sirga.base.Source().Span(xLr, yLr, lenXLr)

		for yHr < len(weightArray) {

			weightY := int32(weightArray[yHr])
			xHr = xHr2

			for xHr < len(weightArray) {

				weight := (weightY*int32(weightArray[xHr]) + image.ImageFilterScale/2) >> image.ImageFilterShift

				if len(fgPtr) > 0 {
					fg += int32(fgPtr[0]) * weight
				}
				totalWeight += int(weight)
				xHr += sirga.base.RXInv()

				if xHr >= filterScale {
					break
				}
				fgPtr = sirga.base.Source().NextX()
			}

			yHr += sirga.base.RYInv()
			if yHr >= filterScale {
				break
			}
			fgPtr = sirga.base.Source().NextY()
		}

		if totalWeight > 0 {
			fg /= int32(totalWeight)
		}

		if fg < 0 {
			fg = 0
		}
		if fg > int32(color.Gray8FullValue()) {
			fg = int32(color.Gray8FullValue())
		}

		span[i] = color.NewGray8WithAlpha[color.Linear](basics.Int8u(fg), color.Gray8FullValue())
		baseFilter.interpolator.Next()
	}
}

// SpanImageResampleGray implements general grayscale resampling with configurable interpolation.
// This is a port of AGG's span_image_resample_gray template class.
type SpanImageResampleGray[Source GraySourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageResample[Source, Interpolator]
}

// NewSpanImageResampleGray creates a new general grayscale resampling filter.
func NewSpanImageResampleGray[Source GraySourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageResampleGray[Source, Interpolator] {
	return &SpanImageResampleGray[Source, Interpolator]{
		base: NewSpanImageResample[Source, Interpolator](),
	}
}

// NewSpanImageResampleGrayWithParams creates a new general grayscale resampling filter with parameters.
func NewSpanImageResampleGrayWithParams[Source GraySourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageResampleGray[Source, Interpolator] {
	return &SpanImageResampleGray[Source, Interpolator]{
		base: NewSpanImageResampleWithParams(src, interpolator, filter),
	}
}

// Generate generates a span of grayscale pixels using general resampling.
func (sirg *SpanImageResampleGray[Source, Interpolator]) Generate(span []color.Gray8[color.Linear], x, y, length int) {
	// Get base filter from the resample structure
	baseFilter := sirg.base.base
	baseFilter.interpolator.Begin(float64(x)+baseFilter.FilterDxDbl(), float64(y)+baseFilter.FilterDyDbl(), length)

	filter := baseFilter.Filter()
	if filter == nil {
		// Fallback to bilinear if no filter
		bilinear := NewSpanImageFilterGrayBilinearWithParams(baseFilter.Source(), baseFilter.Interpolator())
		bilinear.Generate(span, x, y, length)
		return
	}

	diameter := filter.Diameter()
	filterScale := diameter << image.ImageSubpixelShift
	weightArray := filter.WeightArray()

	for i := 0; i < length; i++ {
		sx, sy := baseFilter.interpolator.Coordinates()

		// Get local scale from interpolator if it supports it
		rx := image.ImageSubpixelScale
		ry := image.ImageSubpixelScale
		rxInv := image.ImageSubpixelScale
		ryInv := image.ImageSubpixelScale

		// Adjust scale using base method
		sirg.base.AdjustScale(&rx, &ry)

		rxInv = image.ImageSubpixelScale * image.ImageSubpixelScale / rx
		ryInv = image.ImageSubpixelScale * image.ImageSubpixelScale / ry

		radiusX := (diameter * rx) >> 1
		radiusY := (diameter * ry) >> 1
		lenXLr := (diameter*rx + image.ImageSubpixelMask) >> image.ImageSubpixelShift

		sx += baseFilter.FilterDxInt() - radiusX
		sy += baseFilter.FilterDyInt() - radiusY

		var fg int32 = 0

		yLr := sy >> image.ImageSubpixelShift
		yHr := ((image.ImageSubpixelMask - (sy & image.ImageSubpixelMask)) * ryInv) >> image.ImageSubpixelShift
		totalWeight := 0
		xLr := sx >> image.ImageSubpixelShift
		xHr := ((image.ImageSubpixelMask - (sx & image.ImageSubpixelMask)) * rxInv) >> image.ImageSubpixelShift
		xHr2 := xHr

		fgPtr := sirg.base.Source().Span(xLr, yLr, lenXLr)

		for yHr < len(weightArray) {

			weightY := int32(weightArray[yHr])
			xHr = xHr2

			for xHr < len(weightArray) {

				weight := (weightY*int32(weightArray[xHr]) + image.ImageFilterScale/2) >> image.ImageFilterShift

				if len(fgPtr) > 0 {
					fg += int32(fgPtr[0]) * weight
				}
				totalWeight += int(weight)
				xHr += rxInv

				if xHr >= filterScale {
					break
				}
				fgPtr = sirg.base.Source().NextX()
			}

			yHr += ryInv
			if yHr >= filterScale {
				break
			}
			fgPtr = sirg.base.Source().NextY()
		}

		if totalWeight > 0 {
			fg /= int32(totalWeight)
		}

		if fg < 0 {
			fg = 0
		}
		if fg > int32(color.Gray8FullValue()) {
			fg = int32(color.Gray8FullValue())
		}

		span[i] = color.NewGray8WithAlpha[color.Linear](basics.Int8u(fg), color.Gray8FullValue())
		baseFilter.interpolator.Next()
	}
}
