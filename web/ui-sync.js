// --- Control visibility and demo descriptions ---

export function syncControlVisibility(demoType) {
  document.getElementById("aaControls").style.display =
    demoType === "aa" ? "flex" : "none";
  document.getElementById("dashControls").style.display =
    demoType === "conv_dash_marker" ? "flex" : "none";
  document.getElementById("gouraudControls").style.display =
    demoType === "gouraud" ? "flex" : "none";
  document.getElementById("gouraudMeshControls").style.display =
    demoType === "gouraud_mesh" ? "flex" : "none";
  document.getElementById("imageFilterControls").style.display =
    demoType === "imagefilters" ? "flex" : "none";
  document.getElementById("imageFltrGraphControls").style.display =
    demoType === "image_fltr_graph" ? "flex" : "none";
  document.getElementById("sboolControls").style.display =
    demoType === "sbool" ? "flex" : "none";
  document.getElementById("convstrokeControls").style.display =
    demoType === "convstroke" ? "flex" : "none";
  document.getElementById("convcontourControls").style.display =
    demoType === "convcontour" ? "flex" : "none";
  document.getElementById("gammaControls").style.display =
    demoType === "gamma" ? "flex" : "none";
  document.getElementById("lionControls").style.display =
    demoType === "lion" ? "flex" : "none";
  document.getElementById("lionoutlineControls").style.display =
    demoType === "lionoutline" ? "flex" : "none";
  document.getElementById("lionLensControls").style.display =
    demoType === "lion_lens" ? "flex" : "none";
  document.getElementById("roundedrectControls").style.display =
    demoType === "roundedrect" ? "flex" : "none";
  document.getElementById("componentControls").style.display =
    demoType === "component" ? "flex" : "none";
  document.getElementById("perspectiveControls").style.display =
    demoType === "perspective" ? "flex" : "none";
  document.getElementById("transCurveControls").style.display =
    demoType === "trans_curve" ? "flex" : "none";
  document.getElementById("transCurve2Controls").style.display =
    demoType === "trans_curve2" ? "flex" : "none";
  document.getElementById("blurControls").style.display =
    demoType === "blur" ? "flex" : "none";
  document.getElementById("circlesControls").style.display =
    demoType === "circles" ? "flex" : "none";
  document.getElementById("compositingControls").style.display =
    demoType === "compositing" ? "flex" : "none";
  document.getElementById("multiClipControls").style.display =
    demoType === "multi_clip" ? "flex" : "none";
  document.getElementById("distortionsControls").style.display =
    demoType === "distortions" ? "flex" : "none";
  document.getElementById("alphaMask2Controls").style.display =
    demoType === "alpha_mask2" ? "flex" : "none";
  document.getElementById("image1Controls").style.display =
    demoType === "image1" ? "flex" : "none";
  document.getElementById("imageTransformsControls").style.display =
    demoType === "image_transforms" ? "flex" : "none";
  document.getElementById("patternFillControls").style.display =
    demoType === "pattern_fill" ? "flex" : "none";
  document.getElementById("patternPerspectiveControls").style.display =
    demoType === "pattern_perspective" ? "flex" : "none";
  document.getElementById("patternResampleControls").style.display =
    demoType === "pattern_resample" ? "flex" : "none";
  document.getElementById("gammaTunerControls").style.display =
    demoType === "gamma_tuner" ? "flex" : "none";
  document.getElementById("bezierDivControls").style.display =
    demoType === "bezier_div" ? "flex" : "none";
  document.getElementById("bezierDivControls2").style.display =
    demoType === "bezier_div" ? "flex" : "none";
  document.getElementById("rasterizersControls").style.display =
    demoType === "rasterizers" ? "flex" : "none";
  document.getElementById("bsplineControls").style.display =
    demoType === "bspline" ? "flex" : "none";
  document.getElementById("gpcTestControls").style.display =
    demoType === "gpc_test" ? "flex" : "none";
  document.getElementById("gradientsContourControls").style.display =
    demoType === "gradients_contour" ? "flex" : "none";
  document.getElementById("flashRasterizer2Controls").style.display =
    demoType === "flash_rasterizer2" ? "flex" : "none";
}

