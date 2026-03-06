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
