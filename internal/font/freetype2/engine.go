//go:build freetype

// Package freetype2 provides the main FontEngine implementation with enhanced multi-face support.
package freetype2

/*
#cgo pkg-config: freetype2
#include <ft2build.h>
#include FT_FREETYPE_H
#include <stdlib.h>

// Helper functions for the main engine
static FT_Library* new_library() {
    return (FT_Library*)malloc(sizeof(FT_Library));
}

static void free_library(FT_Library* lib) {
    free(lib);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"

	"agg_go/internal/font"
	"agg_go/internal/path"
	"agg_go/internal/transform"
)

// FontEngine is the main FreeType2 font engine with enhanced multi-face support.
// This corresponds to AGG's fman::font_engine_freetype_base class.
type FontEngine struct {
	*FontEngineBase

	// FreeType library handle
	library *C.FT_Library

	// Face management.
	// Unlike the earlier Go port, AGG does not mirror FT_Face objects in a
	// separate engine-owned C array. Track loaded faces directly here.
	loadedFaces []*LoadedFace
}

// NewFontEngine creates a new FreeType2 font engine.
// flag32 determines whether to use 32-bit or 16-bit precision for path storage.
// maxFaces specifies a Go-only policy limit for simultaneously tracked faces.
// AGG tracks loaded_face objects too, but does not expose a matching configurable cap.
func NewFontEngine(flag32 bool, maxFaces uint32, ftMemory unsafe.Pointer) (*FontEngine, error) {
	if maxFaces == 0 {
		maxFaces = 32 // Go-side default policy; AGG does not expose this knob
	}

	engine := &FontEngine{
		FontEngineBase: NewFontEngineBase(flag32, maxFaces),
		loadedFaces:    make([]*LoadedFace, 0, maxFaces),
	}

	// Initialize FreeType library
	engine.library = C.new_library()
	if engine.library == nil {
		return nil, errors.New("failed to allocate FreeType library handle")
	}

	var err C.FT_Error
	if ftMemory != nil {
		// TODO: Support custom memory management if needed
		// This would require implementing FreeType's memory interface
		err = C.FT_Init_FreeType(engine.library)
	} else {
		err = C.FT_Init_FreeType(engine.library)
	}

	if err != 0 {
		C.free_library(engine.library)
		return nil, fmt.Errorf("failed to initialize FreeType library: error %d", err)
	}

	engine.libraryInitialized = true

	return engine, nil
}

// LoadFace loads a font face from memory buffer.
// This corresponds to AGG's font_engine_freetype_base::load_face method.
func (fe *FontEngine) LoadFace(buffer []byte, bytes uint) (LoadedFaceInterface, error) {
	if uint32(len(fe.loadedFaces)) >= fe.maxFaces {
		return nil, errors.New("maximum number of faces exceeded")
	}

	if len(buffer) == 0 || bytes == 0 {
		return nil, errors.New("invalid buffer or size")
	}

	var ftFace C.FT_Face
	err := C.FT_New_Memory_Face(*fe.library,
		(*C.FT_Byte)(unsafe.Pointer(&buffer[0])),
		C.FT_Long(bytes),
		0, // face_index - could be parameterized for TTC files
		&ftFace)

	if err != 0 {
		fe.lastError = fmt.Errorf("failed to load font face from memory: FreeType error %d", err)
		return nil, fe.lastError
	}

	return fe.createLoadedFace(ftFace)
}

// LoadFaceFile loads a font face from a file.
// This corresponds to AGG's font_engine_freetype_base::load_face_file method.
func (fe *FontEngine) LoadFaceFile(fileName string) (LoadedFaceInterface, error) {
	if uint32(len(fe.loadedFaces)) >= fe.maxFaces {
		return nil, errors.New("maximum number of faces exceeded")
	}

	if fileName == "" {
		return nil, errors.New("invalid file name")
	}

	cFileName := C.CString(fileName)
	defer C.free(unsafe.Pointer(cFileName))

	var ftFace C.FT_Face
	err := C.FT_New_Face(*fe.library, cFileName, 0, &ftFace)

	if err != 0 {
		fe.lastError = fmt.Errorf("failed to load font face from file %s: FreeType error %d", fileName, err)
		return nil, fe.lastError
	}

	return fe.createLoadedFace(ftFace)
}

// createLoadedFace creates a LoadedFace wrapper and stores it in the engine.
// This corresponds to AGG's font_engine_freetype_base::create_loaded_face method.
func (fe *FontEngine) createLoadedFace(ftFace C.FT_Face) (*LoadedFace, error) {
	// Create the loaded face wrapper
	loadedFace := NewLoadedFace(fe, ftFace)

	// Store the loaded face wrapper
	fe.loadedFaces = append(fe.loadedFaces, loadedFace)

	return loadedFace, nil
}

// UnloadFace removes a loaded face from the engine.
// This corresponds to AGG's font_engine_freetype_base::unload_face method.
func (fe *FontEngine) UnloadFace(face LoadedFaceInterface) error {
	loadedFace, ok := face.(*LoadedFace)
	if !ok {
		return errors.New("invalid face type")
	}
	return fe.closeLoadedFace(loadedFace, true)
}

// SetGamma sets gamma correction for the rasterizer.
// This corresponds to AGG's template<class GammaF> void gamma(const GammaF& f).
func (fe *FontEngine) SetGamma(gamma float64) {
	// Apply gamma to the base engine's rasterizer
	fe.FontEngineBase.SetGamma(gamma)
}

// pathStorage16ForTests exposes the internal 16-bit path storage for package-local tests.
func (fe *FontEngine) pathStorage16ForTests() *path.PathStorageInteger[int16] {
	return fe.pathStorage16
}

// pathStorage32ForTests exposes the internal 32-bit path storage for package-local tests.
func (fe *FontEngine) pathStorage32ForTests() *path.PathStorageInteger[int32] {
	return fe.pathStorage32
}

// pathStorageForTests exposes the active internal path storage for package-local tests.
func (fe *FontEngine) pathStorageForTests() font.IntegerPathStorage {
	if fe.flag32 {
		return fe.pathStorage32
	}
	return fe.pathStorage16
}

// DecomposeFTOutline decomposes a FreeType outline into AGG path commands.
// This is a key method that converts vector font outlines to AGG's path format.
// It corresponds to AGG's decompose_ft_outline function.
func (fe *FontEngine) DecomposeFTOutline(outline *C.FT_Outline, flipY bool, affine *transform.TransAffine) error {
	if outline == nil || outline.n_contours <= 0 {
		return nil // Empty outline is valid
	}

	// Clear the appropriate path storage
	var pathStorage font.IntegerPathStorage
	if fe.flag32 {
		fe.pathStorage32.RemoveAll()
		pathStorage = fe.pathStorage32
	} else {
		fe.pathStorage16.RemoveAll()
		pathStorage = fe.pathStorage16
	}

	return fe.decomposeOutlineToPath(outline, flipY, affine, pathStorage)
}

// decomposeOutlineToPath performs the actual outline decomposition.
// This implements the complex FreeType outline walking algorithm from AGG.
func (fe *FontEngine) decomposeOutlineToPath(outline *C.FT_Outline, flipY bool, affine *transform.TransAffine, pathStorage font.IntegerPathStorage) error {
	first := 0

	for n := 0; n < int(outline.n_contours); n++ {
		// Get contour endpoints
		lastPtr := uintptr(unsafe.Pointer(outline.contours)) + uintptr(n)*unsafe.Sizeof(C.short(0))
		last := int(*(*C.short)(unsafe.Pointer(lastPtr)))

		if first > last {
			return errors.New("invalid contour bounds")
		}

		// Get starting points
		vStartPtr := uintptr(unsafe.Pointer(outline.points)) + uintptr(first)*unsafe.Sizeof(C.FT_Vector{})
		vStart := (*C.FT_Vector)(unsafe.Pointer(vStartPtr))

		vLastPtr := uintptr(unsafe.Pointer(outline.points)) + uintptr(last)*unsafe.Sizeof(C.FT_Vector{})
		vLast := (*C.FT_Vector)(unsafe.Pointer(vLastPtr))

		// Check tag of first point
		firstTagPtr := uintptr(unsafe.Pointer(outline.tags)) + uintptr(first)
		firstTag := *(*C.char)(unsafe.Pointer(firstTagPtr))

		// A contour cannot start with a cubic control point
		if (int(firstTag) & 3) == 3 { // FT_CURVE_TAG_CUBIC
			return errors.New("contour cannot start with cubic control point")
		}

		// Determine starting point
		startPoint := vStart
		if (int(firstTag) & 1) == 0 { // FT_CURVE_TAG_CONIC
			// First point is conic control - check last point
			lastTagPtr := uintptr(unsafe.Pointer(outline.tags)) + uintptr(last)
			lastTag := *(*C.char)(unsafe.Pointer(lastTagPtr))

			if int(lastTag)&1 == 1 { // FT_CURVE_TAG_ON
				// Use last point as start
				startPoint = vLast
			} else {
				// Both first and last are conic - use middle point
				middlePoint := C.FT_Vector{
					x: (vStart.x + vLast.x) / 2,
					y: (vStart.y + vLast.y) / 2,
				}
				startPoint = &middlePoint
			}
		}

		// Convert and move to starting point
		x, y := transformFTPoint(startPoint.x, startPoint.y, flipY, affine)

		// Move to start point using the interface method
		pathStorage.MoveTo64(int64(x), int64(y))

		// Process the contour points
		err := fe.processContourPoints(outline, first, last, flipY, affine, pathStorage, startPoint)
		if err != nil {
			return err
		}

		// Close the polygon
		pathStorage.ClosePolygon()

		first = last + 1
	}

	return nil
}

// processContourPoints processes the points in a single contour.
func (fe *FontEngine) processContourPoints(outline *C.FT_Outline, first, last int, flipY bool,
	affine *transform.TransAffine,
	pathStorage font.IntegerPathStorage, startPoint *C.FT_Vector,
) error {
	i := first
	for i < last {
		i++

		// Get current point and tag
		pointPtr := uintptr(unsafe.Pointer(outline.points)) + uintptr(i)*unsafe.Sizeof(C.FT_Vector{})
		point := (*C.FT_Vector)(unsafe.Pointer(pointPtr))

		tagPtr := uintptr(unsafe.Pointer(outline.tags)) + uintptr(i)
		tag := int(*(*C.char)(unsafe.Pointer(tagPtr))) & 3

		switch tag {
		case 1: // FT_CURVE_TAG_ON - straight line
			x, y := transformFTPoint(point.x, point.y, flipY, affine)

			pathStorage.LineTo64(int64(x), int64(y))

		case 0: // FT_CURVE_TAG_CONIC - quadratic curve
			err := fe.processConicCurve(outline, &i, last, flipY, affine, pathStorage, startPoint)
			if err != nil {
				return err
			}

		default: // FT_CURVE_TAG_CUBIC - cubic curve
			err := fe.processCubicCurve(outline, &i, last, flipY, affine, pathStorage, startPoint)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// processConicCurve handles quadratic Bézier curves.
func (fe *FontEngine) processConicCurve(outline *C.FT_Outline, i *int, last int, flipY bool,
	affine *transform.TransAffine,
	pathStorage font.IntegerPathStorage, startPoint *C.FT_Vector,
) error {
	// Get control point
	controlPtr := uintptr(unsafe.Pointer(outline.points)) + uintptr(*i)*unsafe.Sizeof(C.FT_Vector{})
	vControl := (*C.FT_Vector)(unsafe.Pointer(controlPtr))

	for {
		if *i >= last {
			break
		}

		*i++
		nextPtr := uintptr(unsafe.Pointer(outline.points)) + uintptr(*i)*unsafe.Sizeof(C.FT_Vector{})
		nextPoint := (*C.FT_Vector)(unsafe.Pointer(nextPtr))

		nextTagPtr := uintptr(unsafe.Pointer(outline.tags)) + uintptr(*i)
		nextTag := int(*(*C.char)(unsafe.Pointer(nextTagPtr))) & 3

		if nextTag == 1 { // FT_CURVE_TAG_ON - end of curve
			x1, y1 := transformFTPoint(vControl.x, vControl.y, flipY, affine)
			x2, y2 := transformFTPoint(nextPoint.x, nextPoint.y, flipY, affine)

			pathStorage.Curve3_64(int64(x1), int64(y1), int64(x2), int64(y2))
			break
		}

		if nextTag != 0 { // Not conic
			return errors.New("invalid curve sequence")
		}

		// Multiple conic points - create intermediate curve
		vMiddle := C.FT_Vector{
			x: (vControl.x + nextPoint.x) / 2,
			y: (vControl.y + nextPoint.y) / 2,
		}

		x1, y1 := transformFTPoint(vControl.x, vControl.y, flipY, affine)
		x2, y2 := transformFTPoint(vMiddle.x, vMiddle.y, flipY, affine)

		pathStorage.Curve3_64(int64(x1), int64(y1), int64(x2), int64(y2))

		vControl = nextPoint
	}

	// Handle curve back to start if needed
	if *i >= last {
		x1, y1 := transformFTPoint(vControl.x, vControl.y, flipY, affine)
		x2, y2 := transformFTPoint(startPoint.x, startPoint.y, flipY, affine)

		pathStorage.Curve3_64(int64(x1), int64(y1), int64(x2), int64(y2))
	}

	return nil
}

// processCubicCurve handles cubic Bézier curves.
func (fe *FontEngine) processCubicCurve(outline *C.FT_Outline, i *int, last int, flipY bool,
	affine *transform.TransAffine,
	pathStorage font.IntegerPathStorage, startPoint *C.FT_Vector,
) error {
	if *i+1 > last {
		return errors.New("insufficient points for cubic curve")
	}

	// Get first control point
	ctrl1Ptr := uintptr(unsafe.Pointer(outline.points)) + uintptr(*i)*unsafe.Sizeof(C.FT_Vector{})
	vCtrl1 := (*C.FT_Vector)(unsafe.Pointer(ctrl1Ptr))

	*i++

	// Get second control point
	ctrl2Ptr := uintptr(unsafe.Pointer(outline.points)) + uintptr(*i)*unsafe.Sizeof(C.FT_Vector{})
	vCtrl2 := (*C.FT_Vector)(unsafe.Pointer(ctrl2Ptr))

	// Verify second control point is cubic
	tag2Ptr := uintptr(unsafe.Pointer(outline.tags)) + uintptr(*i)
	tag2 := int(*(*C.char)(unsafe.Pointer(tag2Ptr))) & 3
	if tag2 != 3 {
		return errors.New("second cubic control point has wrong tag")
	}

	var endPoint *C.FT_Vector
	if *i < last {
		*i++
		endPtr := uintptr(unsafe.Pointer(outline.points)) + uintptr(*i)*unsafe.Sizeof(C.FT_Vector{})
		endPoint = (*C.FT_Vector)(unsafe.Pointer(endPtr))
	} else {
		// Curve back to start
		endPoint = startPoint
	}

	x1, y1 := transformFTPoint(vCtrl1.x, vCtrl1.y, flipY, affine)
	x2, y2 := transformFTPoint(vCtrl2.x, vCtrl2.y, flipY, affine)
	x3, y3 := transformFTPoint(endPoint.x, endPoint.y, flipY, affine)

	pathStorage.Curve4_64(int64(x1), int64(y1), int64(x2), int64(y2), int64(x3), int64(y3))

	return nil
}

func transformFTPoint(x, y C.FT_Pos, flipY bool, affine *transform.TransAffine) (float64, float64) {
	fx := float64(x) / 64.0
	fy := float64(y) / 64.0
	if flipY {
		fy = -fy
	}
	if affine != nil {
		affine.Transform(&fx, &fy)
	}
	return fx, fy
}

func (fe *FontEngine) closeLoadedFace(face *LoadedFace, releaseFTFace bool) error {
	if face == nil {
		return nil
	}

	removed := false
	for i, lf := range fe.loadedFaces {
		if lf == face {
			fe.loadedFaces = append(fe.loadedFaces[:i], fe.loadedFaces[i+1:]...)
			removed = true
			break
		}
	}

	if !removed && face.engine == fe {
		return errors.New("face not found in engine")
	}

	face.engine = nil
	if releaseFTFace && face.ftFace != nil {
		C.FT_Done_Face(face.ftFace)
		face.ftFace = nil
	}
	face.faceName = ""
	face.closed = true
	return nil
}

// Close cleans up all resources used by the font engine.
// Unlike AGG's font_engine_freetype_base destructor, this Go port also closes
// any still-tracked faces before releasing the FreeType library so the engine
// can provide deterministic, idempotent teardown under Go ownership.
func (fe *FontEngine) Close() error {
	// Clean up loaded faces
	for len(fe.loadedFaces) > 0 {
		face := fe.loadedFaces[len(fe.loadedFaces)-1]
		if face != nil {
			_ = fe.closeLoadedFace(face, true)
			continue
		}
		fe.loadedFaces = fe.loadedFaces[:len(fe.loadedFaces)-1]
	}
	fe.loadedFaces = nil

	// Clean up FreeType library
	if fe.libraryInitialized && fe.library != nil {
		C.FT_Done_FreeType(*fe.library)
		C.free_library(fe.library)
		fe.library = nil
		fe.libraryInitialized = false
	}

	// Clean up base resources
	return fe.FontEngineBase.Close()
}
