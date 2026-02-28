# AGG2D Parity Ledger

This ledger tracks parity between the original AGG2D API and the Go port.

Primary references:

- `../agg-2.6/agg-src/agg2d/agg2d.h`
- `../agg-2.6/agg-src/agg2d/agg2d.cpp`

Status values:

- `unassessed`: mapping exists, behavior not yet audited in detail.
- `close`: largely aligned, minor deltas possible.
- `placeholder`: known simplified/non-fidelity implementation.
- `delta`: intentional or known mismatch from original API/behavior.
- `decl-only`: declared in C++ header, no C++ implementation found.

## Setup

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `attach(buf,width,height,stride)` | `agg2d.h:261`, `agg2d.cpp:125` | `(*agg2d.Agg2D).Attach` | close | `internal/agg2d/agg2d_test.go` | |
| `attach(Image&)` | `agg2d.h:262`, `agg2d.cpp:153` | `(*agg2d.Agg2D).Attach` (via `Image`) | unassessed | `internal/agg2d/image_test.go` | API shape differs in Go. |
| `clipBox(x1,y1,x2,y2)` | `agg2d.h:264`, `agg2d.cpp:159` | `(*agg2d.Agg2D).ClipBox` | close | `internal/agg2d/agg2d_test.go` | Renderer clip propagation audit pending. |
| `clipBox() const` | `agg2d.h:265`, `agg2d.cpp:246` | `(*agg2d.Agg2D).GetClipBox` | close | `internal/agg2d/agg2d_test.go` | |
| `clearAll(Color)` | `agg2d.h:267`, `agg2d.cpp:252` | `(*agg2d.Agg2D).ClearAll` | close | `internal/agg2d/agg2d_test.go` | |
| `clearAll(r,g,b,a)` | `agg2d.h:268`, `agg2d.cpp:258` | `(*agg2d.Agg2D).ClearAllRGBA` | close | `internal/agg2d/agg2d_test.go` | |
| `clearClipBox(Color)` | `agg2d.h:270`, `agg2d.cpp:264` | `(*agg2d.Agg2D).ClearClipBox` | close | `internal/agg2d/utilities_test.go` | |
| `clearClipBox(r,g,b,a)` | `agg2d.h:271`, `agg2d.cpp:270` | `(*agg2d.Agg2D).ClearClipBoxRGBA` | close | `internal/agg2d/utilities_test.go` | |

## Conversions

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `worldToScreen(x,y)` | `agg2d.h:275`, `agg2d.cpp:276` | `(*agg2d.Agg2D).WorldToScreen` | close | `internal/agg2d/utilities_test.go` | |
| `screenToWorld(x,y)` | `agg2d.h:276`, `agg2d.cpp:282` | `(*agg2d.Agg2D).ScreenToWorld` | close | `internal/agg2d/utilities_test.go` | |
| `worldToScreen(scalar)` | `agg2d.h:277`, `agg2d.cpp:289` | `(*agg2d.Agg2D).WorldToScreenScalar` | close | `internal/agg2d/utilities_test.go` | |
| `screenToWorld(scalar)` | `agg2d.h:278`, `agg2d.cpp:302` | `(*agg2d.Agg2D).ScreenToWorldScalar` | close | `internal/agg2d/utilities_test.go` | |
| `alignPoint(x,y)` | `agg2d.h:279`, `agg2d.cpp:315` | `(*agg2d.Agg2D).AlignPoint` | close | `internal/agg2d/utilities_test.go` | |
| `inBox(worldX,worldY)` | `agg2d.h:280`, `agg2d.cpp:325` | `(*agg2d.Agg2D).InBox` | close | `internal/agg2d/utilities_test.go` | |

