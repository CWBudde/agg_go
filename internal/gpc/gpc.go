// Package gpc implements the General Polygon Clipper (GPC) algorithm.
//
// This is a Go port of the GPC library by Alan Murta, which provides
// polygon clipping operations including union, intersection, difference,
// and exclusive-or for arbitrary polygons with holes.
//
// Note: The GPC algorithm is subject to licensing restrictions for commercial use.
// This implementation is provided for non-commercial use. For commercial applications,
// consider alternatives like Clipper2 or Boost.Geometry.
//
// Original GPC by Alan Murta (email: gpc@cs.man.ac.uk)
// Version: 2.32, Date: 17th December 2004
// Copyright: (C) 1997-2004, Advanced Interfaces Group, University of Manchester.
package gpc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
)

const (
	// Version represents the GPC version this implementation is based on
	Version = "2.32"

	// Epsilon is used for floating-point comparisons
	// Increase to encourage merging of near coincident edges
	Epsilon = 1e-15
)

// GPCOp represents the type of boolean operation to perform
type GPCOp int

const (
	// GPCDiff performs set difference (A - B)
	GPCDiff GPCOp = iota
	// GPCInt performs set intersection (A ∩ B)
	GPCInt
	// GPCXor performs exclusive or (A ⊕ B)
	GPCXor
	// GPCUnion performs set union (A ∪ B)
	GPCUnion
)

// String returns the string representation of the operation
func (op GPCOp) String() string {
	switch op {
	case GPCDiff:
		return "Difference"
	case GPCInt:
		return "Intersection"
	case GPCXor:
		return "Exclusive-or"
	case GPCUnion:
		return "Union"
	default:
		return "Unknown"
	}
}

// GPCVertex represents a polygon vertex with x,y coordinates
type GPCVertex struct {
	X, Y float64
}

// Equal checks if two vertices are approximately equal within epsilon
func (v GPCVertex) Equal(other GPCVertex) bool {
	return math.Abs(v.X-other.X) <= Epsilon && math.Abs(v.Y-other.Y) <= Epsilon
}

// String returns the string representation of the vertex
func (v GPCVertex) String() string {
	return fmt.Sprintf("(%.6f, %.6f)", v.X, v.Y)
}

// GPCVertexList represents an array of vertices forming a contour
type GPCVertexList struct {
	NumVertices int
	Vertices    []GPCVertex
}

// NewGPCVertexList creates a new vertex list with the given capacity
func NewGPCVertexList(capacity int) *GPCVertexList {
	return &GPCVertexList{
		NumVertices: 0,
		Vertices:    make([]GPCVertex, 0, capacity),
	}
}

// AddVertex adds a vertex to the list
func (vl *GPCVertexList) AddVertex(x, y float64) {
	vl.Vertices = append(vl.Vertices, GPCVertex{X: x, Y: y})
	vl.NumVertices = len(vl.Vertices)
}

// GetVertex returns the vertex at the given index
func (vl *GPCVertexList) GetVertex(index int) (GPCVertex, error) {
	if index < 0 || index >= vl.NumVertices {
		return GPCVertex{}, fmt.Errorf("vertex index %d out of range [0, %d)", index, vl.NumVertices)
	}
	return vl.Vertices[index], nil
}

// Clear removes all vertices from the list
func (vl *GPCVertexList) Clear() {
	vl.Vertices = vl.Vertices[:0]
	vl.NumVertices = 0
}

// GPCPolygon represents a complete polygon with multiple contours
type GPCPolygon struct {
	NumContours int
	Hole        []bool           // Hole flags for each contour
	Contours    []*GPCVertexList // Array of contours
}

// NewGPCPolygon creates a new empty polygon
func NewGPCPolygon() *GPCPolygon {
	return &GPCPolygon{
		NumContours: 0,
		Hole:        make([]bool, 0),
		Contours:    make([]*GPCVertexList, 0),
	}
}

// AddContour adds a contour to the polygon
func (p *GPCPolygon) AddContour(contour *GPCVertexList, isHole bool) error {
	if contour == nil {
		return errors.New("contour cannot be nil")
	}
	if contour.NumVertices < 3 {
		return fmt.Errorf("contour must have at least 3 vertices, got %d", contour.NumVertices)
	}

	p.Contours = append(p.Contours, contour)
	p.Hole = append(p.Hole, isHole)
	p.NumContours = len(p.Contours)
	return nil
}

// GetContour returns the contour at the given index
func (p *GPCPolygon) GetContour(index int) (*GPCVertexList, bool, error) {
	if index < 0 || index >= p.NumContours {
		return nil, false, fmt.Errorf("contour index %d out of range [0, %d)", index, p.NumContours)
	}
	return p.Contours[index], p.Hole[index], nil
}

// Clear removes all contours from the polygon
func (p *GPCPolygon) Clear() {
	p.Contours = p.Contours[:0]
	p.Hole = p.Hole[:0]
	p.NumContours = 0
}

// IsEmpty returns true if the polygon has no contours
func (p *GPCPolygon) IsEmpty() bool {
	return p.NumContours == 0
}

// Validate checks if the polygon is valid
func (p *GPCPolygon) Validate() error {
	if len(p.Contours) != p.NumContours {
		return fmt.Errorf("contour count mismatch: expected %d, got %d", p.NumContours, len(p.Contours))
	}
	if len(p.Hole) != p.NumContours {
		return fmt.Errorf("hole flag count mismatch: expected %d, got %d", p.NumContours, len(p.Hole))
	}

	for i, contour := range p.Contours {
		if contour == nil {
			return fmt.Errorf("contour %d is nil", i)
		}
		if contour.NumVertices < 3 {
			return fmt.Errorf("contour %d has insufficient vertices: %d", i, contour.NumVertices)
		}
	}
	return nil
}

// GPCTristrip represents a triangle strip set
type GPCTristrip struct {
	NumStrips int
	Strips    []*GPCVertexList // Array of triangle strips
}

// NewGPCTristrip creates a new empty tristrip
func NewGPCTristrip() *GPCTristrip {
	return &GPCTristrip{
		NumStrips: 0,
		Strips:    make([]*GPCVertexList, 0),
	}
}

// AddStrip adds a strip to the tristrip
func (ts *GPCTristrip) AddStrip(strip *GPCVertexList) error {
	if strip == nil {
		return errors.New("strip cannot be nil")
	}
	if strip.NumVertices < 3 {
		return fmt.Errorf("strip must have at least 3 vertices, got %d", strip.NumVertices)
	}

	ts.Strips = append(ts.Strips, strip)
	ts.NumStrips = len(ts.Strips)
	return nil
}

// GetStrip returns the strip at the given index
func (ts *GPCTristrip) GetStrip(index int) (*GPCVertexList, error) {
	if index < 0 || index >= ts.NumStrips {
		return nil, fmt.Errorf("strip index %d out of range [0, %d)", index, ts.NumStrips)
	}
	return ts.Strips[index], nil
}

