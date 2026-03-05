// Common helpers shared by image demos (image1, image_transforms, image_alpha, pattern_fill).
package main

import (
	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/image"
)

// imageClipSource wraps imagePixFmt + ImageAccessorClip to implement the
// span.RGBASourceInterface needed by SpanImageFilterRGBABilinearClip.
type imageClipSource struct {
	accessor *image.ImageAccessorClip[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *imageClipSource) Width() int                  { return s.ipf.Width() }
func (s *imageClipSource) Height() int                 { return s.ipf.Height() }
func (s *imageClipSource) ColorType() string           { return "RGBA8" }
func (s *imageClipSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *imageClipSource) Span(x, y, length int) []basics.Int8u {
	return s.accessor.Span(x, y, length)
}
func (s *imageClipSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *imageClipSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *imageClipSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

// createSpheresImage creates a procedural image with colorful spheres on a dark background.
// This is used in place of spheres.bmp from the original AGG demos.
func createSpheresImage(w, h int) *agg.Image {
	img := agg.CreateImage(w, h)
	imgCtx := agg.NewContextForImage(img)

	// Dark gradient background
	imgCtx.SetColor(agg.RGBA(0.05, 0.05, 0.12, 1.0))
	imgCtx.FillRectangle(0, 0, float64(w), float64(h))

	type sphere struct {
		x, y, r    float64
		r0, g0, b0 float64
	}
	spheres := []sphere{
		{float64(w) * 0.22, float64(h) * 0.30, float64(w) * 0.18, 0.9, 0.2, 0.1},
		{float64(w) * 0.65, float64(h) * 0.28, float64(w) * 0.15, 0.1, 0.4, 0.9},
		{float64(w) * 0.45, float64(h) * 0.68, float64(w) * 0.20, 0.1, 0.8, 0.3},
		{float64(w) * 0.78, float64(h) * 0.65, float64(w) * 0.12, 0.9, 0.7, 0.1},
		{float64(w) * 0.15, float64(h) * 0.72, float64(w) * 0.10, 0.7, 0.1, 0.8},
	}

	for _, sp := range spheres {
		// Soft shadow
		imgCtx.SetColor(agg.RGBA(0, 0, 0, 0.35))
		imgCtx.FillCircle(sp.x+sp.r*0.15, sp.y+sp.r*0.15, sp.r)

		// Simple radial fill approximation with a filled circle
		imgCtx.SetColor(agg.RGBA(sp.r0, sp.g0, sp.b0, 0.85))
		imgCtx.FillCircle(sp.x, sp.y, sp.r)

		// Specular highlight
		hx := sp.x - sp.r*0.30
		hy := sp.y - sp.r*0.30
		hr := sp.r * 0.30
		imgCtx.SetColor(agg.RGBA(1.0, 1.0, 1.0, 0.6))
		imgCtx.FillCircle(hx, hy, hr)
	}

	return img
}
