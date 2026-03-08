package simd

import (
	"runtime"
	"testing"
)

func TestDetectFeatures(t *testing.T) {
	t.Parallel()

	ResetDetection()
	t.Cleanup(ResetDetection)

	features := DetectFeatures()
	if features.Architecture != runtime.GOARCH {
		t.Fatalf("DetectFeatures architecture = %q, want %q", features.Architecture, runtime.GOARCH)
	}

	if runtime.GOARCH == "amd64" && !features.HasSSE2 {
		t.Fatal("amd64 should report SSE2 support")
	}
	if runtime.GOARCH == "arm64" && !features.HasNEON {
		t.Fatal("arm64 should report NEON support")
	}
}

func TestFillRGBA(t *testing.T) {
	t.Parallel()

	dst := make([]byte, 5*4)
	FillRGBA(dst, 1, 2, 3, 4, 5)

	for i := 0; i < 5; i++ {
		p := i * 4
		if got := dst[p : p+4]; got[0] != 1 || got[1] != 2 || got[2] != 3 || got[3] != 4 {
			t.Fatalf("pixel %d = %v, want [1 2 3 4]", i, got)
		}
	}
}

func TestImplementationNameWithForcedFeatures(t *testing.T) {
	t.Cleanup(ResetDetection)
	actual := DetectFeatures()

	switch runtime.GOARCH {
	case "amd64":
		SetForcedFeatures(Features{Architecture: "amd64", HasSSE2: true})
		if got := ImplementationName(); got != "sse2" {
			t.Fatalf("ImplementationName() = %q, want sse2", got)
		}

		if actual.HasSSSE3 && actual.HasSSE41 {
			SetForcedFeatures(Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true})
			if got := ImplementationName(); got != "sse41" {
				t.Fatalf("ImplementationName() = %q, want sse41", got)
			}
		}

		if actual.HasAVX2 {
			SetForcedFeatures(Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true})
			if got := ImplementationName(); got != "avx2" {
				t.Fatalf("ImplementationName() = %q, want avx2", got)
			}
		}

		SetForcedFeatures(Features{Architecture: "amd64", HasAVX2: true, ForceGeneric: true})
		if got := ImplementationName(); got != "generic" {
			t.Fatalf("ImplementationName() = %q, want generic", got)
		}
	case "arm64":
		SetForcedFeatures(Features{Architecture: "arm64", HasNEON: true})
		if got := ImplementationName(); got != "neon" {
			t.Fatalf("ImplementationName() = %q, want neon", got)
		}

		SetForcedFeatures(Features{Architecture: "arm64", HasNEON: true, ForceGeneric: true})
		if got := ImplementationName(); got != "generic" {
			t.Fatalf("ImplementationName() = %q, want generic", got)
		}
	default:
		SetForcedFeatures(Features{Architecture: runtime.GOARCH})
		if got := ImplementationName(); got != "generic" {
			t.Fatalf("ImplementationName() = %q, want generic", got)
		}
	}
}

func TestFillRGBAForcedPaths(t *testing.T) {
	t.Cleanup(ResetDetection)

	cases := []struct {
		name     string
		features Features
		wantImpl string
	}{
		{
			name:     "generic",
			features: Features{Architecture: runtime.GOARCH, ForceGeneric: true},
			wantImpl: "generic",
		},
	}

	actual := DetectFeatures()
	if runtime.GOARCH == "amd64" {
		cases = append(cases,
			struct {
				name     string
				features Features
				wantImpl string
			}{
				name:     "sse2",
				features: Features{Architecture: "amd64", HasSSE2: true},
				wantImpl: "sse2",
			},
		)
		if actual.HasSSSE3 && actual.HasSSE41 {
			cases = append(cases, struct {
				name     string
				features Features
				wantImpl string
			}{
				name:     "sse41",
				features: Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true},
				wantImpl: "sse41",
			})
		}
		if actual.HasAVX2 {
			cases = append(cases, struct {
				name     string
				features Features
				wantImpl string
			}{
				name:     "avx2",
				features: Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true},
				wantImpl: "avx2",
			})
		}
	}

	if runtime.GOARCH == "arm64" {
		cases = append(cases, struct {
			name     string
			features Features
			wantImpl string
		}{
			name:     "neon",
			features: Features{Architecture: "arm64", HasNEON: true},
			wantImpl: "neon",
		})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ResetDetection()
			SetForcedFeatures(tc.features)

			if got := ImplementationName(); got != tc.wantImpl {
				t.Fatalf("ImplementationName() = %q, want %q", got, tc.wantImpl)
			}

			dst := make([]byte, 7*4)
			FillRGBA(dst, 0x11, 0x22, 0x33, 0x44, 7)
			for i := 0; i < 7; i++ {
				p := i * 4
				if got := dst[p : p+4]; got[0] != 0x11 || got[1] != 0x22 || got[2] != 0x33 || got[3] != 0x44 {
					t.Fatalf("pixel %d = %v, want [17 34 51 68]", i, got)
				}
			}
		})
	}
}

func TestBlendSolidHspanRGBAForcedPaths(t *testing.T) {
	t.Cleanup(ResetDetection)

	type tc struct {
		name      string
		features  Features
		wantImpl  string
		premulSrc bool
	}

	cases := []tc{
		{
			name:      "generic_plain",
			features:  Features{Architecture: runtime.GOARCH, ForceGeneric: true},
			wantImpl:  "generic",
			premulSrc: false,
		},
		{
			name:      "generic_pre",
			features:  Features{Architecture: runtime.GOARCH, ForceGeneric: true},
			wantImpl:  "generic",
			premulSrc: true,
		},
	}

	actual := DetectFeatures()
	if runtime.GOARCH == "amd64" {
		cases = append(cases, tc{name: "sse2_plain", features: Features{Architecture: "amd64", HasSSE2: true}, wantImpl: "sse2"})
		if actual.HasSSSE3 && actual.HasSSE41 {
			cases = append(cases, tc{name: "sse41_plain", features: Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true}, wantImpl: "sse41"})
		}
		if actual.HasAVX2 {
			cases = append(cases, tc{name: "avx2_plain", features: Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true}, wantImpl: "avx2"})
		}
	}

	if runtime.GOARCH == "arm64" {
		cases = append(cases, tc{name: "neon_plain", features: Features{Architecture: "arm64", HasNEON: true}, wantImpl: "neon"})
	}

	covers := []byte{0, 255, 17, 255, 64, 255, 200}
	base := []byte{
		10, 20, 30, 40,
		50, 60, 70, 80,
		90, 100, 110, 120,
		130, 140, 150, 160,
		170, 180, 190, 200,
		210, 220, 230, 240,
		250, 240, 230, 220,
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ResetDetection()
			SetForcedFeatures(tc.features)
			if got := ImplementationName(); got != tc.wantImpl {
				t.Fatalf("ImplementationName() = %q, want %q", got, tc.wantImpl)
			}

			got := append([]byte(nil), base...)
			want := append([]byte(nil), base...)
			BlendSolidHspanRGBA(got, covers, 20, 40, 60, 255, tc.premulSrc)
			blendSolidHspanRGBAGeneric(want, covers, 20, 40, 60, 255, tc.premulSrc)
			if string(got) != string(want) {
				t.Fatalf("BlendSolidHspanRGBA mismatch:\n got=%v\nwant=%v", got, want)
			}
		})
	}
}

