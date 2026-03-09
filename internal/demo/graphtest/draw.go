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
}

type Graph struct {
	nodes []node
	edges []edge
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
		nodes: make([]node, numNodes),
		edges: make([]edge, 0, numEdges),
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
		g.edges = append(g.edges, edge{n1: n1, n2: n2})
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

	rng := rand.New(rand.NewSource(100))
	if cfg.DrawEdges {
		for _, e := range g.edges {
			n1 := g.nodes[e.n1]
			n2 := g.nodes[e.n2]
			x1, y1 := n1.x*w, n1.y*h
			x2, y2 := n2.x*w, n2.y*h

			r := uint8(rng.Intn(128))
			gch := uint8(rng.Intn(128))
			b := uint8(rng.Intn(128))
			a8 := uint8(255)
			if cfg.Translucent {
				a8 = 80
			}
			col := agg.NewColor(r, gch, b, a8)
			ctx.SetColor(col)
			ctx.SetLineWidth(cfg.Width)
			a.NoFill()

			switch cfg.Mode {
			case 0:
				a.ResetPath()
				a.MoveTo(x1, y1)
				a.LineTo(x2, y2)
				a.DrawPath(agg.StrokeOnly)
				drawArrowHead(ctx, x1, y1, x2, y2, 8.0, col)
			case 1:
				drawCurve(a, x1, y1, x2, y2, cfg.Width, false)
				drawArrowHeadOnCurve(ctx, x1, y1, x2, y2, 8.0, col)
			case 2:
				drawCurve(a, x1, y1, x2, y2, cfg.Width, true)
				drawArrowHeadOnCurve(ctx, x1, y1, x2, y2, 8.0, col)
			}
		}
	}

	if cfg.DrawNodes {
		outerR := 5.0 * cfg.Width
		innerR := 4.0
		for _, n := range g.nodes {
			x, y := n.x*w, n.y*h
			ctx.SetColor(agg.NewColor(115, 47, 0, 220))
			ctx.FillCircle(x, y, outerR)
			ctx.SetColor(agg.NewColor(154, 74, 0, 255))
			ctx.SetLineWidth(1.0)
			ctx.DrawCircle(x, y, outerR)
			ctx.SetColor(agg.NewColor(248, 202, 80, 230))
			ctx.FillCircle(x, y, innerR)
		}
	}
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
	const steps = 48

	if dashed {
		for i := 0; i < steps; i++ {
			if i%9 >= 6 {
				continue
			}
			t0 := float64(i) / steps
			t1 := float64(i+1) / steps
			px0, py0 := cubicPoint(x1, y1, cx1, cy1, cx2, cy2, x2, y2, t0)
			px1, py1 := cubicPoint(x1, y1, cx1, cy1, cx2, cy2, x2, y2, t1)
			a.ResetPath()
			a.MoveTo(px0, py0)
			a.LineTo(px1, py1)
			a.LineWidth(width)
			a.DrawPath(agg.StrokeOnly)
		}
		return
	}

	a.ResetPath()
	for i := 0; i <= steps; i++ {
		t := float64(i) / steps
		px, py := cubicPoint(x1, y1, cx1, cy1, cx2, cy2, x2, y2, t)
		if i == 0 {
			a.MoveTo(px, py)
		} else {
			a.LineTo(px, py)
		}
	}
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
