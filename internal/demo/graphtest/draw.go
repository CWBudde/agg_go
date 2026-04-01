package graphtest

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/basics"
	icolor "github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	checkboxctrl "github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	rboxctrl "github.com/cwbudde/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/vcgen"
)

type Config struct {
	Mode         int
	Width        float64
	Translucent  bool
	DrawNodes    bool
	DrawEdges    bool
	NumNodes     int
	NumEdges     int
	ShowControls bool
}

type node struct {
	x float64
	y float64
}

type edge struct {
	n1 int
	n2 int
}

type Graph struct {
	nodes    []node
	edges    []edge
	prepared map[[2]int]*preparedGraph
}

type preparedEdge struct {
	x1, y1   float64
	x2, y2   float64
	cx1, cy1 float64
	cx2, cy2 float64
	arrowX0  float64
	arrowY0  float64
	arrowX1  float64
	arrowY1  float64
}

type preparedGraph struct {
	nodes []node
	edges []preparedEdge
}

func NewGraph(numNodes, numEdges int) *Graph {
	if numNodes <= 0 {
		numNodes = 200
	}
	if numEdges <= 0 {
		numEdges = 100
	}

	rng := newClibcRandSeed(100)
	g := &Graph{
		nodes:    make([]node, numNodes),
		edges:    make([]edge, 0, numEdges),
		prepared: make(map[[2]int]*preparedGraph),
	}
	for i := range g.nodes {
		g.nodes[i] = node{
			x: rng.randDouble()*0.75 + 0.2,
			y: rng.randDouble()*0.85 + 0.1,
		}
	}
	for len(g.edges) < numEdges {
		n1 := rng.randN(numNodes)
		n2 := rng.randN(numNodes)
		if n1 == n2 {
			continue
		}
		g.edges = append(g.edges, edge{
			n1: n1,
			n2: n2,
		})
	}
	return g
}

func Draw(ctx *agg.Context, g *Graph, cfg Config) {
	if g == nil {
		g = NewGraph(cfg.NumNodes, cfg.NumEdges)
	}
	if cfg.Width <= 0 {
		cfg.Width = 2.0
	}
	if cfg.Mode < 0 || cfg.Mode > 2 {
		cfg.Mode = 1
	}
	if !cfg.DrawNodes && !cfg.DrawEdges {
		cfg.DrawNodes = true
		cfg.DrawEdges = true
	}

	a := ctx.GetAgg2D()
	a.ResetTransformations()
	ctx.Clear(agg.White)

	w := float64(ctx.GetImage().Width())
	h := float64(ctx.GetImage().Height())
	prepared := g.prepare(int(w), int(h))
	colorRng := newClibcRandSeed(100)

	if cfg.DrawEdges {
		ctx.SetLineWidth(cfg.Width)
		a.NoFill()
		if cfg.Mode == 2 {
			a.AddDash(6, 3)
			a.DashStart(0)
		}
		for _, e := range prepared.edges {
			r := uint8(colorRng.randN(128))
			gc := uint8(colorRng.randN(128))
			b := uint8(colorRng.randN(128))
			a8 := uint8(255)
			if cfg.Translucent {
				a8 = 80
			}
			col := agg.NewColor(r, gc, b, a8)
			ctx.SetColor(col)

			switch cfg.Mode {
			case 0:
				line := &lineSource{x1: e.x1, y1: e.y1, x2: e.x2, y2: e.y2}
				markers := vcgen.NewVCGenMarkersTerm()
				stroke := conv.NewConvStrokeWithMarkers(line, markers)
				stroke.SetWidth(cfg.Width)
				stroke.SetShorten(10.0)
				ah := shapes.NewArrowhead()
				ah.Head(0, 10, 5, 0)
				arrow := conv.NewConvMarker(stroke.Markers(), &arrowheadShapes{ah: ah})
				concat := conv.NewConvConcat(stroke, arrow)
				ras := a.GetInternalRasterizer()
				ras.Reset()
				ras.AddPath(&convToRasSource{src: concat}, 0)
				a.RenderRasterizerWithColor(col)
			case 1:
				drawPreparedCurve(a, e, cfg.Width)
				drawArrowHead(ctx, e.arrowX0, e.arrowY0, e.arrowX1, e.arrowY1, 10.0, col)
			case 2:
				drawPreparedCurve(a, e, cfg.Width)
				drawArrowHead(ctx, e.arrowX0, e.arrowY0, e.arrowX1, e.arrowY1, 10.0, col)
			}
		}
		if cfg.Mode == 2 {
			a.NoDashes()
		}
	}

	if cfg.DrawNodes {
		outerR := 5.0 * cfg.Width

		for _, n := range prepared.nodes {
			a.ResetPath()
			a.AddEllipse(n.x, n.y, outerR, outerR, agg.CCW)
			a.FillRadialGradient(
				n.x, n.y, outerR,
				agg.NewColor(255, 255, 0, 64),
				agg.NewColor(0, 0, 255, 255),
				1.0,
			)
			a.LineColor(agg.Transparent)
			a.DrawPath(agg.FillOnly)
		}
	}

	if cfg.ShowControls {
		drawControls(ctx, cfg)
	}
}

