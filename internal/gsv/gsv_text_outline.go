// Package gsv provides Gouraud Shaded Vector text rendering support.
// This file implements the GSVTextOutline wrapper for stroked text rendering.
package gsv

import (
	"agg_go/internal/basics"
	"agg_go/internal/conv"
	"agg_go/internal/transform"
)

// GSVTextOutline wraps GSVText with stroke and transformation support.
// This is equivalent to AGG's gsv_text_outline template class.
type GSVTextOutline struct {
	text      *GSVText
	stroke    *conv.ConvStroke
	transform *transform.TransAffine
	width     float64
}

// NewGSVTextOutline creates a new GSV text outline renderer.
// It combines vector text rendering with stroke conversion and optional transformation.
func NewGSVTextOutline(text *GSVText) *GSVTextOutline {
	if text == nil {
		text = NewGSVText()
	}

	return &GSVTextOutline{
		text:      text,
		stroke:    conv.NewConvStroke(&gsvTextAdapter{text: text}),
		transform: transform.NewTransAffine(), // Identity transform
		width:     1.0,
	}
}

// NewGSVTextOutlineWithTransform creates a new GSV text outline renderer with a custom transform.
func NewGSVTextOutlineWithTransform(text *GSVText, trans *transform.TransAffine) *GSVTextOutline {
	outline := NewGSVTextOutline(text)
	if trans != nil {
		outline.transform = trans
	}
	return outline
}

// SetWidth sets the stroke width for the text outline.
func (gto *GSVTextOutline) SetWidth(width float64) {
	gto.width = width
	gto.stroke.SetWidth(width)
}

// Width returns the current stroke width.
func (gto *GSVTextOutline) Width() float64 {
	return gto.width
}

// SetTransform sets the transformation matrix for the text.
func (gto *GSVTextOutline) SetTransform(trans *transform.TransAffine) {
	if trans != nil {
		gto.transform = trans
	}
}

// Transform returns the current transformation matrix.
func (gto *GSVTextOutline) Transform() *transform.TransAffine {
	return gto.transform
}

// Text returns the underlying GSVText renderer for direct configuration.
func (gto *GSVTextOutline) Text() *GSVText {
	return gto.text
}

// Stroke returns the underlying stroke converter for advanced configuration.
func (gto *GSVTextOutline) Stroke() *conv.ConvStroke {
	return gto.stroke
}

// Rewind resets the outline renderer to begin generating stroked text vertices.
func (gto *GSVTextOutline) Rewind(pathID uint) {
	// Configure stroke properties
	gto.stroke.SetWidth(gto.width)
	gto.stroke.SetLineJoin(basics.RoundJoin)
	gto.stroke.SetLineCap(basics.RoundCap)

	// Start the stroke conversion - this will internally rewind the source
	gto.stroke.Rewind(pathID)
}

// Vertex generates the next vertex in the stroked and transformed text path.
func (gto *GSVTextOutline) Vertex() (x, y float64, cmd basics.PathCommand) {
	// Get the next stroked vertex
	x, y, cmd = gto.stroke.Vertex()

	// Apply transformation if not identity
	if !gto.transform.IsIdentity(transform.AffineEpsilon) {
		gto.transform.Transform(&x, &y)
	}

	return x, y, cmd
}

// gsvTextAdapter adapts GSVText to work with ConvStroke by implementing VertexSource.
// It converts individual line segments (MoveTo + LineTo) into continuous paths suitable for stroking.
type gsvTextAdapter struct {
	text      *GSVText
	vertices  []VertexCmd
	vertexIdx int
	processed bool
}

// VertexCmd holds a vertex with its command
type VertexCmd struct {
	X, Y float64
	Cmd  basics.PathCommand
}

// Rewind implements the VertexSource interface.
func (adapter *gsvTextAdapter) Rewind(pathID uint) {
	if !adapter.processed {
		adapter.processVertices()
	}
	adapter.vertexIdx = 0
}

