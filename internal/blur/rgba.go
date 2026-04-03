package blur

import (
	"unsafe"

	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/effects"
)

// ---------------------------------------------------------------------------
// StackBlurRGBA8 — mirrors C++ AGG's stack_blur<rgba8>
// ---------------------------------------------------------------------------

// StackBlurRGBA8 implements AGG's stack_blur for RGBA8 images.  It is
// reusable: temporary buffers are kept between calls to avoid repeated
// allocations, just like the C++ original.
type StackBlurRGBA8[CS color.Space] struct {
	buf   *array.PodVector[color.RGBA8[CS]]
	stack *array.PodVector[color.RGBA8[CS]]
}

func NewStackBlurRGBA8[CS color.Space]() *StackBlurRGBA8[CS] {
	return &StackBlurRGBA8[CS]{
		buf:   array.NewPodVector[color.RGBA8[CS]](),
		stack: array.NewPodVector[color.RGBA8[CS]](),
	}
}

func (sb *StackBlurRGBA8[CS]) BlurX(img PixelReadWriter[color.RGBA8[CS]], radius int) {
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
		var sum, sumIn, sumOut effects.StackBlurCalcRGBA[uint32, CS]
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
			var result color.RGBA8[CS]
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

func (sb *StackBlurRGBA8[CS]) BlurY(img PixelReadWriter[color.RGBA8[CS]], radius int) {
	sb.BlurX(&Transposer[color.RGBA8[CS]]{Img: img}, radius)
}

func (sb *StackBlurRGBA8[CS]) Blur(img PixelReadWriter[color.RGBA8[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// RawRGBA8Image adapts a flat RGBA8 byte slice to PixelReadWriter.
type RawRGBA8Image[CS color.Space] struct {
	Pixels []byte
	W, H   int
	Stride int
}

func (r *RawRGBA8Image[CS]) Width() int  { return r.W }
func (r *RawRGBA8Image[CS]) Height() int { return r.H }

func (r *RawRGBA8Image[CS]) Pixel(x, y int) color.RGBA8[CS] {
	i := y*r.Stride + x*4
	return color.NewRGBA8[CS](r.Pixels[i], r.Pixels[i+1], r.Pixels[i+2], r.Pixels[i+3])
}

func (r *RawRGBA8Image[CS]) CopyColorHspan(x, y, length int, colors []color.RGBA8[CS]) {
	off := y*r.Stride + x*4
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*4)
	copy(r.Pixels[off:off+len(colors)*4], src)
}

func (r *RawRGBA8Image[CS]) CopyColorVspan(x, y, length int, colors []color.RGBA8[CS]) {
	off := y*r.Stride + x*4
	for _, c := range colors {
		r.Pixels[off] = c.R
		r.Pixels[off+1] = c.G
		r.Pixels[off+2] = c.B
		r.Pixels[off+3] = c.A
		off += r.Stride
	}
}

func (sb *StackBlurRGBA8[CS]) BlurRGBA8(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*4 || len(pixels) < (h-1)*stride+w*4 {
		return
	}
	sb.Blur(&RawRGBA8Image[CS]{Pixels: pixels, W: w, H: h, Stride: stride}, radius)
}

// ---------------------------------------------------------------------------
// StackBlurRGBA16 — mirrors C++ AGG's stack_blur<rgba16>
// ---------------------------------------------------------------------------

type StackBlurRGBA16[CS color.Space] struct {
	buf   *array.PodVector[color.RGBA16[CS]]
	stack *array.PodVector[color.RGBA16[CS]]
}

func NewStackBlurRGBA16[CS color.Space]() *StackBlurRGBA16[CS] {
	return &StackBlurRGBA16[CS]{
		buf:   array.NewPodVector[color.RGBA16[CS]](),
		stack: array.NewPodVector[color.RGBA16[CS]](),
	}
}

func (sb *StackBlurRGBA16[CS]) BlurX(img PixelReadWriter[color.RGBA16[CS]], radius int) {
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

func (sb *StackBlurRGBA16[CS]) BlurY(img PixelReadWriter[color.RGBA16[CS]], radius int) {
	sb.BlurX(&Transposer[color.RGBA16[CS]]{Img: img}, radius)
}

func (sb *StackBlurRGBA16[CS]) Blur(img PixelReadWriter[color.RGBA16[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

type RawRGBA16Image[CS color.Space] struct {
	Pixels []byte
	W, H   int
	Stride int
}

func (r *RawRGBA16Image[CS]) Width() int  { return r.W }
func (r *RawRGBA16Image[CS]) Height() int { return r.H }

func (r *RawRGBA16Image[CS]) Pixel(x, y int) color.RGBA16[CS] {
	off := y*r.Stride + x*8
	return *(*color.RGBA16[CS])(unsafe.Pointer(&r.Pixels[off]))
}

func (r *RawRGBA16Image[CS]) CopyColorHspan(x, y, length int, colors []color.RGBA16[CS]) {
	off := y*r.Stride + x*8
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*8)
	copy(r.Pixels[off:off+len(colors)*8], src)
}

func (r *RawRGBA16Image[CS]) CopyColorVspan(x, y, length int, colors []color.RGBA16[CS]) {
	off := y*r.Stride + x*8
	for _, c := range colors {
		*(*color.RGBA16[CS])(unsafe.Pointer(&r.Pixels[off])) = c
		off += r.Stride
	}
}

func (sb *StackBlurRGBA16[CS]) BlurRGBA16(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*8 || len(pixels) < (h-1)*stride+w*8 {
		return
	}
	sb.Blur(&RawRGBA16Image[CS]{Pixels: pixels, W: w, H: h, Stride: stride}, radius)
}
