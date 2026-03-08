// Port of AGG C++ gouraud_mesh.cpp – Gouraud-shaded mesh rendering.
//
// Renders a grid of randomly coloured vertices as Gouraud-shaded triangles.
// Uses agg2d.GouraudTriangle for each triangle in the mesh.
// Default: 10×10 mesh on an 800×600 canvas.
package main

import (
	"math/rand"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
)

const (
	meshCols = 10
	meshRows = 10
)

type meshPoint struct {
	x, y  float64
	color agg.Color
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	w := ctx.GetImage().Width()
	h := ctx.GetImage().Height()
	ctx.Clear(agg.Black)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	rng := rand.New(rand.NewSource(1234))

	cellW := float64(w-80) / float64(meshCols-1)
	cellH := float64(h-80) / float64(meshRows-1)
	startX, startY := 40.0, 40.0

	// Build mesh vertices.
	vertices := make([]meshPoint, meshRows*meshCols)
	for i := 0; i < meshRows; i++ {
		for j := 0; j < meshCols; j++ {
			x := startX + float64(j)*cellW
			y := startY + float64(i)*cellH
			vertices[i*meshCols+j] = meshPoint{
				x: x,
				y: y,
				color: agg.NewColor(
					uint8(rng.Intn(256)),
					uint8(rng.Intn(256)),
					uint8(rng.Intn(256)),
					255,
				),
			}
		}
	}

	// Render two triangles per quad cell.
	for i := 0; i < meshRows-1; i++ {
		for j := 0; j < meshCols-1; j++ {
			p00 := vertices[i*meshCols+j]
			p10 := vertices[i*meshCols+j+1]
			p01 := vertices[(i+1)*meshCols+j]
			p11 := vertices[(i+1)*meshCols+j+1]

			// Triangle 1: top-left, top-right, bottom-left.
			a.GouraudTriangle(
				p00.x, p00.y,
				p10.x, p10.y,
				p01.x, p01.y,
				p00.color, p10.color, p01.color,
				0,
			)

			// Triangle 2: top-right, bottom-right, bottom-left.
			a.GouraudTriangle(
				p10.x, p10.y,
				p11.x, p11.y,
				p01.x, p01.y,
				p10.color, p11.color, p01.color,
				0,
			)
		}
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Gouraud Mesh",
		Width:  800,
		Height: 600,
	}, &demo{})
}
