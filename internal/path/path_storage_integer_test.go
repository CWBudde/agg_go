package path

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestPathStorageInteger_NewPathStorageInteger(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	if storage == nil {
		t.Fatal("NewPathStorageInteger returned nil")
	}

	if storage.Size() != 0 {
		t.Errorf("New storage should be empty, got size %d", storage.Size())
	}

	if storage.coordShift != DefaultCoordShift {
		t.Errorf("Expected coordShift %d, got %d", DefaultCoordShift, storage.coordShift)
	}
}

func TestPathStorageInteger_NewPathStorageIntegerWithShift(t *testing.T) {
	customShift := 4
	storage := NewPathStorageIntegerWithShift[int32](customShift)

	if storage == nil {
		t.Fatal("NewPathStorageIntegerWithShift returned nil")
	}

	if storage.coordShift != customShift {
		t.Errorf("Expected coordShift %d, got %d", customShift, storage.coordShift)
	}
}

func TestPathStorageInteger_MoveTo(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	storage.MoveTo(100, 200)

	if storage.Size() != 1 {
		t.Errorf("Expected size 1 after MoveTo, got %d", storage.Size())
	}

	x, y, cmd := storage.Vertex(0)
	expectedX := 100.0 / 64.0 // 100 / DefaultCoordScale
	expectedY := 200.0 / 64.0 // 200 / DefaultCoordScale

	if math.Abs(x-expectedX) > 1e-9 {
		t.Errorf("X = %f, expected %f", x, expectedX)
	}
	if math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Y = %f, expected %f", y, expectedY)
	}
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Command = %v, expected %v", cmd, basics.PathCmdMoveTo)
	}
}

func TestPathStorageInteger_LineTo(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	storage.LineTo(300, 400)

	if storage.Size() != 1 {
		t.Errorf("Expected size 1 after LineTo, got %d", storage.Size())
	}

	x, y, cmd := storage.Vertex(0)
	expectedX := 300.0 / 64.0
	expectedY := 400.0 / 64.0

	if math.Abs(x-expectedX) > 1e-9 {
		t.Errorf("X = %f, expected %f", x, expectedX)
	}
	if math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Y = %f, expected %f", y, expectedY)
	}
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Command = %v, expected %v", cmd, basics.PathCmdLineTo)
	}
}

func TestPathStorageInteger_Curve3(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	storage.Curve3(100, 200, 300, 400)

	if storage.Size() != 2 {
		t.Errorf("Expected size 2 after Curve3, got %d", storage.Size())
	}

	// Check control point
	x1, y1, cmd1 := storage.Vertex(0)
	expectedX1 := 100.0 / 64.0
	expectedY1 := 200.0 / 64.0

	if math.Abs(x1-expectedX1) > 1e-9 {
		t.Errorf("Control X = %f, expected %f", x1, expectedX1)
	}
	if math.Abs(y1-expectedY1) > 1e-9 {
		t.Errorf("Control Y = %f, expected %f", y1, expectedY1)
	}
	if cmd1 != basics.PathCmdCurve3 {
		t.Errorf("Control Command = %v, expected %v", cmd1, basics.PathCmdCurve3)
	}

	// Check end point
	x2, y2, cmd2 := storage.Vertex(1)
	expectedX2 := 300.0 / 64.0
	expectedY2 := 400.0 / 64.0

	if math.Abs(x2-expectedX2) > 1e-9 {
		t.Errorf("End X = %f, expected %f", x2, expectedX2)
	}
	if math.Abs(y2-expectedY2) > 1e-9 {
		t.Errorf("End Y = %f, expected %f", y2, expectedY2)
	}
	if cmd2 != basics.PathCmdCurve3 {
		t.Errorf("End Command = %v, expected %v", cmd2, basics.PathCmdCurve3)
	}
}

func TestPathStorageInteger_Curve4(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	storage.Curve4(100, 200, 300, 400, 500, 600)

	if storage.Size() != 3 {
		t.Errorf("Expected size 3 after Curve4, got %d", storage.Size())
	}

	// Check all three vertices have Curve4 command
	for i := uint32(0); i < 3; i++ {
		_, _, cmd := storage.Vertex(i)
		if cmd != basics.PathCmdCurve4 {
			t.Errorf("Vertex %d: Command = %v, expected %v", i, cmd, basics.PathCmdCurve4)
		}
	}
}

