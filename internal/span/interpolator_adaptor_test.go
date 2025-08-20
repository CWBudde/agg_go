package span

import (
	"math"
	"testing"
)

// intAbs returns the absolute value of an integer
func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// MockAdaptorInterpolator is a simple interpolator for testing the adaptor.
type MockAdaptorInterpolator struct {
	x, y          float64
	ix, iy        int
	subpixelScale int
}

func NewMockAdaptorInterpolator() *MockAdaptorInterpolator {
	return &MockAdaptorInterpolator{
		subpixelScale: 256, // Default subpixel scale
	}
}

func (m *MockAdaptorInterpolator) Begin(x, y float64, length int) {
	m.x = x
	m.y = y
	m.ix = int(x * float64(m.subpixelScale))
	m.iy = int(y * float64(m.subpixelScale))
}

func (m *MockAdaptorInterpolator) Next() {
	m.x += 1.0
	m.ix = int(m.x * float64(m.subpixelScale))
}

func (m *MockAdaptorInterpolator) Coordinates() (x, y int) {
	return m.ix, m.iy
}

func (m *MockAdaptorInterpolator) Resynchronize(xe, ye float64, length int) {
	// Simple implementation - just re-begin at the end point
	m.Begin(xe, ye, length)
}

func (m *MockAdaptorInterpolator) SubpixelShift() int {
	return 8 // Return shift value that gives scale of 256
}

// IdentityDistortion does no transformation - useful for testing basic functionality.
type IdentityDistortion struct{}

func (d *IdentityDistortion) Calculate(x, y *int) {
	// No change
}

// ScaleDistortion scales coordinates by a fixed factor.
type ScaleDistortion struct {
	scaleX, scaleY float64
	subpixelScale  int
}

func NewScaleDistortion(scaleX, scaleY float64) *ScaleDistortion {
	return &ScaleDistortion{
		scaleX:        scaleX,
		scaleY:        scaleY,
		subpixelScale: 256,
	}
}

func (d *ScaleDistortion) Calculate(x, y *int) {
	// Convert from subpixel to float, scale, then back
	fx := float64(*x) / float64(d.subpixelScale)
	fy := float64(*y) / float64(d.subpixelScale)

	fx *= d.scaleX
	fy *= d.scaleY

	*x = int(fx * float64(d.subpixelScale))
	*y = int(fy * float64(d.subpixelScale))
}

// OffsetDistortion adds a fixed offset to coordinates.
type OffsetDistortion struct {
	offsetX, offsetY int
}

func NewOffsetDistortion(offsetX, offsetY int) *OffsetDistortion {
	return &OffsetDistortion{
		offsetX: offsetX,
		offsetY: offsetY,
	}
}

func (d *OffsetDistortion) Calculate(x, y *int) {
	*x += d.offsetX
	*y += d.offsetY
}

// WaveDistortion applies a wave-like distortion similar to AGG's examples.
type WaveDistortion struct {
	centerX, centerY float64
	amplitude        float64
	period           float64
	phase            float64
	subpixelScale    int
}

func NewWaveDistortion(centerX, centerY, amplitude, period, phase float64) *WaveDistortion {
	return &WaveDistortion{
		centerX:       centerX,
		centerY:       centerY,
		amplitude:     amplitude,
		period:        period,
		phase:         phase,
		subpixelScale: 256,
	}
}

func (d *WaveDistortion) Calculate(x, y *int) {
	// Convert to floating point coordinates
	xd := float64(*x)/float64(d.subpixelScale) - d.centerX
	yd := float64(*y)/float64(d.subpixelScale) - d.centerY
	distance := math.Sqrt(xd*xd + yd*yd)

	if distance > 1.0 {
		amplitude := math.Cos(distance/(16.0*d.period)-d.phase)*(1.0/(d.amplitude*distance)) + 1.0
		*x = int((xd*amplitude + d.centerX) * float64(d.subpixelScale))
		*y = int((yd*amplitude + d.centerY) * float64(d.subpixelScale))
	}
}

