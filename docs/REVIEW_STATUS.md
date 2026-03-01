# AGG 2.6 Go Port - File Checklist

This is a comprehensive checklist of files that need to be ported from the original AGG 2.6 C++ codebase to Go. Please always check the original C++ implementation for reference in ../agg-2.6

## Core Header Files (include/)

### Basic Types and Configuration

#### agg_basics.h - Core types, enums, path commands, geometry utilities

Go files:

- internal/basics/types.go
- internal/basics/constants.go
- internal/basics/path.go
- internal/basics/math.go

**Allocators (Templates → Go Generics)**

- [x] pod_allocator<T> → Generic allocator interface
- [x] obj_allocator<T> → Object allocator with constructors

**Basic Types**

- [x] int8, int8u, int16, int16u, int32, int32u, int64, int64u type definitions
- [x] cover_type (unsigned char)
- [x] Enums: cover_scale_e, poly_subpixel_scale_e, filling_rule_e
- [x] Path command enums: path_commands_e, path_flags_e

**Rounding Functions**

- [x] iround(), uround(), ifloor(), ufloor(), iceil(), uceil()
- [x] Platform-specific optimizations (FISTP, QIFIST)

**Template Structs → Go Generics**

- [x] saturation<Limit> → Saturation[T] with limit parameter
- [x] mul_one<Shift> → MulOne with shift parameter
- [x] rect_base<T> → Rect[T] generic struct
- [x] point_base<T> → Point[T] generic struct
- [x] vertex_base<T> → Vertex[T] generic struct
- [x] row_info<T> → RowInfo[T] generic struct
- [x] const_row_info<T> → ConstRowInfo[T] generic struct

**Geometry Functions**

- [x] intersect_rectangles(), unite_rectangles()
- [x] is_equal_eps() epsilon comparison

**Path Utility Functions**

- [x] is_vertex(), is_drawing(), is_stop(), is_move_to()
- [x] is_line_to(), is_curve(), is_curve3(), is_curve4()
- [x] is_end_poly(), is_close(), is_next_poly()
- [x] is_cw(), is_ccw(), is_oriented(), is_closed()
- [x] get_close_flag(), clear_orientation(), get_orientation(), set_orientation()

**Constants**

- [x] pi constant
- [x] deg2rad(), rad2deg() conversions

#### agg_config.h - Configuration definitions

Go files:

- internal/config/config.go

- [x] Configuration constants (mostly compile-time in C++)
- [x] Type overrides mechanism for Go

#### agg_array.h - Dynamic array implementation

Go files:

- internal/array/interfaces.go
- internal/array/algorithms.go
- internal/array/comparators.go
- internal/array/pod_arrays.go
- internal/array/pod_bvector.go
- internal/array/block_allocator.go

**POD Array Types (Templates → Go Generics)**

- [x] pod_array_adaptor<T> → PodArrayAdaptor[T]
- [x] pod_auto_array<T, Size> → PodAutoArray[T] with size
- [x] pod_auto_vector<T, Size> → PodAutoVector[T] with size
- [x] pod_array<T> → PodArray[T] dynamic array
- [x] pod_vector<T> → PodVector[T] growable vector
- [x] pod_bvector<T, S> → PodBVector[T] block vector

**Block Allocator**

- [x] block_allocator class → BlockAllocator struct
- [x] allocate() with alignment support
- [x] block management

**Algorithms (Templates → Go Generics)**

- [x] quick_sort<Array, Less> → QuickSort[T] with comparator
- [x] swap_elements<T> → SwapElements[T]
- [x] remove_duplicates<Array, Equal> → RemoveDuplicates[T]
- [x] invert_container<Array> → InvertContainer[T]
- [x] binary_search_pos<Array, Value, Less> → BinarySearchPos[T]
- [x] range_adaptor<Array> → RangeAdaptor[T]

**Comparison Functions**

- [x] int_less(), int_greater()
- [x] unsigned_less(), unsigned_greater()

#### agg_math.h - Mathematical functions and constants

Go files:

- internal/basics/math.go
- internal/basics/constants.go

**Constants**

- [x] vertex_dist_epsilon
- [x] intersection_epsilon

**Geometric Calculations**

- [x] cross_product()
- [x] point_in_triangle()
- [x] calc_distance()
- [x] calc_sq_distance()
- [x] calc_line_point_distance()
- [x] calc_segment_point_u()
- [x] calc_segment_point_sq_distance() (2 overloads)
- [x] calc_intersection()
- [x] intersection_exists()
- [x] calc_orthogonal()
- [x] dilate_triangle()
- [x] calc_triangle_area()
- [x] calc_polygon_area<Storage>() → CalcPolygonArea[T]()

**Fast Math**

- [x] fast_sqrt() with lookup tables
- [x] g_sqrt_table[1024] lookup table
- [x] g_elder_bit_table[256] lookup table
- [x] besj() Bessel function

---

### Color and Pixel Formats

#### agg_color_gray.h - Grayscale color handling

Go files:

- internal/color/gray.go
- internal/color/conversion.go

**Template Types → Go Generics**

- [x] gray8T<Colorspace> → Gray8[CS] generic struct
- [x] gray16T<Colorspace> → Gray16[CS] generic struct
- [x] gray32T<Colorspace> → Gray32[CS] generic struct

**Core Gray8 Methods**

- [x] luminance(rgba) - ITU-R BT.709 calculation
- [x] luminance(rgba8) - Optimized 8-bit version
- [x] convert() methods between colorspaces
- [x] convert() from/to RGBA types
- [x] convert_from_sRGB() → ConvertFromSRGB()
- [x] convert_to_sRGB() → ConvertToSRGB()
- [x] make_rgba8(), make_srgba8(), make_rgba16(), make_rgba32()
- [x] Constructors and operators

**Gray16 and Gray32 Variants**

- [x] gray16 type with 16-bit precision
- [x] gray32 type with 32-bit precision
- [x] Conversion methods for each precision

#### agg_color_rgba.h - RGBA color handling

Go files:

- internal/color/rgba.go
- internal/color/rgb.go
- internal/color/conversion.go

**Order Structs (Component Ordering)**

- [x] order_rgb → OrderRGB constants
- [x] order_bgr → OrderBGR constants
- [x] order_rgba → OrderRGBA constants
- [x] order_argb → OrderARGB constants
- [x] order_abgr → OrderABGR constants
- [x] order_bgra → OrderBGRA constants

**Colorspace Tags**

- [x] linear struct → Linear type tag
- [x] sRGB struct → SRGB type tag

**Base RGBA Type (float64)**

- [x] rgba struct → RGBA base type
- [x] clear(), transparent(), opacity() methods
- [x] premultiply(), demultiply() methods
- [x] gradient() interpolation
- [x] Operators: +=, _=, +, _
- [x] no_color() static method
- [x] from_wavelength() static method

**Template Types → Go Generics**

- [x] rgba8T<Colorspace> → RGBA8[CS] generic struct
- [x] rgba16T<Colorspace> → RGBA16[CS] generic struct
- [x] rgba32T<Colorspace> → RGBA32[CS] generic struct

**RGBA8 Core Methods**

- [x] convert() between colorspaces (linear ↔ sRGB)
- [x] convert() to/from float rgba
- [x] premultiply(), demultiply() operations
- [x] gradient() interpolation
- [x] clear(), transparent() methods
- [x] add(), subtract(), multiply() blend operations
- [x] apply_gamma_dir(), apply_gamma_inv()

**RGBA16 and RGBA32 Variants**

- [x] 16-bit and 32-bit precision versions
- [x] Corresponding conversion methods

**Helper Functions**

- [x] rgba_pre() → RGBAPre() premultiplied constructor
- [x] rgb_conv_rgba8() → RGBConvRGBA8()
- [x] rgb_conv_rgba16() → RGBConvRGBA16()

**sRGB Conversion Tables**

- [x] sRGB_conv<T> → SRGBConv[T] conversion utilities
- [x] Lookup tables for sRGB ↔ linear conversion

#### agg_pixfmt_base.h - Base pixel format definitions

Go files:

- internal/pixfmt/base.go

**Pixel Format Tags**

- [x] pixfmt_gray_tag → PixFmtGrayTag
- [x] pixfmt_rgb_tag → PixFmtRGBTag
- [x] pixfmt_rgba_tag → PixFmtRGBATag

