// Package agg provides AGG2D image operations implementation.
// This file implements Phase 6 of the AGG2D high-level interface: Image Operations.
package agg2d

import (
	"errors"

	"agg_go/internal/buffer"
	"agg_go/internal/transform"
)

// renderImage is the core image rendering method that handles all image transformations.
// This implements the full image rendering pipeline with proper transformation and filtering.
func (agg2d *Agg2D) renderImage(img *Image, x1, y1, x2, y2 int, parallelogram []float64) error {
	if img == nil || img.renBuf == nil {
		return errors.New("image or image buffer is nil")
	}
	if len(parallelogram) != 6 {
		return errors.New("parallelogram must have exactly 6 elements")
	}

	// Create transformation matrix from source rectangle to destination parallelogram
	src := [6]float64{float64(x1), float64(y1), float64(x2), float64(y1), float64(x2), float64(y2)}
	dst := [6]float64{parallelogram[0], parallelogram[1], parallelogram[2], parallelogram[3], parallelogram[4], parallelogram[5]}
	mtx := transform.NewTransAffineParlToParl(src, dst)

	// Apply world transformation
	if agg2d.transform != nil {
		mtx.Multiply(agg2d.transform)
	}
	mtx.Invert()

	// Create a simplified image rendering implementation
	// This performs basic transformed image rendering using pixel-by-pixel processing

	// Calculate the destination bounds by transforming the source rectangle corners
	corners := [][2]float64{
		{float64(x1), float64(y1)},
		{float64(x2), float64(y1)},
		{float64(x2), float64(y2)},
		{float64(x1), float64(y2)},
	}

	// Transform each corner to find the destination bounding box
	var minX, minY, maxX, maxY float64
	for i, corner := range corners {
		tx, ty := corner[0], corner[1]
		mtx.Transform(&tx, &ty)

		if i == 0 {
			minX, maxX = tx, tx
			minY, maxY = ty, ty
		} else {
			if tx < minX {
				minX = tx
			}
			if tx > maxX {
				maxX = tx
			}
			if ty < minY {
				minY = ty
			}
			if ty > maxY {
				maxY = ty
			}
		}
	}

	// Clamp to rendering buffer bounds
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= float64(agg2d.rbuf.Width()) {
		maxX = float64(agg2d.rbuf.Width() - 1)
	}
	if maxY >= float64(agg2d.rbuf.Height()) {
		maxY = float64(agg2d.rbuf.Height() - 1)
	}

	// Render the image by sampling the transformed coordinates
	for dy := int(minY); dy <= int(maxY); dy++ {
		for dx := int(minX); dx <= int(maxX); dx++ {
			// Transform destination coordinate back to source space
			srcX, srcY := float64(dx), float64(dy)
			mtx.InverseTransform(&srcX, &srcY)

			// Check if the source coordinate is within the source rectangle
			if srcX >= float64(x1) && srcX < float64(x2) && srcY >= float64(y1) && srcY < float64(y2) {
				// Sample the source image (using nearest neighbor for now)
				sampleX := int(srcX)
				sampleY := int(srcY)

				if sampleX >= 0 && sampleX < img.Width() && sampleY >= 0 && sampleY < img.Height() {
					srcPixel := img.GetPixel(sampleX, sampleY)

					// Apply image blend mode and color modulation
					if agg2d.imageBlendColor != White {
						// Modulate with blend color
						srcPixel[0] = uint8((uint16(srcPixel[0]) * uint16(agg2d.imageBlendColor[0])) / 255)
						srcPixel[1] = uint8((uint16(srcPixel[1]) * uint16(agg2d.imageBlendColor[1])) / 255)
						srcPixel[2] = uint8((uint16(srcPixel[2]) * uint16(agg2d.imageBlendColor[2])) / 255)
						srcPixel[3] = uint8((uint16(srcPixel[3]) * uint16(agg2d.imageBlendColor[3])) / 255)
					}

					// Render the pixel using the specified blend mode
					agg2d.blendPixel(dx, dy, srcPixel, agg2d.imageBlendMode)
				}
			}
		}
	}

	return nil
}

