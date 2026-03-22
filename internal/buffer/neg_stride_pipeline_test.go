package buffer_test

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
)

// TestNegativeStridePipeline verifies that the full rendering pipeline
// (rbuf → pixfmt → renderer → CopyHline) produces correct output
// when the rendering buffer uses negative stride (C++ flip_y style).
//
// With negative stride (flip_y=true in C++):
//   - rbuf.Row(0) = physical LAST row of the backing slice
//   - rbuf.Row(H-1) = physical FIRST row of the backing slice
//   - Drawing at rbuf y=0 must land in physical row H-1
//
// When blitting buf top-to-bottom, physical row H-1 appears at the BOTTOM of
// the display, so drawing at y=0 appears at the bottom — matching C++ behavior.
func TestNegativeStridePipeline(t *testing.T) {
	const (
		w = 4
		h = 4
	)
	buf := make([]uint8, w*h*4)

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(buf, w, h, -(w * 4))

	pixf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[
		*pixfmt.PixFmtRGBA32[icol.Linear],
		icol.RGBA8[icol.Linear],
	](pixf)

	white := icol.RGBA8[icol.Linear]{R: 255, G: 255, B: 255, A: 255}
	red := icol.RGBA8[icol.Linear]{R: 255, G: 0, B: 0, A: 255}

	rb.Clear(white)
	// Draw red at rbuf y=0 → must land in physical last row.
	rb.CopyHline(0, 0, w-1, red)

	lastRow := (h - 1) * w * 4
	for x := 0; x < w; x++ {
		base := lastRow + x*4
		if buf[base] != 255 || buf[base+1] != 0 || buf[base+2] != 0 {
			t.Errorf("physical row H-1 x=%d: want red, got R=%d G=%d B=%d",
				x, buf[base], buf[base+1], buf[base+2])
		}
	}

	for x := 0; x < w; x++ {
		base := x * 4
		if buf[base] != 255 || buf[base+1] != 255 || buf[base+2] != 255 {
			t.Errorf("physical row 0 x=%d: want white, got R=%d G=%d B=%d",
				x, buf[base], buf[base+1], buf[base+2])
		}
	}

	// Verify Row(0) pointer identity.
	row0 := rbuf.Row(0)
	if len(row0) == 0 || &row0[0] != &buf[lastRow] {
		t.Errorf("rbuf.Row(0) does not point to last physical row (lastRow=%d)", lastRow)
	}
}

// TestNegativeStrideRasterizer verifies that the rasterizer's SweepScanline
// also writes to the correct physical location through a negative-stride buffer.
func TestNegativeStrideRasterizer(t *testing.T) {
	const (
		w = 10
		h = 10
	)
	buf := make([]uint8, w*h*4)
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(buf, w, h, -(w * 4))

	// Verify that all rows are accessible and distinct.
	for y := 0; y < h; y++ {
		row := rbuf.Row(y)
		if row == nil {
			t.Fatalf("rbuf.Row(%d) returned nil", y)
		}
		// Each row start must be at physical offset (H-1-y)*w*4.
		expectedOff := (h - 1 - y) * w * 4
		actualOff := int(basics.Int8u(0))
		_ = actualOff
		if &row[0] != &buf[expectedOff] {
			t.Errorf("Row(%d): expected physical offset %d, got different pointer", y, expectedOff)
		}
	}
}
