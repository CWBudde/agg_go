package color

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"agg_go/internal/basics"
)

// ---------- scalar conversion tests ----------

func TestScalarSRGBToLinearAndBack(t *testing.T) {
	// Check endpoints and a sweep for near-inverse behavior.
	// Using tight epsilon because functions are exact piecewise except pow rounding.
	const eps = 1e-12

	edges := []float64{0, 0.0031308, 0.04045, 0.5, 1}
	for _, v := range edges {
		lin := ConvertFromSRGB(v)
		back := ConvertToSRGB(lin)
		if !nearlyEqual(v, back, eps) {
			t.Fatalf("edge mismatch: v=%g -> lin=%g -> srgb=%g", v, lin, back)
		}
	}

	// Dense sweep
	for i := 0; i <= 10000; i++ {
		v := float64(i) / 10000.0
		lin := ConvertFromSRGB(v)
		back := ConvertToSRGB(lin)
		if !nearlyEqual(v, back, 2e-12) { // slightly looser in sweep
			t.Fatalf("sweep mismatch: v=%g -> lin=%g -> srgb=%g", v, lin, back)
		}
	}
}

// ---------- table vs. scalar correctness ----------

func TestSRGB8ToLinear8AgainstScalar(t *testing.T) {
	// The table computes: lin8 = round(ConvertFromSRGB(i/255)*255)
	for i := 0; i < 256; i++ {
		initSRGBTables()
		got := srgb8ToLinear8(basics.Int8u(i))
		want := basics.Int8u(ConvertFromSRGB(float64(i)/255.0)*255.0 + 0.5)
		if got != want {
			t.Fatalf("srgb8->linear8 mismatch at %d: got %d want %d", i, got, want)
		}
	}
}

func TestLinear8ToSRGB8AgainstScalar(t *testing.T) {
	// The table computes: s8 = round(ConvertToSRGB(i/255)*255)
	for i := 0; i < 256; i++ {
		initSRGBTables()
		got := linear8ToSrgb8(basics.Int8u(i))
		want := basics.Int8u(ConvertToSRGB(float64(i)/255.0)*255.0 + 0.5)
		if got != want {
			t.Fatalf("linear8->srgb8 mismatch at %d: got %d want %d", i, got, want)
		}
	}
}

func TestSRGB8ToLinearF32AgainstScalar(t *testing.T) {
	// F32 table stores the float32 of ConvertFromSRGB(i/255).
	for i := 0; i < 256; i++ {
		got := srgb8ToLinearF32(basics.Int8u(i))
		want := float32(ConvertFromSRGB(float64(i) / 255.0))
		if !f32NearlyEqual(got, want, 1e-7) {
			t.Fatalf("srgb8->linearF32 mismatch at %d: got %g want %g", i, got, want)
		}
	}
}

// ---------- round-trip properties (with expected quantization) ----------

func TestU8RoundTripLandsInCollapsedBucket(t *testing.T) {
	for i := 0; i < 256; i++ {
		l8 := srgb8ToLinear8(basics.Int8u(i))

		// Find the contiguous bucket [lo,hi] of all sRGB values that map to this l8
		lo, hi := 0, 255
		for j := i; j >= 0; j-- {
			if srgb8ToLinear8(basics.Int8u(j)) == l8 {
				lo = j
			} else {
				break
			}
		}
		for j := i; j < 256; j++ {
			if srgb8ToLinear8(basics.Int8u(j)) == l8 {
				hi = j
			} else {
				break
			}
		}

		back := linear8ToSrgb8(l8)
		if int(back) < lo || int(back) > hi {
			t.Fatalf("roundtrip landed outside collapsed bucket: i=%d l8=%d back=%d bucket=[%d,%d]",
				i, l8, back, lo, hi)
		}
	}
}

