// Package gsv provides Gouraud Shaded Vector text rendering support.
// This is a direct port of AGG's gsv_text functionality with vector font support.
package gsv

import (
	"encoding/binary"
	"os"

	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// Status represents the internal state of the GSV text renderer.
type Status int

const (
	StatusInitial Status = iota
	StatusNextChar
	StatusStartGlyph
	StatusGlyph
)

// GSVText implements AGG's gsv_text class - a vector-based text renderer.
// It supports ASCII characters 32-127 using embedded vector font data.
// Text is rendered as path vertices that can be stroked or filled.
type GSVText struct {
	// Current position and configuration
	x, y      float64 // Current cursor position
	startX    float64 // Line start X position for newlines
	width     float64 // Character width scaling
	height    float64 // Character height
	space     float64 // Additional character spacing
	lineSpace float64 // Additional line spacing

	// Text content
	text    string
	curChar int // Current character index

	// Font data
	font       []byte // Current font data
	loadedFont []byte // Font data loaded from file

	// Rendering state
	status Status
	flip   bool // Flip Y coordinate

	// Font parsing data
	indices     []byte  // Character index table
	glyphs      []byte  // Glyph coordinate data
	beginGlyph  []byte  // Current glyph start
	endGlyph    []byte  // Current glyph end
	w, h        float64 // Scaled width and height factors
	glyphOffset int     // Current offset in glyph data
}

// NewGSVText creates a new GSV text renderer with default settings.
func NewGSVText() *GSVText {
	return &GSVText{
		x:      0.0,
		y:      0.0,
		startX: 0.0,
		width:  10.0,
		height: 0.0,
		font:   GSVDefaultFont,
		status: StatusInitial,
	}
}

// SetFont sets the font data to use for rendering.
// If font is nil, the default embedded font is used.
func (gsv *GSVText) SetFont(font []byte) {
	switch {
	case font != nil:
		gsv.font = font
	case len(gsv.loadedFont) > 0:
		gsv.font = gsv.loadedFont
	default:
		gsv.font = GSVDefaultFont
	}
}

// SetFlip controls Y coordinate flipping for different coordinate systems.
func (gsv *GSVText) SetFlip(flip bool) {
	gsv.flip = flip
}

// LoadFont loads font data from a file.
func (gsv *GSVText) LoadFont(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	gsv.loadedFont = data
	gsv.font = data
	return nil
}

// SetSize sets the text height and optional width.
// If width is 0, it defaults to the same as height.
func (gsv *GSVText) SetSize(height, width float64) {
	gsv.height = height
	gsv.width = width
}

// SetSpace sets additional character spacing.
func (gsv *GSVText) SetSpace(space float64) {
	gsv.space = space
}

// SetLineSpace sets additional line spacing for multi-line text.
func (gsv *GSVText) SetLineSpace(lineSpace float64) {
	gsv.lineSpace = lineSpace
}

// SetStartPoint sets the starting position for text rendering.
func (gsv *GSVText) SetStartPoint(x, y float64) {
	gsv.x = x
	gsv.startX = x
	gsv.y = y
}

// SetText sets the text string to be rendered.
func (gsv *GSVText) SetText(text string) {
	gsv.text = text
}

// MeasureText returns the rendered width of str without modifying the current text state.
// Unlike TextWidth, this is safe to call at any time without affecting gsv.text.
func (gsv *GSVText) MeasureText(str string) float64 {
	saved := gsv.text
	gsv.text = str
	w := gsv.TextWidth()
	gsv.text = saved
	return w
}

// TextWidth calculates the rendered width of the current text using actual bounding rect.
// This matches the C++ text_width() which calls bounding_rect_single internally.
// Note: this method modifies the internal rendering state (same as C++).
func (gsv *GSVText) TextWidth() float64 {
	if gsv.text == "" || len(gsv.font) == 0 {
		return 0.0
	}

	// Save state that gets modified during vertex generation
	savedX, savedY := gsv.x, gsv.y
	savedCurChar := gsv.curChar
	savedStatus := gsv.status
	savedGlyphOffset := gsv.glyphOffset
	savedBeginGlyph := gsv.beginGlyph
	savedEndGlyph := gsv.endGlyph
	savedIndices := gsv.indices
	savedGlyphs := gsv.glyphs
	savedW, savedH := gsv.w, gsv.h

	gsv.Rewind(0)

	var x1, y1, x2, y2 float64
	first := true

	for {
		x, y, cmd := gsv.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsVertex(cmd) {
			if first {
				x1, y1, x2, y2 = x, y, x, y
				first = false
			} else {
				if x < x1 {
					x1 = x
				}
				if y < y1 {
					y1 = y
				}
				if x > x2 {
					x2 = x
				}
				if y > y2 {
					y2 = y
				}
			}
		}
	}

	// Restore state
	gsv.x, gsv.y = savedX, savedY
	gsv.curChar = savedCurChar
	gsv.status = savedStatus
	gsv.glyphOffset = savedGlyphOffset
	gsv.beginGlyph = savedBeginGlyph
	gsv.endGlyph = savedEndGlyph
	gsv.indices = savedIndices
	gsv.glyphs = savedGlyphs
	gsv.w, gsv.h = savedW, savedH

	if first {
		return 0.0
	}
	_ = y1 // suppress unused warning
	return x2 - x1
}

// Rewind resets the renderer to begin generating vertices for the current text.
func (gsv *GSVText) Rewind(pathID uint) {
	gsv.status = StatusInitial
	if len(gsv.font) == 0 {
		return
	}

	// Parse font header:
	// font[0:2] = uint16 offset to index table
	// font[4:6] = uint16 base height
	baseHeight := float64(gsv.getValue16(gsv.font[4:]))
	gsv.indices = gsv.font[gsv.getValue16(gsv.font):]
	gsv.glyphs = gsv.indices[257*2:]

	// Calculate scaling factors
	gsv.h = gsv.height / baseHeight
	if gsv.width == 0.0 {
		gsv.w = gsv.h
	} else {
		gsv.w = gsv.width / baseHeight
	}

	if gsv.flip {
		gsv.h = -gsv.h
	}

	gsv.curChar = 0
}

// Vertex generates the next vertex in the text path.
// Returns the vertex coordinates and command type.
func (gsv *GSVText) Vertex() (x, y float64, cmd basics.PathCommand) {
	for {
		switch gsv.status {
		case StatusInitial:
			if len(gsv.font) == 0 {
				return 0, 0, basics.PathCmdStop
			}
			gsv.status = StatusNextChar
			// fall through to StatusNextChar

		case StatusNextChar:
			if gsv.curChar >= len(gsv.text) {
				return 0, 0, basics.PathCmdStop
			}

			char := gsv.text[gsv.curChar]
			gsv.curChar++

			if char == '\n' {
				gsv.x = gsv.startX
				// C++: m_y -= m_flip ? -m_height - m_line_space : m_height + m_line_space
				if gsv.flip {
					gsv.y += gsv.height + gsv.lineSpace
				} else {
					gsv.y -= gsv.height + gsv.lineSpace
				}
				continue
			}

			// idx is the char value, shifted left by 1 to get the byte offset into
			// the 16-bit index table (each entry is 2 bytes).
			idx := int(char&0xFF) << 1
			startOffset := gsv.getValue16(gsv.indices[idx:])
			endOffset := gsv.getValue16(gsv.indices[idx+2:])

			gsv.beginGlyph = gsv.glyphs[startOffset:]
			gsv.endGlyph = gsv.glyphs[endOffset:]
			gsv.glyphOffset = 0
			gsv.status = StatusStartGlyph
			// fall through to StatusStartGlyph

		case StatusStartGlyph:
			x = gsv.x
			y = gsv.y
			gsv.status = StatusGlyph
			return x, y, basics.PathCmdMoveTo

		case StatusGlyph:
			// Check if we've consumed all bytes for this glyph.
			// beginGlyph starts at startOffset; endGlyph starts at endOffset.
			// len(beginGlyph) - len(endGlyph) = endOffset - startOffset = glyph byte count.
			if gsv.glyphOffset >= len(gsv.beginGlyph)-len(gsv.endGlyph) {
				gsv.status = StatusNextChar
				gsv.x += gsv.space
				continue
			}

			// Read signed delta x and y. Each glyph vertex is 2 bytes: (dx, yc).
			// The high bit of yc is the pen-up/move flag; the remaining 7 bits
			// form a signed value via arithmetic shift (matches C++ int8 behavior).
			dx := int(int8(gsv.beginGlyph[gsv.glyphOffset]))
			yc := int8(gsv.beginGlyph[gsv.glyphOffset+1])
			gsv.glyphOffset += 2

			yf := yc & -128     // 0x80: pen-up (move) flag
			yc = (yc << 1) >> 1 // sign-extend 7-bit value
			dy := int(yc)

			gsv.x += float64(dx) * gsv.w
			gsv.y += float64(dy) * gsv.h
			x = gsv.x
			y = gsv.y

			if yf != 0 {
				return x, y, basics.PathCmdMoveTo
			}
			return x, y, basics.PathCmdLineTo
		}
	}
}

// getValue16 reads a 16-bit little-endian value from font data.
// The GSV font format always stores 16-bit values in little-endian byte order.
func (gsv *GSVText) getValue16(data []byte) uint16 {
	if len(data) < 2 {
		return 0
	}
	return binary.LittleEndian.Uint16(data)
}
