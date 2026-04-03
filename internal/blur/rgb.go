package blur

import (
	"unsafe"

	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/effects"
)

// ---------------------------------------------------------------------------
// StackBlurRGB8 — mirrors C++ AGG's stack_blur<rgb8>
// ---------------------------------------------------------------------------

type StackBlurRGB8[CS color.Space] struct {
	buf   *array.PodVector[color.RGB8[CS]]
	stack *array.PodVector[color.RGB8[CS]]
}

func NewStackBlurRGB8[CS color.Space]() *StackBlurRGB8[CS] {
	return &StackBlurRGB8[CS]{
		buf:   array.NewPodVector[color.RGB8[CS]](),
		stack: array.NewPodVector[color.RGB8[CS]](),
	}
}

func (sb *StackBlurRGB8[CS]) BlurX(img PixelReadWriter[color.RGB8[CS]], radius int) {
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

func (sb *StackBlurRGB8[CS]) BlurY(img PixelReadWriter[color.RGB8[CS]], radius int) {
	sb.BlurX(&Transposer[color.RGB8[CS]]{Img: img}, radius)
}

func (sb *StackBlurRGB8[CS]) Blur(img PixelReadWriter[color.RGB8[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// RawRGB8Image adapts a flat RGB8 byte slice to PixelReadWriter.
type RawRGB8Image[CS color.Space] struct {
	Pixels []byte
	W, H   int
	Stride int
}

func (r *RawRGB8Image[CS]) Width() int  { return r.W }
func (r *RawRGB8Image[CS]) Height() int { return r.H }

func (r *RawRGB8Image[CS]) Pixel(x, y int) color.RGB8[CS] {
	off := y*r.Stride + x*3
	return color.NewRGB8[CS](r.Pixels[off], r.Pixels[off+1], r.Pixels[off+2])
}

func (r *RawRGB8Image[CS]) CopyColorHspan(x, y, length int, colors []color.RGB8[CS]) {
	off := y*r.Stride + x*3
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*3)
	copy(r.Pixels[off:off+len(colors)*3], src)
}

func (r *RawRGB8Image[CS]) CopyColorVspan(x, y, length int, colors []color.RGB8[CS]) {
	off := y*r.Stride + x*3
	for _, c := range colors {
		r.Pixels[off] = c.R
		r.Pixels[off+1] = c.G
		r.Pixels[off+2] = c.B
		off += r.Stride
	}
}

func (sb *StackBlurRGB8[CS]) BlurRGB8(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*3 || len(pixels) < (h-1)*stride+w*3 {
		return
	}
	sb.Blur(&RawRGB8Image[CS]{Pixels: pixels, W: w, H: h, Stride: stride}, radius)
}

// ---------------------------------------------------------------------------
// StackBlurRGB16 — mirrors C++ AGG's stack_blur<rgb16>
// ---------------------------------------------------------------------------

type StackBlurRGB16[CS color.Space] struct {
	buf   *array.PodVector[color.RGB16[CS]]
	stack *array.PodVector[color.RGB16[CS]]
}

func NewStackBlurRGB16[CS color.Space]() *StackBlurRGB16[CS] {
	return &StackBlurRGB16[CS]{
		buf:   array.NewPodVector[color.RGB16[CS]](),
		stack: array.NewPodVector[color.RGB16[CS]](),
	}
}

func (sb *StackBlurRGB16[CS]) BlurX(img PixelReadWriter[color.RGB16[CS]], radius int) {
	if radius < 1 {
		return
	}
	radius = min(radius, 254)

	w := img.Width()
	h := img.Height()
	wm := w - 1
	div := radius*2 + 1

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

func (sb *StackBlurRGB16[CS]) BlurY(img PixelReadWriter[color.RGB16[CS]], radius int) {
	sb.BlurX(&Transposer[color.RGB16[CS]]{Img: img}, radius)
}

func (sb *StackBlurRGB16[CS]) Blur(img PixelReadWriter[color.RGB16[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

type RawRGB16Image[CS color.Space] struct {
	Pixels []byte
	W, H   int
	Stride int
}

func (r *RawRGB16Image[CS]) Width() int  { return r.W }
func (r *RawRGB16Image[CS]) Height() int { return r.H }

func (r *RawRGB16Image[CS]) Pixel(x, y int) color.RGB16[CS] {
	off := y*r.Stride + x*6
	return *(*color.RGB16[CS])(unsafe.Pointer(&r.Pixels[off]))
}

func (r *RawRGB16Image[CS]) CopyColorHspan(x, y, length int, colors []color.RGB16[CS]) {
	off := y*r.Stride + x*6
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*6)
	copy(r.Pixels[off:off+len(colors)*6], src)
}

func (r *RawRGB16Image[CS]) CopyColorVspan(x, y, length int, colors []color.RGB16[CS]) {
	off := y*r.Stride + x*6
	for _, c := range colors {
		*(*color.RGB16[CS])(unsafe.Pointer(&r.Pixels[off])) = c
		off += r.Stride
	}
}

func (sb *StackBlurRGB16[CS]) BlurRGB16(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*6 || len(pixels) < (h-1)*stride+w*6 {
		return
	}
	sb.Blur(&RawRGB16Image[CS]{Pixels: pixels, W: w, H: h, Stride: stride}, radius)
}
