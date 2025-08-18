package path

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestSerializedIntegerPathAdaptor_NewSerializedIntegerPathAdaptor(t *testing.T) {
	adaptor := NewSerializedIntegerPathAdaptor[int32]()

	if adaptor == nil {
		t.Fatal("NewSerializedIntegerPathAdaptor returned nil")
	}

	if !adaptor.IsEmpty() {
		t.Error("New adaptor should be empty")
	}

	if adaptor.Size() != 0 {
		t.Errorf("New adaptor size should be 0, got %d", adaptor.Size())
	}
}

func TestSerializedIntegerPathAdaptor_NewSerializedIntegerPathAdaptorWithData(t *testing.T) {
	// Create some test data by serializing a path
	storage := NewPathStorageInteger[int32]()
	storage.MoveTo(64, 128)
	storage.LineTo(192, 256)

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, 10.0, 20.0)

	if adaptor == nil {
		t.Fatal("NewSerializedIntegerPathAdaptorWithData returned nil")
	}

	if adaptor.IsEmpty() {
		t.Error("Adaptor with data should not be empty")
	}

	if adaptor.Size() != 2 {
		t.Errorf("Adaptor size should be 2, got %d", adaptor.Size())
	}
}

func TestSerializedIntegerPathAdaptor_Init(t *testing.T) {
	adaptor := NewSerializedIntegerPathAdaptor[int32]()

	// Create test data
	storage := NewPathStorageInteger[int32]()
	storage.MoveTo(64, 128)

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor.Init(data, 5.0, 10.0, 2.0, 8)

	if adaptor.IsEmpty() {
		t.Error("Adaptor should not be empty after Init")
	}

	if adaptor.Size() != 1 {
		t.Errorf("Adaptor size should be 1, got %d", adaptor.Size())
	}
}

func TestSerializedIntegerPathAdaptor_Vertex_SingleVertex(t *testing.T) {
	// Create a path with a single vertex
	storage := NewPathStorageInteger[int32]()
	storage.MoveTo(64, 128) // Will be (1, 2) in float coordinates

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, 0.0, 0.0)
	adaptor.Rewind(0)

	// First vertex
	x, y, cmd := adaptor.Vertex()
	expectedX := 1.0 // 64 / 64
	expectedY := 2.0 // 128 / 64

	if math.Abs(x-expectedX) > 1e-9 {
		t.Errorf("X = %f, expected %f", x, expectedX)
	}
	if math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Y = %f, expected %f", y, expectedY)
	}
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Command = %v, expected %v", cmd, basics.PathCmdMoveTo)
	}

	// Second call should return EndPoly
	_, _, cmd2 := adaptor.Vertex()
	if (cmd2 & basics.PathCmdMask) != basics.PathCmdEndPoly {
		t.Errorf("Second call: expected PathCmdEndPoly, got %v", cmd2)
	}

	// Third call should return Stop
	_, _, cmd3 := adaptor.Vertex()
	if cmd3 != basics.PathCmdStop {
		t.Errorf("Third call: expected PathCmdStop, got %v", cmd3)
	}
}

func TestSerializedIntegerPathAdaptor_Vertex_WithTransformation(t *testing.T) {
	// Create a path
	storage := NewPathStorageInteger[int32]()
	storage.MoveTo(64, 128) // Will be (1, 2) in float coordinates

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Test with transformation parameters
	dx, dy, scale := 10.0, 20.0, 2.0
	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, dx, dy)
	adaptor.SetTransform(dx, dy, scale)
	adaptor.Rewind(0)

	x, y, cmd := adaptor.Vertex()
	expectedX := dx + (1.0 * scale) // 10 + (1 * 2) = 12
	expectedY := dy + (2.0 * scale) // 20 + (2 * 2) = 24

	if math.Abs(x-expectedX) > 1e-9 {
		t.Errorf("Transformed X = %f, expected %f", x, expectedX)
	}
	if math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Transformed Y = %f, expected %f", y, expectedY)
	}
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Command = %v, expected %v", cmd, basics.PathCmdMoveTo)
	}
}

