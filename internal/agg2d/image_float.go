// Package agg2d float image rendering (L5). Float twin of the image copy/blend
// path: a float source image is wrapped in a float pixfmt and transferred into
// the target through the base renderer's CopyFrom/BlendFrom rectangle ops.
//
// Affine/perspective image transforms (TransformImage*) live in
// image_transform_float.go; this file covers the rectangle-aligned copy/blend.
package agg2d

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/pixfmt"
)

// CopyImageFloat copies a float source image into the target at (dstX, dstY)
// with no blending (straight pixel replacement within the clip box).
func (a *Agg2DFloat) CopyImageFloat(img *ImageFloat, dstX, dstY int) {
	if a.renBase == nil || img == nil || img.renBuf == nil {
		return
	}
	src := pixfmt.NewPixFmtRGBA128Plain(img.renBuf)
	a.renBase.CopyFrom(src, nil, dstX, dstY)
}

// BlendImageFloat blends a float source image into the target at (dstX, dstY)
// with the given uniform coverage (0..255).
func (a *Agg2DFloat) BlendImageFloat(img *ImageFloat, dstX, dstY int, cover basics.Int8u) {
	if a.renBase == nil || img == nil || img.renBuf == nil {
		return
	}
	src := pixfmt.NewPixFmtRGBA128Plain(img.renBuf)
	a.renBase.BlendFrom(src, nil, dstX, dstY, cover)
}
