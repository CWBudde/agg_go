# Code TODOs

This checklist tracks remaining TODO items and missing features in the AGG Go port. Regenerate TODO comments with:

`rg -n "TODO|FIXME|XXX|HACK" --glob "**/*.go" -S --sort path`

## Core Library TODOs

- [ ] **Font System**

  - [ ] `internal/font/freetype2/engine.go:83`: Support custom memory management (optional enhancement)
