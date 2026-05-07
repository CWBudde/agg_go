package main

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
)

func TestNewGammaControlUsesCPPBorderExtra(t *testing.T) {
	ctrl := newGammaControl()
	ctrl.Rewind(0)
	x, y, cmd := ctrl.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Fatalf("first background command = %v, want %v", cmd, basics.PathCmdMoveTo)
	}
	if x != 8.0 || y != 8.0 {
		t.Fatalf("first background vertex = (%v,%v), want (8,8)", x, y)
	}
}
