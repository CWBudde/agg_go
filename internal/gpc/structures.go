// Internal GPC data structures (vertex, edge, polygon, scanbeam and
// intersection tables) and the list operations on them, split out of gpc.go.

package gpc

import "math"

// Internal data structures for the GPC algorithm
// These mirror the C structures but adapted for Go

// vertexType represents edge intersection classes
type vertexType int

const (
	vtxNUL vertexType = iota // Empty non-intersection
	vtxEMX                   // External maximum
	vtxELI                   // External left intermediate
	vtxTED                   // Top edge
	vtxERI                   // External right intermediate
	vtxRED                   // Right edge
	vtxIMM                   // Internal maximum and minimum
	vtxIMN                   // Internal minimum
	vtxEMN                   // External minimum
	vtxEMM                   // External maximum and minimum
	vtxLED                   // Left edge
	vtxILI                   // Internal left intermediate
	vtxBED                   // Bottom edge
	vtxIRI                   // Internal right intermediate
	vtxIMX                   // Internal maximum
	vtxFUL                   // Full non-intersection
)

// hState represents horizontal edge states
type hState int

const (
	hsNH hState = iota // No horizontal edge
	hsBH               // Bottom horizontal edge
	hsTH               // Top horizontal edge
)

// bundleState represents edge bundle state
type bundleState int

const (
	bsUnbundled  bundleState = iota // Isolated edge not within a bundle
	bsBundleHead                    // Bundle head node
	bsBundleTail                    // Passive bundle tail node
)

// vertexNode represents internal vertex list datatype
type vertexNode struct {
	X    float64     // X coordinate component
	Y    float64     // Y coordinate component
	Next *vertexNode // Pointer to next vertex in list
}

// polygonNode represents internal contour / tristrip type
type polygonNode struct {
	Active int            // Active flag / vertex count
	Hole   bool           // Hole / external contour flag
	V      [2]*vertexNode // Left and right vertex list ptrs
	Next   *polygonNode   // Pointer to next polygon contour
	Proxy  *polygonNode   // Pointer to actual structure used
}

// edgeNode represents an edge in the active edge table
type edgeNode struct {
	Vertex    GPCVertex       // Piggy-backed contour vertex data
	Bot       GPCVertex       // Edge lower (x, y) coordinate
	Top       GPCVertex       // Edge upper (x, y) coordinate
	XB        float64         // Scanbeam bottom x coordinate
	XT        float64         // Scanbeam top x coordinate
	DX        float64         // Change in x for a unit y increase
	Type      int             // Clip / subject edge flag
	Bundle    [2][2]int       // Bundle edge flags
	BSide     [2]int          // Bundle left / right indicators
	BState    [2]bundleState  // Edge bundle state
	OutP      [2]*polygonNode // Output polygon / tristrip pointer
	Prev      *edgeNode       // Previous edge in the AET
	Next      *edgeNode       // Next edge in the AET
	Pred      *edgeNode       // Edge connected at the lower end
	Succ      *edgeNode       // Edge connected at the upper end
	NextBound *edgeNode       // Pointer to next bound in LMT
}

// lmtNode represents local minima table
type lmtNode struct {
	Y          float64   // Y coordinate at local minimum
	FirstBound *edgeNode // Pointer to bound list
	Next       *lmtNode  // Pointer to next local minimum
}

// sbTree represents scanbeam tree
type sbTree struct {
	Y    float64 // Scanbeam node y value
	Less *sbTree // Pointer to nodes with lower y
	More *sbTree // Pointer to nodes with higher y
}

// itNode represents intersection table
type itNode struct {
	IE    [2]*edgeNode // Intersecting edge (bundle) pair
	Point GPCVertex    // Point of intersection
	Next  *itNode      // The next intersection table node
}

// stNode represents sorted edge table
type stNode struct {
	Edge *edgeNode // Pointer to AET edge
	XB   float64   // Scanbeam bottom x coordinate
	XT   float64   // Scanbeam top x coordinate
	DX   float64   // Change in x for a unit y increase
	Prev *stNode   // Previous edge in sorted list
}

// boundingBox represents contour axis-aligned bounding box
type boundingBox struct {
	XMin, YMin, XMax, YMax float64
}

// Constants for edge types
const (
	CLIP = 0
	SUBJ = 1
)

// Constants for left/right
const (
	LEFT  = 0
	RIGHT = 1
)

// Constants for above/below
const (
	ABOVE = 0
	BELOW = 1
)

