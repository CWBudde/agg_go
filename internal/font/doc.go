// Package font ports AGG's font cache manager layer.
//
// The package sits between a font engine and the Agg2D text pipeline:
// a FontEngine prepares glyph metrics and serialized glyph data, FontCache
// stores that data per font signature, and FontCacheManager coordinates cache
// lookup, kerning, and adaptor setup for vector and scanline glyph rendering.
//
// This structure follows agg_font_cache_manager.h closely. The FreeType-backed
// engines live under the build-tagged subpackages in internal/font/freetype
// and internal/font/freetype2.
package font
