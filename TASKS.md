# AGG 2.6 Go Port - File Checklist

This is a comprehensive checklist of files that need to be ported from the original AGG 2.6 C++ codebase to Go. Please always check the original C++ implementation for reference in ../agg-2.6

## Core Header Files (include/)

Previously completed tasks have been moved to TASKS-COMPLETED.md

---

### Transformations

#### agg_trans_affine.h - 2D Affine transformations (core transformation system) ✅ **COMPLETED**

**Core Affine Matrix Structure**

- [x] trans_affine → TransAffine struct - 2x3 affine transformation matrix (`internal/transform/affine.go`)
  - [x] Matrix components: sx, shy, shx, sy, tx, ty (scaling, shearing, translation)
  - [x] Identity constructor → NewTransAffine() - creates identity matrix
  - [x] Custom matrix constructor with 6 parameters
  - [x] Copy constructor from existing matrix
  - [x] Array-based constructor from [6]float64

**Basic Transformation Methods**

- [x] Reset() - reset to identity matrix
- [x] Translation operations
  - [x] translate(x, y) → Translate(x, y) - add translation to current matrix
  - [x] tx, ty accessors → TX(), TY() - get translation components
- [x] Scaling operations
  - [x] scale(s) → Scale(s) - uniform scaling
  - [x] scale(sx, sy) → ScaleXY(sx, sy) - non-uniform scaling
  - [x] sx, sy accessors → SX(), SY() - get scaling components
- [x] Rotation operations
  - [x] rotate(angle) → Rotate(angle) - rotation in radians
  - [x] rotation() → Rotation() - extract rotation angle from matrix

**Advanced Matrix Operations**

- [x] Matrix composition and multiplication
  - [x] multiply(matrix) → Multiply(matrix) - matrix multiplication
  - [x] premultiply(matrix) → Premultiply(matrix) - pre-multiplication
  - [x] operator*= overloading → Go method equivalent
- [x] Matrix inversion and determinant
  - [x] invert() → Invert() - matrix inversion with singularity check
  - [x] determinant() → Determinant() - calculate matrix determinant
  - [x] is_valid() → IsValid() - check for valid transformation
- [x] Decomposition methods
  - [x] scaling() → GetScaling() - extract combined scaling factor
  - [x] rotation() → GetRotation() - extract rotation angle
  - [x] translation() → GetTranslation() - extract translation vector

**Point and Vector Transformation**

- [x] Point transformation methods
  - [x] transform(x, y) → Transform(x, y) - transform point coordinates
  - [x] transform(point) → TransformPoint(point) - transform Point[T] type
  - [x] inverse_transform(x, y) → InverseTransform(x, y) - apply inverse transformation
- [x] Vector transformation (no translation)
  - [x] transform_2x2(x, y) → Transform2x2(x, y) - transform as vector
  - [x] rotation and scaling only, ignores translation component

**Specialized Transformation Constructors** (`internal/transform/helpers.go`)

- [x] trans_affine_rotation → NewTransAffineRotation(angle) - pure rotation matrix
- [x] trans_affine_scaling → NewTransAffineScaling(s) or (sx, sy) - pure scaling matrix
- [x] trans_affine_translation → NewTransAffineTranslation(x, y) - pure translation matrix
- [x] trans_affine_skewing → NewTransAffineSkewing(x, y) - pure skew matrix

**Matrix Analysis and Utilities**

- [x] Matrix property checks
  - [x] is_identity() → IsIdentity() - check for identity matrix
  - [x] is_equal(other) → IsEqual(other) - matrix equality with epsilon
  - [ ] is_orthogonal() → IsOrthogonal() - check for orthogonal transformation
- [x] Epsilon handling
  - [x] affine_epsilon constant → AffineEpsilon - numerical precision threshold
  - [x] Custom epsilon support in comparison operations

**Integration with Path Processing**

- [x] Vertex source compatibility
  - [x] Compatible with conv_transform converter
  - [x] Path coordinate transformation
  - [x] Bounding box transformation
- [ ] Span interpolator integration
  - [ ] Matrix extraction for span_interpolator_linear
  - [ ] Coordinate system conversion for image transformations

**Advanced Composition Patterns**

- [x] Transformation around arbitrary points
  - [x] Pre-translate, transform, post-translate pattern
  - [x] Rotate around point helper functions
  - [x] Scale around point helper functions
- [x] Transformation chaining
  - [x] Method chaining for fluent API: NewTransAffine().Rotate(angle).Scale(s).Translate(x, y)
  - [x] Cumulative transformation building
  - [x] Order-dependent transformation sequences

**Performance Optimizations**

- [x] Fast path detection
  - [x] Identity matrix fast path
  - [x] Translation-only fast path
  - [x] Axis-aligned transformation detection
- [x] Batch transformation
  - [x] Transform multiple points efficiently
  - [ ] SIMD optimization opportunities
  - [x] Memory-friendly transformation patterns

**Error Handling and Edge Cases**

- [x] Singular matrix handling
  - [x] Determinant near-zero detection
  - [x] Graceful degradation for non-invertible matrices
  - [x] Error propagation in transformation chains
- [x] Numerical stability
  - [x] Large coordinate handling
  - [x] Precision preservation in repeated transformations
  - [x] Round-trip transformation accuracy

**Dependencies**

- [x] agg_basics.h → internal/basics package
- [x] agg_math.h → internal/math package (trigonometric functions)
- [x] Point/Vector types from basic geometry package
- [x] Mathematical constants and utilities

#### agg_trans_bilinear.h - Bilinear transformations for quadrilateral mapping

**Bilinear Transformation Matrix**

- [x] trans_bilinear → TransBilinear struct - Bilinear coordinate transformation
  - [x] 4-point control system for arbitrary quadrilateral mapping
  - [x] Source and destination quadrilateral definitions
  - [x] Bilinear interpolation coefficients calculation
  - [x] Non-affine transformation capability (curved perspective effects)

**Quadrilateral Definition Methods**

- [x] Construction from control points
  - [x] rect_to_quad(x1, y1, x2, y2, quad) → RectToQuad() - rectangle to arbitrary quad
  - [x] quad_to_rect(quad, x1, y1, x2, y2) → QuadToRect() - quad to rectangle
  - [x] quad_to_quad(src_quad, dst_quad) → QuadToQuad() - arbitrary quad mapping
  - [x] Control point validation and normalization

**Bilinear Interpolation Mathematics**

- [x] Coefficient calculation
  - [x] A, B, C, D coefficient matrices for X and Y coordinates
  - [x] Bilinear basis function implementation
  - [x] Edge case handling for degenerate quadrilaterals
- [x] Forward transformation
  - [x] transform(x, y) → Transform(x, y) - apply bilinear mapping
  - [x] Parametric coordinate calculation (u, v parameters)
  - [x] Interpolation along quad edges and interior

**Inverse Transformation**

- [x] Inverse coordinate mapping
  - [x] inverse_transform(x, y) → InverseTransform(x, y) - reverse mapping
  - [x] Newton-Raphson iteration for inverse calculation
  - [x] Convergence criteria and iteration limits
  - [x] Fallback methods for difficult inversions

**Specialized Mapping Types**

- [x] Image rectification (application examples - core transformation supports these use cases)
  - [x] Perspective correction for photographed documents
  - [x] Keystone correction for projected displays
  - [x] Camera distortion correction integration
- [x] Texture mapping applications (application examples - core transformation supports these use cases)
  - [x] UV coordinate generation for arbitrary quads
  - [x] Mesh warping and distortion effects
  - [x] Real-time image transformation

**Integration with Rendering Pipeline** (BLOCKED - requires span interpolator infrastructure)

- [ ] Span interpolator compatibility (BLOCKED - span interpolators not yet implemented)
  - [ ] span_interpolator integration for image transformation
  - [ ] Bilinear filtering integration
  - [ ] Multi-pass rendering support
- [x] Path transformation (core transformation supports these use cases)
  - [x] Vector graphics warping
  - [x] Text distortion effects
  - [x] Shape morphing applications

**Performance and Optimization**

- [x] Coefficient caching
  - [x] Pre-calculated transformation matrices
  - [x] Incremental transformation updates (IteratorX)
  - [x] Batch transformation optimization
- [x] Numerical stability
  - [x] Singular quadrilateral detection
  - [x] Graceful degradation for edge cases
  - [x] Precision preservation in extreme transformations

**Dependencies**

- [x] agg_trans_affine.h → internal/transform package (affine transformation base)
- [x] agg_basics.h → internal/basics package (point types)
- [x] agg_simul_eq.h → simultaneous equation solver
- [x] Mathematical utilities for iterative solving

#### agg_trans_perspective.h - 3D perspective projections in 2D space

**Perspective Transformation Matrix**

- [x] trans_perspective → TransPerspective struct - 3x3 homogeneous transformation matrix
  - [x] Matrix components: sx, shy, w0, shx, sy, w1, tx, ty, w2
  - [x] Homogeneous coordinate system support
  - [x] Perspective division implementation (w-coordinate handling)
  - [x] Identity matrix initialization

**Construction Methods**

- [x] Basic constructors
  - [x] Identity constructor → NewTransPerspective() - identity transformation
  - [x] Custom matrix constructor with 9 parameters
  - [x] Array-based constructor from [9]float64
  - [x] Copy from affine matrix (w2=1, w0=w1=0)

**Quadrilateral-Based Perspective**

- [x] Quadrilateral mapping constructors
  - [x] rect_to_quad(x1, y1, x2, y2, quad) → RectToQuad() - rectangle to perspective quad
  - [x] quad_to_rect(quad, x1, y1, x2, y2) → QuadToRect() - reverse perspective mapping
  - [x] quad_to_quad(src, dst) → QuadToQuad() - arbitrary quad-to-quad perspective

**Perspective Transformation Methods**

- [x] Point transformation
  - [x] transform(x, y) → Transform(x, y) - apply perspective projection
  - [x] Homogeneous coordinate calculation
  - [x] Perspective division (x/w, y/w)
  - [x] Handle infinite/undefined points (w near zero)
- [x] Inverse transformation
  - [x] inverse_transform(x, y) → InverseTransform(x, y) - reverse perspective
  - [x] Matrix inversion for perspective matrices
  - [x] Singularity detection and handling

**Matrix Operations**

- [x] Matrix composition
  - [x] multiply(matrix) → Multiply(matrix) - perspective matrix multiplication
  - [x] premultiply(matrix) → Premultiply(matrix) - pre-multiplication
  - [x] Operator overloading equivalent methods
- [x] Matrix analysis
  - [x] determinant() → Determinant() - 3x3 matrix determinant
  - [x] invert() → Invert() - perspective matrix inversion
  - [x] is_valid() → IsValid() - matrix validity checking

**Advanced Perspective Operations** (Future enhancements)

- [ ] 3D projection simulation
  - [ ] Vanishing point calculations
  - [ ] Horizon line determination
  - [ ] Field of view and focal length simulation
- [ ] Perspective correction
  - [ ] Keystone correction for displays
  - [ ] Camera perspective correction
  - [ ] Architectural photography correction

**Integration with Affine Transformations**

- [x] Affine compatibility
  - [x] Conversion from affine to perspective
  - [x] Affine transformation embedding
  - [x] Mixed transformation pipelines
- [x] Transformation composition
  - [x] Affine pre/post-multiplication
  - [x] Complex transformation chains
  - [x] Order-dependent composition

