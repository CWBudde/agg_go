// Package platform provides the core platform support infrastructure for AGG applications.
// It's designed to create interactive demo examples with basic window management,
// event handling, and rendering capabilities.
//
// This is a Go port of the AGG 2.6 platform support system, adapted to be
// platform-agnostic and focused on testing and demonstration purposes.
package platform

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agg_go/internal/buffer"
	"agg_go/internal/platform/types"
)

// Type aliases to avoid breaking existing code while using shared types
type (
	WindowFlags   = types.WindowFlags
	PixelFormat   = types.PixelFormat
	InputFlags    = types.InputFlags
	KeyCode       = types.KeyCode
	EventCallback = types.EventCallback
)

// Re-export constants for backward compatibility
const (
	WindowResize          = types.WindowResize
	WindowHWBuffer        = types.WindowHWBuffer
	WindowKeepAspectRatio = types.WindowKeepAspectRatio
	WindowProcessAllKeys  = types.WindowProcessAllKeys
)

// Re-export pixel format constants for backward compatibility
const (
	PixelFormatUndefined = types.PixelFormatUndefined
	PixelFormatBW        = types.PixelFormatBW
	PixelFormatGray8     = types.PixelFormatGray8
	PixelFormatSGray8    = types.PixelFormatSGray8
	PixelFormatGray16    = types.PixelFormatGray16
	PixelFormatGray32    = types.PixelFormatGray32
	PixelFormatRGB555    = types.PixelFormatRGB555
	PixelFormatRGB565    = types.PixelFormatRGB565
	PixelFormatRGB24     = types.PixelFormatRGB24
	PixelFormatSRGB24    = types.PixelFormatSRGB24
	PixelFormatBGR24     = types.PixelFormatBGR24
	PixelFormatSBGR24    = types.PixelFormatSBGR24
	PixelFormatRGBA32    = types.PixelFormatRGBA32
	PixelFormatSRGBA32   = types.PixelFormatSRGBA32
	PixelFormatARGB32    = types.PixelFormatARGB32
	PixelFormatSARGB32   = types.PixelFormatSARGB32
	PixelFormatABGR32    = types.PixelFormatABGR32
	PixelFormatSABGR32   = types.PixelFormatSABGR32
	PixelFormatBGRA32    = types.PixelFormatBGRA32
	PixelFormatSBGRA32   = types.PixelFormatSBGRA32
	PixelFormatRGB48     = types.PixelFormatRGB48
	PixelFormatSRGB48    = types.PixelFormatSRGB48
	PixelFormatBGR48     = types.PixelFormatBGR48
	PixelFormatSBGR48    = types.PixelFormatSBGR48
	PixelFormatRGBA64    = types.PixelFormatRGBA64
	PixelFormatSRGBA64   = types.PixelFormatSRGBA64
	PixelFormatARGB64    = types.PixelFormatARGB64
	PixelFormatSARGB64   = types.PixelFormatSARGB64
	PixelFormatABGR64    = types.PixelFormatABGR64
	PixelFormatSABGR64   = types.PixelFormatSABGR64
	PixelFormatBGRA64    = types.PixelFormatBGRA64
	PixelFormatSBGRA64   = types.PixelFormatSBGRA64
	PixelFormatRGB96     = types.PixelFormatRGB96
	PixelFormatSRGB96    = types.PixelFormatSRGB96
	PixelFormatBGR96     = types.PixelFormatBGR96
	PixelFormatSBGR96    = types.PixelFormatSBGR96
	PixelFormatRGBA128   = types.PixelFormatRGBA128
	PixelFormatSRGBA128  = types.PixelFormatSRGBA128
	PixelFormatARGB128   = types.PixelFormatARGB128
	PixelFormatSARGB128  = types.PixelFormatSARGB128
	PixelFormatABGR128   = types.PixelFormatABGR128
	PixelFormatSABGR128  = types.PixelFormatSABGR128
	PixelFormatBGRA128   = types.PixelFormatBGRA128
	PixelFormatSBGRA128  = types.PixelFormatSBGRA128
)

// Re-export input flag constants for backward compatibility with examples
const (
	MouseLeft  = types.MouseLeft
	MouseRight = types.MouseRight
	KbdShift   = types.KbdShift
	KbdCtrl    = types.KbdCtrl
)

