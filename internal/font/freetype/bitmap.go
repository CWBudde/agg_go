package freetype

// GlyphBitmap is the raw rendered bitmap of the current glyph, as returned by
// FontEngineFreetype.CurrentBitmap. Data is empty when no bitmap is rendered.
type GlyphBitmap struct {
	Data      []byte
	Width     int
	Height    int
	Pitch     int
	Left      int
	Top       int
	PixelMode uint8
}
