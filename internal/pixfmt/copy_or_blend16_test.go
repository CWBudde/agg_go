package pixfmt

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
)

// C++ copy_or_blend_pix stores an opaque colour at full cover verbatim, so an
// opaque white over black must stay 65535 even though the C++-exact 16-bit lerp
// yields lerp(0, 65535, 65535) == 65534.
func TestCopyOrBlendPixOpaque16(t *testing.T) {
	const w = 4
	white16 := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0xFFFF, B: 0xFFFF, A: 0xFFFF}
	full := []basics.Int8u{255, 255, 255, 255}

	rgba := func() *PixFmtRGBA64[color.Linear] {
		data := make([]basics.Int8u, w*8)
		return NewPixFmtRGBA64Linear(buffer.NewRenderingBufferU8WithData(data, w, 1, w*8))
	}
	checkRGBA := func(name string, pf *PixFmtRGBA64[color.Linear], n int) {
		t.Helper()
		for x := 0; x < n; x++ {
			if got := pf.Pixel(x, 0); got != white16 {
				t.Errorf("%s: pixel %d = %+v, want %+v", name, x, got, white16)
			}
		}
	}

	pf := rgba()
	pf.BlendSolidHspan(0, 0, w, white16, full)
	checkRGBA("rgba BlendSolidHspan covers", pf, w)

	pf = rgba()
	pf.BlendSolidHspan(0, 0, w, white16, nil)
	checkRGBA("rgba BlendSolidHspan nil covers", pf, w)

	pf = rgba()
	pf.BlendPixel(0, 0, white16, 255)
	checkRGBA("rgba BlendPixel", pf, 1)

	gray := func() *PixFmtGray16 {
		data := make([]basics.Int16u, w)
		return NewPixFmtGray16(buffer.NewRenderingBufferU16WithData(data, w, 1, w*2))
	}
	g := color.Gray16[color.Linear]{V: 0xFFFF, A: 0xFFFF}
	full16 := []basics.Int16u{0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF}
	for name, draw := range map[string]func(*PixFmtGray16){
		"BlendPixel":      func(p *PixFmtGray16) { p.BlendPixel(0, 0, g, 0xFFFF) },
		"BlendHline":      func(p *PixFmtGray16) { p.BlendHline(0, 0, 1, g, 0xFFFF) },
		"BlendSolidHspan": func(p *PixFmtGray16) { p.BlendSolidHspan(0, 0, 1, g, full16) },
	} {
		p := gray()
		draw(p)
		if got := p.Pixel(0, 0).V; got != 0xFFFF {
			t.Errorf("gray16 %s: V = %d, want 65535", name, got)
		}
	}
}
