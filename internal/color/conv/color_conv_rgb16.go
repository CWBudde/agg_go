// Color conversion functors for 16-bit RGB formats.
// This corresponds to agg_color_conv_rgb16.h from the original AGG library.
package conv

import (
	"agg_go/internal/basics"
	"encoding/binary"
)

// ColorConvGray16ToGray8 converts 16-bit grayscale to 8-bit grayscale.
type ColorConvGray16ToGray8 struct{}

// NewColorConvGray16ToGray8 creates a new Gray16 to Gray8 converter.
func NewColorConvGray16ToGray8() *ColorConvGray16ToGray8 {
	return &ColorConvGray16ToGray8{}
}

// CopyRow converts 16-bit grayscale to 8-bit by taking the high byte.
func (c *ColorConvGray16ToGray8) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width || len(src) < width*2 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 2
		// Extract high byte from 16-bit value
		gray16 := binary.LittleEndian.Uint16(src[srcIdx:])
		dst[i] = basics.Int8u(gray16 >> 8)
	}
}

// ColorConvRGB24RGB48 converts RGB24 to RGB48 by duplicating each byte.
type ColorConvRGB24RGB48 struct {
	I1, I3 int // Channel indices for R and B swapping
}

// NewColorConvRGB24RGB48 creates a new RGB24 to RGB48 converter.
func NewColorConvRGB24RGB48(i1, i3 int) *ColorConvRGB24RGB48 {
	return &ColorConvRGB24RGB48{I1: i1, I3: i3}
}

// CopyRow converts RGB24 to RGB48 with channel reordering.
func (c *ColorConvRGB24RGB48) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*6 || len(src) < width*3 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 3
		dstIdx := i * 6

		// Read RGB24 components
		r := src[srcIdx+c.I1]
		g := src[srcIdx+1]
		b := src[srcIdx+c.I3]

		// Convert to RGB48 by duplicating bytes: (value << 8) | value
		binary.LittleEndian.PutUint16(dst[dstIdx:], uint16(r)<<8|uint16(r))
		binary.LittleEndian.PutUint16(dst[dstIdx+2:], uint16(g)<<8|uint16(g))
		binary.LittleEndian.PutUint16(dst[dstIdx+4:], uint16(b)<<8|uint16(b))
	}
}

// Common RGB24 to RGB48 conversions
func NewColorConvRGB24ToRGB48() *ColorConvRGB24RGB48 { return NewColorConvRGB24RGB48(0, 2) }
func NewColorConvBGR24ToBGR48() *ColorConvRGB24RGB48 { return NewColorConvRGB24RGB48(0, 2) }
func NewColorConvRGB24ToBGR48() *ColorConvRGB24RGB48 { return NewColorConvRGB24RGB48(2, 0) }
func NewColorConvBGR24ToRGB48() *ColorConvRGB24RGB48 { return NewColorConvRGB24RGB48(2, 0) }

// ColorConvRGB48RGB24 converts RGB48 to RGB24 by taking high bytes.
type ColorConvRGB48RGB24 struct {
	I1, I3 int // Channel indices for R and B swapping
}

// NewColorConvRGB48RGB24 creates a new RGB48 to RGB24 converter.
func NewColorConvRGB48RGB24(i1, i3 int) *ColorConvRGB48RGB24 {
	return &ColorConvRGB48RGB24{I1: i1, I3: i3}
}

// CopyRow converts RGB48 to RGB24 with channel reordering.
func (c *ColorConvRGB48RGB24) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*3 || len(src) < width*6 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 6
		dstIdx := i * 3

		// Read RGB48 components (high bytes only)
		r := binary.LittleEndian.Uint16(src[srcIdx:]) >> 8
		g := binary.LittleEndian.Uint16(src[srcIdx+2:]) >> 8
		b := binary.LittleEndian.Uint16(src[srcIdx+4:]) >> 8

		// Write RGB24 with channel reordering
		dst[dstIdx+c.I1] = basics.Int8u(r)
		dst[dstIdx+1] = basics.Int8u(g)
		dst[dstIdx+c.I3] = basics.Int8u(b)
	}
}

