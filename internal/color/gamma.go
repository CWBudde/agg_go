package color

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// lut8Like is the minimal 8-bit gamma-LUT contract used by the color helpers.
type lut8Like interface {
	Dir(basics.Int8u) basics.Int8u
	Inv(basics.Int8u) basics.Int8u
}

// lut16Like is the minimal 16-bit gamma-LUT contract used by the 16-bit color
// helpers.
type lut16Like interface {
	Dir(basics.Int8u) basics.Int16u
	Inv(basics.Int16u) basics.Int8u
}

// lut32Like is the floating-point gamma-function contract used by RGBA32/Gray32
// helpers.
type lut32Like interface {
	DirFloat(v float32) float32
	InvFloat(v float32) float32
}
