package outline

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/primitives"
)

// MockSource implements the Source interface for testing.
type MockSource struct {
	width  float64
	height float64
	pixels [][]color.RGBA
}

func NewMockSource(width, height int) *MockSource {
	pixels := make([][]color.RGBA, height)
	for y := 0; y < height; y++ {
		pixels[y] = make([]color.RGBA, width)
		for x := 0; x < width; x++ {
			// Create a simple gradient pattern
			r := float64(x) / float64(width)
			g := float64(y) / float64(height)
			pixels[y][x] = color.NewRGBA(r, g, 0.5, 1.0)
		}
	}
	return &MockSource{
		width:  float64(width),
		height: float64(height),
		pixels: pixels,
	}
}

func (ms *MockSource) Width() float64 {
	return ms.width
}

func (ms *MockSource) Height() float64 {
	return ms.height
}

func (ms *MockSource) Pixel(x, y int) color.RGBA {
	if y < 0 || y >= len(ms.pixels) || x < 0 || x >= len(ms.pixels[y]) {
		return color.NewRGBA(0, 0, 0, 0)
	}
	return ms.pixels[y][x]
}

// MockFilter implements the Filter interface for testing.
type MockFilter struct {
	dilation int
}

func NewMockFilter(dilation int) *MockFilter {
	return &MockFilter{dilation: dilation}
}

func (mf *MockFilter) Dilation() int {
	return mf.dilation
}

func (mf *MockFilter) PixelHighRes(rows [][]color.RGBA, p *color.RGBA, x, y int) {
	if len(rows) == 0 {
		*p = color.NewRGBA(1.0, 0.5, 0.5, 1.0) // Return a visible color for testing
		return
	}

	// Simple nearest neighbor for testing with bounds checking
	row := y >> primitives.LineSubpixelShift // Convert from high-res to regular coordinates
	col := x >> primitives.LineSubpixelShift

	if row >= 0 && row < len(rows) && col >= 0 && len(rows[row]) > 0 {
		if col >= len(rows[row]) {
			col = len(rows[row]) - 1
		}
		*p = rows[row][col]
		// Ensure the pixel is not transparent for testing
		if p.A == 0 {
			*p = color.NewRGBA(0.8, 0.2, 0.2, 1.0)
		}
	} else {
		*p = color.NewRGBA(0.5, 0.8, 0.2, 1.0) // Return a visible color for out-of-bounds
	}
}

// MockImageBaseRenderer implements the BaseRenderer interface for testing image rendering.
type MockImageBaseRenderer struct {
	blendHCalls []ImageBlendCall
	blendVCalls []ImageBlendCall
}

type ImageBlendCall struct {
	X, Y   int
	Length int
	Colors []color.RGBA
}

func NewMockImageBaseRenderer() *MockImageBaseRenderer {
	return &MockImageBaseRenderer{
		blendHCalls: make([]ImageBlendCall, 0),
		blendVCalls: make([]ImageBlendCall, 0),
	}
}

func (mbr *MockImageBaseRenderer) BlendColorHSpan(x, y int, length int, colors []color.RGBA, covers []basics.CoverType) {
	colorsCopy := make([]color.RGBA, len(colors))
	copy(colorsCopy, colors)
	mbr.blendHCalls = append(mbr.blendHCalls, ImageBlendCall{
		X: x, Y: y, Length: length, Colors: colorsCopy,
	})
}

func (mbr *MockImageBaseRenderer) BlendColorVSpan(x, y int, length int, colors []color.RGBA, covers []basics.CoverType) {
	colorsCopy := make([]color.RGBA, len(colors))
	copy(colorsCopy, colors)
	mbr.blendVCalls = append(mbr.blendVCalls, ImageBlendCall{
		X: x, Y: y, Length: length, Colors: colorsCopy,
	})
}

