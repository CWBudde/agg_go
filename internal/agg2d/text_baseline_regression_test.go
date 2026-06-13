package agg2d

import "testing"

// TestRasterTextBaselineMixedHeightString is the pixel-level regression test for
// the FreeType raster-glyph vertical-baseline bug (PLAN.md §3.1): short glyphs
// such as `.`, `,`, and `-` must land on the baseline band when rendered through
// RasterFontCache, not float up above the x-height.
//
// It renders glyphs into a real RGBA buffer and measures the inked vertical
// extent of each. Assertions are expressed relative to a capital "H" (cap band)
// and a lowercase "x" (x-height band) so they hold for any system font rather
// than depending on absolute pixel positions.
//
// Runs for real only under `-tags freetype` (otherwise Font() returns an error
// and the test skips, matching the other FreeType tests in this package). The
// companion unit test TestRasterTextYPhaseMatchesYUpQuantization locks in the
// sub-pixel Y-phase half of the fix; this test locks in the gross geometry.
//
// C++ reference: AGG positions raster glyphs from the integer baseline via the
// FreeType bitmap `top` bearing (agg_font_cache_manager.h init_embedded_adaptors
// + agg_font_freetype.h decompose_ft_bitmap_*); see internal/agg2d/text.go
// renderShapedRasterMask (`dstY = baseY - top + 1`).
func TestRasterTextBaselineMixedHeightString(t *testing.T) {
	fontPath := findSystemFont()
	if fontPath == "" {
		t.Skip("no system font found for FreeType testing")
	}

	const (
		w, h          = 240, 240
		baseX, baseY  = 40.0, 140.0
		fontHeightPts = 48.0
	)

	type inkBox struct{ minX, minY, maxX, maxY, count int }

	measure := func(s string) (inkBox, bool) {
		buf := make([]uint8, w*h*4)
		a := NewAgg2D()
		a.Attach(buf, w, h, w*4)
		a.ClearAll(Color{255, 255, 255, 255})
		if err := a.Font(fontPath, fontHeightPts, false, false, RasterFontCache, 0.0); err != nil {
			return inkBox{}, false // FreeType not built in (no `freetype` tag).
		}
		a.FillColor(Color{0, 0, 0, 255})
		a.Text(baseX, baseY, s, false, 0, 0)

		box := inkBox{minX: w, minY: h, maxX: -1, maxY: -1}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if buf[(y*w+x)*4] < 128 { // dark (inked) pixel on white background
					box.count++
					if x < box.minX {
						box.minX = x
					}
					if y < box.minY {
						box.minY = y
					}
					if x > box.maxX {
						box.maxX = x
					}
					if y > box.maxY {
						box.maxY = y
					}
				}
			}
		}
		return box, true
	}

	hBox, ok := measure("H")
	if !ok {
		t.Skip("FreeType not available (build without `-tags freetype`)")
	}
	xBox, _ := measure("x")
	dotBox, _ := measure(".")
	commaBox, _ := measure(",")
	dashBox, _ := measure("-")
	strBox, _ := measure("0.2 H,x-y")

	// Sanity: every reference glyph must actually produce ink.
	for name, b := range map[string]inkBox{"H": hBox, "x": xBox, ".": dotBox, ",": commaBox, "-": dashBox, "0.2 H,x-y": strBox} {
		if b.count == 0 {
			t.Fatalf("glyph %q produced no ink; cannot evaluate baseline geometry", name)
		}
	}

	// Screen coordinates are y-down: the baseline is at the largest inked Y of a
	// cap glyph, the cap top at the smallest.
	capTop, baseline := hBox.minY, hBox.maxY
	capHeight := baseline - capTop
	capMid := (capTop + baseline) / 2
	xTop := xBox.minY // top of the x-height band

	if capHeight < 8 {
		t.Fatalf("cap height too small to evaluate (%d px); font/render setup wrong", capHeight)
	}

	// 1. The period must not float above the x-height band — this is the exact
	//    regression: short glyphs rendering "above x-height".
	if dotBox.minY < xTop {
		t.Errorf("period floats above x-height: dot.minY=%d < x.minY=%d (baseline=%d, capTop=%d)",
			dotBox.minY, xTop, baseline, capTop)
	}

	// 2. The period must sit in the lower (baseline) half of the cap band.
	if dotBox.minY <= capMid {
		t.Errorf("period not in the baseline band: dot.minY=%d should be > capMid=%d", dotBox.minY, capMid)
	}

	// 3. The period's ink must reach near the baseline (not detached upward).
	if d := baseline - dotBox.maxY; d > capHeight/3 {
		t.Errorf("period detached from baseline: baseline=%d dot.maxY=%d gap=%d > %d",
			baseline, dotBox.maxY, d, capHeight/3)
	}

	// 4. The comma is a baseline mark with a descender: it must extend below the
	//    baseline. This also confirms baseline orientation is correct (not flipped).
	if commaBox.maxY <= baseline {
		t.Errorf("comma does not descend below baseline: comma.maxY=%d, baseline=%d", commaBox.maxY, baseline)
	}

	// 5. The hyphen is a mid-line mark: above the baseline and below the cap top.
	if dashBox.maxY >= baseline {
		t.Errorf("hyphen reaches the baseline: dash.maxY=%d, baseline=%d", dashBox.maxY, baseline)
	}
	if dashBox.minY <= capTop {
		t.Errorf("hyphen sits at/above cap top: dash.minY=%d, capTop=%d", dashBox.minY, capTop)
	}

	// 6. In the full mixed-height string nothing floats above the cap line and the
	//    descenders extend below the baseline — the line stays in a sane band.
	if strBox.minY < capTop-capHeight/4 {
		t.Errorf("string ink floats above the cap line: str.minY=%d, capTop=%d", strBox.minY, capTop)
	}
	if strBox.maxY <= baseline {
		t.Errorf("string has no descender ink below baseline: str.maxY=%d, baseline=%d", strBox.maxY, baseline)
	}
}
