# Example Parity Ledger

This file tracks parity against the upstream AGG demo set.

Baseline:

- The workspace does not currently contain `../agg-2.4`.
- The available upstream source tree is `../agg-2.6/agg-src/examples`.
- That directory currently contains 72 top-level `*.cpp` files, including helper/generator utilities in addition to interactive demos.

## Ported Upstream Demos

Standalone examples in `examples/`:

- `aa_demo.cpp` -> `examples/core/intermediate/aa_demo`
- `aa_test.cpp` -> `examples/core/intermediate/aa_test`
- `agg2d_demo.cpp` -> `examples/core/intermediate/agg2d_demo`
- `bspline.cpp` -> `examples/core/intermediate/bspline`
- `circles.cpp` -> `examples/core/basic/circles`
- `compositing.cpp` -> `examples/core/intermediate/compositing`
- `conv_dash_marker.cpp` -> `examples/core/intermediate/conv_dash_marker`
- `gamma_correction.cpp` -> `examples/core/intermediate/controls/gamma_correction`
- `gouraud.cpp` -> `examples/core/intermediate/gouraud`
- `gradients.cpp` -> `examples/core/intermediate/gradients`
- `image_filters.cpp` -> `examples/core/intermediate/image_filters`
- `lion.cpp` -> `examples/core/intermediate/lion`
- `parse_lion.cpp` -> `internal/demo/lion` and `examples/core/intermediate/lion`
- `rasterizers.cpp` -> `examples/core/intermediate/rasterizers`
- `rasterizers2.cpp` -> `examples/core/intermediate/rasterizers2`
- `rounded_rect.cpp` -> `examples/core/basic/rounded_rect`
- `scanline_boolean.cpp` -> `examples/core/intermediate/scanline_boolean`

Custom Go examples not mapped 1:1 to a single upstream demo:

- `examples/core/basic/basic_demo`
- `examples/core/basic/colors_gray`
- `examples/core/basic/colors_rgba`
- `examples/core/basic/embedded_fonts_hello`
- `examples/core/basic/hello_world`
- `examples/core/basic/lines`
- `examples/core/basic/shapes`
- `examples/core/intermediate/text_rendering`
- `examples/core/intermediate/controls/rbox_demo`
- `examples/core/intermediate/controls/slider_demo`
- `examples/core/intermediate/controls/spline_demo`
- `examples/core/advanced/advanced_rendering`

## Remaining Missing Upstream Demos

Still missing as either a standalone example or web demo:

- `alpha_gradient.cpp`
- `alpha_mask.cpp`
- `alpha_mask2.cpp`
- `alpha_mask3.cpp`
- `bezier_div.cpp`
- `blend_color.cpp`
- `blur.cpp`
- `component_rendering.cpp`
- `compositing2.cpp`
- `conv_contour.cpp`
- `conv_stroke.cpp`
- `distortions.cpp`
- `flash_rasterizer.cpp`
- `flash_rasterizer2.cpp`
- `freetype_test.cpp`
- `gamma_ctrl.cpp`
- `gamma_tuner.cpp`
- `gouraud_mesh.cpp`
- `gpc_test.cpp`
- `gradient_focal.cpp`
- `gradients_contour.cpp`
- `graph_test.cpp`
- `idea.cpp`
- `image1.cpp`
- `image_alpha.cpp`
- `image_filters2.cpp`
- `image_fltr_graph.cpp`
- `image_perspective.cpp`
- `image_resample.cpp`
- `image_transforms.cpp`
- `interactive_polygon.cpp`
- `line_patterns.cpp`
- `line_patterns_clip.cpp`
- `line_thickness.cpp`
- `lion_lens.cpp`
- `lion_outline.cpp`
- `make_arrows.cpp`
- `make_gb_poly.cpp`
- `mol_view.cpp`
- `multi_clip.cpp`
- `pattern_fill.cpp`
- `pattern_perspective.cpp`
- `pattern_resample.cpp`
- `perspective.cpp`
- `polymorphic_renderer.cpp`
- `raster_text.cpp`
- `rasterizer_compound.cpp`
- `scanline_boolean2.cpp`
- `simple_blur.cpp`
- `trans_curve1.cpp`
- `trans_curve1_ft.cpp`
- `trans_curve2.cpp`
- `trans_curve2_ft.cpp`
- `trans_polar.cpp`
- `truetype_test.cpp`

## Next Priority

Recommended next parity batch, based on current library coverage:

1. `conv_stroke.cpp`
2. `conv_contour.cpp`
3. `gradient_focal.cpp`
4. `image_transforms.cpp`
5. `image_resample.cpp`
6. `raster_text.cpp`
7. `line_patterns.cpp`
8. `line_patterns.cpp`
