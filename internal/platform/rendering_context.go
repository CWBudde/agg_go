package platform

import (
	"math"

	"agg_go/internal/buffer"
	"agg_go/internal/transform"
)

// RenderingContext provides enhanced rendering capabilities for platform support.
// It integrates the rendering buffer management with coordinate transformations
// and provides utilities for image manipulation and display.
type RenderingContext struct {
	platformSupport *PlatformSupport
	resizeMatrix    *transform.TransAffine
}

// NewRenderingContext creates a new rendering context attached to the given platform support.
func NewRenderingContext(ps *PlatformSupport) *RenderingContext {
	rc := &RenderingContext{
		platformSupport: ps,
		resizeMatrix:    transform.NewTransAffine(),
	}
	return rc
}

// PlatformSupport returns the underlying platform support instance.
func (rc *RenderingContext) PlatformSupport() *PlatformSupport {
	return rc.platformSupport
}

// WindowBuffer returns the main window rendering buffer.
func (rc *RenderingContext) WindowBuffer() *buffer.RenderingBuffer[uint8] {
	return rc.platformSupport.WindowBuffer()
}

// ImageBuffer returns the specified image buffer.
func (rc *RenderingContext) ImageBuffer(idx int) *buffer.RenderingBuffer[uint8] {
	return rc.platformSupport.ImageBuffer(idx)
}

// SetupResizeTransform sets up the resize transformation matrix based on the current
// window dimensions and the initial dimensions. This is typically called when
// the window is resized to maintain proper scaling of graphics elements.
func (rc *RenderingContext) SetupResizeTransform(width, height int) {
	if rc.platformSupport.windowFlags&WindowKeepAspectRatio != 0 {
		// Calculate uniform scaling that maintains aspect ratio
		scaleX := float64(width) / float64(rc.platformSupport.initialWidth)
		scaleY := float64(height) / float64(rc.platformSupport.initialHeight)

		// Use the smaller scale to ensure everything fits
		scale := math.Min(scaleX, scaleY)

		// Calculate centering offsets
		offsetX := (float64(width) - float64(rc.platformSupport.initialWidth)*scale) / 2.0
		offsetY := (float64(height) - float64(rc.platformSupport.initialHeight)*scale) / 2.0

		// Create transformation matrix: translate then scale
		rc.resizeMatrix = transform.NewTransAffineTranslation(offsetX, offsetY)
		rc.resizeMatrix.Multiply(transform.NewTransAffineScaling(scale))
	} else {
		// Simple non-uniform scaling
		scaleX := float64(width) / float64(rc.platformSupport.initialWidth)
		scaleY := float64(height) / float64(rc.platformSupport.initialHeight)
		rc.resizeMatrix = transform.NewTransAffineScalingXY(scaleX, scaleY)
	}
}

// ResizeTransform returns the current resize transformation matrix.
func (rc *RenderingContext) ResizeTransform() *transform.TransAffine {
	return rc.resizeMatrix
}

// TransformPoint applies the resize transformation to a point.
func (rc *RenderingContext) TransformPoint(x, y float64) (float64, float64) {
	rc.resizeMatrix.Transform(&x, &y)
	return x, y
}

// InverseTransformPoint applies the inverse resize transformation to a point.
// This is useful for converting screen coordinates back to logical coordinates.
func (rc *RenderingContext) InverseTransformPoint(x, y float64) (float64, float64) {
	// Create a copy for inversion
	inverse := *rc.resizeMatrix
	inverted := inverse.Invert()
	if inverted != nil {
		inverted.Transform(&x, &y)
		return x, y
	}
	return x, y // Return original coordinates if inversion fails
}

