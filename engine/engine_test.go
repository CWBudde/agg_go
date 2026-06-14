//go:build !aggreal

package engine_test

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"testing"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/engine"
)

type foreignImage struct{}

func (foreignImage) Kind() engine.Kind      { return engine.CPP }
func (foreignImage) Width() int             { return 1 }
func (foreignImage) Height() int            { return 1 }
func (foreignImage) Premultiply() error     { return nil }
func (foreignImage) Demultiply() error      { return nil }
func (foreignImage) ToGoImage() *image.RGBA { return image.NewRGBA(image.Rect(0, 0, 1, 1)) }
func (foreignImage) ToStandardImage() (image.Image, error) {
	return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
}
func (foreignImage) SaveToPNG(string) error { return nil }
func (foreignImage) SaveToJPEG(string, int) error {
	return nil
}

func TestAvailableIncludesPort(t *testing.T) {
	available := engine.Available()
	if len(available) == 0 {
		t.Fatal("expected at least one available engine")
	}
	if available[0] != engine.Port {
		t.Fatalf("expected first available engine to be %q, got %q", engine.Port, available[0])
	}
}

func TestPortCapabilitiesExposeCurrentFacadeSurface(t *testing.T) {
	caps, err := engine.Capabilities(engine.Port)
	if err != nil {
		t.Fatalf("Capabilities() error = %v", err)
	}

	for _, want := range []engine.Capability{
		engine.CapabilitySolidStyle,
		engine.CapabilityPath,
		engine.CapabilityTransforms,
		engine.CapabilityClipBox,
		engine.CapabilityCompositing,
		engine.CapabilityImageDraw,
		engine.CapabilityImageExport,
		engine.CapabilityImageInterop,
		engine.CapabilityGradients,
		engine.CapabilityText,
	} {
		if !containsCapability(caps, want) {
			t.Fatalf("expected capability set to include %q, got %v", want, caps)
		}
	}
}

func TestCapabilitiesCPPUnavailable(t *testing.T) {
	_, err := engine.Capabilities(engine.CPP)
	if err == nil {
		t.Fatal("expected unavailable error for C++ capability query")
	}
	if !errors.Is(err, engine.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}

// unknownCapability is a capability no engine implements, used to exercise the
// unsupported-capability error path. (The port engine now supports every real
// facade capability, so a fictitious one is needed for the negative case.)
const unknownCapability engine.Capability = "imaginary_feature"

func TestRequireCapabilityReturnsTypedUnsupportedError(t *testing.T) {
	err := engine.RequireCapability(engine.Port, unknownCapability, "ImaginaryOp")
	if err == nil {
		t.Fatal("expected unsupported capability error")
	}
	if !errors.Is(err, engine.ErrUnsupportedCapability) {
		t.Fatalf("expected ErrUnsupportedCapability, got %v", err)
	}

	var unsupported *engine.UnsupportedCapabilityError
	if !errors.As(err, &unsupported) {
		t.Fatalf("expected UnsupportedCapabilityError, got %T", err)
	}
	if unsupported.Kind != engine.Port || unsupported.Capability != unknownCapability || unsupported.Operation != "ImaginaryOp" {
		t.Fatalf("unexpected unsupported capability payload: %+v", unsupported)
	}
}

func TestSupportsReflectsCapabilitySet(t *testing.T) {
	if !engine.Supports(engine.Port, engine.CapabilityText) {
		t.Fatal("expected port engine to support text capability")
	}
	if !engine.Supports(engine.Port, engine.CapabilityDashedStroke) {
		t.Fatal("expected port engine to support dashed-stroke capability")
	}
	if engine.Supports(engine.Port, unknownCapability) {
		t.Fatal("did not expect port engine to report an unknown capability")
	}
	if engine.Supports(engine.CPP, engine.CapabilityText) {
		t.Fatal("did not expect unavailable C++ engine to report text capability")
	}
}

func TestNewContextDefaultsToPort(t *testing.T) {
	ctx, err := engine.NewContext(8, 8, engine.Config{})
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}
	if ctx.Kind() != engine.Port {
		t.Fatalf("expected default engine %q, got %q", engine.Port, ctx.Kind())
	}

	ctx.Clear(agg.White)
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(1, 1, 6, 6)

	got := ctx.GetImage().ToGoImage().RGBAAt(4, 4)
	if got.R < 200 || got.G > 40 || got.B > 40 || got.A != 255 {
		t.Fatalf("unexpected rendered color at center: %+v", got)
	}
}