// Horizontal edge state transitions within scanbeam boundary
var nextHState = [3][6]hState{
	//        ABOVE     BELOW     CROSS
	//        L   R     L   R     L   R
	/* NH */ {hsBH, hsTH, hsTH, hsBH, hsNH, hsNH},
	/* BH */ {hsNH, hsNH, hsNH, hsNH, hsTH, hsTH},
	/* TH */ {hsNH, hsNH, hsNH, hsNH, hsBH, hsBH},
}

// Internal helper functions for GPC algorithm

// resetIntersectionTable clears the intersection table
func resetIntersectionTable(it **itNode) {
	for *it != nil {
		next := (*it).Next
		*it = next
	}
}

// insertBound inserts an edge into the bound list, maintaining X-coordinate order
func insertBound(b **edgeNode, e *edgeNode) {
	if *b == nil {
		// Link node e to the tail of the list
		*b = e
		return
	}

	// Do primary sort on the x field
	switch {
	case e.Bot.X < (*b).Bot.X:
		// Insert a new node at the head
		existing := *b
		*b = e
		(*b).NextBound = existing
	case e.Bot.X == (*b).Bot.X:
		// Do secondary sort on the dx field
		if e.DX < (*b).DX {
			// Insert a new node at the head
			existing := *b
			*b = e
			(*b).NextBound = existing
		} else {
			// Head further down the list
			insertBound(&(*b).NextBound, e)
		}
	default:
		// Head further down the list
		insertBound(&(*b).NextBound, e)
	}
}

// boundList returns or creates a bound list for the given Y coordinate
func boundList(lmt **lmtNode, y float64) **edgeNode {
	if *lmt == nil {
		// Add node onto the tail end of the LMT
		*lmt = &lmtNode{
			Y:          y,
			FirstBound: nil,
			Next:       nil,
		}
		return &(*lmt).FirstBound
	}

	if y < (*lmt).Y {
		// Insert a new LMT node before the current node
		existing := *lmt
		*lmt = &lmtNode{
			Y:          y,
			FirstBound: nil,
			Next:       existing,
		}
		return &(*lmt).FirstBound
	}

	if y > (*lmt).Y {
		// Head further up the LMT
		return boundList(&(*lmt).Next, y)
	}

	// Use this existing LMT node
	return &(*lmt).FirstBound
}

// addToScanBeamTree adds a Y coordinate to the scan beam tree
func addToScanBeamTree(entries *int, sbtree **sbTree, y float64) {
	if *sbtree == nil {
		// Add a new tree node here
		*sbtree = &sbTree{
			Y:    y,
			Less: nil,
			More: nil,
		}
		*entries++
		return
	}

	if (*sbtree).Y > y {
		// Head into the 'less' sub-tree
		addToScanBeamTree(entries, &(*sbtree).Less, y)
	} else if (*sbtree).Y < y {
		// Head into the 'more' sub-tree
		addToScanBeamTree(entries, &(*sbtree).More, y)
	}
	// Equal values are ignored (no duplicates)
}

// buildScanBeamTable converts the scan beam tree to a sorted array
func buildScanBeamTable(entries *int, sbt []float64, sbtree *sbTree) {
	if sbtree.Less != nil {
		buildScanBeamTable(entries, sbt, sbtree.Less)
	}
	sbt[*entries] = sbtree.Y
	*entries++
	if sbtree.More != nil {
		buildScanBeamTable(entries, sbt, sbtree.More)
	}
}

// Helper macros adapted from C
func prevIndex(i, n int) int {
	return (i - 1 + n) % n
}

func nextIndex(i, n int) int {
	return (i + 1) % n
}

// optimal checks if vertex i is optimal (not embedded in horizontal edges)
func optimal(vertices []GPCVertex, i, n int) bool {
	return vertices[prevIndex(i, n)].Y != vertices[i].Y || vertices[nextIndex(i, n)].Y != vertices[i].Y
}

// fwdMin checks if vertex i is a forward local minimum
func fwdMin(vertices []GPCVertex, i, n int) bool {
	return vertices[prevIndex(i, n)].Y >= vertices[i].Y && vertices[nextIndex(i, n)].Y > vertices[i].Y
}

// revMin checks if vertex i is a reverse local minimum
func revMin(vertices []GPCVertex, i, n int) bool {
	return vertices[prevIndex(i, n)].Y > vertices[i].Y && vertices[nextIndex(i, n)].Y >= vertices[i].Y
}

