# Text Rendering in AGG2D Go Port

This document describes the text rendering capabilities of the AGG2D Go port, which closely follows the original C++ AGG2D text functionality.

## Overview

Phase 7 of the AGG2D implementation provides comprehensive text rendering support through integration with the FreeType font engine. The implementation includes:

- **Font Loading**: Support for TrueType and OpenType fonts
- **Text Rendering**: High-quality anti-aliased text rendering
- **Text Alignment**: Full control over horizontal and vertical text positioning
- **Font Metrics**: Accurate text measurement and positioning
- **Unicode Support**: Full Unicode text rendering support
- **Cache Management**: Efficient glyph caching for improved performance

## Build Requirements

Text rendering requires FreeType library support. To build with full text functionality:

```bash
# Install FreeType development headers (Linux)
sudo apt-get install libfreetype6-dev

# Or on macOS
brew install freetype

# Build with FreeType support
go build -tags freetype

# Without FreeType, text methods are available but return early
go build  # Regular build - limited text functionality
```

## Basic Usage

### Loading a Font

```go
agg2d := agg.NewAgg2D()

// Load a TrueType font
err := agg2d.Font("/path/to/font.ttf", 16.0, false, false, agg.RasterFontCache, 0.0)
if err != nil {
    log.Fatal(err)
}
```

### Rendering Text

```go
// Set text color
agg2d.FillColor(agg.Black)

// Set alignment
agg2d.TextAlignment(agg.AlignCenter, agg.AlignCenter)

// Render text
agg2d.Text(400, 300, "Hello, World!", false, 0, 0)
```

## API Reference

### Font Configuration

#### `Font(fileName string, height float64, bold, italic bool, cacheType FontCacheType, angle float64) error`
Loads and configures a font for text rendering.

- **fileName**: Path to the font file (TTF/OTF)
- **height**: Font size in points
- **bold**: Bold text flag (currently informational)
- **italic**: Italic text flag (currently informational) 
- **cacheType**: `RasterFontCache` or `VectorFontCache`
- **angle**: Text rotation angle in radians

#### `FontHeight() float64`
Returns the current font height in points.

#### `FlipText(flip bool)`
Sets whether to flip text rendering vertically.

### Text Rendering

#### `Text(x, y float64, str string, roundOff bool, dx, dy float64)`
Renders text at the specified position.

- **x, y**: Base position for text rendering
- **str**: Text string to render (supports Unicode)
- **roundOff**: Whether to round coordinates to pixel boundaries
- **dx, dy**: Additional offset for text positioning

#### `TextWidth(str string) float64`
Calculates the rendered width of a text string.

#### `TextAlignment(alignX, alignY TextAlignment)`
Sets text alignment for both horizontal and vertical positioning.

**Horizontal alignment:**
- `AlignLeft`: Align to the left of the position
- `AlignCenter`: Center on the position
- `AlignRight`: Align to the right of the position

**Vertical alignment:**
- `AlignBottom`: Align baseline to position
- `AlignCenter`: Center on the position  
- `AlignTop`: Align top to position

#### `TextHints(hints bool)` / `GetTextHints() bool`
Enables or disables font hinting for improved text rendering.

## Text Alignment Examples

```go
// Top-left aligned text
agg2d.TextAlignment(agg.AlignLeft, agg.AlignTop)
agg2d.Text(100, 100, "Top Left", false, 0, 0)

// Centered text
agg2d.TextAlignment(agg.AlignCenter, agg.AlignCenter)
agg2d.Text(400, 300, "Centered", false, 0, 0)

// Right-aligned, bottom-aligned text
agg2d.TextAlignment(agg.AlignRight, agg.AlignBottom)
agg2d.Text(700, 500, "Bottom Right", false, 0, 0)
```

## Font Cache Types

### Raster Font Cache (`RasterFontCache`)
- Caches pre-rendered bitmap glyphs
- Fast rendering performance
- Fixed to screen resolution
- Best for UI text and standard rendering

### Vector Font Cache (`VectorFontCache`)
- Caches vector outline data
- Resolution-independent
- Slower rendering but scalable
- Best for high-DPI displays or when scaling

## Unicode and International Text

The text system fully supports Unicode text rendering:

```go
// Various languages
agg2d.Text(100, 100, "Hello World", false, 0, 0)          // English
agg2d.Text(100, 120, "Bonjour le monde", false, 0, 0)     // French
agg2d.Text(100, 140, "こんにちは世界", false, 0, 0)              // Japanese
agg2d.Text(100, 160, "مرحبا بالعالم", false, 0, 0)          // Arabic
agg2d.Text(100, 180, "🌍 🎉 💖 🚀 🎨", false, 0, 0)        // Emoji
```

## Performance Considerations

### Glyph Caching
- Glyphs are automatically cached after first render
- Cache uses two-level indexing (MSB/LSB) for fast lookup
- Maximum of 32 font faces cached by default
- LRU eviction when cache is full

### Kerning Support
- Automatic kerning between character pairs
- Requires font with kerning tables
- Slight performance cost for character pair lookups

### Memory Usage
- Block allocator for efficient memory management
- Minimal memory fragmentation
- Glyph data shared across similar renderings

## Error Handling

### Without FreeType
When built without FreeType support:
- Font loading returns an error
- Text rendering methods return early (no-op)
- TextWidth() returns 0
- No crashes or panics occur

### Font Loading Errors
Common font loading issues:
- File not found
- Unsupported font format
- Corrupted font file
- Insufficient memory

### Graceful Degradation
- Missing characters render as empty space
- Invalid font faces fall back to default
- Out-of-memory conditions handled gracefully

## Implementation Details

### Architecture
The text rendering system consists of several components:

1. **Font Engine** (`internal/font/freetype/`): FreeType integration
2. **Cache Manager** (`internal/font/cache_manager.go`): Glyph caching
3. **Glyph Data** (`internal/font/glyph.go`): Glyph structures
4. **AGG2D Integration** (`agg2d_text.go`): High-level API

### Thread Safety
- Font engines are not thread-safe
- Each AGG2D context should be used from a single thread
- Multiple contexts can be used concurrently

### Build Tags
- `//go:build freetype`: Full FreeType implementation
- `//go:build !freetype`: Stub implementation for graceful degradation

## Examples

See `examples/text_rendering/main.go` for a comprehensive text rendering example that demonstrates:

- Font loading from common system locations
- All text alignment combinations
- Colored text rendering
- Text measurement and bounding boxes
- Unicode and international text
- Error handling without FreeType

## Testing

Comprehensive test coverage in `agg2d_text_test.go`:

- Text alignment verification
- Font parameter handling
- Text measurement accuracy
- Unicode text rendering
- Performance benchmarks
- Error condition handling

Run tests with:

```bash
go test -v ./...                    # Basic tests
go test -tags freetype -v ./...     # Tests with FreeType
go test -bench=. ./...              # Performance benchmarks
```

## Contributing

When contributing to text rendering:

1. Ensure compatibility with original C++ AGG2D behavior
2. Add comprehensive tests for new functionality
3. Handle FreeType unavailability gracefully
4. Document Unicode and international text considerations
5. Follow Go naming conventions and error handling patterns

## Future Enhancements

Potential improvements for future versions:

- **Text Layout**: Multi-line text with line breaks
- **Text Effects**: Shadows, outlines, gradients  
- **Advanced Typography**: Ligatures, complex scripts
- **Alternative Backends**: Native font engines (CoreText, DirectWrite)
- **Font Metrics**: Advanced typography metrics and baselines
- **Caching Improvements**: Persistent disk cache, better LRU algorithms