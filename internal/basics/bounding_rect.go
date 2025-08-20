package basics

// VertexSource interface for providing vertices
type VertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd PathCommand)
}

// GetID interface for getting path IDs from a collection
type GetID interface {
	Get(index uint) uint
}

// SliceGetID implements GetID interface for a slice of uints
type SliceGetID []uint

func (s SliceGetID) Get(index uint) uint {
	if index >= uint(len(s)) {
		return 0
	}
	return s[index]
}

// BoundingRect calculates the bounding rectangle for multiple paths from a vertex source.
// It takes a vertex source, a GetID interface for path IDs, start index, number of paths,
// and returns the bounding rectangle and whether it's valid (contains vertices).
func BoundingRect[T ~int | ~int32 | ~float32 | ~float64](
	vs VertexSource,
	gi GetID,
	start, num uint,
) (Rect[T], bool) {
	rect := Rect[T]{
		X1: T(1), Y1: T(1),
		X2: T(0), Y2: T(0),
	}

	first := true

	for i := uint(0); i < num; i++ {
		pathID := gi.Get(start + i)
		vs.Rewind(pathID)

		for {
			x, y, cmd := vs.Vertex()
			if IsStop(cmd) {
				break
			}

			if IsVertex(cmd) {
				if first {
					rect.X1 = T(x)
					rect.Y1 = T(y)
					rect.X2 = T(x)
					rect.Y2 = T(y)
					first = false
				} else {
					if T(x) < rect.X1 {
						rect.X1 = T(x)
					}
					if T(y) < rect.Y1 {
						rect.Y1 = T(y)
					}
					if T(x) > rect.X2 {
						rect.X2 = T(x)
					}
					if T(y) > rect.Y2 {
						rect.Y2 = T(y)
					}
				}
			}
		}
	}

	return rect, rect.X1 <= rect.X2 && rect.Y1 <= rect.Y2
}

// BoundingRectSingle calculates the bounding rectangle for a single path from a vertex source.
// It takes a vertex source and a path ID, and returns the bounding rectangle and whether it's valid.
func BoundingRectSingle[T ~int | ~int32 | ~float32 | ~float64](
	vs VertexSource,
	pathID uint,
) (Rect[T], bool) {
	rect := Rect[T]{
		X1: T(1), Y1: T(1),
		X2: T(0), Y2: T(0),
	}

	first := true
	vs.Rewind(pathID)

	for {
		x, y, cmd := vs.Vertex()
		if IsStop(cmd) {
			break
		}

		if IsVertex(cmd) {
			if first {
				rect.X1 = T(x)
				rect.Y1 = T(y)
				rect.X2 = T(x)
				rect.Y2 = T(y)
				first = false
			} else {
				if T(x) < rect.X1 {
					rect.X1 = T(x)
				}
				if T(y) < rect.Y1 {
					rect.Y1 = T(y)
				}
				if T(x) > rect.X2 {
					rect.X2 = T(x)
				}
				if T(y) > rect.Y2 {
					rect.Y2 = T(y)
				}
			}
		}
	}

	return rect, rect.X1 <= rect.X2 && rect.Y1 <= rect.Y2
}
