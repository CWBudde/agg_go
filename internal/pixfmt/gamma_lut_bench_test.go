package pixfmt

import (
	"math"
	"strconv"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
)

type benchmarkGrayRowSource struct {
	row []basics.Int8u
}

func (s *benchmarkGrayRowSource) RowData(y int) []basics.Int8u {
	if y != 0 {
		return nil
	}
	return s.row
}

func (s *benchmarkGrayRowSource) Width() int  { return len(s.row) }
func (s *benchmarkGrayRowSource) Height() int { return 1 }

func BenchmarkPixFmtRGBA32ApplyGammaDir(b *testing.B) {
	var dirTable [256]basics.Int8u
	for i := range dirTable {
		v := math.Pow(float64(i)/255.0, 2.2)
		dirTable[i] = basics.Int8u(v*255.0 + 0.5)
	}

	sizes := []struct {
		name          string
		width, height int
	}{
		{name: "256x256", width: 256, height: 256},
		{name: "1024x1024", width: 1024, height: 1024},
	}

	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			buf := make([]basics.Int8u, tc.width*tc.height*4)
			for i := range buf {
				buf[i] = basics.Int8u((i*29 + 17) & 0xff)
			}
			rbuf := buffer.NewRenderingBufferU8WithData(buf, tc.width, tc.height, tc.width*4)
			pf := NewPixFmtRGBA32[color.Linear](rbuf)

			b.ReportAllocs()
			b.SetBytes(int64(len(buf)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pf.ApplyGammaDir(func(v basics.Int8u) basics.Int8u { return dirTable[v] })
			}
		})
	}
}

func BenchmarkPixFmtRGBA32BlendFromLUT(b *testing.B) {
	for _, length := range []int{64, 256, 1024, 4096} {
		b.Run("Len_"+strconv.Itoa(length), func(b *testing.B) {
			dst := make([]basics.Int8u, length*4)
			rbuf := buffer.NewRenderingBufferU8WithData(dst, length, 1, length*4)
			pf := NewPixFmtRGBA32[color.Linear](rbuf)

			srcRow := make([]basics.Int8u, length)
			for i := range srcRow {
				srcRow[i] = basics.Int8u((i*37 + 11) & 0xff)
			}
			src := &benchmarkGrayRowSource{row: srcRow}

			colorLUT := make([]color.RGBA8[color.Linear], 256)
			for i := range colorLUT {
				colorLUT[i] = color.RGBA8[color.Linear]{
					R: basics.Int8u(i),
					G: basics.Int8u(255 - i),
					B: basics.Int8u((i * 3) & 0xff),
					A: 255,
				}
			}

			b.ReportAllocs()
			b.SetBytes(int64(length * 5))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pf.BlendFromLUT(src, colorLUT, 0, 0, 0, 0, length, basics.CoverFull)
			}
		})
	}
}
