"use strict";

const go = new Go();
let wasmInstance;
let canvas;
let ctx;
let imageData;
let pixels;

window.onerror = function(message, source, lineno, colno, error) {
  console.error("Global JS Error:", message, "at", source, ":", lineno);
  updateStatus("JS Error: " + message);
};

async function init() {
  console.log("Initializing AGG Go Web Demo...");
  try {
    const result = await WebAssembly.instantiateStreaming(
      fetch("main.wasm"),
      go.importObject,
    ).catch(err => {
        console.error("WASM Fetch/Instantiate Error:", err);
        throw err;
    });
    
    wasmInstance = result.instance;
    go.run(wasmInstance).catch(err => {
        console.error("WASM Runtime Error:", err);
        updateStatus("WASM Error: " + err.message);
    });

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

    // Initial render
    renderSelectedDemo();

    // Setup event listeners
    const selector = document.getElementById("demoSelector");
    selector.addEventListener("change", () => {
      const aaControls = document.getElementById("aaControls");
      const dashControls = document.getElementById("dashControls");
      const gouraudControls = document.getElementById("gouraudControls");
      aaControls.style.display = selector.value === "aa" ? "flex" : "none";
      dashControls.style.display = selector.value === "dash" ? "flex" : "none";
      gouraudControls.style.display = selector.value === "gouraud" ? "flex" : "none";
      renderSelectedDemo();
    });

    document
      .getElementById("renderBtn")
      .addEventListener("click", renderSelectedDemo);

    const zoomSlider = document.getElementById("zoomSlider");
    zoomSlider.addEventListener("input", () => {
      const val = parseFloat(zoomSlider.value);
      document.getElementById("zoomValue").textContent = val;
      setAAZoom(val);
      renderSelectedDemo();
    });

    const dilationSlider = document.getElementById("dilationSlider");
    dilationSlider.addEventListener("input", () => {
      const val = parseFloat(dilationSlider.value);
      document.getElementById("dilationValue").textContent = val;
      setGouraudDilation(val);
      renderSelectedDemo();
    });

    const dashWidthSlider = document.getElementById("dashWidthSlider");
    dashWidthSlider.addEventListener("input", () => {
      const val = parseFloat(dashWidthSlider.value);
      document.getElementById("dashWidthValue").textContent = val;
      setDashWidth(val);
      renderSelectedDemo();
    });

    const dashClosedBox = document.getElementById("dashClosed");
    dashClosedBox.addEventListener("change", () => {
      setDashClosed(dashClosedBox.checked);
      renderSelectedDemo();
    });

    // Mouse events for interactivity
    let isDragging = false;

    canvas.addEventListener("mousedown", (e) => {
      const rect = canvas.getBoundingClientRect();
      const x = (e.clientX - rect.left) * (canvas.width / rect.width);
      const y = (e.clientY - rect.top) * (canvas.height / rect.height);

      if (onMouseDown(selector.value, x, y)) {
        isDragging = true;
        renderSelectedDemo();
      }
    });

    window.addEventListener("mousemove", (e) => {
      if (!isDragging) return;

      const rect = canvas.getBoundingClientRect();
      const x = (e.clientX - rect.left) * (canvas.width / rect.width);
      const y = (e.clientY - rect.top) * (canvas.height / rect.height);

      if (onMouseMove(selector.value, x, y)) {
        renderSelectedDemo();
      }
    });

    window.addEventListener("mouseup", () => {
      if (!isDragging) return;
      isDragging = false;
      onMouseUp(selector.value);
      renderSelectedDemo();
    });

    updateStatus("Ready");
  } catch (err) {
    console.error("Failed to load WASM:", err);
    updateStatus("Error: " + err.message);
  }
}

const demoDescriptions = {
  lines:
    "Basic line drawing with different thicknesses. Showcases the core rendering pipeline and anti-aliased lines.",
  circles:
    "Simple concentric circles. Demonstrates basic shape primitive rendering.",
  starburst:
    "A collection of lines radiating from a center point. Showcases line rendering at various angles.",
  rects:
    "Filled and stroked rectangles, including rounded rectangles. Demonstrates alpha blending and semi-transparent fills.",
  lion: "The classic AGG signature demo. High-quality vector graphics consisting of hundreds of paths parsed from the original AGG lion data.",
  gradients:
    "Linear and radial gradient fills. Demonstrates the advanced span generation and multi-stop color interpolation.",
  aa: "Anti-aliasing showcase. Lines and circles drawn at sub-pixel offsets to demonstrate the precision and smoothness of AGG's rasterizer.",
  blend:
    "Compositing and blend modes. Showcases how different layers can be combined using standard and advanced blend modes like Multiply, Screen, and Overlay.",
  bspline:
    "B-Spline curve smoothing. Demonstrates the creation of smooth, continuous curves from a set of control points.",
  "dash": "Advanced line styling. Showcases various dash patterns and line thicknesses applied to both simple lines and complex paths.",
  "gouraud": "Smooth color interpolation across triangles. Demonstrates AGG's capability to render gradient-shaded meshes with sub-pixel precision and adjustable dilation.",
  "aatest": "Comprehensive anti-aliasing precision test. Renders radial lines, various ellipse sizes, and gradient-filled triangles at fractional offsets to verify the rasterizer's quality."
  };


function renderSelectedDemo() {
  const selector = document.getElementById("demoSelector");
  const demoType = selector.value;

  updateStatus("Rendering " + demoType + "...");
  document.getElementById("demoDesc").textContent =
    demoDescriptions[demoType] || "";

  try {
    // Perform rendering in Go and copy pixels back to JS
    renderDemo(demoType, pixels);

    // Update canvas
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
