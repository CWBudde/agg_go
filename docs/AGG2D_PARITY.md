# Agg2D Public API Parity

This document tracks how the public Go `Agg2D` wrapper maps to the original C++
interface in `../agg-2.6/agg-src/agg2d/agg2d.h`.

The intent is close interface parity, not a wholesale Go redesign, so the
original AGG2D documentation and demo code remain useful references.

## Naming Rules

- C++ getter overloads become explicit `Get...` methods in Go.
  Examples: `clipBox() const` -> `GetClipBox()`, `blendMode() const` -> `GetBlendMode()`.
- Whole-image overloads that would collide with rectangle/parallelogram
  overloads use a `...Simple` suffix.
  Examples: `transformImage(img, dstX1, dstY1, dstX2, dstY2)` -> `TransformImageSimple(...)`.
- Three-color radial-gradient overloads use `...MultiStop`.
  Examples: `fillRadialGradient(x, y, r, c1, c2, c3)` -> `FillRadialGradientMultiStop(...)`.
- Position-only radial-gradient overloads use `...Pos`.
  Examples: `fillRadialGradient(x, y, r)` -> `FillRadialGradientPos(...)`.
- Rounded-rectangle overloads use descriptive names because Go cannot overload
  by argument count.
  Examples: `roundedRect(..., rx, ry)` -> `RoundedRectXY(...)`,
  `roundedRect(..., rxBottom, ryBottom, rxTop, ryTop)` -> `RoundedRectVariableRadii(...)`.

## Direct Mappings

These method families keep the upstream name directly:

- setup and state: `Attach`, `ClipBox`, `ClearAll`, `ClearClipBox`
- coordinate conversion: `WorldToScreen`, `ScreenToWorld`, `AlignPoint`, `InBox`
- drawing state: `BlendMode`, `ImageBlendMode`, `ImageBlendColor`, `MasterAlpha`, `AntiAliasGamma`
- colors and gradients: `FillColor`, `LineColor`, `FillLinearGradient`, `LineLinearGradient`, `FillRadialGradient`, `LineRadialGradient`
- transforms: `ResetTransformations`, `Affine`, `Rotate`, `Scale`, `Skew`, `Translate`, `Parallelogram`, `Viewport`
- shapes and paths: `Line`, `Triangle`, `Rectangle`, `Ellipse`, `Arc`, `Star`, `Polygon`, `Polyline`, `MoveTo`, `LineTo`, `ArcTo`, `ClosePolygon`, `DrawPath`
- text: `FlipText`, `Font`, `FontHeight`, `TextAlignment`, `TextHints`, `TextWidth`, `Text`
- images: `ImageFilter`, `ImageResample`, `TransformImage`, `TransformImagePath`, `BlendImage`, `CopyImage`

## Deliberate Go Differences

These differences are intentional and currently unavoidable:

- Getter overloads are named explicitly with `Get...`.
- Overload families that differ only by parameter list use descriptive suffixes
  such as `Simple`, `MultiStop`, and `Pos`.
- Public image-transform methods accept Go slices for parallelograms instead of
  raw `double*`.
- The public package exposes Go `Color` values rather than the internal array
  representation used inside `internal/agg2d`.
- A few Go-only helpers remain available for convenience, such as `FontGSV`,
  `PushTransform`, `PopTransform`, `DrawCircle`, `FillCircle`, and
  `SaveImagePPM`.

## Getter Mapping Reference

- `clipBox() const` -> `GetClipBox()`
- `fillColor() const` -> `GetFillColor()`
- `lineColor() const` -> `GetLineColor()`
- `blendMode() const` -> `GetBlendMode()`
- `imageBlendMode() const` -> `GetImageBlendMode()`
- `imageBlendColor() const` -> `GetImageBlendColor()`
- `masterAlpha() const` -> `GetMasterAlpha()`
- `antiAliasGamma() const` -> `GetAntiAliasGamma()`
- `lineCap() const` -> `GetLineCap()`
- `lineJoin() const` -> `GetLineJoin()`
- `fillEvenOdd() const` -> `GetFillEvenOdd()`
- `imageFilter() const` -> `GetImageFilter()`
- `imageResample() const` -> `GetImageResample()`
- `textHints() const` -> `GetTextHints()`
- `transformations() const` -> `GetTransformations()`

## Status

As of Phase 10.0, the remaining interface differences are mostly overload-name
translation rather than missing method families.
