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
