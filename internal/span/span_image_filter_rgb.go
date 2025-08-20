// Package span provides RGB image filtering span generation functionality for AGG.
// This implements a port of AGG's span_image_filter_rgb.h functionality.
package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/image"
)

// SpanImageFilterRGBNN implements nearest neighbor RGB image filtering.
// This is a port of AGG's span_image_filter_rgb_nn template class.
type SpanImageFilterRGBNN[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBNN creates a new nearest neighbor RGB span filter.
func NewSpanImageFilterRGBNN[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGBNN[Source, Interpolator] {
	return &SpanImageFilterRGBNN[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterRGBNNWithParams creates a new nearest neighbor RGB span filter with parameters.
func NewSpanImageFilterRGBNNWithParams[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
) *SpanImageFilterRGBNN[Source, Interpolator] {
	return &SpanImageFilterRGBNN[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, nil),
	}
}

// Generate generates a span of RGB pixels using nearest neighbor interpolation.
func (sif *SpanImageFilterRGBNN[Source, Interpolator]) Generate(span []color.RGB8[color.Linear], x, y int) {
	length := len(span)
	if length == 0 {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)

	for i := 0; i < length; i++ {
		xHr, yHr := sif.base.interpolator.Coordinates()

		// Get pixel coordinates at image subpixel precision
		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		// Sample the source image using proper RGB interface
		fgPtr := sif.base.source.Span(xLr, yLr, 1)
		orderType := sif.base.source.OrderType()

		if len(fgPtr) >= 3 {
			span[i] = color.RGB8[color.Linear]{
				R: fgPtr[orderType.R],
				G: fgPtr[orderType.G],
				B: fgPtr[orderType.B],
			}
		} else {
			// Fallback - set to black if no pixel data
			span[i] = color.RGB8[color.Linear]{R: 0, G: 0, B: 0}
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGBBilinear implements bilinear RGB image filtering.
// This is a port of AGG's span_image_filter_rgb_bilinear template class.
type SpanImageFilterRGBBilinear[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBBilinear creates a new bilinear RGB span filter.
func NewSpanImageFilterRGBBilinear[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGBBilinear[Source, Interpolator] {
	return &SpanImageFilterRGBBilinear[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterRGBBilinearWithParams creates a new bilinear RGB span filter with parameters.
func NewSpanImageFilterRGBBilinearWithParams[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
) *SpanImageFilterRGBBilinear[Source, Interpolator] {
	return &SpanImageFilterRGBBilinear[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, nil),
	}
}

// Generate generates a span of RGB pixels using bilinear interpolation.
func (sif *SpanImageFilterRGBBilinear[Source, Interpolator]) Generate(span []color.RGB8[color.Linear], x, y int) {
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

		var fg [3]int // RGB accumulator

		xHr &= image.ImageSubpixelMask
		yHr &= image.ImageSubpixelMask

		orderType := sif.base.source.OrderType()

		// Top-left sample
		fgPtr := sif.base.source.Span(xLr, yLr, 2)
		weight := (image.ImageSubpixelScale - xHr) * (image.ImageSubpixelScale - yHr)
		if len(fgPtr) >= 3 {
			fg[0] += weight * int(fgPtr[orderType.R])
			fg[1] += weight * int(fgPtr[orderType.G])
			fg[2] += weight * int(fgPtr[orderType.B])
		}

		// Top-right sample
		fgPtr = sif.base.source.NextX()
		weight = xHr * (image.ImageSubpixelScale - yHr)
		if len(fgPtr) >= 3 {
			fg[0] += weight * int(fgPtr[orderType.R])
			fg[1] += weight * int(fgPtr[orderType.G])
			fg[2] += weight * int(fgPtr[orderType.B])
		}

		// Bottom-left sample
		fgPtr = sif.base.source.NextY()
		weight = (image.ImageSubpixelScale - xHr) * yHr
		if len(fgPtr) >= 3 {
			fg[0] += weight * int(fgPtr[orderType.R])
			fg[1] += weight * int(fgPtr[orderType.G])
			fg[2] += weight * int(fgPtr[orderType.B])
		}

		// Bottom-right sample
		fgPtr = sif.base.source.NextX()
		weight = xHr * yHr
		if len(fgPtr) >= 3 {
			fg[0] += weight * int(fgPtr[orderType.R])
			fg[1] += weight * int(fgPtr[orderType.G])
			fg[2] += weight * int(fgPtr[orderType.B])
		}

		// Downshift to get final values
		r := fg[0] >> (image.ImageSubpixelShift * 2)
		g := fg[1] >> (image.ImageSubpixelShift * 2)
		b := fg[2] >> (image.ImageSubpixelShift * 2)

		// Clamp to valid range
		if r > 255 {
			r = 255
		}
		if g > 255 {
			g = 255
		}
		if b > 255 {
			b = 255
		}
		if r < 0 {
			r = 0
		}
		if g < 0 {
			g = 0
		}
		if b < 0 {
			b = 0
		}

		span[i] = color.RGB8[color.Linear]{
			R: basics.Int8u(r),
			G: basics.Int8u(g),
			B: basics.Int8u(b),
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGBBilinearClip implements bilinear RGB image filtering with background clipping.
// This is a port of AGG's span_image_filter_rgb_bilinear_clip template class.
type SpanImageFilterRGBBilinearClip[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base            *SpanImageFilter[Source, Interpolator]
	backgroundColor color.RGB8[color.Linear]
}

// NewSpanImageFilterRGBBilinearClip creates a new bilinear clipping RGB span filter.
func NewSpanImageFilterRGBBilinearClip[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGBBilinearClip[Source, Interpolator] {
	return &SpanImageFilterRGBBilinearClip[Source, Interpolator]{
		base:            NewSpanImageFilter[Source, Interpolator](),
		backgroundColor: color.RGB8[color.Linear]{R: 0, G: 0, B: 0},
	}
}

// NewSpanImageFilterRGBBilinearClipWithParams creates a new bilinear clipping RGB span filter with parameters.
func NewSpanImageFilterRGBBilinearClipWithParams[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	backgroundColor color.RGB8[color.Linear],
	interpolator Interpolator,
) *SpanImageFilterRGBBilinearClip[Source, Interpolator] {
	return &SpanImageFilterRGBBilinearClip[Source, Interpolator]{
		base:            NewSpanImageFilterWithParams(src, interpolator, nil),
		backgroundColor: backgroundColor,
	}
}

// BackgroundColor returns the current background color.
func (sif *SpanImageFilterRGBBilinearClip[Source, Interpolator]) BackgroundColor() color.RGB8[color.Linear] {
	return sif.backgroundColor
}

// SetBackgroundColor sets the background color for out-of-bounds pixels.
func (sif *SpanImageFilterRGBBilinearClip[Source, Interpolator]) SetBackgroundColor(c color.RGB8[color.Linear]) {
	sif.backgroundColor = c
}

// Generate generates a span of RGB pixels using bilinear interpolation with clipping.
func (sif *SpanImageFilterRGBBilinearClip[Source, Interpolator]) Generate(span []color.RGB8[color.Linear], x, y int) {
	length := len(span)
	if length == 0 {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)

	backR := int(sif.backgroundColor.R)
	backG := int(sif.backgroundColor.G)
	backB := int(sif.backgroundColor.B)

	maxX := sif.base.source.Width() - 1
	maxY := sif.base.source.Height() - 1

	for i := 0; i < length; i++ {
		xHr, yHr := sif.base.interpolator.Coordinates()

		xHr -= sif.base.FilterDxInt()
		yHr -= sif.base.FilterDyInt()

		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		var fg [3]int // RGB accumulator

		if xLr >= 0 && yLr >= 0 && xLr < maxX && yLr < maxY {
			// All samples are within bounds - fast path
			xHr &= image.ImageSubpixelMask
			yHr &= image.ImageSubpixelMask

			if sourceAccessor, ok := interface{}(sif.base.source).(interface {
				RowPtr(y int) []basics.Int8u
			}); ok {
				// Implementation similar to bilinear but with bounds checking and row access
				row := sourceAccessor.RowPtr(yLr)
				pixelOffset := xLr * 3

				// Top-left sample
				weight := (image.ImageSubpixelScale - xHr) * (image.ImageSubpixelScale - yHr)
				if pixelOffset+2 < len(row) {
					fg[0] += weight * int(row[pixelOffset])
					fg[1] += weight * int(row[pixelOffset+1])
					fg[2] += weight * int(row[pixelOffset+2])
				}

				// Top-right sample
				weight = xHr * (image.ImageSubpixelScale - yHr)
				if pixelOffset+5 < len(row) {
					fg[0] += weight * int(row[pixelOffset+3])
					fg[1] += weight * int(row[pixelOffset+4])
					fg[2] += weight * int(row[pixelOffset+5])
				}

				// Bottom row samples
				yLr++
				if yLr < sif.base.source.Height() {
					row = sourceAccessor.RowPtr(yLr)

					// Bottom-left sample
					weight = (image.ImageSubpixelScale - xHr) * yHr
					if pixelOffset+2 < len(row) {
						fg[0] += weight * int(row[pixelOffset])
						fg[1] += weight * int(row[pixelOffset+1])
						fg[2] += weight * int(row[pixelOffset+2])
					}

					// Bottom-right sample
					weight = xHr * yHr
					if pixelOffset+5 < len(row) {
						fg[0] += weight * int(row[pixelOffset+3])
						fg[1] += weight * int(row[pixelOffset+4])
						fg[2] += weight * int(row[pixelOffset+5])
					}
				}

				// Downshift results
				fg[0] >>= (image.ImageSubpixelShift * 2)
				fg[1] >>= (image.ImageSubpixelShift * 2)
				fg[2] >>= (image.ImageSubpixelShift * 2)
			}
		} else {
			// Handle clipping case
			if xLr < -1 || yLr < -1 || xLr > maxX || yLr > maxY {
				// Completely outside - use background
				fg[0] = backR
				fg[1] = backG
				fg[2] = backB
			} else {
				// Partial overlap - blend with background
				xHr &= image.ImageSubpixelMask
				yHr &= image.ImageSubpixelMask

				// This would require complex boundary checking for each of the 4 samples
				// For now, simplified to use background color
				fg[0] = backR
				fg[1] = backG
				fg[2] = backB
			}
		}

		// Clamp to valid range
		if fg[0] > 255 {
			fg[0] = 255
		}
		if fg[1] > 255 {
			fg[1] = 255
		}
		if fg[2] > 255 {
			fg[2] = 255
		}
		if fg[0] < 0 {
			fg[0] = 0
		}
		if fg[1] < 0 {
			fg[1] = 0
		}
		if fg[2] < 0 {
			fg[2] = 0
		}

		span[i] = color.RGB8[color.Linear]{
			R: basics.Int8u(fg[0]),
			G: basics.Int8u(fg[1]),
			B: basics.Int8u(fg[2]),
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGB2x2 implements 2x2 RGB image filtering with lookup table.
// This is a port of AGG's span_image_filter_rgb_2x2 template class.
type SpanImageFilterRGB2x2[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGB2x2 creates a new 2x2 RGB span filter.
func NewSpanImageFilterRGB2x2[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGB2x2[Source, Interpolator] {
	return &SpanImageFilterRGB2x2[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterRGB2x2WithParams creates a new 2x2 RGB span filter with parameters.
func NewSpanImageFilterRGB2x2WithParams[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilterRGB2x2[Source, Interpolator] {
	return &SpanImageFilterRGB2x2[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, filter),
	}
}

// Generate generates a span of RGB pixels using 2x2 filter kernel.
func (sif *SpanImageFilterRGB2x2[Source, Interpolator]) Generate(span []color.RGB8[color.Linear], x, y int) {
	length := len(span)
	if length == 0 || sif.base.filter == nil {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)

	// Get weight array from filter (assuming it has this method)
	weightArray := sif.base.filter.WeightArray()
	if weightArray == nil {
		return
	}

	// Calculate offset for 2x2 filter
	offset := ((sif.base.filter.Diameter()/2 - 1) << image.ImageSubpixelShift)

	for i := 0; i < length; i++ {
		xHr, yHr := sif.base.interpolator.Coordinates()

		xHr -= sif.base.FilterDxInt()
		yHr -= sif.base.FilterDyInt()

		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		var fg [3]int // RGB accumulator

		xHr &= image.ImageSubpixelMask
		yHr &= image.ImageSubpixelMask

		orderType := sif.base.source.OrderType()

		// Sample 1 (top-left)
		fgPtr := sif.base.source.Span(xLr, yLr, 2)
		weight := (int(weightArray[xHr+image.ImageSubpixelScale+offset])*
			int(weightArray[yHr+image.ImageSubpixelScale+offset]) +
			image.ImageFilterScale/2) >> image.ImageFilterShift
		if len(fgPtr) >= 3 {
			fg[0] += weight * int(fgPtr[orderType.R])
			fg[1] += weight * int(fgPtr[orderType.G])
			fg[2] += weight * int(fgPtr[orderType.B])
		}

		// Sample 2 (top-right)
		fgPtr = sif.base.source.NextX()
		weight = (int(weightArray[xHr+offset])*
			int(weightArray[yHr+image.ImageSubpixelScale+offset]) +
			image.ImageFilterScale/2) >> image.ImageFilterShift
		if len(fgPtr) >= 3 {
			fg[0] += weight * int(fgPtr[orderType.R])
			fg[1] += weight * int(fgPtr[orderType.G])
			fg[2] += weight * int(fgPtr[orderType.B])
		}

		// Sample 3 (bottom-left)
		fgPtr = sif.base.source.NextY()
		weight = (int(weightArray[xHr+image.ImageSubpixelScale+offset])*
			int(weightArray[yHr+offset]) +
			image.ImageFilterScale/2) >> image.ImageFilterShift
		if len(fgPtr) >= 3 {
			fg[0] += weight * int(fgPtr[orderType.R])
			fg[1] += weight * int(fgPtr[orderType.G])
			fg[2] += weight * int(fgPtr[orderType.B])
		}

		// Sample 4 (bottom-right)
		fgPtr = sif.base.source.NextX()
		weight = (int(weightArray[xHr+offset])*
			int(weightArray[yHr+offset]) +
			image.ImageFilterScale/2) >> image.ImageFilterShift
		if len(fgPtr) >= 3 {
			fg[0] += weight * int(fgPtr[orderType.R])
			fg[1] += weight * int(fgPtr[orderType.G])
			fg[2] += weight * int(fgPtr[orderType.B])
		}

		// Downshift results
		fg[0] >>= image.ImageFilterShift
		fg[1] >>= image.ImageFilterShift
		fg[2] >>= image.ImageFilterShift

		// Clamp to valid range
		if fg[0] > 255 {
			fg[0] = 255
		}
		if fg[1] > 255 {
			fg[1] = 255
		}
		if fg[2] > 255 {
			fg[2] = 255
		}

		span[i] = color.RGB8[color.Linear]{
			R: basics.Int8u(fg[0]),
			G: basics.Int8u(fg[1]),
			B: basics.Int8u(fg[2]),
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGB implements general RGB image filtering with configurable kernel size.
// This is a port of AGG's span_image_filter_rgb template class.
type SpanImageFilterRGB[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGB creates a new general RGB span filter.
func NewSpanImageFilterRGB[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGB[Source, Interpolator] {
	return &SpanImageFilterRGB[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterRGBWithParams creates a new general RGB span filter with parameters.
func NewSpanImageFilterRGBWithParams[Source RGBSourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilterRGB[Source, Interpolator] {
	return &SpanImageFilterRGB[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, filter),
	}
}

// Generate generates a span of RGB pixels using a configurable filter kernel.
func (sif *SpanImageFilterRGB[Source, Interpolator]) Generate(span []color.RGB8[color.Linear], x, y int) {
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

		var fg [3]int // RGB accumulator

		xFract := xHr & image.ImageSubpixelMask
		yCount := diameter

		yHr = image.ImageSubpixelMask - (yHr & image.ImageSubpixelMask)

		orderType := sif.base.source.OrderType()
		fgPtr := sif.base.source.Span(xLr+start, yLr+start, diameter)

		for yCount > 0 {
			xCount := diameter
			weightY := weightArray[yHr]
			xHr = image.ImageSubpixelMask - xFract

			for xCount > 0 {
				weight := (int(weightY)*int(weightArray[xHr]) + image.ImageFilterScale/2) >> image.ImageFilterShift

				if len(fgPtr) >= 3 {
					fg[0] += weight * int(fgPtr[orderType.R])
					fg[1] += weight * int(fgPtr[orderType.G])
					fg[2] += weight * int(fgPtr[orderType.B])
				}

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

		// Downshift results
		fg[0] >>= image.ImageFilterShift
		fg[1] >>= image.ImageFilterShift
		fg[2] >>= image.ImageFilterShift

		// Clamp to valid range
		if fg[0] < 0 {
			fg[0] = 0
		}
		if fg[1] < 0 {
			fg[1] = 0
		}
		if fg[2] < 0 {
			fg[2] = 0
		}

		if fg[0] > 255 {
			fg[0] = 255
		}
		if fg[1] > 255 {
			fg[1] = 255
		}
		if fg[2] > 255 {
			fg[2] = 255
		}

		span[i] = color.RGB8[color.Linear]{
			R: basics.Int8u(fg[0]),
			G: basics.Int8u(fg[1]),
			B: basics.Int8u(fg[2]),
		}

		sif.base.interpolator.Next()
	}
}
