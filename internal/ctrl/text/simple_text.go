// Package text provides simplified text rendering for AGG controls.
// This is a temporary solution until GSVText is fully implemented.
package text

import (
	"agg_go/internal/basics"
	"agg_go/internal/path"
)

// SimpleText provides basic vector text rendering using path storage.
// This is a simplified replacement for AGG's GSVText until the full
// Gouraud Shaded Vector text system is implemented.
type SimpleText struct {
	storage   *path.PathStorage
	x, y      float64
	size      float64
	text      string
	thickness float64
}

// NewSimpleText creates a new simple text renderer.
func NewSimpleText() *SimpleText {
	return &SimpleText{
		storage:   path.NewPathStorage(),
		x:         0.0,
		y:         0.0,
		size:      10.0,
		text:      "",
		thickness: 1.0,
	}
}

// SetText sets the text to be rendered.
func (st *SimpleText) SetText(text string) {
	st.text = text
	st.generatePath()
}

// SetPosition sets the text position.
func (st *SimpleText) SetPosition(x, y float64) {
	st.x = x
	st.y = y
	st.generatePath()
}

// SetSize sets the text size.
func (st *SimpleText) SetSize(size float64) {
	st.size = size
	st.generatePath()
}

// SetThickness sets the text thickness (stroke width).
func (st *SimpleText) SetThickness(thickness float64) {
	st.thickness = thickness
}

// GetStorage returns the path storage containing the text geometry.
func (st *SimpleText) GetStorage() *path.PathStorage {
	return st.storage
}

// Thickness returns the current text thickness.
func (st *SimpleText) Thickness() float64 {
	return st.thickness
}

// generatePath creates the vector path for the text.
// This is a very simplified implementation that renders basic shapes for common characters.
func (st *SimpleText) generatePath() {
	st.storage.RemoveAll()

	if st.text == "" {
		return
	}

	charWidth := st.size * 0.6 // Character width relative to size
	spacing := st.size * 0.1   // Character spacing

	currentX := st.x

	for _, char := range st.text {
		st.renderChar(char, currentX, st.y, st.size)
		currentX += charWidth + spacing
	}
}

