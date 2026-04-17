package engine_test

import (
	"errors"
	"image"
	"image/color"
	"testing"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/engine"
)

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
