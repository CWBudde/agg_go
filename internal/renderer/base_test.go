package renderer

import (
	"testing"

	"agg_go/internal/basics"
)

// MockPixelFormat is a typed mock pixel format for tests.
type MockPixelFormat[C comparable] struct {
	width, height int
	pixels        map[[2]int]C
}

func NewMockPixelFormat[C comparable](w, h int) *MockPixelFormat[C] {
	return &MockPixelFormat[C]{
		width:  w,
		height: h,
		pixels: make(map[[2]int]C),
	}
}

func (m *MockPixelFormat[C]) Width() int    { return m.width }
func (m *MockPixelFormat[C]) Height() int   { return m.height }
func (m *MockPixelFormat[C]) PixWidth() int { return 4 }

func (m *MockPixelFormat[C]) CopyPixel(x, y int, c C) { m.pixels[[2]int{x, y}] = c }
func (m *MockPixelFormat[C]) BlendPixel(x, y int, c C, cover basics.Int8u) {
	m.pixels[[2]int{x, y}] = c
}

func (m *MockPixelFormat[C]) Pixel(x, y int) C {
	if p, ok := m.pixels[[2]int{x, y}]; ok {
		return p
	}
	var zero C
	return zero
}

func (m *MockPixelFormat[C]) CopyHline(x, y, length int, c C) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x + i, y}] = c
	}
}

func (m *MockPixelFormat[C]) BlendHline(x, y, length int, c C, cover basics.Int8u) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x + i, y}] = c
	}
}

func (m *MockPixelFormat[C]) CopyVline(x, y, length int, c C) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x, y + i}] = c
	}
}

func (m *MockPixelFormat[C]) BlendVline(x, y, length int, c C, cover basics.Int8u) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x, y + i}] = c
	}
}

func (m *MockPixelFormat[C]) BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x + i, y}] = c
	}
}

func (m *MockPixelFormat[C]) BlendSolidVspan(x, y, length int, c C, covers []basics.Int8u) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x, y + i}] = c
	}
}

func (m *MockPixelFormat[C]) CopyColorHspan(x, y, length int, colors []C) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x + i, y}] = colors[i]
	}
}

func (m *MockPixelFormat[C]) BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x + i, y}] = colors[i]
	}
}

func (m *MockPixelFormat[C]) CopyColorVspan(x, y, length int, colors []C) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x, y + i}] = colors[i]
	}
}

func (m *MockPixelFormat[C]) BlendColorVspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x, y + i}] = colors[i]
	}
}

func (m *MockPixelFormat[C]) CopyBar(x1, y1, x2, y2 int, c C) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			m.pixels[[2]int{x, y}] = c
		}
	}
}

func (m *MockPixelFormat[C]) BlendBar(x1, y1, x2, y2 int, c C, cover basics.Int8u) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			m.pixels[[2]int{x, y}] = c
		}
	}
}

func (m *MockPixelFormat[C]) Clear(c C) {
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			m.pixels[[2]int{x, y}] = c
		}
	}
}

func (m *MockPixelFormat[C]) Fill(c C) {
	m.Clear(c)
}

func TestRendererBaseTypedBasics(t *testing.T) {
	// Use string as a simple color type for validation
	pf := NewMockPixelFormat[string](50, 50)
	r := NewRendererBaseWithPixfmt[*MockPixelFormat[string], string](pf)

	if r.Width() != 50 || r.Height() != 50 {
		t.Fatalf("expected 50x50, got %dx%d", r.Width(), r.Height())
	}

	// Clipping and pixel ops
	if !r.ClipBox(10, 10, 20, 20) {
		t.Fatal("ClipBox should succeed")
	}

	r.CopyPixel(12, 12, "red")
	if got := r.Pixel(12, 12); got != "red" {
		t.Fatalf("expected red, got %v", got)
	}

	// Outside clip should return zero value of C (empty string)
	if got := r.Pixel(5, 5); got != "" {
		t.Fatalf("expected zero value, got %q", got)
	}

	// H/V lines within clip
	r.CopyHline(10, 15, 20, "green")
	if got := r.Pixel(17, 15); got != "green" {
		t.Fatalf("expected green, got %v", got)
	}

	r.CopyVline(18, 10, 20, "blue")
	if got := r.Pixel(18, 17); got != "blue" {
		t.Fatalf("expected blue, got %v", got)
	}
}

