// Package text provides text rendering for AGG controls using GSVText + ConvStroke,
// matching the C++ ctrl implementation: conv_stroke<gsv_text>.
package text

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
)

// SimpleText wraps GSVText + ConvStroke, exactly as the C++ ctrl does with
// conv_stroke<gsv_text>. Vertex() returns the stroked (filled) polygon path.
type SimpleText struct {
	gsvText   *gsv.GSVText
	stroke    *conv.ConvStroke
	thickness float64
	size      float64
}

// NewSimpleText creates a new text renderer backed by GSVText + ConvStroke.
func NewSimpleText() *SimpleText {
	t := gsv.NewGSVText()
	s := conv.NewConvStroke(t)
	s.SetLineJoin(basics.RoundJoin)
	s.SetLineCap(basics.RoundCap)
	s.SetWidth(1.0)
	return &SimpleText{
		gsvText:   t,
		stroke:    s,
		thickness: 1.0,
		size:      10.0,
	}
}

// SetText sets the text string.
func (st *SimpleText) SetText(text string) {
	st.gsvText.SetText(text)
}

// SetPosition sets the text start position.
func (st *SimpleText) SetPosition(x, y float64) {
	st.gsvText.SetStartPoint(x, y)
}

// SetSize sets the text height and optional width.
// If width is omitted, proportional width (0) is used.
func (st *SimpleText) SetSize(size float64, width ...float64) {
	st.size = size
	w := 0.0
	if len(width) > 0 {
		w = width[0]
	}
	st.gsvText.SetSize(size, w)
}

// SetThickness sets the stroke width used to render the text.
func (st *SimpleText) SetThickness(thickness float64) {
	st.thickness = thickness
	st.stroke.SetWidth(thickness)
}

// Thickness returns the current stroke width.
func (st *SimpleText) Thickness() float64 {
	return st.thickness
}

// Rewind prepares the text for vertex iteration.
func (st *SimpleText) Rewind(pathID uint) {
	st.stroke.Rewind(0)
}

// Vertex returns successive vertices of the stroked text path.
func (st *SimpleText) Vertex() (x, y float64, cmd basics.PathCommand) {
	return st.stroke.Vertex()
}

// EstimateTextWidth returns the rendered width of the given text string.
func (st *SimpleText) EstimateTextWidth(text string) float64 {
	st.gsvText.SetText(text)
	return st.gsvText.TextWidth()
}

// EstimateTextHeight returns the text height.
func (st *SimpleText) EstimateTextHeight() float64 {
	return st.size
}
