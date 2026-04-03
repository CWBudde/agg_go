package blur

import (
	"unsafe"

	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/effects"
)

// ---------------------------------------------------------------------------
// StackBlurGray8 — mirrors C++ AGG's stack_blur<gray8>
// ---------------------------------------------------------------------------

type StackBlurGray8[CS color.Space] struct {
	buf   *array.PodVector[color.Gray8[CS]]
	stack *array.PodVector[color.Gray8[CS]]
}

func NewStackBlurGray8[CS color.Space]() *StackBlurGray8[CS] {
	return &StackBlurGray8[CS]{
		buf:   array.NewPodVector[color.Gray8[CS]](),
		stack: array.NewPodVector[color.Gray8[CS]](),
	}
}

func (sb *StackBlurGray8[CS]) BlurX(img PixelReadWriter[color.Gray8[CS]], radius int) {
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
		var sum, sumIn, sumOut effects.StackBlurCalcGray[uint32, CS]
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
			var result color.Gray8[CS]
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

func (sb *StackBlurGray8[CS]) BlurY(img PixelReadWriter[color.Gray8[CS]], radius int) {
	sb.BlurX(&Transposer[color.Gray8[CS]]{Img: img}, radius)
}

func (sb *StackBlurGray8[CS]) Blur(img PixelReadWriter[color.Gray8[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// RawGray8Image adapts a flat Gray8 byte slice (2 bytes per pixel: V, A) to
// PixelReadWriter.
type RawGray8Image[CS color.Space] struct {
	Pixels []byte
	W, H   int
	Stride int
}

func (r *RawGray8Image[CS]) Width() int  { return r.W }
func (r *RawGray8Image[CS]) Height() int { return r.H }

func (r *RawGray8Image[CS]) Pixel(x, y int) color.Gray8[CS] {
	off := y*r.Stride + x*2
	return color.Gray8[CS]{V: r.Pixels[off], A: r.Pixels[off+1]}
}

func (r *RawGray8Image[CS]) CopyColorHspan(x, y, length int, colors []color.Gray8[CS]) {
	off := y*r.Stride + x*2
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*2)
	copy(r.Pixels[off:off+len(colors)*2], src)
}

func (r *RawGray8Image[CS]) CopyColorVspan(x, y, length int, colors []color.Gray8[CS]) {
	off := y*r.Stride + x*2
	for _, c := range colors {
		r.Pixels[off] = c.V
		r.Pixels[off+1] = c.A
		off += r.Stride
	}
}

func (sb *StackBlurGray8[CS]) BlurGray8(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*2 || len(pixels) < (h-1)*stride+w*2 {
		return
	}
	sb.Blur(&RawGray8Image[CS]{Pixels: pixels, W: w, H: h, Stride: stride}, radius)
}

// ---------------------------------------------------------------------------
// StackBlurGray16 — mirrors C++ AGG's stack_blur<gray16>
// ---------------------------------------------------------------------------

type StackBlurGray16[CS color.Space] struct {
	buf   *array.PodVector[color.Gray16[CS]]
	stack *array.PodVector[color.Gray16[CS]]
}

func NewStackBlurGray16[CS color.Space]() *StackBlurGray16[CS] {
	return &StackBlurGray16[CS]{
		buf:   array.NewPodVector[color.Gray16[CS]](),
		stack: array.NewPodVector[color.Gray16[CS]](),
	}
}

func (sb *StackBlurGray16[CS]) BlurX(img PixelReadWriter[color.Gray16[CS]], radius int) {
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
		var sum, sumIn, sumOut effects.StackBlurCalcGray16[uint64, CS]
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
			var result color.Gray16[CS]
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

func (sb *StackBlurGray16[CS]) BlurY(img PixelReadWriter[color.Gray16[CS]], radius int) {
	sb.BlurX(&Transposer[color.Gray16[CS]]{Img: img}, radius)
}

func (sb *StackBlurGray16[CS]) Blur(img PixelReadWriter[color.Gray16[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

type RawGray16Image[CS color.Space] struct {
	Pixels []byte
	W, H   int
	Stride int
}

func (r *RawGray16Image[CS]) Width() int  { return r.W }
func (r *RawGray16Image[CS]) Height() int { return r.H }

func (r *RawGray16Image[CS]) Pixel(x, y int) color.Gray16[CS] {
	off := y*r.Stride + x*4
	return *(*color.Gray16[CS])(unsafe.Pointer(&r.Pixels[off]))
}

func (r *RawGray16Image[CS]) CopyColorHspan(x, y, length int, colors []color.Gray16[CS]) {
	off := y*r.Stride + x*4
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*4)
	copy(r.Pixels[off:off+len(colors)*4], src)
}

func (r *RawGray16Image[CS]) CopyColorVspan(x, y, length int, colors []color.Gray16[CS]) {
	off := y*r.Stride + x*4
	for _, c := range colors {
		*(*color.Gray16[CS])(unsafe.Pointer(&r.Pixels[off])) = c
		off += r.Stride
	}
}

func (sb *StackBlurGray16[CS]) BlurGray16(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*4 || len(pixels) < (h-1)*stride+w*4 {
		return
	}
	sb.Blur(&RawGray16Image[CS]{Pixels: pixels, W: w, H: h, Stride: stride}, radius)
}
