// Package agg provides image operations for 2D graphics.
// This file contains image loading, transformation, and rendering functionality.
package agg

import (
	"errors"
	"image"
	_ "image/gif" // Import for gif decoding
	"image/jpeg"
	"image/png"
	"os"

	"agg_go/internal/agg2d"
	"agg_go/internal/buffer"
)

// ImageFilter represents different image filtering options
const (
	ImageFilterBilinear ImageFilter = iota
	ImageFilterHanning
	ImageFilterHermite
	ImageFilterQuadric
	ImageFilterBicubic
	ImageFilterCatrom
	ImageFilterSpline16
	ImageFilterSpline36
	ImageFilterBlackman144
)

// Aliases for backwards compatibility
const (
	NoFilter    = ImageFilterBilinear // Default filter
	Bilinear    = ImageFilterBilinear
	Hanning     = ImageFilterHanning
	Hermite     = ImageFilterHermite
	Quadric     = ImageFilterQuadric
	Bicubic     = ImageFilterBicubic
	Catrom      = ImageFilterCatrom
	Spline16    = ImageFilterSpline16
	Spline36    = ImageFilterSpline36
	Blackman144 = ImageFilterBlackman144
)

// ImageResample defines image resampling modes.
const (
	NoResample        ImageResample = iota // No resampling
	ResampleAlways                         // Always resample
	ResampleOnZoomOut                      // Resample only when zooming out
)

// Image represents a raster image that can be used as a rendering target.
// This matches the C++ Agg2D::Image structure.
type Image struct {
	renBuf *buffer.RenderingBuffer[uint8]
	Data   []uint8 // Raw pixel data (RGBA format)
	width  int     // Width in pixels
	height int     // Height in pixels
}

// NewImage creates a new image with the specified buffer.
func NewImage(buf []uint8, width, height, stride int) *Image {
	img := &Image{
		renBuf: buffer.NewRenderingBuffer[uint8](),
		Data:   buf,
		width:  width,
		height: height,
	}
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

// Attach attaches a buffer to the image.
func (img *Image) Attach(buf []uint8, width, height, stride int) {
	img.renBuf.Attach(buf, width, height, stride)
	img.Data = buf
	img.width = width
	img.height = height
}

// ToInternalImage converts this Image to the internal agg2d.Image type.
func (img *Image) ToInternalImage() *agg2d.Image {
	if img == nil {
		return nil
	}
	return agg2d.NewImage(img.Data, img.width, img.height, img.renBuf.Stride())
}

// ToGoImage converts the AGG image to a standard Go image.RGBA.
func (img *Image) ToGoImage() *image.RGBA {
	if img == nil {
		return nil
	}

	goImg := image.NewRGBA(image.Rect(0, 0, img.width, img.height))

	// Copy pixel data from AGG format (RGBA) to Go image format (RGBA)
	stride := img.renBuf.Stride()
	for y := 0; y < img.height; y++ {
		srcRow := y * stride
		dstRow := y * goImg.Stride
		for x := 0; x < img.width; x++ {
			srcIdx := srcRow + x*4
			dstIdx := dstRow + x*4

			// AGG format is RGBA, Go image is also RGBA, so direct copy
			if srcIdx+3 < len(img.Data) && dstIdx+3 < len(goImg.Pix) {
				goImg.Pix[dstIdx] = img.Data[srcIdx]     // R
				goImg.Pix[dstIdx+1] = img.Data[srcIdx+1] // G
				goImg.Pix[dstIdx+2] = img.Data[srcIdx+2] // B
				goImg.Pix[dstIdx+3] = img.Data[srcIdx+3] // A
			}
		}
	}

	return goImg
}

// Context image methods

// DrawImage draws an image at the specified coordinates.
func (ctx *Context) DrawImage(img *Image, x, y float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	return ctx.agg2d.TransformImageSimple(img, x, y, x+float64(img.Width()), y+float64(img.Height()))
}

// DrawImageScaled draws an image scaled to the specified width and height.
func (ctx *Context) DrawImageScaled(img *Image, x, y, width, height float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	return ctx.agg2d.TransformImageSimple(img, x, y, x+width, y+height)
}

// DrawImageTransformed draws an image with a transformation matrix.
func (ctx *Context) DrawImageTransformed(img *Image, transform *Transformations) error {
	if img == nil {
		return errors.New("image is nil")
	}
	if transform == nil {
		return ctx.DrawImage(img, 0, 0)
	}

	// Transform the four corners of the image
	w, h := float64(img.Width()), float64(img.Height())
	corners := [][2]float64{
		{0, 0}, {w, 0}, {w, h}, {0, h},
	}

	for i, corner := range corners {
		x, y := transform.Transform(corner[0], corner[1])
		corners[i] = [2]float64{x, y}
	}

	// Create parallelogram from first three corners
	parallelogram := []float64{
		corners[0][0], corners[0][1], // First corner
		corners[1][0], corners[1][1], // Second corner
		corners[3][0], corners[3][1], // Fourth corner (opposite of second)
	}

	return ctx.agg2d.TransformImageParallelogram(img, 0, 0, img.Width(), img.Height(), parallelogram)
}

// DrawImageRegion draws a region of an image to the specified destination.
func (ctx *Context) DrawImageRegion(img *Image, srcX, srcY, srcW, srcH int, dstX, dstY, dstW, dstH float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return ctx.agg2d.TransformImage(img, srcX, srcY, srcX+srcW, srcY+srcH, dstX, dstY, dstX+dstW, dstY+dstH)
}

// Image loading functions

// LoadImageFromFile loads an image from a file.
func LoadImageFromFile(filename string) (*Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return NewImageFromStandardImage(img)
}

// NewImageFromStandardImage creates an AGG Image from a standard Go image.
func NewImageFromStandardImage(img image.Image) (*Image, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	stride := width * 4 // RGBA format

	// Create buffer and copy image data
	buffer := make([]uint8, height*stride)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

			// Convert from 16-bit to 8-bit
			index := y*stride + x*4
			buffer[index] = uint8(r >> 8)   // R
			buffer[index+1] = uint8(g >> 8) // G
			buffer[index+2] = uint8(b >> 8) // B
			buffer[index+3] = uint8(a >> 8) // A
		}
	}

	return NewImage(buffer, width, height, stride), nil
}

