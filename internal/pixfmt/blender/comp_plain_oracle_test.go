package blender

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

// TestCompositeBlenderPlainMatchesCppOracle replays the grid of
// testdata/comp_plain_oracle.cpp through CompositeBlenderPlain.BlendPix and
// compares it with stock AGG premultiply -> comp_op -> demultiply.
//
// The bridge is not bit-exact with AGG: Go premultiplies the source and the
// destination in float64, while AGG rounds both with rgba8::multiply before
// the operator and rounds the premultiplied result before demultiplying it
// (PLAN.md 8.2, comp-op integer source premultiply). The comparison is
// therefore made in premultiplied space (premulDelta), where those roundings
// stay within premulTolerance; anything larger is a semantic difference, such
// as leaving premultiplied data in the straight buffer.
func TestCompositeBlenderPlainMatchesCppOracle(t *testing.T) {
	colors := []basics.Int8u{0, 1, 77, 128, 200, 255}
	dstAlphas := []basics.Int8u{0, 1, 64, 128, 200, 254, 255}
	srcAlphas := []basics.Int8u{0, 1, 77, 128, 179, 255}
	covers := []basics.Int8u{1, 64, 128, 255}
	nc := len(colors)

	for _, tc := range []struct {
		name string
		op   CompOp
	}{
		{"src_over", CompOpSrcOver},
		{"xor", CompOpXor},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", "comp_plain_"+tc.name+".bin"))
			if err != nil {
				t.Fatalf("read oracle: %v", err)
			}
			bl := NewCompositeBlenderPlain[color.Linear, order.RGBA](tc.op)
			k, worst, differing := 0, 0, 0
			for _, da := range dstAlphas {
				for _, sa := range srcAlphas {
					for _, cover := range covers {
						for i := range nc {
							for j := range nc {
								px := []basics.Int8u{colors[i], colors[nc-1-i], colors[(i+2)%nc], da}
								bl.BlendPix(px, colors[j], colors[(j+3)%nc], colors[nc-1-j], sa, cover)
								exp := want[k*4 : k*4+4]
								k++
								d := premulDelta(px, exp)
								if d > 0 {
									differing++
								}
								worst = max(worst, d)
								if d > premulTolerance {
									t.Errorf("dst=(%d,%d,%d,%d) src=(%d,%d,%d,%d) cover=%d: got %v, want %v",
										colors[i], colors[nc-1-i], colors[(i+2)%nc], da,
										colors[j], colors[(j+3)%nc], colors[nc-1-j], sa, cover, px, exp)
								}
							}
						}
					}
				}
			}
			if k*4 != len(want) {
				t.Fatalf("replayed %d entries, oracle has %d", k, len(want)/4)
			}
			t.Logf("%d/%d entries differ, max delta %d", differing, k, worst)
		})
	}
}

// premulTolerance is the largest premultiplied-space difference the rounding
// in the two bridges can produce: the source premultiply and the result can
// each be one step off, and the two add up for src_over.
const premulTolerance = 2

// premulDelta compares two straight pixels in premultiplied space: the alpha
// difference, and the difference of each colour channel times its own alpha.
// Straight colour is not comparable directly: at alpha 1 AGG's integer
// premultiply keeps no colour at all, while the float bridge keeps all of it.
func premulDelta(got, want []basics.Int8u) int {
	d := absInt(int(got[3]) - int(want[3]))
	for c := range 3 {
		d = max(d, absInt(premul(got[c], got[3])-premul(want[c], want[3])))
	}
	return d
}

func premul(c, a basics.Int8u) int {
	return (int(c)*int(a) + 127) / 255
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
