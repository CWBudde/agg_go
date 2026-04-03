package agg

import (
	"math"

	"github.com/cwbudde/agg_go/internal/blur"
	"github.com/cwbudde/agg_go/internal/color"
)

// ---------------------------------------------------------------------------
// Image interfaces — re-exported from internal/blur
//
// These mirror the implicit template contracts from C++ AGG.  In C++ an image
// processing algorithm such as stack_blur is templated on Img and calls
// img.pixel(x, y) and img.copy_color_hspan(); the compiler resolves the
// concrete pixfmt at instantiation.  In Go we express the same contract as an
// explicit interface.
// ---------------------------------------------------------------------------

// PixelReader is the read-only image contract used by analysis algorithms
// (e.g. Sobel gradient).  Any AGG pixel-format adapter satisfies this.
type PixelReader[C any] = blur.PixelReader[C]

// PixelReadWriter extends PixelReader with the span-write capabilities that
// C++ AGG's in-place blur algorithms require.
type PixelReadWriter[C any] = blur.PixelReadWriter[C]

// LuminanceFunc converts a color value to a scalar luminance in [0, 1].
// It separates the colour-to-luminance conversion from the analysis algorithm,
// following AGG's pattern of parameterising algorithms on calculator types.
type LuminanceFunc[C any] func(C) float64

// ---------------------------------------------------------------------------
// Predefined luminance extractors
// ---------------------------------------------------------------------------

// LuminanceRGBA8Linear returns a LuminanceFunc for color.RGBA8[color.Linear],
// using BT.709 weights (the same weights AGG uses internally).
func LuminanceRGBA8Linear() LuminanceFunc[color.RGBA8[color.Linear]] {
	return func(c color.RGBA8[color.Linear]) float64 {
		return (float64(c.R)*0.2126 + float64(c.G)*0.7152 + float64(c.B)*0.0722) / 255.0
	}
}

// LuminanceRGBA8SRGB returns a LuminanceFunc for color.RGBA8[color.SRGB].
// It linearises channels through the sRGB LUT before applying BT.709 weights.
func LuminanceRGBA8SRGB() LuminanceFunc[color.RGBA8[color.SRGB]] {
	return func(c color.RGBA8[color.SRGB]) float64 {
		v := color.LuminanceFromRGBA8SRGB(c)
		return float64(v) / 255.0
	}
}

// ---------------------------------------------------------------------------
// Sobel gradient
// ---------------------------------------------------------------------------

// sobelKernelX and sobelKernelY are the 3×3 Sobel kernels applied to the
// luminance channel to estimate the image gradient in the X and Y directions.
var (
	sobelKernelX = [3][3]float64{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	sobelKernelY = [3][3]float64{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}
)

// SobelGradient computes per-pixel Sobel gradient magnitudes from any readable
// pixel format.  It reads pixels through img.Pixel(x, y) and converts each to
// luminance via lum, exactly as a C++ AGG template would.
//
// The returned slice has one float32 per pixel in the range [0, 1], where 1
// represents the theoretical maximum gradient magnitude achievable for 8-bit
// input data (≈ 1442 luminance units).  Boundary pixels are handled by
// clamping (border replication).
func SobelGradient[C any, Img PixelReader[C]](img Img, lum LuminanceFunc[C]) []float32 {
	w, h := img.Width(), img.Height()
	out := make([]float32, w*h)
	const maxMag = 4.0 * math.Sqrt2 // Sobel max for luminance in [0,1]

	for y := range h {
		for x := range w {
			var gx, gy float64
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					nx := clampInt(x+kx, 0, w-1)
					ny := clampInt(y+ky, 0, h-1)
					l := lum(img.Pixel(nx, ny))
					gx += l * sobelKernelX[ky+1][kx+1]
					gy += l * sobelKernelY[ky+1][kx+1]
				}
			}
			mag := math.Sqrt(gx*gx+gy*gy) / maxMag
			if mag > 1.0 {
				mag = 1.0
			}
			out[y*w+x] = float32(mag)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