// renderChar renders a single character at the specified position.
// This is a very basic implementation with simple geometric shapes.
func (st *SimpleText) renderChar(char rune, x, y, size float64) {
	w := size * 0.5 // Character width
	h := size       // Character height

	switch char {
	case ' ':
		// Space - do nothing

	case '0':
		st.renderRect(x, y, w, h, true) // Hollow rectangle

	case '1':
		st.renderLine(x+w/2, y, x+w/2, y+h)

	case '2':
		st.renderLine(x, y+h/2, x+w, y+h/2)
		st.renderLine(x, y+h, x+w, y+h)
		st.renderLine(x+w, y+h/2, x+w, y)
		st.renderLine(x+w, y, x, y)
		st.renderLine(x, y, x, y+h/2)

	case '3':
		st.renderLine(x, y, x+w, y)
		st.renderLine(x+w, y, x+w, y+h/2)
		st.renderLine(x+w/2, y+h/2, x+w, y+h/2)
		st.renderLine(x+w, y+h/2, x+w, y+h)
		st.renderLine(x+w, y+h, x, y+h)

	case '4':
		st.renderLine(x, y, x, y+h/2)
		st.renderLine(x, y+h/2, x+w, y+h/2)
		st.renderLine(x+w, y, x+w, y+h)

	case '5':
		st.renderLine(x, y, x+w, y)
		st.renderLine(x, y, x, y+h/2)
		st.renderLine(x, y+h/2, x+w, y+h/2)
		st.renderLine(x+w, y+h/2, x+w, y+h)
		st.renderLine(x+w, y+h, x, y+h)

	case '6':
		st.renderLine(x, y, x+w, y)
		st.renderLine(x, y, x, y+h)
		st.renderLine(x, y+h, x+w, y+h)
		st.renderLine(x+w, y+h, x+w, y+h/2)
		st.renderLine(x+w, y+h/2, x, y+h/2)

	case '7':
		st.renderLine(x, y, x+w, y)
		st.renderLine(x+w, y, x+w/2, y+h)

	case '8':
		st.renderRect(x, y, w, h/2, true)
		st.renderRect(x, y+h/2, w, h/2, true)

	case '9':
		st.renderLine(x, y, x+w, y)
		st.renderLine(x, y, x, y+h/2)
		st.renderLine(x, y+h/2, x+w, y+h/2)
		st.renderLine(x+w, y, x+w, y+h)

	case '.':
		// Small dot
		st.renderRect(x+w/2-size*0.05, y+h-size*0.1, size*0.1, size*0.1, false)

	case '-':
		st.renderLine(x, y+h/2, x+w, y+h/2)

	case ':':
		// Two dots vertically aligned
		dotSize := size * 0.1
		st.renderRect(x+w/2-dotSize/2, y+h/3-dotSize/2, dotSize, dotSize, false)
		st.renderRect(x+w/2-dotSize/2, y+2*h/3-dotSize/2, dotSize, dotSize, false)

	case '%':
		// Simple % symbol - diagonal line with dots
		st.renderLine(x, y+h, x+w, y)
		dotSize := size * 0.15
		st.renderRect(x, y, dotSize, dotSize, false)
		st.renderRect(x+w-dotSize, y+h-dotSize, dotSize, dotSize, false)

	case 'V', 'v':
		// V shape
		st.renderLine(x, y, x+w/2, y+h)
		st.renderLine(x+w/2, y+h, x+w, y)

	case 'a', 'A':
		// A shape
		st.renderLine(x, y+h, x+w/2, y)
		st.renderLine(x+w/2, y, x+w, y+h)
		st.renderLine(x+w/4, y+h/2, x+3*w/4, y+h/2)

	case 'l', 'L':
		// L shape
		st.renderLine(x, y, x, y+h)
		st.renderLine(x, y+h, x+w, y+h)

	case 'u', 'U':
		// U shape
		st.renderLine(x, y, x, y+h)
		st.renderLine(x, y+h, x+w, y+h)
		st.renderLine(x+w, y+h, x+w, y)

	case 'e', 'E':
		// E shape
		st.renderLine(x, y, x, y+h)
		st.renderLine(x, y, x+w, y)
		st.renderLine(x, y+h/2, x+3*w/4, y+h/2)
		st.renderLine(x, y+h, x+w, y+h)

	case 'r', 'R':
		// R shape (simplified)
		st.renderLine(x, y, x, y+h)
		st.renderLine(x, y, x+w, y)
		st.renderLine(x+w, y, x+w, y+h/2)
		st.renderLine(x+w, y+h/2, x, y+h/2)
		st.renderLine(x+w/2, y+h/2, x+w, y+h)

	case 's', 'S':
		// S shape (simplified)
		st.renderLine(x+w, y, x, y)
		st.renderLine(x, y, x, y+h/2)
		st.renderLine(x, y+h/2, x+w, y+h/2)
		st.renderLine(x+w, y+h/2, x+w, y+h)
		st.renderLine(x+w, y+h, x, y+h)

	case 't', 'T':
		// T shape
		st.renderLine(x, y, x+w, y)
		st.renderLine(x+w/2, y, x+w/2, y+h)

	case 'i', 'I':
		// I shape
		st.renderLine(x, y, x+w, y)
		st.renderLine(x+w/2, y, x+w/2, y+h)
		st.renderLine(x, y+h, x+w, y+h)

	case 'n', 'N':
		// N shape
		st.renderLine(x, y, x, y+h)
		st.renderLine(x, y, x+w, y+h)
		st.renderLine(x+w, y+h, x+w, y)

	case 'o', 'O':
		// O shape (rectangle outline)
		st.renderRect(x, y, w, h, true)

	case 'f', 'F':
		// F shape
		st.renderLine(x, y, x, y+h)
		st.renderLine(x, y, x+w, y)
		st.renderLine(x, y+h/2, x+3*w/4, y+h/2)

	case 'c', 'C':
		// C shape
		st.renderLine(x+w, y, x, y)
		st.renderLine(x, y, x, y+h)
		st.renderLine(x, y+h, x+w, y+h)

	default:
		// For unknown characters, render a simple rectangle
		st.renderRect(x, y, w, h, true)
	}
}

// renderLine adds a line to the path storage.
func (st *SimpleText) renderLine(x1, y1, x2, y2 float64) {
	st.storage.MoveTo(x1, y1)
	st.storage.LineTo(x2, y2)
}

// renderRect adds a rectangle to the path storage.
func (st *SimpleText) renderRect(x, y, w, h float64, hollow bool) {
	st.storage.MoveTo(x, y)
	st.storage.LineTo(x+w, y)
	st.storage.LineTo(x+w, y+h)
	st.storage.LineTo(x, y+h)
	st.storage.LineTo(x, y) // Close the path manually

	// If hollow, we just have the outline
	// If filled, we'd need to add fill logic, but for stroked text this is fine
}

// Rewind implements the vertex source interface.
func (st *SimpleText) Rewind(pathID uint) {
	st.storage.Rewind(pathID)
}

// Vertex implements the vertex source interface.
func (st *SimpleText) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmdUint := st.storage.NextVertex()
	return x, y, basics.PathCommand(cmdUint)
}

// EstimateTextWidth estimates the width of text when rendered.
func (st *SimpleText) EstimateTextWidth(text string) float64 {
	if text == "" {
		return 0.0
	}

	charWidth := st.size * 0.6
	spacing := st.size * 0.1
	return float64(len(text))*(charWidth+spacing) - spacing
}

// EstimateTextHeight estimates the height of text when rendered.
func (st *SimpleText) EstimateTextHeight() float64 {
	return st.size
}
