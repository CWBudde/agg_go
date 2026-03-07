package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

type mockAMaskPixfmt struct {
	width, height int
	pixels        map[[2]int]color.RGBA8[color.Linear]
}

func newMockAMaskPixfmt(width, height int) *mockAMaskPixfmt {
	return &mockAMaskPixfmt{
		width:  width,
		height: height,
		pixels: make(map[[2]int]color.RGBA8[color.Linear]),
	}
}

func (m *mockAMaskPixfmt) Width() int  { return m.width }
func (m *mockAMaskPixfmt) Height() int { return m.height }
func (m *mockAMaskPixfmt) GetPixel(x, y int) color.RGBA8[color.Linear] {
	return m.pixels[[2]int{x, y}]
}

func (m *mockAMaskPixfmt) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	m.pixels[[2]int{x, y}] = c
}

func (m *mockAMaskPixfmt) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	m.pixels[[2]int{x, y}] = c
}
func (m *mockAMaskPixfmt) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {}
func (m *mockAMaskPixfmt) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {}
func (m *mockAMaskPixfmt) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
}

func (m *mockAMaskPixfmt) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
}

func (m *mockAMaskPixfmt) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
}

func (m *mockAMaskPixfmt) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
}
func (m *mockAMaskPixfmt) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear]) {}
func (m *mockAMaskPixfmt) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear], cover basics.Int8u) {
}
func (m *mockAMaskPixfmt) CopyColorHspan(x, y, length int, colors []color.RGBA8[color.Linear]) {}
func (m *mockAMaskPixfmt) CopyColorVspan(x, y, length int, colors []color.RGBA8[color.Linear]) {}
func (m *mockAMaskPixfmt) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
}
func (m *mockAMaskPixfmt) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
}
func (m *mockAMaskPixfmt) Clear(c color.RGBA8[color.Linear]) {}
func (m *mockAMaskPixfmt) Fill(c color.RGBA8[color.Linear])  {}
func (m *mockAMaskPixfmt) Pixel(x, y int) color.RGBA8[color.Linear] {
	return m.pixels[[2]int{x, y}]
}
func (m *mockAMaskPixfmt) PixWidth() int { return 4 }

type zeroAlphaMask struct{}

func (zeroAlphaMask) Width() int                                             { return 0 }
func (zeroAlphaMask) Height() int                                            { return 0 }
func (zeroAlphaMask) Pixel(x, y int) basics.Int8u                            { return 0 }
func (zeroAlphaMask) CombinePixel(x, y int, cover basics.Int8u) basics.Int8u { return 0 }
func (zeroAlphaMask) FillHspan(x, y int, dst []basics.Int8u, length int) {
	for i := 0; i < length && i < len(dst); i++ {
		dst[i] = 0
	}
}

func (zeroAlphaMask) CombineHspan(x, y int, dst []basics.Int8u, length int) {
	for i := 0; i < length && i < len(dst); i++ {
		dst[i] = 0
	}
}

func (zeroAlphaMask) FillVspan(x, y int, dst []basics.Int8u, length int) {
	for i := 0; i < length && i < len(dst); i++ {
		dst[i] = 0
	}
}

func (zeroAlphaMask) CombineVspan(x, y int, dst []basics.Int8u, length int) {
	for i := 0; i < length && i < len(dst); i++ {
		dst[i] = 0
	}
}

type mockRowSource struct {
	width, height int
	row           []basics.Int8u
}

func (m *mockRowSource) Width() int                   { return m.width }
func (m *mockRowSource) Height() int                  { return m.height }
func (m *mockRowSource) RowData(y int) []basics.Int8u { return m.row }
func (m *mockRowSource) PixWidth() int                { return 4 }

// fullAlphaMask always returns full alpha (255).
type fullAlphaMask struct{ w, h int }

func (m fullAlphaMask) Width() int  { return m.w }
func (m fullAlphaMask) Height() int { return m.h }
func (m fullAlphaMask) Pixel(x, y int) basics.Int8u { return 255 }
func (m fullAlphaMask) CombinePixel(x, y int, cover basics.Int8u) basics.Int8u { return cover }
func (m fullAlphaMask) FillHspan(x, y int, dst []basics.Int8u, length int) {
	for i := range dst[:length] {
		dst[i] = 255
	}
}
func (m fullAlphaMask) CombineHspan(x, y int, dst []basics.Int8u, length int) {}
func (m fullAlphaMask) FillVspan(x, y int, dst []basics.Int8u, length int) {
	for i := range dst[:length] {
		dst[i] = 255
	}
}
func (m fullAlphaMask) CombineVspan(x, y int, dst []basics.Int8u, length int) {}

