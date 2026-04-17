//go:build freetype

package agg

import "testing"

// TestFreeTypeOutlineTextCFF verifies that CFF/OpenType fonts (which use cubic
// Bezier outlines) produce non-empty glyph paths. Regression test for a bug
// where the outline decomposer compared the FreeType curve tag against 3
// instead of FT_CURVE_TAG_CUBIC (0x02), causing every cubic contour on a CFF
// font to be silently rejected — glyphs loaded but emitted zero vertices.
func TestFreeTypeOutlineTextCFF(t *testing.T) {
	txt, err := NewFreeTypeOutlineText()
	if err != nil {
		t.Fatalf("NewFreeTypeOutlineText() error = %v", err)
	}
	defer func() { _ = txt.Close() }()

	fontPaths := []string{
		"/usr/share/fonts/opentype/urw-base35/NimbusRoman-Regular.otf",
		"/usr/share/fonts/opentype/urw-base35/URWBookman-Light.otf",
	}

	loaded := false
	for _, fontPath := range fontPaths {
		if err := txt.LoadFont(fontPath); err == nil {
			loaded = true
			break
		}
	}
	if !loaded {
		t.Skip("no CFF/OpenType test font available")
	}

	txt.SetFlip(true)
	txt.SetSize(24, 0)

	if got := txt.MeasureText("AV"); got <= 0 {
		t.Fatalf("MeasureText(AV) on CFF font = %v, want > 0", got)
	}

	txt.SetText("AV")
	txt.SetStartPoint(0, 0)
	txt.Rewind(0)

	vertices := 0
	sawMove := false
	sawCubic := false
	for {
		_, _, cmd := txt.Vertex()
		if cmd == PathCmdStop {
			break
		}
		vertices++
		if cmd == PathCmdMoveTo {
			sawMove = true
		}
		if IsPathCurve4(cmd) {
			sawCubic = true
		}
	}

	if vertices == 0 {
		t.Fatal("expected outline vertices from CFF font (got 0 — cubic tag check likely wrong)")
	}
	if !sawMove {
		t.Fatal("expected move-to command from CFF font")
	}
	if !sawCubic {
		t.Fatal("expected at least one cubic curve (Curve4) command — CFF fonts use cubic beziers")
	}
}
