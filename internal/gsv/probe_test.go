package gsv_test

import (
	"fmt"
	"testing"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
)

func TestGSVDirection(t *testing.T) {
	for _, flip := range []bool{false, true} {
		tx := gsv.NewGSVText()
		tx.SetSize(20, 0)
		tx.SetFlip(flip)
		tx.SetStartPoint(0, 100)
		tx.SetText("A")
		tx.Rewind(0)
		fmt.Printf("flip=%v:\n", flip)
		minY, maxY := 1e9, -1e9
		for {
			_, y, cmd := tx.Vertex()
			if cmd == basics.PathCmdStop { break }
			if y < minY { minY = y }
			if y > maxY { maxY = y }
		}
		fmt.Printf("  y range: %.1f..%.1f (start=100)\n", minY, maxY)
	}
}
