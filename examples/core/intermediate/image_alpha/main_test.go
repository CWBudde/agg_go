package main

import (
	"bytes"
	"testing"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

func TestImageAlphaSRGBA8MatchesCXXColorTypeConversion(t *testing.T) {
	got := imageAlphaSRGBA8(129, 255, 74, 236)
	want := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{
		R: 129,
		G: 255,
		B: 74,
		A: 236,
	})
	if got != want {
		t.Fatalf("imageAlphaSRGBA8() = %+v, want C++ srgba8 converted to rgba8 %+v", got, want)
	}
}

func TestImageAlphaRandomEllipseUsesCXXRGBAOrder(t *testing.T) {
	rng := newClibcRandSeed1()
	e := nextImageAlphaEllipse(rng, 320, 300)

	wantColor := imageAlphaSRGBA8(236, 74, 255, 81)
	if e.color != wantColor {
		t.Fatalf("first ellipse color = %+v, want GCC C++ rand argument order converted from srgba8 %+v", e.color, wantColor)
	}
}

func TestLoadPPMImageMatchesCXXSRGB24ToBGR24Conversion(t *testing.T) {
	img, err := loadPPMImage(bytes.NewReader([]byte{
		'P', '6', '\n',
		'1', ' ', '1', '\n',
		'2', '5', '5', '\n',
		81, 255, 74,
	}))
	if err != nil {
		t.Fatalf("loadPPMImage() error = %v", err)
	}

	want := imageAlphaSRGBA8(81, 255, 74, 255)
	if got := img.Data[:3]; got[0] != want.B || got[1] != want.G || got[2] != want.R {
		t.Fatalf("stored pixel BGR = %v, want C++ pixfmt_bgr24 linear bytes [%d %d %d]", got, want.B, want.G, want.R)
	}
}

func TestImageAlphaRunnerConfigMatchesCXXSaveImgConversion(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.FlipY {
		t.Fatal("image_alpha must use FlipY=true to match C++ platform_support")
	}
	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatal("image_alpha renders linear pixfmt_bgr24-equivalent pixels and must encode output like C++ save_img")
	}
	if cfg.DisableLinearRGBToSRGB {
		t.Fatal("image_alpha should use the standard linear-to-sRGB output path")
	}
	if cfg.Width != 320 || cfg.Height != 300 {
		t.Fatalf("runner dimensions = %dx%d, want 320x300", cfg.Width, cfg.Height)
	}
	_ = lowlevelrunner.Config{}
}

func TestImageAlphaClipEllipseUsesCXXVertexCount(t *testing.T) {
	ps := buildTransformedEllipsePath(320, 300, transform.NewTransAffine())
	ps.Rewind(0)

	vertices := 0
	for {
		_, _, cmd := ps.NextVertex()
		pathCmd := basics.PathCommand(cmd)
		if basics.IsStop(pathCmd) {
			break
		}
		if basics.IsVertex(pathCmd) {
			vertices++
		}
	}

	if vertices != imageAlphaClipEllipseSteps {
		t.Fatalf("clip ellipse vertices = %d, want C++ ellipse.init(..., %d)", vertices, imageAlphaClipEllipseSteps)
	}
}

func TestImageAlphaClipSourceUsesCXXEllipseEndPolyFlags(t *testing.T) {
	src := newImageAlphaClipSource(320, 300, transform.NewTransAffine())
	src.Rewind(0)

	var x, y float64
	for i := 0; i < imageAlphaClipEllipseSteps; i++ {
		cmd := basics.PathCommand(src.Vertex(&x, &y))
		if !basics.IsVertex(cmd) {
			t.Fatalf("clip source command %d = %v, want vertex", i, cmd)
		}
	}

	cmd := basics.PathCommand(src.Vertex(&x, &y))
	if !basics.IsEndPoly(cmd) || !basics.IsClosed(uint32(cmd)) || !basics.IsCCW(uint32(cmd)) {
		t.Fatalf("clip source end command = %v, want C++ ellipse end_poly|close|ccw", cmd)
	}
}

func TestImageAlphaBackgroundEllipseUsesCXXStepCount(t *testing.T) {
	if imageAlphaBackgroundEllipseSteps != 50 {
		t.Fatalf("background ellipse steps = %d, want C++ ell.init(..., 50)", imageAlphaBackgroundEllipseSteps)
	}
}

func TestImageAlphaAlphaByteUsesCXXTruncation(t *testing.T) {
	if got := imageAlphaAlphaByte(0.501); got != 127 {
		t.Fatalf("imageAlphaAlphaByte(0.501) = %d, want C++ int8u(value*255.0) truncation", got)
	}
}

func TestImageAlphaRenderTargetUsesCXXBGR24(t *testing.T) {
	img := agg.NewImage(make([]uint8, 2*2*4), 2, 2, -2*4)
	target := newImageAlphaRenderTarget(img)

	if target.pixf.PixWidth() != 3 {
		t.Fatalf("render target pix width = %d, want C++ pixfmt_bgr24 width 3", target.pixf.PixWidth())
	}

	target.renBase.Clear(color.NewRGBA8[color.Linear](10, 20, 30, 255))
	target.copyToImage(img)

	goImg := img.ToGoImage()
	got := goImg.RGBAAt(0, 0)
	if got.R != 10 || got.G != 20 || got.B != 30 || got.A != 255 {
		t.Fatalf("copied BGR24 target pixel = RGBA(%d,%d,%d,%d), want RGBA(10,20,30,255)", got.R, got.G, got.B, got.A)
	}
}

func TestImageAlphaUsesCXXRGBBilinearFilterWithClipAccessor(t *testing.T) {
	src := &imageClipSource{
		accessor: image.NewImageAccessorClip(&imagePixFmt{}, []basics.Int8u{0, 0, 0, 0}),
		ipf:      &imagePixFmt{},
	}
	interp := span.NewSpanInterpolatorLinear(transform.NewTransAffine(), 8)

	var _ *span.SpanImageFilterRGBBilinear[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]] = newImageAlphaRGBBilinear(src, interp)
}