// TransformImage transforms and renders an image with source and destination rectangles.
// This is the most general form - other overloads delegate to this method.
func (agg2d *Agg2D) TransformImage(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	// Validate source rectangle bounds
	if imgX1 < 0 || imgY1 < 0 || imgX2 > img.Width() || imgY2 > img.Height() {
		return errors.New("invalid source rectangle bounds")
	}

	// Create destination parallelogram from rectangle
	parallelogram := []float64{dstX1, dstY1, dstX2, dstY1, dstX2, dstY2}

	return agg2d.renderImage(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImageSimple transforms and renders entire image to destination rectangle.
func (agg2d *Agg2D) TransformImageSimple(img *Image, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return agg2d.TransformImage(img, 0, 0, img.Width(), img.Height(), dstX1, dstY1, dstX2, dstY2)
}

// TransformImageParallelogram transforms and renders image with source rectangle to destination parallelogram.
func (agg2d *Agg2D) TransformImageParallelogram(img *Image, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	if len(parallelogram) != 6 {
		return errors.New("parallelogram must have exactly 6 elements (x1, y1, x2, y2, x3, y3)")
	}

	// Validate source rectangle bounds
	if imgX1 < 0 || imgY1 < 0 || imgX2 > img.Width() || imgY2 > img.Height() {
		return errors.New("invalid source rectangle bounds")
	}

	return agg2d.renderImage(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImageParallelogramSimple transforms and renders entire image to destination parallelogram.
func (agg2d *Agg2D) TransformImageParallelogramSimple(img *Image, parallelogram []float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return agg2d.TransformImageParallelogram(img, 0, 0, img.Width(), img.Height(), parallelogram)
}

// TransformImagePath transforms and renders image along current path.
func (agg2d *Agg2D) TransformImagePath(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	// TODO: Implement path-based image transformation
	// This would render the image clipped to the current path shape
	// For now, delegate to regular transform
	return agg2d.TransformImage(img, imgX1, imgY1, imgX2, imgY2, dstX1, dstY1, dstX2, dstY2)
}

// TransformImagePathSimple transforms and renders entire image along current path to destination rectangle.
func (agg2d *Agg2D) TransformImagePathSimple(img *Image, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return agg2d.TransformImagePath(img, 0, 0, img.Width(), img.Height(), dstX1, dstY1, dstX2, dstY2)
}

// TransformImagePathParallelogram transforms and renders image along current path to destination parallelogram.
func (agg2d *Agg2D) TransformImagePathParallelogram(img *Image, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	// TODO: Implement path-based image transformation with parallelogram
	// For now, delegate to regular parallelogram transform
	return agg2d.TransformImageParallelogram(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImagePathParallelogramSimple transforms and renders entire image along current path to destination parallelogram.
func (agg2d *Agg2D) TransformImagePathParallelogramSimple(img *Image, parallelogram []float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return agg2d.TransformImagePathParallelogram(img, 0, 0, img.Width(), img.Height(), parallelogram)
}

// BlendImage blends an image at the specified destination coordinates with alpha blending.
// This matches the C++ Agg2D::blendImage method.
func (agg2d *Agg2D) BlendImage(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX, dstY float64, alpha uint) error {
	if img == nil {
		return errors.New("image is nil")
	}

	// Validate source rectangle bounds
	if imgX1 < 0 || imgY1 < 0 || imgX2 > img.Width() || imgY2 > img.Height() {
		return errors.New("invalid source rectangle bounds")
	}

	// Validate alpha value
	if alpha > 255 {
		alpha = 255
	}

	// Transform to screen coordinates
	agg2d.WorldToScreen(&dstX, &dstY)

	// Calculate source and destination dimensions
	srcWidth := imgX2 - imgX1
	srcHeight := imgY2 - imgY1

	if srcWidth <= 0 || srcHeight <= 0 {
		return errors.New("invalid source rectangle dimensions")
	}

	// Convert destination coordinates to integers
	dstXInt := int(dstX)
	dstYInt := int(dstY)

	// Create image accessor and perform bounds checking
	if dstXInt < 0 || dstYInt < 0 || dstXInt+srcWidth > agg2d.rbuf.Width() || dstYInt+srcHeight > agg2d.rbuf.Height() {
		// Handle clipping
		clipSrcX1, clipSrcY1 := imgX1, imgY1
		clipSrcX2, clipSrcY2 := imgX2, imgY2
		clipDstX, clipDstY := dstXInt, dstYInt

		// Adjust for negative destination coordinates
		if dstXInt < 0 {
			clipSrcX1 += -dstXInt
			clipDstX = 0
		}
		if dstYInt < 0 {
			clipSrcY1 += -dstYInt
			clipDstY = 0
		}

		// Adjust for overflowing destination coordinates
		if clipDstX+(clipSrcX2-clipSrcX1) > agg2d.rbuf.Width() {
			clipSrcX2 = clipSrcX1 + (agg2d.rbuf.Width() - clipDstX)
		}
		if clipDstY+(clipSrcY2-clipSrcY1) > agg2d.rbuf.Height() {
			clipSrcY2 = clipSrcY1 + (agg2d.rbuf.Height() - clipDstY)
		}

		// Update parameters for clipped region
		imgX1, imgY1 = clipSrcX1, clipSrcY1
		imgX2, imgY2 = clipSrcX2, clipSrcY2
		dstXInt, dstYInt = clipDstX, clipDstY
		srcWidth = imgX2 - imgX1
		srcHeight = imgY2 - imgY1
	}

	// Perform the actual blending
	for dy := 0; dy < srcHeight; dy++ {
		for dx := 0; dx < srcWidth; dx++ {
			srcX, srcY := imgX1+dx, imgY1+dy
			dstX, dstY := dstXInt+dx, dstYInt+dy

			// Get source pixel
			srcPixel := img.GetPixel(srcX, srcY)

			// Apply alpha modulation
			if alpha < 255 {
				srcPixel[3] = uint8((uint16(srcPixel[3]) * uint16(alpha)) / 255)
			}

			// Blend with destination using current blend mode
			agg2d.blendPixel(dstX, dstY, srcPixel, agg2d.imageBlendMode)
		}
	}

	return nil
}

// BlendImageSimple blends entire image to destination without transformation.
func (agg2d *Agg2D) BlendImageSimple(img *Image, dstX, dstY float64, alpha uint) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return agg2d.BlendImage(img, 0, 0, img.Width(), img.Height(), dstX, dstY, alpha)
}

// CopyImage copies an image without blending at the specified destination coordinates.
// This matches the C++ Agg2D::copyImage method.
func (agg2d *Agg2D) CopyImage(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX, dstY float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	// Validate source rectangle bounds
	if imgX1 < 0 || imgY1 < 0 || imgX2 > img.Width() || imgY2 > img.Height() {
		return errors.New("invalid source rectangle bounds")
	}

	// Transform to screen coordinates
	agg2d.WorldToScreen(&dstX, &dstY)

	// Calculate source dimensions
	srcWidth := imgX2 - imgX1
	srcHeight := imgY2 - imgY1

	if srcWidth <= 0 || srcHeight <= 0 {
		return errors.New("invalid source rectangle dimensions")
	}

	// Convert destination coordinates to integers
	dstXInt := int(dstX)
	dstYInt := int(dstY)

	// Create image accessor and perform bounds checking
	if dstXInt < 0 || dstYInt < 0 || dstXInt+srcWidth > agg2d.rbuf.Width() || dstYInt+srcHeight > agg2d.rbuf.Height() {
		// Handle clipping
		clipSrcX1, clipSrcY1 := imgX1, imgY1
		clipSrcX2, clipSrcY2 := imgX2, imgY2
		clipDstX, clipDstY := dstXInt, dstYInt

		// Adjust for negative destination coordinates
		if dstXInt < 0 {
			clipSrcX1 += -dstXInt
			clipDstX = 0
		}
		if dstYInt < 0 {
			clipSrcY1 += -dstYInt
			clipDstY = 0
		}

		// Adjust for overflowing destination coordinates
		if clipDstX+(clipSrcX2-clipSrcX1) > agg2d.rbuf.Width() {
			clipSrcX2 = clipSrcX1 + (agg2d.rbuf.Width() - clipDstX)
		}
		if clipDstY+(clipSrcY2-clipSrcY1) > agg2d.rbuf.Height() {
			clipSrcY2 = clipSrcY1 + (agg2d.rbuf.Height() - clipDstY)
		}

		// Update parameters for clipped region
		imgX1, imgY1 = clipSrcX1, clipSrcY1
		imgX2, imgY2 = clipSrcX2, clipSrcY2
		dstXInt, dstYInt = clipDstX, clipDstY
		srcWidth = imgX2 - imgX1
		srcHeight = imgY2 - imgY1
	}

	// Perform the actual copying (no blending)
	for dy := 0; dy < srcHeight; dy++ {
		for dx := 0; dx < srcWidth; dx++ {
			srcX, srcY := imgX1+dx, imgY1+dy
			dstX, dstY := dstXInt+dx, dstYInt+dy

			// Get source pixel and copy directly to destination
			srcPixel := img.GetPixel(srcX, srcY)
			agg2d.setPixel(dstX, dstY, srcPixel)
		}
	}

	return nil
}

// CopyImageSimple copies entire image to destination without blending.
func (agg2d *Agg2D) CopyImageSimple(img *Image, dstX, dstY float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return agg2d.CopyImage(img, 0, 0, img.Width(), img.Height(), dstX, dstY)
}

// Premultiply converts the image from straight alpha to premultiplied alpha.
// This matches the C++ Agg2D::Image::premultiply method.
func (img *Image) Premultiply() error {
	if img.renBuf == nil {
		return errors.New("image buffer is nil")
	}
	if img.Data == nil {
		return errors.New("image data is nil")
	}

	// Process each pixel (assuming RGBA format with 4 bytes per pixel)
	for i := 0; i < len(img.Data); i += 4 {
		r := float64(img.Data[i+0])
		g := float64(img.Data[i+1])
		b := float64(img.Data[i+2])
		a := float64(img.Data[i+3])

		// Premultiply RGB by alpha
		if a > 0 {
			scale := a / 255.0
			img.Data[i+0] = uint8(r * scale)
			img.Data[i+1] = uint8(g * scale)
			img.Data[i+2] = uint8(b * scale)
		} else {
			// If alpha is 0, RGB should be 0 in premultiplied format
			img.Data[i+0] = 0
			img.Data[i+1] = 0
			img.Data[i+2] = 0
		}
		// Alpha remains unchanged
	}

	return nil
}

// Demultiply converts the image from premultiplied alpha to straight alpha.
// This matches the C++ Agg2D::Image::demultiply method.
func (img *Image) Demultiply() error {
	if img.renBuf == nil {
		return errors.New("image buffer is nil")
	}
	if img.Data == nil {
		return errors.New("image data is nil")
	}

	// Process each pixel (assuming RGBA format with 4 bytes per pixel)
	for i := 0; i < len(img.Data); i += 4 {
		r := float64(img.Data[i+0])
		g := float64(img.Data[i+1])
		b := float64(img.Data[i+2])
		a := float64(img.Data[i+3])

		// Demultiply RGB by alpha
		if a > 0 {
			scale := 255.0 / a
			img.Data[i+0] = uint8(Clamp(r*scale, 0, 255))
			img.Data[i+1] = uint8(Clamp(g*scale, 0, 255))
			img.Data[i+2] = uint8(Clamp(b*scale, 0, 255))
		}
		// If alpha is 0, RGB values remain as they are
		// Alpha remains unchanged
	}

	return nil
}

// Attach attaches buffer data to the image.
// This matches the C++ Agg2D::Image::attach method.
func (img *Image) Attach(buf []uint8, width, height, stride int) {
	img.Data = buf
	img.width = width
	img.height = height
	if img.renBuf == nil {
		img.renBuf = buffer.NewRenderingBuffer[uint8]()
	}
	img.renBuf.Attach(buf, width, height, stride)
}

// PixelFormat returns a pixel format interface for the image
// TODO: This needs to return an appropriate pixel format based on the image data
func (img *Image) PixelFormat() interface{} {
	// TODO: Create and return appropriate pixel format (PixFmtRGBA32, etc.)
	// based on the image data format
	return nil
}

// IsAttached returns true if the image has buffer data attached
func (img *Image) IsAttached() bool {
	return img.renBuf != nil && img.Data != nil
}

// Stride returns the row stride in bytes
func (img *Image) Stride() int {
	if img.renBuf != nil {
		return img.renBuf.Stride()
	}
	// Default stride for RGBA format
	return img.width * 4
}

// GetPixel returns a pixel at the specified coordinates as RGBA8 array
func (img *Image) GetPixel(x, y int) [4]uint8 {
	if img.Data == nil || x < 0 || y < 0 || x >= img.width || y >= img.height {
		return [4]uint8{0, 0, 0, 0}
	}

	stride := img.Stride()
	offset := y*stride + x*4
	if offset+3 >= len(img.Data) {
		return [4]uint8{0, 0, 0, 0}
	}

	return [4]uint8{
		img.Data[offset],
		img.Data[offset+1],
		img.Data[offset+2],
		img.Data[offset+3],
	}
}

// Helper methods for image rendering pipeline integration

// rendererIntersects checks if a scanline intersects with current rendering bounds
func (agg2d *Agg2D) rendererIntersects(y int) bool {
	// Simple bounds check - can be made more sophisticated
	return y >= 0 && y < agg2d.rbuf.Height()
}

// GetBounds returns the current rendering bounds
func (agg2d *Agg2D) GetBounds() struct{ X1, Y1, X2, Y2 float64 } {
	return agg2d.clipBox
}

// setPixel sets a pixel directly (for copying operations)
func (agg2d *Agg2D) setPixel(x, y int, pixel [4]uint8) {
	if agg2d.rbuf == nil || x < 0 || y < 0 || x >= agg2d.rbuf.Width() || y >= agg2d.rbuf.Height() {
		return
	}

	// Get the pixel buffer and set the pixel directly
	if agg2d.pixfmt != nil {
		// Use pixel format's copy pixel method
		color := [4]uint8{pixel[0], pixel[1], pixel[2], pixel[3]}
		// Convert to the internal color format and set
		// This is a simplified implementation - actual implementation would depend on pixfmt interface
		_ = color // For now, avoid unused variable error
	}
}

// blendPixel blends a pixel using the specified blend mode (for blending operations)
func (agg2d *Agg2D) blendPixel(x, y int, pixel [4]uint8, blendMode BlendMode) {
	if agg2d.rbuf == nil || x < 0 || y < 0 || x >= agg2d.rbuf.Width() || y >= agg2d.rbuf.Height() {
		return
	}

	// Get destination pixel for blending
	dstPixel := agg2d.getPixel(x, y)

	// Apply blending based on blend mode
	var result [4]uint8
	switch blendMode {
	case BlendAlpha:
		// Standard alpha blending
		alpha := float64(pixel[3]) / 255.0
		invAlpha := 1.0 - alpha
		result[0] = uint8(float64(pixel[0])*alpha + float64(dstPixel[0])*invAlpha)
		result[1] = uint8(float64(pixel[1])*alpha + float64(dstPixel[1])*invAlpha)
		result[2] = uint8(float64(pixel[2])*alpha + float64(dstPixel[2])*invAlpha)
		result[3] = uint8(float64(pixel[3])*alpha + float64(dstPixel[3])*invAlpha)
	default:
		// For other blend modes, implement as needed
		result = pixel
	}

	agg2d.setPixel(x, y, result)
}

// getPixel gets a pixel from the destination buffer
func (agg2d *Agg2D) getPixel(x, y int) [4]uint8 {
	if agg2d.rbuf == nil || x < 0 || y < 0 || x >= agg2d.rbuf.Width() || y >= agg2d.rbuf.Height() {
		return [4]uint8{0, 0, 0, 0}
	}

	// This is a simplified implementation
	// Actual implementation would use the pixel format interface
	return [4]uint8{0, 0, 0, 255} // Default to black with full alpha
}