// Re-export key constants for backward compatibility with examples
const (
	// ASCII set
	KeyBackspace = types.KeyBackspace
	KeyTab       = types.KeyTab
	KeyClear     = types.KeyClear
	KeyReturn    = types.KeyReturn
	KeyPause     = types.KeyPause
	KeyEscape    = types.KeyEscape
	KeyDelete    = types.KeyDelete

	// Keypad
	KeyKP0        = types.KeyKP0
	KeyKP1        = types.KeyKP1
	KeyKP2        = types.KeyKP2
	KeyKP3        = types.KeyKP3
	KeyKP4        = types.KeyKP4
	KeyKP5        = types.KeyKP5
	KeyKP6        = types.KeyKP6
	KeyKP7        = types.KeyKP7
	KeyKP8        = types.KeyKP8
	KeyKP9        = types.KeyKP9
	KeyKPPeriod   = types.KeyKPPeriod
	KeyKPDivide   = types.KeyKPDivide
	KeyKPMultiply = types.KeyKPMultiply
	KeyKPMinus    = types.KeyKPMinus
	KeyKPPlus     = types.KeyKPPlus
	KeyKPEnter    = types.KeyKPEnter
	KeyKPEquals   = types.KeyKPEquals

	// Arrow keys and navigation
	KeyUp       = types.KeyUp
	KeyDown     = types.KeyDown
	KeyRight    = types.KeyRight
	KeyLeft     = types.KeyLeft
	KeyInsert   = types.KeyInsert
	KeyHome     = types.KeyHome
	KeyEnd      = types.KeyEnd
	KeyPageUp   = types.KeyPageUp
	KeyPageDown = types.KeyPageDown

	// Function keys
	KeyF1  = types.KeyF1
	KeyF2  = types.KeyF2
	KeyF3  = types.KeyF3
	KeyF4  = types.KeyF4
	KeyF5  = types.KeyF5
	KeyF6  = types.KeyF6
	KeyF7  = types.KeyF7
	KeyF8  = types.KeyF8
	KeyF9  = types.KeyF9
	KeyF10 = types.KeyF10
	KeyF11 = types.KeyF11
	KeyF12 = types.KeyF12

	// Modifier keys
	KeyNumLock    = types.KeyNumLock
	KeyCapsLock   = types.KeyCapsLock
	KeyScrollLock = types.KeyScrollLock
	KeyRShift     = types.KeyRShift
	KeyLShift     = types.KeyLShift
	KeyRCtrl      = types.KeyRCtrl
	KeyLCtrl      = types.KeyLCtrl
	KeyRAlt       = types.KeyRAlt
	KeyLAlt       = types.KeyLAlt
)

// PlatformSupport provides the core platform support functionality for AGG applications.
// It manages rendering buffers, handles events, and provides basic window operations.
type PlatformSupport struct {
	// Window configuration
	format      PixelFormat
	flipY       bool
	bpp         int
	windowFlags WindowFlags
	caption     string
	waitMode    bool

	// Window dimensions
	initialWidth  int
	initialHeight int
	currentWidth  int
	currentHeight int

	// Rendering buffers
	windowBuffer buffer.RenderingBuffer[uint8]
	imageBuffers [maxImages]buffer.RenderingBuffer[uint8]

	// Timer
	startTime time.Time

	// Event handlers
	onInitHandler       func()
	onResizeHandler     func(width, height int)
	onIdleHandler       func()
	onMouseMoveHandler  func(x, y int, flags InputFlags)
	onMouseDownHandler  func(x, y int, flags InputFlags)
	onMouseUpHandler    func(x, y int, flags InputFlags)
	onKeyHandler        func(x, y int, key KeyCode, flags InputFlags)
	onCtrlChangeHandler func()
	onDrawHandler       func()
	onPostDrawHandler   func(rawHandler RawEventHandler)
}

const (
	maxImages = 16 // Maximum number of image buffers
)

// BMP file header structures
type BMPFileHeader struct {
	Type      uint16 // File type, must be 'BM'
	Size      uint32 // Size of file in bytes
	Reserved1 uint16 // Reserved, must be 0
	Reserved2 uint16 // Reserved, must be 0
	OffBits   uint32 // Offset to bitmap data in bytes
}