**Base Blender Template → Go Generic**

- [x] blender_base<ColorT, Order> → BlenderBase[C, O]
- [x] get() methods for pixel extraction
- [x] set() methods for pixel setting

#### agg_pixfmt_gray.h - Grayscale pixel formats

Go files:

- internal/pixfmt/pixfmt_gray.go
- internal/pixfmt/pixfmt_gray16.go
- internal/pixfmt/pixfmt_gray32.go
- internal/pixfmt/blender_gray.go
- internal/pixfmt/blender_gray16.go
- internal/pixfmt/blender_gray32.go
- internal/pixfmt/gamma_gray.go

**Blender Types**

- [x] blender_gray<ColorT> → BlenderGray[C]
- [x] blender_gray_pre<ColorT> → BlenderGrayPre[C]
- [x] blend_pix() methods for both

**Gamma Application**

- [x] apply_gamma_dir_gray<ColorT, GammaLut> → ApplyGammaDirGray[C]
- [x] apply_gamma_inv_gray<ColorT, GammaLut> → ApplyGammaInvGray[C]

**Main Pixel Format Template**

- [x] pixfmt_alpha_blend_gray<Blender, RenBuf> → PixFmtAlphaBlendGray[B]
- [x] Core pixel operations (copy_pixel, blend_pixel, etc.)
- [x] Span operations (copy_hline, blend_hline, etc.)
- [x] copy_from() for buffer copying

**Concrete Types**

- [x] pixfmt_gray8 → PixFmtGray8
- [x] pixfmt_sgray8 → PixFmtSGray8
- [x] pixfmt_gray16 → PixFmtGray16
- [x] pixfmt_gray32 → PixFmtGray32

#### agg_pixfmt_rgb.h - RGB pixel formats

Go files:

- internal/pixfmt/pixfmt_rgb.go
- internal/pixfmt/blender_rgb.go

**Gamma Application Classes**

- [x] apply_gamma_dir_rgb<ColorT, Order, GammaLut>
- [x] apply_gamma_inv_rgb<ColorT, Order, GammaLut>

**Blender Types**

- [x] blender_rgb<ColorT, Order> → BlenderRGB[C, O]
- [x] blender_rgb_pre<ColorT, Order> → BlenderRGBPre[C, O]
- [x] blender_rgb_gamma<ColorT, Order, Gamma> → BlenderRGBGamma[C, O]

**Main Pixel Format Template**

- [x] pixfmt_alpha_blend_rgb<Blender, RenBuf, Step, Offset>
- [x] pixel_type nested struct → RGBPixelType
- [x] row_data(), make_pix(), copy_pixel(), blend_pixel()
- [x] Hline operations (copy_hline, blend_hline, etc.)
- [x] Solid color operations (fill, blend*solid*\*)
- [x] copy_from(), blend_from() for compositing

**Concrete RGB24 Types**

- [x] pixfmt_rgb24 → PixFmtRGB24
- [x] pixfmt_bgr24 → PixFmtBGR24
- [x] pixfmt_srgb24 → PixFmtSRGB24
- [x] pixfmt_sbgr24 → PixFmtSBGR24

**RGB48 Types (16-bit per channel)**

- [x] pixfmt_rgb48 → PixFmtRGB48
- [x] pixfmt_bgr48 → PixFmtBGR48

**Gamma Variants**

- [x] pixfmt_rgb24_gamma<Gamma> → PixFmtRGB24Gamma[G]
- [x] Similar for all RGB formats

#### agg_pixfmt_rgb_packed.h - Packed RGB pixel formats

**Packed Formats (555, 565, etc.)**

- [x] pixfmt_rgb555 → PixFmtRGB555
- [x] pixfmt_rgb565 → PixFmtRGB565
- [x] pixfmt_bgr555 → PixFmtBGR555
- [x] pixfmt_bgr565 → PixFmtBGR565
- [x] Packing/unpacking utilities
- [x] Bit-shift operations for packed formats

#### agg_pixfmt_rgba.h - RGBA pixel formats

Go files:

- internal/pixfmt/pixfmt_rgba.go
- internal/pixfmt/blender_rgba.go
- internal/pixfmt/gamma_rgba.go

**Blender Types**

- [x] blender_rgba<ColorT, Order> → BlenderRGBA[C, O]
- [x] blender_rgba_pre<ColorT, Order> → BlenderRGBAPre[C, O]
- [x] blender_rgba_plain<ColorT, Order> → BlenderRGBAPlain[C, O]
- [x] Composite blenders (multiply, screen, overlay, etc.)

**Main RGBA Pixel Format**

- [x] pixfmt_alpha_blend_rgba<Blender, RenBuf>
- [x] Full alpha channel support
- [x] Premultiplied/non-premultiplied operations

**Concrete RGBA32 Types**

- [x] pixfmt_rgba32 → PixFmtRGBA32
- [x] pixfmt_argb32 → PixFmtARGB32
- [x] pixfmt_abgr32 → PixFmtABGR32
- [x] pixfmt_bgra32 → PixFmtBGRA32

**RGBA64 Types (16-bit per channel)**

- [x] pixfmt_rgba64 → PixFmtRGBA64
- [x] pixfmt_argb64 → PixFmtARGB64
- [x] Similar variants

#### agg_pixfmt_transposer.h - Pixel format transposer

**Transposer Wrapper**

- [x] pixfmt_transposer<PixFmt> → PixFmtTransposer[P]
- [x] Transposes x/y coordinates
- [x] Wraps another pixel format

#### agg_pixfmt_amask_adaptor.h - Alpha mask adaptor

**Alpha Mask Adaptor**

- [x] pixfmt_amask_adaptor<PixFmt, AlphaMask> → PixFmtAMaskAdaptor[P, A]
- [x] Applies alpha mask to pixel format operations
- [x] combine_pixel() with mask

---

### Scanlines

#### agg_scanline_bin.h

Go files:

- internal/scanline/scanline_bin.go

**scanline_bin class:**

- [x] span struct (x, len members)
- [x] coord_type typedef
- [x] const_iterator typedef
- [x] Default constructor
- [x] reset(min_x, max_x) method
- [x] add_cell(x, cover) method
- [x] add_span(x, len, cover) method
- [x] add_cells(x, len, covers) method
- [x] finalize(y) method
- [x] reset_spans() method
- [x] y() accessor method
- [x] num_spans() accessor method
- [x] begin() accessor method
- [x] Private members (m_last_x, m_y, m_spans, m_cur_span)
- [x] Copy constructor and assignment operator (prohibited)

**scanline32_bin class:**

- [x] span struct with constructors
- [x] coord_type typedef
- [x] span_array_type typedef
- [x] const_iterator nested class
- [x] Default constructor
- [x] reset(min_x, max_x) method
- [x] add_cell(x, cover) method
- [x] add_span(x, len, cover) method
- [x] add_cells(x, len, covers) method
- [x] finalize(y) method
- [x] reset_spans() method
- [x] y() accessor method
- [x] num_spans() accessor method
- [x] begin() accessor method
- [x] Private members (m_max_len, m_last_x, m_y, m_spans)
- [x] Copy constructor and assignment operator (prohibited)

#### agg_scanline_p.h

Go files:

- internal/scanline/scanline_p8.go

**scanline_p8 class:**

- [x] self_type typedef
- [x] cover_type typedef (int8u)
- [x] coord_type typedef (int16)
- [x] span struct (x, len, covers pointer)
- [x] iterator and const_iterator typedefs
- [x] Default constructor
- [x] reset(min_x, max_x) method with memory allocation
- [x] add_cell(x, cover) method
- [x] add_cells(x, len, covers) method with memcpy
- [x] add_span(x, len, cover) method for solid spans
- [x] finalize(y) method
- [x] reset_spans() method
- [x] y() accessor method
- [x] num_spans() accessor method
- [x] begin() accessor method
- [x] Private members (m_last_x, m_y, m_covers, m_cover_ptr, m_spans, m_cur_span)
- [x] Copy constructor and assignment operator (prohibited)

**scanline32_p8 class:**

