package blender

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

// TestBlenderRGBA8PlainFixedMatchesMatplotlib pins the blender against values
// taken from Matplotlib 3.10.9's own Agg output. Each row is a translucent fill
// composited over opaque white; `plain` is what the agg24-svn blender produces
// and `fixed` is what Matplotlib's fixed_blender_rgba_plain produces. The two
// differ by one LSB on most channels, which is the whole reason this blender
// exists -- if a change makes them agree, the transcription is wrong.
func TestBlenderRGBA8PlainFixedMatchesMatplotlib(t *testing.T) {
	tests := []struct {
		name  string
		src   [3]basics.Int8u
		alpha basics.Int8u
		fixed [3]basics.Int8u
		plain [3]basics.Int8u
	}{
		{"step filled blue", [3]basics.Int8u{107, 158, 230}, 140, [3]basics.Int8u{173, 201, 241}, [3]basics.Int8u{174, 202, 241}},
		{"stacked orange", [3]basics.Int8u{219, 107, 48}, 204, [3]basics.Int8u{226, 136, 89}, [3]basics.Int8u{226, 137, 89}},
		{"stackplot blue", [3]basics.Int8u{51, 140, 191}, 194, [3]basics.Int8u{100, 167, 206}, [3]basics.Int8u{100, 168, 206}},
	}

	var (
		fixed BlenderRGBA8PlainFixed[color.Linear, order.RGBA]
		plain BlenderRGBA8Plain[color.Linear, order.RGBA]
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := []basics.Int8u{255, 255, 255, 255}
			fixed.BlendPix(got, tt.src[0], tt.src[1], tt.src[2], tt.alpha, 255)
			want := []basics.Int8u{tt.fixed[0], tt.fixed[1], tt.fixed[2], 255}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("fixed blender = %v, want %v", got, want)
				}
			}

			ref := []basics.Int8u{255, 255, 255, 255}
			plain.BlendPix(ref, tt.src[0], tt.src[1], tt.src[2], tt.alpha, 255)
			for i := 0; i < 3; i++ {
				if ref[i] != tt.plain[i] {
					t.Fatalf("agg24-svn blender = %v, want %v (test data is stale)", ref[:3], tt.plain)
				}
			}
		})
	}
}

// TestBlenderRGBA8PlainFixedEdgeCases covers the paths the parity cases never
// reach: zero source alpha (no write at all) and a fully transparent result.
func TestBlenderRGBA8PlainFixedEdgeCases(t *testing.T) {
	var b BlenderRGBA8PlainFixed[color.Linear, order.RGBA]

	px := []basics.Int8u{10, 20, 30, 40}
	b.BlendPix(px, 200, 200, 200, 0, 255)
	if px[0] != 10 || px[1] != 20 || px[2] != 30 || px[3] != 40 {
		t.Fatalf("alpha 0 must not write, got %v", px)
	}

	px = []basics.Int8u{10, 20, 30, 40}
	b.BlendPix(px, 200, 200, 200, 200, 0)
	if px[0] != 10 || px[1] != 20 || px[2] != 30 || px[3] != 40 {
		t.Fatalf("cover 0 must not write, got %v", px)
	}

	// Opaque source over anything reproduces the source exactly.
	px = []basics.Int8u{10, 20, 30, 40}
	b.BlendPix(px, 200, 100, 50, 255, 255)
	if px[0] != 200 || px[1] != 100 || px[2] != 50 || px[3] != 255 {
		t.Fatalf("opaque blend = %v, want [200 100 50 255]", px)
	}
}

// TestBlenderRGBA8PlainFixedRoundTripsPlainStorage guards the accessors, which
// must stay straight (non-premultiplied) like BlenderRGBA8Plain's.
func TestBlenderRGBA8PlainFixedRoundTripsPlainStorage(t *testing.T) {
	var b BlenderRGBA8PlainFixed[color.Linear, order.RGBA]
	px := make([]basics.Int8u, 4)
	b.SetPlain(px, 1, 2, 3, 4)
	if r, g, bb, a := b.GetPlain(px); r != 1 || g != 2 || bb != 3 || a != 4 {
		t.Fatalf("round trip = %d %d %d %d, want 1 2 3 4", r, g, bb, a)
	}
}
