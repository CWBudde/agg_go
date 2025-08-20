// Color conversion functors for 8-bit RGB formats.
// This corresponds to agg_color_conv_rgb8.h from the original AGG library.
package conv

import (
	"encoding/binary"

	"agg_go/internal/basics"
)

// Standard RGB8 conversion functors

// ColorConvRGB24 swaps R and B channels in 24-bit RGB data.
// Used for RGB24 ↔ BGR24 conversions.
type ColorConvRGB24 struct{}

// NewColorConvRGB24 creates a new RGB24 converter.
func NewColorConvRGB24() *ColorConvRGB24 {
	return &ColorConvRGB24{}
}

// CopyRow swaps R and B channels for each pixel in the row.
func (c *ColorConvRGB24) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	bytesNeeded := width * 3
	if len(dst) < bytesNeeded || len(src) < bytesNeeded {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 3
		dstIdx := i * 3

		// Swap R and B channels (keep G the same)
		dst[dstIdx] = src[srcIdx+2]   // R → B
		dst[dstIdx+1] = src[srcIdx+1] // G → G
		dst[dstIdx+2] = src[srcIdx]   // B → R
	}
}

// Type aliases for common RGB24 conversions
type (
	ColorConvRGB24ToBGR24 = ColorConvRGB24
	ColorConvBGR24ToRGB24 = ColorConvRGB24
)

// ColorConvRGBA32 performs channel reordering for 32-bit RGBA data.
// Template parameters I1, I2, I3, I4 specify the channel mapping.
type ColorConvRGBA32 struct {
	I1, I2, I3, I4 int // Channel indices for reordering
}

// NewColorConvRGBA32 creates a new RGBA32 converter with channel mapping.
func NewColorConvRGBA32(i1, i2, i3, i4 int) *ColorConvRGBA32 {
	return &ColorConvRGBA32{I1: i1, I2: i2, I3: i3, I4: i4}
}

// CopyRow reorders RGBA channels according to the mapping.
func (c *ColorConvRGBA32) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	bytesNeeded := width * 4
	if len(dst) < bytesNeeded || len(src) < bytesNeeded {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 4
		dstIdx := i * 4

		// Read source pixel
		tmp := [4]basics.Int8u{
			src[srcIdx],
			src[srcIdx+1],
			src[srcIdx+2],
			src[srcIdx+3],
		}

		// Write with channel reordering
		dst[dstIdx] = tmp[c.I1]
		dst[dstIdx+1] = tmp[c.I2]
		dst[dstIdx+2] = tmp[c.I3]
		dst[dstIdx+3] = tmp[c.I4]
	}
}

// Common RGBA32 conversion functions
func NewColorConvARGB32ToABGR32() *ColorConvRGBA32 { return NewColorConvRGBA32(0, 3, 2, 1) }
func NewColorConvARGB32ToBGRA32() *ColorConvRGBA32 { return NewColorConvRGBA32(3, 2, 1, 0) }
func NewColorConvARGB32ToRGBA32() *ColorConvRGBA32 { return NewColorConvRGBA32(1, 2, 3, 0) }
func NewColorConvBGRA32ToABGR32() *ColorConvRGBA32 { return NewColorConvRGBA32(3, 0, 1, 2) }
func NewColorConvBGRA32ToARGB32() *ColorConvRGBA32 { return NewColorConvRGBA32(3, 2, 1, 0) }
func NewColorConvBGRA32ToRGBA32() *ColorConvRGBA32 { return NewColorConvRGBA32(2, 1, 0, 3) }
func NewColorConvRGBA32ToABGR32() *ColorConvRGBA32 { return NewColorConvRGBA32(3, 2, 1, 0) }
func NewColorConvRGBA32ToARGB32() *ColorConvRGBA32 { return NewColorConvRGBA32(3, 0, 1, 2) }
func NewColorConvRGBA32ToBGRA32() *ColorConvRGBA32 { return NewColorConvRGBA32(2, 1, 0, 3) }
func NewColorConvABGR32ToARGB32() *ColorConvRGBA32 { return NewColorConvRGBA32(0, 3, 2, 1) }
func NewColorConvABGR32ToBGRA32() *ColorConvRGBA32 { return NewColorConvRGBA32(1, 2, 3, 0) }
func NewColorConvABGR32ToRGBA32() *ColorConvRGBA32 { return NewColorConvRGBA32(3, 2, 1, 0) }

