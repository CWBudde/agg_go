//go:build arm64

package simd

import (
	"runtime"

	"golang.org/x/sys/cpu"
)

// detectFeaturesImpl reads arm64 SIMD capabilities from x/sys/cpu.
func detectFeaturesImpl() Features {
	return Features{
		HasNEON:      cpu.ARM64.HasASIMD,
		Architecture: runtime.GOARCH,
	}
}
