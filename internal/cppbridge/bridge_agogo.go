//go:build agogo && cgo

package cppbridge

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#cgo CPPFLAGS: -I${SRCDIR}
#include "bridge.h"
*/
import "C"

import "fmt"

// Metadata describes the compiled in-repo native bridge.
type Metadata struct {
	BuildID string
	Stub    bool
}

// CurrentMetadata returns metadata for the compiled bridge in the current
// `agogo`+cgo build.
func CurrentMetadata() Metadata {
	return Metadata{
		BuildID: C.GoString(C.agg_go_cpp_bridge_build_id()),
		Stub:    C.agg_go_cpp_bridge_is_stub() != 0,
	}
}

// Probe validates the bridge and returns an error when the compiled bridge is
// not usable by the higher-level engine adapter.
func Probe() error {
	if code := int(C.agg_go_cpp_bridge_probe()); code == 0 {
		return nil
	}
	msg := C.GoString(C.agg_go_cpp_bridge_last_error())
	if msg == "" {
		msg = "unknown native bridge failure"
	}
	return fmt.Errorf("%s", msg)
}
