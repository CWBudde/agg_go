//go:build agogo && cgo

package engine

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#cgo CPPFLAGS: -I${SRCDIR}
#include "cpp_native.h"
*/
import "C"

import "fmt"

type cppNativeMetadata struct {
	BuildID string
	Stub    bool
}

func currentCPPNativeMetadata() cppNativeMetadata {
	return cppNativeMetadata{
		BuildID: C.GoString(C.agg_go_cpp_bridge_build_id()),
		Stub:    C.agg_go_cpp_bridge_is_stub() != 0,
	}
}

func probeCPPNative() error {
	if code := int(C.agg_go_cpp_bridge_probe()); code == 0 {
		return nil
	}
	msg := C.GoString(C.agg_go_cpp_bridge_last_error())
	if msg == "" {
		msg = "unknown native bridge failure"
	}
	return fmt.Errorf("%s", msg)
}

func cppNativeLastError() string {
	if msg := C.GoString(C.agg_go_cpp_bridge_last_error()); msg != "" {
		return msg
	}
	return "unknown native bridge failure"
}