// ClearWindow clears the window buffer with the specified color components.
func (rc *RenderingContext) ClearWindow(r, g, b, a uint8) {
	buf := rc.WindowBuffer()
	if buf.Buf() == nil {
		return
	}

	data := buf.Buf()
	bpp := rc.platformSupport.bpp / 8

	// Fill buffer based on pixel format
	switch rc.platformSupport.format {
	case PixelFormatRGBA32, PixelFormatSRGBA32:
		for i := 0; i < len(data); i += 4 {
			if i+3 < len(data) {
				data[i] = r
				data[i+1] = g
				data[i+2] = b
				data[i+3] = a
			}
		}
	case PixelFormatBGRA32, PixelFormatSBGRA32:
		for i := 0; i < len(data); i += 4 {
			if i+3 < len(data) {
				data[i] = b
				data[i+1] = g
				data[i+2] = r
				data[i+3] = a
			}
		}
	case PixelFormatRGB24, PixelFormatSRGB24:
		for i := 0; i < len(data); i += 3 {
			if i+2 < len(data) {
				data[i] = r
				data[i+1] = g
				data[i+2] = b
			}
		}
	case PixelFormatBGR24, PixelFormatSBGR24:
		for i := 0; i < len(data); i += 3 {
			if i+2 < len(data) {
				data[i] = b
				data[i+1] = g
				data[i+2] = r
			}
		}
	case PixelFormatGray8, PixelFormatSGray8:
		// Convert to grayscale using standard luminance formula
		gray := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
		for i := 0; i < len(data); i++ {
			data[i] = gray
		}
	default:
		// For other formats, use a simple approach
		for i := 0; i < len(data); i += bpp {
			for j := 0; j < bpp && i+j < len(data); j++ {
				switch j {
				case 0:
					data[i+j] = r
				case 1:
					data[i+j] = g
				case 2:
					data[i+j] = b
				case 3:
					data[i+j] = a
				}
			}
		}
	}
}

// ClearImage clears the specified image buffer with the given color.
func (rc *RenderingContext) ClearImage(idx int, r, g, b, a uint8) {
	buf := rc.ImageBuffer(idx)
	if buf == nil || buf.Buf() == nil {
		return
	}

	data := buf.Buf()
	bpp := rc.platformSupport.bpp / 8

	// Use the same clearing logic as ClearWindow
	switch rc.platformSupport.format {
	case PixelFormatRGBA32, PixelFormatSRGBA32:
		for i := 0; i < len(data); i += 4 {
			if i+3 < len(data) {
				data[i] = r
				data[i+1] = g
				data[i+2] = b
				data[i+3] = a
			}
		}
	case PixelFormatBGRA32, PixelFormatSBGRA32:
		for i := 0; i < len(data); i += 4 {
			if i+3 < len(data) {
				data[i] = b
				data[i+1] = g
				data[i+2] = r
				data[i+3] = a
			}
		}
	default:
		// Simplified approach for other formats
		for i := 0; i < len(data); i += bpp {
			for j := 0; j < bpp && i+j < len(data); j++ {
				switch j {
				case 0:
					data[i+j] = r
				case 1:
					data[i+j] = g
				case 2:
					data[i+j] = b
				case 3:
					data[i+j] = a
				}
			}
		}
	}
}

// GetPixel gets a pixel value from the window buffer at the specified coordinates.
func (rc *RenderingContext) GetPixel(x, y int) (r, g, b, a uint8, ok bool) {
	buf := rc.WindowBuffer()
	if buf.Buf() == nil {
		return 0, 0, 0, 0, false
	}

	if x < 0 || y < 0 || x >= buf.Width() || y >= buf.Height() {
		return 0, 0, 0, 0, false
	}

	data := buf.Buf()
	bpp := rc.platformSupport.bpp / 8
	offset := y*buf.Stride() + x*bpp

	if offset+bpp > len(data) {
		return 0, 0, 0, 0, false
	}

	switch rc.platformSupport.format {
	case PixelFormatRGBA32, PixelFormatSRGBA32:
		return data[offset], data[offset+1], data[offset+2], data[offset+3], true
	case PixelFormatBGRA32, PixelFormatSBGRA32:
		return data[offset+2], data[offset+1], data[offset], data[offset+3], true
	case PixelFormatRGB24, PixelFormatSRGB24:
		return data[offset], data[offset+1], data[offset+2], 255, true
	case PixelFormatBGR24, PixelFormatSBGR24:
		return data[offset+2], data[offset+1], data[offset], 255, true
	case PixelFormatGray8, PixelFormatSGray8:
		gray := data[offset]
		return gray, gray, gray, 255, true
	default:
		// For other formats, make a best effort
		if bpp >= 3 {
			return data[offset], data[offset+1], data[offset+2], 255, true
		} else if bpp >= 1 {
			gray := data[offset]
			return gray, gray, gray, 255, true
		}
		return 0, 0, 0, 0, false
	}
}

