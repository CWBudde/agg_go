package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

// Test with identity transform
func TestConvTransform_Identity(t *testing.T) {
	vertices := []Vertex{
		{10, 20, basics.PathCmdMoveTo},
		{30, 40, basics.PathCmdLineTo},
		{50, 60, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine() // Identity transform
	conv := NewConvTransform(source, transformer)

	// Test that identity transform doesn't change coordinates
	conv.Rewind(0)
	for i, expected := range vertices {
		x, y, cmd := conv.Vertex()
		if x != expected.X || y != expected.Y || cmd != expected.Cmd {
			t.Errorf("Vertex %d: expected (%v, %v, %v), got (%v, %v, %v)",
				i, expected.X, expected.Y, expected.Cmd, x, y, cmd)
		}
	}
}

// Test with translation
func TestConvTransform_Translation(t *testing.T) {
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 20, basics.PathCmdLineTo},
		{30, 40, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().Translate(100, 200)
	conv := NewConvTransform(source, transformer)

	expected := []Vertex{
		{100, 200, basics.PathCmdMoveTo},
		{110, 220, basics.PathCmdLineTo},
		{130, 240, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop}, // Stop command coordinates unchanged
	}

	conv.Rewind(0)
	for i, exp := range expected {
		x, y, cmd := conv.Vertex()
		if basics.IsVertex(cmd) {
			if math.Abs(x-exp.X) > 1e-10 || math.Abs(y-exp.Y) > 1e-10 || cmd != exp.Cmd {
				t.Errorf("Vertex %d: expected (%v, %v, %v), got (%v, %v, %v)",
					i, exp.X, exp.Y, exp.Cmd, x, y, cmd)
			}
		} else {
			if cmd != exp.Cmd {
				t.Errorf("Command %d: expected %v, got %v", i, exp.Cmd, cmd)
			}
		}
	}
}

// Test with scaling
func TestConvTransform_Scaling(t *testing.T) {
	vertices := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{2, 3, basics.PathCmdLineTo},
		{4, 5, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().ScaleXY(2, 3)
	conv := NewConvTransform(source, transformer)

	expected := []Vertex{
		{2, 3, basics.PathCmdMoveTo},
		{4, 9, basics.PathCmdLineTo},
		{8, 15, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	conv.Rewind(0)
	for i, exp := range expected {
		x, y, cmd := conv.Vertex()
		if basics.IsVertex(cmd) {
			if math.Abs(x-exp.X) > 1e-10 || math.Abs(y-exp.Y) > 1e-10 || cmd != exp.Cmd {
				t.Errorf("Vertex %d: expected (%v, %v, %v), got (%v, %v, %v)",
					i, exp.X, exp.Y, exp.Cmd, x, y, cmd)
			}
		} else {
			if cmd != exp.Cmd {
				t.Errorf("Command %d: expected %v, got %v", i, exp.Cmd, cmd)
			}
		}
	}
}

// Test with rotation
func TestConvTransform_Rotation(t *testing.T) {
	vertices := []Vertex{
		{1, 0, basics.PathCmdMoveTo},
		{0, 1, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().Rotate(math.Pi / 2) // 90 degrees
	conv := NewConvTransform(source, transformer)

	// After 90-degree rotation: (1,0) -> (0,1), (0,1) -> (-1,0)
	expected := []Vertex{
		{0, 1, basics.PathCmdMoveTo},
		{-1, 0, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	conv.Rewind(0)
	for i, exp := range expected {
		x, y, cmd := conv.Vertex()
		if basics.IsVertex(cmd) {
			if math.Abs(x-exp.X) > 1e-10 || math.Abs(y-exp.Y) > 1e-10 || cmd != exp.Cmd {
				t.Errorf("Vertex %d: expected (%v, %v, %v), got (%v, %v, %v)",
					i, exp.X, exp.Y, exp.Cmd, x, y, cmd)
			}
		} else {
			if cmd != exp.Cmd {
				t.Errorf("Command %d: expected %v, got %v", i, exp.Cmd, cmd)
			}
		}
	}
}

// Test with combined transformations
func TestConvTransform_Combined(t *testing.T) {
	vertices := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{2, 2, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	// Scale by 2, then translate by (10, 20)
	transformer := transform.NewTransAffine().ScaleXY(2, 2).Translate(10, 20)
	conv := NewConvTransform(source, transformer)

	expected := []Vertex{
		{12, 22, basics.PathCmdMoveTo}, // (1*2+10, 1*2+20)
		{14, 24, basics.PathCmdLineTo}, // (2*2+10, 2*2+20)
		{0, 0, basics.PathCmdStop},
	}

	conv.Rewind(0)
	for i, exp := range expected {
		x, y, cmd := conv.Vertex()
		if basics.IsVertex(cmd) {
			if math.Abs(x-exp.X) > 1e-10 || math.Abs(y-exp.Y) > 1e-10 || cmd != exp.Cmd {
				t.Errorf("Vertex %d: expected (%v, %v, %v), got (%v, %v, %v)",
					i, exp.X, exp.Y, exp.Cmd, x, y, cmd)
			}
		} else {
			if cmd != exp.Cmd {
				t.Errorf("Command %d: expected %v, got %v", i, exp.Cmd, cmd)
			}
		}
	}
}

// Test with curves (should preserve curve commands)
func TestConvTransform_Curves(t *testing.T) {
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 10, basics.PathCmdCurve3},
		{20, 0, basics.PathCmdCurve3},
		{30, 30, basics.PathCmdCurve4},
		{40, 0, basics.PathCmdCurve4},
		{50, 50, basics.PathCmdCurve4},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().Translate(5, 10)
	conv := NewConvTransform(source, transformer)

	expected := []Vertex{
		{5, 10, basics.PathCmdMoveTo},
		{15, 20, basics.PathCmdCurve3},
		{25, 10, basics.PathCmdCurve3},
		{35, 40, basics.PathCmdCurve4},
		{45, 10, basics.PathCmdCurve4},
		{55, 60, basics.PathCmdCurve4},
		{0, 0, basics.PathCmdStop},
	}

	conv.Rewind(0)
	for i, exp := range expected {
		x, y, cmd := conv.Vertex()
		if basics.IsVertex(cmd) {
			if math.Abs(x-exp.X) > 1e-10 || math.Abs(y-exp.Y) > 1e-10 || cmd != exp.Cmd {
				t.Errorf("Vertex %d: expected (%v, %v, %v), got (%v, %v, %v)",
					i, exp.X, exp.Y, exp.Cmd, x, y, cmd)
			}
		} else {
			if cmd != exp.Cmd {
				t.Errorf("Command %d: expected %v, got %v", i, exp.Cmd, cmd)
			}
		}
	}
}

// Test attach method
func TestConvTransform_Attach(t *testing.T) {
	vertices1 := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdStop},
	}
	vertices2 := []Vertex{
		{2, 2, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdStop},
	}

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)
	transformer := transform.NewTransAffine().Translate(10, 10)
	conv := NewConvTransform(source1, transformer)

	// Test with first source
	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	if x != 11 || y != 11 || cmd != basics.PathCmdMoveTo {
		t.Errorf("First source: expected (11, 11, MoveTo), got (%v, %v, %v)", x, y, cmd)
	}

	// Attach second source
	conv.Attach(source2)
	conv.Rewind(0)
	x, y, cmd = conv.Vertex()
	if x != 12 || y != 12 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Second source: expected (12, 12, MoveTo), got (%v, %v, %v)", x, y, cmd)
	}
}

// Test SetTransformer method
func TestConvTransform_SetTransformer(t *testing.T) {
	vertices := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer1 := transform.NewTransAffine().Translate(10, 10)
	transformer2 := transform.NewTransAffine().ScaleXY(2, 2)
	conv := NewConvTransform(source, transformer1)

	// Test with first transformer
	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	if x != 11 || y != 11 || cmd != basics.PathCmdMoveTo {
		t.Errorf("First transformer: expected (11, 11, MoveTo), got (%v, %v, %v)", x, y, cmd)
	}

	// Set second transformer
	conv.SetTransformer(transformer2)
	conv.Rewind(0)
	x, y, cmd = conv.Vertex()
	if x != 2 || y != 2 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Second transformer: expected (2, 2, MoveTo), got (%v, %v, %v)", x, y, cmd)
	}
}

// Test rewind functionality
func TestConvTransform_Rewind(t *testing.T) {
	vertices := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{2, 2, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine()
	conv := NewConvTransform(source, transformer)

	// Read all vertices
	conv.Rewind(0)
	conv.Vertex() // MoveTo
	conv.Vertex() // LineTo
	conv.Vertex() // Stop

	// Rewind and read again
	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	if x != 1 || y != 1 || cmd != basics.PathCmdMoveTo {
		t.Errorf("After rewind: expected (1, 1, MoveTo), got (%v, %v, %v)", x, y, cmd)
	}
}

// Test Transformer method (AGG-compatible name)
func TestConvTransform_Transformer(t *testing.T) {
	vertices := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer1 := transform.NewTransAffine().Translate(10, 10)
	transformer2 := transform.NewTransAffine().ScaleXY(3, 3)
	conv := NewConvTransform(source, transformer1)

	// Test with first transformer
	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	if x != 11 || y != 11 || cmd != basics.PathCmdMoveTo {
		t.Errorf("First transformer: expected (11, 11, MoveTo), got (%v, %v, %v)", x, y, cmd)
	}

	// Use AGG-compatible Transformer method
	conv.Transformer(transformer2)
	conv.Rewind(0)
	x, y, cmd = conv.Vertex()
	if x != 3 || y != 3 || cmd != basics.PathCmdMoveTo {
		t.Errorf("After Transformer method: expected (3, 3, MoveTo), got (%v, %v, %v)", x, y, cmd)
	}
}

// Test with empty path
func TestConvTransform_EmptyPath(t *testing.T) {
	source := NewMockVertexSource([]Vertex{})
	transformer := transform.NewTransAffine().Translate(10, 10)
	conv := NewConvTransform(source, transformer)

	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Empty path: expected Stop command, got (%v, %v, %v)", x, y, cmd)
	}
}

// Test with single point
func TestConvTransform_SinglePoint(t *testing.T) {
	vertices := []Vertex{
		{5, 7, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().ScaleXY(2, 3).Translate(1, 1)
	conv := NewConvTransform(source, transformer)

	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	// Transform: scale (5,7) -> (10,21), then translate -> (11,22)
	expectedX, expectedY := 11.0, 22.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Single point: expected (%v, %v, MoveTo), got (%v, %v, %v)", expectedX, expectedY, x, y, cmd)
	}

	// Next should be stop
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Single point second call: expected Stop, got (%v, %v, %v)", x, y, cmd)
	}
}

// Test with polygon and path flags
func TestConvTransform_PolygonWithFlags(t *testing.T) {
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().Translate(5, 5)
	conv := NewConvTransform(source, transformer)

	expected := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose}, // Non-vertex commands unchanged
		{0, 0, basics.PathCmdStop},
	}

	conv.Rewind(0)
	for i, exp := range expected {
		x, y, cmd := conv.Vertex()
		if basics.IsVertex(cmd) {
			if math.Abs(x-exp.X) > 1e-10 || math.Abs(y-exp.Y) > 1e-10 || cmd != exp.Cmd {
				t.Errorf("Polygon vertex %d: expected (%v, %v, %v), got (%v, %v, %v)",
					i, exp.X, exp.Y, exp.Cmd, x, y, cmd)
			}
		} else {
			if cmd != exp.Cmd {
				t.Errorf("Polygon command %d: expected %v, got %v", i, exp.Cmd, cmd)
			}
		}
	}
}

// Test numerical precision with very small numbers
func TestConvTransform_NumericalPrecision(t *testing.T) {
	vertices := []Vertex{
		{1e-10, 1e-10, basics.PathCmdMoveTo},
		{1e-9, 1e-9, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().ScaleXY(1e6, 1e6) // Scale up very small numbers
	conv := NewConvTransform(source, transformer)

	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	expectedX, expectedY := 1e-4, 1e-4
	if math.Abs(x-expectedX) > 1e-15 || math.Abs(y-expectedY) > 1e-15 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Small numbers: expected (%v, %v, MoveTo), got (%v, %v, %v)", expectedX, expectedY, x, y, cmd)
	}

	x, y, cmd = conv.Vertex()
	expectedX, expectedY = 1e-3, 1e-3
	if math.Abs(x-expectedX) > 1e-15 || math.Abs(y-expectedY) > 1e-15 || cmd != basics.PathCmdLineTo {
		t.Errorf("Small numbers line: expected (%v, %v, LineTo), got (%v, %v, %v)", expectedX, expectedY, x, y, cmd)
	}
}

// Test with very large numbers
func TestConvTransform_LargeNumbers(t *testing.T) {
	vertices := []Vertex{
		{1e12, 1e12, basics.PathCmdMoveTo},
		{2e12, 2e12, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().ScaleXY(0.5, 0.5)
	conv := NewConvTransform(source, transformer)

	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	expectedX, expectedY := 5e11, 5e11
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Large numbers: expected (%v, %v, MoveTo), got (%v, %v, %v)", expectedX, expectedY, x, y, cmd)
	}
}

// Test with all path command types
func TestConvTransform_AllCommands(t *testing.T) {
	vertices := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{2, 2, basics.PathCmdLineTo},
		{3, 3, basics.PathCmdCurve3},
		{4, 4, basics.PathCmdCurve3},
		{5, 5, basics.PathCmdCurve4},
		{6, 6, basics.PathCmdCurve4},
		{7, 7, basics.PathCmdCurve4},
		{0, 0, basics.PathCmdEndPoly},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().Translate(1, 1)
	conv := NewConvTransform(source, transformer)

	expected := []Vertex{
		{2, 2, basics.PathCmdMoveTo},
		{3, 3, basics.PathCmdLineTo},
		{4, 4, basics.PathCmdCurve3},
		{5, 5, basics.PathCmdCurve3},
		{6, 6, basics.PathCmdCurve4},
		{7, 7, basics.PathCmdCurve4},
		{8, 8, basics.PathCmdCurve4},
		{0, 0, basics.PathCmdEndPoly}, // Non-vertex command unchanged
		{0, 0, basics.PathCmdStop},
	}

	conv.Rewind(0)
	for i, exp := range expected {
		x, y, cmd := conv.Vertex()
		if basics.IsVertex(cmd) {
			if math.Abs(x-exp.X) > 1e-10 || math.Abs(y-exp.Y) > 1e-10 || cmd != exp.Cmd {
				t.Errorf("Command %d: expected (%v, %v, %v), got (%v, %v, %v)",
					i, exp.X, exp.Y, exp.Cmd, x, y, cmd)
			}
		} else {
			if cmd != exp.Cmd {
				t.Errorf("Non-vertex command %d: expected %v, got %v", i, exp.Cmd, cmd)
			}
		}
	}
}

// Test with inverse transformation
func TestConvTransform_InverseTransform(t *testing.T) {
	vertices := []Vertex{
		{10, 20, basics.PathCmdMoveTo},
		{30, 40, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	// Create a transformation and its inverse
	transformer := transform.NewTransAffine().ScaleXY(2, 2).Translate(10, 10)
	// Manually create inverse: first untranslate, then unscale
	inverseTransformer := transform.NewTransAffine().Translate(-10, -10).ScaleXY(0.5, 0.5)

	// Apply transform then inverse - should get back original
	conv1 := NewConvTransform(source, transformer)
	conv2 := NewConvTransform(conv1, inverseTransformer)

	conv2.Rewind(0)

	tolerance := 1e-10
	for i, expected := range vertices {
		x, y, cmd := conv2.Vertex()
		if basics.IsVertex(cmd) {
			if math.Abs(x-expected.X) > tolerance || math.Abs(y-expected.Y) > tolerance || cmd != expected.Cmd {
				t.Errorf("Inverse transform vertex %d: expected (%v, %v, %v), got (%v, %v, %v)",
					i, expected.X, expected.Y, expected.Cmd, x, y, cmd)
			}
		} else {
			if cmd != expected.Cmd {
				t.Errorf("Inverse transform command %d: expected %v, got %v", i, expected.Cmd, cmd)
			}
		}
	}
}
