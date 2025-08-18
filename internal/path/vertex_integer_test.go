package path

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestVertexInteger_NewVertexInteger(t *testing.T) {
	tests := []struct {
		name     string
		x, y     int32
		cmd      uint32
		expected VertexInteger[int32]
	}{
		{
			name: "MoveTo command",
			x:    10, y: 20, cmd: CmdMoveTo,
			expected: VertexInteger[int32]{X: 20, Y: 40}, // x<<1 | (cmd&1), y<<1 | ((cmd>>1)&1)
		},
		{
			name: "LineTo command",
			x:    30, y: 40, cmd: CmdLineTo,
			expected: VertexInteger[int32]{X: 61, Y: 80}, // x<<1 | (cmd&1), y<<1 | ((cmd>>1)&1)
		},
		{
			name: "Curve3 command",
			x:    50, y: 60, cmd: CmdCurve3,
			expected: VertexInteger[int32]{X: 100, Y: 121}, // x<<1 | (cmd&1), y<<1 | ((cmd>>1)&1)
		},
		{
			name: "Curve4 command",
			x:    70, y: 80, cmd: CmdCurve4,
			expected: VertexInteger[int32]{X: 141, Y: 161}, // x<<1 | (cmd&1), y<<1 | ((cmd>>1)&1)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vertex := NewVertexInteger(tt.x, tt.y, tt.cmd)
			if vertex.X != tt.expected.X {
				t.Errorf("X = %d, expected %d", vertex.X, tt.expected.X)
			}
			if vertex.Y != tt.expected.Y {
				t.Errorf("Y = %d, expected %d", vertex.Y, tt.expected.Y)
			}
		})
	}
}

func TestVertexInteger_Vertex(t *testing.T) {
	tests := []struct {
		name        string
		vertex      VertexInteger[int32]
		dx, dy      float64
		scale       float64
		coordShift  int
		expectedX   float64
		expectedY   float64
		expectedCmd basics.PathCommand
	}{
		{
			name:   "MoveTo with default shift",
			vertex: NewVertexIntegerFromFloat[int32](64.0, 128.0, CmdMoveTo, DefaultCoordShift),
			dx:     0, dy: 0, scale: 1.0, coordShift: DefaultCoordShift,
			expectedX: 64.0, expectedY: 128.0, expectedCmd: basics.PathCmdMoveTo,
		},
		{
			name:   "LineTo with offset and scale",
			vertex: NewVertexIntegerFromFloat[int32](32.0, 64.0, CmdLineTo, DefaultCoordShift),
			dx:     10, dy: 20, scale: 2.0, coordShift: DefaultCoordShift,
			expectedX: 74.0, expectedY: 148.0, expectedCmd: basics.PathCmdLineTo, // 10 + (32 * 2), 20 + (64 * 2)
		},
		{
			name:   "Curve3 with custom shift",
			vertex: NewVertexIntegerFromFloat[int32](16.0, 32.0, CmdCurve3, 4),
			dx:     0, dy: 0, scale: 1.0, coordShift: 4, // shift=4 means scale=16
			expectedX: 16.0, expectedY: 32.0, expectedCmd: basics.PathCmdCurve3,
		},
		{
			name:   "Curve4 with all parameters",
			vertex: NewVertexIntegerFromFloat[int32](8.0, 16.0, CmdCurve4, 3),
			dx:     5.0, dy: 10.0, scale: 0.5, coordShift: 3, // shift=3 means scale=8
			expectedX: 9.0, expectedY: 18.0, expectedCmd: basics.PathCmdCurve4, // 5 + (8 * 0.5), 10 + (16 * 0.5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, cmd := tt.vertex.Vertex(tt.dx, tt.dy, tt.scale, tt.coordShift)

			if math.Abs(x-tt.expectedX) > 1e-9 {
				t.Errorf("X = %f, expected %f", x, tt.expectedX)
			}
			if math.Abs(y-tt.expectedY) > 1e-9 {
				t.Errorf("Y = %f, expected %f", y, tt.expectedY)
			}
			if cmd != tt.expectedCmd {
				t.Errorf("Command = %v, expected %v", cmd, tt.expectedCmd)
			}
		})
	}
}

func TestVertexInteger_VertexSimple(t *testing.T) {
	vertex := NewVertexIntegerFromFloat[int32](64.0, 128.0, CmdMoveTo, DefaultCoordShift)
	x, y, cmd := vertex.VertexSimple()

	expectedX := 64.0
	expectedY := 128.0
	expectedCmd := basics.PathCmdMoveTo

	if math.Abs(x-expectedX) > 1e-9 {
		t.Errorf("X = %f, expected %f", x, expectedX)
	}
	if math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Y = %f, expected %f", y, expectedY)
	}
	if cmd != expectedCmd {
		t.Errorf("Command = %v, expected %v", cmd, expectedCmd)
	}
}

