// Package glyph provides glyph rasterization functionality for AGG.
// This package implements binary glyph rasterizers that can render text
// from embedded font data.
package glyph

import (
	"agg_go/internal/basics"
	"unsafe"
)

// GlyphRect represents the bounding rectangle and positioning information for a glyph
type GlyphRect struct {
	X1, Y1, X2, Y2 int     // Bounding rectangle
	DX, DY         float64 // Advance vector
}

// GlyphGenerator defines the interface for glyph generation
type GlyphGenerator interface {
	// Prepare sets up rendering for a specific glyph at the given position
	Prepare(r *GlyphRect, x, y float64, glyph rune, flip bool)

	// Span returns the coverage data for a specific row of the prepared glyph
	Span(y int) []basics.CoverType

	// Height returns the font height
	Height() float64

	// BaseLine returns the font baseline position
	BaseLine() float64

	// Width calculates the total width of a string
	Width(str string) float64
}

// GlyphRasterBin implements binary glyph rasterization from embedded font data.
// This is the Go equivalent of AGG's glyph_raster_bin template class.
type GlyphRasterBin struct {
	font      []byte                // Font data
	bigEndian bool                  // Byte order flag
	span      [256]basics.CoverType // Span buffer for coverage data

	// Current glyph state
	bits           []byte // Current glyph bitmap data
	glyphWidth     int    // Current glyph width
	glyphByteWidth int    // Current glyph width in bytes
}

// NewGlyphRasterBin creates a new binary glyph rasterizer
func NewGlyphRasterBin(font []byte) *GlyphRasterBin {
	g := &GlyphRasterBin{
		font: font,
	}

	// Detect byte order
	test := int(1)
	if *(*byte)(unsafe.Pointer(&test)) == 0 {
		g.bigEndian = true
	}

	return g
}

// Font returns the current font data
func (g *GlyphRasterBin) Font() []byte {
	return g.font
}

// SetFont sets the font data
func (g *GlyphRasterBin) SetFont(font []byte) {
	g.font = font
}

// Height returns the font height
func (g *GlyphRasterBin) Height() float64 {
	if len(g.font) < 1 {
		return 0
	}
	return float64(g.font[0])
}

// BaseLine returns the font baseline position
func (g *GlyphRasterBin) BaseLine() float64 {
	if len(g.font) < 2 {
		return 0
	}
	return float64(g.font[1])
}

// Width calculates the total width of a string
func (g *GlyphRasterBin) Width(str string) float64 {
	if len(g.font) < 4 {
		return 0
	}

	startChar := int(g.font[2])
	numChars := int(g.font[3])

	width := 0
	for _, r := range str {
		glyph := int(r)
		if glyph < startChar || glyph >= startChar+numChars {
			continue
		}

		offset := g.getValue(g.font[4+(glyph-startChar)*2:])
		if 4+numChars*2+offset >= len(g.font) {
			continue
		}
		bits := g.font[4+numChars*2+offset:]
		if len(bits) > 0 {
			width += int(bits[0])
		}
	}
	return float64(width)
}

// Prepare sets up rendering for a specific glyph
func (g *GlyphRasterBin) Prepare(r *GlyphRect, x, y float64, glyph rune, flip bool) {
	if len(g.font) < 4 {
		r.X1, r.Y1, r.X2, r.Y2 = 1, 1, 0, 0 // Invalid rectangle
		r.DX, r.DY = 0, 0
		return
	}

	startChar := int(g.font[2])
	numChars := int(g.font[3])
	glyphInt := int(glyph)

	if glyphInt < startChar || glyphInt >= startChar+numChars {
		r.X1, r.Y1, r.X2, r.Y2 = 1, 1, 0, 0 // Invalid rectangle
		r.DX, r.DY = 0, 0
		return
	}

	offset := g.getValue(g.font[4+(glyphInt-startChar)*2:])
	if 4+numChars*2+offset >= len(g.font) {
		r.X1, r.Y1, r.X2, r.Y2 = 1, 1, 0, 0 // Invalid rectangle
		r.DX, r.DY = 0, 0
		return
	}

	g.bits = g.font[4+numChars*2+offset:]
	if len(g.bits) == 0 {
		r.X1, r.Y1, r.X2, r.Y2 = 1, 1, 0, 0 // Invalid rectangle
		r.DX, r.DY = 0, 0
		return
	}

	g.glyphWidth = int(g.bits[0])
	g.glyphByteWidth = (g.glyphWidth + 7) >> 3
	g.bits = g.bits[1:] // Skip width byte

	r.X1 = int(x)
	r.X2 = r.X1 + g.glyphWidth - 1

	if flip {
		r.Y1 = int(y) - int(g.font[0]) + int(g.font[1])
		r.Y2 = r.Y1 + int(g.font[0]) - 1
	} else {
		r.Y1 = int(y) - int(g.font[1]) + 1
		r.Y2 = r.Y1 + int(g.font[0]) - 1
	}

	r.DX = float64(g.glyphWidth)
	r.DY = 0
}

// Span returns the coverage data for a specific row
func (g *GlyphRasterBin) Span(y int) []basics.CoverType {
	if len(g.bits) == 0 || g.glyphByteWidth == 0 || len(g.font) == 0 {
		return nil
	}

	// Flip y coordinate as AGG does
	flippedY := int(g.font[0]) - y - 1
	if flippedY < 0 || flippedY >= int(g.font[0]) {
		return nil
	}

	// Calculate the offset into the bitmap data
	offset := flippedY * g.glyphByteWidth
	if offset >= len(g.bits) {
		return nil
	}

	// Clear the span buffer
	for i := 0; i < g.glyphWidth; i++ {
		g.span[i] = 0
	}

	// Extract bits and convert to coverage values
	bits := g.bits[offset:]
	val := uint8(0)
	if len(bits) > 0 {
		val = bits[0]
	}
	nb := 0

	for i := 0; i < g.glyphWidth; i++ {
		if (val & 0x80) != 0 {
			g.span[i] = basics.CoverFull
		} else {
			g.span[i] = 0
		}
		val <<= 1
		nb++
		if nb >= 8 && (i/8+1) < len(bits) {
			val = bits[i/8+1]
			nb = 0
		}
	}

	return g.span[:g.glyphWidth]
}

// getValue extracts a 16-bit value from the font data, handling endianness
func (g *GlyphRasterBin) getValue(data []byte) int {
	if len(data) < 2 {
		return 0
	}

	if g.bigEndian {
		return int(data[0])<<8 | int(data[1])
	}
	return int(data[0]) | int(data[1])<<8
}
