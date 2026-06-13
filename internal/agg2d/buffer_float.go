// Package agg2d float-backed image and buffer management for the float (128-bit,
// 4 x float32) Agg2D variant. This file is the float twin of buffer.go's Image
// and provides the boundary conversions called for by PLAN.md Phase 4 (L3).
package agg2d

import (
	"image"

	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
)

// Boundary contract (PLAN.md Phase 4)
//
// ImageFloat stores STRAIGHT (non-premultiplied) RGBA float32 data with channels
// nominally in [0,1], laid out as 4 contiguous float32 per pixel in channel order
// R,G,B,A. This mirrors the 8-bit Image's straight-data contract: premultiply /
// demultiply happens inside the pixfmt blenders, not at this boundary.
//
// Conversions to/from external formats honor each format's own alpha convention:
//   - color.RGBA32          : straight  <-> straight   (GetPixel/SetPixel)
//   - image.NRGBA64 (Go)    : straight  <-> straight   (non-premultiplied, 16-bit)
//   - image.RGBA    (Go)    : straight  <-> premultiplied (Go mandates premul, 8-bit)
//   - 8-bit AGG Image       : straight  <-> straight   (channel scale by 255)

const imageFloatChannels = 4

// ImageFloat represents a float (128-bit) raster image usable as a rendering
// target or image source. It is the float counterpart of Image.
type ImageFloat struct {
	renBuf *buffer.RenderingBufferF32
	Data   []float32 // straight RGBA float32, 4 per pixel, channels in [0,1]
	width  int
	height int
}

// NewImageFloat creates a float image wrapping the given buffer. stride is in
// bytes per row (matching the rendering buffer convention).
func NewImageFloat(buf []float32, width, height, stride int) *ImageFloat {
	img := &ImageFloat{
		Data:   buf,
		width:  width,
		height: height,
	}
	img.renBuf = buffer.NewRenderingBufferF32()
	img.renBuf.Attach(buf, width, height, stride)
	return img
}

// NewImageFloatEmpty allocates a zeroed float image of the given dimensions with
// a tightly packed row stride.
func NewImageFloatEmpty(width, height int) *ImageFloat {
	data := make([]float32, width*height*imageFloatChannels)
	return NewImageFloat(data, width, height, width*imageFloatChannels*4)
}

// Width returns the image width in pixels.
func (img *ImageFloat) Width() int { return img.width }

// Height returns the image height in pixels.
func (img *ImageFloat) Height() int { return img.height }

// Stride returns the row stride in bytes.
func (img *ImageFloat) Stride() int {
	if img.renBuf != nil {
		return img.renBuf.Stride()
	}
	return img.width * imageFloatChannels * 4
}

// IsAttached reports whether the image has buffer data attached.
func (img *ImageFloat) IsAttached() bool {
	return img.renBuf != nil && img.Data != nil
}

// Attach attaches buffer data to the image. stride is in bytes per row.
func (img *ImageFloat) Attach(buf []float32, width, height, stride int) {
	img.Data = buf
	img.width = width
	img.height = height
	if img.renBuf == nil {
		img.renBuf = buffer.NewRenderingBufferF32()
	}
	img.renBuf.Attach(buf, width, height, stride)
}

// pixel returns the 4-float slice for the pixel at (x,y), or nil if out of bounds.
func (img *ImageFloat) pixel(x, y int) []float32 {
	if img.renBuf == nil || x < 0 || y < 0 || x >= img.width || y >= img.height {
		return nil
	}
	row := img.renBuf.Row(y)
	off := x * imageFloatChannels
	if off < 0 || off+imageFloatChannels > len(row) {
		return nil
	}
	return row[off : off+imageFloatChannels]
}

// GetPixel returns the straight (non-premultiplied) color at (x,y). Out-of-bounds
// coordinates return transparent black.
func (img *ImageFloat) GetPixel(x, y int) color.RGBA32[color.Linear] {
	px := img.pixel(x, y)
	if px == nil {
		return color.RGBA32[color.Linear]{}
	}
	return color.RGBA32[color.Linear]{R: px[0], G: px[1], B: px[2], A: px[3]}
}

// SetPixel writes the straight (non-premultiplied) color at (x,y). Out-of-bounds
// coordinates are a no-op.
func (img *ImageFloat) SetPixel(x, y int, c color.RGBA32[color.Linear]) {
	px := img.pixel(x, y)
	if px == nil {
		return
	}
	px[0], px[1], px[2], px[3] = c.R, c.G, c.B, c.A
}

// Premultiply converts the image in place from straight to premultiplied alpha.
func (img *ImageFloat) Premultiply() {
	for i := 0; i+imageFloatChannels <= len(img.Data); i += imageFloatChannels {
		a := img.Data[i+3]
		if a <= 0 {
			img.Data[i+0], img.Data[i+1], img.Data[i+2] = 0, 0, 0
		} else if a < 1 {
			img.Data[i+0] *= a
			img.Data[i+1] *= a
			img.Data[i+2] *= a
		}
	}
}