// SaveImageToPNG saves an AGG Image to a PNG file.
func (img *Image) SaveToPNG(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	stdImg, err := img.ToStandardImage()
	if err != nil {
		return err
	}

	return png.Encode(file, stdImg)
}

// SaveImageToJPEG saves an AGG Image to a JPEG file.
func (img *Image) SaveToJPEG(filename string, quality int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	stdImg, err := img.ToStandardImage()
	if err != nil {
		return err
	}

	options := &jpeg.Options{Quality: quality}
	return jpeg.Encode(file, stdImg, options)
}

// ToStandardImage converts an AGG Image to a standard Go image.
func (img *Image) ToStandardImage() (image.Image, error) {
	if img == nil || img.renBuf == nil {
		return nil, errors.New("image or buffer is nil")
	}

	width := img.Width()
	height := img.Height()
	stride := img.renBuf.Stride()

	bounds := image.Rect(0, 0, width, height)
	stdImg := image.NewRGBA(bounds)

	// Copy pixel data
	buffer := img.renBuf.Buf()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcIndex := y*stride + x*4
			dstIndex := y*stdImg.Stride + x*4

			if srcIndex+3 < len(buffer) {
				stdImg.Pix[dstIndex] = buffer[srcIndex]     // R
				stdImg.Pix[dstIndex+1] = buffer[srcIndex+1] // G
				stdImg.Pix[dstIndex+2] = buffer[srcIndex+2] // B
				stdImg.Pix[dstIndex+3] = buffer[srcIndex+3] // A
			}
		}
	}

	return stdImg, nil
}

// Image filtering methods

// SetImageFilter sets the image filtering method for subsequent image operations.
func (ctx *Context) SetImageFilter(filter ImageFilter) {
	ctx.agg2d.ImageFilter(filter)
}

// GetImageFilter returns the current image filtering method.
func (ctx *Context) GetImageFilter() ImageFilter {
	return ctx.agg2d.GetImageFilter()
}

// Image resampling methods

// SetImageResample sets the image resampling method.
func (ctx *Context) SetImageResample(resample ImageResample) {
	ctx.agg2d.ImageResample(resample)
}

// GetImageResample returns the current image resampling method.
func (ctx *Context) GetImageResample() ImageResample {
	return ctx.agg2d.GetImageResample()
}

// Advanced image operations

// DrawImageRotated draws an image rotated by the specified angle (in radians).
func (ctx *Context) DrawImageRotated(img *Image, x, y, angle float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	// Create rotation transformation
	transform := Rotation(angle)
	transform.AffineMatrix[4] = x // Set translation
	transform.AffineMatrix[5] = y

	return ctx.DrawImageTransformed(img, transform)
}

// DrawImageRotatedDegrees draws an image rotated by the specified angle (in degrees).
func (ctx *Context) DrawImageRotatedDegrees(img *Image, x, y, degrees float64) error {
	return ctx.DrawImageRotated(img, x, y, degrees*3.14159265359/180.0)
}

// DrawImageSkewed draws an image with skewing transformation.
func (ctx *Context) DrawImageSkewed(img *Image, x, y, skewX, skewY float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	// Create skewing transformation
	transform := Skewing(skewX, skewY)
	transform.AffineMatrix[4] = x // Set translation
	transform.AffineMatrix[5] = y

	return ctx.DrawImageTransformed(img, transform)
}

// CreateImagePattern creates a repeating pattern from an image.
func (ctx *Context) CreateImagePattern(img *Image, x, y, width, height float64, repeatX, repeatY bool) error {
	if img == nil {
		return errors.New("image is nil")
	}

	// For now, just draw the image once
	// TODO: Implement proper pattern support when span generators are ready
	return ctx.DrawImageScaled(img, x, y, width, height)
}

// Image creation utilities

// CreateImage creates a new blank image with the specified dimensions.
func CreateImage(width, height int) *Image {
	stride := width * 4 // RGBA format
	buffer := make([]uint8, height*stride)
	return NewImage(buffer, width, height, stride)
}

// CreateImageFromColor creates a new image filled with a single color.
func CreateImageFromColor(width, height int, color Color) *Image {
	img := CreateImage(width, height)

	// Fill with color
	stride := img.renBuf.Stride()
	buffer := img.renBuf.Buf()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := y*stride + x*4
			buffer[index] = color.R   // R
			buffer[index+1] = color.G // G
			buffer[index+2] = color.B // B
			buffer[index+3] = color.A // A
		}
	}

	return img
}

// CloneImage creates a copy of an existing image.
func CloneImage(src *Image) (*Image, error) {
	if src == nil {
		return nil, errors.New("source image is nil")
	}

	width := src.Width()
	height := src.Height()
	stride := src.renBuf.Stride()

	// Create new buffer and copy data
	srcBuffer := src.renBuf.Buf()
	dstBuffer := make([]uint8, len(srcBuffer))
	copy(dstBuffer, srcBuffer)

	return NewImage(dstBuffer, width, height, stride), nil
}
