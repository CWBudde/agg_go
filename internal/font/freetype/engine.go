//go:build freetype

// Package freetype provides FreeType font engine integration for AGG graphics library.
// This requires FreeType library to be installed and uses CGO for integration.
package freetype

/*
#cgo pkg-config: freetype2
#include <ft2build.h>
#include FT_FREETYPE_H
#include <stdlib.h>
#include <string.h>

// Helper functions to work around CGO limitations
static FT_Library* new_library() {
    return (FT_Library*)malloc(sizeof(FT_Library));
}

static void free_library(FT_Library* lib) {
    free(lib);
}

static FT_Face* new_face_array(int size) {
    return (FT_Face*)calloc(size, sizeof(FT_Face));
}

static void free_face_array(FT_Face* faces) {
    free(faces);
}

static char** new_name_array(int size) {
    return (char**)calloc(size, sizeof(char*));
}

static void free_name_array(char** names, int size) {
    int i;
    for (i = 0; i < size; i++) {
        if (names[i]) free(names[i]);
    }
    free(names);
}

static void set_name_in_array(char** names, int index, const char* name) {
    names[index] = strdup(name);
}

static char* get_name_from_array(char** names, int index) {
    return names[index];
}

static FT_Face get_face_from_array(FT_Face* faces, int index) {
    return faces[index];
}

static void set_face_in_array(FT_Face* faces, int index, FT_Face face) {
    faces[index] = face;
}

static int has_kerning(FT_Face face) {
    return FT_HAS_KERNING(face);
}

// Helper functions for 26.6 fixed point conversions
static double int26p6_to_dbl(long p) {
    return (double)p / 64.0;
}

static long dbl_to_int26p6(double p) {
    return (long)(p * 64.0 + 0.5);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"

	"agg_go/internal/basics"
	"agg_go/internal/font"
	"agg_go/internal/path"
	"agg_go/internal/transform"
)

// CRC32 table for font signature generation (AUTODIN II polynomial)
var crc32Table = [256]uint32{
	0x00000000, 0x77073096, 0xee0e612c, 0x990951ba,
	0x076dc419, 0x706af48f, 0xe963a535, 0x9e6495a3,
	0x0edb8832, 0x79dcb8a4, 0xe0d5e91e, 0x97d2d988,
	0x09b64c2b, 0x7eb17cbd, 0xe7b82d07, 0x90bf1d91,
	0x1db71064, 0x6ab020f2, 0xf3b97148, 0x84be41de,
	0x1adad47d, 0x6ddde4eb, 0xf4d4b551, 0x83d385c7,
	0x136c9856, 0x646ba8c0, 0xfd62f97a, 0x8a65c9ec,
	0x14015c4f, 0x63066cd9, 0xfa0f3d63, 0x8d080df5,
	0x3b6e20c8, 0x4c69105e, 0xd56041e4, 0xa2677172,
	0x3c03e4d1, 0x4b04d447, 0xd20d85fd, 0xa50ab56b,
	0x35b5a8fa, 0x42b2986c, 0xdbbbc9d6, 0xacbcf940,
	0x32d86ce3, 0x45df5c75, 0xdcd60dcf, 0xabd13d59,
	0x26d930ac, 0x51de003a, 0xc8d75180, 0xbfd06116,
	0x21b4f4b5, 0x56b3c423, 0xcfba9599, 0xb8bda50f,
	0x2802b89e, 0x5f058808, 0xc60cd9b2, 0xb10be924,
	0x2f6f7c87, 0x58684c11, 0xc1611dab, 0xb6662d3d,
	0x76dc4190, 0x01db7106, 0x98d220bc, 0xefd5102a,
	0x71b18589, 0x06b6b51f, 0x9fbfe4a5, 0xe8b8d433,
	0x7807c9a2, 0x0f00f934, 0x9609a88e, 0xe10e9818,
	0x7f6a0dbb, 0x086d3d2d, 0x91646c97, 0xe6635c01,
	0x6b6b51f4, 0x1c6c6162, 0x856530d8, 0xf262004e,
	0x6c0695ed, 0x1b01a57b, 0x8208f4c1, 0xf50fc457,
	0x65b0d9c6, 0x12b7e950, 0x8bbeb8ea, 0xfcb9887c,
	0x62dd1ddf, 0x15da2d49, 0x8cd37cf3, 0xfbd44c65,
	0x4db26158, 0x3ab551ce, 0xa3bc0074, 0xd4bb30e2,
	0x4adfa541, 0x3dd895d7, 0xa4d1c46d, 0xd3d6f4fb,
	0x4369e96a, 0x346ed9fc, 0xad678846, 0xda60b8d0,
	0x44042d73, 0x33031de5, 0xaa0a4c5f, 0xdd0d7cc9,
	0x5005713c, 0x270241aa, 0xbe0b1010, 0xc90c2086,
	0x5768b525, 0x206f85b3, 0xb966d409, 0xce61e49f,
	0x5edef90e, 0x29d9c998, 0xb0d09822, 0xc7d7a8b4,
	0x59b33d17, 0x2eb40d81, 0xb7bd5c3b, 0xc0ba6cad,
	0xedb88320, 0x9abfb3b6, 0x03b6e20c, 0x74b1d29a,
	0xead54739, 0x9dd277af, 0x04db2615, 0x73dc1683,
	0xe3630b12, 0x94643b84, 0x0d6d6a3e, 0x7a6a5aa8,
	0xe40ecf0b, 0x9309ff9d, 0x0a00ae27, 0x7d079eb1,
	0xf00f9344, 0x8708a3d2, 0x1e01f268, 0x6906c2fe,
	0xf762575d, 0x806567cb, 0x196c3671, 0x6e6b06e7,
	0xfed41b76, 0x89d32be0, 0x10da7a5a, 0x67dd4acc,
	0xf9b9df6f, 0x8ebeeff9, 0x17b7be43, 0x60b08ed5,
	0xd6d6a3e8, 0xa1d1937e, 0x38d8c2c4, 0x4fdff252,
	0xd1bb67f1, 0xa6bc5767, 0x3fb506dd, 0x48b2364b,
	0xd80d2bda, 0xaf0a1b4c, 0x36034af6, 0x41047a60,
	0xdf60efc3, 0xa867df55, 0x316e8eef, 0x4669be79,
	0xcb61b38c, 0xbc66831a, 0x256fd2a0, 0x5268e236,
	0xcc0c7795, 0xbb0b4703, 0x220216b9, 0x5505262f,
	0xc5ba3bbe, 0xb2bd0b28, 0x2bb45a92, 0x5cb36a04,
	0xc2d7ffa7, 0xb5d0cf31, 0x2cd99e8b, 0x5bdeae1d,
	0x9b64c2b0, 0xec63f226, 0x756aa39c, 0x026d930a,
	0x9c0906a9, 0xeb0e363f, 0x72076785, 0x05005713,
	0x95bf4a82, 0xe2b87a14, 0x7bb12bae, 0x0cb61b38,
	0x92d28e9b, 0xe5d5be0d, 0x7cdcefb7, 0x0bdbdf21,
	0x86d3d2d4, 0xf1d4e242, 0x68ddb3f8, 0x1fda836e,
	0x81be16cd, 0xf6b9265b, 0x6fb077e1, 0x18b74777,
	0x88085ae6, 0xff0f6a70, 0x66063bca, 0x11010b5c,
	0x8f659eff, 0xf862ae69, 0x616bffd3, 0x166ccf45,
	0xa00ae278, 0xd70dd2ee, 0x4e048354, 0x3903b3c2,
	0xa7672661, 0xd06016f7, 0x4969474d, 0x3e6e77db,
	0xaed16a4a, 0xd9d65adc, 0x40df0b66, 0x37d83bf0,
	0xa9bcae53, 0xdebb9ec5, 0x47b2cf7f, 0x30b5ffe9,
	0xbdbdf21c, 0xcabac28a, 0x53b39330, 0x24b4a3a6,
	0xbad03605, 0xcdd70693, 0x54de5729, 0x23d967bf,
	0xb3667a2e, 0xc4614ab8, 0x5d681b02, 0x2a6f2b94,
	0xb40bbe37, 0xc30c8ea1, 0x5a05df1b, 0x2d02ef8d,
}

// calcCRC32 calculates CRC32 checksum for the given data.
func calcCRC32(data []byte) uint32 {
	crc := uint32(0xFFFFFFFF)
	for _, b := range data {
		crc = (crc >> 8) ^ crc32Table[(crc^uint32(b))&0xFF]
	}
	return ^crc
}

// FontEngineFreetype implements the FontEngine interface using FreeType.
type FontEngineFreetype struct {
	// Configuration
	flag32             bool
	changeStamp        int
	lastError          int
	name               string
	nameLen            uint
	faceIndex          uint
	charMap            C.FT_Encoding
	signature          string
	height             uint
	width              uint
	hinting            bool
	flipY              bool
	libraryInitialized bool
	resolution         int
	glyphRendering     GlyphRenderingType
	affine             *transform.TransAffine

	// FreeType handles
	library     *C.FT_Library
	faces       *C.FT_Face // Array of font faces
	faceNames   **C.char   // Array of face name strings
	numFaces    uint
	maxFaces    uint
	currentFace C.FT_Face

	// Current glyph information
	glyphIndex uint
	dataSize   uint
	dataType   font.GlyphDataType
	bounds     basics.Rect[int]
	advanceX   float64
	advanceY   float64

	// Path storage for outline fonts
	pathStorage *path.PathStorageStl
}

// GlyphRenderingType defines how glyphs should be rendered.
type GlyphRenderingType int

const (
	GlyphRenderingNative GlyphRenderingType = iota
	GlyphRenderingOutline
	GlyphRenderingAAGray8
	GlyphRenderingAAMono
	GlyphRenderingMono
)

// NewFontEngineFreetype creates a new FreeType font engine.
func NewFontEngineFreetype(flag32 bool, maxFaces uint) (*FontEngineFreetype, error) {
	if maxFaces == 0 {
		maxFaces = 32
	}

	engine := &FontEngineFreetype{
		flag32:      flag32,
		maxFaces:    maxFaces,
		resolution:  72, // Default DPI
		hinting:     true,
		flipY:       false,
		pathStorage: path.NewPathStorageStl(),
		affine:      transform.NewTransAffine(),
	}

	// Initialize FreeType library
	engine.library = C.new_library()
	if C.FT_Init_FreeType(engine.library) != 0 {
		C.free_library(engine.library)
		return nil, errors.New("failed to initialize FreeType library")
	}

	engine.libraryInitialized = true

	// Allocate face arrays
	engine.faces = C.new_face_array(C.int(maxFaces))
	engine.faceNames = C.new_name_array(C.int(maxFaces))

	engine.updateSignature()
	return engine, nil
}

// Close cleans up FreeType resources.
func (fe *FontEngineFreetype) Close() error {
	if fe.libraryInitialized {
		// Clean up faces
		for i := uint(0); i < fe.numFaces; i++ {
			face := C.get_face_from_array(fe.faces, C.int(i))
			if face != nil {
				C.FT_Done_Face(face)
			}
		}

		C.free_face_array(fe.faces)
		C.free_name_array(fe.faceNames, C.int(fe.maxFaces))

		C.FT_Done_FreeType(*fe.library)
		C.free_library(fe.library)
		fe.libraryInitialized = false
	}
	return nil
}

// SetResolution sets the font rendering resolution in DPI.
func (fe *FontEngineFreetype) SetResolution(dpi uint) {
	fe.resolution = int(dpi)
	fe.updateCharSize()
}

// LoadFont loads a font from file or memory.
func (fe *FontEngineFreetype) LoadFont(fontName string, faceIndex uint, renType GlyphRenderingType,
	fontMem []byte,
) error {
	var face C.FT_Face
	var err C.FT_Error

	fe.glyphRendering = renType

	if len(fontMem) > 0 {
		// Load from memory
		err = C.FT_New_Memory_Face(*fe.library,
			(*C.FT_Byte)(unsafe.Pointer(&fontMem[0])),
			C.FT_Long(len(fontMem)),
			C.FT_Long(faceIndex),
			&face)
	} else {
		// Load from file
		cFontName := C.CString(fontName)
		defer C.free(unsafe.Pointer(cFontName))
		err = C.FT_New_Face(*fe.library, cFontName, C.FT_Long(faceIndex), &face)
	}

	if err != 0 {
		fe.lastError = int(err)
		return fmt.Errorf("failed to load font %s: FreeType error %d", fontName, err)
	}

	// Store the face
	if fe.numFaces >= fe.maxFaces {
		return errors.New("maximum number of faces exceeded")
	}

	C.set_face_in_array(fe.faces, C.int(fe.numFaces), face)
	C.set_name_in_array(fe.faceNames, C.int(fe.numFaces), C.CString(fontName))
	fe.numFaces++

	fe.currentFace = face
	fe.faceIndex = faceIndex
	fe.name = fontName
	fe.nameLen = uint(len(fontName))

	// Set character map to Unicode
	fe.charMap = C.FT_ENCODING_UNICODE
	C.FT_Select_Charmap(fe.currentFace, fe.charMap)

	fe.updateCharSize()
	fe.updateSignature()
	fe.changeStamp++

	return nil
}

// updateCharSize updates the character size in FreeType.
func (fe *FontEngineFreetype) updateCharSize() {
	if fe.currentFace != nil {
		C.FT_Set_Char_Size(fe.currentFace,
			C.FT_F26Dot6(fe.width),
			C.FT_F26Dot6(fe.height),
			C.FT_UInt(fe.resolution),
			C.FT_UInt(fe.resolution))
	}
}

// updateSignature updates the font signature string with CRC32 hash.
func (fe *FontEngineFreetype) updateSignature() {
	// Create signature string similar to AGG C++ implementation
	sigStr := fmt.Sprintf("%s_%d_%d_%t_%t_%d",
		fe.name, fe.height, fe.width, fe.hinting, fe.flipY, int(fe.glyphRendering))

	// Calculate CRC32 hash for uniqueness (similar to AGG)
	crc := calcCRC32([]byte(sigStr))
	fe.signature = fmt.Sprintf("%s_%08x", sigStr, crc)
}

// SetHeight sets the font height in 26.6 fixed point (1/64th of a point).
func (fe *FontEngineFreetype) SetHeight(h float64) {
	fe.height = uint(h * 64.0)
	fe.updateCharSize()
	fe.updateSignature()
	fe.changeStamp++
}

// SetWidth sets the font width in 26.6 fixed point.
func (fe *FontEngineFreetype) SetWidth(w float64) {
	fe.width = uint(w * 64.0)
	fe.updateCharSize()
	fe.updateSignature()
	fe.changeStamp++
}

// SetHinting enables or disables font hinting.
func (fe *FontEngineFreetype) SetHinting(h bool) {
	fe.hinting = h
	fe.updateSignature()
	fe.changeStamp++
}

// SetFlipY sets whether to flip Y coordinates.
func (fe *FontEngineFreetype) SetFlipY(f bool) {
	fe.flipY = f
	fe.updateSignature()
	fe.changeStamp++
}

// SetTransform sets the affine transformation matrix.
func (fe *FontEngineFreetype) SetTransform(affine *transform.TransAffine) {
	fe.affine = affine
	fe.changeStamp++
}

// FontSignature returns the unique font signature.
func (fe *FontEngineFreetype) FontSignature() string {
	return fe.signature
}

// ChangeStamp returns the change stamp for cache invalidation.
func (fe *FontEngineFreetype) ChangeStamp() int {
	return fe.changeStamp
}

// GetHeight returns the current font height.
func (fe *FontEngineFreetype) GetHeight() float64 {
	return float64(fe.height) / 64.0
}

// GetWidth returns the current font width.
func (fe *FontEngineFreetype) GetWidth() float64 {
	return float64(fe.width) / 64.0
}

// GetHinting returns whether hinting is enabled.
func (fe *FontEngineFreetype) GetHinting() bool {
	return fe.hinting
}

// GetFlipY returns whether Y coordinates are flipped.
func (fe *FontEngineFreetype) GetFlipY() bool {
	return fe.flipY
}

// GetAscender returns the font ascender.
func (fe *FontEngineFreetype) GetAscender() float64 {
	if fe.currentFace != nil {
		return float64(fe.currentFace.ascender) * fe.GetHeight() / float64(fe.currentFace.units_per_EM)
	}
	return 0
}

// GetDescender returns the font descender.
func (fe *FontEngineFreetype) GetDescender() float64 {
	if fe.currentFace != nil {
		return float64(fe.currentFace.descender) * fe.GetHeight() / float64(fe.currentFace.units_per_EM)
	}
	return 0
}

// NumFaces returns the number of loaded faces.
func (fe *FontEngineFreetype) NumFaces() uint {
	return fe.numFaces
}

// Name returns the current font name.
func (fe *FontEngineFreetype) Name() string {
	return fe.name
}

// LastError returns the last FreeType error code.
func (fe *FontEngineFreetype) LastError() int {
	return fe.lastError
}

// PrepareGlyph prepares a glyph for rendering.
func (fe *FontEngineFreetype) PrepareGlyph(glyphCode uint) bool {
	if fe.currentFace == nil {
		return false
	}

	// Get glyph index
	fe.glyphIndex = uint(C.FT_Get_Char_Index(fe.currentFace, C.FT_ULong(glyphCode)))
	if fe.glyphIndex == 0 {
		return false
	}

	// Load glyph
	loadFlags := C.FT_LOAD_DEFAULT
	if !fe.hinting {
		loadFlags |= C.FT_LOAD_NO_HINTING
	}

	err := C.FT_Load_Glyph(fe.currentFace, C.FT_UInt(fe.glyphIndex), C.FT_Int32(loadFlags))
	if err != 0 {
		fe.lastError = int(err)
		return false
	}

	glyph := fe.currentFace.glyph

	// Set bounds and advance
	fe.bounds = basics.Rect[int]{
		X1: int(glyph.bitmap_left),
		Y1: int(int(glyph.bitmap_top) - int(glyph.bitmap.rows)),
		X2: int(int(glyph.bitmap_left) + int(glyph.bitmap.width)),
		Y2: int(glyph.bitmap_top),
	}

	fe.advanceX = float64(glyph.advance.x) / 64.0
	fe.advanceY = float64(glyph.advance.y) / 64.0

	// Determine data type and size based on rendering type
	switch fe.glyphRendering {
	case GlyphRenderingOutline:
		fe.dataType = font.GlyphDataOutline
		fe.dataSize = 0 // Outline data is stored in path

		// Clear previous path and decompose the outline
		fe.pathStorage.RemoveAll()
		if glyph.format == C.FT_GLYPH_FORMAT_OUTLINE {
			if !fe.decomposeFTOutline(&glyph.outline, fe.flipY, fe.pathStorage) {
				fe.lastError = -1
				return false
			}
		}

	case GlyphRenderingAAGray8:
		fe.dataType = font.GlyphDataGray8
		// Render to bitmap if not already done
		if glyph.format != C.FT_GLYPH_FORMAT_BITMAP {
			err = C.FT_Render_Glyph(glyph, C.FT_RENDER_MODE_NORMAL)
			if err != 0 {
				fe.lastError = int(err)
				return false
			}
		}
		fe.dataSize = uint(int(glyph.bitmap.rows) * int(glyph.bitmap.pitch))

	case GlyphRenderingAAMono:
		fe.dataType = font.GlyphDataMono
		// Render to monochrome bitmap
		if glyph.format != C.FT_GLYPH_FORMAT_BITMAP {
			err = C.FT_Render_Glyph(glyph, C.FT_RENDER_MODE_MONO)
			if err != 0 {
				fe.lastError = int(err)
				return false
			}
		}
		fe.dataSize = uint(int(glyph.bitmap.rows) * int(glyph.bitmap.pitch))

	default:
		fe.dataType = font.GlyphDataInvalid
		fe.dataSize = 0
	}

	return true
}

// GlyphIndex returns the current glyph index.
func (fe *FontEngineFreetype) GlyphIndex() uint {
	return fe.glyphIndex
}

// DataSize returns the size of the current glyph data.
func (fe *FontEngineFreetype) DataSize() uint {
	return fe.dataSize
}

// DataType returns the type of the current glyph data.
func (fe *FontEngineFreetype) DataType() font.GlyphDataType {
	return fe.dataType
}

// Bounds returns the bounding rectangle of the current glyph.
func (fe *FontEngineFreetype) Bounds() basics.Rect[int] {
	return fe.bounds
}

// AdvanceX returns the horizontal advance of the current glyph.
func (fe *FontEngineFreetype) AdvanceX() float64 {
	return fe.advanceX
}

// AdvanceY returns the vertical advance of the current glyph.
func (fe *FontEngineFreetype) AdvanceY() float64 {
	return fe.advanceY
}

// WriteGlyphTo writes the current glyph data to the provided buffer.
func (fe *FontEngineFreetype) WriteGlyphTo(data []byte) {
	if fe.currentFace == nil || fe.dataSize == 0 {
		return
	}

	glyph := fe.currentFace.glyph

	switch fe.dataType {
	case font.GlyphDataGray8:
		bitmap := &glyph.bitmap
		srcData := unsafe.Slice((*byte)(bitmap.buffer), fe.dataSize)
		copy(data, srcData)
	case font.GlyphDataMono:
		bitmap := &glyph.bitmap
		srcData := unsafe.Slice((*byte)(bitmap.buffer), fe.dataSize)
		copy(data, srcData)
	}
}

// AddKerning adds kerning offset between two glyphs.
func (fe *FontEngineFreetype) AddKerning(first, second uint) (dx, dy float64) {
	if fe.currentFace == nil || C.has_kerning(fe.currentFace) == 0 {
		return 0, 0
	}

	var delta C.FT_Vector
	err := C.FT_Get_Kerning(fe.currentFace, C.FT_UInt(first), C.FT_UInt(second),
		C.FT_KERNING_DEFAULT, &delta)
	if err != 0 {
		return 0, 0
	}

	dx = float64(delta.x) / 64.0
	dy = float64(delta.y) / 64.0
	return dx, dy
}

// PathAdaptor returns the path storage for vector fonts.
func (fe *FontEngineFreetype) PathAdaptor() *path.PathStorageStl {
	return fe.pathStorage
}

// decomposeFTOutline converts a FreeType outline to AGG path commands.
// This is a port of AGG's decompose_ft_outline function.
func (fe *FontEngineFreetype) decomposeFTOutline(outline *C.FT_Outline, flipY bool, pathStorage *path.PathStorageStl) bool {
	if outline.n_contours <= 0 {
		return true // Empty outline is valid
	}

	first := 0

	for n := 0; n < int(outline.n_contours); n++ {
		last := int(C.short(uintptr(unsafe.Pointer(uintptr(unsafe.Pointer(outline.contours)) + uintptr(n)*unsafe.Sizeof(C.short(0))))))

		// Bounds checking - ensure indices are within valid range
		if first < 0 || last < 0 || first >= int(outline.n_points) || last >= int(outline.n_points) {
			// Invalid indices - return false to avoid crash
			return false
		}

		// Get starting points from outline using safer array indexing
		vStartOriginal := (*C.FT_Vector)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.points)) + uintptr(first)*unsafe.Sizeof(C.FT_Vector{})))
		vLastOriginal := (*C.FT_Vector)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.points)) + uintptr(last)*unsafe.Sizeof(C.FT_Vector{})))

		// Storage for modified start/last points - ensures memory remains valid throughout loop
		vStartStorage := *vStartOriginal
		vLastStorage := *vLastOriginal

		// Pointers to the active start/last points
		vStart := vStartOriginal
		vLast := vLastOriginal

		vControl := *vStart
		point := vStart
		tags := (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.tags)) + uintptr(first)))
		tag := int(*tags) & 1 // FT_CURVE_TAG_ON = 1

		// A contour cannot start with a cubic control point
		if (int(*tags) & 3) == 3 { // FT_CURVE_TAG_CUBIC = 3
			return false
		}

		// Check first point to determine origin
		if (int(*tags) & 1) == 0 { // FT_CURVE_TAG_CONIC
			// First point is conic control
			lastTag := (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.tags)) + uintptr(last)))
			if int(*lastTag)&1 == 1 { // FT_CURVE_TAG_ON
				// Start at last point if it is on the curve
				vStart = vLast
			} else {
				// If both first and last points are conic, start at their middle
				// Modify the storage variables directly, following C++ implementation
				vStartStorage.x = (vStartStorage.x + vLastStorage.x) / 2
				vStartStorage.y = (vStartStorage.y + vLastStorage.y) / 2
				vLastStorage = vStartStorage

				// Point to our storage variables
				vStart = &vStartStorage
				vLast = &vLastStorage
			}
		}

		// Convert starting point and move to it
		x1 := float64(C.int26p6_to_dbl(C.long(vStart.x)))
		y1 := float64(C.int26p6_to_dbl(C.long(vStart.y)))
		if flipY {
			y1 = -y1
		}
		fe.affine.Transform(&x1, &y1)
		pathStorage.MoveTo(x1, y1)

		// Process outline points
		for i := first; i < last; {
			i++
			point = (*C.FT_Vector)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.points)) + uintptr(i)*unsafe.Sizeof(C.FT_Vector{})))
			tags = (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.tags)) + uintptr(i)))
			tag = int(*tags) & 3

			switch tag {
			case 1: // FT_CURVE_TAG_ON - emit a single line_to
				x1 = float64(C.int26p6_to_dbl(C.long(point.x)))
				y1 = float64(C.int26p6_to_dbl(C.long(point.y)))
				if flipY {
					y1 = -y1
				}
				fe.affine.Transform(&x1, &y1)
				pathStorage.LineTo(x1, y1)

			case 0: // FT_CURVE_TAG_CONIC - consume conic arcs
				vControl = *point

				for {
					if i >= last {
						break
					}

					i++
					point = (*C.FT_Vector)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.points)) + uintptr(i)*unsafe.Sizeof(C.FT_Vector{})))
					tags = (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.tags)) + uintptr(i)))
					tag = int(*tags) & 3

					vec := *point

					if tag == 1 { // FT_CURVE_TAG_ON
						x1 = float64(C.int26p6_to_dbl(C.long(vControl.x)))
						y1 = float64(C.int26p6_to_dbl(C.long(vControl.y)))
						x2 := float64(C.int26p6_to_dbl(C.long(vec.x)))
						y2 := float64(C.int26p6_to_dbl(C.long(vec.y)))
						if flipY {
							y1 = -y1
							y2 = -y2
						}
						fe.affine.Transform(&x1, &y1)
						fe.affine.Transform(&x2, &y2)
						pathStorage.Curve3(x1, y1, x2, y2)
						break
					}

					if tag != 0 { // Not FT_CURVE_TAG_CONIC
						return false
					}

					// Calculate middle point
					vMiddle := C.FT_Vector{
						x: (vControl.x + vec.x) / 2,
						y: (vControl.y + vec.y) / 2,
					}

					x1 = float64(C.int26p6_to_dbl(C.long(vControl.x)))
					y1 = float64(C.int26p6_to_dbl(C.long(vControl.y)))
					x2 := float64(C.int26p6_to_dbl(C.long(vMiddle.x)))
					y2 := float64(C.int26p6_to_dbl(C.long(vMiddle.y)))
					if flipY {
						y1 = -y1
						y2 = -y2
					}
					fe.affine.Transform(&x1, &y1)
					fe.affine.Transform(&x2, &y2)
					pathStorage.Curve3(x1, y1, x2, y2)

					vControl = vec
				}

				// If we broke out early, create final curve to start
				if i >= last {
					x1 = float64(C.int26p6_to_dbl(C.long(vControl.x)))
					y1 = float64(C.int26p6_to_dbl(C.long(vControl.y)))
					x2 := float64(C.int26p6_to_dbl(C.long(vStart.x)))
					y2 := float64(C.int26p6_to_dbl(C.long(vStart.y)))
					if flipY {
						y1 = -y1
						y2 = -y2
					}
					fe.affine.Transform(&x1, &y1)
					fe.affine.Transform(&x2, &y2)
					pathStorage.Curve3(x1, y1, x2, y2)
				}

			default: // FT_CURVE_TAG_CUBIC
				if i+1 > last {
					return false
				}

				vec1 := *point
				i++
				point = (*C.FT_Vector)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.points)) + uintptr(i)*unsafe.Sizeof(C.FT_Vector{})))
				tags = (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.tags)) + uintptr(i)))

				if (int(*tags) & 3) != 3 { // Not FT_CURVE_TAG_CUBIC
					return false
				}

				vec2 := *point

				if i < last {
					i++
					point = (*C.FT_Vector)(unsafe.Pointer(uintptr(unsafe.Pointer(outline.points)) + uintptr(i)*unsafe.Sizeof(C.FT_Vector{})))
					vec := *point

					x1 = float64(C.int26p6_to_dbl(C.long(vec1.x)))
					y1 = float64(C.int26p6_to_dbl(C.long(vec1.y)))
					x2 := float64(C.int26p6_to_dbl(C.long(vec2.x)))
					y2 := float64(C.int26p6_to_dbl(C.long(vec2.y)))
					x3 := float64(C.int26p6_to_dbl(C.long(vec.x)))
					y3 := float64(C.int26p6_to_dbl(C.long(vec.y)))
					if flipY {
						y1 = -y1
						y2 = -y2
						y3 = -y3
					}
					fe.affine.Transform(&x1, &y1)
					fe.affine.Transform(&x2, &y2)
					fe.affine.Transform(&x3, &y3)
					pathStorage.Curve4(x1, y1, x2, y2, x3, y3)
				} else {
					x1 = float64(C.int26p6_to_dbl(C.long(vec1.x)))
					y1 = float64(C.int26p6_to_dbl(C.long(vec1.y)))
					x2 := float64(C.int26p6_to_dbl(C.long(vec2.x)))
					y2 := float64(C.int26p6_to_dbl(C.long(vec2.y)))
					x3 := float64(C.int26p6_to_dbl(C.long(vStart.x)))
					y3 := float64(C.int26p6_to_dbl(C.long(vStart.y)))
					if flipY {
						y1 = -y1
						y2 = -y2
						y3 = -y3
					}
					fe.affine.Transform(&x1, &y1)
					fe.affine.Transform(&x2, &y2)
					fe.affine.Transform(&x3, &y3)
					pathStorage.Curve4(x1, y1, x2, y2, x3, y3)
				}
			}
		}

		pathStorage.ClosePolygon(basics.PathFlagsNone)
		first = last + 1
	}

	return true
}
