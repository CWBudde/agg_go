package agg2d

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

// Source-linked premultiply/demultiply tests for the float Agg2D path.
//
// These tie the float color (color.RGBA32, the Go equivalent of C++ agg::rgba32
// selected by AGG2D_USE_FLOAT_FORMAT) and the ImageFloat boundary helper to the
// exact C++ formulas:
//
//	agg_color_rgba.h  rgba32T<Colorspace>::premultiply()  (line ~1243)
//	    if (a < 1) { if (a <= 0) r=g=b=0; else { r*=a; g*=a; b*=a; } }
//	agg_color_rgba.h  rgba32T<Colorspace>::demultiply()   (line ~1262)
//	    if (a < 1) { if (a <= 0) r=g=b=0; else { r/=a; g/=a; b/=a; } }
//
// The `a < 1` guard is significant: a fully opaque pixel (a == 1) is left
// untouched by both operations, matching C++ exactly.
//
// NOTE: the transformed-image half of the PLAN.md §4.5 verification item
// (TransformImage* premul/demul behavior) stays deferred — affine/perspective
// image transforms are not yet ported to the float path (see image_float.go and
// docs/AGG_DELTAS.md "Float Agg2D Variant"). This file covers only the
// straight↔premultiplied boundary contract.

func TestRGBA32PremultiplyDemultiplySourceLinked(t *testing.T) {
	const tol = 1e-6
	cases := []struct {
		name                string
		r, g, b, a          float32
		wantR, wantG, wantB float32 // expected premultiplied RGB
	}{
		// a in (0,1): r*=a etc.
		{"half-alpha", 1.0, 0.5, 0.25, 0.5, 0.5, 0.25, 0.125},
		// a == 1: the `a < 1` guard leaves RGB untouched.
		{"opaque-untouched", 0.8, 0.4, 0.2, 1.0, 0.8, 0.4, 0.2},
		// a == 0: RGB zeroed.
		{"zero-alpha-zeroes-rgb", 0.9, 0.7, 0.3, 0.0, 0.0, 0.0, 0.0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := color.NewRGBA32[color.Linear](tc.r, tc.g, tc.b, tc.a)
			c.Premultiply()
			if !feqTol(c.R, tc.wantR, tol) || !feqTol(c.G, tc.wantG, tol) || !feqTol(c.B, tc.wantB, tol) {
				t.Fatalf("Premultiply() = {%v,%v,%v,a=%v}, want RGB {%v,%v,%v}",
					c.R, c.G, c.B, c.A, tc.wantR, tc.wantG, tc.wantB)
			}
			if !feqTol(c.A, tc.a, tol) {
				t.Fatalf("Premultiply() changed alpha: got %v, want %v", c.A, tc.a)
			}

			// Demultiply must recover the straight values whenever a > 0.
			c.Demultiply()
			if tc.a > 0 {
				if !feqTol(c.R, tc.r, tol) || !feqTol(c.G, tc.g, tol) || !feqTol(c.B, tc.b, tol) {
					t.Fatalf("Demultiply() round-trip = {%v,%v,%v}, want {%v,%v,%v}",
						c.R, c.G, c.B, tc.r, tc.g, tc.b)
				}
			} else if !feqTol(c.R, 0, tol) || !feqTol(c.G, 0, tol) || !feqTol(c.B, 0, tol) {
				t.Fatalf("Demultiply() with a==0 = {%v,%v,%v}, want zeros", c.R, c.G, c.B)
			}
		})
	}
}

// TestImageFloatPremultiplyMatchesPerPixelColor verifies the ImageFloat boundary
// helper applies the same C++ rgba32 premultiply per pixel, including the opaque
// `a == 1` no-op and the `a == 0` zeroing, across a multi-pixel buffer.
func TestImageFloatPremultiplyMatchesPerPixelColor(t *testing.T) {
	img := NewImageFloatEmpty(3, 1)
	img.SetPixel(0, 0, color.NewRGBA32[color.Linear](1.0, 0.5, 0.25, 0.5)) // partial
	img.SetPixel(1, 0, color.NewRGBA32[color.Linear](0.8, 0.4, 0.2, 1.0))  // opaque
	img.SetPixel(2, 0, color.NewRGBA32[color.Linear](0.9, 0.7, 0.3, 0.0))  // transparent

	img.Premultiply()
	wantFloatPixel(t, img, 0, 0, 0.5, 0.25, 0.125, 0.5) // r*=a
	wantFloatPixel(t, img, 1, 0, 0.8, 0.4, 0.2, 1.0)    // a==1: untouched
	wantFloatPixel(t, img, 2, 0, 0, 0, 0, 0)            // a==0: RGB zeroed

	img.Demultiply()
	wantFloatPixel(t, img, 0, 0, 1.0, 0.5, 0.25, 0.5) // recovered
	wantFloatPixel(t, img, 1, 0, 0.8, 0.4, 0.2, 1.0)  // a==1: untouched
	wantFloatPixel(t, img, 2, 0, 0, 0, 0, 0)          // a==0 stays zeroed
}