func (g *Graph) prepare(width, height int) *preparedGraph {
	key := [2]int{width, height}
	if pg, ok := g.prepared[key]; ok {
		return pg
	}

	w := float64(width)
	h := float64(height)
	pg := &preparedGraph{
		nodes: make([]node, len(g.nodes)),
		edges: make([]preparedEdge, len(g.edges)),
	}

	for i, n := range g.nodes {
		pg.nodes[i] = node{x: n.x * w, y: n.y * h}
	}
	for i, e := range g.edges {
		n1 := pg.nodes[e.n1]
		n2 := pg.nodes[e.n2]
		cx1, cy1, cx2, cy2 := curveControls(n1.x, n1.y, n2.x, n2.y)
		ax0, ay0 := cubicPoint(n1.x, n1.y, cx1, cy1, cx2, cy2, n2.x, n2.y, 0.92)
		ax1, ay1 := cubicPoint(n1.x, n1.y, cx1, cy1, cx2, cy2, n2.x, n2.y, 1.0)
		pg.edges[i] = preparedEdge{
			x1:      n1.x,
			y1:      n1.y,
			x2:      n2.x,
			y2:      n2.y,
			cx1:     cx1,
			cy1:     cy1,
			cx2:     cx2,
			cy2:     cy2,
			arrowX0: ax0,
			arrowY0: ay0,
			arrowX1: ax1,
			arrowY1: ay1,
		}
	}

	g.prepared[key] = pg
	return pg
}

func cubicPoint(x1, y1, cx1, cy1, cx2, cy2, x2, y2, t float64) (float64, float64) {
	u := 1.0 - t
	tt := t * t
	uu := u * u
	uuu := uu * u
	ttt := tt * t
	x := uuu*x1 + 3*uu*t*cx1 + 3*u*tt*cx2 + ttt*x2
	y := uuu*y1 + 3*uu*t*cy1 + 3*u*tt*cy2 + ttt*y2
	return x, y
}

func curveControls(x1, y1, x2, y2 float64) (float64, float64, float64, float64) {
	k := 0.5
	dx := x2 - x1
	dy := y2 - y1
	cx1 := x1 - dy*k
	cy1 := y1 + dx*k
	cx2 := x2 + dy*k
	cy2 := y2 - dx*k
	return cx1, cy1, cx2, cy2
}

func drawPreparedCurve(a *agg.Agg2D, e preparedEdge, width float64) {
	a.ResetPath()
	a.MoveTo(e.x1, e.y1)
	a.CubicCurveTo(e.cx1, e.cy1, e.cx2, e.cy2, e.x2, e.y2)
	a.LineWidth(width)
	a.DrawPath(agg.StrokeOnly)
}

func drawArrowHead(ctx *agg.Context, x1, y1, x2, y2, size float64, col agg.Color) {
	dx := x2 - x1
	dy := y2 - y1
	l := math.Hypot(dx, dy)
	if l < 1e-6 {
		return
	}
	ux, uy := dx/l, dy/l
	px, py := -uy, ux
	tx, ty := x2-ux*size, y2-uy*size
	lx, ly := tx+px*size*0.5, ty+py*size*0.5
	rx, ry := tx-px*size*0.5, ty-py*size*0.5

	a := ctx.GetAgg2D()
	ctx.SetColor(col)
	a.ResetPath()
	a.MoveTo(x2, y2)
	a.LineTo(lx, ly)
	a.LineTo(rx, ry)
	a.ClosePolygon()
	a.NoLine()
	a.DrawPath(agg.FillOnly)
}

type lineSource struct {
	x1, y1 float64
	x2, y2 float64
	f      int
}

func (l *lineSource) Rewind(pathID uint) { l.f = 0 }

func (l *lineSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	switch l.f {
	case 0:
		l.f++
		return l.x1, l.y1, basics.PathCmdMoveTo
	case 1:
		l.f++
		return l.x2, l.y2, basics.PathCmdLineTo
	default:
		return 0, 0, basics.PathCmdStop
	}
}

