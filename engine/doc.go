// Package engine provides an opt-in high-level facade for selecting a rendering
// engine at runtime.
//
// The default engine is the native Go port in the root agg package. A future
// C++ reference/performance engine can plug into the same facade without
// changing existing agg package callers. The optional in-repo C++ engine is
// reserved for builds using the "agogo" build tag; the current real AGG-backed
// development path additionally uses "aggreal" while dependency detection is
// still being migrated into this repo. Requesting the C++ engine from other
// builds returns a typed unavailable error instead of silently falling back.
package engine
