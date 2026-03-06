package simd

import (
	"runtime"
	"testing"
)

func BenchmarkFillRGBA(b *testing.B) {
	sizes := []int{64, 256, 1024, 4096}

	for _, tc := range benchmarkCases() {
		b.Run(tc.name, func(b *testing.B) {
			for _, pixels := range sizes {
				b.Run(pixelBenchName(pixels), func(b *testing.B) {
					dst := make([]byte, pixels*4)
					ResetDetection()
					SetForcedFeatures(tc.features)
					b.Cleanup(ResetDetection)
					b.SetBytes(int64(len(dst)))
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						FillRGBA(dst, 1, 2, 3, 4, pixels)
					}
				})
			}
		})
	}
}

func BenchmarkBlendSolidHspanRGBA(b *testing.B) {
	sizes := []int{64, 256, 1024, 4096}

	for _, tc := range benchmarkCases() {
		b.Run(tc.name, func(b *testing.B) {
			for _, pixels := range sizes {
				b.Run(pixelBenchName(pixels), func(b *testing.B) {
					dst := make([]byte, pixels*4)
					covers := make([]byte, pixels)
					for i := range covers {
						switch {
						case i%8 < 3:
							covers[i] = 255
						case i%8 == 3:
							covers[i] = 128
						case i%8 == 4:
							covers[i] = 64
						default:
							covers[i] = 0
						}
					}

					ResetDetection()
					SetForcedFeatures(tc.features)
					b.Cleanup(ResetDetection)
					b.SetBytes(int64(len(dst) + len(covers)))
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						BlendSolidHspanRGBA(dst, covers, 20, 40, 60, 255, false)
					}
				})
			}
		})
	}
}

func benchmarkCases() []struct {
	name     string
	features Features
} {
	cases := []struct {
		name     string
		features Features
	}{
		{
			name:     "generic",
			features: Features{Architecture: runtime.GOARCH, ForceGeneric: true},
		},
	}

	switch runtime.GOARCH {
	case "amd64":
		cases = append(cases, struct {
			name     string
			features Features
		}{
			name:     "sse2",
			features: Features{Architecture: "amd64", HasSSE2: true},
		})
		if DetectFeatures().HasAVX2 {
			cases = append(cases, struct {
				name     string
				features Features
			}{
				name:     "avx2",
				features: Features{Architecture: "amd64", HasSSE2: true, HasAVX2: true},
			})
		}
	case "arm64":
		cases = append(cases, struct {
			name     string
			features Features
		}{
			name:     "neon",
			features: Features{Architecture: "arm64", HasNEON: true},
		})
	}

	return cases
}

func pixelBenchName(pixels int) string {
	return "pixels_" + itoa(pixels)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
