// Package gsv provides Gouraud Shaded Vector text rendering support.
// This is a direct port of AGG's gsv_text functionality with vector font support.
package gsv

import (
	"encoding/binary"
	"os"
	"unsafe"

	"agg_go/internal/basics"
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
	status    Status
	bigEndian bool
	flip      bool // Flip Y coordinate

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
	gsv := &GSVText{
		x:         0.0,
		y:         0.0,
		startX:    0.0,
		width:     10.0,
		height:    0.0,
		space:     0.0,
		lineSpace: 0.0,
		text:      "",
		curChar:   0,
		font:      GSVDefaultFont,
		status:    StatusInitial,
		bigEndian: isBigEndian(),
		flip:      false,
	}
	return gsv
}

// SetFont sets the font data to use for rendering.
// If font is nil, the default embedded font is used.
func (gsv *GSVText) SetFont(font []byte) {
	if font != nil {
		gsv.font = font
	} else if len(gsv.loadedFont) > 0 {
		gsv.font = gsv.loadedFont
	} else {
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

// TextWidth calculates the width of the current text when rendered.
// This uses a simplified calculation based on the font's base dimensions.
func (gsv *GSVText) TextWidth() float64 {
	if gsv.text == "" || len(gsv.font) == 0 {
		return 0.0
	}

	// Parse font header to get base height
	baseHeight := gsv.getValue16(gsv.font[4:])
	if baseHeight == 0 {
		return 0.0
	}

	// Calculate scaling factors
	h := gsv.height / float64(baseHeight)
	w := gsv.width
	if w == 0.0 {
		w = h
	} else {
		w = w / float64(baseHeight)
	}

	// Estimate width based on character count and average character width
	// This is a simplified calculation - AGG does full vertex traversal
	avgCharWidth := w * float64(baseHeight) * 0.6 // Approximate average character width
	totalSpacing := gsv.space * float64(len(gsv.text)-1)

	return avgCharWidth*float64(len(gsv.text)) + totalSpacing
}

// Rewind resets the renderer to begin generating vertices for the current text.
func (gsv *GSVText) Rewind(pathID uint) {
	gsv.status = StatusInitial
	if len(gsv.font) == 0 {
		return
	}

	// Parse font header
	gsv.indices = gsv.font[gsv.getValue16(gsv.font[:]):] // Skip to index table
	baseHeight := gsv.getValue16(gsv.font[4:])
	gsv.glyphs = gsv.indices[257*2:] // Skip 256 character indices + null terminator

	// Calculate scaling factors
	gsv.h = gsv.height / float64(baseHeight)
	if gsv.width == 0.0 {
		gsv.w = gsv.h
	} else {
		gsv.w = gsv.width / float64(baseHeight)
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

		case StatusNextChar:
			if gsv.curChar >= len(gsv.text) {
				return 0, 0, basics.PathCmdStop
			}

			char := gsv.text[gsv.curChar]
			gsv.curChar++

			if char == '\n' {
				// Handle newline
				gsv.x = gsv.startX
				if gsv.flip {
					gsv.y -= -(gsv.height + gsv.lineSpace)
				} else {
					gsv.y += gsv.height + gsv.lineSpace
				}
				continue
			}

			// Get character index (0-255)
			charIndex := int(char) & 0xFF

			// Get glyph data offsets from index table
			startOffset := gsv.getValue16(gsv.indices[charIndex*2:])
			endOffset := gsv.getValue16(gsv.indices[charIndex*2+2:])

			gsv.beginGlyph = gsv.glyphs[startOffset:]
			gsv.endGlyph = gsv.glyphs[endOffset:]
			gsv.glyphOffset = 0
			gsv.status = StatusStartGlyph

		case StatusStartGlyph:
			x = gsv.x
			y = gsv.y
			gsv.status = StatusGlyph
			return x, y, basics.PathCmdMoveTo

		case StatusGlyph:
			// Check if we've processed all glyph data
			if gsv.glyphOffset >= len(gsv.beginGlyph) ||
				(len(gsv.endGlyph) > 0 && gsv.glyphOffset >= (len(gsv.beginGlyph)-len(gsv.endGlyph))) {
				gsv.status = StatusNextChar
				gsv.x += gsv.space
				continue
			}

			// Read coordinate delta (dx, dy) from glyph data
			if gsv.glyphOffset+1 >= len(gsv.beginGlyph) {
				gsv.status = StatusNextChar
				gsv.x += gsv.space
				continue
			}

			dx := int8(gsv.beginGlyph[gsv.glyphOffset])
			yc := int8(gsv.beginGlyph[gsv.glyphOffset+1])
			gsv.glyphOffset += 2

			// Check move/line flag (high bit of Y coordinate)
			yf := (yc & -128) != 0 // -128 is 0x80 as int8
			yc = (yc << 1) >> 1    // Sign extend after clearing flag bit
			dy := int(yc)

			// Apply deltas with scaling
			gsv.x += float64(dx) * gsv.w
			gsv.y += float64(dy) * gsv.h

			x = gsv.x
			y = gsv.y

			if yf {
				return x, y, basics.PathCmdMoveTo
			} else {
				return x, y, basics.PathCmdLineTo
			}
		}
	}
}

// getValue16 reads a 16-bit value from font data, handling endianness.
func (gsv *GSVText) getValue16(data []byte) uint16 {
	if len(data) < 2 {
		return 0
	}

	if gsv.bigEndian {
		return binary.BigEndian.Uint16(data)
	} else {
		return binary.LittleEndian.Uint16(data)
	}
}

// isBigEndian detects system endianness.
func isBigEndian() bool {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return b == 0x01
}
