package agg

import "testing"

func feqf(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= 1e-6
}

func TestPublicAgg2DFloatSolidFill(t *testing.T) {
	img := NewImageFloat(20, 20)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(0, 0, 0, 0))

	a.FillColor(NewColor(255, 0, 0, 255))
	a.ResetPath()
	a.MoveTo(2, 2)
	a.LineTo(18, 2)
	a.LineTo(18, 18)
	a.LineTo(2, 18)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	r, g, b, al := img.GetPixelFloat(10, 10)
	if !feqf(r, 1.0) || !feqf(g, 0.0) || !feqf(b, 0.0) || !feqf(al, 1.0) {
		t.Fatalf("center = {%v,%v,%v,%v}, want opaque red", r, g, b, al)
	}
	_, _, _, ca := img.GetPixelFloat(0, 0)
	if ca != 0 {
		t.Fatalf("corner alpha = %v, want 0", ca)
	}
}

func TestPublicAgg2DFloatGradientAndImageOps(t *testing.T) {
	img := NewImageFloat(40, 10)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(0, 0, 0, 0))

	a.FillLinearGradient(0, 0, 40, 0, NewColor(255, 0, 0, 255), NewColor(0, 0, 255, 255), 1.0)
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(40, 0)
	a.LineTo(40, 10)
	a.LineTo(0, 10)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	lr, _, _, _ := img.GetPixelFloat(3, 5)
	_, _, rb, _ := img.GetPixelFloat(36, 5)
	if lr <= 0 || rb <= 0 {
		t.Fatalf("gradient endpoints not rendered: lr=%v rb=%v", lr, rb)
	}

	// CopyImage onto a fresh target.
	dst := NewImageFloat(60, 20)
	b := NewAgg2DFloat()
	b.AttachImage(dst)
	b.ClearAll(NewColor(0, 0, 0, 0))
	b.CopyImage(img, 5, 5)
	_, _, _, a2 := dst.GetPixelFloat(20, 8)
	if a2 <= 0 {
		t.Fatalf("copied region alpha = %v, want > 0", a2)
	}
}

func TestPublicAgg2DFloatBoundaryToRGBA(t *testing.T) {
	img := NewImageFloat(1, 1)
	img.SetPixelFloat(0, 0, 1.0, 0.5, 0.0, 0.5)
	rgba := img.ToRGBA()
	got := rgba.RGBAAt(0, 0)
	// Go image.RGBA is premultiplied: a=128, r=round(1*0.5*255)=128
	if got.A < 126 || got.A > 130 || got.R < 126 || got.R > 130 {
		t.Fatalf("ToRGBA pixel = %+v, want ~{128,64,0,128}", got)
	}
}

func TestPublicAgg2DFloatTransformAndFillMode(t *testing.T) {
	img := NewImageFloat(24, 24)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(0, 0, 0, 0))

	// Translate then fill a small rect; verify it moved.
	a.FillColor(NewColor(255, 0, 0, 255))
	a.Translate(10, 10)
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(6, 0)
	a.LineTo(6, 6)
	a.LineTo(0, 6)
	a.ClosePolygon()
	a.DrawPath(FillOnly)
	if _, _, _, al := img.GetPixelFloat(13, 13); al <= 0 {
		t.Fatalf("translated fill missing at (13,13): alpha=%v", al)
	}
	if _, _, _, al := img.GetPixelFloat(3, 3); al != 0 {
		t.Fatalf("untranslated location should be empty: alpha=%v", al)
	}

	// Fill-mode toggles are reachable from the public API.
	a.FillEvenOdd(true)
	if !a.GetFillEvenOdd() {
		t.Fatal("GetFillEvenOdd should be true after FillEvenOdd(true)")
	}
	a.NoFill()
	a.NoLine()
}

