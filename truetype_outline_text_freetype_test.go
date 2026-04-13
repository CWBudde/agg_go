//go:build freetype

package agg

import "testing"

func TestFreeTypeOutlineTextMeasureAndVertex(t *testing.T) {
	txt, err := NewFreeTypeOutlineText()
	if err != nil {
		t.Fatalf("NewFreeTypeOutlineText() error = %v", err)
	}
	defer func() { _ = txt.Close() }()

	fontPaths := []string{
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/System/Library/Fonts/Arial.ttf",
		"C:\\Windows\\Fonts\\arial.ttf",
	}

	loaded := false
	for _, fontPath := range fontPaths {
		if err := txt.LoadFont(fontPath); err == nil {
			loaded = true
			break
		}
	}
	if !loaded {
		t.Skip("no test TrueType font available")
	}

	txt.SetFlip(true)
	txt.SetSize(24, 0)

	if got := txt.MeasureText("AV"); got <= 0 {
		t.Fatalf("MeasureText(AV) = %v, want > 0", got)
	}

	txt.SetText("AV")
	txt.SetStartPoint(10, 20)
	txt.Rewind(0)

	vertices := 0
	sawMove := false
	sawCurve := false
	for {
		_, _, cmd := txt.Vertex()
		if cmd == PathCmdStop {
			break
		}
		vertices++
		if cmd == PathCmdMoveTo {
			sawMove = true
		}
		if IsPathCurve3(cmd) || IsPathCurve4(cmd) {
			sawCurve = true
		}
	}

	if vertices == 0 {
		t.Fatal("expected outline vertices")
	}
	if !sawMove {
		t.Fatal("expected move-to command")
	}
	if !sawCurve {
		t.Fatal("expected curve command in outline path")
	}
}
