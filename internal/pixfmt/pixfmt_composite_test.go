package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt/blender"
)

type rgbaRowSource struct {
	pixels []color.RGBA8[color.Linear]
}

func (s *rgbaRowSource) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if y != 0 || x < 0 || x >= len(s.pixels) {
		return color.RGBA8[color.Linear]{}
	}
	return s.pixels[x]
}

func (s *rgbaRowSource) Width() int  { return len(s.pixels) }
func (s *rgbaRowSource) Height() int { return 1 }

type rgbaRowDataSource struct {
	pixels []color.RGBA8[color.Linear]
	row    []basics.Int8u
}

func newRGBARowDataSource(pixels []color.RGBA8[color.Linear]) *rgbaRowDataSource {
	row := make([]basics.Int8u, len(pixels)*4)
	for i, px := range pixels {
		off := i * 4
		row[off+0] = px.R
		row[off+1] = px.G
		row[off+2] = px.B
		row[off+3] = px.A
	}
	return &rgbaRowDataSource{pixels: pixels, row: row}
}

func (s *rgbaRowDataSource) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if y != 0 || x < 0 || x >= len(s.pixels) {
		return color.RGBA8[color.Linear]{}
	}
	return s.pixels[x]
}

func (s *rgbaRowDataSource) Width() int  { return len(s.pixels) }
func (s *rgbaRowDataSource) Height() int { return 1 }
func (s *rgbaRowDataSource) RowData(y int) []basics.Int8u {
	if y != 0 {
		return nil
	}
	return s.row
}

func pixelBytes(buf []basics.Int8u, x int) []basics.Int8u {
	off := x * 4
	return buf[off : off+4]
}

func TestPixFmtCompositeRGBA32PreBlendFrom(t *testing.T) {
	buf := make([]basics.Int8u, 3*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 3, 1, 3*4)
	pf := NewPixFmtCompositeRGBA32Pre(rbuf, blender.CompOpSrcOver)

	src := &rgbaRowSource{
		pixels: []color.RGBA8[color.Linear]{
			{R: 128, G: 0, B: 0, A: 128},
			{R: 0, G: 128, B: 0, A: 128},
			{R: 0, G: 0, B: 128, A: 128},
		},
	}

	pf.BlendFrom(src, 0, 0, 0, 0, 3, basics.CoverFull)

	for x, want := range src.pixels {
		if got := pf.GetPixel(x, 0); got != want {
			t.Fatalf("pixel %d: got %+v want %+v", x, got, want)
		}
	}
}

