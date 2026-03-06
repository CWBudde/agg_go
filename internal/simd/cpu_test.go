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

	switch runtime.GOARCH {
	case "amd64":
		SetForcedFeatures(Features{Architecture: "amd64", HasSSE2: true})
		if got := ImplementationName(); got != "sse2" {
			t.Fatalf("ImplementationName() = %q, want sse2", got)
		}

		SetForcedFeatures(Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true})
		if got := ImplementationName(); got != "avx2" {
			t.Fatalf("ImplementationName() = %q, want avx2", got)
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
			struct {
				name     string
				features Features
				wantImpl string
			}{
				name:     "avx2",
				features: Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true},
				wantImpl: "avx2",
			},
		)
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
