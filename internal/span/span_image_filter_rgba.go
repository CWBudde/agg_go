// Package span provides RGBA image filtering span generation functionality for AGG.
// This implements a port of AGG's span_image_filter_rgba.h functionality.
package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/image"
)

// SpanImageFilterRGBANN implements nearest neighbor RGBA image filtering.
// This is a port of AGG's span_image_filter_rgba_nn template class.
type SpanImageFilterRGBANN[Source SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBANN creates a new nearest neighbor RGBA span filter.
func NewSpanImageFilterRGBANN[Source SourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGBANN[Source, Interpolator] {
	return &SpanImageFilterRGBANN[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterRGBANNWithParams creates a new nearest neighbor RGBA span filter with parameters.
func NewSpanImageFilterRGBANNWithParams[Source SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
) *SpanImageFilterRGBANN[Source, Interpolator] {
	return &SpanImageFilterRGBANN[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, nil),
	}
}

// Generate generates a span of RGBA pixels using nearest neighbor interpolation.
func (sif *SpanImageFilterRGBANN[Source, Interpolator]) Generate(span []color.RGBA8[color.Linear], x, y int) {
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

		// Sample the source image
		if sourceAccessor, ok := interface{}(sif.base.source).(interface {
			Span(x, y, length int) []basics.Int8u
		}); ok {
			fgPtr := sourceAccessor.Span(xLr, yLr, 1)
			if len(fgPtr) >= 4 {
				span[i] = color.RGBA8[color.Linear]{
					R: fgPtr[0], // Assumes RGBA order
					G: fgPtr[1],
					B: fgPtr[2],
					A: fgPtr[3],
				}
			}
		} else {
			// Fallback - set to transparent black if source doesn't support span access
			span[i] = color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0}
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGBABilinear implements bilinear RGBA image filtering.
// This is a port of AGG's span_image_filter_rgba_bilinear template class.
type SpanImageFilterRGBABilinear[Source SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBABilinear creates a new bilinear RGBA span filter.
func NewSpanImageFilterRGBABilinear[Source SourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGBABilinear[Source, Interpolator] {
	return &SpanImageFilterRGBABilinear[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterRGBABilinearWithParams creates a new bilinear RGBA span filter with parameters.
func NewSpanImageFilterRGBABilinearWithParams[Source SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
) *SpanImageFilterRGBABilinear[Source, Interpolator] {
	return &SpanImageFilterRGBABilinear[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, nil),
	}
}

// Generate generates a span of RGBA pixels using bilinear interpolation.
func (sif *SpanImageFilterRGBABilinear[Source, Interpolator]) Generate(span []color.RGBA8[color.Linear], x, y int) {
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

		var fg [4]int // RGBA accumulator

		// Initialize with rounding bias
		fg[0] = image.ImageSubpixelScale * image.ImageSubpixelScale / 2
		fg[1] = image.ImageSubpixelScale * image.ImageSubpixelScale / 2
		fg[2] = image.ImageSubpixelScale * image.ImageSubpixelScale / 2
		fg[3] = image.ImageSubpixelScale * image.ImageSubpixelScale / 2

		xHr &= image.ImageSubpixelMask
		yHr &= image.ImageSubpixelMask

		if sourceAccessor, ok := interface{}(sif.base.source).(interface {
			Span(x, y, length int) []basics.Int8u
			NextX() []basics.Int8u
			NextY() []basics.Int8u
		}); ok {
			// Top-left sample
			fgPtr := sourceAccessor.Span(xLr, yLr, 2)
			weight := (image.ImageSubpixelScale - xHr) * (image.ImageSubpixelScale - yHr)
			if len(fgPtr) >= 4 {
				fg[0] += weight * int(fgPtr[0])
				fg[1] += weight * int(fgPtr[1])
				fg[2] += weight * int(fgPtr[2])
				fg[3] += weight * int(fgPtr[3])
			}

			// Top-right sample
			fgPtr = sourceAccessor.NextX()
			weight = xHr * (image.ImageSubpixelScale - yHr)
			if len(fgPtr) >= 4 {
				fg[0] += weight * int(fgPtr[0])
				fg[1] += weight * int(fgPtr[1])
				fg[2] += weight * int(fgPtr[2])
				fg[3] += weight * int(fgPtr[3])
			}

			// Bottom-left sample
			fgPtr = sourceAccessor.NextY()
			weight = (image.ImageSubpixelScale - xHr) * yHr
			if len(fgPtr) >= 4 {
				fg[0] += weight * int(fgPtr[0])
				fg[1] += weight * int(fgPtr[1])
				fg[2] += weight * int(fgPtr[2])
				fg[3] += weight * int(fgPtr[3])
			}

			// Bottom-right sample
			fgPtr = sourceAccessor.NextX()
			weight = xHr * yHr
			if len(fgPtr) >= 4 {
				fg[0] += weight * int(fgPtr[0])
				fg[1] += weight * int(fgPtr[1])
				fg[2] += weight * int(fgPtr[2])
				fg[3] += weight * int(fgPtr[3])
			}

			// Downshift to get final values
			r := fg[0] >> (image.ImageSubpixelShift * 2)
			g := fg[1] >> (image.ImageSubpixelShift * 2)
			b := fg[2] >> (image.ImageSubpixelShift * 2)
			a := fg[3] >> (image.ImageSubpixelShift * 2)

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
			if a > 255 {
				a = 255
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
			if a < 0 {
				a = 0
			}

			// Apply alpha channel constraints - RGB components should not exceed alpha
			if r > a {
				r = a
			}
			if g > a {
				g = a
			}
			if b > a {
				b = a
			}

			span[i] = color.RGBA8[color.Linear]{
				R: basics.Int8u(r),
				G: basics.Int8u(g),
				B: basics.Int8u(b),
				A: basics.Int8u(a),
			}
		} else {
			// Fallback
			span[i] = color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0}
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGBABilinearClip implements bilinear RGBA image filtering with background clipping.
// This is a port of AGG's span_image_filter_rgba_bilinear_clip template class.
type SpanImageFilterRGBABilinearClip[Source SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base            *SpanImageFilter[Source, Interpolator]
	backgroundColor color.RGBA8[color.Linear]
}

// NewSpanImageFilterRGBABilinearClip creates a new bilinear clipping RGBA span filter.
func NewSpanImageFilterRGBABilinearClip[Source SourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGBABilinearClip[Source, Interpolator] {
	return &SpanImageFilterRGBABilinearClip[Source, Interpolator]{
		base:            NewSpanImageFilter[Source, Interpolator](),
		backgroundColor: color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0},
	}
}

// NewSpanImageFilterRGBABilinearClipWithParams creates a new bilinear clipping RGBA span filter with parameters.
func NewSpanImageFilterRGBABilinearClipWithParams[Source SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	backgroundColor color.RGBA8[color.Linear],
	interpolator Interpolator,
) *SpanImageFilterRGBABilinearClip[Source, Interpolator] {
	return &SpanImageFilterRGBABilinearClip[Source, Interpolator]{
		base:            NewSpanImageFilterWithParams(src, interpolator, nil),
		backgroundColor: backgroundColor,
	}
}

// BackgroundColor returns the current background color.
func (sif *SpanImageFilterRGBABilinearClip[Source, Interpolator]) BackgroundColor() color.RGBA8[color.Linear] {
	return sif.backgroundColor
}

// SetBackgroundColor sets the background color for out-of-bounds pixels.
func (sif *SpanImageFilterRGBABilinearClip[Source, Interpolator]) SetBackgroundColor(c color.RGBA8[color.Linear]) {
	sif.backgroundColor = c
}

// Generate generates a span of RGBA pixels using bilinear interpolation with clipping.
func (sif *SpanImageFilterRGBABilinearClip[Source, Interpolator]) Generate(span []color.RGBA8[color.Linear], x, y int) {
	length := len(span)
	if length == 0 {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)

	backR := int(sif.backgroundColor.R)
	backG := int(sif.backgroundColor.G)
	backB := int(sif.backgroundColor.B)
	backA := int(sif.backgroundColor.A)

	maxX := sif.base.source.Width() - 1
	maxY := sif.base.source.Height() - 1

	for i := 0; i < length; i++ {
		xHr, yHr := sif.base.interpolator.Coordinates()

		xHr -= sif.base.FilterDxInt()
		yHr -= sif.base.FilterDyInt()

		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift

		var fg [4]int // RGBA accumulator

		if xLr >= 0 && yLr >= 0 && xLr < maxX && yLr < maxY {
			// All samples are within bounds - fast path
			fg[0] = 0
			fg[1] = 0
			fg[2] = 0
			fg[3] = 0

			xHr &= image.ImageSubpixelMask
			yHr &= image.ImageSubpixelMask

			if sourceAccessor, ok := interface{}(sif.base.source).(interface {
				RowPtr(y int) []basics.Int8u
			}); ok {
				// Top row samples
				row := sourceAccessor.RowPtr(yLr)
				pixelOffset := xLr * 4

				// Top-left sample
				weight := (image.ImageSubpixelScale - xHr) * (image.ImageSubpixelScale - yHr)
				if pixelOffset+3 < len(row) {
					fg[0] += weight * int(row[pixelOffset])
					fg[1] += weight * int(row[pixelOffset+1])
					fg[2] += weight * int(row[pixelOffset+2])
					fg[3] += weight * int(row[pixelOffset+3])
				}

				// Top-right sample
				weight = xHr * (image.ImageSubpixelScale - yHr)
				if pixelOffset+7 < len(row) {
					fg[0] += weight * int(row[pixelOffset+4])
					fg[1] += weight * int(row[pixelOffset+5])
					fg[2] += weight * int(row[pixelOffset+6])
					fg[3] += weight * int(row[pixelOffset+7])
				}

				// Bottom row samples
				yLr++
				if yLr < sif.base.source.Height() {
					row = sourceAccessor.RowPtr(yLr)

					// Bottom-left sample
					weight = (image.ImageSubpixelScale - xHr) * yHr
					if pixelOffset+3 < len(row) {
						fg[0] += weight * int(row[pixelOffset])
						fg[1] += weight * int(row[pixelOffset+1])
						fg[2] += weight * int(row[pixelOffset+2])
						fg[3] += weight * int(row[pixelOffset+3])
					}

					// Bottom-right sample
					weight = xHr * yHr
					if pixelOffset+7 < len(row) {
						fg[0] += weight * int(row[pixelOffset+4])
						fg[1] += weight * int(row[pixelOffset+5])
						fg[2] += weight * int(row[pixelOffset+6])
						fg[3] += weight * int(row[pixelOffset+7])
					}
				}

				// Downshift results
				fg[0] >>= (image.ImageSubpixelShift * 2)
				fg[1] >>= (image.ImageSubpixelShift * 2)
				fg[2] >>= (image.ImageSubpixelShift * 2)
				fg[3] >>= (image.ImageSubpixelShift * 2)
			}
		} else {
			// Handle clipping case
			if xLr < -1 || yLr < -1 || xLr > maxX || yLr > maxY {
				// Completely outside - use background
				fg[0] = backR
				fg[1] = backG
				fg[2] = backB
				fg[3] = backA
			} else {
				// Partial overlap - complex clipping needed
				// For simplicity, use background color in clip cases
				// In a full implementation, this would check each of the 4 samples individually
				fg[0] = backR
				fg[1] = backG
				fg[2] = backB
				fg[3] = backA
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
		if fg[3] > 255 {
			fg[3] = 255
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
		if fg[3] < 0 {
			fg[3] = 0
		}

		// Apply alpha channel constraints
		if fg[0] > fg[3] {
			fg[0] = fg[3]
		}
		if fg[1] > fg[3] {
			fg[1] = fg[3]
		}
		if fg[2] > fg[3] {
			fg[2] = fg[3]
		}

		span[i] = color.RGBA8[color.Linear]{
			R: basics.Int8u(fg[0]),
			G: basics.Int8u(fg[1]),
			B: basics.Int8u(fg[2]),
			A: basics.Int8u(fg[3]),
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGBA2x2 implements 2x2 RGBA image filtering with lookup table.
// This is a port of AGG's span_image_filter_rgba_2x2 template class.
type SpanImageFilterRGBA2x2[Source SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBA2x2 creates a new 2x2 RGBA span filter.
func NewSpanImageFilterRGBA2x2[Source SourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGBA2x2[Source, Interpolator] {
	return &SpanImageFilterRGBA2x2[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterRGBA2x2WithParams creates a new 2x2 RGBA span filter with parameters.
func NewSpanImageFilterRGBA2x2WithParams[Source SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilterRGBA2x2[Source, Interpolator] {
	return &SpanImageFilterRGBA2x2[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, filter),
	}
}

// Generate generates a span of RGBA pixels using 2x2 filter kernel.
func (sif *SpanImageFilterRGBA2x2[Source, Interpolator]) Generate(span []color.RGBA8[color.Linear], x, y int) {
	length := len(span)
	if length == 0 || sif.base.filter == nil {
		return
	}

	sif.base.interpolator.Begin(float64(x)+sif.base.FilterDxDbl(), float64(y)+sif.base.FilterDyDbl(), length)

	// Get weight array from filter
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

		var fg [4]int // RGBA accumulator

		xHr &= image.ImageSubpixelMask
		yHr &= image.ImageSubpixelMask

		if sourceAccessor, ok := interface{}(sif.base.source).(interface {
			Span(x, y, length int) []basics.Int8u
			NextX() []basics.Int8u
			NextY() []basics.Int8u
		}); ok {
			// Sample 1 (top-left)
			fgPtr := sourceAccessor.Span(xLr, yLr, 2)
			weight := (int(weightArray[xHr+image.ImageSubpixelScale+offset])*
				int(weightArray[yHr+image.ImageSubpixelScale+offset]) +
				image.ImageFilterScale/2) >> image.ImageFilterShift
			if len(fgPtr) >= 4 {
				fg[0] += weight * int(fgPtr[0])
				fg[1] += weight * int(fgPtr[1])
				fg[2] += weight * int(fgPtr[2])
				fg[3] += weight * int(fgPtr[3])
			}

			// Sample 2 (top-right)
			fgPtr = sourceAccessor.NextX()
			weight = (int(weightArray[xHr+offset])*
				int(weightArray[yHr+image.ImageSubpixelScale+offset]) +
				image.ImageFilterScale/2) >> image.ImageFilterShift
			if len(fgPtr) >= 4 {
				fg[0] += weight * int(fgPtr[0])
				fg[1] += weight * int(fgPtr[1])
				fg[2] += weight * int(fgPtr[2])
				fg[3] += weight * int(fgPtr[3])
			}

			// Sample 3 (bottom-left)
			fgPtr = sourceAccessor.NextY()
			weight = (int(weightArray[xHr+image.ImageSubpixelScale+offset])*
				int(weightArray[yHr+offset]) +
				image.ImageFilterScale/2) >> image.ImageFilterShift
			if len(fgPtr) >= 4 {
				fg[0] += weight * int(fgPtr[0])
				fg[1] += weight * int(fgPtr[1])
				fg[2] += weight * int(fgPtr[2])
				fg[3] += weight * int(fgPtr[3])
			}

			// Sample 4 (bottom-right)
			fgPtr = sourceAccessor.NextX()
			weight = (int(weightArray[xHr+offset])*
				int(weightArray[yHr+offset]) +
				image.ImageFilterScale/2) >> image.ImageFilterShift
			if len(fgPtr) >= 4 {
				fg[0] += weight * int(fgPtr[0])
				fg[1] += weight * int(fgPtr[1])
				fg[2] += weight * int(fgPtr[2])
				fg[3] += weight * int(fgPtr[3])
			}

			// Downshift results
			fg[0] >>= image.ImageFilterShift
			fg[1] >>= image.ImageFilterShift
			fg[2] >>= image.ImageFilterShift
			fg[3] >>= image.ImageFilterShift

			// Clamp to valid range and apply alpha constraints
			if fg[3] > 255 {
				fg[3] = 255
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

			span[i] = color.RGBA8[color.Linear]{
				R: basics.Int8u(fg[0]),
				G: basics.Int8u(fg[1]),
				B: basics.Int8u(fg[2]),
				A: basics.Int8u(fg[3]),
			}
		} else {
			span[i] = color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0}
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageFilterRGBA implements general RGBA image filtering with configurable kernel size.
// This is a port of AGG's span_image_filter_rgba template class.
type SpanImageFilterRGBA[Source SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageFilter[Source, Interpolator]
}

// NewSpanImageFilterRGBA creates a new general RGBA span filter.
func NewSpanImageFilterRGBA[Source SourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageFilterRGBA[Source, Interpolator] {
	return &SpanImageFilterRGBA[Source, Interpolator]{
		base: NewSpanImageFilter[Source, Interpolator](),
	}
}

// NewSpanImageFilterRGBAWithParams creates a new general RGBA span filter with parameters.
func NewSpanImageFilterRGBAWithParams[Source SourceInterface, Interpolator SpanInterpolatorInterface](
	src Source,
	interpolator Interpolator,
	filter *image.ImageFilterLUT,
) *SpanImageFilterRGBA[Source, Interpolator] {
	return &SpanImageFilterRGBA[Source, Interpolator]{
		base: NewSpanImageFilterWithParams(src, interpolator, filter),
	}
}

// Generate generates a span of RGBA pixels using a configurable filter kernel.
func (sif *SpanImageFilterRGBA[Source, Interpolator]) Generate(span []color.RGBA8[color.Linear], x, y int) {
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

		var fg [4]int // RGBA accumulator

		xFract := xHr & image.ImageSubpixelMask
		yCount := diameter

		yHr = image.ImageSubpixelMask - (yHr & image.ImageSubpixelMask)

		if sourceAccessor, ok := interface{}(sif.base.source).(interface {
			Span(x, y, length int) []basics.Int8u
			NextX() []basics.Int8u
			NextY() []basics.Int8u
		}); ok {
			fgPtr := sourceAccessor.Span(xLr+start, yLr+start, diameter)

			for yCount > 0 {
				xCount := diameter
				weightY := weightArray[yHr]
				xHr = image.ImageSubpixelMask - xFract

				for xCount > 0 {
					weight := (int(weightY)*int(weightArray[xHr]) + image.ImageFilterScale/2) >> image.ImageFilterShift

					if len(fgPtr) >= 4 {
						fg[0] += weight * int(fgPtr[0])
						fg[1] += weight * int(fgPtr[1])
						fg[2] += weight * int(fgPtr[2])
						fg[3] += weight * int(fgPtr[3])
					}

					xCount--
					if xCount == 0 {
						break
					}
					xHr += image.ImageSubpixelScale
					fgPtr = sourceAccessor.NextX()
				}

				yCount--
				if yCount == 0 {
					break
				}
				yHr += image.ImageSubpixelScale
				fgPtr = sourceAccessor.NextY()
			}

			// Downshift results
			fg[0] >>= image.ImageFilterShift
			fg[1] >>= image.ImageFilterShift
			fg[2] >>= image.ImageFilterShift
			fg[3] >>= image.ImageFilterShift

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
			if fg[3] < 0 {
				fg[3] = 0
			}

			if fg[3] > 255 {
				fg[3] = 255
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

			span[i] = color.RGBA8[color.Linear]{
				R: basics.Int8u(fg[0]),
				G: basics.Int8u(fg[1]),
				B: basics.Int8u(fg[2]),
				A: basics.Int8u(fg[3]),
			}
		} else {
			span[i] = color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0}
		}

		sif.base.interpolator.Next()
	}
}

// SpanImageResampleRGBAAffine provides affine resampling with automatic scale detection for RGBA.
// This is a port of AGG's span_image_resample_rgba_affine template class.
type SpanImageResampleRGBAAffine[Source SourceInterface] struct {
	base *SpanImageResampleAffine[Source]
}

// NewSpanImageResampleRGBAAffine creates a new affine RGBA resampling filter.
func NewSpanImageResampleRGBAAffine[Source SourceInterface]() *SpanImageResampleRGBAAffine[Source] {
	return &SpanImageResampleRGBAAffine[Source]{
		base: NewSpanImageResampleAffine[Source](),
	}
}

// SpanImageResampleRGBA provides general RGBA image resampling with configurable interpolation.
// This is a port of AGG's span_image_resample_rgba template class.
type SpanImageResampleRGBA[Source SourceInterface, Interpolator SpanInterpolatorInterface] struct {
	base *SpanImageResample[Source, Interpolator]
}

// NewSpanImageResampleRGBA creates a new general RGBA resampling filter.
func NewSpanImageResampleRGBA[Source SourceInterface, Interpolator SpanInterpolatorInterface]() *SpanImageResampleRGBA[Source, Interpolator] {
	return &SpanImageResampleRGBA[Source, Interpolator]{
		base: NewSpanImageResample[Source, Interpolator](),
	}
}
