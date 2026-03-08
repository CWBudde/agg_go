package main

import (
	"math"

	"agg_go/internal/demo/gpctest"
)

var (
	gpcTestScene   = 3
	gpcTestOp      = 2
	gpcTestCenterX = math.NaN()
	gpcTestCenterY = math.NaN()
)

func setGPCTestScene(v int) {
	if v < 0 {
		v = 0
	}
	if v > 4 {
		v = 4
	}
	gpcTestScene = v
}

func setGPCTestOperation(v int) {
	if v < 0 {
		v = 0
	}
	if v > 5 {
		v = 5
	}
	gpcTestOp = v
}

func setGPCTestCenter(x, y float64) {
	gpcTestCenterX = x
	gpcTestCenterY = y
}

func handleGPCTestMouseDown(x, y float64) bool {
	setGPCTestCenter(x, y)
	return true
}

func handleGPCTestMouseMove(x, y float64) bool {
	setGPCTestCenter(x, y)
	return true
}

func handleGPCTestMouseUp() {}

func drawGPCTestDemo() {
	gpctest.Draw(ctx, gpctest.Config{
		Scene:     gpcTestScene,
		Operation: gpcTestOp,
		CenterX:   gpcTestCenterX,
		CenterY:   gpcTestCenterY,
	})
}
