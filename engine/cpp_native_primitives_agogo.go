//go:build agogo && cgo

package engine

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#cgo CPPFLAGS: -I${SRCDIR}
#include <stdlib.h>
#include "cpp_native.h"
*/
import "C"

import (
	"fmt"
	"image"
	"math"
	"runtime"
	"unsafe"

	agg "github.com/cwbudde/agg_go"
)

type cppNativeFillRule int

const (
	cppNativeFillRuleNonZero cppNativeFillRule = iota
	cppNativeFillRuleEvenOdd
)

type cppNativeLineCap int

const (
	cppNativeLineCapButt cppNativeLineCap = iota
	cppNativeLineCapRound
	cppNativeLineCapSquare
)

type cppNativeLineJoin int

const (
	cppNativeLineJoinMiter cppNativeLineJoin = iota
	cppNativeLineJoinRound
	cppNativeLineJoinBevel
)

type cppNativeStrokeOptions struct {
	Width      float32
	LineCap    cppNativeLineCap
	LineJoin   cppNativeLineJoin
	MiterLimit float32
	// Dashes is a flat list of (dashLen, gapLen) pairs; an empty slice strokes
	// solid. DashStart is the phase offset along the path.
	Dashes    []float32
	DashStart float32
}

func defaultCPPNativeStrokeOptions() cppNativeStrokeOptions {
	return cppNativeStrokeOptions{
		Width:      1,
		LineCap:    cppNativeLineCapButt,
		LineJoin:   cppNativeLineJoinMiter,
		MiterLimit: 4,
	}
}

type cppNativeImage struct {
	ptr *C.AggGoCPPImage
}

func newCPPNativeImage(width, height int) (*cppNativeImage, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("image dimensions must be positive, got %dx%d", width, height)
	}
	ptr := C.agg_go_cpp_image_create(C.uint32_t(width), C.uint32_t(height))
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_image_create failed: %s", cppNativeLastError())
	}
	img := &cppNativeImage{ptr: ptr}
	runtime.SetFinalizer(img, (*cppNativeImage).close)
	return img, nil
}

func (img *cppNativeImage) close() error {
	runtime.SetFinalizer(img, nil)
	if img == nil || img.ptr == nil {
		return nil
	}
	C.agg_go_cpp_image_free(img.ptr)
	img.ptr = nil
	return nil
}

func (img *cppNativeImage) width() int {
	if img == nil || img.ptr == nil {
		return 0
	}
	return int(C.agg_go_cpp_image_width(img.ptr))
}

func (img *cppNativeImage) height() int {
	if img == nil || img.ptr == nil {
		return 0
	}
	return int(C.agg_go_cpp_image_height(img.ptr))
}

func (img *cppNativeImage) stride() int {
	if img == nil || img.ptr == nil {
		return 0
	}
	return int(C.agg_go_cpp_image_stride(img.ptr))
}

func (img *cppNativeImage) clear(r, g, b, a uint8) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("image is nil")
	}
	if code := int(C.agg_go_cpp_image_clear(img.ptr, C.uint8_t(r), C.uint8_t(g), C.uint8_t(b), C.uint8_t(a))); code != 0 {
		return fmt.Errorf("agg_go_cpp_image_clear failed: %s", cppNativeLastError())
	}
	return nil
}

func (img *cppNativeImage) pixelsRGBA() ([]byte, error) {
	if img == nil || img.ptr == nil {
		return nil, fmt.Errorf("image is nil")
	}
	size := img.stride() * img.height()
	if size == 0 {
		return nil, nil
	}
	ptr := C.agg_go_cpp_image_pixels(img.ptr)
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_image_pixels returned nil")
	}
	return C.GoBytes(unsafe.Pointer(ptr), C.int(size)), nil
}

func (img *cppNativeImage) pixelView() ([]byte, error) {
	if img == nil || img.ptr == nil {
		return nil, fmt.Errorf("image is nil")
	}
	size := img.stride() * img.height()
	if size == 0 {
		return nil, nil
	}
	ptr := C.agg_go_cpp_image_pixels_mut(img.ptr)
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_image_pixels_mut returned nil")
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size), nil
}