func TestPublicAgg2DFloatTransformImage(t *testing.T) {
	// Opaque source with a recognizable solid color.
	src := NewImageFloat(8, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src.SetPixelFloat(x, y, 0.0, 0.8, 0.2, 1.0)
		}
	}

	dst := NewImageFloat(40, 40)
	a := NewAgg2DFloat()
	a.AttachImage(dst)
	a.ClearAll(NewColor(0, 0, 0, 0))

	// Affine scale the 8x8 source into a 24x24 destination rectangle.
	if err := a.TransformImageSimple(src, 8, 8, 32, 32); err != nil {
		t.Fatalf("TransformImageSimple: %v", err)
	}
	if _, g, _, al := dst.GetPixelFloat(20, 20); g <= 0 || al <= 0 {
		t.Fatalf("transformed region missing at (20,20): g=%v a=%v", g, al)
	}
	if _, _, _, al := dst.GetPixelFloat(2, 2); al != 0 {
		t.Fatalf("outside transformed region should be empty: alpha=%v", al)
	}

	// Perspective quad maps the source to a non-affine quadrangle.
	dst2 := NewImageFloat(40, 40)
	b := NewAgg2DFloat()
	b.AttachImage(dst2)
	b.ClearAll(NewColor(0, 0, 0, 0))
	quad := [8]float64{6, 8, 34, 5, 32, 36, 9, 31}
	if err := b.TransformImageQuadSimple(src, quad); err != nil {
		t.Fatalf("TransformImageQuadSimple: %v", err)
	}
	if _, _, _, al := dst2.GetPixelFloat(20, 20); al <= 0 {
		t.Fatalf("perspective region missing at (20,20): alpha=%v", al)
	}
}

func TestPublicContextFloat(t *testing.T) {
	ctx := NewContextFloat(20, 20)
	ctx.Clear(NewColor(0, 0, 0, 0))
	ctx.SetColor(NewColor(0, 0, 255, 255))
	ctx.FillRectangle(2, 2, 16, 16)

	img := ctx.GetImage()
	_, _, bb, ba := img.GetPixelFloat(10, 10)
	if !feqf(bb, 1.0) || ba <= 0 {
		t.Fatalf("context fill center = blue? b=%v a=%v", bb, ba)
	}
	if ctx.Width() != 20 || ctx.Height() != 20 {
		t.Fatalf("ctx dims = %dx%d, want 20x20", ctx.Width(), ctx.Height())
	}
}

func TestPublicAgg2DFloatBlendMode(t *testing.T) {
	img := NewImageFloat(24, 24)
	a := NewAgg2DFloat()
	a.AttachImage(img)

	if a.GetBlendMode() != BlendAlpha {
		t.Fatalf("default blend mode = %v, want BlendAlpha", a.GetBlendMode())
	}

	// Opaque background; Multiply an opaque fill over it. Premultiplied == straight
	// for opaque content, so the result is the component-wise product.
	a.ClearAll(NewColor(255, 128, 64, 255)) // ~(1.0, 0.502, 0.251)
	a.SetBlendMode(BlendMultiply)
	if a.GetBlendMode() != BlendMultiply {
		t.Fatalf("blend mode = %v after set, want BlendMultiply", a.GetBlendMode())
	}
	a.FillColor(NewColor(128, 128, 128, 255)) // ~0.502
	a.ResetPath()
	a.MoveTo(4, 4)
	a.LineTo(20, 4)
	a.LineTo(20, 20)
	a.LineTo(4, 20)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	r, g, b, al := img.GetPixelFloat(12, 12)
	near := func(v, want float32) bool {
		d := v - want
		if d < 0 {
			d = -d
		}
		return d <= 0.01
	}
	// r = 1.0*0.502, g = 0.502*0.502, b = 0.251*0.502
	if !near(r, 0.502) || !near(g, 0.252) || !near(b, 0.126) || !near(al, 1.0) {
		t.Fatalf("multiply center = {%v,%v,%v,%v}, want ~{0.502,0.252,0.126,1}", r, g, b, al)
	}
}
