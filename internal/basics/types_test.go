package basics

import (
	"testing"
)

func TestRowInfo(t *testing.T) {
	t.Run("NewRowInfo", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		rowInfo := NewRowInfo(10, 20, data)

		if rowInfo.X1 != 10 {
			t.Errorf("Expected X1=10, got %d", rowInfo.X1)
		}
		if rowInfo.X2 != 20 {
			t.Errorf("Expected X2=20, got %d", rowInfo.X2)
		}
		if len(rowInfo.Ptr) != 5 {
			t.Errorf("Expected Ptr length=5, got %d", len(rowInfo.Ptr))
		}
	})

	t.Run("NewConstRowInfo", func(t *testing.T) {
		data := []float64{1.1, 2.2, 3.3}
		constRowInfo := NewConstRowInfo(5, 15, data)

		if constRowInfo.X1 != 5 {
			t.Errorf("Expected X1=5, got %d", constRowInfo.X1)
		}
		if constRowInfo.X2 != 15 {
			t.Errorf("Expected X2=15, got %d", constRowInfo.X2)
		}
		if len(constRowInfo.Ptr) != 3 {
			t.Errorf("Expected Ptr length=3, got %d", len(constRowInfo.Ptr))
		}
	})
}

func TestTypeAliases(t *testing.T) {
	t.Run("Point aliases", func(t *testing.T) {
		// Test PointI
		pi := PointI{X: 10, Y: 20}
		if pi.X != 10 || pi.Y != 20 {
			t.Errorf("PointI failed: expected (10,20), got (%d,%d)", pi.X, pi.Y)
		}

		// Test PointF
		pf := PointF{X: 1.5, Y: 2.5}
		if pf.X != 1.5 || pf.Y != 2.5 {
			t.Errorf("PointF failed: expected (1.5,2.5), got (%f,%f)", pf.X, pf.Y)
		}

		// Test PointD
		pd := PointD{X: 3.14, Y: 2.71}
		if pd.X != 3.14 || pd.Y != 2.71 {
			t.Errorf("PointD failed: expected (3.14,2.71), got (%f,%f)", pd.X, pd.Y)
		}
	})

	t.Run("Rect aliases", func(t *testing.T) {
		// Test RectI
		ri := RectI{X1: 0, Y1: 0, X2: 100, Y2: 200}
		if ri.X1 != 0 || ri.Y1 != 0 || ri.X2 != 100 || ri.Y2 != 200 {
			t.Errorf("RectI failed: expected (0,0,100,200), got (%d,%d,%d,%d)", ri.X1, ri.Y1, ri.X2, ri.Y2)
		}

		// Test RectF
		rf := RectF{X1: 0.1, Y1: 0.2, X2: 10.5, Y2: 20.7}
		if rf.X1 != 0.1 || rf.Y1 != 0.2 || rf.X2 != 10.5 || rf.Y2 != 20.7 {
			t.Errorf("RectF failed: expected (0.1,0.2,10.5,20.7), got (%f,%f,%f,%f)", rf.X1, rf.Y1, rf.X2, rf.Y2)
		}

		// Test RectD
		rd := RectD{X1: 1.1, Y1: 2.2, X2: 3.3, Y2: 4.4}
		if rd.X1 != 1.1 || rd.Y1 != 2.2 || rd.X2 != 3.3 || rd.Y2 != 4.4 {
			t.Errorf("RectD failed: expected (1.1,2.2,3.3,4.4), got (%f,%f,%f,%f)", rd.X1, rd.Y1, rd.X2, rd.Y2)
		}
	})

	t.Run("Vertex aliases", func(t *testing.T) {
		// Test VertexI
		vi := VertexI{X: 50, Y: 60, Cmd: 1}
		if vi.X != 50 || vi.Y != 60 || vi.Cmd != 1 {
			t.Errorf("VertexI failed: expected (50,60,1), got (%d,%d,%d)", vi.X, vi.Y, vi.Cmd)
		}

		// Test VertexF
		vf := VertexF{X: 5.5, Y: 6.6, Cmd: 2}
		if vf.X != 5.5 || vf.Y != 6.6 || vf.Cmd != 2 {
			t.Errorf("VertexF failed: expected (5.5,6.6,2), got (%f,%f,%d)", vf.X, vf.Y, vf.Cmd)
		}

		// Test VertexD
		vd := VertexD{X: 7.7, Y: 8.8, Cmd: 3}
		if vd.X != 7.7 || vd.Y != 8.8 || vd.Cmd != 3 {
			t.Errorf("VertexD failed: expected (7.7,8.8,3), got (%f,%f,%d)", vd.X, vd.Y, vd.Cmd)
		}
	})
}

