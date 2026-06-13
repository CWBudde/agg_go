// Package scene defines a backend-neutral corpus of render scenes that can be
// drawn identically through any engine.Context implementation. The same scenes
// drive the cross-backend conformance tests (tests/conformance), the
// engine-compare render tool (cmd/engine-compare), and the corpus benchmark.
//
// The package builds in the default (no-tag) build and never references a
// concrete backend directly. Whether the optional C++ engine is reachable is
// decided at runtime via engine.Available(); scenes only ever go through the
// engine.Context / engine.Image facade. Each scene declares the facade
// capabilities it exercises so callers can skip it on a backend that does not
// support all of them (see engine.Supports).
package scene

import (
	"errors"
	"strings"

	"github.com/cwbudde/agg_go/engine"
)

// ErrAssetUnavailable signals that a scene could not run because a required
// runtime asset (currently only a usable font for the text scene) was not
// found. Callers should treat this as a skip, not a failure.
var ErrAssetUnavailable = errors.New("scene asset unavailable")

// Scene is a backend-neutral description of a render operation sequence.
//
// Draw must be deterministic: no time, no randomness, no goroutines, and no
// reliance on map iteration order. It receives a context already constructed
// for some engine and an Assets bundle owned by that same engine. Draw returns
// an error so capability or engine-mismatch failures surface instead of
// panicking; it returns ErrAssetUnavailable when a required asset is missing.
type Scene struct {
	Name   string
	Width  int
	Height int
	// Caps lists every facade capability the Draw closure exercises. A scene is
	// skipped on any engine that does not report all of these via
	// engine.Supports. The always-available solid_style/path capabilities may be
	// omitted, but listing them is harmless.
	Caps []engine.Capability
	Draw func(ctx engine.Context, assets *Assets) error
}

// SupportedBy reports whether the given engine kind supports every capability
// the scene requires.
func (s Scene) SupportedBy(kind engine.Kind) bool {
	for _, c := range s.Caps {
		if !engine.Supports(kind, c) {
			return false
		}
	}
	return true
}

// All returns a copy of the corpus in a stable, deterministic order.
func All() []Scene {
	out := make([]Scene, len(corpus))
	copy(out, corpus)
	return out
}

// ByName returns the named scene and true, or a zero Scene and false.
func ByName(name string) (Scene, bool) {
	for _, s := range corpus {
		if s.Name == name {
			return s, true
		}
	}
	return Scene{}, false
}

// Filter returns the scenes whose Name contains substr ("" returns all).
func Filter(substr string) []Scene {
	if substr == "" {
		return All()
	}
	var out []Scene
	for _, s := range corpus {
		if strings.Contains(s.Name, substr) {
			out = append(out, s)
		}
	}
	return out
}
