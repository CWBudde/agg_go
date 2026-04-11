package agg

import (
	"fmt"
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

// ---------------------------------------------------------------------------
// Mock pixel formats for blur variant tests
// ---------------------------------------------------------------------------

type mockRGBA16Image struct {
	pixels []color.RGBA16[color.Linear]
	w, h   int
}

func newMockRGBA16Image(w, h int) *mockRGBA16Image {
	return &mockRGBA16Image{pixels: make([]color.RGBA16[color.Linear], w*h), w: w, h: h}
}

func (m *mockRGBA16Image) Width() int  { return m.w }
func (m *mockRGBA16Image) Height() int { return m.h }
func (m *mockRGBA16Image) Pixel(x, y int) color.RGBA16[color.Linear] {
	return m.pixels[y*m.w+x]
}

func (m *mockRGBA16Image) CopyColorHspan(x, y, length int, colors []color.RGBA16[color.Linear]) {
	copy(m.pixels[y*m.w+x:], colors[:length])
}

func (m *mockRGBA16Image) CopyColorVspan(x, y, length int, colors []color.RGBA16[color.Linear]) {
	for i, c := range colors[:length] {
		m.pixels[(y+i)*m.w+x] = c
	}
}

type mockRGB8Image struct {
	pixels []color.RGB8[color.Linear]
	w, h   int
}

func newMockRGB8Image(w, h int) *mockRGB8Image {
	return &mockRGB8Image{pixels: make([]color.RGB8[color.Linear], w*h), w: w, h: h}
}

func (m *mockRGB8Image) Width() int  { return m.w }
func (m *mockRGB8Image) Height() int { return m.h }
func (m *mockRGB8Image) Pixel(x, y int) color.RGB8[color.Linear] {
	return m.pixels[y*m.w+x]
}

func (m *mockRGB8Image) CopyColorHspan(x, y, length int, colors []color.RGB8[color.Linear]) {
	copy(m.pixels[y*m.w+x:], colors[:length])
}

func (m *mockRGB8Image) CopyColorVspan(x, y, length int, colors []color.RGB8[color.Linear]) {
	for i, c := range colors[:length] {
		m.pixels[(y+i)*m.w+x] = c
	}
}

type mockGray8Image struct {
	pixels []color.Gray8[color.Linear]
	w, h   int
}

func newMockGray8Image(w, h int) *mockGray8Image {
	return &mockGray8Image{pixels: make([]color.Gray8[color.Linear], w*h), w: w, h: h}
}

func (m *mockGray8Image) Width() int  { return m.w }
func (m *mockGray8Image) Height() int { return m.h }
func (m *mockGray8Image) Pixel(x, y int) color.Gray8[color.Linear] {
	return m.pixels[y*m.w+x]
}

func (m *mockGray8Image) CopyColorHspan(x, y, length int, colors []color.Gray8[color.Linear]) {
	copy(m.pixels[y*m.w+x:], colors[:length])
}

func (m *mockGray8Image) CopyColorVspan(x, y, length int, colors []color.Gray8[color.Linear]) {
	for i, c := range colors[:length] {
		m.pixels[(y+i)*m.w+x] = c
	}
}

// ---------------------------------------------------------------------------
// StackBlurRGBA16 tests
// ---------------------------------------------------------------------------

func TestStackBlurRGBA16PreservesOpaque(t *testing.T) {
	img := newMockRGBA16Image(8, 8)
	want := color.RGBA16[color.Linear]{R: 50000, G: 30000, B: 10000, A: 65535}
	for i := range img.pixels {
		img.pixels[i] = want
	}
	sb := NewStackBlurRGBA16[color.Linear]()
	sb.Blur(img, 2)
	for i, got := range img.pixels {
		if got != want {
			t.Fatalf("pixel[%d] = %v, want %v", i, got, want)
		}
	}
}

func TestStackBlurRGBA16Spreads(t *testing.T) {
	img := newMockRGBA16Image(5, 5)
	img.pixels[2*5+2] = color.RGBA16[color.Linear]{R: 65535, G: 0, B: 0, A: 65535}
	sb := NewStackBlurRGBA16[color.Linear]()
	sb.Blur(img, 1)
	centre := img.Pixel(2, 2)
	if centre.R == 65535 {
		t.Fatal("centre should have been blurred down from 65535")
	}
	if centre.R == 0 {
		t.Fatal("centre should still have some red after blur")
	}
}

// ---------------------------------------------------------------------------
// StackBlurRGB tests
// ---------------------------------------------------------------------------

func TestStackBlurRGBPreservesOpaque(t *testing.T) {
	img := newMockRGB8Image(8, 8)
	want := color.RGB8[color.Linear]{R: 200, G: 100, B: 50}
	for i := range img.pixels {
		img.pixels[i] = want
	}
	sb := NewStackBlurRGB[color.Linear]()
	sb.Blur(img, 2)
	for i, got := range img.pixels {
		if got != want {
			t.Fatalf("pixel[%d] = %v, want %v", i, got, want)
		}
	}
}

func TestStackBlurRGBSpreads(t *testing.T) {
	img := newMockRGB8Image(5, 5)
	img.pixels[2*5+2] = color.RGB8[color.Linear]{R: 255, G: 0, B: 0}
	sb := NewStackBlurRGB[color.Linear]()
	sb.Blur(img, 1)
	centre := img.Pixel(2, 2)
	if centre.R == 255 {
		t.Fatal("centre should have been blurred down from 255")
	}
	if centre.R == 0 {
		t.Fatal("centre should still have some red after blur")
	}
}

// ---------------------------------------------------------------------------
// StackBlurGray tests
// ---------------------------------------------------------------------------

func TestStackBlurGrayPreservesOpaque(t *testing.T) {
	img := newMockGray8Image(8, 8)
	want := color.Gray8[color.Linear]{V: 128, A: 255}
	for i := range img.pixels {
		img.pixels[i] = want
	}
	sb := NewStackBlurGray[color.Linear]()
	sb.Blur(img, 2)
	for i, got := range img.pixels {
		if got != want {
			t.Fatalf("pixel[%d] = %v, want %v", i, got, want)
		}
	}
}

func TestStackBlurGraySpreads(t *testing.T) {
	img := newMockGray8Image(5, 5)
	img.pixels[2*5+2] = color.Gray8[color.Linear]{V: 255, A: 255}
	sb := NewStackBlurGray[color.Linear]()
	sb.Blur(img, 1)
	centre := img.Pixel(2, 2)
	if centre.V == 255 {
		t.Fatal("centre should have been blurred down from 255")
	}
	if centre.V == 0 {
		t.Fatal("centre should still have some value after blur")
	}
}

// ---------------------------------------------------------------------------
// Benchmarks — typed API
// ---------------------------------------------------------------------------

func BenchmarkStackBlurRGBA16(b *testing.B) {
	for _, size := range []int{64, 256} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			img := newMockRGBA16Image(size, size)
			want := color.RGBA16[color.Linear]{R: 30000, G: 20000, B: 10000, A: 65535}
			for i := range img.pixels {
				img.pixels[i] = want
			}
			sb := NewStackBlurRGBA16[color.Linear]()
			b.ResetTimer()
			for range b.N {
				sb.Blur(img, 5)
			}
		})
	}
}

func BenchmarkStackBlurRGB(b *testing.B) {
	for _, size := range []int{64, 256} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			img := newMockRGB8Image(size, size)
			want := color.RGB8[color.Linear]{R: 200, G: 100, B: 50}
			for i := range img.pixels {
				img.pixels[i] = want
			}
			sb := NewStackBlurRGB[color.Linear]()
			b.ResetTimer()
			for range b.N {
				sb.Blur(img, 5)
			}
		})
	}
}

func BenchmarkStackBlurGray(b *testing.B) {
	for _, size := range []int{64, 256} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			img := newMockGray8Image(size, size)
			want := color.Gray8[color.Linear]{V: 128, A: 255}
			for i := range img.pixels {
				img.pixels[i] = want
			}
			sb := NewStackBlurGray[color.Linear]()
			b.ResetTimer()
			for range b.N {
				sb.Blur(img, 5)
			}
		})
	}
}
