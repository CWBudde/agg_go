package graphtest

import (
	"math"
	"math/rand"

	agg "github.com/MeKo-Christian/agg_go"
)

type Config struct {
	Mode        int
	Width       float64
	Translucent bool
	DrawNodes   bool
	DrawEdges   bool
	NumNodes    int
	NumEdges    int
}

type node struct {
	x float64
	y float64
}

type edge struct {
	n1 int
	n2 int
	r  uint8
	g  uint8
	b  uint8
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
	r, g, b  uint8
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

	rng := rand.New(rand.NewSource(100))
	g := &Graph{
		nodes:    make([]node, numNodes),
		edges:    make([]edge, 0, numEdges),
		prepared: make(map[[2]int]*preparedGraph),
	}
	for i := range g.nodes {
		g.nodes[i] = node{
			x: rng.Float64()*0.75 + 0.2,
			y: rng.Float64()*0.85 + 0.1,
		}
	}
	for len(g.edges) < numEdges {
		n1 := rng.Intn(numNodes)
		n2 := rng.Intn(numNodes)
		if n1 == n2 {
			continue
		}
		g.edges = append(g.edges, edge{
			n1: n1,
			n2: n2,
			r:  uint8(rng.Intn(128)),
			g:  uint8(rng.Intn(128)),
			b:  uint8(rng.Intn(128)),
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

	if cfg.DrawEdges {
		ctx.SetLineWidth(cfg.Width)
		a.NoFill()
		if cfg.Mode == 2 {
			a.AddDash(9, 5)
			a.DashStart(0)
		}
		for _, e := range prepared.edges {
			a8 := uint8(255)
			if cfg.Translucent {
				a8 = 80
			}
			col := agg.NewColor(e.r, e.g, e.b, a8)
			ctx.SetColor(col)

			switch cfg.Mode {
			case 0:
				a.ResetPath()
				a.MoveTo(e.x1, e.y1)
				a.LineTo(e.x2, e.y2)
				a.DrawPath(agg.StrokeOnly)
				drawArrowHead(ctx, e.arrowX0, e.arrowY0, e.arrowX1, e.arrowY1, 8.0, col)
			case 1:
				drawPreparedCurve(a, e, cfg.Width)
				drawArrowHead(ctx, e.arrowX0, e.arrowY0, e.arrowX1, e.arrowY1, 8.0, col)
			case 2:
				drawPreparedCurve(a, e, cfg.Width)
				drawArrowHead(ctx, e.arrowX0, e.arrowY0, e.arrowX1, e.arrowY1, 8.0, col)
			}
		}
		if cfg.Mode == 2 {
			a.NoDashes()
		}
	}

	if cfg.DrawNodes {
		outerR := 5.0 * cfg.Width
		innerR := 4.0

		a.ResetPath()
		for _, n := range prepared.nodes {
			x, y := n.x, n.y
			a.AddEllipse(x, y, outerR, outerR, agg.CCW)
		}
		a.FillColor(agg.NewColor(115, 47, 0, 220))
		a.NoLine()
		a.DrawPath(agg.FillOnly)

		a.ResetPath()
		for _, n := range prepared.nodes {
			x, y := n.x, n.y
			a.AddEllipse(x, y, outerR, outerR, agg.CCW)
		}
		a.FillColor(agg.Transparent)
		a.LineColor(agg.NewColor(154, 74, 0, 255))
		a.LineWidth(1.0)
		a.DrawPath(agg.StrokeOnly)

		a.ResetPath()
		for _, n := range prepared.nodes {
			x, y := n.x, n.y
			a.AddEllipse(x, y, innerR, innerR, agg.CCW)
		}
		a.FillColor(agg.NewColor(248, 202, 80, 230))
		a.NoLine()
		a.DrawPath(agg.FillOnly)
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
			r:       e.r,
			g:       e.g,
			b:       e.b,
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

func drawCurve(a *agg.Agg2D, x1, y1, x2, y2, width float64, dashed bool) {
	cx1, cy1, cx2, cy2 := curveControls(x1, y1, x2, y2)
	a.ResetPath()
	a.MoveTo(x1, y1)
	a.CubicCurveTo(cx1, cy1, cx2, cy2, x2, y2)
	a.LineWidth(width)
	a.DrawPath(agg.StrokeOnly)
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

func drawArrowHeadOnCurve(ctx *agg.Context, x1, y1, x2, y2, size float64, col agg.Color) {
	cx1, cy1, cx2, cy2 := curveControls(x1, y1, x2, y2)
	p0x, p0y := cubicPoint(x1, y1, cx1, cy1, cx2, cy2, x2, y2, 0.92)
	p1x, p1y := cubicPoint(x1, y1, cx1, cy1, cx2, cy2, x2, y2, 1.0)
	drawArrowHead(ctx, p0x, p0y, p1x, p1y, size, col)
}
