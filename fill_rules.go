package agg

import "github.com/MeKo-Christian/agg_go/internal/agg2d"

// FillRuleExamples re-exports the helper examples for the supported fill rules.
type FillRuleExamples = agg2d.FillRuleExamples

// PathWindingDirection identifies polygon winding direction for fill-rule
// helpers.
type PathWindingDirection = agg2d.PathWindingDirection

const (
	// Clockwise indicates clockwise vertex winding.
	Clockwise PathWindingDirection = agg2d.Clockwise
	// CounterClockwise indicates counter-clockwise vertex winding.
	CounterClockwise PathWindingDirection = agg2d.CounterClockwise
)

// CalculateWindingNumber computes the winding number of polygon around point.
func CalculateWindingNumber(point [2]float64, polygon [][2]float64) int {
	return agg2d.CalculateWindingNumber(point, polygon)
}
