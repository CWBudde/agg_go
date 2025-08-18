package bezierarc

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

const epsilon = 1e-10

// TestArcToBezier tests the basic arc to Bezier conversion function.
func TestArcToBezier(t *testing.T) {
	tests := []struct {
		name        string
		cx, cy      float64
		rx, ry      float64
		startAngle  float64
		sweepAngle  float64
		expectedLen int
	}{
		{
			name: "Quarter circle",
			cx:   0, cy: 0,
			rx: 10, ry: 10,
			startAngle:  0,
			sweepAngle:  math.Pi / 2,
			expectedLen: 8,
		},
		{
			name: "Half circle",
			cx:   100, cy: 50,
			rx: 20, ry: 20,
			startAngle:  0,
			sweepAngle:  math.Pi,
			expectedLen: 8,
		},
		{
			name: "Elliptical arc",
			cx:   0, cy: 0,
			rx: 30, ry: 20,
			startAngle:  math.Pi / 4,
			sweepAngle:  math.Pi / 3,
			expectedLen: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			curve := ArcToBezier(tt.cx, tt.cy, tt.rx, tt.ry, tt.startAngle, tt.sweepAngle)

			if len(curve) != tt.expectedLen {
				t.Errorf("Expected curve length %d, got %d", tt.expectedLen, len(curve))
			}

			// Verify that the start point is correct
			startX := tt.cx + tt.rx*math.Cos(tt.startAngle)
			startY := tt.cy + tt.ry*math.Sin(tt.startAngle)
			if math.Abs(curve[0]-startX) > epsilon || math.Abs(curve[1]-startY) > epsilon {
				t.Errorf("Start point mismatch: expected (%.10f, %.10f), got (%.10f, %.10f)",
					startX, startY, curve[0], curve[1])
			}

			// Verify that the end point is correct
			endX := tt.cx + tt.rx*math.Cos(tt.startAngle+tt.sweepAngle)
			endY := tt.cy + tt.ry*math.Sin(tt.startAngle+tt.sweepAngle)
			if math.Abs(curve[6]-endX) > epsilon || math.Abs(curve[7]-endY) > epsilon {
				t.Errorf("End point mismatch: expected (%.10f, %.10f), got (%.10f, %.10f)",
					endX, endY, curve[6], curve[7])
			}
		})
	}
}

// TestNewBezierArc tests the creation of new bezier arcs.
func TestNewBezierArc(t *testing.T) {
	arc := NewBezierArc()
	if arc == nil {
		t.Fatal("NewBezierArc() returned nil")
	}

	if arc.vertex != 26 {
		t.Errorf("Expected vertex index 26 (uninitialized), got %d", arc.vertex)
	}

	if arc.cmd != basics.PathCmdLineTo {
		t.Errorf("Expected initial command PathCmdLineTo, got %v", arc.cmd)
	}
}

// TestNewBezierArcWithParams tests creation with parameters.
func TestNewBezierArcWithParams(t *testing.T) {
	x, y := 100.0, 50.0
	rx, ry := 30.0, 20.0
	startAngle, sweepAngle := 0.0, math.Pi/2

	arc := NewBezierArcWithParams(x, y, rx, ry, startAngle, sweepAngle)
	if arc == nil {
		t.Fatal("NewBezierArcWithParams() returned nil")
	}

	if arc.numVertices == 0 {
		t.Error("Expected arc to be initialized with vertices")
	}
}

// TestBezierArcInit tests the initialization of bezier arcs.
func TestBezierArcInit(t *testing.T) {
	tests := []struct {
		name                string
		x, y                float64
		rx, ry              float64
		startAngle          float64
		sweepAngle          float64
		expectedMinVertices uint
		expectedCmd         basics.PathCommand
	}{
		{
			name: "Quarter circle",
			x:    0, y: 0,
			rx: 10, ry: 10,
			startAngle:          0,
			sweepAngle:          math.Pi / 2,
			expectedMinVertices: 2,
			expectedCmd:         basics.PathCmdCurve4,
		},
		{
			name: "Full circle",
			x:    50, y: 50,
			rx: 25, ry: 25,
			startAngle:          0,
			sweepAngle:          2 * math.Pi,
			expectedMinVertices: 14, // Should need multiple segments
			expectedCmd:         basics.PathCmdCurve4,
		},
		{
			name: "Degenerate arc (zero sweep)",
			x:    0, y: 0,
			rx: 10, ry: 10,
			startAngle:          0,
			sweepAngle:          1e-12, // Essentially zero
			expectedMinVertices: 4,
			expectedCmd:         basics.PathCmdLineTo,
		},
		{
			name: "Clockwise arc",
			x:    0, y: 0,
			rx: 15, ry: 15,
			startAngle:          0,
			sweepAngle:          -math.Pi / 2,
			expectedMinVertices: 2,
			expectedCmd:         basics.PathCmdCurve4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arc := NewBezierArc()
			arc.Init(tt.x, tt.y, tt.rx, tt.ry, tt.startAngle, tt.sweepAngle)

			if arc.numVertices < tt.expectedMinVertices {
				t.Errorf("Expected at least %d vertices, got %d", tt.expectedMinVertices, arc.numVertices)
			}

			if arc.cmd != tt.expectedCmd {
				t.Errorf("Expected command %v, got %v", tt.expectedCmd, arc.cmd)
			}

			// Verify vertices array has correct number of elements
			vertices := arc.Vertices()
			if len(vertices) != int(arc.numVertices) {
				t.Errorf("Vertices() length %d doesn't match NumVertices() %d", len(vertices), arc.numVertices)
			}
		})
	}
}