func TestBlendHlineRGBAForcedPaths(t *testing.T) {
	t.Cleanup(ResetDetection)

	type tc struct {
		name      string
		features  Features
		wantImpl  string
		premulSrc bool
	}

	cases := []tc{
		{name: "generic_plain", features: Features{Architecture: runtime.GOARCH, ForceGeneric: true}, wantImpl: "generic"},
		{name: "generic_pre", features: Features{Architecture: runtime.GOARCH, ForceGeneric: true}, wantImpl: "generic", premulSrc: true},
	}

	actual := DetectFeatures()
	if runtime.GOARCH == "amd64" {
		cases = append(cases, tc{name: "sse2_plain", features: Features{Architecture: "amd64", HasSSE2: true}, wantImpl: "sse2"})
		if actual.HasSSSE3 && actual.HasSSE41 {
			cases = append(cases, tc{name: "sse41_plain", features: Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true}, wantImpl: "sse41"})
		}
		if actual.HasAVX2 {
			cases = append(cases, tc{name: "avx2_plain", features: Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true}, wantImpl: "avx2"})
		}
	}
	if runtime.GOARCH == "arm64" {
		cases = append(cases, tc{name: "neon_plain", features: Features{Architecture: "arm64", HasNEON: true}, wantImpl: "neon"})
	}

	base := []byte{
		10, 20, 30, 40,
		50, 60, 70, 80,
		90, 100, 110, 120,
		130, 140, 150, 160,
		170, 180, 190, 200,
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ResetDetection()
			SetForcedFeatures(tc.features)
			if got := ImplementationName(); got != tc.wantImpl {
				t.Fatalf("ImplementationName() = %q, want %q", got, tc.wantImpl)
			}

			got := append([]byte(nil), base...)
			want := append([]byte(nil), base...)
			BlendHlineRGBA(got, 20, 40, 60, 200, 128, len(base)/4, tc.premulSrc)
			blendHlineRGBAGeneric(want, 20, 40, 60, 200, 128, len(base)/4, tc.premulSrc)
			if string(got) != string(want) {
				t.Fatalf("BlendHlineRGBA mismatch:\n got=%v\nwant=%v", got, want)
			}
		})
	}
}

// TestBlendHlineRGBAComprehensive validates bit-identical output for all SIMD
// implementations across a wide range of uniform-coverage blend parameters.
func TestBlendHlineRGBAComprehensive(t *testing.T) {
	t.Cleanup(ResetDetection)

	implCases := []struct {
		name     string
		features Features
	}{
		{"generic", Features{Architecture: runtime.GOARCH, ForceGeneric: true}},
	}
	actual := DetectFeatures()
	if runtime.GOARCH == "amd64" {
		implCases = append(implCases, struct {
			name     string
			features Features
		}{"sse2", Features{Architecture: "amd64", HasSSE2: true}})
		if actual.HasSSSE3 && actual.HasSSE41 {
			implCases = append(implCases, struct {
				name     string
				features Features
			}{"sse41", Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true}})
		}
		if actual.HasAVX2 {
			implCases = append(implCases, struct {
				name     string
				features Features
			}{"avx2", Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true}})
		}
	}
	if runtime.GOARCH == "arm64" {
		implCases = append(implCases, struct {
			name     string
			features Features
		}{"neon", Features{Architecture: "arm64", HasNEON: true}})
	}

	type scenario struct {
		name       string
		r, g, b, a uint8
		cover      uint8
		premul     bool
		count      int
		base       []byte
	}

	mkBase := func(n int, val byte) []byte {
		b := make([]byte, n*4)
		for i := range b {
			b[i] = val
		}
		return b
	}
	mkVaryBase := func(n int) []byte {
		b := make([]byte, n*4)
		for i := range b {
			b[i] = byte(i * 7 % 256)
		}
		return b
	}

	scenarios := []scenario{
		// Opaque + full cover → fill path
		{name: "opaque_full", r: 200, g: 100, b: 50, a: 255, cover: 255, count: 16, base: mkBase(16, 128)},
		// Full cover, partial alpha
		{name: "alpha128_full", r: 200, g: 100, b: 50, a: 128, cover: 255, count: 16, base: mkBase(16, 128)},
		// Partial cover, opaque src
		{name: "cover128_opaque", r: 200, g: 100, b: 50, a: 255, cover: 128, count: 16, base: mkBase(16, 128)},
		// Partial cover, partial alpha
		{name: "alpha64_cover128", r: 100, g: 150, b: 200, a: 64, cover: 128, count: 16, base: mkBase(16, 100)},
		// Single pixel
		{name: "single", r: 255, g: 0, b: 0, a: 200, cover: 100, count: 1, base: []byte{50, 60, 70, 80}},
		// Two pixels (loop boundary)
		{name: "two", r: 0, g: 255, b: 0, a: 255, cover: 200, count: 2, base: mkBase(2, 50)},
		// Three pixels (one SIMD loop + one tail)
		{name: "three", r: 100, g: 100, b: 100, a: 255, cover: 200, count: 3, base: mkBase(3, 200)},
		// Large span (tests SIMD loop)
		{name: "large_64", r: 30, g: 60, b: 90, a: 255, cover: 180, count: 64, base: mkVaryBase(64)},
		// Odd count (tail handling)
		{name: "odd_9", r: 100, g: 150, b: 200, a: 255, cover: 64, count: 9, base: mkVaryBase(9)},
		// Zero alpha (no-op)
		{name: "zero_alpha", r: 255, g: 0, b: 0, a: 0, cover: 255, count: 8, base: mkBase(8, 128)},
		// Zero cover (no-op)
		{name: "zero_cover", r: 255, g: 0, b: 0, a: 255, cover: 0, count: 8, base: mkBase(8, 128)},
		// Premultiplied source
		{name: "premul_full", r: 100, g: 50, b: 25, a: 128, cover: 255, premul: true, count: 16, base: mkVaryBase(16)},
		{name: "premul_partial", r: 100, g: 50, b: 25, a: 128, cover: 128, premul: true, count: 9, base: mkVaryBase(9)},
		// Black on white
		{name: "black_on_white", r: 0, g: 0, b: 0, a: 255, cover: 180, count: 32, base: mkBase(32, 255)},
		// White on black
		{name: "white_on_black", r: 255, g: 255, b: 255, a: 255, cover: 180, count: 32, base: mkBase(32, 0)},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			for _, impl := range implCases {
				t.Run(impl.name, func(t *testing.T) {
					ResetDetection()
					SetForcedFeatures(impl.features)

					got := append([]byte(nil), sc.base...)
					want := append([]byte(nil), sc.base...)

					BlendHlineRGBA(got, sc.r, sc.g, sc.b, sc.a, sc.cover, sc.count, sc.premul)
					blendHlineRGBAGeneric(want, sc.r, sc.g, sc.b, sc.a, sc.cover, sc.count, sc.premul)

					if string(got) != string(want) {
						for i := 0; i < sc.count; i++ {
							p := i * 4
							if got[p] != want[p] || got[p+1] != want[p+1] || got[p+2] != want[p+2] || got[p+3] != want[p+3] {
								t.Fatalf("pixel %d: got=[%d,%d,%d,%d] want=[%d,%d,%d,%d]",
									i, got[p], got[p+1], got[p+2], got[p+3],
									want[p], want[p+1], want[p+2], want[p+3])
							}
						}
						t.Fatal("mismatch but no differing pixel found")
					}
				})
			}
		})
	}
}

