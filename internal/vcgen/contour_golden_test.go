package vcgen

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
)

// contourGoldenCpp is the vertex output of C++ AGG 2.6 vcgen_contour
// (clang++ -O0 -ffp-contract=off, agg_vcgen_contour.cpp) for closed polygons
// added without orientation flags, i.e. move_to, line_to..., then
// end_poly|close. Header: "P <polygon> <width> <auto_detect>"; vertex lines
// "cmd x y"; end_poly lines carry only the command (C++ leaves x/y untouched).
const contourGoldenCpp = `
P cw_square 10 0
1 5 5
2 5 95
2 95 95
2 95 5
95
P cw_square 10 1
1 -5 -5
2 -5 105
2 105 105
2 105 -5
95
P cw_square -7.5 0
1 -3.7500000000000142 -3.75
2 -3.75 103.75000000000001
2 103.75000000000001 103.75
2 103.75 -3.7500000000000142
95
P cw_square -7.5 1
1 3.75 3.75
2 3.75 96.25
2 96.25 96.25
2 96.25 3.75
95
P ccw_square 10 0
1 -5 -5
2 105 -5
2 105 105
2 -5 105
95
P ccw_square 10 1
1 -5 -5
2 105 -5
2 105 105
2 -5 105
95
P ccw_square -7.5 0
1 3.75 3.75
2 96.25 3.75
2 96.25 96.25
2 3.75 96.25
95
P ccw_square -7.5 1
1 3.75 3.75
2 96.25 3.75
2 96.25 96.25
2 3.75 96.25
95
P cw_irregular 10 0
1 16.630183270899565 28.879134535165619
2 34.941412463011758 173.78651616657197
2 119.92249569717443 144.78597867127729
2 192.60797152755543 163.4407799798571
2 156.89347681092011 46.863278357632353
2 89.987263795760612 65.328433920955945
95
P cw_irregular 10 1
1 4.3698167291004211 11.620865464834381
2 26.558587536988242 187.21348383342803
2 120.32750430282557 155.21402132872271
2 207.39202847244457 177.5592200201429
2 163.60652318907989 34.636721642367633
2 91.012736204239388 54.671566079044041
95
P cw_irregular -7.5 0
1 5.90236254682533 13.778149098625789
2 27.606440652741185 185.53511287507104
2 120.27687822711924 153.91051599654202
2 205.54402135433344 175.79441501510718
2 162.76739239180992 36.165041231775689
2 90.884552153179527 56.003674559283034
95
P cw_irregular -7.5 1
1 15.09763745317467 26.721850901374211
2 33.893559347258822 175.46488712492899
2 119.97312177288076 146.08948400345798
2 194.45597864566656 165.20558498489282
2 157.73260760819008 45.334958768224311
2 90.115447846820473 63.996325440716966
95
P ccw_irregular 10 0
1 91.01273620423936 54.671566079044041
2 163.60652318907989 34.636721642367633
2 207.39202847244457 177.55922002014296
2 120.3275043028256 155.21402132872271
2 26.558587536988242 187.21348383342803
2 4.3698167291004353 11.620865464834395
95
P ccw_irregular 10 1
1 91.01273620423936 54.671566079044041
2 163.60652318907989 34.636721642367633
2 207.39202847244457 177.55922002014296
2 120.3275043028256 155.21402132872271
2 26.558587536988242 187.21348383342803
2 4.3698167291004353 11.620865464834395
95
P ccw_irregular -7.5 0
1 90.115447846820473 63.996325440716966
2 157.73260760819008 45.334958768224283
2 194.45597864566656 165.20558498489282
2 119.97312177288077 146.08948400345798
2 33.893559347258829 175.46488712492896
2 15.097637453174677 26.721850901374239
95
P ccw_irregular -7.5 1
1 90.115447846820473 63.996325440716966
2 157.73260760819008 45.334958768224283
2 194.45597864566656 165.20558498489282
2 119.97312177288077 146.08948400345798
2 33.893559347258829 175.46488712492896
2 15.097637453174677 26.721850901374239
95
`

var contourGoldenPolys = map[string][]float64{
	"cw_square":     {0, 0, 0, 100, 100, 100, 100, 0},
	"ccw_square":    {0, 0, 100, 0, 100, 100, 0, 100},
	"cw_irregular":  {10.5, 20.25, 30.75, 180.5, 120.125, 150.0, 200.0, 170.5, 160.25, 40.75, 90.5, 60.0},
	"ccw_irregular": {90.5, 60.0, 160.25, 40.75, 200.0, 170.5, 120.125, 150.0, 30.75, 180.5, 10.5, 20.25},
}

// contourGoldenTol absorbs last-bit differences from FMA contraction in
// basics.MathStroke on arm64 (the C++ oracle is built with -ffp-contract=off).
// It is far below anything an orientation error would produce (>= 2*width).
const contourGoldenTol = 1e-9

// TestVCGenContourGoldenCpp checks the contour generator, including
// auto_detect_orientation (signed calc_polygon_area > 0 => CCW), against C++.
func TestVCGenContourGoldenCpp(t *testing.T) {
	type rec struct {
		header string
		lines  [][]float64
	}
	var recs []rec
	for _, ln := range strings.Split(strings.TrimSpace(contourGoldenCpp), "\n") {
		f := strings.Fields(ln)
		if f[0] == "P" {
			recs = append(recs, rec{header: ln})
			continue
		}
		vals := make([]float64, len(f))
		for i, s := range f {
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				t.Fatal(err)
			}
			vals[i] = v
		}
		recs[len(recs)-1].lines = append(recs[len(recs)-1].lines, vals)
	}
	if len(recs) != 16 {
		t.Fatalf("parsed %d golden records, want 16", len(recs))
	}

	for _, r := range recs {
		f := strings.Fields(r.header)
		poly := contourGoldenPolys[f[1]]
		w, _ := strconv.ParseFloat(f[2], 64)

		vc := NewVCGenContour()
		vc.Width(w)
		vc.AutoDetectOrientation(f[3] == "1")
		vc.AddVertex(poly[0], poly[1], basics.PathCmdMoveTo)
		for i := 2; i < len(poly); i += 2 {
			vc.AddVertex(poly[i], poly[i+1], basics.PathCmdLineTo)
		}
		vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathFlagClose)
		vc.Rewind(0)

		for i, want := range r.lines {
			x, y, cmd := vc.Vertex()
			if wantCmd := cppPathCmd(uint32(want[0])); cmd != wantCmd {
				t.Fatalf("%s: vertex %d cmd = %#x, want %#x (C++ %#x)", r.header, i, cmd, wantCmd, uint32(want[0]))
			}
			if len(want) == 3 && (math.Abs(x-want[1]) > contourGoldenTol || math.Abs(y-want[2]) > contourGoldenTol) {
				t.Fatalf("%s: vertex %d = (%.17g, %.17g), want (%.17g, %.17g)", r.header, i, x, y, want[1], want[2])
			}
		}
		if _, _, cmd := vc.Vertex(); !basics.IsStop(cmd) {
			t.Fatalf("%s: extra vertex with cmd %d", r.header, cmd)
		}
	}
}

// cppPathCmd maps a C++ AGG command word to the Go encoding: C++ uses
// path_cmd_end_poly = 0x0F while basics.PathCmdEndPoly is 8; flag bits match.
func cppPathCmd(c uint32) basics.PathCommand {
	if c&0x0F == 0x0F {
		return basics.PathCmdEndPoly | basics.PathCommand(c&0xF0)
	}
	return basics.PathCommand(c)
}
