// Package gsv provides Gouraud Shaded Vector text rendering support.
// This file implements the GSVTextOutline wrapper for stroked text rendering.
package gsv

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// GSVTextOutline wraps GSVText with stroke and transformation support.
// This is equivalent to AGG's gsv_text_outline template class:
//
//	conv_stroke<gsv_text>                         m_polyline
//	conv_transform<conv_stroke<gsv_text>, Trans>  m_trans
type GSVTextOutline struct {
	text      *GSVText
	stroke    *conv.ConvStroke
	transform *transform.TransAffine
	width     float64
}

// NewGSVTextOutline creates a new GSV text outline renderer.
// GSVText is used directly as the source for ConvStroke, matching the C++ design.
func NewGSVTextOutline(text *GSVText) *GSVTextOutline {
	if text == nil {
		text = NewGSVText()
	}

	return &GSVTextOutline{
		text:      text,
		stroke:    conv.NewConvStroke(text),
		transform: transform.NewTransAffine(),
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
// Matches C++ gsv_text_outline::rewind which sets round join/cap then rewinds.
func (gto *GSVTextOutline) Rewind(pathID uint) {
	gto.stroke.SetWidth(gto.width)
	gto.stroke.SetLineJoin(basics.RoundJoin)
	gto.stroke.SetLineCap(basics.RoundCap)
	gto.stroke.Rewind(pathID)
}

// Vertex generates the next vertex in the stroked and transformed text path.
func (gto *GSVTextOutline) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmd = gto.stroke.Vertex()
	if !gto.transform.IsIdentity(transform.AffineEpsilon) {
		gto.transform.Transform(&x, &y)
	}
	return x, y, cmd
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
		filled:  false,
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
		gts.outline.text.Rewind(pathID)
	} else {
		gts.outline.Rewind(pathID)
	}
}

// Vertex generates the next vertex in the text path.
func (gts *GSVTextSimple) Vertex() (x, y float64, cmd basics.PathCommand) {
	if gts.filled {
		x, y, cmd = gts.outline.text.Vertex()
		if !gts.outline.transform.IsIdentity(transform.AffineEpsilon) {
			gts.outline.transform.Transform(&x, &y)
		}
		return x, y, cmd
	}
	return gts.outline.Vertex()
}

// EstimateWidth returns the rendered width of the current text.
func (gts *GSVTextSimple) EstimateWidth() float64 {
	return gts.outline.text.TextWidth()
}
