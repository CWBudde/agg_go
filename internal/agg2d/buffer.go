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
	agg2d.FlipText(false)
	agg2d.ImageFilter(ImageFilterBilinear)
	agg2d.ImageResample(NoResample)
	agg2d.masterAlpha = 1.0
	agg2d.antiAliasGamma = 1.0
	agg2d.blendMode = BlendAlpha

	// Initialize rendering pipeline
	agg2d.initializeRendering()
	agg2d.updateRasterizerGamma()
}

// initializeRendering sets up the rendering pipeline
func (agg2d *Agg2D) initializeRendering() {
	// Initialize pixel format with the attached buffer
	width := agg2d.rbuf.Width()
	height := agg2d.rbuf.Height()

	if width > 0 && height > 0 {
		// Create pixel format
		agg2d.pixfmt = pixfmt.NewPixFmtRGBA32[color.Linear](agg2d.rbuf)
		agg2d.pixfmtPre = pixfmt.NewPixFmtRGBA32Pre[color.Linear](agg2d.rbuf)
		agg2d.renBase = newBaseRendererAdapter[color.RGBA8[color.Linear]](agg2d.pixfmt)
		agg2d.renBasePre = newBaseRendererAdapter[color.RGBA8[color.Linear]](agg2d.pixfmtPre)

		// Create composite pixel format with default source-over blending
		agg2d.pixfmtComp = pixfmt.NewPixFmtCompositeRGBA32(agg2d.rbuf, blender.CompOpSrcOver)
		agg2d.pixfmtCompPre = pixfmt.NewPixFmtCompositeRGBA32Pre(agg2d.rbuf, blender.CompOpSrcOver)
		agg2d.renBaseComp = newBaseRendererAdapter[color.RGBA8[color.Linear]](agg2d.pixfmtComp)
		agg2d.renBaseCompPre = newBaseRendererAdapter[color.RGBA8[color.Linear]](agg2d.pixfmtCompPre)

		// Reapply current clip box to renderer adapters.
		agg2d.ClipBox(agg2d.clipBox.X1, agg2d.clipBox.Y1, agg2d.clipBox.X2, agg2d.clipBox.Y2)

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
	if agg2d.renBase == nil {
		return
	}

	clearColor := color.RGBA8[color.Linear]{R: c[0], G: c[1], B: c[2], A: c[3]}
	agg2d.renBase.Clear(clearColor)
}

// ClipBox sets the clipping rectangle.
func (agg2d *Agg2D) ClipBox(x1, y1, x2, y2 float64) {
	agg2d.clipBox.X1 = x1
	agg2d.clipBox.Y1 = y1
	agg2d.clipBox.X2 = x2
	agg2d.clipBox.Y2 = y2

	rx1, ry1 := int(x1), int(y1)
	rx2, ry2 := int(x2), int(y2)
	if agg2d.renBase != nil {
		agg2d.renBase.ClipBox(rx1, ry1, rx2, ry2)
	}
	if agg2d.renBasePre != nil {
		agg2d.renBasePre.ClipBox(rx1, ry1, rx2, ry2)
	}
	if agg2d.renBaseComp != nil {
		agg2d.renBaseComp.ClipBox(rx1, ry1, rx2, ry2)
	}
	if agg2d.renBaseCompPre != nil {
		agg2d.renBaseCompPre.ClipBox(rx1, ry1, rx2, ry2)
	}

	if agg2d.rasterizer != nil {
		agg2d.rasterizer.ClipBox(x1, y1, x2, y2)
	}
}

// WorldToScreen transforms world coordinates to screen coordinates.
func (agg2d *Agg2D) WorldToScreen(x, y *float64) {
	agg2d.transform.Transform(x, y)
}

// ScreenToWorld transforms screen coordinates to world coordinates.
func (agg2d *Agg2D) ScreenToWorld(x, y *float64) {
	agg2d.transform.InverseTransform(x, y)
}