type BMPInfoHeader struct {
	Size          uint32 // Size of this header in bytes
	Width         int32  // Width of bitmap in pixels
	Height        int32  // Height of bitmap in pixels
	Planes        uint16 // Number of color planes, must be 1
	BitCount      uint16 // Number of bits per pixel
	Compression   uint32 // Compression method used
	SizeImage     uint32 // Size of bitmap in bytes
	XPelsPerMeter int32  // Horizontal resolution in pixels per meter
	YPelsPerMeter int32  // Vertical resolution in pixels per meter
	ClrUsed       uint32 // Number of colors in color table
	ClrImportant  uint32 // Number of important colors used
}

// NewPlatformSupport creates a new platform support instance with the specified pixel format and Y-axis orientation.
func NewPlatformSupport(format PixelFormat, flipY bool) *PlatformSupport {
	ps := &PlatformSupport{
		format:   format,
		flipY:    flipY,
		bpp:      format.BPP(),
		waitMode: false,
		caption:  "AGG Application",
	}

	// Initialize buffers
	ps.windowBuffer = *buffer.NewRenderingBuffer[uint8]()
	for i := range ps.imageBuffers {
		ps.imageBuffers[i] = *buffer.NewRenderingBuffer[uint8]()
	}

	return ps
}

// Caption sets the window caption (title).
func (ps *PlatformSupport) Caption(caption string) {
	ps.caption = caption
}

// GetCaption returns the current window caption.
func (ps *PlatformSupport) GetCaption() string {
	return ps.caption
}

// Format returns the pixel format.
func (ps *PlatformSupport) Format() PixelFormat {
	return ps.format
}

// FlipY returns whether the Y-axis is flipped.
func (ps *PlatformSupport) FlipY() bool {
	return ps.flipY
}

// BPP returns the bits per pixel.
func (ps *PlatformSupport) BPP() int {
	return ps.bpp
}

// WaitMode returns the current wait mode setting.
func (ps *PlatformSupport) WaitMode() bool {
	return ps.waitMode
}

// SetWaitMode sets the wait mode. When true, the application waits for events
// and doesn't call OnIdle(). When false, it calls OnIdle() when the event queue is empty.
func (ps *PlatformSupport) SetWaitMode(waitMode bool) {
	ps.waitMode = waitMode
}

// Init initializes the platform support with the specified window dimensions and flags.
func (ps *PlatformSupport) Init(width, height int, flags WindowFlags) error {
	ps.initialWidth = width
	ps.initialHeight = height
	ps.currentWidth = width
	ps.currentHeight = height
	ps.windowFlags = flags

	// Calculate stride based on pixel format
	stride := width * ps.bpp / 8
	bufferSize := stride * height

	// Initialize window buffer
	windowData := make([]uint8, bufferSize)
	ps.windowBuffer.Attach(windowData, width, height, stride)

	// Call initialization handler
	if ps.onInitHandler != nil {
		ps.onInitHandler()
	}

	return nil
}

// WindowFlags returns the current window flags.
func (ps *PlatformSupport) WindowFlags() WindowFlags {
	return ps.windowFlags
}

// Width returns the current window width.
func (ps *PlatformSupport) Width() int {
	return ps.currentWidth
}

// Height returns the current window height.
func (ps *PlatformSupport) Height() int {
	return ps.currentHeight
}

// InitialWidth returns the initial window width.
func (ps *PlatformSupport) InitialWidth() int {
	return ps.initialWidth
}

// InitialHeight returns the initial window height.
func (ps *PlatformSupport) InitialHeight() int {
	return ps.initialHeight
}

// WindowBuffer returns a reference to the main rendering buffer.
func (ps *PlatformSupport) WindowBuffer() *buffer.RenderingBuffer[uint8] {
	return &ps.windowBuffer
}

// ImageBuffer returns a reference to the specified image buffer.
func (ps *PlatformSupport) ImageBuffer(idx int) *buffer.RenderingBuffer[uint8] {
	if idx >= 0 && idx < maxImages {
		return &ps.imageBuffers[idx]
	}
	return nil
}

// CreateImage creates an image buffer with the specified dimensions.
// If width or height is 0, uses the current window dimensions.
func (ps *PlatformSupport) CreateImage(idx int, width, height int) bool {
	if idx < 0 || idx >= maxImages {
		return false
	}

	if width == 0 {
		width = ps.currentWidth
	}
	if height == 0 {
		height = ps.currentHeight
	}

	stride := width * ps.bpp / 8
	bufferSize := stride * height
	imageData := make([]uint8, bufferSize)

	ps.imageBuffers[idx].Attach(imageData, width, height, stride)
	return true
}

