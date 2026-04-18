//go:build agogo && cgo

package cppbridge

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#cgo CPPFLAGS: -I${SRCDIR}
#include "bridge.h"
*/
import "C"

import (
	"fmt"
	"image"
	"math"
	"runtime"
	"unsafe"
)

// FillRule selects the winding rule used by native path fill operations.
type FillRule int

const (
	FillRuleNonZero FillRule = iota
	FillRuleEvenOdd
)

// LineCap selects the line ending style used by native stroke operations.
type LineCap int

const (
	LineCapButt LineCap = iota
	LineCapRound
	LineCapSquare
)

// LineJoin selects the corner join style used by native stroke operations.
type LineJoin int

const (
	LineJoinMiter LineJoin = iota
	LineJoinRound
	LineJoinBevel
)

// StrokeOptions configures the native stroke primitive currently exposed by the
// in-repo bridge.
type StrokeOptions struct {
	Width      float32
	LineCap    LineCap
	LineJoin   LineJoin
	MiterLimit float32
}

// DefaultStrokeOptions returns the bridge's current default stroke options.
func DefaultStrokeOptions() StrokeOptions {
	return StrokeOptions{
		Width:      1,
		LineCap:    LineCapButt,
		LineJoin:   LineJoinMiter,
		MiterLimit: 4,
	}
}

// Image is the first in-repo native image handle migrated from the previous
// AGoGo bridge surface.
type Image struct {
	ptr *C.AggGoCPPImage
}

// NewImage allocates a new RGBA image in the native bridge.
func NewImage(width, height int) (*Image, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("image dimensions must be positive, got %dx%d", width, height)
	}
	ptr := C.agg_go_cpp_image_create(C.uint32_t(width), C.uint32_t(height))
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_image_create failed: %s", lastError())
	}
	img := &Image{ptr: ptr}
	runtime.SetFinalizer(img, (*Image).Close)
	return img, nil
}

// Close frees the native image handle.
func (img *Image) Close() error {
	runtime.SetFinalizer(img, nil)
	if img == nil || img.ptr == nil {
		return nil
	}
	C.agg_go_cpp_image_free(img.ptr)
	img.ptr = nil
	return nil
}

func (img *Image) Width() int {
	if img == nil || img.ptr == nil {
		return 0
	}
	return int(C.agg_go_cpp_image_width(img.ptr))
}

func (img *Image) Height() int {
	if img == nil || img.ptr == nil {
		return 0
	}
	return int(C.agg_go_cpp_image_height(img.ptr))
}

func (img *Image) Stride() int {
	if img == nil || img.ptr == nil {
		return 0
	}
	return int(C.agg_go_cpp_image_stride(img.ptr))
}

// Clear fills the full native image with a solid RGBA color.
func (img *Image) Clear(r, g, b, a uint8) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("image is nil")
	}
	if code := int(C.agg_go_cpp_image_clear(img.ptr, C.uint8_t(r), C.uint8_t(g), C.uint8_t(b), C.uint8_t(a))); code != 0 {
		return fmt.Errorf("agg_go_cpp_image_clear failed: %s", lastError())
	}
	return nil
}

// PixelsRGBA copies the native image into a Go RGBA buffer.
func (img *Image) PixelsRGBA() ([]byte, error) {
	if img == nil || img.ptr == nil {
		return nil, fmt.Errorf("image is nil")
	}
	stride := img.Stride()
	height := img.Height()
	size := stride * height
	if size == 0 {
		return nil, nil
	}
	ptr := C.agg_go_cpp_image_pixels(img.ptr)
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_image_pixels returned nil")
	}
	return C.GoBytes(unsafe.Pointer(ptr), C.int(size)), nil
}

