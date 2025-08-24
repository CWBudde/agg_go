package main

import (
	"fmt"
	"math"
)

// Debug version of the dt algorithm
func dt(spanf, spang, spanr []float32, spann []int, length int) {
	square := func(x int) float32 { return float32(x * x) }

	fmt.Printf("Input spanf: %v\n", spanf[:length])

	k := 0
	spann[0] = 0
	spang[0] = -math.MaxFloat32
	spang[1] = math.MaxFloat32

	fmt.Printf("Initial state: k=%d, spann[0]=%d, spang[0]=%f, spang[1]=%f\n",
		k, spann[0], spang[0], spang[1])

	// First pass: build lower envelope of parabolas
	for q := 1; q <= length-1; q++ {
		var s float32
		if 2*q-2*spann[k] != 0 {
			s = ((spanf[q] + square(q)) - (spanf[spann[k]] + square(spann[k]))) / float32(2*q-2*spann[k])
		} else {
			s = math.MaxFloat32
		}

		fmt.Printf("q=%d: initial s=%f\n", q, s)

		for s <= spang[k] {
			k--
			if 2*q-2*spann[k] != 0 {
				s = ((spanf[q] + square(q)) - (spanf[spann[k]] + square(spann[k]))) / float32(2*q-2*spann[k])
			} else {
				s = math.MaxFloat32
			}
			fmt.Printf("  k decremented to %d, new s=%f\n", k, s)
		}

		k++
		spann[k] = q
		spang[k] = s
		spang[k+1] = math.MaxFloat32

		fmt.Printf("  Final: k=%d, spann[%d]=%d, spang[%d]=%f\n", k, k, spann[k], k, spang[k])
	}

	// Second pass: query the envelope
	k = 0
	for q := 0; q <= length-1; q++ {
		for spang[k+1] < float32(q) {
			k++
			fmt.Printf("  q=%d: k incremented to %d\n", q, k)
		}
		spanr[q] = square(q-spann[k]) + spanf[spann[k]]
		fmt.Printf("q=%d: k=%d, spann[k]=%d, result=%f\n", q, k, spann[k], spanr[q])
	}

	fmt.Printf("Final spanr: %v\n", spanr[:length])
}

func main() {
	length := 7
	spanf := []float32{0, math.MaxFloat32, math.MaxFloat32, 0, math.MaxFloat32, math.MaxFloat32, 0}
	spang := make([]float32, length+1)
	spanr := make([]float32, length)
	spann := make([]int, length)

	dt(spanf, spang, spanr, spann, length)

	expected := []float32{0, 1, 4, 0, 1, 4, 0}
	fmt.Printf("Expected: %v\n", expected)
	fmt.Printf("Got:      %v\n", spanr[:length])
}