// loadBMP loads a BMP image file and converts it to the platform's pixel format
func (ps *PlatformSupport) loadBMP(filename string) ([]uint8, int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, err
	}
	defer file.Close()

	// Read BMP file header
	var fileHeader BMPFileHeader
	if err := binary.Read(file, binary.LittleEndian, &fileHeader); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read BMP file header: %v", err)
	}

	// Verify BMP signature
	if fileHeader.Type != 0x4D42 { // "BM" in little endian
		return nil, 0, 0, fmt.Errorf("not a valid BMP file")
	}

	// Read BMP info header
	var infoHeader BMPInfoHeader
	if err := binary.Read(file, binary.LittleEndian, &infoHeader); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read BMP info header: %v", err)
	}

	// Validate basic BMP properties
	if infoHeader.Planes != 1 {
		return nil, 0, 0, fmt.Errorf("unsupported number of planes: %d", infoHeader.Planes)
	}
	if infoHeader.Compression != 0 {
		return nil, 0, 0, fmt.Errorf("compressed BMP files are not supported")
	}
	if infoHeader.BitCount != 24 && infoHeader.BitCount != 32 {
		return nil, 0, 0, fmt.Errorf("unsupported bit depth: %d", infoHeader.BitCount)
	}

	width := int(infoHeader.Width)
	height := int(infoHeader.Height)
	if width <= 0 || height <= 0 {
		return nil, 0, 0, fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	// BMP images are typically stored bottom-to-top
	flipVertical := height > 0
	if height < 0 {
		height = -height
		flipVertical = false
	}

	// Seek to pixel data
	if _, err := file.Seek(int64(fileHeader.OffBits), 0); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to seek to pixel data: %v", err)
	}

	// Calculate stride and buffer size for our target format
	targetStride := width * ps.bpp / 8
	buffer := make([]uint8, height*targetStride)

	// Read pixel data
	srcStride := ((width*int(infoHeader.BitCount) + 31) / 32) * 4 // BMP row padding
	rowData := make([]uint8, srcStride)

	for y := 0; y < height; y++ {
		if _, err := file.Read(rowData); err != nil {
			return nil, 0, 0, fmt.Errorf("failed to read pixel data: %v", err)
		}

		// Determine target row (handle vertical flip)
		targetY := y
		if flipVertical != ps.flipY {
			targetY = height - 1 - y
		}

		// Convert pixels to target format
		for x := 0; x < width; x++ {
			var r, g, b, a uint8
			srcIdx := x * int(infoHeader.BitCount) / 8

			// Read source pixel (BMP is BGR format)
			if infoHeader.BitCount == 24 {
				b = rowData[srcIdx]
				g = rowData[srcIdx+1]
				r = rowData[srcIdx+2]
				a = 255
			} else { // 32-bit BGRA
				b = rowData[srcIdx]
				g = rowData[srcIdx+1]
				r = rowData[srcIdx+2]
				a = rowData[srcIdx+3]
			}

			// Write to target buffer (assuming RGBA format)
			dstIdx := targetY*targetStride + x*ps.bpp/8
			switch ps.bpp {
			case 32:
				buffer[dstIdx] = r
				buffer[dstIdx+1] = g
				buffer[dstIdx+2] = b
				buffer[dstIdx+3] = a
			case 24:
				buffer[dstIdx] = r
				buffer[dstIdx+1] = g
				buffer[dstIdx+2] = b
			}
		}
	}

	return buffer, width, height, nil
}