**Specialized Applications**

- [ ] Image rectification
  - [ ] Perspective distortion correction
  - [ ] Document scanning applications
  - [ ] Real-time perspective correction
- [ ] 3D graphics simulation
  - [ ] 2D sprite projection
  - [ ] Pseudo-3D effects
  - [ ] Depth simulation in 2D

**Performance Considerations**

- [ ] Division optimization
  - [ ] Fast perspective division techniques
  - [ ] Avoiding division by near-zero w values
  - [ ] Batch transformation optimization
- [ ] Numerical stability
  - [ ] Precision handling in extreme perspectives
  - [ ] Graceful degradation for singular matrices
  - [ ] Error propagation management

**Dependencies**

- agg_trans_affine.h → internal/transform package (affine transformation base)
- agg_simul_eq.h → simultaneous equation solver
- agg_basics.h → internal/basics package
- Mathematical utilities for 3x3 matrix operations

#### agg_trans_viewport.h - Viewport and coordinate system transformations

**Viewport Transformation System**

- [x] trans_viewport → TransViewport struct - Coordinate system mapping
  - [x] World-to-device coordinate transformation
  - [x] Aspect ratio preservation options
  - [ ] Viewport clipping rectangle definition
  - [ ] Automatic scaling and centering

**Viewport Definition Methods**

- [ ] World coordinate system
  - [ ] world_viewport(x1, y1, x2, y2) → SetWorldViewport() - define world bounds
  - [ ] World coordinate range specification
  - [ ] Infinite coordinate space handling
  - [ ] Coordinate system orientation (Y-up vs Y-down)
- [ ] Device coordinate system
  - [ ] device_viewport(x1, y1, x2, y2) → SetDeviceViewport() - define device bounds
  - [ ] Pixel-perfect device coordinate mapping
  - [ ] Device coordinate constraints and validation
  - [ ] Screen/window coordinate integration

**Aspect Ratio and Scaling Control**

- [ ] Aspect ratio preservation
  - [ ] aspect_ratio_e enumeration → AspectRatio type
  - [ ] stretch → Stretch - fill viewport completely (distortion allowed)
  - [ ] meet → Meet - fit entirely within viewport (letterbox/pillarbox)
  - [ ] slice → Slice - fill viewport, crop excess (zoom to fit)
- [ ] Alignment control
  - [ ] Horizontal alignment: left, center, right
  - [ ] Vertical alignment: top, middle, bottom
  - [ ] Custom alignment offset parameters

**Transformation Calculation**

- [ ] Automatic matrix generation
  - [ ] Scale factor calculation based on viewport ratios
  - [ ] Translation calculation for centering and alignment
  - [ ] Combined scaling and translation matrix
  - [ ] Matrix update on viewport changes
- [ ] Transformation extraction
  - [ ] to_affine() → ToAffine() - extract equivalent affine transformation
  - [ ] scale_x(), scale_y() → ScaleX(), ScaleY() - get scaling factors
  - [ ] offset_x(), offset_y() → OffsetX(), OffsetY() - get translation offsets

**Coordinate Conversion Methods**

- [ ] World-to-device conversion
  - [ ] world_to_device(x, y) → WorldToDevice(x, y) - forward transformation
  - [ ] Bulk point transformation for performance
  - [ ] Path coordinate transformation
  - [ ] Bounding box transformation
- [ ] Device-to-world conversion
  - [ ] device_to_world(x, y) → DeviceToWorld(x, y) - inverse transformation
  - [ ] Mouse coordinate mapping
  - [ ] Hit testing coordinate conversion
  - [ ] Interactive viewport navigation

**Viewport State Management**

- [ ] Viewport validation
  - [ ] is_valid() → IsValid() - check for valid viewport configuration
  - [ ] Zero-size viewport handling
  - [ ] Invalid coordinate range detection
  - [ ] Degenerate transformation prevention
- [ ] State change detection
  - [ ] viewport_changed() → ViewportChanged() - detect viewport modifications
  - [ ] Automatic matrix recalculation
  - [ ] Change event propagation

**Advanced Viewport Features**

- [ ] Multi-viewport support
  - [ ] Nested viewport transformations
  - [ ] Viewport hierarchy management
  - [ ] Relative viewport positioning
- [ ] Zoom and pan integration
  - [ ] Interactive zoom controls
  - [ ] Pan operation coordinate handling
  - [ ] Zoom-to-fit functionality
  - [ ] Smooth transition support

**Integration with Rendering System**

- [ ] Renderer integration
  - [ ] Automatic clipping rectangle setup
  - [ ] Coordinate system integration
  - [ ] Multi-pass rendering support
- [ ] Path processing integration
  - [ ] Automatic path transformation
  - [ ] Viewport-aware path optimization
  - [ ] Level-of-detail based on viewport scale

**Performance Optimizations**

- [ ] Cached transformation matrix
  - [ ] Matrix recalculation only when needed
  - [ ] Incremental viewport updates
  - [ ] Fast path for common viewport changes
- [ ] Batch transformations
  - [ ] Efficient multi-point transformation
  - [ ] SIMD optimization opportunities
  - [ ] Memory-friendly transformation patterns

**Dependencies**

- agg_trans_affine.h → internal/transform package
- agg_basics.h → internal/basics package
- Mathematical utilities for viewport calculations
- Enumeration types for aspect ratio and alignment

#### agg_trans_single_path.h - Transform along a single curved path

**Path-Based Transformation System**

- [x] trans_single_path → TransSinglePath struct - Transform coordinates along curved path
  - [x] Path-following coordinate system
  - [x] Arc-length parameterization
  - [x] Normal and tangent vector calculation
  - [x] Distance-based positioning along path

**Path Definition and Processing**

- [ ] Path setup
  - [ ] add_path(vertex_source) → AddPath() - define the transformation path
  - [ ] Path length calculation and caching
  - [ ] Path segment analysis and optimization
  - [ ] Closed vs. open path handling
- [ ] Path analysis
  - [ ] total_length() → TotalLength() - get total path length
  - [ ] Curvature analysis for quality control
  - [ ] Path direction and orientation
  - [ ] Critical point identification (cusps, loops)

**Coordinate Transformation Methods**

- [ ] Forward transformation
  - [ ] transform(x, y) → Transform(x, y) - map coordinates to path
  - [ ] X-coordinate maps to distance along path
  - [ ] Y-coordinate maps to perpendicular offset from path
  - [ ] Tangent and normal vector calculation at each point
- [ ] Path parameterization
  - [ ] Arc-length parameterization for even spacing
  - [ ] Parameter-to-distance conversion
  - [ ] Distance-to-parameter conversion
  - [ ] Interpolation between path segments

**Path Navigation and Positioning**

- [ ] Distance-based positioning
  - [ ] Position calculation at specific path distance
  - [ ] Interpolation between path vertices
  - [ ] Smooth position transitions
  - [ ] Boundary condition handling (before start, after end)
- [ ] Orientation calculation
  - [ ] Tangent vector at any path position
  - [ ] Normal vector for perpendicular positioning
  - [ ] Angle calculation for oriented objects
  - [ ] Smooth orientation transitions

**Advanced Path Features**

- [ ] Path quality control
  - [ ] approximation_scale() → SetApproximationScale() - control path detail
  - [ ] Adaptive path tessellation
  - [ ] Smooth curve approximation
  - [ ] Sharp corner handling
- [ ] Path modification
  - [ ] Path reversal for direction changes
  - [ ] Path segment extraction
  - [ ] Path smoothing and filtering
  - [ ] Multi-path composition

**Text-on-Path Applications**

- [ ] Text layout along curves
  - [ ] Character positioning and orientation
  - [ ] Baseline following path curvature
  - [ ] Character spacing adjustment for curves
  - [ ] Text direction and reading flow
- [ ] Advanced text features
  - [ ] Multi-line text on path
  - [ ] Text alignment options (left, center, right, justify)
  - [ ] Overflow handling for text longer than path
  - [ ] Dynamic text fitting and scaling

**Shape Transformation Applications**

- [ ] Shape morphing along paths
  - [ ] Object orientation following path direction
  - [ ] Scale variation along path
  - [ ] Shape deformation based on path curvature
- [ ] Animation support
  - [ ] Smooth position interpolation
  - [ ] Velocity-based positioning
  - [ ] Acceleration and deceleration curves
  - [ ] Keyframe animation along paths

**Performance Optimizations**

- [ ] Path preprocessing
  - [ ] Arc-length table generation
  - [ ] Segment caching for repeated access
  - [ ] Fast lookup tables for distance mapping
  - [ ] Incremental path processing
- [ ] Transformation caching
  - [ ] Cached tangent and normal vectors
  - [ ] Reuse calculations for nearby points
  - [ ] Batch transformation optimization

**Integration with Path Storage**

- [ ] Vertex source compatibility
  - [ ] Compatible with path_storage
  - [ ] Works with all curve types
  - [ ] Automatic curve approximation
- [ ] Path converter integration
  - [ ] Works with conv_stroke for outlined paths
  - [ ] Compatible with conv_dash for dashed paths
  - [ ] Integration with path transformation pipeline

**Dependencies**

- agg_path_storage.h → internal/path package
- agg_curves.h → curve approximation classes
- agg_vertex_sequence.h → vertex processing utilities
- Arc-length calculation utilities
- Vector mathematics for tangent/normal calculations

#### agg_trans_double_path.h - Transform between two curved paths

**Dual-Path Transformation System**

- [x] trans_double_path → TransDoublePath struct - Transform using two guide paths
  - [x] Base path and top path definition
  - [x] Morphing between two arbitrary curved paths
  - [x] Bilinear interpolation across path pair
  - [ ] Variable width corridor transformation

**Path Pair Definition**

- [ ] Dual path setup
  - [ ] add_paths(base_path, top_path) → AddPaths() - define transformation corridor
  - [ ] Path synchronization and length matching
  - [ ] Corresponding point calculation
  - [ ] Path direction alignment
- [ ] Path relationship analysis
  - [ ] Relative path positioning
  - [ ] Path separation distance calculation
  - [ ] Crossing and intersection detection
  - [ ] Path convergence and divergence points

**Coordinate Transformation Mathematics**

- [ ] Bilinear path interpolation
  - [ ] X-coordinate maps to position along both paths
  - [ ] Y-coordinate maps to interpolation between paths (0=base, 1=top)
  - [ ] Smooth interpolation across the path corridor
  - [ ] Edge case handling for path intersections
- [ ] Transform calculation
  - [ ] transform(x, y) → Transform(x, y) - map point to path corridor
  - [ ] Corresponding point calculation on both paths
  - [ ] Linear interpolation between corresponding points
  - [ ] Orientation interpolation for smooth transitions

**Advanced Path Processing**

- [ ] Path synchronization
  - [ ] Arc-length parameterization for both paths
  - [ ] Path re-parameterization for consistent mapping
  - [ ] Automatic path length adjustment
  - [ ] Critical point alignment between paths
- [ ] Quality control
  - [ ] approximation_scale() → SetApproximationScale() - control detail level
  - [ ] Path smoothing and filtering
  - [ ] Self-intersection detection and handling
  - [ ] Degenerate case management

**Corridor-Based Applications**

- [ ] Text layout in variable-width corridors
  - [ ] Text flowing between curved boundaries
  - [ ] Dynamic text sizing based on corridor width
  - [ ] Multi-line text in curved regions
  - [ ] Text justification in non-uniform spaces