func TestNewContextCPPUnavailable(t *testing.T) {
	_, err := engine.NewContext(16, 16, engine.Config{Kind: engine.CPP})
	if err == nil {
		t.Fatal("expected error when requesting unavailable C++ engine")
	}
	if !errors.Is(err, engine.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}

	var unavailable *engine.UnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected UnavailableError, got %T", err)
	}
	if unavailable.Kind != engine.CPP {
		t.Fatalf("expected unavailable kind %q, got %q", engine.CPP, unavailable.Kind)
	}
}

func TestDrawImageThroughPortEngine(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 255, A: 255})

	ctx, err := engine.NewContext(4, 4, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}
	img, err := engine.NewImageFromGoImage(src, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewImageFromGoImage() error = %v", err)
	}

	ctx.Clear(agg.Transparent)
	if err := ctx.DrawImageScaled(img, 0, 0, 4, 4); err != nil {
		t.Fatalf("DrawImageScaled() error = %v", err)
	}

	got := ctx.GetImage().ToGoImage().RGBAAt(2, 2)
	if got.R < 200 || got.A != 255 {
		t.Fatalf("unexpected scaled image pixel: %+v", got)
	}
}

func TestNewImageAndContextForImagePort(t *testing.T) {
	img, err := engine.NewImage(4, 4, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}
	ctx, err := engine.NewContextForImage(img)
	if err != nil {
		t.Fatalf("NewContextForImage() error = %v", err)
	}

	ctx.Clear(agg.Blue)
	got := img.ToGoImage().RGBAAt(1, 1)
	if got.B < 200 || got.A != 255 {
		t.Fatalf("unexpected attached image pixel: %+v", got)
	}
}

func TestNewImageFromBufferUsesCallerBuffer(t *testing.T) {
	buf := make([]byte, 4*4*4)
	img, err := engine.NewImageFromBuffer(buf, 4, 4, 16, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewImageFromBuffer() error = %v", err)
	}
	ctx, err := engine.NewContextForImage(img)
	if err != nil {
		t.Fatalf("NewContextForImage() error = %v", err)
	}

	ctx.Clear(agg.Green)
	if buf[0] != 0 || buf[1] < 200 || buf[2] != 0 || buf[3] != 255 {
		t.Fatalf("caller buffer was not updated as expected: %v", buf[:4])
	}
}

func TestEngineMismatchError(t *testing.T) {
	ctx, err := engine.NewContext(4, 4, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	cppImg, err := engine.NewImage(1, 1, engine.Config{Kind: engine.CPP})
	if err == nil {
		t.Fatal("expected unavailable C++ image creation to fail in this build")
	}
	if cppImg != nil {
		t.Fatal("expected nil image from unavailable C++ image creation")
	}

	err = ctx.DrawImageScaled(foreignImage{}, 0, 0, 1, 1)
	if err == nil {
		t.Fatal("expected engine mismatch error")
	}
	if !errors.Is(err, engine.ErrEngineMismatch) {
		t.Fatalf("expected ErrEngineMismatch, got %v", err)
	}

	var mismatch *engine.EngineMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("expected EngineMismatchError, got %T", err)
	}
	if mismatch.ContextKind != engine.Port || mismatch.ResourceKind != engine.CPP {
		t.Fatalf("unexpected mismatch payload: %+v", mismatch)
	}
}

