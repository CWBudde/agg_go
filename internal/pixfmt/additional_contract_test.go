package pixfmt

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
)

type sumMask struct{}

func (sumMask) Calculate(p []basics.Int8u) basics.Int8u {
	if len(p) < 2 {
		return 0
	}
	return p[0] + p[1]
}

func TestPixelTypeSetGetAndSetComponent(t *testing.T) {
	p := PixelType[int]{Components: []int{1, 2, 3}}
	p.Set(9)
	if got := p.Components; got[0] != 9 || got[1] != 9 || got[2] != 9 {
		t.Fatalf("Set() = %v", got)
	}

	p.SetComponent(1, 4)
	p.SetComponent(-1, 7)
	p.SetComponent(99, 8)
	if got := p.Components; got[0] != 9 || got[1] != 4 || got[2] != 9 {
		t.Fatalf("SetComponent() = %v", got)
	}

	if got := p.Get(1); got != 4 {
		t.Fatalf("Get(1) = %d", got)
	}
	if got := p.Get(-1); got != 0 {
		t.Fatalf("Get(-1) = %d", got)
	}
	if got := p.Get(99); got != 0 {
		t.Fatalf("Get(99) = %d", got)
	}
}

func TestAlphaMaskConstructorsAndAccessors(t *testing.T) {
	rbuf := buffer.NewRenderingBufferU8WithData([]basics.Int8u{
		10, 20, 30,
		40, 50, 60,
	}, 3, 2, 3)

	mask := NewAlphaMaskU8WithBuffer(rbuf, 1, 0, OneComponentMaskU8{})
	if mask.Width() != 3 || mask.Height() != 2 {
		t.Fatalf("AlphaMaskU8 size = (%d,%d)", mask.Width(), mask.Height())
	}
	if _, ok := mask.MaskFunction().(OneComponentMaskU8); !ok {
		t.Fatalf("AlphaMaskU8 mask func = %T", mask.MaskFunction())
	}

	noclip := NewAMaskNoClipU8WithBuffer(rbuf, 1, 0, OneComponentMaskU8{})
	if noclip.Width() != 3 || noclip.Height() != 2 {
		t.Fatalf("AMaskNoClipU8 size = (%d,%d)", noclip.Width(), noclip.Height())
	}
	if _, ok := noclip.MaskFunction().(OneComponentMaskU8); !ok {
		t.Fatalf("AMaskNoClipU8 mask func = %T", noclip.MaskFunction())
	}

	gray := NewAlphaMaskGray8()
	if gray.Pixel(0, 0) != 0 {
		t.Fatalf("detached gray mask Pixel() = %d", gray.Pixel(0, 0))
	}

	if _, ok := NewAMaskNoClipRGB24Gray().MaskFunction().(RGBToGrayMaskU8); !ok {
		t.Fatalf("NewAMaskNoClipRGB24Gray returned unexpected mask function")
	}
	if _, ok := NewAlphaMaskBGR24Gray().MaskFunction().(RGBToGrayMaskU8); !ok {
		t.Fatalf("NewAlphaMaskBGR24Gray returned unexpected mask function")
	}
	if _, ok := NewAMaskNoClipBGR24Gray().MaskFunction().(RGBToGrayMaskU8); !ok {
		t.Fatalf("NewAMaskNoClipBGR24Gray returned unexpected mask function")
	}
}

func TestMaskSpanGenericPaths(t *testing.T) {
	src := []basics.Int8u{5, 7, 20, 30, 40}
	dst := make([]basics.Int8u, 3)

	fillMaskSpan(dst, src, 1, 0, 3, sumMask{})
	if dst[0] != 12 || dst[1] != 27 || dst[2] != 50 {
		t.Fatalf("fillMaskSpan generic = %v", dst)
	}

	dst = []basics.Int8u{100, 100, 100}
	combineMaskSpan(dst, src, 1, 0, 3, sumMask{})
	if dst[0] != basics.Int8u((CoverFull+100*12)>>CoverShift) ||
		dst[1] != basics.Int8u((CoverFull+100*27)>>CoverShift) ||
		dst[2] != basics.Int8u((CoverFull+100*50)>>CoverShift) {
		t.Fatalf("combineMaskSpan generic = %v", dst)
	}
}

func TestAMaskNoClipSpanOperations(t *testing.T) {
	rbuf := buffer.NewRenderingBufferU8WithData([]basics.Int8u{
		10, 20, 30,
		40, 50, 60,
		70, 80, 90,
	}, 3, 3, 3)
	mask := NewAMaskNoClipU8WithBuffer(rbuf, 1, 0, OneComponentMaskU8{})

	if got := mask.Pixel(1, 1); got != 50 {
		t.Fatalf("Pixel(1,1) = %d", got)
	}
	if got := mask.CombinePixel(2, 0, 100); got != basics.Int8u((CoverFull+100*30)>>CoverShift) {
		t.Fatalf("CombinePixel(2,0) = %d", got)
	}

	h := []basics.Int8u{1, 1, 1}
	mask.FillHspan(0, 2, h, 3)
	if h[0] != 70 || h[1] != 80 || h[2] != 90 {
		t.Fatalf("FillHspan() = %v", h)
	}

	mask.CombineHspan(0, 0, h, 3)
	if h[0] != basics.Int8u((CoverFull+70*10)>>CoverShift) ||
		h[1] != basics.Int8u((CoverFull+80*20)>>CoverShift) ||
		h[2] != basics.Int8u((CoverFull+90*30)>>CoverShift) {
		t.Fatalf("CombineHspan() = %v", h)
	}

	v := []basics.Int8u{0, 0, 0}
	mask.FillVspan(1, 0, v, 3)
	if v[0] != 20 || v[1] != 50 || v[2] != 80 {
		t.Fatalf("FillVspan() = %v", v)
	}

	mask.CombineVspan(2, 0, v, 3)
	if v[0] != basics.Int8u((CoverFull+20*30)>>CoverShift) ||
		v[1] != basics.Int8u((CoverFull+50*60)>>CoverShift) ||
		v[2] != basics.Int8u((CoverFull+80*90)>>CoverShift) {
		t.Fatalf("CombineVspan() = %v", v)
	}
}
