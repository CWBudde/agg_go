package basics

// Clipping flag constants for the Cyrus-Beck line clipping algorithm
const (
	ClippingFlagsX1Clipped = 4
	ClippingFlagsX2Clipped = 1
	ClippingFlagsY1Clipped = 8
	ClippingFlagsY2Clipped = 2
	ClippingFlagsXClipped  = ClippingFlagsX1Clipped | ClippingFlagsX2Clipped
	ClippingFlagsYClipped  = ClippingFlagsY1Clipped | ClippingFlagsY2Clipped
)

// ClippingFlags determines the clipping code of the vertex according to the
// Cyrus-Beck line clipping algorithm
//
//	      |        |
//	0110  |  0010  | 0011
//	      |        |
//
// -------+--------+-------- clip_box.y2
//
//	      |        |
//	0100  |  0000  | 0001
//	      |        |
//
// -------+--------+-------- clip_box.y1
//
//	      |        |
//	1100  |  1000  | 1001
//	      |        |
//	clip_box.x1  clip_box.x2
func ClippingFlags[T CoordType](x, y T, clipBox Rect[T]) uint32 {
	var flags uint32
	if x > clipBox.X2 {
		flags |= ClippingFlagsX2Clipped
	}
	if y > clipBox.Y2 {
		flags |= ClippingFlagsY2Clipped
	}
	if x < clipBox.X1 {
		flags |= ClippingFlagsX1Clipped
	}
	if y < clipBox.Y1 {
		flags |= ClippingFlagsY1Clipped
	}
	return flags
}

// ClippingFlagsX determines clipping flags for X coordinate only
func ClippingFlagsX[T CoordType](x T, clipBox Rect[T]) uint32 {
	var flags uint32
	if x > clipBox.X2 {
		flags |= ClippingFlagsX2Clipped
	}
	if x < clipBox.X1 {
		flags |= ClippingFlagsX1Clipped
	}
	return flags
}

// ClippingFlagsY determines clipping flags for Y coordinate only
func ClippingFlagsY[T CoordType](y T, clipBox Rect[T]) uint32 {
	var flags uint32
	if y > clipBox.Y2 {
		flags |= ClippingFlagsY2Clipped
	}
	if y < clipBox.Y1 {
		flags |= ClippingFlagsY1Clipped
	}
	return flags
}

// ClipLiangBarsky implements the Liang-Barsky line clipping algorithm.
// Returns the number of clipped points (0-2) and stores them in x and y arrays.
// This is a direct translation of the AGG C++ implementation.
func ClipLiangBarsky[T CoordType](
	x1, y1, x2, y2 T,
	clipBox Rect[T],
	x, y []T,
) uint32 {
	const nearzero = 1e-30

	fx1, fy1 := float64(x1), float64(y1)
	fx2, fy2 := float64(x2), float64(y2)
	deltax := fx2 - fx1
	deltay := fy2 - fy1

	var xin, xout, yin, yout float64
	var tinx, tiny, toutx, touty float64
	var tin1, tin2, tout1 float64
	var np uint32

	if deltax == 0.0 {
		// bump off of the vertical
		if float64(x1) > float64(clipBox.X1) {
			deltax = -nearzero
		} else {
			deltax = nearzero
		}
	}

	if deltay == 0.0 {
		// bump off of the horizontal
		if float64(y1) > float64(clipBox.Y1) {
			deltay = -nearzero
		} else {
			deltay = nearzero
		}
	}

	if deltax > 0.0 {
		// points to right
		xin = float64(clipBox.X1)
		xout = float64(clipBox.X2)
	} else {
		xin = float64(clipBox.X2)
		xout = float64(clipBox.X1)
	}

	if deltay > 0.0 {
		// points up
		yin = float64(clipBox.Y1)
		yout = float64(clipBox.Y2)
	} else {
		yin = float64(clipBox.Y2)
		yout = float64(clipBox.Y1)
	}

	tinx = (xin - float64(x1)) / deltax
	tiny = (yin - float64(y1)) / deltay

	if tinx < tiny {
		// hits x first
		tin1 = tinx
		tin2 = tiny
	} else {
		// hits y first
		tin1 = tiny
		tin2 = tinx
	}

	if tin1 <= 1.0 {
		if 0.0 < tin1 {
			x[np] = T(xin)
			y[np] = T(yin)
			np++
		}

		if tin2 <= 1.0 {
			toutx = (xout - float64(x1)) / deltax
			touty = (yout - float64(y1)) / deltay

			if toutx < touty {
				tout1 = toutx
			} else {
				tout1 = touty
			}

			if tin2 > 0.0 || tout1 > 0.0 {
				if tin2 <= tout1 {
					if tin2 > 0.0 {
						if tinx > tiny {
							x[np] = T(xin)
							y[np] = T(float64(y1) + tinx*deltay)
						} else {
							x[np] = T(float64(x1) + tiny*deltax)
							y[np] = T(yin)
						}
						np++
					}

					if tout1 < 1.0 {
						if toutx < touty {
							x[np] = T(xout)
							y[np] = T(float64(y1) + toutx*deltay)
						} else {
							x[np] = T(float64(x1) + touty*deltax)
							y[np] = T(yout)
						}
					} else {
						x[np] = x2
						y[np] = y2
					}
					np++
				} else {
					if tinx > tiny {
						x[np] = T(xin)
						y[np] = T(yout)
					} else {
						x[np] = T(xout)
						y[np] = T(yin)
					}
					np++
				}
			}
		}
	}
	return np
}

