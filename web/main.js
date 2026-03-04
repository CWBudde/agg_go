"use strict";

const go = new Go();
let wasmInstance;
let canvas;
let ctx;
let imageData;
let pixels;

window.onerror = function (message, source, lineno, colno, error) {
  console.error("Global JS Error:", message, "at", source, ":", lineno);
  updateStatus("JS Error: " + message);
};

// --- URL parameter helpers ---

function getURLParams() {
  return new URLSearchParams(window.location.search);
}

function updateURL(params) {
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
const ALL_DEMO_PARAMS = [
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
  "gx0",
  "gy0",
  "gx1",
  "gy1",
  "gx2",
  "gy2",
  // imagefilters
  "flt",
  "frad",
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
];

function clearDemoParams() {
  const nulls = {};
  for (const k of ALL_DEMO_PARAMS) nulls[k] = null;
  updateURL(nulls);
}

// --- Per-demo persist / restore ---

const demoURLHandlers = {
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

function persistDemoParams(demoType) {
  demoURLHandlers[demoType]?.persist();
}

function restoreDemoParams(demoType, params) {
  demoURLHandlers[demoType]?.restore(params);
}

// --- Theme Management ---

function initTheme() {
  const themeBtn = document.getElementById("themeCycle");
  const themeIcon = document.getElementById("themeIcon");
  const themeLabel = document.getElementById("themeLabel");

  const themes = ["auto", "light", "dark"];
  const themeInfo = {
    auto: { icon: "🌓", label: "Auto" },
    light: { icon: "☀️", label: "Light" },
    dark: { icon: "🌑", label: "Dark" },
  };

  let currentTheme = localStorage.getItem("agg-theme") || "auto";

  const applyTheme = (theme) => {
    const isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
    const effectiveTheme =
      theme === "auto" ? (isDark ? "dark" : "light") : theme;

    document.body.setAttribute("data-theme", theme);
    document.body.setAttribute("data-effective-theme", effectiveTheme);

    // Update button content
    themeIcon.textContent = themeInfo[theme].icon;
    let label = themeInfo[theme].label;
    if (theme === "auto") {
      label += ` (${effectiveTheme === "dark" ? "Dark" : "Light"})`;
    }
    themeLabel.textContent = label;

    localStorage.setItem("agg-theme", theme);
    currentTheme = theme;
  };

  themeBtn.addEventListener("click", () => {
    const nextIndex = (themes.indexOf(currentTheme) + 1) % themes.length;
    applyTheme(themes[nextIndex]);
  });

  // Listen for system theme changes
  window
    .matchMedia("(prefers-color-scheme: dark)")
    .addEventListener("change", () => {
      if (currentTheme === "auto") {
        applyTheme("auto");
      }
    });

  applyTheme(currentTheme);
}

// --- Initialization ---

async function init() {
  console.log("Initializing AGG Go Web Demo...");
  try {
    const result = await WebAssembly.instantiateStreaming(
      fetch("main.wasm"),
      go.importObject,
    ).catch((err) => {
      console.error("WASM Fetch/Instantiate Error:", err);
      throw err;
    });

    wasmInstance = result.instance;
    go.run(wasmInstance).catch((err) => {
      console.error("WASM Runtime Error:", err);
      updateStatus("WASM Error: " + err.message);
    });

    initTheme();

    // Hide loading screen
    document.getElementById("loading").style.display = "none";

    // Setup canvas
    const dims = getCanvasDimensions();
    canvas = document.getElementById("aggCanvas");
    canvas.width = dims.width;
    canvas.height = dims.height;
    ctx = canvas.getContext("2d");

    imageData = ctx.createImageData(dims.width, dims.height);
    pixels = new Uint8ClampedArray(dims.width * dims.height * 4);

    // Restore state from URL params
    const params = getURLParams();
    const selector = document.getElementById("demoSelector");
    if (params.has("demo")) {
      selector.value = params.get("demo");
    }
    syncControlVisibility(selector.value);
    restoreDemoParams(selector.value, params);

    // Initial render
    renderSelectedDemo();

    // --- Event listeners ---

    selector.addEventListener("change", () => {
      clearDemoParams();
      updateURL({ demo: selector.value });
      syncControlVisibility(selector.value);
      renderSelectedDemo();
    });

    document
      .getElementById("renderBtn")
      .addEventListener("click", renderSelectedDemo);

    // aa controls
    document.getElementById("zoomSlider").addEventListener("input", () => {
      const val = parseFloat(document.getElementById("zoomSlider").value);
      document.getElementById("zoomValue").textContent = val;
      setAAZoom(val);
      persistDemoParams("aa");
      renderSelectedDemo();
    });

    // conv_dash_marker controls
    document.getElementById("dashWidthSlider").addEventListener("input", () => {
      const val = parseFloat(document.getElementById("dashWidthSlider").value);
      document.getElementById("dashWidthValue").textContent = val;
      setDashWidth(val);
      persistDemoParams("conv_dash_marker");
      renderSelectedDemo();
    });
    document
      .getElementById("dashSmoothSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("dashSmoothSlider").value,
        );
        document.getElementById("dashSmoothValue").textContent = val.toFixed(2);
        setDashSmooth(val);
        persistDemoParams("conv_dash_marker");
        renderSelectedDemo();
      });
    document
      .getElementById("dashCapSelector")
      .addEventListener("change", () => {
        setDashCap(parseInt(document.getElementById("dashCapSelector").value));
        persistDemoParams("conv_dash_marker");
        renderSelectedDemo();
      });
    document.getElementById("dashClosed").addEventListener("change", () => {
      setDashClosed(document.getElementById("dashClosed").checked);
      persistDemoParams("conv_dash_marker");
      renderSelectedDemo();
    });
    document.getElementById("dashEvenOdd").addEventListener("change", () => {
      setDashEvenOdd(document.getElementById("dashEvenOdd").checked);
      persistDemoParams("conv_dash_marker");
      renderSelectedDemo();
    });

    // gouraud controls
    document.getElementById("dilationSlider").addEventListener("input", () => {
      const val = parseFloat(document.getElementById("dilationSlider").value);
      document.getElementById("dilationValue").textContent = val;
      setGouraudDilation(val);
      persistDemoParams("gouraud");
      renderSelectedDemo();
    });

    // imagefilters controls
    const filterSelector = document.getElementById("filterSelector");
    filterSelector.addEventListener("change", () => {
      const val = parseInt(filterSelector.value);
      setImageFilter(val);
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
      persistDemoParams("imagefilters");
      renderSelectedDemo();
    });
    document
      .getElementById("filterRadiusSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("filterRadiusSlider").value,
        );
        document.getElementById("filterRadiusValue").textContent = val;
        setImageFilterRadius(parseInt(filterSelector.value), val);
        persistDemoParams("imagefilters");
        renderSelectedDemo();
      });

    // sbool controls
    document
      .getElementById("sboolOpSelector")
      .addEventListener("change", () => {
        setSBoolOp(parseInt(document.getElementById("sboolOpSelector").value));
        persistDemoParams("sbool");
        renderSelectedDemo();
      });

    // convstroke controls
    document
      .getElementById("strokeJoinSelector")
      .addEventListener("change", () => {
        setStrokeJoin(
          parseInt(document.getElementById("strokeJoinSelector").value),
        );
        persistDemoParams("convstroke");
        renderSelectedDemo();
      });
    document
      .getElementById("strokeCapSelector")
      .addEventListener("change", () => {
        setStrokeCap(
          parseInt(document.getElementById("strokeCapSelector").value),
        );
        persistDemoParams("convstroke");
        renderSelectedDemo();
      });
    document
      .getElementById("strokeWidthSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("strokeWidthSlider").value,
        );
        document.getElementById("strokeWidthValue").textContent = val;
        setStrokeWidth(val);
        persistDemoParams("convstroke");
        renderSelectedDemo();
      });
    document
      .getElementById("strokeMiterSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("strokeMiterSlider").value,
        );
        document.getElementById("strokeMiterValue").textContent = val;
        setStrokeMiterLimit(val);
        persistDemoParams("convstroke");
        renderSelectedDemo();
      });

    // convcontour controls
    document
      .getElementById("contourWidthSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("contourWidthSlider").value,
        );
        document.getElementById("contourWidthValue").textContent = val;
        setContourWidth(val);
        persistDemoParams("convcontour");
        renderSelectedDemo();
      });
    document
      .getElementById("contourCloseModeSelector")
      .addEventListener("change", () => {
        setContourCloseMode(
          parseInt(document.getElementById("contourCloseModeSelector").value),
        );
        persistDemoParams("convcontour");
        renderSelectedDemo();
      });
    document
      .getElementById("contourAutoDetect")
      .addEventListener("change", () => {
        setContourAutoDetect(
          document.getElementById("contourAutoDetect").checked,
        );
        persistDemoParams("convcontour");
        renderSelectedDemo();
      });

    // gamma controls
    document.getElementById("gammaSlider").addEventListener("input", () => {
      const val = parseFloat(document.getElementById("gammaSlider").value);
      document.getElementById("gammaValue").textContent = val.toFixed(2);
      setGammaValue(val);
      renderSelectedDemo();
    });
    document
      .getElementById("gammaThickSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("gammaThickSlider").value,
        );
        document.getElementById("gammaThickValue").textContent = val.toFixed(1);
        setGammaThickness(val);
        renderSelectedDemo();
      });
    document
      .getElementById("gammaContrastSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("gammaContrastSlider").value,
        );
        document.getElementById("gammaContrastValue").textContent =
          val.toFixed(2);
        setGammaContrast(val);
        renderSelectedDemo();
      });

    // lionoutline width slider
    document
      .getElementById("lionOutlineWidthSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("lionOutlineWidthSlider").value,
        );
        document.getElementById("lionOutlineWidthValue").textContent =
          val.toFixed(1);
        setLionOutlineWidth(val);
        persistDemoParams("lionoutline");
        renderSelectedDemo();
      });

    // lion_lens controls
    document
      .getElementById("lionLensScaleSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("lionLensScaleSlider").value,
        );
        document.getElementById("lionLensScaleValue").textContent =
          val.toFixed(2);
        setLionLensScale(val);
        renderSelectedDemo();
      });
    document
      .getElementById("lionLensRadiusSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("lionLensRadiusSlider").value,
        );
        document.getElementById("lionLensRadiusValue").textContent = val;
        setLionLensRadius(val);
        renderSelectedDemo();
      });

    // roundedrect controls
    document.getElementById("rrRadiusSlider").addEventListener("input", () => {
      const val = parseFloat(document.getElementById("rrRadiusSlider").value);
      document.getElementById("rrRadiusValue").textContent = val.toFixed(1);
      setRRRadius(val);
      persistDemoParams("roundedrect");
      renderSelectedDemo();
    });
    document.getElementById("rrOffsetSlider").addEventListener("input", () => {
      const val = parseFloat(document.getElementById("rrOffsetSlider").value);
      document.getElementById("rrOffsetValue").textContent = val.toFixed(1);
      setRROffset(val);
      persistDemoParams("roundedrect");
      renderSelectedDemo();
    });
    document.getElementById("rrDarkBg").addEventListener("change", () => {
      setRRDarkBg(document.getElementById("rrDarkBg").checked);
      persistDemoParams("roundedrect");
      renderSelectedDemo();
    });

    // component controls
    document.getElementById("compAlphaSlider").addEventListener("input", () => {
      const val = parseInt(document.getElementById("compAlphaSlider").value);
      document.getElementById("compAlphaValue").textContent = val;
      setCompAlpha(val);
      persistDemoParams("component");
      renderSelectedDemo();
    });

    // perspective controls
    document
      .getElementById("perspectiveTypeSelector")
      .addEventListener("change", () => {
        setPerspectiveType(
          parseInt(document.getElementById("perspectiveTypeSelector").value),
        );
        persistDemoParams("perspective");
        renderSelectedDemo();
      });

    // trans_curve controls
    document
      .getElementById("transCurveAnimate")
      .addEventListener("change", () => {
        toggleTransCurveAnimate();
        renderSelectedDemo();
      });

    // trans_curve2 controls
    document
      .getElementById("transCurve2Animate")
      .addEventListener("change", () => {
        toggleTransCurve2Animate();
        renderSelectedDemo();
      });

    // blur controls
    document
      .getElementById("blurRadiusSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("blurRadiusSlider").value,
        );
        document.getElementById("blurRadiusValue").textContent = val;
        setBlurRadius(val);
        renderSelectedDemo();
      });
    document
      .getElementById("blurMethodSelector")
      .addEventListener("change", () => {
        setBlurMethod(
          parseInt(document.getElementById("blurMethodSelector").value),
        );
        renderSelectedDemo();
      });

    // circles controls
    document
      .getElementById("circlesSelectivitySlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("circlesSelectivitySlider").value,
        );
        document.getElementById("circlesSelectivityValue").textContent =
          val.toFixed(2);
        setCirclesSelectivity(val);
        renderSelectedDemo();
      });
    document
      .getElementById("circlesSizeSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("circlesSizeSlider").value,
        );
        document.getElementById("circlesSizeValue").textContent =
          val.toFixed(1);
        setCirclesSize(val);
        renderSelectedDemo();
      });

    document
      .getElementById("circlesZLowSlider")
      .addEventListener("input", () => {
        const low = parseFloat(document.getElementById("circlesZLowSlider").value);
        let high = parseFloat(document.getElementById("circlesZHighSlider").value);
        if (low > high) {
          high = low;
          document.getElementById("circlesZHighSlider").value = high;
          document.getElementById("circlesZHighValue").textContent = high.toFixed(2);
        }
        document.getElementById("circlesZLowValue").textContent = low.toFixed(2);
        setCirclesZRange(low, high);
        renderSelectedDemo();
      });
    document
      .getElementById("circlesZHighSlider")
      .addEventListener("input", () => {
        let low = parseFloat(document.getElementById("circlesZLowSlider").value);
        const high = parseFloat(document.getElementById("circlesZHighSlider").value);
        if (high < low) {
          low = high;
          document.getElementById("circlesZLowSlider").value = low;
          document.getElementById("circlesZLowValue").textContent = low.toFixed(2);
        }
        document.getElementById("circlesZHighValue").textContent = high.toFixed(2);
        setCirclesZRange(low, high);
        renderSelectedDemo();
      });

    // gouraud_mesh controls
    document
      .getElementById("meshColsSlider")
      .addEventListener("input", () => {
        const cols = parseInt(document.getElementById("meshColsSlider").value);
        const rows = parseInt(document.getElementById("meshRowsSlider").value);
        document.getElementById("meshColsValue").textContent = cols;
        setMeshSize(cols, rows);
      });
    document
      .getElementById("meshRowsSlider")
      .addEventListener("input", () => {
        const cols = parseInt(document.getElementById("meshColsSlider").value);
        const rows = parseInt(document.getElementById("meshRowsSlider").value);
        document.getElementById("meshRowsValue").textContent = rows;
        setMeshSize(cols, rows);
      });

    // compositing controls
    document.getElementById("compOpSelector").addEventListener("change", () => {
      setCompOp(parseInt(document.getElementById("compOpSelector").value));
      persistDemoParams("compositing");
      renderSelectedDemo();
    });
    document
      .getElementById("compAlphaSrcSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("compAlphaSrcSlider").value,
        );
        document.getElementById("compAlphaSrcValue").textContent =
          val.toFixed(2);
        setCompAlphaSrc(val);
        persistDemoParams("compositing");
        renderSelectedDemo();
      });
    document
      .getElementById("compAlphaDstSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("compAlphaDstSlider").value,
        );
        document.getElementById("compAlphaDstValue").textContent =
          val.toFixed(2);
        setCompAlphaDst(val);
        persistDemoParams("compositing");
        renderSelectedDemo();
      });

    // multi_clip controls
    document
      .getElementById("multiClipNSlider")
      .addEventListener("input", () => {
        const val = parseFloat(
          document.getElementById("multiClipNSlider").value,
        );
        document.getElementById("multiClipNValue").textContent = val;
        setMultiClipN(val);
        persistDemoParams("multi_clip");
        renderSelectedDemo();
      });

    // alpha_mask2 controls
    document
      .getElementById("am2EllipsesSlider")
      .addEventListener("input", () => {
        const val = parseInt(document.getElementById("am2EllipsesSlider").value);
        document.getElementById("am2EllipsesValue").textContent = val;
        setAlphaMask2NumEllipses(val);
        renderSelectedDemo();
      });

    // Mouse events for draggable-point demos
    let isDragging = false;

    canvas.addEventListener("contextmenu", (e) => e.preventDefault());

    canvas.addEventListener("mousedown", (e) => {
      const rect = canvas.getBoundingClientRect();
      const x = (e.clientX - rect.left) * (canvas.width / rect.width);
      const y = (e.clientY - rect.top) * (canvas.height / rect.height);
      const right = e.button === 2;
      if (onMouseDown(selector.value, x, y, right)) {
        isDragging = true;
        renderSelectedDemo();
      }
    });

    window.addEventListener("mousemove", (e) => {
      if (!isDragging) return;
      const rect = canvas.getBoundingClientRect();
      const x = (e.clientX - rect.left) * (canvas.width / rect.width);
      const y = (e.clientY - rect.top) * (canvas.height / rect.height);
      const right = (e.buttons & 2) !== 0;
      if (onMouseMove(selector.value, x, y, right)) {
        renderSelectedDemo();
      }
    });

    window.addEventListener("mouseup", () => {
      if (!isDragging) return;
      isDragging = false;
      onMouseUp(selector.value);
      renderSelectedDemo();
      // Persist node positions after drag for all interactive demos
      persistDemoParams(selector.value);
    });

    // Animation loop for demos that need it (gouraud_mesh, trans_curve)
    function animate() {
      const demoType = selector.value;
      if (
        demoType === "gouraud_mesh" ||
        (demoType === "trans_curve" &&
          document.getElementById("transCurveAnimate").checked) ||
        (demoType === "trans_curve2" &&
          document.getElementById("transCurve2Animate").checked) ||
        demoType === "distortions"
      ) {
        renderSelectedDemo();
      }
      requestAnimationFrame(animate);
    }
    requestAnimationFrame(animate);

    updateStatus("Ready");
  } catch (err) {
    console.error("Failed to load WASM:", err);
    updateStatus("Error: " + err.message);
  }
}

