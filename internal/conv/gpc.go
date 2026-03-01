package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/gpc"
)

// GPCOp represents the type of boolean operation to perform
type GPCOp int

const (
	// GPCOr performs set union (A ∪ B)
	GPCOr GPCOp = iota
	// GPCAnd performs set intersection (A ∩ B)
	GPCAnd
	// GPCXor performs exclusive or (A ⊕ B)
	GPCXor
	// GPCAMinusB performs set difference (A - B)
	GPCAMinusB
	// GPCBMinusA performs set difference (B - A)
	GPCBMinusA
)

// String returns the string representation of the operation
func (op GPCOp) String() string {
	switch op {
	case GPCOr:
		return "Union"
	case GPCAnd:
		return "Intersection"
	case GPCXor:
		return "Exclusive-or"
	case GPCAMinusB:
		return "A-minus-B"
	case GPCBMinusA:
		return "B-minus-A"
	default:
		return "Unknown"
	}
}

// conversionStatus represents the internal state of the GPC conversion
type conversionStatus int

const (
	statusMoveTo conversionStatus = iota
	statusLineTo
	statusStop
)

// ContourHeader represents metadata for a single contour during accumulation
type ContourHeader struct {
	numVertices int
	holeFlag    bool
	vertices    []gpc.GPCVertex
}

func normalizeContourVertices(vertices []gpc.GPCVertex) []gpc.GPCVertex {
	if len(vertices) > 2 && vertices[0].Equal(vertices[len(vertices)-1]) {
		return vertices[:len(vertices)-1]
	}
	return vertices
}

// ConvGPC performs boolean clipping operations on two vertex sources using GPC.
//
// The low-level GPC package has working polygon clipping and polygon text I/O
// helpers. Remaining review work here is converter parity: some converter-driven
// example cases still emit empty output and need deeper comparison with AGG's
// original agg_conv_gpc.h behavior.
type ConvGPC[VSA VertexSource, VSB VertexSource] struct {
	srcA      VSA
	srcB      VSB
	operation GPCOp
	status    conversionStatus
	vertex    int
	contour   int

	// Accumulation storage
	vertexAccumulator  []gpc.GPCVertex
	contourAccumulator []ContourHeader

	// GPC polygons
	polyA  *gpc.GPCPolygon
	polyB  *gpc.GPCPolygon
	result *gpc.GPCPolygon
}

// NewConvGPC creates a new GPC converter with two vertex sources
func NewConvGPC[VSA VertexSource, VSB VertexSource](sourceA VSA, sourceB VSB, op GPCOp) *ConvGPC[VSA, VSB] {
	return &ConvGPC[VSA, VSB]{
		srcA:               sourceA,
		srcB:               sourceB,
		operation:          op,
		status:             statusMoveTo,
		vertex:             -1,
		contour:            -1,
		vertexAccumulator:  make([]gpc.GPCVertex, 0),
		contourAccumulator: make([]ContourHeader, 0),
		polyA:              gpc.NewGPCPolygon(),
		polyB:              gpc.NewGPCPolygon(),
		result:             gpc.NewGPCPolygon(),
	}
}

// Attach1 sets the first vertex source
func (c *ConvGPC[VSA, VSB]) Attach1(source VSA) {
	c.srcA = source
}

// Attach2 sets the second vertex source
func (c *ConvGPC[VSA, VSB]) Attach2(source VSB) {
	c.srcB = source
}

// Operation sets the boolean operation to perform
func (c *ConvGPC[VSA, VSB]) Operation(op GPCOp) {
	c.operation = op
}

// Rewind implements the VertexSource interface - prepares the clipped result
func (c *ConvGPC[VSA, VSB]) Rewind(pathID uint) {
	// Clear previous result
	c.freeResult()

	// Process source A and B into GPC polygons
	c.srcA.Rewind(pathID)
	c.srcB.Rewind(pathID)
	c.addToPolygon(c.srcA, c.polyA)
	c.addToPolygon(c.srcB, c.polyB)

	// Perform the clipping operation
	var err error
	switch c.operation {
	case GPCOr:
		c.result, err = gpc.PolygonClip(gpc.GPCUnion, c.polyA, c.polyB)
	case GPCAnd:
		c.result, err = gpc.PolygonClip(gpc.GPCInt, c.polyA, c.polyB)
	case GPCXor:
		c.result, err = gpc.PolygonClip(gpc.GPCXor, c.polyA, c.polyB)
	case GPCAMinusB:
		c.result, err = gpc.PolygonClip(gpc.GPCDiff, c.polyA, c.polyB)
	case GPCBMinusA:
		c.result, err = gpc.PolygonClip(gpc.GPCDiff, c.polyB, c.polyA)
	}

	// On error, result will be empty but we continue
	if err != nil {
		// Log error but continue with empty result - maintains vertex source interface contract
		// In a production system, consider adding a logging framework
		c.result.Clear()
	}

	// Start extracting vertices from result
	c.startExtracting()
}