export const demoDescriptions = {
  agg2d:
    "Port of the original agg2d_demo.cpp. Showcases the high-level Agg2D API: viewport mapping, aqua-style gradient buttons with rounded rectangles, filled ellipses, arc-based path construction, blend modes (Add, Overlay), and radial gradient fills.",
  lion: "Port of AGG's lion.cpp. The classic AGG signature demo rendered as filled paths. Left-drag to rotate and scale; right-drag to skew. Adjust the Alpha slider to control global opacity.",
  gradients:
    "Linear and radial gradient fills. Demonstrates the advanced span generation and multi-stop color interpolation.",
  gradient_focal:
    "Port of AGG's gradient_focal demo. Renders a reflected radial-focus gradient (focal offset inside the circle) using a gamma-aware 4-stop LUT. Web parameters are URL-driven: gfg (gamma), gfx (focus-x), gfy (focus-y).",
  aa: "Anti-aliasing showcase. Lines and circles drawn at sub-pixel offsets to demonstrate the precision and smoothness of AGG's rasterizer.",
  blend:
    "Blend-mode gallery. Shows multiple blend modes side-by-side on overlapping RGB circles for quick visual comparison.",
  bspline:
    "B-Spline curve smoothing. Demonstrates the creation of smooth, continuous curves from a set of control points.",
  interactive_polygon:
    "Port of AGG's interactive_polygon.cpp helper. Drag a vertex to move one point, drag an edge to move the adjacent segment, or drag inside the shape to move the whole polygon.",
  conv_dash_marker:
    "Port of AGG's conv_dash_marker demo. Applies conv_smooth_poly1 to soften corners, then conv_dash to create dash patterns, and conv_marker to place arrowheads at line endpoints. Adjust smoothness, stroke width, cap style, and fill rule. Drag the three control points to reshape the paths.",
  line_thickness:
    "Port of AGG's line_thickness demo. Renders variable-thickness anti-aliased line sets and applies slight blur. Web parameters are URL-driven: ltf (thickness factor), ltb (blur radius), ltm (monochrome 0/1), lti (invert 0/1).",
  line_patterns:
    "Port of AGG's line_patterns demo core. Draws nine patterned Bezier curves using the original line pattern assets. URL parameters: lpsx (scale_x), lpst (start_x).",
  line_patterns_clip:
    "Port of AGG's line_patterns_clip demo core. Image-patterned polyline clipped to an inner rectangle. URL parameters: lpcsx (scale_x), lpcst (start_x).",
  scanline_boolean2:
    "Port of AGG's scanline_boolean2 demo core using polygon clipping-backed boolean operations. URL parameters: sb2m (mode 0..4), sb2f (fill rule 0/1), sb2s (scanline type 0..2), sb2o (operation 0..6), sb2x/sb2y (center).",
  gpc_test:
    "Port of AGG's gpc_test demo core using general polygon clipping operations. Drag with left mouse to move the scene center.",
  gradients_contour:
    "Port of AGG's gradients_contour demo. Contour-based gradients using Distance Transform — colors follow the shape of an arbitrary path. Supports star, Great Britain outline, spiral, and glyph shapes with four gradient modes (Contour, Auto-Contour, Conic/Angle, Flat) and 2–11 color stops.",
  flash_rasterizer2:
    "Port of AGG's flash_rasterizer2.cpp. Alternative Flash compound-shape rasterization: decomposes each shape into per-fill-style sub-shapes. For each style, paths with a matching left-fill are added forward and paths with a matching right-fill are added reversed (inverted polygon winding). A clipping rasterizer discards the spurious edge from the clipper origin. Select any of the 24 shape frames with the slider.",
  polymorphic_renderer:
    "Port of AGG's polymorphic_renderer.cpp. Demonstrates how Go interfaces provide natural polymorphic rendering — the same scanline rendering code works uniformly across any pixel-format backend without virtual dispatch. In C++ this required a virtual base class and an explicit factory switch; in Go a single interface value suffices. Drag the three corner handles to reshape the filled triangle.",
  gouraud:
    "Smooth color interpolation across triangles. Demonstrates AGG's capability to render gradient-shaded meshes with sub-pixel precision and adjustable dilation.",
  imagefilters:
    "Comparison of different image interpolation filters. Rotates and scales a procedurally generated image using filters like Bilinear, Bicubic, Sinc, and Lanczos to showcase quality vs. performance.",
  image_fltr_graph:
    "Port of AGG's image_fltr_graph demo. Plots raw filter shape (red), unnormalized discrete accumulation (green), and normalized LUT weights (blue). Use HTML controls/URL params: ifgr (radius), ifgm (bitmask of enabled filters).",
  sbool:
    "Boolean operations on vector shapes. Demonstrates combining multiple paths using filling rules to achieve Union and XOR-like effects with interactive polygons.",
  aatest:
    "Comprehensive anti-aliasing precision test. Renders radial lines, various ellipse sizes, and gradient-filled triangles at fractional offsets to verify the rasterizer's quality.",
  convstroke:
    "Line join and cap style showcase. Port of AGG's classic conv_stroke demo. Drag the three control points to reshape the path; use the controls to change join style (Miter/Round/Bevel), cap style (Butt/Square/Round), stroke width, and miter limit.",
  convcontour:
    "Contour tool and polygon orientation. Port of AGG's conv_contour demo. Expands or shrinks a closed path by a given width using the contour converter. The glyph is defined with quadratic bezier curves, processed through conv_curve → conv_transform → conv_contour. Adjust the width slider and orientation flags to see the effect.",
  gamma:
    "Gamma correction showcase. Port of AGG's gamma_correction demo. Renders colored ellipses over a four-quadrant background (dark, light, reddish) to demonstrate how the anti-aliasing gamma affects line quality. Click and drag on the canvas to resize the ellipses. Adjust gamma, line thickness, and background contrast with the sliders.",
  lionoutline:
    "Lion outline rendering. Port of AGG's lion_outline demo. The classic lion vector art is rendered as stroked outlines instead of filled polygons. Left-drag to rotate and scale the lion; right-drag to apply shear. Adjust the line width with the slider.",
  roundedrect:
    "Rounded rectangle demo. Port of AGG's rounded_rect demo. Drag the two corner control points to resize the rectangle. Adjust the corner radius and sub-pixel offset with sliders; toggle white-on-black rendering with the checkbox.",
  alphagrad:
    "Alpha channel gradient. Port of AGG's alpha_gradient demo. A circle is filled with a circular colour gradient (dark teal → yellow-green → dark red); its alpha channel is independently modulated by an XY-product gradient mapped over a draggable parallelogram. Drag the three teal control points to reshape the parallelogram and watch the transparency pattern change. Dragging inside the triangle moves all three together.",
  component:
    "Component (channel) rendering. Port of AGG's component_rendering demo. Three large circles are each rendered into an individual color channel using Multiply blend mode, producing classic CMY subtractive color mixing: Cyan darkens the Red channel, Magenta the Green, Yellow the Blue. Where all three overlap the result is black. The Alpha slider controls how strongly each channel is darkened.",
  rasterizers:
    "Aliased vs Anti-Aliased rasterization. Comparison between the standard AA rasterizer and aliased (threshold-based) rendering. Drag the triangle nodes to see how edges behave under different rendering modes and gamma settings.",
  flash_rasterizer:
    "Compound rasterization. Demonstrates AGG's ability to render overlapping shapes with multiple styles in a single pass using the compound rasterizer. This is highly efficient for complex vector scenes with many layers.",
  graph_test:
    "Port of AGG's graph_test demo core. Renders a deterministic random graph with curved, arrowed edges and radial node markers.",
  rasterizer_compound:
    "Port of AGG's rasterizer_compound demo. Shows layered compound AA rasterization (direct/inverse layer order) over a yellow→cyan gradient background. URL params: rcw, rca1..rca4, rcio.",
  perspective:
    "Perspective and Bilinear transformations. Apply non-linear distortions to the lion vector art by dragging the four corners of the control quadrilateral. Switch between Bilinear and Perspective modes to see the difference in projection.",
  bezier_div:
    "Bezier curve subdivision comparison. Shows two methods of rendering cubic Bezier curves: Subdivision (Green) and Incremental (Red). Drag the four control points to see how both algorithms handle various curve shapes and cusps.",
  gouraud_mesh:
    "Animated Gouraud-shaded mesh. A grid of triangles with varying colors and positions, rendered efficiently using compound rasterization and smooth Gouraud shading. Drag points to manually distort the mesh.",
  trans_curve:
    "Along-a-curve transformation. Bends complex vector shapes (the lion) along an interactive B-Spline path. Drag the six control points to reshape the path. Toggle animation to watch the lion flow along the moving curve.",
  trans_curve2:
    "Double path transformation. Bends vector shapes (the lion) between two interactive B-Spline curves. Drag the 12 control points to reshape the envelope. Toggle animation to watch the lion morph between the moving curves.",
  gamma_ctrl:
    "Interactive gamma correction control. Port of AGG's gamma_ctrl demo. Use the spline control points to adjust the gamma curve and see its effect on various primitives, text, and rotated shapes.",
  gamma_tuner:
    "RGB gamma tuning tool. Port of AGG's gamma_tuner demo. Calibrate gamma for R, G, and B channels independently using horizontal, vertical, and checkered test patterns.",
  lion_lens:
    "Dynamic lens magnification effect. Port of AGG's lion_lens demo. Applies a TransWarpMagnifier to the lion vector art. Click and drag to move the lens; use the sliders to adjust scale and radius.",
  distortions:
    "Animated image distortions. Applies Wave and Swirl effects to the selected source image (Spheres/Test Grid) using custom coordinate distortion interpolators. Click and drag to move the distortion center. URL parameter: dimg (0/1).",
  trans_polar:
    "Polar coordinate transformations. Bends the lion vector art into a circular or spiral shape using a non-linear polar transformer. Click and drag to adjust the radius and spiral intensity.",
  circles:
    "Random circles demo. A scatter plot prototype using B-Spline color interpolation. Renders thousands of small circles with colors controlled by splines. Click to regenerate the points.",
  blur: "Gaussian and Stack blur demonstration. Renders a complex path with a shadow and applies recursive or stack blur to the entire canvas. Use the controls to adjust radius and method.",
  simple_blur:
    "Simple 3x3 box blur. Renders the classic lion and then applies a simple box blur inside a draggable elliptical region. Click and drag to move the blurred area.",
  alpha_mask:
    "Alpha masking showcase. Port of AGG's alpha_mask demo. Renders the classic lion vector art through a dynamic alpha mask generated from overlapping random ellipses. Demonstrates the PixFmtAMaskAdaptor's ability to apply transparency patterns to any rendering operation.",
  alpha_mask2:
    "Alpha Mask 2 — Lion with ellipse mask. Renders the classic lion vector art through a grayscale alpha mask built from random semi-transparent ellipses. Left-drag to rotate and scale the lion; right-drag to skew. Adjust the Ellipses slider to change the mask density.",
  compositing:
    "Porter-Duff and SVG compositing operations. Port of AGG's compositing demo. Demonstrates various rules for combining source and destination shapes, such as SrcOver, Multiply, Screen, and Xor. Adjust the source and destination opacity and select different operations to see how they affect the overlapping regions.",
  multi_clip:
    "Multi-region clipping. Port of AGG's multi_clip demo. Showcases the RendererMClip which allows restricting all rendering operations to a set of multiple rectangular regions. Adjust the grid size slider to change the number of clipping boxes and watch the lion art being clipped into a grid.",
  image1:
    "Image affine transformation. Port of AGG's image1.cpp. A large ellipse is filled with a bilinear-filtered image that rotates and scales with the ellipse. Use the angle and scale sliders to transform the image.",
  image_resample:
    "Port of AGG's image_resample demo (closest current affine equivalent). Renders a clipped transformed image inside a 4-point quad; affine modes and resample blur are URL-driven: irt, irb, irx0..iry3.",
  image_perspective:
    "Port of AGG's image_perspective demo. Modes and quad are URL-driven: ipt (0..2), ipx0..ipy3.",
  pattern_perspective:
    "Port of AGG's pattern_perspective demo. Reflect-wrapped AGG pattern warped through affine/bilinear/perspective; URL params: ppt (0..2), ppx0..ppy3.",
  pattern_resample:
    "Port of AGG's pattern_resample demo. Includes perspective lerp/exact and resample paths with gamma+blur controls via URL: prt (0..5), prg, prb, prx0..pry3.",
  image_transforms:
    "Image transform examples. Port of AGG's image_transforms.cpp. A 14-pointed star polygon is filled with an affine-filtered image. Choose between 7 different examples showing how the image matrix can be set independently from the polygon matrix. Drag the image center point (red handle) to reposition it.",
  image_alpha:
    "Image alpha from brightness. Port of AGG's image_alpha.cpp. A background of random colored ellipses is overlaid by a large rotated ellipse filled with an image. The alpha of each image pixel is derived from its brightness via a piecewise linear lookup table, making bright areas more transparent.",
  pattern_fill:
    "Pattern fill (tiled star). Port of AGG's pattern_fill.cpp. A 14-pointed star polygon is filled with a small repeating star pattern rendered into an offscreen buffer. Use the sliders to rotate and scale both the polygon and the pattern tile.",
};
