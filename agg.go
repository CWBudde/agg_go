package agg

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
)

// NewContext creates a new rendering context with the specified dimensions.
// The context uses RGBA8 pixel format by default.
func NewContext(width, height int) *Context {
	// Create image buffer
	img := &Image{
		Width:  width,
		Height: height,
		Stride: width * 4, // 4 bytes per pixel (RGBA)
		Data:   make([]uint8, width*height*4),
	}

	ctx := &Context{
		width:        width,
		height:       height,
		currentColor: White,
		currentPath:  NewPath(),
	}

	// Initialize internal buffer
	ctx.buffer = buffer.NewRenderingBuffer[uint8]()
	ctx.buffer.Attach(img.Data, width, height, width*4)

	return ctx
}

// NewImage creates a new image with the specified dimensions.
func NewImage(width, height int) *Image {
	return &Image{
		Width:  width,
		Height: height,
		Stride: width * 4,
		Data:   make([]uint8, width*height*4),
	}
}

// SetColor sets the current drawing color.
func (ctx *Context) SetColor(c Color) {
	ctx.currentColor = c
}

// GetColor returns the current drawing color.
func (ctx *Context) GetColor() Color {
	return ctx.currentColor
}

// Clear fills the entire context with the specified color.
func (ctx *Context) Clear(c Color) {
	rgba := c.ConvertToRGBA()
	r8 := uint8(rgba.R * 255)
	g8 := uint8(rgba.G * 255)
	b8 := uint8(rgba.B * 255)
	a8 := uint8(rgba.A * 255)

	// Fill buffer with color
	for i := 0; i < len(ctx.buffer.Buf()); i += 4 {
		data := ctx.buffer.Buf()
		if i+3 < len(data) {
			data[i] = r8   // R
			data[i+1] = g8 // G
			data[i+2] = b8 // B
			data[i+3] = a8 // A
		}
	}
}

// GetImage returns the underlying image data.
func (ctx *Context) GetImage() *Image {
	return &Image{
		Width:  ctx.width,
		Height: ctx.height,
		Stride: ctx.width * 4,
		Data:   ctx.buffer.Buf(),
	}
}

// NewPath creates a new path for drawing.
func NewPath() *Path {
	return &Path{
		commands: make([]pathCommand, 0),
	}
}

// MoveTo moves the current point to the specified coordinates.
func (p *Path) MoveTo(x, y float64) {
	p.commands = append(p.commands, pathCommand{
		cmd:    basics.PathCmdMoveTo,
		points: []Point{{X: x, Y: y}},
	})
}

// LineTo draws a line from the current point to the specified coordinates.
func (p *Path) LineTo(x, y float64) {
	p.commands = append(p.commands, pathCommand{
		cmd:    basics.PathCmdLineTo,
		points: []Point{{X: x, Y: y}},
	})
}

// Close closes the current sub-path.
func (p *Path) Close() {
	if len(p.commands) > 0 {
		// For now, just add an end poly command
		p.commands = append(p.commands, pathCommand{
			cmd:    basics.PathCmdEndPoly,
			points: []Point{},
		})
	}
}

// DrawPath sets the current path for drawing operations.
func (ctx *Context) DrawPath(path *Path) {
	ctx.currentPath = path
}

// Fill fills the current path with the current color.
func (ctx *Context) Fill() {
	// This is a placeholder implementation
	// In a full implementation, this would:
	// 1. Rasterize the path using internal/rasterizer
	// 2. Generate scanlines using internal/scanline
	// 3. Render using internal/renderer
	//
	// For now, we'll implement a simple pixel setting function
	// as a proof of concept
}

// Stroke draws the outline of the current path with the current color.
func (ctx *Context) Stroke() {
	// Placeholder implementation
	// This would use stroke conversion and then fill
}

// Convenience drawing functions

// DrawCircle draws a circle at the specified center with the given radius.
func (ctx *Context) DrawCircle(cx, cy, r float64) {
	path := NewPath()

	// Simple circle approximation using lines (for now)
	steps := 32
	for i := 0; i <= steps; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(steps)
		x := cx + r*math.Cos(angle)
		y := cy + r*math.Sin(angle)

		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()

	ctx.DrawPath(path)
}

// DrawRectangle draws a rectangle with the specified coordinates and dimensions.
func (ctx *Context) DrawRectangle(x, y, w, h float64) {
	path := NewPath()
	path.MoveTo(x, y)
	path.LineTo(x+w, y)
	path.LineTo(x+w, y+h)
	path.LineTo(x, y+h)
	path.Close()

	ctx.DrawPath(path)
}

// DrawRoundedRectangle draws a rounded rectangle.
func (ctx *Context) DrawRoundedRectangle(x, y, w, h, r float64) {
	// Simplified rounded rectangle (just a regular rectangle for now)
	// A full implementation would use arc segments
	ctx.DrawRectangle(x, y, w, h)
}

// Transform applies a transformation to subsequent drawing operations.
func (ctx *Context) Transform(t Transform) {
	// Placeholder - would modify the internal transformation matrix
	ctx.currentTransform = t
}

// Width returns the context width.
func (ctx *Context) Width() int {
	return ctx.width
}

// Height returns the context height.
func (ctx *Context) Height() int {
	return ctx.height
}

// DrawEllipse draws an ellipse outline at the specified center with the given radii.
func (ctx *Context) DrawEllipse(cx, cy, rx, ry float64) {
	// Convert to integer coordinates for primitive rendering
	x := int(cx)
	y := int(cy)
	radiusX := int(rx)
	radiusY := int(ry)

	// For now, implement a simple pixel-level ellipse drawing
	// In a full implementation, this would use the primitives renderer
	ctx.drawEllipseOutline(x, y, radiusX, radiusY)
}

