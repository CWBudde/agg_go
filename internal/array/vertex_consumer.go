package array

import "agg_go/internal/basics"

// PodBVectorConsumer adapts PodBVector to implement VertexConsumer
type PodBVectorConsumer struct {
	vector *PodBVector[basics.PointD]
}

// NewPodBVectorConsumer creates a new consumer wrapping a PodBVector
func NewPodBVectorConsumer(vector *PodBVector[basics.PointD]) *PodBVectorConsumer {
	return &PodBVectorConsumer{vector: vector}
}

// Add implements VertexConsumer
func (pc *PodBVectorConsumer) Add(x, y float64) {
	pc.vector.Add(basics.PointD{X: x, Y: y})
}

// RemoveAll implements VertexConsumer
func (pc *PodBVectorConsumer) RemoveAll() {
	pc.vector.RemoveAll()
}