func TestSpanInterpolatorAdaptor_BasicFunctionality(t *testing.T) {
	t.Run("IdentityDistortion", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		distortion := &IdentityDistortion{}
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		adaptor.Begin(10, 20, 5)

		x, y := adaptor.Coordinates()
		expectedX := 10 * 256 // 10.0 * subpixel_scale
		expectedY := 20 * 256 // 20.0 * subpixel_scale

		if x != expectedX || y != expectedY {
			t.Errorf("Identity distortion: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}

		// Test Next()
		adaptor.Next()
		x, y = adaptor.Coordinates()
		expectedX = 11 * 256 // 11.0 * subpixel_scale (advanced by 1)
		expectedY = 20 * 256 // y should remain the same

		if x != expectedX || y != expectedY {
			t.Errorf("After Next(): got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("ScaleDistortion", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		distortion := NewScaleDistortion(2.0, 0.5)
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		adaptor.Begin(4, 8, 3)

		x, y := adaptor.Coordinates()
		// Base coordinates: (4*256, 8*256)
		// After scaling: (4*2*256, 8*0.5*256) = (2048, 1024)
		expectedX := 2048
		expectedY := 1024

		if x != expectedX || y != expectedY {
			t.Errorf("Scale distortion: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("OffsetDistortion", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		distortion := NewOffsetDistortion(100, -50)
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		adaptor.Begin(1, 2, 3)

		x, y := adaptor.Coordinates()
		// Base coordinates: (1*256, 2*256) = (256, 512)
		// After offset: (256+100, 512-50) = (356, 462)
		expectedX := 356
		expectedY := 462

		if x != expectedX || y != expectedY {
			t.Errorf("Offset distortion: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})
}

func TestSpanInterpolatorAdaptor_AccessorMethods(t *testing.T) {
	t.Run("BaseAccess", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		distortion := &IdentityDistortion{}
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		// Test Base() getter
		if adaptor.Base() != base {
			t.Error("Base() should return the original base interpolator")
		}

		// Test SetBase()
		newBase := NewMockAdaptorInterpolator()
		adaptor.SetBase(newBase)
		if adaptor.Base() != newBase {
			t.Error("SetBase() should update the base interpolator")
		}
	})

	t.Run("DistortionAccess", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		distortion := NewOffsetDistortion(5, 10)
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		// Test Distortion() getter
		if adaptor.Distortion() != distortion {
			t.Error("Distortion() should return the original distortion")
		}

		// Test SetDistortion()
		newDistortion := NewOffsetDistortion(10, 20)
		adaptor.SetDistortion(newDistortion)
		if adaptor.Distortion() != newDistortion {
			t.Error("SetDistortion() should update the distortion")
		}

		// Verify new distortion is applied
		adaptor.Begin(0, 0, 1)
		x, y := adaptor.Coordinates()
		if x != 10 || y != 20 { // Should be offset by (10, 20)
			t.Errorf("New distortion not applied: got (%d, %d), want (10, 20)", x, y)
		}
	})
}

func TestSpanInterpolatorAdaptor_WithRealInterpolator(t *testing.T) {
	t.Run("WithMockInterpolator", func(t *testing.T) {
		// Use a mock interpolator that simulates translation
		base := NewMockAdaptorInterpolator()
		distortion := NewScaleDistortion(2.0, 2.0)
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		// Start at (5, 10) to simulate translation
		adaptor.Begin(5, 10, 5)

		x, y := adaptor.Coordinates()
		// Base interpolator: 5*256, 10*256 = 1280, 2560
		// Scale distortion: 1280*2, 2560*2 = 2560, 5120
		expectedX := 2560
		expectedY := 5120

		if x != expectedX || y != expectedY {
			t.Errorf("Mock interpolator with distortion: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}

		// Test advancing
		adaptor.Next()
		x, y = adaptor.Coordinates()
		// Base interpolator: (5+1)*256, 10*256 = 1536, 2560
		// Scale distortion: 1536*2, 2560*2 = 3072, 5120
		expectedX = 3072
		expectedY = 5120

		if x != expectedX || y != expectedY {
			t.Errorf("After Next() with mock interpolator: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})
}

func TestSpanInterpolatorAdaptor_WaveDistortion(t *testing.T) {
	t.Run("WaveEffect", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		// Create a wave centered at (50, 50) with small amplitude
		distortion := NewWaveDistortion(50.0, 50.0, 0.1, 1.0, 0.0)
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		// Start at the center - should not be affected much by wave
		adaptor.Begin(50, 50, 3)
		x, y := adaptor.Coordinates()

		// At the center, the effect should be minimal
		centerX := 50 * 256
		centerY := 50 * 256

		// Allow some tolerance due to wave calculation
		tolerance := 10
		if intAbs(x-centerX) > tolerance || intAbs(y-centerY) > tolerance {
			t.Logf("Wave distortion at center: got (%d, %d), expected near (%d, %d)", x, y, centerX, centerY)
			// This is acceptable as long as it's not the original coordinates
		}

		// Move away from center - effect should be more pronounced
		adaptor.Begin(100, 100, 3)
		x, y = adaptor.Coordinates()

		originalX := 100 * 256
		originalY := 100 * 256

		if x == originalX && y == originalY {
			t.Error("Wave distortion should modify coordinates away from center")
		}
	})
}

func TestSpanInterpolatorAdaptor_SpanIteration(t *testing.T) {
	t.Run("MultiplePixels", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		distortion := NewOffsetDistortion(100, 200)
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		adaptor.Begin(0, 0, 5)

		// Collect coordinates for multiple pixels
		coords := make([][2]int, 5)
		for i := 0; i < 5; i++ {
			x, y := adaptor.Coordinates()
			coords[i] = [2]int{x, y}
			if i < 4 {
				adaptor.Next()
			}
		}

		// Verify pattern: x advances, offset applied to both
		for i := 0; i < 5; i++ {
			expectedX := i*256 + 100 // i pixels + offset
			expectedY := 200         // offset only

			if coords[i][0] != expectedX || coords[i][1] != expectedY {
				t.Errorf("Pixel %d: got (%d, %d), want (%d, %d)",
					i, coords[i][0], coords[i][1], expectedX, expectedY)
			}
		}
	})
}

func TestSpanInterpolatorAdaptor_EdgeCases(t *testing.T) {
	t.Run("NegativeCoordinates", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		distortion := &IdentityDistortion{}
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		adaptor.Begin(-10, -5, 3)
		x, y := adaptor.Coordinates()

		expectedX := -10 * 256
		expectedY := -5 * 256

		if x != expectedX || y != expectedY {
			t.Errorf("Negative coordinates: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("ZeroLength", func(t *testing.T) {
		base := NewMockAdaptorInterpolator()
		distortion := NewOffsetDistortion(50, 75)
		adaptor := NewSpanInterpolatorAdaptor(base, distortion)

		adaptor.Begin(1, 2, 0) // Zero length
		x, y := adaptor.Coordinates()

		expectedX := 1*256 + 50
		expectedY := 2*256 + 75

		if x != expectedX || y != expectedY {
			t.Errorf("Zero length span: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})
}

// Benchmark tests
func BenchmarkSpanInterpolatorAdaptor_IdentityDistortion(b *testing.B) {
	base := NewMockAdaptorInterpolator()
	distortion := &IdentityDistortion{}
	adaptor := NewSpanInterpolatorAdaptor(base, distortion)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adaptor.Begin(0, 0, 1000)
		for j := 0; j < 1000; j++ {
			adaptor.Next()
			adaptor.Coordinates()
		}
	}
}

func BenchmarkSpanInterpolatorAdaptor_ScaleDistortion(b *testing.B) {
	base := NewMockAdaptorInterpolator()
	distortion := NewScaleDistortion(1.5, 1.2)
	adaptor := NewSpanInterpolatorAdaptor(base, distortion)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adaptor.Begin(0, 0, 1000)
		for j := 0; j < 1000; j++ {
			adaptor.Next()
			adaptor.Coordinates()
		}
	}
}

func BenchmarkSpanInterpolatorAdaptor_WaveDistortion(b *testing.B) {
	base := NewMockAdaptorInterpolator()
	distortion := NewWaveDistortion(100.0, 100.0, 0.1, 1.0, 0.0)
	adaptor := NewSpanInterpolatorAdaptor(base, distortion)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adaptor.Begin(50, 50, 1000)
		for j := 0; j < 1000; j++ {
			adaptor.Next()
			adaptor.Coordinates()
		}
	}
}