function syncControlVisibility(demoType) {
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
  document.getElementById("sboolControls").style.display =
    demoType === "sbool" ? "flex" : "none";
  document.getElementById("convstrokeControls").style.display =
    demoType === "convstroke" ? "flex" : "none";
  document.getElementById("convcontourControls").style.display =
    demoType === "convcontour" ? "flex" : "none";
  document.getElementById("gammaControls").style.display =
    demoType === "gamma" ? "flex" : "none";
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
  document.getElementById("alphaMask2Controls").style.display =
    demoType === "alpha_mask2" ? "flex" : "none";
}

const demoDescriptions = {
  agg2d:
    "Port of the original agg2d_demo.cpp. Showcases the high-level Agg2D API: viewport mapping, aqua-style gradient buttons with rounded rectangles, filled ellipses, arc-based path construction, blend modes (Add, Overlay), and radial gradient fills.",
  lion: "The classic AGG signature demo. High-quality vector graphics consisting of hundreds of paths parsed from the original AGG lion data.",
  gradients:
    "Linear and radial gradient fills. Demonstrates the advanced span generation and multi-stop color interpolation.",
  aa: "Anti-aliasing showcase. Lines and circles drawn at sub-pixel offsets to demonstrate the precision and smoothness of AGG's rasterizer.",
  blend:
    "Compositing and blend modes. Showcases how different layers can be combined using standard and advanced blend modes like Multiply, Screen, and Overlay.",
  bspline:
    "B-Spline curve smoothing. Demonstrates the creation of smooth, continuous curves from a set of control points.",
  conv_dash_marker:
    "Port of AGG's conv_dash_marker demo. Applies conv_smooth_poly1 to soften corners, then conv_dash to create dash patterns, and conv_marker to place arrowheads at line endpoints. Adjust smoothness, stroke width, cap style, and fill rule. Drag the three control points to reshape the paths.",
  gouraud:
    "Smooth color interpolation across triangles. Demonstrates AGG's capability to render gradient-shaded meshes with sub-pixel precision and adjustable dilation.",
  imagefilters:
    "Comparison of different image interpolation filters. Rotates and scales a procedurally generated image using filters like Bilinear, Bicubic, Sinc, and Lanczos to showcase quality vs. performance.",
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
    "Animated image distortions. Applies Wave and Swirl effects to a procedurally generated image using custom coordinate distortion interpolators. Click and drag to move the distortion center.",
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
};

function renderSelectedDemo() {
  const selector = document.getElementById("demoSelector");
  const demoType = selector.value;

  updateStatus("Rendering " + demoType + "...");
  document.getElementById("demoDesc").textContent =
    demoDescriptions[demoType] || "";

  try {
    renderDemo(demoType, pixels);
    imageData.data.set(pixels);
    ctx.putImageData(imageData, 0, 0);
    updateStatus("Rendered " + demoType);
  } catch (err) {
    console.error("Render Error:", err);
    updateStatus("Render Error: " + err.message);
  }
}

function updateStatus(msg) {
  document.getElementById("statusMsg").textContent = msg;
}

// Start initialization
init();
