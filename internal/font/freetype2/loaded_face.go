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
	"errors"
	"fmt"
	"math"
	"unsafe"

	"agg_go/internal/basics"
	"agg_go/internal/fonts"
	"agg_go/internal/rasterizer"
	"agg_go/internal/scanline"
	"agg_go/internal/transform"
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
	dpi       uint32
	height    float64
	width     float64
	rendering GlyphRendering
	hinting   bool
	flipY     bool
	charMap   CharEncoding

	// Transformation matrix
	affine *transform.TransAffine

	closed bool
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
	return float64(lf.ftFace.ascender) * lf.height / float64(lf.ftFace.height)
}

// Descent returns the typographic descender.
func (lf *LoadedFace) Descent() float64 {
	if lf.ftFace == nil {
		return 0
	}
	return float64(lf.ftFace.descender) * lf.height / float64(lf.ftFace.height)
}

// AscentB returns the maximum ascender (bounding box).
func (lf *LoadedFace) AscentB() float64 {
	if lf.ftFace == nil {
		return 0
	}
	return float64(lf.ftFace.bbox.yMax) * lf.height / float64(lf.ftFace.height)
}

// DescentB returns the maximum descender (bounding box).
func (lf *LoadedFace) DescentB() float64 {
	if lf.ftFace == nil {
		return 0
	}
	return float64(lf.ftFace.bbox.yMin) * lf.height / float64(lf.ftFace.height)
}

// Hinting returns whether hinting is enabled.
func (lf *LoadedFace) Hinting() bool {
	return lf.hinting
}

// FlipY returns whether Y coordinates are flipped.
func (lf *LoadedFace) FlipY() bool {
	return lf.flipY
}

