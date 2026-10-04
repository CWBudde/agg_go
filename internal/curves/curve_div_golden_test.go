package curves

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
)

// TestCurveDivGoldenCpp compares Curve3Div/Curve4Div vertex streams bit-exactly
// against vertices dumped from C++ AGG 2.6 curve3_div/curve4_div.
//
// testdata/curve_div_golden.txt was generated with
//
//	clang++ -O0 -ffp-contract=off -I agg-src/include golden.cpp agg-src/src/agg_curves.cpp
//
// over the 17 test curves listed in agg-src/examples/bezier_div.cpp plus
// collinear/degenerate curves that exercise the curve4_div switch cases 0, 1
// and 2 (p2 or p3 exactly on the chord). Angles and cusp limits are given in
// degrees and converted with deg*pi/180, as the C++ generator does.
//
// Record format:
//
//	C4 x1 y1 x2 y2 x3 y3 x4 y4 scale angleDeg cuspDeg n
//	C3 x1 y1 x2 y2 x3 y3 scale angleDeg n
//
// each followed by n lines "cmd x y".
func TestCurveDivGoldenCpp(t *testing.T) {
	f, err := os.Open("testdata/curve_div_golden.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	line := 0
	next := func() []float64 {
		if !sc.Scan() {
			t.Fatalf("unexpected EOF after line %d", line)
		}
		line++
		fields := strings.Fields(sc.Text())
		out := make([]float64, 0, len(fields))
		for i, s := range fields {
			if i == 0 && (s == "C3" || s == "C4") {
				continue
			}
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				t.Fatalf("line %d: %v", line, err)
			}
			out = append(out, v)
		}
		return out
	}

	type vertexSource interface {
		Rewind(uint)
		Vertex() (float64, float64, basics.PathCommand)
	}

	records := 0
	for sc.Scan() {
		line++
		header := sc.Text()
		fields := strings.Fields(header)
		if len(fields) == 0 {
			continue
		}
		vals := make([]float64, 0, len(fields)-1)
		for _, s := range fields[1:] {
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				t.Fatalf("line %d: %v", line, err)
			}
			vals = append(vals, v)
		}

		var src vertexSource
		var n int
		switch fields[0] {
		case "C4":
			c := NewCurve4Div()
			c.SetApproximationScale(vals[8])
			c.SetAngleTolerance(vals[9] * math.Pi / 180.0)
			c.SetCuspLimit(vals[10] * math.Pi / 180.0)
			c.Init(vals[0], vals[1], vals[2], vals[3], vals[4], vals[5], vals[6], vals[7])
			src, n = c, int(vals[11])
		case "C3":
			c := NewCurve3Div()
			c.SetApproximationScale(vals[6])
			c.SetAngleTolerance(vals[7] * math.Pi / 180.0)
			c.Init(vals[0], vals[1], vals[2], vals[3], vals[4], vals[5])
			src, n = c, int(vals[8])
		default:
			t.Fatalf("line %d: unknown record %q", line, fields[0])
		}

		src.Rewind(0)
		mismatch := false
		for i := 0; i < n; i++ {
			want := next()
			x, y, cmd := src.Vertex()
			if mismatch {
				continue
			}
			if uint32(cmd) != uint32(want[0]) || x != want[1] || y != want[2] {
				t.Errorf("%s: vertex %d = (%d, %.17g, %.17g), want (%d, %.17g, %.17g)",
					header, i, cmd, x, y, uint32(want[0]), want[1], want[2])
				mismatch = true
			}
		}
		if !mismatch {
			if _, _, cmd := src.Vertex(); !basics.IsStop(cmd) {
				t.Errorf("%s: more than %d vertices", header, n)
			}
		}
		records++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if records == 0 {
		t.Fatal("no golden records")
	}
}
