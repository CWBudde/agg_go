This directory stores canonical screenshots captured from the precompiled C++
AGG examples in `../agg-2.6/build/examples/`.

Capture details:

- Captured on 2026-03-16.
- Source binaries were run with the direct `--screenshot` flag.
- The X11 platform code renders offscreen and writes `screenshot.ppm`.
- The captured `PPM` files were converted to `PNG` and copied into
  `examples/` here.

Current scope:

- 60 example screenshots imported successfully.
- This is an initial canonical C++ corpus, not yet a complete mirror of the
  entire AGG example set.
- These files are intended to drive visual-parity work in Phase 8.2 and,
  where mapped to parity rows, support Phase 8.3.

Next work:

- Map individual C++ screenshots to specific Go demo or visual-test scenarios.
- Re-run the remaining C++ examples with a more robust retry/stability pass for
  the demos that did not capture in the first import.
- Replace Go-derived references with these C++ references where the scenario is
  a direct match.
