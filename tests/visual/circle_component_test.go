package visual

import (
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	scanlinestorage "agg_go/internal/scanline"
	"agg_go/internal/shapes"
)

// ScanlineAdapter adapts the scanline storage to the rasterizer interface
type ScanlineAdapter struct {
	sl *scanlinestorage.ScanlineU8
}

func (sa *ScanlineAdapter) ResetSpans() {
	sa.sl.ResetSpans()
}

func (sa *ScanlineAdapter) AddCell(x int, cover uint32) {
	sa.sl.AddCell(x, uint(cover))
}

func (sa *ScanlineAdapter) AddSpan(x, length int, cover uint32) {
	sa.sl.AddSpan(x, length, uint(cover))
}

func (sa *ScanlineAdapter) Finalize(y int) {
	sa.sl.Finalize(y)
}

func (sa *ScanlineAdapter) NumSpans() int {
	return sa.sl.NumSpans()
}

// savePNG saves the RGBA buffer as a PNG file for visual verification
func savePNG(filename string, pixelData []uint8, width, height int) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Create RGBA image from our buffer
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Copy pixel data - Go's image.RGBA expects RGBA format (which we have)
	copy(img.Pix, pixelData)
	
	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Encode as PNG
	return png.Encode(file, img)
}

// TestCircleComponentPipeline tests the complete AGG rendering pipeline 
// by drawing a circle using only low-level components, without agg2d.
// This demonstrates the step-by-step process from path to pixels.
func TestCircleComponentPipeline(t *testing.T) {
	// Test parameters
	width, height := 200, 200
	centerX, centerY := 100.0, 100.0
	radius := 50.0
	
	// Stage 1: Set up rendering buffer and pixel format
	t.Log("Stage 1: Setting up rendering buffer and pixel format")
	
	// Create raw RGBA buffer
	stride := width * 4 // 4 bytes per pixel (RGBA)
	pixelData := make([]uint8, height*stride)
	
	// Initialize with white background
	for i := 0; i < len(pixelData); i += 4 {
		pixelData[i] = 255   // R
		pixelData[i+1] = 255 // G  
		pixelData[i+2] = 255 // B
		pixelData[i+3] = 255 // A
	}
	
	// Create rendering buffer
	renderingBuffer := buffer.NewRenderingBufferWithData(pixelData, width, height, stride)
	if renderingBuffer.Width() != width || renderingBuffer.Height() != height {
		t.Fatalf("Rendering buffer setup failed: got %dx%d, expected %dx%d", 
			renderingBuffer.Width(), renderingBuffer.Height(), width, height)
	}
	t.Log("✓ Rendering buffer created successfully")
	
	// Create RGBA pixel format with alpha blending
	blenderInstance := blender.NewCompositeBlender[color.Linear, interface{}](blender.CompOpSrcOver)
	pixfmt := pixfmt.NewPixFmtAlphaBlendRGBA[blender.CompositeBlender[color.Linear, interface{}], color.Linear](renderingBuffer, blenderInstance)
	if pixfmt.Width() != width || pixfmt.Height() != height {
		t.Fatalf("Pixel format setup failed: got %dx%d, expected %dx%d",
			pixfmt.Width(), pixfmt.Height(), width, height)
	}
	t.Log("✓ RGBA pixel format created successfully")
	
	// Stage 2: Set up base renderer
	t.Log("Stage 2: Setting up base renderer")
	
	baseRenderer := renderer.NewRendererBaseWithPixfmt(pixfmt)
	if baseRenderer.Width() != width || baseRenderer.Height() != height {
		t.Fatalf("Base renderer setup failed: got %dx%d, expected %dx%d",
			baseRenderer.Width(), baseRenderer.Height(), width, height)
	}
	t.Log("✓ Base renderer created successfully")
	
	// Stage 3: Create circle path using Ellipse shape
	t.Log("Stage 3: Generating circle path")
	
	ellipse := shapes.NewEllipseWithParams(centerX, centerY, radius, radius, 0, false)
	t.Logf("✓ Circle path created: center=(%.1f,%.1f), radius=%.1f", centerX, centerY, radius)
	
	// Stage 4: Set up rasterizer and scanline
	t.Log("Stage 4: Setting up rasterizer and scanline")
	
	// Create clipper (real clipper for this test)
	clipper := rasterizer.NewRasterizerSlClip[rasterizer.RasConvDbl]()
	
	// Create rasterizer
	cellBlockLimit := uint32(512)
	rasterizerInstance := rasterizer.NewRasterizerScanlineAA[*rasterizer.RasterizerSlClip[rasterizer.RasConvDbl], rasterizer.RasConvDbl](cellBlockLimit, clipper)
	rasterizerInstance.FillingRule(basics.FillNonZero)
	
	// Set clipping box to cover our canvas
	rasterizerInstance.ClipBox(0, 0, float64(width), float64(height))
	t.Log("✓ Rasterizer created and clipping box set")
	
	// Create scanline container
	scanlineContainer := scanlinestorage.NewScanlineU8()
	t.Log("✓ Scanline container created successfully")
	
	// Stage 5: Rasterize the circle path
	t.Log("Stage 5: Rasterizing circle path")
	
	// Reset rasterizer
	rasterizerInstance.Reset()
	
	// Add the ellipse path to the rasterizer
	ellipse.Rewind(0)
	var x, y float64
	
	pathStarted := false
	vertexCount := 0
	
	for {
		cmd := ellipse.Vertex(&x, &y)
		
		if cmd == basics.PathCmdStop {
			break
		}
		
		vertexCount++
		if vertexCount <= 5 {
			t.Logf("Vertex %d: cmd=%d, (%.2f, %.2f)", vertexCount, uint32(cmd), x, y)
		}
		
		switch cmd {
		case basics.PathCmdMoveTo:
			rasterizerInstance.MoveToD(x, y)
			pathStarted = true
		case basics.PathCmdLineTo:
			if pathStarted {
				rasterizerInstance.LineToD(x, y)
			}
		default:
			if cmd&basics.PathCmdEndPoly != 0 {
				// End of polygon - close it if needed
				if uint32(cmd)&uint32(basics.PathFlagsClose) != 0 {
					t.Log("Closing polygon")
					rasterizerInstance.ClosePolygon()
				}
			}
		}
	}
	
	t.Logf("Total vertices generated: %d", vertexCount)
	
	t.Log("✓ Circle path added to rasterizer")
	
	// Stage 6: Render scanlines to pixels manually (simplified approach)
	t.Log("Stage 6: Rendering scanlines to pixels manually")
	
	// Define the fill color (red circle)
	fillColor := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}
	
	// Simplified approach: Create adapter for scanline interface compatibility
	scanlineAdapter := &ScanlineAdapter{scanlineContainer}
	
	// Manual scanline processing to demonstrate the pipeline
	scanlineCount := 0
	spanCount := 0
	pixelCount := 0
	
	if rasterizerInstance.RewindScanlines() {
		t.Log("Rasterizer has scanlines to process")
		
		// Initialize scanline bounds once for the entire sweep
		scanlineContainer.Reset(rasterizerInstance.MinX(), rasterizerInstance.MaxX())
		
		// Sweep through all scanlines - each call processes one scanline
		for rasterizerInstance.SweepScanline(scanlineAdapter) {
			scanlineCount++
			
			// Process each scanline by iterating through spans
			y := scanlineContainer.Y()
			
			// Get spans and manually render pixels
			spans := scanlineContainer.Begin()
			if len(spans) > 0 {
				t.Logf("Scanline %d (y=%d): %d spans", scanlineCount, y, len(spans))
				
				// Verify span properties for debugging
				if scanlineCount <= 5 {
					for i, span := range spans {
						t.Logf("  Span %d: X=%d, Len=%d, covers[0]=%d", i, span.X, span.Len, 
							func() uint8 { if len(span.Covers) > 0 { return span.Covers[0] } else { return 0 } }())
					}
				}
				
				// Assert reasonable span properties for circle
				if y >= 70 && y <= 130 { // Within circle's vertical bounds
					totalSpanLength := 0
					for _, span := range spans {
						totalSpanLength += int(span.Len)
						// Spans should have positive length
						if span.Len <= 0 {
							t.Errorf("Scanline %d: span has invalid length %d", scanlineCount, span.Len)
						}
						// X coordinates should be within canvas bounds  
						if span.X < 0 || span.X >= 200 {
							t.Errorf("Scanline %d: span X coordinate %d out of bounds", scanlineCount, span.X)
						}
					}
					// For a circle, expect reasonable total span length
					if totalSpanLength > 120 || totalSpanLength < 10 {
						t.Logf("Warning: Scanline %d has unusual total span length: %d", scanlineCount, totalSpanLength)
					}
				}
			}
			
			for _, span := range spans {
				spanCount++
				// Render each pixel in the span
				for px := 0; px < int(span.Len); px++ {
					x := int(span.X) + px
					if x >= 0 && x < width && y >= 0 && y < height {
						cover := basics.Int8u(basics.CoverFull)
						if px < len(span.Covers) {
							cover = span.Covers[px]
						}
						// Blend the pixel
						baseRenderer.BlendPixel(x, y, fillColor, cover)
						pixelCount++
						
						// Log first few pixels for debugging
						if pixelCount <= 5 {
							t.Logf("  Pixel (%d,%d) cover=%d", x, y, cover)
						}
					}
				}
			}
		}
	} else {
		t.Log("Rasterizer has no scanlines to process")
	}
	
	t.Logf("Processed %d scanlines, %d spans, %d pixels", scanlineCount, spanCount, pixelCount)
	
	// Diagnostic information for debugging
	t.Logf("Rasterizer bounds: MinX=%d, MaxX=%d, MinY=%d, MaxY=%d", 
		rasterizerInstance.MinX(), rasterizerInstance.MaxX(),
		rasterizerInstance.MinY(), rasterizerInstance.MaxY())
	
	if scanlineCount > 0 {
		avgSpansPerScanline := float64(spanCount) / float64(scanlineCount)
		avgPixelsPerSpan := float64(pixelCount) / float64(spanCount)
		t.Logf("Average spans per scanline: %.2f", avgSpansPerScanline)
		t.Logf("Average pixels per span: %.2f", avgPixelsPerSpan)
		
		// For a circle, we expect reasonable ratios
		if avgSpansPerScanline > 5.0 {
			t.Logf("Warning: High spans-per-scanline ratio may indicate fragmented rendering")
		}
		if avgPixelsPerSpan < 5.0 {
			t.Logf("Warning: Low pixels-per-span ratio may indicate rendering issues")
		}
	}
	
	t.Log("✓ Scanlines rendered to pixels manually")
	
	// Stage 7: Verify pixel results
	t.Log("Stage 7: Verifying pixel results")
	
	verifyPixel := func(x, y int, expected color.RGBA8[color.Linear], description string) {
		actual := pixfmt.Pixel(x, y)
		if actual.R != expected.R || actual.G != expected.G || actual.B != expected.B {
			t.Errorf("%s at (%d,%d): expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
				description, x, y, expected.R, expected.G, expected.B, actual.R, actual.G, actual.B)
		} else {
			t.Logf("✓ %s at (%d,%d): RGB(%d,%d,%d)", description, x, y, actual.R, actual.G, actual.B)
		}
	}
	
	// Test center of circle (should be red)
	verifyPixel(100, 100, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}, "Circle center")
	
	// Test point clearly inside circle (should be red)
	verifyPixel(100, 120, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}, "Inside circle")
	
	// Test point clearly outside circle (should be white background)
	verifyPixel(50, 50, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}, "Outside circle")
	
	// Test edge pixels for anti-aliasing (should be intermediate values)
	edgePixel := pixfmt.Pixel(100, 50) // Top edge of circle
	if edgePixel.R == 255 && edgePixel.G == 255 && edgePixel.B == 255 {
		t.Error("Edge pixel should show anti-aliasing but appears to be pure white")
	} else if edgePixel.R == 255 && edgePixel.G == 0 && edgePixel.B == 0 {
		t.Error("Edge pixel should show anti-aliasing but appears to be pure red")
	} else {
		t.Logf("✓ Edge anti-aliasing detected at (100,50): RGB(%d,%d,%d)", edgePixel.R, edgePixel.G, edgePixel.B)
	}
	
	// Test multiple edge positions for proper anti-aliasing
	edgeTests := []struct {
		x, y int
		desc string
	}{
		{100, 50, "top edge"},
		{100, 150, "bottom edge"}, 
		{50, 100, "left edge"},
		{150, 100, "right edge"},
		{75, 75, "top-left diagonal"},
		{125, 125, "bottom-right diagonal"},
	}
	
	antiAliasedCount := 0
	for _, test := range edgeTests {
		pixel := pixfmt.Pixel(test.x, test.y)
		// Anti-aliased pixels should have intermediate red values (not pure red or white)
		if pixel.R > 50 && pixel.R < 200 && pixel.G < 100 && pixel.B < 100 {
			antiAliasedCount++
			t.Logf("✓ Anti-aliasing at %s (%d,%d): RGB(%d,%d,%d)", test.desc, test.x, test.y, pixel.R, pixel.G, pixel.B)
		} else if pixel.R == 255 && pixel.G == 0 && pixel.B == 0 {
			t.Logf("  Solid fill at %s (%d,%d): RGB(%d,%d,%d)", test.desc, test.x, test.y, pixel.R, pixel.G, pixel.B)
		} else if pixel.R == 255 && pixel.G == 255 && pixel.B == 255 {
			t.Logf("  Background at %s (%d,%d): RGB(%d,%d,%d)", test.desc, test.x, test.y, pixel.R, pixel.G, pixel.B)
		} else {
			t.Logf("  Unexpected color at %s (%d,%d): RGB(%d,%d,%d)", test.desc, test.x, test.y, pixel.R, pixel.G, pixel.B)
		}
	}
	
	if antiAliasedCount == 0 {
		t.Error("No anti-aliased edge pixels found - circle should have smooth edges")
	} else {
		t.Logf("✓ Found %d anti-aliased edge pixels out of %d tested", antiAliasedCount, len(edgeTests))
	}
	
	// Check for diagonal artifact (should NOT exist)
	diagonalArtifacts := 0
	for i := 0; i < 100; i++ { // Check diagonal from (0,0) to (100,100)
		if i < width && i < height {
			pixel := pixfmt.Pixel(i, i)
			// Calculate distance from circle center
			dx := float64(i) - centerX
			dy := float64(i) - centerY  
			distance := math.Sqrt(dx*dx + dy*dy)
			
			// If this diagonal pixel is red but should be outside circle, it's an artifact
			if distance > radius+2 && pixel.R > 200 && pixel.G < 50 && pixel.B < 50 {
				diagonalArtifacts++
				if diagonalArtifacts <= 3 { // Log first few
					t.Logf("Diagonal artifact detected at (%d,%d): RGB(%d,%d,%d), distance=%.1f", 
						i, i, pixel.R, pixel.G, pixel.B, distance)
				}
			}
		}
	}
	
	if diagonalArtifacts > 0 {
		t.Errorf("Found %d diagonal artifacts - these indicate improper scanline processing", diagonalArtifacts)
	} else {
		t.Log("✓ No diagonal artifacts detected")
	}
	
	// Check for alternating line pattern (horizontal stripes)
	stripeArtifacts := 0
	for y := 70; y < 130; y += 2 { // Check every other line in circle area
		redInLine := 0
		for x := 70; x < 130; x++ {
			pixel := pixfmt.Pixel(x, y)
			if pixel.R > 200 && pixel.G < 50 && pixel.B < 50 {
				redInLine++
			}
		}
		// If alternating lines, some will be empty when they shouldn't be
		if redInLine == 0 {
			stripeArtifacts++
		}
	}
	
	if stripeArtifacts > 2 { // Allow for some edge cases
		t.Errorf("Found %d empty horizontal lines in circle area - indicates alternating stripe artifact", stripeArtifacts)
	} else {
		t.Log("✓ No alternating stripe artifacts detected")
	}
	
	// Count red pixels to verify approximate circle area
	redPixelCount := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pixfmt.Pixel(x, y)
			if pixel.R > 200 && pixel.G < 50 && pixel.B < 50 {
				redPixelCount++
			}
		}
	}
	
	// Expected area of circle: π * r²
	expectedArea := 3.14159 * radius * radius
	tolerance := 0.2 // 20% tolerance for anti-aliasing effects
	
	if float64(redPixelCount) < expectedArea*(1-tolerance) || float64(redPixelCount) > expectedArea*(1+tolerance) {
		t.Errorf("Circle area verification failed: got %d red pixels, expected approximately %.0f (±%.0f%%)",
			redPixelCount, expectedArea, tolerance*100)
	} else {
		t.Logf("✓ Circle area verified: %d pixels (expected ~%.0f)", redPixelCount, expectedArea)
	}
	
	// Stage 8: Save result as PNG for visual verification
	t.Log("Stage 8: Saving result as PNG file")
	
	outputPath := "/home/christian/Code/agg_go/tests/visual/output/circle_component_test.png"
	if err := savePNG(outputPath, pixelData, width, height); err != nil {
		t.Logf("Warning: Failed to save PNG: %v", err)
	} else {
		t.Logf("✓ PNG saved to: %s", outputPath)
		t.Logf("  You can view this file to visually verify the circle rendering")
	}
	
	t.Log("✓ All verification tests passed - low-level component pipeline works correctly!")
}

