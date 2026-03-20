package curves_test

import (
	"fmt"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/curves"
)

func TestBSplineValues(t *testing.T) {
	splineRX := []float64{0.000000, 0.200000, 0.400000, 0.910484, 0.957258, 1.000000}
	splineRY := []float64{1.000000, 0.800000, 0.600000, 0.066667, 0.169697, 0.600000}

	spR := curves.NewBSplineFromPoints(splineRX, splineRY)

	// First scatter point z = 17767/32768 ≈ 0.5422
	z1 := float64(17767) / 32768.0
	r1 := spR.Get(z1) * 0.8
	fmt.Printf("z=%.4f -> spR=%.4f, r_byte=%d\n", z1, spR.Get(z1), int(r1*255))

	// A few more test values
	for _, z := range []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0} {
		rv := spR.Get(z)
		fmt.Printf("  z=%.1f -> spR=%.4f byte=%d\n", z, rv, int(rv*0.8*255))
	}
}
