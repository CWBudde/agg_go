package scene

import (
	"fmt"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/engine"
)

// sourceImageSize is the edge length of the deterministic source image used by
// image scenes. A small power-of-two size with hard internal edges makes
// transform/scale sampling differences between backends easy to see.
const sourceImageSize = 64

// Assets carries per-engine resources a scene may need. It must be built for
// the same engine kind as the context it will be used with: a Port image drawn
// into a CPP context returns engine.ErrEngineMismatch.
type Assets struct {
	// Source is a deterministic RGBA image owned by the engine that created it.
	// It is nil only if BuildAssets failed for an engine that has no image
	// support, which does not happen for the engines in this corpus.
	Source engine.Image
}

// BuildAssets creates the per-engine source image used by image scenes. It must
// be called once per engine kind; the returned Assets is owned by that kind.
func BuildAssets(kind engine.Kind) (*Assets, error) {
	img, err := engine.NewImage(sourceImageSize, sourceImageSize, engine.Config{Kind: kind})
	if err != nil {
		return nil, fmt.Errorf("create source image for %s: %w", kind, err)
	}
	ctx, err := engine.NewContextForImage(img)
	if err != nil {
		return nil, fmt.Errorf("attach context to source image for %s: %w", kind, err)
	}
	paintSourcePattern(ctx)
	return &Assets{Source: img}, nil
}

// paintSourcePattern draws four solid quadrants plus a black diagonal using only
// the always-available solid-fill facade surface, so every backend produces the
// same source bytes.
func paintSourcePattern(ctx engine.Context) {
	h := float64(sourceImageSize) / 2

	ctx.Clear(agg.White)

	ctx.SetFillColor(agg.NewColorRGB(220, 40, 40)) // top-left: red
	ctx.FillRectangle(0, 0, h, h)
	ctx.SetFillColor(agg.NewColorRGB(40, 160, 40)) // top-right: green
	ctx.FillRectangle(h, 0, h, h)
	ctx.SetFillColor(agg.NewColorRGB(40, 60, 200)) // bottom-left: blue
	ctx.FillRectangle(0, h, h, h)
	ctx.SetFillColor(agg.NewColorRGB(230, 200, 40)) // bottom-right: yellow
	ctx.FillRectangle(h, h, h, h)

	// Diagonal hard edge across the whole tile.
	ctx.SetStrokeColor(agg.Black)
	ctx.SetLineWidth(2)
	ctx.DrawLine(0, 0, float64(sourceImageSize), float64(sourceImageSize))
}