// ClipMovePoint moves a point to the clipping boundary
func ClipMovePoint[T CoordType](
	x1, y1, x2, y2 T,
	clipBox Rect[T],
	x, y *T,
	flags uint32,
) bool {
	var bound T

	if flags&ClippingFlagsXClipped != 0 {
		if x1 == x2 {
			return false
		}
		if flags&ClippingFlagsX1Clipped != 0 {
			bound = clipBox.X1
		} else {
			bound = clipBox.X2
		}
		*y = T(float64(bound-x1)*float64(y2-y1)/float64(x2-x1) + float64(y1))
		*x = bound
	}

	flags = ClippingFlagsY(*y, clipBox)
	if flags&ClippingFlagsYClipped != 0 {
		if y1 == y2 {
			return false
		}
		if flags&ClippingFlagsY1Clipped != 0 {
			bound = clipBox.Y1
		} else {
			bound = clipBox.Y2
		}
		*x = T(float64(bound-y1)*float64(x2-x1)/float64(y2-y1) + float64(x1))
		*y = bound
	}
	return true
}

// ClipLineSegment clips a line segment to a rectangle
// Returns: ret >= 4        - Fully clipped
//
//	(ret & 1) != 0  - First point has been moved
//	(ret & 2) != 0  - Second point has been moved
func ClipLineSegment[T CoordType](
	x1, y1, x2, y2 *T,
	clipBox Rect[T],
) uint32 {
	f1 := ClippingFlags(*x1, *y1, clipBox)
	f2 := ClippingFlags(*x2, *y2, clipBox)
	var ret uint32

	if (f2 | f1) == 0 {
		// Fully visible
		return 0
	}

	if (f1&ClippingFlagsXClipped) != 0 &&
		(f1&ClippingFlagsXClipped) == (f2&ClippingFlagsXClipped) {
		// Fully clipped
		return 4
	}

	if (f1&ClippingFlagsYClipped) != 0 &&
		(f1&ClippingFlagsYClipped) == (f2&ClippingFlagsYClipped) {
		// Fully clipped
		return 4
	}

	tx1, ty1 := *x1, *y1
	tx2, ty2 := *x2, *y2

	if f1 != 0 {
		if !ClipMovePoint(tx1, ty1, tx2, ty2, clipBox, x1, y1, f1) {
			return 4
		}
		if *x1 == *x2 && *y1 == *y2 {
			return 4
		}
		ret |= 1
	}

	if f2 != 0 {
		if !ClipMovePoint(tx1, ty1, tx2, ty2, clipBox, x2, y2, f2) {
			return 4
		}
		if *x1 == *x2 && *y1 == *y2 {
			return 4
		}
		ret |= 2
	}

	return ret
}
