package agg

import (
	"github.com/cwbudde/agg_go/internal/font"
	"github.com/cwbudde/agg_go/internal/font/freetype"
	"github.com/cwbudde/agg_go/internal/path"
)

// FreeTypeOutlineText exposes FreeType-backed outline glyph paths without
// going through Agg2D's text engine. Callers can measure text or iterate the
// full outline path and then render it through ordinary AGG path operations.
type FreeTypeOutlineText struct {
	engine     *freetype.FontEngineFreetype
	cache      *font.FontCacheManager
	path       *path.PathStorageStl
	text       string
	startX     float64
	startY     float64
	height     float64
	width      float64
	resolution uint
	hinting    bool
	flip       bool
	dirty      bool
}

// NewFreeTypeOutlineText creates a public outline-only FreeType text source.
func NewFreeTypeOutlineText() (*FreeTypeOutlineText, error) {
	engine, err := freetype.NewFontEngineFreetype(false, 32)
	if err != nil {
		return nil, err
	}

	t := &FreeTypeOutlineText{
		engine:     engine,
		cache:      font.NewFontCacheManager(engine, 32),
		path:       path.NewPathStorageStl(),
		resolution: 72,
		hinting:    true,
		dirty:      true,
	}
	t.engine.SetResolution(t.resolution)
	t.engine.SetHinting(t.hinting)
	t.engine.SetFlipY(t.flip)

	return t, nil
}

// Close releases the underlying FreeType resources.
func (t *FreeTypeOutlineText) Close() error {
	if t == nil || t.engine == nil {
		return nil
	}

	err := t.engine.Close()
	t.engine = nil
	t.cache = nil
	t.path = nil

	return err
}

// LoadFont loads a font face from fontPath in outline mode.
func (t *FreeTypeOutlineText) LoadFont(fontPath string) error {
	if t == nil || t.engine == nil {
		return nil
	}

	t.engine.SetResolution(t.resolution)
	t.engine.SetHinting(t.hinting)
	t.engine.SetFlipY(t.flip)
	if err := t.engine.LoadFont(fontPath, 0, freetype.GlyphRenderingOutline, nil); err != nil {
		return err
	}

	t.applySize()
	t.dirty = true

	return nil
}

// SetResolution sets the outline engine DPI.
func (t *FreeTypeOutlineText) SetResolution(dpi uint) {
	if t == nil || t.engine == nil || dpi == 0 {
		return
	}

	t.resolution = dpi
	t.engine.SetResolution(dpi)
	t.dirty = true
}

// SetHinting enables or disables FreeType hinting.
func (t *FreeTypeOutlineText) SetHinting(hinting bool) {
	if t == nil || t.engine == nil {
		return
	}

	t.hinting = hinting
	t.engine.SetHinting(hinting)
	t.dirty = true
}

// SetFlip controls whether the outline Y coordinates are flipped.
func (t *FreeTypeOutlineText) SetFlip(flip bool) {
	if t == nil || t.engine == nil {
		return
	}

	t.flip = flip
	t.engine.SetFlipY(flip)
	t.dirty = true
}

// SetSize sets the nominal text height and optional width.
func (t *FreeTypeOutlineText) SetSize(height, width float64) {
	if t == nil || t.engine == nil {
		return
	}

	t.height = height
	t.width = width
	t.applySize()
	t.dirty = true
}

// SetStartPoint sets the baseline origin for the current text path.
func (t *FreeTypeOutlineText) SetStartPoint(x, y float64) {
	if t == nil {
		return
	}

	t.startX = x
	t.startY = y
	t.dirty = true
}

// SetText sets the current text string.
func (t *FreeTypeOutlineText) SetText(text string) {
	if t == nil {
		return
	}

	t.text = text
	t.dirty = true
}

// MeasureText returns the advance width of str with the current font settings.
func (t *FreeTypeOutlineText) MeasureText(str string) float64 {
	if t == nil || t.cache == nil || str == "" {
		return 0
	}

	return t.measureText(str)
}

// TextWidth returns the width of the currently configured text.
func (t *FreeTypeOutlineText) TextWidth() float64 {
	if t == nil {
		return 0
	}

	return t.MeasureText(t.text)
}

// GetAscender returns the current font ascender.
func (t *FreeTypeOutlineText) GetAscender() float64 {
	if t == nil || t.engine == nil {
		return 0
	}

	return t.engine.GetAscender()
}

// GetDescender returns the current font descender.
func (t *FreeTypeOutlineText) GetDescender() float64 {
	if t == nil || t.engine == nil {
		return 0
	}

	return t.engine.GetDescender()
}

// Rewind resets vertex iteration for the current text path.
func (t *FreeTypeOutlineText) Rewind(pathID uint) {
	if t == nil {
		return
	}

	t.rebuildPath()
	if t.path != nil {
		t.path.Rewind(pathID)
	}
}

// Vertex returns the next outline-path vertex.
func (t *FreeTypeOutlineText) Vertex() (x, y float64, cmd PathCommand) {
	if t == nil || t.path == nil {
		return 0, 0, PathCmdStop
	}

	x, y, raw := t.path.NextVertex()
	return x, y, PathCommand(raw)
}

func (t *FreeTypeOutlineText) applySize() {
	if t == nil || t.engine == nil {
		return
	}

	if t.height > 0 {
		t.engine.SetHeight(t.height)
	}
	t.engine.SetWidth(t.width)
}

func (t *FreeTypeOutlineText) measureText(str string) float64 {
	x := 0.0
	y := 0.0
	first := true
	var prevGlyphIndex uint

	for _, r := range str {
		glyph := t.cache.Glyph(uint(r))
		if glyph == nil {
			continue
		}
		if !first {
			t.cache.AddKerning(&x, &y, prevGlyphIndex, glyph.GlyphIndex)
		}
		x += glyph.AdvanceX
		y += glyph.AdvanceY
		first = false
		prevGlyphIndex = glyph.GlyphIndex
	}

	return x
}

func (t *FreeTypeOutlineText) rebuildPath() {
	if t == nil || !t.dirty || t.path == nil || t.cache == nil {
		return
	}

	t.path.RemoveAll()
	if t.text == "" {
		t.dirty = false
		return
	}

	currentX := t.startX
	currentY := t.startY
	first := true
	var prevGlyphIndex uint

	for _, r := range t.text {
		glyph := t.cache.Glyph(uint(r))
		if glyph == nil {
			continue
		}

		if !first {
			t.cache.AddKerning(&currentX, &currentY, prevGlyphIndex, glyph.GlyphIndex)
		}

		t.cache.InitEmbeddedAdaptors(glyph, currentX, currentY)
		if glyph.DataType == font.GlyphDataOutline {
			if src := t.cache.PathAdaptor(); src != nil {
				t.path.ConcatPath(src, 0)
			}
		}

		currentX += glyph.AdvanceX
		currentY += glyph.AdvanceY
		prevGlyphIndex = glyph.GlyphIndex
		first = false
	}

	t.dirty = false
}
