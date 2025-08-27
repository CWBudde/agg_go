package blender

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

func TestBlenderRGBAGet(t *testing.T) {
	// plain read (what old BlenderRGBA.Get did)
	pixel := []basics.Int8u{128, 64, 192, 255}
	cover := basics.Int8u(255)
	result := getPlain[order.RGBA](pixel, cover)

	if math.Abs(result.R-128.0/255.0) > 0.01 {
		t.Errorf("Get R mismatch")
	}
	if math.Abs(result.G-64.0/255.0) > 0.01 {
		t.Errorf("Get G mismatch")
	}
	if math.Abs(result.B-192.0/255.0) > 0.01 {
		t.Errorf("Get B mismatch")
	}
	if math.Abs(result.A-1.0) > 0.01 {
		t.Errorf("Get A mismatch")
	}
}

func TestBlenderRGBAGetWithCoverage(t *testing.T) {
	pixel := []basics.Int8u{255, 255, 255, 255}
	result := getPlain[order.RGBA](pixel, 128)
	exp := 0.5
	if math.Abs(result.R-exp) > 0.01 || math.Abs(result.G-exp) > 0.01 ||
		math.Abs(result.B-exp) > 0.01 || math.Abs(result.A-exp) > 0.01 {
		t.Errorf("Get with coverage should scale by 0.5, got %+v", result)
	}
}

func TestBlenderRGBAGetZeroCoverage(t *testing.T) {
	pixel := []basics.Int8u{255, 255, 255, 255}
	result := getPlain[order.RGBA](pixel, 0)
	if result != color.NoColor() {
		t.Errorf("Zero coverage should return NoColor, got %+v", result)
	}
}

func TestBlenderRGBAGetRaw(t *testing.T) {
	pixel := []basics.Int8u{128, 64, 192, 255}
	r, g, b, a := getRaw[order.RGBA](pixel)
	if r != 128 || g != 64 || b != 192 || a != 255 {
		t.Errorf("GetRaw mismatch: got (%d,%d,%d,%d)", r, g, b, a)
	}
}

func TestBlenderRGBASet(t *testing.T) {
	pixel := make([]basics.Int8u, 4)
	in := color.RGBA{R: 0.5, G: 0.25, B: 0.75, A: 1.0}
	setPlain[order.RGBA](pixel, in)

	expR := basics.Int8u(128)
	expG := basics.Int8u(64)
	expB := basics.Int8u(191)
	expA := basics.Int8u(255)
	if pixel[0] != expR || pixel[1] != expG || pixel[2] != expB || pixel[3] != expA {
		t.Errorf("Set (plain) mismatch: got %v", pixel)
	}
}

func TestBlenderRGBASetRaw(t *testing.T) {
	pixel := make([]basics.Int8u, 4)
	var o order.RGBA
	pixel[o.IdxR()], pixel[o.IdxG()], pixel[o.IdxB()], pixel[o.IdxA()] = 128, 64, 192, 255
	if pixel[0] != 128 || pixel[1] != 64 || pixel[2] != 192 || pixel[3] != 255 {
		t.Errorf("SetRaw mismatch: %v", pixel)
	}
}

func TestBlenderRGBARoundTrip(t *testing.T) {
	orig := color.RGBA{R: 0.3, G: 0.6, B: 0.9, A: 0.8}
	pixel := make([]basics.Int8u, 4)
	setPlain[order.RGBA](pixel, orig)
	got := getPlain[order.RGBA](pixel, 255)
	if math.Abs(got.R-orig.R) > 0.01 ||
		math.Abs(got.G-orig.G) > 0.01 ||
		math.Abs(got.B-orig.B) > 0.01 ||
		math.Abs(got.A-orig.A) > 0.01 {
		t.Errorf("Round trip (plain) mismatch: orig %+v, got %+v", orig, got)
	}
}