// ColorConvRGB24RGBA32 converts RGB24 to RGBA32 with alpha = 255.
type ColorConvRGB24RGBA32 struct {
	I1, I2, I3, A int // Channel indices and alpha position
}

// NewColorConvRGB24RGBA32 creates a new RGB24 to RGBA32 converter.
func NewColorConvRGB24RGBA32(i1, i2, i3, a int) *ColorConvRGB24RGBA32 {
	return &ColorConvRGB24RGBA32{I1: i1, I2: i2, I3: i3, A: a}
}

// CopyRow converts RGB24 to RGBA32 with full alpha.
func (c *ColorConvRGB24RGBA32) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*4 || len(src) < width*3 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 3
		dstIdx := i * 4

		// Read RGB
		r := src[srcIdx]
		g := src[srcIdx+1]
		b := src[srcIdx+2]

		// Write RGBA with channel mapping
		dst[dstIdx+c.I1] = r
		dst[dstIdx+c.I2] = g
		dst[dstIdx+c.I3] = b
		dst[dstIdx+c.A] = 255
	}
}

// Common RGB24 to RGBA32 conversions
func NewColorConvRGB24ToARGB32() *ColorConvRGB24RGBA32 { return NewColorConvRGB24RGBA32(1, 2, 3, 0) }
func NewColorConvRGB24ToABGR32() *ColorConvRGB24RGBA32 { return NewColorConvRGB24RGBA32(3, 2, 1, 0) }
func NewColorConvRGB24ToBGRA32() *ColorConvRGB24RGBA32 { return NewColorConvRGB24RGBA32(2, 1, 0, 3) }
func NewColorConvRGB24ToRGBA32() *ColorConvRGB24RGBA32 { return NewColorConvRGB24RGBA32(0, 1, 2, 3) }
func NewColorConvBGR24ToARGB32() *ColorConvRGB24RGBA32 { return NewColorConvRGB24RGBA32(3, 2, 1, 0) }
func NewColorConvBGR24ToABGR32() *ColorConvRGB24RGBA32 { return NewColorConvRGB24RGBA32(1, 2, 3, 0) }
func NewColorConvBGR24ToBGRA32() *ColorConvRGB24RGBA32 { return NewColorConvRGB24RGBA32(0, 1, 2, 3) }
func NewColorConvBGR24ToRGBA32() *ColorConvRGB24RGBA32 { return NewColorConvRGB24RGBA32(2, 1, 0, 3) }

// ColorConvRGBA32RGB24 converts RGBA32 to RGB24 (drops alpha).
type ColorConvRGBA32RGB24 struct {
	I1, I2, I3 int // Channel indices for RGB extraction
}

// NewColorConvRGBA32RGB24 creates a new RGBA32 to RGB24 converter.
func NewColorConvRGBA32RGB24(i1, i2, i3 int) *ColorConvRGBA32RGB24 {
	return &ColorConvRGBA32RGB24{I1: i1, I2: i2, I3: i3}
}

// CopyRow converts RGBA32 to RGB24 by dropping alpha.
func (c *ColorConvRGBA32RGB24) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*3 || len(src) < width*4 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 4
		dstIdx := i * 3

		// Extract RGB channels
		dst[dstIdx] = src[srcIdx+c.I1]
		dst[dstIdx+1] = src[srcIdx+c.I2]
		dst[dstIdx+2] = src[srcIdx+c.I3]
	}
}