## General Attributes

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `blendMode(m)` | `agg2d.h:284`, `agg2d.cpp:176` | `SetBlendMode` | close | `internal/agg2d/blend_modes_test.go` | |
| `blendMode() const` | `agg2d.h:285`, `agg2d.cpp:184` | `GetBlendMode` | close | `internal/agg2d/blend_modes_test.go` | |
| `imageBlendMode(m)` | `agg2d.h:287`, `agg2d.cpp:190` | `SetImageBlendMode` | close | `internal/agg2d/blend_modes_test.go` | |
| `imageBlendMode() const` | `agg2d.h:288`, `agg2d.cpp:196` | `GetImageBlendMode` | close | `internal/agg2d/blend_modes_test.go` | |
| `imageBlendColor(Color)` | `agg2d.h:290`, `agg2d.cpp:202` | `SetImageBlendColor` | close | `internal/agg2d/blend_modes_test.go` | |
| `imageBlendColor(r,g,b,a)` | `agg2d.h:291`, `agg2d.cpp:208` | `SetImageBlendColorRGBA` | close | `internal/agg2d/blend_modes_test.go` | |
| `imageBlendColor() const` | `agg2d.h:292`, `agg2d.cpp:214` | `GetImageBlendColor` | close | `internal/agg2d/blend_modes_test.go` | |
| `masterAlpha(a)` | `agg2d.h:294`, `agg2d.cpp:220` | `SetMasterAlpha` | close | `internal/agg2d/rendering_test.go` | |
| `masterAlpha() const` | `agg2d.h:295`, `agg2d.cpp:227` | `GetMasterAlpha` | close | `internal/agg2d/rendering_test.go` | |
| `antiAliasGamma(g)` | `agg2d.h:297`, `agg2d.cpp:233` | `SetAntiAliasGamma` | close | `internal/agg2d/rendering_test.go` | |
| `antiAliasGamma() const` | `agg2d.h:298`, `agg2d.cpp:240` | `GetAntiAliasGamma` | close | `internal/agg2d/rendering_test.go` | |
| `fillColor(Color)` | `agg2d.h:300`, `agg2d.cpp:424` | `FillColor` | close | `internal/agg2d/agg2d_test.go` | |
| `fillColor(r,g,b,a)` | `agg2d.h:301`, `agg2d.cpp:431` | `FillColorRGBA` | close | `internal/agg2d/agg2d_test.go` | |
| `noFill()` | `agg2d.h:302`, `agg2d.cpp:437` | `NoFill` | close | `internal/agg2d/utilities_test.go` | |
| `lineColor(Color)` | `agg2d.h:304`, `agg2d.cpp:443` | `LineColor` | close | `internal/agg2d/agg2d_test.go` | |
| `lineColor(r,g,b,a)` | `agg2d.h:305`, `agg2d.cpp:450` | `LineColorRGBA` | close | `internal/agg2d/agg2d_test.go` | |
| `noLine()` | `agg2d.h:306`, `agg2d.cpp:456` | `NoLine` | close | `internal/agg2d/utilities_test.go` | |
| `fillColor() const` | `agg2d.h:308`, `agg2d.cpp:462` | `GetFillColor` | close | `internal/agg2d/agg2d_test.go` | |
| `lineColor() const` | `agg2d.h:309`, `agg2d.cpp:468` | `GetLineColor` | close | `internal/agg2d/agg2d_test.go` | |
| `fillLinearGradient(...)` | `agg2d.h:311`, `agg2d.cpp:474` | `FillLinearGradient` | close | `internal/agg2d/gradient_test.go` | Transform-order audit pending. |
| `lineLinearGradient(...)` | `agg2d.h:312`, `agg2d.cpp:507` | `LineLinearGradient` | close | `internal/agg2d/gradient_test.go` | Transform-order audit pending. |
| `fillRadialGradient(c1,c2,profile)` | `agg2d.h:314`, `agg2d.cpp:540` | `FillRadialGradient` | placeholder | `internal/agg2d/gradient_test.go` | Uses non-fidelity helper path currently. |
| `lineRadialGradient(c1,c2,profile)` | `agg2d.h:315`, `agg2d.cpp:571` | `LineRadialGradient` | placeholder | `internal/agg2d/gradient_test.go` | Uses non-fidelity helper path currently. |
| `fillRadialGradient(c1,c2,c3)` | `agg2d.h:317`, `agg2d.cpp:602` | `FillRadialGradientMultiStop` | placeholder | `internal/agg2d/gradient_test.go` | |
| `lineRadialGradient(c1,c2,c3)` | `agg2d.h:318`, `agg2d.cpp:625` | `LineRadialGradientMultiStop` | placeholder | `internal/agg2d/gradient_test.go` | |
| `fillRadialGradient(x,y,r)` | `agg2d.h:320`, `agg2d.cpp:647` | `FillRadialGradientPos` | placeholder | `internal/agg2d/gradient_test.go` | |
| `lineRadialGradient(x,y,r)` | `agg2d.h:321`, `agg2d.cpp:659` | `LineRadialGradientPos` | placeholder | `internal/agg2d/gradient_test.go` | |
| `lineWidth(w)` | `agg2d.h:323`, `agg2d.cpp:671` | `LineWidth` | close | `internal/agg2d/rendering_test.go` | |
| `lineWidth() const` | `agg2d.h:324`, `agg2d.cpp:679` | `GetLineWidth` | close | `internal/agg2d/rendering_test.go` | C++ signature typo preserved in source. |
| `lineCap(cap)` | `agg2d.h:326`, `agg2d.cpp:701` | `LineCap` | close | `internal/agg2d/rendering_test.go` | |
| `lineCap() const` | `agg2d.h:327`, `agg2d.cpp:709` | `GetLineCap` | close | `internal/agg2d/rendering_test.go` | |
| `lineJoin(join)` | `agg2d.h:329`, `agg2d.cpp:716` | `LineJoin` | close | `internal/agg2d/rendering_test.go` | |
| `lineJoin() const` | `agg2d.h:330`, `agg2d.cpp:724` | `GetLineJoin` | close | `internal/agg2d/rendering_test.go` | |
| `fillEvenOdd(flag)` | `agg2d.h:332`, `agg2d.cpp:686` | `FillEvenOdd` | close | `internal/agg2d/fill_rules.go` | |
| `fillEvenOdd() const` | `agg2d.h:333`, `agg2d.cpp:694` | `GetFillEvenOdd` | close | `internal/agg2d/fill_rules.go` | |

