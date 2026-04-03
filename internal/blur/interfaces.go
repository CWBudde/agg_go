// Package blur implements AGG's stack_blur family of algorithms for various
// pixel formats.  The package mirrors C++ AGG's agg_blur.h, providing reusable
// blur instances parameterised on colour type and colour space.
package blur

// PixelReader is the read-only image contract used by analysis algorithms
// (e.g. Sobel gradient).  Any AGG pixel-format adapter satisfies this.
type PixelReader[C any] interface {
	Width() int
	Height() int
	Pixel(x, y int) C
}

// PixelReadWriter extends PixelReader with the span-write capabilities that
// C++ AGG's in-place blur algorithms require: pixel() for reading,
// copy_color_hspan() for writing processed rows back, and copy_color_vspan()
// for writing processed columns.  The vertical span method is needed by
// pixfmt_transposer which maps copy_color_hspan → copy_color_vspan on the
// underlying image (see agg_pixfmt_transposer.h).
type PixelReadWriter[C any] interface {
	PixelReader[C]
	CopyColorHspan(x, y, length int, colors []C)
	CopyColorVspan(x, y, length int, colors []C)
}

// Transposer swaps the X/Y axes of a PixelReadWriter so that a horizontal
// blur pass applied to the transposed view becomes a vertical blur.  This is
// the Go equivalent of C++ AGG's pixfmt_transposer (agg_pixfmt_transposer.h).
type Transposer[C any] struct {
	Img PixelReadWriter[C]
}

func (t *Transposer[C]) Width() int       { return t.Img.Height() }
func (t *Transposer[C]) Height() int      { return t.Img.Width() }
func (t *Transposer[C]) Pixel(x, y int) C { return t.Img.Pixel(y, x) }

func (t *Transposer[C]) CopyColorHspan(x, y, length int, colors []C) {
	t.Img.CopyColorVspan(y, x, length, colors)
}

func (t *Transposer[C]) CopyColorVspan(x, y, length int, colors []C) {
	t.Img.CopyColorHspan(y, x, length, colors)
}