// Rendering returns the current glyph rendering mode.
func (lf *LoadedFace) Rendering() GlyphRendering {
	return lf.rendering
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

	prepared.AdvanceX = float64(C.ft_26dot6_to_double(glyph.advance.x))
	prepared.AdvanceY = float64(C.ft_26dot6_to_double(glyph.advance.y))

	// Determine data type and size based on rendering mode
	switch lf.rendering {
	case GlyphRenOutline:
		prepared.DataType = fonts.FmanGlyphDataOutline
		if glyph.format != C.FT_GLYPH_FORMAT_OUTLINE {
			return nil, false
		}
		if !lf.prepareOutlineGlyph(glyph, prepared) {
			return nil, false
		}

	case GlyphRenNativeGray8, GlyphRenAggGray8:
		prepared.DataType = fonts.FmanGlyphDataGray8
		if !lf.prepareGray8Glyph(glyph, prepared) {
			return nil, false
		}

	case GlyphRenNativeMono, GlyphRenAggMono:
		prepared.DataType = fonts.FmanGlyphDataMono
		if !lf.prepareMonoGlyph(glyph, prepared) {
			return nil, false
		}

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
	if lf.rendering == GlyphRenOutline || lf.rendering == GlyphRenAggMono || lf.rendering == GlyphRenAggGray8 {
		if lf.affine != nil {
			lf.affine.Transform2x2(&dx, &dy)
		}
	}
	return dx, dy
}

// WriteGlyphTo writes the current glyph bitmap data to the provided buffer.
// This corresponds to AGG's loaded_face::write_glyph_to method.
func (lf *LoadedFace) WriteGlyphTo(prepared *PreparedGlyph, data []byte) {
	if lf.ftFace == nil || prepared.DataSize == 0 {
		return
	}

	switch prepared.DataType {
	case fonts.FmanGlyphDataGray8:
		if len(data) >= lf.engine.scanlinesAA.ByteSize() {
			lf.engine.scanlinesAA.Serialize(data)
		}

	case fonts.FmanGlyphDataMono:
		if len(data) >= lf.engine.scanlinesBin.ByteSize() {
			lf.engine.scanlinesBin.Serialize(data)
		}

	case fonts.FmanGlyphDataOutline:
		if lf.engine.flag32 {
			serialized, err := lf.engine.pathStorage32.Serialize()
			if err == nil {
				copy(data, serialized)
			}
		} else {
			serialized, err := lf.engine.pathStorage16.Serialize()
			if err == nil {
				copy(data, serialized)
			}
		}
	}
}

func (lf *LoadedFace) prepareOutlineGlyph(glyph *C.FT_GlyphSlotRec, prepared *PreparedGlyph) bool {
	if decomErr := lf.engine.DecomposeFTOutline(&glyph.outline, lf.flipY, lf.affine); decomErr != nil {
		return false
	}

	var bounds basics.Rect[float64]
	if lf.engine.flag32 {
		bounds = lf.engine.pathStorage32.BoundingRect()
		prepared.DataSize = lf.engine.pathStorage32.ByteSize()
	} else {
		bounds = lf.engine.pathStorage16.BoundingRect()
		prepared.DataSize = lf.engine.pathStorage16.ByteSize()
	}

	if prepared.DataSize == 0 {
		prepared.Bounds = basics.Rect[int]{}
		return true
	}

	prepared.Bounds = basics.Rect[int]{
		X1: int(math.Floor(bounds.X1)),
		Y1: int(math.Floor(bounds.Y1)),
		X2: int(math.Ceil(bounds.X2)),
		Y2: int(math.Ceil(bounds.Y2)),
	}
	if lf.affine != nil {
		lf.affine.Transform2x2(&prepared.AdvanceX, &prepared.AdvanceY)
	}
	return true
}

func (lf *LoadedFace) prepareGray8Glyph(glyph *C.FT_GlyphSlotRec, prepared *PreparedGlyph) bool {
	if lf.rendering == GlyphRenNativeGray8 {
		if glyph.format != C.FT_GLYPH_FORMAT_BITMAP {
			if err := C.FT_Render_Glyph(glyph, C.FT_RENDER_MODE_NORMAL); err != 0 {
				return false
			}
		}
		if !lf.decomposeBitmapGray8(&glyph.bitmap, int(glyph.bitmap_left), int(glyph.bitmap_top)) {
			return false
		}
	} else {
		if glyph.format != C.FT_GLYPH_FORMAT_OUTLINE {
			return false
		}
		if !lf.rasterizeOutlineToGray8(&glyph.outline) {
			return false
		}
		if lf.affine != nil {
			lf.affine.Transform2x2(&prepared.AdvanceX, &prepared.AdvanceY)
		}
	}

	if lf.engine.scanlinesAA.MinX() > lf.engine.scanlinesAA.MaxX() {
		prepared.Bounds = basics.Rect[int]{}
		prepared.DataSize = 0
		return true
	}

	prepared.Bounds = basics.Rect[int]{
		X1: lf.engine.scanlinesAA.MinX(),
		Y1: lf.engine.scanlinesAA.MinY(),
		X2: lf.engine.scanlinesAA.MaxX() + 1,
		Y2: lf.engine.scanlinesAA.MaxY() + 1,
	}
	prepared.DataSize = uint32(lf.engine.scanlinesAA.ByteSize())
	return true
}

func (lf *LoadedFace) prepareMonoGlyph(glyph *C.FT_GlyphSlotRec, prepared *PreparedGlyph) bool {
	if lf.rendering == GlyphRenNativeMono {
		if glyph.format != C.FT_GLYPH_FORMAT_BITMAP {
			if err := C.FT_Render_Glyph(glyph, C.FT_RENDER_MODE_MONO); err != 0 {
				return false
			}
		}
		if !lf.decomposeBitmapMono(&glyph.bitmap, int(glyph.bitmap_left), int(glyph.bitmap_top)) {
			return false
		}
	} else {
		if glyph.format != C.FT_GLYPH_FORMAT_OUTLINE {
			return false
		}
		if !lf.rasterizeOutlineToMono(&glyph.outline) {
			return false
		}
		if lf.affine != nil {
			lf.affine.Transform2x2(&prepared.AdvanceX, &prepared.AdvanceY)
		}
	}

	if lf.engine.scanlinesBin.MinX() > lf.engine.scanlinesBin.MaxX() {
		prepared.Bounds = basics.Rect[int]{}
		prepared.DataSize = 0
		return true
	}

	prepared.Bounds = basics.Rect[int]{
		X1: lf.engine.scanlinesBin.MinX(),
		Y1: lf.engine.scanlinesBin.MinY(),
		X2: lf.engine.scanlinesBin.MaxX() + 1,
		Y2: lf.engine.scanlinesBin.MaxY() + 1,
	}
	prepared.DataSize = uint32(lf.engine.scanlinesBin.ByteSize())
	return true
}

func (lf *LoadedFace) decomposeBitmapGray8(bitmap *C.FT_Bitmap, left, top int) bool {
	if bitmap == nil {
		return false
	}
	lf.engine.scanlinesAA.Prepare()
	sl := lf.engine.scanlineU8
	width := int(bitmap.width)
	rows := int(bitmap.rows)
	pitch := absInt(int(bitmap.pitch))
	if width <= 0 || rows <= 0 || bitmap.buffer == nil {
		return true
	}

	src := unsafe.Slice((*byte)(bitmap.buffer), rows*pitch)
	for row := 0; row < rows; row++ {
		y := bitmapRowY(top, row, lf.flipY)
		sl.Reset(left, left+width-1)
		x := 0
		for x < width {
			if src[row*pitch+x] == 0 {
				x++
				continue
			}
			start := x
			for x < width && src[row*pitch+x] != 0 {
				x++
			}
			covers := make([]scanline.CoverType, x-start)
			copy(covers, src[row*pitch+start:row*pitch+x])
			sl.AddCells(left+start, x-start, covers)
		}
		if sl.NumSpans() == 0 {
			continue
		}
		sl.Finalize(y)
		lf.engine.scanlinesAA.Render(scanlineU8Wrapper{sl})
	}
	return true
}

func (lf *LoadedFace) decomposeBitmapMono(bitmap *C.FT_Bitmap, left, top int) bool {
	if bitmap == nil {
		return false
	}
	lf.engine.scanlinesBin.Prepare()
	sl := lf.engine.scanlineBin
	width := int(bitmap.width)
	rows := int(bitmap.rows)
	pitch := absInt(int(bitmap.pitch))
	if width <= 0 || rows <= 0 || bitmap.buffer == nil {
		return true
	}

	src := unsafe.Slice((*byte)(bitmap.buffer), rows*pitch)
	for row := 0; row < rows; row++ {
		y := bitmapRowY(top, row, lf.flipY)
		sl.Reset(left, left+width-1)
		x := 0
		for x < width {
			if !monoBitmapPixel(src[row*pitch:], x) {
				x++
				continue
			}
			start := x
			for x < width && monoBitmapPixel(src[row*pitch:], x) {
				x++
			}
			sl.AddSpan(left+start, x-start, 0)
		}
		if sl.NumSpans() == 0 {
			continue
		}
		sl.Finalize(y)
		lf.engine.scanlinesBin.RenderBinScanline(sl)
	}
	return true
}

func (lf *LoadedFace) rasterizeOutlineToGray8(outline *C.FT_Outline) bool {
	if outline == nil {
		return false
	}
	if err := lf.engine.DecomposeFTOutline(outline, lf.flipY, lf.affine); err != nil {
		return false
	}
	lf.engine.scanlinesAA.Prepare()
	ras := rasterizer.NewRasterizerScanlineAAWithGamma[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(), lf.engine.gammaFunc.Apply,
	)
	if lf.engine.flag32 {
		ras.AddPath(rasterizerVertexSourceAdapter{src: lf.engine.curves32}, 0)
	} else {
		ras.AddPath(rasterizerVertexSourceAdapter{src: lf.engine.curves16}, 0)
	}
	return renderRasterizerToAAStorage(ras, lf.engine.scanlineU8, lf.engine.scanlinesAA)
}

func (lf *LoadedFace) rasterizeOutlineToMono(outline *C.FT_Outline) bool {
	if outline == nil {
		return false
	}
	if err := lf.engine.DecomposeFTOutline(outline, lf.flipY, lf.affine); err != nil {
		return false
	}
	lf.engine.scanlinesBin.Prepare()
	ras := rasterizer.NewRasterizerScanlineAAWithGamma[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(), lf.engine.gammaFunc.Apply,
	)
	if lf.engine.flag32 {
		ras.AddPath(rasterizerVertexSourceAdapter{src: lf.engine.curves32}, 0)
	} else {
		ras.AddPath(rasterizerVertexSourceAdapter{src: lf.engine.curves16}, 0)
	}
	return renderRasterizerToBinStorage(ras, lf.engine.scanlineBin, lf.engine.scanlinesBin)
}

type scanlineU8Wrapper struct {
	sl *scanline.ScanlineU8
}

func (w scanlineU8Wrapper) Y() int                                 { return w.sl.Y() }
func (w scanlineU8Wrapper) NumSpans() int                          { return w.sl.NumSpans() }
func (w scanlineU8Wrapper) ResetSpans()                            { w.sl.ResetSpans() }
func (w scanlineU8Wrapper) AddSpan(x, len int, cover basics.Int8u) { w.sl.AddSpan(x, len, uint(cover)) }
func (w scanlineU8Wrapper) AddCells(x, len int, covers []basics.Int8u) {
	w.sl.AddCells(x, len, covers)
}
func (w scanlineU8Wrapper) Finalize(y int) { w.sl.Finalize(y) }
func (w scanlineU8Wrapper) Begin() scanline.ScanlineIterator {
	return &scanlineU8Iterator{spans: w.sl.Spans(), idx: 0}
}

type scanlineU8Iterator struct {
	spans []scanline.Span
	idx   int
}

func (it *scanlineU8Iterator) GetSpan() scanline.SpanInfo {
	if it.idx >= len(it.spans) {
		return scanline.SpanInfo{}
	}
	span := it.spans[it.idx]
	return scanline.SpanInfo{
		X:      int(span.X),
		Len:    int(span.Len),
		Covers: span.Covers[:int(span.Len)],
	}
}

func (it *scanlineU8Iterator) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

type rasterizerVertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type rasterizerVertexSourceAdapter struct {
	src rasterizerVertexSource
}

func (a rasterizerVertexSourceAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a rasterizerVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type rasterizerScanlineU8Adapter struct {
	sl *scanline.ScanlineU8
}

func (a *rasterizerScanlineU8Adapter) ResetSpans() {
	a.sl.ResetSpans()
}

func (a *rasterizerScanlineU8Adapter) AddCell(x int, cover uint32) {
	a.sl.AddCell(x, uint(cover))
}

func (a *rasterizerScanlineU8Adapter) AddSpan(x, len int, cover uint32) {
	a.sl.AddSpan(x, len, uint(cover))
}

func (a *rasterizerScanlineU8Adapter) Finalize(y int) {
	a.sl.Finalize(y)
}

func (a *rasterizerScanlineU8Adapter) NumSpans() int {
	return a.sl.NumSpans()
}

type rasterizerScanlineBinAdapter struct {
	sl *scanline.ScanlineBin
}

func (a *rasterizerScanlineBinAdapter) ResetSpans() {
	a.sl.ResetSpans()
}

func (a *rasterizerScanlineBinAdapter) AddCell(x int, cover uint32) {
	a.sl.AddCell(x, uint(cover))
}

func (a *rasterizerScanlineBinAdapter) AddSpan(x, len int, cover uint32) {
	a.sl.AddSpan(x, len, uint(cover))
}

func (a *rasterizerScanlineBinAdapter) Finalize(y int) {
	a.sl.Finalize(y)
}

func (a *rasterizerScanlineBinAdapter) NumSpans() int {
	return a.sl.NumSpans()
}

func renderRasterizerToAAStorage(ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip], sl *scanline.ScanlineU8, storage *scanline.ScanlineStorageAA[uint8]) bool {
	if !ras.RewindScanlines() {
		return true
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	adapter := &rasterizerScanlineU8Adapter{sl: sl}
	for ras.SweepScanline(adapter) {
		storage.Render(scanlineU8Wrapper{sl})
	}
	return true
}

func renderRasterizerToBinStorage(ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip], sl *scanline.ScanlineBin, storage *scanline.ScanlineStorageBin) bool {
	if !ras.RewindScanlines() {
		return true
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	adapter := &rasterizerScanlineBinAdapter{sl: sl}
	for ras.SweepScanline(adapter) {
		storage.RenderBinScanline(sl)
	}
	return true
}

func bitmapRowY(top, row int, flipY bool) int {
	if flipY {
		return -top + row
	}
	return top - row - 1
}

func monoBitmapPixel(row []byte, x int) bool {
	if x < 0 {
		return false
	}
	byteIdx := x >> 3
	if byteIdx >= len(row) {
		return false
	}
	mask := byte(0x80 >> (x & 7))
	return row[byteIdx]&mask != 0
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Close cleans up resources used by this loaded face.
func (lf *LoadedFace) Close() error {
	if lf.closed {
		return nil
	}
	lf.closed = true

	if lf.engine != nil {
		engine := lf.engine
		lf.engine = nil
		return engine.closeLoadedFace(lf, true)
	}

	if lf.ftFace != nil {
		C.FT_Done_Face(lf.ftFace)
		lf.ftFace = nil
	}
	lf.faceName = ""
	return nil
}
