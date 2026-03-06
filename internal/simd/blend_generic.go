//go:build (!amd64 && !arm64) || purego

package simd

func selectImplementationArch(features Features) implementation {
	_ = features
	return genericImplementation()
}