func TestPathStorageInteger_ComplexPath(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	// Build a complex path
	storage.MoveTo(0, 0)
	storage.LineTo(100, 0)
	storage.LineTo(100, 100)
	storage.Curve3(150, 150, 200, 100)
	storage.Curve4(250, 50, 300, 150, 350, 100)

	expectedSize := uint32(1 + 1 + 1 + 2 + 3) // MoveTo + LineTo + LineTo + Curve3 + Curve4
	if storage.Size() != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, storage.Size())
	}

	// Test vertex iteration
	commands := []basics.PathCommand{
		basics.PathCmdMoveTo, basics.PathCmdLineTo, basics.PathCmdLineTo,
		basics.PathCmdCurve3, basics.PathCmdCurve3,
		basics.PathCmdCurve4, basics.PathCmdCurve4, basics.PathCmdCurve4,
	}

	for i, expectedCmd := range commands {
		_, _, cmd := storage.Vertex(uint32(i))
		if cmd != expectedCmd {
			t.Errorf("Vertex %d: Command = %v, expected %v", i, cmd, expectedCmd)
		}
	}
}

func TestPathStorageInteger_RemoveAll(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	// Add some vertices
	storage.MoveTo(0, 0)
	storage.LineTo(100, 100)

	if storage.Size() != 2 {
		t.Errorf("Expected size 2 before RemoveAll, got %d", storage.Size())
	}

	storage.RemoveAll()

	if storage.Size() != 0 {
		t.Errorf("Expected size 0 after RemoveAll, got %d", storage.Size())
	}
}

func TestPathStorageInteger_BoundingRect(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	// Test empty path
	bounds := storage.BoundingRect()
	if bounds.X1 != 0 || bounds.Y1 != 0 || bounds.X2 != 0 || bounds.Y2 != 0 {
		t.Errorf("Empty path bounds should be (0,0,0,0), got (%f,%f,%f,%f)",
			bounds.X1, bounds.Y1, bounds.X2, bounds.Y2)
	}

	// Add some vertices
	storage.MoveTo(64, 128)   // Will be (1, 2) in float
	storage.LineTo(192, 256)  // Will be (3, 4) in float
	storage.LineTo(-64, -128) // Will be (-1, -2) in float

	bounds = storage.BoundingRect()
	expectedX1, expectedY1 := -1.0, -2.0
	expectedX2, expectedY2 := 3.0, 4.0

	if math.Abs(bounds.X1-expectedX1) > 1e-9 {
		t.Errorf("X1 = %f, expected %f", bounds.X1, expectedX1)
	}
	if math.Abs(bounds.Y1-expectedY1) > 1e-9 {
		t.Errorf("Y1 = %f, expected %f", bounds.Y1, expectedY1)
	}
	if math.Abs(bounds.X2-expectedX2) > 1e-9 {
		t.Errorf("X2 = %f, expected %f", bounds.X2, expectedX2)
	}
	if math.Abs(bounds.Y2-expectedY2) > 1e-9 {
		t.Errorf("Y2 = %f, expected %f", bounds.Y2, expectedY2)
	}
}

func TestPathStorageInteger_Serialization(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	// Build a simple path
	storage.MoveTo(64, 128)
	storage.LineTo(192, 256)

	// Test ByteSize
	expectedSize := storage.Size() * 8 // 2 vertices * 8 bytes per vertex (2 int32s)
	if storage.ByteSize() != expectedSize {
		t.Errorf("ByteSize = %d, expected %d", storage.ByteSize(), expectedSize)
	}

	// Test Serialize
	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	if len(data) != int(expectedSize) {
		t.Errorf("Serialized data length = %d, expected %d", len(data), expectedSize)
	}
}