// Common RGB48 to RGB24 conversions
func NewColorConvRGB48ToRGB24() *ColorConvRGB48RGB24 { return NewColorConvRGB48RGB24(0, 2) }
func NewColorConvBGR48ToBGR24() *ColorConvRGB48RGB24 { return NewColorConvRGB48RGB24(0, 2) }
func NewColorConvRGB48ToBGR24() *ColorConvRGB48RGB24 { return NewColorConvRGB48RGB24(2, 0) }
func NewColorConvBGR48ToRGB24() *ColorConvRGB48RGB24 { return NewColorConvRGB48RGB24(2, 0) }

// ColorConvRGBAAAToRGB24 converts 10-bit packed RGB (AAA format) to RGB24.
// Format: RRRRRRRRRRGGGGGGGGGGBBBBBBBBBBAA (32-bit)
type ColorConvRGBAAAToRGB24 struct {
	R, B int // Red and blue channel positions in output
}

// NewColorConvRGBAAAToRGB24 creates a new 10-bit AAA to RGB24 converter.
func NewColorConvRGBAAAToRGB24(r, b int) *ColorConvRGBAAAToRGB24 {
	return &ColorConvRGBAAAToRGB24{R: r, B: b}
}

// CopyRow converts 10-bit RGB AAA format to RGB24.
func (c *ColorConvRGBAAAToRGB24) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*3 || len(src) < width*4 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 4
		dstIdx := i * 3

		// Read 32-bit packed value
		rgb := binary.LittleEndian.Uint32(src[srcIdx:])

		// Extract 10-bit components and convert to 8-bit
		dst[dstIdx+c.R] = basics.Int8u(rgb >> 22) // R: bits 31-22
		dst[dstIdx+1] = basics.Int8u(rgb >> 12)   // G: bits 21-12
		dst[dstIdx+c.B] = basics.Int8u(rgb >> 2)  // B: bits 11-2
	}
}

// Common 10-bit AAA conversions
func NewColorConvRGBAAAToRGB24Std() *ColorConvRGBAAAToRGB24 { return NewColorConvRGBAAAToRGB24(0, 2) }
func NewColorConvRGBAAAToBGR24() *ColorConvRGBAAAToRGB24    { return NewColorConvRGBAAAToRGB24(2, 0) }
func NewColorConvBGRAAAToRGB24() *ColorConvRGBAAAToRGB24    { return NewColorConvRGBAAAToRGB24(2, 0) }
func NewColorConvBGRAAAToBGR24() *ColorConvRGBAAAToRGB24    { return NewColorConvRGBAAAToRGB24(0, 2) }

// ColorConvRGBBBAToRGB24 converts 10-bit packed RGB (BBA format) to RGB24.
// Format: RRRRRRRRRRRRGGGGGGGGGGBBBBBBBBAA (32-bit)
type ColorConvRGBBBAToRGB24 struct {
	R, B int // Red and blue channel positions in output
}

// NewColorConvRGBBBAToRGB24 creates a new 10-bit BBA to RGB24 converter.
func NewColorConvRGBBBAToRGB24(r, b int) *ColorConvRGBBBAToRGB24 {
	return &ColorConvRGBBBAToRGB24{R: r, B: b}
}

// CopyRow converts 10-bit RGB BBA format to RGB24.
func (c *ColorConvRGBBBAToRGB24) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*3 || len(src) < width*4 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 4
		dstIdx := i * 3

		// Read 32-bit packed value
		rgb := binary.LittleEndian.Uint32(src[srcIdx:])

		// Extract components with different bit positions
		dst[dstIdx+c.R] = basics.Int8u(rgb >> 24) // R: bits 31-24
		dst[dstIdx+1] = basics.Int8u(rgb >> 13)   // G: bits 23-13
		dst[dstIdx+c.B] = basics.Int8u(rgb >> 2)  // B: bits 12-2
	}
}

// Common 10-bit BBA conversions
func NewColorConvRGBBBAToRGB24Std() *ColorConvRGBBBAToRGB24 { return NewColorConvRGBBBAToRGB24(0, 2) }
func NewColorConvRGBBBAToBGR24() *ColorConvRGBBBAToRGB24    { return NewColorConvRGBBBAToRGB24(2, 0) }

// ColorConvBGRABBToRGB24 converts 10-bit packed BGR (ABB format) to RGB24.
// Format: BBBBBBBBBBGGGGGGGGGGAARRRRRRRRR (32-bit)
type ColorConvBGRABBToRGB24 struct {
	B, R int // Blue and red channel positions in output
}