## Transformations

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `transformations() const` | `agg2d.h:337`, `agg2d.cpp:333` | `GetTransformations` | close | `internal/agg2d/transform_test.go` | |
| `transformations(const Transformations&)` | `agg2d.h:338`, `agg2d.cpp:342` | `SetTransformations` | close | `internal/agg2d/transform_test.go` | |
| `resetTransformations()` | `agg2d.h:339`, `agg2d.cpp:351` | `ResetTransformations` | close | `internal/agg2d/transform_test.go` | |
| `affine(const Affine&)` | `agg2d.h:340`, `agg2d.cpp:364` | `Affine` | close | `internal/agg2d/transform_test.go` | |
| `affine(const Transformations&)` | `agg2d.h:341`, `agg2d.cpp:372` | `AffineFromMatrix` | close | `internal/agg2d/transform_test.go` | |
| `rotate(angle)` | `agg2d.h:342`, `agg2d.cpp:358` | `Rotate` | close | `internal/agg2d/transform_test.go` | |
| `scale(sx,sy)` | `agg2d.h:343`, `agg2d.cpp:379` | `Scale` | close | `internal/agg2d/transform_test.go` | |
| `skew(sx,sy)` | `agg2d.h:344`, `agg2d.cpp:359` | `Skew` | close | `internal/agg2d/transform_test.go` | |
| `translate(x,y)` | `agg2d.h:345`, `agg2d.cpp:360` | `Translate` | close | `internal/agg2d/transform_test.go` | |
| `parallelogram(...)` | `agg2d.h:346`, `agg2d.cpp:388` | `Parallelogram` | close | `internal/agg2d/transform_test.go` | Go signature shape differs. |
| `viewport(...)` | `agg2d.h:347`, `agg2d.cpp:397` | `Viewport` | close | `internal/agg2d/transform_test.go` | |