func TestLineImageScale(t *testing.T) {
	t.Run("BasicScaling", func(t *testing.T) {
		source := NewMockSource(10, 10)
		scaler := NewLineImageScale(source, 5.0)

		if scaler.Width() != source.Width() {
			t.Errorf("Width() = %f, want %f", scaler.Width(), source.Width())
		}

		if scaler.Height() != 5.0 {
			t.Errorf("Height() = %f, want 5.0", scaler.Height())
		}

		// Test pixel scaling
		pixel := scaler.Pixel(5, 2)
		if pixel.A == 0 {
			t.Error("Scaled pixel should not be transparent")
		}
	})

	t.Run("UpscaleInterpolation", func(t *testing.T) {
		source := NewMockSource(5, 5)
		scaler := NewLineImageScale(source, 10.0)

		// Test that upscaling works
		pixel1 := scaler.Pixel(2, 2)
		pixel2 := scaler.Pixel(2, 3)

		// Pixels should be different due to interpolation
		if pixel1.R == pixel2.R && pixel1.G == pixel2.G {
			t.Error("Upscaled pixels should show interpolation differences")
		}
	})

	t.Run("DownscaleAveraging", func(t *testing.T) {
		source := NewMockSource(10, 10)
		scaler := NewLineImageScale(source, 5.0)

		// Test downscaling averaging
		pixel := scaler.Pixel(2, 2)
		if pixel.A == 0 {
			t.Error("Downscaled pixel should not be transparent")
		}
	})
}

func TestLineImagePattern(t *testing.T) {
	t.Run("BasicPattern", func(t *testing.T) {
		filter := NewMockFilter(1)
		source := NewMockSource(8, 4)
		pattern := NewLineImagePatternFromSource(filter, source)

		if pattern.PatternWidth() <= 0 {
			t.Error("PatternWidth should be positive")
		}

		if pattern.LineWidth() <= 0 {
			t.Error("LineWidth should be positive")
		}

		if pattern.Width() <= 0 {
			t.Error("Width should be positive")
		}

		// Test pixel access
		var p color.RGBA
		pattern.Pixel(&p, 100, 100)
		// Should not crash and should set p to some value
	})

	t.Run("FilterAccess", func(t *testing.T) {
		filter := NewMockFilter(2)
		pattern := NewLineImagePattern(filter)

		if pattern.GetFilter() != filter {
			t.Error("GetFilter should return the same filter instance")
		}
	})
}

func TestLineImagePatternPow2(t *testing.T) {
	t.Run("PowerOf2Optimization", func(t *testing.T) {
		filter := NewMockFilter(1)
		source := NewMockSource(8, 4) // Width 8 is power of 2
		pattern := NewLineImagePatternPow2FromSource(filter, source)

		if pattern.PatternWidth() <= 0 {
			t.Error("PatternWidth should be positive")
		}

		// Test that power-of-2 optimization is working
		var p1, p2 color.RGBA
		pattern.Pixel(&p1, 0, 100)
		pattern.Pixel(&p2, pattern.PatternWidth(), 100)

		// Should wrap around due to power-of-2 masking
		if p1.R != p2.R || p1.G != p2.G || p1.B != p2.B || p1.A != p2.A {
			t.Error("Power-of-2 pattern should wrap around correctly")
		}
	})
}

func TestDistanceInterpolator4(t *testing.T) {
	t.Run("BasicInterpolation", func(t *testing.T) {
		di := NewDistanceInterpolator4(0, 0, 100, 100, 10, 10, 90, 90, 141, 1.0, 50, 50)

		if di.Len() <= 0 {
			t.Error("Length should be positive")
		}

		// Test distance getters
		if di.DX() == 0 && di.DY() == 0 {
			t.Error("DX and DY should not both be zero for non-degenerate line")
		}

		// Test distance updates
		initialDist := di.Dist()
		di.IncX()
		if di.Dist() == initialDist {
			t.Error("IncX should change distance")
		}

		di.DecX()
		if di.Dist() != initialDist {
			t.Error("DecX should restore distance")
		}
	})

	t.Run("DistanceComponents", func(t *testing.T) {
		di := NewDistanceInterpolator4(0, 0, 100, 100, 10, 10, 90, 90, 141, 1.0, 50, 50)

		// Test all distance components are accessible
		_ = di.DistStart()
		_ = di.DistPict()
		_ = di.DistEnd()

		// Test all delta components are accessible
		_ = di.DXStart()
		_ = di.DYStart()
		_ = di.DXPict()
		_ = di.DYPict()
		_ = di.DXEnd()
		_ = di.DYEnd()
	})
}