func TestBlendColorHspanRGBAForcedPaths(t *testing.T) {
	t.Cleanup(ResetDetection)

	type tc struct {
		name     string
		features Features
		wantImpl string
	}
	cases := []tc{
		{name: "generic", features: Features{Architecture: runtime.GOARCH, ForceGeneric: true}, wantImpl: "generic"},
	}
	actual := DetectFeatures()
	if runtime.GOARCH == "amd64" {
		cases = append(cases, tc{name: "sse2", features: Features{Architecture: "amd64", HasSSE2: true}, wantImpl: "sse2"})
		if actual.HasSSSE3 && actual.HasSSE41 {
			cases = append(cases, tc{name: "sse41", features: Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true}, wantImpl: "sse41"})
		}
		if actual.HasAVX2 {
			cases = append(cases, tc{name: "avx2", features: Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true}, wantImpl: "avx2"})
		}
	}
	if runtime.GOARCH == "arm64" {
		cases = append(cases, tc{name: "neon", features: Features{Architecture: "arm64", HasNEON: true}, wantImpl: "neon"})
	}

	dst := make([]byte, 5*4)
	for i := range dst {
		dst[i] = byte(50 + i*3)
	}
	srcColors := []byte{
		200, 100, 50, 255,
		150, 200, 100, 200,
		50, 150, 200, 128,
		100, 50, 200, 255,
		200, 200, 100, 180,
	}
	covers := []byte{255, 128, 200, 64, 255}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ResetDetection()
			SetForcedFeatures(tc.features)
			if got := ImplementationName(); got != tc.wantImpl {
				t.Fatalf("ImplementationName() = %q, want %q", got, tc.wantImpl)
			}
			got := append([]byte(nil), dst...)
			want := append([]byte(nil), dst...)
			BlendColorHspanRGBA(got, srcColors, covers, len(covers), false)
			blendColorHspanRGBAGeneric(want, srcColors, covers, len(covers), false)
			if string(got) != string(want) {
				t.Fatalf("BlendColorHspanRGBA mismatch:\n got=%v\nwant=%v", got, want)
			}
		})
	}
}