// Clear removes all strips from the tristrip
func (ts *GPCTristrip) Clear() {
	ts.Strips = ts.Strips[:0]
	ts.NumStrips = 0
}

// IsEmpty returns true if the tristrip has no strips
func (ts *GPCTristrip) IsEmpty() bool {
	return ts.NumStrips == 0
}

// ReadPolygon reads a polygon from an io.Reader
func ReadPolygon(reader io.Reader, readHoleFlags bool) (*GPCPolygon, error) {
	if reader == nil {
		return nil, errors.New("reader cannot be nil")
	}

	scanner := bufio.NewReader(reader)

	var numContours int
	if _, err := fmt.Fscan(scanner, &numContours); err != nil {
		return nil, fmt.Errorf("failed to read contour count: %w", err)
	}
	if numContours < 0 {
		return nil, fmt.Errorf("invalid contour count: %d", numContours)
	}

	polygon := NewGPCPolygon()
	for i := 0; i < numContours; i++ {
		isHole := false
		numVertices := 0

		if readHoleFlags {
			holeFlag := 0
			if _, err := fmt.Fscan(scanner, &holeFlag, &numVertices); err != nil {
				return nil, fmt.Errorf("failed to read contour %d header: %w", i, err)
			}
			if holeFlag != 0 && holeFlag != 1 {
				return nil, fmt.Errorf("invalid hole flag for contour %d: %d", i, holeFlag)
			}
			isHole = holeFlag == 1
		} else {
			if _, err := fmt.Fscan(scanner, &numVertices); err != nil {
				return nil, fmt.Errorf("failed to read contour %d vertex count: %w", i, err)
			}
		}

		if numVertices < 3 {
			return nil, fmt.Errorf("contour %d must have at least 3 vertices, got %d", i, numVertices)
		}

		contour := NewGPCVertexList(numVertices)
		for j := 0; j < numVertices; j++ {
			var x, y float64
			if _, err := fmt.Fscan(scanner, &x, &y); err != nil {
				return nil, fmt.Errorf("failed to read contour %d vertex %d: %w", i, j, err)
			}
			contour.AddVertex(x, y)
		}

		if err := polygon.AddContour(contour, isHole); err != nil {
			return nil, fmt.Errorf("failed to add contour %d: %w", i, err)
		}
	}

	validateErr := polygon.Validate()
	return polygon, validateErr
}