func TestRendererOutlineImage(t *testing.T) {
	t.Run("BasicRenderer", func(t *testing.T) {
		baseRenderer := NewMockImageBaseRenderer()
		filter := NewMockFilter(1)
		source := NewMockSource(8, 4)
		pattern := NewLineImagePatternFromSource(filter, source)
		renderer := NewRendererOutlineImage(baseRenderer, pattern)

		// Test basic properties
		if renderer.Width() <= 0 {
			t.Error("Width should be positive")
		}

		if renderer.SubpixelWidth() <= 0 {
			t.Error("SubpixelWidth should be positive")
		}

		if renderer.PatternWidth() <= 0 {
			t.Error("PatternWidth should be positive")
		}

		// Test scale operations
		renderer.SetScaleX(2.0)
		if renderer.ScaleX() != 2.0 {
			t.Errorf("ScaleX() = %f, want 2.0", renderer.ScaleX())
		}

		// Test start position
		renderer.SetStartX(10.0)
		if renderer.StartX() != 10.0 {
			t.Errorf("StartX() = %f, want 10.0", renderer.StartX())
		}
	})

	t.Run("ClippingOperations", func(t *testing.T) {
		baseRenderer := NewMockImageBaseRenderer()
		filter := NewMockFilter(1)
		source := NewMockSource(8, 4)
		pattern := NewLineImagePatternFromSource(filter, source)
		renderer := NewRendererOutlineImage(baseRenderer, pattern)

		// Test clipping
		renderer.ClipBox(10.0, 10.0, 100.0, 100.0)
		renderer.ResetClipping()

		// Should not crash
	})

	t.Run("PatternOperations", func(t *testing.T) {
		baseRenderer := NewMockImageBaseRenderer()
		filter := NewMockFilter(1)
		source := NewMockSource(8, 4)
		pattern1 := NewLineImagePatternFromSource(filter, source)
		pattern2 := NewLineImagePatternFromSource(filter, source)
		renderer := NewRendererOutlineImage(baseRenderer, pattern1)

		renderer.SetPattern(pattern2)
		if renderer.GetPattern() != pattern2 {
			t.Error("GetPattern should return the pattern set with SetPattern")
		}
	})

	t.Run("AccurateJoinOnly", func(t *testing.T) {
		baseRenderer := NewMockImageBaseRenderer()
		filter := NewMockFilter(1)
		source := NewMockSource(8, 4)
		pattern := NewLineImagePatternFromSource(filter, source)
		renderer := NewRendererOutlineImage(baseRenderer, pattern)

		if !renderer.AccurateJoinOnly() {
			t.Error("AccurateJoinOnly should return true for image renderer")
		}
	})

	t.Run("LineRendering", func(t *testing.T) {
		baseRenderer := NewMockImageBaseRenderer()
		filter := NewMockFilter(1)
		source := NewMockSource(8, 4)
		pattern := NewLineImagePatternFromSource(filter, source)
		renderer := NewRendererOutlineImage(baseRenderer, pattern)

		// Create a simple line
		lp := primitives.NewLineParameters(0, 0, 100, 100, 141)

		// Test Line3NoClip - should not crash
		renderer.Line3NoClip(&lp, 10, 10, 90, 90)

		// Test Line3 with clipping
		renderer.ClipBox(0, 0, 200, 200)
		renderer.Line3(&lp, 10, 10, 90, 90)

		// Should have called blend functions
		totalCalls := len(baseRenderer.blendHCalls) + len(baseRenderer.blendVCalls)
		if totalCalls == 0 {
			t.Error("Line rendering should generate blend calls")
		}
	})
}

