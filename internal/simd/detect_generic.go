//go:build !amd64 && !arm64

package simd

import "runtime"

func detectFeaturesImpl() Features {
	return Features{
		Architecture: runtime.GOARCH,
	}
}
