package agg2d

import "testing"

// Attach must reset the rasterizer gamma (matching 8-bit Attach). Otherwise a
// master alpha set before a re-Attach leaks into subsequent rendering as a
// stale coverage scale.
func TestAgg2DFloatAttachResetsRasterizerGamma(t *testing.T) {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(10, 10)
	a.AttachImageFloat(img)

	// Set a half master alpha; this installs a coverage-scaling gamma.
	a.SetMasterAlpha(0.5)

	// Re-attach (Attach resets masterAlpha to 1.0 and must reset gamma too).
	a.AttachImageFloat(img)
	a.ClearAll(NewColor(0, 0, 0, 0))

	// Fill an opaque interior; full coverage + masterAlpha 1.0 -> alpha 1.0.
	a.FillColor(NewColor(255, 0, 0, 255))
	a.ResetPath()
	a.MoveTo(1, 1)
	a.LineTo(9, 1)
	a.LineTo(9, 9)
	a.LineTo(1, 9)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	center := img.GetPixel(5, 5)
	if !approxF(center.A, 1.0) {
		t.Fatalf("interior alpha = %v, want 1.0 (stale master-alpha gamma leaked across Attach)", center.A)
	}
}
