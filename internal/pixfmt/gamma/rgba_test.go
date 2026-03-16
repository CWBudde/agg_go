package gamma

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
)

func TestApplyGammaRGBAOnlyTouchesRGB(t *testing.T) {
	gamma := NewSimpleGammaLut(2.2)
	pixel := []basics.Int8u{128, 64, 192, 77}
	origAlpha := pixel[3]

	NewApplyGammaDirRGBA[color.Linear](gamma).Apply(pixel)
	if pixel[3] != origAlpha {
		t.Fatalf("direct gamma changed alpha: %d -> %d", origAlpha, pixel[3])
	}

	NewApplyGammaInvRGBA[color.Linear](gamma).Apply(pixel)
	if pixel[3] != origAlpha {
		t.Fatalf("inverse gamma changed alpha: %d -> %d", origAlpha, pixel[3])
	}
}

func TestApplyGammaRGBAIgnoresShortBuffers(t *testing.T) {
	gamma := NewSimpleGammaLut(2.2)
	pixel := []basics.Int8u{1, 2, 3}
	orig := append([]basics.Int8u(nil), pixel...)

	NewApplyGammaDirRGBA[color.Linear](gamma).Apply(pixel)
	if string(pixel) != string(orig) {
		t.Fatalf("short direct buffer changed: %v -> %v", orig, pixel)
	}

	NewApplyGammaInvRGBA[color.Linear](gamma).Apply(pixel)
	if string(pixel) != string(orig) {
		t.Fatalf("short inverse buffer changed: %v -> %v", orig, pixel)
	}
}

func TestPixFmtRGBAGammaDimensions(t *testing.T) {
	rbuf := buffer.NewRenderingBufferU8WithData(make([]basics.Int8u, 4*3*4), 4, 3, 16)
	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	gammaPF := NewPixFmtRGBA32Gamma(pf, 2.2)
	linearPF := NewPixFmtRGBA32Linear(pf)

	if gammaPF.Width() != 4 || gammaPF.Height() != 3 || gammaPF.PixWidth() != 4 {
		t.Fatalf("gamma pixfmt dims = (%d,%d,%d)", gammaPF.Width(), gammaPF.Height(), gammaPF.PixWidth())
	}
	if linearPF.Width() != 4 || linearPF.Height() != 3 || linearPF.PixWidth() != 4 {
		t.Fatalf("linear pixfmt dims = (%d,%d,%d)", linearPF.Width(), linearPF.Height(), linearPF.PixWidth())
	}
}

func TestRGBAMultiplierPremultiplyAndDemultiply(t *testing.T) {
	rgba := []basics.Int8u{100, 50, 25, 128}
	RGBAMultiplier[order.RGBA]{}.Premultiply(rgba)
	if rgba != nil && (rgba[0] != 50 || rgba[1] != 25 || rgba[2] != 12 || rgba[3] != 128) {
		t.Fatalf("Premultiply RGBA = %v", rgba)
	}
	RGBAMultiplier[order.RGBA]{}.Demultiply(rgba)
	if rgba[0] != 100 || rgba[1] != 50 || rgba[2] != 24 || rgba[3] != 128 {
		t.Fatalf("Demultiply RGBA = %v", rgba)
	}

	zero := []basics.Int8u{90, 80, 70, 0}
	RGBAMultiplier[order.RGBA]{}.Premultiply(zero)
	if zero[0] != 0 || zero[1] != 0 || zero[2] != 0 || zero[3] != 0 {
		t.Fatalf("Premultiply zero alpha = %v", zero)
	}
}

func TestRGBAMultiplierRespectsComponentOrder(t *testing.T) {
	argb := []basics.Int8u{128, 90, 45, 18}
	RGBAMultiplier[order.ARGB]{}.Premultiply(argb)
	if argb[0] != 128 || argb[1] != 45 || argb[2] != 22 || argb[3] != 9 {
		t.Fatalf("Premultiply ARGB = %v", argb)
	}

	bgra := []basics.Int8u{20, 40, 80, 128}
	RGBAMultiplier[order.BGRA]{}.Premultiply(bgra)
	if bgra[0] != 10 || bgra[1] != 20 || bgra[2] != 40 || bgra[3] != 128 {
		t.Fatalf("Premultiply BGRA = %v", bgra)
	}

	abgr := []basics.Int8u{128, 20, 40, 80}
	RGBAMultiplier[order.ABGR]{}.Premultiply(abgr)
	if abgr[0] != 128 || abgr[1] != 10 || abgr[2] != 20 || abgr[3] != 40 {
		t.Fatalf("Premultiply ABGR = %v", abgr)
	}
}