type arrowheadShapes struct{ ah *shapes.Arrowhead }

func (a *arrowheadShapes) Rewind(shapeIndex uint) { a.ah.Rewind(uint32(shapeIndex)) }
func (a *arrowheadShapes) Vertex() (x, y float64, cmd basics.PathCommand) {
	var vx, vy float64
	c := a.ah.Vertex(&vx, &vy)
	return vx, vy, c
}

type convToRasSource struct{ src conv.VertexSource }

func (a *convToRasSource) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *convToRasSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRandSeed(seed int32) *clibcRand {
	if seed == 0 {
		seed = 1
	}

	r := &clibcRand{}
	r.state[0] = seed
	for i := 1; i < len(r.state); i++ {
		next := (16807 * int64(r.state[i-1])) % 2147483647
		r.state[i] = int32(next)
	}
	r.fptr = 3
	r.rptr = 0
	for i := 0; i < 310; i++ {
		r.next()
	}
	return r
}

func (r *clibcRand) next() int32 {
	r.state[r.fptr] += r.state[r.rptr]
	result := int32(uint32(r.state[r.fptr]) >> 1)
	r.fptr++
	if r.fptr >= len(r.state) {
		r.fptr = 0
	}
	r.rptr++
	if r.rptr >= len(r.state) {
		r.rptr = 0
	}
	return result
}

func (r *clibcRand) randN(n int) int {
	return int(r.next()) % n
}

func (r *clibcRand) randDouble() float64 {
	return float64(r.next()) / 2147483647.0
}

type ctrlPathSource interface {
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) icolor.RGBA
}

type ctrlPathAdapter struct {
	ctrl ctrlPathSource
}

func (a *ctrlPathAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func toAggColor(c icolor.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255 + 0.5)
		}
	}
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

func renderCtrl(ctx *agg.Context, ctrl ctrlPathSource) {
	a := ctx.GetAgg2D()
	ras := a.GetInternalRasterizer()
	adapter := &ctrlPathAdapter{ctrl: ctrl}

	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(ctrl.Color(pathID)))
	}
}

func drawControls(ctx *agg.Context, cfg Config) {
	typeCtrl := rboxctrl.NewDefaultRboxCtrl(0, 0, 110, 95, false)
	typeCtrl.SetTextSize(8.0, 0.0)
	for _, item := range []string{"Poygons Bin", "Poygons AA", "Dashed curves", "Bezier curves", "Solid lines"} {
		typeCtrl.AddItem(item)
	}
	typeCtrl.SetCurItem(4 - cfg.Mode)

	widthCtrl := sliderctrl.NewSliderCtrl(190, 8, 390, 15, false)
	widthCtrl.SetNumSteps(20)
	widthCtrl.SetRange(0.0, 5.0)
	widthCtrl.SetValue(cfg.Width)
	widthCtrl.SetLabel("Width=%1.2f")
	widthCtrl.SetTextThickness(1.0)

	benchmarkCtrl := checkboxctrl.NewDefaultCheckboxCtrl(398, 6, "Benchmark", false)
	drawNodesCtrl := checkboxctrl.NewDefaultCheckboxCtrl(398, 21, "Draw Nodes", false)
	drawEdgesCtrl := checkboxctrl.NewDefaultCheckboxCtrl(488, 21, "Draw Edges", false)
	draftCtrl := checkboxctrl.NewDefaultCheckboxCtrl(488, 6, "Draft Mode", false)
	translucentCtrl := checkboxctrl.NewDefaultCheckboxCtrl(190, 21, "Translucent Mode", false)

	benchmarkCtrl.SetChecked(false)
	drawNodesCtrl.SetChecked(cfg.DrawNodes)
	drawEdgesCtrl.SetChecked(cfg.DrawEdges)
	draftCtrl.SetChecked(false)
	translucentCtrl.SetChecked(cfg.Translucent)

	benchmarkCtrl.SetTextSize(8.0, 0.0)
	drawNodesCtrl.SetTextSize(8.0, 0.0)
	drawEdgesCtrl.SetTextSize(8.0, 0.0)
	draftCtrl.SetTextSize(8.0, 0.0)
	translucentCtrl.SetTextSize(8.0, 0.0)

	for _, ctrl := range []ctrlPathSource{
		typeCtrl,
		widthCtrl,
		benchmarkCtrl,
		drawNodesCtrl,
		drawEdgesCtrl,
		draftCtrl,
		translucentCtrl,
	} {
		renderCtrl(ctx, ctrl)
	}
}