func TestLUTsAreMonotonic(t *testing.T) {
	for i := 0; i < 255; i++ {
		if srgb8ToLinear8(basics.Int8u(i)) > srgb8ToLinear8(basics.Int8u(i+1)) {
			t.Fatalf("srgb8->linear8 not monotonic at %d", i)
		}
		if linear8ToSrgb8(basics.Int8u(i)) > linear8ToSrgb8(basics.Int8u(i+1)) {
			t.Fatalf("linear8->srgb8 not monotonic at %d", i)
		}
	}
}

func TestRoundTripFloat32WithinSmallEpsilon(t *testing.T) {
	const eps = 5e-7 // f32-compatible tolerance
	r := rand.New(rand.NewSource(123))
	for i := 0; i < 20000; i++ {
		v := r.Float64()
		l := ConvertFromSRGB(v)
		s := ConvertToSRGB(l)
		if !nearlyEqual(v, s, eps) {
			t.Fatalf("f32 roundtrip mismatch: v=%g back=%g", v, s)
		}
	}
}

// ---------- alpha passthrough ----------

func TestAlphaPassthroughU8(t *testing.T) {
	for i := 0; i < 256; i++ {
		a := basics.Int8u(i)
		if alphaU8FromSRGB(a) != a || alphaU8ToSRGB(a) != a {
			t.Fatalf("alpha passthrough failed at %d", i)
		}
	}
}

func TestAlphaPassthroughF32(t *testing.T) {
	for i := 0; i < 256; i++ {
		a8 := basics.Int8u(i)
		f := alphaF32FromSRGB(a8)
		back := alphaF32ToSRGB(f)
		// back is rounded to nearest; must be exactly original
		if back != a8 {
			t.Fatalf("alpha f32 passthrough failed: a=%d f=%g back=%d", a8, f, back)
		}
	}
}

// ---------- struct conversions ----------

func TestRGBA8ConversionsKeepAlpha(t *testing.T) {
	src := RGBA8[Linear]{R: 10, G: 20, B: 30, A: 77}
	s := ConvertRGBA8LinearToSRGB(src)
	if s.A != src.A {
		t.Fatalf("alpha changed on Linear->sRGB: got %d want %d", s.A, src.A)
	}
	back := ConvertRGBA8SRGBToLinear(s)
	if back.A != src.A {
		t.Fatalf("alpha changed on sRGB->Linear: got %d want %d", back.A, src.A)
	}
}

func TestRGB8ConversionsRoundTripIntoBucket(t *testing.T) {
	src := RGB8[SRGB]{R: 5, G: 120, B: 240}
	lin := ConvertRGB8SRGBToLinear(src)
	back := ConvertRGB8LinearToSRGB(lin)

	chk := func(name string, orig, got basics.Int8u) {
		lo, hi, _ := srgbCollapsedBucket(int(orig))
		if int(got) < lo || int(got) > hi {
			t.Fatalf("%s roundtrip outside collapsed bucket: orig=%d back=%d bucket=[%d,%d]", name, orig, got, lo, hi)
		}
	}

	chk("R", src.R, back.R)
	chk("G", src.G, back.G)
	chk("B", src.B, back.B)
}

func TestGray8ConversionsBucketAndAlpha(t *testing.T) {
	g := Gray8[SRGB]{V: 17, A: 200}
	l := ConvertGray8SRGBToLinear(g)
	s := ConvertGray8LinearToSRGB(l)

	// Value must land inside the collapsed bucket for the original V.
	lo, hi, _ := srgbCollapsedBucket(int(g.V))
	if int(s.V) < lo || int(s.V) > hi {
		t.Fatalf("Gray8 V roundtrip outside bucket: V=%d back=%d bucket=[%d,%d]", g.V, s.V, lo, hi)
	}

	// Alpha is passthrough-exact.
	if s.A != g.A {
		t.Fatalf("Gray8 alpha changed: %d -> %d", g.A, s.A)
	}
}

