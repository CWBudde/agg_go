// Package outline provides pattern filter adapters for outline rendering.
package outline

import (
	"agg_go/internal/color"
	"agg_go/internal/span"
)

// PatternFilterRGBAAdapter adapts span pattern filters to work with outline renderer Filter interface.
// This bridges the gap between the span package's pattern filters and the outline package's requirements.
type PatternFilterRGBAAdapter struct {
	filter *span.PatternFilterBilinearRGBA8[color.Linear]
}

// NewPatternFilterRGBAAdapter creates a new pattern filter adapter.
func NewPatternFilterRGBAAdapter() *PatternFilterRGBAAdapter {
	return &PatternFilterRGBAAdapter{
		filter: span.NewPatternFilterBilinearRGBA8[color.Linear](),
	}
}

// Dilation returns the filter dilation.
func (pfa *PatternFilterRGBAAdapter) Dilation() int {
	return pfa.filter.Dilation()
}

// PixelHighRes performs high-resolution pixel sampling with type conversion.
func (pfa *PatternFilterRGBAAdapter) PixelHighRes(rows [][]color.RGBA, p *color.RGBA, x, y int) {
	// Convert [][]color.RGBA to [][]color.RGBA8[Linear] for the filter
	if len(rows) == 0 || y < 0 || y >= len(rows) {
		*p = color.NewRGBA(0, 0, 0, 0)
		return
	}

	// Create converted buffer on the fly - only converting the rows we need
	convertedRows := make([][]color.RGBA8[color.Linear], len(rows))
	for i, row := range rows {
		if row != nil {
			convertedRow := make([]color.RGBA8[color.Linear], len(row))
			for j, pixel := range row {
				// Convert from color.RGBA to color.RGBA8[Linear]
				convertedRow[j] = color.RGBA8[color.Linear]{
					R: uint8(pixel.R * 255),
					G: uint8(pixel.G * 255),
					B: uint8(pixel.B * 255),
					A: uint8(pixel.A * 255),
				}
			}
			convertedRows[i] = convertedRow
		}
	}

	// Sample using the bilinear filter
	var rgba8Result color.RGBA8[color.Linear]
	pfa.filter.PixelHighRes(convertedRows, &rgba8Result, x, y)

	// Convert back to color.RGBA
	*p = color.NewRGBA(
		float64(rgba8Result.R)/255.0,
		float64(rgba8Result.G)/255.0,
		float64(rgba8Result.B)/255.0,
		float64(rgba8Result.A)/255.0,
	)
}