// loadPPM loads a PPM P6 (binary) image file
func (ps *PlatformSupport) loadPPM(filename string) ([]uint8, int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, err
	}
	defer file.Close()

	// Read PPM header
	var magic string
	var width, height, maxVal int

	// Read magic number
	if _, err := fmt.Fscanf(file, "%s", &magic); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read PPM magic: %v", err)
	}
	if magic != "P6" {
		return nil, 0, 0, fmt.Errorf("unsupported PPM format: %s (only P6 supported)", magic)
	}

	// Read dimensions
	if _, err := fmt.Fscanf(file, "%d %d", &width, &height); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read PPM dimensions: %v", err)
	}
	if width <= 0 || height <= 0 {
		return nil, 0, 0, fmt.Errorf("invalid PPM dimensions: %dx%d", width, height)
	}

	// Read maximum value
	if _, err := fmt.Fscanf(file, "%d", &maxVal); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read PPM max value: %v", err)
	}
	if maxVal != 255 {
		return nil, 0, 0, fmt.Errorf("unsupported PPM max value: %d (only 255 supported)", maxVal)
	}

	// Skip whitespace after header
	var dummy byte
	file.Read([]byte{dummy})

	// Calculate target stride and allocate buffer
	targetStride := width * ps.bpp / 8
	buffer := make([]uint8, height*targetStride)

	// Read pixel data (PPM is RGB)
	pixelData := make([]uint8, width*height*3)
	if _, err := file.Read(pixelData); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read PPM pixel data: %v", err)
	}

	// Convert to target format
	for y := 0; y < height; y++ {
		targetY := y
		if ps.flipY {
			targetY = height - 1 - y
		}

		for x := 0; x < width; x++ {
			srcIdx := (y*width + x) * 3
			r := pixelData[srcIdx]
			g := pixelData[srcIdx+1]
			b := pixelData[srcIdx+2]

			dstIdx := targetY*targetStride + x*ps.bpp/8
			switch ps.bpp {
			case 32:
				buffer[dstIdx] = r
				buffer[dstIdx+1] = g
				buffer[dstIdx+2] = b
				buffer[dstIdx+3] = 255 // Full alpha
			case 24:
				buffer[dstIdx] = r
				buffer[dstIdx+1] = g
				buffer[dstIdx+2] = b
			}
		}
	}

	return buffer, width, height, nil
}

// loadPNG loads a PNG image file using Go's standard library
func (ps *PlatformSupport) loadPNG(filename string) ([]uint8, int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, err
	}
	defer file.Close()

	// Decode PNG image
	img, err := png.Decode(file)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to decode PNG: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate target stride and allocate buffer
	targetStride := width * ps.bpp / 8
	buffer := make([]uint8, height*targetStride)

	// Convert image to target format
	for y := 0; y < height; y++ {
		targetY := y
		if ps.flipY {
			targetY = height - 1 - y
		}

		for x := 0; x < width; x++ {
			srcR, srcG, srcB, srcA := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

			// Convert from 16-bit to 8-bit
			r := uint8(srcR >> 8)
			g := uint8(srcG >> 8)
			b := uint8(srcB >> 8)
			a := uint8(srcA >> 8)

			dstIdx := targetY*targetStride + x*ps.bpp/8
			switch ps.bpp {
			case 32:
				buffer[dstIdx] = r
				buffer[dstIdx+1] = g
				buffer[dstIdx+2] = b
				buffer[dstIdx+3] = a
			case 24:
				buffer[dstIdx] = r
				buffer[dstIdx+1] = g
				buffer[dstIdx+2] = b
			}
		}
	}

	return buffer, width, height, nil
}

// LoadImage loads an image from file.
func (ps *PlatformSupport) LoadImage(idx int, filename string) bool {
	if idx < 0 || idx >= maxImages {
		return false
	}

	// Determine file format from extension
	ext := strings.ToLower(filepath.Ext(filename))

	// Try to append extension if none provided
	if ext == "" {
		// Try .bmp first (default for AGG), then .ppm, then .png
		for _, tryExt := range []string{".bmp", ".ppm", ".png"} {
			tryFilename := filename + tryExt
			if _, err := os.Stat(tryFilename); err == nil {
				filename = tryFilename
				ext = tryExt
				break
			}
		}

		// If still no extension, default to .bmp
		if ext == "" {
			filename += ".bmp"
			ext = ".bmp"
		}
	}

	// Load image based on format
	var buffer []uint8
	var width, height int
	var err error

	switch ext {
	case ".bmp":
		buffer, width, height, err = ps.loadBMP(filename)
	case ".ppm":
		buffer, width, height, err = ps.loadPPM(filename)
	case ".png":
		buffer, width, height, err = ps.loadPNG(filename)
	default:
		return false
	}

	if err != nil {
		fmt.Printf("Error loading image %s: %v\n", filename, err)
		return false
	}

	// Attach buffer to image slot
	stride := width * ps.bpp / 8
	ps.imageBuffers[idx].Attach(buffer, width, height, stride)
	return true
}

