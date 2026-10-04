package rasterizer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/agg_go/internal/scanline"
)

// Golden values in this file were dumped from AGG 2.6 C++
// (rasterizer_scanline_aa<> and rasterizer_scanline_aa_nogamma<> with the
// default rasterizer_sl_clip_int, swept with scanline_u8; clang++ -O0
// -ffp-contract=off). Both C++ variants produce identical output.

// statusRasterizer is the subset shared by both scanline AA variants that
// the C++ parity scenarios below exercise.
type statusRasterizer interface {
	MoveToD(x, y float64)
	LineToD(x, y float64)
	EdgeD(x1, y1, x2, y2 float64)
	RewindScanlines() bool
	SweepScanline(sl scanline.Scanline) bool
	MinX() int
	MinY() int
	MaxX() int
	MaxY() int
}

func newParityRasterizers() map[string]statusRasterizer {
	return map[string]statusRasterizer{
		"gamma": NewRasterizerScanlineAA[int, IntConv, *RasterizerSlClip[int, IntConv]](
			IntConv{}, NewRasterizerSlClip[int, IntConv](IntConv{})),
		"nogamma": NewRasterizerScanlineAANoGamma[int, IntConv, *RasterizerSlClip[int, IntConv]](
			IntConv{}, NewRasterizerSlClip[int, IntConv](IntConv{})),
	}
}

func sweepSpansCpp(t *testing.T, ras statusRasterizer) []string {
	t.Helper()
	if !ras.RewindScanlines() {
		return nil
	}
	sl := scanline.NewScanlineU8()
	sl.Reset(ras.MinX(), ras.MaxX())
	var out []string
	for ras.SweepScanline(sl) {
		for _, span := range sl.Begin() {
			var b strings.Builder
			fmt.Fprintf(&b, "y=%d x=%d len=%d", sl.Y(), span.X, span.Len)
			for i := 0; i < int(span.Len); i++ {
				fmt.Fprintf(&b, " %d", span.Covers[i])
			}
			out = append(out, b.String())
		}
	}
	return out
}

func checkBoundsCpp(t *testing.T, ras statusRasterizer, want [4]int) {
	t.Helper()
	got := [4]int{ras.MinX(), ras.MinY(), ras.MaxX(), ras.MaxY()}
	if got != want {
		t.Errorf("bounds = %v, want %v (C++)", got, want)
	}
}

// edge_d() leaves the rasterizer in status_move_to, so the pending
// line_to contour before it is NOT auto-closed by rewind_scanlines().
var cppEdgeAfterContourSpans = []string{
	"y=10 x=10 len=31 4 21 39 57 73 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 77 226",
	"y=11 x=14 len=27 1 16 34 52 70 88 107 125 143 161 179 197 215 233 250 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=12 x=29 len=13 14 237 255 255 255 255 255 255 255 255 255 255 255",
	"y=13 x=29 len=13 49 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=14 x=29 len=14 121 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=15 x=29 len=14 193 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=16 x=28 len=15 13 250 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=17 x=28 len=16 80 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=18 x=28 len=16 152 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=19 x=28 len=17 223 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=20 x=27 len=18 39 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=21 x=27 len=19 110 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=22 x=27 len=19 182 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=23 x=26 len=20 8 245 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=24 x=26 len=21 69 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=25 x=26 len=21 141 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=26 x=26 len=22 213 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=27 x=25 len=23 28 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255",
	"y=28 x=25 len=24 86 230 230 230 230 230 230 230 230 230 230 230 230 230 230 230 230 230 230 230 230 230 255 255",
	"y=29 x=48 len=1 175",
	"y=30 x=48 len=1 39",
}

func TestRasterizerScanlineAA_EdgeStatusMatchesCpp(t *testing.T) {
	for name, ras := range newParityRasterizers() {
		t.Run(name, func(t *testing.T) {
			ras.MoveToD(10.3, 10.7)
			ras.LineToD(30.2, 12.1)
			ras.LineToD(25.5, 28.9)
			ras.EdgeD(40.1, 10.2, 48.7, 30.4)
			checkBoundsCpp(t, ras, [4]int{10, 10, 48, 30})
			got := sweepSpansCpp(t, ras)
			if strings.Join(got, "\n") != strings.Join(cppEdgeAfterContourSpans, "\n") {
				t.Fatalf("spans mismatch\nwant:\n%s\n\ngot:\n%s",
					strings.Join(cppEdgeAfterContourSpans, "\n"), strings.Join(got, "\n"))
			}
		})
	}
}

