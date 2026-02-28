package agg

import "agg_go/internal/agg2d"

type FillRuleExamples = agg2d.FillRuleExamples

type PathWindingDirection = agg2d.PathWindingDirection

const (
	Clockwise        PathWindingDirection = agg2d.Clockwise
	CounterClockwise PathWindingDirection = agg2d.CounterClockwise
)

func CalculateWindingNumber(point [2]float64, polygon [][2]float64) int {
	return agg2d.CalculateWindingNumber(point, polygon)
}
