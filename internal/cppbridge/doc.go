// Package cppbridge provides the in-repo native bridge boundary for the
// optional `agogo` C++ engine.
//
// The bridge is intentionally kept internal while the higher-level `engine`
// adapter is still under construction. This package owns the native build
// boundary and bridge metadata so the repository no longer depends on the
// external AGoGo module at runtime.
package cppbridge