// notFMax checks if vertex i is not a forward maximum
func notFMax(vertices []GPCVertex, i, n int) bool {
	return vertices[nextIndex(i, n)].Y > vertices[i].Y
}

// notRMax checks if vertex i is not a reverse maximum
func notRMax(vertices []GPCVertex, i, n int) bool {
	return vertices[prevIndex(i, n)].Y > vertices[i].Y
}

// countOptimalVertices counts vertices that are not embedded in horizontal edges
func countOptimalVertices(contour *GPCVertexList) int {
	if contour.NumVertices <= 0 {
		return 0
	}

	result := 0
	for i := 0; i < contour.NumVertices; i++ {
		if optimal(contour.Vertices, i, contour.NumVertices) {
			result++
		}
	}
	return result
}

// buildLocalMinimaTable constructs the Local Minima Table from input polygons
func buildLocalMinimaTable(lmt **lmtNode, sbtree **sbTree, sbtEntries *int,
	polygon *GPCPolygon, edgeType int, operation GPCOp,
) []*edgeNode {
	if polygon.NumContours == 0 {
		return nil
	}

	// Count total optimal vertices
	totalVertices := 0
	for c := 0; c < polygon.NumContours; c++ {
		contour, _, err := polygon.GetContour(c)
		if err == nil {
			totalVertices += countOptimalVertices(contour)
		}
	}

	if totalVertices == 0 {
		return nil
	}

	// Create edge table
	edgeTable := make([]*edgeNode, totalVertices)
	for i := range edgeTable {
		edgeTable[i] = &edgeNode{}
	}

	edgeIndex := 0

	// Process each contour
	for c := 0; c < polygon.NumContours; c++ {
		contour, _, err := polygon.GetContour(c)
		if err != nil {
			continue
		}

		if contour.NumVertices <= 0 {
			continue
		}

		// Collect optimal vertices
		optimalVertices := make([]GPCVertex, 0, contour.NumVertices)
		for i := 0; i < contour.NumVertices; i++ {
			if optimal(contour.Vertices, i, contour.NumVertices) {
				optimalVertices = append(optimalVertices, contour.Vertices[i])
				// Record vertex in scanbeam tree
				addToScanBeamTree(sbtEntries, sbtree, contour.Vertices[i].Y)
			}
		}

		numVertices := len(optimalVertices)
		if numVertices < 3 {
			continue // Skip degenerate contours
		}

		// Process forward local minima
		for minIdx := 0; minIdx < numVertices; minIdx++ {
			if fwdMin(optimalVertices, minIdx, numVertices) {
				// Find next local maximum
				numEdges := 1
				maxIdx := nextIndex(minIdx, numVertices)
				for notFMax(optimalVertices, maxIdx, numVertices) {
					numEdges++
					maxIdx = nextIndex(maxIdx, numVertices)
				}

				// Build edge list for this minimum
				if edgeIndex+numEdges <= len(edgeTable) {
					e := edgeTable[edgeIndex : edgeIndex+numEdges]
					edgeIndex += numEdges
					v := minIdx

					for i := 0; i < numEdges; i++ {
						e[i].Bot = optimalVertices[v]
						e[i].XB = optimalVertices[v].X
						v = nextIndex(v, numVertices)
						e[i].Top = optimalVertices[v]

						// Calculate dx (change in x per unit y)
						if e[i].Top.Y != e[i].Bot.Y {
							e[i].DX = (e[i].Top.X - e[i].Bot.X) / (e[i].Top.Y - e[i].Bot.Y)
						} else {
							e[i].DX = 0 // Horizontal edge
						}

						e[i].Type = edgeType
						e[i].Vertex = e[i].Bot

						// Set edge linkages
						if numEdges > 1 && i < numEdges-1 {
							e[i].Succ = e[i+1]
						}
						if numEdges > 1 && i > 0 {
							e[i].Pred = e[i-1]
						}

						// Set bundle sides based on operation
						if operation == GPCDiff {
							e[i].BSide[CLIP] = RIGHT
						} else {
							e[i].BSide[CLIP] = LEFT
						}
						e[i].BSide[SUBJ] = LEFT
					}

					// Insert into bound list
					insertBound(boundList(lmt, optimalVertices[minIdx].Y), e[0])
				}
			}
		}

		// Process reverse local minima
		for minIdx := 0; minIdx < numVertices; minIdx++ {
			if revMin(optimalVertices, minIdx, numVertices) {
				// Find previous local maximum
				numEdges := 1
				maxIdx := prevIndex(minIdx, numVertices)
				for notRMax(optimalVertices, maxIdx, numVertices) {
					numEdges++
					maxIdx = prevIndex(maxIdx, numVertices)
				}

				// Build edge list for this minimum
				if edgeIndex+numEdges <= len(edgeTable) {
					e := edgeTable[edgeIndex : edgeIndex+numEdges]
					edgeIndex += numEdges
					v := minIdx

					for i := 0; i < numEdges; i++ {
						e[i].Bot = optimalVertices[v]
						e[i].XB = optimalVertices[v].X
						v = prevIndex(v, numVertices)
						e[i].Top = optimalVertices[v]

						// Calculate dx (change in x per unit y)
						if e[i].Top.Y != e[i].Bot.Y {
							e[i].DX = (e[i].Top.X - e[i].Bot.X) / (e[i].Top.Y - e[i].Bot.Y)
						} else {
							e[i].DX = 0 // Horizontal edge
						}

						e[i].Type = edgeType
						e[i].Vertex = e[i].Bot

						// Set edge linkages
						if numEdges > 1 && i < numEdges-1 {
							e[i].Succ = e[i+1]
						}
						if numEdges > 1 && i > 0 {
							e[i].Pred = e[i-1]
						}

						// Set bundle sides
						if operation == GPCDiff {
							e[i].BSide[CLIP] = RIGHT
						} else {
							e[i].BSide[CLIP] = LEFT
						}
						e[i].BSide[SUBJ] = LEFT
					}

					// Insert into bound list
					insertBound(boundList(lmt, optimalVertices[minIdx].Y), e[0])
				}
			}
		}
	}

	return edgeTable
}

