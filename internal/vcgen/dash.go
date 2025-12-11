package vcgen

import (
	"math"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// Status represents the state of the dash generator
type DashStatus int

const (
	DashStatusInitial DashStatus = iota
	DashStatusReady
	DashStatusPolyline
	DashStatusStop
)

// MaxDashes is the maximum number of dash elements (dash + gap pairs)
const MaxDashes = 32

// VCGenDash is the dash vertex generator.
// This is a port of AGG's vcgen_dash class.
type VCGenDash struct {
	dashes        [MaxDashes]float64                      // Dash pattern array
	totalDashLen  float64                                 // Total length of one dash cycle
	numDashes     uint                                    // Number of elements in dash array
	dashStart     float64                                 // Dash start offset
	shorten       float64                                 // Path shortening distance
	currDashStart float64                                 // Current dash start position
	currDash      uint                                    // Current dash index
	currRest      float64                                 // Remaining distance in current segment
	v1            *array.VertexDist        // Current vertex
	v2            *array.VertexDist        // Next vertex
	srcVertices   *array.VertexDistSequence // Source vertices storage
	closed        uint                     // Whether path is closed
	status        DashStatus                              // Current generator status
	srcVertex     uint                                    // Current source vertex index
}

// NewVCGenDash creates a new dash vertex generator
func NewVCGenDash() *VCGenDash {
	return &VCGenDash{
		totalDashLen:  0.0,
		numDashes:     0,
		dashStart:     0.0,
		shorten:       0.0,
		currDashStart: 0.0,
		currDash:      0,
		srcVertices:   array.NewVertexDistSequence(),
		closed:        0,
		status:        DashStatusInitial,
		srcVertex:     0,
	}
}

// RemoveAllDashes clears all dash patterns
func (d *VCGenDash) RemoveAllDashes() {
	d.totalDashLen = 0.0
	d.numDashes = 0
	d.currDashStart = 0.0
	d.currDash = 0
}

// AddDash adds a dash pattern (dash length + gap length)
func (d *VCGenDash) AddDash(dashLen, gapLen float64) {
	if d.numDashes < MaxDashes {
		d.totalDashLen += dashLen + gapLen
		d.dashes[d.numDashes] = dashLen
		d.numDashes++
		d.dashes[d.numDashes] = gapLen
		d.numDashes++
	}
}

// DashStart sets the dash start offset
func (d *VCGenDash) DashStart(ds float64) {
	d.dashStart = ds
	d.calcDashStart(math.Abs(ds))
}

// GetDashStart returns the current dash start offset
func (d *VCGenDash) GetDashStart() float64 {
	return d.dashStart
}

// calcDashStart calculates the dash start position within the pattern
func (d *VCGenDash) calcDashStart(ds float64) {
	d.currDash = 0
	d.currDashStart = 0.0
	for ds > 0.0 {
		if ds > d.dashes[d.currDash] {
			ds -= d.dashes[d.currDash]
			d.currDash++
			d.currDashStart = 0.0
			if d.currDash >= d.numDashes {
				d.currDash = 0
			}
		} else {
			d.currDashStart = ds
			ds = 0.0
		}
	}
}

// Shorten sets the path shortening distance
func (d *VCGenDash) Shorten(s float64) {
	d.shorten = s
}

// GetShorten returns the current path shortening distance
func (d *VCGenDash) GetShorten() float64 {
	return d.shorten
}

// RemoveAll clears all vertices and resets state
func (d *VCGenDash) RemoveAll() {
	d.status = DashStatusInitial
	d.srcVertices.RemoveAll()
	d.closed = 0
}

// AddVertex adds a vertex to the path
func (d *VCGenDash) AddVertex(x, y float64, cmd basics.PathCommand) {
	d.status = DashStatusInitial
	if basics.IsMoveTo(cmd) {
		d.srcVertices.ModifyLast(array.VertexDist{X: x, Y: y, Dist: 0.0})
	} else {
		if basics.IsVertex(cmd) {
			d.srcVertices.Add(array.VertexDist{X: x, Y: y, Dist: 0.0})
		} else {
			if basics.GetCloseFlag(uint32(cmd)) != 0 {
				d.closed = 1
			} else {
				d.closed = 0
			}
		}
	}
}

// PrepareSrc prepares the source vertices for processing
func (d *VCGenDash) PrepareSrc() {
	if d.status == DashStatusInitial {
		d.srcVertices.Close(d.closed != 0)
		array.ShortenPath(d.srcVertices, d.shorten, d.closed != 0)
	}
	d.status = DashStatusReady
	d.srcVertex = 0
}

// Rewind resets the vertex iterator
func (d *VCGenDash) Rewind(pathID uint) {
	d.PrepareSrc()
}

// Vertex returns the next vertex in the dash sequence
func (d *VCGenDash) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdMoveTo
	for !basics.IsStop(cmd) {
		switch d.status {
		case DashStatusInitial:
			d.Rewind(0)

		case DashStatusReady:
			if d.numDashes < 2 || d.srcVertices.Size() < 2 {
				cmd = basics.PathCmdStop
				break
			}
			d.status = DashStatusPolyline
			d.srcVertex = 1
			d.v1 = &[]array.VertexDist{d.srcVertices.Get(0)}[0]
			d.v2 = &[]array.VertexDist{d.srcVertices.Get(1)}[0]
			d.currRest = d.v1.Dist
			x = d.v1.X
			y = d.v1.Y
			if d.dashStart >= 0.0 {
				d.calcDashStart(d.dashStart)
			}
			return x, y, basics.PathCmdMoveTo

		case DashStatusPolyline:
			dashRest := d.dashes[d.currDash] - d.currDashStart

			if (d.currDash & 1) != 0 {
				cmd = basics.PathCmdMoveTo
			} else {
				cmd = basics.PathCmdLineTo
			}

			if d.currRest > dashRest {
				d.currRest -= dashRest
				d.currDash++
				if d.currDash >= d.numDashes {
					d.currDash = 0
				}
				d.currDashStart = 0.0
				x = d.v2.X - (d.v2.X-d.v1.X)*d.currRest/d.v1.Dist
				y = d.v2.Y - (d.v2.Y-d.v1.Y)*d.currRest/d.v1.Dist
			} else {
				d.currDashStart += d.currRest
				x = d.v2.X
				y = d.v2.Y
				d.srcVertex++
				d.v1 = d.v2
				d.currRest = d.v1.Dist
				if d.closed != 0 {
					if d.srcVertex > uint(d.srcVertices.Size()) {
						d.status = DashStatusStop
					} else {
						var nextIdx int
						if d.srcVertex >= uint(d.srcVertices.Size()) {
							nextIdx = 0
						} else {
							nextIdx = int(d.srcVertex)
						}
						d.v2 = &[]array.VertexDist{d.srcVertices.Get(nextIdx)}[0]
					}
				} else {
					if d.srcVertex >= uint(d.srcVertices.Size()) {
						d.status = DashStatusStop
					} else {
						d.v2 = &[]array.VertexDist{d.srcVertices.Get(int(d.srcVertex))}[0]
					}
				}
			}
			return x, y, cmd

		case DashStatusStop:
			cmd = basics.PathCmdStop
			break
		}
	}
	return 0, 0, basics.PathCmdStop
}