// WritePolygon writes a polygon to an io.Writer
func WritePolygon(writer io.Writer, polygon *GPCPolygon, writeHoleFlags bool) error {
	if polygon == nil {
		return errors.New("polygon cannot be nil")
	}

	// Simple text format: number of contours, then for each contour:
	// hole_flag num_vertices, then vertices as x,y pairs
	_, err := fmt.Fprintf(writer, "%d\n", polygon.NumContours)
	if err != nil {
		return err
	}

	for i := 0; i < polygon.NumContours; i++ {
		contour := polygon.Contours[i]
		isHole := polygon.Hole[i]

		if writeHoleFlags {
			holeFlag := 0
			if isHole {
				holeFlag = 1
			}
			_, err = fmt.Fprintf(writer, "%d %d\n", holeFlag, contour.NumVertices)
		} else {
			_, err = fmt.Fprintf(writer, "%d\n", contour.NumVertices)
		}
		if err != nil {
			return err
		}

		for j := 0; j < contour.NumVertices; j++ {
			v := contour.Vertices[j]
			_, err = fmt.Fprintf(writer, "%.6f %.6f\n", v.X, v.Y)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AddContour adds a contour to an existing polygon (standalone function)
func AddContour(polygon *GPCPolygon, contour *GPCVertexList, isHole bool) error {
	if polygon == nil {
		return errors.New("polygon cannot be nil")
	}
	return polygon.AddContour(contour, isHole)
}

// copyPolygon creates a deep copy of a polygon
func copyPolygon(src *GPCPolygon) *GPCPolygon {
	result := NewGPCPolygon()

	for i := 0; i < src.NumContours; i++ {
		contour, isHole, err := src.GetContour(i)
		if err == nil {
			newContour := NewGPCVertexList(contour.NumVertices)
			for j := 0; j < contour.NumVertices; j++ {
				vertex, err := contour.GetVertex(j)
				if err == nil {
					newContour.AddVertex(vertex.X, vertex.Y)
				}
			}
			// PolygonClip validates its inputs before copying, so every contour
			// has at least three vertices and AddContour cannot fail here.
			_ = result.AddContour(newContour, isHole)
		}
	}

	return result
}

// PolygonClip performs boolean clipping operations on two polygons
func PolygonClip(operation GPCOp, subjectPolygon, clipPolygon *GPCPolygon) (*GPCPolygon, error) {
	if subjectPolygon == nil || clipPolygon == nil {
		return nil, errors.New("input polygons cannot be nil")
	}

	if err := subjectPolygon.Validate(); err != nil {
		return nil, fmt.Errorf("subject polygon validation failed: %w", err)
	}

	if err := clipPolygon.Validate(); err != nil {
		return nil, fmt.Errorf("clip polygon validation failed: %w", err)
	}

	// Handle trivial cases
	if ((subjectPolygon.NumContours == 0) && (clipPolygon.NumContours == 0)) ||
		((subjectPolygon.NumContours == 0) && ((operation == GPCInt) || (operation == GPCDiff))) ||
		((clipPolygon.NumContours == 0) && (operation == GPCInt)) {
		return NewGPCPolygon(), nil
	}

	// Handle cases with one empty polygon
	if subjectPolygon.NumContours == 0 {
		switch operation {
		case GPCUnion, GPCXor:
			return copyPolygon(clipPolygon), nil
		case GPCInt, GPCDiff:
			return NewGPCPolygon(), nil
		default:
			return copyPolygon(clipPolygon), nil
		}
	}

	if clipPolygon.NumContours == 0 {
		switch operation {
		case GPCUnion, GPCXor, GPCDiff:
			return copyPolygon(subjectPolygon), nil
		case GPCInt:
			return NewGPCPolygon(), nil
		default:
			return NewGPCPolygon(), nil
		}
	}

	// Use the complete scanline algorithm
	return polygonClipComplete(operation, subjectPolygon, clipPolygon)
}

// polygonClipComplete implements the complete GPC scanline algorithm
func polygonClipComplete(operation GPCOp, subjectPolygon, clipPolygon *GPCPolygon) (*GPCPolygon, error) {
	// Identify potentially contributing contours using bounding box overlap
	if ((operation == GPCInt) || (operation == GPCDiff)) &&
		(subjectPolygon.NumContours > 0) && (clipPolygon.NumContours > 0) {
		minimaxTest(subjectPolygon, clipPolygon, operation)
	}

	// Build Local Minima Table
	var lmt *lmtNode
	var sbtree *sbTree
	sbtEntries := 0

	// Build LMT and store edge heaps (returned but not used - memory managed by GC)
	if subjectPolygon.NumContours > 0 {
		_ = buildLocalMinimaTable(&lmt, &sbtree, &sbtEntries, subjectPolygon, SUBJ, operation)
	}
	if clipPolygon.NumContours > 0 {
		_ = buildLocalMinimaTable(&lmt, &sbtree, &sbtEntries, clipPolygon, CLIP, operation)
	}

	// Return empty result if no contours contribute
	if lmt == nil {
		return NewGPCPolygon(), nil
	}

	// Build scanbeam table from scanbeam tree
	sbt := make([]float64, sbtEntries)
	scanbeam := 0
	buildScanBeamTable(&scanbeam, sbt, sbtree)
	scanbeam = 0

	// Initialize for scan-line algorithm
	var aet *edgeNode
	var outPoly *polygonNode
	var cf *polygonNode
	parity := [2]int{LEFT, LEFT}

	// Invert clip polygon for difference operation
	if operation == GPCDiff {
		parity[CLIP] = RIGHT
	}

	localMin := lmt

	// Process each scanbeam
	for scanbeam < sbtEntries {
		// Set yb and yt to the bottom and top of the scanbeam
		yb := sbt[scanbeam]
		scanbeam++
		var yt, dy float64
		if scanbeam < sbtEntries {
			yt = sbt[scanbeam]
			dy = yt - yb
		}

		// === SCANBEAM BOUNDARY PROCESSING ===========================

		// If LMT node corresponding to yb exists
		if localMin != nil && localMin.Y == yb {
			// Add edges starting at this local minimum to the AET
			for edge := localMin.FirstBound; edge != nil; edge = edge.NextBound {
				addEdgeToAET(&aet, edge, nil)
			}
			localMin = localMin.Next
		}

		// Set dummy previous x value
		px := -math.MaxFloat64
		_ = px // Used for vertex tracking

		// Create bundles within AET
		if aet != nil {
			e0 := aet
			// Set up bundle fields of first edge
			aet.Bundle[ABOVE][aet.Type] = 0
			if aet.Top.Y != yb {
				aet.Bundle[ABOVE][aet.Type] = 1
			}
			aet.Bundle[ABOVE][1-aet.Type] = 0
			aet.BState[ABOVE] = bsUnbundled

			// Process remaining edges
			for nextEdge := aet.Next; nextEdge != nil; nextEdge = nextEdge.Next {
				// Set up bundle fields of next edge
				nextEdge.Bundle[ABOVE][nextEdge.Type] = 0
				if nextEdge.Top.Y != yb {
					nextEdge.Bundle[ABOVE][nextEdge.Type] = 1
				}
				nextEdge.Bundle[ABOVE][1-nextEdge.Type] = 0
				nextEdge.BState[ABOVE] = bsUnbundled

				// Bundle edges above the scanbeam boundary if they coincide
				if nextEdge.Bundle[ABOVE][nextEdge.Type] != 0 {
					if eq(e0.XB, nextEdge.XB) && eq(e0.DX, nextEdge.DX) && (e0.Top.Y != yb) {
						nextEdge.Bundle[ABOVE][nextEdge.Type] ^= e0.Bundle[ABOVE][nextEdge.Type]
						nextEdge.Bundle[ABOVE][1-nextEdge.Type] = e0.Bundle[ABOVE][1-nextEdge.Type]
						nextEdge.BState[ABOVE] = bsBundleHead
						e0.Bundle[ABOVE][CLIP] = 0
						e0.Bundle[ABOVE][SUBJ] = 0
						e0.BState[ABOVE] = bsBundleTail
					}
					e0 = nextEdge
				}
			}
		}

		horiz := [2]hState{hsNH, hsNH}

		// Process each edge at this scanbeam boundary
		for edge := aet; edge != nil; edge = edge.Next {
			exists := [2]int{
				edge.Bundle[ABOVE][CLIP] + (edge.Bundle[BELOW][CLIP] << 1),
				edge.Bundle[ABOVE][SUBJ] + (edge.Bundle[BELOW][SUBJ] << 1),
			}

			if exists[CLIP] != 0 || exists[SUBJ] != 0 {
				// Set bundle side
				edge.BSide[CLIP] = parity[CLIP]
				edge.BSide[SUBJ] = parity[SUBJ]

				// Determine contributing status and quadrant occupancies
				var contributing bool
				var br, bl, tr, tl int

				switch operation {
				case GPCDiff, GPCInt:
					contributing = (exists[CLIP] != 0 && (parity[SUBJ] != 0 || horiz[SUBJ] != hsNH)) ||
						(exists[SUBJ] != 0 && (parity[CLIP] != 0 || horiz[CLIP] != hsNH)) ||
						(exists[CLIP] != 0 && exists[SUBJ] != 0 && (parity[CLIP] == parity[SUBJ]))
					br = parity[CLIP] & parity[SUBJ]
					bl = (parity[CLIP] ^ edge.Bundle[ABOVE][CLIP]) & (parity[SUBJ] ^ edge.Bundle[ABOVE][SUBJ])
					tr = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH)) & (parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH))
					tl = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH) ^ edge.Bundle[BELOW][CLIP]) &
						(parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH) ^ edge.Bundle[BELOW][SUBJ])
				case GPCXor:
					contributing = exists[CLIP] != 0 || exists[SUBJ] != 0
					br = parity[CLIP] ^ parity[SUBJ]
					bl = (parity[CLIP] ^ edge.Bundle[ABOVE][CLIP]) ^ (parity[SUBJ] ^ edge.Bundle[ABOVE][SUBJ])
					tr = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH)) ^ (parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH))
					tl = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH) ^ edge.Bundle[BELOW][CLIP]) ^
						(parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH) ^ edge.Bundle[BELOW][SUBJ])
				case GPCUnion:
					contributing = (exists[CLIP] != 0 && (parity[SUBJ] == 0 || horiz[SUBJ] != hsNH)) ||
						(exists[SUBJ] != 0 && (parity[CLIP] == 0 || horiz[CLIP] != hsNH)) ||
						(exists[CLIP] != 0 && exists[SUBJ] != 0 && (parity[CLIP] == parity[SUBJ]))
					br = parity[CLIP] | parity[SUBJ]
					bl = (parity[CLIP] ^ edge.Bundle[ABOVE][CLIP]) | (parity[SUBJ] ^ edge.Bundle[ABOVE][SUBJ])
					tr = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH)) | (parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH))
					tl = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH) ^ edge.Bundle[BELOW][CLIP]) |
						(parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH) ^ edge.Bundle[BELOW][SUBJ])
				}

				// Update parity
				parity[CLIP] ^= edge.Bundle[ABOVE][CLIP]
				parity[SUBJ] ^= edge.Bundle[ABOVE][SUBJ]

				// Update horizontal state
				if exists[CLIP] != 0 {
					horiz[CLIP] = nextHState[horiz[CLIP]][((exists[CLIP]-1)<<1)+parity[CLIP]]
				}
				if exists[SUBJ] != 0 {
					horiz[SUBJ] = nextHState[horiz[SUBJ]][((exists[SUBJ]-1)<<1)+parity[SUBJ]]
				}

				vclass := tr + (tl << 1) + (br << 2) + (bl << 3)

				if contributing {
					xb := edge.XB

					switch vclass {
					case int(vtxEMN), int(vtxIMN):
						addLocalMin(&outPoly, edge, xb, yb)
						px = xb
						cf = edge.OutP[ABOVE]
					case int(vtxERI):
						if xb != px {
							addRight(cf, xb, yb)
							px = xb
						}
						edge.OutP[ABOVE] = cf
						cf = nil
					case int(vtxELI):
						addLeft(edge.OutP[BELOW], xb, yb)
						px = xb
						cf = edge.OutP[BELOW]
					case int(vtxEMX):
						if xb != px {
							addLeft(cf, xb, yb)
							px = xb
						}
						mergeRight(cf, edge.OutP[BELOW], outPoly)
						cf = nil
					case int(vtxILI):
						if xb != px {
							addLeft(cf, xb, yb)
							px = xb
						}
						edge.OutP[ABOVE] = cf
						cf = nil
					case int(vtxIRI):
						addRight(edge.OutP[BELOW], xb, yb)
						px = xb
						cf = edge.OutP[BELOW]
						edge.OutP[BELOW] = nil
					case int(vtxIMX):
						if xb != px {
							addRight(cf, xb, yb)
							px = xb
						}
						mergeLeft(cf, edge.OutP[BELOW], outPoly)
						cf = nil
						edge.OutP[BELOW] = nil
					case int(vtxIMM):
						if xb != px {
							addRight(cf, xb, yb)
							px = xb
						}
						mergeLeft(cf, edge.OutP[BELOW], outPoly)
						edge.OutP[BELOW] = nil
						addLocalMin(&outPoly, edge, xb, yb)
						cf = edge.OutP[ABOVE]
					case int(vtxEMM):
						if xb != px {
							addLeft(cf, xb, yb)
							px = xb
						}
						mergeRight(cf, edge.OutP[BELOW], outPoly)
						edge.OutP[BELOW] = nil
						addLocalMin(&outPoly, edge, xb, yb)
						cf = edge.OutP[ABOVE]
					case int(vtxLED):
						if edge.Bot.Y == yb && edge.OutP[BELOW] != nil {
							addLeft(edge.OutP[BELOW], xb, yb)
						}
						edge.OutP[ABOVE] = edge.OutP[BELOW]
						px = xb
					case int(vtxRED):
						if edge.Bot.Y == yb && edge.OutP[BELOW] != nil {
							addRight(edge.OutP[BELOW], xb, yb)
						}
						edge.OutP[ABOVE] = edge.OutP[BELOW]
						px = xb
					}
				}
			}
		}

		// Delete terminating edges from the AET, otherwise compute xt
		var prevEdge *edgeNode
		edge := aet
		for edge != nil {
			if edge.Top.Y == yb {
				nextEdge := edge.Next
				if prevEdge != nil {
					prevEdge.Next = nextEdge
				} else {
					aet = nextEdge
				}
				if nextEdge != nil {
					nextEdge.Prev = prevEdge
				}

				// Copy bundle head state to the adjacent tail edge if required
				if (edge.BState[BELOW] == bsBundleHead) && prevEdge != nil {
					if prevEdge.BState[BELOW] == bsBundleTail {
						prevEdge.OutP[BELOW] = edge.OutP[BELOW]
						prevEdge.BState[BELOW] = bsUnbundled
						if prevEdge.Prev != nil && prevEdge.Prev.BState[BELOW] == bsBundleTail {
							prevEdge.BState[BELOW] = bsBundleHead
						}
					}
				}
				edge = nextEdge
			} else {
				if edge.Top.Y == yt {
					edge.XT = edge.Top.X
				} else {
					edge.XT = edge.Bot.X + edge.DX*(yt-edge.Bot.Y)
				}
				prevEdge = edge
				edge = edge.Next
			}
		}

		if scanbeam < sbtEntries {
			// === SCANBEAM INTERIOR PROCESSING ===========================
			var it *itNode
			buildIntersectionTable(&it, aet, dy)

			// Process each node in the intersection table
			for intersect := it; intersect != nil; intersect = intersect.Next {
				e0 := intersect.IE[0]
				e1 := intersect.IE[1]

				// Only generate output for contributing intersections
				if (e0.Bundle[ABOVE][CLIP] != 0 || e0.Bundle[ABOVE][SUBJ] != 0) &&
					(e1.Bundle[ABOVE][CLIP] != 0 || e1.Bundle[ABOVE][SUBJ] != 0) {

					p := e0.OutP[ABOVE]
					q := e1.OutP[ABOVE]
					ix := intersect.Point.X
					iy := intersect.Point.Y + yb

					in := [2]int{
						boolToInt((e0.Bundle[ABOVE][CLIP] != 0 && e0.BSide[CLIP] == 0) ||
							(e1.Bundle[ABOVE][CLIP] != 0 && e1.BSide[CLIP] != 0) ||
							(e0.Bundle[ABOVE][CLIP] == 0 && e1.Bundle[ABOVE][CLIP] == 0 &&
								e0.BSide[CLIP] != 0 && e1.BSide[CLIP] != 0)),
						boolToInt((e0.Bundle[ABOVE][SUBJ] != 0 && e0.BSide[SUBJ] == 0) ||
							(e1.Bundle[ABOVE][SUBJ] != 0 && e1.BSide[SUBJ] != 0) ||
							(e0.Bundle[ABOVE][SUBJ] == 0 && e1.Bundle[ABOVE][SUBJ] == 0 &&
								e0.BSide[SUBJ] != 0 && e1.BSide[SUBJ] != 0)),
					}

					// Determine quadrant occupancies
					var tr, tl, br, bl int
					switch operation {
					case GPCDiff, GPCInt:
						tr = in[CLIP] & in[SUBJ]
						tl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP]) & (in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ])
						br = (in[CLIP] ^ e0.Bundle[ABOVE][CLIP]) & (in[SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
						bl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP] ^ e0.Bundle[ABOVE][CLIP]) &
							(in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
					case GPCXor:
						tr = in[CLIP] ^ in[SUBJ]
						tl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP]) ^ (in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ])
						br = (in[CLIP] ^ e0.Bundle[ABOVE][CLIP]) ^ (in[SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
						bl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP] ^ e0.Bundle[ABOVE][CLIP]) ^
							(in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
					case GPCUnion:
						tr = in[CLIP] | in[SUBJ]
						tl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP]) | (in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ])
						br = (in[CLIP] ^ e0.Bundle[ABOVE][CLIP]) | (in[SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
						bl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP] ^ e0.Bundle[ABOVE][CLIP]) |
							(in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
					}

					vclass := tr + (tl << 1) + (br << 2) + (bl << 3)

					switch vclass {
					case int(vtxEMN):
						addLocalMin(&outPoly, e0, ix, iy)
						e1.OutP[ABOVE] = e0.OutP[ABOVE]
					case int(vtxERI):
						if p != nil {
							addRight(p, ix, iy)
							e1.OutP[ABOVE] = p
							e0.OutP[ABOVE] = nil
						}
					case int(vtxELI):
						if q != nil {
							addLeft(q, ix, iy)
							e0.OutP[ABOVE] = q
							e1.OutP[ABOVE] = nil
						}
					case int(vtxEMX):
						if p != nil && q != nil {
							addLeft(p, ix, iy)
							mergeRight(p, q, outPoly)
							e0.OutP[ABOVE] = nil
							e1.OutP[ABOVE] = nil
						}
					case int(vtxIMN):
						addLocalMin(&outPoly, e0, ix, iy)
						e1.OutP[ABOVE] = e0.OutP[ABOVE]
					case int(vtxILI):
						if p != nil {
							addLeft(p, ix, iy)
							e1.OutP[ABOVE] = p
							e0.OutP[ABOVE] = nil
						}
					case int(vtxIRI):
						if q != nil {
							addRight(q, ix, iy)
							e0.OutP[ABOVE] = q
							e1.OutP[ABOVE] = nil
						}
					case int(vtxIMX):
						if p != nil && q != nil {
							addRight(p, ix, iy)
							mergeLeft(p, q, outPoly)
							e0.OutP[ABOVE] = nil
							e1.OutP[ABOVE] = nil
						}
					case int(vtxIMM):
						if p != nil && q != nil {
							addRight(p, ix, iy)
							mergeLeft(p, q, outPoly)
							addLocalMin(&outPoly, e0, ix, iy)
							e1.OutP[ABOVE] = e0.OutP[ABOVE]
						}
					case int(vtxEMM):
						if p != nil && q != nil {
							addLeft(p, ix, iy)
							mergeRight(p, q, outPoly)
							addLocalMin(&outPoly, e0, ix, iy)
							e1.OutP[ABOVE] = e0.OutP[ABOVE]
						}
					}
				}

				// Swap bundle sides in response to edge crossing
				if e0 != nil && e1 != nil {
					if e0.Bundle[ABOVE][CLIP] != 0 {
						e1.BSide[CLIP] = 1 - e1.BSide[CLIP]
					}
					if e1.Bundle[ABOVE][CLIP] != 0 {
						e0.BSide[CLIP] = 1 - e0.BSide[CLIP]
					}
					if e0.Bundle[ABOVE][SUBJ] != 0 {
						e1.BSide[SUBJ] = 1 - e1.BSide[SUBJ]
					}
					if e1.Bundle[ABOVE][SUBJ] != 0 {
						e0.BSide[SUBJ] = 1 - e0.BSide[SUBJ]
					}

					// Swap e0 and e1 in the AET.
					prevEdge := e0.Prev
					nextEdge := e1.Next
					if nextEdge != nil {
						nextEdge.Prev = e0
					}

					if e0.BState[ABOVE] == bsBundleHead {
						search := true
						for search {
							if prevEdge != nil {
								prevEdge = prevEdge.Prev
								if prevEdge == nil || prevEdge.BState[ABOVE] != bsBundleTail {
									search = false
								}
							} else {
								search = false
							}
						}
					}

					if prevEdge == nil {
						aet.Prev = e1
						e1.Next = aet
						aet = e0.Next
					} else {
						prevEdge.Next.Prev = e1
						e1.Next = prevEdge.Next
						prevEdge.Next = e0.Next
					}
					e0.Next.Prev = prevEdge
					e1.Next.Prev = e1
					e0.Next = nextEdge
				}
			}

			// Prepare for next scanbeam
			for edge := aet; edge != nil; {
				nextEdge := edge.Next
				succEdge := edge.Succ

				if edge.Top.Y == yt && succEdge != nil {
					// Replace AET edge by its successor.
					succEdge.OutP[BELOW] = edge.OutP[ABOVE]
					succEdge.BState[BELOW] = edge.BState[ABOVE]
					succEdge.Bundle[BELOW][CLIP] = edge.Bundle[ABOVE][CLIP]
					succEdge.Bundle[BELOW][SUBJ] = edge.Bundle[ABOVE][SUBJ]

					prevEdge := edge.Prev
					if prevEdge != nil {
						prevEdge.Next = succEdge
					} else {
						aet = succEdge
					}
					if nextEdge != nil {
						nextEdge.Prev = succEdge
					}
					succEdge.Prev = prevEdge
					succEdge.Next = nextEdge
				} else {
					// Update this edge.
					edge.OutP[BELOW] = edge.OutP[ABOVE]
					edge.BState[BELOW] = edge.BState[ABOVE]
					edge.Bundle[BELOW][CLIP] = edge.Bundle[ABOVE][CLIP]
					edge.Bundle[BELOW][SUBJ] = edge.Bundle[ABOVE][SUBJ]
					edge.XB = edge.XT
				}

				edge.OutP[ABOVE] = nil
				edge = nextEdge
			}
		}
	}

	// Convert output polygons to result format
	return convertPolygonNodesToGPCPolygon(outPoly), nil
}

