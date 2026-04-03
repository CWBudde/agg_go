package effects

import (
	"math"

	"github.com/cwbudde/agg_go/internal/array"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
)

// StackBlurCalcRGBA is the RGBA accumulator used by the generic stack-blur
// implementation, mirroring AGG's stack_blur_calc_rgba helper.  The CS type
// parameter carries the colour-space tag so that the calculator produces
// correctly typed output without unsafe conversions.
type StackBlurCalcRGBA[T ~uint32 | ~uint64, CS color.Space] struct {
	r, g, b, a T
}

// Clear resets all color channels to zero.
func (calc *StackBlurCalcRGBA[T, CS]) Clear() {
	calc.r, calc.g, calc.b, calc.a = 0, 0, 0, 0
}

// Add adds a color value to the accumulator.
func (calc *StackBlurCalcRGBA[T, CS]) Add(v color.RGBA8[CS]) {
	calc.r += T(v.R)
	calc.g += T(v.G)
	calc.b += T(v.B)
	calc.a += T(v.A)
}

// AddWeighted adds a weighted color value to the accumulator.
func (calc *StackBlurCalcRGBA[T, CS]) AddWeighted(v color.RGBA8[CS], weight int) {
	w := T(weight)
	calc.r += T(v.R) * w
	calc.g += T(v.G) * w
	calc.b += T(v.B) * w
	calc.a += T(v.A) * w
}

// Sub subtracts a color value from the accumulator.
func (calc *StackBlurCalcRGBA[T, CS]) Sub(v color.RGBA8[CS]) {
	calc.r -= T(v.R)
	calc.g -= T(v.G)
	calc.b -= T(v.B)
	calc.a -= T(v.A)
}

// SubCalc subtracts another calculator from this one.
func (calc *StackBlurCalcRGBA[T, CS]) SubCalc(other StackBlurCalcRGBA[T, CS]) {
	calc.r -= other.r
	calc.g -= other.g
	calc.b -= other.b
	calc.a -= other.a
}

// AddCalc adds another calculator to this one.
func (calc *StackBlurCalcRGBA[T, CS]) AddCalc(other StackBlurCalcRGBA[T, CS]) {
	calc.r += other.r
	calc.g += other.g
	calc.b += other.b
	calc.a += other.a
}

// CalcPix calculates the final pixel value using division.
func (calc *StackBlurCalcRGBA[T, CS]) CalcPix(result *color.RGBA8[CS], div int) {
	d := T(div)
	result.R = basics.Int8u(calc.r / d)
	result.G = basics.Int8u(calc.g / d)
	result.B = basics.Int8u(calc.b / d)
	result.A = basics.Int8u(calc.a / d)
}

// CalcPixMulShr calculates the final pixel value using multiplication and bit shift.
func (calc *StackBlurCalcRGBA[T, CS]) CalcPixMulShr(result *color.RGBA8[CS], mul, shr int) {
	m := T(mul)
	result.R = basics.Int8u((calc.r * m) >> shr)
	result.G = basics.Int8u((calc.g * m) >> shr)
	result.B = basics.Int8u((calc.b * m) >> shr)
	result.A = basics.Int8u((calc.a * m) >> shr)
}

// StackBlurCalcRGBA16 mirrors AGG's stack_blur_calc_rgba for 16-bit RGBA channels.
type StackBlurCalcRGBA16[T ~uint32 | ~uint64, CS color.Space] struct {
	r, g, b, a T
}

// Clear resets all color channels to zero.
func (calc *StackBlurCalcRGBA16[T, CS]) Clear() {
	calc.r, calc.g, calc.b, calc.a = 0, 0, 0, 0
}

// Add adds a color value to the accumulator.
func (calc *StackBlurCalcRGBA16[T, CS]) Add(v color.RGBA16[CS]) {
	calc.r += T(v.R)
	calc.g += T(v.G)
	calc.b += T(v.B)
	calc.a += T(v.A)
}

