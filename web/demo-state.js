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
};

export function persistDemoParams(demoType) {
  demoURLHandlers[demoType]?.persist();
}

export function restoreDemoParams(demoType, params) {
  demoURLHandlers[demoType]?.restore(params);
}
