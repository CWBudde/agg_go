package pixfmt

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
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
		// The straight-source composite pixfmt treats its buffer as straight
		// alpha (Agg2D's convention): dst {128,0,0,128} is straight red at 50%
		// alpha. dst-out keeps the destination colour and scales only its alpha
		// by (1-Sa), so the red stays 128 and the alpha halves to 64. xor
		// likewise demultiplies back to straight green.
		{
			name: "dst_out",
			op:   blender.CompOpDstOut,
			dst:  [4]basics.Int8u{128, 0, 0, 128},
			src:  color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 128},
			want: [4]basics.Int8u{128, 0, 0, 64},
		},
		{
			name: "xor",
			op:   blender.CompOpXor,
			dst:  [4]basics.Int8u{128, 0, 0, 128},
			src:  color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 128},
			want: [4]basics.Int8u{64, 127, 0, 127},
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

func TestPixFmtCompositeRGBA32PixWidth(t *testing.T) {
	buf := make([]basics.Int8u, 4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrcOver)
	if pf.PixWidth() != 4 {
		t.Errorf("PixWidth() expected 4, got %d", pf.PixWidth())
	}
}

func TestPixFmtCompositeRGBA32Pixel(t *testing.T) {
	buf := make([]basics.Int8u, 4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)
	c := color.RGBA8[color.Linear]{R: 10, G: 20, B: 30, A: 255}
	pf.CopyPixel(0, 0, c)
	if got := pf.GetPixel(0, 0); got != c {
		t.Errorf("GetPixel after CopyPixel: got %+v want %+v", got, c)
	}
}

func TestPixFmtCompositeRGBA32GetCompOp(t *testing.T) {
	buf := make([]basics.Int8u, 4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrcOver)
	if pf.GetCompOp() != blender.CompOpSrcOver {
		t.Errorf("GetCompOp() expected CompOpSrcOver")
	}
}

func TestPixFmtCompositeRGBA32BlendHline(t *testing.T) {
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 4, 1, 4*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	c := color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 255}
	pf.BlendHline(0, 0, 4, c, basics.CoverFull)

	for x := 0; x < 4; x++ {
		if got := pf.GetPixel(x, 0); got != c {
			t.Errorf("BlendHline x=%d: got %+v want %+v", x, got, c)
		}
	}
}