// TestBlendColorHspanRGBAComprehensive validates bit-identical output across all
// SIMD paths for per-pixel color and coverage blending.
func TestBlendColorHspanRGBAComprehensive(t *testing.T) {
	t.Cleanup(ResetDetection)

	implCases := []struct {
		name     string
		features Features
	}{
		{"generic", Features{Architecture: runtime.GOARCH, ForceGeneric: true}},
	}
	actual := DetectFeatures()
	if runtime.GOARCH == "amd64" {
		implCases = append(implCases, struct {
			name     string
			features Features
		}{"sse2", Features{Architecture: "amd64", HasSSE2: true}})
		if actual.HasSSSE3 && actual.HasSSE41 {
			implCases = append(implCases, struct {
				name     string
				features Features
			}{"sse41", Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true}})
		}
		if actual.HasAVX2 {
			implCases = append(implCases, struct {
				name     string
				features Features
			}{"avx2", Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true}})
		}
	}
	if runtime.GOARCH == "arm64" {
		implCases = append(implCases, struct {
			name     string
			features Features
		}{"neon", Features{Architecture: "arm64", HasNEON: true}})
	}

	type scenario struct {
		name      string
		premul    bool
		count     int
		dst       []byte
		srcColors []byte
		covers    []byte // nil = use covers==nil path
	}

	mkDst := func(n int) []byte {
		b := make([]byte, n*4)
		for i := range b {
			b[i] = byte(i*7%256 + 10)
		}
		return b
	}
	mkSrc := func(n int) []byte {
		b := make([]byte, n*4)
		for i := 0; i < n; i++ {
			b[i*4+0] = byte(200 - i*13%200)
			b[i*4+1] = byte(100 + i*17%155)
			b[i*4+2] = byte(50 + i*23%200)
			b[i*4+3] = byte(128 + i*11%127) // non-zero alpha
		}
		return b
	}
	mkCovers := func(n int) []byte {
		c := make([]byte, n)
		for i := range c {
			switch i % 5 {
			case 0:
				c[i] = 255
			case 1:
				c[i] = 128
			case 2:
				c[i] = 0
			case 3:
				c[i] = 64
			case 4:
				c[i] = 200
			}
		}
		return c
	}

	scenarios := []scenario{
		// Single pixel
		{
			name: "single", count: 1,
			dst:       []byte{50, 60, 70, 80},
			srcColors: []byte{200, 100, 50, 255},
			covers:    []byte{200},
		},
		// Two pixels (SIMD loop boundary)
		{
			name: "two", count: 2,
			dst:       mkDst(2),
			srcColors: mkSrc(2),
			covers:    mkCovers(2),
		},
		// Three pixels (one loop + one tail)
		{
			name: "three", count: 3,
			dst:       mkDst(3),
			srcColors: mkSrc(3),
			covers:    mkCovers(3),
		},
		// Large span — exercises the full loop
		{
			name: "large_32", count: 32,
			dst:       mkDst(32),
			srcColors: mkSrc(32),
			covers:    mkCovers(32),
		},
		// Odd count
		{
			name: "odd_9", count: 9,
			dst:       mkDst(9),
			srcColors: mkSrc(9),
			covers:    mkCovers(9),
		},
		// All covers = 255 (opaque)
		{
			name: "all_full_cover", count: 16,
			dst:       mkDst(16),
			srcColors: mkSrc(16),
			covers: func() []byte {
				c := make([]byte, 16)
				for i := range c {
					c[i] = 255
				}
				return c
			}(),
		},
		// All covers = 0 (no-op)
		{
			name: "all_zero_cover", count: 8,
			dst:       mkDst(8),
			srcColors: mkSrc(8),
			covers:    func() []byte { return make([]byte, 8) }(),
		},
		// Some transparent source pixels
		{
			name: "transparent_src", count: 4,
			dst:       mkDst(4),
			srcColors: []byte{200, 100, 50, 0, 150, 100, 50, 255, 200, 100, 50, 0, 100, 200, 50, 128},
			covers:    []byte{255, 255, 128, 200},
		},
		// nil covers → generic path
		{
			name: "nil_covers", count: 8,
			dst:       mkDst(8),
			srcColors: mkSrc(8),
			covers:    nil,
		},
		// Premultiplied source → always generic
		{
			name: "premul_8", premul: true, count: 8,
			dst:       mkDst(8),
			srcColors: mkSrc(8),
			covers:    mkCovers(8),
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			for _, impl := range implCases {
				t.Run(impl.name, func(t *testing.T) {
					ResetDetection()
					SetForcedFeatures(impl.features)

					got := append([]byte(nil), sc.dst...)
					want := append([]byte(nil), sc.dst...)

					BlendColorHspanRGBA(got, sc.srcColors, sc.covers, sc.count, sc.premul)
					blendColorHspanRGBAGeneric(want, sc.srcColors, sc.covers, sc.count, sc.premul)

					if string(got) != string(want) {
						for i := 0; i < sc.count; i++ {
							p := i * 4
							if got[p] != want[p] || got[p+1] != want[p+1] || got[p+2] != want[p+2] || got[p+3] != want[p+3] {
								cv := byte(255)
								if sc.covers != nil {
									cv = sc.covers[i]
								}
								t.Fatalf("pixel %d (cover=%d src=[%d,%d,%d,%d]): got=[%d,%d,%d,%d] want=[%d,%d,%d,%d]",
									i, cv,
									sc.srcColors[i*4], sc.srcColors[i*4+1], sc.srcColors[i*4+2], sc.srcColors[i*4+3],
									got[p], got[p+1], got[p+2], got[p+3],
									want[p], want[p+1], want[p+2], want[p+3])
							}
						}
						t.Fatal("mismatch but no differing pixel found")
					}
				})
			}
		})
	}
}

