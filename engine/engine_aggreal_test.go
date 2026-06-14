//go:build agogo && cgo && aggreal

package engine_test

import (
	"os"
	"testing"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/engine"
)

func TestAvailableIncludesCPPWithAggReal(t *testing.T) {
	available := engine.Available()
	if len(available) < 2 {
		t.Fatalf("expected C++ engine to be advertised in aggreal build, got %v", available)
	}
	found := false
	for _, kind := range available {
		if kind == engine.CPP {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected available engines to include %q, got %v", engine.CPP, available)
	}
}

func TestCapabilitiesCPPExposeCurrentRealSubset(t *testing.T) {
	caps, err := engine.Capabilities(engine.CPP)
	if err != nil {
		t.Fatalf("Capabilities(CPP) error = %v", err)
	}
	for _, want := range []engine.Capability{
		engine.CapabilitySolidStyle,
		engine.CapabilityPath,
		engine.CapabilityTransforms,
		engine.CapabilityClipBox,
		engine.CapabilityCompositing,
		engine.CapabilityImageDraw,
		engine.CapabilityImageExport,
		engine.CapabilityGradients,
		engine.CapabilityText,
		engine.CapabilityDashedStroke,
	} {
		found := false
		for _, cap := range caps {
			if cap == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected cpp capability set to include %q, got %v", want, caps)
		}
	}
	if !engine.Supports(engine.CPP, engine.CapabilityText) {
		t.Fatal("expected current real C++ subset to report text capability")
	}
}

func TestNewContextCPPWorksWithAggReal(t *testing.T) {
	ctx, err := engine.NewContext(16, 16, engine.Config{Kind: engine.CPP})
	if err != nil {
		t.Fatalf("NewContext(CPP) error = %v", err)
	}

	ctx.Clear(agg.White)
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(2, 2, 10, 10)

	got := ctx.GetImage().ToGoImage().RGBAAt(6, 6)
	if got.R < 200 || got.G > 40 || got.B > 40 || got.A != 255 {
		t.Fatalf("unexpected rendered color at center: %+v", got)
	}
}

func TestCPPStyleAndTransformReadbackWithAggReal(t *testing.T) {
	ctx, err := engine.NewContext(16, 16, engine.Config{Kind: engine.CPP})
	if err != nil {
		t.Fatalf("NewContext(CPP) error = %v", err)
	}

	fill := agg.NewColor(10, 20, 30, 200)
	stroke := agg.NewColor(40, 50, 60, 255)
	ctx.SetFillColor(fill)
	ctx.SetStrokeColor(stroke)
	ctx.SetLineWidth(2.5)
	ctx.SetLineCap(agg.CapRound)
	ctx.SetLineJoin(agg.JoinBevel)

	if got := ctx.GetFillColor(); got != fill {
		t.Errorf("GetFillColor() = %+v, want %+v", got, fill)
	}
	if got := ctx.GetStrokeColor(); got != stroke {
		t.Errorf("GetStrokeColor() = %+v, want %+v", got, stroke)
	}
	if got := ctx.GetLineWidth(); got != 2.5 {
		t.Errorf("GetLineWidth() = %v, want 2.5", got)
	}
	if got := ctx.GetLineCap(); got != agg.CapRound {
		t.Errorf("GetLineCap() = %v, want %v", got, agg.CapRound)
	}
	if got := ctx.GetLineJoin(); got != agg.JoinBevel {
		t.Errorf("GetLineJoin() = %v, want %v", got, agg.JoinBevel)
	}

	// Fresh transform is identity; readback comes from the native matrix.
	if got := ctx.GetTransform(); !got.IsIdentity() {
		t.Fatalf("fresh GetTransform() = %v, want identity", got.AffineMatrix)
	}
	// Translation of identity is convention-independent.
	ctx.Translate(5, 7)
	if got := ctx.GetTransform().AffineMatrix; got != [6]float64{1, 0, 0, 1, 5, 7} {
		t.Fatalf("after Translate GetTransform() = %v, want {1,0,0,1,5,7}", got)
	}
	ctx.ResetTransform()
	if got := ctx.GetTransform(); !got.IsIdentity() {
		t.Fatalf("after ResetTransform GetTransform() = %v, want identity", got.AffineMatrix)
	}
}

func TestCPPCompOpSrcKeepsStraightAlphaWithAggReal(t *testing.T) {
	// comp_op_src must replace the destination with the straight source colour,
	// not a premultiplied one. The C++ comp-op adaptor demultiplies on write to
	// match the port's CompositeBlenderPlain; without it the stored RGB would be
	// premultiplied (e.g. ~25,38,138 instead of 40,60,220 for alpha 160).
	ctx, err := engine.NewContext(64, 64, engine.Config{Kind: engine.CPP})
	if err != nil {
		t.Fatalf("NewContext(CPP) error = %v", err)
	}
	ctx.Clear(agg.White)
	ctx.SetBlendMode(agg.BlendSrc)
	ctx.SetFillColor(agg.NewColor(40, 60, 220, 160))
	ctx.FillRectangle(8, 8, 48, 48)

	got := ctx.GetImage().ToGoImage().RGBAAt(32, 32)
	// Straight colour is (40,60,220,160); the premultiplied bug would store
	// roughly (25,38,138,160). Allow 1-LSB slack from the integer
	// premultiply/demultiply round-trip — the distinction from premultiplied is
	// ~80 LSB, far larger than the tolerance.
	within := func(got, want uint8) bool {
		d := int(got) - int(want)
		return d >= -2 && d <= 2
	}
	if !within(got.R, 40) || !within(got.G, 60) || !within(got.B, 220) || got.A != 160 {
		t.Fatalf("comp_op_src center pixel = %+v, want ~straight (40,60,220,160), not premultiplied", got)
	}
}

func TestCPPCompOpSrcDoesNotWipeBackgroundWithAggReal(t *testing.T) {
	// comp_op_src must only affect the rendered shape, not the whole buffer. The
	// earlier layer-then-composite path wiped the untouched background to clear.
	ctx, err := engine.NewContext(64, 64, engine.Config{Kind: engine.CPP})
	if err != nil {
		t.Fatalf("NewContext(CPP) error = %v", err)
	}
	ctx.Clear(agg.White)
	ctx.SetBlendMode(agg.BlendSrc)
	ctx.SetFillColor(agg.NewColor(40, 60, 220, 160))
	ctx.FillRectangle(20, 20, 24, 24)

	// A pixel well outside the rectangle must still be the opaque white clear.
	bg := ctx.GetImage().ToGoImage().RGBAAt(4, 4)
	if bg.R != 255 || bg.G != 255 || bg.B != 255 || bg.A != 255 {
		t.Fatalf("background outside src rect = %+v, want opaque white", bg)
	}
}

func TestCPPGradientUnderSrcBlendDoesNotWipeBackgroundWithAggReal(t *testing.T) {
	// A gradient fill under a non-src-over blend must apply the operator only
	// within the shape's coverage. The earlier gradient path composited the whole
	// layer rectangle, so src wiped the untouched background to transparent.
	ctx, err := engine.NewContext(256, 256, engine.Config{Kind: engine.CPP})
	if err != nil {
		t.Fatalf("NewContext(CPP) error = %v", err)
	}
	ctx.Clear(agg.White)
	// Opaque green background block.
	ctx.SetFillColor(agg.NewColorRGB(40, 130, 60))
	ctx.FillRectangle(30, 30, 150, 150)
	// Translucent gradient circle under src.
	ctx.SetBlendMode(agg.BlendSrc)
	ctx.SetLinearGradient(96, 96, 220, 220,
		agg.NewColor(220, 40, 40, 160), agg.NewColor(40, 60, 220, 160))
	ctx.FillCircle(140, 140, 70)

	img := ctx.GetImage().ToGoImage()

	// Background inside the green block but outside the circle must survive.
	bg := img.RGBAAt(40, 40)
	within := func(got, want uint8) bool {
		d := int(got) - int(want)
		return d >= -2 && d <= 2
	}
	if !within(bg.R, 40) || !within(bg.G, 130) || !within(bg.B, 60) || bg.A != 255 {
		t.Fatalf("background outside gradient circle = %+v, want opaque green ~(40,130,60,255)", bg)
	}

	// Circle interior must hold the straight translucent gradient (alpha ~160),
	// proving src replaced the destination rather than wiping or src-over blending.
	center := img.RGBAAt(140, 140)
	if center.A < 150 || center.A > 170 {
		t.Fatalf("gradient circle center alpha = %d, want straight translucent ~160", center.A)
	}
	if center.R == bg.R && center.G == bg.G && center.B == bg.B {
		t.Fatalf("gradient circle center = %+v, expected gradient colour not the background", center)
	}
}

func TestCPPDashedStrokeReducesInkWithAggReal(t *testing.T) {
	inkOnLine := func(dashed bool) int {
		ctx, err := engine.NewContext(120, 20, engine.Config{Kind: engine.CPP})
		if err != nil {
			t.Fatalf("NewContext(CPP) error = %v", err)
		}
		ctx.Clear(agg.White)
		ctx.SetStrokeColor(agg.NewColorRGB(0, 0, 0))
		ctx.SetLineWidth(3)
		ctx.SetLineCap(agg.CapButt)
		if dashed {
			ctx.AddDash(8, 8)
		}
		ctx.BeginPath()
		ctx.MoveTo(5, 10)
		ctx.LineTo(115, 10)
		ctx.Stroke()

		img := ctx.GetImage().ToGoImage()
		ink := 0
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				p := img.RGBAAt(x, y)
				if p.R < 250 || p.G < 250 || p.B < 250 {
					ink++
				}
			}
		}
		return ink
	}

	solid := inkOnLine(false)
	dashed := inkOnLine(true)
	if solid == 0 || dashed == 0 {
		t.Fatalf("expected both strokes to draw ink: solid=%d dashed=%d", solid, dashed)
	}
	if dashed >= solid {
		t.Fatalf("expected dashed ink (%d) < solid ink (%d) under real AGG backend", dashed, solid)
	}
}