// close_polygon() only closes from status_line_to; a trailing bare move_to
// must not emit a closing edge (which would extend the cell bounds).
var cppStrayMoveToSpans = []string{
	"y=10 x=10 len=11 62 128 128 128 128 128 128 128 128 128 96",
	"y=11 x=10 len=11 6 168 255 255 255 255 255 255 255 255 192",
	"y=12 x=11 len=10 3 151 255 255 255 255 255 255 255 192",
	"y=13 x=12 len=9 1 133 255 255 255 255 255 255 192",
	"y=14 x=14 len=7 115 255 255 255 255 255 192",
	"y=15 x=15 len=6 98 253 255 255 255 192",
	"y=16 x=16 len=5 82 249 255 255 192",
	"y=17 x=17 len=4 67 244 255 192",
	"y=18 x=18 len=3 55 237 192",
	"y=19 x=19 len=2 43 165",
	"y=20 x=20 len=1 9",
}

func TestRasterizerScanlineAA_StrayMoveToMatchesCpp(t *testing.T) {
	for name, ras := range newParityRasterizers() {
		t.Run(name, func(t *testing.T) {
			ras.MoveToD(10.25, 10.5)
			ras.LineToD(20.75, 10.5)
			ras.LineToD(20.75, 20.25)
			ras.MoveToD(150.5, 170.5) // stray trailing move_to
			checkBoundsCpp(t, ras, [4]int{10, 10, 20, 20})
			got := sweepSpansCpp(t, ras)
			checkBoundsCpp(t, ras, [4]int{10, 10, 20, 20})
			if strings.Join(got, "\n") != strings.Join(cppStrayMoveToSpans, "\n") {
				t.Fatalf("spans mismatch\nwant:\n%s\n\ngot:\n%s",
					strings.Join(cppStrayMoveToSpans, "\n"), strings.Join(got, "\n"))
			}
		})
	}
}

// TestRasterizerScanlineAA_DefaultCellBlockLimitMatchesCpp checks the default
// cell_block_limit (1024 blocks of 4096 cells in AGG). The zigzag below
// produces 1,200,001 cells, more than a 256-block limit (1,048,576) can hold.
func TestRasterizerScanlineAA_DefaultCellBlockLimitMatchesCpp(t *testing.T) {
	const (
		wantCells    = 1200001
		wantCoverSum = 306330439
		wantFNV      = uint64(11108252005346207514)
	)
	gammaRas := NewRasterizerScanlineAA[int, IntConv, *RasterizerSlClip[int, IntConv]](
		IntConv{}, NewRasterizerSlClip[int, IntConv](IntConv{}))
	noGammaRas := NewRasterizerScanlineAANoGamma[int, IntConv, *RasterizerSlClip[int, IntConv]](
		IntConv{}, NewRasterizerSlClip[int, IntConv](IntConv{}))
	cases := map[string]struct {
		ras   statusRasterizer
		cells func() uint32
	}{
		"gamma":   {gammaRas, gammaRas.outline.TotalCells},
		"nogamma": {noGammaRas, noGammaRas.outline.TotalCells},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ras := tc.ras
			const n = 1000
			for i := 0; i <= n; i++ {
				x := 1.25 + float64(i)*2.0
				y := 0.5
				if i&1 == 1 {
					y = 1200.5
				}
				if i == 0 {
					ras.MoveToD(x, y)
				} else {
					ras.LineToD(x, y)
				}
			}
			if !ras.RewindScanlines() {
				t.Fatal("expected scanlines")
			}
			if got := tc.cells(); got != wantCells {
				t.Fatalf("total cells = %d, want %d (C++)", got, wantCells)
			}
			sl := scanline.NewScanlineU8()
			sl.Reset(ras.MinX(), ras.MaxX())
			var sum uint64
			h := uint64(1469598103934665603)
			for ras.SweepScanline(sl) {
				for _, span := range sl.Begin() {
					for i := 0; i < int(span.Len); i++ {
						c := uint64(span.Covers[i])
						sum += c
						h = (h ^ (c + 256*uint64((int(span.X)+i)&0xffff))) * 1099511628211
					}
				}
			}
			if sum != wantCoverSum || h != wantFNV {
				t.Fatalf("coverage sum/hash = %d/%d, want %d/%d (C++)", sum, h, wantCoverSum, wantFNV)
			}
		})
	}
}
