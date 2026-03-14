// Package molview ports AGG's mol_view.cpp demo.
package molview

import (
	"bufio"
	"bytes"
	_ "embed"
	"math"
	"strconv"
	"strings"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	BaseWidth    = 400.0
	BaseHeight   = 400.0
	DefaultCount = 100
)

const (
	atomGeneral = iota
	atomN
	atomO
	atomS
	atomP
	atomHalogen
	atomColors
)

type Atom struct {
	X, Y     float64
	Label    string
	Charge   int
	ColorIdx int
}

type Bond struct {
	Idx1, Idx2 uint
	X1, Y1     float64
	X2, Y2     float64
	Order      int
	Stereo     int
	Topology   int
}

type Molecule struct {
	Name       string
	Atoms      []Atom
	Bonds      []Bond
	AverageLen float64
}

type State struct {
	MoleculeIdx int
	Thickness   float64
	TextSize    float64
	CenterX     float64
	CenterY     float64
	Scale       float64
	Angle       float64
	AutoRotate  bool
}

type DragState struct {
	Active    bool
	Right     bool
	PDX       float64
	PDY       float64
	PrevScale float64
	PrevAngle float64
}

type layout struct {
	minX, minY float64
	maxX, maxY float64
	baseScale  float64
	textSize   float64
	thickness  float64
}

type point struct {
	x, y float64
}

type contour []point

//go:embed 1.sdf
var sdfData []byte

var molecules = loadMolecules()

var atomPalette = [atomColors]agg.Color{
	agg.Black,
	agg.NewColor(0, 0, 120, 255),
	agg.NewColor(200, 0, 0, 255),
	agg.NewColor(120, 120, 0, 255),
	agg.NewColor(80, 50, 0, 255),
	agg.NewColor(0, 200, 0, 255),
}

func DefaultState() State {
	return State{
		Thickness:  1.0,
		TextSize:   2.5,
		CenterX:    BaseWidth * 0.5,
		CenterY:    BaseHeight * 0.5,
		Scale:      1.0,
		AutoRotate: true,
	}
}

func MoleculeCount() int {
	return len(molecules)
}

func MoleculeName(idx int) string {
	if idx < 0 || idx >= len(molecules) {
		return ""
	}
	return molecules[idx].Name
}

func (s *State) Clamp() {
	if len(molecules) == 0 {
		s.MoleculeIdx = 0
	} else {
		if s.MoleculeIdx < 0 {
			s.MoleculeIdx = 0
		}
		if s.MoleculeIdx >= len(molecules) {
			s.MoleculeIdx = len(molecules) - 1
		}
	}
	if s.Thickness < 0.1 {
		s.Thickness = 0.1
	}
	if s.Thickness > 3.0 {
		s.Thickness = 3.0
	}
	if s.TextSize < 0.5 {
		s.TextSize = 0.5
	}
	if s.TextSize > 5.0 {
		s.TextSize = 5.0
	}
	if s.Scale < 0.05 {
		s.Scale = 0.05
	}
	if s.Scale > 20.0 {
		s.Scale = 20.0
	}
}

func (s *State) Advance() {
	if !s.AutoRotate {
		return
	}
	s.Angle += agg.Deg2RadFunc(0.1)
}

