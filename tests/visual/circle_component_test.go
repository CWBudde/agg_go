package visual

import (
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	scanlinestorage "github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

// savePNG saves the RGBA buffer as a PNG file for visual verification
func savePNG(filename string, pixelData []uint8, width, height int) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, pixelData)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func renderCircle(t *testing.T, width, height int, centerX, centerY, radius float64) (pf *pixfmt.PixFmtRGBA32[color.Linear], pixelData []uint8) {
	t.Helper()

	stride := width * 4
	pixelData = make([]uint8, height*stride)
	for i := 0; i < len(pixelData); i += 4 {
		pixelData[i] = 255
		pixelData[i+1] = 255
		pixelData[i+2] = 255
		pixelData[i+3] = 255
	}

	renderingBuffer := buffer.NewRenderingBufferWithData(pixelData, width, height, stride)
	pf = pixfmt.NewPixFmtRGBA32Linear(renderingBuffer)
	baseRenderer := renderer.NewRendererBaseWithPixfmt(pf)

	conv := rasterizer.RasConvInt{}
	clipper := rasterizer.NewRasterizerSlNoClip()
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](conv, clipper)
	ras.FillingRule(basics.FillEvenOdd)
	ras.ClipBox(0, 0, float64(width), float64(height))

	ellipse := shapes.NewEllipseWithParams(centerX, centerY, radius, radius, 0, false)
	ellipse.Rewind(0)

	var x, y float64
	for {
		cmd := ellipse.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}
		switch cmd {
		case basics.PathCmdMoveTo:
			ras.MoveToD(x, y)
		case basics.PathCmdLineTo:
			ras.LineToD(x, y)
		default:
			if cmd&basics.PathCmdEndPoly != 0 && uint32(cmd)&uint32(basics.PathFlagsClose) != 0 {
				ras.ClosePolygon()
			}
		}
	}

	sl := scanlinestorage.NewScanlineU8()
	fillColor := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(sl) {
			y := sl.Y()
			for _, span := range sl.Begin() {
				for px := 0; px < int(span.Len); px++ {
					x := int(span.X) + px
					if x >= 0 && x < width && y >= 0 && y < height {
						cover := basics.Int8u(basics.CoverFull)
						if px < len(span.Covers) {
							cover = span.Covers[px]
						}
						baseRenderer.BlendPixel(x, y, fillColor, cover)
					}
				}
			}
		}
	}
	return pf, pixelData
}

func TestCircleComponentPipeline(t *testing.T) {
	width, height := 200, 200
	centerX, centerY := 100.0, 100.0
	radius := 50.0

	pf, _ := renderCircle(t, width, height, centerX, centerY, radius)

	// Center and interior must be red-dominant.
	for _, pt := range []struct{ x, y int }{{100, 100}, {100, 120}} {
		p := pf.Pixel(pt.x, pt.y)
		if p.R <= p.G+20 || p.R <= p.B+20 {
			t.Errorf("expected red-dominant at (%d,%d), got RGB(%d,%d,%d)", pt.x, pt.y, p.R, p.G, p.B)
		}
	}

	// Outside must be white.
	p := pf.Pixel(50, 50)
	if p.R != 255 || p.G != 255 || p.B != 255 {
		t.Errorf("expected white outside circle at (50,50), got RGB(%d,%d,%d)", p.R, p.G, p.B)
	}

	// No diagonal artifacts outside the circle.
	for i := 0; i < 100; i++ {
		if i >= width || i >= height {
			continue
		}
		dx := float64(i) - centerX
		dy := float64(i) - centerY
		if math.Sqrt(dx*dx+dy*dy) > radius+2 {
			p := pf.Pixel(i, i)
			if p.R > 200 && p.G < 50 && p.B < 50 {
				t.Errorf("diagonal artifact at (%d,%d): RGB(%d,%d,%d)", i, i, p.R, p.G, p.B)
			}
		}
	}

	// No alternating horizontal stripes inside the circle.
	stripeArtifacts := 0
	for y := 70; y < 130; y += 2 {
		redInLine := 0
		for x := 70; x < 130; x++ {
			p := pf.Pixel(x, y)
			if p.R > 200 && p.G < 50 && p.B < 50 {
				redInLine++
			}
		}
		if redInLine == 0 {
			stripeArtifacts++
		}
	}
	if stripeArtifacts > 2 {
		t.Errorf("found %d empty horizontal lines in circle area (stripe artifact)", stripeArtifacts)
	}

	// Approximate area check.
	redPixelCount := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := pf.Pixel(x, y)
			if p.R > 200 && p.G < 50 && p.B < 50 {
				redPixelCount++
			}
		}
	}
	expectedArea := math.Pi * radius * radius
	if float64(redPixelCount) < expectedArea*0.2 || float64(redPixelCount) > expectedArea*1.3 {
		t.Errorf("circle area out of expected range: got %d pixels, expected ~%.0f", redPixelCount, expectedArea)
	}
}

// TestCircleVisualDemo creates a larger image for visual inspection.
func TestCircleVisualDemo(t *testing.T) {
	_, pixelData := renderCircle(t, 400, 400, 200, 200, 100)
	outputPath := filepath.Join("output", "circle_demo_400x400.png")
	if err := savePNG(outputPath, pixelData, 400, 400); err != nil {
		t.Errorf("failed to save PNG: %v", err)
	}
}

// TestComponentStages tests each stage of the pipeline individually.
func TestComponentStages(t *testing.T) {
	t.Run("BufferCreation", func(t *testing.T) {
		width, height := 100, 100
		stride := width * 4
		data := make([]uint8, height*stride)
		buf := buffer.NewRenderingBufferWithData(data, width, height, stride)
		if buf.Width() != width || buf.Height() != height {
			t.Errorf("buffer dimensions: got %dx%d, expected %dx%d", buf.Width(), buf.Height(), width, height)
		}
		if len(buf.Row(0)) == 0 {
			t.Error("row access returned empty slice")
		}
	})

	t.Run("PixelFormat", func(t *testing.T) {
		width, height := 10, 10
		stride := width * 4
		data := make([]uint8, height*stride)
		buf := buffer.NewRenderingBufferWithData(data, width, height, stride)
		pf := pixfmt.NewPixFmtRGBA32Linear(buf)
		testColor := color.RGBA8[color.Linear]{R: 128, G: 64, B: 192, A: 255}
		pf.CopyPixel(5, 5, testColor)
		got := pf.Pixel(5, 5)
		if got.R != testColor.R || got.G != testColor.G || got.B != testColor.B {
			t.Errorf("pixel copy/retrieve: expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
				testColor.R, testColor.G, testColor.B, got.R, got.G, got.B)
		}
	})

	t.Run("PathGeneration", func(t *testing.T) {
		ellipse := shapes.NewEllipseWithParams(50, 50, 25, 25, 0, false)
		ellipse.Rewind(0)
		var x, y float64
		vertexCount := 0
		for {
			cmd := ellipse.Vertex(&x, &y)
			if cmd == basics.PathCmdStop {
				break
			}
			vertexCount++
			if x < 0 || x > 100 || y < 0 || y > 100 {
				t.Errorf("path vertex out of bounds: (%.2f, %.2f)", x, y)
			}
		}
		if vertexCount < 8 {
			t.Errorf("too few vertices: got %d, expected at least 8", vertexCount)
		}
	})
}
