package agg

import (
	"unsafe"

	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/effects"
)

// ---------------------------------------------------------------------------
// StackBlurRGBA16  —  16-bit variant of StackBlur
//
// Mirrors C++ AGG's stack_blur instantiated with rgba16 / stack_blur_calc_rgba
// using 64-bit accumulators.  The algorithm is identical to StackBlur; only the
// colour type and accumulator width differ.
// ---------------------------------------------------------------------------

// StackBlurRGBA16 implements AGG's stack_blur for RGBA16 images.  It is
// reusable: temporary buffers are kept between calls to avoid repeated
// allocations, just like the C++ original.  The CS type parameter selects the
// colour-space tag (color.Linear or color.SRGB).
type StackBlurRGBA16[CS color.Space] struct {
	buf   *array.PodVector[color.RGBA16[CS]]
	stack *array.PodVector[color.RGBA16[CS]]
}

// NewStackBlurRGBA16 creates a new StackBlurRGBA16 instance.
func NewStackBlurRGBA16[CS color.Space]() *StackBlurRGBA16[CS] {
	return &StackBlurRGBA16[CS]{
		buf:   array.NewPodVector[color.RGBA16[CS]](),
		stack: array.NewPodVector[color.RGBA16[CS]](),
	}
}

// BlurX applies a horizontal stack blur pass.  It reads pixels through
// img.Pixel(x, y) and writes blurred rows back with img.CopyColorHspan(),
// matching C++ AGG's stack_blur::blur_x exactly.
func (sb *StackBlurRGBA16[CS]) BlurX(img PixelReadWriter[color.RGBA16[CS]], radius int) {
	if radius < 1 {
		return
	}
	radius = min(radius, 254)

	w := img.Width()
	h := img.Height()
	wm := w - 1
	div := radius*2 + 1

	// 16-bit channels exceed the 8-bit mul/shr table range, so always
	// use the division path (matches C++ AGG's max_val > 255 check).
	divSum := (radius + 1) * (radius + 1)

	sb.buf.Allocate(w, 128)
	sb.stack.Allocate(div, 32)

	for y := range h {
		var sum, sumIn, sumOut effects.StackBlurCalcRGBA16[uint64, CS]
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
			var result color.RGBA16[CS]
			sum.CalcPix(&result, divSum)
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
func (sb *StackBlurRGBA16[CS]) BlurY(img PixelReadWriter[color.RGBA16[CS]], radius int) {
	sb.BlurX(&pixelImageTransposer[color.RGBA16[CS]]{img: img}, radius)
}

// Blur applies both horizontal and vertical blur passes.
func (sb *StackBlurRGBA16[CS]) Blur(img PixelReadWriter[color.RGBA16[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// ---------------------------------------------------------------------------
// rawRGBA16Image — adapts a flat RGBA16 byte slice (8 bytes per pixel) to the
// PixelReadWriter[color.RGBA16[CS]] interface.
// ---------------------------------------------------------------------------

type rawRGBA16Image[CS color.Space] struct {
	pixels []byte
	w, h   int
	stride int // bytes per row
}

func (r *rawRGBA16Image[CS]) Width() int  { return r.w }
func (r *rawRGBA16Image[CS]) Height() int { return r.h }

func (r *rawRGBA16Image[CS]) Pixel(x, y int) color.RGBA16[CS] {
	off := y*r.stride + x*8
	return *(*color.RGBA16[CS])(unsafe.Pointer(&r.pixels[off]))
}

func (r *rawRGBA16Image[CS]) CopyColorHspan(x, y, length int, colors []color.RGBA16[CS]) {
	off := y*r.stride + x*8
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*8)
	copy(r.pixels[off:off+len(colors)*8], src)
}

func (r *rawRGBA16Image[CS]) CopyColorVspan(x, y, length int, colors []color.RGBA16[CS]) {
	off := y*r.stride + x*8
	for _, c := range colors {
		*(*color.RGBA16[CS])(unsafe.Pointer(&r.pixels[off])) = c
		off += r.stride
	}
}

// BlurRGBA16 applies a stack blur to a raw RGBA16 byte buffer.  The stride
// parameter is the number of bytes per row; pass w*8 for tightly packed
// buffers.  The blur algorithm averages raw channel values and is agnostic to
// colour-space interpretation.
func (sb *StackBlurRGBA16[CS]) BlurRGBA16(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*8 || len(pixels) < (h-1)*stride+w*8 {
		return
	}
	sb.Blur(&rawRGBA16Image[CS]{pixels: pixels, w: w, h: h, stride: stride}, radius)
}
