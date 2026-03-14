// Package main ports AGG's idea.cpp demo.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/idea"
)

type demo struct {
	state idea.State
}

func newDemo() *demo {
	return &demo{state: idea.DefaultState()}
}

func (d *demo) Render(ctx *agg.Context) {
	idea.Draw(ctx, d.state)
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Idea",
		Width:  250,
		Height: 280,
	}, newDemo())
}
