// Package image provides image accessor implementations for AGG.
// This package handles various pixel access patterns and boundary conditions.
package image

import "agg_go/internal/basics"

// PixelFormat defines the interface that pixel formats must implement for image accessors.
type PixelFormat interface {
	Width() int
	Height() int
	PixWidth() int                  // Bytes per pixel
	PixPtr(x, y int) []basics.Int8u // Returns slice starting at pixel
}

// WrapMode defines the interface for coordinate wrapping strategies.
type WrapMode interface {
	Call(v int) basics.Int32u
	Inc() basics.Int32u
}

// ImageAccessorClip provides pixel access with background color for out-of-bounds coordinates.
// When accessing pixels outside the image bounds, returns a pre-configured background color.
type ImageAccessorClip[PixFmt PixelFormat] struct {
	pixfmt   *PixFmt
	bkBuffer []basics.Int8u
	x, x0, y int
	pixPtr   []basics.Int8u
}

// NewImageAccessorClip creates a new clipping image accessor with background color.
func NewImageAccessorClip[PixFmt PixelFormat](pixf *PixFmt, backgroundColor []basics.Int8u) *ImageAccessorClip[PixFmt] {
	accessor := &ImageAccessorClip[PixFmt]{
		pixfmt:   pixf,
		bkBuffer: make([]basics.Int8u, (*pixf).PixWidth()),
	}

	// Copy background color to buffer
	copy(accessor.bkBuffer, backgroundColor)

	return accessor
}

// Attach attaches a new pixel format to this accessor.
func (ia *ImageAccessorClip[PixFmt]) Attach(pixf *PixFmt) {
	ia.pixfmt = pixf
}

// SetBackgroundColor sets the background color for out-of-bounds pixels.
func (ia *ImageAccessorClip[PixFmt]) SetBackgroundColor(backgroundColor []basics.Int8u) {
	copy(ia.bkBuffer, backgroundColor)
}

// pixel returns a pointer to the current pixel or background if out of bounds.
func (ia *ImageAccessorClip[PixFmt]) pixel() []basics.Int8u {
	if ia.y >= 0 && ia.y < (*ia.pixfmt).Height() &&
		ia.x >= 0 && ia.x < (*ia.pixfmt).Width() {
		return (*ia.pixfmt).PixPtr(ia.x, ia.y)
	}
	return ia.bkBuffer
}

// Span initializes access for a horizontal span of pixels.
// Returns pointer to first pixel data or background if out of bounds.
func (ia *ImageAccessorClip[PixFmt]) Span(x, y, length int) []basics.Int8u {
	if y >= 0 && y < (*ia.pixfmt).Height() &&
		x >= 0 && x < (*ia.pixfmt).Width() && x+length <= (*ia.pixfmt).Width() {
		ia.x = x
		ia.x0 = x
		ia.y = y
		ia.pixPtr = (*ia.pixfmt).PixPtr(x, y)
		return ia.pixPtr
	}

	ia.x = x
	ia.x0 = x
	ia.y = y
	ia.pixPtr = nil
	return ia.bkBuffer // Return background directly for out-of-bounds span
}

// NextX moves to the next pixel in the current span.
func (ia *ImageAccessorClip[PixFmt]) NextX() []basics.Int8u {
	if ia.pixPtr != nil {
		pixWidth := (*ia.pixfmt).PixWidth()
		ia.pixPtr = ia.pixPtr[pixWidth:]
		return ia.pixPtr
	}
	ia.x++
	return ia.pixel()
}

// NextY moves to the next row at the original x position.
func (ia *ImageAccessorClip[PixFmt]) NextY() []basics.Int8u {
	ia.y++
	ia.x = ia.x0

	if ia.pixPtr != nil && ia.y >= 0 && ia.y < (*ia.pixfmt).Height() {
		ia.pixPtr = (*ia.pixfmt).PixPtr(ia.x, ia.y)
		return ia.pixPtr
	}

	ia.pixPtr = nil
	return ia.pixel()
}

// ImageAccessorNoClip provides direct pixel access without bounds checking.
// This is the fastest accessor but assumes all coordinates are valid.
type ImageAccessorNoClip[PixFmt PixelFormat] struct {
	pixfmt *PixFmt
	x, y   int
	pixPtr []basics.Int8u
}

// NewImageAccessorNoClip creates a new no-clipping image accessor.
func NewImageAccessorNoClip[PixFmt PixelFormat](pixf *PixFmt) *ImageAccessorNoClip[PixFmt] {
	return &ImageAccessorNoClip[PixFmt]{
		pixfmt: pixf,
	}
}

// Attach attaches a new pixel format to this accessor.
func (ia *ImageAccessorNoClip[PixFmt]) Attach(pixf *PixFmt) {
	ia.pixfmt = pixf
}

// Span initializes access for a horizontal span of pixels.
func (ia *ImageAccessorNoClip[PixFmt]) Span(x, y, length int) []basics.Int8u {
	ia.x = x
	ia.y = y
	ia.pixPtr = (*ia.pixfmt).PixPtr(x, y)
	return ia.pixPtr
}

// NextX moves to the next pixel in the current span.
func (ia *ImageAccessorNoClip[PixFmt]) NextX() []basics.Int8u {
	pixWidth := (*ia.pixfmt).PixWidth()
	ia.pixPtr = ia.pixPtr[pixWidth:]
	return ia.pixPtr
}

// NextY moves to the next row at the original x position.
func (ia *ImageAccessorNoClip[PixFmt]) NextY() []basics.Int8u {
	ia.y++
	ia.pixPtr = (*ia.pixfmt).PixPtr(ia.x, ia.y)
	return ia.pixPtr
}