- [x] self_type typedef
- [x] cover_type typedef (int8u)
- [x] coord_type typedef (int32)
- [x] span struct with constructors
- [x] span_array_type typedef
- [x] const_iterator nested class
- [x] Default constructor
- [x] reset(min_x, max_x) method
- [x] add_cell(x, cover) method
- [x] add_cells(x, len, covers) method
- [x] add_span(x, len, cover) method
- [x] finalize(y) method
- [x] reset_spans() method
- [x] y() accessor method
- [x] num_spans() accessor method
- [x] begin() accessor method
- [x] Private members for 32-bit coordinate handling
- [x] Copy constructor and assignment operator (prohibited)

#### agg_scanline_u.h

**scanline_u8 class:**

- [x] self_type typedef
- [x] cover_type typedef (int8u)
- [x] coord_type typedef (int16)
- [x] span struct (x, len, covers array pointer)
- [x] iterator and const_iterator typedefs
- [x] Default constructor
- [x] reset(min_x, max_x) method
- [x] add_cell(x, cover) method
- [x] add_cells(x, len, covers) method
- [x] add_span(x, len, cover) method
- [x] finalize(y) method
- [x] reset_spans() method
- [x] y() accessor method
- [x] num_spans() accessor method
- [x] begin() accessor method
- [x] Private members (m_min_x, m_last_x, m_y, m_covers, m_spans, m_cur_span)

**scanline32_u8 class:**

- [x] Similar structure adapted for 32-bit coordinates
- [x] 32-bit typedefs and member variables
- [x] All corresponding methods adapted for larger coordinate space

#### agg_scanline_storage_aa.h

**scanline_cell_storage<T> template class:**

- [x] extra_span struct (len, ptr members)
- [x] value_type typedef
- [x] Destructor with memory cleanup
- [x] Default constructor
- [x] Copy constructor with deep copy
- [x] Assignment operator with proper cleanup
- [x] remove_all() method
- [x] add_cells(cells, num_cells) method with dynamic allocation
- [x] operator[] const overload for cell access
- [x] operator[] non-const overload for cell access
- [x] copy_extra_storage() private helper method
- [x] Private members (m_cells, m_extra_storage)

**scanline_storage_aa class:**

- [x] Embedded span and scanline structs
- [x] Constructor and destructor
- [x] min_x(), max_x() accessor methods
- [x] reset(min_x, max_x) method
- [x] add_cells() method
- [x] finalize() method
- [x] size() accessor method
- [x] operator[] for scanline access
- [x] Memory management methods

**scanline_storage_aa8 typedef:**

- [x] Concrete instantiation for int8u cover type

#### agg_scanline_storage_bin.h

**scanline_storage_bin class:**

- [x] Similar structure to AA storage but for binary scanlines
- [x] span and scanline structs for binary data
- [x] Constructor and destructor
- [x] prepare() method (equivalent to C++ prepare, not reset)
- [x] render() method (handles span processing, not separate add_span/finalize)
- [x] Access methods for binary scanline data (sweep, bounds, etc.)

**serialized_scanlines_adaptor_bin class:**

- [x] Binary scanline deserialization support
- [x] cover_type typedef (BinaryCoverType = bool)
- [x] embedded_scanline nested class with iterator
- [x] Constructor and data initialization
- [x] rewind_scanlines() method
- [x] sweep_scanline() methods (generic and embedded)
- [x] Coordinate offset support (dx, dy parameters)

#### agg_scanline_boolean_algebra.h

**Boolean operation functors (all template-based):**

- [x] sbool_combine_spans_bin template functor
- [x] sbool_combine_spans_empty template functor
- [x] sbool_add_span_empty template functor
- [x] sbool_add_span_bin template functor
- [x] sbool_add_span_aa template functor
- [x] sbool_intersect_spans_aa template functor with cover_scale_e enum
- [x] sbool_unite_spans_aa template functor
- [x] sbool_xor_spans_aa template functor
- [x] sbool_subtract_spans_aa template functor
- [x] Additional boolean operation functors

**Main algorithm templates:**

- [x] sbool_intersect_shapes template function
- [x] sbool_unite_shapes template function
- [x] sbool_xor_shapes template function
- [x] sbool_subtract_shapes template function

**Template adaptation considerations:**

- [x] Convert C++ functors to Go function types or interfaces
- [x] Adapt template parameters to Go generics or concrete types
- [x] Handle iterator patterns with Go-idiomatic approaches
- [x] Memory management adaptation for Go's garbage collector

---

### Rasterizers

#### agg_rasterizer_cells_aa.h

**Status: COMPLETED** ✅ - Fixed horizontal line rendering and cell generation logic

**rasterizer_cells_aa<Cell> template class:**

- [x] cell_block_scale_e enum (cell_block_shift, cell_block_size, cell_block_mask, cell_block_pool)
- [x] sorted_y struct (start, num members)
- [x] cell_type typedef
- [x] self_type typedef
- [x] Destructor with block memory cleanup
- [x] Constructor with cell_block_limit parameter
- [x] reset() method
- [x] style(style_cell) method
- [x] line(x1, y1, x2, y2) method for line rasterization
- [x] min_x() accessor method
- [x] min_y() accessor method
- [x] max_x() accessor method
- [x] max_y() accessor method
- [x] sort_cells() method
- [x] total_cells() accessor method
- [x] scanline_num_cells(y) method
- [x] scanline_cells(y) method
- [x] sorted() accessor method
- [x] set_curr_cell(x, y) private method
- [x] add_curr_cell() private method
- [x] render_hline() private method
- [x] allocate_block() private method
- [x] Private members (m_num_blocks, m_max_blocks, m_curr_block, m_num_cells, etc.)
- [x] Copy constructor and assignment operator (prohibited)

#### agg_rasterizer_scanline_aa.h

**Status: COMPLETED** ✅ - Fixed coordinate system conversions, generic type parameters, and interface compatibility

**rasterizer_scanline_aa<Clip> template class:**

- [x] status enum (status_initial, status_move_to, status_line_to, status_closed)
- [x] clip_type typedef
- [x] conv_type typedef
- [x] coord_type typedef
- [x] aa_scale_e enum (aa_shift, aa_scale, aa_mask, aa_scale2, aa_mask2)
- [x] Default constructor with cell_block_limit
- [x] Template constructor with gamma function
- [x] reset() method
- [x] reset_clipping() method
- [x] clip_box(x1, y1, x2, y2) method
- [x] filling_rule(filling_rule) method
- [x] auto_close(flag) method
- [x] gamma() template method for gamma correction
- [x] apply_gamma(cover) method
- [x] move_to(x, y) method (integer coordinates)
- [x] line_to(x, y) method (integer coordinates)
- [x] move_to_d(x, y) method (double coordinates)
- [x] line_to_d(x, y) method (double coordinates)
- [x] close_polygon() method
- [x] add_path() template method
- [x] add_vertex(x, y, cmd) method
- [x] edge(x1, y1, x2, y2) method
- [x] edge_d(x1, y1, x2, y2) method
- [x] sort() method
- [x] rewind_scanlines() method
- [x] calculate_alpha() method
- [x] sweep_scanline() template method
- [x] navigate_scanline(y) method
- [x] hit_test(tx, ty) method
- [x] Private members (m_outline, m_clipper, m_filling_rule, m_gamma, etc.)

#### agg_rasterizer_scanline_aa_nogamma.h

**Status: COMPLETED** ✅ - All functionality implemented and tested

**rasterizer_scanline_aa_nogamma<Clip> template class:**

- [x] Similar structure to rasterizer_scanline_aa but without gamma correction
- [x] Simplified apply_gamma() method (no gamma table)
- [x] All other methods matching rasterizer_scanline_aa interface
- [x] Performance-optimized implementation

#### agg_rasterizer_compound_aa.h

**Status: COMPLETED** ✅ - Fixed cell sorting crashes and completed missing methods

**cell_style_aa struct:**

- [x] Position members (x, y)
- [x] Coverage members (cover, area)
- [x] Style members (left, right)
- [x] initial() method
- [x] style(c) method
- [x] not_equal(ex, ey, c) method

**layer_order_e enum:**

- [x] layer_unsorted constant
- [x] layer_direct constant
- [x] layer_inverse constant

**rasterizer_compound_aa<Clip> template class:**

