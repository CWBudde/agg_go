// Package engine provides an opt-in high-level facade for selecting a rendering
// engine at runtime.
//
// The default engine is the native Go port in the root agg package. A future
// C++ reference/performance engine can plug into the same facade without
// changing existing agg package callers.
package engine