// ToGoImage converts the native image to a standard library RGBA image.
func (img *Image) ToGoImage() (*image.RGBA, error) {
	pixels, err := img.PixelsRGBA()
	if err != nil {
		return nil, err
	}
	return &image.RGBA{
		Pix:    pixels,
		Stride: img.Stride(),
		Rect:   image.Rect(0, 0, img.Width(), img.Height()),
	}, nil
}

// BlitFrom copies a source rectangle from src into the destination image.
func (img *Image) BlitFrom(src *Image, dstX, dstY, srcX, srcY, width, height int) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("destination image is nil")
	}
	if src == nil || src.ptr == nil {
		return fmt.Errorf("source image is nil")
	}
	code := int(C.agg_go_cpp_image_blit(
		img.ptr,
		src.ptr,
		C.int(dstX),
		C.int(dstY),
		C.int(srcX),
		C.int(srcY),
		C.uint32_t(width),
		C.uint32_t(height),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_image_blit failed: %s", lastError())
	}
	return nil
}

// Path is the first in-repo native path handle migrated from the previous
// AGoGo bridge surface.
type Path struct {
	ptr *C.AggGoCPPPath
}

// NewPath allocates a new native path handle.
func NewPath() (*Path, error) {
	ptr := C.agg_go_cpp_path_create()
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_path_create failed: %s", lastError())
	}
	path := &Path{ptr: ptr}
	runtime.SetFinalizer(path, (*Path).Close)
	return path, nil
}

// Close frees the native path handle.
func (p *Path) Close() error {
	runtime.SetFinalizer(p, nil)
	if p == nil || p.ptr == nil {
		return nil
	}
	C.agg_go_cpp_path_free(p.ptr)
	p.ptr = nil
	return nil
}

// Reset clears the path so it can be reused.
func (p *Path) Reset() error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if code := int(C.agg_go_cpp_path_reset(p.ptr)); code != 0 {
		return fmt.Errorf("agg_go_cpp_path_reset failed: %s", lastError())
	}
	return nil
}

func (p *Path) MoveTo(x, y float32) error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if code := int(C.agg_go_cpp_path_move_to(p.ptr, C.float(x), C.float(y))); code != 0 {
		return fmt.Errorf("agg_go_cpp_path_move_to failed: %s", lastError())
	}
	return nil
}

func (p *Path) LineTo(x, y float32) error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if code := int(C.agg_go_cpp_path_line_to(p.ptr, C.float(x), C.float(y))); code != 0 {
		return fmt.Errorf("agg_go_cpp_path_line_to failed: %s", lastError())
	}
	return nil
}

func (p *Path) ClosePath() error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if code := int(C.agg_go_cpp_path_close(p.ptr)); code != 0 {
		return fmt.Errorf("agg_go_cpp_path_close failed: %s", lastError())
	}
	return nil
}

// Transform returns a transformed copy of the native path.
func (p *Path) Transform(matrix *Matrix) (*Path, error) {
	if p == nil || p.ptr == nil {
		return nil, fmt.Errorf("path is nil")
	}
	if matrix == nil || matrix.ptr == nil {
		return nil, fmt.Errorf("matrix is nil")
	}
	ptr := C.agg_go_cpp_path_transform(p.ptr, matrix.ptr)
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_path_transform failed: %s", lastError())
	}
	path := &Path{ptr: ptr}
	runtime.SetFinalizer(path, (*Path).Close)
	return path, nil
}

// FillPath fills the native path into the native image using a simple RGBA
// solid color operation. This is an intermediate migrated bridge primitive, not
// the final AGG-backed engine adapter.
func FillPath(img *Image, path *Path, rule FillRule, r, g, b, a uint8) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("image is nil")
	}
	if path == nil || path.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	code := int(C.agg_go_cpp_render_fill_path(
		img.ptr,
		path.ptr,
		C.int(rule),
		C.uint8_t(r),
		C.uint8_t(g),
		C.uint8_t(b),
		C.uint8_t(a),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_render_fill_path failed: %s", lastError())
	}
	return nil
}

