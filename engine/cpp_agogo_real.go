//go:build agogo && cgo && aggreal

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
		return fmt.Sprintf("the %q build is using a stub C++ backend %q", cppBuildTag, meta.BuildID)
	}
	if err := probeCPPNative(); err != nil {
		return fmt.Sprintf("the %q build tag is enabled, but the real AGG-backed probe failed: %v", cppBuildTag, err)
	}
	return cppBridgeNotImplementedReason()
}
