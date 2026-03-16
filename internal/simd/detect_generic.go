//go:build !amd64 && !arm64

package simd

import "runtime"

// detectFeaturesImpl reports the current architecture for non-amd64/non-arm64
// builds, which use the generic implementation only.
func detectFeaturesImpl() Features {
	return Features{
		Architecture: runtime.GOARCH,
	}
}
