# Fundamental AGG Concepts (Go Translation)

This document introduces the core concepts of Anti-Grain Geometry (AGG) as implemented in the Go port, adapting the original C++ library to Go's idioms and best practices.

## Table of Contents

- [Anti-Aliasing Principles](#anti-aliasing-principles)
- [Vector Graphics Rendering Pipeline](#vector-graphics-rendering-pipeline)
- [Coordinate System and Transformations](#coordinate-system-and-transformations)
- [Color Models and Gamma Correction](#color-models-and-gamma-correction)
- [Scanline Rendering Fundamentals](#scanline-rendering-fundamentals)
- [Coverage Calculation Methods](#coverage-calculation-methods)
- [Go-Specific Considerations](#go-specific-considerations)

---

## Anti-Aliasing Principles

### What is Anti-Aliasing?

Anti-aliasing is the technique of smoothing jagged edges (aliasing artifacts) that appear when rendering vector graphics on pixel-based displays. In AGG Go, anti-aliasing is achieved through **subpixel accuracy** and **coverage calculation**.

### Subpixel Accuracy in Go

Unlike traditional pixel-based rendering, AGG Go maintains floating-point precision throughout the rendering pipeline:

```go
// Example: Drawing a line with subpixel positioning
ctx := agg.NewContext(800, 600)
ctx.SetColor(agg.Red)

// These coordinates have subpixel precision
ctx.DrawLine(100.3, 150.7, 250.1, 200.9)
```

The Go port preserves the original AGG's subpixel accuracy using Go's `float64` type throughout the pipeline, ensuring smooth edges regardless of zoom level.

### Coverage-Based Anti-Aliasing

AGG Go calculates the **coverage** of each pixel - what percentage of the pixel is covered by the geometric shape:

```go
// Conceptual representation of coverage calculation
type Coverage struct {
    Value uint32 // 0 to 256 (where 256 = 100% coverage)
}

// In practice, this is handled internally by the rasterizer
func (r *Rasterizer) calculateCoverage(pixel Pixel, shape Geometry) Coverage {
    // Complex geometric calculation determining overlap
    // Returns fractional coverage for anti-aliasing
}
```

### Quality Levels

AGG Go provides different anti-aliasing quality settings:

```go
// Through the Context API
ctx.SetAntiAliasing(agg.AAModeHigh)    // Best quality, slower
ctx.SetAntiAliasing(agg.AAModeMedium)  // Balanced
ctx.SetAntiAliasing(agg.AAModeOff)     // No anti-aliasing, fastest
```

---

## Vector Graphics Rendering Pipeline

### Pipeline Overview

The AGG Go rendering pipeline transforms vector graphics into high-quality rasterized output through these stages:

```
Path Definition → Transformation → Conversion → Rasterization → Scanline Generation → Pixel Rendering
```

### 1. Path Definition

Paths in Go are built using method chaining for clean, readable code:

```go
ctx := agg.NewContext(800, 600)

// Building a complex path
ctx.BeginPath()
ctx.MoveTo(100, 100)
ctx.LineTo(200, 150)
ctx.CurveTo(250, 100, 300, 150, 350, 100) // Cubic Bézier
ctx.ClosePath()
ctx.Fill()
```

Internally, paths are stored efficiently using Go slices:

```go
type PathStorage struct {
    vertices []Vertex  // Dynamic array of path vertices
    commands []Command // Corresponding path commands (MoveTo, LineTo, etc.)
}
```

### 2. Transformation Stage

Transformations use Go's math package for precision:

```go
import "math"

// Create and apply transformations
transform := agg.NewTransform()
transform.Scale(1.5, 1.5)
transform.Rotate(math.Pi / 4) // 45 degrees
transform.Translate(50, 30)

ctx.ApplyTransform(transform)
ctx.DrawRectangle(0, 0, 100, 100) // Will be transformed
```

### 3. Conversion Stage

Path converters modify the geometry before rasterization:

```go
// Stroke conversion example
ctx.SetStrokeWidth(5.0)
ctx.SetStrokeJoin(agg.JoinRound)
ctx.SetStrokeCap(agg.CapRound)

// Dash pattern conversion
ctx.SetDashPattern([]float64{10, 5, 2, 5}) // dash, gap, dash, gap
ctx.SetDashOffset(3.0)
```

### 4. Rasterization

The rasterizer converts vector paths to coverage data:

```go
// Simplified internal rasterization process
type Rasterizer struct {
    cells []Cell  // Coverage cells for anti-aliasing
    width, height int
}

func (r *Rasterizer) AddPath(path *Path) {
    // Scan-convert the path into coverage cells
    // Each cell contains coverage information for anti-aliasing
}
```

### 5. Scanline Generation

Scanlines are horizontal strips of pixels with associated coverage data:

```go
type Scanline struct {
    Y     int        // Y coordinate
    Spans []Span     // Horizontal spans with coverage
}

type Span struct {
    X        int      // Starting X coordinate
    Length   int      // Length of span
    Coverage []uint8  // Coverage values (0-255)
}
```

### 6. Pixel Rendering

The final stage applies colors and blending:

```go
// Renderer applies color and blending to pixels
type Renderer struct {
    pixelFormat PixelFormat
    buffer      []uint8
    blender     Blender
}

func (r *Renderer) RenderScanline(scanline Scanline, color Color) {
    // Apply color and blending for each span
}
```

---

## Coordinate System and Transformations

### Coordinate System

AGG Go uses a standard 2D coordinate system:

- **Origin (0,0)**: Top-left corner
- **X-axis**: Increases rightward
- **Y-axis**: Increases downward (screen coordinates)
- **Units**: Pixels with subpixel precision

```go
// Working with coordinates
ctx := agg.NewContext(800, 600)

// Top-left quadrant
ctx.DrawCircle(200, 150, 50)

// Bottom-right quadrant
ctx.DrawCircle(600, 450, 50)
```

### Transformation Matrix

AGG Go uses 2x3 affine transformation matrices:

```go
type Transform struct {
    // Matrix elements: [sx, shx, shy, sy, tx, ty]
    // Represents: [sx shx tx]
    //            [shy sy ty]
    //            [0   0   1]
    sx, shx, shy, sy, tx, ty float64
}

func (t *Transform) Apply(x, y float64) (float64, float64) {
    newX := t.sx*x + t.shx*y + t.tx
    newY := t.shy*x + t.sy*y + t.ty
    return newX, newY
}
```

### Common Transformations

```go
// Identity transformation (no change)
identity := agg.NewTransform()

// Translation
translate := agg.NewTransform().Translate(50, 100)

// Scaling
scale := agg.NewTransform().Scale(2.0, 1.5)

// Rotation around origin
rotate := agg.NewTransform().Rotate(math.Pi / 3) // 60 degrees

// Combined transformations (applied right-to-left)
combined := agg.NewTransform().
    Translate(100, 50).   // 3. Finally translate
    Rotate(math.Pi/4).    // 2. Then rotate
    Scale(1.5, 1.5)       // 1. First scale
```

### Viewport Transformations

Convert between different coordinate systems:

```go
type Viewport struct {
    worldX1, worldY1, worldX2, worldY2 float64 // World coordinates
    deviceX1, deviceY1, deviceX2, deviceY2 float64 // Device coordinates
}

func (v *Viewport) WorldToDevice(worldX, worldY float64) (float64, float64) {
    // Transform from world coordinates to device coordinates
    scaleX := (v.deviceX2 - v.deviceX1) / (v.worldX2 - v.worldX1)
    scaleY := (v.deviceY2 - v.deviceY1) / (v.worldY2 - v.worldY1)

    deviceX := v.deviceX1 + (worldX - v.worldX1) * scaleX
    deviceY := v.deviceY1 + (worldY - v.worldY1) * scaleY

    return deviceX, deviceY
}
```

---

## Color Models and Gamma Correction

### Color Types in Go

AGG Go provides type-safe color handling:

```go
// RGBA color with 8-bit channels
type RGBA8 struct {
    R, G, B, A uint8
}

// RGBA color with floating-point channels (0.0 to 1.0)
type RGBA struct {
    R, G, B, A float64
}

// Grayscale color
type Gray8 struct {
    V uint8 // Value (luminance)
    A uint8 // Alpha
}

// Pre-multiplied alpha for efficient blending
type RGBAPre struct {
    R, G, B, A float64 // RGB values pre-multiplied by alpha
}
```

### Color Spaces

AGG Go supports multiple color spaces with proper conversions:

```go
// Linear RGB (gamma = 1.0)
type LinearRGB struct {
    RGBA
}

// sRGB (gamma ≈ 2.2)
type SRGB struct {
    RGBA
}

// Convert between color spaces
func LinearToSRGB(linear LinearRGB) SRGB {
    // Apply gamma correction
    gamma := 1.0 / 2.2
    return SRGB{
        R: math.Pow(linear.R, gamma),
        G: math.Pow(linear.G, gamma),
        B: math.Pow(linear.B, gamma),
        A: linear.A, // Alpha is not gamma corrected
    }
}
```

### Gamma Correction

Gamma correction ensures consistent color appearance across devices:

```go
type GammaLUT struct {
    forward  []uint8 // Linear to gamma-corrected
    inverse  []uint8 // Gamma-corrected to linear
    gamma    float64
}

func NewGammaLUT(gamma float64) *GammaLUT {
    lut := &GammaLUT{
        forward: make([]uint8, 256),
        inverse: make([]uint8, 256),
        gamma:   gamma,
    }

    // Build lookup tables for fast conversion
    for i := 0; i < 256; i++ {
        linear := float64(i) / 255.0
        corrected := math.Pow(linear, 1.0/gamma)
        lut.forward[i] = uint8(corrected * 255.0)

        corrected = float64(i) / 255.0
        linear = math.Pow(corrected, gamma)
        lut.inverse[i] = uint8(linear * 255.0)
    }

    return lut
}
```

### Working with Colors

```go
// Creating colors
red := agg.RGBA8{R: 255, G: 0, B: 0, A: 255}
transparentBlue := agg.RGBA8{R: 0, G: 0, B: 255, A: 128}

// Named colors for convenience
ctx.SetColor(agg.Red)
ctx.SetColor(agg.Blue.WithAlpha(0.5))

// Color interpolation
color1 := agg.RGBA8{R: 255, G: 0, B: 0, A: 255}   // Red
color2 := agg.RGBA8{R: 0, G: 0, B: 255, A: 255}   // Blue
interpolated := agg.LerpColor(color1, color2, 0.5) // Purple
```

---

## Scanline Rendering Fundamentals

### What are Scanlines?

Scanlines are horizontal rows of pixels that make up the final rendered image. AGG Go processes images one scanline at a time for memory efficiency:

```go
type Scanline interface {
    // Reset scanline for reuse
    Reset()

    // Add a span (horizontal segment) to this scanline
    AddSpan(x, length int, coverage []uint8)

    // Get Y coordinate of this scanline
    Y() int

    // Iterate over spans in this scanline
    Spans() []Span
}
```

### Scanline Storage Types

AGG Go provides different scanline storage formats optimized for different use cases:

```go
// Packed scanline - memory efficient
type ScanlineP struct {
    y       int
    spans   []Span
    packed  bool
}

// Unpacked scanline - faster processing
type ScanlineU struct {
    y           int
    minX, maxX  int
    coverage    []uint8  // Direct coverage array
}

// Binary scanline - for solid fills (no anti-aliasing)
type ScanlineBin struct {
    y       int
    spans   []SpanBin
}
```

### Scanline Processing Example

```go
func RenderPath(path *Path, renderer *Renderer) {
    rasterizer := NewRasterizer()
    scanline := NewScanlineU()

    // Convert path to coverage data
    rasterizer.AddPath(path)

    // Process each scanline
    for scanline.Reset(); rasterizer.SweepScanline(scanline); {
        // Render this scanline with current color/pattern
        renderer.RenderScanline(scanline)
    }
}
```

### Memory Efficiency

AGG Go reuses scanline objects to minimize garbage collection:

```go
type ScanlinePool struct {
    scanlines []*ScanlineU
    index     int
}

func (p *ScanlinePool) Get() *ScanlineU {
    if p.index >= len(p.scanlines) {
        return NewScanlineU()
    }
    scanline := p.scanlines[p.index]
    p.index++
    scanline.Reset()
    return scanline
}

func (p *ScanlinePool) Put(scanline *ScanlineU) {
    if p.index > 0 {
        p.index--
        p.scanlines[p.index] = scanline
    }
}
```

---

## Coverage Calculation Methods

### Understanding Coverage

Coverage represents how much of a pixel is covered by a geometric shape, ranging from 0 (not covered) to 256 (fully covered) in AGG Go's internal representation:

```go
type Coverage uint32

const (
    CoverageNone = Coverage(0)   // 0% coverage
    CoverageFull = Coverage(256) // 100% coverage
)

// Convert coverage to alpha value (0-255)
func (c Coverage) ToAlpha() uint8 {
    return uint8(c >> 8) // Divide by 256
}
```

### Analytical Anti-Aliasing

AGG Go uses analytical methods to calculate exact coverage:

```go
// Calculate coverage for a line segment crossing a pixel
func LineSegmentCoverage(x1, y1, x2, y2 float64, pixelX, pixelY int) Coverage {
    // Convert to pixel-relative coordinates
    px := float64(pixelX)
    py := float64(pixelY)

    // Calculate intersection points with pixel boundaries
    // This is a simplified version - actual implementation is more complex

    if x1 == x2 { // Vertical line
        if x1 >= px && x1 <= px+1 {
            // Line passes through pixel
            coverage := calculateVerticalLineCoverage(x1, y1, y2, px, py)
            return Coverage(coverage * 256)
        }
        return CoverageNone
    }

    // General case involves complex geometric calculations
    return calculateGeneralLineCoverage(x1, y1, x2, y2, px, py)
}
```

### Supersampling (Alternative Approach)

While AGG Go primarily uses analytical methods, it also supports supersampling for comparison:

```go
func SupersampleCoverage(shape Geometry, pixel Rect, samples int) Coverage {
    covered := 0
    total := samples * samples

    stepX := 1.0 / float64(samples)
    stepY := 1.0 / float64(samples)

    for i := 0; i < samples; i++ {
        for j := 0; j < samples; j++ {
            x := pixel.X + float64(i)*stepX + stepX/2
            y := pixel.Y + float64(j)*stepY + stepY/2

            if shape.Contains(x, y) {
                covered++
            }
        }
    }

    return Coverage((covered * 256) / total)
}
```

### Gamma-Aware Coverage

For accurate color reproduction, AGG Go applies gamma correction to coverage values:

```go
func GammaCorrectedCoverage(coverage Coverage, gamma float64) Coverage {
    if coverage == 0 {
        return 0
    }

    // Convert to 0-1 range
    normalized := float64(coverage) / 256.0

    // Apply gamma correction
    corrected := math.Pow(normalized, 1.0/gamma)

    // Convert back to coverage range
    return Coverage(corrected * 256.0)
}
```

---

## Go-Specific Considerations

### Memory Management

Unlike C++ AGG's manual memory management, AGG Go leverages Go's garbage collector:

```go
// Efficient buffer reuse to minimize GC pressure
type BufferPool struct {
    buffers [][]uint8
    sizes   []int
}

func (p *BufferPool) GetBuffer(size int) []uint8 {
    for i, s := range p.sizes {
        if s >= size {
            buffer := p.buffers[i]
            // Remove from pool
            p.buffers = append(p.buffers[:i], p.buffers[i+1:]...)
            p.sizes = append(p.sizes[:i], p.sizes[i+1:]...)
            return buffer[:size]
        }
    }

    // Create new buffer
    return make([]uint8, size)
}

func (p *BufferPool) PutBuffer(buffer []uint8) {
    p.buffers = append(p.buffers, buffer)
    p.sizes = append(p.sizes, cap(buffer))
}
```

### Slice Usage Patterns

AGG Go uses slices extensively, following Go idioms:

```go
// Growing slices efficiently
type VertexArray struct {
    vertices []Vertex
}

func (va *VertexArray) AddVertex(v Vertex) {
    if len(va.vertices) == cap(va.vertices) {
        // Double capacity when full
        newVertices := make([]Vertex, len(va.vertices), cap(va.vertices)*2)
        copy(newVertices, va.vertices)
        va.vertices = newVertices
    }

    va.vertices = append(va.vertices, v)
}

// Slicing for zero-copy subranges
func (va *VertexArray) Range(start, end int) []Vertex {
    return va.vertices[start:end] // Zero-copy slice
}
```

### Interface Design

AGG Go uses interfaces for pluggable components:

```go
// Renderer interface allows different rendering backends
type Renderer interface {
    RenderScanline(scanline Scanline, color Color)
    SetPixelFormat(format PixelFormat)
    Clear(color Color)
}

// Blender interface for different blend modes
type Blender interface {
    BlendPixel(dst, src Color, coverage uint8) Color
}

// Pixel format interface abstracts different color formats
type PixelFormat interface {
    BytesPerPixel() int
    SetPixel(x, y int, color Color)
    GetPixel(x, y int) Color
    BlendPixel(x, y int, color Color, coverage uint8)
}
```

### Error Handling

AGG Go follows Go's error handling conventions:

```go
func LoadImage(filename string) (*Image, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to open image file: %w", err)
    }
    defer file.Close()

    img, format, err := image.Decode(file)
    if err != nil {
        return nil, fmt.Errorf("failed to decode image (%s): %w", format, err)
    }

    return NewImageFromGoImage(img), nil
}

func (ctx *Context) SaveToPNG(filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    return png.Encode(file, ctx.image.ToGoImage())
}
```

### Concurrency Considerations

AGG Go rendering contexts are not thread-safe by design (following Go conventions), but can be used safely with proper synchronization:

```go
// Example: Parallel rendering of multiple regions
func ParallelRender(width, height, numWorkers int, drawFunc func(*Context)) *Image {
    result := NewImage(nil, width, height, width*4)

    var wg sync.WaitGroup
    rowsPerWorker := height / numWorkers

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)

        go func(workerID int) {
            defer wg.Done()

            startY := workerID * rowsPerWorker
            endY := startY + rowsPerWorker
            if workerID == numWorkers-1 {
                endY = height // Handle remainder
            }

            // Create separate context for this worker
            ctx := NewContext(width, endY-startY)
            ctx.SetViewport(0, startY, width, endY)

            drawFunc(ctx)

            // Copy worker result to final image (needs synchronization)
            mutex := &sync.Mutex{}
            mutex.Lock()
            result.CopyRegion(ctx.image, 0, startY, width, endY-startY)
            mutex.Unlock()
        }(i)
    }

    wg.Wait()
    return result
}
```

---

## Summary

This document covered the fundamental concepts of AGG adapted for Go:

1. **Anti-aliasing**: Coverage-based approach with subpixel accuracy
2. **Rendering Pipeline**: Six-stage process from paths to pixels
3. **Coordinates**: Standard 2D system with affine transformations
4. **Colors**: Type-safe color handling with gamma correction
5. **Scanlines**: Horizontal pixel processing for memory efficiency
6. **Coverage**: Analytical calculation methods for high-quality anti-aliasing
7. **Go Integration**: Memory management, interfaces, and concurrency patterns

These concepts form the foundation for understanding and effectively using the AGG Go port for high-quality 2D graphics rendering.
