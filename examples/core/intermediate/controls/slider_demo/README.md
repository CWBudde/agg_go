# Slider Control Demo

This example demonstrates the comprehensive slider control implementation for the AGG Go port. The slider control is a faithful port of AGG's `slider_ctrl_impl` and `slider_ctrl` classes from C++.

## Features Demonstrated

### 1. Basic Slider Configuration

- **Horizontal slider** with standard 0-100 range
- **Default colors** matching AGG's original design
- **Value formatting** with printf-style labels

### 2. Temperature Slider

- **Custom range** (-10°C to 40°C)
- **Custom colors** (red pointer for temperature)
- **Decimal precision** formatting

### 3. Volume Slider with Steps

- **Discrete steps** (0-10 with 11 positions)
- **Step quantization** - values snap to nearest step
- **Custom colors** (green theme for volume)

### 4. Percentage Slider

- **Normalized range** (0-1 displayed as percentage)
- **Custom background** and pointer colors
- **Progress indication** styling

### 5. Scientific Precision Slider

- **High precision** range (0.001-0.999)
- **Three decimal places** display
- **Gray text** styling for technical data

### 6. Descending Mode Slider

- **Visual triangle indicator** pointing left (descending = true)
- **Yellow triangle** color customization
- **Same value behavior** as normal sliders (descending only affects visuals)

## Key Implementation Features

### AGG-Compatible Design

- **6 rendering paths**: Background, Triangle, Text, Pointer Preview, Pointer, Step marks
- **RGBA color system**: Full floating-point color support with alpha
- **Vertex generation**: Complete path tessellation for all visual elements
- **Hit testing**: Accurate mouse interaction detection
- **Coordinate transformation**: Support for affine transformations

### Mouse Interaction

- **Drag handling**: Proper delta calculation for smooth dragging
- **Click detection**: Hit testing on pointer handle specifically
- **Preview values**: Real-time preview during drag operations
- **Value commitment**: Values update on mouse release

### Keyboard Navigation

- **Arrow key support**: Left/right and up/down navigation
- **Step-aware**: Respects discrete steps when configured
- **Consistent behavior**: Matches C++ AGG keyboard handling

### Rendering Architecture

- **Path-based rendering**: Each visual element is a separate path
- **Color customization**: Full control over all visual elements
- **Vertex streaming**: Efficient vertex generation for complex shapes
- **Step tick marks**: Automatic generation of step indicators

## Running the Demo

```bash
# From the project root
go run ./examples/controls/slider_demo/

# Or build and run
go build ./examples/controls/slider_demo/
./slider_demo
```

## Integration with AGG Rendering

This demo shows the control logic and vertex generation. In a complete AGG application, you would:

1. **Render each path** using AGG's rasterizer and renderer
2. **Handle mouse events** from your platform layer
3. **Apply transformations** for different coordinate systems
4. **Composite with other controls** in a GUI system

Example integration:

```go
slider := slider.NewSliderCtrl(x, y, x+width, y+height, flipY)

// Configure the slider
slider.SetRange(minVal, maxVal)
slider.SetValue(currentVal)
slider.SetLabel("Value: %.2f")

// In your render loop, for each path:
for pathID := uint(0); pathID < slider.NumPaths(); pathID++ {
    slider.Rewind(pathID)
    color := slider.Color(pathID)

    // Use AGG rasterizer to render the path
    rasterizer.Reset()
    rasterizer.AddPath(slider, pathID)
    renderScanlines(rasterizer, scanline, renderer, color)
}

// In your mouse event handler:
if slider.InRect(mouseX, mouseY) {
    if mouseDown {
        slider.OnMouseButtonDown(mouseX, mouseY)
    } else if dragging {
        slider.OnMouseMove(mouseX, mouseY, true)
    }
}
```

## C++ AGG Compatibility

This implementation maintains full compatibility with the original C++ AGG slider control:

- **Same API surface**: Method names and behavior match C++ equivalents
- **Identical rendering**: 6-path structure with same visual elements
- **Compatible colors**: Default colors match C++ AGG exactly
- **Same interaction model**: Mouse and keyboard handling identical to original

The slider control demonstrates the high fidelity of the Go AGG port and serves as a template for implementing other AGG controls.