// TestBezierArcVertex tests vertex generation.
func TestBezierArcVertex(t *testing.T) {
	arc := NewBezierArcWithParams(0, 0, 10, 10, 0, math.Pi/2)
	arc.Rewind(0)

	var x, y float64
	vertexCount := 0
	firstCmd := true

	for {
		cmd := arc.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}

		vertexCount++

		if firstCmd {
			if cmd != basics.PathCmdMoveTo {
				t.Errorf("Expected first command to be MoveTo, got %v", cmd)
			}
			firstCmd = false
		} else if cmd != basics.PathCmdCurve4 {
			t.Errorf("Expected subsequent commands to be Curve4, got %v", cmd)
		}

		// Verify coordinates are reasonable
		if math.IsNaN(x) || math.IsNaN(y) {
			t.Errorf("Got NaN coordinates: (%f, %f)", x, y)
		}
	}

	if vertexCount == 0 {
		t.Error("No vertices were generated")
	}

	expectedVertexCount := int(arc.NumVertices() / 2)
	if vertexCount != expectedVertexCount {
		t.Errorf("Expected %d vertices, got %d", expectedVertexCount, vertexCount)
	}
}

// TestBezierArcRewind tests the rewind functionality.
func TestBezierArcRewind(t *testing.T) {
	arc := NewBezierArcWithParams(0, 0, 10, 10, 0, math.Pi/2)

	// Generate some vertices
	arc.Rewind(0)
	var x, y float64
	arc.Vertex(&x, &y) // Move to first vertex
	arc.Vertex(&x, &y) // Move to second vertex

	// Rewind and verify we start over
	arc.Rewind(0)
	cmd := arc.Vertex(&x, &y)
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("After rewind, expected first command to be MoveTo, got %v", cmd)
	}
}

// TestNewBezierArcSVG tests SVG arc creation.
func TestNewBezierArcSVG(t *testing.T) {
	svg := NewBezierArcSVG()
	if svg == nil {
		t.Fatal("NewBezierArcSVG() returned nil")
	}

	if svg.radiiOk {
		t.Error("Expected new SVG arc to have radiiOk = false initially")
	}
}

// TestBezierArcSVGInit tests SVG arc initialization.
func TestBezierArcSVGInit(t *testing.T) {
	tests := []struct {
		name          string
		x1, y1        float64
		rx, ry        float64
		angle         float64
		largeArcFlag  bool
		sweepFlag     bool
		x2, y2        float64
		expectRadiiOk bool
	}{
		{
			name: "Simple arc",
			x1:   0, y1: 0,
			rx: 10, ry: 10,
			angle:        0,
			largeArcFlag: false,
			sweepFlag:    true,
			x2:           10, y2: 10,
			expectRadiiOk: true,
		},
		{
			name: "Large arc flag",
			x1:   0, y1: 0,
			rx: 5, ry: 5,
			angle:        0,
			largeArcFlag: true,
			sweepFlag:    false,
			x2:           10, y2: 0,
			expectRadiiOk: true,
		},
		{
			name: "Radii too small (should scale up)",
			x1:   0, y1: 0,
			rx: 1, ry: 1, // Very small radii
			angle:        0,
			largeArcFlag: false,
			sweepFlag:    true,
			x2:           20, y2: 0, // Large distance
			expectRadiiOk: false, // Should be scaled up
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg := NewBezierArcSVG()
			svg.Init(tt.x1, tt.y1, tt.rx, tt.ry, tt.angle, tt.largeArcFlag, tt.sweepFlag, tt.x2, tt.y2)

			if svg.RadiiOk() != tt.expectRadiiOk {
				t.Errorf("Expected radiiOk %v, got %v", tt.expectRadiiOk, svg.RadiiOk())
			}

			if svg.NumVertices() == 0 {
				t.Error("Expected SVG arc to generate vertices")
			}

			// Verify endpoint precision - first and last vertices should match exactly
			vertices := svg.Vertices()
			if len(vertices) >= 4 {
				if math.Abs(vertices[0]-tt.x1) > epsilon || math.Abs(vertices[1]-tt.y1) > epsilon {
					t.Errorf("Start point mismatch: expected (%.10f, %.10f), got (%.10f, %.10f)",
						tt.x1, tt.y1, vertices[0], vertices[1])
				}

				lastIdx := len(vertices) - 2
				if math.Abs(vertices[lastIdx]-tt.x2) > epsilon || math.Abs(vertices[lastIdx+1]-tt.y2) > epsilon {
					t.Errorf("End point mismatch: expected (%.10f, %.10f), got (%.10f, %.10f)",
						tt.x2, tt.y2, vertices[lastIdx], vertices[lastIdx+1])
				}
			}
		})
	}
}