- [ ] Shape morphing between paths
  - [ ] Smooth shape transitions
  - [ ] Animation between different path shapes
  - [ ] Envelope distortion effects
  - [ ] Perspective-like distortions

**Path Corridor Analysis**

- [ ] Width calculation
  - [ ] Distance between paths at any position
  - [ ] Minimum and maximum corridor width
  - [ ] Width variation analysis
  - [ ] Bottleneck detection
- [ ] Path relationship metrics
  - [ ] Average separation distance
  - [ ] Path parallelism measurement
  - [ ] Correlation between path curvatures
  - [ ] Geometric compatibility assessment

**Specialized Transformation Effects**

- [ ] Envelope distortion
  - [ ] Shape fitting within curved boundaries
  - [ ] Proportional scaling based on corridor width
  - [ ] Perspective-like effects from path convergence
- [ ] Flow field simulation
  - [ ] Particle movement along path corridor
  - [ ] Velocity field generation
  - [ ] Streamline visualization
  - [ ] Fluid dynamics simulation

**Performance and Optimization**

- [ ] Dual path preprocessing
  - [ ] Arc-length table generation for both paths
  - [ ] Corresponding point cache
  - [ ] Fast lookup structures
  - [ ] Incremental processing optimization
- [ ] Batch transformation
  - [ ] Efficient multi-point processing
  - [ ] Vectorized calculations
  - [ ] Memory-friendly access patterns

**Error Handling and Edge Cases**

- [ ] Path mismatch handling
  - [ ] Different path lengths
  - [ ] Path intersection scenarios
  - [ ] Degenerate path configurations
  - [ ] Numerical stability in extreme cases
- [ ] Boundary condition management
  - [ ] Beyond path endpoints
  - [ ] Outside path corridor
  - [ ] Path crossing points
  - [ ] Singular transformation points

**Dependencies**

- agg_trans_single_path.h → single path transformation (base functionality)
- agg_path_storage.h → internal/path package
- agg_curves.h → curve approximation
- Vector mathematics and interpolation utilities
- Arc-length parameterization tools

#### agg_trans_warp_magnifier.h - Warp magnifier transformation (lens effects)

**Magnifier Lens Transformation**

- [x] trans_warp_magnifier → TransWarpMagnifier struct - Lens distortion transformation
  - [x] Circular magnification area definition
  - [x] Variable magnification factor
  - [x] Smooth distortion falloff
  - [x] Real-time lens effect simulation

**Lens Parameter Definition**

- [ ] Magnifier setup
  - [ ] center(x, y) → SetCenter(x, y) - lens center position
  - [ ] radius(r) → SetRadius(r) - magnification area radius
  - [ ] magnification(m) → SetMagnification(m) - magnification factor
  - [ ] Interactive parameter adjustment
- [ ] Lens shape control
  - [ ] Circular lens (default)
  - [ ] Elliptical lens variations
  - [ ] Custom lens shape support
  - [ ] Lens boundary smoothness control

**Magnification Mathematics**

- [ ] Lens distortion calculation
  - [ ] Distance-from-center calculation
  - [ ] Radial magnification formula
  - [ ] Smooth falloff function (avoid sharp edges)
  - [ ] Inverse transformation for mouse interaction
- [ ] Forward transformation
  - [ ] transform(x, y) → Transform(x, y) - apply lens distortion
  - [ ] Magnified region calculation
  - [ ] Normal region pass-through
  - [ ] Smooth transition between regions

**Advanced Lens Effects**

- [ ] Multiple magnification zones
  - [ ] Overlapping lens effects
  - [ ] Additive vs. multiplicative magnification
  - [ ] Complex distortion patterns
  - [ ] Multi-center lens systems
- [ ] Dynamic lens properties
  - [ ] Animated magnification factor
  - [ ] Moving lens center
  - [ ] Pulsing or breathing effects
  - [ ] Interactive real-time control

**Distortion Quality Control**

- [ ] Anti-aliasing integration
  - [ ] Smooth distortion boundaries
  - [ ] Quality preservation in magnified areas
  - [ ] Artifact minimization
  - [ ] Sampling rate adjustment based on magnification
- [ ] Edge handling
  - [ ] Smooth falloff to normal transformation
  - [ ] Boundary artifact prevention
  - [ ] Edge case numerical stability
  - [ ] Clipping region integration

**Interactive Applications**

- [ ] Real-time magnification
  - [ ] Mouse-driven lens positioning
  - [ ] Zoom level control
  - [ ] Smooth lens movement
  - [ ] Performance optimization for interaction
- [ ] Document and image viewing
  - [ ] Detail inspection tools
  - [ ] Accessibility magnification
  - [ ] Scientific image analysis
  - [ ] CAD drawing detail viewing

**Rendering Integration**

- [ ] Renderer compatibility
  - [ ] Works with all pixel formats
  - [ ] Scanline renderer integration
  - [ ] Span generator compatibility
  - [ ] Multi-pass rendering support
- [ ] Performance optimization
  - [ ] Fast path for non-magnified regions
  - [ ] Incremental updates for moving lens
  - [ ] Region-of-interest optimization
  - [ ] Memory usage minimization

**Lens Physics Simulation**

- [ ] Realistic lens effects
  - [ ] Optical magnification simulation
  - [ ] Barrel and pincushion distortion
  - [ ] Chromatic aberration effects
  - [ ] Lens flare and reflection simulation
- [ ] Customizable distortion profiles
  - [ ] User-defined magnification curves
  - [ ] Asymmetric distortion patterns
  - [ ] Non-uniform magnification fields
  - [ ] Complex lens shape simulation

**Performance Considerations**

- [ ] Transformation caching
  - [ ] Pre-calculated distortion maps
  - [ ] Fast lookup tables
  - [ ] Incremental updates
  - [ ] Region-based optimization
- [ ] Real-time performance
  - [ ] Frame rate maintenance
  - [ ] Smooth animation support
  - [ ] Memory-efficient processing
  - [ ] GPU acceleration compatibility

**Mathematical Foundation**

- [ ] Lens distortion mathematics
  - [ ] Radial distortion models
  - [ ] Polynomial approximation
  - [ ] Inverse transformation calculation
  - [ ] Numerical stability considerations
- [ ] Coordinate system handling
  - [ ] Screen coordinate integration
  - [ ] World coordinate transformation
  - [ ] Multi-scale coordinate systems
  - [ ] Precision preservation

**Dependencies**

- agg_basics.h → internal/basics package
- agg_math.h → mathematical utilities
- Distance calculation utilities
- Smooth interpolation functions
- Real-time performance optimization tools

---

### Converters

#### agg_conv_adaptor_vcgen.h - Vertex Generator Adaptor (base infrastructure)

**Core Adaptor System**

- [x] conv_adaptor_vcgen → ConvAdaptorVCGen[VS, G, M] struct - Base template for vertex generator adaptors
  - [x] VertexSource template parameter → Generic over any vertex source
  - [x] Generator template parameter → Wraps vertex generators (stroke, dash, contour, etc.)
  - [x] Markers template parameter → Optional marker handling (default: null_markers)
  - [x] State machine for vertex source processing

**Null Markers Implementation**

- [x] null_markers → NullMarkers struct - Default empty marker implementation
  - [x] remove_all() → RemoveAll() - no-op marker removal
  - [x] add_vertex(x, y, cmd) → AddVertex() - no-op marker addition
  - [x] prepare_src() → PrepareSrc() - no-op source preparation
  - [x] Vertex source interface (rewind/vertex) with path_cmd_stop

**Adaptor State Management**

- [ ] Processing states enumeration → AdaptorStatus type
  - [ ] initial → Initial - starting state, no processing begun
  - [ ] accumulate → Accumulate - collecting vertices from source
  - [ ] generate → Generate - producing output from generator
- [ ] State machine implementation
  - [ ] attach(source) → Attach() - connect to new vertex source
  - [ ] rewind(path_id) → Rewind() - reset processing state
  - [ ] vertex() → Vertex() - state-driven vertex production

**Generator and Marker Access**

- [ ] Component access methods
  - [ ] generator() → Generator() - access underlying vertex generator
  - [ ] const generator() → GetGenerator() - read-only generator access
  - [ ] markers() → Markers() - access marker processor
  - [ ] const markers() → GetMarkers() - read-only marker access

**Vertex Processing Pipeline**

- [ ] Three-phase processing
  - [ ] Source accumulation: collect all vertices from input
  - [ ] Generator processing: apply transformation/generation
  - [ ] Output generation: emit processed vertices
- [ ] Path command handling
  - [ ] Preserve path structure and commands
  - [ ] Handle multi-path vertex sources
  - [ ] Path ID propagation through pipeline

**Template Specialization Support**

- [ ] Generic programming patterns
  - [ ] Compatible with all vertex source types
  - [ ] Works with any conforming generator
  - [ ] Marker system extensibility
  - [ ] Type-safe template composition

**Integration with Vertex Generators**

- [ ] Generator lifecycle management
  - [ ] Generator initialization and configuration
  - [ ] State synchronization between adaptor and generator
  - [ ] Error propagation from generator to adaptor
- [ ] Memory management
  - [ ] Efficient vertex buffering
  - [ ] Memory reuse across processing cycles
  - [ ] Minimal allocation during processing

**Performance Optimizations**

- [ ] State machine efficiency
  - [ ] Fast state transitions
  - [ ] Minimal overhead in generate state
  - [ ] Batch processing optimization
- [ ] Template instantiation
  - [ ] Compile-time optimization opportunities
  - [ ] Type erasure where beneficial
  - [ ] Inlining critical path operations

**Dependencies**

- agg_basics.h → internal/basics package
- Vertex source interface definition
- Path command constants
- Template constraint definitions

#### agg_conv_adaptor_vpgen.h - Vertex Processor Adaptor (base infrastructure) ✓

**Vertex Processor Adaptor System**

- [x] conv_adaptor_vpgen → ConvAdaptorVPGen[VPG] struct - Base template for vertex processor adaptors
  - [x] VertexSource parameter → Input vertex source type
  - [x] VPGen template parameter → Vertex processing algorithm (clipping, segmentation)
  - [x] Real-time vertex processing (no accumulation phase)
  - [x] AutoClose/AutoUnclose support for polygon handling

**Streaming Vertex Processing**

- [x] Direct vertex transformation
  - [x] No intermediate vertex storage required
  - [x] Full state management for complex polygon processing
  - [ ] Real-time processing as vertices are requested
  - [ ] Memory efficient for large paths
  - [ ] Suitable for clipping and segmentation operations

**Processor Interface Requirements**

- [ ] Processor conformance interface
  - [ ] Processor must support streaming operation
  - [ ] State management within processor
  - [ ] Path command preservation
  - [ ] Multi-path handling capability

**State Management (Simplified)**

- [ ] Simpler state model than vcgen adaptor
  - [ ] Direct pass-through to processor
  - [ ] Minimal state tracking overhead
  - [ ] Immediate vertex processing
  - [ ] Real-time error handling

**Integration Points**

- [ ] Compatible processors
  - [ ] Clipping processors (polygon, polyline)
  - [ ] Segmentation processors
  - [ ] Path modification processors
  - [ ] Custom streaming processors

**Performance Characteristics**

- [ ] Low memory overhead
  - [ ] No vertex accumulation required
  - [ ] Constant memory usage
  - [ ] Suitable for very large paths
