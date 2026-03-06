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
		name   string
		r, g, b, a uint8
		premul bool
		covers []byte
		base   []byte // per-pixel dst, len must be len(covers)*4
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
		{"run_then_partial", 255, 128, 0, 255, false,
			append(mkCovers(8, 255), mkCovers(4, 100)...),
			mkBase(12, 64)},
		// Alternating 0/255 covers
		{"alternating", 200, 100, 50, 255, false,
			func() []byte {
				c := make([]byte, 16)
				for i := range c {
					if i%2 == 0 {
						c[i] = 255
					}
				}
				return c
			}(),
			mkBase(16, 128)},
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
