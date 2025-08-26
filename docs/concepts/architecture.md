# AGG Go Architecture Overview

This document provides a comprehensive overview of the Anti-Grain Geometry (AGG) Go port architecture, covering the rendering pipeline, data structures, interfaces, and performance characteristics.

## Table of Contents

- [Complete Rendering Pipeline Flow](#complete-rendering-pipeline-flow)
- [Data Structures Relationships](#data-structures-relationships)
- [Interface Hierarchy and Composition Patterns](#interface-hierarchy-and-composition-patterns)
- [Performance Characteristics and Trade-offs](#performance-characteristics-and-trade-offs)
- [Memory Management Patterns](#memory-management-patterns)

---

## Complete Rendering Pipeline Flow

### Pipeline Stages Diagram

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ Path Definition │───▶│  Transformation  │───▶│   Conversion    │
│   (types.go)    │    │ (transform pkg)  │    │   (conv pkg)    │
│                 │    │                  │    │                 │
│ • MoveTo/LineTo │    │ • Affine matrix  │    │ • Stroke gen    │
│ • CurveTo       │    │ • Scale/Rotate   │    │ • Dash pattern  │
│ • ClosePath     │    │ • Translate      │    │ • Contour gen   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Rasterization  │◀───│  Scanline Gen.   │◀───│ Pixel Rendering │
│ (rasterizer)    │    │  (scanline pkg)  │    │ (renderer pkg)  │
│                 │    │                  │    │                 │
│ • Vector→pixels │    │ • Horizontal     │    │ • Color blend   │
│ • Coverage calc │    │   strips         │    │ • Alpha comp    │
│ • Anti-aliasing │    │ • Span storage   │    │ • Final output  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Detailed Flow Description

#### 1. **Path Definition** (`types.go`, `context.go`)

```go
// User builds paths using high-level API
ctx := agg.NewContext(800, 600)
ctx.BeginPath()
ctx.MoveTo(100, 100)
ctx.LineTo(200, 150)
ctx.CurveTo(250, 100, 300, 150, 350, 100)
ctx.ClosePath()

// Internally stored as vertex sequences
type PathStorage struct {
    vertices []basics.Vertex
    commands []PathCommand
}
```

#### 2. **Transformation** (`internal/transform/`)

```go
// Affine transformations applied to all coordinates
type Transform struct {
    sx, shx, shy, sy, tx, ty float64  // 2x3 matrix elements
}

func (t *Transform) Apply(x, y float64) (float64, float64) {
    return t.sx*x + t.shx*y + t.tx,
           t.shy*x + t.sy*y + t.ty
}
```

#### 3. **Conversion** (`internal/conv/`)

```go
// Path converters modify geometry before rasterization
type Stroke struct {
    width     float64
    lineJoin  LineJoin
    lineCap   LineCap
    miterLimit float64
}

// Converts path to outlined strokes
func (s *Stroke) Convert(path *PathStorage) *PathStorage {
    // Complex stroke generation algorithm
    // Handles joins, caps, and width expansion
}
```

#### 4. **Rasterization** (`internal/rasterizer/`)

```go
// Core rasterization converts vector paths to coverage data
type RasterizerScanlineAA[Clip ClipInterface, Conv ConverterInterface] struct {
    cells      []Cell
    outline    *OutlineAA
    clipper    Clip
    converter  Conv
}

type Cell struct {
    x       int    // Cell X coordinate
    y       int    // Cell Y coordinate
    cover   int32  // Coverage accumulation
    area    int32  // Area accumulation
}
```

#### 5. **Scanline Generation** (`internal/scanline/`)

```go
// Converts cells to horizontal spans for rendering
type ScanlineU struct {
    y        int
    minX     int
    maxX     int
    coverage []uint8    // Coverage array
}

type Span struct {
    x        int      // Starting X
    length   int      // Span length
    coverage []uint8  // Per-pixel coverage
}
```

#### 6. **Pixel Rendering** (`internal/renderer/`, `internal/pixfmt/`)

```go
// Final pixel output with blending
type RendererScanline[PixFmt PixelFormat] struct {
    pixfmt PixFmt
}

func (r *RendererScanline[PixFmt]) RenderScanline(sl Scanline, color Color) {
    // Apply color and alpha blending to each span
    for _, span := range sl.Spans() {
        r.pixfmt.BlendHorizontalSpan(span.X, sl.Y(), span.Length,
                                     color, span.Coverage)
    }
}
```

---

## Data Structures Relationships

### Core Data Structure Hierarchy

```
                    ┌─────────────┐
                    │   Context   │ (Public API)
                    │  (agg2d.go) │
                    └─────┬───────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │ Internal AGG2D  │ (Implementation)
                 │(internal/agg2d/)│
                 └─────┬───────────┘
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
┌──────────────┐ ┌─────────────┐ ┌──────────────┐
│ PathStorage  │ │ Rasterizer  │ │ Renderer     │
│   (path)     │ │(rasterizer) │ │ (renderer)   │
└─────┬────────┘ └─────┬───────┘ └──────┬───────┘
      │                │                │
      ▼                ▼                ▼
┌──────────────┐ ┌─────────────┐ ┌──────────────┐
│ VertexArray  │ │ ScanlineU   │ │ PixelFormat  │
│   (array)    │ │ (scanline)  │ │  (pixfmt)    │
└──────────────┘ └─────────────┘ └──────────────┘
```

### Container Relationships

```go
// Array containers form the foundation
type PodArray[T any] struct {
    data []T          // Go slice as backing store
    size int          // Current size
}

// Specialized arrays for different data types
type VertexBlockStorage struct {
    vertices    PodArray[basics.Vertex]
    commands    PodArray[PathCommand]
    blockSize   int
}

// Path storage builds on vertex arrays
type PathStorage struct {
    storage     *VertexBlockStorage
    totalSize   int
    iterator    int
}
```

### Generic Type Relationships

```go
// Generic hierarchy enables type-safe composition
type PixelFormat interface {
    BlendPixel(x, y int, color Color, cover uint8)
}

type PixFmtRGBA32[Blender any] struct {
    buffer   *buffer.RenderingBuffer
    blender  Blender
}

// Composed types maintain type safety
type PixFmtAlphaBlendRGBA[Blender BlenderInterface, ColorSpace ColorSpaceInterface] struct {
    buffer     *buffer.RenderingBuffer
    blender    Blender
    colorSpace ColorSpace
}
```

---

## Interface Hierarchy and Composition Patterns

### Core Interface Hierarchy

```go
// Base interfaces define minimal contracts
type ArrayInterface[T any] interface {
    Size() int
    At(i int) T
    Data() []T
}

// Extended interfaces add mutation capabilities
type MutableArrayInterface[T any] interface {
    ArrayInterface[T]
    Set(i int, val T)
}

// Growable interfaces add dynamic sizing
type GrowableArrayInterface[T any] interface {
    MutableArrayInterface[T]
    PushBack(val T)
    RemoveLast()
    Resize(newSize int)
}
```

### Converter Interface Composition

```go
// Generic converter interface
type ConverterInterface interface {
    Rewind(pathID int)
    Vertex(x, y *float64) int
}

// Clip interface for geometric clipping
type ClipInterface interface {
    ConverterInterface
    ClipBox(x1, y1, x2, y2 float64)
}

// Composition pattern used in rasterizer
type RasterizerScanlineAA[Clip ClipInterface, Conv ConverterInterface] struct {
    clipper   Clip      // Geometric clipping
    converter Conv      // Path conversion
    outline   *OutlineAA
}
```

### Renderer Interface Patterns

```go
// Base renderer interface
type RendererBase[PixFmt PixelFormat] interface {
    AttachPixFmt(pixfmt PixFmt)
    RenderScanline(sl Scanline, color Color)
    Clear(color Color)
}

// Anti-aliased outline renderer
type OutlineAARenderer[PixFmt PixelFormat] struct {
    pixfmt    PixFmt
    primitive OutlinePrimitive
}

// Solid scanline renderer
type RendererScanlineSolid[PixFmt PixelFormat] struct {
    pixfmt PixFmt
    color  Color
}
```

### Blender Strategy Pattern

```go
// Blender interface enables different blend modes
type BlenderInterface interface {
    BlendPix(p *Color, cr, cg, cb, alpha, cover uint)
}

// Specific blender implementations
type BlenderRGBA8 struct{}
type BlenderRGBA8Pre struct{}

// Pixel formats compose with blenders
type PixFmtRGBA32[Blender BlenderInterface] struct {
    buffer  *buffer.RenderingBuffer
    blender Blender
}
```

### Interface Usage Examples

```go
// Polymorphic rendering through interfaces
func RenderPath[R RendererBase[PF], PF PixelFormat](
    renderer R, path *PathStorage, color Color) {

    rasterizer := NewRasterizerScanlineAA[NoClip, PathConverter]()
    scanline := NewScanlineU()

    rasterizer.AddPath(path, 0)

    for scanline.Reset(); rasterizer.SweepScanline(scanline); {
        renderer.RenderScanline(scanline, color)
    }
}

// Type-safe composition
func NewRGBARenderer() *OutlineAARenderer[PixFmtRGBA32[BlenderRGBA8]] {
    pixfmt := NewPixFmtRGBA32(buffer, BlenderRGBA8{})
    return NewOutlineAARenderer(pixfmt)
}
```

---

## Performance Characteristics and Trade-offs

### Go vs C++ Performance Profile

| **Aspect**            | **C++ AGG**              | **Go AGG**              | **Trade-off**                          |
| --------------------- | ------------------------ | ----------------------- | -------------------------------------- |
| **Memory Management** | Manual (new/delete)      | Garbage Collected       | GC pauses vs memory safety             |
| **Type Safety**       | Templates (compile-time) | Generics (compile-time) | Similar performance, better ergonomics |
| **Virtual Calls**     | vtable overhead          | Interface dispatch      | Similar overhead, better composability |
| **Array Access**      | Raw pointers             | Slice bounds checking   | Safety vs minimal overhead             |
| **Memory Layout**     | Custom allocators        | Go runtime managed      | Less control, better safety            |

### Hot Path Optimizations

```go
// Critical path: rasterizer cell processing
func (ras *RasterizerScanlineAA) AddCurve(x1, y1, x2, y2, x3, y3 float64) {
    // Minimize allocations in hot path
    if ras.cellPool == nil {
        ras.cellPool = make([]Cell, 0, 256) // Pre-allocate
    }

    // Reuse slices to avoid GC pressure
    ras.cellPool = ras.cellPool[:0] // Reset length, keep capacity

    // Inline critical calculations
    subdivisionLimit := 1.0 / float64(ras.gammaValue)

    // Avoid function calls in inner loops
    for i := 0; i < len(ras.cells); i++ {
        cell := &ras.cells[i]  // Direct access, no bounds check needed
        // ... cell processing
    }
}
```

### Memory Allocation Patterns

```go
// Efficient buffer reuse pattern
type BufferPool struct {
    buffers [][]uint8
    sizes   []int
    mu      sync.Mutex
}

func (p *BufferPool) GetBuffer(size int) []uint8 {
    p.mu.Lock()
    defer p.mu.Unlock()

    // Find suitable buffer
    for i, s := range p.sizes {
        if s >= size {
            buffer := p.buffers[i]
            // Remove from pool
            p.buffers = append(p.buffers[:i], p.buffers[i+1:]...)
            p.sizes = append(p.sizes[:i], p.sizes[i+1:]...)
            return buffer[:size]
        }
    }

    // Create new buffer if none suitable
    return make([]uint8, size)
}
```

### Garbage Collection Impact

```go
// GC-friendly patterns used throughout AGG Go
type Scanline struct {
    y        int
    spans    []Span     // Pre-allocated slice
    spanPool []Span     // Reuse pool
}

func (sl *Scanline) Reset() {
    // Reset length but keep capacity
    sl.spans = sl.spans[:0]
    // Don't set to nil - keeps memory allocated
}

// Avoid creating garbage in hot paths
func (ras *RasterizerScanlineAA) SweepScanline(sl *Scanline) bool {
    // Reuse existing slices rather than allocate new ones
    sl.Reset() // Resets length, keeps capacity

    if ras.scanY > ras.maxY {
        return false
    }

    // Build spans in pre-allocated slice
    ras.buildSpans(sl)
    return true
}
```

### Performance Benchmarks

```go
// Typical performance characteristics
const (
    // Memory overhead per operation
    VertexSize      = 16 bytes  // vs 8 bytes in C++ (slice header)
    CellSize        = 16 bytes  // vs 12 bytes in C++ (alignment)
    SliceOverhead   = 24 bytes  // vs 8 bytes for raw pointer

    // Time complexity
    RasterizationComplexity = "O(n×k)" // n = vertices, k = coverage
    ScanlineGenComplexity   = "O(w×h)" // w = width, h = affected height
    BlendingComplexity      = "O(p)"   // p = pixels touched
)
```

---

## Memory Management Patterns

### Go Garbage Collection vs C++ Manual Management

#### C++ AGG Pattern

```cpp
// C++ manual memory management
template<class T> class pod_array {
    T* data;
    unsigned size;
    unsigned capacity;

public:
    pod_array() : data(nullptr), size(0), capacity(0) {}

    ~pod_array() {
        delete[] data;  // Manual cleanup
    }

    void resize(unsigned new_size) {
        if (new_size > capacity) {
            T* new_data = new T[new_size];  // Manual allocation
            std::memcpy(new_data, data, size * sizeof(T));
            delete[] data;  // Manual cleanup
            data = new_data;
            capacity = new_size;
        }
        size = new_size;
    }
};
```

#### Go AGG Pattern

```go
// Go garbage-collected pattern
type PodArray[T any] struct {
    data []T  // Go slice - automatically managed
    size int
}

func (pa *PodArray[T]) Resize(newSize int) {
    if newSize > cap(pa.data) {
        // Go runtime handles allocation/copying
        newData := make([]T, newSize, newSize*2) // Growth strategy
        copy(newData, pa.data)
        pa.data = newData  // Old data eligible for GC
    }
    pa.data = pa.data[:newSize]
    pa.size = newSize
}
```

### Slice Management Best Practices

```go
// Efficient slice growth pattern
type VertexArray struct {
    vertices []basics.Vertex
}

func (va *VertexArray) PushBack(vertex basics.Vertex) {
    // Go's built-in append handles growth efficiently
    va.vertices = append(va.vertices, vertex)
}

// Reuse pattern for hot paths
type ScanlinePool struct {
    pool []ScanlineU
    idx  int
}

func (sp *ScanlinePool) Get() *ScanlineU {
    if sp.idx >= len(sp.pool) {
        sp.pool = append(sp.pool, ScanlineU{})
    }

    sl := &sp.pool[sp.idx]
    sp.idx++
    sl.Reset() // Clear but don't deallocate
    return sl
}

func (sp *ScanlinePool) Put(sl *ScanlineU) {
    if sp.idx > 0 {
        sp.idx--  // Return to pool
    }
}
```

### Buffer Management Strategies

```go
// Rendering buffer with controlled allocation
type RenderingBuffer struct {
    buffer    []uint8     // Main buffer
    width     int
    height    int
    stride    int         // Bytes per row
    rowPtrs   [][]uint8   // Cached row pointers
}

func NewRenderingBuffer(width, height, bytesPerPixel int) *RenderingBuffer {
    stride := width * bytesPerPixel

    // Single allocation for main buffer
    buffer := make([]uint8, height*stride)

    // Pre-calculate row pointers to avoid arithmetic in hot paths
    rowPtrs := make([][]uint8, height)
    for y := 0; y < height; y++ {
        start := y * stride
        end := start + stride
        rowPtrs[y] = buffer[start:end:end] // Fixed capacity slice
    }

    return &RenderingBuffer{
        buffer:  buffer,
        width:   width,
        height:  height,
        stride:  stride,
        rowPtrs: rowPtrs,
    }
}

// Fast row access without bounds checking (slice is pre-bounded)
func (rb *RenderingBuffer) Row(y int) []uint8 {
    return rb.rowPtrs[y] // No bounds check needed
}
```

### Memory Reuse Patterns

```go
// Pattern: Reuse expensive objects
type RasterizationContext struct {
    rasterizer  *RasterizerScanlineAA
    scanline    *ScanlineU
    cellPool    []Cell
    spanPool    []Span
}

func (ctx *RasterizationContext) RenderPath(path *PathStorage) {
    // Reuse rasterizer state
    ctx.rasterizer.Reset()
    ctx.scanline.Reset()

    // Add path to rasterizer
    ctx.rasterizer.AddPath(path, 0)

    // Render using pre-allocated scanline
    for ctx.rasterizer.SweepScanline(ctx.scanline) {
        // Process scanline...
    }
}

// Object pooling for frequently allocated structures
type CellPool struct {
    cells [][]Cell
    index int
}

func (cp *CellPool) GetCells(size int) []Cell {
    if cp.index >= len(cp.cells) {
        cp.cells = append(cp.cells, make([]Cell, size))
    }

    cells := cp.cells[cp.index]
    cp.index++

    // Reset cells but keep allocation
    for i := range cells {
        cells[i] = Cell{} // Zero value
    }

    return cells
}
```

### Concurrency and Memory

```go
// Thread-local context pattern for parallel rendering
type RenderingWorker struct {
    id       int
    context  *RasterizationContext  // Worker-owned context
    buffer   *RenderingBuffer       // Shared, read-only after init
    region   image.Rectangle        // Worker's area
}

func (w *RenderingWorker) RenderRegion(paths []*PathStorage) {
    // No synchronization needed - each worker has own context
    for _, path := range paths {
        w.context.RenderPath(path)
    }
}

// Synchronization only needed for final composition
func ParallelRender(paths []*PathStorage, numWorkers int) *RenderingBuffer {
    workers := make([]*RenderingWorker, numWorkers)

    // Each worker gets independent memory context
    for i := range workers {
        workers[i] = &RenderingWorker{
            id:      i,
            context: NewRasterizationContext(),
            region:  calculateWorkerRegion(i, numWorkers),
        }
    }

    // Workers run independently
    var wg sync.WaitGroup
    for _, worker := range workers {
        wg.Add(1)
        go func(w *RenderingWorker) {
            defer wg.Done()
            w.RenderRegion(paths)
        }(worker)
    }

    wg.Wait()
    return compositeResults(workers) // Synchronization point
}
```

---

## Summary

The AGG Go port architecture successfully adapts C++ AGG's proven design to Go's programming model:

1. **Pipeline Fidelity**: Maintains the original six-stage rendering pipeline
2. **Type Safety**: Uses Go generics to replace C++ templates with compile-time type safety
3. **Interface Design**: Employs composition over inheritance for flexible, testable components
4. **Performance**: Achieves comparable performance through careful memory management and GC-aware patterns
5. **Memory Management**: Leverages Go's GC while minimizing allocation pressure through object reuse
6. **Concurrency**: Supports parallel rendering through thread-local contexts and minimal synchronization

The architecture demonstrates how to successfully port a complex C++ graphics library to Go while preserving performance characteristics and improving type safety and memory safety.