func TestPixFmtCompositeRGBA32BlendPixelCompositeOps(t *testing.T) {
	tests := []struct {
		name string
		op   blender.CompOp
		dst  [4]basics.Int8u
		src  color.RGBA8[color.Linear]
		want [4]basics.Int8u
	}{
		{
			name: "clear",
			op:   blender.CompOpClear,
			dst:  [4]basics.Int8u{255, 0, 0, 255},
			src:  color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255},
			want: [4]basics.Int8u{0, 0, 0, 0},
		},
		{
			name: "src",
			op:   blender.CompOpSrc,
			dst:  [4]basics.Int8u{255, 0, 0, 255},
			src:  color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255},
			want: [4]basics.Int8u{0, 255, 0, 255},
		},
		{
			name: "dst",
			op:   blender.CompOpDst,
			dst:  [4]basics.Int8u{255, 0, 0, 255},
			src:  color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255},
			want: [4]basics.Int8u{255, 0, 0, 255},
		},
		{
			name: "dst_out",
			op:   blender.CompOpDstOut,
			dst:  [4]basics.Int8u{128, 0, 0, 128},
			src:  color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 128},
			want: [4]basics.Int8u{64, 0, 0, 64},
		},
		{
			name: "xor",
			op:   blender.CompOpXor,
			dst:  [4]basics.Int8u{128, 0, 0, 128},
			src:  color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 128},
			want: [4]basics.Int8u{64, 64, 0, 127},
		},
		{
			name: "multiply",
			op:   blender.CompOpMultiply,
			dst:  [4]basics.Int8u{255, 0, 0, 255},
			src:  color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255},
			want: [4]basics.Int8u{0, 0, 0, 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := []basics.Int8u{tt.dst[0], tt.dst[1], tt.dst[2], tt.dst[3]}
			rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
			pf := NewPixFmtCompositeRGBA32(rbuf, tt.op)

			pf.BlendPixel(0, 0, tt.src, basics.CoverFull)

			got := [4]basics.Int8u{buf[0], buf[1], buf[2], buf[3]}
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestPixFmtCompositeRGBA32PreBlendFromCompositeOpsMatchManualBlend(t *testing.T) {
	tests := []struct {
		name  string
		op    blender.CompOp
		src   []color.RGBA8[color.Linear]
		dst   []basics.Int8u
		cover basics.Int8u
	}{
		{
			name: "src",
			op:   blender.CompOpSrc,
			src: []color.RGBA8[color.Linear]{
				{R: 64, G: 0, B: 0, A: 64},
				{R: 0, G: 96, B: 0, A: 96},
				{R: 0, G: 0, B: 128, A: 128},
			},
			dst:   []basics.Int8u{10, 20, 30, 40, 40, 50, 60, 70, 70, 80, 90, 100},
			cover: basics.CoverFull,
		},
		{
			name: "dst",
			op:   blender.CompOpDst,
			src: []color.RGBA8[color.Linear]{
				{R: 64, G: 0, B: 0, A: 64},
				{R: 0, G: 96, B: 0, A: 96},
				{R: 0, G: 0, B: 128, A: 128},
			},
			dst:   []basics.Int8u{10, 20, 30, 40, 40, 50, 60, 70, 70, 80, 90, 100},
			cover: basics.CoverFull,
		},
		{
			name: "xor_partial_cover",
			op:   blender.CompOpXor,
			src: []color.RGBA8[color.Linear]{
				{R: 128, G: 0, B: 0, A: 128},
				{R: 0, G: 128, B: 0, A: 128},
				{R: 0, G: 0, B: 128, A: 128},
			},
			dst:   []basics.Int8u{90, 20, 10, 140, 15, 100, 25, 150, 30, 35, 110, 160},
			cover: 200,
		},
		{
			name: "plus_partial_cover",
			op:   blender.CompOpPlus,
			src: []color.RGBA8[color.Linear]{
				{R: 32, G: 16, B: 0, A: 64},
				{R: 0, G: 32, B: 16, A: 64},
				{R: 16, G: 0, B: 32, A: 64},
			},
			dst:   []basics.Int8u{100, 90, 80, 120, 50, 40, 30, 60, 10, 20, 30, 40},
			cover: 180,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := append([]basics.Int8u(nil), tt.dst...)
			bl := blender.NewCompositeBlenderPre[color.Linear, order.RGBA](tt.op)
			for i, srcPx := range tt.src {
				bl.BlendPix(pixelBytes(expected, i), srcPx.R, srcPx.G, srcPx.B, srcPx.A, tt.cover)
			}

			for _, src := range []struct {
				name string
				src  interface {
					GetPixel(x, y int) color.RGBA8[color.Linear]
					Width() int
					Height() int
				}
			}{
				{name: "getpixel", src: &rgbaRowSource{pixels: tt.src}},
				{name: "rowdata", src: newRGBARowDataSource(tt.src)},
			} {
				t.Run(src.name, func(t *testing.T) {
					buf := append([]basics.Int8u(nil), tt.dst...)
					rbuf := buffer.NewRenderingBufferU8WithData(buf, len(tt.src), 1, len(tt.src)*4)
					pf := NewPixFmtCompositeRGBA32Pre(rbuf, tt.op)

					pf.BlendFrom(src.src, 0, 0, 0, 0, len(tt.src), tt.cover)

					for i := range tt.src {
						got := [4]basics.Int8u{buf[i*4+0], buf[i*4+1], buf[i*4+2], buf[i*4+3]}
						want := [4]basics.Int8u{
							expected[i*4+0], expected[i*4+1], expected[i*4+2], expected[i*4+3],
						}
						if got != want {
							t.Fatalf("pixel %d: got %v want %v", i, got, want)
						}
					}
				})
			}
		})
	}
}

func TestPixFmtCompositeRGBA32SetCompOpAffectsSubsequentBlends(t *testing.T) {
	buf := []basics.Int8u{20, 30, 40, 50}
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpDst)

	src := color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255}
	pf.BlendPixel(0, 0, src, basics.CoverFull)
	if got := [4]basics.Int8u{buf[0], buf[1], buf[2], buf[3]}; got != [4]basics.Int8u{20, 30, 40, 50} {
		t.Fatalf("CompOpDst should preserve destination, got %v", got)
	}

	pf.SetCompOp(blender.CompOpSrc)
	pf.BlendPixel(0, 0, src, basics.CoverFull)
	if got := [4]basics.Int8u{buf[0], buf[1], buf[2], buf[3]}; got != [4]basics.Int8u{0, 255, 0, 255} {
		t.Fatalf("SetCompOp should switch to source replacement, got %v", got)
	}
}
