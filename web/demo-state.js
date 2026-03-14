// --- Per-demo persist / restore ---

import { getURLParams, updateURL } from "./url-state.js";

export const demoURLHandlers = {
  aa: {
    persist() {
      const n = getAANodes();
      const zoom = parseFloat(document.getElementById("zoomSlider").value);
      updateURL({
        zoom,
        x0: n.x0.toFixed(1),
        y0: n.y0.toFixed(1),
        x1: n.x1.toFixed(1),
        y1: n.y1.toFixed(1),
        x2: n.x2.toFixed(1),
        y2: n.y2.toFixed(1),
      });
    },
    restore(p) {
      const zoom = p.has("zoom") ? parseFloat(p.get("zoom")) : null;
      if (zoom !== null) {
        setAAZoom(zoom);
        document.getElementById("zoomSlider").value = zoom;
        document.getElementById("zoomValue").textContent = zoom;
      }
      const keys = ["x0", "y0", "x1", "y1", "x2", "y2"];
      if (keys.every((k) => p.has(k))) {
        const vals = keys.map((k) => parseFloat(p.get(k)));
        setAANodes(...vals);
      }
    },
  },

  conv_dash_marker: {
    persist() {
      const n = getDashNodes();
      updateURL({
        dw: parseFloat(document.getElementById("dashWidthSlider").value),
        ds: parseFloat(document.getElementById("dashSmoothSlider").value),
        dcap: document.getElementById("dashCapSelector").value,
        dc: document.getElementById("dashClosed").checked ? "1" : "0",
        deo: document.getElementById("dashEvenOdd").checked ? "1" : "0",
        dx0: n.x0.toFixed(1),
        dy0: n.y0.toFixed(1),
        dx1: n.x1.toFixed(1),
        dy1: n.y1.toFixed(1),
        dx2: n.x2.toFixed(1),
        dy2: n.y2.toFixed(1),
      });
    },
    restore(p) {
      if (p.has("dw")) {
        const val = parseFloat(p.get("dw"));
        setDashWidth(val);
        document.getElementById("dashWidthSlider").value = val;
        document.getElementById("dashWidthValue").textContent = val;
      }
      if (p.has("ds")) {
        const val = parseFloat(p.get("ds"));
        setDashSmooth(val);
        document.getElementById("dashSmoothSlider").value = val;
        document.getElementById("dashSmoothValue").textContent = val.toFixed(2);
      }
      if (p.has("dcap")) {
        const val = parseInt(p.get("dcap"));
        setDashCap(val);
        document.getElementById("dashCapSelector").value = val;
      }
      if (p.has("dc")) {
        const checked = p.get("dc") === "1";
        setDashClosed(checked);
        document.getElementById("dashClosed").checked = checked;
      }
      if (p.has("deo")) {
        const checked = p.get("deo") === "1";
        setDashEvenOdd(checked);
        document.getElementById("dashEvenOdd").checked = checked;
      }
      const keys = ["dx0", "dy0", "dx1", "dy1", "dx2", "dy2"];
      if (keys.every((k) => p.has(k))) {
        const vals = keys.map((k) => parseFloat(p.get(k)));
        setDashNodes(...vals);
      }
    },
  },

  gouraud: {
    persist() {
      const n = getGouraudNodes();
      updateURL({
        dil: parseFloat(document.getElementById("dilationSlider").value),
        gopa: parseFloat(document.getElementById("gouraudOpacitySlider").value),
        gx0: n.x0.toFixed(1),
        gy0: n.y0.toFixed(1),
        gx1: n.x1.toFixed(1),
        gy1: n.y1.toFixed(1),
        gx2: n.x2.toFixed(1),
        gy2: n.y2.toFixed(1),
      });
    },
    restore(p) {
      if (p.has("dil")) {
        const val = parseFloat(p.get("dil"));
        setGouraudDilation(val);
        document.getElementById("dilationSlider").value = val;
        document.getElementById("dilationValue").textContent = val;
      }
      if (p.has("gopa")) {
        const val = parseFloat(p.get("gopa"));
        setGouraudOpacity(val);
        document.getElementById("gouraudOpacitySlider").value = val;
        document.getElementById("gouraudOpacityValue").textContent =
          val.toFixed(2);
      }
      const keys = ["gx0", "gy0", "gx1", "gy1", "gx2", "gy2"];
      if (keys.every((k) => p.has(k))) {
        const vals = keys.map((k) => parseFloat(p.get(k)));
        setGouraudNodes(...vals);
      }
    },
  },

  imagefilters: {
    persist() {
      const flt = parseInt(document.getElementById("filterSelector").value);
      const frad = parseFloat(
        document.getElementById("filterRadiusSlider").value,
      );
      updateURL({ flt, frad });
    },
    restore(p) {
      if (p.has("flt")) {
        const val = parseInt(p.get("flt"));
        setImageFilter(val);
        document.getElementById("filterSelector").value = val;
        const hasRadius = val >= 12;
        document.getElementById("radiusLabel").style.display = hasRadius
          ? "inline"
          : "none";
        document.getElementById("filterRadiusSlider").style.display = hasRadius
          ? "inline"
          : "none";
        document.getElementById("filterRadiusValue").style.display = hasRadius
          ? "inline"
          : "none";
      }
      if (p.has("frad")) {
        const val = parseFloat(p.get("frad"));
        const flt = parseInt(document.getElementById("filterSelector").value);
        setImageFilterRadius(flt, val);
        document.getElementById("filterRadiusSlider").value = val;
        document.getElementById("filterRadiusValue").textContent = val;
      }
    },
  },

  image_fltr_graph: {
    persist() {
      let mask = 0;
      document.querySelectorAll(".ifg-filter").forEach((el) => {
        const idx = parseInt(el.dataset.index);
        if (el.checked) {
          mask |= 1 << idx;
        }
      });
      updateURL({
        ifgr: parseFloat(document.getElementById("ifgRadiusSlider").value),
        ifgm: mask >>> 0,
      });
    },
    restore(p) {
      if (p.has("ifgr")) {
        const val = parseFloat(p.get("ifgr"));
        setImageFltrGraphRadius(val);
        document.getElementById("ifgRadiusSlider").value = val;
        document.getElementById("ifgRadiusValue").textContent = val.toFixed(1);
      } else {
        const val = parseFloat(
          document.getElementById("ifgRadiusSlider").value,
        );
        setImageFltrGraphRadius(val);
        document.getElementById("ifgRadiusValue").textContent = val.toFixed(1);
      }

      if (p.has("ifgm")) {
        const mask = parseInt(p.get("ifgm")) >>> 0;
        document.querySelectorAll(".ifg-filter").forEach((el) => {
          const idx = parseInt(el.dataset.index);
          el.checked = ((mask >> idx) & 1) !== 0;
        });
      }

      let mask = 0;
      document.querySelectorAll(".ifg-filter").forEach((el) => {
        const idx = parseInt(el.dataset.index);
        if (el.checked) {
          mask |= 1 << idx;
        }
      });
      setImageFltrGraphMask(mask >>> 0);
    },
  },

  image_filters2: {
    persist() {
      updateURL({
        if2f: parseInt(document.getElementById("if2FilterSelector").value),
        if2r: parseFloat(document.getElementById("if2RadiusSlider").value),
        if2n: document.getElementById("if2Normalize").checked ? "1" : "0",
      });
    },
    restore(p) {
      const syncRadiusVisibility = () => {
        const hasRadius =
          parseInt(document.getElementById("if2FilterSelector").value) >= 14;
        document.getElementById("if2RadiusLabel").style.display = hasRadius
          ? "inline"
          : "none";
        document.getElementById("if2RadiusSlider").style.display = hasRadius
          ? "inline"
          : "none";
        document.getElementById("if2RadiusValue").style.display = hasRadius
          ? "inline"
          : "none";
      };
      if (p.has("if2f")) {
        const val = parseInt(p.get("if2f"));
        setImageFilters2Filter(val);
        document.getElementById("if2FilterSelector").value = val;
      } else {
        setImageFilters2Filter(
          parseInt(document.getElementById("if2FilterSelector").value),
        );
      }
      if (p.has("if2r")) {
        const val = parseFloat(p.get("if2r"));
        setImageFilters2Radius(val);
        document.getElementById("if2RadiusSlider").value = val;
        document.getElementById("if2RadiusValue").textContent = val.toFixed(1);
      } else {
        document.getElementById("if2RadiusValue").textContent = parseFloat(
          document.getElementById("if2RadiusSlider").value,
        ).toFixed(1);
      }
      if (p.has("if2n")) {
        const checked = p.get("if2n") === "1";
        setImageFilters2Normalize(checked);
        document.getElementById("if2Normalize").checked = checked;
      } else {
        setImageFilters2Normalize(
          document.getElementById("if2Normalize").checked,
        );
      }
      syncRadiusVisibility();
    },
  },

  sbool: {
    persist() {
      const n = getSBoolNodes();
      updateURL({
        sop: parseInt(document.getElementById("sboolOpSelector").value),
        p1x0: n.p1x0.toFixed(1),
        p1y0: n.p1y0.toFixed(1),
        p1x1: n.p1x1.toFixed(1),
        p1y1: n.p1y1.toFixed(1),
        p1x2: n.p1x2.toFixed(1),
        p1y2: n.p1y2.toFixed(1),
        p1x3: n.p1x3.toFixed(1),
        p1y3: n.p1y3.toFixed(1),
        p2x0: n.p2x0.toFixed(1),
        p2y0: n.p2y0.toFixed(1),
        p2x1: n.p2x1.toFixed(1),
        p2y1: n.p2y1.toFixed(1),
        p2x2: n.p2x2.toFixed(1),
        p2y2: n.p2y2.toFixed(1),
        p2x3: n.p2x3.toFixed(1),
        p2y3: n.p2y3.toFixed(1),
      });
    },
    restore(p) {
      if (p.has("sop")) {
        const val = parseInt(p.get("sop"));
        setSBoolOp(val);
        document.getElementById("sboolOpSelector").value = val;
      }
      const keys = [
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
      ];
      if (keys.every((k) => p.has(k))) {
        const vals = keys.map((k) => parseFloat(p.get(k)));
        setSBoolNodes(...vals);
      }
    },
  },

  convstroke: {
    persist() {
      const n = getStrokeNodes();
      updateURL({
        sj: parseInt(document.getElementById("strokeJoinSelector").value),
        sc: parseInt(document.getElementById("strokeCapSelector").value),
        sw: parseFloat(document.getElementById("strokeWidthSlider").value),
        sml: parseFloat(document.getElementById("strokeMiterSlider").value),
        sx0: n.x0.toFixed(1),
        sy0: n.y0.toFixed(1),
        sx1: n.x1.toFixed(1),
        sy1: n.y1.toFixed(1),
        sx2: n.x2.toFixed(1),
        sy2: n.y2.toFixed(1),
      });
    },
    restore(p) {
      if (p.has("sj")) {
        const val = parseInt(p.get("sj"));
        setStrokeJoin(val);
        document.getElementById("strokeJoinSelector").value = val;
      }
      if (p.has("sc")) {
        const val = parseInt(p.get("sc"));
        setStrokeCap(val);
        document.getElementById("strokeCapSelector").value = val;
      }
      if (p.has("sw")) {
        const val = parseFloat(p.get("sw"));
        setStrokeWidth(val);
        document.getElementById("strokeWidthSlider").value = val;
        document.getElementById("strokeWidthValue").textContent = val;
      }
      if (p.has("sml")) {
        const val = parseFloat(p.get("sml"));
        setStrokeMiterLimit(val);
        document.getElementById("strokeMiterSlider").value = val;
        document.getElementById("strokeMiterValue").textContent = val;
      }
      const keys = ["sx0", "sy0", "sx1", "sy1", "sx2", "sy2"];
      if (keys.every((k) => p.has(k))) {
        const vals = keys.map((k) => parseFloat(p.get(k)));
        setStrokeNodes(...vals);
      }
    },
  },

  convcontour: {
    persist() {
      updateURL({
        ccw: parseFloat(document.getElementById("contourWidthSlider").value),
        ccm: parseInt(
          document.getElementById("contourCloseModeSelector").value,
        ),
        ccad: document.getElementById("contourAutoDetect").checked ? "1" : "0",
      });
    },
    restore(p) {
      if (p.has("ccw")) {
        const val = parseFloat(p.get("ccw"));
        setContourWidth(val);
        document.getElementById("contourWidthSlider").value = val;
        document.getElementById("contourWidthValue").textContent = val;
      }
      if (p.has("ccm")) {
        const val = parseInt(p.get("ccm"));
        setContourCloseMode(val);
        document.getElementById("contourCloseModeSelector").value = val;
      }
      if (p.has("ccad")) {
        const checked = p.get("ccad") === "1";
        setContourAutoDetect(checked);
        document.getElementById("contourAutoDetect").checked = checked;
      }
    },
  },

  lionoutline: {
    persist() {
      updateURL({
        low: parseFloat(
          document.getElementById("lionOutlineWidthSlider").value,
        ),
      });
    },
    restore(p) {
      if (p.has("low")) {
        const val = parseFloat(p.get("low"));
        setLionOutlineWidth(val);
        document.getElementById("lionOutlineWidthSlider").value = val;
        document.getElementById("lionOutlineWidthValue").textContent =
          val.toFixed(1);
      }
    },
  },

  roundedrect: {
    persist() {
      const n = getRRNodes();
      updateURL({
        rrr: parseFloat(document.getElementById("rrRadiusSlider").value),
        rro: parseFloat(document.getElementById("rrOffsetSlider").value),
        rrd: document.getElementById("rrDarkBg").checked ? "1" : "0",
        rrx0: n.x0.toFixed(1),
        rry0: n.y0.toFixed(1),
        rrx1: n.x1.toFixed(1),
        rry1: n.y1.toFixed(1),
      });
    },
    restore(p) {
      if (p.has("rrr")) {
        const val = parseFloat(p.get("rrr"));
        setRRRadius(val);
        document.getElementById("rrRadiusSlider").value = val;
        document.getElementById("rrRadiusValue").textContent = val.toFixed(1);
      }
      if (p.has("rro")) {
        const val = parseFloat(p.get("rro"));
        setRROffset(val);
        document.getElementById("rrOffsetSlider").value = val;
        document.getElementById("rrOffsetValue").textContent = val.toFixed(1);
      }
      if (p.has("rrd")) {
        const checked = p.get("rrd") === "1";
        setRRDarkBg(checked);
        document.getElementById("rrDarkBg").checked = checked;
      }
      const keys = ["rrx0", "rry0", "rrx1", "rry1"];
      if (keys.every((k) => p.has(k))) {
        const vals = keys.map((k) => parseFloat(p.get(k)));
        setRRNodes(...vals);
      }
    },
  },

  component: {
    persist() {
      updateURL({
        ca: parseInt(document.getElementById("compAlphaSlider").value),
      });
    },
    restore(p) {
      if (p.has("ca")) {
        const val = parseInt(p.get("ca"));
        setCompAlpha(val);
        document.getElementById("compAlphaSlider").value = val;
        document.getElementById("compAlphaValue").textContent = val;
      }
    },
  },

  alphagrad: {
    persist() {
      const n = getAlphaGradNodes();
      updateURL({
        agx0: n.x0.toFixed(1),
        agy0: n.y0.toFixed(1),
        agx1: n.x1.toFixed(1),
        agy1: n.y1.toFixed(1),
        agx2: n.x2.toFixed(1),
        agy2: n.y2.toFixed(1),
      });
    },
    restore(p) {
      const keys = ["agx0", "agy0", "agx1", "agy1", "agx2", "agy2"];
      if (keys.every((k) => p.has(k))) {
        const vals = keys.map((k) => parseFloat(p.get(k)));
        setAlphaGradNodes(...vals);
      }
    },
  },

  perspective: {
    persist() {
      updateURL({
        pt: parseInt(document.getElementById("perspectiveTypeSelector").value),
      });
    },
    restore(p) {
      if (p.has("pt")) {
        const val = parseInt(p.get("pt"));
        setPerspectiveType(val);
        document.getElementById("perspectiveTypeSelector").value = val;
      }
    },
  },

  compositing: {
    persist() {
      updateURL({
        compop: document.getElementById("compOpSelector").value,
        cas: parseFloat(document.getElementById("compAlphaSrcSlider").value),
        cad: parseFloat(document.getElementById("compAlphaDstSlider").value),
      });
    },
    restore(p) {
      if (p.has("compop")) {
        const val = parseInt(p.get("compop"));
        setCompOp(val);
        document.getElementById("compOpSelector").value = val;
      }
      if (p.has("cas")) {
        const val = parseFloat(p.get("cas"));
        setCompAlphaSrc(val);
        document.getElementById("compAlphaSrcSlider").value = val;
        document.getElementById("compAlphaSrcValue").textContent =
          val.toFixed(2);
      }
      if (p.has("cad")) {
        const val = parseFloat(p.get("cad"));
        setCompAlphaDst(val);
        document.getElementById("compAlphaDstSlider").value = val;
        document.getElementById("compAlphaDstValue").textContent =
          val.toFixed(2);
      }
    },
  },

  multi_clip: {
    persist() {
      updateURL({
        mcn: parseFloat(document.getElementById("multiClipNSlider").value),
      });
    },
    restore(p) {
      if (p.has("mcn")) {
        const val = parseFloat(p.get("mcn"));
        setMultiClipN(val);
        document.getElementById("multiClipNSlider").value = val;
        document.getElementById("multiClipNValue").textContent = val;
      }
    },
  },

  gradient_focal: {
    persist() {
      // URL-only controls (no visible widget panel).
    },
    restore(p) {
      if (p.has("gfg")) {
        setGradientFocalGamma(parseFloat(p.get("gfg")));
      }
      if (p.has("gfx")) {
        setGradientFocalFX(parseFloat(p.get("gfx")));
      }
      if (p.has("gfy")) {
        setGradientFocalFY(parseFloat(p.get("gfy")));
      }
    },
  },

  line_thickness: {
    persist() {
      // URL-only controls (no visible widget panel).
    },
    restore(p) {
      if (p.has("ltf")) {
        setLineThicknessFactor(parseFloat(p.get("ltf")));
      }
      if (p.has("ltb")) {
        setLineThicknessBlur(parseFloat(p.get("ltb")));
      }
      if (p.has("ltm")) {
        setLineThicknessMono(p.get("ltm") === "1");
      }
      if (p.has("lti")) {
        setLineThicknessInvert(p.get("lti") === "1");
      }
    },
  },

  rasterizer_compound: {
    persist() {
      // URL-only controls (no visible widget panel).
    },
    restore(p) {
      if (p.has("rcw")) {
        setCompoundWidth(parseFloat(p.get("rcw")));
      }
      if (p.has("rca1")) {
        setCompoundAlpha1(parseFloat(p.get("rca1")));
      }
      if (p.has("rca2")) {
        setCompoundAlpha2(parseFloat(p.get("rca2")));
      }
      if (p.has("rca3")) {
        setCompoundAlpha3(parseFloat(p.get("rca3")));
      }
      if (p.has("rca4")) {
        setCompoundAlpha4(parseFloat(p.get("rca4")));
      }
      if (p.has("rcio")) {
        setCompoundInvert(p.get("rcio") === "1");
      }
    },
  },

  image_resample: {
    persist() {
      // URL-only controls (no visible widget panel).
    },
    restore(p) {
      if (p.has("irt")) {
        setImageResampleType(parseInt(p.get("irt"), 10));
      }
      if (p.has("irb")) {
        setImageResampleBlur(parseFloat(p.get("irb")));
      }
      const keys = [
        "irx0",
        "iry0",
        "irx1",
        "iry1",
        "irx2",
        "iry2",
        "irx3",
        "iry3",
      ];
      if (keys.every((k) => p.has(k))) {
        setImageResampleQuad(
          parseFloat(p.get("irx0")),
          parseFloat(p.get("iry0")),
          parseFloat(p.get("irx1")),
          parseFloat(p.get("iry1")),
          parseFloat(p.get("irx2")),
          parseFloat(p.get("iry2")),
          parseFloat(p.get("irx3")),
          parseFloat(p.get("iry3")),
        );
      }
    },
  },
  image_perspective: {
    persist() {
      // URL-only controls (no visible widget panel).
    },
    restore(p) {
      if (p.has("ipt")) {
        setImagePerspectiveType(parseInt(p.get("ipt"), 10));
      }
      const keys = [
        "ipx0",
        "ipy0",
        "ipx1",
        "ipy1",
        "ipx2",
        "ipy2",
        "ipx3",
        "ipy3",
      ];
      if (keys.every((k) => p.has(k))) {
        setImagePerspectiveQuad(
          parseFloat(p.get("ipx0")),
          parseFloat(p.get("ipy0")),
          parseFloat(p.get("ipx1")),
          parseFloat(p.get("ipy1")),
          parseFloat(p.get("ipx2")),
          parseFloat(p.get("ipy2")),
          parseFloat(p.get("ipx3")),
          parseFloat(p.get("ipy3")),
        );
      }
    },
  },
  pattern_perspective: {
    persist() {
      updateURL({
        ppt: parseInt(
          document.getElementById("patternPerspectiveTypeSelector").value,
          10,
        ),
      });
    },
    restore(p) {
      if (p.has("ppt")) {
        const val = parseInt(p.get("ppt"), 10);
        setPatternPerspectiveType(val);
        document.getElementById("patternPerspectiveTypeSelector").value = val;
      }
      const keys = [
        "ppx0",
        "ppy0",
        "ppx1",
        "ppy1",
        "ppx2",
        "ppy2",
        "ppx3",
        "ppy3",
      ];
      if (keys.every((k) => p.has(k))) {
        setPatternPerspectiveQuad(
          parseFloat(p.get("ppx0")),
          parseFloat(p.get("ppy0")),
          parseFloat(p.get("ppx1")),
          parseFloat(p.get("ppy1")),
          parseFloat(p.get("ppx2")),
          parseFloat(p.get("ppy2")),
          parseFloat(p.get("ppx3")),
          parseFloat(p.get("ppy3")),
        );
      }
    },
  },
  pattern_resample: {
    persist() {
      updateURL({
        prt: parseInt(
          document.getElementById("patternResampleTypeSelector").value,
          10,
        ),
        prg: parseFloat(
          document.getElementById("patternResampleGammaSlider").value,
        ),
        prb: parseFloat(
          document.getElementById("patternResampleBlurSlider").value,
        ),
      });
    },
    restore(p) {
      if (p.has("prt")) {
        const val = parseInt(p.get("prt"), 10);
        setPatternResampleType(val);
        document.getElementById("patternResampleTypeSelector").value = val;
      }
      if (p.has("prg")) {
        const val = parseFloat(p.get("prg"));
        setPatternResampleGamma(val);
        document.getElementById("patternResampleGammaSlider").value = val;
        document.getElementById("patternResampleGammaValue").textContent =
          val.toFixed(2);
      }
      if (p.has("prb")) {
        const val = parseFloat(p.get("prb"));
        setPatternResampleBlur(val);
        document.getElementById("patternResampleBlurSlider").value = val;
        document.getElementById("patternResampleBlurValue").textContent =
          val.toFixed(2);
      }
      const keys = [
        "prx0",
        "pry0",
        "prx1",
        "pry1",
        "prx2",
        "pry2",
        "prx3",
        "pry3",
      ];
      if (keys.every((k) => p.has(k))) {
        setPatternResampleQuad(
          parseFloat(p.get("prx0")),
          parseFloat(p.get("pry0")),
          parseFloat(p.get("prx1")),
          parseFloat(p.get("pry1")),
          parseFloat(p.get("prx2")),
          parseFloat(p.get("pry2")),
          parseFloat(p.get("prx3")),
          parseFloat(p.get("pry3")),
        );
      }
    },
  },

  line_patterns_clip: {
    persist() {
      // URL-only controls (no visible widget panel).
    },
    restore(p) {
      if (p.has("lpcsx")) {
        setLinePatternClipScaleX(parseFloat(p.get("lpcsx")));
      }
      if (p.has("lpcst")) {
        setLinePatternClipStartX(parseFloat(p.get("lpcst")));
      }
    },
  },

  line_patterns: {
    persist() {
      // URL-only controls (no visible widget panel).
    },
    restore(p) {
      if (p.has("lpsx")) {
        setLinePatternScaleX(parseFloat(p.get("lpsx")));
      }
      if (p.has("lpst")) {
        setLinePatternStartX(parseFloat(p.get("lpst")));
      }
    },
  },

  scanline_boolean2: {
    persist() {
      // URL-only controls (no visible widget panel).
    },
    restore(p) {
      if (p.has("sb2m")) {
        setScanlineBoolean2Mode(parseInt(p.get("sb2m"), 10));
      }
      if (p.has("sb2f")) {
        setScanlineBoolean2FillRule(parseInt(p.get("sb2f"), 10));
      }
      if (p.has("sb2s")) {
        setScanlineBoolean2Scanline(parseInt(p.get("sb2s"), 10));
      }
      if (p.has("sb2o")) {
        setScanlineBoolean2Operation(parseInt(p.get("sb2o"), 10));
      }
      if (p.has("sb2x") && p.has("sb2y")) {
        setScanlineBoolean2Center(
          parseFloat(p.get("sb2x")),
          parseFloat(p.get("sb2y")),
        );
      }
    },
  },

  gpc_test: {
    persist() {
      updateURL({
        gsc: parseInt(document.getElementById("gpcSceneSelector").value, 10),
        gop: parseInt(document.getElementById("gpcOpSelector").value, 10),
      });
    },
    restore(p) {
      if (p.has("gsc")) {
        const val = parseInt(p.get("gsc"), 10);
        setGPCTestScene(val);
        document.getElementById("gpcSceneSelector").value = String(val);
      }
      if (p.has("gop")) {
        const val = parseInt(p.get("gop"), 10);
        setGPCTestOperation(val);
        document.getElementById("gpcOpSelector").value = String(val);
      }
      if (p.has("gcx") && p.has("gcy")) {
        setGPCTestCenter(parseFloat(p.get("gcx")), parseFloat(p.get("gcy")));
      }
    },
  },

  gradients_contour: {
    persist() {
      updateURL({
        gcp: parseInt(document.getElementById("gcPolygonSelector").value, 10),
        gcg: parseInt(document.getElementById("gcGradientSelector").value, 10),
        gcr: document.getElementById("gcReflect").checked ? 1 : 0,
        gcc: parseInt(document.getElementById("gcColorsSlider").value, 10),
        gc1: parseFloat(document.getElementById("gcC1Slider").value),
        gc2: parseFloat(document.getElementById("gcC2Slider").value),
        gd1: parseFloat(document.getElementById("gcD1Slider").value),
        gd2: parseFloat(document.getElementById("gcD2Slider").value),
      });
    },
    restore(p) {
      if (p.has("gcp")) {
        const val = parseInt(p.get("gcp"), 10);
        setGradientsContourPolygon(val);
        document.getElementById("gcPolygonSelector").value = String(val);
      }
      if (p.has("gcg")) {
        const val = parseInt(p.get("gcg"), 10);
        setGradientsContourGradient(val);
        document.getElementById("gcGradientSelector").value = String(val);
      }
      if (p.has("gcr")) {
        const val = p.get("gcr") === "1";
        setGradientsContourReflect(val);
        document.getElementById("gcReflect").checked = val;
      }
      if (p.has("gcc")) {
        const val = parseInt(p.get("gcc"), 10);
        setGradientsContourColors(val);
        document.getElementById("gcColorsSlider").value = String(val);
        document.getElementById("gcColorsValue").textContent = String(val);
      }
      if (p.has("gc1")) {
        const val = parseFloat(p.get("gc1"));
        setGradientsContourC1(val);
        document.getElementById("gcC1Slider").value = String(val);
        document.getElementById("gcC1Value").textContent = String(val);
      }
      if (p.has("gc2")) {
        const val = parseFloat(p.get("gc2"));
        setGradientsContourC2(val);
        document.getElementById("gcC2Slider").value = String(val);
        document.getElementById("gcC2Value").textContent = String(val);
      }
      if (p.has("gd1")) {
        const val = parseFloat(p.get("gd1"));
        setGradientsContourD1(val);
        document.getElementById("gcD1Slider").value = String(val);
        document.getElementById("gcD1Value").textContent = String(val);
      }
      if (p.has("gd2")) {
        const val = parseFloat(p.get("gd2"));
        setGradientsContourD2(val);
        document.getElementById("gcD2Slider").value = String(val);
        document.getElementById("gcD2Value").textContent = String(val);
      }
    },
  },

  flash_rasterizer2: {
    persist() {
      updateURL({
        fr2s: parseInt(document.getElementById("fr2ShapeSlider").value, 10),
      });
    },
    restore(p) {
      if (p.has("fr2s")) {
        const val = parseInt(p.get("fr2s"), 10);
        setFlash2ShapeIdx(val);
        document.getElementById("fr2ShapeSlider").value = String(val);
        document.getElementById("fr2ShapeValue").textContent = String(val);
      }
    },
  },

  distortions: {
    persist() {
      updateURL({
        dimg: parseInt(
          document.getElementById("distortionsImageSelector").value,
          10,
        ),
      });
    },
    restore(p) {
      if (p.has("dimg")) {
        const val = parseInt(p.get("dimg"), 10);
        setDistortionsImage(val);
        document.getElementById("distortionsImageSelector").value = String(val);
      }
    },
  },
};

export function persistDemoParams(demoType) {
  demoURLHandlers[demoType]?.persist();
}

export function restoreDemoParams(demoType, params) {
  demoURLHandlers[demoType]?.restore(params);
}
