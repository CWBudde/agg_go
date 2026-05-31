package agg2d

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestNewAgg2DFloatDefaults(t *testing.T) {
	a := NewAgg2DFloat()
	if a == nil {
		t.Fatal("NewAgg2DFloat() = nil")
	}
	if a.lineWidth != 1.0 {
		t.Errorf("lineWidth = %v, want 1.0", a.lineWidth)
	}
	if a.fillColor != White {
		t.Errorf("fillColor = %v, want White", a.fillColor)
	}
	if a.lineColor != Black {
		t.Errorf("lineColor = %v, want Black", a.lineColor)
	}
	if a.masterAlpha != 1.0 {
		t.Errorf("masterAlpha = %v, want 1.0", a.masterAlpha)
	}
	// Color-agnostic subsystems must be wired in the constructor.
	if a.path == nil || a.transform == nil || a.rasterizer == nil ||
		a.scanline == nil || a.convCurve == nil || a.convStroke == nil {
		t.Fatal("core subsystems not initialized")
	}
	if a.spanAllocator == nil || a.fillGradientLUT == nil || a.lineGradientLUT == nil {
		t.Fatal("float span allocator / gradient LUTs not initialized")
	}
	if len(a.fillGradientLUT) != 256 || len(a.lineGradientLUT) != 256 {
		t.Fatalf("gradient LUT lengths = %d/%d, want 256/256",
			len(a.fillGradientLUT), len(a.lineGradientLUT))
	}
	if a.fillLinearSpanGenerator == nil || a.lineLinearSpanGenerator == nil ||
		a.fillRadialSpanGenerator == nil || a.lineRadialSpanGenerator == nil {
		t.Fatal("float gradient span generators not initialized")
	}
}

func TestAgg2DFloatAttachWiresRenderers(t *testing.T) {
	a := NewAgg2DFloat()
	buf := make([]float32, 4*3*4)
	a.Attach(buf, 4, 3, 4*4*4)

	if a.pixfmt == nil || a.pixfmtPre == nil {
		t.Fatal("float pixfmt not wired after Attach")
	}
	if a.renBase == nil || a.renBasePre == nil {
		t.Fatal("float base renderers not wired after Attach")
	}
	if a.pixfmt.Width() != 4 || a.pixfmt.Height() != 3 {
		t.Fatalf("pixfmt dims = %dx%d, want 4x3", a.pixfmt.Width(), a.pixfmt.Height())
	}
}

func TestAgg2DFloatClearAll(t *testing.T) {
	a := NewAgg2DFloat()
	buf := make([]float32, 2*2*4)
	a.Attach(buf, 2, 2, 2*4*4)

	a.ClearAll(NewColor(255, 128, 0, 255))

	// Buffer holds straight RGBA float32; 8-bit color scales to [0,1].
	got := a.pixfmt.Pixel(0, 0)
	if got.R != 1.0 || !approxF(got.G, 128.0/255.0) || got.B != 0.0 || got.A != 1.0 {
		t.Fatalf("Pixel(0,0) = %+v, want ~{1,0.502,0,1}", got)
	}
	// last pixel too
	last := a.pixfmt.Pixel(1, 1)
	if last.R != 1.0 || last.A != 1.0 {
		t.Fatalf("Pixel(1,1) = %+v, want R=1,A=1", last)
	}
}

func TestAgg2DFloatAttachImage(t *testing.T) {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(5, 4)
	a.AttachImageFloat(img)
	if a.pixfmt == nil || a.pixfmt.Width() != 5 || a.pixfmt.Height() != 4 {
		t.Fatalf("AttachImageFloat did not wire 5x4 pixfmt: %v", a.pixfmt)
	}
	bounds := a.GetBounds()
	if bounds.X2 != 5 || bounds.Y2 != 4 {
		t.Fatalf("GetBounds = %+v, want X2=5,Y2=4", bounds)
	}
}

func approxF(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= 1e-6
}

// compile-time: ensure the float color type flows through the gradient LUT field.
var _ = func(a *Agg2DFloat) []color.RGBA32[color.Linear] { return a.fillGradientLUT }