// Helper function to convert bool to int for vertex classification
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// convertPolygonNodesToGPCPolygon converts internal polygon nodes to GPCPolygon
func convertPolygonNodesToGPCPolygon(outPoly *polygonNode) *GPCPolygon {
	result := NewGPCPolygon()

	if outPoly == nil {
		return result
	}

	// Count valid contours first
	numContours := countContours(outPoly)
	if numContours == 0 {
		return result
	}

	// Process each polygon node
	for poly := outPoly; poly != nil; poly = poly.Next {
		if poly.Active > 2 {
			// Create vertex list from the internal vertex chain
			contour := NewGPCVertexList(poly.Active)

			// Traverse the vertex chain
			for v := poly.Proxy.V[LEFT]; v != nil; v = v.Next {
				contour.AddVertex(v.X, v.Y)
			}

			// Add the contour to the result polygon
			err := result.AddContour(contour, poly.Proxy.Hole)
			if err != nil {
				// Continue with other contours even if one fails
				continue
			}
		}
	}

	return result
}

// convertTriStripNodesToGPCTristrip converts internal triangle strip nodes to GPCTristrip
func convertTriStripNodesToGPCTristrip(tlist *polygonNode) *GPCTristrip {
	result := NewGPCTristrip()

	if tlist == nil {
		return result
	}

	// Count valid strips first
	numStrips := countTristrips(tlist)
	if numStrips == 0 {
		return result
	}

	// Process each triangle strip node
	for strip := tlist; strip != nil; strip = strip.Next {
		if strip.Active > 2 {
			// Create vertex list from the internal vertex chain
			stripVertices := NewGPCVertexList(strip.Active)

			// Traverse the vertex chain
			for v := strip.V[LEFT]; v != nil; v = v.Next {
				stripVertices.AddVertex(v.X, v.Y)
			}

			// Add the strip to the result tristrip
			err := result.AddStrip(stripVertices)
			if err != nil {
				// Continue with other strips even if one fails
				continue
			}
		}
	}

	return result
}

