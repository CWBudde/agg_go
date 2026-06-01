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
	engine               *freetype.FontEngineFreetype
	cache                *font.FontCacheManager
	path                 *path.PathStorageStl
	text                 string
	startX               float64
	startY               float64
	height               float64
	width                float64
	resolution           uint
	hinting              bool
	forceAutohint        bool
	snapOutlineX         bool
	ttInterpreterVersion uint
	flip                 bool
	dirty                bool
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
	t.engine.SetForceAutohint(t.forceAutohint)
	t.engine.SetSnapOutlineX(t.snapOutlineX)
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
	t.engine.SetForceAutohint(t.forceAutohint)
	t.engine.SetSnapOutlineX(t.snapOutlineX)
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

// SetForceAutohint forces FreeType to use its auto-hinter instead of the
// font's native TrueType bytecode hints. Has no effect when hinting is
// disabled. Produces stem positions that more aggressively grid-fit to
// integer pixel boundaries, closer to Java Graphics2D's text AA output.
func (t *FreeTypeOutlineText) SetForceAutohint(force bool) {
	if t == nil || t.engine == nil {
		return
	}

	t.forceAutohint = force
	t.engine.SetForceAutohint(force)
	t.dirty = true
}

// SetSnapOutlineX enables integer-snapping of the leftmost outline X
// coordinate after FT_Load_Glyph. When enabled, the whole outline is
// translated so the leftmost vertex lands on an integer pixel (Java-like
// aggressive stem grid-fit). Advance metrics are preserved.
func (t *FreeTypeOutlineText) SetSnapOutlineX(snap bool) {
	if t == nil || t.engine == nil {
		return
	}

	t.snapOutlineX = snap
	t.engine.SetSnapOutlineX(snap)
	t.dirty = true
}

// SetTrueTypeInterpreterVersion selects the TrueType bytecode interpreter
// version used by FreeType. Common values: 35 (classic Apple, aggressive
// X+Y grid-fit, closest to Java greyscale AA), 38 (intermediate), 40
// (FreeType default, sub-pixel biased, no X grid-fit). Pass 0 to leave
// FreeType's default untouched.
func (t *FreeTypeOutlineText) SetTrueTypeInterpreterVersion(v uint) error {
	if t == nil || t.engine == nil {
		return nil
	}

	t.ttInterpreterVersion = v
	if err := t.engine.SetTrueTypeInterpreterVersion(v); err != nil {
		return err
	}

	t.dirty = true
	return nil
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

// FirstGlyphMinX returns the minimum X coordinate of the first rune's outline
// when rendered at origin (0, 0) — effectively the left side bearing of the
// first glyph with the current font settings (including hinting / autohint).
// Returns ok=false if no text is configured, the glyph has no outline, or the
// outline is empty. The returned value is in pixel units and is typically a
// small non-negative fractional number.
func (t *FreeTypeOutlineText) FirstGlyphMinX() (float64, bool) {
	if t == nil || t.cache == nil || t.text == "" {
		return 0, false
	}

	var firstRune rune
	haveRune := false
	for _, r := range t.text {
		firstRune = r
		haveRune = true
		break
	}
	if !haveRune {
		return 0, false
	}

	glyph := t.cache.Glyph(uint(firstRune))
	if glyph == nil || glyph.DataType != font.GlyphDataOutline {
		return 0, false
	}

	t.cache.InitEmbeddedAdaptors(glyph, 0, 0)
	src := t.cache.PathAdaptor()
	if src == nil {
		return 0, false
	}

	src.Rewind(0)
	minX := 0.0
	haveVertex := false
	for {
		x, _, raw := src.NextVertex()
		cmd := PathCommand(raw)
		if cmd == PathCmdStop {
			break
		}
		if !IsPathVertex(cmd) {
			continue
		}
		if !haveVertex || x < minX {
			minX = x
			haveVertex = true
		}
	}
	if !haveVertex {
		return 0, false
	}

	t.dirty = true

	return minX, true
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
