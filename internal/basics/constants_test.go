package basics

import (
	"math"
	"testing"
)

func TestSaturation(t *testing.T) {
	t.Run("Apply method", func(t *testing.T) {
		saturation := NewSaturationInt(100)

		tests := []struct {
			input    int
			expected int
		}{
			{50, 50},   // Below limit
			{100, 100}, // At limit
			{150, 100}, // Above limit
			{0, 0},     // Zero
			{-50, -50}, // Negative below limit
		}

		for _, tt := range tests {
			result := saturation.Apply(tt.input)
			if result != tt.expected {
				t.Errorf("Apply(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("IRound method", func(t *testing.T) {
		saturation := NewSaturationInt(127)

		tests := []struct {
			input    float64
			expected int
		}{
			{50.3, 50},     // Normal rounding
			{50.7, 51},     // Normal rounding up
			{150.0, 127},   // Above positive limit
			{-150.0, -127}, // Below negative limit
			{-50.3, -50},   // Negative normal rounding
			{-50.7, -51},   // Negative normal rounding down
			{0.0, 0},       // Zero
		}

		for _, tt := range tests {
			result := saturation.IRound(tt.input)
			if result != tt.expected {
				t.Errorf("IRound(%f) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Different types", func(t *testing.T) {
		// Test with int32
		sat32 := NewSaturationInt32(1000)
		result32 := sat32.Apply(1500)
		if result32 != 1000 {
			t.Errorf("int32 saturation failed: expected 1000, got %d", result32)
		}

		// Test with uint32
		satU32 := NewSaturationUint32(255)
		resultU32 := satU32.Apply(300)
		if resultU32 != 255 {
			t.Errorf("uint32 saturation failed: expected 255, got %d", resultU32)
		}
	})
}

func TestMulOne(t *testing.T) {
	t.Run("Apply method", func(t *testing.T) {
		mulOne := NewMulOne[uint32](8)

		tests := []struct {
			input    uint32
			expected uint32
		}{
			{256, 1},  // 256 >> 8 = 1
			{512, 2},  // 512 >> 8 = 2
			{1024, 4}, // 1024 >> 8 = 4
			{255, 0},  // 255 >> 8 = 0
			{0, 0},    // 0 >> 8 = 0
		}

		for _, tt := range tests {
			result := mulOne.Apply(tt.input)
			if result != tt.expected {
				t.Errorf("Apply(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Mul method", func(t *testing.T) {
		mulOne := NewMulOne[uint32](8) // shift = 8, so we're working with 8-bit precision

		tests := []struct {
			a, b     uint32
			expected uint32
		}{
			{255, 255, 255}, // Full scale multiplication
			{255, 128, 128}, // Half scale
			{128, 128, 64},  // Quarter result
			{0, 255, 0},     // Zero multiplication
		}

		for _, tt := range tests {
			result := mulOne.Mul(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Mul(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		}
	})

	t.Run("Different shift values", func(t *testing.T) {
		// Test with shift = 4
		mulOne4 := NewMulOne[uint32](4)
		result4 := mulOne4.Apply(32) // 32 >> 4 = 2
		if result4 != 2 {
			t.Errorf("Shift 4: expected 2, got %d", result4)
		}

		// Test with shift = 16
		mulOne16 := NewMulOne[uint32](16)
		result16 := mulOne16.Apply(65536) // 65536 >> 16 = 1
		if result16 != 1 {
			t.Errorf("Shift 16: expected 1, got %d", result16)
		}
	})

	t.Run("Mul precision test", func(t *testing.T) {
		mulOne := NewMulOne[uint32](8)

		// Test the multiplication formula: (a * b + (1 << (shift-1))) >> shift
		// For shift=8: (a * b + 128) >> 8, but then add rounding: (q + (q >> shift)) >> shift
		a, b := uint32(200), uint32(150)
		q := a*b + (1 << (8 - 1))       // q = 30000 + 128 = 30128
		expected := (q + (q >> 8)) >> 8 // expected = (30128 + 117) >> 8 = 30245 >> 8 = 118
		result := mulOne.Mul(a, b)

		if result != expected {
			t.Errorf("Mul precision test failed: expected %d, got %d", expected, result)
		}
	})
}

func TestRoundingFunctions(t *testing.T) {
	t.Run("IRound", func(t *testing.T) {
		tests := []struct {
			input    float64
			expected int
		}{
			{3.4, 3},
			{3.5, 4},
			{3.6, 4},
			{-3.4, -3},
			{-3.5, -4},
			{-3.6, -4},
			{0.0, 0},
			{0.5, 1},
			{-0.5, -1},
		}

		for _, tt := range tests {
			result := IRound(tt.input)
			if result != tt.expected {
				t.Errorf("IRound(%f) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("URound", func(t *testing.T) {
		tests := []struct {
			input    float64
			expected uint32
		}{
			{3.4, 3},
			{3.5, 4},
			{3.6, 4},
			{0.0, 0},
			{0.5, 1},
			{-1.0, 0}, // Negative values should return 0
		}

		for _, tt := range tests {
			result := URound(tt.input)
			if result != tt.expected {
				t.Errorf("URound(%f) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("IFloor", func(t *testing.T) {
		tests := []struct {
			input    float64
			expected int
		}{
			{3.9, 3},
			{3.1, 3},
			{-3.1, -4},
			{-3.9, -4},
			{0.0, 0},
		}

		for _, tt := range tests {
			result := IFloor(tt.input)
			if result != tt.expected {
				t.Errorf("IFloor(%f) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("ICeil", func(t *testing.T) {
		tests := []struct {
			input    float64
			expected int
		}{
			{3.1, 4},
			{3.9, 4},
			{-3.1, -3},
			{-3.9, -3},
			{0.0, 0},
			{3.0, 3}, // Exact integer
		}

		for _, tt := range tests {
			result := ICeil(tt.input)
			if result != tt.expected {
				t.Errorf("ICeil(%f) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})
}

func TestConversionFunctions(t *testing.T) {
	t.Run("Deg2RadF", func(t *testing.T) {
		tests := []struct {
			degrees   float64
			expected  float64
			tolerance float64
		}{
			{0, 0, 1e-10},
			{90, math.Pi / 2, 1e-10},
			{180, math.Pi, 1e-10},
			{360, 2 * math.Pi, 1e-10},
			{-90, -math.Pi / 2, 1e-10},
		}

		for _, tt := range tests {
			result := Deg2RadF(tt.degrees)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("Deg2RadF(%f) = %f, want %f", tt.degrees, result, tt.expected)
			}
		}
	})

	t.Run("Rad2DegF", func(t *testing.T) {
		tests := []struct {
			radians   float64
			expected  float64
			tolerance float64
		}{
			{0, 0, 1e-10},
			{math.Pi / 2, 90, 1e-10},
			{math.Pi, 180, 1e-10},
			{2 * math.Pi, 360, 1e-10},
			{-math.Pi / 2, -90, 1e-10},
		}

		for _, tt := range tests {
			result := Rad2DegF(tt.radians)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("Rad2DegF(%f) = %f, want %f", tt.radians, result, tt.expected)
			}
		}
	})
}

func TestFillingRule(t *testing.T) {
	t.Run("Filling rule constants", func(t *testing.T) {
		if FillNonZero != 0 {
			t.Errorf("FillNonZero should be 0, got %d", FillNonZero)
		}
		if FillEvenOdd != 1 {
			t.Errorf("FillEvenOdd should be 1, got %d", FillEvenOdd)
		}
	})
}

func TestMathConstants(t *testing.T) {
	if VertexDistEpsilon != 1e-14 {
		t.Errorf("VertexDistEpsilon expected 1e-14, got %g", VertexDistEpsilon)
	}
	if IntersectionEpsilon != 1.0e-30 {
		t.Errorf("IntersectionEpsilon expected 1.0e-30, got %g", IntersectionEpsilon)
	}
}

// Benchmark tests
func BenchmarkSaturationIRound(b *testing.B) {
	saturation := NewSaturationInt(127)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		saturation.IRound(float64(i % 1000))
	}
}

func BenchmarkMulOneMul(b *testing.B) {
	mulOne := NewMulOne[uint32](8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mulOne.Mul(uint32(i%256), uint32((i+1)%256))
	}
}

func BenchmarkIRound(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IRound(float64(i) + 0.5)
	}
}

// Additional tests merged from constants_extra_test.go

func TestUFloorAndUCeil(t *testing.T) {
	t.Run("UFloor", func(t *testing.T) {
		tests := []struct {
			in       float64
			expected uint32
		}{
			{3.9, 3},
			{3.1, 3},
			{0.0, 0},
			{7.0, 7},  // exact integer
			{-3.1, 0}, // negative clamps to 0
			{-0.1, 0}, // negative clamps to 0
		}
		for _, tt := range tests {
			got := UFloor(tt.in)
			if got != tt.expected {
				t.Errorf("UFloor(%v) = %d, want %d", tt.in, got, tt.expected)
			}
		}
	})

	t.Run("UCeil", func(t *testing.T) {
		tests := []struct {
			in       float64
			expected uint32
		}{
			{3.1, 4},
			{3.9, 4},
			{0.0, 0},
			{7.0, 7},  // exact integer
			{-3.1, 0}, // negative clamps to 0
			{-0.1, 0}, // negative clamps to 0
		}
		for _, tt := range tests {
			got := UCeil(tt.in)
			if got != tt.expected {
				t.Errorf("UCeil(%v) = %d, want %d", tt.in, got, tt.expected)
			}
		}
	})
}

func TestSaturationUnsignedIRound(t *testing.T) {
	// For unsigned saturation, lower bound should clamp to 0
	sat := NewSaturationUint32(255)

	tests := []struct {
		in       float64
		expected uint32
	}{
		{-1.2, 0},    // negative clamps to 0
		{-300.0, 0},  // below -limit clamps to 0
		{0.0, 0},     // zero
		{5.4, 5},     // standard rounding down
		{5.5, 6},     // standard rounding up
		{300.0, 255}, // above limit clamps to limit
	}

	for _, tt := range tests {
		got := sat.IRound(tt.in)
		if got != tt.expected {
			t.Errorf("SaturationUint32.IRound(%v) = %d, want %d", tt.in, got, tt.expected)
		}
	}
}

func TestPiConstant(t *testing.T) {
	if Pi != math.Pi {
		t.Errorf("Pi mismatch: got %v, want %v", Pi, math.Pi)
	}
}