// AddWeighted adds a weighted color value to the accumulator.
func (calc *StackBlurCalcRGBA16[T, CS]) AddWeighted(v color.RGBA16[CS], weight int) {
	w := T(weight)
	calc.r += T(v.R) * w
	calc.g += T(v.G) * w
	calc.b += T(v.B) * w
	calc.a += T(v.A) * w
}

// Sub subtracts a color value from the accumulator.
func (calc *StackBlurCalcRGBA16[T, CS]) Sub(v color.RGBA16[CS]) {
	calc.r -= T(v.R)
	calc.g -= T(v.G)
	calc.b -= T(v.B)
	calc.a -= T(v.A)
}

// SubCalc subtracts another calculator from this one.
func (calc *StackBlurCalcRGBA16[T, CS]) SubCalc(other StackBlurCalcRGBA16[T, CS]) {
	calc.r -= other.r
	calc.g -= other.g
	calc.b -= other.b
	calc.a -= other.a
}

// AddCalc adds another calculator to this one.
func (calc *StackBlurCalcRGBA16[T, CS]) AddCalc(other StackBlurCalcRGBA16[T, CS]) {
	calc.r += other.r
	calc.g += other.g
	calc.b += other.b
	calc.a += other.a
}

// CalcPix calculates the final pixel value using division.
func (calc *StackBlurCalcRGBA16[T, CS]) CalcPix(result *color.RGBA16[CS], div int) {
	d := T(div)
	result.R = basics.Int16u(calc.r / d)
	result.G = basics.Int16u(calc.g / d)
	result.B = basics.Int16u(calc.b / d)
	result.A = basics.Int16u(calc.a / d)
}

// CalcPixMulShr calculates the final pixel value using multiplication and bit shift.
func (calc *StackBlurCalcRGBA16[T, CS]) CalcPixMulShr(result *color.RGBA16[CS], mul, shr int) {
	m := T(mul)
	result.R = basics.Int16u((calc.r * m) >> shr)
	result.G = basics.Int16u((calc.g * m) >> shr)
	result.B = basics.Int16u((calc.b * m) >> shr)
	result.A = basics.Int16u((calc.a * m) >> shr)
}

// StackBlurCalcRGB mirrors AGG's stack_blur_calc_rgb for 8-bit RGB channels.
type StackBlurCalcRGB[T ~uint32 | ~uint64, CS color.Space] struct {
	r, g, b T
}

// Clear resets all color channels to zero.
func (calc *StackBlurCalcRGB[T, CS]) Clear() {
	calc.r, calc.g, calc.b = 0, 0, 0
}

// Add adds a color value to the accumulator.
func (calc *StackBlurCalcRGB[T, CS]) Add(v color.RGB8[CS]) {
	calc.r += T(v.R)
	calc.g += T(v.G)
	calc.b += T(v.B)
}

// AddWeighted adds a weighted color value to the accumulator.
func (calc *StackBlurCalcRGB[T, CS]) AddWeighted(v color.RGB8[CS], weight int) {
	w := T(weight)
	calc.r += T(v.R) * w
	calc.g += T(v.G) * w
	calc.b += T(v.B) * w
}

// Sub subtracts a color value from the accumulator.
func (calc *StackBlurCalcRGB[T, CS]) Sub(v color.RGB8[CS]) {
	calc.r -= T(v.R)
	calc.g -= T(v.G)
	calc.b -= T(v.B)
}

// SubCalc subtracts another calculator from this one.
func (calc *StackBlurCalcRGB[T, CS]) SubCalc(other StackBlurCalcRGB[T, CS]) {
	calc.r -= other.r
	calc.g -= other.g
	calc.b -= other.b
}

// AddCalc adds another calculator to this one.
func (calc *StackBlurCalcRGB[T, CS]) AddCalc(other StackBlurCalcRGB[T, CS]) {
	calc.r += other.r
	calc.g += other.g
	calc.b += other.b
}

