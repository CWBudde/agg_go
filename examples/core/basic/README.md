# Basic Examples

These are small, self-contained demos that exercise the very basics of the library. Each example renders a simple image or prints out simple diagnostics to verify the pipeline end-to-end. Keep expectations minimal: these are sanity checks and building blocks for richer examples.

## What’s Included

- hello_world: Creates a canvas, clears the background, draws a rectangle and a circle, and prints a few image stats.
- shapes: Draws a few ellipses and writes a simple PPM file (ellipse_test.ppm).
- colors_gray: Demonstrates grayscale buffers, conversions, gradients, and blending with ASCII output to the console.
- colors_rgba: Demonstrates RGBA colors, arithmetic, conversions, blending, and premultiplication with console output.
- rounded_rect: Renders a few ellipses and saves a PNG (rounded_rect_demo.png) to showcase a simple pipeline.
- lines: Draws basic lines (grid, diagonals, starburst) plus thick lines with various widths; saves `lines_demo.png`.
- embedded_fonts_hello: Renders "Hello World" using embedded bitmap fonts and saves `embedded_fonts_hello.png`.

## Running

- Single example: `just run-example basic/<name>` (e.g., `just run-example basic/hello_world`)
- All basic examples: `just run-examples-basic`

When running all, generated files are written to `examples/basic/_out/`:

- shapes → `ellipse_test.ppm`
- rounded_rect → `rounded_rect_demo.png`
- lines → `lines_demo.png`
- embedded_fonts_hello → `embedded_fonts_hello.png`

Other examples print diagnostics to the console.