- [x] style_info struct (start_cell, num_cells, last_x)
- [x] cell_info struct (x, area, cover)
- [x] clip_type typedef
- [x] conv_type typedef
- [x] coord_type typedef
- [x] aa_scale_e enum constants
- [x] Default constructor
- [x] reset() method
- [x] reset_clipping() method
- [x] clip_box(x1, y1, x2, y2) method
- [x] filling_rule(filling_rule) method
- [x] layer_order(order) method
- [x] styles(left, right) method
- [x] move_to(x, y) method
- [x] line_to(x, y) method
- [x] move_to_d(x, y) method
- [x] line_to_d(x, y) method
- [x] add_vertex(x, y, cmd) method
- [x] edge(x1, y1, x2, y2) method
- [x] edge_d(x1, y1, x2, y2) method
- [x] sort() method
- [x] navigate_scanline(y) method
- [x] hit_test(tx, ty) method
- [x] allocate_cover_buffer() method
- [x] sweep_styles() method
- [x] scanline_start() method
- [x] scanline_length() method
- [x] style(style_id) method
- [x] add_path() method
- [x] Private members for style and layer management

#### agg_rasterizer_sl_clip.h

**Coordinate conversion structs:**

- [x] ras_conv_int struct (coord_type typedef, mul_div, xi, yi, upscale, downscale methods)
- [x] ras_conv_int_sat struct (saturated integer conversion)
- [x] ras_conv_int_3x struct (3x integer conversion for sub-pixel rendering)
- [x] ras_conv_dbl struct (double precision conversion)
- [x] ras_conv_dbl_3x struct (3x double conversion)

**Clipping template classes:**

- [x] rasterizer_sl_no_clip<Conv> template class
- [x] rasterizer_sl_clip_int<Conv> template class
- [x] rasterizer_sl_clip_int_sat<Conv> template class
- [x] rasterizer_sl_clip_int_3x<Conv> template class
- [x] rasterizer_sl_clip_dbl<Conv> template class
- [x] rasterizer_sl_clip_dbl_3x<Conv> template class

**Each clipping class includes:**

- [x] conv_type typedef
- [x] coord_type typedef
- [x] Constructor with clipping bounds
- [x] reset_clipping() method
- [x] clip_box() method
- [x] move_to() method
- [x] line_to() method
- [x] Private clipping implementation

#### agg_rasterizer_outline.h

**rasterizer_outline<Renderer> template class:**

- [x] Constructor with renderer
- [x] attach(renderer) method
- [x] move_to(x, y) method
- [x] line_to(x, y) method
- [x] move_to_d(x, y) method
- [x] line_to_d(x, y) method
- [x] close() method
- [x] add_path() template method
- [x] add_vertex(x, y, cmd) method
- [x] render_all_paths() template method
- [x] render_ctrl() template method
- [x] Private outline rendering implementation

#### agg_rasterizer_outline_aa.h

**rasterizer_outline_aa<Renderer> template class:**

- [x] Similar structure to rasterizer_outline but with anti-aliasing
- [x] Enhanced line rendering with coverage calculation
- [x] Anti-aliased endpoint handling
- [x] Smooth line joining algorithms

**Template adaptation considerations:**

- [x] Convert Cell template parameters to Go generics or interfaces
- [x] Adapt Clip template parameters to interface-based design
- [x] Convert Renderer template parameters to interface types
- [x] Replace C++ functors with Go function types
- [x] Adapt memory management for Go's garbage collector
- [x] Convert enums to Go constants or typed constants
- [x] Handle coordinate conversion with Go methods or interfaces

---

### Renderers

#### agg_renderer_base.h - Base renderer template class

**Template Class renderer_base<PixelFormat>**

- [x] pixfmt_type, color_type, row_data typedefs
- [x] Default constructor
- [x] Parameterized constructor with pixel format
- [x] attach() method for pixel format attachment

**Pixel Format Access**

- [x] ren() const method - pixel format accessor
- [x] ren() non-const method - pixel format accessor
- [x] width() const method
- [x] height() const method

**Clipping Operations**

- [x] clip_box(x1, y1, x2, y2) method with bounds checking
- [x] reset_clipping(visibility) method
- [x] clip_box_naked(x1, y1, x2, y2) method - no bounds checking
- [x] inbox(x, y) const method - point-in-clip test

**Clipping Accessors**

- [x] clip_box() const method
- [x] xmin(), ymin(), xmax(), ymax() accessors
- [x] bounding_clip_box() const method
- [x] bounding_xmin(), bounding_ymin(), bounding_xmax(), bounding_ymax() accessors

**Buffer Operations**

- [x] clear(color) method - clear entire buffer
- [x] fill(color) method - fill with blending

**Pixel Operations**

- [x] copy_pixel(x, y, color) method
- [x] blend_pixel(x, y, color, cover) method
- [x] pixel(x, y) const method - get pixel color

**Line Operations**

- [x] copy_hline(x1, y, x2, color) method
- [x] copy_vline(x, y1, y2, color) method
- [x] blend_hline(x1, y, x2, color, cover) method
- [x] blend_vline(x, y1, y2, color, cover) method

**Rectangle Operations**

- [x] copy_bar(x1, y1, x2, y2, color) method
- [x] blend_bar(x1, y1, x2, y2, color, cover) method

**Span Operations**

- [x] blend_solid_hspan(x, y, len, color, covers) method
- [x] blend_solid_vspan(x, y, len, color, covers) method
- [x] copy_color_hspan(x, y, len, colors) method
- [x] copy_color_vspan(x, y, len, colors) method
- [x] blend_color_hspan(x, y, len, colors, covers, cover) method
- [x] blend_color_vspan(x, y, len, colors, covers, cover) method

**Buffer Copying**

- [x] copy_from() template method for buffer-to-buffer copying

#### agg_renderer_scanline.h - Scanline rendering functions and classes

**Free Functions**

- [x] render_scanline_aa_solid<Scanline, BaseRenderer, ColorT>() function
- [x] render_scanlines_aa_solid<Rasterizer, Scanline, BaseRenderer, ColorT>() function
- [x] render_scanline_aa<Scanline, BaseRenderer, SpanAllocator, SpanGenerator>() function
- [x] render_scanlines_aa<Rasterizer, Scanline, BaseRenderer, SpanAllocator, SpanGenerator>() function
- [x] render_scanline_bin_solid<Scanline, BaseRenderer, ColorT>() function
- [x] render_scanlines_bin_solid<Rasterizer, Scanline, BaseRenderer, ColorT>() function

**Template Class renderer_scanline_aa_solid<BaseRenderer>**

- [x] base_ren_type, color_type typedefs
- [x] Constructor with base renderer
- [x] attach(base_ren) method
- [x] color(color) setter method
- [x] color() const getter method
- [x] prepare() method
- [x] render<Scanline>(scanline) template method

**Template Class renderer_scanline_aa<BaseRenderer, SpanAllocator, SpanGenerator>**

- [x] base_ren_type, alloc_type, span_gen_type typedefs
- [x] Constructor with base renderer
- [x] attach(base_ren, span_allocator, span_generator) method
- [x] prepare() method
- [x] render<Scanline>(scanline) template method

**Template Class renderer_scanline_bin_solid<BaseRenderer>**

- [x] base_ren_type, color_type typedefs
- [x] Constructor and attach method
- [x] color management methods
- [x] prepare() method
- [x] render<Scanline>(scanline) template method for binary scanlines

**Template Class renderer_scanline_bin<BaseRenderer, SpanAllocator, SpanGenerator>**

- [x] Similar structure to renderer_scanline_aa but for binary scanlines
- [x] Base renderer and span generator management
- [x] Binary scanline rendering

#### agg_renderer_primitives.h - Primitive drawing operations

**Template Class renderer_primitives<BaseRenderer>**

- [x] base_ren_type, color_type typedefs
- [x] Constructor with base renderer
- [x] attach(base_ren) method

**Color Management**

- [x] fill_color(color) setter method
- [x] line_color(color) setter method
- [x] fill_color() const getter method
- [x] line_color() const getter method

**Rectangle Operations**

- [x] rectangle(x1, y1, x2, y2) method - outlined rectangle
- [x] solid_rectangle(x1, y1, x2, y2) method - filled rectangle
- [x] outlined_rectangle(x1, y1, x2, y2) method - outlined with different line color