// TestBlendSolidHspanRGBAComprehensive validates bit-identical output for all
// SIMD implementations across a wide range of parameters.
func TestBlendSolidHspanRGBAComprehensive(t *testing.T) {
	t.Cleanup(ResetDetection)

	implCases := []struct {
		name     string
		features Features
	}{
		{"generic", Features{Architecture: runtime.GOARCH, ForceGeneric: true}},
	}
	actual := DetectFeatures()
	if runtime.GOARCH == "amd64" {
		implCases = append(implCases, struct {
			name     string
			features Features
		}{"sse2", Features{Architecture: "amd64", HasSSE2: true}})
		if actual.HasSSSE3 && actual.HasSSE41 {
			implCases = append(implCases, struct {
				name     string
				features Features
			}{"sse41", Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true}})
		}
		if actual.HasAVX2 {
			implCases = append(implCases, struct {
				name     string
				features Features
			}{"avx2", Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true}})
		}
	}
	if runtime.GOARCH == "arm64" {
		implCases = append(implCases, struct {
			name     string
			features Features
		}{"neon", Features{Architecture: "arm64", HasNEON: true}})
	}

	// Test scenarios covering edge cases and common patterns
	type scenario struct {
		name       string
		r, g, b, a uint8
		premul     bool
		covers     []byte
		base       []byte // per-pixel dst, len must be len(covers)*4
	}

	mkBase := func(n int, val byte) []byte {
		b := make([]byte, n*4)
		for i := range b {
			b[i] = val
		}
		return b
	}

	mkCovers := func(n int, val byte) []byte {
		c := make([]byte, n)
		for i := range c {
			c[i] = val
		}
		return c
	}

	// Mixed cover pattern for larger spans
	mkMixedCovers := func(n int) []byte {
		c := make([]byte, n)
		for i := range c {
			switch i % 5 {
			case 0:
				c[i] = 0
			case 1:
				c[i] = 255
			case 2:
				c[i] = 128
			case 3:
				c[i] = 1
			case 4:
				c[i] = 254
			}
		}
		return c
	}

	scenarios := []scenario{
		// All-zero covers (no-op)
		{"all_zero_covers", 255, 0, 0, 255, false, mkCovers(16, 0), mkBase(16, 128)},
		// All-full covers with opaque source
		{"all_full_opaque", 200, 100, 50, 255, false, mkCovers(16, 255), mkBase(16, 128)},
		// All-full covers with transparent source
		{"all_full_alpha128", 200, 100, 50, 128, false, mkCovers(16, 255), mkBase(16, 128)},
		// Single pixel
		{"single_pixel", 255, 0, 0, 255, false, []byte{128}, []byte{50, 60, 70, 80}},
		// Two pixels (boundary for 2-pixel loop)
		{"two_pixels", 0, 255, 0, 200, false, []byte{255, 64}, []byte{10, 20, 30, 40, 50, 60, 70, 80}},
		// Odd count (tests tail handling)
		{"odd_count_9", 100, 150, 200, 255, false, mkMixedCovers(9), mkBase(9, 100)},
		// Large span with mixed covers
		{"large_mixed_64", 30, 60, 90, 255, false, mkMixedCovers(64), mkBase(64, 200)},
		// Large span with partial alpha
		{"large_alpha_64", 30, 60, 90, 100, false, mkMixedCovers(64), mkBase(64, 200)},
		// Premultiplied source
		{"premul_mixed", 100, 50, 25, 128, true, mkMixedCovers(16), mkBase(16, 150)},
		// Run of 255 covers then partial (tests run-fill boundary)
		{
			"run_then_partial", 255, 128, 0, 255, false,
			append(mkCovers(8, 255), mkCovers(4, 100)...),
			mkBase(12, 64),
		},
		// Alternating 0/255 covers
		{
			"alternating", 200, 100, 50, 255, false,
			func() []byte {
				c := make([]byte, 16)
				for i := range c {
					if i%2 == 0 {
						c[i] = 255
					}
				}
				return c
			}(),
			mkBase(16, 128),
		},
		// Source with alpha=0 (should be no-op for non-premul)
		{"zero_alpha", 255, 0, 0, 0, false, mkCovers(8, 255), mkBase(8, 128)},
		// White on black
		{"white_on_black", 255, 255, 255, 255, false, mkMixedCovers(32), mkBase(32, 0)},
		// Black on white
		{"black_on_white", 0, 0, 0, 255, false, mkMixedCovers(32), mkBase(32, 255)},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			for _, impl := range implCases {
				t.Run(impl.name, func(t *testing.T) {
					ResetDetection()
					SetForcedFeatures(impl.features)

					got := append([]byte(nil), sc.base...)
					want := append([]byte(nil), sc.base...)

					BlendSolidHspanRGBA(got, sc.covers, sc.r, sc.g, sc.b, sc.a, sc.premul)
					blendSolidHspanRGBAGeneric(want, sc.covers, sc.r, sc.g, sc.b, sc.a, sc.premul)

					if string(got) != string(want) {
						// Find first mismatch pixel for diagnostics
						for i := 0; i < len(sc.covers); i++ {
							p := i * 4
							if got[p] != want[p] || got[p+1] != want[p+1] || got[p+2] != want[p+2] || got[p+3] != want[p+3] {
								t.Fatalf("pixel %d (cover=%d): got=[%d,%d,%d,%d] want=[%d,%d,%d,%d]",
									i, sc.covers[i],
									got[p], got[p+1], got[p+2], got[p+3],
									want[p], want[p+1], want[p+2], want[p+3])
							}
						}
						t.Fatal("mismatch but no differing pixel found")
					}
				})
			}
		})
	}
}

// allImplCases returns the forced-feature sets available on the current CPU.
func allImplCases() []struct {
	name     string
	features Features
} {
	cases := []struct {
		name     string
		features Features
	}{
		{"generic", Features{Architecture: runtime.GOARCH, ForceGeneric: true}},
	}
	actual := DetectFeatures()
	if runtime.GOARCH == "amd64" {
		cases = append(cases, struct {
			name     string
			features Features
		}{"sse2", Features{Architecture: "amd64", HasSSE2: true}})
		if actual.HasSSSE3 && actual.HasSSE41 {
			cases = append(cases, struct {
				name     string
				features Features
			}{"sse41", Features{Architecture: "amd64", HasSSE2: true, HasSSSE3: true, HasSSE41: true}})
		}
		if actual.HasAVX2 {
			cases = append(cases, struct {
				name     string
				features Features
			}{"avx2", Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true}})
		}
	}
	if runtime.GOARCH == "arm64" {
		cases = append(cases, struct {
			name     string
			features Features
		}{"neon", Features{Architecture: "arm64", HasNEON: true}})
	}
	return cases
}

