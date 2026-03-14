package main

import "github.com/MeKo-Christian/agg_go/internal/demo/idea"

var ideaState = idea.DefaultState()

func setIdeaRotate(v bool) {
	ideaState.Rotate = v
}

func setIdeaEvenOdd(v bool) {
	ideaState.EvenOdd = v
}

func setIdeaDraft(v bool) {
	ideaState.Draft = v
}

func setIdeaRoundoff(v bool) {
	ideaState.Roundoff = v
}

func setIdeaAngleDelta(v float64) {
	ideaState.AngleDelta = v
	ideaState.Clamp()
}

func drawIdeaDemo() {
	ideaState.Advance()
	idea.Draw(ctx, ideaState)
}
