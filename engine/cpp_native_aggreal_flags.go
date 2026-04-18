//go:build agogo && cgo && aggreal

package engine

/*
#cgo CPPFLAGS: -DAGG_GO_CPP_REAL -I/usr/include/agg2
#cgo pkg-config: freetype2
#cgo LDFLAGS: -L/usr/lib/x86_64-linux-gnu -lagg -laggfontfreetype -lm
*/
import "C"
