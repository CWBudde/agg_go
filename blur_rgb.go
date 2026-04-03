package agg

import (
	"unsafe"

	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/effects"
)

// ---------------------------------------------------------------------------
// StackBlurRGB — stack blur for 8-bit RGB pixel formats (3 bytes per pixel).
//
// Mirrors the RGBA StackBlur but uses effects.StackBlurCalcRGB and
// color.RGB8 instead of the four-channel variants.
// ---------------------------------------------------------------------------

// StackBlurRGB implements AGG's stack_blur for RGB8 images.  It is reusable:
// temporary buffers are kept between calls to avoid repeated allocations, just
// like the C++ original.
type StackBlurRGB[CS color.Space] struct {
	buf   *array.PodVector[color.RGB8[CS]]
	stack *array.PodVector[color.RGB8[CS]]
}

// NewStackBlurRGB creates a new StackBlurRGB instance.
func NewStackBlurRGB[CS color.Space]() *StackBlurRGB[CS] {
	return &StackBlurRGB[CS]{
		buf:   array.NewPodVector[color.RGB8[CS]](),
		stack: array.NewPodVector[color.RGB8[CS]](),
	}
}

// BlurX applies a horizontal stack blur pass for RGB8 images.
func (sb *StackBlurRGB[CS]) BlurX(img PixelReadWriter[color.RGB8[CS]], radius int) {
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
		var sum, sumIn, sumOut effects.StackBlurCalcRGB[uint32, CS]
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
			var result color.RGB8[CS]
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
func (sb *StackBlurRGB[CS]) BlurY(img PixelReadWriter[color.RGB8[CS]], radius int) {
	sb.BlurX(&pixelImageTransposer[color.RGB8[CS]]{img: img}, radius)
}

// Blur applies both horizontal and vertical blur passes.
func (sb *StackBlurRGB[CS]) Blur(img PixelReadWriter[color.RGB8[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// ---------------------------------------------------------------------------
// rawRGB8Image — raw byte-buffer adapter for RGB8 (3 bytes per pixel)
// ---------------------------------------------------------------------------

type rawRGB8Image[CS color.Space] struct {
	pixels []byte
	w, h   int
	stride int
}

func (r *rawRGB8Image[CS]) Width() int  { return r.w }
func (r *rawRGB8Image[CS]) Height() int { return r.h }

func (r *rawRGB8Image[CS]) Pixel(x, y int) color.RGB8[CS] {
	off := y*r.stride + x*3
	return color.NewRGB8[CS](r.pixels[off], r.pixels[off+1], r.pixels[off+2])
}

func (r *rawRGB8Image[CS]) CopyColorHspan(x, y, length int, colors []color.RGB8[CS]) {
	off := y*r.stride + x*3
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*3)
	copy(r.pixels[off:off+len(colors)*3], src)
}

func (r *rawRGB8Image[CS]) CopyColorVspan(x, y, length int, colors []color.RGB8[CS]) {
	off := y*r.stride + x*3
	for _, c := range colors {
		r.pixels[off] = c.R
		r.pixels[off+1] = c.G
		r.pixels[off+2] = c.B
		off += r.stride
	}
}

// BlurRGB8 applies a stack blur to a raw RGB8 byte buffer.  The stride
// parameter is the number of bytes per row; pass w*3 for tightly packed
// buffers.
func (sb *StackBlurRGB[CS]) BlurRGB8(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*3 || len(pixels) < (h-1)*stride+w*3 {
		return
	}
	sb.Blur(&rawRGB8Image[CS]{pixels: pixels, w: w, h: h, stride: stride}, radius)
}

// ---------------------------------------------------------------------------
// StackBlurRGB16 — stack blur for 16-bit RGB pixel formats (6 bytes per pixel).
// ---------------------------------------------------------------------------

// StackBlurRGB16 implements AGG's stack_blur for RGB16 images.  It is
// reusable: temporary buffers are kept between calls to avoid repeated
// allocations.
type StackBlurRGB16[CS color.Space] struct {
	buf   *array.PodVector[color.RGB16[CS]]
	stack *array.PodVector[color.RGB16[CS]]
}

// NewStackBlurRGB16 creates a new StackBlurRGB16 instance.
func NewStackBlurRGB16[CS color.Space]() *StackBlurRGB16[CS] {
	return &StackBlurRGB16[CS]{
		buf:   array.NewPodVector[color.RGB16[CS]](),
		stack: array.NewPodVector[color.RGB16[CS]](),
	}
}

// BlurX applies a horizontal stack blur pass for RGB16 images.
func (sb *StackBlurRGB16[CS]) BlurX(img PixelReadWriter[color.RGB16[CS]], radius int) {
	if radius < 1 {
		return
	}
	radius = min(radius, 254)

	w := img.Width()
	h := img.Height()
	wm := w - 1
	div := radius*2 + 1

	// 16-bit channels exceed the 8-bit mul/shr table range.
	divSum := (radius + 1) * (radius + 1)

	sb.buf.Allocate(w, 128)
	sb.stack.Allocate(div, 32)

	for y := range h {
		var sum, sumIn, sumOut effects.StackBlurCalcRGB16[uint64, CS]
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
			var result color.RGB16[CS]
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
func (sb *StackBlurRGB16[CS]) BlurY(img PixelReadWriter[color.RGB16[CS]], radius int) {
	sb.BlurX(&pixelImageTransposer[color.RGB16[CS]]{img: img}, radius)
}

// Blur applies both horizontal and vertical blur passes.
func (sb *StackBlurRGB16[CS]) Blur(img PixelReadWriter[color.RGB16[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// ---------------------------------------------------------------------------
// rawRGB16Image — raw byte-buffer adapter for RGB16 (6 bytes per pixel)
// ---------------------------------------------------------------------------

type rawRGB16Image[CS color.Space] struct {
	pixels []byte
	w, h   int
	stride int
}

func (r *rawRGB16Image[CS]) Width() int  { return r.w }
func (r *rawRGB16Image[CS]) Height() int { return r.h }

func (r *rawRGB16Image[CS]) Pixel(x, y int) color.RGB16[CS] {
	off := y*r.stride + x*6
	p := (*color.RGB16[CS])(unsafe.Pointer(&r.pixels[off]))
	return *p
}

func (r *rawRGB16Image[CS]) CopyColorHspan(x, y, length int, colors []color.RGB16[CS]) {
	off := y*r.stride + x*6
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*6)
	copy(r.pixels[off:off+len(colors)*6], src)
}

func (r *rawRGB16Image[CS]) CopyColorVspan(x, y, length int, colors []color.RGB16[CS]) {
	off := y*r.stride + x*6
	for _, c := range colors {
		p := (*color.RGB16[CS])(unsafe.Pointer(&r.pixels[off]))
		*p = c
		off += r.stride
	}
}

// BlurRGB16 applies a stack blur to a raw RGB16 byte buffer.  The stride
// parameter is the number of bytes per row; pass w*6 for tightly packed
// buffers.
func (sb *StackBlurRGB16[CS]) BlurRGB16(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*6 || len(pixels) < (h-1)*stride+w*6 {
		return
	}
	sb.Blur(&rawRGB16Image[CS]{pixels: pixels, w: w, h: h, stride: stride}, radius)
}
