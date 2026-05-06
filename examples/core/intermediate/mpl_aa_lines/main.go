// Package main demonstrates thin-line anti-aliasing over bright colored backgrounds.
//
// This example reproduces the rendering scenario from matplotlib-go's
// unstructured_showcase: dark contour lines (~1.15 px wide) drawn over bright
// viridis-colormap backgrounds (green/yellow).
//
// It is designed to help diagnose the anti-aliasing coverage difference between
// the Go AGG reimplementation and the reference C++ matplotlib AGG backend.
// Specifically, matplotlib produces more "halo" pixels (partially-covered edge
// pixels) around thin lines on bright backgrounds, making lines look softer and
// slightly wider.  The Go reimplementation renders sharper/crisper edges with
// fewer partial-coverage pixels.
//
// Output (output/mpl_aa_lines.png) has three side-by-side panels:
//
//	Left   – direct rendering: 5 coloured bands (viridis-like) × 4 line widths.
//	         A red rectangle marks the sub-region magnified in the other panels.
//	Middle – 16× pixel-exact zoom: each pixel drawn as a 16×16 solid block so
//	         individual coverage gradations are clearly visible.
//	Right  – coverage heat-map of the same region:
//	           white=background, red=partial coverage, black=fully covered.
//
// Coverage statistics are printed to stdout.
//
// Run:
//
//	go run ./examples/core/intermediate/mpl_aa_lines/
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	agg "github.com/cwbudde/agg_go"
)

const (
	totalWidth  = 960
	totalHeight = 480

	// Preview scene (left panel).
	previewW = 320
	previewH = 480

	// Sub-region to magnify (preview coordinates).
	zoomSrcX = 8
	zoomSrcY = 120
	zoomSrcW = 20
	zoomSrcH = 30

	// Each preview pixel becomes a zoomScale×zoomScale block.
	zoomScale = 16
	// Panel dimensions (must equal zoomSrcW/H * zoomScale).
	zoomPanelW = zoomSrcW * zoomScale // 320
	zoomPanelH = zoomSrcH * zoomScale // 480
)

// viridisColors holds 5 representative viridis samples (dark-purple→yellow).
// The problematic range for the AA difference is the green and yellow entries.
var viridisColors = [5][3]uint8{
	{68, 1, 84},    // 0.00 dark purple
	{59, 82, 139},  // 0.25 blue
	{33, 145, 140}, // 0.50 teal
	{94, 201, 98},  // 0.75 green  ← strong AA difference visible here
	{253, 231, 37}, // 1.00 yellow ← strong AA difference visible here
}

// lineWidths is the set of stroke widths demonstrated in each band.
var lineWidths = []float64{0.5, 1.0, 1.15, 2.0}

// ink is the contour-line colour used in the unstructured_showcase example.
var ink = agg.RGBA(20.0/255, 31.0/255, 46.0/255, 242.0/255)

// rgb8 converts uint8 components to an agg.Color.
func rgb8(r, g, b uint8) agg.Color { return agg.NewColor(r, g, b, 255) }

// ─────────────────────────────────────────────────────────────────────────────

// renderPreview draws the left panel (previewW × previewH) into ctx.
func renderPreview(ctx *agg.Context) {
	bandH := float64(previewH) / float64(len(viridisColors))

	for bi, vc := range viridisColors {
		y0 := float64(bi) * bandH
		y1 := y0 + bandH

		// Fill background band.
		ctx.SetColor(rgb8(vc[0], vc[1], vc[2]))
		ctx.BeginPath()
		ctx.MoveTo(0, y0)
		ctx.LineTo(float64(previewW), y0)
		ctx.LineTo(float64(previewW), y1)
		ctx.LineTo(0, y1)
		ctx.ClosePath()
		ctx.Fill()

		// Draw 4 horizontal lines at increasing widths.
		ctx.SetColor(ink)
		for wi, lw := range lineWidths {
			lineY := y0 + bandH*float64(wi+1)/float64(len(lineWidths)+1)
			ctx.SetLineWidth(lw)
			ctx.BeginPath()
			ctx.MoveTo(8, lineY)
			ctx.LineTo(float64(previewW)-8, lineY)
			ctx.Stroke()
		}

		// Also draw a diagonal line at the contour width (1.15 px).
		ctx.SetLineWidth(1.15)
		ctx.BeginPath()
		ctx.MoveTo(8, y0+4)
		ctx.LineTo(float64(previewW)-8, y1-4)
		ctx.Stroke()
	}
}