func (img *cppNativeImage) toGoImage() (*image.RGBA, error) {
	pixels, err := img.pixelsRGBA()
	if err != nil {
		return nil, err
	}
	return &image.RGBA{
		Pix:    pixels,
		Stride: img.stride(),
		Rect:   image.Rect(0, 0, img.width(), img.height()),
	}, nil
}

func (img *cppNativeImage) blitFrom(src *cppNativeImage, dstX, dstY, srcX, srcY, width, height int) error {
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
		return fmt.Errorf("agg_go_cpp_image_blit failed: %s", cppNativeLastError())
	}
	return nil
}

func (img *cppNativeImage) compositeFrom(src *cppNativeImage, dstX, dstY int, clip image.Rectangle, blendMode agg.BlendMode) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("destination image is nil")
	}
	if src == nil || src.ptr == nil {
		return fmt.Errorf("source image is nil")
	}
	code := int(C.agg_go_cpp_image_composite(
		img.ptr,
		src.ptr,
		C.int(dstX),
		C.int(dstY),
		C.int(clip.Min.X),
		C.int(clip.Min.Y),
		C.int(clip.Max.X),
		C.int(clip.Max.Y),
		C.int(blendMode),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_image_composite failed: %s", cppNativeLastError())
	}
	return nil
}

func (img *cppNativeImage) compositeScaledFrom(
	src *cppNativeImage,
	srcX,
	srcY,
	srcW,
	srcH,
	dstX,
	dstY,
	dstW,
	dstH int,
	clip image.Rectangle,
	blendMode agg.BlendMode,
) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("destination image is nil")
	}
	if src == nil || src.ptr == nil {
		return fmt.Errorf("source image is nil")
	}
	code := int(C.agg_go_cpp_image_composite_scaled(
		img.ptr,
		src.ptr,
		C.int(srcX),
		C.int(srcY),
		C.uint32_t(srcW),
		C.uint32_t(srcH),
		C.int(dstX),
		C.int(dstY),
		C.uint32_t(dstW),
		C.uint32_t(dstH),
		C.int(clip.Min.X),
		C.int(clip.Min.Y),
		C.int(clip.Max.X),
		C.int(clip.Max.Y),
		C.int(blendMode),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_image_composite_scaled failed: %s", cppNativeLastError())
	}
	return nil
}

func (img *cppNativeImage) compositeQuadFrom(
	src *cppNativeImage,
	srcX,
	srcY,
	srcW,
	srcH int,
	quad [8]float64,
	clip image.Rectangle,
	blendMode agg.BlendMode,
) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("destination image is nil")
	}
	if src == nil || src.ptr == nil {
		return fmt.Errorf("source image is nil")
	}
	code := int(C.agg_go_cpp_image_composite_quad(
		img.ptr,
		src.ptr,
		C.int(srcX),
		C.int(srcY),
		C.uint32_t(srcW),
		C.uint32_t(srcH),
		(*C.double)(unsafe.Pointer(&quad[0])),
		C.int(clip.Min.X),
		C.int(clip.Min.Y),
		C.int(clip.Max.X),
		C.int(clip.Max.Y),
		C.int(blendMode),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_image_composite_quad failed: %s", cppNativeLastError())
	}
	return nil
}

type cppNativePath struct {
	ptr *C.AggGoCPPPath
}

func newCPPNativePath() (*cppNativePath, error) {
	ptr := C.agg_go_cpp_path_create()
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_path_create failed: %s", cppNativeLastError())
	}
	path := &cppNativePath{ptr: ptr}
	runtime.SetFinalizer(path, (*cppNativePath).close)
	return path, nil
}

func (p *cppNativePath) close() error {
	runtime.SetFinalizer(p, nil)
	if p == nil || p.ptr == nil {
		return nil
	}
	C.agg_go_cpp_path_free(p.ptr)
	p.ptr = nil
	return nil
}

func (p *cppNativePath) reset() error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if code := int(C.agg_go_cpp_path_reset(p.ptr)); code != 0 {
		return fmt.Errorf("agg_go_cpp_path_reset failed: %s", cppNativeLastError())
	}
	return nil
}