// CalcPix calculates the final pixel value using division.
func (calc *StackBlurCalcRGB[T, CS]) CalcPix(result *color.RGB8[CS], div int) {
	d := T(div)
	result.R = basics.Int8u(calc.r / d)
	result.G = basics.Int8u(calc.g / d)
	result.B = basics.Int8u(calc.b / d)
}

// CalcPixMulShr calculates the final pixel value using multiplication and bit shift.
func (calc *StackBlurCalcRGB[T, CS]) CalcPixMulShr(result *color.RGB8[CS], mul, shr int) {
	m := T(mul)
	result.R = basics.Int8u((calc.r * m) >> shr)
	result.G = basics.Int8u((calc.g * m) >> shr)
	result.B = basics.Int8u((calc.b * m) >> shr)
}

// StackBlurCalcRGB16 mirrors AGG's stack_blur_calc_rgb for 16-bit RGB channels.
type StackBlurCalcRGB16[T ~uint32 | ~uint64, CS color.Space] struct {
	r, g, b T
}

// Clear resets all color channels to zero.
func (calc *StackBlurCalcRGB16[T, CS]) Clear() {
	calc.r, calc.g, calc.b = 0, 0, 0
}

// Add adds a color value to the accumulator.
func (calc *StackBlurCalcRGB16[T, CS]) Add(v color.RGB16[CS]) {
	calc.r += T(v.R)
	calc.g += T(v.G)
	calc.b += T(v.B)
}

// AddWeighted adds a weighted color value to the accumulator.
func (calc *StackBlurCalcRGB16[T, CS]) AddWeighted(v color.RGB16[CS], weight int) {
	w := T(weight)
	calc.r += T(v.R) * w
	calc.g += T(v.G) * w
	calc.b += T(v.B) * w
}

// Sub subtracts a color value from the accumulator.
func (calc *StackBlurCalcRGB16[T, CS]) Sub(v color.RGB16[CS]) {
	calc.r -= T(v.R)
	calc.g -= T(v.G)
	calc.b -= T(v.B)
}

// SubCalc subtracts another calculator from this one.
func (calc *StackBlurCalcRGB16[T, CS]) SubCalc(other StackBlurCalcRGB16[T, CS]) {
	calc.r -= other.r
	calc.g -= other.g
	calc.b -= other.b
}

// AddCalc adds another calculator to this one.
func (calc *StackBlurCalcRGB16[T, CS]) AddCalc(other StackBlurCalcRGB16[T, CS]) {
	calc.r += other.r
	calc.g += other.g
	calc.b += other.b
}

// CalcPix calculates the final pixel value using division.
func (calc *StackBlurCalcRGB16[T, CS]) CalcPix(result *color.RGB16[CS], div int) {
	d := T(div)
	result.R = basics.Int16u(calc.r / d)
	result.G = basics.Int16u(calc.g / d)
	result.B = basics.Int16u(calc.b / d)
}

// CalcPixMulShr calculates the final pixel value using multiplication and bit shift.
func (calc *StackBlurCalcRGB16[T, CS]) CalcPixMulShr(result *color.RGB16[CS], mul, shr int) {
	m := T(mul)
	result.R = basics.Int16u((calc.r * m) >> shr)
	result.G = basics.Int16u((calc.g * m) >> shr)
	result.B = basics.Int16u((calc.b * m) >> shr)
}

// StackBlurCalcGray mirrors AGG's stack_blur_calc_gray for 8-bit grayscale channels.
type StackBlurCalcGray[T ~uint32 | ~uint64, CS color.Space] struct {
	v, a T
}

// Clear resets all channels to zero.
func (calc *StackBlurCalcGray[T, CS]) Clear() {
	calc.v, calc.a = 0, 0
}

// Add adds a color value to the accumulator.
func (calc *StackBlurCalcGray[T, CS]) Add(c color.Gray8[CS]) {
	calc.v += T(c.V)
	calc.a += T(c.A)
}