func Draw(ctx *agg.Context, st State) {
	st.Clamp()
	ctx.Clear(agg.White)
	if len(molecules) == 0 {
		return
	}

	mol := molecules[st.MoleculeIdx]
	lay := computeLayout(mol, st)
	angle := st.Angle
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	frameScale, offX, offY := fitFrame(ctx.Width(), ctx.Height())
	centerX := offX + st.CenterX*frameScale
	centerY := offY + st.CenterY*frameScale
	baseScale := lay.baseScale * st.Scale * frameScale
	midX := (lay.minX + lay.maxX) * 0.5
	midY := (lay.minY + lay.maxY) * 0.5
	tm := textMatrix(centerX, centerY, baseScale, angle, midX, midY)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	for _, bond := range mol.Bonds {
		for _, poly := range bondPolygons(bond, st.Thickness*lay.thickness) {
			fillPolygon(ctx, poly, midX, midY, centerX, centerY, baseScale, cosA, sinA)
		}
	}

	for _, atom := range mol.Atoms {
		if atom.Label == "C" {
			continue
		}
		x, y := transformPoint(atom.X, atom.Y, midX, midY, centerX, centerY, baseScale, cosA, sinA)
		ctx.SetColor(agg.White)
		ctx.FillCircle(x, y, lay.textSize*2.5*baseScale)
	}

	text := gsv.NewGSVText()
	text.SetFlip(true)
	text.SetSize(lay.textSize*3.0, 0)
	outline := gsv.NewGSVTextOutlineWithTransform(text, tm)
	outline.SetWidth(st.Thickness * lay.thickness)
	textAdapter := &gsvAdapter{src: outline}
	ras := a.GetInternalRasterizer()

	for _, atom := range mol.Atoms {
		if atom.Label == "C" {
			continue
		}
		text.SetText(atom.Label)
		text.SetStartPoint(atom.X-lay.textSize*1.5, atom.Y-lay.textSize*1.5)
		ras.Reset()
		ras.AddPath(textAdapter, 0)
		a.RenderRasterizerWithColor(atomPalette[atom.ColorIdx])
	}

	a.FontGSV(10)
	a.FillColor(agg.Black)
	a.NoLine()
	a.Text(10, BaseHeight-20, mol.Name, false, 0, 0)
}

func BeginDrag(st *State, drag *DragState, x, y float64, right bool) {
	st.Clamp()
	drag.Active = true
	drag.Right = right
	drag.PDX = st.CenterX - x
	drag.PDY = st.CenterY - y
	drag.PrevScale = st.Scale
	drag.PrevAngle = st.Angle + math.Pi
}

func UpdateDrag(st *State, drag *DragState, x, y float64, right bool) bool {
	if !drag.Active || drag.Right != right {
		return false
	}
	if right {
		st.CenterX = x + drag.PDX
		st.CenterY = y + drag.PDY
		return true
	}
	dx := x - st.CenterX
	dy := y - st.CenterY
	base := math.Hypot(drag.PDX, drag.PDY)
	cur := math.Hypot(dx, dy)
	if base < 1e-6 || cur < 1e-6 {
		return false
	}
	st.Scale = drag.PrevScale * cur / base
	st.Angle = drag.PrevAngle + math.Atan2(dy, dx) - math.Atan2(drag.PDY, drag.PDX)
	st.Clamp()
	return true
}

func EndDrag(drag *DragState) {
	drag.Active = false
}

func PrevMolecule(st *State) {
	if st.MoleculeIdx > 0 {
		st.MoleculeIdx--
	}
}

func NextMolecule(st *State) {
	if st.MoleculeIdx+1 < len(molecules) {
		st.MoleculeIdx++
	}
}

func fitFrame(w, h int) (scale, offX, offY float64) {
	sx := float64(w) / BaseWidth
	sy := float64(h) / BaseHeight
	scale = math.Min(sx, sy)
	if scale > 1.0 {
		scale = 1.0
	}
	if scale <= 0 {
		scale = 1.0
	}
	offX = (float64(w) - BaseWidth*scale) * 0.5
	offY = (float64(h) - BaseHeight*scale) * 0.5
	return scale, offX, offY
}

func computeLayout(m Molecule, st State) layout {
	minX, minY := 1e100, 1e100
	maxX, maxY := -1e100, -1e100
	for _, a := range m.Atoms {
		if a.X < minX {
			minX = a.X
		}
		if a.Y < minY {
			minY = a.Y
		}
		if a.X > maxX {
			maxX = a.X
		}
		if a.Y > maxY {
			maxY = a.Y
		}
	}
	scaleX := BaseWidth / (maxX - minX)
	scaleY := BaseHeight / (maxY - minY)
	baseScale := math.Min(scaleX, scaleY) * 0.80
	textSize := m.AverageLen * st.TextSize / 4.0
	thickness := m.AverageLen / math.Sqrt(math.Max(st.Scale, 0.0001)) / 8.0
	return layout{
		minX:      minX,
		minY:      minY,
		maxX:      maxX,
		maxY:      maxY,
		baseScale: baseScale,
		textSize:  textSize,
		thickness: thickness,
	}
}