// saveBMP saves an image buffer to a BMP file
func (ps *PlatformSupport) saveBMP(filename string, buffer []uint8, width, height, stride int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Calculate BMP parameters
	bitsPerPixel := uint16(ps.bpp)
	imageSize := uint32(height * stride)
	fileSize := 54 + imageSize // File header (14) + Info header (40) + Image data

	// Create BMP file header
	fileHeader := BMPFileHeader{
		Type:      0x4D42, // "BM"
		Size:      fileSize,
		Reserved1: 0,
		Reserved2: 0,
		OffBits:   54, // Offset to pixel data (14 + 40)
	}

	// Create BMP info header
	infoHeader := BMPInfoHeader{
		Size:          40,
		Width:         int32(width),
		Height:        int32(height), // Positive = bottom-to-top
		Planes:        1,
		BitCount:      bitsPerPixel,
		Compression:   0, // No compression
		SizeImage:     imageSize,
		XPelsPerMeter: 2835, // 72 DPI
		YPelsPerMeter: 2835, // 72 DPI
		ClrUsed:       0,
		ClrImportant:  0,
	}

	// Write headers
	if err := binary.Write(file, binary.LittleEndian, &fileHeader); err != nil {
		return fmt.Errorf("failed to write BMP file header: %v", err)
	}
	if err := binary.Write(file, binary.LittleEndian, &infoHeader); err != nil {
		return fmt.Errorf("failed to write BMP info header: %v", err)
	}

	// Calculate BMP row stride (must be multiple of 4)
	bmpStride := ((width*int(bitsPerPixel) + 31) / 32) * 4
	rowPadding := bmpStride - (width * int(bitsPerPixel) / 8)
	padding := make([]uint8, rowPadding)

	// Write pixel data (BMP is bottom-to-top, BGR format)
	for y := height - 1; y >= 0; y-- {
		srcY := y
		if ps.flipY {
			srcY = height - 1 - y
		}

		for x := 0; x < width; x++ {
			srcIdx := srcY*stride + x*ps.bpp/8

			// Convert from source format to BGR(A)
			switch ps.bpp {
			case 32:
				// Source is RGBA, write as BGRA
				r := buffer[srcIdx]
				g := buffer[srcIdx+1]
				b := buffer[srcIdx+2]
				a := buffer[srcIdx+3]
				file.Write([]byte{b, g, r, a})
			case 24:
				// Source is RGB, write as BGR
				r := buffer[srcIdx]
				g := buffer[srcIdx+1]
				b := buffer[srcIdx+2]
				file.Write([]byte{b, g, r})
			}
		}

		// Write row padding
		if rowPadding > 0 {
			file.Write(padding)
		}
	}

	return nil
}

// savePPM saves an image buffer to a PPM P6 file
func (ps *PlatformSupport) savePPM(filename string, buffer []uint8, width, height, stride int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write PPM header
	header := fmt.Sprintf("P6\n%d %d\n255\n", width, height)
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write PPM header: %v", err)
	}

	// Write pixel data (PPM is top-to-bottom, RGB format)
	pixelData := make([]byte, 3)
	for y := 0; y < height; y++ {
		srcY := y
		if ps.flipY {
			srcY = height - 1 - y
		}

		for x := 0; x < width; x++ {
			srcIdx := srcY*stride + x*ps.bpp/8

			// Extract RGB components
			switch ps.bpp {
			case 32:
				pixelData[0] = buffer[srcIdx]   // R
				pixelData[1] = buffer[srcIdx+1] // G
				pixelData[2] = buffer[srcIdx+2] // B
			case 24:
				pixelData[0] = buffer[srcIdx]   // R
				pixelData[1] = buffer[srcIdx+1] // G
				pixelData[2] = buffer[srcIdx+2] // B
			}

			if _, err := file.Write(pixelData); err != nil {
				return fmt.Errorf("failed to write PPM pixel data: %v", err)
			}
		}
	}

	return nil
}

// savePNG saves an image buffer to a PNG file
func (ps *PlatformSupport) savePNG(filename string, buffer []uint8, width, height, stride int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create Go image from buffer
	bounds := image.Rect(0, 0, width, height)
	img := image.NewRGBA(bounds)

	// Copy pixel data
	for y := 0; y < height; y++ {
		srcY := y
		if ps.flipY {
			srcY = height - 1 - y
		}

		for x := 0; x < width; x++ {
			srcIdx := srcY*stride + x*ps.bpp/8
			dstIdx := y*img.Stride + x*4

			switch ps.bpp {
			case 32:
				img.Pix[dstIdx] = buffer[srcIdx]     // R
				img.Pix[dstIdx+1] = buffer[srcIdx+1] // G
				img.Pix[dstIdx+2] = buffer[srcIdx+2] // B
				img.Pix[dstIdx+3] = buffer[srcIdx+3] // A
			case 24:
				img.Pix[dstIdx] = buffer[srcIdx]     // R
				img.Pix[dstIdx+1] = buffer[srcIdx+1] // G
				img.Pix[dstIdx+2] = buffer[srcIdx+2] // B
				img.Pix[dstIdx+3] = 255              // A
			}
		}
	}

	// Encode as PNG
	return png.Encode(file, img)
}

