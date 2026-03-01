// Package agg provides AGG2D image operations implementation.
// This file implements Phase 6 of the AGG2D high-level interface: Image Operations.
package agg2d

import (
	"errors"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt/blender"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

type imageSampleGenerator interface {
	Generate(colors []color.RGBA8[color.Linear], x, y int)
}

type imageSamplePreparer interface {
	Prepare()
}

type imageSpanGenerator struct {
	sample      imageSampleGenerator
	blendMode   BlendMode
	blendColor  Color
	compBlender blender.CompositeBlender[color.Linear, order.RGBA]
	hasCompOp   bool
}

func newImageSpanGenerator(sample imageSampleGenerator, blendMode BlendMode, blendColor Color) *imageSpanGenerator {
	sg := &imageSpanGenerator{
		sample:     sample,
		blendMode:  blendMode,
		blendColor: blendColor,
	}
	if blendMode != BlendDst {
		sg.compBlender = blender.NewCompositeBlender[color.Linear, order.RGBA](blendModeToCompOp(blendMode))
		sg.hasCompOp = true
	}
	return sg
}

func (sg *imageSpanGenerator) Prepare() {
	if preparer, ok := sg.sample.(imageSamplePreparer); ok {
		preparer.Prepare()
	}
}

func (sg *imageSpanGenerator) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if sg.sample == nil || length <= 0 || len(colors) == 0 {
		return
	}
	if length < len(colors) {
		colors = colors[:length]
	}

	sg.sample.Generate(colors, x, y)

	if sg.hasCompOp {
		srcR, srcG, srcB := sg.blendColor[0], sg.blendColor[1], sg.blendColor[2]
		for i := range colors {
			pixel := []basics.Int8u{colors[i].R, colors[i].G, colors[i].B, colors[i].A}
			sg.compBlender.BlendPix(pixel, srcR, srcG, srcB, 255, basics.CoverFull)
			colors[i].R = pixel[0]
			colors[i].G = pixel[1]
			colors[i].B = pixel[2]
			colors[i].A = pixel[3]
		}
	}

	// AGG applies alpha from imageBlendColor after blend conversion.
	if sg.blendColor[3] != 255 {
		alpha := sg.blendColor[3]
		for i := range colors {
			colors[i].R = color.RGBA8Multiply(colors[i].R, alpha)
			colors[i].G = color.RGBA8Multiply(colors[i].G, alpha)
			colors[i].B = color.RGBA8Multiply(colors[i].B, alpha)
			colors[i].A = color.RGBA8Multiply(colors[i].A, alpha)
		}
	}
}

func (agg2d *Agg2D) newImageFilterGenerator(
	source *imagePixelFormat,
	interpolator *span.SpanInterpolatorLinear[*transform.TransAffine],
) imageSampleGenerator {
	if agg2d.imageFilter == NoFilter {
		return span.NewSpanImageFilterRGBANNWithParams[*imagePixelFormat, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interpolator)
	}

	resample := agg2d.imageResample == ResampleAlways
	if agg2d.imageResample == ResampleOnZoomOut && interpolator != nil {
		if tr := interpolator.Transformer(); tr != nil {
			sx, sy := tr.GetScalingAbs()
			if sx > 1.125 || sy > 1.125 {
				resample = true
			}
		}
	}

	if resample {
		return span.NewSpanImageResampleRGBAAffineWithParams[*imagePixelFormat](
			source,
			interpolator,
			agg2d.imageFilterLUT,
		)
	}

	if agg2d.imageFilter == Bilinear {
		return span.NewSpanImageFilterRGBABilinearWithParams[*imagePixelFormat, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interpolator)
	}

	if agg2d.imageFilterLUT == nil {
		return span.NewSpanImageFilterRGBABilinearWithParams[*imagePixelFormat, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interpolator)
	}

	if agg2d.imageFilterLUT.Diameter() == 2 {
		return span.NewSpanImageFilterRGBA2x2WithParams[*imagePixelFormat, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			interpolator,
			agg2d.imageFilterLUT,
		)
	}

	return span.NewSpanImageFilterRGBAWithParams[*imagePixelFormat, *span.SpanInterpolatorLinear[*transform.TransAffine]](
		source,
		interpolator,
		agg2d.imageFilterLUT,
	)
}

func (agg2d *Agg2D) setImagePathRect(x1, y1, x2, y2 float64) {
	agg2d.ResetPath()
	agg2d.MoveTo(x1, y1)
	agg2d.LineTo(x2, y1)
	agg2d.LineTo(x2, y2)
	agg2d.LineTo(x1, y2)
	agg2d.ClosePolygon()
}

func (agg2d *Agg2D) setImagePathParallelogram(parallelogram []float64) {
	agg2d.ResetPath()
	agg2d.MoveTo(parallelogram[0], parallelogram[1])
	agg2d.LineTo(parallelogram[2], parallelogram[3])
	agg2d.LineTo(parallelogram[4], parallelogram[5])
	agg2d.LineTo(
		parallelogram[0]+parallelogram[4]-parallelogram[2],
		parallelogram[1]+parallelogram[5]-parallelogram[3],
	)
	agg2d.ClosePolygon()
}