// Common RGBA32 to RGB24 conversions
func NewColorConvARGB32ToRGB24() *ColorConvRGBA32RGB24 { return NewColorConvRGBA32RGB24(1, 2, 3) }
func NewColorConvABGR32ToRGB24() *ColorConvRGBA32RGB24 { return NewColorConvRGBA32RGB24(3, 2, 1) }
func NewColorConvBGRA32ToRGB24() *ColorConvRGBA32RGB24 { return NewColorConvRGBA32RGB24(2, 1, 0) }
func NewColorConvRGBA32ToRGB24() *ColorConvRGBA32RGB24 { return NewColorConvRGBA32RGB24(0, 1, 2) }
func NewColorConvARGB32ToBGR24() *ColorConvRGBA32RGB24 { return NewColorConvRGBA32RGB24(3, 2, 1) }
func NewColorConvABGR32ToBGR24() *ColorConvRGBA32RGB24 { return NewColorConvRGBA32RGB24(1, 2, 3) }
func NewColorConvBGRA32ToBGR24() *ColorConvRGBA32RGB24 { return NewColorConvRGBA32RGB24(0, 1, 2) }
func NewColorConvRGBA32ToBGR24() *ColorConvRGBA32RGB24 { return NewColorConvRGBA32RGB24(2, 1, 0) }

// ColorConvRGB555RGB24 converts RGB555 to RGB24.
type ColorConvRGB555RGB24 struct {
	R, B int // Red and blue channel positions
}

// NewColorConvRGB555RGB24 creates a new RGB555 to RGB24 converter.
func NewColorConvRGB555RGB24(r, b int) *ColorConvRGB555RGB24 {
	return &ColorConvRGB555RGB24{R: r, B: b}
}

// CopyRow converts RGB555 packed format to RGB24.
func (c *ColorConvRGB555RGB24) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*3 || len(src) < width*2 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 2
		dstIdx := i * 3

		// Read 16-bit RGB555 value
		rgb := binary.LittleEndian.Uint16(src[srcIdx:])

		// Extract 5-bit components (match C++ AGG exactly)
		dst[dstIdx+c.R] = basics.Int8u((rgb >> 7) & 0xF8)
		dst[dstIdx+1] = basics.Int8u((rgb >> 2) & 0xF8)
		dst[dstIdx+c.B] = basics.Int8u((rgb << 3) & 0xF8)
	}
}

// Common RGB555 conversions
func NewColorConvRGB555ToBGR24() *ColorConvRGB555RGB24 { return NewColorConvRGB555RGB24(2, 0) }
func NewColorConvRGB555ToRGB24() *ColorConvRGB555RGB24 { return NewColorConvRGB555RGB24(0, 2) }

// ColorConvRGB24RGB555 converts RGB24 to RGB555.
type ColorConvRGB24RGB555 struct {
	R, B int // Red and blue channel positions
}

// NewColorConvRGB24RGB555 creates a new RGB24 to RGB555 converter.
func NewColorConvRGB24RGB555(r, b int) *ColorConvRGB24RGB555 {
	return &ColorConvRGB24RGB555{R: r, B: b}
}

// CopyRow converts RGB24 to RGB555 packed format.
func (c *ColorConvRGB24RGB555) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*2 || len(src) < width*3 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 3
		dstIdx := i * 2

		// Pack RGB into 15-bit RGB555
		r := uint16(src[srcIdx+c.R]) >> 3
		g := uint16(src[srcIdx+1]) >> 3
		b := uint16(src[srcIdx+c.B]) >> 3

		rgb555 := (r << 10) | (g << 5) | b
		binary.LittleEndian.PutUint16(dst[dstIdx:], rgb555)
	}
}

// Common RGB24 to RGB555 conversions
func NewColorConvBGR24ToRGB555() *ColorConvRGB24RGB555 { return NewColorConvRGB24RGB555(2, 0) }
func NewColorConvRGB24ToRGB555() *ColorConvRGB24RGB555 { return NewColorConvRGB24RGB555(0, 2) }

// ColorConvRGB565RGB24 converts RGB565 to RGB24.
type ColorConvRGB565RGB24 struct {
	R, B int // Red and blue channel positions
}

