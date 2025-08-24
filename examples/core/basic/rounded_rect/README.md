# Rounded Rectangle Demo

This example demonstrates the AGG Go rendering pipeline using ellipses as a foundation for more complex shape rendering. While the original AGG `rounded_rect.cpp` example is fully interactive, this simplified version focuses on showcasing the core rendering capabilities.

## What it demonstrates

- **High-level API usage**: Using the `agg.NewContext()` API for simple graphics
- **Color management**: Setting different colors for rendering
- **Shape rendering**: Both filled and outlined ellipses
- **Image export**: Converting AGG's internal format to standard PNG

## Output

The example generates `rounded_rect_demo.png` showing:

- Blue filled ellipse
- Red outlined ellipse
- Green filled ellipse
- Purple outlined ellipse
- Orange filled ellipse

## Running the example

```bash
# From the project root
cd examples/basic/rounded_rect
go run main.go

# Or build first
go build .
./rounded_rect
```

## Next steps for full rounded rectangles

This example uses ellipses because the high-level Fill() function is not yet implemented. For true rounded rectangles with AGG-quality anti-aliasing, you would:

1. Use `internal/shapes/rounded_rect.go` which provides the full vertex source implementation
2. Use the low-level rendering pipeline with:
   - `internal/rasterizer/RasterizerScanlineAA`
   - `internal/scanline/ScanlineP8`
   - `internal/renderer/scanline/RendererScanlineAASolid`
   - `internal/renderer/scanline/RenderScanlines()`

## Implementation status

✅ **Available components:**

- Core rendering pipeline (rasterizer, scanline, renderer)
- RoundedRect vertex source in `internal/shapes/`
- ConvStroke for outlines
- Full pixel format support

❌ **Missing for interactive version:**

- UI controls (sliders, checkboxes)
- Mouse interaction handling
- Real-time parameter adjustment

The fundamental math and rendering is ready - the gaps are in interactive UI components.

## Related examples

- `../hello_world/` - Basic AGG usage
- `../shapes/` - More shape rendering examples
- See `../../../TASKS-EXAMPLES.md` for roadmap of all planned examples