// Vertex implements the VertexSource interface - returns the next vertex
func (c *ConvGPC[VSA, VSB]) Vertex() (x, y float64, cmd basics.PathCommand) {
	if c.status == statusMoveTo {
		if c.nextContour() {
			if c.nextVertex(&x, &y) {
				c.status = statusLineTo
				return x, y, basics.PathCmdMoveTo
			}
			c.status = statusStop
			return 0, 0, basics.PathCmdEndPoly | basics.PathFlagClose
		}
	} else {
		if c.nextVertex(&x, &y) {
			return x, y, basics.PathCmdLineTo
		} else {
			c.status = statusMoveTo
		}
		return 0, 0, basics.PathCmdEndPoly | basics.PathFlagClose
	}
	return 0, 0, basics.PathCmdStop
}

// addToPolygon converts a vertex source to a GPC polygon
func (c *ConvGPC[VSA, VSB]) addToPolygon(vs VertexSource, polygon *gpc.GPCPolygon) {
	// Clear existing contours
	polygon.Clear()
	c.contourAccumulator = c.contourAccumulator[:0]

	var startX, startY float64
	var lineTo bool
	var orientation uint

	vs.Rewind(0)
	for {
		x, y, cmd := vs.Vertex()

		if basics.IsStop(cmd) {
			break
		}

		if basics.IsVertex(cmd) {
			if basics.IsMoveTo(cmd) {
				if lineTo {
					c.endContour(orientation)
					orientation = 0
				}
				c.startContour()
				startX = x
				startY = y
			}
			c.addVertex(x, y)
			lineTo = true
		} else {
			if basics.IsEndPoly(cmd) {
				orientation = uint(basics.GetOrientation(uint32(cmd)))
				if lineTo && basics.IsClosed(uint32(cmd)) {
					c.addVertex(startX, startY)
				}
			}
		}
	}

	if lineTo {
		c.endContour(orientation)
	}

	c.makePolygon(polygon)
}

// startContour begins accumulating a new contour
func (c *ConvGPC[VSA, VSB]) startContour() {
	header := ContourHeader{
		numVertices: 0,
		holeFlag:    false,
		vertices:    nil,
	}
	c.contourAccumulator = append(c.contourAccumulator, header)
	c.vertexAccumulator = c.vertexAccumulator[:0]
}

// addVertex adds a vertex to the current contour
func (c *ConvGPC[VSA, VSB]) addVertex(x, y float64) {
	vertex := gpc.GPCVertex{X: x, Y: y}
	c.vertexAccumulator = append(c.vertexAccumulator, vertex)
}

// endContour finishes the current contour and prepares it for GPC
func (c *ConvGPC[VSA, VSB]) endContour(orientation uint) {
	if len(c.contourAccumulator) == 0 {
		return
	}

	normalizedVertices := normalizeContourVertices(c.vertexAccumulator)
	if len(normalizedVertices) > 2 {
		// Get the last contour header
		headerIdx := len(c.contourAccumulator) - 1
		header := &c.contourAccumulator[headerIdx]

		header.numVertices = len(normalizedVertices)
		// Set hole flag based on orientation - clockwise polygons are holes
		// This matches the C++ logic: if(is_cw(orientation)) h.hole_flag = 1;
		header.holeFlag = (orientation & uint(basics.PathFlagsCW)) != 0

		// Copy vertices
		header.vertices = make([]gpc.GPCVertex, len(normalizedVertices))
		copy(header.vertices, normalizedVertices)
	} else {
		// Remove incomplete contour (equivalent to C++ remove_last)
		if len(c.contourAccumulator) > 0 {
			c.contourAccumulator = c.contourAccumulator[:len(c.contourAccumulator)-1]
		}
	}
}

// makePolygon creates a GPC polygon from accumulated contours
func (c *ConvGPC[VSA, VSB]) makePolygon(polygon *gpc.GPCPolygon) {
	polygon.Clear()

	for _, header := range c.contourAccumulator {
		if header.numVertices > 0 && header.vertices != nil {
			contour := gpc.NewGPCVertexList(header.numVertices)
			for _, vertex := range header.vertices {
				contour.AddVertex(vertex.X, vertex.Y)
			}
			err := polygon.AddContour(contour, header.holeFlag)
			if err != nil {
				// Skip this contour but continue processing others
				// In a production system, consider logging this error
				continue
			}
		}
	}
}

// startExtracting prepares for vertex extraction from the result
func (c *ConvGPC[VSA, VSB]) startExtracting() {
	c.status = statusMoveTo
	c.contour = -1
	c.vertex = -1
}

// nextContour moves to the next contour in the result
func (c *ConvGPC[VSA, VSB]) nextContour() bool {
	c.contour++
	if c.contour < c.result.NumContours {
		c.vertex = -1
		return true
	}
	return false
}

// nextVertex gets the next vertex from the current contour
func (c *ConvGPC[VSA, VSB]) nextVertex(x, y *float64) bool {
	if c.result.NumContours == 0 || c.contour >= c.result.NumContours {
		return false
	}

	contour, _, err := c.result.GetContour(c.contour)
	if err != nil {
		return false
	}

	c.vertex++
	if c.vertex < contour.NumVertices {
		vertex, err := contour.GetVertex(c.vertex)
		if err != nil {
			return false
		}
		*x = vertex.X
		*y = vertex.Y
		return true
	}
	return false
}

// freeResult clears the result polygon
func (c *ConvGPC[VSA, VSB]) freeResult() {
	c.result.Clear()
}