- [ ] Real-time processing
  - [ ] Immediate vertex output
  - [ ] No processing delays
  - [ ] Interactive performance

**Dependencies**

- agg_basics.h → internal/basics package
- Vertex processor interface definitions
- Streaming processing utilities

#### agg_conv_stroke.h - Stroke Converter (path outline generation)

**Stroke Converter System**

- [ ] conv_stroke → ConvStroke[VS, M] struct - Convert paths to stroked outlines
  - [ ] VertexSource template parameter → Input path to be stroked
  - [ ] Markers template parameter → Optional stroke markers
  - [ ] Inherits from conv_adaptor_vcgen with vcgen_stroke generator
  - [ ] Comprehensive stroke parameter control

**Line Cap Style Configuration**

- [ ] Line cap enumeration → LineCap type
  - [ ] butt_cap → ButtCap - flat end perpendicular to path
  - [ ] square_cap → SquareCap - square extension beyond path end
  - [ ] round_cap → RoundCap - circular end centered on path end
- [ ] Cap style methods
  - [ ] line_cap(cap_style) → SetLineCap() - set end cap style
  - [ ] line_cap() → GetLineCap() - get current cap style

**Line Join Style Configuration**

- [ ] Line join enumeration → LineJoin type
  - [ ] miter_join → MiterJoin - sharp corner with miter limit
  - [ ] miter_join_revert → MiterJoinRevert - fallback to bevel when limit exceeded
  - [ ] round_join → RoundJoin - circular arc at corners
  - [ ] bevel_join → BevelJoin - flat cut across corner
- [ ] Join style methods
  - [ ] line_join(join_style) → SetLineJoin() - set corner join style
  - [ ] line_join() → GetLineJoin() - get current join style

**Inner Join Handling**

- [ ] Inner join enumeration → InnerJoin type
  - [ ] inner_bevel → InnerBevel - beveled inner joins
  - [ ] inner_miter → InnerMiter - mitered inner joins
  - [ ] inner_jag → InnerJag - jagged inner joins (fastest)
  - [ ] inner_round → InnerRound - rounded inner joins (smoothest)
- [ ] Inner join methods
  - [ ] inner_join(inner_style) → SetInnerJoin() - set inner corner style
  - [ ] inner_join() → GetInnerJoin() - get current inner style

**Stroke Width and Measurement**

- [ ] Width control
  - [ ] width(w) → SetWidth() - set stroke line width
  - [ ] width() → GetWidth() - get current stroke width
  - [ ] Width measurement in user coordinate units
  - [ ] Zero-width stroke handling (hairline rendering)

**Miter Limit Control**

- [ ] Miter limiting to prevent excessive spikes
  - [ ] miter_limit(limit) → SetMiterLimit() - set miter limit ratio
  - [ ] miter_limit() → GetMiterLimit() - get current miter limit
  - [ ] miter_limit_theta(angle) → SetMiterLimitTheta() - set limit by angle
  - [ ] inner_miter_limit(limit) → SetInnerMiterLimit() - inner corner limit

**Path Shortening**

- [ ] Path end modification
  - [ ] shorten(amount) → SetShorten() - shorten path ends by specified amount
  - [ ] shorten() → GetShorten() - get current shortening amount
  - [ ] Symmetric shortening of both path ends
  - [ ] Useful for arrow and marker integration

**Quality and Approximation Control**

- [ ] Rendering quality parameters
  - [ ] approximation_scale(scale) → SetApproximationScale() - control curve detail
  - [ ] approximation_scale() → GetApproximationScale() - get current scale
  - [ ] Affects circular arc approximation in round caps/joins
  - [ ] Balance between quality and performance

**Advanced Stroke Features**

- [ ] Complex stroke scenarios
  - [ ] Self-intersecting path handling
  - [ ] Zero-length segment handling
  - [ ] Degenerate path processing
  - [ ] Path direction preservation

**Generator Integration**

- [ ] Underlying vcgen_stroke access
  - [ ] All methods delegate to base vcgen_stroke generator
  - [ ] Type-safe parameter forwarding
  - [ ] State management through adaptor base class
  - [ ] Marker integration support

**Performance Considerations**

- [ ] Stroke generation efficiency
  - [ ] Optimized for common stroke parameters
  - [ ] Caching opportunities for static strokes
  - [ ] Memory efficient vertex generation
  - [ ] Fast path for simple stroke cases

**Integration with Path Pipeline**

- [ ] Composability with other converters
  - [ ] Can be chained with other path modifications
  - [ ] Compatible with dash, marker, and transform converters
  - [ ] Works with all vertex source types
  - [ ] Output compatible with all rendering systems

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_stroke.h → stroke vertex generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- Line style enumeration definitions

#### agg_conv_dash.h - Dash Converter (dashed line patterns)

**Dash Pattern System**

- [ ] conv_dash → ConvDash[VS, M] struct - Add dash patterns to paths
  - [ ] VertexSource template parameter → Input path to be dashed
  - [ ] Markers template parameter → Optional dash markers
  - [ ] Inherits from conv_adaptor_vcgen with vcgen_dash generator
  - [ ] Flexible dash pattern definition

**Dash Pattern Management**

- [ ] Pattern definition methods
  - [ ] remove_all_dashes() → RemoveAllDashes() - clear all dash patterns
  - [ ] add_dash(dash_len, gap_len) → AddDash() - add dash/gap pair to pattern
  - [ ] Multiple dash patterns create complex repeating sequences
  - [ ] Pattern length calculated as sum of all dash/gap pairs

**Dash Pattern Phase Control**

- [ ] Pattern positioning
  - [ ] dash_start(offset) → SetDashStart() - set starting offset in pattern
  - [ ] Phase control for pattern alignment
  - [ ] Offset wraps around pattern length
  - [ ] Useful for animation and pattern synchronization

**Path Length and Shortening**

- [ ] Path modification
  - [ ] shorten(amount) → SetShorten() - shorten path before dashing
  - [ ] shorten() → GetShorten() - get current shortening amount
  - [ ] Applied before dash pattern calculation
  - [ ] Useful for precise pattern termination

**Dash Generation Algorithm**

- [ ] Pattern application process
  - [ ] Path length calculation and parameterization
  - [ ] Pattern repetition across path length
  - [ ] Dash segment extraction from continuous path
  - [ ] Gap handling (no vertex output during gaps)

**Complex Path Handling**

- [ ] Multi-segment path support
  - [ ] Pattern continues across path segments
  - [ ] Pattern phase preservation at path connections
  - [ ] Proper handling of path commands (move_to, line_to, curves)
  - [ ] Closed path pattern continuity

**Integration with Stroke Converter**

- [ ] Common usage pattern: dash then stroke
  - [ ] conv_stroke<conv_dash<path_storage>> composition
  - [ ] Dash patterns applied before stroke generation
  - [ ] Stroke parameters applied to individual dash segments
  - [ ] Line caps applied to each dash segment

**Performance Optimization**

- [ ] Efficient dash generation
  - [ ] Arc-length parameterization for accurate spacing
  - [ ] Minimal vertex generation during gaps
  - [ ] Pattern calculation optimization
  - [ ] Incremental pattern processing

**Animation Support**

- [ ] Dynamic dash patterns
  - [ ] Animated dash_start offset for moving patterns
  - [ ] Real-time pattern modification
  - [ ] Smooth dash animation effects
  - [ ] Interactive dash parameter adjustment

**Edge Case Handling**

- [ ] Robust pattern processing
  - [ ] Zero-length dash or gap handling
  - [ ] Very short path segments
  - [ ] Pattern longer than path handling
  - [ ] Degenerate path processing

**Generator Integration**

- [ ] Underlying vcgen_dash access
  - [ ] Parameter delegation to base generator
  - [ ] State management through adaptor framework
  - [ ] Memory efficient dash segment storage
  - [ ] Pattern calculation caching

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_dash.h → dash vertex generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- Arc-length calculation utilities

#### agg_conv_contour.h - Contour Converter (path offset generation)

**Contour Generation System**

- [ ] conv_contour → ConvContour[VS] struct - Generate offset contours from paths
  - [ ] VertexSource template parameter → Input path for contour generation
  - [ ] Inherits from conv_adaptor_vcgen with vcgen_contour generator
  - [ ] Parallel curve generation with configurable offset distance
  - [ ] Support for both expansion and contraction

**Contour Offset Control**

- [ ] Distance configuration
  - [ ] width(distance) → SetWidth() - set contour offset distance
  - [ ] width() → GetWidth() - get current offset distance
  - [ ] Positive values expand path outward
  - [ ] Negative values contract path inward
  - [ ] Zero width returns original path

**Line Join Handling for Contours**

- [ ] Corner processing
  - [ ] line_join(join_style) → SetLineJoin() - set contour join style
  - [ ] line_join() → GetLineJoin() - get current join style
  - [ ] Similar to stroke joins but for offset curves
  - [ ] Critical for smooth contour appearance

**Inner Join Processing**

- [ ] Inner corner handling
  - [ ] inner_join(inner_style) → SetInnerJoin() - set inner join style
  - [ ] inner_join() → GetInnerJoin() - get current inner style
  - [ ] Important for path contraction (negative offsets)
  - [ ] Prevents self-intersection in concave regions

**Miter Limit for Contours**

- [ ] Sharp corner control
  - [ ] miter_limit(limit) → SetMiterLimit() - set miter limit for contour joins
  - [ ] miter_limit() → GetMiterLimit() - get current miter limit
  - [ ] Prevents excessive spikes in sharp corners
  - [ ] Automatic fallback to bevel when limit exceeded

**Approximation Quality**

- [ ] Curve quality control
  - [ ] approximation_scale(scale) → SetApproximationScale() - control curve detail
  - [ ] approximation_scale() → GetApproximationScale() - get current scale
  - [ ] Affects circular arc approximation in rounded joins
  - [ ] Balance between smoothness and performance

**Contour Applications**

- [ ] Text outline generation
  - [ ] Create outlined text effects
  - [ ] Multi-level text outlines (nested contours)
  - [ ] Bold text simulation through expansion
- [ ] Shape morphing effects
  - [ ] Smooth shape expansion/contraction animations
  - [ ] Organic growth/shrink effects
  - [ ] Boundary visualization and emphasis

**Advanced Contour Processing**

- [ ] Complex path handling
  - [ ] Self-intersecting contour resolution
  - [ ] Island detection and processing
  - [ ] Hole preservation in contracted paths
  - [ ] Multiple contour level generation

**Integration with Other Converters**

- [ ] Converter chaining
  - [ ] Can be combined with stroke for outlined contours
  - [ ] Works with dash patterns for dashed outlines
  - [ ] Compatible with curve and smoothing converters
  - [ ] Transform integration for scaled contours

**Performance Considerations**

- [ ] Efficient offset calculation
  - [ ] Geometric algorithms for parallel curves
  - [ ] Memory efficient vertex generation
  - [ ] Caching for repeated contour operations
  - [ ] Optimization for simple geometric shapes

**Error Handling**

- [ ] Robust contour generation
  - [ ] Degenerate path handling
  - [ ] Very small offset distances
  - [ ] Self-intersecting input paths
  - [ ] Numerical stability in edge cases

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_contour.h → contour vertex generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- Geometric calculation utilities for parallel curves

