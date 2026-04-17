//go:build agogo && cgo

package engine

import (
	"fmt"

	"github.com/cwbudde/agg_go/internal/cppbridge"
)

func cppAvailable() bool {
	meta := cppbridge.CurrentMetadata()
	if meta.Stub {
		return false
	}
	return cppbridge.Probe() == nil
}

func cppUnavailableReason() string {
	meta := cppbridge.CurrentMetadata()
	if meta.Stub {
		return fmt.Sprintf("the %q build tag is enabled, but the in-repo C++ bridge build %q is a stub and is rejected", cppBuildTag, meta.BuildID)
	}
	if err := cppbridge.Probe(); err != nil {
		return fmt.Sprintf("the %q build tag is enabled, but the in-repo C++ bridge probe failed: %v", cppBuildTag, err)
	}
	return cppBridgeNotImplementedReason()
}
