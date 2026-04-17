package engine_test

import (
	"errors"
	"image"
	"image/color"
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
func (foreignImage) SaveToPNG(string) error { return nil }

func TestAvailableIncludesPort(t *testing.T) {
	available := engine.Available()
	if len(available) == 0 {
		t.Fatal("expected at least one available engine")
	}
	if available[0] != engine.Port {
		t.Fatalf("expected first available engine to be %q, got %q", engine.Port, available[0])
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