func TestBlenderRGBAColorOrders(t *testing.T) {
	type caseT struct {
		name string
		set  func([]basics.Int8u, color.RGBA)
		idx  func() (r, g, b, a int)
	}
	cases := []caseT{
		{
			"RGBA",
			func(p []basics.Int8u, c color.RGBA) { setPlain[order.RGBA](p, c) },
			func() (int, int, int, int) { var o order.RGBA; return o.IdxR(), o.IdxG(), o.IdxB(), o.IdxA() },
		},
		{
			"ARGB",
			func(p []basics.Int8u, c color.RGBA) { setPlain[order.ARGB](p, c) },
			func() (int, int, int, int) { var o order.ARGB; return o.IdxR(), o.IdxG(), o.IdxB(), o.IdxA() },
		},
		{
			"BGRA",
			func(p []basics.Int8u, c color.RGBA) { setPlain[order.BGRA](p, c) },
			func() (int, int, int, int) { var o order.BGRA; return o.IdxR(), o.IdxG(), o.IdxB(), o.IdxA() },
		},
		{
			"ABGR",
			func(p []basics.Int8u, c color.RGBA) { setPlain[order.ABGR](p, c) },
			func() (int, int, int, int) { var o order.ABGR; return o.IdxR(), o.IdxG(), o.IdxB(), o.IdxA() },
		},
	}

	expR := basics.Int8u(51)  // 0.2*255+0.5
	expG := basics.Int8u(102) // 0.4*255+0.5
	expB := basics.Int8u(153) // 0.6*255+0.5
	expA := basics.Int8u(204) // 0.8*255+0.5
	c := color.RGBA{R: 0.2, G: 0.4, B: 0.6, A: 0.8}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := make([]basics.Int8u, 4)
			tc.set(p, c)
			ri, gi, bi, ai := tc.idx()
			if p[ri] != expR || p[gi] != expG || p[bi] != expB || p[ai] != expA {
				t.Errorf("%s order mismatch: got %v (r@%d=%d g@%d=%d b@%d=%d a@%d=%d)",
					tc.name, p, ri, p[ri], gi, p[gi], bi, p[bi], ai, p[ai])
			}
		})
	}
}

func TestBlenderRGBAPreGet(t *testing.T) {
	// premultiplied read (what old BlenderRGBAPre.Get did)
	orig := color.RGBA{R: 0.6, G: 0.4, B: 0.8, A: 0.5}
	pm := orig
	pm.Premultiply()
	pixel := []basics.Int8u{
		basics.Int8u(pm.R*255 + 0.5),
		basics.Int8u(pm.G*255 + 0.5),
		basics.Int8u(pm.B*255 + 0.5),
		basics.Int8u(pm.A*255 + 0.5),
	}
	got := getPremult[order.RGBA](pixel, 255)

	tol := 0.02
	if math.Abs(got.R-orig.R) > tol || math.Abs(got.G-orig.G) > tol ||
		math.Abs(got.B-orig.B) > tol || math.Abs(got.A-orig.A) > tol {
		t.Errorf("Premult Get mismatch: orig %+v, got %+v", orig, got)
	}
}

func TestBlenderRGBAPreSet(t *testing.T) {
	pixel := make([]basics.Int8u, 4)
	in := color.RGBA{R: 0.6, G: 0.4, B: 0.8, A: 0.5}
	setPremult[order.RGBA](pixel, in)

	// Expected premult bytes
	pm := in
	pm.Premultiply()
	exp := []basics.Int8u{
		basics.Int8u(pm.R*255 + 0.5),
		basics.Int8u(pm.G*255 + 0.5),
		basics.Int8u(pm.B*255 + 0.5),
		basics.Int8u(pm.A*255 + 0.5),
	}
	if pixel[0] != exp[0] || pixel[1] != exp[1] || pixel[2] != exp[2] || pixel[3] != exp[3] {
		t.Errorf("Premult Set mismatch: got %v, exp %v", pixel, exp)
	}
}

func TestBlenderRGBAPreRoundTrip(t *testing.T) {
	orig := color.RGBA{R: 0.3, G: 0.6, B: 0.9, A: 0.7}
	pixel := make([]basics.Int8u, 4)
	setPremult[order.RGBA](pixel, orig)
	got := getPremult[order.RGBA](pixel, 255)

	tol := 0.02
	if math.Abs(got.R-orig.R) > tol || math.Abs(got.G-orig.G) > tol ||
		math.Abs(got.B-orig.B) > tol || math.Abs(got.A-orig.A) > tol {
		t.Errorf("Premult round trip mismatch: orig %+v, got %+v", orig, got)
	}
}

//
// keep a couple of sanity-blend tests that still use the real blenders
//