// Demultiply converts the image in place from premultiplied to straight alpha.
func (img *ImageFloat) Demultiply() {
	for i := 0; i+imageFloatChannels <= len(img.Data); i += imageFloatChannels {
		a := img.Data[i+3]
		if a <= 0 {
			img.Data[i+0], img.Data[i+1], img.Data[i+2] = 0, 0, 0
		} else if a < 1 {
			inv := 1.0 / a
			img.Data[i+0] *= inv
			img.Data[i+1] *= inv
			img.Data[i+2] *= inv
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// Boundary conversions
////////////////////////////////////////////////////////////////////////////////

// clampUnit clamps a float32 to [0,1].
func clampUnit(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// roundToU16 maps a unit float to a 16-bit channel value with rounding.
func roundToU16(v float32) uint16 {
	return uint16(clampUnit(v)*65535 + 0.5)
}

// roundToU8 maps a unit float to an 8-bit channel value with rounding.
func roundToU8(v float32) uint8 {
	return uint8(clampUnit(v)*255 + 0.5)
}

// ToNRGBA64 returns a Go image.NRGBA64 (straight, 16-bit) copy of the image.
func (img *ImageFloat) ToNRGBA64() *image.NRGBA64 {
	dst := image.NewNRGBA64(image.Rect(0, 0, img.width, img.height))
	for y := range img.height {
		for x := range img.width {
			c := img.GetPixel(x, y)
			o := dst.PixOffset(x, y)
			r, g, b, a := roundToU16(c.R), roundToU16(c.G), roundToU16(c.B), roundToU16(c.A)
			dst.Pix[o+0] = uint8(r >> 8)
			dst.Pix[o+1] = uint8(r)
			dst.Pix[o+2] = uint8(g >> 8)
			dst.Pix[o+3] = uint8(g)
			dst.Pix[o+4] = uint8(b >> 8)
			dst.Pix[o+5] = uint8(b)
			dst.Pix[o+6] = uint8(a >> 8)
			dst.Pix[o+7] = uint8(a)
		}
	}
	return dst
}

// NewImageFloatFromNRGBA64 builds a float image from a Go image.NRGBA64 (straight).
func NewImageFloatFromNRGBA64(src *image.NRGBA64) *ImageFloat {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	img := NewImageFloatEmpty(w, h)
	for y := range h {
		for x := range w {
			c := src.NRGBA64At(b.Min.X+x, b.Min.Y+y)
			img.SetPixel(x, y, color.RGBA32[color.Linear]{
				R: float32(c.R) / 65535,
				G: float32(c.G) / 65535,
				B: float32(c.B) / 65535,
				A: float32(c.A) / 65535,
			})
		}
	}
	return img
}

// ToRGBA returns a Go image.RGBA (alpha-premultiplied, 8-bit) copy of the image.
func (img *ImageFloat) ToRGBA() *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, img.width, img.height))
	for y := range img.height {
		for x := range img.width {
			c := img.GetPixel(x, y)
			a := clampUnit(c.A)
			o := dst.PixOffset(x, y)
			dst.Pix[o+0] = roundToU8(c.R * a)
			dst.Pix[o+1] = roundToU8(c.G * a)
			dst.Pix[o+2] = roundToU8(c.B * a)
			dst.Pix[o+3] = roundToU8(a)
		}
	}
	return dst
}

// NewImageFloatFromRGBA builds a straight float image from a Go image.RGBA
// (alpha-premultiplied). RGB channels are demultiplied by alpha.
func NewImageFloatFromRGBA(src *image.RGBA) *ImageFloat {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	img := NewImageFloatEmpty(w, h)
	for y := range h {
		for x := range w {
			o := src.PixOffset(b.Min.X+x, b.Min.Y+y)
			a := float32(src.Pix[o+3]) / 255
			r := float32(src.Pix[o+0]) / 255
			g := float32(src.Pix[o+1]) / 255
			bl := float32(src.Pix[o+2]) / 255
			if a > 0 {
				inv := 1.0 / a
				r, g, bl = r*inv, g*inv, bl*inv
			} else {
				r, g, bl = 0, 0, 0
			}
			img.SetPixel(x, y, color.RGBA32[color.Linear]{R: r, G: g, B: bl, A: a})
		}
	}
	return img
}

// ToImage8 returns an 8-bit AGG Image holding straight RGBA8 data (channels
// scaled by 255, no premultiply).
func (img *ImageFloat) ToImage8() *Image {
	w, h := img.width, img.height
	data := make([]uint8, w*h*4)
	for y := range h {
		for x := range w {
			c := img.GetPixel(x, y)
			o := (y*w + x) * 4
			data[o+0] = roundToU8(c.R)
			data[o+1] = roundToU8(c.G)
			data[o+2] = roundToU8(c.B)
			data[o+3] = roundToU8(c.A)
		}
	}
	return NewImage(data, w, h, w*4)
}

// NewImageFloatFromImage8 builds a straight float image from an 8-bit AGG Image
// holding straight RGBA8 data (channels scaled to [0,1], no demultiply).
func NewImageFloatFromImage8(src *Image) *ImageFloat {
	w, h := src.Width(), src.Height()
	img := NewImageFloatEmpty(w, h)
	for y := range h {
		for x := range w {
			px := src.GetPixel(x, y)
			img.SetPixel(x, y, color.RGBA32[color.Linear]{
				R: float32(px[0]) / 255,
				G: float32(px[1]) / 255,
				B: float32(px[2]) / 255,
				A: float32(px[3]) / 255,
			})
		}
	}
	return img
}
