package transform

import "math"

// swapArrays swaps two arrays element by element
func swapArrays[T any](a1, a2 []T) {
	for i := range a1 {
		a1[i], a2[i] = a2[i], a1[i]
	}
}

// matrixPivot performs pivoting operation for Gaussian elimination
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

// SimulEq solves simultaneous equations using Gaussian elimination
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

// solve4x2 is a specialized version for 4x4 left matrix and 4x2 right matrix
// This is the specific case needed by TransBilinear
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