func TestSerializedIntegerPathAdaptor_ComplexPath(t *testing.T) {
	// Create a complex path
	storage := NewPathStorageInteger[int32]()
	storage.MoveTo(64, 128)
	storage.LineTo(192, 256)
	storage.Curve3(320, 384, 448, 512)

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, 0.0, 0.0)
	adaptor.Rewind(0)

	// Expected commands in order
	expectedCommands := []basics.PathCommand{
		basics.PathCmdMoveTo, basics.PathCmdLineTo,
		basics.PathCmdCurve3, basics.PathCmdCurve3,
		basics.PathCmdEndPoly, basics.PathCmdStop,
	}

	for i, expectedCmd := range expectedCommands {
		_, _, cmd := adaptor.Vertex()
		if i < 4 { // For the actual vertices
			if cmd != expectedCmd {
				t.Errorf("Vertex %d: Command = %v, expected %v", i, cmd, expectedCmd)
			}
		} else { // For EndPoly and Stop
			if i == 4 && (cmd&basics.PathCmdMask) != basics.PathCmdEndPoly {
				t.Errorf("Expected PathCmdEndPoly, got %v", cmd)
			} else if i == 5 && cmd != basics.PathCmdStop {
				t.Errorf("Expected PathCmdStop, got %v", cmd)
			}
		}
	}
}

func TestSerializedIntegerPathAdaptor_MultipleMoveToPolygonClosing(t *testing.T) {
	// Create a path with multiple MoveTo commands to test polygon closing
	storage := NewPathStorageInteger[int32]()
	storage.MoveTo(64, 128)
	storage.LineTo(192, 256)
	storage.LineTo(320, 384)
	storage.MoveTo(448, 512) // This should trigger polygon closing
	storage.LineTo(576, 640)

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, 0.0, 0.0)
	adaptor.Rewind(0)

	// Read vertices until we hit the second MoveTo
	var commands []basics.PathCommand
	for i := 0; i < 10; i++ { // Safety limit
		_, _, cmd := adaptor.Vertex()
		commands = append(commands, cmd)
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// We should see: MoveTo, LineTo, LineTo, EndPoly (from second MoveTo), MoveTo, LineTo, EndPoly, Stop
	expectedPattern := []basics.PathCommand{
		basics.PathCmdMoveTo, basics.PathCmdLineTo, basics.PathCmdLineTo,
	}

	if len(commands) < len(expectedPattern) {
		t.Fatalf("Not enough commands received: %v", commands)
	}

	for i, expected := range expectedPattern {
		if commands[i] != expected {
			t.Errorf("Command %d: got %v, expected %v", i, commands[i], expected)
		}
	}

	// Check that we get an EndPoly before the second MoveTo
	hasEndPoly := false
	for _, cmd := range commands {
		if (cmd & basics.PathCmdMask) == basics.PathCmdEndPoly {
			hasEndPoly = true
			break
		}
	}
	if !hasEndPoly {
		t.Error("Expected to find PathCmdEndPoly for polygon closing")
	}
}

func TestSerializedIntegerPathAdaptor_Rewind(t *testing.T) {
	// Create a path
	storage := NewPathStorageInteger[int32]()
	storage.MoveTo(64, 128)
	storage.LineTo(192, 256)

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, 0.0, 0.0)

	// Read all vertices
	adaptor.Rewind(0)
	_, _, cmd1 := adaptor.Vertex()
	_, _, _ = adaptor.Vertex()
	_, _, _ = adaptor.Vertex() // Should be EndPoly

	// Rewind and read again
	adaptor.Rewind(0)
	_, _, cmdFirst := adaptor.Vertex()

	if cmdFirst != cmd1 {
		t.Errorf("After rewind, first command = %v, expected %v", cmdFirst, cmd1)
	}
}