func TestIntersectRectangles(t *testing.T) {
	tests := []struct {
		name       string
		r1, r2     RectI
		expected   RectI
		intersects bool
	}{
		{
			name:       "Overlapping rectangles",
			r1:         RectI{X1: 0, Y1: 0, X2: 10, Y2: 10},
			r2:         RectI{X1: 5, Y1: 5, X2: 15, Y2: 15},
			expected:   RectI{X1: 5, Y1: 5, X2: 10, Y2: 10},
			intersects: true,
		},
		{
			name:       "Non-overlapping rectangles",
			r1:         RectI{X1: 0, Y1: 0, X2: 5, Y2: 5},
			r2:         RectI{X1: 10, Y1: 10, X2: 15, Y2: 15},
			expected:   RectI{X1: 10, Y1: 10, X2: 5, Y2: 5},
			intersects: false,
		},
		{
			name:       "Touching rectangles",
			r1:         RectI{X1: 0, Y1: 0, X2: 5, Y2: 5},
			r2:         RectI{X1: 5, Y1: 5, X2: 10, Y2: 10},
			expected:   RectI{X1: 5, Y1: 5, X2: 5, Y2: 5},
			intersects: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, intersects := IntersectRectangles(tt.r1, tt.r2)
			if intersects != tt.intersects {
				t.Errorf("IntersectRectangles() intersects = %v, want %v", intersects, tt.intersects)
			}
			if result != tt.expected {
				t.Errorf("IntersectRectangles() result = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUniteRectangles(t *testing.T) {
	tests := []struct {
		name     string
		r1, r2   RectI
		expected RectI
	}{
		{
			name:     "Two separate rectangles",
			r1:       RectI{X1: 0, Y1: 0, X2: 5, Y2: 5},
			r2:       RectI{X1: 10, Y1: 10, X2: 15, Y2: 15},
			expected: RectI{X1: 0, Y1: 0, X2: 15, Y2: 15},
		},
		{
			name:     "Overlapping rectangles",
			r1:       RectI{X1: 0, Y1: 0, X2: 10, Y2: 10},
			r2:       RectI{X1: 5, Y1: 5, X2: 15, Y2: 15},
			expected: RectI{X1: 0, Y1: 0, X2: 15, Y2: 15},
		},
		{
			name:     "One rectangle inside another",
			r1:       RectI{X1: 0, Y1: 0, X2: 20, Y2: 20},
			r2:       RectI{X1: 5, Y1: 5, X2: 15, Y2: 15},
			expected: RectI{X1: 0, Y1: 0, X2: 20, Y2: 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniteRectangles(tt.r1, tt.r2)
			if result != tt.expected {
				t.Errorf("UniteRectangles() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRectangleFunctionsWithFloats(t *testing.T) {
	t.Run("Float rectangles intersection", func(t *testing.T) {
		r1 := RectD{X1: 0.5, Y1: 0.5, X2: 10.5, Y2: 10.5}
		r2 := RectD{X1: 5.2, Y1: 5.2, X2: 15.7, Y2: 15.7}
		result, intersects := IntersectRectangles(r1, r2)

		if !intersects {
			t.Error("Expected rectangles to intersect")
		}

		expected := RectD{X1: 5.2, Y1: 5.2, X2: 10.5, Y2: 10.5}
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Float rectangles union", func(t *testing.T) {
		r1 := RectD{X1: 1.1, Y1: 2.2, X2: 5.5, Y2: 6.6}
		r2 := RectD{X1: 3.3, Y1: 4.4, X2: 7.7, Y2: 8.8}
		result := UniteRectangles(r1, r2)

		expected := RectD{X1: 1.1, Y1: 2.2, X2: 7.7, Y2: 8.8}
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