// NewColorConvRGB565RGB24 creates a new RGB565 to RGB24 converter.
func NewColorConvRGB565RGB24(r, b int) *ColorConvRGB565RGB24 {
	return &ColorConvRGB565RGB24{R: r, B: b}
}

// CopyRow converts RGB565 packed format to RGB24.
func (c *ColorConvRGB565RGB24) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*3 || len(src) < width*2 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 2
		dstIdx := i * 3

		// Read 16-bit RGB565 value
		rgb := binary.LittleEndian.Uint16(src[srcIdx:])

		// Extract components (match C++ AGG exactly)
		dst[dstIdx+c.R] = basics.Int8u((rgb >> 8) & 0xF8)
		dst[dstIdx+1] = basics.Int8u((rgb >> 3) & 0xFC)
		dst[dstIdx+c.B] = basics.Int8u((rgb << 3) & 0xF8)
	}
}

// Common RGB565 conversions
func NewColorConvRGB565ToBGR24() *ColorConvRGB565RGB24 { return NewColorConvRGB565RGB24(2, 0) }
func NewColorConvRGB565ToRGB24() *ColorConvRGB565RGB24 { return NewColorConvRGB565RGB24(0, 2) }

// ColorConvRGB24RGB565 converts RGB24 to RGB565.
type ColorConvRGB24RGB565 struct {
	R, B int // Red and blue channel positions
}

// NewColorConvRGB24RGB565 creates a new RGB24 to RGB565 converter.
func NewColorConvRGB24RGB565(r, b int) *ColorConvRGB24RGB565 {
	return &ColorConvRGB24RGB565{R: r, B: b}
}

// CopyRow converts RGB24 to RGB565 packed format.
func (c *ColorConvRGB24RGB565) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*2 || len(src) < width*3 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 3
		dstIdx := i * 2

		// Pack RGB into 16-bit RGB565
		r := uint16(src[srcIdx+c.R]) >> 3
		g := uint16(src[srcIdx+1]) >> 2
		b := uint16(src[srcIdx+c.B]) >> 3

		rgb565 := (r << 11) | (g << 5) | b
		binary.LittleEndian.PutUint16(dst[dstIdx:], rgb565)
	}
}

// Common RGB24 to RGB565 conversions
func NewColorConvBGR24ToRGB565() *ColorConvRGB24RGB565 { return NewColorConvRGB24RGB565(2, 0) }
func NewColorConvRGB24ToRGB565() *ColorConvRGB24RGB565 { return NewColorConvRGB24RGB565(0, 2) }

// ColorConvRGB24Gray8 converts RGB24 to 8-bit grayscale using luminance formula.
type ColorConvRGB24Gray8 struct {
	R, B int // Red and blue channel positions
}

// NewColorConvRGB24Gray8 creates a new RGB24 to grayscale converter.
func NewColorConvRGB24Gray8(r, b int) *ColorConvRGB24Gray8 {
	return &ColorConvRGB24Gray8{R: r, B: b}
}

// CopyRow converts RGB24 to grayscale using ITU-R BT.601 luminance weights.
func (c *ColorConvRGB24Gray8) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width || len(src) < width*3 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 3

		// ITU-R BT.601 luminance: Y = 0.299*R + 0.587*G + 0.114*B
		// Using integer math: (77*R + 150*G + 29*B) >> 8
		r := int(src[srcIdx+c.R])
		g := int(src[srcIdx+1])
		b := int(src[srcIdx+c.B])

		gray := (77*r + 150*g + 29*b) >> 8
		dst[i] = basics.Int8u(gray)
	}
}

// Common RGB to grayscale conversions
func NewColorConvRGB24ToGray8() *ColorConvRGB24Gray8 { return NewColorConvRGB24Gray8(0, 2) }
func NewColorConvBGR24ToGray8() *ColorConvRGB24Gray8 { return NewColorConvRGB24Gray8(2, 0) }
