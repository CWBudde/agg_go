package agg

// Agg2D intentionally stays close to the original C++ AGG2D interface in
// `../agg-2.6/agg-src/agg2d/agg2d.h`.
//
// Go cannot overload methods, so a few naming adjustments are required:
// `Get...` methods expose C++ getter overloads, whole-image overloads use a
// `...Simple` suffix, three-color radial gradients use `...MultiStop`, and
// position-only radial gradient overloads use `...Pos`.
//
// The goal of this wrapper is API parity at the public boundary while keeping
// existing Go callers stable, so the original AGG2D documentation remains a
// useful reference.

// GetFillColor returns the current fill color.
// This is the Go getter form of the C++ `fillColor() const` overload.
func (a *Agg2D) GetFillColor() Color {
	c := a.impl.GetFillColor()
	return Color{R: c[0], G: c[1], B: c[2], A: c[3]}
}

// GetLineColor returns the current stroke color.
// This is the Go getter form of the C++ `lineColor() const` overload.
func (a *Agg2D) GetLineColor() Color {
	c := a.impl.GetLineColor()
	return Color{R: c[0], G: c[1], B: c[2], A: c[3]}
}

// ClearAllRGBA fills the entire attached buffer using explicit RGBA components.
// This mirrors the C++ `clearAll(r, g, b, a)` overload.
func (a *Agg2D) ClearAllRGBA(r, g, b, alpha uint8) {
	a.ClearAll(Color{R: r, G: g, B: b, A: alpha})
}

// ClearClipBoxRGBA fills the current clip box using explicit RGBA components.
// This mirrors the C++ `clearClipBox(r, g, b, a)` overload.
func (a *Agg2D) ClearClipBoxRGBA(r, g, b, alpha uint8) {
	a.ClearClipBox(Color{R: r, G: g, B: b, A: alpha})
}

// FillColorRGBA sets the fill color using explicit RGBA components.
// This mirrors the C++ `fillColor(r, g, b, a)` overload.
func (a *Agg2D) FillColorRGBA(r, g, b, alpha uint8) {
	a.FillColor(Color{R: r, G: g, B: b, A: alpha})
}

// LineColorRGBA sets the stroke color using explicit RGBA components.
// This mirrors the C++ `lineColor(r, g, b, a)` overload.
func (a *Agg2D) LineColorRGBA(r, g, b, alpha uint8) {
	a.LineColor(Color{R: r, G: g, B: b, A: alpha})
}

// ImageBlendColorRGBA sets the image blend color using explicit RGBA
// components. This mirrors the C++ `imageBlendColor(r, g, b, a)` overload.
func (a *Agg2D) ImageBlendColorRGBA(r, g, b, alpha uint8) {
	a.ImageBlendColor(Color{R: r, G: g, B: b, A: alpha})
}

// FillRadialGradientPos updates only the position and radius of the current
// fill radial gradient. This mirrors the C++ `fillRadialGradient(x, y, r)`
// overload.
func (a *Agg2D) FillRadialGradientPos(x, y, r float64) {
	a.impl.FillRadialGradientPos(x, y, r)
}

// LineRadialGradientPos updates only the position and radius of the current
// stroke radial gradient. This mirrors the C++ `lineRadialGradient(x, y, r)`
// overload.
func (a *Agg2D) LineRadialGradientPos(x, y, r float64) {
	a.impl.LineRadialGradientPos(x, y, r)
}

// Parallelogram multiplies the current transform with the parallelogram mapping
// used by the original AGG2D interface.
func (a *Agg2D) Parallelogram(x1, y1, x2, y2, x3, y3 float64) {
	a.impl.Parallelogram(x1, y1, x2, y2, x3, y3)
}