**Ellipse Operations**

- [x] ellipse(x, y, rx, ry) method - outlined ellipse with Bresenham algorithm
- [x] solid_ellipse(x, y, rx, ry) method - filled ellipse
- [x] outlined_ellipse(x, y, rx, ry) method - outlined with different line color

**Line Drawing**

- [x] line(x1, y1, x2, y2, last) method using DDA algorithm
- [x] move_to(x, y) method for path building
- [x] line_to(x, y, last) method for path building

**Accessors**

- [x] ren() const method - base renderer accessor
- [x] rbuf() const method - rendering buffer accessor

**Private Members**

- [x] m_ren pointer to base renderer
- [x] m_fill_color member
- [x] m_line_color member
- [x] m_curr_x, m_curr_y current position members

#### agg_renderer_markers.h - Marker shape rendering

**Template Class renderer_markers<BaseRenderer> (inherits from renderer_primitives)**

- [x] base_type, base_ren_type, color_type typedefs
- [x] Inheritance from renderer_primitives<BaseRenderer>

**Visibility and Basic Operations**

- [x] visible(x, y, r) const method - visibility test within bounds

**Basic Shape Markers**

- [x] square(x, y, r) method - solid square marker
- [x] diamond(x, y, r) method - solid diamond marker
- [x] circle(x, y, r) method - solid circle marker using ellipse algorithm
- [x] crossed_circle(x, y, r) method - circle with cross pattern

**Semi-ellipse Markers (Direction-specific)**

- [x] semiellipse_left(x, y, r) method
- [x] semiellipse_right(x, y, r) method
- [x] semiellipse_up(x, y, r) method
- [x] semiellipse_down(x, y, r) method

**Triangle Markers (Direction-specific)**

- [x] triangle_left(x, y, r) method
- [x] triangle_right(x, y, r) method
- [x] triangle_up(x, y, r) method
- [x] triangle_down(x, y, r) method

**Ray and Line Markers**

- [x] four_rays(x, y, r) method - plus sign pattern
- [x] cross(x, y, r) method - diagonal cross pattern
- [x] x(x, y, r) method - X pattern
- [x] dash(x, y, r) method - horizontal dash
- [x] dot(x, y, r) method - small filled circle
- [x] pixel(x, y, color) method - single pixel marker

#### agg_renderer_outline_aa.h - Anti-aliased outline rendering

**Template Class renderer_outline_aa<BaseRenderer>**

- [x] base_ren_type, color_type, coord_type typedefs
- [x] Constructor with base renderer
- [x] attach(base_ren) method

**Line Pattern Support**

- [x] pattern(line_pattern) method
- [x] pattern() const getter method
- [x] pattern_scale() setter method
- [x] pattern_scale() const getter method
- [x] pattern_start() setter method

**Line Join and Cap Settings**

- [x] line_join(join_type) method - miter, round, bevel
- [x] line_cap(cap_type) method - butt, square, round
- [x] inner_join(join_type) method
- [x] width(line_width) setter method
- [x] width() const getter method

**Rendering Methods**

- [x] move_to(x, y) method
- [x] line_to(x, y) method
- [x] move_to_d(x, y) method - double precision
- [x] line_to_d(x, y) method - double precision
- [x] close_polygon() method
- [x] add_path<VertexSource>(vs, path_id) template method
- [x] add_vertex(x, y, cmd) method

**Accuracy Control**

- [x] accuracy(approximation_scale) setter method
- [x] accuracy() const getter method

#### agg_renderer_outline_image.h - Image-based outline rendering

**Template Class renderer_outline_image<BaseRenderer, ImagePattern>**

- [x] base_ren_type, color_type, order_type typedefs
- [x] pattern_type typedef
- [x] Constructor with base renderer and pattern
- [x] attach(base_ren) method

**Pattern Management**

- [x] pattern(image_pattern) setter method
- [x] pattern() const getter method
- [x] pattern_scale_x(), pattern_scale_y() setters
- [x] pattern_scale() unified setter method

**Rendering Methods**

- [x] move_to(x, y) method
- [x] line_to(x, y) method
- [x] move_to_d(x, y) method
- [x] line_to_d(x, y) method
- [x] Pattern-based line stroke rendering

**Image Pattern Application**

- [x] Subpixel pattern positioning
- [x] Pattern scaling and rotation
- [x] Pattern tiling along line path

#### agg_renderer_mclip.h - Multi-clipping renderer

**Template Class renderer_mclip<PixelFormat>**

- [x] pixfmt_type, color_type typedefs
- [x] base_ren_type typedef
- [x] Constructor with pixel format
- [x] attach(pixfmt) method

**Clipping Region Management**

- [x] first_clip_box() method
- [x] add_clip_box(x1, y1, x2, y2) method
- [x] remove_last_clip_box() method
- [x] clear_clip_boxes() method
- [x] clip_box_count() const method

**Multi-region Clipping Operations**

- [x] copy_pixel(x, y, color) method with multi-clip
- [x] blend_pixel(x, y, color, cover) method with multi-clip
- [x] copy_hline(x1, y, x2, color) method with multi-clip
- [x] blend_hline(x1, y, x2, color, cover) method with multi-clip
- [x] copy_vline(x, y1, y2, color) method with multi-clip
- [x] blend_vline(x, y1, y2, color, cover) method with multi-clip

**Clipping Logic**

- [x] inbox_all(x, y) const method - point in all clips
- [x] inbox_any(x, y) const method - point in any clip
- [x] Intersection and union clipping operations

#### agg_renderer_raster_text.h - Raster text rendering

**Template Class renderer_raster_text<BaseRenderer, GlyphRasterizer>**

- [x] base_ren_type, color_type typedefs
- [x] glyph_ras_type, glyph_type typedefs
- [x] Constructor with base renderer and glyph rasterizer
- [x] attach(base_ren, glyph_ras) method

**Text Rendering**

- [x] render_text(x, y, text_string) method
- [x] render_glyph(x, y, glyph) method
- [x] Character positioning and advancement

**Font and Style Management**

- [x] color(text_color) setter method
- [x] color() const getter method
- [x] font_size(size) setter method
- [x] font_height() const getter method
- [x] baseline() const getter method

**Text Positioning**

- [x] move_to(x, y) method
- [x] text_out(text_string) method
- [x] Horizontal and vertical alignment support
- [x] Character and line spacing controls

**Glyph Operations**

- [x] Embedded raster font support
- [x] Glyph caching and reuse
- [x] Unicode text support

---

### Geometric Primitives

#### agg_arc.h - Arc generation

**arc class**

- [x] Default constructor: `arc()`
- [x] Parameterized constructor: `arc(double x, double y, double rx, double ry, double a1, double a2, bool ccw=true)`
- [x] init() method: `init(double x, double y, double rx, double ry, double a1, double a2, bool ccw=true)`
- [x] approximation_scale(double s) setter method
- [x] approximation_scale() const getter method
- [x] rewind(unsigned) method for path rewinding
- [x] vertex(double* x, double* y) method for vertex generation
- [x] normalize(double a1, double a2, bool ccw) private method for angle normalization

**Member Variables**

- [x] Position and radius: m_x, m_y, m_rx, m_ry (double)
- [x] Angle parameters: m_angle, m_start, m_end, m_da (double)
- [x] Scale parameter: m_scale (double)
- [x] State flags: m_ccw, m_initialized (bool)
- [x] Path command: m_path_cmd (unsigned)

#### agg_ellipse.h - Ellipse generation

**ellipse class**

- [x] Default constructor: `ellipse()`
- [x] Parameterized constructor: `ellipse(double x, double y, double rx, double ry, unsigned num_steps=0, bool cw=false)`
- [x] init() inline method: `init(double x, double y, double rx, double ry, unsigned num_steps=0, bool cw=false)`
- [x] approximation_scale(double scale) inline setter method
- [x] rewind(unsigned path_id) inline method
- [x] vertex(double* x, double* y) inline method
- [x] calc_num_steps() private inline method for step calculation

**Member Variables**

- [x] Position and radius: m_x, m_y, m_rx, m_ry (double)
- [x] Scale parameter: m_scale (double)
- [x] Step tracking: m_num, m_step (unsigned)
- [x] Direction flag: m_cw (bool)

