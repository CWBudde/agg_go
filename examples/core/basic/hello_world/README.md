# Hello World Example

This is the most basic AGG Go example that demonstrates fundamental library usage.

## What This Example Does

- Creates an 800x600 rendering context
- Clears the background with a light blue color
- Fills a red rectangle
- Draws a green stroked circle
- Fills a custom triangle using the explicit path API

## AGG Concepts Demonstrated

- **Context Creation**: Using `agg.NewContext()` to set up a rendering surface
- **Color Management**: Setting colors with `agg.RGB()` and predefined colors like `agg.Red`
- **Immediate-Mode Helpers**: Using `FillRectangle()` and `DrawCircle()` for common shapes
- **Path Construction**: Using `BeginPath()`, `MoveTo()`, `LineTo()`, `ClosePath()`, and `Fill()`
- **Image Output**: Rendering into the Context-owned backing image

## How to Run

```bash
just run-example basic/hello_world
```

## Expected Output

An interactive demo window showing:

- A red filled rectangle on a light blue background
- A green stroked circle
- A black filled triangle built from manual path commands

## Related Examples

- [shapes](../shapes/) - More comprehensive shape drawing
- [colors_rgba](../colors_rgba/) - Color system exploration
- [basic_demo](../basic_demo/) - Extended basic functionality