// NewColorConvBGRABBToRGB24 creates a new 10-bit ABB to RGB24 converter.
func NewColorConvBGRABBToRGB24(b, r int) *ColorConvBGRABBToRGB24 {
	return &ColorConvBGRABBToRGB24{B: b, R: r}
}

// CopyRow converts 10-bit BGR ABB format to RGB24.
func (c *ColorConvBGRABBToRGB24) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*3 || len(src) < width*4 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 4
		dstIdx := i * 3

		// Read 32-bit packed value
		bgr := binary.LittleEndian.Uint32(src[srcIdx:])

		// Extract BGR components
		dst[dstIdx+c.R] = basics.Int8u(bgr >> 3)  // R: bits 12-3
		dst[dstIdx+1] = basics.Int8u(bgr >> 14)   // G: bits 23-14
		dst[dstIdx+c.B] = basics.Int8u(bgr >> 24) // B: bits 31-24
	}
}

// Common 10-bit ABB conversions
func NewColorConvBGRABBToRGB24Std() *ColorConvBGRABBToRGB24 { return NewColorConvBGRABBToRGB24(2, 0) }
func NewColorConvBGRABBToBGR24() *ColorConvBGRABBToRGB24    { return NewColorConvBGRABBToRGB24(0, 2) }

// ColorConvRGBA64RGBA32 converts RGBA64 to RGBA32 by taking high bytes.
type ColorConvRGBA64RGBA32 struct {
	I1, I2, I3, I4 int // Channel indices for reordering
}

// NewColorConvRGBA64RGBA32 creates a new RGBA64 to RGBA32 converter.
func NewColorConvRGBA64RGBA32(i1, i2, i3, i4 int) *ColorConvRGBA64RGBA32 {
	return &ColorConvRGBA64RGBA32{I1: i1, I2: i2, I3: i3, I4: i4}
}

// CopyRow converts RGBA64 to RGBA32 with channel reordering.
func (c *ColorConvRGBA64RGBA32) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*4 || len(src) < width*8 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 8
		dstIdx := i * 4

		// Read RGBA64 components (high bytes only)
		rgba64 := [4]basics.Int8u{
			basics.Int8u(binary.LittleEndian.Uint16(src[srcIdx:]) >> 8),
			basics.Int8u(binary.LittleEndian.Uint16(src[srcIdx+2:]) >> 8),
			basics.Int8u(binary.LittleEndian.Uint16(src[srcIdx+4:]) >> 8),
			basics.Int8u(binary.LittleEndian.Uint16(src[srcIdx+6:]) >> 8),
		}

		// Write RGBA32 with channel reordering
		dst[dstIdx] = rgba64[c.I1]
		dst[dstIdx+1] = rgba64[c.I2]
		dst[dstIdx+2] = rgba64[c.I3]
		dst[dstIdx+3] = rgba64[c.I4]
	}
}

// Common RGBA64 to RGBA32 conversions
func NewColorConvRGBA64ToRGBA32() *ColorConvRGBA64RGBA32 { return NewColorConvRGBA64RGBA32(0, 1, 2, 3) }
func NewColorConvARGB64ToARGB32() *ColorConvRGBA64RGBA32 { return NewColorConvRGBA64RGBA32(0, 1, 2, 3) }
func NewColorConvBGRA64ToBGRA32() *ColorConvRGBA64RGBA32 { return NewColorConvRGBA64RGBA32(0, 1, 2, 3) }
func NewColorConvABGR64ToABGR32() *ColorConvRGBA64RGBA32 { return NewColorConvRGBA64RGBA32(0, 1, 2, 3) }
func NewColorConvARGB64ToABGR32() *ColorConvRGBA64RGBA32 { return NewColorConvRGBA64RGBA32(0, 3, 2, 1) }
func NewColorConvARGB64ToBGRA32() *ColorConvRGBA64RGBA32 { return NewColorConvRGBA64RGBA32(3, 2, 1, 0) }
func NewColorConvARGB64ToRGBA32() *ColorConvRGBA64RGBA32 { return NewColorConvRGBA64RGBA32(1, 2, 3, 0) }

// ColorConvRGB24RGBA64 converts RGB24 to RGBA64 with full alpha.
type ColorConvRGB24RGBA64 struct {
	I1, I2, I3, A int // Channel indices and alpha position
}