#### agg_ellipse_bresenham.h - Bresenham ellipse algorithm

**ellipse_bresenham_interpolator class**

- [x] Constructor: `ellipse_bresenham_interpolator(int rx, int ry)`
- [x] dx() const getter method
- [x] dy() const getter method
- [x] operator++() increment operator for pixel stepping

**Member Variables**

- [x] Radius squared: m_rx2, m_ry2 (int)
- [x] Double radius squared: m_two_rx2, m_two_ry2 (int)
- [x] Current deltas: m_dx, m_dy (int)
- [x] Increments: m_inc_x, m_inc_y (int)
- [x] Current function value: m_cur_f (int)

#### agg_rounded_rect.h - Rounded rectangle generation

**rounded_rect class**

- [x] Default constructor: `rounded_rect()`
- [x] Parameterized constructor: `rounded_rect(double x1, double y1, double x2, double y2, double r)`
- [x] rect(double x1, double y1, double x2, double y2) method
- [x] radius(double r) single radius setter
- [x] radius(double rx, double ry) x/y radius setter
- [x] radius(double rx_bottom, double ry_bottom, double rx_top, double ry_top) top/bottom setter
- [x] radius(double rx1, double ry1, double rx2, double ry2, double rx3, double ry3, double rx4, double ry4) individual corner setter
- [x] normalize_radius() method
- [x] approximation_scale(double s) inline setter
- [x] approximation_scale() const inline getter
- [x] rewind(unsigned) method
- [x] vertex(double* x, double* y) method with state machine

**Member Variables**

- [x] Rectangle bounds: m_x1, m_y1, m_x2, m_y2 (double)
- [x] Corner radii: m_rx1, m_ry1, m_rx2, m_ry2, m_rx3, m_ry3, m_rx4, m_ry4 (double)
- [x] State tracking: m_status (unsigned)
- [x] Composed arc object: m_arc (arc)

#### agg_arrowhead.h - Arrowhead generation

**arrowhead class**

- [x] Default constructor: `arrowhead()`
- [x] head(double d1, double d2, double d3, double d4) inline head configuration
- [x] head() inline getter
- [x] no_head() inline method to disable head
- [x] tail(double d1, double d2, double d3, double d4) inline tail configuration
- [x] tail() inline getter
- [x] no_tail() inline method to disable tail
- [x] rewind(unsigned path_id) method
- [x] vertex(double* x, double* y) method

**Member Variables**

- [x] Head parameters: m_head_d1, m_head_d2, m_head_d3, m_head_d4 (double)
- [x] Tail parameters: m_tail_d1, m_tail_d2, m_tail_d3, m_tail_d4 (double)
- [x] Enable flags: m_head_flag, m_tail_flag (bool)
- [x] Coordinate array: m_coord[16] (double)
- [x] Command array: m_cmd[8] (unsigned)
- [x] Current state: m_curr_id, m_curr_coord (unsigned)

#### Implementation Files (.cpp)

**agg_arc.cpp**

- [x] Constructor implementation with angle normalization
- [x] init() method with parameter validation
- [x] approximation_scale() with scale setting
- [x] rewind() path reset logic
- [x] vertex() trigonometric vertex calculation
- [x] normalize() private angle normalization algorithm

**agg_arrowhead.cpp**

- [x] Constructor with coordinate and command array initialization
- [x] rewind() path selection logic (head/tail/both)
- [x] vertex() coordinate lookup and command generation

**agg_rounded_rect.cpp**

- [x] Constructor with radius initialization
- [x] rect() bounds setting with validation
- [x] radius() methods with various parameter combinations
- [x] normalize_radius() clamping to valid ranges
- [x] rewind() state machine initialization
- [x] vertex() complex state machine for corner generation using composed arc

#### Special Porting Considerations

- [x] All classes implement the same vertex source interface (rewind/vertex pattern)
- [x] No templates used - direct struct/class conversion
- [x] Mathematical dependencies on <cmath> functions (sin, cos, atan2, etc.)
- [x] State machines in vertex() methods need careful Go translation
- [x] Inline methods should be regular Go methods
- [x] Double precision floating point throughout
- [x] Path command constants from agg_basics.h dependency

---

### Curves and Paths

- [x] agg_curves.h - Curve approximation
- [x] agg_bezier_arc.h - Bezier arc
- [x] agg_bspline.h - B-spline curves
- [x] agg_path_storage.h - Path storage
- [x] agg_path_storage_integer.h - Integer path storage
- [x] agg_path_length.h - Path length calculation

---

### Transformations

- [x] agg_trans_affine.h - Affine transformations
- [x] agg_trans_bilinear.h - Bilinear transformations
- [x] agg_trans_perspective.h - Perspective transformations
- [x] agg_trans_viewport.h - Viewport transformations
- [x] agg_trans_single_path.h - Single path transformation
- [x] agg_trans_double_path.h - Double path transformation
- [x] agg_trans_warp_magnifier.h - Warp magnifier transformation

---

### Converters

- [x] agg_conv_adaptor_vcgen.h - Vertex generator adaptor
- [x] agg_conv_adaptor_vpgen.h - Vertex processor adaptor
- [x] agg_conv_bspline.h - B-spline converter
- [x] agg_conv_clip_polygon.h - Polygon clipping converter
- [x] agg_conv_clip_polyline.h - Polyline clipping converter
- [x] agg_conv_close_polygon.h - Polygon closing converter
- [x] agg_conv_concat.h - Path concatenation converter
- [x] agg_conv_contour.h - Contour converter
- [x] agg_conv_curve.h - Curve converter
- [x] agg_conv_dash.h - Dash converter
- [x] agg_conv_gpc.h - General Polygon Clipper converter
- [x] agg_conv_marker.h - Marker converter
- [x] agg_conv_marker_adaptor.h - Marker adaptor converter
- [x] agg_conv_segmentator.h - Segmentator converter
- [x] agg_conv_shorten_path.h - Path shortening converter
- [x] agg_conv_smooth_poly1.h - Polygon smoothing converter
- [x] agg_conv_stroke.h - Stroke converter
- [x] agg_conv_transform.h - Transform converter
- [x] agg_conv_unclose_polygon.h - Polygon unclosing converter

---

### Vertex Generators

- [x] agg_vcgen_bspline.h - B-spline vertex generator
- [x] agg_vcgen_contour.h - Contour vertex generator
- [x] agg_vcgen_dash.h - Dash vertex generator
- [x] agg_vcgen_markers_term.h - Terminal markers vertex generator
- [x] agg_vcgen_smooth_poly1.h - Polygon smoothing vertex generator
- [x] agg_vcgen_stroke.h - Stroke vertex generator
- [x] agg_vcgen_vertex_sequence.h - Vertex sequence generator

---

### Vertex Processors

- [x] agg_vpgen_clip_polygon.h - Polygon clipping vertex processor
- [x] agg_vpgen_clip_polyline.h - Polyline clipping vertex processor
- [x] agg_vpgen_segmentator.h - Segmentator vertex processor

---

### Spans and Gradients

- [x] agg_span_allocator.h - Span allocator
- [x] agg_span_converter.h - Span converter
- [x] agg_span_solid.h - Solid color span
- [x] agg_span_gradient.h - Gradient span
- [x] agg_span_gradient_alpha.h - Alpha gradient span
- [x] agg_span_gradient_contour.h - Contour gradient span
- [x] agg_span_gradient_image.h - Image gradient span
- [x] agg_span_gouraud.h - Gouraud shading span
- [x] agg_span_gouraud_gray.h - Grayscale Gouraud span
- [x] agg_span_gouraud_rgba.h - RGBA Gouraud span

---

### Image Processing

- [x] agg_image_accessors.h - Image accessors
- [x] agg_image_filters.h - Image filters
- [x] agg_span_image_filter.h - Image filter span
- [x] agg_span_image_filter_gray.h - Grayscale image filter span
- [x] agg_span_image_filter_rgb.h - RGB image filter span
- [x] agg_span_image_filter_rgba.h - RGBA image filter span

---

### Pattern Processing

- [x] agg_pattern_filters_rgba.h - RGBA pattern filters
- [x] agg_span_pattern_gray.h - Grayscale pattern span
- [x] agg_span_pattern_rgb.h - RGB pattern span
- [x] agg_span_pattern_rgba.h - RGBA pattern span

