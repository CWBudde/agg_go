//go:build freetype

// Package freetype2 provides LoadedFace implementation for individual font face management.
package freetype2

/*
#cgo pkg-config: freetype2
#include <ft2build.h>
#include FT_FREETYPE_H
#include <stdlib.h>
#include <string.h>

// Helper functions for LoadedFace operations
static char* get_face_family_name(FT_Face face) {
    return face->family_name;
}

static char* get_face_style_name(FT_Face face) {
    return face->style_name;
}

static int face_is_scalable(FT_Face face) {
    return FT_IS_SCALABLE(face);
}

static int face_has_kerning(FT_Face face) {
    return FT_HAS_KERNING(face);
}

static long face_units_per_em(FT_Face face) {
    return face->units_per_EM;
}

// 26.6 fixed point conversion helpers
static double ft_26dot6_to_double(FT_Pos pos) {
    return (double)pos / 64.0;
}

static FT_F26Dot6 double_to_ft_26dot6(double d) {
    return (FT_F26Dot6)(d * 64.0);
}
*/
import "C"

import (
	"agg_go/internal/basics"
	"agg_go/internal/fonts"
	"agg_go/internal/transform"
	"errors"
	"fmt"
	"unsafe"
)

// LoadedFace represents a loaded font face with instance management capabilities.
// This corresponds to AGG's fman::loaded_face class.
type LoadedFace struct {
	// Reference to parent engine
	engine *FontEngine
	
	// FreeType face handle
	ftFace C.FT_Face
	
	// Face name (dynamically allocated)
	faceName string
	
	// Current instance settings
	dpi        uint32
	height     float64
	width      float64
	rendering  GlyphRendering
	hinting    bool
	flipY      bool
	charMap    CharEncoding
	
	// Transformation matrix
	affine *transform.TransAffine
}

// NewLoadedFace creates a new loaded face wrapper around a FreeType face.
func NewLoadedFace(engine *FontEngine, ftFace C.FT_Face) *LoadedFace {
	lf := &LoadedFace{
		engine:    engine,
		ftFace:    ftFace,
		dpi:       0,
		height:    0,
		width:     0,
		rendering: GlyphRenNativeGray8,
		hinting:   false,
		flipY:     true, // Default to flip Y as in AGG
		charMap:   EncodingNone,
		affine:    transform.NewTransAffine(),
	}
	
	lf.setFaceName()
	return lf
}

// setFaceName sets the face name by combining family and style names.
func (lf *LoadedFace) setFaceName() {
	if lf.ftFace == nil {
		lf.faceName = "unknown"
		return
	}
	
	familyName := C.get_face_family_name(lf.ftFace)
	styleName := C.get_face_style_name(lf.ftFace)
	
	if familyName != nil && styleName != nil {
		lf.faceName = fmt.Sprintf("%s %s", C.GoString(familyName), C.GoString(styleName))
	} else if familyName != nil {
		lf.faceName = C.GoString(familyName)
	} else {
		lf.faceName = "unknown"
	}
}

// NumFaces returns the number of faces in this font file.
func (lf *LoadedFace) NumFaces() uint32 {
	if lf.ftFace == nil {
		return 0
	}
	return uint32(lf.ftFace.num_faces)
}

// Name returns the face name.
func (lf *LoadedFace) Name() string {
	return lf.faceName
}

// Resolution returns the current resolution in DPI.
func (lf *LoadedFace) Resolution() uint32 {
	return lf.dpi
}

// Height returns the current font height.
func (lf *LoadedFace) Height() float64 {
	return lf.height
}

// Width returns the current font width.
func (lf *LoadedFace) Width() float64 {
	return lf.width
}

// Ascent returns the typographic ascender.
func (lf *LoadedFace) Ascent() float64 {
	if lf.ftFace == nil {
		return 0
	}
	return float64(lf.ftFace.ascender) * lf.height / float64(C.face_units_per_em(lf.ftFace))
}

// Descent returns the typographic descender.
func (lf *LoadedFace) Descent() float64 {
	if lf.ftFace == nil {
		return 0
	}
	return float64(lf.ftFace.descender) * lf.height / float64(C.face_units_per_em(lf.ftFace))
}

// AscentB returns the maximum ascender (bounding box).
func (lf *LoadedFace) AscentB() float64 {
	if lf.ftFace == nil {
		return 0
	}
	return float64(lf.ftFace.bbox.yMax) * lf.height / float64(C.face_units_per_em(lf.ftFace))
}

// DescentB returns the maximum descender (bounding box).
func (lf *LoadedFace) DescentB() float64 {
	if lf.ftFace == nil {
		return 0
	}
	return float64(lf.ftFace.bbox.yMin) * lf.height / float64(C.face_units_per_em(lf.ftFace))
}

// Hinting returns whether hinting is enabled.
func (lf *LoadedFace) Hinting() bool {
	return lf.hinting
}

