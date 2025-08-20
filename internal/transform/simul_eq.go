// Package transform provides mathematical utilities for geometric transformations.
// This file implements simultaneous linear equation solving using Gaussian elimination.
//
// The implementation mirrors the C++ AGG library's agg_simul_eq.h, providing:
// - Gaussian elimination with partial pivoting for numerical stability
// - Support for multiple right-hand sides (solving AX = B where X and B are matrices)
// - Efficient specialized version for 4x4 systems used by bilinear transformations
//
// The algorithm uses partial pivoting to find the largest element in each column
// for numerical stability, then performs forward elimination followed by back
// substitution to solve the system.
package transform

import "math"

// swapArrays swaps two arrays element by element.
// This is a generic utility function used during row pivoting operations.
func swapArrays[T any](a1, a2 []T) {
	for i := range a1 {
		a1[i], a2[i] = a2[i], a1[i]
	}
}

// matrixPivot performs partial pivoting for Gaussian elimination.
// It finds the row with the largest absolute value in the specified column
// (starting from 'row') and swaps it to the pivot position for numerical stability.
//
// Parameters:
//   - m: the augmented matrix to pivot
//   - row: the current pivot row position
//
// Returns:
//   - -1 if the matrix is singular (pivot element is zero)
//   - 0 if no row swap was needed
//   - k>0 if row k was swapped with the pivot row
func matrixPivot(m [][]float64, row int) int {
	rows := len(m)
	k := row
	maxVal := -1.0

	for i := row; i < rows; i++ {
		if tmp := math.Abs(m[i][row]); tmp > maxVal && tmp != 0.0 {
			maxVal = tmp
			k = i
		}
	}

	if m[k][row] == 0.0 {
		return -1
	}

	if k != row {
		swapArrays(m[k], m[row])
		return k
	}
	return 0
}

// SimulEq solves simultaneous linear equations using Gaussian elimination with partial pivoting.
// It solves the matrix equation AX = B where A is the coefficient matrix, B is the right-hand
// side matrix, and X is the solution matrix.
//
// The function can solve multiple systems simultaneously by providing multiple columns in the
// right-hand side matrix (B). This is equivalent to solving AX₁ = B₁, AX₂ = B₂, etc.
//
// Parameters:
//   - left: coefficient matrix A (Size×Size)
//   - right: right-hand side matrix B (Size×RightCols)
//   - result: solution matrix X (Size×RightCols) - modified in place
//
// Returns:
//   - true if the system was solved successfully
//   - false if the matrix is singular (no unique solution exists)
//
// Example:
//
//	// Solve: 2x + 3y = 8, x + 4y = 7
//	left := [][]float64{{2, 3}, {1, 4}}
//	right := [][]float64{{8}, {7}}
//	result := make([][]float64, 2)
//	for i := range result { result[i] = make([]float64, 1) }
//	success := SimulEq(left, right, result)
//	// result[0][0] = x, result[1][0] = y
func SimulEq(left, right, result [][]float64) bool {
	size := len(left)
	rightCols := len(right[0])
	// Create augmented matrix
	tmp := make([][]float64, size)
	for i := range tmp {
		tmp[i] = make([]float64, size+rightCols)
		// Copy left matrix
		for j := 0; j < size; j++ {
			tmp[i][j] = left[i][j]
		}
		// Copy right matrix
		for j := 0; j < rightCols; j++ {
			tmp[i][size+j] = right[i][j]
		}
	}

	// Forward elimination
	for k := 0; k < size; k++ {
		if matrixPivot(tmp, k) < 0 {
			return false // Singularity
		}

		a1 := tmp[k][k]

		// Normalize pivot row
		for j := k; j < size+rightCols; j++ {
			tmp[k][j] /= a1
		}

		// Eliminate column
		for i := k + 1; i < size; i++ {
			a1 = tmp[i][k]
			for j := k; j < size+rightCols; j++ {
				tmp[i][j] -= a1 * tmp[k][j]
			}
		}
	}

	// Back substitution
	for k := 0; k < rightCols; k++ {
		for m := size - 1; m >= 0; m-- {
			result[m][k] = tmp[m][size+k]
			for j := m + 1; j < size; j++ {
				result[m][k] -= tmp[m][j] * result[j][k]
			}
		}
	}
	return true
}

// solve4x2 is a specialized version for 4x4 left matrix and 4x2 right matrix.
// This is optimized for the specific case needed by TransBilinear transformations
// when mapping arbitrary quadrilaterals to quadrilaterals.
//
// Parameters:
//   - left: 4×4 coefficient matrix (fixed arrays for performance)
//   - right: 4×2 right-hand side matrix
//   - result: 4×2 solution matrix (modified in place)
//
// Returns:
//   - true if the system was solved successfully
//   - false if the matrix is singular
func solve4x2(left *[4][4]float64, right [4][2]float64, result *[4][2]float64) bool {
	// Convert to slices for generic function
	leftSlice := make([][]float64, 4)
	rightSlice := make([][]float64, 4)
	resultSlice := make([][]float64, 4)

	for i := 0; i < 4; i++ {
		leftSlice[i] = left[i][:]
		rightSlice[i] = right[i][:]
		resultSlice[i] = result[i][:]
	}

	success := SimulEq(leftSlice, rightSlice, resultSlice)

	// Copy result back
	if success {
		for i := 0; i < 4; i++ {
			for j := 0; j < 2; j++ {
				result[i][j] = resultSlice[i][j]
			}
		}
	}

	return success
}