// processVertices converts GSV text line segments into paths suitable for stroking
func (adapter *gsvTextAdapter) processVertices() {
	adapter.vertices = nil
	adapter.text.Rewind(0)

	var pendingMoveX, pendingMoveY float64
	var hasPendingMove bool

	for {
		x, y, cmd := adapter.text.Vertex()
		if basics.IsStop(cmd) {
			break
		}

		if basics.IsMoveTo(cmd) {
			// Store the MoveTo for when we get the LineTo
			pendingMoveX, pendingMoveY = x, y
			hasPendingMove = true
		} else if basics.IsLineTo(cmd) && hasPendingMove {
			// We have a MoveTo-LineTo pair, create a complete path
			adapter.vertices = append(adapter.vertices, VertexCmd{X: pendingMoveX, Y: pendingMoveY, Cmd: basics.PathCmdMoveTo})
			adapter.vertices = append(adapter.vertices, VertexCmd{X: x, Y: y, Cmd: basics.PathCmdLineTo})
			// Add EndPoly to mark end of this sub-path
			adapter.vertices = append(adapter.vertices, VertexCmd{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly})
			hasPendingMove = false
		}
	}

	adapter.processed = true
}

// Vertex implements the VertexSource interface.
func (adapter *gsvTextAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	if adapter.vertexIdx >= len(adapter.vertices) {
		return 0, 0, basics.PathCmdStop
	}

	vertex := adapter.vertices[adapter.vertexIdx]
	adapter.vertexIdx++
	return vertex.X, vertex.Y, vertex.Cmd
}

// GSVTextSimple provides a simplified interface for common text rendering tasks.
// It combines GSVText and GSVTextOutline with sensible defaults.
type GSVTextSimple struct {
	outline *GSVTextOutline
	filled  bool // If true, render filled text; if false, render stroked
}

// NewGSVTextSimple creates a simplified GSV text renderer.
func NewGSVTextSimple() *GSVTextSimple {
	text := NewGSVText()
	outline := NewGSVTextOutline(text)

	return &GSVTextSimple{
		outline: outline,
		filled:  false, // Default to stroked text
	}
}

// SetText sets the text to render.
func (gts *GSVTextSimple) SetText(text string) {
	gts.outline.text.SetText(text)
}

// SetPosition sets the text position.
func (gts *GSVTextSimple) SetPosition(x, y float64) {
	gts.outline.text.SetStartPoint(x, y)
}

// SetSize sets the text size (height and optional width).
func (gts *GSVTextSimple) SetSize(height, width float64) {
	gts.outline.text.SetSize(height, width)
}

// SetStrokeWidth sets the stroke width (only applies to stroked text).
func (gts *GSVTextSimple) SetStrokeWidth(width float64) {
	gts.outline.SetWidth(width)
}

// SetFilled sets whether to render filled (true) or stroked (false) text.
func (gts *GSVTextSimple) SetFilled(filled bool) {
	gts.filled = filled
}

// SetTransform applies a transformation to the text.
func (gts *GSVTextSimple) SetTransform(trans *transform.TransAffine) {
	gts.outline.SetTransform(trans)
}

// SetSpacing sets character and line spacing.
func (gts *GSVTextSimple) SetSpacing(charSpace, lineSpace float64) {
	gts.outline.text.SetSpace(charSpace)
	gts.outline.text.SetLineSpace(lineSpace)
}

// Rewind resets the renderer for vertex generation.
func (gts *GSVTextSimple) Rewind(pathID uint) {
	if gts.filled {
		// For filled text, use the raw GSVText output
		gts.outline.text.Rewind(pathID)
	} else {
		// For stroked text, use the outline
		gts.outline.Rewind(pathID)
	}
}

// Vertex generates the next vertex in the text path.
func (gts *GSVTextSimple) Vertex() (x, y float64, cmd basics.PathCommand) {
	if gts.filled {
		x, y, cmd = gts.outline.text.Vertex()

		// Apply transformation if not identity
		if !gts.outline.transform.IsIdentity(transform.AffineEpsilon) {
			gts.outline.transform.Transform(&x, &y)
		}

		return x, y, cmd
	} else {
		return gts.outline.Vertex()
	}
}

// EstimateWidth estimates the rendered width of the current text.
func (gts *GSVTextSimple) EstimateWidth() float64 {
	return gts.outline.text.TextWidth()
}
