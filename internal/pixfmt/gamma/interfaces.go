package gamma

import "agg_go/internal/basics"

// Keep a tiny interface (exactly what you need)
type LUT8 interface {
	Dir(basics.Int8u) basics.Int8u
	Inv(basics.Int8u) basics.Int8u
}