// FlipY returns whether Y coordinates are flipped.
func (lf *LoadedFace) FlipY() bool {
	return lf.flipY
}

// Transform returns the current affine transformation.
func (lf *LoadedFace) Transform() *transform.TransAffine {
	return lf.affine
}

// CharMap returns the current character encoding.
func (lf *LoadedFace) CharMap() CharEncoding {
	return lf.charMap
}

// SelectInstance selects and configures a font instance with the specified parameters.
// This corresponds to AGG's loaded_face::select_instance method.
func (lf *LoadedFace) SelectInstance(height, width float64, hinting bool, rendering GlyphRendering) {
	// Adjust rendering to what this face can support
	rendering = lf.CapableRendering(rendering)
	
	// Only update if parameters have changed
	if lf.height != height || lf.width != width || lf.hinting != hinting || lf.rendering != rendering {
		lf.height = height
		lf.width = width
		lf.hinting = hinting
		lf.rendering = rendering
		lf.updateCharSize()
	}
}

// CapableRendering returns the best rendering mode this face can support.
// This corresponds to AGG's loaded_face::capable_rendering method.
func (lf *LoadedFace) CapableRendering(rendering GlyphRendering) GlyphRendering {
	if lf.ftFace == nil {
		return GlyphRenNativeGray8
	}
	
	switch rendering {
	case GlyphRenNativeMono, GlyphRenNativeGray8:
		// Native bitmap modes are always supported
		return rendering
		
	case GlyphRenOutline:
		// Outline mode requires scalable fonts
		if C.face_is_scalable(lf.ftFace) == 0 {
			return GlyphRenNativeGray8
		}
		return rendering
		
	case GlyphRenAggMono:
		// AGG mono mode requires scalable fonts
		if C.face_is_scalable(lf.ftFace) == 0 {
			return GlyphRenNativeMono
		}
		return rendering
		
	case GlyphRenAggGray8:
		// AGG gray8 mode requires scalable fonts
		if C.face_is_scalable(lf.ftFace) == 0 {
			return GlyphRenNativeGray8
		}
		return rendering
		
	default:
		return GlyphRenNativeGray8
	}
}

// updateCharSize updates the character size in FreeType.
func (lf *LoadedFace) updateCharSize() {
	if lf.ftFace == nil {
		return
	}
	
	if lf.dpi > 0 {
		// Use resolution-based sizing
		C.FT_Set_Char_Size(lf.ftFace,
			C.double_to_ft_26dot6(C.double(lf.width)),
			C.double_to_ft_26dot6(C.double(lf.height)),
			C.FT_UInt(lf.dpi),
			C.FT_UInt(lf.dpi))
	} else {
		// Use pixel-based sizing
		C.FT_Set_Pixel_Sizes(lf.ftFace,
			C.FT_UInt(lf.width),
			C.FT_UInt(lf.height))
	}
}

// SetHinting enables or disables font hinting.
func (lf *LoadedFace) SetHinting(hinting bool) {
	lf.hinting = hinting
}

// SetFlipY sets whether to flip Y coordinates.
func (lf *LoadedFace) SetFlipY(flipY bool) {
	lf.flipY = flipY
}

// SetTransform sets the affine transformation matrix.
func (lf *LoadedFace) SetTransform(affine *transform.TransAffine) {
	lf.affine = affine
}

// SetCharMap sets the character encoding map.
func (lf *LoadedFace) SetCharMap(encoding CharEncoding) error {
	if lf.ftFace == nil {
		return errors.New("no face loaded")
	}
	
	var ftEncoding C.FT_Encoding
	switch encoding {
	case EncodingNone:
		ftEncoding = C.FT_ENCODING_NONE
	case EncodingUnicode:
		ftEncoding = C.FT_ENCODING_UNICODE
	case EncodingSymbol:
		ftEncoding = C.FT_ENCODING_MS_SYMBOL
	case EncodingAdobeLatin1:
		ftEncoding = C.FT_ENCODING_ADOBE_LATIN_1
	case EncodingAdobeCustom:
		ftEncoding = C.FT_ENCODING_ADOBE_CUSTOM
	case EncodingAdobeExpert:
		ftEncoding = C.FT_ENCODING_ADOBE_EXPERT
	default:
		return fmt.Errorf("unsupported encoding: %d", encoding)
	}
	
	err := C.FT_Select_Charmap(lf.ftFace, ftEncoding)
	if err != 0 {
		return fmt.Errorf("failed to set charmap: FreeType error %d", err)
	}
	
	lf.charMap = encoding
	return nil
}