func TestPathStorageInteger_VertexIteration(t *testing.T) {
	storage := NewPathStorageInteger[int32]()

	// Test empty path iteration
	storage.Rewind(0)
	_, _, cmd := storage.VertexIterate()
	if cmd != basics.PathCmdStop {
		t.Errorf("Empty path should return PathCmdStop, got %v", cmd)
	}

	// Build a simple path
	storage.MoveTo(64, 128)
	storage.LineTo(192, 256)

	// Test iteration
	storage.Rewind(0)

	// First vertex should be MoveTo
	x1, y1, cmd1 := storage.VertexIterate()
	if cmd1 != basics.PathCmdMoveTo {
		t.Errorf("First vertex: expected PathCmdMoveTo, got %v", cmd1)
	}
	if math.Abs(x1-1.0) > 1e-9 || math.Abs(y1-2.0) > 1e-9 {
		t.Errorf("First vertex: coordinates (%f, %f), expected (1.0, 2.0)", x1, y1)
	}

	// Second vertex should be LineTo
	x2, y2, cmd2 := storage.VertexIterate()
	if cmd2 != basics.PathCmdLineTo {
		t.Errorf("Second vertex: expected PathCmdLineTo, got %v", cmd2)
	}
	if math.Abs(x2-3.0) > 1e-9 || math.Abs(y2-4.0) > 1e-9 {
		t.Errorf("Second vertex: coordinates (%f, %f), expected (3.0, 4.0)", x2, y2)
	}

	// Third call should return EndPoly
	_, _, cmd3 := storage.VertexIterate()
	if (cmd3 & basics.PathCmdMask) != basics.PathCmdEndPoly {
		t.Errorf("Third vertex: expected PathCmdEndPoly, got %v", cmd3)
	}

	// Fourth call should return Stop
	_, _, cmd4 := storage.VertexIterate()
	if cmd4 != basics.PathCmdStop {
		t.Errorf("Fourth vertex: expected PathCmdStop, got %v", cmd4)
	}
}

func TestPathStorageInteger_DifferentTypes(t *testing.T) {
	// Test with int16
	storage16 := NewPathStorageInteger[int16]()
	storage16.MoveTo(64, 128)
	if storage16.Size() != 1 {
		t.Errorf("int16 storage size = %d, expected 1", storage16.Size())
	}

	// Test with int64
	storage64 := NewPathStorageInteger[int64]()
	storage64.LineTo(192, 256)
	if storage64.Size() != 1 {
		t.Errorf("int64 storage size = %d, expected 1", storage64.Size())
	}
}

func TestPathStorageInteger_CustomCoordShift(t *testing.T) {
	// Test with custom coordinate shift
	customShift := 4 // scale = 16
	storage := NewPathStorageIntegerWithShift[int32](customShift)

	storage.MoveTo(16, 32) // Will be (1, 2) with scale 16

	x, y, cmd := storage.Vertex(0)
	expectedX := 1.0
	expectedY := 2.0

	if math.Abs(x-expectedX) > 1e-9 {
		t.Errorf("Custom shift X = %f, expected %f", x, expectedX)
	}
	if math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Custom shift Y = %f, expected %f", y, expectedY)
	}
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Custom shift command = %v, expected %v", cmd, basics.PathCmdMoveTo)
	}
}

// Benchmark tests
func BenchmarkPathStorageInteger_MoveTo(b *testing.B) {
	storage := NewPathStorageInteger[int32]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.MoveTo(int32(i), int32(i*2))
	}
}

func BenchmarkPathStorageInteger_LineTo(b *testing.B) {
	storage := NewPathStorageInteger[int32]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.LineTo(int32(i), int32(i*2))
	}
}

func BenchmarkPathStorageInteger_Curve4(b *testing.B) {
	storage := NewPathStorageInteger[int32]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ii := int32(i)
		storage.Curve4(ii, ii*2, ii*3, ii*4, ii*5, ii*6)
	}
}

func BenchmarkPathStorageInteger_VertexIteration(b *testing.B) {
	storage := NewPathStorageInteger[int32]()

	// Build a path with 1000 vertices
	for i := 0; i < 1000; i++ {
		storage.LineTo(int32(i), int32(i*2))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.Rewind(0)
		for {
			_, _, cmd := storage.VertexIterate()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkPathStorageInteger_Serialization(b *testing.B) {
	storage := NewPathStorageInteger[int32]()

	// Build a path with 100 vertices
	for i := 0; i < 100; i++ {
		storage.LineTo(int32(i), int32(i*2))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = storage.Serialize()
	}
}
