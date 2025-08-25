package effects

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// SlightBlur provides a special-purpose filter for applying a Gaussian blur
// with a radius small enough that the blur only affects adjacent pixels.
// It uses a Gaussian curve with standard deviation of r/2 as per HTML/CSS spec.
// This filter is useful for smoothing artifacts caused by detail rendered
// at the pixel scale, e.g. single-pixel lines.
type SlightBlur struct {
	g0, g1 float64
	buf    []basics.Int8u
}

// PixFmtInterface defines the interface for pixel formats that can be blurred.
type PixFmtInterface interface {
	Width() int
	Height() int
	GetPixel(x, y int) color.RGBA8[color.Linear]
	CopyPixel(x, y int, c color.RGBA8[color.Linear])
}

// NewSlightBlur creates a new SlightBlur with the specified radius.
// Default radius is 1.33, which provides good quality for single-pixel smoothing.
func NewSlightBlur(radius float64) *SlightBlur {
	if radius <= 0 {
		radius = 1.33
	}

	sb := &SlightBlur{}
	sb.SetRadius(radius)
	return sb
}

// SetRadius sets the blur radius and recalculates the Gaussian coefficients.
func (sb *SlightBlur) SetRadius(r float64) {
	if r > 0 {
		// Sample the gaussian curve at 0 and r/2 standard deviations.
		// At 3 standard deviations, the response is < 0.005.
		pi := 3.14159
		n := 2 / r
		sb.g0 = 1 / math.Sqrt(2*pi)
		sb.g1 = sb.g0 * math.Exp(-n*n)

		// Normalize
		sum := sb.g0 + 2*sb.g1
		sb.g0 /= sum
		sb.g1 /= sum
	} else {
		sb.g0 = 1
		sb.g1 = 0
	}
}

// Blur applies the slight blur to the specified rectangle of the image.
func (sb *SlightBlur) Blur(img PixFmtInterface, bounds basics.RectI) {
	// Make sure we stay within the image area
	imageBounds := basics.RectI{X1: 0, Y1: 0, X2: img.Width() - 1, Y2: img.Height() - 1}
	bounds.Clip(imageBounds)

	w := bounds.X2 - bounds.X1 + 1
	h := bounds.Y2 - bounds.Y1 + 1

	if w < 3 || h < 3 {
		return
	}

	// Allocate temporary buffer for the blur operation (4 bytes per pixel for RGBA)
	bufSize := w * h * 4
	if len(sb.buf) < bufSize {
		sb.buf = make([]basics.Int8u, bufSize)
	}

	// Apply horizontal blur first
	sb.blurHorizontal(img, bounds)
	// Then apply vertical blur
	sb.blurVertical(img, bounds)
}

// blurHorizontal applies horizontal Gaussian blur
func (sb *SlightBlur) blurHorizontal(img PixFmtInterface, bounds basics.RectI) {
	w := bounds.X2 - bounds.X1 + 1
	h := bounds.Y2 - bounds.Y1 + 1

	// Process each row
	for y := 0; y < h; y++ {
		imgY := bounds.Y1 + y
		bufOffset := y * w * 4

		for x := 0; x < w; x++ {
			imgX := bounds.X1 + x
			pixOffset := bufOffset + x*4

			// Get pixels with clamping
			leftX := imgX - 1
			rightX := imgX + 1
			if leftX < 0 {
				leftX = 0
			}
			if rightX >= img.Width() {
				rightX = img.Width() - 1
			}

			left := img.GetPixel(leftX, imgY)
			center := img.GetPixel(imgX, imgY)
			right := img.GetPixel(rightX, imgY)

			// Apply Gaussian weights
			blurredR := float64(left.R)*sb.g1 + float64(center.R)*sb.g0 + float64(right.R)*sb.g1
			blurredG := float64(left.G)*sb.g1 + float64(center.G)*sb.g0 + float64(right.G)*sb.g1
			blurredB := float64(left.B)*sb.g1 + float64(center.B)*sb.g0 + float64(right.B)*sb.g1
			blurredA := float64(left.A)*sb.g1 + float64(center.A)*sb.g0 + float64(right.A)*sb.g1

			// Clamp and store in buffer
			sb.buf[pixOffset] = basics.Int8u(math.Min(255, math.Max(0, blurredR)))
			sb.buf[pixOffset+1] = basics.Int8u(math.Min(255, math.Max(0, blurredG)))
			sb.buf[pixOffset+2] = basics.Int8u(math.Min(255, math.Max(0, blurredB)))
			sb.buf[pixOffset+3] = basics.Int8u(math.Min(255, math.Max(0, blurredA)))
		}
	}
}

// blurVertical applies vertical Gaussian blur
func (sb *SlightBlur) blurVertical(img PixFmtInterface, bounds basics.RectI) {
	w := bounds.X2 - bounds.X1 + 1
	h := bounds.Y2 - bounds.Y1 + 1

	// Process each column
	for x := 0; x < w; x++ {
		imgX := bounds.X1 + x

		for y := 0; y < h; y++ {
			imgY := bounds.Y1 + y

			// Get pixel indices in buffer with clamping
			topY := y - 1
			bottomY := y + 1
			if topY < 0 {
				topY = 0
			}
			if bottomY >= h {
				bottomY = h - 1
			}

			// Get pixel data from horizontal buffer
			topOffset := topY*w*4 + x*4
			centerOffset := y*w*4 + x*4
			bottomOffset := bottomY*w*4 + x*4

			// Apply Gaussian weights
			blurredR := float64(sb.buf[topOffset])*sb.g1 + float64(sb.buf[centerOffset])*sb.g0 + float64(sb.buf[bottomOffset])*sb.g1
			blurredG := float64(sb.buf[topOffset+1])*sb.g1 + float64(sb.buf[centerOffset+1])*sb.g0 + float64(sb.buf[bottomOffset+1])*sb.g1
			blurredB := float64(sb.buf[topOffset+2])*sb.g1 + float64(sb.buf[centerOffset+2])*sb.g0 + float64(sb.buf[bottomOffset+2])*sb.g1
			blurredA := float64(sb.buf[topOffset+3])*sb.g1 + float64(sb.buf[centerOffset+3])*sb.g0 + float64(sb.buf[bottomOffset+3])*sb.g1

			// Clamp values and write back to image
			blurredPixel := color.RGBA8[color.Linear]{
				R: basics.Int8u(math.Min(255, math.Max(0, blurredR))),
				G: basics.Int8u(math.Min(255, math.Max(0, blurredG))),
				B: basics.Int8u(math.Min(255, math.Max(0, blurredB))),
				A: basics.Int8u(math.Min(255, math.Max(0, blurredA))),
			}
			img.CopyPixel(imgX, imgY, blurredPixel)
		}
	}
}

// ApplySlightBlur is a convenience function for applying slight blur to an image.
func ApplySlightBlur(img PixFmtInterface, bounds basics.RectI, radius float64) {
	if radius > 0 {
		blur := NewSlightBlur(radius)
		blur.Blur(img, bounds)
	}
}

// ApplySlightBlurFull applies slight blur to the entire image.
func ApplySlightBlurFull(img PixFmtInterface, radius float64) {
	if radius > 0 {
		bounds := basics.RectI{X1: 0, Y1: 0, X2: img.Width() - 1, Y2: img.Height() - 1}
		ApplySlightBlur(img, bounds, radius)
	}
}