## Basic Shapes

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `line(...)` | `agg2d.h:353`, `agg2d.cpp:739` | `Line` | close | `internal/agg2d/agg2d_test.go` | |
| `triangle(...)` | `agg2d.h:354`, `agg2d.cpp:748` | `Triangle` | close | `internal/agg2d/agg2d_test.go` | |
| `rectangle(...)` | `agg2d.h:355`, `agg2d.cpp:760` | `Rectangle` | close | `internal/agg2d/agg2d_test.go` | |
| `roundedRect(...,r)` | `agg2d.h:356`, `agg2d.cpp:773` | `RoundedRect` | close | `internal/agg2d/agg2d_test.go` | |
| `roundedRect(...,rx,ry)` | `agg2d.h:357`, `agg2d.cpp:788` | `RoundedRectXY` | close | `internal/agg2d/agg2d_test.go` | |
| `roundedRect(...,rxB,ryB,rxT,ryT)` | `agg2d.h:358`, `agg2d.cpp:803` | `RoundedRectVariableRadii` | close | `internal/agg2d/agg2d_test.go` | |
| `ellipse(...)` | `agg2d.h:361`, `agg2d.cpp:820` | `Ellipse` | close | `internal/agg2d/agg2d_test.go` | |
| `arc(...)` | `agg2d.h:362`, `agg2d.cpp:832` | `Arc` | close | `internal/agg2d/agg2d_test.go` | |
| `star(...)` | `agg2d.h:363`, `agg2d.cpp:843` | `Star` | close | `internal/agg2d/agg2d_test.go` | |
| `curve(x1..x3)` | `agg2d.h:364`, `agg2d.cpp:865` | `Curve` | close | `internal/agg2d/agg2d_test.go` | |
| `curve(x1..x4)` | `agg2d.h:365`, `agg2d.cpp:875` | `Curve4` | close | `internal/agg2d/agg2d_test.go` | |
| `polygon(...)` | `agg2d.h:366`, `agg2d.cpp:885` | `Polygon` | close | `internal/agg2d/agg2d_test.go` | |
| `polyline(...)` | `agg2d.h:367`, `agg2d.cpp:896` | `Polyline` | close | `internal/agg2d/agg2d_test.go` | |

## Text

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `flipText(flip)` | `agg2d.h:372`, `agg2d.cpp:906` | `FlipText` | close | `internal/agg2d/text_test.go` | |
| `font(...)` | `agg2d.h:373`, `agg2d.cpp:912` | `Font` | close | `internal/agg2d/text_test.go` | |
| `fontHeight() const` | `agg2d.h:378`, `agg2d.cpp:947` | `FontHeight` | close | `internal/agg2d/text_test.go` | |
| `textAlignment(ax,ay)` | `agg2d.h:379`, `agg2d.cpp:953` | `TextAlignment` | close | `internal/agg2d/text_test.go` | |
| `textHints() const` | `agg2d.h:380`, `agg2d.cpp:981` | `GetTextHints` | close | `internal/agg2d/text_test.go` | |
| `textHints(hints)` | `agg2d.h:381`, `agg2d.cpp:987` | `TextHints` | close | `internal/agg2d/text_test.go` | |
| `textWidth(str)` | `agg2d.h:382`, `agg2d.cpp:960` | `TextWidth` | placeholder | `internal/agg2d/text_test.go` | Kerning/metrics parity still incomplete. |
| `text(x,y,str,...)` | `agg2d.h:383`, `agg2d.cpp:995` | `Text` | placeholder | `internal/agg2d/text_test.go` | Raster glyph path still simplified. |

