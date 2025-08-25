package ctrl

import (
	"agg_go/internal/renderer/scanline"
)

// RasterizerInterface defines the interface that rasterizers must implement for control rendering.
// This extends the scanline.RasterizerInterface with the AddPath method needed for controls.
type RasterizerInterface interface {
	scanline.RasterizerInterface
	Reset()
	AddPath(vs VertexSourceInterface, pathID uint)
}

// ScanlineInterface defines the interface that scanlines must implement for control rendering.
// This uses the same interface as the scanline package.
type ScanlineInterface = scanline.ScanlineInterface

// RendererInterface defines the interface that renderers must implement for control rendering.
// This uses the scanline package's renderer interfaces.
type RendererInterface[C any] interface {
	scanline.RendererInterface[C]
}

// BaseRendererInterface defines the interface for base renderers used in control rendering.
// This uses the same interface as the scanline package.
type BaseRendererInterface[C any] = scanline.BaseRendererInterface[C]

// VertexSourceInterface defines the interface for vertex sources used in control rendering.
type VertexSourceInterface interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd uint32)
}

// RenderCtrl renders a control using the specified rasterizer, scanline, and renderer.
// This is the Go equivalent of the C++ render_ctrl template function.
//
// The function iterates through all paths in the control, rasterizes each one,
// and renders it with the path's associated color using anti-aliased solid rendering.
func RenderCtrl[R RasterizerInterface, S ScanlineInterface, Ren BaseRendererInterface[C], C any](
	ras R, sl S, renderer Ren, ctrl Ctrl[C],
) {
	numPaths := ctrl.NumPaths()

	for i := uint(0); i < numPaths; i++ {
		// Reset rasterizer for new path
		ras.Reset()

		// Add the control's path to the rasterizer
		// We need to create an adapter since ctrl implements vertex source differently
		adapter := &ctrlVertexSourceAdapter[C]{ctrl: ctrl, pathID: i}
		ras.AddPath(adapter, i)

		// Get the color for this path
		color := ctrl.Color(i)

		// Render the scanlines with anti-aliased solid color
		// This corresponds to the C++ render_scanlines_aa_solid call
		scanline.RenderScanlinesAASolid(ras, sl, renderer, color)
	}
}

// RenderCtrlRS renders a control using render storage variant.
// This corresponds to the C++ render_ctrl_rs template function.
func RenderCtrlRS[R RasterizerInterface, S ScanlineInterface, Ren RendererInterface[C], C any](
	ras R, sl S, renderer Ren, ctrl Ctrl[C],
) {
	numPaths := ctrl.NumPaths()

	for i := uint(0); i < numPaths; i++ {
		// Reset rasterizer for new path
		ras.Reset()

		// Add the control's path to the rasterizer
		adapter := &ctrlVertexSourceAdapter[C]{ctrl: ctrl, pathID: i}
		ras.AddPath(adapter, i)

		// Set color on renderer and render
		// This corresponds to the C++ r.color(c.color(i)) call
		renderer.SetColor(ctrl.Color(i))

		// Render the scanlines using the renderer's current color
		// This corresponds to the C++ render_scanlines call
		scanline.RenderScanlines(ras, sl, renderer)
	}
}

// ctrlVertexSourceAdapter adapts a Ctrl to the VertexSourceInterface needed by rasterizers.
type ctrlVertexSourceAdapter[C any] struct {
	ctrl   Ctrl[C]
	pathID uint
}

// Rewind rewinds to the beginning of the specified path.
func (adapter *ctrlVertexSourceAdapter[C]) Rewind(pathID uint) {
	adapter.pathID = pathID
	adapter.ctrl.Rewind(pathID)
}

// Vertex returns the next vertex in the current path.
func (adapter *ctrlVertexSourceAdapter[C]) Vertex() (x, y float64, cmd uint32) {
	x, y, pathCmd := adapter.ctrl.Vertex()
	return x, y, uint32(pathCmd)
}

// SimpleRenderCtrl provides a simplified control rendering function that works with
// the existing AGG Go implementation. This is a temporary solution until the
// generic rendering system is fully integrated.
func SimpleRenderCtrl[C any](ctrl Ctrl[C], renderFunc func(pathID uint, vertices []Vertex, color C)) {
	numPaths := ctrl.NumPaths()

	for pathID := uint(0); pathID < numPaths; pathID++ {
		// Collect vertices for this path
		var vertices []Vertex
		ctrl.Rewind(pathID)

		for {
			x, y, cmd := ctrl.Vertex()
			if cmd == 0 { // PathCmdStop
				break
			}

			vertices = append(vertices, Vertex{
				X:   x,
				Y:   y,
				Cmd: uint32(cmd),
			})
		}

		// Render the path with its color
		color := ctrl.Color(pathID)
		renderFunc(pathID, vertices, color)
	}
}

// Vertex represents a single vertex with coordinates and command.
type Vertex struct {
	X, Y float64
	Cmd  uint32
}

// Helper function to create a simple render function for testing
func CreateTestRenderFunc[C any]() func(pathID uint, vertices []Vertex, color C) {
	return func(pathID uint, vertices []Vertex, color C) {
		// This is a no-op render function for testing
		// Real implementations would draw the vertices using the provided color
		_ = pathID
		_ = vertices
		_ = color
	}
}
