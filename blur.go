package agg

import (
	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/effects"
)

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

// CopyColorHspan in the transposed view maps to CopyColorVspan on the
// underlying image, matching C++ AGG's pixfmt_transposer::copy_color_hspan.
func (t *pixelImageTransposer[C]) CopyColorHspan(x, y, length int, colors []C) {
	t.img.CopyColorVspan(y, x, length, colors)
}

// CopyColorVspan in the transposed view maps to CopyColorHspan on the
// underlying image, matching C++ AGG's pixfmt_transposer::copy_color_vspan.
func (t *pixelImageTransposer[C]) CopyColorVspan(x, y, length int, colors []C) {
	t.img.CopyColorHspan(y, x, length, colors)
}

// ---------------------------------------------------------------------------
// Public convenience wrappers for raw RGBA8 byte buffers
// ---------------------------------------------------------------------------

// rawRGBA8Image adapts a flat RGBA8 byte slice (row-major, 4 bytes/pixel) to
// the PixelReadWriter[color.RGBA8[color.Linear]] interface so that callers
// outside the agg_go module can use StackBlur without importing internal types.
type rawRGBA8Image struct {
	pixels []byte
	w, h   int
}

func (r *rawRGBA8Image) Width() int  { return r.w }
func (r *rawRGBA8Image) Height() int { return r.h }

func (r *rawRGBA8Image) Pixel(x, y int) color.RGBA8[color.Linear] {
	i := (y*r.w + x) * 4
	return color.NewRGBA8[color.Linear](r.pixels[i], r.pixels[i+1], r.pixels[i+2], r.pixels[i+3])
}

func (r *rawRGBA8Image) CopyColorHspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	for j, c := range colors {
		i := (y*r.w + x + j) * 4
		r.pixels[i] = c.R
		r.pixels[i+1] = c.G
		r.pixels[i+2] = c.B
		r.pixels[i+3] = c.A
	}
}

func (r *rawRGBA8Image) CopyColorVspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	for j, c := range colors {
		i := ((y+j)*r.w + x) * 4
		r.pixels[i] = c.R
		r.pixels[i+1] = c.G
		r.pixels[i+2] = c.B
		r.pixels[i+3] = c.A
	}
}

// BlurRGBA8 applies a stack blur to a raw RGBA8 byte buffer (row-major,
// 4 bytes per pixel, len = w*h*4).  This is the recommended entry point for
// callers outside the agg_go module.
func (sb *StackBlur) BlurRGBA8(pixels []byte, w, h, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || len(pixels) < w*h*4 {
		return
	}
	sb.Blur(&rawRGBA8Image{pixels: pixels, w: w, h: h}, radius)
}