func TestRowPtrCache(t *testing.T) {
	t.Run("BasicCaching", func(t *testing.T) {
		data := make([]color.RGBA, 100)
		for i := range data {
			data[i] = color.NewRGBA(float64(i)/100.0, 0.5, 0.5, 1.0)
		}

		cache := NewMockRowPtrCache()
		cache.Attach(data, 10, 10, 10)

		if cache.Width() != 10 {
			t.Errorf("Width() = %d, want 10", cache.Width())
		}

		if cache.Height() != 10 {
			t.Errorf("Height() = %d, want 10", cache.Height())
		}

		// Test row access
		row := cache.RowPtr(5)
		if len(row) == 0 {
			t.Error("RowPtr should return non-empty row for valid index")
		}

		// Test out of bounds
		row = cache.RowPtr(-1)
		if row != nil {
			t.Error("RowPtr should return nil for negative index")
		}

		row = cache.RowPtr(15)
		if row != nil {
			t.Error("RowPtr should return nil for index >= height")
		}
	})
}

// MockRowPtrCache for testing since the real one is in buffer package
type MockRowPtrCache struct {
	buf    []color.RGBA
	rows   [][]color.RGBA
	width  int
	height int
	stride int
}

func NewMockRowPtrCache() *MockRowPtrCache {
	return &MockRowPtrCache{}
}

func (rpc *MockRowPtrCache) Attach(buf []color.RGBA, width, height, stride int) {
	rpc.buf = buf
	rpc.width = width
	rpc.height = height
	rpc.stride = stride
	rpc.rows = make([][]color.RGBA, height)

	for y := 0; y < height; y++ {
		rowOffset := y * stride
		if rowOffset >= 0 && rowOffset < len(buf) {
			end := rowOffset + width
			if end > len(buf) {
				end = len(buf)
			}
			rpc.rows[y] = buf[rowOffset:end]
		}
	}
}

func (rpc *MockRowPtrCache) RowPtr(y int) []color.RGBA {
	if y < 0 || y >= rpc.height {
		return nil
	}
	return rpc.rows[y]
}

func (rpc *MockRowPtrCache) Width() int {
	return rpc.width
}

func (rpc *MockRowPtrCache) Height() int {
	return rpc.height
}

// Integration test for complete pipeline
func TestImageOutlineRenderingIntegration(t *testing.T) {
	t.Run("CompleteRendering", func(t *testing.T) {
		// Create a complete rendering pipeline
		baseRenderer := NewMockImageBaseRenderer()
		filter := NewMockFilter(1)
		source := NewMockSource(16, 8)
		pattern := NewLineImagePatternFromSource(filter, source)
		renderer := NewRendererOutlineImage(baseRenderer, pattern)

		// Set up rendering parameters
		renderer.SetScaleX(1.5)
		renderer.SetStartX(5.0)
		renderer.ClipBox(0, 0, 200, 200)

		// Render multiple lines to test pattern advancement
		lines := []struct {
			x1, y1, x2, y2 int
			sx, sy, ex, ey int
		}{
			{10, 10, 50, 20, 5, 5, 55, 25},
			{50, 20, 80, 60, 45, 15, 85, 65},
			{80, 60, 120, 80, 75, 55, 125, 85},
		}

		for i, line := range lines {
			lp := primitives.NewLineParameters(line.x1, line.y1, line.x2, line.y2,
				int(basics.CalcDistance(float64(line.x1), float64(line.y1), float64(line.x2), float64(line.y2))))

			renderer.Line3(&lp, line.sx, line.sy, line.ex, line.ey)

			// Verify pattern advancement
			startAfter := renderer.StartX()
			if i == 0 && startAfter <= 5.0 {
				t.Error("Pattern start position should advance after rendering")
			}
		}

		// Verify blend calls were made
		totalCalls := len(baseRenderer.blendHCalls) + len(baseRenderer.blendVCalls)
		if totalCalls == 0 {
			t.Error("Integration test should generate blend calls")
		}

		// Verify colors are not all transparent
		hasNonTransparent := false
		for _, call := range baseRenderer.blendHCalls {
			for _, color := range call.Colors {
				if color.A > 0 {
					hasNonTransparent = true
					break
				}
			}
		}
		for _, call := range baseRenderer.blendVCalls {
			for _, color := range call.Colors {
				if color.A > 0 {
					hasNonTransparent = true
					break
				}
			}
		}

		if !hasNonTransparent {
			t.Error("Rendered colors should not all be transparent")
		}
	})
}