// TestPremultiplyRGBAComprehensive validates bit-identical output for
// PremultiplyRGBA across all forced implementation paths.
func TestPremultiplyRGBAComprehensive(t *testing.T) {
	t.Cleanup(ResetDetection)

	type scenario struct {
		name string
		buf  []byte // initial pixel data (R,G,B,A order)
	}

	mkBuf := func(pixels [][4]byte) []byte {
		b := make([]byte, len(pixels)*4)
		for i, p := range pixels {
			b[i*4+0] = p[0]
			b[i*4+1] = p[1]
			b[i*4+2] = p[2]
			b[i*4+3] = p[3]
		}
		return b
	}

	// Build a varied N-pixel buffer for stress tests.
	mkVary := func(n int) []byte {
		b := make([]byte, n*4)
		for i := 0; i < n; i++ {
			b[i*4+0] = byte(i * 97 % 256)
			b[i*4+1] = byte(i * 131 % 256)
			b[i*4+2] = byte(i * 173 % 256)
			b[i*4+3] = byte(i * 53 % 256) // varying alpha
		}
		return b
	}

	scenarios := []scenario{
		// Zero alpha: RGB must become 0.
		{"zero_alpha", mkBuf([][4]byte{{200, 100, 50, 0}, {128, 64, 32, 0}})},
		// Full alpha (255): channels unchanged.
		{"full_alpha", mkBuf([][4]byte{{200, 100, 50, 255}, {128, 64, 32, 255}})},
		// Half alpha.
		{"alpha_128", mkBuf([][4]byte{{200, 100, 50, 128}, {255, 128, 0, 128}})},
		// Single pixel, partial alpha.
		{"single_partial", mkBuf([][4]byte{{180, 90, 45, 64}})},
		// Two pixels (loop boundary for 2-pixel inner loop).
		{"two_pixels", mkBuf([][4]byte{{100, 150, 200, 100}, {50, 80, 120, 200}})},
		// Three pixels (two + one tail).
		{"three_pixels", mkBuf([][4]byte{{100, 150, 200, 100}, {50, 80, 120, 200}, {255, 128, 64, 1}})},
		// Five pixels (one 4-pixel SIMD iteration + one tail).
		{"five_pixels", mkVary(5)},
		// Eight pixels (exact two 4-pixel SIMD iterations).
		{"eight_pixels", mkVary(8)},
		// Twelve pixels (three 4-pixel iterations).
		{"twelve_pixels", mkVary(12)},
		// Thirteen pixels (three SIMD + one tail).
		{"thirteen_pixels", mkVary(13)},
		// Large span.
		{"large_64", mkVary(64)},
		// Odd large count.
		{"odd_31", mkVary(31)},
		// All pixels opaque.
		{"all_opaque_16", func() []byte {
			b := mkVary(16)
			for i := 0; i < 16; i++ {
				b[i*4+3] = 255
			}
			return b
		}()},
		// Row of transparent pixels (zero-alpha guard).
		{"all_transparent_8", mkBuf(func() [][4]byte {
			p := make([][4]byte, 8)
			for i := range p {
				p[i] = [4]byte{200, 100, 50, 0}
			}
			return p
		}())},
		// Mixed opaque + transparent pixels.
		{"mixed_opaque_transparent", mkBuf([][4]byte{
			{200, 100, 50, 255},
			{200, 100, 50, 0},
			{100, 200, 150, 128},
			{255, 255, 255, 0},
			{0, 0, 0, 255},
		})},
		// Boundary values: alpha = 1 and alpha = 254.
		{"alpha_boundary", mkBuf([][4]byte{
			{255, 255, 255, 1},
			{255, 255, 255, 254},
			{128, 128, 128, 1},
			{128, 128, 128, 254},
		})},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			for _, impl := range allImplCases() {
				t.Run(impl.name, func(t *testing.T) {
					ResetDetection()
					SetForcedFeatures(impl.features)

					count := len(sc.buf) / 4
					got := append([]byte(nil), sc.buf...)
					want := append([]byte(nil), sc.buf...)

					PremultiplyRGBA(got, count)
					premultiplyRGBAGeneric(want, count)

					for i := 0; i < count; i++ {
						p := i * 4
						if got[p] != want[p] || got[p+1] != want[p+1] || got[p+2] != want[p+2] || got[p+3] != want[p+3] {
							t.Fatalf("pixel %d (alpha=%d): got=[%d,%d,%d,%d] want=[%d,%d,%d,%d]",
								i, sc.buf[p+3],
								got[p], got[p+1], got[p+2], got[p+3],
								want[p], want[p+1], want[p+2], want[p+3])
						}
					}
				})
			}
		})
	}
}

// TestDemultiplyRGBAComprehensive validates bit-identical output for
// DemultiplyRGBA across all forced implementation paths.
func TestDemultiplyRGBAComprehensive(t *testing.T) {
	t.Cleanup(ResetDetection)

	type scenario struct {
		name string
		buf  []byte
	}

	mkBuf := func(pixels [][4]byte) []byte {
		b := make([]byte, len(pixels)*4)
		for i, p := range pixels {
			b[i*4+0] = p[0]
			b[i*4+1] = p[1]
			b[i*4+2] = p[2]
			b[i*4+3] = p[3]
		}
		return b
	}

	// Build premultiplied pixels: R/G/B are already ≤ alpha.
	mkPremulVary := func(n int) []byte {
		b := make([]byte, n*4)
		for i := 0; i < n; i++ {
			a := byte(i*53%255 + 1) // alpha 1..255
			r := byte(uint32(i*97%256) * uint32(a) / 255)
			g := byte(uint32(i*131%256) * uint32(a) / 255)
			bv := byte(uint32(i*173%256) * uint32(a) / 255)
			b[i*4+0] = r
			b[i*4+1] = g
			b[i*4+2] = bv
			b[i*4+3] = a
		}
		return b
	}

	scenarios := []scenario{
		// Zero alpha: RGB must become 0.
		{"zero_alpha", mkBuf([][4]byte{{0, 0, 0, 0}, {0, 0, 0, 0}})},
		// Full alpha: channels unchanged.
		{"full_alpha", mkBuf([][4]byte{{200, 100, 50, 255}, {128, 64, 32, 255}})},
		// Half alpha.
		{"alpha_128", mkBuf([][4]byte{{100, 50, 25, 128}})},
		// Single partial pixel.
		{"single_partial", mkBuf([][4]byte{{90, 45, 22, 64}})},
		// Two pixels.
		{"two_pixels", mkBuf([][4]byte{{50, 75, 100, 100}, {25, 40, 60, 200}})},
		// Five pixels (SIMD + tail).
		{"five_pixels", mkPremulVary(5)},
		// Eight pixels.
		{"eight_pixels", mkPremulVary(8)},
		// Large span.
		{"large_64", mkPremulVary(64)},
		// Mixed opaque + zero alpha.
		{"mixed_zero_opaque", mkBuf([][4]byte{
			{200, 100, 50, 255},
			{0, 0, 0, 0},
			{100, 50, 25, 128},
			{0, 0, 0, 0},
		})},
		// Boundary: alpha = 1.
		{"alpha_1", mkBuf([][4]byte{{1, 0, 0, 1}, {0, 1, 0, 1}, {0, 0, 1, 1}})},
		// Boundary: alpha = 254.
		{"alpha_254", mkBuf([][4]byte{{200, 100, 50, 254}})},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			for _, impl := range allImplCases() {
				t.Run(impl.name, func(t *testing.T) {
					ResetDetection()
					SetForcedFeatures(impl.features)

					count := len(sc.buf) / 4
					got := append([]byte(nil), sc.buf...)
					want := append([]byte(nil), sc.buf...)

					DemultiplyRGBA(got, count)
					demultiplyRGBAGeneric(want, count)

					for i := 0; i < count; i++ {
						p := i * 4
						if got[p] != want[p] || got[p+1] != want[p+1] || got[p+2] != want[p+2] || got[p+3] != want[p+3] {
							t.Fatalf("pixel %d (alpha=%d): got=[%d,%d,%d,%d] want=[%d,%d,%d,%d]",
								i, sc.buf[p+3],
								got[p], got[p+1], got[p+2], got[p+3],
								want[p], want[p+1], want[p+2], want[p+3])
						}
					}
				})
			}
		})
	}
}

