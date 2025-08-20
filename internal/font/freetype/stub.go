//go:build !freetype

// Package freetype provides a stub implementation when FreeType is not available.
package freetype

import (
	"agg_go/internal/basics"
	"agg_go/internal/path"
	"errors"
)

// FontEngineFreetype is a stub implementation when FreeType is not available.
type FontEngineFreetype struct {
	signature string
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

// GlyphDataType defines the type of glyph data stored in a cache entry.
type GlyphDataType int

const (
	GlyphDataInvalid GlyphDataType = iota
	GlyphDataMono
	GlyphDataGray8
	GlyphDataOutline
)

// NewFontEngineFreetype returns an error indicating FreeType is not available.
func NewFontEngineFreetype(flag32 bool, maxFaces uint) (*FontEngineFreetype, error) {
	return nil, errors.New("FreeType support not compiled in - rebuild with 'freetype' build tag")
}

// Stub methods to satisfy the FontEngine interface

func (fe *FontEngineFreetype) Close() error {
	return errors.New("FreeType not available")
}

func (fe *FontEngineFreetype) FontSignature() string {
	return ""
}

func (fe *FontEngineFreetype) ChangeStamp() int {
	return 0
}

func (fe *FontEngineFreetype) PrepareGlyph(glyphCode uint) bool {
	return false
}

func (fe *FontEngineFreetype) GlyphIndex() uint {
	return 0
}

func (fe *FontEngineFreetype) DataSize() uint {
	return 0
}

func (fe *FontEngineFreetype) DataType() interface{} {
	return GlyphDataInvalid
}

func (fe *FontEngineFreetype) Bounds() interface{} {
	return basics.Rect[int]{}
}

func (fe *FontEngineFreetype) AdvanceX() float64 {
	return 0
}

func (fe *FontEngineFreetype) AdvanceY() float64 {
	return 0
}

func (fe *FontEngineFreetype) WriteGlyphTo(data []byte) {
}

func (fe *FontEngineFreetype) AddKerning(first, second uint) (dx, dy float64) {
	return 0, 0
}

func (fe *FontEngineFreetype) PathAdaptor() *path.PathStorageStl {
	return nil
}

func (fe *FontEngineFreetype) SetResolution(dpi uint) {
}

func (fe *FontEngineFreetype) LoadFont(fontName string, faceIndex uint, renType GlyphRenderingType, fontMem []byte) error {
	return errors.New("FreeType not available")
}

func (fe *FontEngineFreetype) SetHeight(h float64) {
}

func (fe *FontEngineFreetype) SetWidth(w float64) {
}

func (fe *FontEngineFreetype) SetHinting(h bool) {
}

func (fe *FontEngineFreetype) SetFlipY(f bool) {
}

func (fe *FontEngineFreetype) SetTransform(affine interface{}) {
}

func (fe *FontEngineFreetype) GetHeight() float64 {
	return 0
}

func (fe *FontEngineFreetype) GetWidth() float64 {
	return 0
}

func (fe *FontEngineFreetype) GetHinting() bool {
	return false
}

func (fe *FontEngineFreetype) GetFlipY() bool {
	return false
}

func (fe *FontEngineFreetype) GetAscender() float64 {
	return 0
}

func (fe *FontEngineFreetype) GetDescender() float64 {
	return 0
}

func (fe *FontEngineFreetype) NumFaces() uint {
	return 0
}

func (fe *FontEngineFreetype) Name() string {
	return ""
}

func (fe *FontEngineFreetype) LastError() int {
	return -1
}
