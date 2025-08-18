package transform

import (
	"math"
	"testing"
)

func TestSwapArrays(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{4, 5, 6}
	expected_a := []int{4, 5, 6}
	expected_b := []int{1, 2, 3}

	swapArrays(a, b)

	for i := range a {
		if a[i] != expected_a[i] {
			t.Errorf("a[%d] = %d, want %d", i, a[i], expected_a[i])
		}
		if b[i] != expected_b[i] {
			t.Errorf("b[%d] = %d, want %d", i, b[i], expected_b[i])
		}
	}
}

func TestMatrixPivot(t *testing.T) {
	// Test with a 3x4 matrix that needs pivoting
	m := [][]float64{
		{0.0, 1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0, 7.0},
		{8.0, 9.0, 10.0, 11.0},
	}

	// Should swap row 0 with row 2 (largest absolute value in column 0)
	result := matrixPivot(m, 0)

	if result != 2 {
		t.Errorf("matrixPivot returned %d, want 2", result)
	}

	// Check that rows were swapped
	expected := [][]float64{
		{8.0, 9.0, 10.0, 11.0},
		{4.0, 5.0, 6.0, 7.0},
		{0.0, 1.0, 2.0, 3.0},
	}

	for i := range m {
		for j := range m[i] {
			if m[i][j] != expected[i][j] {
				t.Errorf("m[%d][%d] = %f, want %f", i, j, m[i][j], expected[i][j])
			}
		}
	}
}

func TestMatrixPivotSingular(t *testing.T) {
	// Test with a singular matrix (zero pivot)
	m := [][]float64{
		{0.0, 1.0},
		{0.0, 2.0},
	}

	result := matrixPivot(m, 0)
	if result != -1 {
		t.Errorf("matrixPivot returned %d, want -1 (singular)", result)
	}
}

func TestSimulEq_Identity(t *testing.T) {
	// Test identity system: I * x = b should give x = b
	left := [][]float64{
		{1.0, 0.0},
		{0.0, 1.0},
	}
	right := [][]float64{
		{5.0},
		{7.0},
	}
	result := make([][]float64, 2)
	for i := range result {
		result[i] = make([]float64, 1)
	}

	success := SimulEq(left, right, result)

	if !success {
		t.Fatal("SimulEq failed to solve identity system")
	}

	expected := [][]float64{{5.0}, {7.0}}
	for i := range result {
		for j := range result[i] {
			if math.Abs(result[i][j]-expected[i][j]) > 1e-10 {
				t.Errorf("result[%d][%d] = %f, want %f", i, j, result[i][j], expected[i][j])
			}
		}
	}
}

func TestSimulEq_2x2System(t *testing.T) {
	// Test system: 2x + 3y = 8, 4x + 6y = 16
	// This should be singular (second equation is 2x first)
	left := [][]float64{
		{2.0, 3.0},
		{4.0, 6.0},
	}
	right := [][]float64{
		{8.0},
		{16.0},
	}
	result := make([][]float64, 2)
	for i := range result {
		result[i] = make([]float64, 1)
	}

	success := SimulEq(left, right, result)

	// This system is actually consistent (dependent equations)
	// but our implementation should detect singularity
	if success {
		// If it succeeds, check if the solution is correct
		x, y := result[0][0], result[1][0]
		if math.Abs(2*x+3*y-8) > 1e-10 {
			t.Errorf("Solution doesn't satisfy first equation: 2*%f + 3*%f = %f, want 8", x, y, 2*x+3*y)
		}
	}
}

func TestSimulEq_3x3System(t *testing.T) {
	// Test system with known solution
	// 2x + 3y + z = 11
	// x + 2y + 3z = 14
	// 3x + y + 2z = 11
	// Solution: x=1, y=2, z=3
	left := [][]float64{
		{2.0, 3.0, 1.0},
		{1.0, 2.0, 3.0},
		{3.0, 1.0, 2.0},
	}
	right := [][]float64{
		{11.0},
		{14.0},
		{11.0},
	}
	result := make([][]float64, 3)
	for i := range result {
		result[i] = make([]float64, 1)
	}

	success := SimulEq(left, right, result)

	if !success {
		t.Fatal("SimulEq failed to solve 3x3 system")
	}

	expected := []float64{1.0, 2.0, 3.0}
	tolerance := 1e-10
	for i := range result {
		if math.Abs(result[i][0]-expected[i]) > tolerance {
			t.Errorf("result[%d] = %f, want %f", i, result[i][0], expected[i])
		}
	}

	// Verify solution by substitution
	for i := 0; i < 3; i++ {
		computed := 0.0
		for j := 0; j < 3; j++ {
			computed += left[i][j] * result[j][0]
		}
		if math.Abs(computed-right[i][0]) > tolerance {
			t.Errorf("Verification failed for equation %d: computed %f, expected %f", i, computed, right[i][0])
		}
	}
}

func TestSolve4x2(t *testing.T) {
	// Test the specialized 4x2 solver
	left := [4][4]float64{
		{1.0, 0.0, 1.0, 0.0},
		{1.0, 0.0, 0.0, 1.0},
		{1.0, 1.0, 1.0, 1.0},
		{1.0, 2.0, 0.0, 2.0},
	}
	right := [4][2]float64{
		{1.0, 2.0},
		{3.0, 4.0},
		{5.0, 6.0},
		{7.0, 8.0},
	}
	var result [4][2]float64

	success := solve4x2(&left, right, &result)

	if !success {
		t.Fatal("solve4x2 failed")
	}

	// Verify the solution by substitution
	tolerance := 1e-10
	for i := 0; i < 4; i++ {
		for j := 0; j < 2; j++ {
			computed := left[i][0]*result[0][j] + left[i][1]*result[1][j] +
				left[i][2]*result[2][j] + left[i][3]*result[3][j]
			expected := right[i][j]
			if math.Abs(computed-expected) > tolerance {
				t.Errorf("Solution verification failed at [%d,%d]: computed %f, expected %f",
					i, j, computed, expected)
			}
		}
	}
}

func BenchmarkSimulEq3x3(b *testing.B) {
	left := [][]float64{
		{2.0, 3.0, 1.0},
		{1.0, 2.0, 3.0},
		{3.0, 1.0, 2.0},
	}
	right := [][]float64{
		{10.0},
		{14.0},
		{12.0},
	}
	result := make([][]float64, 3)
	for i := range result {
		result[i] = make([]float64, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SimulEq(left, right, result)
	}
}

func BenchmarkSolve4x2(b *testing.B) {
	left := [4][4]float64{
		{1.0, 0.0, 1.0, 0.0},
		{1.0, 0.0, 0.0, 1.0},
		{1.0, 1.0, 1.0, 1.0},
		{1.0, 2.0, 0.0, 2.0},
	}
	right := [4][2]float64{
		{1.0, 2.0},
		{3.0, 4.0},
		{5.0, 6.0},
		{7.0, 8.0},
	}
	var result [4][2]float64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solve4x2(&left, right, &result)
	}
}
