package span

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func approxF128(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= 1.0/255.0
}

func TestSpanGouraudRGBA128Creation(t *testing.T) {
	sg := NewSpanGouraudRGBA128()
	if sg == nil {
		t.Fatal("NewSpanGouraudRGBA128 returned nil")
	}
}

func TestSpanGouraudRGBA128WithTriangle(t *testing.T) {
	c1 := color.NewRGBA32[color.Linear](1, 0, 0, 1) // Red
	c2 := color.NewRGBA32[color.Linear](0, 1, 0, 1) // Green
	c3 := color.NewRGBA32[color.Linear](0, 0, 1, 1) // Blue

	sg := NewSpanGouraudRGBA128WithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	if sg == nil {
		t.Fatal("NewSpanGouraudRGBA128WithTriangle returned nil")
	}
	coord := sg.Coord()
	if coord[0].Color.R != 1 || coord[0].Color.G != 0 {
		t.Errorf("first vertex color not set correctly: %+v", coord[0].Color)
	}
	if coord[2].Color.B != 1 || coord[2].Color.R != 0 {
		t.Errorf("third vertex color not set correctly: %+v", coord[2].Color)
	}
}

// TestSpanGouraudRGBA128SolidColor verifies that a triangle with all three
// vertices the same color generates that exact color across a span.
func TestSpanGouraudRGBA128SolidColor(t *testing.T) {
	c := color.NewRGBA32[color.Linear](0.25, 0.5, 0.75, 1.0)
	sg := NewSpanGouraudRGBA128WithTriangle(c, c, c, 0, 0, 40, 0, 20, 40, 0)
	sg.Prepare()

	out := make([]color.RGBA32[color.Linear], 20)
	sg.Generate(out, 5, 20, 10)
	for i := range 10 {
		g := out[i]
		if !approxF128(g.R, 0.25) || !approxF128(g.G, 0.5) || !approxF128(g.B, 0.75) || !approxF128(g.A, 1.0) {
			t.Fatalf("solid span pixel %d = %+v, want {0.25,0.5,0.75,1}", i, g)
		}
	}
}

// TestSpanGouraudRGBA128Interpolates verifies color varies across a horizontal
// gradient triangle (left red, right green at the same Y band).
func TestSpanGouraudRGBA128Interpolates(t *testing.T) {
	red := color.NewRGBA32[color.Linear](1, 0, 0, 1)
	green := color.NewRGBA32[color.Linear](0, 1, 0, 1)
	// Wide flat-ish triangle so a mid scanline spans red->green horizontally.
	sg := NewSpanGouraudRGBA128WithTriangle(red, green, green, 0, 0, 100, 0, 50, 60, 0)
	sg.Prepare()

	out := make([]color.RGBA32[color.Linear], 100)
	sg.Generate(out, 0, 1, 100)

	// Somewhere across the span the green channel must rise above the red as we
	// move from the red vertex toward the green ones.
	left := out[2]
	right := out[90]
	if !(left.R > right.R) {
		t.Fatalf("expected red to fall left->right: left.R=%v right.R=%v", left.R, right.R)
	}
	if !(right.G > left.G) {
		t.Fatalf("expected green to rise left->right: left.G=%v right.G=%v", left.G, right.G)
	}
}