// NewColorConvRGB24RGBA64 creates a new RGB24 to RGBA64 converter.
func NewColorConvRGB24RGBA64(i1, i2, i3, a int) *ColorConvRGB24RGBA64 {
	return &ColorConvRGB24RGBA64{I1: i1, I2: i2, I3: i3, A: a}
}

// CopyRow converts RGB24 to RGBA64 with full alpha.
func (c *ColorConvRGB24RGBA64) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*8 || len(src) < width*3 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 3
		dstIdx := i * 8

		// Read RGB24
		r := src[srcIdx]
		g := src[srcIdx+1]
		b := src[srcIdx+2]

		// Convert to 16-bit by duplicating bytes
		r16 := uint16(r)<<8 | uint16(r)
		g16 := uint16(g)<<8 | uint16(g)
		b16 := uint16(b)<<8 | uint16(b)

		// Write RGBA64 with channel mapping
		binary.LittleEndian.PutUint16(dst[dstIdx+c.I1*2:], r16)
		binary.LittleEndian.PutUint16(dst[dstIdx+c.I2*2:], g16)
		binary.LittleEndian.PutUint16(dst[dstIdx+c.I3*2:], b16)
		binary.LittleEndian.PutUint16(dst[dstIdx+c.A*2:], 65535) // Full alpha
	}
}

// Common RGB24 to RGBA64 conversions
func NewColorConvRGB24ToARGB64() *ColorConvRGB24RGBA64 { return NewColorConvRGB24RGBA64(1, 2, 3, 0) }
func NewColorConvRGB24ToABGR64() *ColorConvRGB24RGBA64 { return NewColorConvRGB24RGBA64(3, 2, 1, 0) }
func NewColorConvRGB24ToBGRA64() *ColorConvRGB24RGBA64 { return NewColorConvRGB24RGBA64(2, 1, 0, 3) }
func NewColorConvRGB24ToRGBA64() *ColorConvRGB24RGBA64 { return NewColorConvRGB24RGBA64(0, 1, 2, 3) }
func NewColorConvBGR24ToARGB64() *ColorConvRGB24RGBA64 { return NewColorConvRGB24RGBA64(3, 2, 1, 0) }
func NewColorConvBGR24ToABGR64() *ColorConvRGB24RGBA64 { return NewColorConvRGB24RGBA64(1, 2, 3, 0) }
func NewColorConvBGR24ToBGRA64() *ColorConvRGB24RGBA64 { return NewColorConvRGB24RGBA64(0, 1, 2, 3) }
func NewColorConvBGR24ToRGBA64() *ColorConvRGB24RGBA64 { return NewColorConvRGB24RGBA64(2, 1, 0, 3) }

// ColorConvRGB24Gray16 converts RGB24 to 16-bit grayscale.
type ColorConvRGB24Gray16 struct {
	R, B int // Red and blue channel positions
}

// NewColorConvRGB24Gray16 creates a new RGB24 to Gray16 converter.
func NewColorConvRGB24Gray16(r, b int) *ColorConvRGB24Gray16 {
	return &ColorConvRGB24Gray16{R: r, B: b}
}

// CopyRow converts RGB24 to 16-bit grayscale using luminance formula.
func (c *ColorConvRGB24Gray16) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	if len(dst) < width*2 || len(src) < width*3 {
		return
	}

	for i := 0; i < width; i++ {
		srcIdx := i * 3
		dstIdx := i * 2

		// ITU-R BT.601 luminance: Y = 0.299*R + 0.587*G + 0.114*B
		// Using integer math for 16-bit result
		r := int(src[srcIdx+c.R])
		g := int(src[srcIdx+1])
		b := int(src[srcIdx+c.B])

		gray16 := uint16(77*r + 150*g + 29*b) // Result in range [0, 65280]
		binary.LittleEndian.PutUint16(dst[dstIdx:], gray16)
	}
}

// Common RGB24 to Gray16 conversions
func NewColorConvRGB24ToGray16() *ColorConvRGB24Gray16 { return NewColorConvRGB24Gray16(0, 2) }
func NewColorConvBGR24ToGray16() *ColorConvRGB24Gray16 { return NewColorConvRGB24Gray16(2, 0) }
