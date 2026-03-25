package main

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
)

func TestBackgroundOnlyPixelMatchesCurrentPipeline(t *testing.T) {
	img := agg.NewImage(make([]byte, frameWidth*frameHeight*4), frameWidth, frameHeight, frameWidth*4)
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pf)

	bg := color.ConvertFromRGBA[color.Linear](color.RGBA{R: 1.0, G: 0.95, B: 0.95, A: 1.0})
	renBase.Clear(bg)

	goImg := img.ToGoImage()
	got := goImg.RGBAAt(0, 0)
	if got.R != 255 || got.G != 242 || got.B != 242 || got.A != 255 {
		t.Fatalf("background-only pixel = %v, want rgba(255,242,242,255)", got)
	}
}

func TestBackgroundOnlyPixelDiffersFromCPPReference(t *testing.T) {
	refPath := filepath.Join("..", "..", "..", "..", "tests", "visual", "reference", "cpp", "examples", "blend_color.png")
	f, err := os.Open(refPath)
	if err != nil {
		t.Fatalf("open cpp reference: %v", err)
	}
	defer f.Close()

	refImg, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode cpp reference: %v", err)
	}

	ref := refImg.At(0, 0)
	r, g, b, a := ref.RGBA()
	if r>>8 != 255 || g>>8 != 249 || b>>8 != 249 || a>>8 != 255 {
		t.Fatalf("cpp reference pixel changed: got rgba(%d,%d,%d,%d), want rgba(255,249,249,255)", r>>8, g>>8, b>>8, a>>8)
	}
}
