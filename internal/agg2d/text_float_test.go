package agg2d

import "testing"

// renderText8 runs a text scene into a straight-RGBA8 buffer.
func renderText8(w, h int, scene func(*Agg2D)) []uint8 {
	a := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a.Attach(buf, w, h, w*4)
	scene(a)
	return buf
}

// renderTextFloat runs a text scene into a float image.
func renderTextFloat(w, h int, scene func(*Agg2DFloat)) *ImageFloat {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	a.AttachImageFloat(img)
	scene(a)
	return img
}

// inkCount8 counts pixels that differ from the (opaque) background, a proxy for
// "did we actually draw glyphs".
func inkCount8(buf []uint8, w, h int, bg [4]int) int {
	n := 0
	for y := range h {
		for x := range w {
			c := pixel8(buf, w, x, y)
			if c != bg {
				n++
			}
		}
	}
	return n
}

// maxImageDiff returns the largest per-channel difference between the 8-bit
// buffer and the float image across the whole frame.
func maxImageDiff(buf []uint8, img *ImageFloat, w, h int) (maxDiff, samples int) {
	for y := range h {
		for x := range w {
			c8 := pixel8(buf, w, x, y)
			cf := pixelFloatAsU8(img, x, y)
			if d := maxChanDiff(c8, cf); d > maxDiff {
				maxDiff = d
			}
			samples++
		}
	}
	return maxDiff, samples
}

// TestParityTextGSV verifies the built-in GSV stroke font renders identically
// through the 8-bit and float pipelines (no cgo / FreeType needed).
func TestParityTextGSV(t *testing.T) {
	const w, h = 96, 32
	bg := NewColor(255, 255, 255, 255)
	fg := NewColor(20, 40, 160, 255)

	scene8 := func(a *Agg2D) {
		a.ClearAll(bg)
		a.FontGSV(18)
		a.FillColor(fg)
		a.Text(4, 22, "Hi!", false, 0, 0)
	}
	sceneF := func(a *Agg2DFloat) {
		a.ClearAll(bg)
		a.FontGSV(18)
		a.FillColor(fg)
		a.Text(4, 22, "Hi!", false, 0, 0)
	}

	buf := renderText8(w, h, scene8)
	img := renderTextFloat(w, h, sceneF)

	if ink := inkCount8(buf, w, h, [4]int{255, 255, 255, 255}); ink < 20 {
		t.Fatalf("GSV text rendered too little ink in 8-bit reference: %d pixels", ink)
	}

	maxDiff, _ := maxImageDiff(buf, img, w, h)
	if maxDiff > 2 {
		t.Errorf("GSV text float/8-bit max channel diff = %d (tol 2)", maxDiff)
	}

	// The float image must also contain ink (guards against a silently blank float path).
	inkF := 0
	for y := range h {
		for x := range w {
			if pixelFloatAsU8(img, x, y) != [4]int{255, 255, 255, 255} {
				inkF++
			}
		}
	}
	if inkF < 20 {
		t.Errorf("GSV text rendered too little ink in float path: %d pixels", inkF)
	}
}

// TestParityTextFreeTypeOutline verifies FreeType vector (outline) glyphs render
// to parity between the 8-bit and float pipelines. Outline glyphs flow through
// the shared rasterizer/scanline fill+stroke path.
func TestParityTextFreeTypeOutline(t *testing.T) {
	fontPath := findSystemFont()
	if fontPath == "" {
		t.Skip("no system font available")
	}

	const w, h = 120, 40
	bg := NewColor(255, 255, 255, 255)
	fg := NewColor(0, 0, 0, 255)

	scene8 := func(a *Agg2D) {
		a.ClearAll(bg)
		if err := a.Font(fontPath, 24.0, false, false, VectorFontCache, 0.0); err != nil {
			t.Skipf("FreeType unavailable: %v", err)
		}
		a.FillColor(fg)
		a.LineColor(fg)
		a.Text(6, 30, "Age", false, 0, 0)
	}
	sceneF := func(a *Agg2DFloat) {
		a.ClearAll(bg)
		if err := a.Font(fontPath, 24.0, false, false, VectorFontCache, 0.0); err != nil {
			t.Skipf("FreeType unavailable: %v", err)
		}
		a.FillColor(fg)
		a.LineColor(fg)
		a.Text(6, 30, "Age", false, 0, 0)
	}

	buf := renderText8(w, h, scene8)
	img := renderTextFloat(w, h, sceneF)

	if ink := inkCount8(buf, w, h, [4]int{255, 255, 255, 255}); ink < 50 {
		t.Fatalf("outline text rendered too little ink in 8-bit reference: %d pixels", ink)
	}

	maxDiff, _ := maxImageDiff(buf, img, w, h)
	if maxDiff > 2 {
		t.Errorf("outline text float/8-bit max channel diff = %d (tol 2)", maxDiff)
	}
}

// TestParityTextFreeTypeRaster verifies FreeType raster (gray8) glyph bitmaps
// render to parity. These flow through the BlendSolidHspan coverage path.
func TestParityTextFreeTypeRaster(t *testing.T) {
	fontPath := findSystemFont()
	if fontPath == "" {
		t.Skip("no system font available")
	}

	const w, h = 120, 40
	bg := NewColor(255, 255, 255, 255)
	fg := NewColor(0, 0, 0, 255)

	scene8 := func(a *Agg2D) {
		a.ClearAll(bg)
		if err := a.Font(fontPath, 24.0, false, false, RasterFontCache, 0.0); err != nil {
			t.Skipf("FreeType unavailable: %v", err)
		}
		a.FillColor(fg)
		a.Text(6, 30, "Age", false, 0, 0)
	}
	sceneF := func(a *Agg2DFloat) {
		a.ClearAll(bg)
		if err := a.Font(fontPath, 24.0, false, false, RasterFontCache, 0.0); err != nil {
			t.Skipf("FreeType unavailable: %v", err)
		}
		a.FillColor(fg)
		a.Text(6, 30, "Age", false, 0, 0)
	}

	buf := renderText8(w, h, scene8)
	img := renderTextFloat(w, h, sceneF)

	if ink := inkCount8(buf, w, h, [4]int{255, 255, 255, 255}); ink < 50 {
		t.Fatalf("raster text rendered too little ink in 8-bit reference: %d pixels", ink)
	}

	maxDiff, _ := maxImageDiff(buf, img, w, h)
	if maxDiff > 2 {
		t.Errorf("raster text float/8-bit max channel diff = %d (tol 2)", maxDiff)
	}
}

// TestAgg2DFloatTextWidthParity checks that TextWidth agrees between the 8-bit
// and float pipelines for both the GSV and FreeType backends.
func TestAgg2DFloatTextWidthParity(t *testing.T) {
	const str = "Width"

	a8 := NewAgg2D()
	buf := make([]uint8, 64*16*4)
	a8.Attach(buf, 64, 16, 64*4)
	a8.FontGSV(14)
	aF := NewAgg2DFloat()
	imgF := NewImageFloatEmpty(64, 16)
	aF.AttachImageFloat(imgF)
	aF.FontGSV(14)

	w8 := a8.TextWidth(str)
	wF := aF.TextWidth(str)
	if diff := w8 - wF; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("GSV TextWidth mismatch: 8-bit=%v float=%v", w8, wF)
	}
}