// SetPixel sets a pixel value in the window buffer at the specified coordinates.
func (rc *RenderingContext) SetPixel(x, y int, r, g, b, a uint8) bool {
	buf := rc.WindowBuffer()
	if buf.Buf() == nil {
		return false
	}

	if x < 0 || y < 0 || x >= buf.Width() || y >= buf.Height() {
		return false
	}

	data := buf.Buf()
	bpp := rc.platformSupport.bpp / 8
	offset := y*buf.Stride() + x*bpp

	if offset+bpp > len(data) {
		return false
	}

	switch rc.platformSupport.format {
	case PixelFormatRGBA32, PixelFormatSRGBA32:
		data[offset] = r
		data[offset+1] = g
		data[offset+2] = b
		data[offset+3] = a
	case PixelFormatBGRA32, PixelFormatSBGRA32:
		data[offset] = b
		data[offset+1] = g
		data[offset+2] = r
		data[offset+3] = a
	case PixelFormatRGB24, PixelFormatSRGB24:
		data[offset] = r
		data[offset+1] = g
		data[offset+2] = b
	case PixelFormatBGR24, PixelFormatSBGR24:
		data[offset] = b
		data[offset+1] = g
		data[offset+2] = r
	case PixelFormatGray8, PixelFormatSGray8:
		// Convert to grayscale
		gray := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
		data[offset] = gray
	default:
		// Best effort for other formats
		for i := 0; i < bpp && offset+i < len(data); i++ {
			switch i {
			case 0:
				data[offset+i] = r
			case 1:
				data[offset+i] = g
			case 2:
				data[offset+i] = b
			case 3:
				data[offset+i] = a
			}
		}
	}
	return true
}

// BlendPixel blends a pixel with the existing pixel in the window buffer using alpha blending.
func (rc *RenderingContext) BlendPixel(x, y int, r, g, b, a uint8) bool {
	if a == 0 {
		return true // Fully transparent, no change
	}
	if a == 255 {
		return rc.SetPixel(x, y, r, g, b, a) // Fully opaque, replace
	}

	// Get existing pixel
	existingR, existingG, existingB, existingA, ok := rc.GetPixel(x, y)
	if !ok {
		return false
	}

	// Alpha blending: new = src * alpha + dst * (1 - alpha)
	alpha := float64(a) / 255.0
	invAlpha := 1.0 - alpha

	blendedR := uint8(float64(r)*alpha + float64(existingR)*invAlpha)
	blendedG := uint8(float64(g)*alpha + float64(existingG)*invAlpha)
	blendedB := uint8(float64(b)*alpha + float64(existingB)*invAlpha)
	blendedA := uint8(math.Max(float64(a), float64(existingA)))

	return rc.SetPixel(x, y, blendedR, blendedG, blendedB, blendedA)
}

// DrawLine draws a simple line using Bresenham's algorithm.
// This is a basic implementation for testing purposes.
func (rc *RenderingContext) DrawLine(x0, y0, x1, y1 int, r, g, b, a uint8) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)

	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}

	err := dx - dy
	x, y := x0, y0

	for {
		rc.SetPixel(x, y, r, g, b, a)

		if x == x1 && y == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// DrawRectangle draws a simple rectangle outline.
func (rc *RenderingContext) DrawRectangle(x, y, width, height int, r, g, b, a uint8) {
	// Top edge
	rc.DrawLine(x, y, x+width-1, y, r, g, b, a)
	// Bottom edge
	rc.DrawLine(x, y+height-1, x+width-1, y+height-1, r, g, b, a)
	// Left edge
	rc.DrawLine(x, y, x, y+height-1, r, g, b, a)
	// Right edge
	rc.DrawLine(x+width-1, y, x+width-1, y+height-1, r, g, b, a)
}

// FillRectangle fills a rectangle with the specified color.
func (rc *RenderingContext) FillRectangle(x, y, width, height int, r, g, b, a uint8) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			rc.SetPixel(x+dx, y+dy, r, g, b, a)
		}
	}
}

