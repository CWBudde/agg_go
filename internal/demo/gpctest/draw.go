package gpctest

import (
	"math"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/internal/demo/scanlineboolean2"
)

type Config struct {
	Scene     int
	Operation int
	CenterX   float64
	CenterY   float64
}

func Draw(ctx *agg.Context, cfg Config) {
	scene := cfg.Scene
	if scene < 0 {
		scene = 0
	}
	if scene > 4 {
		scene = 4
	}
	op := cfg.Operation
	if op < 0 {
		op = 0
	}
	if op > 5 {
		op = 5
	}
	if math.IsNaN(cfg.CenterX) || math.IsNaN(cfg.CenterY) {
		cfg.CenterX = float64(ctx.GetImage().Width()) * 0.5
		cfg.CenterY = float64(ctx.GetImage().Height()) * 0.5
	}

	scanlineboolean2.Draw(ctx, scanlineboolean2.Config{
		Mode:         scene,
		FillRule:     1,
		ScanlineType: 1,
		Operation:    op,
		CenterX:      cfg.CenterX,
		CenterY:      cfg.CenterY,
	})
}
