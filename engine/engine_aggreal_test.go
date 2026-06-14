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