// addEdgeToAET adds an edge to the Active Edge Table, maintaining X-coordinate order
func addEdgeToAET(aet **edgeNode, edge, prev *edgeNode) {
	if *aet == nil {
		// Append edge onto the tail end of the AET
		*aet = edge
		edge.Prev = prev
		edge.Next = nil
		return
	}

	// Do primary sort on the xb field
	switch {
	case edge.XB < (*aet).XB:
		// Insert edge here (before the AET edge)
		edge.Prev = prev
		edge.Next = *aet
		(*aet).Prev = edge
		*aet = edge
	case edge.XB == (*aet).XB:
		// Do secondary sort on the dx field
		if edge.DX < (*aet).DX {
			// Insert edge here (before the AET edge)
			edge.Prev = prev
			edge.Next = *aet
			(*aet).Prev = edge
			*aet = edge
		} else {
			// Head further into the AET
			addEdgeToAET(&(*aet).Next, edge, *aet)
		}
	default:
		// Head further into the AET
		addEdgeToAET(&(*aet).Next, edge, *aet)
	}
}

// addIntersection records an edge intersection in the intersection table
func addIntersection(it **itNode, edge0, edge1 *edgeNode, x, y float64) {
	if *it == nil {
		// Append a new node to the tail of the list
		*it = &itNode{
			IE:    [2]*edgeNode{edge0, edge1},
			Point: GPCVertex{X: x, Y: y},
			Next:  nil,
		}
		return
	}

	if (*it).Point.Y > y {
		// Insert a new node mid-list
		existing := *it
		*it = &itNode{
			IE:    [2]*edgeNode{edge0, edge1},
			Point: GPCVertex{X: x, Y: y},
			Next:  existing,
		}
	} else {
		// Head further down the list
		addIntersection(&(*it).Next, edge0, edge1, x, y)
	}
}

// addSortedEdge adds an edge to the sorted edge table for intersection detection
func addSortedEdge(st **stNode, it **itNode, edge *edgeNode, dy float64) {
	if *st == nil {
		// Append edge onto the tail end of the ST
		*st = &stNode{
			Edge: edge,
			XB:   edge.XB,
			XT:   edge.XT,
			DX:   edge.DX,
			Prev: nil,
		}
		return
	}

	den := ((*st).XT - (*st).XB) - (edge.XT - edge.XB)

	// If new edge and ST edge don't cross
	if (edge.XT >= (*st).XT) || (edge.DX == (*st).DX) || (math.Abs(den) <= Epsilon) {
		// No intersection - insert edge here (before the ST edge)
		existing := *st
		*st = &stNode{
			Edge: edge,
			XB:   edge.XB,
			XT:   edge.XT,
			DX:   edge.DX,
			Prev: existing,
		}
	} else {
		// Compute intersection between new edge and ST edge
		r := (edge.XB - (*st).XB) / den
		x := (*st).XB + r*((*st).XT-(*st).XB)
		y := r * dy

		// Insert the edge pointers and the intersection point in the IT
		addIntersection(it, (*st).Edge, edge, x, y)

		// Head further into the ST
		addSortedEdge(&(*st).Prev, it, edge, dy)
	}
}

