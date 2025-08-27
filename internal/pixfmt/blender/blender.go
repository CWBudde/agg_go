package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

// The pixel-format depends only on S and a value of B implementing this.
type RGBABlender[S color.Space, O order.RGBAOrder] interface {
	// Blend "plain" src (r,g,b,a) into dst[0:4] with coverage.
	// The concrete blender decides how to handle premul/plain storage + order.
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u)

	// Write/read a *plain* RGBA color to/from the framebuffer pixel.
	// (For premul storage these do premultiply/demultiply.)
	SetPlain(dst []basics.Int8u, r, g, b, a basics.Int8u)
	GetPlain(src []basics.Int8u) (r, g, b, a basics.Int8u)
}
