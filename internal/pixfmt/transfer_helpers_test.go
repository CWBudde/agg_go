package pixfmt

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
)

type mockRowDataSource struct {
	rows  [][]basics.Int8u
	width int
}

func (m mockRowDataSource) RowData(y int) []basics.Int8u {
	if y < 0 || y >= len(m.rows) {
		return nil
	}
	return m.rows[y]
}

func (m mockRowDataSource) Width() int  { return m.width }
func (m mockRowDataSource) Height() int { return len(m.rows) }

type mockPixWidthSource struct {
	mockRowDataSource
	pixWidth int
}

func (m mockPixWidthSource) PixWidth() int { return m.pixWidth }

func TestDetectBytesPerPixel(t *testing.T) {
	withPixWidth := mockPixWidthSource{
		mockRowDataSource: mockRowDataSource{
			rows:  [][]basics.Int8u{{1, 2, 3, 4, 5, 6}},
			width: 2,
		},
		pixWidth: 3,
	}
	if got := detectBytesPerPixel(withPixWidth, 0); got != 3 {
		t.Fatalf("detectBytesPerPixel() with PixWidth = %d, want 3", got)
	}

	fromRowLength := mockRowDataSource{
		rows:  [][]basics.Int8u{{1, 2, 3, 4, 5, 6, 7, 8}},
		width: 2,
	}
	if got := detectBytesPerPixel(fromRowLength, 0); got != 4 {
		t.Fatalf("detectBytesPerPixel() from row length = %d, want 4", got)
	}

	fallback := mockRowDataSource{
		rows:  [][]basics.Int8u{{1, 2, 3}},
		width: 0,
	}
	if got := detectBytesPerPixel(fallback, 0); got != 4 {
		t.Fatalf("detectBytesPerPixel() fallback = %d, want 4", got)
	}
	if got := detectBytesPerPixel(fromRowLength, 5); got != 4 {
		t.Fatalf("detectBytesPerPixel() out-of-range sample = %d, want 4", got)
	}
}

func TestDecodeRGBA8FromRowData(t *testing.T) {
	tests := []struct {
		name   string
		row    []basics.Int8u
		bpp    int
		pixelX int
		want   [4]basics.Int8u
		wantOK bool
	}{
		{"gray", []basics.Int8u{77}, 1, 0, [4]basics.Int8u{77, 77, 77, 255}, true},
		{"gray+alpha", []basics.Int8u{80, 90}, 2, 0, [4]basics.Int8u{80, 80, 80, 90}, true},
		{"rgb", []basics.Int8u{10, 20, 30}, 3, 0, [4]basics.Int8u{10, 20, 30, 255}, true},
		{"rgba", []basics.Int8u{1, 2, 3, 4}, 4, 0, [4]basics.Int8u{1, 2, 3, 4}, true},
		{"negative x", []basics.Int8u{1, 2, 3, 4}, 4, -1, [4]basics.Int8u{}, false},
		{"out of bounds", []basics.Int8u{1, 2, 3}, 4, 0, [4]basics.Int8u{}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := decodeRGBA8FromRowData(tc.row, tc.bpp, tc.pixelX)
			if ok != tc.wantOK {
				t.Fatalf("decodeRGBA8FromRowData() ok=%v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if got.R != tc.want[0] || got.G != tc.want[1] || got.B != tc.want[2] || got.A != tc.want[3] {
				t.Fatalf("decodeRGBA8FromRowData() = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					got.R, got.G, got.B, got.A,
					tc.want[0], tc.want[1], tc.want[2], tc.want[3],
				)
			}
		})
	}
}