// buildIntersectionTable constructs intersection table for the current scanbeam
func buildIntersectionTable(it **itNode, aet *edgeNode, dy float64) {
	var st *stNode

	// Build intersection table for the current scanbeam
	resetIntersectionTable(it)
	st = nil

	// Process each AET edge
	for edge := aet; edge != nil; edge = edge.Next {
		if (edge.BState[ABOVE] == bsBundleHead) ||
			edge.Bundle[ABOVE][CLIP] != 0 || edge.Bundle[ABOVE][SUBJ] != 0 {
			addSortedEdge(&st, it, edge, dy)
		}
	}

	// Free the sorted edge table
	for st != nil {
		prev := st.Prev
		st = prev
	}
}

// countContours counts valid contours in the polygon output
func countContours(polygon *polygonNode) int {
	nc := 0
	for p := polygon; p != nil; p = p.Next {
		if p.Active != 0 {
			// Count the vertices in the current contour
			nv := 0
			for v := p.Proxy.V[LEFT]; v != nil; v = v.Next {
				nv++
			}

			// Record valid vertex counts in the active field
			if nv > 2 {
				p.Active = nv
				nc++
			} else {
				// Invalid contour: mark as inactive
				p.Active = 0
			}
		}
	}
	return nc
}

// addLeft adds a vertex to the left end of a polygon's vertex list
func addLeft(p *polygonNode, x, y float64) {
	nv := &vertexNode{
		X:    x,
		Y:    y,
		Next: p.Proxy.V[LEFT],
	}
	p.Proxy.V[LEFT] = nv
}

// addRight adds a vertex to the right end of a polygon's vertex list
func addRight(p *polygonNode, x, y float64) {
	nv := &vertexNode{
		X:    x,
		Y:    y,
		Next: nil,
	}

	if p.Proxy.V[RIGHT] != nil {
		p.Proxy.V[RIGHT].Next = nv
		p.Proxy.V[RIGHT] = nv
	} else {
		// If no right pointer, this becomes both left and right
		p.Proxy.V[LEFT] = nv
		p.Proxy.V[RIGHT] = nv
	}
}

// mergeLeft merges left polygon chains and labels contour as hole
func mergeLeft(p, q, list *polygonNode) {
	// Label contour as a hole
	q.Proxy.Hole = true

	if p.Proxy != q.Proxy {
		// Assign p's vertex list to the left end of q's list
		if p.Proxy.V[RIGHT] != nil {
			p.Proxy.V[RIGHT].Next = q.Proxy.V[LEFT]
		}
		q.Proxy.V[LEFT] = p.Proxy.V[LEFT]

		// Redirect any p.Proxy references to q.Proxy
		target := p.Proxy
		for l := list; l != nil; l = l.Next {
			if l.Proxy == target {
				l.Active = 0 // Mark as inactive
				l.Proxy = q.Proxy
			}
		}
	}
}

// mergeRight merges right polygon chains and labels contour as external
func mergeRight(p, q, list *polygonNode) {
	// Label contour as external
	q.Proxy.Hole = false

	if p.Proxy != q.Proxy {
		// Assign p's vertex list to the right end of q's list
		if q.Proxy.V[RIGHT] != nil {
			q.Proxy.V[RIGHT].Next = p.Proxy.V[LEFT]
		}
		q.Proxy.V[RIGHT] = p.Proxy.V[RIGHT]

		// Redirect any p.Proxy references to q.Proxy
		target := p.Proxy
		for l := list; l != nil; l = l.Next {
			if l.Proxy == target {
				l.Active = 0 // Mark as inactive
				l.Proxy = q.Proxy
			}
		}
	}
}

// addLocalMin adds a local minimum vertex and creates new polygon node
func addLocalMin(p **polygonNode, edge *edgeNode, x, y float64) {
	existing := *p

	nv := &vertexNode{
		X:    x,
		Y:    y,
		Next: nil,
	}

	*p = &polygonNode{
		Proxy:  nil, // Will be set to self below
		Active: 1,   // TRUE equivalent
		Next:   existing,
		V:      [2]*vertexNode{nv, nv}, // Both LEFT and RIGHT point to new vertex
	}

	// Initialize proxy to point to p itself
	(*p).Proxy = *p

	// Assign polygon p to the edge
	edge.OutP[ABOVE] = *p
}