## Path Commands

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `resetPath()` | `agg2d.h:387`, `agg2d.cpp:1082` | `ResetPath` | close | `internal/agg2d/agg2d_test.go` | |
| `moveTo(x,y)` | `agg2d.h:389`, `agg2d.cpp:1085` | `MoveTo` | close | `internal/agg2d/agg2d_test.go` | |
| `moveRel(dx,dy)` | `agg2d.h:390`, `agg2d.cpp:1091` | `MoveRel` | close | `internal/agg2d/agg2d_test.go` | |
| `lineTo(x,y)` | `agg2d.h:392`, `agg2d.cpp:1098` | `LineTo` | close | `internal/agg2d/agg2d_test.go` | |
| `lineRel(dx,dy)` | `agg2d.h:393`, `agg2d.cpp:1105` | `LineRel` | close | `internal/agg2d/agg2d_test.go` | |
| `horLineTo(x)` | `agg2d.h:395`, `agg2d.cpp:1112` | `HorLineTo` | close | `internal/agg2d/agg2d_test.go` | |
| `horLineRel(dx)` | `agg2d.h:396`, `agg2d.cpp:1119` | `HorLineRel` | close | `internal/agg2d/agg2d_test.go` | |
| `verLineTo(y)` | `agg2d.h:398`, `agg2d.cpp:1126` | `VerLineTo` | close | `internal/agg2d/agg2d_test.go` | |
| `verLineRel(dy)` | `agg2d.h:399`, `agg2d.cpp:1133` | `VerLineRel` | close | `internal/agg2d/agg2d_test.go` | |
| `arcTo(...)` | `agg2d.h:401`, `agg2d.cpp:1140` | `ArcTo` | close | `internal/agg2d/agg2d_test.go` | |
| `arcRel(...)` | `agg2d.h:407`, `agg2d.cpp:1151` | `ArcRel` | close | `internal/agg2d/agg2d_test.go` | |
| `quadricCurveTo(xCtrl,yCtrl,xTo,yTo)` | `agg2d.h:413`, `agg2d.cpp:1162` | `QuadricCurveTo` | close | `internal/agg2d/agg2d_test.go` | |
| `quadricCurveRel(dxCtrl,dyCtrl,dxTo,dyTo)` | `agg2d.h:415`, `agg2d.cpp:1170` | `QuadricCurveRel` | close | `internal/agg2d/agg2d_test.go` | |
| `quadricCurveTo(xTo,yTo)` | `agg2d.h:417`, `agg2d.cpp:1178` | `QuadricCurveToSmooth` | close | `internal/agg2d/agg2d_test.go` | Name differs to avoid overload in Go. |
| `quadricCurveRel(dxTo,dyTo)` | `agg2d.h:418`, `agg2d.cpp:1185` | `QuadricCurveRelSmooth` | close | `internal/agg2d/agg2d_test.go` | Name differs to avoid overload in Go. |
| `cubicCurveTo(xCtrl1,yCtrl1,xCtrl2,yCtrl2,xTo,yTo)` | `agg2d.h:420`, `agg2d.cpp:1192` | `CubicCurveTo` | close | `internal/agg2d/agg2d_test.go` | |
| `cubicCurveRel(dxCtrl1,dyCtrl1,dxCtrl2,dyCtrl2,dxTo,dyTo)` | `agg2d.h:424`, `agg2d.cpp:1201` | `CubicCurveRel` | close | `internal/agg2d/agg2d_test.go` | |
| `cubicCurveTo(xCtrl2,yCtrl2,xTo,yTo)` | `agg2d.h:428`, `agg2d.cpp:1210` | `CubicCurveToSmooth` | close | `internal/agg2d/agg2d_test.go` | Name differs to avoid overload in Go. |
| `cubicCurveRel(xCtrl2,yCtrl2,xTo,yTo)` | `agg2d.h:431`, `agg2d.cpp:1218` | `CubicCurveRelSmooth` | close | `internal/agg2d/agg2d_test.go` | Name differs to avoid overload in Go. |
| `addEllipse(cx,cy,rx,ry,dir)` | `agg2d.h:434`, `agg2d.cpp:1225` | `AddEllipse` | close | `internal/agg2d/agg2d_test.go` | |
| `closePolygon()` | `agg2d.h:435`, `agg2d.cpp:1234` | `ClosePolygon` | close | `internal/agg2d/agg2d_test.go` | |
| `drawPath(flag)` | `agg2d.h:437`, `agg2d.cpp:1367` | `DrawPath` | close | `internal/agg2d/rendering_test.go` | |
| `drawPathNoTransform(flag)` | `agg2d.h:438` | `DrawPathNoTransform` | delta | `internal/agg2d/rendering_test.go` | C++ declaration found, implementation not found in `agg2d.cpp`. |