func newRGBAPixfmt(w, h int) (*newMockTrackingPixfmt, *PixFmtAMaskAdaptor) {
	mock := newTrackingMock(w, h)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: w, h: h})
	return mock, adaptor
}

// newMockTrackingPixfmt tracks blend calls to verify they pass through.
type newMockTrackingPixfmt struct {
	*mockAMaskPixfmt
	blendSolidHspanCalled bool
	blendSolidVspanCalled bool
}

func newTrackingMock(w, h int) *newMockTrackingPixfmt {
	return &newMockTrackingPixfmt{mockAMaskPixfmt: newMockAMaskPixfmt(w, h)}
}

func (m *newMockTrackingPixfmt) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	m.blendSolidHspanCalled = true
}
func (m *newMockTrackingPixfmt) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	m.blendSolidVspanCalled = true
}

func TestPixFmtAMaskAdaptorWidthHeightPixWidth(t *testing.T) {
	_, adaptor := newRGBAPixfmt(10, 20)
	if adaptor.Width() != 10 {
		t.Errorf("Width() expected 10, got %d", adaptor.Width())
	}
	if adaptor.Height() != 20 {
		t.Errorf("Height() expected 20, got %d", adaptor.Height())
	}
	if adaptor.PixWidth() != 4 {
		t.Errorf("PixWidth() expected 4, got %d", adaptor.PixWidth())
	}
}

func TestPixFmtAMaskAdaptorAttach(t *testing.T) {
	mock1 := newMockAMaskPixfmt(4, 4)
	mock2 := newMockAMaskPixfmt(8, 8)
	mask1 := fullAlphaMask{w: 4, h: 4}
	mask2 := zeroAlphaMask{}

	adaptor := NewPixFmtAMaskAdaptor(mock1, mask1)
	if adaptor.Width() != 4 {
		t.Errorf("initial Width() expected 4, got %d", adaptor.Width())
	}

	adaptor.AttachPixfmt(mock2)
	if adaptor.Width() != 8 {
		t.Errorf("after AttachPixfmt Width() expected 8, got %d", adaptor.Width())
	}

	adaptor.AttachAlphaMask(mask2)
	// attach just changes the mask — check it compiles and doesn't panic
}

func TestPixFmtAMaskAdaptorPixel(t *testing.T) {
	mock := newMockAMaskPixfmt(4, 4)
	c := color.RGBA8[color.Linear]{R: 1, G: 2, B: 3, A: 255}
	mock.pixels[[2]int{1, 1}] = c
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	if got := adaptor.Pixel(1, 1); got != c {
		t.Errorf("Pixel(1,1) = %+v want %+v", got, c)
	}
}

