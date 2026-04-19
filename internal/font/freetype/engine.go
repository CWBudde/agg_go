//go:build freetype

// Package freetype provides the build-tagged CGO wrapper around FreeType used by
// AGG's font-cache pipeline.
package freetype

/*
#cgo pkg-config: freetype2
#include <ft2build.h>
#include FT_FREETYPE_H
#include FT_MODULE_H
#include <stdlib.h>
#include <string.h>

// Set the TrueType bytecode interpreter version globally for the given
// library. Values are 35 (classic Apple), 38 (Infinality/older CT), 40
// (default modern ClearType). Returns 0 on success, FT_Error otherwise.
static int agg_set_tt_interpreter_version(FT_Library lib, unsigned int version) {
    FT_UInt v = version;
    return FT_Property_Set(lib, "truetype", "interpreter-version", &v);
}

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
	"math"
	"unsafe"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/font"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt/gamma"
	isc "github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
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

// FontEngineFreetype implements font.FontEngine on top of a FreeType library
// instance, closely following AGG's font_engine_freetype_* classes.
type FontEngineFreetype struct {
	// Configuration
	flag32               bool
	changeStamp          int
	lastError            int
	name                 string
	nameLen              uint
	faceIndex            uint
	charMap              C.FT_Encoding
	signature            string
	height               uint
	width                uint
	hinting              bool
	forceAutohint        bool
	snapOutlineX         bool
	ttInterpreterVersion uint
	flipY                bool
	libraryInitialized   bool
	resolution           int
	glyphRendering       GlyphRenderingType
	affine               *transform.TransAffine
	gammaFunc            gamma.GammaFunction
	gammaTable           [256]basics.Int8u

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
	bitmapData []byte
	bitmapW    int
	bitmapH    int
	bitmapPitch int
	bitmapLeft int
	bitmapTop  int
	bitmapMode uint8

	// Path storage for outline fonts
	pathStorage  *path.PathStorageStl
	scanlineU8   *isc.ScanlineU8
	scanlineBin  *isc.ScanlineBin
	scanlinesAA  *isc.ScanlineStorageAA[basics.Int8u]
	scanlinesBin *isc.ScanlineStorageBin
}

// GlyphRenderingType selects the glyph representation requested from FreeType.
type GlyphRenderingType int

const (
	GlyphRenderingNative GlyphRenderingType = iota
	GlyphRenderingOutline
	GlyphRenderingAAGray8
	GlyphRenderingAAMono
	GlyphRenderingMono
)

// NewFontEngineFreetype creates a FreeType engine with a bounded face cache.
func NewFontEngineFreetype(flag32 bool, maxFaces uint) (*FontEngineFreetype, error) {
	if maxFaces == 0 {
		maxFaces = 32
	}

	engine := &FontEngineFreetype{
		flag32:       flag32,
		maxFaces:     maxFaces,
		resolution:   72, // Default DPI
		hinting:      true,
		flipY:        false,
		pathStorage:  path.NewPathStorageStl(),
		scanlineU8:   isc.NewScanlineU8(),
		scanlineBin:  isc.NewScanlineBin(),
		scanlinesAA:  isc.NewScanlineStorageAA[basics.Int8u](),
		scanlinesBin: isc.NewScanlineStorageBin(),
		affine:       transform.NewTransAffine(),
		gammaFunc:    gamma.NewGammaNone(),
	}
	engine.rebuildGammaTable()

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
	fe.updateSignature()
	fe.changeStamp++
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
		fe.applyFaceTransform()
	}
}

func (fe *FontEngineFreetype) applyFaceTransform() {
	if fe.currentFace == nil {
		return
	}

	mtx := transform.NewTransAffine()
	if fe.affine != nil {
		*mtx = *fe.affine
	}

	matrix := C.FT_Matrix{
		xx: C.FT_Fixed(dblToPlainFX(mtx.SX)),
		xy: C.FT_Fixed(dblToPlainFX(mtx.SHX)),
		yx: C.FT_Fixed(dblToPlainFX(mtx.SHY)),
		yy: C.FT_Fixed(dblToPlainFX(mtx.SY)),
	}
	delta := C.FT_Vector{
		x: C.FT_Pos(dblToPlainFX(mtx.TX)),
		y: C.FT_Pos(dblToPlainFX(mtx.TY)),
	}
	C.FT_Set_Transform(fe.currentFace, &matrix, &delta)
}

func (fe *FontEngineFreetype) applyFaceTransformWithSubpixelOffset(offsetX, offsetY float64) {
	if fe.currentFace == nil {
		return
	}

	mtx := transform.NewTransAffine()
	if fe.affine != nil {
		*mtx = *fe.affine
	}

	matrix := C.FT_Matrix{
		xx: C.FT_Fixed(dblToPlainFX(mtx.SX)),
		xy: C.FT_Fixed(dblToPlainFX(mtx.SHX)),
		yx: C.FT_Fixed(dblToPlainFX(mtx.SHY)),
		yy: C.FT_Fixed(dblToPlainFX(mtx.SY)),
	}
	delta := C.FT_Vector{
		x: C.dbl_to_int26p6(C.double(offsetX)),
		y: C.dbl_to_int26p6(C.double(offsetY)),
	}
	C.FT_Set_Transform(fe.currentFace, &matrix, &delta)
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func dblToPlainFX(v float64) int {
	return int(v * 65536.0)
}

func usesGammaInSignature(rendering GlyphRenderingType) bool {
	switch rendering {
	case GlyphRenderingAAGray8, GlyphRenderingAAMono:
		return true
	default:
		return false
	}
}

func usesAffineInSignature(rendering GlyphRenderingType) bool {
	switch rendering {
	case GlyphRenderingOutline, GlyphRenderingAAGray8, GlyphRenderingAAMono:
		return true
	default:
		return false
	}
}

func buildGammaTable(gammaFunc gamma.GammaFunction) [256]basics.Int8u {
	var table [256]basics.Int8u
	for i := range table {
		value := basics.URound(gammaFunc.Apply(float64(i)/255.0) * 255.0)
		if value > 255 {
			value = 255
		}
		table[i] = basics.Int8u(value)
	}
	return table
}

func (fe *FontEngineFreetype) rebuildGammaTable() {
	fe.gammaTable = buildGammaTable(fe.gammaFunc)
}

// updateSignature updates the font signature string following AGG's cache key layout.
func (fe *FontEngineFreetype) updateSignature() {
	gammaHash := uint32(0)
	if usesGammaInSignature(fe.glyphRendering) {
		gammaHash = calcCRC32(fe.gammaTable[:])
	}

	fe.signature = fmt.Sprintf("%s,%d,%d,%d,%d:%dx%d,%d,%d,%d,%d,%d,%08X",
		fe.name,
		int(fe.charMap),
		fe.faceIndex,
		int(fe.glyphRendering),
		fe.resolution,
		fe.height,
		fe.width,
		boolInt(fe.hinting),
		boolInt(fe.forceAutohint),
		boolInt(fe.snapOutlineX),
		fe.ttInterpreterVersion,
		boolInt(fe.flipY),
		gammaHash,
	)

	if usesAffineInSignature(fe.glyphRendering) {
		mtx := [6]float64{1, 0, 0, 1, 0, 0}
		if fe.affine != nil {
			fe.affine.StoreTo(mtx[:])
		}
		fe.signature += fmt.Sprintf(",%08X%08X%08X%08X%08X%08X",
			dblToPlainFX(mtx[0]),
			dblToPlainFX(mtx[1]),
			dblToPlainFX(mtx[2]),
			dblToPlainFX(mtx[3]),
			dblToPlainFX(mtx[4]),
			dblToPlainFX(mtx[5]),
		)
	}
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

// SetForceAutohint forces FreeType to use its auto-hinter instead of
// the font's native TrueType bytecode hints. Has no effect when hinting
// is disabled. Useful for Java-like stem grid-fitting behaviour when
// the font's built-in hints leave stems at fractional pixel positions.
func (fe *FontEngineFreetype) SetForceAutohint(f bool) {
	fe.forceAutohint = f
	fe.updateSignature()
	fe.changeStamp++
}

// SetSnapOutlineX enables integer-snapping of the glyph outline's leftmost
// X coordinate to the pixel grid after FT_Load_Glyph. When enabled (outline
// rendering only), the whole outline path is translated so the leftmost
// vertex lands on an integer pixel. Mimics Java Graphics2D's aggressive
// stem grid-fitting. Advance metrics are unchanged.
func (fe *FontEngineFreetype) SetSnapOutlineX(s bool) {
	fe.snapOutlineX = s
	fe.updateSignature()
	fe.changeStamp++
}

// SetTrueTypeInterpreterVersion selects the TrueType bytecode interpreter
// version used by FreeType for this engine's library. Passing 0 leaves the
// library default untouched. Common values:
//
//   - 35: classic Apple-style — aggressive X+Y stem grid-fit; closest to older
//     Oracle JDK T2K and modern OpenJDK greyscale-AA output.
//   - 38: intermediate "Infinality" / early ClearType-like behaviour.
//   - 40: FreeType's current default — avoids X-direction grid-fit to preserve
//     advance width for sub-pixel AA.
//
// Because FT_Property_Set applies library-globally, this affects all subsequent
// glyph loads from this engine. Returns an error if the library rejects the
// value (e.g. the requested interpreter module isn't built in).
func (fe *FontEngineFreetype) SetTrueTypeInterpreterVersion(v uint) error {
	fe.ttInterpreterVersion = v
	if !fe.libraryInitialized || v == 0 {
		fe.updateSignature()
		fe.changeStamp++
		return nil
	}

	rc := C.agg_set_tt_interpreter_version(*fe.library, C.uint(v))
	if rc != 0 {
		return fmt.Errorf("FT_Property_Set(truetype.interpreter-version=%d) failed: FT_Error=%d", v, int(rc))
	}

	fe.updateSignature()
	fe.changeStamp++
	return nil
}

// SetFlipY sets whether to flip Y coordinates.
func (fe *FontEngineFreetype) SetFlipY(f bool) {
	fe.flipY = f
	fe.updateSignature()
	fe.changeStamp++
}

// SetTransform sets the affine transformation matrix.
func (fe *FontEngineFreetype) SetTransform(affine *transform.TransAffine) {
	if affine == nil {
		fe.affine = transform.NewTransAffine()
	} else {
		fe.affine = affine
	}
	fe.applyFaceTransform()
	fe.updateSignature()
	fe.changeStamp++
}

// SetGamma configures AGG-style gamma correction for gray8 bitmap glyph coverage.
func (fe *FontEngineFreetype) SetGamma(gammaValue float64) {
	switch {
	case gammaValue <= 0 || gammaValue == 1.0:
		fe.gammaFunc = gamma.NewGammaNone()
	default:
		fe.gammaFunc = gamma.NewGammaPower(gammaValue)
	}
	fe.rebuildGammaTable()
	fe.updateSignature()
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

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func signedPointerAdd(ptr unsafe.Pointer, offset int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(int64(uintptr(ptr)) + int64(offset)))
}

type scanlineU8StorageWrapper struct {
	sl *isc.ScanlineU8
}

func (w scanlineU8StorageWrapper) Y() int        { return w.sl.Y() }
func (w scanlineU8StorageWrapper) NumSpans() int { return w.sl.NumSpans() }
func (w scanlineU8StorageWrapper) ResetSpans()   { w.sl.ResetSpans() }
func (w scanlineU8StorageWrapper) AddSpan(x, length int, cover basics.Int8u) {
	w.sl.AddSpan(x, length, uint(cover))
}
func (w scanlineU8StorageWrapper) AddCells(x, length int, covers []basics.Int8u) {
	for i := 0; i < length && i < len(covers); i++ {
		w.sl.AddCell(x+i, uint(covers[i]))
	}
}
func (w scanlineU8StorageWrapper) Finalize(y int) { w.sl.Finalize(y) }
func (w scanlineU8StorageWrapper) Begin() isc.ScanlineIterator {
	return w.sl.BeginIterator()
}

func scanlineStorageBoundsAA(storage *isc.ScanlineStorageAA[basics.Int8u]) basics.Rect[int] {
	if storage == nil {
		return basics.Rect[int]{}
	}
	return basics.Rect[int]{
		X1: storage.MinX(),
		Y1: storage.MinY(),
		X2: storage.MaxX() + 1,
		Y2: storage.MaxY() + 1,
	}
}

func scanlineStorageHasBoundsAA(storage *isc.ScanlineStorageAA[basics.Int8u]) bool {
	return storage != nil && storage.MinX() <= storage.MaxX() && storage.MinY() <= storage.MaxY()
}

func scanlineStorageBoundsBin(storage *isc.ScanlineStorageBin) basics.Rect[int] {
	if storage == nil {
		return basics.Rect[int]{}
	}
	return basics.Rect[int]{
		X1: storage.MinX(),
		Y1: storage.MinY(),
		X2: storage.MaxX() + 1,
		Y2: storage.MaxY() + 1,
	}
}

func scanlineStorageHasBoundsBin(storage *isc.ScanlineStorageBin) bool {
	return storage != nil && storage.MinX() <= storage.MaxX() && storage.MinY() <= storage.MaxY()
}

func decomposeFTBitmapGray8(bitmap *C.FT_Bitmap, x, y int, flipY bool, gammaTable []basics.Int8u, sl *isc.ScanlineU8, storage *isc.ScanlineStorageAA[basics.Int8u]) {
	if storage != nil {
		storage.Prepare()
	}
	if bitmap == nil || bitmap.buffer == nil || sl == nil || storage == nil {
		return
	}

	rows := int(bitmap.rows)
	width := int(bitmap.width)
	pitch := int(bitmap.pitch)
	if rows <= 0 || width <= 0 || pitch == 0 {
		return
	}

	sl.Reset(x, x+width)

	buf := unsafe.Pointer(bitmap.buffer)
	step := pitch
	if flipY {
		buf = signedPointerAdd(buf, pitch*(rows-1))
		y += rows
		step = -pitch
	}

	for i := 0; i < rows; i++ {
		sl.ResetSpans()
		for j := 0; j < width; j++ {
			src := *(*byte)(signedPointerAdd(buf, j))
			if src != 0 {
				cover := src
				if len(gammaTable) > int(src) {
					cover = byte(gammaTable[int(src)])
				}
				if cover != 0 {
					sl.AddCell(x+j, uint(cover))
				}
			}
		}
		if sl.NumSpans() > 0 {
			sl.Finalize(y - i - 1)
			storage.Render(scanlineU8StorageWrapper{sl: sl})
		}
		buf = signedPointerAdd(buf, step)
	}
}

func decomposeFTBitmapMono(bitmap *C.FT_Bitmap, x, y int, flipY bool, sl *isc.ScanlineBin, storage *isc.ScanlineStorageBin) {
	if storage != nil {
		storage.Prepare()
	}
	if bitmap == nil || bitmap.buffer == nil || sl == nil || storage == nil {
		return
	}

	rows := int(bitmap.rows)
	width := int(bitmap.width)
	pitch := int(bitmap.pitch)
	rowBytes := absInt(pitch)
	if rows <= 0 || width <= 0 || rowBytes <= 0 {
		return
	}

	sl.Reset(x, x+width)

	buf := unsafe.Pointer(bitmap.buffer)
	step := pitch
	if flipY {
		buf = signedPointerAdd(buf, pitch*(rows-1))
		y += rows
		step = -pitch
	}

	for i := 0; i < rows; i++ {
		sl.ResetSpans()
		for j := 0; j < width; j++ {
			byteIdx := j >> 3
			if byteIdx >= rowBytes {
				continue
			}
			src := *(*byte)(signedPointerAdd(buf, byteIdx))
			bit := uint(7 - (j & 7))
			if ((src >> bit) & 0x1) != 0 {
				sl.AddCell(x+j, 0)
			}
		}
		if sl.NumSpans() > 0 {
			sl.Finalize(y - i - 1)
			storage.RenderBinScanline(sl)
		}
		buf = signedPointerAdd(buf, step)
	}
}

// snapPathStorageMinXToInteger shifts every vertex in the path so that the
// leftmost vertex X coordinate lands on the nearest integer pixel. No-op if
// the path is empty or already snapped. Advance metrics are untouched.
func snapPathStorageMinXToInteger(ps *path.PathStorageStl) {
	if ps == nil {
		return
	}
	total := ps.TotalVertices()
	if total == 0 {
		return
	}

	pathBounds, ok := basics.BoundingRectSingle[float64](
		path.NewPathStorageStlVertexSourceAdapter(ps), 0,
	)
	if !ok {
		return
	}

	shift := math.Round(pathBounds.X1) - pathBounds.X1
	if shift == 0 {
		return
	}

	for i := uint(0); i < total; i++ {
		x, y, cmd := ps.Vertex(i)
		if !basics.IsVertex(basics.PathCommand(cmd)) {
			continue
		}
		ps.ModifyVertex(i, x+shift, y)
	}
}

func outlineBoundsForAgg(pathStorage *path.PathStorageStl) basics.Rect[int] {
	if pathStorage == nil || pathStorage.TotalVertices() == 0 {
		return basics.Rect[int]{}
	}
	pathBounds, ok := basics.BoundingRectSingle[float64](
		path.NewPathStorageStlVertexSourceAdapter(pathStorage), 0,
	)
	if !ok {
		return basics.Rect[int]{}
	}
	return basics.Rect[int]{
		X1: int(math.Floor(pathBounds.X1)),
		Y1: int(math.Floor(pathBounds.Y1)),
		X2: int(math.Ceil(pathBounds.X2)),
		Y2: int(math.Ceil(pathBounds.Y2)),
	}
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

func (fe *FontEngineFreetype) currentLoadFlags() C.FT_Int32 {
	loadFlags := C.FT_LOAD_DEFAULT
	if !fe.hinting {
		loadFlags |= C.FT_LOAD_NO_HINTING
	} else if fe.forceAutohint {
		loadFlags |= C.FT_LOAD_FORCE_AUTOHINT
	}
	return C.FT_Int32(loadFlags)
}

func (fe *FontEngineFreetype) prepareGlyphIndexWithOffset(glyphIndex uint, offsetX, offsetY float64) bool {
	if fe.currentFace == nil {
		return false
	}
	if glyphIndex == 0 {
		return false
	}
	fe.glyphIndex = glyphIndex

	fe.applyFaceTransformWithSubpixelOffset(offsetX, offsetY)
	err := C.FT_Load_Glyph(fe.currentFace, C.FT_UInt(fe.glyphIndex), fe.currentLoadFlags())
	fe.applyFaceTransform()
	if err != 0 {
		fe.lastError = int(err)
		return false
	}

	glyph := fe.currentFace.glyph

	fe.advanceX = float64(glyph.advance.x) / 64.0
	fe.advanceY = float64(glyph.advance.y) / 64.0
	fe.bitmapData = nil
	fe.bitmapW = 0
	fe.bitmapH = 0
	fe.bitmapPitch = 0
	fe.bitmapLeft = 0
	fe.bitmapTop = 0
	fe.bitmapMode = 0

	// Determine data type and size based on rendering type
	switch fe.glyphRendering {
	case GlyphRenderingOutline:
		fe.dataType = font.GlyphDataOutline

		// Clear previous path and decompose the outline
		fe.dataSize = 0
		fe.pathStorage.RemoveAll()
		if glyph.format == C.FT_GLYPH_FORMAT_OUTLINE {
			if !fe.decomposeFTOutline(&glyph.outline, fe.flipY, fe.pathStorage) {
				fe.lastError = -1
				return false
			}
		}
		if fe.snapOutlineX {
			snapPathStorageMinXToInteger(fe.pathStorage)
		}
		fe.bounds = outlineBoundsForAgg(fe.pathStorage)

	case GlyphRenderingAAGray8:
		fe.dataType = font.GlyphDataGray8
		fe.bounds = basics.Rect[int]{}
		fe.dataSize = 0
		// Render to bitmap if not already done
		if glyph.format != C.FT_GLYPH_FORMAT_BITMAP {
			err = C.FT_Render_Glyph(glyph, C.FT_RENDER_MODE_NORMAL)
			if err != 0 {
				fe.lastError = int(err)
				return false
			}
		}
		fe.captureBitmap(glyph)
		top := int(glyph.bitmap_top)
		if fe.flipY {
			top = -top
		}
		decomposeFTBitmapGray8(&glyph.bitmap, int(glyph.bitmap_left), top, fe.flipY, fe.gammaTable[:], fe.scanlineU8, fe.scanlinesAA)
		if scanlineStorageHasBoundsAA(fe.scanlinesAA) {
			fe.bounds = scanlineStorageBoundsAA(fe.scanlinesAA)
			fe.dataSize = uint(fe.scanlinesAA.ByteSize())
		}

	case GlyphRenderingAAMono:
		fe.dataType = font.GlyphDataMono
		fe.bounds = basics.Rect[int]{}
		fe.dataSize = 0
		// Render to monochrome bitmap
		if glyph.format != C.FT_GLYPH_FORMAT_BITMAP {
			err = C.FT_Render_Glyph(glyph, C.FT_RENDER_MODE_MONO)
			if err != 0 {
				fe.lastError = int(err)
				return false
			}
		}
		fe.captureBitmap(glyph)
		top := int(glyph.bitmap_top)
		if fe.flipY {
			top = -top
		}
		decomposeFTBitmapMono(&glyph.bitmap, int(glyph.bitmap_left), top, fe.flipY, fe.scanlineBin, fe.scanlinesBin)
		if scanlineStorageHasBoundsBin(fe.scanlinesBin) {
			fe.bounds = scanlineStorageBoundsBin(fe.scanlinesBin)
			fe.dataSize = uint(fe.scanlinesBin.ByteSize())
		}

	default:
		fe.dataType = font.GlyphDataInvalid
		fe.dataSize = 0
	}

	return true
}

func (fe *FontEngineFreetype) captureBitmap(glyph C.FT_GlyphSlot) {
	if glyph == nil {
		return
	}
	bitmap := glyph.bitmap
	width := int(bitmap.width)
	rows := int(bitmap.rows)
	pitch := int(bitmap.pitch)
	if width <= 0 || rows <= 0 || pitch == 0 || bitmap.buffer == nil {
		return
	}

	rowStride := absInt(pitch)
	size := rowStride * rows
	fe.bitmapData = C.GoBytes(unsafe.Pointer(bitmap.buffer), C.int(size))
	fe.bitmapW = width
	fe.bitmapH = rows
	fe.bitmapPitch = pitch
	fe.bitmapLeft = int(glyph.bitmap_left)
	fe.bitmapTop = int(glyph.bitmap_top)
	fe.bitmapMode = uint8(bitmap.pixel_mode)
}

// PrepareGlyph prepares a glyph for rendering from a Unicode code point.
func (fe *FontEngineFreetype) PrepareGlyph(glyphCode uint) bool {
	if fe.currentFace == nil {
		return false
	}

	glyphIndex := uint(C.FT_Get_Char_Index(fe.currentFace, C.FT_ULong(glyphCode)))
	if glyphIndex == 0 {
		return false
	}
	return fe.prepareGlyphIndexWithOffset(glyphIndex, 0, 0)
}

// PrepareGlyphIndex prepares a glyph for rendering from a font-specific glyph index.
func (fe *FontEngineFreetype) PrepareGlyphIndex(glyphIndex uint) bool {
	return fe.prepareGlyphIndexWithOffset(glyphIndex, 0, 0)
}

// PrepareGlyphIndexSubpixel prepares a glyph for rendering from a font-specific
// glyph index using fractional screen-space offsets in pixels.
func (fe *FontEngineFreetype) PrepareGlyphIndexSubpixel(glyphIndex uint, offsetX, offsetY float64) bool {
	return fe.prepareGlyphIndexWithOffset(glyphIndex, offsetX, offsetY)
}

// CurrentBitmap returns the raw rendered bitmap of the current glyph, if any.
func (fe *FontEngineFreetype) CurrentBitmap() (data []byte, width, height, pitch, left, top int, pixelMode uint8) {
	return fe.bitmapData, fe.bitmapW, fe.bitmapH, fe.bitmapPitch, fe.bitmapLeft, fe.bitmapTop, fe.bitmapMode
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

	switch fe.dataType {
	case font.GlyphDataGray8:
		fe.scanlinesAA.Serialize(data)
	case font.GlyphDataMono:
		fe.scanlinesBin.Serialize(data)
	}
}

// AddKerning adds kerning offset between two glyphs.
func (fe *FontEngineFreetype) AddKerning(first, second uint) (dx, dy float64) {
	if fe.currentFace == nil || first == 0 || second == 0 || C.has_kerning(fe.currentFace) == 0 {
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
	if fe.glyphRendering == GlyphRenderingOutline ||
		fe.glyphRendering == GlyphRenderingAAMono ||
		fe.glyphRendering == GlyphRenderingAAGray8 {
		if fe.affine != nil {
			fe.affine.Transform2x2(&dx, &dy)
		}
	}
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
		lastPtr := uintptr(unsafe.Pointer(outline.contours)) + uintptr(n)*unsafe.Sizeof(C.short(0))
		last := int(*(*C.short)(unsafe.Pointer(lastPtr)))

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
		if (int(*tags) & 3) == 2 { // FT_CURVE_TAG_CUBIC = 0x02
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

				if (int(*tags) & 3) != 2 { // Not FT_CURVE_TAG_CUBIC (0x02)
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