func transformPoint(x, y, midX, midY, centerX, centerY, scale, cosA, sinA float64) (float64, float64) {
	x -= midX
	y -= midY
	rx := x*cosA - y*sinA
	ry := x*sinA + y*cosA
	return centerX + rx*scale, centerY + ry*scale
}

func textMatrix(centerX, centerY, scale, angle, midX, midY float64) *transform.TransAffine {
	mtx := transform.NewTransAffine()
	mtx.Translate(-midX, -midY)
	mtx.Scale(scale)
	mtx.Rotate(angle)
	mtx.Translate(centerX, centerY)
	return mtx
}

func fillPolygon(ctx *agg.Context, poly contour, midX, midY, centerX, centerY, scale, cosA, sinA float64) {
	if len(poly) < 3 {
		return
	}
	ctx.BeginPath()
	x, y := transformPoint(poly[0].x, poly[0].y, midX, midY, centerX, centerY, scale, cosA, sinA)
	ctx.MoveTo(x, y)
	for i := 1; i < len(poly); i++ {
		x, y = transformPoint(poly[i].x, poly[i].y, midX, midY, centerX, centerY, scale, cosA, sinA)
		ctx.LineTo(x, y)
	}
	ctx.ClosePath()
	ctx.Fill()
}

func orthogonal(half, x1, y1, x2, y2 float64) (dx, dy float64) {
	vx := x2 - x1
	vy := y2 - y1
	d := math.Hypot(vx, vy)
	if d == 0 {
		return 0, 0
	}
	return -vy * half / d, vx * half / d
}

func bondPolygons(b Bond, thickness float64) []contour {
	switch {
	case b.Order == 1 && b.Stereo == 1:
		return []contour{solidWedge(b.X1, b.Y1, b.X2, b.Y2, thickness)}
	case b.Order == 1 && b.Stereo == 6:
		return dashedWedge(b.X1, b.Y1, b.X2, b.Y2, thickness, 8)
	case b.Order == 2:
		dx, dy := orthogonal(thickness, b.X1, b.Y1, b.X2, b.Y2)
		return []contour{
			lineRect(b.X1-dx, b.Y1-dy, b.X2-dx, b.Y2-dy, thickness),
			lineRect(b.X1+dx, b.Y1+dy, b.X2+dx, b.Y2+dy, thickness),
		}
	default:
		return []contour{lineRect(b.X1, b.Y1, b.X2, b.Y2, thickness)}
	}
}

func lineRect(x1, y1, x2, y2, thickness float64) contour {
	dx, dy := orthogonal(thickness*0.5, x1, y1, x2, y2)
	return contour{
		{x1 - dx, y1 - dy},
		{x2 - dx, y2 - dy},
		{x2 + dx, y2 + dy},
		{x1 + dx, y1 + dy},
	}
}

func solidWedge(x1, y1, x2, y2, thickness float64) contour {
	dx, dy := orthogonal(thickness*2.0, x1, y1, x2, y2)
	return contour{
		{x1, y1},
		{x2 - dx, y2 - dy},
		{x2 + dx, y2 + dy},
	}
}

func dashedWedge(x1, y1, x2, y2, thickness float64, numDashes int) []contour {
	dx, dy := orthogonal(thickness*2.0, x2, y2, x1, y1)
	xt2, yt2 := x1-dx, y1-dy
	xt3, yt3 := x1+dx, y1+dy
	out := make([]contour, 0, numDashes)
	for i := 0; i < numDashes; i++ {
		k1 := float64(i) / float64(numDashes)
		k2 := k1 + 0.4/float64(numDashes)
		out = append(out, contour{
			{x2 + (xt2-x2)*k1, y2 + (yt2-y2)*k1},
			{x2 + (xt2-x2)*k2, y2 + (yt2-y2)*k2},
			{x2 + (xt3-x2)*k2, y2 + (yt3-y2)*k2},
			{x2 + (xt3-x2)*k1, y2 + (yt3-y2)*k1},
		})
	}
	return out
}