// AddWeighted adds a weighted color value to the accumulator.
func (calc *StackBlurCalcGray[T, CS]) AddWeighted(c color.Gray8[CS], weight int) {
	w := T(weight)
	calc.v += T(c.V) * w
	calc.a += T(c.A) * w
}

// Sub subtracts a color value from the accumulator.
func (calc *StackBlurCalcGray[T, CS]) Sub(c color.Gray8[CS]) {
	calc.v -= T(c.V)
	calc.a -= T(c.A)
}

// SubCalc subtracts another calculator from this one.
func (calc *StackBlurCalcGray[T, CS]) SubCalc(other StackBlurCalcGray[T, CS]) {
	calc.v -= other.v
	calc.a -= other.a
}

// AddCalc adds another calculator to this one.
func (calc *StackBlurCalcGray[T, CS]) AddCalc(other StackBlurCalcGray[T, CS]) {
	calc.v += other.v
	calc.a += other.a
}

// CalcPix calculates the final pixel value using division.
func (calc *StackBlurCalcGray[T, CS]) CalcPix(result *color.Gray8[CS], div int) {
	d := T(div)
	result.V = basics.Int8u(calc.v / d)
	result.A = basics.Int8u(calc.a / d)
}

// CalcPixMulShr calculates the final pixel value using multiplication and bit shift.
func (calc *StackBlurCalcGray[T, CS]) CalcPixMulShr(result *color.Gray8[CS], mul, shr int) {
	m := T(mul)
	result.V = basics.Int8u((calc.v * m) >> shr)
	result.A = basics.Int8u((calc.a * m) >> shr)
}

// StackBlurCalcGray16 mirrors AGG's stack_blur_calc_gray for 16-bit grayscale channels.
type StackBlurCalcGray16[T ~uint32 | ~uint64, CS color.Space] struct {
	v, a T
}

// Clear resets all channels to zero.
func (calc *StackBlurCalcGray16[T, CS]) Clear() {
	calc.v, calc.a = 0, 0
}

// Add adds a color value to the accumulator.
func (calc *StackBlurCalcGray16[T, CS]) Add(c color.Gray16[CS]) {
	calc.v += T(c.V)
	calc.a += T(c.A)
}

// AddWeighted adds a weighted color value to the accumulator.
func (calc *StackBlurCalcGray16[T, CS]) AddWeighted(c color.Gray16[CS], weight int) {
	w := T(weight)
	calc.v += T(c.V) * w
	calc.a += T(c.A) * w
}

// Sub subtracts a color value from the accumulator.
func (calc *StackBlurCalcGray16[T, CS]) Sub(c color.Gray16[CS]) {
	calc.v -= T(c.V)
	calc.a -= T(c.A)
}

// SubCalc subtracts another calculator from this one.
func (calc *StackBlurCalcGray16[T, CS]) SubCalc(other StackBlurCalcGray16[T, CS]) {
	calc.v -= other.v
	calc.a -= other.a
}

// AddCalc adds another calculator to this one.
func (calc *StackBlurCalcGray16[T, CS]) AddCalc(other StackBlurCalcGray16[T, CS]) {
	calc.v += other.v
	calc.a += other.a
}

// CalcPix calculates the final pixel value using division.
func (calc *StackBlurCalcGray16[T, CS]) CalcPix(result *color.Gray16[CS], div int) {
	d := T(div)
	result.V = basics.Int16u(calc.v / d)
	result.A = basics.Int16u(calc.a / d)
}

// CalcPixMulShr calculates the final pixel value using multiplication and bit shift.
func (calc *StackBlurCalcGray16[T, CS]) CalcPixMulShr(result *color.Gray16[CS], mul, shr int) {
	m := T(mul)
	result.V = basics.Int16u((calc.v * m) >> shr)
	result.A = basics.Int16u((calc.a * m) >> shr)
}

// SimpleStackBlur is a direct, reusable two-pass stack blur for RGBA images.
// It keeps AGG-style temporary buffers between calls to avoid repeated
// allocations when the same instance is reused.
type SimpleStackBlur struct {
	buf   *array.PodVector[color.RGBA8[color.Linear]]
	stack *array.PodVector[color.RGBA8[color.Linear]]
}

