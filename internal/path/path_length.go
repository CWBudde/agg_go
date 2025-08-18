package path

import (
	"agg_go/internal/basics"
)

// PathLength calculates the total length of a path.
// This is a direct port of AGG's path_length template function.
//
// The function iterates through all vertices in the path, calculating
// the cumulative distance between consecutive points. When a close
// command is encountered, it adds the distance from the last point
// back to the starting point of the current subpath.
//
// Parameters:
//   - vs: VertexSource providing the path vertices
//   - pathID: ID of the path to measure (default 0)
//
// Returns:
//   - Total length of the path as float64
func PathLength(vs VertexSource, pathID uint) float64 {
	var length float64
	var startX, startY float64
	var x1, y1 float64
	var x2, y2 float64
	first := true

	vs.Rewind(pathID)
	for {
		var cmd uint32
		x2, y2, cmd = vs.NextVertex()

		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}

		if basics.IsVertex(basics.PathCommand(cmd)) {
			if first || basics.IsMoveTo(basics.PathCommand(cmd)) {
				startX = x2
				startY = y2
			} else {
				length += basics.CalcDistance(x1, y1, x2, y2)
			}
			x1 = x2
			y1 = y2
			first = false
		} else {
			if basics.IsClose(cmd) && !first {
				length += basics.CalcDistance(x1, y1, startX, startY)
			}
		}
	}

	return length
}
