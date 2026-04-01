package agg

import (
	"math"

	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/effects"
)

// ---------------------------------------------------------------------------
// Image interfaces
//
// These mirror the implicit template contracts from C++ AGG.  In C++ an image
// processing algorithm such as stack_blur is templated on Img and calls
// img.pixel(x, y) and img.copy_color_hspan(); the compiler resolves the
// concrete pixfmt at instantiation.  In Go we express the same contract as an
// explicit interface.
// ---------------------------------------------------------------------------

// PixelReader is the read-only image contract used by analysis algorithms
// (e.g. Sobel gradient).  Any AGG pixel-format adapter satisfies this.
type PixelReader[C any] interface {
	Width() int
	Height() int
	Pixel(x, y int) C
}

// PixelReadWriter extends PixelReader with the span-write capability that C++
// AGG's in-place blur algorithms require (pixel() for reading, copy_color_hspan()
// for writing processed rows back).
type PixelReadWriter[C any] interface {
	PixelReader[C]
	CopyColorHspan(x, y, length int, colors []C)
}

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
// Stack blur  —  mirrors C++ AGG's stack_blur class from agg_blur.h
//
// The C++ version is:
//
//   template<class ColorT=rgba8,
//            class CalculatorT=stack_blur_calc_rgba<unsigned>>
//   class stack_blur {
//       pod_vector<color_type> m_buf;
//       pod_vector<color_type> m_stack;
//   public:
//       template<class Img> void blur_x(Img& img, unsigned radius);
//       template<class Img> void blur_y(Img& img, unsigned radius);
//       template<class Img> void blur(Img& img, unsigned radius);
//   };
//
// blur_x reads via img.pixel(x,y) and writes via img.copy_color_hspan().
// blur_y delegates to blur_x through a pixfmt_transposer adapter.
// ---------------------------------------------------------------------------

// StackBlur implements AGG's stack_blur for RGBA8 images.  It is reusable:
// temporary buffers are kept between calls to avoid repeated allocations, just
// like the C++ original.
type StackBlur struct {
	buf   *array.PodVector[color.RGBA8[color.Linear]]
	stack *array.PodVector[color.RGBA8[color.Linear]]
}

// NewStackBlur creates a new StackBlur instance.
func NewStackBlur() *StackBlur {
	return &StackBlur{
		buf:   array.NewPodVector[color.RGBA8[color.Linear]](),
		stack: array.NewPodVector[color.RGBA8[color.Linear]](),
	}
}

// BlurX applies a horizontal stack blur pass.  It reads pixels through
// img.Pixel(x, y) and writes blurred rows back with img.CopyColorHspan(),
// matching C++ AGG's stack_blur::blur_x exactly.
func (sb *StackBlur) BlurX(img PixelReadWriter[color.RGBA8[color.Linear]], radius int) {
	if radius < 1 {
		return
	}
	radius = min(radius, 254)

	w := img.Width()
	h := img.Height()
	wm := w - 1
	div := radius*2 + 1

	divSum := (radius + 1) * (radius + 1)
	var mulSum, shrSum int
	if radius < 255 {
		mulSum = int(effects.StackBlur8Mul[radius])
		shrSum = int(effects.StackBlur8Shr[radius])
	}

	sb.buf.Allocate(w, 128)
	sb.stack.Allocate(div, 32)

	for y := range h {
		var sum, sumIn, sumOut effects.StackBlurCalcRGBA[uint32]
		sum.Clear()
		sumIn.Clear()
		sumOut.Clear()

		pix := img.Pixel(0, y)
		for i := 0; i <= radius; i++ {
			sb.stack.Set(i, pix)
			sum.AddWeighted(pix, i+1)
			sumOut.Add(pix)
		}
		for i := 1; i <= radius; i++ {
			pix = img.Pixel(min(i, wm), y)
			sb.stack.Set(i+radius, pix)
			sum.AddWeighted(pix, radius+1-i)
			sumIn.Add(pix)
		}

		stackPtr := radius
		for x := range w {
			var result color.RGBA8[color.Linear]
			if mulSum != 0 {
				sum.CalcPixMulShr(&result, mulSum, shrSum)
			} else {
				sum.CalcPix(&result, divSum)
			}
			sb.buf.Set(x, result)

			sum.SubCalc(sumOut)

			stackStart := stackPtr + div - radius
			if stackStart >= div {
				stackStart -= div
			}
			stackPix := sb.stack.ValueAt(stackStart)
			sumOut.Sub(stackPix)

			pix = img.Pixel(min(x+radius+1, wm), y)
			sb.stack.Set(stackStart, pix)

			sumIn.Add(pix)
			sum.AddCalc(sumIn)

			stackPtr++
			if stackPtr >= div {
				stackPtr = 0
			}
			stackPix = sb.stack.ValueAt(stackPtr)

			sumOut.Add(stackPix)
			sumIn.Sub(stackPix)
		}

		img.CopyColorHspan(0, y, w, sb.buf.Data()[:w])
	}
}

// BlurY applies a vertical stack blur pass.  Following the C++ AGG pattern,
// it wraps the image in a transposing adapter and delegates to BlurX.
func (sb *StackBlur) BlurY(img PixelReadWriter[color.RGBA8[color.Linear]], radius int) {
	sb.BlurX(&pixelImageTransposer[color.RGBA8[color.Linear]]{img: img}, radius)
}

// Blur applies both horizontal and vertical blur passes.
func (sb *StackBlur) Blur(img PixelReadWriter[color.RGBA8[color.Linear]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// ---------------------------------------------------------------------------
// pixelImageTransposer — Go equivalent of C++ AGG's pixfmt_transposer.
//
// It swaps the X/Y axes of a PixelReadWriter so that a horizontal blur pass
// applied to the transposed view becomes a vertical blur.  C++ AGG uses this
// same trick in stack_blur::blur_y.
// ---------------------------------------------------------------------------

type pixelImageTransposer[C any] struct {
	img PixelReadWriter[C]
}

func (t *pixelImageTransposer[C]) Width() int       { return t.img.Height() }
func (t *pixelImageTransposer[C]) Height() int      { return t.img.Width() }
func (t *pixelImageTransposer[C]) Pixel(x, y int) C { return t.img.Pixel(y, x) }
func (t *pixelImageTransposer[C]) CopyColorHspan(x, y, length int, colors []C) {
	// A horizontal span in the transposed view is a vertical span in the
	// original image.  Write pixel-by-pixel since pixfmt only guarantees
	// horizontal span writes.
	for i := range length {
		t.img.CopyColorHspan(y, x+i, 1, colors[i:i+1])
	}
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