// TestPremultiplyRoundtrip verifies that premultiply followed by demultiply
// recovers the original channels within the theoretical precision bound.
// The max round-trip error for a channel x with alpha a is ceil(255/(2*a)),
// since premultiplication maps ~(255/a) distinct input values to the same
// premultiplied byte. Alpha channel must be preserved exactly.
func TestPremultiplyRoundtrip(t *testing.T) {
	t.Cleanup(ResetDetection)

	for _, impl := range allImplCases() {
		t.Run(impl.name, func(t *testing.T) {
			ResetDetection()
			SetForcedFeatures(impl.features)

			const n = 64
			orig := make([]byte, n*4)
			for i := 0; i < n; i++ {
				a := byte(i*53%255 + 1) // alpha 1..255 (never zero)
				orig[i*4+0] = byte(i * 97 % 256)
				orig[i*4+1] = byte(i * 131 % 256)
				orig[i*4+2] = byte(i * 173 % 256)
				orig[i*4+3] = a
			}

			buf := append([]byte(nil), orig...)
			PremultiplyRGBA(buf, n)
			DemultiplyRGBA(buf, n)

			for i := 0; i < n; i++ {
				p := i * 4
				a := int(orig[p+3])
				// Theoretical maximum round-trip error: ceil(255 / (2*a))
				maxErr := (255 + 2*a - 1) / (2 * a)
				if maxErr < 1 {
					maxErr = 1
				}
				for ch := 0; ch < 3; ch++ {
					diff := int(buf[p+ch]) - int(orig[p+ch])
					if diff < -maxErr || diff > maxErr {
						t.Fatalf("pixel %d channel %d alpha=%d: roundtrip got %d want %d (diff %d, max allowed %d)",
							i, ch, a, buf[p+ch], orig[p+ch], diff, maxErr)
					}
				}
				if buf[p+3] != orig[p+3] {
					t.Fatalf("pixel %d: alpha changed from %d to %d", i, orig[p+3], buf[p+3])
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 8.2 — Composite blend mode tests
// ─────────────────────────────────────────────────────────────────────────────

// refCompSrcOver is a float64 reference for Porter-Duff SrcOver on premultiplied dst.
// (r,g,b,a) is straight-alpha src; cv is per-pixel coverage.
func refCompSrcOver(dca [4]byte, r, g, b, a, cv byte) [4]byte {
	sa := float64(rgba8Multiply(a, cv)) / 255.0
	scar := (float64(r) / 255.0) * sa
	scag := (float64(g) / 255.0) * sa
	scab := (float64(b) / 255.0) * sa
	dc := [4]float64{float64(dca[0]) / 255, float64(dca[1]) / 255, float64(dca[2]) / 255, float64(dca[3]) / 255}
	out := [4]float64{
		scar + dc[0]*(1-sa),
		scag + dc[1]*(1-sa),
		scab + dc[2]*(1-sa),
		sa + dc[3]*(1-sa),
	}
	var res [4]byte
	for i, v := range out {
		if v < 0 {
			v = 0
		} else if v > 1 {
			v = 1
		}
		res[i] = byte(v*255 + 0.5)
	}
	return res
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestCompSrcOverHspanRGBAComprehensive verifies CompSrcOverHspanRGBA across all
// SIMD implementation tiers for a range of pixel counts and coverage patterns.
func TestCompSrcOverHspanRGBAComprehensive(t *testing.T) {
	t.Cleanup(ResetDetection)

	type scenario struct {
		name        string
		dst         []byte
		r, g, b, a  byte
		covers      []byte
	}

	mkDst := func(pixels [][4]byte) []byte {
		buf := make([]byte, len(pixels)*4)
		for i, p := range pixels {
			copy(buf[i*4:], p[:])
		}
		return buf
	}

	scenarios := []scenario{
		{
			name: "transparent_src",
			dst:  mkDst([][4]byte{{128, 64, 32, 200}}),
			r: 255, g: 0, b: 0, a: 0,
		},
		{
			name: "opaque_src_over_transparent_dst",
			dst:  mkDst([][4]byte{{0, 0, 0, 0}}),
			r: 200, g: 100, b: 50, a: 255,
		},
		{
			name: "opaque_src_over_opaque_dst",
			dst:  mkDst([][4]byte{{80, 80, 80, 255}}),
			r: 200, g: 100, b: 50, a: 255,
		},
		{
			name: "half_alpha_2px",
			dst:  mkDst([][4]byte{{100, 150, 200, 200}, {50, 50, 50, 128}}),
			r: 200, g: 50, b: 100, a: 128,
		},
		{
			name: "4_pixels_simd_boundary",
			dst:  mkDst([][4]byte{{10, 20, 30, 40}, {50, 60, 70, 80}, {90, 100, 110, 120}, {130, 140, 150, 160}}),
			r: 200, g: 100, b: 50, a: 180,
		},
		{
			name: "5_pixels_tail",
			dst:  mkDst([][4]byte{{10, 20, 30, 40}, {50, 60, 70, 80}, {90, 100, 110, 120}, {130, 140, 150, 160}, {170, 180, 190, 200}}),
			r: 100, g: 150, b: 200, a: 200,
		},
		{
			name:   "variable_covers",
			dst:    mkDst([][4]byte{{100, 100, 100, 255}, {100, 100, 100, 255}, {100, 100, 100, 255}}),
			r: 200, g: 50, b: 50, a: 200,
			covers: []byte{0, 128, 255},
		},
		{
			name:   "zero_cover_skips_pixel",
			dst:    mkDst([][4]byte{{200, 200, 200, 255}}),
			r: 0, g: 0, b: 0, a: 255,
			covers: []byte{0},
		},
		{
			name: "large_span_64px",
			dst: func() []byte {
				b := make([]byte, 64*4)
				for i := range 64 {
					b[i*4+0] = byte(i * 3)
					b[i*4+1] = byte(i * 7)
					b[i*4+2] = byte(i * 11)
					b[i*4+3] = byte(i * 4)
				}
				return b
			}(),
			r: 180, g: 90, b: 45, a: 200,
		},
		{
			name: "odd_count_13px",
			dst: func() []byte {
				b := make([]byte, 13*4)
				for i := range 13 {
					b[i*4+0] = byte(i * 17 % 256)
					b[i*4+1] = byte(i * 31 % 256)
					b[i*4+2] = byte(i * 53 % 256)
					b[i*4+3] = byte(i * 7 % 256)
				}
				return b
			}(),
			r: 200, g: 100, b: 150, a: 220,
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			n := len(sc.dst) / 4
			// Build float64 reference.
			ref := append([]byte(nil), sc.dst...)
			for i := range n {
				cv := byte(255)
				if sc.covers != nil && i < len(sc.covers) {
					cv = sc.covers[i]
				}
				d := [4]byte{ref[i*4], ref[i*4+1], ref[i*4+2], ref[i*4+3]}
				out := refCompSrcOver(d, sc.r, sc.g, sc.b, sc.a, cv)
				copy(ref[i*4:], out[:])
			}

			for _, impl := range allImplCases() {
				t.Run(impl.name, func(t *testing.T) {
					ResetDetection()
					SetForcedFeatures(impl.features)

					dst := append([]byte(nil), sc.dst...)
					CompSrcOverHspanRGBA(dst, sc.covers, sc.r, sc.g, sc.b, sc.a, n)

					for i := range n {
						p := i * 4
						for ch := range 4 {
							got := int(dst[p+ch])
							want := int(ref[p+ch])
							if absInt(got-want) > 1 {
								t.Fatalf("pixel %d ch %d: got %d, want %d (±1 allowed)", i, ch, got, want)
							}
						}
					}
				})
			}
		})
	}
}

// TestCompOtherOpsGeneric verifies DstOver, SrcIn, DstIn, SrcOut, DstOut, Xor
// using float64 reference values.
func TestCompOtherOpsGeneric(t *testing.T) {
	t.Parallel()

	toF := func(v byte) float64 { return float64(v) / 255 }
	from8 := func(v float64) byte {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 255
		}
		return byte(v*255 + 0.5)
	}

	type testOp struct {
		name string
		fn   func(dst, covers []byte, r, g, b, a uint8, count int)
		ref  func(d [4]byte, r, g, b, a byte) [4]byte
	}

	ops := []testOp{
		{
			name: "DstOver",
			fn:   CompDstOverHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				da, sa := toF(d[3]), toF(a)
				return [4]byte{
					from8(toF(d[0]) + toF(r)*sa*(1-da)),
					from8(toF(d[1]) + toF(g)*sa*(1-da)),
					from8(toF(d[2]) + toF(b)*sa*(1-da)),
					from8(da + sa*(1-da)),
				}
			},
		},
		{
			name: "SrcIn",
			fn:   CompSrcInHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				da, sa := toF(d[3]), toF(a)
				return [4]byte{
					from8(toF(r) * sa * da),
					from8(toF(g) * sa * da),
					from8(toF(b) * sa * da),
					from8(sa * da),
				}
			},
		},
		{
			name: "DstIn",
			fn:   CompDstInHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				sa := toF(a)
				return [4]byte{
					from8(toF(d[0]) * sa),
					from8(toF(d[1]) * sa),
					from8(toF(d[2]) * sa),
					from8(toF(d[3]) * sa),
				}
			},
		},
		{
			name: "SrcOut",
			fn:   CompSrcOutHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				da, sa := toF(d[3]), toF(a)
				return [4]byte{
					from8(toF(r) * sa * (1 - da)),
					from8(toF(g) * sa * (1 - da)),
					from8(toF(b) * sa * (1 - da)),
					from8(sa * (1 - da)),
				}
			},
		},
		{
			name: "DstOut",
			fn:   CompDstOutHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				sa := toF(a)
				return [4]byte{
					from8(toF(d[0]) * (1 - sa)),
					from8(toF(d[1]) * (1 - sa)),
					from8(toF(d[2]) * (1 - sa)),
					from8(toF(d[3]) * (1 - sa)),
				}
			},
		},
		{
			name: "Xor",
			fn:   CompXorHspanRGBA,
			ref: func(d [4]byte, r, g, b, a byte) [4]byte {
				da, sa := toF(d[3]), toF(a)
				scar, scag, scab := toF(r)*sa, toF(g)*sa, toF(b)*sa
				return [4]byte{
					from8(scar*(1-da) + toF(d[0])*(1-sa)),
					from8(scag*(1-da) + toF(d[1])*(1-sa)),
					from8(scab*(1-da) + toF(d[2])*(1-sa)),
					from8(sa + da - 2*sa*da),
				}
			},
		},
	}

	dstPixels := [][4]byte{
		{0, 0, 0, 0},
		{255, 255, 255, 255},
		{100, 150, 200, 200},
		{50, 50, 50, 128},
	}
	srcs := [][4]byte{
		{200, 100, 50, 255},
		{200, 100, 50, 128},
		{200, 100, 50, 0},
	}

	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			for _, dp := range dstPixels {
				for _, sp := range srcs {
					dst := []byte{dp[0], dp[1], dp[2], dp[3]}
					want := op.ref(dp, sp[0], sp[1], sp[2], sp[3])
					op.fn(dst, nil, sp[0], sp[1], sp[2], sp[3], 1)
					for ch := range 4 {
						got := int(dst[ch])
						wantV := int(want[ch])
						if absInt(got-wantV) > 1 {
							t.Errorf("dst=%v src=%v ch=%d: got %d, want %d",
								dp, sp, ch, got, wantV)
						}
					}
				}
			}
		})
	}
}

// TestCompClearHspanRGBA verifies that Clear zeroes all pixels in the span.
func TestCompClearHspanRGBA(t *testing.T) {
	t.Parallel()

	dst := []byte{255, 128, 64, 200, 100, 50, 25, 180}
	CompClearHspanRGBA(dst, 2)
	for i, v := range dst {
		if v != 0 {
			t.Fatalf("byte %d not zeroed: got %d", i, v)
		}
	}

	// Partial clear — second pixel must be untouched.
	dst2 := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	CompClearHspanRGBA(dst2, 1)
	for i := range 4 {
		if dst2[i] != 0 {
			t.Fatalf("pixel 0 byte %d not cleared: got %d", i, dst2[i])
		}
	}
	if dst2[4] != 5 {
		t.Fatalf("second pixel modified: dst2[4]=%d want 5", dst2[4])
	}
}