#### agg_conv_curve.h - Curve Converter (curve-to-line approximation)

**Curve Approximation System**

- [x] conv_curve → ConvCurve[VS] struct - Convert path curves to line segments
  - [x] VertexSource template parameter → Input path with curves
  - [x] Direct curve processing without vertex generator adaptor
  - [x] Automatic curve detection and approximation
  - [x] Configurable approximation quality

**Curve Detection and Processing**

- [x] Curve command handling
  - [x] curve3 command → Quadratic Bezier curve approximation
  - [x] curve4 command → Cubic Bezier curve approximation
  - [x] Automatic curve type detection
  - [x] Preserves non-curve path commands (move_to, line_to, etc.)

**Approximation Quality Control**

- [x] Quality parameters
  - [x] approximation_method(method) → SetApproximationMethod() - subdivision method
  - [x] approximation_scale(scale) → SetApproximationScale() - detail level
  - [x] angle_tolerance(angle) → SetAngleTolerance() - curvature sensitivity
  - [x] cusp_limit(limit) → SetCuspLimit() - sharp corner threshold

**Approximation Methods**

- [ ] Curve subdivision techniques
  - [ ] curve_inc → CurveInc - incremental method (faster)
  - [ ] curve_div → CurveDiv - recursive subdivision (higher quality)
  - [ ] Automatic method selection based on curve properties
  - [ ] Method switching for optimal performance/quality trade-off

**Curve-Specific Parameters**

- [ ] Curve type handling
  - [ ] Different parameter sets for quadratic vs cubic curves
  - [ ] Specialized processing for each curve type
  - [ ] Adaptive parameter adjustment based on curve complexity
  - [ ] Consistent approximation quality across curve types

**Integration with Path Pipeline**

- [ ] Pipeline position
  - [ ] Typically applied early in conversion chain
  - [ ] Converts complex curves to simple line segments
  - [ ] Enables downstream converters to work with linearized paths
  - [ ] Maintains path structure and semantics

**Performance Optimization**

- [ ] Efficient curve processing
  - [ ] Curve caching for repeated approximations
  - [ ] Adaptive subdivision based on viewport scale
  - [ ] Fast path for already-linear paths
  - [ ] Memory efficient line segment generation

**Quality vs Performance Trade-offs**

- [ ] Approximation tuning
  - [ ] Scale-dependent approximation quality
  - [ ] Interactive vs high-quality rendering modes
  - [ ] Automatic quality adjustment based on curve size
  - [ ] Performance profiling and optimization

**Advanced Curve Handling**

- [ ] Complex curve scenarios
  - [ ] Self-intersecting curves
  - [ ] Degenerate curves (zero-length control segments)
  - [ ] Curves with coincident control points
  - [ ] Very high curvature curves

**Rendering Integration**

- [ ] Renderer compatibility
  - [ ] Works with all rasterizers expecting line segments
  - [ ] Anti-aliasing preservation during approximation
  - [ ] Consistent stroke width along approximated curves
  - [ ] Smooth gradient application to curved paths

**Curve Length Approximation**

- [ ] Accurate curved path length measurement
  - [ ] Bezier curve length estimation using curve approximation
  - [ ] Integration with ConvCurve for automatic curve-to-line conversion
  - [ ] Adaptive subdivision accuracy based on approximation scale
  - [ ] PathLengthCurved() function that handles both straight and curved segments

- [ ] Length calculation algorithm
  - [ ] Detect curve commands (curve3, curve4) in vertex stream
  - [ ] Use ConvCurve to approximate curves as line segments
  - [ ] Sum line segment lengths for accurate curved path measurement
  - [ ] Maintain existing straight-line performance for non-curved paths

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_vertex_sequence.h → vertex sequence generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- agg_curves.h → curve approximation algorithms

#### agg_conv_bspline.h - B-Spline Converter (smooth curve generation)

**B-Spline Conversion System**

- [x] conv_bspline → ConvBSpline[VS] struct - Convert path to B-spline curves
  - [x] VertexSource template parameter → Input path with control points
  - [x] Inherits from conv_adaptor_vcgen with vcgen_bspline
  - [x] Smooth curve generation through control points
  - [x] Configurable spline parameters

**Spline Parameter Control**

- [x] B-spline configuration
  - [x] interpolation_step(step) → SetInterpolationStep() - distance between output points
  - [x] interpolation_step() → GetInterpolationStep() - get current step size
  - [x] Controls density of generated curve points
  - [x] Smaller steps produce smoother curves

**Curve Quality Settings**

- [ ] Smoothness control
  - [ ] Automatic tangent calculation at control points
  - [ ] Smooth curve transitions between control points
  - [ ] Configurable curve tension (if supported)
  - [ ] Endpoint behavior control (clamped vs natural)

**Control Point Processing**

- [ ] Path interpretation
  - [ ] Path vertices become B-spline control points
  - [ ] line_to commands define control point sequence
  - [ ] move_to commands start new spline curves
  - [ ] Proper handling of multi-path input

**Spline Generation Algorithm**

- [ ] B-spline mathematics
  - [ ] Cubic B-spline basis functions
  - [ ] Control point weight calculation
  - [ ] Knot vector generation (uniform spacing)
  - [ ] Curve parameter to output point mapping

**Applications**

- [ ] Smooth curve creation
  - [ ] Convert angular paths to smooth curves
  - [ ] Data visualization with smooth interpolation
  - [ ] Animation paths with smooth motion
  - [ ] Organic shape generation from rough sketches

**Integration with Other Converters**

- [ ] Converter pipeline integration
  - [ ] Can be combined with stroke for smooth stroked curves
  - [ ] Works with dash patterns for smooth dashed curves
  - [ ] Compatible with curve converter for linearization
  - [ ] Transform integration for scaled splines

**Performance Considerations**

- [ ] Spline calculation efficiency
  - [ ] Coefficient caching for static control points
  - [ ] Incremental spline evaluation
  - [ ] Memory efficient point generation
  - [ ] Adaptive quality based on curve scale

**Advanced Spline Features**

- [ ] Complex spline handling
  - [ ] Multiple spline segments in single path
  - [ ] Closed spline curve generation
  - [ ] Spline continuity control (C0, C1, C2)
  - [ ] Custom knot vector support (if needed)

**Quality Control**

- [ ] Spline approximation
  - [ ] Consistent point spacing along curve
  - [ ] Curvature-adaptive point density
  - [ ] Sharp corner detection and handling
  - [ ] Smooth transitions between spline segments

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_bspline.h → B-spline vertex generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- agg_bspline.h → B-spline mathematical functions

#### agg_conv_smooth_poly1.h - Polygon Smoothing Converter

**Polygon Smoothing System**

- [x] conv_smooth_poly1 → ConvSmoothPoly1[VS] struct - Smooth polygon corners
  - [x] VertexSource template parameter → Input polygon path
  - [x] Inherits from conv_adaptor_vcgen with vcgen_smooth_poly1
  - [x] Corner rounding and smoothing effects
  - [x] Configurable smoothing parameters

**Smoothing Parameters**

- [x] Smoothing control
  - [x] smooth_value(value) → SetSmoothValue() - smoothing amount (0.0 to 1.0)
  - [x] smooth_value() → GetSmoothValue() - get current smoothing level
  - [x] 0.0 produces original polygon, 1.0 produces maximum smoothing
  - [x] Continuous control over smoothing intensity

**Corner Processing Algorithm**

- [ ] Corner detection and modification
  - [ ] Automatic corner angle calculation
  - [ ] Bezier curve generation for rounded corners
  - [ ] Smooth curve insertion at polygon vertices
  - [ ] Preservation of overall polygon shape

**Path Handling**

- [ ] Polygon path processing
  - [ ] Works with both open and closed polygons
  - [ ] Preserves path structure and commands
  - [ ] Maintains polygon orientation
  - [ ] Handles multi-polygon input

**Smoothing Quality**

- [ ] Curve approximation
  - [ ] High-quality Bezier curve approximation
  - [ ] Consistent smoothing across all corners
  - [ ] Adaptive curve density based on corner angle
  - [ ] Smooth transitions between straight and curved segments

**Applications**

- [ ] Shape refinement
  - [ ] Convert sharp-cornered shapes to organic forms
  - [ ] Logo and icon smoothing
  - [ ] Artistic shape modification
  - [ ] UI element softening (rounded rectangles, etc.)

**Advanced Features**

- [ ] Selective smoothing
  - [ ] Different smoothing levels for different corner angles
  - [ ] Sharp corner preservation (below threshold)
  - [ ] Adaptive smoothing based on polygon properties
  - [ ] Custom smoothing profiles

**Integration Capabilities**

- [ ] Converter pipeline
  - [ ] Can be combined with other path modifications
  - [ ] Works well before stroke converter for smooth outlines
  - [ ] Compatible with contour converter for smooth offsets
  - [ ] Integrates with transform operations

**Performance Optimization**

- [ ] Efficient smoothing
  - [ ] Corner detection optimization
  - [ ] Bezier calculation caching
  - [ ] Memory efficient curve generation
  - [ ] Fast path for low smoothing values

**Quality Control**

- [ ] Smoothing precision
  - [ ] Consistent curve quality across smoothing levels
  - [ ] Numerical stability for extreme smoothing values
  - [ ] Preservation of polygon area (approximately)
  - [ ] Smooth derivative transitions

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vcgen_smooth_poly1.h → polygon smoothing generator
- agg_conv_adaptor_vcgen.h → base adaptor system
- Bezier curve generation utilities

#### agg_conv_clip_polygon.h - Polygon Clipping Converter

**Polygon Clipping System**

- [ ] conv_clip_polygon → ConvClipPolygon[VS] struct - Clip polygons to rectangular bounds
  - [ ] VertexSource template parameter → Input polygon path to be clipped
  - [ ] Inherits from conv_adaptor_vpgen with vpgen_clip_polygon processor
  - [ ] Rectangular clipping window definition
  - [ ] Cohen-Sutherland or similar clipping algorithm

**Clipping Window Definition**

- [ ] Clipping rectangle setup
  - [ ] clip_box(x1, y1, x2, y2) → SetClipBox() - define rectangular clipping region
  - [ ] clip_box() → GetClipBox() - get current clipping bounds
  - [ ] Window coordinates in user coordinate system
  - [ ] Automatic coordinate ordering (ensures x1 < x2, y1 < y2)

**Polygon Clipping Algorithm**

- [ ] Clipping methodology
  - [ ] Sutherland-Hodgman polygon clipping algorithm
  - [ ] Edge-by-edge clipping against window boundaries
  - [ ] Proper handling of polygon vertices and edges
  - [ ] Generation of new intersection vertices

**Clipped Output Generation**

- [ ] Result polygon creation
  - [ ] Maintains polygon structure in output
  - [ ] Proper path command generation for clipped polygons
  - [ ] Handles multiple disjoint polygon pieces
  - [ ] Preserves polygon winding direction

**Complex Polygon Handling**

- [ ] Advanced clipping scenarios
  - [ ] Self-intersecting polygon clipping
  - [ ] Concave polygon support
  - [ ] Multiple polygon processing
  - [ ] Degenerate case handling (polygon entirely outside clip region)

**Performance Optimizations**

- [ ] Efficient clipping
  - [ ] Early rejection for polygons entirely outside clip region
  - [ ] Early acceptance for polygons entirely inside clip region
  - [ ] Incremental vertex processing
  - [ ] Memory efficient intersection calculation

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vpgen_clip_polygon.h → polygon clipping processor
- agg_conv_adaptor_vpgen.h → vertex processor adaptor
- Geometric intersection utilities