func TestGray16Approximation(t *testing.T) {
	// Gray16 conversions downsample to 8-bit, convert, then replicate (V<<8|V).
	// Test that property holds and alpha is preserved via the same scheme.
	vals := []basics.Int16u{0x0000, 0x0101, 0x1234, 0x7F7F, 0x8080, 0xFFFF}
	for _, v := range vals {
		g := Gray16[Linear]{V: v, A: 0xBEEF}
		s := ConvertGray16LinearToSRGB(g)
		// Expected: v8 = v>>8; s8 = linear8ToSrgb8(v8); s.V == replicate(s8)
		v8 := basics.Int8u(v >> 8)
		wantV8 := linear8ToSrgb8(v8)
		wantV16 := basics.Int16u(wantV8)
		wantV16 = (wantV16 << 8) | wantV16
		if s.V != wantV16 {
			t.Fatalf("Gray16 Linear->sRGB mismatch: V=%#04x got=%#04x want=%#04x", v, s.V, wantV16)
		}
		// Alpha passthrough via same replication
		a8 := basics.Int8u(g.A >> 8)
		wantA16 := basics.Int16u(alphaU8ToSRGB(a8))
		wantA16 = (wantA16 << 8) | wantA16
		if s.A != wantA16 {
			t.Fatalf("Gray16 alpha mismatch: got=%#04x want=%#04x", s.A, wantA16)
		}

		// And back the other way
		back := ConvertGray16SRGBToLinear(s)
		bv8 := basics.Int8u(s.V >> 8)
		wantBackV8 := srgb8ToLinear8(bv8)
		wantBackV16 := basics.Int16u(wantBackV8)
		wantBackV16 = (wantBackV16 << 8) | wantBackV16
		if back.V != wantBackV16 {
			t.Fatalf("Gray16 sRGB->Linear mismatch: got=%#04x want=%#04x", back.V, wantBackV16)
		}
	}
}

func TestRGBA32FloatPaths(t *testing.T) {
	src := RGBA32[Linear]{R: 0.1, G: 0.5, B: 0.9, A: 0.25}
	s := ConvertRGBA32LinearToSRGB(src)
	back := ConvertRGBA32SRGBToLinear(s)

	// Allow a little float32 error
	const eps = float32(1e-6)
	if !f32NearlyEqual(src.R, back.R, eps) ||
		!f32NearlyEqual(src.G, back.G, eps) ||
		!f32NearlyEqual(src.B, back.B, eps) {
		t.Fatalf("RGBA32 roundtrip mismatch: src=%+v back=%+v", src, back)
	}
	if s.A != src.A || back.A != src.A {
		t.Fatalf("RGBA32 alpha changed across conversions")
	}
}

func TestGray32FloatPaths(t *testing.T) {
	src := Gray32[Linear]{V: 0.42, A: 0.75}
	s := ConvertGray32LinearToSRGB(src)
	back := ConvertGray32SRGBToLinear(s)
	if !f32NearlyEqual(src.V, back.V, 1e-6) {
		t.Fatalf("Gray32 roundtrip mismatch: src=%+v back=%+v", src, back)
	}
	if s.A != src.A || back.A != src.A {
		t.Fatalf("Gray32 alpha changed across conversions")
	}
}

// ---------- RGBA8Pre helper ----------

func TestRGBA8PreHelper(t *testing.T) {
	tests := []struct {
		r, g, b, a float64
	}{
		{0, 0, 0, 0},
		{1, 1, 1, 1},
		{1, 0, 0, 0.5},
		{0.25, 0.5, 0.75, 0.66},
		{-0.1, 1.2, 0.5, 0.5}, // out-of-range inputs; current code doesn't clamp, so test current behavior
	}
	for _, tt := range tests {
		c := RGBA8Pre[Linear](tt.r, tt.g, tt.b, tt.a)
		// Expected current behavior: direct multiply + scale + round, no clamping.
		expR := basics.Int8u(tt.r*tt.a*255.0 + 0.5)
		expG := basics.Int8u(tt.g*tt.a*255.0 + 0.5)
		expB := basics.Int8u(tt.b*tt.a*255.0 + 0.5)
		expA := basics.Int8u(tt.a*255.0 + 0.5)
		if c.R != expR || c.G != expG || c.B != expB || c.A != expA {
			t.Fatalf("RGBA8Pre mismatch for %+v: got=%v want=(%d,%d,%d,%d)", tt, c, expR, expG, expB, expA)
		}
	}
}