// StrokePath strokes the native path into the native image using the current
// bridge's minimal stroke primitive.
func StrokePath(img *Image, path *Path, opts StrokeOptions, r, g, b, a uint8) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("image is nil")
	}
	if path == nil || path.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	code := int(C.agg_go_cpp_render_stroke_path(
		img.ptr,
		path.ptr,
		C.float(opts.Width),
		C.int(opts.LineCap),
		C.int(opts.LineJoin),
		C.float(opts.MiterLimit),
		C.uint8_t(r),
		C.uint8_t(g),
		C.uint8_t(b),
		C.uint8_t(a),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_render_stroke_path failed: %s", lastError())
	}
	return nil
}

// Matrix is the affine transform handle currently exposed by the in-repo
// native bridge.
type Matrix struct {
	ptr *C.AggGoCPPMatrix
}

// NewMatrix allocates a new identity matrix.
func NewMatrix() (*Matrix, error) {
	ptr := C.agg_go_cpp_matrix_create()
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_matrix_create failed: %s", lastError())
	}
	matrix := &Matrix{ptr: ptr}
	runtime.SetFinalizer(matrix, (*Matrix).Close)
	return matrix, nil
}

// Close frees the native matrix handle.
func (m *Matrix) Close() error {
	runtime.SetFinalizer(m, nil)
	if m == nil || m.ptr == nil {
		return nil
	}
	C.agg_go_cpp_matrix_free(m.ptr)
	m.ptr = nil
	return nil
}

func (m *Matrix) Identity() error {
	if m == nil || m.ptr == nil {
		return fmt.Errorf("matrix is nil")
	}
	if code := int(C.agg_go_cpp_matrix_identity(m.ptr)); code != 0 {
		return fmt.Errorf("agg_go_cpp_matrix_identity failed: %s", lastError())
	}
	return nil
}

func (m *Matrix) Translate(tx, ty float32) error {
	if m == nil || m.ptr == nil {
		return fmt.Errorf("matrix is nil")
	}
	if code := int(C.agg_go_cpp_matrix_translate(m.ptr, C.float(tx), C.float(ty))); code != 0 {
		return fmt.Errorf("agg_go_cpp_matrix_translate failed: %s", lastError())
	}
	return nil
}

func (m *Matrix) Scale(sx, sy float32) error {
	if m == nil || m.ptr == nil {
		return fmt.Errorf("matrix is nil")
	}
	if code := int(C.agg_go_cpp_matrix_scale(m.ptr, C.float(sx), C.float(sy))); code != 0 {
		return fmt.Errorf("agg_go_cpp_matrix_scale failed: %s", lastError())
	}
	return nil
}

func (m *Matrix) Rotate(angle float32) error {
	if m == nil || m.ptr == nil {
		return fmt.Errorf("matrix is nil")
	}
	if code := int(C.agg_go_cpp_matrix_rotate(m.ptr, C.float(angle))); code != 0 {
		return fmt.Errorf("agg_go_cpp_matrix_rotate failed: %s", lastError())
	}
	return nil
}

func (m *Matrix) RotateDegrees(degrees float32) error {
	return m.Rotate(float32(float64(degrees) * math.Pi / 180.0))
}

func (m *Matrix) TransformPoint(x, y float64) (float64, float64, error) {
	if m == nil || m.ptr == nil {
		return 0, 0, fmt.Errorf("matrix is nil")
	}
	cx := C.double(x)
	cy := C.double(y)
	if code := int(C.agg_go_cpp_matrix_transform_point(m.ptr, &cx, &cy)); code != 0 {
		return 0, 0, fmt.Errorf("agg_go_cpp_matrix_transform_point failed: %s", lastError())
	}
	return float64(cx), float64(cy), nil
}

func lastError() string {
	if msg := C.GoString(C.agg_go_cpp_bridge_last_error()); msg != "" {
		return msg
	}
	return "unknown native bridge failure"
}