## Image Transformations

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `imageFilter(f)` | `agg2d.h:443`, `agg2d.cpp:1241` | `ImageFilter` | close | `internal/agg2d/image_test.go` | |
| `imageFilter() const` | `agg2d.h:444`, `agg2d.cpp:1261` | `GetImageFilter` | close | `internal/agg2d/image_test.go` | |
| `imageResample(f)` | `agg2d.h:446`, `agg2d.cpp:1268` | `ImageResample` | close | `internal/agg2d/image_test.go` | |
| `imageResample() const` | `agg2d.h:447`, `agg2d.cpp:1275` | `GetImageResample` | close | `internal/agg2d/image_test.go` | |
| `transformImage(img,srcRect,dstRect)` | `agg2d.h:449`, `agg2d.cpp:1282` | `TransformImage` | placeholder | `internal/agg2d/image_test.go` | Simplified path currently. |
| `transformImage(img,dstRect)` | `agg2d.h:453`, `agg2d.cpp:1296` | `TransformImageSimple` | placeholder | `internal/agg2d/image_test.go` | |
| `transformImage(img,srcRect,parallelogram)` | `agg2d.h:456`, `agg2d.cpp:1309` | `TransformImageParallelogram` | placeholder | `internal/agg2d/image_test.go` | |
| `transformImage(img,parallelogram)` | `agg2d.h:460`, `agg2d.cpp:1324` | `TransformImageParallelogramSimple` | placeholder | `internal/agg2d/image_test.go` | |
| `transformImagePath(img,srcRect,dstRect)` | `agg2d.h:463`, `agg2d.cpp:1337` | `TransformImagePath` | placeholder | `internal/agg2d/image_test.go` | |
| `transformImagePath(img,dstRect)` | `agg2d.h:467`, `agg2d.cpp:1345` | `TransformImagePathSimple` | placeholder | `internal/agg2d/image_test.go` | |
| `transformImagePath(img,srcRect,parallelogram)` | `agg2d.h:470`, `agg2d.cpp:1352` | `TransformImagePathParallelogram` | placeholder | `internal/agg2d/image_test.go` | |
| `transformImagePath(img,parallelogram)` | `agg2d.h:474`, `agg2d.cpp:1359` | `TransformImagePathParallelogramSimple` | placeholder | `internal/agg2d/image_test.go` | |

## Image Blending and Copy

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `blendImage(img,srcRect,dst,alpha)` | `agg2d.h:478`, `agg2d.cpp:1768` | `BlendImage` | close | `internal/agg2d/image_test.go` | |
| `blendImage(img,dst,alpha)` | `agg2d.h:481`, `agg2d.cpp:1789` | `BlendImageSimple` | close | `internal/agg2d/image_test.go` | |
| `copyImage(img,srcRect,dst)` | `agg2d.h:485`, `agg2d.cpp:1806` | `CopyImage` | close | `internal/agg2d/image_test.go` | |
| `copyImage(img,dst)` | `agg2d.h:488`, `agg2d.cpp:1818` | `CopyImageSimple` | close | `internal/agg2d/image_test.go` | |

## Auxiliary

| C++ Method | C++ Ref | Go Symbol | Status | Tests | Notes |
|---|---|---|---|---|---|
| `pi()` | `agg2d.h:493` | `agg2d.Pi`, `agg.Pi` | close | `internal/agg2d/agg2d_test.go` | |
| `deg2Rad(v)` | `agg2d.h:494` | `agg2d.Deg2Rad`, `agg.Deg2RadFunc` | close | `internal/agg2d/agg2d_test.go` | |
| `rad2Deg(v)` | `agg2d.h:495` | `agg2d.Rad2Deg` | close | `internal/agg2d/agg2d_test.go` | |

## Current Priority Gaps

- Image transform/render pipeline fidelity (`TransformImage*` and `renderImage` internals).
- Gradient world/screen transform parity (`FillRadialGradient*`, `LineRadialGradient*` internals).
- Text raster rendering fidelity (`Text`, `TextWidth`, glyph scanline rendering).