---

### Interpolators

- [x] agg_span_interpolator_adaptor.h - Interpolator adaptor
- [x] agg_span_interpolator_linear.h - Linear interpolator (including linear_subdiv)
- [x] agg_span_interpolator_persp.h - Perspective interpolator
- [x] agg_span_interpolator_trans.h - Transform interpolator
- [x] agg_span_subdiv_adaptor.h - Subdivision adaptor

---

### Utility and Math

- [x] agg_alpha_mask_u8.h - 8-bit alpha mask
- [x] agg_bitset_iterator.h - Bitset iterator
- [x] agg_blur.h - Blur effects
- [x] agg_bounding_rect.h - Bounding rectangle calculation
- [x] agg_clip_liang_barsky.h - Liang-Barsky clipping algorithm
- [x] agg_dda_line.h - DDA line algorithm
- [x] agg_gamma_functions.h - Gamma correction functions
- [x] agg_gamma_lut.h - Gamma lookup table
- [x] agg_gradient_lut.h - Gradient lookup table
- [x] agg_line_aa_basics.h - Anti-aliased line basics
- [x] agg_math_stroke.h - Stroke mathematics
- [x] agg_shorten_path.h - Path shortening
- [x] agg_simul_eq.h - Simultaneous equations solver
- [x] agg_vertex_sequence.h - Vertex sequence

---

### Text and Fonts

- [x] agg_embedded_raster_fonts.h - Embedded raster fonts
- [x] agg_font_cache_manager.h - Font cache manager
- [x] agg_font_cache_manager2.h - Font cache manager v2
- [x] agg_glyph_raster_bin.h - Binary glyph rasterizer
- [x] agg_gsv_text.h - GSV text rendering

---

### Controls (ctrl/)

- [x] agg_ctrl.h - Base control class
- [x] agg_bezier_ctrl.h - Bezier curve control
- [x] agg_cbox_ctrl.h - Checkbox control
- [x] agg_gamma_ctrl.h - Gamma control
- [x] agg_gamma_spline.h - Gamma spline
- [x] agg_polygon_ctrl.h - Polygon control
- [x] agg_rbox_ctrl.h - Radio button control
- [x] agg_scale_ctrl.h - Scale control
- [x] agg_slider_ctrl.h - Slider control
- [x] agg_spline_ctrl.h - Spline control

---

### Platform Support (platform/)

- [x] agg_platform_support.h - Platform support interface

---

### Utilities (util/)

- [x] agg_color_conv.h - Color conversion utilities
- [x] agg_color_conv_rgb16.h - 16-bit RGB color conversion
- [x] agg_color_conv_rgb8.h - 8-bit RGB color conversion

## Core Implementation Files (src/)

### Basic Implementations

**Note**: All basic implementations are complete but organized into Go packages following Go conventions rather than matching C++ file structure exactly.

- [x] agg_arc.cpp - Arc generation implementation → internal/shapes/arc.go
- [x] agg_arrowhead.cpp - Arrowhead implementation → internal/shapes/arrowhead.go
- [x] agg_bezier_arc.cpp - Bezier arc implementation → internal/bezierarc/bezier_arc.go
- [x] agg_bspline.cpp - B-spline implementation → internal/curves/bspline.go
- [x] agg_color_rgba.cpp - RGBA color implementation → internal/color/rgba.go
- [x] agg_curves.cpp - Curve approximation implementation → internal/curves/curves.go
- [x] agg_embedded_raster_fonts.cpp - Embedded fonts implementation → internal/fonts/embedded_fonts.go
- [x] agg_gsv_text.cpp - GSV text implementation → internal/gsv/gsv_text.go
- [x] agg_image_filters.cpp - Image filters implementation → internal/image/filters.go
- [x] agg_line_aa_basics.cpp - Anti-aliased line basics → internal/primitives/line_aa_basics.go
- [x] agg_line_profile_aa.cpp - Anti-aliased line profile → internal/renderer/outline/line_profile_aa.go
- [x] agg_rounded_rect.cpp - Rounded rectangle implementation → internal/shapes/rounded_rect.go
- [x] agg_sqrt_tables.cpp - Square root tables → internal/basics/math.go (gSqrtTable, FastSqrt)
- [x] agg_trans_affine.cpp - Affine transformation implementation
- [-] agg_trans_double_path.cpp - Double path transformation
- [-] agg_trans_single_path.cpp - Single path transformation
- [x] agg_trans_warp_magnifier.cpp - Warp magnifier implementation

### Vertex Generators

- [-] agg_vcgen_bspline.cpp - B-spline vertex generator
- [-] agg_vcgen_contour.cpp - Contour vertex generator
- [-] agg_vcgen_dash.cpp - Dash vertex generator
- [-] agg_vcgen_markers_term.cpp - Terminal markers implementation
- [-] agg_vcgen_smooth_poly1.cpp - Polygon smoothing implementation
- [-] agg_vcgen_stroke.cpp - Stroke vertex generator

### Vertex Processors

- [-] agg_vpgen_clip_polygon.cpp - Polygon clipping implementation
- [-] agg_vpgen_clip_polyline.cpp - Polyline clipping implementation
- [-] agg_vpgen_segmentator.cpp - Segmentator implementation

### Controls Implementation (src/ctrl/)

- [-] agg_bezier_ctrl.cpp - Bezier control implementation
- [-] agg_cbox_ctrl.cpp - Checkbox control implementation
- [-] agg_gamma_ctrl.cpp - Gamma control implementation
- [-] agg_gamma_spline.cpp - Gamma spline implementation
- [-] agg_polygon_ctrl.cpp - Polygon control implementation
- [-] agg_rbox_ctrl.cpp - Radio button implementation
- [-] agg_scale_ctrl.cpp - Scale control implementation
- [-] agg_slider_ctrl.cpp - Slider control implementation
- [-] agg_spline_ctrl.cpp - Spline control implementation

## AGG2D High-Level Interface

- [-] agg2d.h - AGG2D header
- [-] agg2d.cpp - AGG2D implementation

## Font Support

### FreeType Integration

- [-] agg_font_freetype.h - FreeType font support header
- [-] agg_font_freetype.cpp - FreeType font implementation
- [-] agg_font_freetype2.h - FreeType v2 support header
- [-] agg_font_freetype2.cpp - FreeType v2 implementation

### Win32 TrueType Support

- [-] agg_font_win32_tt.h - Win32 TrueType header
- [-] agg_font_win32_tt.cpp - Win32 TrueType implementation

## General Polygon Clipper (GPC)

- [-] gpc.h - GPC header
- [-] gpc.c - GPC implementation

## Platform Support Implementations

### Cross-Platform

- [-] agg_platform_support.cpp (generic interface)

### Platform-Specific (Optional - for examples)

- [-] src/platform/X11/agg_platform_support.cpp - X11 support
- [-] src/platform/win32/agg_platform_support.cpp - Win32 support
- [-] src/platform/win32/agg_win32_bmp.cpp - Win32 bitmap support
- [-] src/platform/mac/agg_platform_support.cpp - macOS support
- [-] src/platform/mac/agg_mac_pmap.cpp - macOS pixmap support
- [-] src/platform/BeOS/agg_platform_support.cpp - BeOS support
- [-] src/platform/AmigaOS/agg_platform_support.cpp - AmigaOS support
- [-] src/platform/sdl/agg_platform_support.cpp - SDL support
- [-] src/platform/sdl2/agg_platform_support.cpp - SDL2 support
- [-] src/platform/nano-X/agg_platform_support.cpp - nano-X support

## Priority Order

### Phase 1: Core Foundation

1. Basic types and configuration (agg_basics.h, agg_config.h, agg_array.h, agg_math.h)
2. Color handling (agg_color_rgba.h/.cpp, agg_color_gray.h)
3. Rendering buffer (agg_rendering_buffer.h)
4. Basic pixel formats (agg_pixfmt_base.h, agg_pixfmt_rgba.h, agg_pixfmt_rgb.h)

### Phase 2: Core Rendering

1. Scanlines (agg_scanline_u.h, agg_scanline_p.h, agg_scanline_bin.h)
2. Rasterizers (agg_rasterizer_scanline_aa.h, agg_rasterizer_cells_aa.h)
3. Basic renderers (agg_renderer_base.h, agg_renderer_scanline.h)
4. Spans (agg_span_solid.h, agg_span_allocator.h)