// FillEllipse draws a filled ellipse at the specified center with the given radii.
func (ctx *Context) FillEllipse(cx, cy, rx, ry float64) {
	// Convert to integer coordinates for primitive rendering
	x := int(cx)
	y := int(cy)
	radiusX := int(rx)
	radiusY := int(ry)

	// For now, implement a simple pixel-level ellipse filling
	// In a full implementation, this would use the primitives renderer
	ctx.drawSolidEllipse(x, y, radiusX, radiusY)
}

// drawEllipseOutline is a helper method that draws an ellipse outline using our primitives
func (ctx *Context) drawEllipseOutline(x, y, rx, ry int) {
	// This is a simplified implementation that directly accesses the buffer
	// A full implementation would use the renderer_primitives with proper pixel formats

	if ctx.buffer == nil {
		return
	}

	// Use Bresenham ellipse algorithm directly
	ctx.drawBresenhamEllipse(x, y, rx, ry, false)
}

// drawSolidEllipse is a helper method that draws a filled ellipse using our primitives
func (ctx *Context) drawSolidEllipse(x, y, rx, ry int) {
	// This is a simplified implementation that directly accesses the buffer
	// A full implementation would use the renderer_primitives with proper pixel formats

	if ctx.buffer == nil {
		return
	}

	// Use Bresenham ellipse algorithm for filled ellipse
	ctx.drawBresenhamEllipse(x, y, rx, ry, true)
}

// drawBresenhamEllipse implements the Bresenham ellipse algorithm with optional filling
func (ctx *Context) drawBresenhamEllipse(cx, cy, rx, ry int, filled bool) {
	// Get color components
	rgba := ctx.currentColor.ConvertToRGBA()
	r8 := uint8(rgba.R * 255)
	g8 := uint8(rgba.G * 255)
	b8 := uint8(rgba.B * 255)
	a8 := uint8(rgba.A * 255)

	// Simple bounds checking
	if cx-rx < 0 || cx+rx >= ctx.width || cy-ry < 0 || cy+ry >= ctx.height {
		return
	}

	// Use the ellipse Bresenham algorithm
	x := 0
	y := ry
	rx2 := rx * rx
	ry2 := ry * ry
	twoRx2 := 2 * rx2
	twoRy2 := 2 * ry2
	p := 0
	px := 0
	py := twoRx2 * y

	// Region 1
	p = ry2 - (rx2 * ry) + (rx2 / 4)
	for px < py {
		x++
		px += twoRy2
		if p < 0 {
			p += ry2 + px
		} else {
			y--
			py -= twoRx2
			p += ry2 + px - py
		}

		if filled {
			ctx.drawHorizontalLine(cx-x, cx+x, cy+y, r8, g8, b8, a8)
			ctx.drawHorizontalLine(cx-x, cx+x, cy-y, r8, g8, b8, a8)
		} else {
			ctx.setPixel(cx+x, cy+y, r8, g8, b8, a8)
			ctx.setPixel(cx-x, cy+y, r8, g8, b8, a8)
			ctx.setPixel(cx+x, cy-y, r8, g8, b8, a8)
			ctx.setPixel(cx-x, cy-y, r8, g8, b8, a8)
		}
	}

	// Region 2 - simplified integer calculation
	p = ry2*x*x + rx2*(y-1)*(y-1) - rx2*ry2
	for y > 0 {
		y--
		py -= twoRx2
		if p > 0 {
			p += rx2 - py
		} else {
			x++
			px += twoRy2
			p += rx2 - py + px
		}

		if filled {
			ctx.drawHorizontalLine(cx-x, cx+x, cy+y, r8, g8, b8, a8)
			ctx.drawHorizontalLine(cx-x, cx+x, cy-y, r8, g8, b8, a8)
		} else {
			ctx.setPixel(cx+x, cy+y, r8, g8, b8, a8)
			ctx.setPixel(cx-x, cy+y, r8, g8, b8, a8)
			ctx.setPixel(cx+x, cy-y, r8, g8, b8, a8)
			ctx.setPixel(cx-x, cy-y, r8, g8, b8, a8)
		}
	}
}

// setPixel sets a single pixel in the buffer (bounds checking assumed done)
func (ctx *Context) setPixel(x, y int, r, g, b, a uint8) {
	if x < 0 || x >= ctx.width || y < 0 || y >= ctx.height {
		return
	}

	// Access the buffer directly
	data := ctx.buffer.Buf()
	offset := (y*ctx.width + x) * 4
	if offset+3 < len(data) {
		data[offset] = r
		data[offset+1] = g
		data[offset+2] = b
		data[offset+3] = a
	}
}

// drawHorizontalLine draws a horizontal line from x1 to x2 at y
func (ctx *Context) drawHorizontalLine(x1, x2, y int, r, g, b, a uint8) {
	if y < 0 || y >= ctx.height {
		return
	}

	if x1 > x2 {
		x1, x2 = x2, x1
	}

	if x1 < 0 {
		x1 = 0
	}
	if x2 >= ctx.width {
		x2 = ctx.width - 1
	}

	// Access the buffer directly
	data := ctx.buffer.Buf()
	for x := x1; x <= x2; x++ {
		offset := (y*ctx.width + x) * 4
		if offset+3 < len(data) {
			data[offset] = r
			data[offset+1] = g
			data[offset+2] = b
			data[offset+3] = a
		}
	}
}
