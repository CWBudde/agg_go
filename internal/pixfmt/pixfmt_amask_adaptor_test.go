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
