package renderer

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
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

// --- BlendFromColor / BlendFromLUT tests ---

// testColor is a simple RGBA color type for blend tests.
type testColor struct {
	R, G, B, A uint8
}

// mockGraySource implements GraySource for tests.
type mockGraySource struct {
	w, h int
	rows [][]basics.Int8u // rows[y] is the full row data
}

func (s *mockGraySource) Width() int                   { return s.w }
func (s *mockGraySource) Height() int                  { return s.h }
func (s *mockGraySource) RowData(y int) []basics.Int8u { return s.rows[y] }

// blendPixFmt is a mock pixel format that also supports BlendFromColor and BlendFromLUT.
type blendPixFmt struct {
	width, height int
	pixels        map[[2]int]testColor
}

func newBlendPixFmt(w, h int) *blendPixFmt {
	return &blendPixFmt{width: w, height: h, pixels: make(map[[2]int]testColor)}
}

func (m *blendPixFmt) Width() int    { return m.width }
func (m *blendPixFmt) Height() int   { return m.height }
func (m *blendPixFmt) PixWidth() int { return 4 }

func (m *blendPixFmt) CopyPixel(x, y int, c testColor) { m.pixels[[2]int{x, y}] = c }
func (m *blendPixFmt) BlendPixel(x, y int, c testColor, cover basics.Int8u) {
	m.pixels[[2]int{x, y}] = c
}

func (m *blendPixFmt) Pixel(x, y int) testColor {
	if p, ok := m.pixels[[2]int{x, y}]; ok {
		return p
	}
	return testColor{}
}

func (m *blendPixFmt) CopyHline(x, y, length int, c testColor) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x + i, y}] = c
	}
}

func (m *blendPixFmt) BlendHline(x, y, length int, c testColor, cover basics.Int8u) {
	m.CopyHline(x, y, length, c)
}

func (m *blendPixFmt) CopyVline(x, y, length int, c testColor) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x, y + i}] = c
	}
}

func (m *blendPixFmt) BlendVline(x, y, length int, c testColor, cover basics.Int8u) {
	m.CopyVline(x, y, length, c)
}

func (m *blendPixFmt) BlendSolidHspan(x, y, length int, c testColor, covers []basics.Int8u) {
	m.CopyHline(x, y, length, c)
}

func (m *blendPixFmt) BlendSolidVspan(x, y, length int, c testColor, covers []basics.Int8u) {
	m.CopyVline(x, y, length, c)
}

func (m *blendPixFmt) CopyColorHspan(x, y, length int, colors []testColor) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x + i, y}] = colors[i]
	}
}

func (m *blendPixFmt) BlendColorHspan(x, y, length int, colors []testColor, covers []basics.Int8u, cover basics.Int8u) {
	m.CopyColorHspan(x, y, length, colors)
}

func (m *blendPixFmt) CopyColorVspan(x, y, length int, colors []testColor) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x, y + i}] = colors[i]
	}
}

func (m *blendPixFmt) BlendColorVspan(x, y, length int, colors []testColor, covers []basics.Int8u, cover basics.Int8u) {
	m.CopyColorVspan(x, y, length, colors)
}

func (m *blendPixFmt) CopyBar(x1, y1, x2, y2 int, c testColor) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			m.pixels[[2]int{x, y}] = c
		}
	}
}

func (m *blendPixFmt) BlendBar(x1, y1, x2, y2 int, c testColor, cover basics.Int8u) {
	m.CopyBar(x1, y1, x2, y2, c)
}

func (m *blendPixFmt) Clear(c testColor) {
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			m.pixels[[2]int{x, y}] = c
		}
	}
}
func (m *blendPixFmt) Fill(c testColor) { m.Clear(c) }

