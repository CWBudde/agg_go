// Package agg2d: float image source adapter. This is the float (RGBA32) twin of
// imagePixelFormat in adapters.go. It exposes an ImageFloat through the
// source-side accessors used by the RGBA32 image-filter span generators,
// reproducing AGG's image_accessor_clone edge-clamp behavior over float32 rows.
package agg2d

import (
	"github.com/cwbudde/agg_go/internal/color"
)

// imagePixelFormatFloat exposes an ImageFloat as a span.RGBA32SourceInterface.
type imagePixelFormatFloat struct {
	img      *ImageFloat
	x, y, x0 int
	pixelBuf [4]float32
}

func newImagePixelFormatFloat(img *ImageFloat) *imagePixelFormatFloat {
	return &imagePixelFormatFloat{img: img}
}

func (ipf *imagePixelFormatFloat) Width() int  { return ipf.img.Width() }
func (ipf *imagePixelFormatFloat) Height() int { return ipf.img.Height() }

// OrderType returns the channel layout (ImageFloat stores straight R,G,B,A).
func (ipf *imagePixelFormatFloat) OrderType() color.ColorOrder {
	return color.OrderRGBA
}

// RowPtr returns the float32 row bytes for the given row, or nil if out of bounds.
func (ipf *imagePixelFormatFloat) RowPtr(y int) []float32 {
	if ipf.img == nil || ipf.img.renBuf == nil || y < 0 || y >= ipf.img.height {
		return nil
	}
	row := ipf.img.renBuf.Row(y)
	rowLen := ipf.img.width * imageFloatChannels
	if len(row) < rowLen {
		return nil
	}
	return row[:rowLen]
}

// RowData mirrors RowPtr for parity with the 8-bit adapter naming.
func (ipf *imagePixelFormatFloat) RowData(y int) []float32 {
	return ipf.RowPtr(y)
}

// pixelSliceClamped returns the 4-float pixel at clamped (x,y), reproducing
// image_accessor_clone: coordinates outside the image are clamped to the edge.
func (ipf *imagePixelFormatFloat) pixelSliceClamped(x, y int) []float32 {
	if ipf.img == nil || ipf.img.Data == nil || ipf.img.width <= 0 || ipf.img.height <= 0 {
		ipf.pixelBuf = [4]float32{}
		return ipf.pixelBuf[:]
	}

	if x < 0 {
		x = 0
	} else if x >= ipf.img.width {
		x = ipf.img.width - 1
	}
	if y < 0 {
		y = 0
	} else if y >= ipf.img.height {
		y = ipf.img.height - 1
	}

	row := ipf.RowPtr(y)
	if row == nil {
		ipf.pixelBuf = [4]float32{}
		return ipf.pixelBuf[:]
	}
	offset := x * imageFloatChannels
	if offset+imageFloatChannels > len(row) {
		ipf.pixelBuf = [4]float32{}
		return ipf.pixelBuf[:]
	}

	ipf.pixelBuf = [4]float32{row[offset], row[offset+1], row[offset+2], row[offset+3]}
	return ipf.pixelBuf[:]
}

// Span initializes source sampling at (x,y) and returns the first pixel.
func (ipf *imagePixelFormatFloat) Span(x, y, length int) []float32 {
	_ = length
	ipf.x = x
	ipf.x0 = x
	ipf.y = y
	return ipf.pixelSliceClamped(x, y)
}

// NextX advances sampling by one pixel in x.
func (ipf *imagePixelFormatFloat) NextX() []float32 {
	ipf.x++
	return ipf.pixelSliceClamped(ipf.x, ipf.y)
}

// NextY advances sampling by one row at the original x position.
func (ipf *imagePixelFormatFloat) NextY() []float32 {
	ipf.y++
	ipf.x = ipf.x0
	return ipf.pixelSliceClamped(ipf.x, ipf.y)
}