// SaveImage saves an image to file.
func (ps *PlatformSupport) SaveImage(idx int, filename string) bool {
	if idx < 0 || idx >= maxImages || ps.imageBuffers[idx].Buf() == nil {
		return false
	}

	// Get buffer information
	buffer := ps.imageBuffers[idx].Buf()
	width := ps.imageBuffers[idx].Width()
	height := ps.imageBuffers[idx].Height()
	stride := ps.imageBuffers[idx].Stride()

	// Determine file format from extension
	ext := strings.ToLower(filepath.Ext(filename))

	// Default to .bmp if no extension
	if ext == "" {
		filename += ".bmp"
		ext = ".bmp"
	}

	// Save image based on format
	var err error
	switch ext {
	case ".bmp":
		err = ps.saveBMP(filename, buffer, width, height, stride)
	case ".ppm":
		err = ps.savePPM(filename, buffer, width, height, stride)
	case ".png":
		err = ps.savePNG(filename, buffer, width, height, stride)
	default:
		fmt.Printf("Unsupported image format: %s\n", ext)
		return false
	}

	if err != nil {
		fmt.Printf("Error saving image %s: %v\n", filename, err)
		return false
	}

	return true
}

// CopyImageToWindow copies the specified image buffer to the window buffer.
func (ps *PlatformSupport) CopyImageToWindow(idx int) {
	if idx >= 0 && idx < maxImages && ps.imageBuffers[idx].Buf() != nil {
		ps.windowBuffer.CopyFrom(&ps.imageBuffers[idx])
	}
}

// CopyWindowToImage copies the window buffer to the specified image buffer.
func (ps *PlatformSupport) CopyWindowToImage(idx int) {
	if idx >= 0 && idx < maxImages {
		ps.CreateImage(idx, ps.windowBuffer.Width(), ps.windowBuffer.Height())
		ps.imageBuffers[idx].CopyFrom(&ps.windowBuffer)
	}
}

// CopyImageToImage copies one image buffer to another.
func (ps *PlatformSupport) CopyImageToImage(idxTo, idxFrom int) {
	if idxFrom >= 0 && idxFrom < maxImages &&
		idxTo >= 0 && idxTo < maxImages &&
		ps.imageBuffers[idxFrom].Buf() != nil {
		fromBuffer := &ps.imageBuffers[idxFrom]
		ps.CreateImage(idxTo, fromBuffer.Width(), fromBuffer.Height())
		ps.imageBuffers[idxTo].CopyFrom(fromBuffer)
	}
}

// ForceRedraw sets a flag to redraw the window on the next event cycle.
func (ps *PlatformSupport) ForceRedraw() {
	// In a real implementation, this would set a redraw flag or send a message
	// For now, we just call the draw handler immediately
	if ps.onDrawHandler != nil {
		ps.onDrawHandler()
	}
}

// UpdateWindow immediately updates the window with the current buffer content.
func (ps *PlatformSupport) UpdateWindow() {
	// In a real implementation, this would copy the buffer to the actual window
	// For now, this is a no-op since we don't have actual window display
}

// StartTimer starts the timer for elapsed time measurement.
func (ps *PlatformSupport) StartTimer() {
	ps.startTime = time.Now()
}

// ElapsedTime returns the time elapsed since the last StartTimer() call in milliseconds.
func (ps *PlatformSupport) ElapsedTime() float64 {
	return float64(time.Since(ps.startTime).Nanoseconds()) / 1e6
}

// Message displays a message (stub implementation).
func (ps *PlatformSupport) Message(msg string) {
	// In a real implementation, this would show a message box
	fmt.Println("Platform Message:", msg)
}

// ImageExtension returns the default image file extension for this platform.
func (ps *PlatformSupport) ImageExtension() string {
	return ".bmp" // Default to BMP format for AGG compatibility
}