// PrepareGlyph prepares a glyph for rendering and returns the prepared glyph data.
// This corresponds to AGG's loaded_face::prepare_glyph method.
func (lf *LoadedFace) PrepareGlyph(glyphCode uint32) (*PreparedGlyph, bool) {
	if lf.ftFace == nil {
		return nil, false
	}
	
	// Get glyph index
	glyphIndex := uint32(C.FT_Get_Char_Index(lf.ftFace, C.FT_ULong(glyphCode)))
	if glyphIndex == 0 {
		return nil, false
	}
	
	// Load glyph
	loadFlags := C.FT_LOAD_DEFAULT
	if !lf.hinting {
		loadFlags |= C.FT_LOAD_NO_HINTING
	}
	
	err := C.FT_Load_Glyph(lf.ftFace, C.FT_UInt(glyphIndex), C.FT_Int32(loadFlags))
	if err != 0 {
		return nil, false
	}
	
	glyph := lf.ftFace.glyph
	prepared := &PreparedGlyph{
		GlyphCode:  glyphCode,
		GlyphIndex: glyphIndex,
	}
	
	// Set bounds and advance
	prepared.Bounds = basics.Rect[int]{
		X1: int(glyph.bitmap_left),
		Y1: int(int(glyph.bitmap_top) - int(glyph.bitmap.rows)),
		X2: int(int(glyph.bitmap_left) + int(glyph.bitmap.width)),
		Y2: int(glyph.bitmap_top),
	}
	
	prepared.AdvanceX = float64(C.ft_26dot6_to_double(glyph.advance.x))
	prepared.AdvanceY = float64(C.ft_26dot6_to_double(glyph.advance.y))
	
	// Determine data type and size based on rendering mode
	switch lf.rendering {
	case GlyphRenOutline:
		prepared.DataType = fonts.FmanGlyphDataOutline
		prepared.DataSize = 0 // Outline data is handled separately
		
		// Decompose outline if available
		if glyph.format == C.FT_GLYPH_FORMAT_OUTLINE {
			// TODO: Decompose outline to path storage
			// This would be handled by the engine's path storage
		}
		
	case GlyphRenNativeGray8, GlyphRenAggGray8:
		prepared.DataType = fonts.FmanGlyphDataGray8
		
		// Render to bitmap if not already done
		if glyph.format != C.FT_GLYPH_FORMAT_BITMAP {
			err = C.FT_Render_Glyph(glyph, C.FT_RENDER_MODE_NORMAL)
			if err != 0 {
				return nil, false
			}
		}
		prepared.DataSize = uint32(int(glyph.bitmap.rows) * int(glyph.bitmap.pitch))
		
	case GlyphRenNativeMono, GlyphRenAggMono:
		prepared.DataType = fonts.FmanGlyphDataMono
		
		// Render to monochrome bitmap
		if glyph.format != C.FT_GLYPH_FORMAT_BITMAP {
			err = C.FT_Render_Glyph(glyph, C.FT_RENDER_MODE_MONO)
			if err != 0 {
				return nil, false
			}
		}
		prepared.DataSize = uint32(int(glyph.bitmap.rows) * int(glyph.bitmap.pitch))
		
	default:
		prepared.DataType = fonts.FmanGlyphDataInvalid
		prepared.DataSize = 0
	}
	
	return prepared, true
}

// AddKerning adds kerning offset between two glyphs.
// This corresponds to AGG's loaded_face::add_kerning method.
func (lf *LoadedFace) AddKerning(first, second uint32) (dx, dy float64) {
	if lf.ftFace == nil || C.face_has_kerning(lf.ftFace) == 0 {
		return 0, 0
	}
	
	var delta C.FT_Vector
	err := C.FT_Get_Kerning(lf.ftFace, C.FT_UInt(first), C.FT_UInt(second),
		C.FT_KERNING_DEFAULT, &delta)
	if err != 0 {
		return 0, 0
	}
	
	dx = float64(C.ft_26dot6_to_double(delta.x))
	dy = float64(C.ft_26dot6_to_double(delta.y))
	return dx, dy
}

// WriteGlyphTo writes the current glyph bitmap data to the provided buffer.
// This corresponds to AGG's loaded_face::write_glyph_to method.
func (lf *LoadedFace) WriteGlyphTo(prepared *PreparedGlyph, data []byte) {
	if lf.ftFace == nil || prepared.DataSize == 0 {
		return
	}
	
	glyph := lf.ftFace.glyph
	
	switch prepared.DataType {
	case fonts.FmanGlyphDataGray8:
		bitmap := &glyph.bitmap
		if bitmap.buffer != nil && len(data) >= int(prepared.DataSize) {
			srcData := unsafe.Slice((*byte)(bitmap.buffer), prepared.DataSize)
			copy(data, srcData)
		}
		
	case fonts.FmanGlyphDataMono:
		bitmap := &glyph.bitmap
		if bitmap.buffer != nil && len(data) >= int(prepared.DataSize) {
			srcData := unsafe.Slice((*byte)(bitmap.buffer), prepared.DataSize)
			copy(data, srcData)
		}
	}
}

// Close cleans up resources used by this loaded face.
func (lf *LoadedFace) Close() error {
	if lf.ftFace != nil {
		C.FT_Done_Face(lf.ftFace)
		lf.ftFace = nil
	}
	lf.faceName = ""
	return nil
}