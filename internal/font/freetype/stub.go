//go:build !freetype

// Package freetype provides the build-tagged FreeType font engine wrapper used
// by Agg2D text rendering.
package freetype

import (
	"errors"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/font"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// FontEngineFreetype is the no-freetype stub returned in builds without the
// freetype tag.
type FontEngineFreetype struct {
	signature string
}

// GlyphRenderingType mirrors the rendering modes exposed by the real FreeType
// engine so callers can compile without the freetype tag.
type GlyphRenderingType int

const (
	GlyphRenderingNative GlyphRenderingType = iota
	GlyphRenderingOutline
	GlyphRenderingAAGray8
	GlyphRenderingAAMono
	GlyphRenderingMono
)

// Use GlyphDataType from font package to avoid duplication

// NewFontEngineFreetype reports that the build was compiled without FreeType
// support.
func NewFontEngineFreetype(flag32 bool, maxFaces uint) (*FontEngineFreetype, error) {
	return nil, errors.New("FreeType support not compiled in - rebuild with 'freetype' build tag")
}

// The remaining methods satisfy the same interface as the CGO-backed engine
// while reporting that FreeType support is unavailable.

func (fe *FontEngineFreetype) Close() error {
	return errors.New("FreeType not available")
}

func (fe *FontEngineFreetype) FontSignature() string {
	return fe.signature
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

func (fe *FontEngineFreetype) DataType() font.GlyphDataType {
	return font.GlyphDataInvalid
}

func (fe *FontEngineFreetype) Bounds() basics.Rect[int] {
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

func (fe *FontEngineFreetype) SetTransform(affine *transform.TransAffine) {
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