// NewSimpleStackBlur creates a new SimpleStackBlur instance.
func NewSimpleStackBlur() *SimpleStackBlur {
	return &SimpleStackBlur{
		buf:   array.NewPodVector[color.RGBA8[color.Linear]](),
		stack: array.NewPodVector[color.RGBA8[color.Linear]](),
	}
}

// BlurHorizontal applies horizontal blur to an RGBA image.
func (sb *SimpleStackBlur) BlurHorizontal(pixels [][]color.RGBA8[color.Linear], radius int) {
	if radius < 1 || len(pixels) == 0 || len(pixels[0]) == 0 {
		return
	}

	h := len(pixels)
	w := len(pixels[0])
	wm := w - 1
	div := radius*2 + 1

	var mulSum, shrSum int
	if radius < 255 {
		mulSum = int(StackBlur8Mul[radius])
		shrSum = int(StackBlur8Shr[radius])
	}

	// Allocate buffers
	sb.buf.Allocate(w, 128)
	sb.stack.Allocate(div, 32)

	for y := 0; y < h; y++ {
		sum := StackBlurCalcRGBA[uint32, color.Linear]{}
		sumIn := StackBlurCalcRGBA[uint32, color.Linear]{}
		sumOut := StackBlurCalcRGBA[uint32, color.Linear]{}

		sum.Clear()
		sumIn.Clear()
		sumOut.Clear()

		// Initialize the stack with edge pixels
		pix := pixels[y][0]
		for i := 0; i <= radius; i++ {
			sb.stack.Set(i, pix)
			sum.AddWeighted(pix, i+1)
			sumOut.Add(pix)
		}

		// Fill the right side of the stack
		for i := 1; i <= radius; i++ {
			x := i
			if x > wm {
				x = wm
			}
			pix = pixels[y][x]
			sb.stack.Set(i+radius, pix)
			sum.AddWeighted(pix, radius+1-i)
			sumIn.Add(pix)
		}

		stackPtr := radius
		for x := 0; x < w; x++ {
			// Calculate the blurred pixel
			var result color.RGBA8[color.Linear]
			if mulSum != 0 {
				sum.CalcPixMulShr(&result, mulSum, shrSum)
			} else {
				divSum := (radius + 1) * (radius + 1)
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

			xp := x + radius + 1
			if xp > wm {
				xp = wm
			}
			pix = pixels[y][xp]

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

		// Copy the blurred row back
		for i := 0; i < w; i++ {
			pixels[y][i] = sb.buf.ValueAt(i)
		}
	}
}

// BlurVertical applies vertical blur to an RGBA image.
func (sb *SimpleStackBlur) BlurVertical(pixels [][]color.RGBA8[color.Linear], radius int) {
	if radius < 1 || len(pixels) == 0 || len(pixels[0]) == 0 {
		return
	}

	h := len(pixels)
	w := len(pixels[0])
	hm := h - 1
	div := radius*2 + 1

	var mulSum, shrSum int
	if radius < 255 {
		mulSum = int(StackBlur8Mul[radius])
		shrSum = int(StackBlur8Shr[radius])
	}

	// Allocate buffers
	sb.buf.Allocate(h, 128)
	sb.stack.Allocate(div, 32)

	for x := 0; x < w; x++ {
		sum := StackBlurCalcRGBA[uint32, color.Linear]{}
		sumIn := StackBlurCalcRGBA[uint32, color.Linear]{}
		sumOut := StackBlurCalcRGBA[uint32, color.Linear]{}

		sum.Clear()
		sumIn.Clear()
		sumOut.Clear()

		// Initialize the stack with edge pixels
		pix := pixels[0][x]
		for i := 0; i <= radius; i++ {
			sb.stack.Set(i, pix)
			sum.AddWeighted(pix, i+1)
			sumOut.Add(pix)
		}

		// Fill the bottom side of the stack
		for i := 1; i <= radius; i++ {
			y := i
			if y > hm {
				y = hm
			}
			pix = pixels[y][x]
			sb.stack.Set(i+radius, pix)
			sum.AddWeighted(pix, radius+1-i)
			sumIn.Add(pix)
		}

		stackPtr := radius
		for y := 0; y < h; y++ {
			// Calculate the blurred pixel
			var result color.RGBA8[color.Linear]
			if mulSum != 0 {
				sum.CalcPixMulShr(&result, mulSum, shrSum)
			} else {
				divSum := (radius + 1) * (radius + 1)
				sum.CalcPix(&result, divSum)
			}
			sb.buf.Set(y, result)

			sum.SubCalc(sumOut)

			stackStart := stackPtr + div - radius
			if stackStart >= div {
				stackStart -= div
			}
			stackPix := sb.stack.ValueAt(stackStart)

			sumOut.Sub(stackPix)

			yp := y + radius + 1
			if yp > hm {
				yp = hm
			}
			pix = pixels[yp][x]

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

		// Copy the blurred column back
		for i := 0; i < h; i++ {
			pixels[i][x] = sb.buf.ValueAt(i)
		}
	}
}

// Blur applies both horizontal and vertical blur.
func (sb *SimpleStackBlur) Blur(pixels [][]color.RGBA8[color.Linear], radius int) {
	sb.BlurHorizontal(pixels, radius)
	sb.BlurVertical(pixels, radius)
}

// SimpleRecursiveBlur provides a recursive blur implementation for high-quality results.
type SimpleRecursiveBlur struct {
	sum1 *array.PodVector[RecursiveBlurCalcRGBA[float64]]
	sum2 *array.PodVector[RecursiveBlurCalcRGBA[float64]]
	buf  *array.PodVector[color.RGBA8[color.Linear]]
}

// NewSimpleRecursiveBlur creates a new SimpleRecursiveBlur instance.
func NewSimpleRecursiveBlur() *SimpleRecursiveBlur {
	return &SimpleRecursiveBlur{
		sum1: array.NewPodVector[RecursiveBlurCalcRGBA[float64]](),
		sum2: array.NewPodVector[RecursiveBlurCalcRGBA[float64]](),
		buf:  array.NewPodVector[color.RGBA8[color.Linear]](),
	}
}

// BlurHorizontal applies horizontal recursive blur.
func (rb *SimpleRecursiveBlur) BlurHorizontal(pixels [][]color.RGBA8[color.Linear], radius float64) {
	if radius < 0.62 || len(pixels) == 0 || len(pixels[0]) == 0 {
		return
	}

	h := len(pixels)
	w := len(pixels[0])
	if w < 3 {
		return
	}

	// Calculate filter coefficients
	s := radius * 0.5
	var q float64
	if s < 2.5 {
		q = 3.97156 - 4.14554*math.Sqrt(1-0.26891*s)
	} else {
		q = 0.98711*s - 0.96330
	}

	q2 := q * q
	q3 := q2 * q

	b0 := 1.0 / (1.578250 + 2.444130*q + 1.428100*q2 + 0.422205*q3)
	b1 := (2.44413*q + 2.85619*q2 + 1.26661*q3)
	b2 := (-1.42810*q2 + -1.26661*q3)
	b3 := (0.422205 * q3)
	b := 1 - (b1+b2+b3)*b0

	b1 *= b0
	b2 *= b0
	b3 *= b0

	wm := w - 1

	// Allocate buffers
	rb.sum1.Allocate(w, 0)
	rb.sum2.Allocate(w, 0)
	rb.buf.Allocate(w, 0)

	for y := 0; y < h; y++ {
		// Forward pass
		c := RecursiveBlurCalcRGBA[float64]{}
		c.FromPix(pixels[y][0])

		calc0 := rb.sum1.ValueAt(0)
		calc0.Calc(b, b1, b2, b3, c, c, c, c)
		rb.sum1.Set(0, calc0)

		c.FromPix(pixels[y][1])
		calc1 := rb.sum1.ValueAt(1)
		calc1.Calc(b, b1, b2, b3, c, rb.sum1.ValueAt(0), rb.sum1.ValueAt(0), rb.sum1.ValueAt(0))
		rb.sum1.Set(1, calc1)

		c.FromPix(pixels[y][2])
		calc2 := rb.sum1.ValueAt(2)
		calc2.Calc(b, b1, b2, b3, c, rb.sum1.ValueAt(1), rb.sum1.ValueAt(0), rb.sum1.ValueAt(0))
		rb.sum1.Set(2, calc2)

		for x := 3; x < w; x++ {
			c.FromPix(pixels[y][x])
			calcX := rb.sum1.ValueAt(x)
			calcX.Calc(b, b1, b2, b3, c, rb.sum1.ValueAt(x-1), rb.sum1.ValueAt(x-2), rb.sum1.ValueAt(x-3))
			rb.sum1.Set(x, calcX)
		}

		// Backward pass
		calcWm := rb.sum2.ValueAt(wm)
		calcWm.Calc(b, b1, b2, b3, rb.sum1.ValueAt(wm), rb.sum1.ValueAt(wm), rb.sum1.ValueAt(wm), rb.sum1.ValueAt(wm))
		rb.sum2.Set(wm, calcWm)

		var result color.RGBA8[color.Linear]
		calc := rb.sum2.ValueAt(wm)
		calc.ToPix(&result)
		rb.buf.Set(wm, result)

		for x := wm - 1; x >= 0; x-- {
			x1 := x + 1
			x2 := x + 2
			x3 := x + 3
			if x2 >= w {
				x2 = wm
			}
			if x3 >= w {
				x3 = wm
			}

			calcX := rb.sum2.ValueAt(x)
			calcX.Calc(b, b1, b2, b3, rb.sum1.ValueAt(x), rb.sum2.ValueAt(x1), rb.sum2.ValueAt(x2), rb.sum2.ValueAt(x3))
			rb.sum2.Set(x, calcX)

			calc := rb.sum2.ValueAt(x)
			calc.ToPix(&result)
			rb.buf.Set(x, result)
		}

		// Copy the blurred row back
		for i := 0; i < w; i++ {
			pixels[y][i] = rb.buf.ValueAt(i)
		}
	}
}

// RecursiveBlurCalcRGBA implements the calculator for recursive blur with RGBA colors.
type RecursiveBlurCalcRGBA[T ~float64] struct {
	r, g, b, a T
}

// FromPix loads values from a pixel.
func (calc *RecursiveBlurCalcRGBA[T]) FromPix(c color.RGBA8[color.Linear]) {
	calc.r = T(c.R)
	calc.g = T(c.G)
	calc.b = T(c.B)
	calc.a = T(c.A)
}

// Calc performs the recursive filter calculation.
func (calc *RecursiveBlurCalcRGBA[T]) Calc(b1, b2, b3, b4 float64, c1, c2, c3, c4 RecursiveBlurCalcRGBA[T]) {
	calc.r = T(b1)*c1.r + T(b2)*c2.r + T(b3)*c3.r + T(b4)*c4.r
	calc.g = T(b1)*c1.g + T(b2)*c2.g + T(b3)*c3.g + T(b4)*c4.g
	calc.b = T(b1)*c1.b + T(b2)*c2.b + T(b3)*c3.b + T(b4)*c4.b
	calc.a = T(b1)*c1.a + T(b2)*c2.a + T(b3)*c3.a + T(b4)*c4.a
}

// ToPix stores values to a pixel.
func (calc *RecursiveBlurCalcRGBA[T]) ToPix(c *color.RGBA8[color.Linear]) {
	c.R = basics.Int8u(calc.r)
	c.G = basics.Int8u(calc.g)
	c.B = basics.Int8u(calc.b)
	c.A = basics.Int8u(calc.a)
}