// TestBezierArcSVGVertex tests SVG arc vertex generation.
func TestBezierArcSVGVertex(t *testing.T) {
	svg := NewBezierArcSVGWithParams(0, 0, 10, 10, 0, false, true, 10, 10)
	svg.Rewind(0)

	var x, y float64
	vertexCount := 0

	for {
		cmd := svg.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++

		// Verify coordinates are reasonable
		if math.IsNaN(x) || math.IsNaN(y) {
			t.Errorf("Got NaN coordinates: (%f, %f)", x, y)
		}
	}

	if vertexCount == 0 {
		t.Error("No vertices were generated by SVG arc")
	}
}

// TestBezierArcAngleNormalization tests angle normalization.
func TestBezierArcAngleNormalization(t *testing.T) {
	tests := []struct {
		name       string
		startAngle float64
		sweepAngle float64
	}{
		{"Large positive start angle", 4 * math.Pi, math.Pi / 2},
		{"Large negative start angle", -4 * math.Pi, math.Pi / 2},
		{"Large positive sweep", 0, 4 * math.Pi},
		{"Large negative sweep", 0, -4 * math.Pi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arc := NewBezierArcWithParams(0, 0, 10, 10, tt.startAngle, tt.sweepAngle)

			if arc.NumVertices() == 0 {
				t.Error("Arc should generate vertices even with large angles")
			}

			// Should not panic and should generate reasonable output
			arc.Rewind(0)
			var x, y float64
			cmd := arc.Vertex(&x, &y)
			if cmd == basics.PathCmdStop {
				t.Error("Should generate at least one vertex")
			}
		})
	}
}

// TestBezierArcElliptical tests elliptical (non-circular) arcs.
func TestBezierArcElliptical(t *testing.T) {
	// Test different rx and ry values
	tests := []struct {
		name   string
		rx, ry float64
	}{
		{"Wide ellipse", 20, 10},
		{"Tall ellipse", 10, 20},
		{"Very wide", 50, 5},
		{"Very tall", 5, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arc := NewBezierArcWithParams(0, 0, tt.rx, tt.ry, 0, math.Pi)

			if arc.NumVertices() == 0 {
				t.Error("Elliptical arc should generate vertices")
			}

			// Verify that the arc respects the elliptical radii
			arc.Rewind(0)
			var x, y float64
			arc.Vertex(&x, &y) // First vertex (start point)

			expectedX := tt.rx * math.Cos(0)
			expectedY := tt.ry * math.Sin(0)

			if math.Abs(x-expectedX) > epsilon || math.Abs(y-expectedY) > epsilon {
				t.Errorf("Start point doesn't respect elliptical radii: expected (%.6f, %.6f), got (%.6f, %.6f)",
					expectedX, expectedY, x, y)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkArcToBezier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ArcToBezier(0, 0, 10, 10, 0, math.Pi/2)
	}
}

func BenchmarkBezierArcInit(b *testing.B) {
	arc := NewBezierArc()
	for i := 0; i < b.N; i++ {
		arc.Init(0, 0, 10, 10, 0, math.Pi)
	}
}

func BenchmarkBezierArcSVGInit(b *testing.B) {
	svg := NewBezierArcSVG()
	for i := 0; i < b.N; i++ {
		svg.Init(0, 0, 10, 10, 0, false, true, 10, 10)
	}
}