func (p *cppNativePath) moveTo(x, y float32) error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if code := int(C.agg_go_cpp_path_move_to(p.ptr, C.float(x), C.float(y))); code != 0 {
		return fmt.Errorf("agg_go_cpp_path_move_to failed: %s", cppNativeLastError())
	}
	return nil
}

func (p *cppNativePath) lineTo(x, y float32) error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if code := int(C.agg_go_cpp_path_line_to(p.ptr, C.float(x), C.float(y))); code != 0 {
		return fmt.Errorf("agg_go_cpp_path_line_to failed: %s", cppNativeLastError())
	}
	return nil
}

func (p *cppNativePath) closePath() error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if code := int(C.agg_go_cpp_path_close(p.ptr)); code != 0 {
		return fmt.Errorf("agg_go_cpp_path_close failed: %s", cppNativeLastError())
	}
	return nil
}

func (p *cppNativePath) transform(matrix *cppNativeMatrix) (*cppNativePath, error) {
	if p == nil || p.ptr == nil {
		return nil, fmt.Errorf("path is nil")
	}
	if matrix == nil || matrix.ptr == nil {
		return nil, fmt.Errorf("matrix is nil")
	}
	ptr := C.agg_go_cpp_path_transform(p.ptr, matrix.ptr)
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_path_transform failed: %s", cppNativeLastError())
	}
	path := &cppNativePath{ptr: ptr}
	runtime.SetFinalizer(path, (*cppNativePath).close)
	return path, nil
}

func fillCPPNativePath(img *cppNativeImage, path *cppNativePath, rule cppNativeFillRule, r, g, b, a uint8) error {
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
		return fmt.Errorf("agg_go_cpp_render_fill_path failed: %s", cppNativeLastError())
	}
	return nil
}

// fillCPPNativePathComp fills a path directly onto img with the given blend mode
// and clip rectangle, applying the compositing operator per span with coverage.
func fillCPPNativePathComp(img *cppNativeImage, path *cppNativePath, rule cppNativeFillRule, clip image.Rectangle, blendMode agg.BlendMode, r, g, b, a uint8) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("image is nil")
	}
	if path == nil || path.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	code := int(C.agg_go_cpp_render_fill_path_comp(
		img.ptr,
		path.ptr,
		C.int(rule),
		C.int(blendMode),
		C.int(clip.Min.X),
		C.int(clip.Min.Y),
		C.int(clip.Max.X),
		C.int(clip.Max.Y),
		C.uint8_t(r),
		C.uint8_t(g),
		C.uint8_t(b),
		C.uint8_t(a),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_render_fill_path_comp failed: %s", cppNativeLastError())
	}
	return nil
}

func strokeCPPNativePath(img *cppNativeImage, path *cppNativePath, opts cppNativeStrokeOptions, r, g, b, a uint8) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("image is nil")
	}
	if path == nil || path.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	if len(opts.Dashes) >= 2 {
		return strokeCPPNativePathDashed(img, path, opts, r, g, b, a)
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
		return fmt.Errorf("agg_go_cpp_render_stroke_path failed: %s", cppNativeLastError())
	}
	return nil
}