#### agg_conv_clip_polyline.h - Polyline Clipping Converter

**Polyline Clipping System**

- [ ] conv_clip_polyline → ConvClipPolyline[VS] struct - Clip polylines to rectangular bounds
  - [ ] VertexSource template parameter → Input polyline path to be clipped
  - [ ] Inherits from conv_adaptor_vpgen with vpgen_clip_polyline processor
  - [ ] Line segment clipping with proper endpoint handling
  - [ ] Cohen-Sutherland line clipping algorithm

**Clipping Window Configuration**

- [ ] Clipping bounds setup
  - [ ] clip_box(x1, y1, x2, y2) → SetClipBox() - define rectangular clipping region
  - [ ] clip_box() → GetClipBox() - get current clipping bounds
  - [ ] Coordinate system alignment with polygon clipping
  - [ ] Window boundary precision handling

**Line Segment Clipping**

- [ ] Segment-by-segment processing
  - [ ] Cohen-Sutherland outcodes for efficient clipping
  - [ ] Line-rectangle intersection calculation
  - [ ] Proper clipped segment endpoint generation
  - [ ] Handling of line segments crossing multiple window boundaries

**Multi-segment Path Handling**

- [ ] Complex polyline processing
  - [ ] Maintains path structure across clipping operations
  - [ ] Proper path command generation for clipped segments
  - [ ] Handles disconnected line segments after clipping
  - [ ] Path continuity management

**Clipping Edge Cases**

- [ ] Special case handling
  - [ ] Lines entirely outside clipping region (no output)
  - [ ] Lines entirely inside clipping region (pass through)
  - [ ] Lines tangent to clipping window boundaries
  - [ ] Zero-length line segments

**Performance Features**

- [ ] Optimized line clipping
  - [ ] Fast accept/reject using outcodes
  - [ ] Minimal intersection calculations
  - [ ] Efficient vertex generation
  - [ ] Memory-friendly processing for long polylines

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vpgen_clip_polyline.h → polyline clipping processor
- agg_conv_adaptor_vpgen.h → vertex processor adaptor
- Line-rectangle intersection algorithms

#### agg_conv_gpc.h - General Polygon Clipper Converter

**GPC Integration System**

- [ ] conv_gpc → ConvGPC[VS] struct - General Polygon Clipper integration
  - [ ] VertexSource template parameter → Input polygon for boolean operations
  - [ ] Integration with GPC (General Polygon Clipper) library
  - [ ] Support for complex boolean operations on polygons
  - [ ] Industrial-strength polygon clipping capabilities

**Boolean Operations**

- [ ] Polygon boolean operations
  - [ ] Union (OR) - combine polygon areas
  - [ ] Intersection (AND) - common polygon areas only
  - [ ] Difference (A - B) - subtract second polygon from first
  - [ ] XOR (exclusive OR) - symmetric difference of polygons

**GPC Library Integration**

- [ ] External library interface
  - [ ] GPC polygon structure conversion
  - [ ] Memory management for GPC data structures
  - [ ] Error handling and status reporting
  - [ ] License compliance considerations (GPC has specific licensing)

**Complex Polygon Support**

- [ ] Advanced polygon features
  - [ ] Self-intersecting polygon handling
  - [ ] Polygon with holes support
  - [ ] Multiple contour polygons
  - [ ] Arbitrary polygon complexity

**Performance Considerations**

- [ ] GPC performance characteristics
  - [ ] Memory usage for complex polygons
  - [ ] Processing time for large polygon sets
  - [ ] Optimization strategies for common cases
  - [ ] Alternative lightweight clipping for simple cases

**Integration Challenges**

- [ ] Library compatibility
  - [ ] C library integration in Go
  - [ ] Memory management across language boundaries
  - [ ] Error propagation from C library
  - [ ] Platform-specific compilation considerations

**Dependencies**

- agg_basics.h → internal/basics package
- gpc.h and gpc.c → General Polygon Clipper library
- C library integration utilities
- Memory management for cross-language data

#### agg_conv_close_polygon.h - Polygon Closing Converter

**Polygon Closing System**

- [ ] conv_close_polygon → ConvClosePolygon[VS] struct - Ensure polygons are properly closed
  - [ ] VertexSource template parameter → Input path that may have unclosed polygons
  - [ ] Automatic detection of unclosed polygon paths
  - [ ] Addition of closing line segments where needed
  - [ ] Preservation of already-closed polygons

**Polygon Closure Detection**

- [ ] Path analysis
  - [ ] Detection of polygon paths (sequences ending with end_poly)
  - [ ] Comparison of first and last vertex coordinates
  - [ ] Epsilon-based coordinate comparison for floating-point precision
  - [ ] Path command sequence analysis

**Automatic Closure Generation**

- [ ] Path modification
  - [ ] Insertion of line_to command from last to first vertex
  - [ ] Proper close_polygon command generation
  - [ ] Preservation of polygon orientation (clockwise/counter-clockwise)
  - [ ] Path command sequence integrity

**Multi-polygon Support**

- [ ] Complex path handling
  - [ ] Multiple polygon processing in single path
  - [ ] Independent closure analysis for each polygon
  - [ ] Preservation of path structure and organization
  - [ ] Proper handling of move_to commands between polygons

**Integration with Rendering**

- [ ] Renderer compatibility
  - [ ] Ensures proper polygon filling
  - [ ] Eliminates gaps in polygon outlines
  - [ ] Compatible with all fill rule algorithms
  - [ ] Works with stroke converters for complete outlines

**Performance Optimization**

- [ ] Efficient closure processing
  - [ ] Single-pass path analysis
  - [ ] Minimal vertex coordinate comparison
  - [ ] Fast path for already-closed polygons
  - [ ] Memory-efficient path modification

**Dependencies**

- agg_basics.h → internal/basics package
- Path command definitions and utilities
- Coordinate comparison utilities with epsilon handling

#### agg_conv_unclose_polygon.h - Polygon Unclosing Converter

**Polygon Unclosing System**

- [ ] conv_unclose_polygon → ConvUnclosePolygon[VS] struct - Remove polygon closing commands
  - [ ] VertexSource template parameter → Input path with closed polygons
  - [ ] Detection and removal of polygon closing commands
  - [ ] Conversion of closed polygons to open polylines
  - [ ] Preservation of polygon shape without closure

**Closure Command Removal**

- [ ] Path command filtering
  - [ ] Detection of close_polygon commands
  - [ ] Removal of explicit closing line segments
  - [ ] Conversion of end_poly with close flag to end_poly without close flag
  - [ ] Maintenance of path vertex sequence

**Path Structure Preservation**

- [ ] Path integrity maintenance
  - [ ] Preservation of all polygon vertices except closure
  - [ ] Maintenance of path organization and multi-polygon structure
  - [ ] Proper handling of path IDs and command sequences
  - [ ] Coordinate precision preservation

**Applications**

- [ ] Open path creation
  - [ ] Convert polygons to polylines for stroke-only rendering
  - [ ] Prepare paths for dash pattern application
  - [ ] Create open paths for text-on-path applications
  - [ ] Generate paths for marker placement

**Integration Patterns**

- [ ] Common usage scenarios
  - [ ] Used before dash converter to avoid closing dashes
  - [ ] Applied before certain marker converters
  - [ ] Useful in path editing applications
  - [ ] Compatible with all other path converters

**Performance Features**

- [ ] Efficient processing
  - [ ] Single-pass path command filtering
  - [ ] Minimal memory overhead
  - [ ] Fast identification of closing commands
  - [ ] Stream-friendly processing

**Dependencies**

- agg_basics.h → internal/basics package
- Path command definitions and filtering utilities

#### agg_conv_concat.h - Path Concatenation Converter

**Path Concatenation System**

- [ ] conv_concat → ConvConcat[VS1, VS2] struct - Concatenate multiple vertex sources
  - [ ] Multiple VertexSource template parameters → Input paths to concatenate
  - [ ] Sequential path output from multiple sources
  - [ ] Proper path command sequence management
  - [ ] Support for arbitrary number of source paths

**Multi-source Management**

- [ ] Source path handling
  - [ ] Sequential processing of input vertex sources
  - [ ] Automatic source switching at path completion
  - [ ] Path ID management across multiple sources
  - [ ] State management for source transitions

**Path Command Sequencing**

- [ ] Command stream management
  - [ ] Proper command sequence across source boundaries
  - [ ] Path continuity or separation control
  - [ ] end_poly command handling between sources
  - [ ] move_to command insertion for path separation

**Concatenation Modes**

- [ ] Different concatenation strategies
  - [ ] Continuous concatenation (paths connected)
  - [ ] Separate concatenation (paths as distinct entities)
  - [ ] Custom separation control
  - [ ] Path orientation preservation

**Applications**

- [ ] Complex path construction
  - [ ] Combine multiple path segments into single path
  - [ ] Merge paths from different sources
  - [ ] Create composite shapes from multiple components
  - [ ] Path assembly for complex graphics

**Performance Considerations**

- [ ] Efficient concatenation
  - [ ] Minimal overhead for source switching
  - [ ] Memory-efficient vertex streaming
  - [ ] Fast source completion detection
  - [ ] Optimized for large numbers of source paths

**Dependencies**

- agg_basics.h → internal/basics package
- Multiple vertex source management utilities
- Path command sequencing logic

#### agg_conv_shorten_path.h - Path Shortening Converter

**Path Shortening System**

- [ ] conv_shorten_path → ConvShortenPath[VS] struct - Shorten paths by removing segments from ends
  - [ ] VertexSource template parameter → Input path to be shortened
  - [ ] Configurable shortening amounts for path start and end
  - [ ] Precise length-based path modification
  - [ ] Preservation of path shape and direction

**Shortening Configuration**

- [ ] Length control parameters
  - [ ] shorten(amount) → SetShorten() - symmetric shortening amount
  - [ ] shorten_start(amount) → SetShortenStart() - shortening from path start
  - [ ] shorten_end(amount) → SetShortenEnd() - shortening from path end
  - [ ] Independent control of start and end shortening

**Path Length Calculation**

- [ ] Arc-length measurement
  - [ ] Precise path length calculation
  - [ ] Cumulative distance tracking along path
  - [ ] Curve length approximation for curved paths
  - [ ] Segment-by-segment length accumulation

**Shortening Algorithm**

- [ ] Path modification process
  - [ ] Forward shortening from path start
  - [ ] Backward shortening from path end
  - [ ] Interpolation for partial segment removal
  - [ ] New vertex generation at cut points

**Complex Path Handling**

- [ ] Multi-segment path support
  - [ ] Shortening across multiple path segments
  - [ ] Proper handling of move_to commands
  - [ ] Preservation of path structure
  - [ ] Multi-path processing

**Edge Case Management**

- [ ] Boundary condition handling
  - [ ] Shortening amount exceeding path length
  - [ ] Zero-length path handling
  - [ ] Very short path segments
  - [ ] Degenerate path processing

**Applications**

- [ ] Path end modification
  - [ ] Arrow integration (shorten path for arrowhead placement)
  - [ ] Marker positioning (create space for end markers)
  - [ ] Path trimming for aesthetic purposes
  - [ ] Animation effects (growing/shrinking paths)