type gsvAdapter struct {
	src *gsv.GSVTextOutline
}

func (a *gsvAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *gsvAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func loadMolecules() []Molecule {
	sc := bufio.NewScanner(bytes.NewReader(sdfData))
	sc.Buffer(make([]byte, 1024), 1024*1024)
	var lines []string
	for sc.Scan() {
		lines = append(lines, strings.TrimRight(sc.Text(), "\r"))
	}
	out := make([]Molecule, 0, DefaultCount)
	for i := 0; i < len(lines) && len(out) < DefaultCount; {
		m, next, ok := parseMolecule(lines, i)
		if !ok {
			break
		}
		out = append(out, m)
		i = next
	}
	return out
}

func parseMolecule(lines []string, start int) (Molecule, int, bool) {
	if start+4 > len(lines) {
		return Molecule{}, start, false
	}
	m := Molecule{Name: strings.TrimSpace(lines[start])}
	counts := lines[start+3]
	numAtoms := parseIntField(counts, 0, 3)
	numBonds := parseIntField(counts, 3, 3)
	if numAtoms <= 0 || numBonds < 0 || start+4+numAtoms+numBonds > len(lines) {
		return Molecule{}, start, false
	}
	m.Atoms = make([]Atom, 0, numAtoms)
	lineIdx := start + 4
	for i := 0; i < numAtoms; i++ {
		line := lines[lineIdx+i]
		a := Atom{
			X:      parseFloatField(line, 0, 10),
			Y:      parseFloatField(line, 10, 10),
			Label:  parseStringField(line, 31, 3),
			Charge: parseIntField(line, 38, 1),
		}
		if a.Charge != 0 {
			a.Charge = 4 - a.Charge
		}
		switch a.Label {
		case "N":
			a.ColorIdx = atomN
		case "O":
			a.ColorIdx = atomO
		case "S":
			a.ColorIdx = atomS
		case "P":
			a.ColorIdx = atomP
		case "F", "Cl", "Br", "I":
			a.ColorIdx = atomHalogen
		default:
			a.ColorIdx = atomGeneral
		}
		m.Atoms = append(m.Atoms, a)
	}
	lineIdx += numAtoms
	m.Bonds = make([]Bond, 0, numBonds)
	var sumLen float64
	for i := 0; i < numBonds; i++ {
		line := lines[lineIdx+i]
		idx1 := parseIntField(line, 0, 3) - 1
		idx2 := parseIntField(line, 3, 3) - 1
		if idx1 < 0 || idx2 < 0 || idx1 >= len(m.Atoms) || idx2 >= len(m.Atoms) {
			continue
		}
		b := Bond{
			Idx1:     uint(idx1),
			Idx2:     uint(idx2),
			X1:       m.Atoms[idx1].X,
			Y1:       m.Atoms[idx1].Y,
			X2:       m.Atoms[idx2].X,
			Y2:       m.Atoms[idx2].Y,
			Order:    parseIntField(line, 6, 3),
			Stereo:   parseIntField(line, 9, 3),
			Topology: parseIntField(line, 12, 3),
		}
		sumLen += math.Hypot(b.X1-b.X2, b.Y1-b.Y2)
		m.Bonds = append(m.Bonds, b)
	}
	if len(m.Bonds) > 0 {
		m.AverageLen = sumLen / float64(len(m.Bonds))
	}
	next := lineIdx + numBonds
	for next < len(lines) && !strings.HasPrefix(lines[next], "$$$$") {
		next++
	}
	if next < len(lines) {
		next++
	}
	return m, next, true
}

func parseStringField(line string, start, width int) string {
	if start >= len(line) {
		return ""
	}
	end := start + width
	if end > len(line) {
		end = len(line)
	}
	return strings.TrimSpace(line[start:end])
}

func parseIntField(line string, start, width int) int {
	v, _ := strconv.Atoi(parseStringField(line, start, width))
	return v
}

func parseFloatField(line string, start, width int) float64 {
	v, _ := strconv.ParseFloat(parseStringField(line, start, width), 64)
	return v
}
