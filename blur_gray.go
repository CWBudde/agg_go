package agg

import (
	"unsafe"

	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/effects"
)

// ---------------------------------------------------------------------------
// Stack blur for Gray8 images — mirrors the RGBA8 StackBlur pattern but uses
// StackBlurCalcGray[uint32, CS] as the accumulator calculator.
// ---------------------------------------------------------------------------

// StackBlurGray implements AGG's stack_blur for Gray8 images.  Like StackBlur,
// temporary buffers are reused across calls.
type StackBlurGray[CS color.Space] struct {
	buf   *array.PodVector[color.Gray8[CS]]
	stack *array.PodVector[color.Gray8[CS]]
}

// NewStackBlurGray creates a new StackBlurGray instance.
func NewStackBlurGray[CS color.Space]() *StackBlurGray[CS] {
	return &StackBlurGray[CS]{
		buf:   array.NewPodVector[color.Gray8[CS]](),
		stack: array.NewPodVector[color.Gray8[CS]](),
	}
}

// BlurX applies a horizontal stack blur pass for Gray8 images.
func (sb *StackBlurGray[CS]) BlurX(img PixelReadWriter[color.Gray8[CS]], radius int) {
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

// BlurY applies a vertical stack blur pass by delegating through a transposer.
func (sb *StackBlurGray[CS]) BlurY(img PixelReadWriter[color.Gray8[CS]], radius int) {
	sb.BlurX(&pixelImageTransposer[color.Gray8[CS]]{img: img}, radius)
}

// Blur applies both horizontal and vertical blur passes.
func (sb *StackBlurGray[CS]) Blur(img PixelReadWriter[color.Gray8[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// ---------------------------------------------------------------------------
// Raw Gray8 byte-buffer adapter
// ---------------------------------------------------------------------------

// rawGray8Image adapts a flat Gray8 byte slice (2 bytes per pixel: V, A) to
// PixelReadWriter[color.Gray8[CS]].
type rawGray8Image[CS color.Space] struct {
	pixels []byte
	w, h   int
	stride int
}

func (r *rawGray8Image[CS]) Width() int  { return r.w }
func (r *rawGray8Image[CS]) Height() int { return r.h }

func (r *rawGray8Image[CS]) Pixel(x, y int) color.Gray8[CS] {
	off := y*r.stride + x*2
	return color.Gray8[CS]{V: r.pixels[off], A: r.pixels[off+1]}
}

func (r *rawGray8Image[CS]) CopyColorHspan(x, y, length int, colors []color.Gray8[CS]) {
	off := y*r.stride + x*2
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*2)
	copy(r.pixels[off:off+len(colors)*2], src)
}

func (r *rawGray8Image[CS]) CopyColorVspan(x, y, length int, colors []color.Gray8[CS]) {
	off := y*r.stride + x*2
	for _, c := range colors {
		r.pixels[off] = c.V
		r.pixels[off+1] = c.A
		off += r.stride
	}
}

// BlurGray8 applies a stack blur to a raw Gray8 byte buffer.  Each pixel is
// 2 bytes (V, A).  The stride parameter is the number of bytes per row.
func (sb *StackBlurGray[CS]) BlurGray8(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*2 || len(pixels) < (h-1)*stride+w*2 {
		return
	}
	sb.Blur(&rawGray8Image[CS]{pixels: pixels, w: w, h: h, stride: stride}, radius)
}

// ---------------------------------------------------------------------------
// Stack blur for Gray16 images — uses StackBlurCalcGray16[uint64, CS].
// ---------------------------------------------------------------------------

// StackBlurGray16 implements AGG's stack_blur for Gray16 images.  Like
// StackBlur, temporary buffers are reused across calls.
type StackBlurGray16[CS color.Space] struct {
	buf   *array.PodVector[color.Gray16[CS]]
	stack *array.PodVector[color.Gray16[CS]]
}

// NewStackBlurGray16 creates a new StackBlurGray16 instance.
func NewStackBlurGray16[CS color.Space]() *StackBlurGray16[CS] {
	return &StackBlurGray16[CS]{
		buf:   array.NewPodVector[color.Gray16[CS]](),
		stack: array.NewPodVector[color.Gray16[CS]](),
	}
}

// BlurX applies a horizontal stack blur pass for Gray16 images.
func (sb *StackBlurGray16[CS]) BlurX(img PixelReadWriter[color.Gray16[CS]], radius int) {
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

// BlurY applies a vertical stack blur pass by delegating through a transposer.
func (sb *StackBlurGray16[CS]) BlurY(img PixelReadWriter[color.Gray16[CS]], radius int) {
	sb.BlurX(&pixelImageTransposer[color.Gray16[CS]]{img: img}, radius)
}

// Blur applies both horizontal and vertical blur passes.
func (sb *StackBlurGray16[CS]) Blur(img PixelReadWriter[color.Gray16[CS]], radius int) {
	sb.BlurX(img, radius)
	sb.BlurY(img, radius)
}

// ---------------------------------------------------------------------------
// Raw Gray16 byte-buffer adapter
// ---------------------------------------------------------------------------

// rawGray16Image adapts a flat Gray16 byte slice (4 bytes per pixel: V uint16,
// A uint16, native byte order) to PixelReadWriter[color.Gray16[CS]].
type rawGray16Image[CS color.Space] struct {
	pixels []byte
	w, h   int
	stride int
}

func (r *rawGray16Image[CS]) Width() int  { return r.w }
func (r *rawGray16Image[CS]) Height() int { return r.h }

func (r *rawGray16Image[CS]) Pixel(x, y int) color.Gray16[CS] {
	off := y*r.stride + x*4
	return *(*color.Gray16[CS])(unsafe.Pointer(&r.pixels[off]))
}

func (r *rawGray16Image[CS]) CopyColorHspan(x, y, length int, colors []color.Gray16[CS]) {
	off := y*r.stride + x*4
	src := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), len(colors)*4)
	copy(r.pixels[off:off+len(colors)*4], src)
}

func (r *rawGray16Image[CS]) CopyColorVspan(x, y, length int, colors []color.Gray16[CS]) {
	off := y*r.stride + x*4
	for _, c := range colors {
		*(*color.Gray16[CS])(unsafe.Pointer(&r.pixels[off])) = c
		off += r.stride
	}
}

// BlurGray16 applies a stack blur to a raw Gray16 byte buffer.  Each pixel is
// 4 bytes (V uint16, A uint16, native byte order).  The stride parameter is
// the number of bytes per row.
func (sb *StackBlurGray16[CS]) BlurGray16(pixels []byte, w, h, stride, radius int) {
	if radius < 1 || w <= 0 || h <= 0 || stride < w*4 || len(pixels) < (h-1)*stride+w*4 {
		return
	}
	sb.Blur(&rawGray16Image[CS]{pixels: pixels, w: w, h: h, stride: stride}, radius)
}