// ─────────────────────────────────────────────────────────────────────────────

// buildMagnifiedPanels takes the rendered preview image and returns two panels:
//
//	pixelPanel – each pixel of the zoom region rendered as a zoomScale² block.
//	coverPanel – coverage heat-map (see coverToHeat).
func buildMagnifiedPanels(src *image.RGBA) (pixelPanel, coverPanel *image.RGBA) {
	panelRect := image.Rect(0, 0, zoomPanelW, zoomPanelH)
	pixelPanel = image.NewRGBA(panelRect)
	coverPanel = image.NewRGBA(panelRect)

	for py := 0; py < zoomSrcH; py++ {
		for px := 0; px < zoomSrcW; px++ {
			sc := src.RGBAAt(zoomSrcX+px, zoomSrcY+py)

			// Estimate ink coverage from the green channel.
			// bg≈201 (viridis green), ink≈31; this is a rough heuristic.
			const bgG = 201
			const inkG = 31
			rawCover := (float64(bgG) - float64(sc.G)) / float64(bgG-inkG)
			if rawCover < 0 {
				rawCover = 0
			}
			if rawCover > 1 {
				rawCover = 1
			}

			heat := coverToHeat(rawCover)

			for dy := 0; dy < zoomScale; dy++ {
				for dx := 0; dx < zoomScale; dx++ {
					nx := px*zoomScale + dx
					ny := py*zoomScale + dy
					// 1-pixel grid border for the heat-map panel.
					if dx == 0 || dy == 0 {
						pixelPanel.SetRGBA(nx, ny, color.RGBA{80, 80, 80, 255})
						coverPanel.SetRGBA(nx, ny, color.RGBA{80, 80, 80, 255})
					} else {
						pixelPanel.SetRGBA(nx, ny, sc)
						coverPanel.SetRGBA(nx, ny, heat)
					}
				}
			}
		}
	}

	return pixelPanel, coverPanel
}

// coverToHeat maps a coverage value [0,1] to a diagnostic colour:
//
//	0.0 → white  (pure background, no ink)
//	0.5 → red    (50% edge pixel – the "halo" region)
//	1.0 → black  (fully covered line core)
func coverToHeat(c float64) color.RGBA {
	if c < 0.5 {
		t := c * 2 // 0→1
		v := uint8(255 * (1 - t))
		return color.RGBA{255, v, v, 255}
	}
	t := (c - 0.5) * 2 // 0→1
	v := uint8(255 * (1 - t))
	return color.RGBA{v, 0, 0, 255}
}

// ─────────────────────────────────────────────────────────────────────────────

// printCoverageStats counts "dark" pixels (all channels below threshold) in the
// region [x0,x1) × [y0,y1) at several thresholds, mirroring the analysis done
// in the matplotlib-go parity tests.
func printCoverageStats(label string, src *image.RGBA, x0, x1, y0, y1 int) {
	thresholds := []struct {
		pct   int
		limit uint32
	}{
		{12, 0x2000},
		{19, 0x3000},
		{25, 0x4000},
		{31, 0x5000},
	}

	fmt.Printf("=== %s (x=%d..%d, y=%d..%d) ===\n", label, x0, x1, y0, y1)
	for _, th := range thresholds {
		count := 0
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				c := src.RGBAAt(x, y)
				r := uint32(c.R) << 8
				g := uint32(c.G) << 8
				b := uint32(c.B) << 8
				if r < th.limit && g < th.limit && b < th.limit {
					count++
				}
			}
		}
		fmt.Printf("  threshold %2d%% (0x%04X): %d dark pixels\n", th.pct, th.limit, count)
	}
}