// strokeCPPNativePathDashed strokes a dashed outline. opts.Dashes holds an even
// number of (dashLen, gapLen) values; pairs beyond the last complete pair are
// ignored.
func strokeCPPNativePathDashed(img *cppNativeImage, path *cppNativePath, opts cppNativeStrokeOptions, r, g, b, a uint8) error {
	pairCount := len(opts.Dashes) / 2
	dashes := make([]C.float, pairCount*2)
	for i := range dashes {
		dashes[i] = C.float(opts.Dashes[i])
	}
	code := int(C.agg_go_cpp_render_stroke_path_dashed(
		img.ptr,
		path.ptr,
		C.float(opts.Width),
		C.int(opts.LineCap),
		C.int(opts.LineJoin),
		C.float(opts.MiterLimit),
		&dashes[0],
		C.int(pairCount),
		C.float(opts.DashStart),
		C.uint8_t(r),
		C.uint8_t(g),
		C.uint8_t(b),
		C.uint8_t(a),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_render_stroke_path_dashed failed: %s", cppNativeLastError())
	}
	return nil
}

// strokeCPPNativePathComp strokes a path (optionally dashed) directly onto img
// with the given blend mode and clip rectangle, applying the compositing
// operator per span with coverage.
func strokeCPPNativePathComp(img *cppNativeImage, path *cppNativePath, opts cppNativeStrokeOptions, clip image.Rectangle, blendMode agg.BlendMode, r, g, b, a uint8) error {
	if img == nil || img.ptr == nil {
		return fmt.Errorf("image is nil")
	}
	if path == nil || path.ptr == nil {
		return fmt.Errorf("path is nil")
	}
	pairCount := len(opts.Dashes) / 2
	// A non-nil pointer is always passed; the C side ignores it when pairCount is
	// zero. The single-element fallback keeps &dashes[0] valid for solid strokes.
	dashes := make([]C.float, pairCount*2)
	for i := range dashes {
		dashes[i] = C.float(opts.Dashes[i])
	}
	if len(dashes) == 0 {
		dashes = []C.float{0}
	}
	code := int(C.agg_go_cpp_render_stroke_path_comp(
		img.ptr,
		path.ptr,
		C.float(opts.Width),
		C.int(opts.LineCap),
		C.int(opts.LineJoin),
		C.float(opts.MiterLimit),
		&dashes[0],
		C.int(pairCount),
		C.float(opts.DashStart),
		C.int(blendMode),
		C.int(clip.Min.X),
		C.int(clip.Min.Y),
		C.int(clip.Max.X),
		C.int(clip.Max.Y),
		C.uint8_t(r),
		C.uint8_t(g),
		C.uint8_t(b),
		C.uint8_t(a),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_render_stroke_path_comp failed: %s", cppNativeLastError())
	}
	return nil
}

type cppNativeMatrix struct {
	ptr *C.AggGoCPPMatrix
}

func newCPPNativeMatrix() (*cppNativeMatrix, error) {
	ptr := C.agg_go_cpp_matrix_create()
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_matrix_create failed: %s", cppNativeLastError())
	}
	matrix := &cppNativeMatrix{ptr: ptr}
	runtime.SetFinalizer(matrix, (*cppNativeMatrix).close)
	return matrix, nil
}

func (m *cppNativeMatrix) close() error {
	runtime.SetFinalizer(m, nil)
	if m == nil || m.ptr == nil {
		return nil
	}
	C.agg_go_cpp_matrix_free(m.ptr)
	m.ptr = nil
	return nil
}

func (m *cppNativeMatrix) identity() error {
	if m == nil || m.ptr == nil {
		return fmt.Errorf("matrix is nil")
	}
	if code := int(C.agg_go_cpp_matrix_identity(m.ptr)); code != 0 {
		return fmt.Errorf("agg_go_cpp_matrix_identity failed: %s", cppNativeLastError())
	}
	return nil
}

func (m *cppNativeMatrix) translate(tx, ty float32) error {
	if m == nil || m.ptr == nil {
		return fmt.Errorf("matrix is nil")
	}
	if code := int(C.agg_go_cpp_matrix_translate(m.ptr, C.float(tx), C.float(ty))); code != 0 {
		return fmt.Errorf("agg_go_cpp_matrix_translate failed: %s", cppNativeLastError())
	}
	return nil
}

func (m *cppNativeMatrix) scale(sx, sy float32) error {
	if m == nil || m.ptr == nil {
		return fmt.Errorf("matrix is nil")
	}
	if code := int(C.agg_go_cpp_matrix_scale(m.ptr, C.float(sx), C.float(sy))); code != 0 {
		return fmt.Errorf("agg_go_cpp_matrix_scale failed: %s", cppNativeLastError())
	}
	return nil
}

func (m *cppNativeMatrix) rotate(angle float32) error {
	if m == nil || m.ptr == nil {
		return fmt.Errorf("matrix is nil")
	}
	if code := int(C.agg_go_cpp_matrix_rotate(m.ptr, C.float(angle))); code != 0 {
		return fmt.Errorf("agg_go_cpp_matrix_rotate failed: %s", cppNativeLastError())
	}
	return nil
}

func (m *cppNativeMatrix) rotateDegrees(degrees float32) error {
	return m.rotate(float32(float64(degrees) * math.Pi / 180.0))
}

func (m *cppNativeMatrix) transformPoint(x, y float64) (float64, float64, error) {
	if m == nil || m.ptr == nil {
		return 0, 0, fmt.Errorf("matrix is nil")
	}
	cx := C.double(x)
	cy := C.double(y)
	if code := int(C.agg_go_cpp_matrix_transform_point(m.ptr, &cx, &cy)); code != 0 {
		return 0, 0, fmt.Errorf("agg_go_cpp_matrix_transform_point failed: %s", cppNativeLastError())
	}
	return float64(cx), float64(cy), nil
}

type cppNativeFont struct {
	ptr *C.AggGoCPPFont
}

func newCPPNativeFont(fontPath string) (*cppNativeFont, error) {
	if fontPath == "" {
		return nil, fmt.Errorf("font path is empty")
	}
	cPath := C.CString(fontPath)
	defer C.free(unsafe.Pointer(cPath))
	ptr := C.agg_go_cpp_font_create(cPath)
	if ptr == nil {
		return nil, fmt.Errorf("agg_go_cpp_font_create failed: %s", cppNativeLastError())
	}
	font := &cppNativeFont{ptr: ptr}
	runtime.SetFinalizer(font, (*cppNativeFont).close)
	return font, nil
}

func (f *cppNativeFont) close() error {
	runtime.SetFinalizer(f, nil)
	if f == nil || f.ptr == nil {
		return nil
	}
	C.agg_go_cpp_font_free(f.ptr)
	f.ptr = nil
	return nil
}

func (f *cppNativeFont) setSize(size float32) error {
	if f == nil || f.ptr == nil {
		return fmt.Errorf("font is nil")
	}
	if code := int(C.agg_go_cpp_font_set_size(f.ptr, C.float(size))); code != 0 {
		return fmt.Errorf("agg_go_cpp_font_set_size failed: %s", cppNativeLastError())
	}
	return nil
}

func (f *cppNativeFont) setHinting(enabled bool) error {
	if f == nil || f.ptr == nil {
		return fmt.Errorf("font is nil")
	}
	value := 0
	if enabled {
		value = 1
	}
	if code := int(C.agg_go_cpp_font_set_hinting(f.ptr, C.int(value))); code != 0 {
		return fmt.Errorf("agg_go_cpp_font_set_hinting failed: %s", cppNativeLastError())
	}
	return nil
}

func (f *cppNativeFont) setFlipY(enabled bool) error {
	if f == nil || f.ptr == nil {
		return fmt.Errorf("font is nil")
	}
	value := 0
	if enabled {
		value = 1
	}
	if code := int(C.agg_go_cpp_font_set_flip_y(f.ptr, C.int(value))); code != 0 {
		return fmt.Errorf("agg_go_cpp_font_set_flip_y failed: %s", cppNativeLastError())
	}
	return nil
}

func (f *cppNativeFont) renderText(img *cppNativeImage, text string, x, y float32, r, g, b, a uint8) error {
	if f == nil || f.ptr == nil {
		return fmt.Errorf("font is nil")
	}
	if img == nil || img.ptr == nil {
		return fmt.Errorf("image is nil")
	}
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	code := int(C.agg_go_cpp_render_text(
		img.ptr,
		f.ptr,
		cText,
		C.float(x),
		C.float(y),
		C.uint8_t(r),
		C.uint8_t(g),
		C.uint8_t(b),
		C.uint8_t(a),
	))
	if code != 0 {
		return fmt.Errorf("agg_go_cpp_render_text failed: %s", cppNativeLastError())
	}
	return nil
}

func (f *cppNativeFont) textWidth(text string) float64 {
	if f == nil || f.ptr == nil {
		return 0
	}
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	return float64(C.agg_go_cpp_text_width(f.ptr, cText))
}

func (f *cppNativeFont) textHeight() float64 {
	if f == nil || f.ptr == nil {
		return 0
	}
	return float64(C.agg_go_cpp_text_height(f.ptr))
}