func TestBlendAndFillRuleReadback(t *testing.T) {
	ctx, err := engine.NewContext(4, 4, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	ctx.SetBlendMode(agg.BlendMultiply)
	if got := ctx.GetBlendMode(); got != agg.BlendMultiply {
		t.Fatalf("GetBlendMode() = %v, want %v", got, agg.BlendMultiply)
	}

	ctx.FillEvenOdd(true)
	if !ctx.GetFillEvenOdd() {
		t.Fatal("expected even-odd fill rule to be enabled")
	}
}

func TestGradientFacadeReadback(t *testing.T) {
	ctx, err := engine.NewContext(8, 8, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	ctx.SetLinearGradient(0, 0, 8, 0, agg.Red, agg.Blue)
	if got := ctx.GetFillGradientType(); got != agg.LinearGradient {
		t.Fatalf("GetFillGradientType() = %v, want %v", got, agg.LinearGradient)
	}

	ctx.SetStrokeRadialGradient(4, 4, 4, agg.White, agg.Black)
	if got := ctx.GetStrokeGradientType(); got != agg.RadialGradient {
		t.Fatalf("GetStrokeGradientType() = %v, want %v", got, agg.RadialGradient)
	}
}

func TestTextFacadeConfigurationAndValidation(t *testing.T) {
	ctx, err := engine.NewContext(16, 16, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	ctx.TextHints(true)
	if !ctx.GetTextHints() {
		t.Fatal("expected text hints to be enabled")
	}

	ctx.SetTextAlignment(agg.AlignCenter, agg.AlignTop)

	if err := ctx.DrawText("", 1, 1); err == nil {
		t.Fatal("expected empty text to be rejected")
	}
}

func TestImageInteropHelpers(t *testing.T) {
	img, err := engine.NewImage(3, 2, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}
	ctx, err := engine.NewContextForImage(img)
	if err != nil {
		t.Fatalf("NewContextForImage() error = %v", err)
	}

	ctx.Clear(agg.Transparent)
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(0, 0, 3, 2)

	stdImg, err := img.ToStandardImage()
	if err != nil {
		t.Fatalf("ToStandardImage() error = %v", err)
	}
	if got := stdImg.Bounds(); got.Dx() != 3 || got.Dy() != 2 {
		t.Fatalf("unexpected standard image bounds: %v", got)
	}
	r, g, b, a := stdImg.At(1, 1).RGBA()
	if r < 0xf000 || g != 0 || b != 0 || a != 0xffff {
		t.Fatalf("unexpected standard image pixel: r=%#x g=%#x b=%#x a=%#x", r, g, b, a)
	}

	out := t.TempDir() + "/out.jpg"
	if err := img.SaveToJPEG(out, 90); err != nil {
		t.Fatalf("SaveToJPEG() error = %v", err)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", out, err)
	}
	defer f.Close()

	decoded, err := jpeg.Decode(f)
	if err != nil {
		t.Fatalf("jpeg.Decode() error = %v", err)
	}
	if got := decoded.Bounds(); got.Dx() != 3 || got.Dy() != 2 {
		t.Fatalf("unexpected decoded JPEG bounds: %v", got)
	}
}

func TestPortDashedStrokeReducesInkAndRestores(t *testing.T) {
	// inkOnLine strokes a single horizontal line and counts the non-white pixels,
	// optionally applying a dash pattern first.
	inkOnLine := func(dashed bool) int {
		ctx, err := engine.NewContext(120, 20, engine.Config{Kind: engine.Port})
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}
		ctx.Clear(agg.White)
		ctx.SetStrokeColor(agg.NewColorRGB(0, 0, 0))
		ctx.SetLineWidth(3)
		ctx.SetLineCap(agg.CapButt)
		if dashed {
			ctx.AddDash(8, 8)
		} else {
			ctx.RemoveAllDashes()
		}
		ctx.BeginPath()
		ctx.MoveTo(5, 10)
		ctx.LineTo(115, 10)
		ctx.Stroke()

		img := ctx.GetImage().ToGoImage()
		ink := 0
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				if r>>8 < 250 || g>>8 < 250 || b>>8 < 250 {
					ink++
				}
			}
		}
		return ink
	}

	solid := inkOnLine(false)
	dashed := inkOnLine(true)
	if solid == 0 {
		t.Fatal("solid stroke drew no ink")
	}
	if dashed == 0 {
		t.Fatal("dashed stroke drew no ink")
	}
	// An 8-on/8-off pattern should leave roughly half the line unpainted.
	if dashed >= solid {
		t.Fatalf("expected dashed ink (%d) to be less than solid ink (%d)", dashed, solid)
	}
}

func TestPortDashStartRoundTrips(t *testing.T) {
	ctx, err := engine.NewContext(32, 32, engine.Config{Kind: engine.Port})
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}
	ctx.AddDash(10, 5)
	ctx.DashStart(7)
	if got := ctx.GetDashStart(); got != 7 {
		t.Fatalf("GetDashStart() = %v, want 7", got)
	}
}

func containsCapability(caps []engine.Capability, want engine.Capability) bool {
	for _, cap := range caps {
		if cap == want {
			return true
		}
	}
	return false
}
