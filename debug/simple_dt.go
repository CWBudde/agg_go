package main

import (
	"fmt"
	"math"
)

// Simple brute force implementation to verify expected behavior
func bruteForceDistanceTransform(input []float32) []float32 {
	result := make([]float32, len(input))

	for i := 0; i < len(input); i++ {
		if input[i] == 0 {
			result[i] = 0
		} else {
			minDist := float32(math.MaxFloat32)
			// Find distance to nearest zero
			for j := 0; j < len(input); j++ {
				if input[j] == 0 {
					dist := float32((i - j) * (i - j))
					if dist < minDist {
						minDist = dist
					}
				}
			}
			result[i] = minDist
		}
	}
	return result
}

func main() {
	input := []float32{0, math.MaxFloat32, math.MaxFloat32, 0, math.MaxFloat32, math.MaxFloat32, 0}

	fmt.Printf("Input: %v\n", input)

	// What would brute force give us?
	bruteResult := bruteForceDistanceTransform(input)
	fmt.Printf("Brute force result: %v\n", bruteResult)

	expected := []float32{0, 1, 4, 0, 1, 4, 0}
	fmt.Printf("Test expects:       %v\n", expected)

	// Let's manually check each position:
	fmt.Println("\nManual verification:")
	for i, val := range input {
		if val == 0 {
			fmt.Printf("Position %d: is zero -> distance = 0\n", i)
		} else {
			// Find nearest zeros
			leftZero, rightZero := -1, -1
			for j := i - 1; j >= 0; j-- {
				if input[j] == 0 {
					leftZero = j
					break
				}
			}
			for j := i + 1; j < len(input); j++ {
				if input[j] == 0 {
					rightZero = j
					break
				}
			}

			leftDist := float32(math.MaxFloat32)
			rightDist := float32(math.MaxFloat32)
			if leftZero >= 0 {
				leftDist = float32((i - leftZero) * (i - leftZero))
			}
			if rightZero >= 0 {
				rightDist = float32((rightZero - i) * (rightZero - i))
			}

			minDist := leftDist
			nearest := leftZero
			if rightDist < leftDist {
				minDist = rightDist
				nearest = rightZero
			}

			fmt.Printf("Position %d: nearest zero at %d, distance = %d² = %g\n",
				i, nearest, int(math.Sqrt(float64(minDist))), minDist)
		}
	}
}
