# Code TODOs

This checklist is generated from TODO-like comments found in Go source files (`TODO`, `FIXME`, `XXX`, `HACK`). Each file groups its items with line numbers and short descriptions so you can track progress. Regenerate with:

`rg -n "TODO|FIXME|XXX|HACK" --glob "**/*.go" -S --sort path`

- [ ] internal/font/freetype2/cache_integration.go

  - [ ] L47: Convert to Fman adaptors when available (optional enhancement)
  - [ ] L58: Convert to Fman adaptors when available (optional enhancement)

- [ ] internal/font/freetype2/engine.go

  - [ ] L83: Support custom memory management if needed (optional enhancement)

- [ ] internal/font/freetype2/types.go

  - [ ] L168: Add conv_curve wrapper for int16 paths (optional enhancement)
  - [ ] L169: Add conv_curve wrapper for int32 paths (optional enhancement)

- [ ] internal/fonts/embedded_fonts.go

  - [x] L1056: Implement GSE4x8 font data
  - [x] L1063: Implement GSE5x9 font data
  - [x] L1070: Implement GSE6x9 font data
  - [ ] L1077: Implement GSE6x12 font data
  - [ ] L1084: Implement GSE7x11 font data
  - [ ] L1091: Implement GSE7x11_bold font data
  - [ ] L1098: Implement GSE7x15 font data
  - [ ] L1105: Implement GSE7x15_bold font data
  - [ ] L1112: Implement GSE8x16 font data
  - [ ] L1119: Implement GSE8x16_bold font data
  - [ ] L1134: Implement MCS5x11_mono font data
  - [ ] L1141: Implement MCS6x10_mono font data
  - [ ] L1148: Implement MCS6x11_mono font data
  - [ ] L1155: Implement MCS7x12_mono_high font data
  - [ ] L1162: Implement MCS7x12_mono_low font data
  - [ ] L1169: Implement MCS11_prop font data
  - [ ] L1176: Implement MCS11_prop_condensed font data
  - [ ] L1183: Implement MCS12_prop font data
  - [ ] L1190: Implement MCS13_prop font data
  - [ ] L1205: Implement Verdana12_bold font data
  - [ ] L1212: Implement Verdana13 font data
  - [ ] L1219: Implement Verdana13_bold font data
  - [ ] L1226: Implement Verdana14 font data
  - [ ] L1233: Implement Verdana14_bold font data
  - [ ] L1240: Implement Verdana16 font data
  - [ ] L1247: Implement Verdana16_bold font data
  - [ ] L1254: Implement Verdana17 font data
  - [ ] L1261: Implement Verdana17_bold font data
  - [ ] L1268: Implement Verdana18 font data
  - [ ] L1275: Implement Verdana18_bold font data

# Missing Features

- [ ] UI Toolkit: The project lacks a UI toolkit for creating interactive examples with controls like sliders and checkboxes. This makes it difficult to create faithful ports of the original C++ examples that use `agg::ctrl`. The `line_thickness` example was ported without the interactive controls.