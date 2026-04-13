package agg

import "testing"

func TestPublicGSVTextMeasureAndVertex(t *testing.T) {
	txt := NewGSVText()
	txt.SetFlip(true)
	txt.SetSize(18, 0)
	txt.SetText("AG")
	txt.SetStartPoint(10, 20)

	if got := txt.MeasureText("AG"); got <= 0 {
		t.Fatalf("MeasureText(\"AG\") = %v, want > 0", got)
	}

	txt.Rewind(0)
	seenMove := false
	seenLine := false
	for i := 0; i < 4096; i++ {
		_, _, cmd := txt.Vertex()
		switch cmd {
		case GSVPathCmdStop:
			if !seenMove || !seenLine {
				t.Fatalf("expected move and line vertices before stop, got move=%v line=%v", seenMove, seenLine)
			}
			return
		case GSVPathCmdMoveTo:
			seenMove = true
		case GSVPathCmdLineTo:
			seenLine = true
		}
	}

	t.Fatal("vertex stream did not terminate")
}