### Phase 3: Geometric Primitives

1. Path storage (agg_path_storage.h)
2. Basic shapes (agg_ellipse.h, agg_rounded_rect.h/.cpp, agg_arc.h/.cpp)
3. Transformations (agg_trans_affine.h/.cpp)
4. Basic converters (agg_conv_transform.h, agg_conv_stroke.h)

### Phase 4: Advanced Features

1. Curves (agg_curves.h/.cpp, agg_bezier_arc.h/.cpp, agg_bspline.h/.cpp)
2. Complex converters and vertex generators
3. Image processing and filters
4. Gradients and patterns
5. Text rendering
6. High-level AGG2D interface

## Notes

- Files marked with `.h` are header files that define interfaces and templates
- Files marked with `.cpp` are implementation files
- Some headers are template-only and may not have corresponding .cpp files
- Platform-specific files can be implemented as needed for target platforms
- The GPC library may need special licensing consideration
- Font support files are optional depending on text rendering requirements

## Implementation Status Notes

### internal/shapes/*.go and internal/agg2d/paths.go

**Status: COMPLETED** - The old note is stale. Geometric primitives now live under `internal/shapes/`, with dedicated `arc.go`, `ellipse.go`, and `rounded_rect.go` ports plus tests, and Agg2D path helpers use those implementations directly.

### internal/agg2d/text.go

**Status: PARTIAL** - Text rendering is no longer a placeholder. Agg2D now loads fonts, computes width/kerning, and renders outline, gray8, and mono glyph caches through `internal/font/cache_manager.go`. Remaining review focus is parity and edge cases, not basic functionality.

### internal/agg2d/image.go

**Status: COMPLETED** - The old note is stale. Image transforms, parallelogram mapping, path-based image rendering, interpolation, and resampling code paths are implemented and covered by tests.

### internal/agg2d/gradient.go

**Status: COMPLETED** - The old placeholder note is stale. Gradient setup now applies world/screen transforms and uses real gradient matrices rather than a fixed 1:1 mapping assumption.

### internal/agg2d/fill_rules.go

**Status: COMPLETED** - Fill rule selection is wired into the rasterizer immediately and is applied during rendering.

### internal/agg2d/agg2d.go and internal/agg2d/rendering.go

**Status: COMPLETED** - The old note is stale. The Agg2D rendering stack now initializes typed pixfmt/renderers, manages scanline/rasterizer state, and draws fill/stroke paths through the rendering pipeline.

### internal/vcgen/stroke.go

**Status: COMPLETED** - The stroke vertex generator is implemented and no longer a no-op placeholder.

### internal/span/span_image_filter_rgb.go

**Status: COMPLETED** - Reviewed and updated. RGB nearest-neighbor, bilinear, and bilinear-clip paths are implemented, and the bilinear clip path now blends partial out-of-bounds samples per corner instead of falling back to a flat background fill.

### internal/span/span_image_filter_rgba.go

**Status: COMPLETED** - Reviewed and updated. RGBA nearest-neighbor, bilinear, and bilinear-clip paths are implemented, and the bilinear clip path now blends partial out-of-bounds samples per corner instead of falling back to a flat background fill.

### internal/span/span_gradient_contour.go

**Status: PARTIAL** - Improved. The contour gradient path no longer uses the older manual snapped-line rasterization; it now renders the transformed curve through the outline rasterizer/primitives pipeline before the distance transform. Remaining work is a deeper parity review against AGG's original contour-gradient behavior rather than the previous placeholder-style rasterization gap.

### internal/span/converter.go

**Status: COMPLETED** - The alpha and brightness-alpha converters actively modify span colors; the old placeholder note is stale.

### internal/scanline/storage_aa_serialized.go

**Status: COMPLETED** - Serialized AA scanline storage/adaptor support is implemented and tested; the old placeholder note is stale.

### internal/scanline/scanline_p8.go and internal/scanline/scanline32_p8.go

**Status: COMPLETED** - Reviewed and updated. The temporary `unsafe` pointer arithmetic has been removed in favor of explicit slice-based cover tracking.

### internal/renderer/mclip.go

**Status: PARTIAL** - The multi-clip renderer is functional and tested, but should still be reviewed against AGG for exact buffer-operation parity.

### internal/renderer/base.go

**Status: PARTIAL** - Renderer base functionality is implemented; remaining review focus is fidelity of `CopyFrom` and related transfer semantics rather than missing functionality.

### internal/platform/platform_support.go

**Status: PARTIAL** - The generic platform support layer still uses mock/immediate redraw behavior in the fallback path, but this is no longer the whole story: SDL2/X11 backends exist separately under `internal/platform/sdl2/` and `internal/platform/x11/`.

### internal/pixfmt/pixfmt_rgba8.go

**Status: COMPLETED** - The old fixed-order note is stale. RGBA pixfmt now supports multiple component orders through typed blender/order combinations (`RGBA`, `BGRA`, `ARGB`, `ABGR`).

### internal/gpc/gpc.go

**Status: PARTIAL** - Core polygon clipping code exists and is exercised by tests/benchmarks, so the old "operations not implemented" note is stale. Remaining work is correctness review for edge cases, I/O helpers such as `ReadPolygon`, and converter parity where example coverage still documents rough edges.

### internal/fonts/embedded_fonts.go

**Status: PARTIAL** - This note still broadly applies. Embedded font coverage is substantial, but some aliases and fallback mappings remain intentionally shared rather than having fully distinct font data for every historical AGG face.

### internal/font/cache_manager.go

`internal/font/cache_manager.go` is now the authoritative port of AGG's `agg_font_cache_manager.h` for the Agg2D text path. Review remaining behavior there against AGG as needed, but do not reintroduce the removed duplicate v1 cache-manager implementation under `internal/fonts/`.

### internal/fonts/cache_manager2.go

`internal/fonts/cache_manager2.go` remains the separate `fman::font_cache_manager` / embedded-font path, not the Agg2D text path. Review notes for this package should stay clearly separated from `internal/font/cache_manager.go`.

### internal/font/freetype2/*

The FreeType2 port in `internal/font/freetype2/` is significantly closer to `agg_font_freetype2.h/.cpp` than this checklist currently implies:

- serialized glyph cache output now covers native gray/mono, AGG gray/mono, and outline modes rather than caching raw FreeType bitmap buffers
- cache-manager ownership no longer closes the engine
- engine/face close semantics are idempotent
- the previous engine-owned `FT_Face` mirror array has been removed; loaded faces are tracked directly in Go, which is closer to AGG's `loaded_face` ownership model

Remaining review focus should be on the still-intentional Go deltas: convenience wrappers such as `FontManager`, any residual API surface beyond AGG's `fman` types, and any multi-face lifetime behavior still not directly modeled on the original source.

### internal/conv/gpc.go

**Status: PARTIAL** - The converter is no longer a placeholder and does invoke polygon clipping, but it still depends on the remaining correctness review noted for `internal/gpc/gpc.go`.

### internal/color/rgb.go

**Status: COMPLETED** - The old placeholder note is stale. `RGB` now supports the expected arithmetic and RGBA conversion helpers.

### internal/color/conversion.go

**Status: PARTIAL** - This is no longer "only the most common conversions" in the original sense: the file now includes scalar, RGB8, RGBA8, Gray8, Gray16, RGBA32, and Gray32 conversions. Remaining work is breadth/parity, not a placeholder implementation.

### internal/renderer/scanline/helpers.go

**Status: COMPLETED** - The helpers are functional, used by the renderer, and include compound rendering support. The old placeholder note is stale.

### internal/platform/sdl2/sdl2_display.go and internal/platform/x11/x11_display.go

**Status: PARTIAL** - These are no longer empty placeholders; backend code exists for real windows/surfaces. Remaining work is backend completeness and platform parity, not total absence.

### examples/core/intermediate/gradients/main.go

**Status: COMPLETED** - The example renders gradient-filled content and writes a PNG output.

### examples/core/basic/hello_world/main.go

**Status: COMPLETED** - Updated. The example now renders content and writes `hello_world.png` instead of only printing pixel values.
