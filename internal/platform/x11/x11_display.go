package x11

/*
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
    "encoding/binary"
    "fmt"
    "os"
    "time"

    "agg_go/internal/buffer"
    "agg_go/internal/platform/types"
)

// copyBufferToXImage copies the AGG rendering buffer to XImage, handling pixel format conversion
func (x *X11Backend) copyBufferToXImage(buffer *buffer.RenderingBuffer[uint8]) error {
	if buffer == nil || buffer.Buf() == nil {
		return fmt.Errorf("invalid buffer")
	}

	bufWidth := buffer.Width()
	bufHeight := buffer.Height()
	bufStride := buffer.Stride()

	// Ensure buffer dimensions match window dimensions
	if bufWidth != x.width || bufHeight != x.height {
		return fmt.Errorf("buffer dimensions (%dx%d) don't match window (%dx%d)",
			bufWidth, bufHeight, x.width, x.height)
	}

	// Get source and destination pointers
	srcData := buffer.Buf()
	dstData := x.imgData

	// Perform pixel format conversion based on AGG format and X11 format
	switch x.format {
	case types.PixelFormatRGBA32:
		return x.copyRGBA32ToXImage(srcData, dstData, bufStride, x.imgStride)
	case types.PixelFormatBGRA32:
		return x.copyBGRA32ToXImage(srcData, dstData, bufStride, x.imgStride)
	case types.PixelFormatRGB24:
		return x.copyRGB24ToXImage(srcData, dstData, bufStride, x.imgStride)
	case types.PixelFormatBGR24:
		return x.copyBGR24ToXImage(srcData, dstData, bufStride, x.imgStride)
	case types.PixelFormatGray8:
		return x.copyGray8ToXImage(srcData, dstData, bufStride, x.imgStride)
	default:
		// For unsupported formats, do a raw copy
		return x.copyRawToXImage(srcData, dstData, bufStride, x.imgStride)
	}
}

// copyRGBA32ToXImage converts RGBA32 to X11 format
func (x *X11Backend) copyRGBA32ToXImage(src, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < x.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for px := 0; px < x.width; px++ {
			srcPixel := srcRow + px*4
			dstPixel := dstRow + px*4

			if srcPixel+3 < len(src) && dstPixel+3 < len(dst) {
				// RGBA -> BGRA (typical X11 format)
				dst[dstPixel+0] = src[srcPixel+2] // B
				dst[dstPixel+1] = src[srcPixel+1] // G
				dst[dstPixel+2] = src[srcPixel+0] // R
				dst[dstPixel+3] = src[srcPixel+3] // A
			}
		}
	}
	return nil
}

// copyBGRA32ToXImage converts BGRA32 to X11 format
func (x *X11Backend) copyBGRA32ToXImage(src, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < x.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for px := 0; px < x.width; px++ {
			srcPixel := srcRow + px*4
			dstPixel := dstRow + px*4

			if srcPixel+3 < len(src) && dstPixel+3 < len(dst) {
				// BGRA -> BGRA (direct copy for typical X11 format)
				dst[dstPixel+0] = src[srcPixel+0] // B
				dst[dstPixel+1] = src[srcPixel+1] // G
				dst[dstPixel+2] = src[srcPixel+2] // R
				dst[dstPixel+3] = src[srcPixel+3] // A
			}
		}
	}
	return nil
}

// copyRGB24ToXImage converts RGB24 to X11 format
func (x *X11Backend) copyRGB24ToXImage(src, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < x.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for px := 0; px < x.width; px++ {
			srcPixel := srcRow + px*3
			dstPixel := dstRow + px*4 // X11 typically uses 4 bytes even for RGB

			if srcPixel+2 < len(src) && dstPixel+3 < len(dst) {
				// RGB -> BGRX
				dst[dstPixel+0] = src[srcPixel+2] // B
				dst[dstPixel+1] = src[srcPixel+1] // G
				dst[dstPixel+2] = src[srcPixel+0] // R
				dst[dstPixel+3] = 255             // X (unused)
			}
		}
	}
	return nil
}

// copyBGR24ToXImage converts BGR24 to X11 format
func (x *X11Backend) copyBGR24ToXImage(src, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < x.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for px := 0; px < x.width; px++ {
			srcPixel := srcRow + px*3
			dstPixel := dstRow + px*4

			if srcPixel+2 < len(src) && dstPixel+3 < len(dst) {
				// BGR -> BGRX (direct copy)
				dst[dstPixel+0] = src[srcPixel+0] // B
				dst[dstPixel+1] = src[srcPixel+1] // G
				dst[dstPixel+2] = src[srcPixel+2] // R
				dst[dstPixel+3] = 255             // X (unused)
			}
		}
	}
	return nil
}

// copyGray8ToXImage converts Gray8 to X11 format
func (x *X11Backend) copyGray8ToXImage(src, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < x.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for px := 0; px < x.width; px++ {
			srcPixel := srcRow + px
			dstPixel := dstRow + px*4

			if srcPixel < len(src) && dstPixel+3 < len(dst) {
				gray := src[srcPixel]
				// Gray -> BGRX
				dst[dstPixel+0] = gray // B
				dst[dstPixel+1] = gray // G
				dst[dstPixel+2] = gray // R
				dst[dstPixel+3] = 255  // X (unused)
			}
		}
	}
	return nil
}

// copyRawToXImage does a raw memory copy (fallback for unsupported formats)
func (x *X11Backend) copyRawToXImage(src, dst []byte, srcStride, dstStride int) error {
	minStride := srcStride
	if dstStride < srcStride {
		minStride = dstStride
	}

	for y := 0; y < x.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		if srcRow+minStride <= len(src) && dstRow+minStride <= len(dst) {
			copy(dst[dstRow:dstRow+minStride], src[srcRow:srcRow+minStride])
		}
	}
	return nil
}

// loadBMP loads a 24-bit uncompressed BMP file.
func (x *X11Backend) loadBMP(filename string) (*X11ImageSurface, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read BMP file header
	fileHeader := make([]byte, 14)
	if _, err := file.Read(fileHeader); err != nil {
		return nil, fmt.Errorf("failed to read BMP file header: %w", err)
	}

	if string(fileHeader[0:2]) != "BM" {
		return nil, fmt.Errorf("not a BMP file")
	}

	// Read DIB header size
	var dibHeaderSize uint32
	if err := binary.Read(file, binary.LittleEndian, &dibHeaderSize); err != nil {
		return nil, fmt.Errorf("failed to read DIB header size: %w", err)
	}

	// Read the rest of the DIB header (we only support BITMAPINFOHEADER)
	if dibHeaderSize != 40 {
		return nil, fmt.Errorf("unsupported DIB header size: %d", dibHeaderSize)
	}

	dibHeader := make([]byte, 36) // 40 - 4 bytes for size
	if _, err := file.Read(dibHeader); err != nil {
		return nil, fmt.Errorf("failed to read DIB header: %w", err)
	}

	width := int(binary.LittleEndian.Uint32(dibHeader[0:4]))
	height := int(binary.LittleEndian.Uint32(dibHeader[4:8]))
	bpp := binary.LittleEndian.Uint16(dibHeader[10:12])
	compression := binary.LittleEndian.Uint32(dibHeader[12:16])

	if bpp != 24 || compression != 0 {
		return nil, fmt.Errorf("unsupported BMP format: bpp=%d, compression=%d", bpp, compression)
	}

	// Move to pixel data
	pixelDataOffset := binary.LittleEndian.Uint32(fileHeader[10:14])
	if _, err := file.Seek(int64(pixelDataOffset), 0); err != nil {
		return nil, fmt.Errorf("failed to seek to pixel data: %w", err)
	}

	// Read pixel data
	rowSize := (width*3 + 3) &^ 3 // 24-bit BMP rows are padded to 4 bytes
	pixelData := make([]byte, rowSize*height)
	if _, err := file.Read(pixelData); err != nil {
		return nil, fmt.Errorf("failed to read pixel data: %w", err)
	}

	// Create X11ImageSurface
	surface := &X11ImageSurface{
		width:  width,
		height: height,
		bpp:    24,
		stride: width * 3,
		data:   make([]byte, width*3*height),
	}

	// Copy pixel data, flipping vertically
	for y := 0; y < height; y++ {
		srcRow := (height - 1 - y) * rowSize
		dstRow := y * surface.stride
		copy(surface.data[dstRow:dstRow+surface.stride], pixelData[srcRow:srcRow+width*3])
	}

	return surface, nil
}

// saveBMP saves a 24-bit uncompressed BMP file.
func (x *X11Backend) saveBMP(surface *X11ImageSurface, filename string) error {
	if surface.bpp != 24 {
		return fmt.Errorf("only 24-bit BMP saving is supported, got %d bpp", surface.bpp)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	width := surface.width
	height := surface.height
	rowSize := (width*3 + 3) &^ 3 // 24-bit BMP rows are padded to 4 bytes
	imageSize := rowSize * height
	fileSize := 54 + imageSize // BMP header (14) + DIB header (40) + image data

	// BMP File Header (14 bytes)
	fileHeader := make([]byte, 14)
	fileHeader[0] = 'B'
	fileHeader[1] = 'M'
	binary.LittleEndian.PutUint32(fileHeader[2:6], uint32(fileSize)) // File size
	binary.LittleEndian.PutUint16(fileHeader[6:8], 0)                // Reserved
	binary.LittleEndian.PutUint16(fileHeader[8:10], 0)               // Reserved
	binary.LittleEndian.PutUint32(fileHeader[10:14], 54)             // Pixel data offset

	if _, err := file.Write(fileHeader); err != nil {
		return fmt.Errorf("failed to write BMP file header: %w", err)
	}

	// DIB Header (BITMAPINFOHEADER - 40 bytes)
	dibHeader := make([]byte, 40)
	binary.LittleEndian.PutUint32(dibHeader[0:4], 40)                  // DIB header size
	binary.LittleEndian.PutUint32(dibHeader[4:8], uint32(width))       // Width
	binary.LittleEndian.PutUint32(dibHeader[8:12], uint32(height))     // Height
	binary.LittleEndian.PutUint16(dibHeader[12:14], 1)                 // Planes
	binary.LittleEndian.PutUint16(dibHeader[14:16], 24)                // Bits per pixel
	binary.LittleEndian.PutUint32(dibHeader[16:20], 0)                 // Compression (none)
	binary.LittleEndian.PutUint32(dibHeader[20:24], uint32(imageSize)) // Image size
	binary.LittleEndian.PutUint32(dibHeader[24:28], 2835)              // X pixels per meter
	binary.LittleEndian.PutUint32(dibHeader[28:32], 2835)              // Y pixels per meter
	binary.LittleEndian.PutUint32(dibHeader[32:36], 0)                 // Colors used
	binary.LittleEndian.PutUint32(dibHeader[36:40], 0)                 // Important colors

	if _, err := file.Write(dibHeader); err != nil {
		return fmt.Errorf("failed to write DIB header: %w", err)
	}

	// Write pixel data (BMP stores rows bottom-to-top)
	rowBuffer := make([]byte, rowSize)
	for y := height - 1; y >= 0; y-- {
		srcRow := y * surface.stride

		// Copy pixels and handle padding
		for x := 0; x < width; x++ {
			srcPixel := srcRow + x*3
			dstPixel := x * 3

			if srcPixel+2 < len(surface.data) {
				// BMP expects BGR order, assuming surface data is RGB
				rowBuffer[dstPixel+0] = surface.data[srcPixel+2] // B
				rowBuffer[dstPixel+1] = surface.data[srcPixel+1] // G
				rowBuffer[dstPixel+2] = surface.data[srcPixel+0] // R
			}
		}

		// Zero-fill padding bytes
		for i := width * 3; i < rowSize; i++ {
			rowBuffer[i] = 0
		}

		if _, err := file.Write(rowBuffer); err != nil {
			return fmt.Errorf("failed to write pixel row %d: %w", y, err)
		}
	}

	return nil
}

// CreateImageSurface creates an X11 image surface
func (x *X11Backend) CreateImageSurface(width, height int) (types.ImageSurface, error) {
	// For X11, we create a simple byte buffer that can be converted to XImage
	bpp := x.bpp
	if bpp < 24 {
		bpp = 24 // Minimum for X11 color images
	}

	stride := width * bpp / 8
	size := stride * height

	surface := &X11ImageSurface{
		width:  width,
		height: height,
		bpp:    bpp,
		stride: stride,
		data:   make([]byte, size),
	}

	return surface, nil
}

// DestroyImageSurface destroys an X11 image surface
func (x *X11Backend) DestroyImageSurface(surface types.ImageSurface) error {
	// For our simple implementation, just let Go GC handle it
	return nil
}

// GetTicks returns the current tick count in milliseconds
func (x *X11Backend) GetTicks() uint32 {
	return uint32(time.Now().UnixNano()/1e6) - x.startTicks
}

// Delay provides a delay using time.Sleep
func (x *X11Backend) Delay(ms uint32) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// LoadImage loads an image file (basic BMP support)
func (x *X11Backend) LoadImage(filename string) (types.ImageSurface, error) {
	return x.loadBMP(filename)
}

// SaveImage saves an image to file (basic BMP support)
func (x *X11Backend) SaveImage(surface types.ImageSurface, filename string) error {
	if surface == nil || !surface.IsValid() {
		return fmt.Errorf("invalid surface")
	}

	x11Surface, ok := surface.(*X11ImageSurface)
	if !ok {
		return fmt.Errorf("surface is not an X11ImageSurface")
	}

	return x.saveBMP(x11Surface, filename)
}

// GetImageExtension returns the preferred image extension
func (x *X11Backend) GetImageExtension() string {
	return ".bmp"
}

// GetNativeHandle returns the native X11 window handle
func (x *X11Backend) GetNativeHandle() types.NativeHandle {
	return &X11NativeHandle{
		display: x.display,
		window:  x.window,
		gc:      x.gc,
	}
}

// X11ImageSurface represents an image surface in X11
type X11ImageSurface struct {
	width  int
	height int
	bpp    int
	stride int
	data   []byte
}

// Interface methods for ImageSurface
func (x *X11ImageSurface) GetWidth() int   { return x.width }
func (x *X11ImageSurface) GetHeight() int  { return x.height }
func (x *X11ImageSurface) GetData() []byte { return x.data }
func (x *X11ImageSurface) IsValid() bool   { return x.data != nil && x.width > 0 && x.height > 0 }

// X11NativeHandle provides access to native X11 handles
type X11NativeHandle struct {
	display *C.Display
	window  C.Window
	gc      C.GC
}

// Interface methods for NativeHandle
func (x *X11NativeHandle) GetType() string { return "X11" }
func (x *X11NativeHandle) IsValid() bool   { return x.display != nil && x.window != 0 }
