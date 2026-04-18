//go:build agogo && cgo

package engine

import "fmt"

func cppAvailable() bool {
	meta := currentCPPNativeMetadata()
	if meta.Stub {
		return false
	}
	return probeCPPNative() == nil
}

func cppUnavailableReason() string {
	meta := currentCPPNativeMetadata()
	if meta.Stub {
		return fmt.Sprintf("the %q build tag is enabled, but the in-repo C++ backend build %q is a stub and is rejected", cppBuildTag, meta.BuildID)
	}
	if err := probeCPPNative(); err != nil {
		return fmt.Sprintf("the %q build tag is enabled, but the in-repo C++ backend probe failed: %v", cppBuildTag, err)
	}
	return cppBridgeNotImplementedReason()
}