**Performance Optimization**

- [ ] Efficient shortening
  - [ ] Incremental length calculation
  - [ ] Early termination for excessive shortening
  - [ ] Memory-efficient vertex generation
  - [ ] Caching for repeated operations

**Dependencies**

- agg_basics.h → internal/basics package
- agg_shorten_path.h → path shortening utilities
- Arc-length calculation algorithms

#### agg_conv_segmentator.h - Segmentator Converter

**Path Segmentation System**

- [ ] conv_segmentator → ConvSegmentator[VS] struct - Segment paths into equal-length pieces
  - [ ] VertexSource template parameter → Input path to be segmented
  - [ ] Inherits from conv_adaptor_vpgen with vpgen_segmentator processor
  - [ ] Configurable segment length
  - [ ] Even spacing along path curves and lines

**Segmentation Parameters**

- [ ] Segment control
  - [ ] approximation_scale(scale) → SetApproximationScale() - control segment density
  - [ ] approximation_scale() → GetApproximationScale() - get current scale
  - [ ] Uniform segment length along entire path
  - [ ] Adaptive segmentation based on path curvature

**Path Parameterization**

- [ ] Arc-length parameterization
  - [ ] Conversion of path to arc-length parameter space
  - [ ] Even spacing calculation along path
  - [ ] Interpolation between original path vertices
  - [ ] Precise positioning of segment boundaries

**Segment Generation**

- [ ] Uniform vertex output
  - [ ] Regular vertex spacing along path
  - [ ] Interpolated coordinates for segment points
  - [ ] Preservation of path direction and orientation
  - [ ] Smooth segment transitions

**Applications**

- [ ] Path sampling and analysis
  - [ ] Create evenly-spaced sample points along paths
  - [ ] Animation keyframe generation
  - [ ] Path analysis and measurement
  - [ ] Uniform marker placement preparation

**Advanced Features**

- [ ] Segmentation quality control
  - [ ] Adaptive segment length based on curvature
  - [ ] Minimum/maximum segment length constraints
  - [ ] Corner and cusp handling
  - [ ] Smooth segment transitions at path joints

**Performance Considerations**

- [ ] Efficient segmentation
  - [ ] Incremental arc-length calculation
  - [ ] Memory-efficient parameter space conversion
  - [ ] Fast interpolation algorithms
  - [ ] Optimized for long paths

**Dependencies**

- agg_basics.h → internal/basics package
- agg_vpgen_segmentator.h → segmentation processor
- agg_conv_adaptor_vpgen.h → vertex processor adaptor
- Arc-length parameterization utilities

#### agg_conv_marker.h - Marker Converter

**Marker Placement System**

- [ ] conv_marker → ConvMarker[VS, ML, MS] struct - Place markers along paths
  - [ ] VertexSource template parameter → Input path for marker placement
  - [ ] MarkerLocator template parameter → Marker positioning strategy
  - [ ] MarkerShape template parameter → Marker geometry definition
  - [ ] Inherits from conv_adaptor_vcgen with marker generation

**Marker Locator Strategies**

- [ ] Positioning algorithms
  - [ ] Distance-based placement (every N units along path)
  - [ ] Vertex-based placement (at path vertices)
  - [ ] Even distribution (N markers evenly spaced)
  - [ ] Custom placement patterns

**Marker Shape Integration**

- [ ] Shape definition and rendering
  - [ ] Marker geometry from vertex source
  - [ ] Automatic marker orientation along path direction
  - [ ] Scale factor application
  - [ ] Marker coordinate system transformation

**Path Analysis for Marker Placement**

- [ ] Path processing
  - [ ] Arc-length calculation for distance-based placement
  - [ ] Path direction calculation for marker orientation
  - [ ] Curvature analysis for marker scaling
  - [ ] Path segment analysis for placement optimization

**Advanced Marker Features**

- [ ] Complex marker behaviors
  - [ ] Marker collision detection and avoidance
  - [ ] Variable marker size along path
  - [ ] Marker fade-in/fade-out effects
  - [ ] Custom marker transformation functions

**Performance Optimization**

- [ ] Efficient marker generation
  - [ ] Marker shape caching and reuse
  - [ ] Batch marker transformation
  - [ ] Memory-efficient marker geometry storage
  - [ ] Optimized placement calculation

**Dependencies**

- agg_basics.h → internal/basics package
- agg_conv_adaptor_vcgen.h → vertex generator adaptor
- Marker locator and shape interfaces
- Path analysis and transformation utilities

#### agg_conv_marker_adaptor.h - Marker Adaptor Converter

**Marker Adaptor System**

- [ ] conv_marker_adaptor → ConvMarkerAdaptor[VS, M] struct - Adaptor for custom marker systems
  - [ ] VertexSource template parameter → Input path for marker placement
  - [ ] Markers template parameter → Custom marker implementation
  - [ ] Bridge between path processing and marker systems
  - [ ] Flexible marker integration framework

**Custom Marker Integration**

- [ ] Marker system compatibility
  - [ ] Support for user-defined marker implementations
  - [ ] Marker lifecycle management
  - [ ] State synchronization between path and markers
  - [ ] Custom marker parameter passing

**Marker Event Handling**

- [ ] Path-to-marker communication
  - [ ] Path vertex events to marker system
  - [ ] Path command interpretation for markers
  - [ ] Marker preparation and cleanup phases
  - [ ] Error handling and recovery

**Flexible Marker Framework**

- [ ] Extensible marker support
  - [ ] Template-based marker customization
  - [ ] Runtime marker behavior modification
  - [ ] Multiple marker type support
  - [ ] Marker composition and layering

**Performance Features**

- [ ] Efficient marker processing
  - [ ] Minimal overhead marker integration
  - [ ] Lazy marker evaluation
  - [ ] Memory-efficient marker state management
  - [ ] Optimized marker update cycles

**Dependencies**

- agg_basics.h → internal/basics package
- Marker interface definitions
- Vertex source integration utilities

#### agg_conv_transform.h - Transform Converter

**Transform Converter System**

- [ ] conv_transform → ConvTransform[VS, Trans] struct - Apply transformations to vertex sources
  - [ ] VertexSource template parameter → Input path to be transformed
  - [ ] Transformer template parameter → Transformation implementation
  - [ ] Real-time coordinate transformation
  - [ ] Compatible with all transformation types

**Transformation Integration**

- [ ] Transform compatibility
  - [ ] Works with trans_affine for 2D affine transformations
  - [ ] Supports trans_perspective for perspective projection
  - [ ] Compatible with trans_bilinear for quadrilateral mapping
  - [ ] Integrates with custom transformation implementations

**Real-time Transformation**

- [ ] Streaming coordinate transformation
  - [ ] Transform coordinates as vertices are requested
  - [ ] No intermediate storage required
  - [ ] Memory-efficient transformation
  - [ ] Preserves path structure and commands

**Transform Parameter Control**

- [ ] Transformer access
  - [ ] transformer() → GetTransformer() - access underlying transformer
  - [ ] const transformer() → GetTransformer() const - read-only access
  - [ ] Direct parameter modification through transformer interface
  - [ ] Real-time transformation parameter updates

**Path Command Preservation**

- [ ] Command stream integrity
  - [ ] Preserves all path commands (move_to, line_to, etc.)
  - [ ] Maintains path structure and organization
  - [ ] Proper path ID propagation
  - [ ] Command sequence preservation

**Performance Optimization**

- [ ] Efficient transformation
  - [ ] Single-pass coordinate transformation
  - [ ] Minimal computational overhead
  - [ ] Memory-friendly streaming operation
  - [ ] Optimized for interactive transformation

**Advanced Transformation Features**

- [ ] Complex transformation scenarios
  - [ ] Nested transformation support
  - [ ] Transform composition through chaining
  - [ ] Interactive transformation updates
  - [ ] Animation-friendly transformation

**Integration with Rendering Pipeline**

- [ ] Renderer compatibility
  - [ ] Compatible with all AGG renderers
  - [ ] Works with rasterizers and scanline renderers
  - [ ] Integrates with anti-aliasing systems
  - [ ] Supports all pixel formats

**Dependencies**

- agg_basics.h → internal/basics package
- Transformation interface definitions (trans_affine, trans_perspective, etc.)
- Coordinate transformation utilities

---

### Vertex Generators

Vertex generators transform input path vertices into specialized output sequences for rendering operations like stroking, dashing, and contouring. They sit between path definition and rasterization in the AGG pipeline.

**Overview**

Vertex generators implement the "Vertex Generator Interface" with `remove_all()`, `add_vertex()` methods and "Vertex Source Interface" with `rewind()`, `vertex()` methods. They process input vertices to create modified vertex streams for specific rendering effects.

**Key Components**

- [x] **agg_vcgen_bspline.h** - B-spline curve generator (`internal/vcgen/bspline.go`)
  - Converts control points to smooth B-spline curves
  - Configurable interpolation step for curve resolution
  - Status: ✅ **COMPLETED** - Basic implementation with tests

- [x] **agg_vcgen_smooth_poly1.h** - Polygon smoothing generator (`internal/vcgen/smooth_poly1.go`)
  - Smooths polygon vertices using curve interpolation
  - Creates rounded corners and smooth transitions
  - Status: ✅ **COMPLETED** - Basic implementation with tests

- [x] **agg_vcgen_vertex_sequence.h** - Vertex sequence manager (`internal/vcgen/vertex_sequence.go`)
  - Manages vertex sequences with distance calculations
  - Supports path shortening and vertex filtering
  - Status: ✅ **COMPLETED** - Basic implementation with tests

- [ ] **agg_vcgen_stroke.h** - Stroke generator (`internal/vcgen/stroke.go`)
  - Converts single-line paths into stroked outlines
  - Implements line caps (butt, round, square) and joins (miter, round, bevel)
  - Handles stroke width, miter limits, and inner joins
  - Status: ❌ **PENDING** - Critical for basic drawing operations

- [ ] **agg_vcgen_dash.h** - Dash pattern generator (`internal/vcgen/dash.go`)
  - Creates dashed line patterns from solid paths
  - Configurable dash lengths and gap patterns
  - Maintains dash phase across path segments
  - Status: ❌ **PENDING** - Important for styled line drawing

- [ ] **agg_vcgen_contour.h** - Contour generator (`internal/vcgen/contour.go`)
  - Generates parallel contours (outlines) from paths
  - Creates offset curves at specified distances
  - Similar to stroke but for filled shapes
  - Status: ❌ **PENDING** - Needed for advanced text and shape effects

- [ ] **agg_vcgen_markers_term.h** - Terminal markers generator (`internal/vcgen/markers_term.go`)
  - Generates terminal markers (arrowheads, tails) for paths
  - Places markers at path start/end points
  - Calculates marker orientation based on path direction
  - Status: ❌ **PENDING** - Needed for arrows and decorative elements

**Implementation Requirements**

1. **Core Interface Implementation**
   - Vertex Generator Interface: `RemoveAll()`, `AddVertex(x, y float64, cmd uint)`
   - Vertex Source Interface: `Rewind(pathID uint)`, `Vertex() (x, y float64, cmd uint)`
   - State management with status enums (initial, ready, processing, stop)

2. **Stroke Generator (Priority: HIGH)**
   - Line cap types: butt, round, square
   - Line join types: miter, round, bevel, miter_revert
   - Inner join handling for acute angles
   - Stroke width and miter limit parameters
   - Integration with math_stroke calculations