func (agg2d *Agg2D) addCurrentPathToRasterizer() {
	if agg2d.path == nil || agg2d.path.TotalVertices() == 0 || agg2d.rasterizer == nil {
		return
	}

	transformedPath := conv.NewConvTransform(agg2d.convCurve, agg2d.transform)
	transformedPath.Rewind(0)
	for {
		x, y, cmd := transformedPath.Vertex()
		if cmd == basics.PathCmdStop {
			return
		}
		agg2d.rasterizer.AddVertex(x, y, uint32(cmd))
	}
}

// renderImage renders the current path using AGG-style image span interpolation.
func (agg2d *Agg2D) renderImage(img *Image, x1, y1, x2, y2 int, parallelogram []float64) error {
	if img == nil || img.renBuf == nil {
		return errors.New("image or image buffer is nil")
	}
	if len(parallelogram) != 6 {
		return errors.New("parallelogram must have exactly 6 elements")
	}
	if agg2d.rasterizer == nil || agg2d.scanline == nil || agg2d.spanAllocator == nil {
		return errors.New("render pipeline is not initialized")
	}

	src := [6]float64{
		float64(x1), float64(y1),
		float64(x2), float64(y1),
		float64(x2), float64(y2),
	}
	dst := [6]float64{
		parallelogram[0], parallelogram[1],
		parallelogram[2], parallelogram[3],
		parallelogram[4], parallelogram[5],
	}

	mtx := transform.NewTransAffineParlToParl(src, dst)
	if agg2d.transform != nil {
		mtx.Multiply(agg2d.transform)
	}
	mtx.Invert()

	agg2d.rasterizer.Reset()
	agg2d.rasterizer.FillingRule(agg2d.GetFillRule())
	agg2d.addCurrentPathToRasterizer()

	interpolator := span.NewSpanInterpolatorLinearDefault(mtx)
	imageSource := newImagePixelFormat(img)
	sampleGenerator := agg2d.newImageFilterGenerator(imageSource, interpolator)
	spanGenerator := newImageSpanGenerator(sampleGenerator, agg2d.imageBlendMode, agg2d.imageBlendColor)

	var renderer *baseRendererAdapter[color.RGBA8[color.Linear]]
	if agg2d.blendMode == BlendAlpha {
		renderer = agg2d.renBase
	} else {
		renderer = agg2d.renBaseComp
	}
	if renderer == nil {
		return nil
	}

	rasAdapter := rasterizerAdapter{ras: agg2d.rasterizer}
	slAdapter := &scanlineWrapper{sl: agg2d.scanline}
	renscan.RenderScanlinesAA(rasAdapter, slAdapter, renderer, agg2d.spanAllocator, spanGenerator)

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

	agg2d.setImagePathRect(dstX1, dstY1, dstX2, dstY2)

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

	agg2d.setImagePathParallelogram(parallelogram)

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
// The image is clipped to the shape of the current path, matching the original AGG C++ behavior.
func (agg2d *Agg2D) TransformImagePath(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	// Create destination parallelogram from rectangle
	parallelogram := []float64{dstX1, dstY1, dstX2, dstY1, dstX2, dstY2}

	return agg2d.renderImage(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImagePathSimple transforms and renders entire image along current path to destination rectangle.
func (agg2d *Agg2D) TransformImagePathSimple(img *Image, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return agg2d.TransformImagePath(img, 0, 0, img.Width(), img.Height(), dstX1, dstY1, dstX2, dstY2)
}

// TransformImagePathParallelogram transforms and renders image along current path to destination parallelogram.
// The image is clipped to the shape of the current path, matching the original AGG C++ behavior.
func (agg2d *Agg2D) TransformImagePathParallelogram(img *Image, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	if len(parallelogram) < 6 {
		return errors.New("parallelogram requires 6 coordinates (3 points)")
	}

	return agg2d.renderImage(img, imgX1, imgY1, imgX2, imgY2, parallelogram)
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

	// Create image pixel format for the source
	imgPixFmt := newImagePixelFormat(img)

	// Create source rectangle
	srcRect := &basics.RectI{
		X1: imgX1,
		Y1: imgY1,
		X2: imgX2 - 1, // AGG uses inclusive coordinates
		Y2: imgY2 - 1,
	}

	// Use the rendering pipeline for blending.
	// AGG uses the general blend mode for blendImage/copyImage operations.
	var renderer *baseRendererAdapter[color.RGBA8[color.Linear]]
	if agg2d.blendMode == BlendAlpha {
		renderer = agg2d.renBase
	} else {
		renderer = agg2d.renBaseComp
	}

	if renderer != nil {
		renderer.BlendFrom(imgPixFmt, srcRect, dstXInt, dstYInt, basics.Int8u(alpha))
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

	// Create image pixel format for the source
	imgPixFmt := newImagePixelFormat(img)

	// Create source rectangle
	srcRect := &basics.RectI{
		X1: imgX1,
		Y1: imgY1,
		X2: imgX2 - 1, // AGG uses inclusive coordinates
		Y2: imgY2 - 1,
	}

	// Use the rendering pipeline for copying (always uses normal copying, regardless of imageBlendMode)
	// Note: CopyImage always does direct copying without blending, but we could apply imageBlendMode if needed
	if agg2d.renBase != nil {
		agg2d.renBase.CopyFrom(imgPixFmt, srcRect, dstXInt, dstYInt)
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
func (img *Image) PixelFormat() *imagePixelFormat {
	return newImagePixelFormat(img)
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