// ImageAccessorClone provides pixel access with edge clamping for out-of-bounds coordinates.
// When accessing pixels outside the image bounds, returns the nearest edge pixel.
type ImageAccessorClone[PixFmt PixelFormat] struct {
	pixfmt   *PixFmt
	x, x0, y int
	pixPtr   []basics.Int8u
}

// NewImageAccessorClone creates a new cloning image accessor.
func NewImageAccessorClone[PixFmt PixelFormat](pixf *PixFmt) *ImageAccessorClone[PixFmt] {
	return &ImageAccessorClone[PixFmt]{
		pixfmt: pixf,
	}
}

// Attach attaches a new pixel format to this accessor.
func (ia *ImageAccessorClone[PixFmt]) Attach(pixf *PixFmt) {
	ia.pixfmt = pixf
}

// pixel returns a pointer to the current pixel with coordinate clamping.
func (ia *ImageAccessorClone[PixFmt]) pixel() []basics.Int8u {
	x := ia.x
	y := ia.y

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= (*ia.pixfmt).Width() {
		x = (*ia.pixfmt).Width() - 1
	}
	if y >= (*ia.pixfmt).Height() {
		y = (*ia.pixfmt).Height() - 1
	}

	return (*ia.pixfmt).PixPtr(x, y)
}

// Span initializes access for a horizontal span of pixels.
func (ia *ImageAccessorClone[PixFmt]) Span(x, y, length int) []basics.Int8u {
	ia.x = x
	ia.x0 = x
	ia.y = y

	if y >= 0 && y < (*ia.pixfmt).Height() &&
		x >= 0 && x+length <= (*ia.pixfmt).Width() {
		ia.pixPtr = (*ia.pixfmt).PixPtr(x, y)
		return ia.pixPtr
	}

	ia.pixPtr = nil
	return ia.pixel()
}

// NextX moves to the next pixel in the current span.
func (ia *ImageAccessorClone[PixFmt]) NextX() []basics.Int8u {
	if ia.pixPtr != nil {
		ia.x++
		// Check if we're still within bounds
		if ia.x < (*ia.pixfmt).Width() {
			pixWidth := (*ia.pixfmt).PixWidth()
			ia.pixPtr = ia.pixPtr[pixWidth:]
			return ia.pixPtr
		}
		// Out of bounds, fall back to pixel() method
		ia.pixPtr = nil
	} else {
		ia.x++
	}
	return ia.pixel()
}

// NextY moves to the next row at the original x position.
func (ia *ImageAccessorClone[PixFmt]) NextY() []basics.Int8u {
	ia.y++
	ia.x = ia.x0

	if ia.pixPtr != nil && ia.y >= 0 && ia.y < (*ia.pixfmt).Height() {
		ia.pixPtr = (*ia.pixfmt).PixPtr(ia.x, ia.y)
		return ia.pixPtr
	}

	ia.pixPtr = nil
	return ia.pixel()
}

// ImageAccessorWrap provides pixel access with configurable wrapping modes.
// Uses separate wrap modes for X and Y coordinates to handle tiling patterns.
type ImageAccessorWrap[PixFmt PixelFormat, WrapX WrapMode, WrapY WrapMode] struct {
	pixfmt *PixFmt
	rowPtr []basics.Int8u
	x      int
	wrapX  WrapX
	wrapY  WrapY
}

// NewImageAccessorWrap creates a new wrapping image accessor.
func NewImageAccessorWrap[PixFmt PixelFormat, WrapX WrapMode, WrapY WrapMode](
	pixf *PixFmt, wrapX WrapX, wrapY WrapY,
) *ImageAccessorWrap[PixFmt, WrapX, WrapY] {
	return &ImageAccessorWrap[PixFmt, WrapX, WrapY]{
		pixfmt: pixf,
		wrapX:  wrapX,
		wrapY:  wrapY,
	}
}

// Attach attaches a new pixel format to this accessor.
func (ia *ImageAccessorWrap[PixFmt, WrapX, WrapY]) Attach(pixf *PixFmt) {
	ia.pixfmt = pixf
}

// Span initializes access for a horizontal span of pixels.
func (ia *ImageAccessorWrap[PixFmt, WrapX, WrapY]) Span(x, y, length int) []basics.Int8u {
	ia.x = x
	wrappedY := ia.wrapY.Call(y)
	ia.rowPtr = (*ia.pixfmt).PixPtr(0, int(wrappedY))

	wrappedX := ia.wrapX.Call(x)
	pixWidth := (*ia.pixfmt).PixWidth()
	return ia.rowPtr[int(wrappedX)*pixWidth:]
}

// NextX moves to the next pixel in the current span.
func (ia *ImageAccessorWrap[PixFmt, WrapX, WrapY]) NextX() []basics.Int8u {
	wrappedX := ia.wrapX.Inc()
	pixWidth := (*ia.pixfmt).PixWidth()
	return ia.rowPtr[int(wrappedX)*pixWidth:]
}

// NextY moves to the next row at the original x position.
func (ia *ImageAccessorWrap[PixFmt, WrapX, WrapY]) NextY() []basics.Int8u {
	wrappedY := ia.wrapY.Inc()
	ia.rowPtr = (*ia.pixfmt).PixPtr(0, int(wrappedY))

	wrappedX := ia.wrapX.Call(ia.x)
	pixWidth := (*ia.pixfmt).PixWidth()
	return ia.rowPtr[int(wrappedX)*pixWidth:]
}