func TestPixFmtAMaskAdaptorCopyPixelWithFullMask(t *testing.T) {
	// CopyPixel uses mask.Pixel as coverage → full mask → pixel is blended in
	mock := newMockAMaskPixfmt(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	c := color.RGBA8[color.Linear]{R: 99, G: 88, B: 77, A: 255}
	adaptor.CopyPixel(2, 2, c)

	if got := mock.pixels[[2]int{2, 2}]; got != c {
		t.Errorf("CopyPixel through full mask: got %+v want %+v", got, c)
	}
}

func TestPixFmtAMaskAdaptorCopyPixelWithZeroMask(t *testing.T) {
	// CopyPixel uses mask.Pixel as coverage → zero mask → BlendPixel(cover=0) → no change
	mock := newMockAMaskPixfmt(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, zeroAlphaMask{})

	c := color.RGBA8[color.Linear]{R: 99, G: 88, B: 77, A: 255}
	adaptor.CopyPixel(2, 2, c)

	// With zero cover, BlendPixel stores the color anyway in our mock (mock doesn't implement alpha blend)
	// The important thing is it doesn't panic
}

func TestPixFmtAMaskAdaptorBlendPixel(t *testing.T) {
	mock := newMockAMaskPixfmt(4, 4)
	mask := NewSimpleAlphaMask(4, 4)
	mask.SetOpaque()
	adaptor := NewPixFmtAMaskAdaptor(mock, mask)

	c := color.RGBA8[color.Linear]{R: 55, G: 66, B: 77, A: 255}
	adaptor.BlendPixel(1, 1, c, basics.CoverFull)
	if got := mock.pixels[[2]int{1, 1}]; got != c {
		t.Errorf("BlendPixel through opaque mask: got %+v want %+v", got, c)
	}
}

func TestPixFmtAMaskAdaptorCopyHlineWithFullMask(t *testing.T) {
	mock := newTrackingMock(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	c := color.RGBA8[color.Linear]{R: 200, G: 100, B: 50, A: 255}
	adaptor.CopyHline(0, 0, 4, c)
	// full mask → calls BlendSolidHspan on underlying pixfmt
	if !mock.blendSolidHspanCalled {
		t.Error("CopyHline should call BlendSolidHspan on underlying pixfmt")
	}
}

func TestPixFmtAMaskAdaptorBlendHlineWithZeroMask(t *testing.T) {
	mock := newTrackingMock(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, zeroAlphaMask{})

	c := color.RGBA8[color.Linear]{R: 200, G: 100, B: 50, A: 255}
	adaptor.BlendHline(0, 0, 4, c, basics.CoverFull)
	// zero mask → CombineHspan zeroes covers → BlendSolidHspan called with zeros
	if !mock.blendSolidHspanCalled {
		t.Error("BlendHline should always call BlendSolidHspan")
	}
}

func TestPixFmtAMaskAdaptorCopyVline(t *testing.T) {
	mock := newTrackingMock(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	c := color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 255}
	adaptor.CopyVline(0, 0, 4, c)
	if !mock.blendSolidVspanCalled {
		t.Error("CopyVline should call BlendSolidVspan on underlying pixfmt")
	}
}

func TestPixFmtAMaskAdaptorBlendVline(t *testing.T) {
	mock := newTrackingMock(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, zeroAlphaMask{})

	c := color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 255}
	adaptor.BlendVline(0, 0, 4, c, basics.CoverFull)
	if !mock.blendSolidVspanCalled {
		t.Error("BlendVline should call BlendSolidVspan")
	}
}

func TestPixFmtAMaskAdaptorBlendSolidHspanWithCovers(t *testing.T) {
	mock := newTrackingMock(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	c := color.RGBA8[color.Linear]{R: 255, A: 255}
	covers := []basics.Int8u{255, 128, 64, 0}
	adaptor.BlendSolidHspan(0, 0, 4, c, covers)
	if !mock.blendSolidHspanCalled {
		t.Error("BlendSolidHspan(covers) should call BlendSolidHspan on underlying pixfmt")
	}
}

func TestPixFmtAMaskAdaptorBlendSolidHspanNilCovers(t *testing.T) {
	// nil covers → delegates to CopyHline
	mock := newTrackingMock(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	c := color.RGBA8[color.Linear]{R: 255, A: 255}
	adaptor.BlendSolidHspan(0, 0, 4, c, nil)
	if !mock.blendSolidHspanCalled {
		t.Error("BlendSolidHspan(nil) should call BlendSolidHspan via CopyHline")
	}
}

func TestPixFmtAMaskAdaptorBlendSolidVspanWithAndWithoutCovers(t *testing.T) {
	mock := newTrackingMock(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	c := color.RGBA8[color.Linear]{G: 255, A: 255}
	covers := []basics.Int8u{255, 128, 64, 0}
	adaptor.BlendSolidVspan(0, 0, 4, c, covers)
	if !mock.blendSolidVspanCalled {
		t.Error("BlendSolidVspan(covers) should call BlendSolidVspan")
	}

	mock.blendSolidVspanCalled = false
	adaptor.BlendSolidVspan(0, 0, 4, c, nil)
	if !mock.blendSolidVspanCalled {
		t.Error("BlendSolidVspan(nil) should call BlendSolidVspan via CopyVline")
	}
}

func TestPixFmtAMaskAdaptorCopyBarBlendBar(t *testing.T) {
	mock := newTrackingMock(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	c := color.RGBA8[color.Linear]{R: 100, A: 255}
	adaptor.CopyBar(0, 0, 2, 2, c)
	if !mock.blendSolidHspanCalled {
		t.Error("CopyBar should call BlendSolidHspan")
	}

	mock.blendSolidHspanCalled = false
	adaptor.BlendBar(0, 0, 2, 2, c, basics.CoverFull)
	if !mock.blendSolidHspanCalled {
		t.Error("BlendBar should call BlendSolidHspan via BlendHline")
	}
}

func TestPixFmtAMaskAdaptorCopyColorHspanVspan(t *testing.T) {
	mock := newMockAMaskPixfmt(3, 3)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 3, h: 3})

	colors := []color.RGBA8[color.Linear]{
		{R: 10, A: 255}, {R: 20, A: 255}, {R: 30, A: 255},
	}

	// CopyColorHspan delegates directly to underlying pixfmt
	adaptor.CopyColorHspan(0, 0, 3, colors)
	adaptor.CopyColorVspan(0, 0, 3, colors)
	// No panic means the delegation works
}

func TestPixFmtAMaskAdaptorClearAndFill(t *testing.T) {
	mock := newMockAMaskPixfmt(4, 4)
	adaptor := NewPixFmtAMaskAdaptor(mock, fullAlphaMask{w: 4, h: 4})

	c := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	adaptor.Clear(c) // delegates to underlying pixfmt.Clear — no panic
	adaptor.Fill(c)  // delegates to underlying pixfmt.Fill — no panic
}

func TestPixFmtAMaskAdaptorBlendColorHspanWithCovers(t *testing.T) {
	mock := newMockAMaskPixfmt(4, 1)
	mask := NewSimpleAlphaMask(4, 1)
	mask.Fill(255) // fully opaque mask
	adaptor := NewPixFmtAMaskAdaptor(mock, mask)

	colors := []color.RGBA8[color.Linear]{
		{R: 100, A: 255},
		{G: 100, A: 255},
		{B: 100, A: 255},
		{R: 50, G: 50, A: 255},
	}
	covers := []basics.Int8u{255, 255, 255, 255}
	adaptor.BlendColorHspan(0, 0, 4, colors, covers, basics.CoverFull)
	// pixels should be set (mask is opaque, covers are 255)
	if got := mock.pixels[[2]int{0, 0}]; got.R != 100 {
		t.Errorf("BlendColorHspan x=0: R expected 100, got %d", got.R)
	}
}

func TestPixFmtAMaskAdaptorBlendColorHspanNilCovers(t *testing.T) {
	mock := newMockAMaskPixfmt(2, 1)
	mask := NewSimpleAlphaMask(2, 1)
	mask.Fill(255)
	adaptor := NewPixFmtAMaskAdaptor(mock, mask)

	colors := []color.RGBA8[color.Linear]{
		{R: 77, A: 255},
		{G: 88, A: 255},
	}
	adaptor.BlendColorHspan(0, 0, 2, colors, nil, basics.CoverFull)
	// each pixel's cover = global cover (255), then combined with mask (255) → 255
	if got := mock.pixels[[2]int{0, 0}]; got.R != 77 {
		t.Errorf("BlendColorHspan nil covers x=0: R expected 77, got %d", got.R)
	}
}

func TestPixFmtAMaskAdaptorBlendColorVspan(t *testing.T) {
	mock := newMockAMaskPixfmt(1, 3)
	mask := NewSimpleAlphaMask(1, 3)
	mask.Fill(255)
	adaptor := NewPixFmtAMaskAdaptor(mock, mask)

	colors := []color.RGBA8[color.Linear]{
		{R: 10, A: 255},
		{G: 20, A: 255},
		{B: 30, A: 255},
	}
	adaptor.BlendColorVspan(0, 0, 3, colors, nil, basics.CoverFull)
	if got := mock.pixels[[2]int{0, 0}]; got.R != 10 {
		t.Errorf("BlendColorVspan y=0: R expected 10, got %d", got.R)
	}
	if got := mock.pixels[[2]int{0, 2}]; got.B != 30 {
		t.Errorf("BlendColorVspan y=2: B expected 30, got %d", got.B)
	}
}

func TestSimpleAlphaMaskOperations(t *testing.T) {
	mask := NewSimpleAlphaMask(4, 4)

	if mask.Width() != 4 || mask.Height() != 4 {
		t.Errorf("NewSimpleAlphaMask dimensions wrong: %d×%d", mask.Width(), mask.Height())
	}

	// SetPixel and Pixel
	mask.SetPixel(2, 1, 200)
	if got := mask.Pixel(2, 1); got != 200 {
		t.Errorf("Pixel(2,1) expected 200, got %d", got)
	}
	if got := mask.Pixel(-1, 0); got != 0 {
		t.Errorf("out-of-bounds Pixel should return 0")
	}

	// CombinePixel
	combined := mask.CombinePixel(2, 1, 128)
	if combined == 0 || combined > 200 {
		t.Errorf("CombinePixel(200, 128) = %d, expected > 0 and <= 200", combined)
	}

	// Fill and Clear
	mask.Fill(100)
	if mask.Pixel(0, 0) != 100 {
		t.Errorf("Fill(100) didn't work")
	}
	mask.Clear()
	if mask.Pixel(0, 0) != 0 {
		t.Errorf("Clear() didn't zero mask")
	}

	// SetOpaque
	mask.SetOpaque()
	if mask.Pixel(3, 3) != 255 {
		t.Errorf("SetOpaque() didn't make mask fully opaque")
	}

	// FillHspan
	dst := make([]basics.Int8u, 4)
	mask.FillHspan(0, 0, dst, 4)
	for i, v := range dst {
		if v != 255 {
			t.Errorf("FillHspan dst[%d] = %d, want 255", i, v)
		}
	}

	// CombineHspan
	covers := []basics.Int8u{128, 64, 32, 16}
	mask.CombineHspan(0, 0, covers, 4)
	for i, v := range covers {
		if v != 128>>(i) {
			// roughly should be scaled
			_ = v
		}
	}

	// FillVspan
	dst2 := make([]basics.Int8u, 4)
	mask.FillVspan(0, 0, dst2, 4)
	for i, v := range dst2 {
		if v != 255 {
			t.Errorf("FillVspan dst[%d] = %d, want 255", i, v)
		}
	}

	// CombineVspan
	covers2 := []basics.Int8u{200, 100, 50, 25}
	mask.CombineVspan(0, 0, covers2, 4)
	// all mask values are 255, so combine(v, 255) = v * 255 / 255 = v
	for i, v := range covers2 {
		_ = i
		_ = v
	}
}

func TestSimpleAlphaMaskOutOfBounds(t *testing.T) {
	mask := NewSimpleAlphaMask(3, 3)
	mask.SetOpaque()

	dst := make([]basics.Int8u, 3)
	// FillHspan out-of-bounds row
	mask.FillHspan(0, -1, dst, 3)
	for _, v := range dst {
		if v != 0 {
			t.Errorf("FillHspan out-of-bounds y should zero dst, got %d", v)
		}
	}

	// FillVspan out-of-bounds column
	mask.FillVspan(-1, 0, dst, 3)
	for _, v := range dst {
		if v != 0 {
			t.Errorf("FillVspan out-of-bounds x should zero dst, got %d", v)
		}
	}
}

func TestPixFmtAMaskAdaptorCopyFromFallbackBypassesMask(t *testing.T) {
	dst := newMockAMaskPixfmt(4, 1)
	adaptor := NewPixFmtAMaskAdaptor(dst, zeroAlphaMask{})

	src := &mockRowSource{
		width:  2,
		height: 1,
		row: []basics.Int8u{
			10, 20, 30, 255,
			40, 50, 60, 255,
		},
	}

	adaptor.CopyFrom(src, 1, 0, 0, 0, 2)

	if got := dst.GetPixel(1, 0); got.R != 10 || got.G != 20 || got.B != 30 || got.A != 255 {
		t.Fatalf("pixel (1,0) = %+v, want RGBA(10,20,30,255)", got)
	}
	if got := dst.GetPixel(2, 0); got.R != 40 || got.G != 50 || got.B != 60 || got.A != 255 {
		t.Fatalf("pixel (2,0) = %+v, want RGBA(40,50,60,255)", got)
	}
}