// TestCircleVisualDemo creates a larger, clearer image for visual demonstration
func TestCircleVisualDemo(t *testing.T) {
	// Create a larger canvas for better visual inspection
	width, height := 400, 400
	centerX, centerY := 200.0, 200.0
	radius := 100.0
	
	// Set up rendering buffer
	stride := width * 4
	pixelData := make([]uint8, height*stride)
	
	// Initialize with white background
	for i := 0; i < len(pixelData); i += 4 {
		pixelData[i] = 255   // R
		pixelData[i+1] = 255 // G  
		pixelData[i+2] = 255 // B
		pixelData[i+3] = 255 // A
	}
	
	// Set up the rendering pipeline
	renderingBuffer := buffer.NewRenderingBufferWithData(pixelData, width, height, stride)
	blenderInstance := blender.NewCompositeBlender[color.Linear, interface{}](blender.CompOpSrcOver)
	pixfmt := pixfmt.NewPixFmtAlphaBlendRGBA[blender.CompositeBlender[color.Linear, interface{}], color.Linear](renderingBuffer, blenderInstance)
	baseRenderer := renderer.NewRendererBaseWithPixfmt(pixfmt)
	
	// Set up rasterizer
	clipper := rasterizer.NewRasterizerSlClip[rasterizer.RasConvDbl]()
	rasterizerInstance := rasterizer.NewRasterizerScanlineAA[*rasterizer.RasterizerSlClip[rasterizer.RasConvDbl], rasterizer.RasConvDbl](1024, clipper)
	rasterizerInstance.FillingRule(basics.FillNonZero)
	rasterizerInstance.ClipBox(0, 0, float64(width), float64(height))
	
	// Create circle path
	ellipse := shapes.NewEllipseWithParams(centerX, centerY, radius, radius, 0, false)
	
	// Add path to rasterizer
	rasterizerInstance.Reset()
	ellipse.Rewind(0)
	var x, y float64
	
	for {
		cmd := ellipse.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}
		
		switch cmd {
		case basics.PathCmdMoveTo:
			rasterizerInstance.MoveToD(x, y)
		case basics.PathCmdLineTo:
			rasterizerInstance.LineToD(x, y)
		default:
			if cmd&basics.PathCmdEndPoly != 0 {
				if uint32(cmd)&uint32(basics.PathFlagsClose) != 0 {
					rasterizerInstance.ClosePolygon()
				}
			}
		}
	}
	
	// Render the circle
	scanlineContainer := scanlinestorage.NewScanlineU8()
	scanlineAdapter := &ScanlineAdapter{scanlineContainer}
	fillColor := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255} // Red
	
	if rasterizerInstance.RewindScanlines() {
		scanlineContainer.Reset(rasterizerInstance.MinX(), rasterizerInstance.MaxX())
		
		for rasterizerInstance.SweepScanline(scanlineAdapter) {
			y := scanlineContainer.Y()
			spans := scanlineContainer.Begin()
			
			for _, span := range spans {
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
	
	// Save the result
	outputPath := "/home/christian/Code/agg_go/tests/visual/output/circle_demo_400x400.png"
	if err := savePNG(outputPath, pixelData, width, height); err != nil {
		t.Errorf("Failed to save PNG: %v", err)
	} else {
		t.Logf("✓ Large demo circle saved to: %s", outputPath)
		t.Logf("  This 400x400 image shows a red circle with anti-aliasing")
	}
}

// TestComponentStages tests each stage of the pipeline individually
func TestComponentStages(t *testing.T) {
	// Test Stage 1: Buffer creation
	t.Run("BufferCreation", func(t *testing.T) {
		width, height := 100, 100
		stride := width * 4
		data := make([]uint8, height*stride)
		
		buf := buffer.NewRenderingBufferWithData(data, width, height, stride)
		if buf.Width() != width || buf.Height() != height {
			t.Errorf("Buffer dimensions incorrect: got %dx%d, expected %dx%d", 
				buf.Width(), buf.Height(), width, height)
		}
		
		// Test row access
		row := buf.Row(0)
		if len(row) == 0 {
			t.Error("Row access failed - returned empty slice")
		}
	})
	
	// Test Stage 2: Pixel format operations
	t.Run("PixelFormat", func(t *testing.T) {
		width, height := 10, 10
		stride := width * 4
		data := make([]uint8, height*stride)
		
		buf := buffer.NewRenderingBufferWithData(data, width, height, stride)
		blenderInstance := blender.NewCompositeBlender[color.Linear, interface{}](blender.CompOpSrcOver)
		pf := pixfmt.NewPixFmtAlphaBlendRGBA[blender.CompositeBlender[color.Linear, interface{}], color.Linear](buf, blenderInstance)
		
		// Test pixel operations
		testColor := color.RGBA8[color.Linear]{R: 128, G: 64, B: 192, A: 255}
		pf.CopyPixel(5, 5, testColor)
		
		retrieved := pf.Pixel(5, 5)
		if retrieved.R != testColor.R || retrieved.G != testColor.G || retrieved.B != testColor.B {
			t.Errorf("Pixel copy/retrieve failed: expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
				testColor.R, testColor.G, testColor.B, retrieved.R, retrieved.G, retrieved.B)
		}
	})
	
	// Test Stage 3: Path generation
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
			
			// Verify coordinates are reasonable
			if x < 0 || x > 100 || y < 0 || y > 100 {
				t.Errorf("Path vertex out of expected bounds: (%.2f, %.2f)", x, y)
			}
		}
		
		if vertexCount < 8 {
			t.Errorf("Too few vertices generated: got %d, expected at least 8", vertexCount)
		}
	})
}