// ─────────────────────────────────────────────────────────────────────────────

// compositeOutput assembles the three panels side by side.
func compositeOutput(preview *image.RGBA, pixelPanel, coverPanel *image.RGBA) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, totalWidth, totalHeight))

	// Grey filler so panel seams are visible.
	for i := 0; i < len(out.Pix); i += 4 {
		out.Pix[i+0] = 200
		out.Pix[i+1] = 200
		out.Pix[i+2] = 200
		out.Pix[i+3] = 255
	}

	blit := func(dst, src *image.RGBA, offX, offY int) {
		for y := 0; y < src.Bounds().Dy(); y++ {
			for x := 0; x < src.Bounds().Dx(); x++ {
				dx, dy := offX+x, offY+y
				if dx < dst.Bounds().Dx() && dy < dst.Bounds().Dy() {
					dst.SetRGBA(dx, dy, src.RGBAAt(x, y))
				}
			}
		}
	}

	blit(out, preview, 0, 0)
	blit(out, pixelPanel, previewW, 0)
	blit(out, coverPanel, previewW+zoomPanelW, 0)

	// Draw red rectangle in preview marking the magnified sub-region.
	red := color.RGBA{255, 0, 0, 255}
	for x := zoomSrcX; x < zoomSrcX+zoomSrcW; x++ {
		out.SetRGBA(x, zoomSrcY, red)
		out.SetRGBA(x, zoomSrcY+zoomSrcH, red)
	}
	for y := zoomSrcY; y <= zoomSrcY+zoomSrcH; y++ {
		out.SetRGBA(zoomSrcX, y, red)
		out.SetRGBA(zoomSrcX+zoomSrcW, y, red)
	}

	return out
}

// ─────────────────────────────────────────────────────────────────────────────

func main() {
	// Render the preview scene.
	ctx := agg.NewContext(previewW, previewH)
	ctx.Clear(agg.White)
	renderPreview(ctx)

	previewGoImg := ctx.GetImage().ToGoImage()

	// Coverage statistics for the whole preview.
	printCoverageStats("full preview", previewGoImg, 0, previewW, 0, previewH)
	// Zoom-region statistics (the region shown magnified).
	printCoverageStats("zoom region", previewGoImg, zoomSrcX, zoomSrcX+zoomSrcW, zoomSrcY, zoomSrcY+zoomSrcH)

	fmt.Println()
	fmt.Println("Note: compare these counts with the C++ matplotlib reference output.")
	fmt.Println("At threshold 19% Go and matplotlib should be nearly equal (~1%).")
	fmt.Println("At threshold 31% matplotlib typically has 2× more pixels due to")
	fmt.Println("wider AA halos on bright backgrounds.")
	fmt.Println()

	// Build magnified panels.
	pixelPanel, coverPanel := buildMagnifiedPanels(previewGoImg)

	// Composite and save.
	out := compositeOutput(previewGoImg, pixelPanel, coverPanel)

	if err := os.MkdirAll("output", 0o755); err != nil {
		panic(err)
	}
	f, err := os.Create("output/mpl_aa_lines.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		panic(err)
	}

	fmt.Println("Saved output/mpl_aa_lines.png")
	fmt.Println()
	fmt.Println("Panel legend:")
	fmt.Println("  Left   – direct rendering (line widths: 0.5, 1.0, 1.15, 2.0 px per band)")
	fmt.Println("  Middle – 16× pixel zoom of the red-bordered region")
	fmt.Println("  Right  – coverage heat-map (white=bg, red=edge pixel, black=core)")
}
