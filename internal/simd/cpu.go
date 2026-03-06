package simd

import (
	"sync"
)

// Features describes the CPU capabilities relevant to SIMD dispatch.
type Features struct {
	HasSSE    bool
	HasSSE2   bool
	HasSSE3   bool
	HasSSSE3  bool
	HasSSE41  bool
	HasAVX    bool
	HasAVX2   bool
	HasAVX512 bool
	HasNEON   bool

	ForceGeneric bool
	Architecture string
}

type implementation struct {
	name                string
	fillRGBA            func(dst []byte, r, g, b, a uint8, count int)
	blendSolidHspanRGBA func(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool)
}

var (
	detectMutex sync.Mutex
	detectOnce  sync.Once
	detected    Features

	forcedMutex    sync.RWMutex
	forcedFeatures *Features

	implMutex  sync.Mutex
	implOnce   sync.Once
	activeImpl implementation
)

// DetectFeatures returns the detected CPU features, cached across calls.
func DetectFeatures() Features {
	forcedMutex.RLock()
	forced := forcedFeatures
	forcedMutex.RUnlock()
	if forced != nil {
		return *forced
	}

	detectMutex.Lock()
	detectOnce.Do(func() {
		detected = detectFeaturesImpl()
	})
	features := detected
	detectMutex.Unlock()

	return features
}

// SetForcedFeatures overrides hardware detection for tests.
func SetForcedFeatures(f Features) {
	forcedMutex.Lock()
	forced := f
	forcedFeatures = &forced
	forcedMutex.Unlock()
	resetImplementation()
}

// ResetDetection clears forced features and cached detection/dispatch state.
func ResetDetection() {
	forcedMutex.Lock()
	forcedFeatures = nil
	forcedMutex.Unlock()

	detectMutex.Lock()
	detectOnce = sync.Once{}
	detected = Features{}
	detectMutex.Unlock()

	resetImplementation()
}

// ImplementationName returns the currently selected fill implementation name.
func ImplementationName() string {
	return currentImplementation().name
}

// FillRGBA fills count tightly-packed 4-byte pixels in dst with the provided byte pattern.
func FillRGBA(dst []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}

	maxCount := len(dst) / 4
	if count > maxCount {
		count = maxCount
	}
	if count <= 0 {
		return
	}

	currentImplementation().fillRGBA(dst, r, g, b, a, count)
}

func currentImplementation() implementation {
	implMutex.Lock()
	implOnce.Do(func() {
		activeImpl = selectImplementation(DetectFeatures())
	})
	impl := activeImpl
	implMutex.Unlock()
	return impl
}

func resetImplementation() {
	implMutex.Lock()
	implOnce = sync.Once{}
	activeImpl = implementation{}
	implMutex.Unlock()
}

func selectImplementation(features Features) implementation {
	return selectImplementationArch(features)
}

func genericImplementation() implementation {
	return implementation{
		name:                "generic",
		fillRGBA:            fillRGBAGeneric,
		blendSolidHspanRGBA: blendSolidHspanRGBAGeneric,
	}
}

func fillRGBAGeneric(dst []byte, r, g, b, a uint8, count int) {
	for i := 0; i < count; i++ {
		p := i * 4
		dst[p+0] = r
		dst[p+1] = g
		dst[p+2] = b
		dst[p+3] = a
	}
}

// BlendSolidHspanRGBA blends a solid color into tightly packed RGBA pixels.
// dst must contain one 4-byte pixel per cover entry.
func BlendSolidHspanRGBA(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	if len(covers) == 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if len(covers) > maxPixels {
		covers = covers[:maxPixels]
	}
	currentImplementation().blendSolidHspanRGBA(dst, covers, r, g, b, a, premulSrc)
}

func blendSolidHspanRGBAGeneric(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	for i, cv := range covers {
		if cv == 0 {
			continue
		}
		p := i * 4
		if a == 0 && !premulSrc {
			continue
		}
		if a == 255 && cv == 255 {
			dst[p+0] = r
			dst[p+1] = g
			dst[p+2] = b
			dst[p+3] = a
			continue
		}

		if premulSrc {
			rr, gg, bb, aa := r, g, b, a
			if cv != 255 {
				rr = rgba8Multiply(rr, cv)
				gg = rgba8Multiply(gg, cv)
				bb = rgba8Multiply(bb, cv)
				aa = rgba8Multiply(aa, cv)
			}
			if aa == 0 && rr == 0 && gg == 0 && bb == 0 {
				continue
			}
			dst[p+0] = rgba8Prelerp(dst[p+0], rr, aa)
			dst[p+1] = rgba8Prelerp(dst[p+1], gg, aa)
			dst[p+2] = rgba8Prelerp(dst[p+2], bb, aa)
			dst[p+3] = rgba8Prelerp(dst[p+3], aa, aa)
			continue
		}

		alpha := rgba8Multiply(a, cv)
		if alpha == 0 {
			continue
		}
		dst[p+0] = rgba8Lerp(dst[p+0], r, alpha)
		dst[p+1] = rgba8Lerp(dst[p+1], g, alpha)
		dst[p+2] = rgba8Lerp(dst[p+2], b, alpha)
		dst[p+3] = rgba8Prelerp(dst[p+3], alpha, alpha)
	}
}

func rgba8Multiply(a, b uint8) uint8 {
	t := uint32(a)*uint32(b) + 128
	return uint8(((t >> 8) + t) >> 8)
}

func rgba8Lerp(p, q, a uint8) uint8 {
	var greater int32
	if p > q {
		greater = 1
	}
	t := (int32(q)-int32(p))*int32(a) + 128 - greater
	return uint8(int32(p) + (((t >> 8) + t) >> 8))
}

func rgba8Prelerp(p, q, a uint8) uint8 {
	return p + q - rgba8Multiply(p, a)
}