// countTristrips counts the number of triangle strips
func countTristrips(tn *polygonNode) int {
	total := 0
	for t := tn; t != nil; t = t.Next {
		if t.Active > 2 {
			total++
		}
	}
	return total
}

// addVertexToTristrip adds a vertex to a tristrip (different from general addVertex)
func addVertexToTristrip(t **vertexNode, x, y float64) {
	if *t == nil {
		*t = &vertexNode{
			X:    x,
			Y:    y,
			Next: nil,
		}
	} else {
		// Head further down the list
		addVertexToTristrip(&(*t).Next, x, y)
	}
}

// newTristrip creates a new triangle strip
func newTristrip(tn **polygonNode, edge *edgeNode, x, y float64) {
	if *tn == nil {
		*tn = &polygonNode{
			Next:   nil,
			V:      [2]*vertexNode{nil, nil},
			Active: 1,
		}
		addVertexToTristrip(&(*tn).V[LEFT], x, y)
		edge.OutP[ABOVE] = *tn
	} else {
		// Head further down the list
		newTristrip(&(*tn).Next, edge, x, y)
	}
}

// createContourBBoxes creates bounding boxes for all contours
func createContourBBoxes(p *GPCPolygon) []boundingBox {
	if p.NumContours == 0 {
		return nil
	}

	boxes := make([]boundingBox, p.NumContours)

	// Construct contour bounding boxes
	for c := 0; c < p.NumContours; c++ {
		contour, _, err := p.GetContour(c)
		if err != nil {
			continue
		}

		// Initialize bounding box extent
		boxes[c].XMin = math.MaxFloat64
		boxes[c].YMin = math.MaxFloat64
		boxes[c].XMax = -math.MaxFloat64
		boxes[c].YMax = -math.MaxFloat64

		for v := 0; v < contour.NumVertices; v++ {
			vertex := contour.Vertices[v]
			// Adjust bounding box
			if vertex.X < boxes[c].XMin {
				boxes[c].XMin = vertex.X
			}
			if vertex.Y < boxes[c].YMin {
				boxes[c].YMin = vertex.Y
			}
			if vertex.X > boxes[c].XMax {
				boxes[c].XMax = vertex.X
			}
			if vertex.Y > boxes[c].YMax {
				boxes[c].YMax = vertex.Y
			}
		}
	}
	return boxes
}

// minimaxTest performs bounding box overlap test to optimize clipping
func minimaxTest(subj, clip *GPCPolygon, op GPCOp) {
	if subj.NumContours == 0 || clip.NumContours == 0 {
		return
	}

	sBbox := createContourBBoxes(subj)
	cBbox := createContourBBoxes(clip)

	if sBbox == nil || cBbox == nil {
		return
	}

	// Create overlap table
	overlapTable := make([]bool, subj.NumContours*clip.NumContours)

	// Check all subject contour bounding boxes against clip boxes
	for s := 0; s < subj.NumContours; s++ {
		for c := 0; c < clip.NumContours; c++ {
			overlapTable[c*subj.NumContours+s] = !((sBbox[s].XMax < cBbox[c].XMin) ||
				(sBbox[s].XMin > cBbox[c].XMax)) &&
				!((sBbox[s].YMax < cBbox[c].YMin) ||
					(sBbox[s].YMin > cBbox[c].YMax))
		}
	}

	// For each clip contour, search for any subject contour overlaps
	for c := 0; c < clip.NumContours; c++ {
		overlap := false
		for s := 0; s < subj.NumContours && !overlap; s++ {
			overlap = overlapTable[c*subj.NumContours+s]
		}

		if !overlap {
			// Flag non-contributing status by negating vertex count
			contour, _, err := clip.GetContour(c)
			if err == nil {
				contour.NumVertices = -contour.NumVertices
			}
		}
	}

	if op == GPCInt {
		// For each subject contour, search for any clip contour overlaps
		for s := 0; s < subj.NumContours; s++ {
			overlap := false
			for c := 0; c < clip.NumContours && !overlap; c++ {
				overlap = overlapTable[c*subj.NumContours+s]
			}

			if !overlap {
				// Flag non-contributing status by negating vertex count
				contour, _, err := subj.GetContour(s)
				if err == nil {
					contour.NumVertices = -contour.NumVertices
				}
			}
		}
	}
}
