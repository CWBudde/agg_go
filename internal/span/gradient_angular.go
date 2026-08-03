package span

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
)

// GradientAngular maps the complete signed atan2 turn onto [0,d2]. It is kept
// separate from GradientConic, whose two-sided absolute-angle behaviour is part
// of the existing API contract.
type GradientAngular struct{}

func (GradientAngular) Calculate(x, y, d2 int) int {
	if d2 <= 0 {
		return 0
	}
	t := (math.Atan2(float64(y), float64(x)) + math.Pi) / (2 * math.Pi)
	return basics.IRound(t * float64(d2))
}

// GradientReflected maps each complete distance interval onto a mirrored
// end-to-start-to-end ramp. Unlike GradientReflectAdaptor, its period is d2
// rather than 2*d2, matching reflected gradients used by image editors.
type GradientReflected struct{}

func (GradientReflected) Calculate(x, _, d2 int) int {
	if d2 <= 0 {
		return 0
	}
	position := x % d2
	if position < 0 {
		position += d2
	}
	return basics.Abs(position*2 - d2)
}