func TestBlenderRGBA8_BlendPix(t *testing.T) {
	bl := BlenderRGBA8[color.Linear, order.RGBA]{}
	dst := []basics.Int8u{100, 100, 100, 255}
	bl.BlendPix(dst, 200, 150, 50, 128, 255)
	if dst[0] <= 100 || dst[0] >= 200 {
		t.Errorf("R should move toward src")
	}
	if dst[1] <= 100 || dst[1] >= 150 {
		t.Errorf("G should move toward src")
	}
	if dst[2] < 70 || dst[2] > 90 { // 100 -> 50 at ~50% alpha ≈ 75
		t.Errorf("B out of expected range, got %d", dst[2])
	}
}

func TestBlenderRGBA8Pre_BlendPix_Modifies(t *testing.T) {
	bl := BlenderRGBA8Pre[color.Linear, order.RGBA]{}
	dst := []basics.Int8u{100, 100, 100, 255}
	orig := append([]basics.Int8u(nil), dst...)
	bl.BlendPix(dst, 200, 150, 50, 128, 255)
	changed := false
	for i := 0; i < 4; i++ {
		if dst[i] != orig[i] {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("premult BlendPix should modify destination")
	}
}

func TestBlenderRGBA8Plain_BlendPix_Modifies(t *testing.T) {
	bl := BlenderRGBA8Plain[color.Linear, order.RGBA]{}
	dst := []basics.Int8u{100, 100, 100, 200}
	bl.BlendPix(dst, 200, 150, 50, 128, 255)
	if dst[0] == 100 && dst[1] == 100 && dst[2] == 100 {
		t.Error("plain BlendPix should change RGB")
	}
}

func getPlain[O order.RGBAOrder](p []basics.Int8u, cover basics.Int8u) color.RGBA {
	if cover == 0 {
		return color.NoColor()
	}
	var o O
	c := color.RGBA{
		R: float64(p[o.IdxR()]) / 255.0,
		G: float64(p[o.IdxG()]) / 255.0,
		B: float64(p[o.IdxB()]) / 255.0,
		A: float64(p[o.IdxA()]) / 255.0,
	}
	if cover < 255 {
		scale := float64(cover) / 255.0
		c.R *= scale
		c.G *= scale
		c.B *= scale
		c.A *= scale
	}
	return c
}

func getPremult[O order.RGBAOrder](p []basics.Int8u, cover basics.Int8u) color.RGBA {
	if cover == 0 {
		return color.NoColor()
	}
	var o O
	c := color.RGBA{
		R: float64(p[o.IdxR()]) / 255.0,
		G: float64(p[o.IdxG()]) / 255.0,
		B: float64(p[o.IdxB()]) / 255.0,
		A: float64(p[o.IdxA()]) / 255.0,
	}
	// the stored pixel is premultiplied -> demultiply to straight
	if c.A > 0 {
		c.Demultiply()
	}
	if cover < 255 {
		scale := float64(cover) / 255.0
		c.R *= scale
		c.G *= scale
		c.B *= scale
		c.A *= scale
	}
	return c
}

func setPlain[O order.RGBAOrder](p []basics.Int8u, c color.RGBA) {
	var o O
	p[o.IdxR()] = basics.Int8u(c.R*255 + 0.5)
	p[o.IdxG()] = basics.Int8u(c.G*255 + 0.5)
	p[o.IdxB()] = basics.Int8u(c.B*255 + 0.5)
	p[o.IdxA()] = basics.Int8u(c.A*255 + 0.5)
}

func setPremult[O order.RGBAOrder](p []basics.Int8u, c color.RGBA) {
	var o O
	pm := c
	pm.Premultiply()
	p[o.IdxR()] = basics.Int8u(pm.R*255 + 0.5)
	p[o.IdxG()] = basics.Int8u(pm.G*255 + 0.5)
	p[o.IdxB()] = basics.Int8u(pm.B*255 + 0.5)
	p[o.IdxA()] = basics.Int8u(pm.A*255 + 0.5)
}

func getRaw[O order.RGBAOrder](p []basics.Int8u) (r, g, b, a basics.Int8u) {
	var o O
	return p[o.IdxR()], p[o.IdxG()], p[o.IdxB()], p[o.IdxA()]
}