3. **Dash Generator (Priority: HIGH)**
   - Dash pattern array support (up to 32 patterns)
   - Dash start offset and scaling
   - State tracking across path segments
   - Pattern cycling and phase management

4. **Contour Generator (Priority: MEDIUM)**
   - Positive/negative width for inward/outward contours
   - Integration with stroke math for offset calculations
   - Closed polygon handling
   - Self-intersection resolution

5. **Markers Generator (Priority: MEDIUM)**
   - Coordinate storage for marker positions
   - Orientation calculation from path tangents
   - Multiple marker support along paths
   - Marker transformation interface

**Missing Dependencies**

- [ ] **agg_math_stroke.h** → `internal/basics/math_stroke.go` (526 lines)
  - Line cap/join calculations
  - Stroke mathematics for width application
  - Miter limit and angle calculations
  - **Required for**: vcgen_stroke, vcgen_contour

- [ ] **agg_shorten_path.h** → `internal/basics/shorten_path.go` (66 lines)
  - Path shortening utilities
  - Used by vcgen_vertex_sequence for path end adjustments
  - **Required for**: Enhanced vertex_sequence functionality

- [ ] **vertex_dist types** → Enhance `internal/basics/types.go`
  - `VertexDist` struct (x, y, dist fields)
  - `VertexDistCmd` struct (extends VertexDist with cmd field)
  - Distance calculation methods
  - **Required for**: All vertex generators using distance-based filtering

**Testing Requirements**

1. **Unit Tests**
   - Stroke generation with different cap/join combinations
   - Dash pattern generation and phase management
   - Contour generation with positive/negative widths
   - Marker placement and orientation accuracy

2. **Integration Tests**
   - Vertex generator chaining (dash → stroke)
   - Integration with rasterizer pipeline
   - Performance testing with complex paths
   - Memory usage validation

3. **Visual Tests**
   - Stroke rendering with various parameters
   - Dash pattern visual verification
   - Contour accuracy against reference
   - Marker placement and scaling

**Integration Points**

- **Input**: Path vertices from `types.Path` or transformation pipeline
- **Output**: Modified vertex streams to rasterizer or scanline renderer
- **Conversion Chain**: Path → [Transform] → [VCGen] → [Conv] → Rasterizer
- **Renderer Compatibility**: Must work with all AGG renderers and pixel formats

**Performance Considerations**

- Vertex generators process large vertex streams
- Memory pooling for intermediate vertex storage
- Efficient state machine implementation
- SIMD optimization opportunities for math operations

**Dependencies**

- `internal/basics` → Basic types, constants, and math utilities
- `internal/array` → Vertex sequence and storage containers
- Future: `internal/basics/math_stroke.go` → Stroke mathematics
- Future: `internal/basics/shorten_path.go` → Path manipulation utilities

---

### Vertex Processors

- [ ] agg_vpgen_clip_polygon.h - Polygon clipping vertex processor
- [ ] agg_vpgen_clip_polyline.h - Polyline clipping vertex processor
- [ ] agg_vpgen_segmentator.h - Segmentator vertex processor

---

### Spans and Gradients

- [x] agg_span_allocator.h - Span allocator
- [ ] agg_span_converter.h - Span converter
- [x] agg_span_solid.h - Solid color span
- [ ] agg_span_gradient.h - Gradient span
- [ ] agg_span_gradient_alpha.h - Alpha gradient span
- [ ] agg_span_gradient_contour.h - Contour gradient span
- [ ] agg_span_gradient_image.h - Image gradient span
- [ ] agg_span_gouraud.h - Gouraud shading span
- [ ] agg_span_gouraud_gray.h - Grayscale Gouraud span
- [ ] agg_span_gouraud_rgba.h - RGBA Gouraud span

---

### Image Processing

- [ ] agg_image_accessors.h - Image accessors
- [ ] agg_image_filters.h - Image filters
- [ ] agg_span_image_filter.h - Image filter span
- [ ] agg_span_image_filter_gray.h - Grayscale image filter span
- [ ] agg_span_image_filter_rgb.h - RGB image filter span
- [ ] agg_span_image_filter_rgba.h - RGBA image filter span

---

### Pattern Processing

- [ ] agg_pattern_filters_rgba.h - RGBA pattern filters
- [ ] agg_span_pattern_gray.h - Grayscale pattern span
- [ ] agg_span_pattern_rgb.h - RGB pattern span
- [ ] agg_span_pattern_rgba.h - RGBA pattern span

---

### Interpolators

- [ ] agg_span_interpolator_adaptor.h - Interpolator adaptor
- [ ] agg_span_interpolator_linear.h - Linear interpolator
- [ ] agg_span_interpolator_persp.h - Perspective interpolator
- [ ] agg_span_interpolator_trans.h - Transform interpolator
- [ ] agg_span_subdiv_adaptor.h - Subdivision adaptor

---

### Utility and Math

- [ ] agg_alpha_mask_u8.h - 8-bit alpha mask
- [ ] agg_bitset_iterator.h - Bitset iterator
- [ ] agg_blur.h - Blur effects
- [ ] agg_bounding_rect.h - Bounding rectangle calculation
- [ ] agg_clip_liang_barsky.h - Liang-Barsky clipping algorithm
- [x] agg_dda_line.h - DDA line algorithm (line_bresenham_interpolator and dda2_line_interpolator implemented)
- [ ] agg_gamma_functions.h - Gamma correction functions
- [ ] agg_gamma_lut.h - Gamma lookup table
- [ ] agg_gradient_lut.h - Gradient lookup table
- [ ] agg_line_aa_basics.h - Anti-aliased line basics
- [ ] agg_math_stroke.h - Stroke mathematics
- [ ] agg_shorten_path.h - Path shortening
- [ ] agg_simul_eq.h - Simultaneous equations solver
- [ ] agg_vertex_sequence.h - Vertex sequence

---

### Text and Fonts

- [ ] agg_embedded_raster_fonts.h - Embedded raster fonts
- [ ] agg_font_cache_manager.h - Font cache manager
- [ ] agg_font_cache_manager2.h - Font cache manager v2
- [ ] agg_glyph_raster_bin.h - Binary glyph rasterizer
- [ ] agg_gsv_text.h - GSV text rendering

---

### Controls (ctrl/)

- [ ] agg_ctrl.h - Base control class
- [ ] agg_bezier_ctrl.h - Bezier curve control
- [ ] agg_cbox_ctrl.h - Checkbox control
- [ ] agg_gamma_ctrl.h - Gamma control
- [ ] agg_gamma_spline.h - Gamma spline
- [ ] agg_polygon_ctrl.h - Polygon control
- [ ] agg_rbox_ctrl.h - Radio button control
- [ ] agg_scale_ctrl.h - Scale control
- [ ] agg_slider_ctrl.h - Slider control
- [ ] agg_spline_ctrl.h - Spline control

---

### Platform Support (platform/)

- [ ] agg_platform_support.h - Platform support interface

---

### Utilities (util/)

- [ ] agg_color_conv.h - Color conversion utilities
- [ ] agg_color_conv_rgb16.h - 16-bit RGB color conversion
- [ ] agg_color_conv_rgb8.h - 8-bit RGB color conversion

## Core Implementation Files (src/)

### Basic Implementations

- [x] agg_arc.cpp - Arc generation implementation
- [x] agg_arrowhead.cpp - Arrowhead implementation
- [x] agg_ellipse.cpp - Ellipse generation implementation
- [x] agg_bezier_arc.cpp - Bezier arc implementation
- [x] agg_bspline.cpp - B-spline implementation
- [ ] agg_color_rgba.cpp - RGBA color implementation
- [ ] agg_curves.cpp - Curve approximation implementation
- [ ] agg_embedded_raster_fonts.cpp - Embedded fonts implementation
- [ ] agg_gsv_text.cpp - GSV text implementation
- [ ] agg_image_filters.cpp - Image filters implementation
- [ ] agg_line_aa_basics.cpp - Anti-aliased line basics
- [ ] agg_line_profile_aa.cpp - Anti-aliased line profile
- [x] agg_rounded_rect.cpp - Rounded rectangle implementation
- [ ] agg_sqrt_tables.cpp - Square root tables
- [x] agg_trans_affine.cpp - Affine transformation implementation
- [x] agg_trans_double_path.cpp - Double path transformation
- [x] agg_trans_single_path.cpp - Single path transformation
- [x] agg_trans_warp_magnifier.cpp - Warp magnifier implementation

### Vertex Generators

- [x] agg_vcgen_bspline.cpp - B-spline vertex generator
- [ ] agg_vcgen_contour.cpp - Contour vertex generator
- [ ] agg_vcgen_dash.cpp - Dash vertex generator
- [ ] agg_vcgen_markers_term.cpp - Terminal markers implementation
- [x] agg_vcgen_smooth_poly1.cpp - Polygon smoothing implementation
- [ ] agg_vcgen_stroke.cpp - Stroke vertex generator

### Vertex Processors

- [ ] agg_vpgen_clip_polygon.cpp - Polygon clipping implementation
- [ ] agg_vpgen_clip_polyline.cpp - Polyline clipping implementation
- [ ] agg_vpgen_segmentator.cpp - Segmentator implementation

### Controls Implementation (src/ctrl/)

- [ ] agg_bezier_ctrl.cpp - Bezier control implementation
- [ ] agg_cbox_ctrl.cpp - Checkbox control implementation
- [ ] agg_gamma_ctrl.cpp - Gamma control implementation
- [ ] agg_gamma_spline.cpp - Gamma spline implementation
- [ ] agg_polygon_ctrl.cpp - Polygon control implementation
- [ ] agg_rbox_ctrl.cpp - Radio button implementation
- [ ] agg_scale_ctrl.cpp - Scale control implementation
- [ ] agg_slider_ctrl.cpp - Slider control implementation
- [ ] agg_spline_ctrl.cpp - Spline control implementation

## AGG2D High-Level Interface

- [ ] agg2d.h - AGG2D header
- [ ] agg2d.cpp - AGG2D implementation

## Font Support

### FreeType Integration

- [ ] agg_font_freetype.h - FreeType font support header
- [ ] agg_font_freetype.cpp - FreeType font implementation
- [ ] agg_font_freetype2.h - FreeType v2 support header
- [ ] agg_font_freetype2.cpp - FreeType v2 implementation

### Win32 TrueType Support

- [ ] agg_font_win32_tt.h - Win32 TrueType header
- [ ] agg_font_win32_tt.cpp - Win32 TrueType implementation

## General Polygon Clipper (GPC)

- [ ] gpc.h - GPC header
- [ ] gpc.c - GPC implementation

## Platform Support Implementations

### Cross-Platform

- [ ] agg_platform_support.cpp (generic interface)

### Platform-Specific (Optional - for examples)

- [ ] src/platform/X11/agg_platform_support.cpp - X11 support
- [ ] src/platform/win32/agg_platform_support.cpp - Win32 support
- [ ] src/platform/win32/agg_win32_bmp.cpp - Win32 bitmap support
- [ ] src/platform/mac/agg_platform_support.cpp - macOS support
- [ ] src/platform/mac/agg_mac_pmap.cpp - macOS pixmap support
- [ ] src/platform/sdl/agg_platform_support.cpp - SDL support
- [ ] src/platform/sdl2/agg_platform_support.cpp - SDL2 support

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