// BlendFromColor scales the color's alpha by the gray source value and writes it.
// NOTE: The src parameter uses an anonymous interface (not the named GraySource type)
// to match the blendFromColorCapable constraint used by RendererBase type assertions.
func (m *blendPixFmt) BlendFromColor(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, c testColor, xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u,
) {
	row := src.RowData(ysrc)
	if row == nil {
		return
	}
	for i := 0; i < length; i++ {
		idx := xsrc + i
		if idx < 0 || idx >= len(row) {
			continue
		}
		gray := row[idx]
		// Scale: output = color * (gray/255) * (cover/255). Simplified for test:
		// just store the color with alpha = gray * cover / 255.
		scaledA := uint8((int(gray) * int(cover)) / 255)
		m.pixels[[2]int{xdst + i, ydst}] = testColor{R: c.R, G: c.G, B: c.B, A: scaledA}
	}
}

// BlendFromLUT uses the gray source value as an index into colorLUT.
// NOTE: The src parameter uses an anonymous interface (not the named GraySource type)
// to match the blendFromLUTCapable constraint used by RendererBase type assertions.
func (m *blendPixFmt) BlendFromLUT(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, colorLUT []testColor, xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u,
) {
	row := src.RowData(ysrc)
	if row == nil {
		return
	}
	for i := 0; i < length; i++ {
		idx := xsrc + i
		if idx < 0 || idx >= len(row) {
			continue
		}
		lutIdx := int(row[idx])
		if lutIdx >= len(colorLUT) {
			continue
		}
		c := colorLUT[lutIdx]
		// Scale alpha by cover for test.
		scaledA := uint8((int(c.A) * int(cover)) / 255)
		m.pixels[[2]int{xdst + i, ydst}] = testColor{R: c.R, G: c.G, B: c.B, A: scaledA}
	}
}

func TestRendererBase_BlendFromColor(t *testing.T) {
	// 4x4 destination, white background.
	dst := newBlendPixFmt(4, 4)
	dst.Clear(testColor{R: 255, G: 255, B: 255, A: 255})
	ren := NewRendererBaseWithPixfmt[*blendPixFmt, testColor](dst)

	// 3x2 gray source with known values.
	src := &mockGraySource{
		w: 3, h: 2,
		rows: [][]basics.Int8u{
			{128, 255, 0},  // row 0
			{64, 192, 255}, // row 1
		},
	}

	green := testColor{R: 0, G: 255, B: 0, A: 255}
	ren.BlendFromColor(src, green, nil, 0, 0, basics.CoverFull)

	// Check row 0: gray values 128, 255, 0 with CoverFull (255).
	p00 := dst.Pixel(0, 0)
	if p00.A != 128 {
		t.Fatalf("pixel(0,0) alpha: got %d, want 128", p00.A)
	}
	if p00.G != 255 {
		t.Fatalf("pixel(0,0) green: got %d, want 255", p00.G)
	}

	p10 := dst.Pixel(1, 0)
	if p10.A != 255 {
		t.Fatalf("pixel(1,0) alpha: got %d, want 255", p10.A)
	}

	p20 := dst.Pixel(2, 0)
	if p20.A != 0 {
		t.Fatalf("pixel(2,0) alpha: got %d, want 0", p20.A)
	}

	// Check row 1.
	p01 := dst.Pixel(0, 1)
	if p01.A != 64 {
		t.Fatalf("pixel(0,1) alpha: got %d, want 64", p01.A)
	}
	p11 := dst.Pixel(1, 1)
	if p11.A != 192 {
		t.Fatalf("pixel(1,1) alpha: got %d, want 192", p11.A)
	}

	// Pixel (3,0) should still be white (untouched, src is only 3 wide).
	p30 := dst.Pixel(3, 0)
	if p30.R != 255 || p30.G != 255 || p30.B != 255 {
		t.Fatalf("pixel(3,0) should be white, got %+v", p30)
	}
}