// SupportedImageExtensions returns a list of supported image file extensions.
func (ps *PlatformSupport) SupportedImageExtensions() []string {
	return []string{".bmp", ".ppm", ".png"}
}

// Run starts the main event loop (stub implementation).
func (ps *PlatformSupport) Run() int {
	// In a real implementation, this would start the platform-specific event loop
	// For now, just call the draw handler once
	if ps.onDrawHandler != nil {
		ps.onDrawHandler()
	}
	return 0
}

// Event handler setters

// SetOnInit sets the initialization event handler.
func (ps *PlatformSupport) SetOnInit(handler func()) {
	ps.onInitHandler = handler
}

// SetOnResize sets the resize event handler.
func (ps *PlatformSupport) SetOnResize(handler func(width, height int)) {
	ps.onResizeHandler = handler
}

// SetOnIdle sets the idle event handler.
func (ps *PlatformSupport) SetOnIdle(handler func()) {
	ps.onIdleHandler = handler
}

// SetOnMouseMove sets the mouse move event handler.
func (ps *PlatformSupport) SetOnMouseMove(handler func(x, y int, flags InputFlags)) {
	ps.onMouseMoveHandler = handler
}

// SetOnMouseDown sets the mouse button down event handler.
func (ps *PlatformSupport) SetOnMouseDown(handler func(x, y int, flags InputFlags)) {
	ps.onMouseDownHandler = handler
}

// SetOnMouseUp sets the mouse button up event handler.
func (ps *PlatformSupport) SetOnMouseUp(handler func(x, y int, flags InputFlags)) {
	ps.onMouseUpHandler = handler
}

// SetOnKey sets the keyboard event handler.
func (ps *PlatformSupport) SetOnKey(handler func(x, y int, key KeyCode, flags InputFlags)) {
	ps.onKeyHandler = handler
}

// SetOnCtrlChange sets the control change event handler.
func (ps *PlatformSupport) SetOnCtrlChange(handler func()) {
	ps.onCtrlChangeHandler = handler
}

// SetOnDraw sets the draw event handler.
func (ps *PlatformSupport) SetOnDraw(handler func()) {
	ps.onDrawHandler = handler
}

// SetOnPostDraw sets the post-draw event handler.
func (ps *PlatformSupport) SetOnPostDraw(handler func(rawHandler RawEventHandler)) {
	ps.onPostDrawHandler = handler
}

// Trigger event handlers for testing purposes

// TriggerResize triggers a resize event.
func (ps *PlatformSupport) TriggerResize(width, height int) {
	ps.currentWidth = width
	ps.currentHeight = height

	// Update window buffer
	stride := width * ps.bpp / 8
	bufferSize := stride * height
	windowData := make([]uint8, bufferSize)
	ps.windowBuffer.Attach(windowData, width, height, stride)

	if ps.onResizeHandler != nil {
		ps.onResizeHandler(width, height)
	}
}

// TriggerMouseMove triggers a mouse move event.
func (ps *PlatformSupport) TriggerMouseMove(x, y int, flags InputFlags) {
	if ps.onMouseMoveHandler != nil {
		ps.onMouseMoveHandler(x, y, flags)
	}
}

// TriggerMouseDown triggers a mouse button down event.
func (ps *PlatformSupport) TriggerMouseDown(x, y int, flags InputFlags) {
	if ps.onMouseDownHandler != nil {
		ps.onMouseDownHandler(x, y, flags)
	}
}

// TriggerMouseUp triggers a mouse button up event.
func (ps *PlatformSupport) TriggerMouseUp(x, y int, flags InputFlags) {
	if ps.onMouseUpHandler != nil {
		ps.onMouseUpHandler(x, y, flags)
	}
}

// TriggerKey triggers a keyboard event.
func (ps *PlatformSupport) TriggerKey(x, y int, key KeyCode, flags InputFlags) {
	if ps.onKeyHandler != nil {
		ps.onKeyHandler(x, y, key, flags)
	}
}

// TriggerIdle triggers an idle event.
func (ps *PlatformSupport) TriggerIdle() {
	if ps.onIdleHandler != nil {
		ps.onIdleHandler()
	}
}

// TriggerDraw triggers a draw event.
func (ps *PlatformSupport) TriggerDraw() {
	if ps.onDrawHandler != nil {
		ps.onDrawHandler()
	}
}