func TestSerializedIntegerPathAdaptor_EmptyData(t *testing.T) {
	adaptor := NewSerializedIntegerPathAdaptorWithData[int32]([]byte{}, 0.0, 0.0)

	if !adaptor.IsEmpty() {
		t.Error("Adaptor with empty data should be empty")
	}

	_, _, cmd := adaptor.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Empty adaptor should return PathCmdStop, got %v", cmd)
	}
}

func TestSerializedIntegerPathAdaptor_SetTransform(t *testing.T) {
	// Create a path
	storage := NewPathStorageInteger[int32]()
	storage.MoveTo(64, 128)

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, 0.0, 0.0)

	// Test different transformations
	tests := []struct {
		dx, dy, scale        float64
		expectedX, expectedY float64
	}{
		{0.0, 0.0, 1.0, 1.0, 2.0},     // No transformation
		{10.0, 20.0, 1.0, 11.0, 22.0}, // Translation only
		{0.0, 0.0, 2.0, 2.0, 4.0},     // Scale only
		{5.0, 10.0, 0.5, 5.5, 11.0},   // Translation and scale
	}

	for i, test := range tests {
		adaptor.SetTransform(test.dx, test.dy, test.scale)
		adaptor.Rewind(0)

		x, y, _ := adaptor.Vertex()
		if math.Abs(x-test.expectedX) > 1e-9 {
			t.Errorf("Test %d: X = %f, expected %f", i, x, test.expectedX)
		}
		if math.Abs(y-test.expectedY) > 1e-9 {
			t.Errorf("Test %d: Y = %f, expected %f", i, y, test.expectedY)
		}
	}
}

func TestSerializedIntegerPathAdaptor_DifferentTypes(t *testing.T) {
	// Test with int16
	storage16 := NewPathStorageInteger[int16]()
	storage16.MoveTo(64, 128)

	data16, err := storage16.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize int16 storage: %v", err)
	}

	adaptor16 := NewSerializedIntegerPathAdaptorWithData[int16](data16, 0.0, 0.0)
	if adaptor16.Size() != 1 {
		t.Errorf("int16 adaptor size = %d, expected 1", adaptor16.Size())
	}

	// Test with int64
	storage64 := NewPathStorageInteger[int64]()
	storage64.MoveTo(64, 128)

	data64, err := storage64.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize int64 storage: %v", err)
	}

	adaptor64 := NewSerializedIntegerPathAdaptorWithData[int64](data64, 0.0, 0.0)
	if adaptor64.Size() != 1 {
		t.Errorf("int64 adaptor size = %d, expected 1", adaptor64.Size())
	}
}

func TestSerializedIntegerPathAdaptor_CustomCoordShift(t *testing.T) {
	// Create a path with custom coordinate shift
	customShift := 4
	storage := NewPathStorageIntegerWithShift[int32](customShift)
	storage.MoveTo(16, 32) // Will be (1, 2) with scale 16

	data, err := storage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptor[int32]()
	adaptor.Init(data, 0.0, 0.0, 1.0, customShift)

	x, y, cmd := adaptor.Vertex()
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
func BenchmarkSerializedIntegerPathAdaptor_Vertex(b *testing.B) {
	// Create a path with many vertices
	storage := NewPathStorageInteger[int32]()
	for i := 0; i < 1000; i++ {
		storage.LineTo(int32(i*64), int32(i*128))
	}

	data, err := storage.Serialize()
	if err != nil {
		b.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, 0.0, 0.0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		adaptor.Rewind(0)
		for {
			_, _, cmd := adaptor.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkSerializedIntegerPathAdaptor_Rewind(b *testing.B) {
	// Create a path
	storage := NewPathStorageInteger[int32]()
	for i := 0; i < 100; i++ {
		storage.LineTo(int32(i*64), int32(i*128))
	}

	data, err := storage.Serialize()
	if err != nil {
		b.Fatalf("Failed to serialize storage: %v", err)
	}

	adaptor := NewSerializedIntegerPathAdaptorWithData[int32](data, 0.0, 0.0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		adaptor.Rewind(0)
	}
}