// DrawCircle draws a simple circle outline using the midpoint circle algorithm.
func (rc *RenderingContext) DrawCircle(centerX, centerY, radius int, r, g, b, a uint8) {
	x := radius
	y := 0
	err := 0

	for x >= y {
		rc.SetPixel(centerX+x, centerY+y, r, g, b, a)
		rc.SetPixel(centerX+y, centerY+x, r, g, b, a)
		rc.SetPixel(centerX-y, centerY+x, r, g, b, a)
		rc.SetPixel(centerX-x, centerY+y, r, g, b, a)
		rc.SetPixel(centerX-x, centerY-y, r, g, b, a)
		rc.SetPixel(centerX-y, centerY-x, r, g, b, a)
		rc.SetPixel(centerX+y, centerY-x, r, g, b, a)
		rc.SetPixel(centerX+x, centerY-y, r, g, b, a)

		y++
		err += 1 + 2*y
		if 2*(err-x)+1 > 0 {
			x--
			err += 1 - 2*x
		}
	}
}

// GetBufferInfo returns information about the current window buffer.
func (rc *RenderingContext) GetBufferInfo() (width, height, stride, bpp int, format PixelFormat) {
	buf := rc.WindowBuffer()
	return buf.Width(), buf.Height(), buf.Stride(), rc.platformSupport.bpp, rc.platformSupport.format
}

// GetImageInfo returns information about the specified image buffer.
func (rc *RenderingContext) GetImageInfo(idx int) (width, height, stride, bpp int, format PixelFormat, ok bool) {
	buf := rc.ImageBuffer(idx)
	if buf == nil || buf.Buf() == nil {
		return 0, 0, 0, 0, PixelFormatUndefined, false
	}
	return buf.Width(), buf.Height(), buf.Stride(), rc.platformSupport.bpp, rc.platformSupport.format, true
}

// ValidateBufferAccess checks if the given coordinates are within the window buffer bounds.
func (rc *RenderingContext) ValidateBufferAccess(x, y int) bool {
	buf := rc.WindowBuffer()
	return x >= 0 && y >= 0 && x < buf.Width() && y < buf.Height()
}

// Helper function to calculate absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Statistics contains rendering statistics and buffer information.
type Statistics struct {
	// Window buffer info
	WindowWidth  int    `json:"window_width"`
	WindowHeight int    `json:"window_height"`
	WindowStride int    `json:"window_stride"`
	PixelFormat  string `json:"pixel_format"`
	BPP          int    `json:"bpp"`
	FlipY        bool   `json:"flip_y"`

	// Buffer size info
	WindowBufferSize int `json:"window_buffer_size"`

	// Image buffer info
	ActiveImageBuffers int `json:"active_image_buffers"`

	// Transform info
	HasResizeTransform bool    `json:"has_resize_transform"`
	ResizeScaleX       float64 `json:"resize_scale_x,omitempty"`
	ResizeScaleY       float64 `json:"resize_scale_y,omitempty"`
	ResizeTranslateX   float64 `json:"resize_translate_x,omitempty"`
	ResizeTranslateY   float64 `json:"resize_translate_y,omitempty"`
}

// Statistics returns rendering statistics and buffer information.
func (rc *RenderingContext) Statistics() *Statistics {
	stats := &Statistics{}

	// Window buffer info
	buf := rc.WindowBuffer()
	stats.WindowWidth = buf.Width()
	stats.WindowHeight = buf.Height()
	stats.WindowStride = buf.Stride()
	stats.PixelFormat = rc.platformSupport.format.String()
	stats.BPP = rc.platformSupport.bpp
	stats.FlipY = rc.platformSupport.flipY

	// Calculate buffer size
	if buf.Buf() != nil {
		stats.WindowBufferSize = len(buf.Buf())
	} else {
		stats.WindowBufferSize = 0
	}

	// Count active image buffers
	activeImages := 0
	for i := 0; i < maxImages; i++ {
		if imgBuf := rc.ImageBuffer(i); imgBuf != nil && imgBuf.Buf() != nil {
			activeImages++
		}
	}
	stats.ActiveImageBuffers = activeImages

	// Transformation info
	isIdentity := rc.resizeMatrix.IsIdentity(1e-10)
	stats.HasResizeTransform = !isIdentity
	if !isIdentity {
		stats.ResizeScaleX = rc.resizeMatrix.SX
		stats.ResizeScaleY = rc.resizeMatrix.SY
		stats.ResizeTranslateX = rc.resizeMatrix.TX
		stats.ResizeTranslateY = rc.resizeMatrix.TY
	}

	return stats
}
