// Package main ports AGG's mol_view.cpp demo.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/molview"
)

type demo struct {
	state molview.State
	drag  molview.DragState
}

func newDemo() *demo {
	return &demo{state: molview.DefaultState()}
}

func (d *demo) Render(ctx *agg.Context) {
	d.state.Advance()
	molview.Draw(ctx, d.state)
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left && !btn.Right {
		return false
	}
	molview.BeginDrag(&d.state, &d.drag, float64(x), float64(y), btn.Right)
	return true
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left && !btn.Right {
		return false
	}
	return molview.UpdateDrag(&d.state, &d.drag, float64(x), float64(y), btn.Right)
}

func (d *demo) OnMouseUp(_, _ int, _ demorunner.Buttons) bool {
	molview.EndDrag(&d.drag)
	return true
}

func (d *demo) OnKey(key rune) bool {
	switch key {
	case ' ', 'r', 'R':
		d.state.AutoRotate = !d.state.AutoRotate
	case 'n', 'N', '+':
		molview.NextMolecule(&d.state)
	case 'p', 'P', '-':
		molview.PrevMolecule(&d.state)
	default:
		return false
	}
	return true
}

func (d *demo) IsAnimated() bool {
	return d.state.AutoRotate
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Mol View",
		Width:  400,
		Height: 400,
	}, newDemo())
}
