// --- URL parameter helpers ---

export function getURLParams() {
  return new URLSearchParams(window.location.search);
}

export function updateURL(params) {
  const url = new URL(window.location);
  for (const [key, value] of Object.entries(params)) {
    if (value === null || value === undefined) {
      url.searchParams.delete(key);
    } else {
      url.searchParams.set(key, value);
    }
  }
  history.replaceState(null, "", url);
}

// All known demo-specific URL parameter names.
// When switching demos, all of these are cleared so each demo starts clean.
export const ALL_DEMO_PARAMS = [
  // aa
  "zoom",
  "x0",
  "y0",
  "x1",
  "y1",
  "x2",
  "y2",
  // conv_dash_marker
  "dw",
  "ds",
  "dcap",
  "dc",
  "deo",
  "dx0",
  "dy0",
  "dx1",
  "dy1",
  "dx2",
  "dy2",
  // gouraud
  "dil",
  "gopa",
  "gx0",
  "gy0",
  "gx1",
  "gy1",
  "gx2",
  "gy2",
  // imagefilters
  "flt",
  "frad",
  // image_fltr_graph
  "ifgr",
  "ifgm",
  // image_filters2
  "if2f",
  "if2r",
  "if2n",
  // idea
  "idr",
  "ideo",
  "idd",
  "idro",
  "ids",
  // mol_view
  "mvm",
  "mvt",
  "mvz",
  "mvr",
  // sbool
  "sop",
  "p1x0",
  "p1y0",
  "p1x1",
  "p1y1",
  "p1x2",
  "p1y2",
  "p1x3",
  "p1y3",
  "p2x0",
  "p2y0",
  "p2x1",
  "p2y1",
  "p2x2",
  "p2y2",
  "p2x3",
  "p2y3",
  // convstroke
  "sj",
  "sc",
  "sw",
  "sml",
  "sx0",
  "sy0",
  "sx1",
  "sy1",
  "sx2",
  "sy2",
  // convcontour
  "ccw",
  "ccm",
  "ccad",
  // lionoutline
  "low",
  // roundedrect
  "rrr",
  "rro",
  "rrd",
  "rrx0",
  "rry0",
  "rrx1",
  "rry1",
  // component
  "ca",
  // alphagrad
  "agx0",
  "agy0",
  "agx1",
  "agy1",
  "agx2",
  "agy2",
  // perspective
  "pt",
  // compositing
  "compop",
  "cas",
  "cad",
  // multi_clip
  "mcn",
  // gradient_focal
  "gfg",
  "gfx",
  "gfy",
  // line_thickness
  "ltf",
  "ltb",
  "ltm",
  "lti",
  // rasterizer_compound
  "rcw",
  "rca1",
  "rca2",
  "rca3",
  "rca4",
  "rcio",
  // image_resample
  "irt",
  "irb",
  "irx0",
  "iry0",
  "irx1",
  "iry1",
  "irx2",
  "iry2",
  "irx3",
  "iry3",
  // image_perspective
  "ipt",
  "ipx0",
  "ipy0",
  "ipx1",
  "ipy1",
  "ipx2",
  "ipy2",
  "ipx3",
  "ipy3",
  // pattern_perspective
  "ppt",
  "ppx0",
  "ppy0",
  "ppx1",
  "ppy1",
  "ppx2",
  "ppy2",
  "ppx3",
  "ppy3",
  // pattern_resample
  "prt",
  "prg",
  "prb",
  "prx0",
  "pry0",
  "prx1",
  "pry1",
  "prx2",
  "pry2",
  "prx3",
  "pry3",
  // line_patterns_clip
  "lpcsx",
  "lpcst",
  "lpcp",
  // line_patterns
  "lpsx",
  "lpst",
  "lpp",
  // scanline_boolean2
  "sb2m",
  "sb2f",
  "sb2s",
  "sb2o",
  "sb2x",
  "sb2y",
  // distortions
  "dimg",
];

export function clearDemoParams() {
  const nulls = {};
  for (const k of ALL_DEMO_PARAMS) nulls[k] = null;
  updateURL(nulls);
}