func TestRendererBase_BlendFromColor_WithOffset(t *testing.T) {
	dst := newBlendPixFmt(6, 6)
	dst.Clear(testColor{R: 255, G: 255, B: 255, A: 255})
	ren := NewRendererBaseWithPixfmt[*blendPixFmt, testColor](dst)

	src := &mockGraySource{
		w: 2, h: 2,
		rows: [][]basics.Int8u{
			{100, 200},
			{50, 150},
		},
	}

	red := testColor{R: 255, G: 0, B: 0, A: 255}
	ren.BlendFromColor(src, red, nil, 2, 3, basics.CoverFull)

	// Source (0,0) -> Dest (2,3).
	p := dst.Pixel(2, 3)
	if p.A != 100 || p.R != 255 {
		t.Fatalf("pixel(2,3): got %+v, want R=255 A=100", p)
	}

	// Source (1,0) -> Dest (3,3).
	p = dst.Pixel(3, 3)
	if p.A != 200 || p.R != 255 {
		t.Fatalf("pixel(3,3): got %+v, want R=255 A=200", p)
	}

	// Source (0,1) -> Dest (2,4).
	p = dst.Pixel(2, 4)
	if p.A != 50 {
		t.Fatalf("pixel(2,4) alpha: got %d, want 50", p.A)
	}

	// Untouched pixel.
	p = dst.Pixel(0, 0)
	if p.R != 255 || p.G != 255 || p.B != 255 || p.A != 255 {
		t.Fatalf("pixel(0,0) should be white, got %+v", p)
	}
}

func TestRendererBase_BlendFromLUT(t *testing.T) {
	dst := newBlendPixFmt(4, 4)
	dst.Clear(testColor{R: 0, G: 0, B: 0, A: 0})
	ren := NewRendererBaseWithPixfmt[*blendPixFmt, testColor](dst)

	// Build a 256-entry color LUT: index -> color with that index as alpha.
	lut := make([]testColor, 256)
	for i := range lut {
		lut[i] = testColor{R: uint8(i), G: uint8(255 - i), B: 128, A: 255}
	}

	// 3x2 gray source: values serve as LUT indices.
	src := &mockGraySource{
		w: 3, h: 2,
		rows: [][]basics.Int8u{
			{0, 100, 255},
			{50, 200, 128},
		},
	}

	ren.BlendFromLUT(src, lut, nil, 0, 0, basics.CoverFull)

	// Row 0, pixel 0: LUT[0] = {R:0, G:255, B:128, A:255}, cover=255 => A=255.
	p := dst.Pixel(0, 0)
	if p.R != 0 || p.G != 255 || p.B != 128 || p.A != 255 {
		t.Fatalf("pixel(0,0): got %+v, want {0,255,128,255}", p)
	}

	// Row 0, pixel 1: LUT[100] = {R:100, G:155, B:128, A:255}.
	p = dst.Pixel(1, 0)
	if p.R != 100 || p.G != 155 {
		t.Fatalf("pixel(1,0): got %+v, want R=100 G=155", p)
	}

	// Row 1, pixel 2: LUT[128] = {R:128, G:127, B:128, A:255}.
	p = dst.Pixel(2, 1)
	if p.R != 128 || p.G != 127 {
		t.Fatalf("pixel(2,1): got %+v, want R=128 G=127", p)
	}
}

func TestRendererBase_BlendFromLUT_WithOffset(t *testing.T) {
	dst := newBlendPixFmt(6, 6)
	dst.Clear(testColor{R: 0, G: 0, B: 0, A: 0})
	ren := NewRendererBaseWithPixfmt[*blendPixFmt, testColor](dst)

	lut := make([]testColor, 256)
	for i := range lut {
		lut[i] = testColor{R: uint8(i), G: 0, B: 0, A: 255}
	}

	src := &mockGraySource{
		w: 2, h: 1,
		rows: [][]basics.Int8u{
			{10, 20},
		},
	}

	ren.BlendFromLUT(src, lut, nil, 3, 4, basics.CoverFull)

	// Source (0,0) -> Dest (3,4): LUT[10].
	p := dst.Pixel(3, 4)
	if p.R != 10 {
		t.Fatalf("pixel(3,4) R: got %d, want 10", p.R)
	}

	// Source (1,0) -> Dest (4,4): LUT[20].
	p = dst.Pixel(4, 4)
	if p.R != 20 {
		t.Fatalf("pixel(4,4) R: got %d, want 20", p.R)
	}

	// Untouched pixel.
	p = dst.Pixel(0, 0)
	if p.R != 0 || p.A != 0 {
		t.Fatalf("pixel(0,0) should be zero, got %+v", p)
	}
}
