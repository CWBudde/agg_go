package blender

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/order"
)

func TestBlenderRGBA16SetGetPlainRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "PremultipliedFramebuffer",
			run: func(t *testing.T) {
				bl := BlenderRGBA16[color.Linear, order.RGBA]{}
				dst := make([]basics.Int8u, 8)
				bl.SetPlain(dst, 40000, 20000, 10000, 50000)
				r, g, b, a := bl.GetPlain(dst)
				if r < 39990 || r > 40010 || g < 19990 || g > 20010 || b < 9990 || b > 10010 || a != 50000 {
					t.Fatalf("round trip got (%d,%d,%d,%d), want near (40000,20000,10000,50000)", r, g, b, a)
				}
				if bl.IdxR() != 0 || bl.IdxG() != 1 || bl.IdxB() != 2 || bl.IdxA() != 3 {
					t.Fatalf("unexpected RGBA order indices")
				}
			},
		},
		{
			name: "PremultipliedSourceBlender",
			run: func(t *testing.T) {
				bl := BlenderRGBA16Pre[color.Linear, order.RGBA]{}
				dst := make([]basics.Int8u, 8)
				bl.SetPlain(dst, 30000, 15000, 5000, 45000)
				r, g, b, a := bl.GetPlain(dst)
				if r < 29990 || r > 30010 || g < 14990 || g > 15010 || b < 4990 || b > 5010 || a != 45000 {
					t.Fatalf("round trip got (%d,%d,%d,%d), want near (30000,15000,5000,45000)", r, g, b, a)
				}
			},
		},
		{
			name: "PlainFramebuffer",
			run: func(t *testing.T) {
				bl := BlenderRGBA16Plain[color.Linear, order.RGBA]{}
				dst := make([]basics.Int8u, 8)
				bl.SetPlain(dst, 11111, 22222, 33333, 44444)
				r, g, b, a := bl.GetPlain(dst)
				if r != 11111 || g != 22222 || b != 33333 || a != 44444 {
					t.Fatalf("plain round trip got (%d,%d,%d,%d), want (11111,22222,33333,44444)", r, g, b, a)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

func TestBlenderRGBA16BlendPixChangesDestination(t *testing.T) {
	tests := []struct {
		name string
		run  func(dst []basics.Int8u)
	}{
		{
			name: "PlainToPremul",
			run: func(dst []basics.Int8u) {
				BlenderRGBA16[color.Linear, order.RGBA]{}.BlendPix(dst, 50000, 10000, 2000, 32768, 65535)
			},
		},
		{
			name: "PremulToPremul",
			run: func(dst []basics.Int8u) {
				BlenderRGBA16Pre[color.Linear, order.RGBA]{}.BlendPix(dst, 25000, 5000, 1000, 32768, 65535)
			},
		},
		{
			name: "PlainToPlain",
			run: func(dst []basics.Int8u) {
				BlenderRGBA16Plain[color.Linear, order.RGBA]{}.BlendPix(dst, 45000, 12000, 3000, 32768, 65535)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dst := make([]basics.Int8u, 8)
			tc.run(dst)
			zero := true
			for _, v := range dst {
				if v != 0 {
					zero = false
					break
				}
			}
			if zero {
				t.Fatal("BlendPix left destination unchanged")
			}
		})
	}
}

func TestBlendRGBA16Pixel(t *testing.T) {
	dst := make([]basics.Int8u, 8)
	src := color.NewRGBA16[color.Linear](40000, 30000, 20000, 50000)
	var bl BlenderRGBA16[color.Linear, order.RGBA]

	BlendRGBA16Pixel(dst, src, 65535, bl)
	r, g, b, a := bl.GetPlain(dst)
	if r == 0 || g == 0 || b == 0 || a == 0 {
		t.Fatalf("BlendRGBA16Pixel produced zeroed pixel (%d,%d,%d,%d)", r, g, b, a)
	}

	orig := append([]basics.Int8u(nil), dst...)
	transparent := color.NewRGBA16[color.Linear](60000, 60000, 60000, 0)
	BlendRGBA16Pixel(dst, transparent, 65535, bl)
	for i := range dst {
		if dst[i] != orig[i] {
			t.Fatalf("transparent BlendRGBA16Pixel changed dst[%d]: got %d want %d", i, dst[i], orig[i])
		}
	}
}
