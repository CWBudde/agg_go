package order

type RGBAOrder interface {
	IdxR() int
	IdxG() int
	IdxB() int
	IdxA() int
}

type RGBA struct{}

func (RGBA) IdxR() int { return 0 }
func (RGBA) IdxG() int { return 1 }
func (RGBA) IdxB() int { return 2 }
func (RGBA) IdxA() int { return 3 }

type ARGB struct{}

func (ARGB) IdxA() int { return 0 }
func (ARGB) IdxR() int { return 1 }
func (ARGB) IdxG() int { return 2 }
func (ARGB) IdxB() int { return 3 }

type BGRA struct{}

func (BGRA) IdxB() int { return 0 }
func (BGRA) IdxG() int { return 1 }
func (BGRA) IdxR() int { return 2 }
func (BGRA) IdxA() int { return 3 }

type ABGR struct{}

func (ABGR) IdxA() int { return 0 }
func (ABGR) IdxB() int { return 1 }
func (ABGR) IdxG() int { return 2 }
func (ABGR) IdxR() int { return 3 }

type RGBOrder interface {
	IdxR() int
	IdxG() int
	IdxB() int
}

type RGB struct{}

func (RGB) IdxR() int { return 0 }
func (RGB) IdxG() int { return 1 }
func (RGB) IdxB() int { return 2 }

type BGR struct{}

func (BGR) IdxR() int { return 2 }
func (BGR) IdxG() int { return 1 }
func (BGR) IdxB() int { return 0 }

// 32-bit padded RGB orders (one padding byte, not addressed here).
// These still satisfy RGBOrder; indices may be 0..3.
type RGBX32 struct{}     // [R,G,B,X]
func (RGBX32) IdxR() int { return 0 }
func (RGBX32) IdxG() int { return 1 }
func (RGBX32) IdxB() int { return 2 }

type XRGB32 struct{}     // [X,R,G,B]
func (XRGB32) IdxR() int { return 1 }
func (XRGB32) IdxG() int { return 2 }
func (XRGB32) IdxB() int { return 3 }

type BGRX32 struct{}     // [B,G,R,X]
func (BGRX32) IdxR() int { return 2 }
func (BGRX32) IdxG() int { return 1 }
func (BGRX32) IdxB() int { return 0 }

type XBGR32 struct{}     // [X,B,G,R]
func (XBGR32) IdxR() int { return 3 }
func (XBGR32) IdxG() int { return 2 }
func (XBGR32) IdxB() int { return 1 }

// Compile-time interface checks
var (
	_ RGBOrder  = RGB{}
	_ RGBOrder  = BGR{}
	_ RGBOrder  = RGBX32{}
	_ RGBOrder  = XRGB32{}
	_ RGBOrder  = BGRX32{}
	_ RGBOrder  = XBGR32{}
	_ RGBAOrder = RGBA{}
	_ RGBAOrder = ARGB{}
	_ RGBAOrder = BGRA{}
	_ RGBAOrder = ABGR{}
)