func TestRendererBaseCopyFromOverlappingVerticalRegion(t *testing.T) {
	pf := NewMockPixelFormat[string](4, 4)
	r := NewRendererBaseWithPixfmt[*MockPixelFormat[string], string](pf)

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			pf.CopyPixel(x, y, string(rune('A'+y)))
		}
	}

	srcRect := &basics.RectI{X1: 0, Y1: 0, X2: 3, Y2: 2}
	r.CopyFrom(pf, srcRect, 0, 1)

	for x := 0; x < 4; x++ {
		if got := pf.Pixel(x, 1); got != "A" {
			t.Fatalf("row 1 pixel %d = %q, want %q", x, got, "A")
		}
		if got := pf.Pixel(x, 2); got != "B" {
			t.Fatalf("row 2 pixel %d = %q, want %q", x, got, "B")
		}
		if got := pf.Pixel(x, 3); got != "C" {
			t.Fatalf("row 3 pixel %d = %q, want %q", x, got, "C")
		}
	}
}

func TestRendererBaseBlendFromOverlappingVerticalRegion(t *testing.T) {
	pf := NewMockPixelFormat[string](3, 4)
	r := NewRendererBaseWithPixfmt[*MockPixelFormat[string], string](pf)

	for y := 0; y < 4; y++ {
		for x := 0; x < 3; x++ {
			pf.CopyPixel(x, y, string(rune('K'+y)))
		}
	}

	srcRect := &basics.RectI{X1: 0, Y1: 0, X2: 2, Y2: 2}
	r.BlendFrom(pf, srcRect, 0, 1, basics.CoverFull)

	for x := 0; x < 3; x++ {
		if got := pf.Pixel(x, 1); got != "K" {
			t.Fatalf("row 1 pixel %d = %q, want %q", x, got, "K")
		}
		if got := pf.Pixel(x, 2); got != "L" {
			t.Fatalf("row 2 pixel %d = %q, want %q", x, got, "L")
		}
		if got := pf.Pixel(x, 3); got != "M" {
			t.Fatalf("row 3 pixel %d = %q, want %q", x, got, "M")
		}
	}
}

func TestRendererBaseCopyFromUsesRelativeDestinationOffset(t *testing.T) {
	src := NewMockPixelFormat[string](5, 5)
	dst := NewMockPixelFormat[string](8, 8)
	r := NewRendererBaseWithPixfmt[*MockPixelFormat[string], string](dst)

	src.CopyPixel(2, 3, "P")
	src.CopyPixel(3, 3, "Q")
	src.CopyPixel(2, 4, "R")
	src.CopyPixel(3, 4, "S")

	srcRect := &basics.RectI{X1: 2, Y1: 3, X2: 3, Y2: 4}
	r.CopyFrom(src, srcRect, 1, 2)

	if got := dst.Pixel(3, 5); got != "P" {
		t.Fatalf("dst(3,5) = %q, want %q", got, "P")
	}
	if got := dst.Pixel(4, 5); got != "Q" {
		t.Fatalf("dst(4,5) = %q, want %q", got, "Q")
	}
	if got := dst.Pixel(3, 6); got != "R" {
		t.Fatalf("dst(3,6) = %q, want %q", got, "R")
	}
	if got := dst.Pixel(4, 6); got != "S" {
		t.Fatalf("dst(4,6) = %q, want %q", got, "S")
	}

	if got := dst.Pixel(1, 2); got != "" {
		t.Fatalf("unexpected absolute-position copy at (1,2): got %q", got)
	}
}

func TestRendererBaseBlendFromUsesRelativeDestinationOffset(t *testing.T) {
	src := NewMockPixelFormat[string](5, 5)
	dst := NewMockPixelFormat[string](8, 8)
	r := NewRendererBaseWithPixfmt[*MockPixelFormat[string], string](dst)

	src.CopyPixel(1, 1, "A")
	src.CopyPixel(2, 1, "B")
	src.CopyPixel(1, 2, "C")
	src.CopyPixel(2, 2, "D")

	srcRect := &basics.RectI{X1: 1, Y1: 1, X2: 2, Y2: 2}
	r.BlendFrom(src, srcRect, 2, 3, basics.CoverFull)

	if got := dst.Pixel(3, 4); got != "A" {
		t.Fatalf("dst(3,4) = %q, want %q", got, "A")
	}
	if got := dst.Pixel(4, 4); got != "B" {
		t.Fatalf("dst(4,4) = %q, want %q", got, "B")
	}
	if got := dst.Pixel(3, 5); got != "C" {
		t.Fatalf("dst(3,5) = %q, want %q", got, "C")
	}
	if got := dst.Pixel(4, 5); got != "D" {
		t.Fatalf("dst(4,5) = %q, want %q", got, "D")
	}

	if got := dst.Pixel(2, 3); got != "" {
		t.Fatalf("unexpected absolute-position blend at (2,3): got %q", got)
	}
}
