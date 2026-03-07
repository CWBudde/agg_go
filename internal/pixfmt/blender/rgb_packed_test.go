package blender

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestPackedPixelRoundTrips(t *testing.T) {
	tests := []struct {
		name   string
		make   func(r, g, b basics.Int8u) basics.Int16u
		unpack func(pixel basics.Int16u) (r, g, b basics.Int8u)
	}{
		{"RGB555", MakePixel555, UnpackPixel555},
		{"BGR555", MakePixelBGR555, UnpackPixelBGR555},
		{"RGB565", MakePixel565, UnpackPixel565},
		{"BGR565", MakePixelBGR565, UnpackPixelBGR565},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pixel := tc.make(0xE0, 0xA0, 0x40)
			r, g, b := tc.unpack(pixel)
			if r == 0 || g == 0 || b == 0 {
				t.Fatalf("unpacked %s pixel = (%d,%d,%d), expected non-zero quantized channels", tc.name, r, g, b)
			}
		})
	}
}

func TestPackedBlendersModifyPixels(t *testing.T) {
	tests := []struct {
		name      string
		initial   basics.Int16u
		blend     func(*basics.Int16u)
		unpack    func(pixel basics.Int16u) (r, g, b basics.Int8u)
		makeColor func(pixel basics.Int16u) color.RGB8[color.Linear]
	}{
		{
			name:      "RGB555",
			initial:   MakePixel555(0, 0, 0),
			blend:     func(p *basics.Int16u) { BlenderRGB555{}.BlendPix(p, 255, 0, 0, 255, 255) },
			unpack:    UnpackPixel555,
			makeColor: MakeColorRGB555,
		},
		{
			name:      "RGB555Pre",
			initial:   MakePixel555(0, 0, 0),
			blend:     func(p *basics.Int16u) { BlenderRGB555Pre{}.BlendPix(p, 200, 0, 0, 255, 255) },
			unpack:    UnpackPixel555,
			makeColor: MakeColorRGB555,
		},
		{
			name:      "RGB555Gamma",
			initial:   MakePixel555(0, 0, 0),
			blend:     func(p *basics.Int16u) { NewBlenderRGB555Gamma(trivialGamma{}).BlendPix(p, 255, 0, 0, 255, 255) },
			unpack:    UnpackPixel555,
			makeColor: MakeColorRGB555,
		},
		{
			name:      "RGB565",
			initial:   MakePixel565(0, 0, 0),
			blend:     func(p *basics.Int16u) { BlenderRGB565{}.BlendPix(p, 0, 255, 0, 255, 255) },
			unpack:    UnpackPixel565,
			makeColor: MakeColorRGB565,
		},
		{
			name:      "RGB565Pre",
			initial:   MakePixel565(0, 0, 0),
			blend:     func(p *basics.Int16u) { BlenderRGB565Pre{}.BlendPix(p, 0, 200, 0, 255, 255) },
			unpack:    UnpackPixel565,
			makeColor: MakeColorRGB565,
		},
		{
			name:      "RGB565Gamma",
			initial:   MakePixel565(0, 0, 0),
			blend:     func(p *basics.Int16u) { NewBlenderRGB565Gamma(trivialGamma{}).BlendPix(p, 0, 255, 0, 255, 255) },
			unpack:    UnpackPixel565,
			makeColor: MakeColorRGB565,
		},
		{
			name:      "BGR555",
			initial:   MakePixelBGR555(0, 0, 0),
			blend:     func(p *basics.Int16u) { BlenderBGR555{}.BlendPix(p, 0, 0, 255, 255, 255) },
			unpack:    UnpackPixelBGR555,
			makeColor: MakeColorBGR555,
		},
		{
			name:      "BGR565",
			initial:   MakePixelBGR565(0, 0, 0),
			blend:     func(p *basics.Int16u) { BlenderBGR565{}.BlendPix(p, 0, 0, 255, 255, 255) },
			unpack:    UnpackPixelBGR565,
			makeColor: MakeColorBGR565,
		},
		{
			name:      "BGR555Gamma",
			initial:   MakePixelBGR555(0, 0, 0),
			blend:     func(p *basics.Int16u) { NewBlenderBGR555Gamma(trivialGamma{}).BlendPix(p, 0, 0, 255, 255, 255) },
			unpack:    UnpackPixelBGR555,
			makeColor: MakeColorBGR555,
		},
		{
			name:      "BGR565Gamma",
			initial:   MakePixelBGR565(0, 0, 0),
			blend:     func(p *basics.Int16u) { NewBlenderBGR565Gamma(trivialGamma{}).BlendPix(p, 0, 0, 255, 255, 255) },
			unpack:    UnpackPixelBGR565,
			makeColor: MakeColorBGR565,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pixel := tc.initial
			tc.blend(&pixel)
			r, g, b := tc.unpack(pixel)
			if r == 0 && g == 0 && b == 0 {
				t.Fatalf("%s blend left packed pixel black", tc.name)
			}
			got := tc.makeColor(pixel)
			if got.R == 0 && got.G == 0 && got.B == 0 {
				t.Fatalf("%s MakeColor returned black after blend", tc.name)
			}
		})
	}
}
