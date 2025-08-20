// Package renderer provides specialized renderers for AGG.
package renderer

import (
	"agg_go"
	"agg_go/internal/basics"
	scanline_renderer "agg_go/internal/renderer/scanline"
)

// BaseRendererWithCopyBar extends BaseRendererInterface with CopyBar method
type BaseRendererWithCopyBar interface {
	scanline_renderer.BaseRendererInterface
	CopyBar(x1, y1, x2, y2 int, c interface{})
}

// RendererEnlarged is a specialized renderer that magnifies pixels for visualization.
// This renderer is used in the aa_demo example to show anti-aliasing coverage values.
// It implements the RendererInterface to work with the render_scanlines function.
type RendererEnlarged[Ren BaseRendererWithCopyBar] struct {
	baseRenderer Ren         // The underlying base renderer
	size         float64     // Size multiplier for magnification
	color        interface{} // Current color
}

// NewRendererEnlarged creates a new enlarged pixel renderer.
func NewRendererEnlarged[Ren BaseRendererWithCopyBar](
	baseRenderer Ren, size float64) *RendererEnlarged[Ren] {

	return &RendererEnlarged[Ren]{
		baseRenderer: baseRenderer,
		size:         size,
	}
}

// Color returns the current color for rendering.
func (r *RendererEnlarged[Ren]) Color() interface{} {
	return r.color
}

// SetColor sets the current color for rendering.
func (r *RendererEnlarged[Ren]) SetColor(color interface{}) {
	r.color = color
}

// Prepare prepares the renderer for rendering (no-op for this renderer).
func (r *RendererEnlarged[Ren]) Prepare() {
	// Nothing to prepare for this renderer
}

// Render renders a scanline with pixel magnification.
// This method processes each pixel in the scanline and renders a magnified version.
func (r *RendererEnlarged[Ren]) Render(sl scanline_renderer.ScanlineInterface) {
	y := sl.Y()
	numSpans := sl.NumSpans()

	if numSpans == 0 {
		return
	}

	iter := sl.Begin()

	// Process each span in the scanline
	for i := 0; i < numSpans; i++ {
		span := iter.GetSpan()
		x := span.X
		covers := span.Covers
		numPix := span.Len

		// Process each pixel in the span
		for j := 0; j < numPix; j++ {
			if j < len(covers) {
				cover := covers[j]

				// Calculate alpha based on coverage and current color
				var alpha basics.Int8u
				if color, ok := r.color.(agg.RGBA8); ok {
					alpha = basics.Int8u((int(cover) * int(color.A)) >> 8)
				} else {
					alpha = cover
				}

				// Create a color with the calculated alpha
				renderColor := r.color
				if color, ok := r.color.(agg.RGBA8); ok {
					renderColor = agg.RGBA8{R: color.R, G: color.G, B: color.B, A: alpha}
				}

				// Draw the magnified pixel
				currentX := x + j
				r.drawMagnifiedPixel(float64(currentX), float64(y), renderColor)
			}
		}

		// Move to next span if not the last one
		if i < numSpans-1 {
			iter.Next()
		}
	}
}

// drawMagnifiedPixel draws a single magnified pixel.
// This is a helper method to draw the magnified representation of a pixel.
func (r *RendererEnlarged[Ren]) drawMagnifiedPixel(x, y float64, color interface{}) {
	// In the C++ version, they create a new rasterizer/scanline for each pixel
	// and draw a square using render_scanlines_aa_solid.
	// For simplicity in Go, we'll draw directly to the base renderer.

	// Calculate the magnified coordinates
	magX1 := int(x * r.size)
	magY1 := int(y * r.size)
	magX2 := int((x + 1) * r.size)
	magY2 := int((y + 1) * r.size)

	// Draw the magnified pixel as a filled rectangle
	r.baseRenderer.CopyBar(magX1, magY1, magX2-1, magY2-1, color)
}