func TestVertexInteger_DifferentTypes(t *testing.T) {
	// Test with int16
	vertex16 := NewVertexIntegerFromFloat[int16](100.0, 200.0, CmdLineTo, DefaultCoordShift)
	x16, y16, cmd16 := vertex16.VertexSimple()
	if math.Abs(x16-100.0) > 1e-9 || math.Abs(y16-200.0) > 1e-9 || cmd16 != basics.PathCmdLineTo {
		t.Errorf("int16 test failed: x=%f, y=%f, cmd=%v", x16, y16, cmd16)
	}

	// Test with int64
	vertex64 := NewVertexIntegerFromFloat[int64](300.0, 400.0, CmdCurve3, DefaultCoordShift)
	x64, y64, cmd64 := vertex64.VertexSimple()
	if math.Abs(x64-300.0) > 1e-9 || math.Abs(y64-400.0) > 1e-9 || cmd64 != basics.PathCmdCurve3 {
		t.Errorf("int64 test failed: x=%f, y=%f, cmd=%v", x64, y64, cmd64)
	}
}

func TestVertexInteger_CommandPacking(t *testing.T) {
	// Test that commands are properly packed and unpacked from coordinate bits
	tests := []struct {
		cmd      uint32
		expected basics.PathCommand
	}{
		{CmdMoveTo, basics.PathCmdMoveTo},
		{CmdLineTo, basics.PathCmdLineTo},
		{CmdCurve3, basics.PathCmdCurve3},
		{CmdCurve4, basics.PathCmdCurve4},
	}

	for _, tt := range tests {
		vertex := NewVertexInteger[int32](0, 0, tt.cmd)
		_, _, cmd := vertex.VertexSimple()
		if cmd != tt.expected {
			t.Errorf("Command packing failed: input=%d, got=%v, expected=%v",
				tt.cmd, cmd, tt.expected)
		}
	}
}

func TestVertexInteger_NegativeCoordinates(t *testing.T) {
	// Test with negative coordinates
	vertex := NewVertexIntegerFromFloat[int32](-64.0, -128.0, CmdMoveTo, DefaultCoordShift)
	x, y, cmd := vertex.VertexSimple()

	expectedX := -64.0
	expectedY := -128.0
	expectedCmd := basics.PathCmdMoveTo

	if math.Abs(x-expectedX) > 1e-9 {
		t.Errorf("Negative X = %f, expected %f", x, expectedX)
	}
	if math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Negative Y = %f, expected %f", y, expectedY)
	}
	if cmd != expectedCmd {
		t.Errorf("Command = %v, expected %v", cmd, expectedCmd)
	}
}

func TestVertexInteger_CoordinateShiftPrecision(t *testing.T) {
	// Test different coordinate shift values for precision
	tests := []struct {
		shift    int
		x, y     int32
		expected float64 // expected scale factor
	}{
		{shift: 2, x: 4, y: 4, expected: 1.0},     // shift=2 -> scale=4
		{shift: 4, x: 16, y: 16, expected: 1.0},   // shift=4 -> scale=16
		{shift: 6, x: 64, y: 64, expected: 1.0},   // shift=6 -> scale=64 (default)
		{shift: 8, x: 256, y: 256, expected: 1.0}, // shift=8 -> scale=256
	}

	for _, tt := range tests {
		vertex := NewVertexIntegerFromFloat[int32](tt.expected, tt.expected, CmdMoveTo, tt.shift)
		x, y, _ := vertex.Vertex(0, 0, 1.0, tt.shift)

		if math.Abs(x-tt.expected) > 1e-9 {
			t.Errorf("Shift %d: X = %f, expected %f", tt.shift, x, tt.expected)
		}
		if math.Abs(y-tt.expected) > 1e-9 {
			t.Errorf("Shift %d: Y = %f, expected %f", tt.shift, y, tt.expected)
		}
	}
}

// Benchmark tests
func BenchmarkVertexInteger_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewVertexInteger[int32](int32(i), int32(i*2), CmdLineTo)
	}
}

func BenchmarkVertexInteger_Vertex(b *testing.B) {
	vertex := NewVertexInteger[int32](100, 200, CmdLineTo)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = vertex.Vertex(0, 0, 1.0, DefaultCoordShift)
	}
}

func BenchmarkVertexInteger_VertexSimple(b *testing.B) {
	vertex := NewVertexInteger[int32](100, 200, CmdLineTo)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = vertex.VertexSimple()
	}
}
