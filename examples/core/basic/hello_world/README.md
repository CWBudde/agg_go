# Hello World Example

This is the most basic AGG Go example that demonstrates fundamental library usage.

## What This Example Does

- Creates an 800x600 rendering context
- Clears the background with a light blue color
- Draws a red rectangle
- Draws a green circle
- Outputs the rendered image data

## AGG Concepts Demonstrated

- **Context Creation**: Using `agg.NewContext()` to set up a rendering surface
- **Color Management**: Setting colors with `agg.RGB()` and predefined colors like `agg.Red`
- **Basic Shapes**: Drawing rectangles and circles with `DrawRectangle()` and `DrawCircle()`
- **Filling**: Using `Fill()` to render filled shapes
- **Image Output**: Getting the final rendered image with `GetImage()`

## How to Run

```bash
cd examples/core/basic/hello_world
go run main.go
```

## Expected Output

The program will print:

- Context creation confirmation
- Image dimensions and byte size
- Sample pixel values from the rendered image
- Success confirmation message

Since this is a console example, no image file is saved, but the rendered data is verified through pixel sampling.

## Related Examples

- [shapes](../shapes/) - More comprehensive shape drawing
- [colors_rgba](../colors_rgba/) - Color system exploration
- [basic_demo](../basic_demo/) - Extended basic functionality