// TristripClip performs boolean clipping operations and returns triangle strips
func TristripClip(operation GPCOp, subjectPolygon, clipPolygon *GPCPolygon) (*GPCTristrip, error) {
	if subjectPolygon == nil || clipPolygon == nil {
		return nil, errors.New("input polygons cannot be nil")
	}

	if err := subjectPolygon.Validate(); err != nil {
		return nil, fmt.Errorf("subject polygon validation failed: %w", err)
	}

	if err := clipPolygon.Validate(); err != nil {
		return nil, fmt.Errorf("clip polygon validation failed: %w", err)
	}

	// Handle trivial cases
	if ((subjectPolygon.NumContours == 0) && (clipPolygon.NumContours == 0)) ||
		((subjectPolygon.NumContours == 0) && ((operation == GPCInt) || (operation == GPCDiff))) ||
		((clipPolygon.NumContours == 0) && (operation == GPCInt)) {
		return NewGPCTristrip(), nil
	}

	// Identify potentially contributing contours using bounding box overlap
	if ((operation == GPCInt) || (operation == GPCDiff)) &&
		(subjectPolygon.NumContours > 0) && (clipPolygon.NumContours > 0) {
		minimaxTest(subjectPolygon, clipPolygon, operation)
	}

	// Build Local Minima Table
	var lmt *lmtNode
	var sbtree *sbTree
	sbtEntries := 0

	// Build LMT and store edge heaps (returned but not used - memory managed by GC)
	if subjectPolygon.NumContours > 0 {
		_ = buildLocalMinimaTable(&lmt, &sbtree, &sbtEntries, subjectPolygon, SUBJ, operation)
	}
	if clipPolygon.NumContours > 0 {
		_ = buildLocalMinimaTable(&lmt, &sbtree, &sbtEntries, clipPolygon, CLIP, operation)
	}

	// Return empty result if no contours contribute
	if lmt == nil {
		return NewGPCTristrip(), nil
	}

	// Build scanbeam table from scanbeam tree
	sbt := make([]float64, sbtEntries)
	scanbeam := 0
	buildScanBeamTable(&scanbeam, sbt, sbtree)
	scanbeam = 0

	// Initialize for scan-line algorithm
	var aet *edgeNode
	var tlist *polygonNode // Triangle strip list
	parity := [2]int{LEFT, LEFT}

	// Invert clip polygon for difference operation
	if operation == GPCDiff {
		parity[CLIP] = RIGHT
	}

	localMin := lmt

	// Process each scanbeam
	for scanbeam < sbtEntries {
		// Set yb and yt to the bottom and top of the scanbeam
		yb := sbt[scanbeam]
		scanbeam++
		var yt, dy float64
		if scanbeam < sbtEntries {
			yt = sbt[scanbeam]
			dy = yt - yb
		}

		// === SCANBEAM BOUNDARY PROCESSING ===========================

		// If LMT node corresponding to yb exists
		if localMin != nil && localMin.Y == yb {
			// Add edges starting at this local minimum to the AET
			for edge := localMin.FirstBound; edge != nil; edge = edge.NextBound {
				addEdgeToAET(&aet, edge, nil)
			}
			localMin = localMin.Next
		}

		// Set dummy previous x value
		px := -math.MaxFloat64
		_ = px // Used for vertex tracking

		// Create bundles within AET
		if aet != nil {
			e0 := aet
			// Set up bundle fields of first edge
			aet.Bundle[ABOVE][aet.Type] = 0
			if aet.Top.Y != yb {
				aet.Bundle[ABOVE][aet.Type] = 1
			}
			aet.Bundle[ABOVE][1-aet.Type] = 0
			aet.BState[ABOVE] = bsUnbundled

			// Process remaining edges
			for nextEdge := aet.Next; nextEdge != nil; nextEdge = nextEdge.Next {
				// Set up bundle fields of next edge
				nextEdge.Bundle[ABOVE][nextEdge.Type] = 0
				if nextEdge.Top.Y != yb {
					nextEdge.Bundle[ABOVE][nextEdge.Type] = 1
				}
				nextEdge.Bundle[ABOVE][1-nextEdge.Type] = 0
				nextEdge.BState[ABOVE] = bsUnbundled

				// Bundle edges above the scanbeam boundary if they coincide
				if nextEdge.Bundle[ABOVE][nextEdge.Type] != 0 {
					if eq(e0.XB, nextEdge.XB) && eq(e0.DX, nextEdge.DX) && (e0.Top.Y != yb) {
						nextEdge.Bundle[ABOVE][nextEdge.Type] ^= e0.Bundle[ABOVE][nextEdge.Type]
						nextEdge.Bundle[ABOVE][1-nextEdge.Type] = e0.Bundle[ABOVE][1-nextEdge.Type]
						nextEdge.BState[ABOVE] = bsBundleHead
						e0.Bundle[ABOVE][CLIP] = 0
						e0.Bundle[ABOVE][SUBJ] = 0
						e0.BState[ABOVE] = bsBundleTail
					}
					e0 = nextEdge
				}
			}
		}

		horiz := [2]hState{hsNH, hsNH}

		// Process each edge at this scanbeam boundary
		for edge := aet; edge != nil; edge = edge.Next {
			exists := [2]int{
				edge.Bundle[ABOVE][CLIP] + (edge.Bundle[BELOW][CLIP] << 1),
				edge.Bundle[ABOVE][SUBJ] + (edge.Bundle[BELOW][SUBJ] << 1),
			}

			if exists[CLIP] != 0 || exists[SUBJ] != 0 {
				// Set bundle side
				edge.BSide[CLIP] = parity[CLIP]
				edge.BSide[SUBJ] = parity[SUBJ]

				// Determine contributing status and quadrant occupancies
				var contributing bool
				var br, bl, tr, tl int

				switch operation {
				case GPCDiff, GPCInt:
					contributing = (exists[CLIP] != 0 && (parity[SUBJ] != 0 || horiz[SUBJ] != hsNH)) ||
						(exists[SUBJ] != 0 && (parity[CLIP] != 0 || horiz[CLIP] != hsNH)) ||
						(exists[CLIP] != 0 && exists[SUBJ] != 0 && (parity[CLIP] == parity[SUBJ]))
					br = parity[CLIP] & parity[SUBJ]
					bl = (parity[CLIP] ^ edge.Bundle[ABOVE][CLIP]) & (parity[SUBJ] ^ edge.Bundle[ABOVE][SUBJ])
					tr = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH)) & (parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH))
					tl = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH) ^ edge.Bundle[BELOW][CLIP]) &
						(parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH) ^ edge.Bundle[BELOW][SUBJ])
				case GPCXor:
					contributing = exists[CLIP] != 0 || exists[SUBJ] != 0
					br = parity[CLIP] ^ parity[SUBJ]
					bl = (parity[CLIP] ^ edge.Bundle[ABOVE][CLIP]) ^ (parity[SUBJ] ^ edge.Bundle[ABOVE][SUBJ])
					tr = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH)) ^ (parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH))
					tl = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH) ^ edge.Bundle[BELOW][CLIP]) ^
						(parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH) ^ edge.Bundle[BELOW][SUBJ])
				case GPCUnion:
					contributing = (exists[CLIP] != 0 && (parity[SUBJ] == 0 || horiz[SUBJ] != hsNH)) ||
						(exists[SUBJ] != 0 && (parity[CLIP] == 0 || horiz[CLIP] != hsNH)) ||
						(exists[CLIP] != 0 && exists[SUBJ] != 0 && (parity[CLIP] == parity[SUBJ]))
					br = parity[CLIP] | parity[SUBJ]
					bl = (parity[CLIP] ^ edge.Bundle[ABOVE][CLIP]) | (parity[SUBJ] ^ edge.Bundle[ABOVE][SUBJ])
					tr = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH)) | (parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH))
					tl = (parity[CLIP] ^ boolToInt(horiz[CLIP] != hsNH) ^ edge.Bundle[BELOW][CLIP]) |
						(parity[SUBJ] ^ boolToInt(horiz[SUBJ] != hsNH) ^ edge.Bundle[BELOW][SUBJ])
				}

				// Update parity
				parity[CLIP] ^= edge.Bundle[ABOVE][CLIP]
				parity[SUBJ] ^= edge.Bundle[ABOVE][SUBJ]

				// Update horizontal state
				if exists[CLIP] != 0 {
					horiz[CLIP] = nextHState[horiz[CLIP]][((exists[CLIP]-1)<<1)+parity[CLIP]]
				}
				if exists[SUBJ] != 0 {
					horiz[SUBJ] = nextHState[horiz[SUBJ]][((exists[SUBJ]-1)<<1)+parity[SUBJ]]
				}

				vclass := tr + (tl << 1) + (br << 2) + (bl << 3)

				if contributing {
					xb := edge.XB

					switch vclass {
					case int(vtxEMN):
						newTristrip(&tlist, edge, xb, yb)
					case int(vtxERI):
						// Add vertex to right side of strip
						if edge.OutP[ABOVE] != nil {
							addVertexToTristrip(&edge.OutP[ABOVE].V[RIGHT], xb, yb)
							edge.OutP[ABOVE].Active++
						}
					case int(vtxELI):
						// Add vertex to left side of strip
						if edge.OutP[BELOW] != nil {
							addVertexToTristrip(&edge.OutP[BELOW].V[LEFT], xb, yb)
							edge.OutP[BELOW].Active++
							edge.OutP[ABOVE] = edge.OutP[BELOW]
						}
					case int(vtxEMX):
						// Terminate strip
						if edge.OutP[BELOW] != nil {
							addVertexToTristrip(&edge.OutP[BELOW].V[RIGHT], xb, yb)
							edge.OutP[BELOW].Active++
							edge.OutP[ABOVE] = nil
						}
					case int(vtxIMN):
						newTristrip(&tlist, edge, xb, yb)
					case int(vtxILI):
						newTristrip(&tlist, edge, xb, yb)
					case int(vtxIRI):
						// Add vertex to right side
						if edge.OutP[BELOW] != nil {
							addVertexToTristrip(&edge.OutP[BELOW].V[RIGHT], xb, yb)
							edge.OutP[BELOW].Active++
							edge.OutP[ABOVE] = nil
						}
					case int(vtxIMX):
						// Add vertex to left side
						if edge.OutP[ABOVE] != nil {
							addVertexToTristrip(&edge.OutP[ABOVE].V[LEFT], xb, yb)
							edge.OutP[ABOVE].Active++
							edge.OutP[ABOVE] = nil
						}
					}
				}
			}
		}

		// Delete terminating edges from the AET, otherwise compute xt
		var prevEdge *edgeNode
		edge := aet
		for edge != nil {
			if edge.Top.Y == yb {
				nextEdge := edge.Next
				if prevEdge != nil {
					prevEdge.Next = nextEdge
				} else {
					aet = nextEdge
				}
				if nextEdge != nil {
					nextEdge.Prev = prevEdge
				}

				// Copy bundle head state to the adjacent tail edge if required
				if (edge.BState[BELOW] == bsBundleHead) && prevEdge != nil {
					if prevEdge.BState[BELOW] == bsBundleTail {
						prevEdge.OutP[BELOW] = edge.OutP[BELOW]
						prevEdge.BState[BELOW] = bsUnbundled
						if prevEdge.Prev != nil && prevEdge.Prev.BState[BELOW] == bsBundleTail {
							prevEdge.BState[BELOW] = bsBundleHead
						}
					}
				}
				edge = nextEdge
			} else {
				if edge.Top.Y == yt {
					edge.XT = edge.Top.X
				} else {
					edge.XT = edge.Bot.X + edge.DX*(yt-edge.Bot.Y)
				}
				prevEdge = edge
				edge = edge.Next
			}
		}

		if scanbeam < sbtEntries {
			// === SCANBEAM INTERIOR PROCESSING ===========================
			var it *itNode
			buildIntersectionTable(&it, aet, dy)

			// Process each node in the intersection table
			for intersect := it; intersect != nil; intersect = intersect.Next {
				e0 := intersect.IE[0]
				e1 := intersect.IE[1]

				// Only generate output for contributing intersections
				if (e0.Bundle[ABOVE][CLIP] != 0 || e0.Bundle[ABOVE][SUBJ] != 0) &&
					(e1.Bundle[ABOVE][CLIP] != 0 || e1.Bundle[ABOVE][SUBJ] != 0) {

					p := e0.OutP[ABOVE]
					q := e1.OutP[ABOVE]
					ix := intersect.Point.X
					iy := intersect.Point.Y + yb

					in := [2]int{
						boolToInt((e0.Bundle[ABOVE][CLIP] != 0 && e0.BSide[CLIP] == 0) ||
							(e1.Bundle[ABOVE][CLIP] != 0 && e1.BSide[CLIP] != 0) ||
							(e0.Bundle[ABOVE][CLIP] == 0 && e1.Bundle[ABOVE][CLIP] == 0 &&
								e0.BSide[CLIP] != 0 && e1.BSide[CLIP] != 0)),
						boolToInt((e0.Bundle[ABOVE][SUBJ] != 0 && e0.BSide[SUBJ] == 0) ||
							(e1.Bundle[ABOVE][SUBJ] != 0 && e1.BSide[SUBJ] != 0) ||
							(e0.Bundle[ABOVE][SUBJ] == 0 && e1.Bundle[ABOVE][SUBJ] == 0 &&
								e0.BSide[SUBJ] != 0 && e1.BSide[SUBJ] != 0)),
					}

					// Determine quadrant occupancies
					var tr, tl, br, bl int
					switch operation {
					case GPCDiff, GPCInt:
						tr = in[CLIP] & in[SUBJ]
						tl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP]) & (in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ])
						br = (in[CLIP] ^ e0.Bundle[ABOVE][CLIP]) & (in[SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
						bl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP] ^ e0.Bundle[ABOVE][CLIP]) &
							(in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
					case GPCXor:
						tr = in[CLIP] ^ in[SUBJ]
						tl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP]) ^ (in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ])
						br = (in[CLIP] ^ e0.Bundle[ABOVE][CLIP]) ^ (in[SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
						bl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP] ^ e0.Bundle[ABOVE][CLIP]) ^
							(in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
					case GPCUnion:
						tr = in[CLIP] | in[SUBJ]
						tl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP]) | (in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ])
						br = (in[CLIP] ^ e0.Bundle[ABOVE][CLIP]) | (in[SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
						bl = (in[CLIP] ^ e1.Bundle[ABOVE][CLIP] ^ e0.Bundle[ABOVE][CLIP]) |
							(in[SUBJ] ^ e1.Bundle[ABOVE][SUBJ] ^ e0.Bundle[ABOVE][SUBJ])
					}

					vclass := tr + (tl << 1) + (br << 2) + (bl << 3)

					switch vclass {
					case int(vtxEMN):
						newTristrip(&tlist, e0, ix, iy)
						e1.OutP[ABOVE] = e0.OutP[ABOVE]
					case int(vtxERI):
						if p != nil {
							addVertexToTristrip(&p.V[RIGHT], ix, iy)
							p.Active++
							e1.OutP[ABOVE] = p
							e0.OutP[ABOVE] = nil
						}
					case int(vtxELI):
						if q != nil {
							addVertexToTristrip(&q.V[LEFT], ix, iy)
							q.Active++
							e0.OutP[ABOVE] = q
							e1.OutP[ABOVE] = nil
						}
					}
				}

				// Swap edge bundles and x coordinates
				if e0 != nil && e1 != nil {
					e0.OutP[ABOVE], e1.OutP[ABOVE] = e1.OutP[ABOVE], e0.OutP[ABOVE]
					e0.Bundle[ABOVE][CLIP], e1.Bundle[ABOVE][CLIP] = e1.Bundle[ABOVE][CLIP], e0.Bundle[ABOVE][CLIP]
					e0.Bundle[BELOW][CLIP], e1.Bundle[BELOW][CLIP] = e1.Bundle[BELOW][CLIP], e0.Bundle[BELOW][CLIP]
					e0.Bundle[ABOVE][SUBJ], e1.Bundle[ABOVE][SUBJ] = e1.Bundle[ABOVE][SUBJ], e0.Bundle[ABOVE][SUBJ]
					e0.Bundle[BELOW][SUBJ], e1.Bundle[BELOW][SUBJ] = e1.Bundle[BELOW][SUBJ], e0.Bundle[BELOW][SUBJ]
					e0.BSide[CLIP], e1.BSide[CLIP] = e1.BSide[CLIP], e0.BSide[CLIP]
					e0.BSide[SUBJ], e1.BSide[SUBJ] = e1.BSide[SUBJ], e0.BSide[SUBJ]
					e0.BState[ABOVE], e1.BState[ABOVE] = e1.BState[ABOVE], e0.BState[ABOVE]
					e0.BState[BELOW], e1.BState[BELOW] = e1.BState[BELOW], e0.BState[BELOW]
					e0.XB, e1.XB = e1.XB, e0.XB
					e0.XT, e1.XT = e1.XT, e0.XT
				}
			}
		}

		// Copy bundle below to bundle above for next scanbeam
		for edge := aet; edge != nil; edge = edge.Next {
			edge.Bundle[ABOVE][CLIP] = edge.Bundle[BELOW][CLIP]
			edge.Bundle[ABOVE][SUBJ] = edge.Bundle[BELOW][SUBJ]
			edge.BState[ABOVE] = edge.BState[BELOW]
		}
	}

	// Convert triangle strip nodes to result format
	return convertTriStripNodesToGPCTristrip(tlist), nil
}

