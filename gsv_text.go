package agg

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/gsv"
)

// GSVPathCommand is the public AGG path verb type emitted by GSVText.
type GSVPathCommand = basics.PathCommand

const (
	GSVPathCmdStop   GSVPathCommand = basics.PathCmdStop
	GSVPathCmdMoveTo GSVPathCommand = basics.PathCmdMoveTo
	GSVPathCmdLineTo GSVPathCommand = basics.PathCmdLineTo
)

// GSVText exposes AGG's built-in stroke-vector font path source without going
// through the Agg2D text engine. Callers can measure or iterate vertices and
// then render via ordinary AGG path or rasterizer APIs.
type GSVText struct {
	impl *gsv.GSVText
}

// NewGSVText creates a new public wrapper around AGG's built-in GSV path
// source.
func NewGSVText() *GSVText {
	return &GSVText{impl: gsv.NewGSVText()}
}

// SetFontData replaces the active vector-font data. Passing nil restores the
// embedded default font.
func (t *GSVText) SetFontData(font []byte) {
	if t == nil {
		return
	}
	t.ensure()
	t.impl.SetFont(font)
}

// SetFlip controls whether GSV text uses flipped Y coordinates.
func (t *GSVText) SetFlip(flip bool) {
	if t == nil {
		return
	}
	t.ensure()
	t.impl.SetFlip(flip)
}

// SetSize sets the nominal text height and optional width. A width of zero uses
// the proportional default, matching upstream AGG behavior.
func (t *GSVText) SetSize(height, width float64) {
	if t == nil {
		return
	}
	t.ensure()
	t.impl.SetSize(height, width)
}

// SetSpace sets additional inter-character spacing.
func (t *GSVText) SetSpace(space float64) {
	if t == nil {
		return
	}
	t.ensure()
	t.impl.SetSpace(space)
}

// SetLineSpace sets additional spacing between lines in multi-line text.
func (t *GSVText) SetLineSpace(lineSpace float64) {
	if t == nil {
		return
	}
	t.ensure()
	t.impl.SetLineSpace(lineSpace)
}

// SetStartPoint sets the starting text position.
func (t *GSVText) SetStartPoint(x, y float64) {
	if t == nil {
		return
	}
	t.ensure()
	t.impl.SetStartPoint(x, y)
}

// SetText sets the current text string.
func (t *GSVText) SetText(text string) {
	if t == nil {
		return
	}
	t.ensure()
	t.impl.SetText(text)
}

// MeasureText returns the rendered width of str without changing caller state.
func (t *GSVText) MeasureText(str string) float64 {
	if t == nil {
		return 0
	}
	t.ensure()
	return t.impl.MeasureText(str)
}

// TextWidth returns the width of the currently configured text.
func (t *GSVText) TextWidth() float64 {
	if t == nil {
		return 0
	}
	t.ensure()
	return t.impl.TextWidth()
}

// Rewind resets vertex iteration for the current text.
func (t *GSVText) Rewind(pathID uint) {
	if t == nil {
		return
	}
	t.ensure()
	t.impl.Rewind(pathID)
}

// Vertex returns the next GSV path vertex.
func (t *GSVText) Vertex() (x, y float64, cmd GSVPathCommand) {
	if t == nil {
		return 0, 0, GSVPathCmdStop
	}
	t.ensure()
	return t.impl.Vertex()
}

func (t *GSVText) ensure() {
	if t.impl == nil {
		t.impl = gsv.NewGSVText()
	}
}