func TestCPPTextWorksWithAggReal(t *testing.T) {
	fontPath := "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	if _, err := os.Stat(fontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}

	ctx, err := engine.NewContext(120, 40, engine.Config{Kind: engine.CPP})
	if err != nil {
		t.Fatalf("NewContext(CPP) error = %v", err)
	}
	ctx.Clear(agg.White)
	ctx.SetFillColor(agg.Black)
	if err := ctx.LoadFont(fontPath); err != nil {
		t.Fatalf("LoadFont() error = %v", err)
	}
	ctx.TextHints(true)

	width, height := ctx.MeasureText("Hello")
	if width <= 0 || height <= 0 {
		t.Fatalf("unexpected text metrics: width=%v height=%v", width, height)
	}

	if err := ctx.DrawText("Hello", 10, 20); err != nil {
		t.Fatalf("DrawText() error = %v", err)
	}

	img := ctx.GetImage().ToGoImage()
	nonWhite := false
	for y := 0; y < img.Bounds().Dy() && !nonWhite; y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			p := img.RGBAAt(x, y)
			if p.R != agg.White.R || p.G != agg.White.G || p.B != agg.White.B || p.A != agg.White.A {
				nonWhite = true
				break
			}
		}
	}
	if !nonWhite {
		t.Fatal("expected rendered text to modify at least one pixel")
	}
}
