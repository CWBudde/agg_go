//go:build typed_renderer

package renderer

import (
    "testing"

    "agg_go/internal/basics"
)

// MockPixelFormatT is a typed mock pixel format for tests.
type MockPixelFormatT[C comparable] struct {
    width, height int
    pixels        map[[2]int]C
}

func NewMockPixelFormatT[C comparable](w, h int) *MockPixelFormatT[C] {
    return &MockPixelFormatT[C]{
        width:  w,
        height: h,
        pixels: make(map[[2]int]C),
    }
}

func (m *MockPixelFormatT[C]) Width() int    { return m.width }
func (m *MockPixelFormatT[C]) Height() int   { return m.height }
func (m *MockPixelFormatT[C]) PixWidth() int { return 4 }

func (m *MockPixelFormatT[C]) CopyPixel(x, y int, c C) { m.pixels[[2]int{x, y}] = c }
func (m *MockPixelFormatT[C]) BlendPixel(x, y int, c C, cover basics.Int8u) {
    m.pixels[[2]int{x, y}] = c
}
func (m *MockPixelFormatT[C]) Pixel(x, y int) C {
    if p, ok := m.pixels[[2]int{x, y}]; ok {
        return p
    }
    var zero C
    return zero
}

func (m *MockPixelFormatT[C]) CopyHline(x, y, length int, c C) {
    for i := 0; i < length; i++ {
        m.pixels[[2]int{x + i, y}] = c
    }
}
func (m *MockPixelFormatT[C]) BlendHline(x, y, length int, c C, cover basics.Int8u) {
    for i := 0; i < length; i++ {
        m.pixels[[2]int{x + i, y}] = c
    }
}
func (m *MockPixelFormatT[C]) CopyVline(x, y, length int, c C) {
    for i := 0; i < length; i++ {
        m.pixels[[2]int{x, y + i}] = c
    }
}
func (m *MockPixelFormatT[C]) BlendVline(x, y, length int, c C, cover basics.Int8u) {
    for i := 0; i < length; i++ {
        m.pixels[[2]int{x, y + i}] = c
    }
}

func (m *MockPixelFormatT[C]) BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u) {
    for i := 0; i < length; i++ {
        m.pixels[[2]int{x + i, y}] = c
    }
}
func (m *MockPixelFormatT[C]) BlendSolidVspan(x, y, length int, c C, covers []basics.Int8u) {
    for i := 0; i < length; i++ {
        m.pixels[[2]int{x, y + i}] = c
    }
}

func (m *MockPixelFormatT[C]) CopyColorHspan(x, y, length int, colors []C) {
    for i := 0; i < length && i < len(colors); i++ {
        m.pixels[[2]int{x + i, y}] = colors[i]
    }
}
func (m *MockPixelFormatT[C]) BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
    for i := 0; i < length && i < len(colors); i++ {
        m.pixels[[2]int{x + i, y}] = colors[i]
    }
}
func (m *MockPixelFormatT[C]) CopyColorVspan(x, y, length int, colors []C) {
    for i := 0; i < length && i < len(colors); i++ {
        m.pixels[[2]int{x, y + i}] = colors[i]
    }
}
func (m *MockPixelFormatT[C]) BlendColorVspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
    for i := 0; i < length && i < len(colors); i++ {
        m.pixels[[2]int{x, y + i}] = colors[i]
    }
}

func TestRendererBaseTTypedBasics(t *testing.T) {
    // Use string as a simple color type for validation
    pf := NewMockPixelFormatT[string](50, 50)
    r := NewRendererBaseTWithPixfmt[*MockPixelFormatT[string], string](pf)

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