func TestPixFmtCompositeRGBA32BlendVline(t *testing.T) {
	buf := make([]basics.Int8u, 1*4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 4, 1*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	c := color.RGBA8[color.Linear]{R: 50, G: 100, B: 150, A: 255}
	pf.BlendVline(0, 0, 4, c, basics.CoverFull)

	for y := 0; y < 4; y++ {
		if got := pf.GetPixel(0, y); got != c {
			t.Errorf("BlendVline y=%d: got %+v want %+v", y, got, c)
		}
	}
}

func TestPixFmtCompositeRGBA32BlendSolidHspan(t *testing.T) {
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 4, 1, 4*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	c := color.RGBA8[color.Linear]{R: 200, G: 100, B: 50, A: 255}
	covers := []basics.Int8u{255, 255, 128, 0}
	pf.BlendSolidHspan(0, 0, 4, c, covers)

	if got := pf.GetPixel(0, 0); got != c {
		t.Errorf("BlendSolidHspan full cover x=0: got %+v want %+v", got, c)
	}
	// zero cover → no change
	if got := pf.GetPixel(3, 0); got.R != 0 {
		t.Errorf("BlendSolidHspan zero cover x=3: should be zero, got %+v", got)
	}
}

func TestPixFmtCompositeRGBA32CopyHlineVline(t *testing.T) {
	buf := make([]basics.Int8u, 4*4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 4, 4, 4*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	c := color.RGBA8[color.Linear]{R: 77, G: 88, B: 99, A: 255}
	pf.CopyHline(0, 1, 4, c)
	for x := 0; x < 4; x++ {
		if got := pf.GetPixel(x, 1); got != c {
			t.Errorf("CopyHline x=%d: got %+v want %+v", x, got, c)
		}
	}

	pf.CopyVline(2, 0, 4, c)
	for y := 0; y < 4; y++ {
		if got := pf.GetPixel(2, y); got != c {
			t.Errorf("CopyVline y=%d: got %+v want %+v", y, got, c)
		}
	}
}

func TestPixFmtCompositeRGBA32CopyAndBlendBar(t *testing.T) {
	buf := make([]basics.Int8u, 6*6*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 6, 6, 6*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	c := color.RGBA8[color.Linear]{R: 111, G: 222, B: 33, A: 255}
	pf.CopyBar(1, 1, 4, 4, c)

	for y := 1; y <= 4; y++ {
		for x := 1; x <= 4; x++ {
			if got := pf.GetPixel(x, y); got != c {
				t.Errorf("CopyBar (%d,%d): got %+v want %+v", x, y, got, c)
			}
		}
	}

	d := color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255}
	pf.BlendBar(1, 1, 2, 2, d, basics.CoverFull)
	if got := pf.GetPixel(1, 1); got != d {
		t.Errorf("BlendBar (%d,%d): got %+v want %+v", 1, 1, got, d)
	}
}

func TestPixFmtCompositeRGBA32CopyColorSpans(t *testing.T) {
	buf := make([]basics.Int8u, 3*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 3, 1, 3*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	colors := []color.RGBA8[color.Linear]{
		{R: 10, A: 255}, {R: 20, A: 255}, {R: 30, A: 255},
	}
	pf.CopyColorHspan(0, 0, 3, colors)
	for x, want := range colors {
		if got := pf.GetPixel(x, 0); got != want {
			t.Errorf("CopyColorHspan x=%d: got %+v want %+v", x, got, want)
		}
	}

	buf2 := make([]basics.Int8u, 1*3*4)
	rbuf2 := buffer.NewRenderingBufferU8WithData(buf2, 1, 3, 1*4)
	pf2 := NewPixFmtCompositeRGBA32(rbuf2, blender.CompOpSrc)
	pf2.CopyColorVspan(0, 0, 3, colors)
	for y, want := range colors {
		if got := pf2.GetPixel(0, y); got != want {
			t.Errorf("CopyColorVspan y=%d: got %+v want %+v", y, got, want)
		}
	}
}

func TestPixFmtCompositeRGBA32BlendColorSpans(t *testing.T) {
	buf := make([]basics.Int8u, 3*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 3, 1, 3*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	colors := []color.RGBA8[color.Linear]{
		{R: 50, A: 255}, {R: 100, A: 255}, {R: 150, A: 255},
	}
	pf.BlendColorHspan(0, 0, 3, colors, nil, basics.CoverFull)
	for x, want := range colors {
		if got := pf.GetPixel(x, 0); got != want {
			t.Errorf("BlendColorHspan x=%d: got %+v want %+v", x, got, want)
		}
	}

	buf2 := make([]basics.Int8u, 1*3*4)
	rbuf2 := buffer.NewRenderingBufferU8WithData(buf2, 1, 3, 1*4)
	pf2 := NewPixFmtCompositeRGBA32(rbuf2, blender.CompOpSrc)
	pf2.BlendColorVspan(0, 0, 3, colors, nil, basics.CoverFull)
	for y, want := range colors {
		if got := pf2.GetPixel(0, y); got != want {
			t.Errorf("BlendColorVspan y=%d: got %+v want %+v", y, got, want)
		}
	}
}

func TestPixFmtCompositeRGBA32SolidVspan(t *testing.T) {
	buf := make([]basics.Int8u, 1*4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 4, 1*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	c := color.RGBA8[color.Linear]{R: 180, G: 90, B: 45, A: 255}
	pf.BlendSolidVspan(0, 0, 4, c, nil)
	for y := 0; y < 4; y++ {
		if got := pf.GetPixel(0, y); got != c {
			t.Errorf("BlendSolidVspan y=%d: got %+v want %+v", y, got, c)
		}
	}
}

func TestPixFmtCompositeRGBA32ClearAndFill(t *testing.T) {
	buf := make([]basics.Int8u, 2*2*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 2, 2, 2*4)
	pf := NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrc)

	c := color.RGBA8[color.Linear]{R: 55, G: 66, B: 77, A: 255}
	pf.Clear(c)
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if got := pf.GetPixel(x, y); got != c {
				t.Errorf("Clear (%d,%d): got %+v want %+v", x, y, got, c)
			}
		}
	}

	zero := color.RGBA8[color.Linear]{}
	pf.Fill(zero)
	if got := pf.GetPixel(0, 0); got != zero {
		t.Errorf("Fill zero: got %+v", got)
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

// TestPixFmtCompositeRGBA32FastPathMatchesScalar locks the BlendSolidHspan /
// BlendHline fast-path wiring (clip + cover-slice alignment) for the operators
// that have a plainSpanKernel (src-over, xor). For each case it renders through
// the pixfmt (which takes the fast path) and compares against an independent
// scalar reference computed straight from CompositeBlenderPlain, asserting a
// byte-for-byte match. Destinations are seeded translucent to exercise the
// straight<->premult bridge the premult-dst kernels got wrong.
func TestPixFmtCompositeRGBA32FastPathMatchesScalar(t *testing.T) {
	const width = 24
	ops := []struct {
		name string
		op   blender.CompOp
	}{
		{"src_over", blender.CompOpSrcOver},
		{"xor", blender.CompOpXor},
		{"dst_out", blender.CompOpDstOut},
		{"clear", blender.CompOpClear},
		{"multiply", blender.CompOpMultiply},
		{"src_in", blender.CompOpSrcIn},
	}
	src := color.RGBA8[color.Linear]{R: 40, G: 60, B: 220, A: 160}

	seed := func() []basics.Int8u {
		buf := make([]basics.Int8u, width*4)
		for i := 0; i < width; i++ {
			p := i * 4
			buf[p+0] = basics.Int8u(20 + i*7)
			buf[p+1] = basics.Int8u(200 - i*5)
			buf[p+2] = basics.Int8u(60 + i*3)
			buf[p+3] = basics.Int8u(90 + i*6) // translucent destination
		}
		return buf
	}

	ramp := make([]basics.Int8u, width)
	for i := range ramp {
		ramp[i] = basics.Int8u((i * 37) % 256)
	}

	type spanCase struct {
		name    string
		x, n    int
		covers  []basics.Int8u
		isHline bool
	}
	cases := []spanCase{
		{"hspan_full_x0", 0, width, nil, false},
		{"hspan_full_xoff", 5, 16, nil, false},
		{"hspan_clip_negx", -3, 16, nil, false},
		{"hspan_ramp_x0", 0, width, ramp, false},
		{"hspan_ramp_xoff", 4, 18, ramp, false},
		{"hline_full_x0", 0, width, nil, true},
		{"hline_clip_negx", -2, 20, nil, true},
	}

	for _, opc := range ops {
		op := opc.op
		for _, tc := range cases {
			t.Run(opc.name+"/"+tc.name, func(t *testing.T) {
				// Fast path: render through the pixfmt.
				fpBuf := seed()
				rbuf := buffer.NewRenderingBufferU8WithData(fpBuf, width, 1, width*4)
				pf := NewPixFmtCompositeRGBA32(rbuf, op)
				if tc.isHline {
					pf.BlendHline(tc.x, 0, tc.n, src, basics.CoverFull)
				} else {
					pf.BlendSolidHspan(tc.x, 0, tc.n, src, tc.covers)
				}

				// Scalar reference: same clip + covers indexing, straight from the blender.
				refBuf := seed()
				bl := blender.NewCompositeBlenderPlain[color.Linear, order.RGBA](op)
				startX := max(0, tc.x)
				endX := min(tc.x+tc.n, width)
				for i := startX; i < endX; i++ {
					cover := basics.Int8u(basics.CoverFull)
					if !tc.isHline && tc.covers != nil {
						ci := i - tc.x
						if ci >= len(tc.covers) {
							continue
						}
						cover = tc.covers[ci]
					}
					bl.BlendPix(refBuf[i*4:i*4+4], src.R, src.G, src.B, src.A, cover)
				}

				for i := 0; i < width*4; i++ {
					if fpBuf[i] != refBuf[i] {
						t.Fatalf("px=%d byte=%d: fast-path %d != scalar %d", i/4, i%4, fpBuf[i], refBuf[i])
					}
				}
			})
		}
	}
}

// TestPixFmtCompositeRGBA32StraightOverOpaqueFaithful is the end-to-end (through
// the pixfmt's wired fast path) regression guard for the §5.5 premultiplied-
// storage bug: a translucent source xor'd / dst-out over an opaque destination
// must read back as a STRAIGHT (demultiplied) value, not premultiplied. The bug
// would store ~(15,48,22,95) for xor; the faithful straight result is (40,130,60,95).
// Covers both BlendSolidHspan and BlendHline.
func TestPixFmtCompositeRGBA32StraightOverOpaqueFaithful(t *testing.T) {
	src := color.RGBA8[color.Linear]{R: 40, G: 60, B: 220, A: 160}
	want := color.RGBA8[color.Linear]{R: 40, G: 130, B: 60, A: 95}
	ops := []struct {
		name string
		op   blender.CompOp
	}{
		{"xor", blender.CompOpXor},
		{"dst_out", blender.CompOpDstOut},
	}
	for _, oc := range ops {
		for _, path := range []string{"hspan", "hline"} {
			t.Run(oc.name+"/"+path, func(t *testing.T) {
				buf := []basics.Int8u{40, 130, 60, 255} // 1px opaque green
				rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
				pf := NewPixFmtCompositeRGBA32(rbuf, oc.op)
				if path == "hspan" {
					pf.BlendSolidHspan(0, 0, 1, src, nil)
				} else {
					pf.BlendHline(0, 0, 1, src, basics.CoverFull)
				}
				if got := pf.GetPixel(0, 0); got != want {
					t.Errorf("%s/%s: got %+v want %+v (premultiplied-storage bug would give ~{15,48,22,95})",
						oc.name, path, got, want)
				}
			})
		}
	}
}