// ---------- Gray8 -> RGBA8 helpers ----------

func TestMakeRGBA8FromGray8(t *testing.T) {
	gL := Gray8[Linear]{V: 17, A: 200}
	rgbL := MakeRGBA8FromGray8Linear[Linear](gL)
	if rgbL.R != gL.V || rgbL.G != gL.V || rgbL.B != gL.V || rgbL.A != gL.A {
		t.Fatalf("MakeRGBA8FromGray8Linear mismatch: %+v -> %+v", gL, rgbL)
	}
	gS := Gray8[SRGB]{V: 33, A: 111}
	rgbS := MakeRGBA8FromGray8SRGB[SRGB](gS)
	if rgbS.R != gS.V || rgbS.G != gS.V || rgbS.B != gS.V || rgbS.A != gS.A {
		t.Fatalf("MakeRGBA8FromGray8SRGB mismatch: %+v -> %+v", gS, rgbS)
	}

	rgbS2 := MakeSRGBA8FromGray8Linear[SRGB](gL)
	expectV := linear8ToSrgb8(gL.V)
	if rgbS2.R != expectV || rgbS2.G != expectV || rgbS2.B != expectV || rgbS2.A != gL.A {
		t.Fatalf("MakeSRGBA8FromGray8Linear mismatch: %+v -> %+v (expect V=%d)", gL, rgbS2, expectV)
	}

	rgbL2 := MakeRGBA8FromGray8SRGB_ToLinear[Linear](gS)
	expectV2 := srgb8ToLinear8(gS.V)
	if rgbL2.R != expectV2 || rgbL2.G != expectV2 || rgbL2.B != expectV2 || rgbL2.A != gS.A {
		t.Fatalf("MakeRGBA8FromGray8SRGB_ToLinear mismatch: %+v -> %+v (expect V=%d)", gS, rgbL2, expectV2)
	}
}

// ---------- fuzz (Go 1.18+) ----------

func FuzzScalarNearInverse(f *testing.F) {
	seed := []float64{0, 1e-6, 0.0031308, 0.04045, 0.5, 0.999999, 1}
	for _, s := range seed {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, v float64) {
		// Limit inputs to [0,1], but still exercise out-of-range by clamping.
		v = clamp01(v)
		lin := ConvertFromSRGB(v)
		back := ConvertToSRGB(lin)
		if !nearlyEqual(v, back, 1e-9) {
			t.Fatalf("near-inverse failed: v=%g -> lin=%g -> srgb=%g", v, lin, back)
		}
	})
}

// ---------- benchmarks ----------

func BenchmarkScalarConvertFromSRGB(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	vals := make([]float64, 1<<12)
	for i := range vals {
		vals[i] = r.Float64()
	}
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink += ConvertFromSRGB(vals[i&(len(vals)-1)])
	}
	_ = sink
}

func BenchmarkTableSRGB8ToLinear8(b *testing.B) {
	initSRGBTables()
	var sink basics.Int8u
	for i := 0; i < b.N; i++ {
		sink += srgb8ToLinear8(basics.Int8u(i))
	}
	_ = sink
}

// ---------- small helpers ----------

func nearlyEqual(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

func f32NearlyEqual(a, b, eps float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// helper: find the [lo,hi] bucket of sRGB codes that map to the same linear8 value as 'i'.
func srgbCollapsedBucket(i int) (lo, hi int, l8 basics.Int8u) {
	l8 = srgb8ToLinear8(basics.Int8u(i))
	lo, hi = i, i
	for j := i - 1; j >= 0; j-- {
		if srgb8ToLinear8(basics.Int8u(j)) == l8 {
			lo = j
		} else {
			break
		}
	}
	for j := i + 1; j < 256; j++ {
		if srgb8ToLinear8(basics.Int8u(j)) == l8 {
			hi = j
		} else {
			break
		}
	}
	return lo, hi, l8
}