// PolygonToTristrip converts a polygon to triangle strips
func PolygonToTristrip(polygon *GPCPolygon) (*GPCTristrip, error) {
	if polygon == nil {
		return nil, errors.New("polygon cannot be nil")
	}

	if err := polygon.Validate(); err != nil {
		return nil, fmt.Errorf("polygon validation failed: %w", err)
	}

	// Create an empty clipping polygon (as per the C++ implementation)
	emptyClip := NewGPCPolygon()

	// Use TristripClip with GPC_DIFF operation against empty polygon
	// This effectively converts the subject polygon to triangle strips
	return TristripClip(GPCDiff, polygon, emptyClip)
}

// Helper functions for floating-point comparisons
func eq(a, b float64) bool {
	return math.Abs(a-b) <= Epsilon
}

// Helper function to compute polygon orientation
func isClockwise(vertices []GPCVertex) bool {
	if len(vertices) < 3 {
		return false
	}

	// Calculate signed area using the shoelace formula
	// For clockwise orientation, the signed area should be negative
	area := 0.0
	n := len(vertices)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += (vertices[i].X * vertices[j].Y) - (vertices[j].X * vertices[i].Y)
	}

	// Clockwise orientation has negative area
	return area < 0
}

// Helper function to validate contour winding
func validateContourWinding(contour *GPCVertexList, shouldBeClockwise bool) bool {
	if contour == nil || contour.NumVertices < 3 {
		return false
	}

	return isClockwise(contour.Vertices) == shouldBeClockwise
}
