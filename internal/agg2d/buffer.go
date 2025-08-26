// Package agg2d buffer management for AGG2D high-level interface.
// This file contains buffer attachment and management methods.
package agg2d

import (
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
)

// Image represents a raster image that can be used as a rendering target.
// This matches the C++ Agg2D::Image structure.
type Image struct {
	renBuf *buffer.RenderingBuffer[uint8]
	Data   []uint8 // Raw pixel data (RGBA format)
	width  int     // Width in pixels
	height int     // Height in pixels
}

// NewImage creates a new Image with the given buffer, dimensions, and stride.
func NewImage(buf []uint8, width, height, stride int) *Image {
	img := &Image{
		Data:   buf,
		width:  width,
		height: height,
	}

	// Create a rendering buffer from the image data
	img.renBuf = buffer.NewRenderingBuffer[uint8]()
	img.renBuf.Attach(buf, width, height, stride)

	return img
}

// Width returns the image width.
func (img *Image) Width() int {
	return img.width
}

// Height returns the image height.
func (img *Image) Height() int {
	return img.height
}

// Attach attaches a rendering buffer to the AGG2D context.
// This matches the C++ Agg2D::attach method.
func (agg2d *Agg2D) Attach(buf []uint8, width, height, stride int) {
	agg2d.rbuf.Attach(buf, width, height, stride)

	// Reset clipping and transformations
	agg2d.ResetTransformations()
	agg2d.LineWidth(1.0)
	agg2d.LineColor(Black)
	agg2d.FillColor(White)
	agg2d.TextAlignment(AlignLeft, AlignBottom)
	agg2d.ClipBox(0, 0, float64(width), float64(height))
	agg2d.LineCap(CapRound)
	agg2d.LineJoin(JoinRound)
	agg2d.ImageFilter(ImageFilterBilinear)
	agg2d.ImageResample(NoResample)
	agg2d.masterAlpha = 1.0
	agg2d.antiAliasGamma = 1.0
	agg2d.blendMode = BlendAlpha

	// Initialize rendering pipeline
	agg2d.initializeRendering()
}

// initializeRendering sets up the rendering pipeline
func (agg2d *Agg2D) initializeRendering() {
	// Initialize pixel format with the attached buffer
	width := agg2d.rbuf.Width()
	height := agg2d.rbuf.Height()

	if width > 0 && height > 0 {
		// Create pixel format
		agg2d.pixfmt = pixfmt.NewPixFmtRGBA32(agg2d.rbuf)
		agg2d.renBase = &baseRendererAdapter[color.RGBA8[color.Linear]]{pf: agg2d.pixfmt}

		// Create composite pixel format with default source-over blending
		agg2d.pixfmtComp = pixfmt.NewPixFmtCompositeRGBA32(agg2d.rbuf, blender.CompOpSrcOver)
		agg2d.renBaseComp = &baseRendererAdapter[color.RGBA8[color.Linear]]{pf: agg2d.pixfmtComp}

		// Initialize rasterizer if needed
		// Note: The rasterizer is already created in NewAgg2D with the correct types
		if agg2d.rasterizer != nil {
			// Reset rasterizer clip box
			agg2d.rasterizer.Reset()
			agg2d.rasterizer.ClipBox(0, 0, float64(width), float64(height))
		}
	}
}

// ClearAll fills the entire buffer with the specified color.
func (agg2d *Agg2D) ClearAll(c Color) {
	// Simple implementation - fill the entire buffer
	buf := agg2d.rbuf.Buf()
	width := agg2d.rbuf.Width()
	height := agg2d.rbuf.Height()
	stride := agg2d.rbuf.Stride()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := y*stride + x*4
			if offset+3 < len(buf) {
				buf[offset] = c[0]   // R
				buf[offset+1] = c[1] // G
				buf[offset+2] = c[2] // B
				buf[offset+3] = c[3] // A
			}
		}
	}
}

// ClipBox sets the clipping rectangle.
func (agg2d *Agg2D) ClipBox(x1, y1, x2, y2 float64) {
	agg2d.clipBox.X1 = x1
	agg2d.clipBox.Y1 = y1
	agg2d.clipBox.X2 = x2
	agg2d.clipBox.Y2 = y2

	// if agg2d.rasterizer != nil {
	//     agg2d.rasterizer.ClipBox(x1, y1, x2, y2)
	// }
}

// WorldToScreen transforms world coordinates to screen coordinates.
func (agg2d *Agg2D) WorldToScreen(x, y *float64) {
	agg2d.transform.Transform(x, y)
}

// ScreenToWorld transforms screen coordinates to world coordinates.
func (agg2d *Agg2D) ScreenToWorld(x, y *float64) {
	agg2d.transform.InverseTransform(x, y)
}
