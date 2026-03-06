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
	name     string
	fillRGBA func(dst []byte, r, g, b, a uint8, count int)
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
		name:     "generic",
		fillRGBA: fillRGBAGeneric,
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
