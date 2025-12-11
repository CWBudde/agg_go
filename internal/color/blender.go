package color

import "agg_go/internal/basics"

// Blendable is a constraint for colors that can blend with cover values.
// This uses the "Self type" pattern where Self is the implementing type itself.
// Colors like RGBA8[Linear] satisfy Blendable[RGBA8[Linear]] because they have
// an AddWithCover method that takes their own type.
type Blendable[Self any] interface {
	AddWithCover(src Self, cover basics.Int8u)
}
