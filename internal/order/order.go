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
