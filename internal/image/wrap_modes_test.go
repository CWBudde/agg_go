package image

import (
	"testing"

	"agg_go/internal/basics"
)

func TestWrapModeRepeat(t *testing.T) {
	tests := []struct {
		name     string
		size     basics.Int32u
		input    int
		expected basics.Int32u
	}{
		{"size 10, positive", 10, 15, 5},
		{"size 10, negative", 10, -5, 5},
		{"size 10, zero", 10, 0, 0},
		{"size 10, exact multiple", 10, 20, 0},
		{"size 8, wrap around", 8, 11, 3},
		{"size 1, any value", 1, 999, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrap := NewWrapModeRepeat(tt.size)
			result := wrap.Call(tt.input)
			if result != tt.expected {
				t.Errorf("WrapModeRepeat.Call(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWrapModeRepeat_Inc(t *testing.T) {
	wrap := NewWrapModeRepeat(5)

	// Test incrementing through a cycle
	// Inc() increments from current position, so starting at 0, we get 1,2,3,4,0,1,2
	expected := []basics.Int32u{1, 2, 3, 4, 0, 1, 2}

	wrap.Call(0) // Start at 0
	for i, exp := range expected {
		result := wrap.Inc()
		if result != exp {
			t.Errorf("Inc() step %d: got %d, want %d", i, result, exp)
		}
	}
}

func TestWrapModeRepeatPow2(t *testing.T) {
	tests := []struct {
		name     string
		size     basics.Int32u
		input    int
		expected basics.Int32u
	}{
		{"size 8, positive", 8, 15, 7},
		{"size 8, within bounds", 8, 3, 3},
		{"size 16, large value", 16, 31, 15},
		{"size 4, exact boundary", 4, 4, 0},
		{"size 2, alternating", 2, 5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrap := NewWrapModeRepeatPow2(tt.size)
			result := wrap.Call(tt.input)
			if result != tt.expected {
				t.Errorf("WrapModeRepeatPow2.Call(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWrapModeRepeatPow2_Inc(t *testing.T) {
	wrap := NewWrapModeRepeatPow2(4)

	// Test incrementing through a cycle
	// Inc() increments from current position, so starting at 0, we get 1,2,3,0,1,2
	expected := []basics.Int32u{1, 2, 3, 0, 1, 2}

	wrap.Call(0) // Start at 0
	for i, exp := range expected {
		result := wrap.Inc()
		if result != exp {
			t.Errorf("Inc() step %d: got %d, want %d", i, result, exp)
		}
	}
}

func TestWrapModeRepeatAutoPow2(t *testing.T) {
	// Test power-of-2 size (should use bitmasking)
	t.Run("power of 2", func(t *testing.T) {
		wrap := NewWrapModeRepeatAutoPow2(8)
		result := wrap.Call(15)
		expected := basics.Int32u(7)
		if result != expected {
			t.Errorf("Auto pow2 with size 8: got %d, want %d", result, expected)
		}
	})

	// Test non-power-of-2 size (should use modulo)
	t.Run("non power of 2", func(t *testing.T) {
		wrap := NewWrapModeRepeatAutoPow2(10)
		result := wrap.Call(15)
		expected := basics.Int32u(5)
		if result != expected {
			t.Errorf("Auto pow2 with size 10: got %d, want %d", result, expected)
		}
	})
}

func TestWrapModeReflect(t *testing.T) {
	tests := []struct {
		name     string
		size     basics.Int32u
		input    int
		expected basics.Int32u
	}{
		{"size 5, forward", 5, 3, 3},
		{"size 5, first reflect", 5, 7, 2}, // 5+2 -> reflects to 5-2-1 = 2
		{"size 5, second cycle", 5, 12, 2},
		{"size 4, exact boundary", 4, 4, 3}, // reflects to 4-1 = 3
		{"size 3, multiple reflections", 3, 8, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrap := NewWrapModeReflect(tt.size)
			result := wrap.Call(tt.input)
			if result != tt.expected {
				t.Errorf("WrapModeReflect.Call(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWrapModeReflect_Inc(t *testing.T) {
	wrap := NewWrapModeReflect(3)

	// Test incrementing through reflect cycle
	// For size=3, internal value sequence: 0,1,2,3,4,5,0,1,2...
	// Mapped output: 0,1,2,2,1,0,0,1,2...
	// Starting at 0, Inc() gives: 1,2,2,1,0,0...
	expected := []basics.Int32u{1, 2, 2, 1, 0, 0}

	wrap.Call(0) // Start at 0
	for i, exp := range expected {
		result := wrap.Inc()
		if result != exp {
			t.Errorf("Inc() step %d: got %d, want %d", i, result, exp)
		}
	}
}

func TestWrapModeReflectPow2(t *testing.T) {
	tests := []struct {
		name     string
		size     basics.Int32u
		input    int
		expected basics.Int32u
	}{
		{"size 4, forward", 4, 2, 2},
		{"size 4, reflect", 4, 6, 1}, // reflects
		{"size 8, large value", 8, 10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrap := NewWrapModeReflectPow2(tt.size)
			result := wrap.Call(tt.input)
			if result != tt.expected {
				t.Errorf("WrapModeReflectPow2.Call(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWrapModeReflectAutoPow2(t *testing.T) {
	// Test power-of-2 size
	t.Run("power of 2", func(t *testing.T) {
		wrap := NewWrapModeReflectAutoPow2(4)
		result := wrap.Call(6)
		expected := basics.Int32u(1) // Should reflect
		if result != expected {
			t.Errorf("Auto reflect pow2 with size 4: got %d, want %d", result, expected)
		}
	})

	// Test non-power-of-2 size
	t.Run("non power of 2", func(t *testing.T) {
		wrap := NewWrapModeReflectAutoPow2(3)
		result := wrap.Call(4)
		expected := basics.Int32u(1) // 4%6=4, 4>=3 so 6-4-1=1
		if result != expected {
			t.Errorf("Auto reflect pow2 with size 3: got %d, want %d", result, expected)
		}
	})
}

// Benchmark tests to verify performance improvements of optimized variants
func BenchmarkWrapModeRepeat(b *testing.B) {
	wrap := NewWrapModeRepeat(100)
	for i := 0; i < b.N; i++ {
		wrap.Call(i * 7) // Use varying input
	}
}

func BenchmarkWrapModeRepeatPow2(b *testing.B) {
	wrap := NewWrapModeRepeatPow2(128) // Power of 2
	for i := 0; i < b.N; i++ {
		wrap.Call(i * 7)
	}
}

func BenchmarkWrapModeRepeatAutoPow2_Pow2(b *testing.B) {
	wrap := NewWrapModeRepeatAutoPow2(128) // Power of 2
	for i := 0; i < b.N; i++ {
		wrap.Call(i * 7)
	}
}

func BenchmarkWrapModeRepeatAutoPow2_NonPow2(b *testing.B) {
	wrap := NewWrapModeRepeatAutoPow2(100) // Non-power of 2
	for i := 0; i < b.N; i++ {
		wrap.Call(i * 7)
	}
}

func BenchmarkWrapModeReflect(b *testing.B) {
	wrap := NewWrapModeReflect(100)
	for i := 0; i < b.N; i++ {
		wrap.Call(i * 7)
	}
}

func BenchmarkWrapModeReflectPow2(b *testing.B) {
	wrap := NewWrapModeReflectPow2(128)
	for i := 0; i < b.N; i++ {
		wrap.Call(i * 7)
	}
}

// Test edge cases and boundary conditions
func TestWrapModes_EdgeCases(t *testing.T) {
	t.Run("size 1", func(t *testing.T) {
		wrap := NewWrapModeRepeat(1)
		for i := -10; i <= 10; i++ {
			result := wrap.Call(i)
			if result != 0 {
				t.Errorf("Size 1 should always return 0, got %d for input %d", result, i)
			}
		}
	})

	t.Run("negative inputs", func(t *testing.T) {
		wrap := NewWrapModeRepeat(5)
		result := wrap.Call(-7)
		// Should wrap to positive equivalent
		if result >= 5 {
			t.Errorf("Negative input should wrap to valid range, got %d", result)
		}
	})

	t.Run("large inputs", func(t *testing.T) {
		wrap := NewWrapModeRepeat(10)
		result := wrap.Call(1000000)
		if result >= 10 {
			t.Errorf("Large input should wrap to valid range, got %d", result)
		}
	})
}
