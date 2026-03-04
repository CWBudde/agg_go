// Common helpers shared by image demos (image1, image_transforms, image_alpha, pattern_fill).
package main

import (
	"math"

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
		x, y, r float64
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

		// Sphere body
		n := 16
		for i := 0; i < n; i++ {
			t := float64(i) / float64(n-1) // 0..1 from highlight to shadow
			fac := 1.0 - t
			ri := sp.r * (1.0 - t*0.0) // same radius, we just change fill color
			// Highlight: lighter color at top-left
			cr := sp.r0*fac + 0.05*(1-fac)
			cg := sp.g0*fac + 0.05*(1-fac)
			cb := sp.b0*fac + 0.05*(1-fac)
			alpha := 1.0 / float64(n) * 2.0
			imgCtx.SetColor(agg.RGBA(cr, cg, cb, alpha))
			imgCtx.FillCircle(sp.x, sp.y, ri)
			_ = ri
		}
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

// createStarPath fills a path with a star (n-pointed) centered at (cx,cy)
// with outer radius r1 and inner radius r2. startAngle is in degrees.
func createStarPath(ps interface {
	MoveTo(x, y float64)
	LineTo(x, y float64)
	ClosePolygon()
}, cx, cy, r1, r2 float64, n int, startAngleDeg float64) {
	start := startAngleDeg * math.Pi / 180.0
	for i := 0; i < n; i++ {
		a := math.Pi*2.0*float64(i)/float64(n) - math.Pi/2.0 + start
		var dx, dy float64
		if i&1 != 0 {
			dx = math.Cos(a) * r1
			dy = math.Sin(a) * r1
			ps.LineTo(cx+dx, cy+dy)
		} else {
			dx = math.Cos(a) * r2
			dy = math.Sin(a) * r2
			if i == 0 {
				ps.MoveTo(cx+dx, cy+dy)
			} else {
				ps.LineTo(cx+dx, cy+dy)
			}
		}
	}
	ps.ClosePolygon()
}
