"use strict";

import { getURLParams } from "./url-state.js";
import { persistDemoParams, restoreDemoParams } from "./demo-state.js";
import { initTheme } from "./theme.js";
import { syncControlVisibility, demoDescriptions } from "./ui-sync.js";
import { setupEventHandlers } from "./event-handlers.js";

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

    // Setup all event listeners
    setupEventHandlers(canvas, selector, renderSelectedDemo, persistDemoParams);

    updateStatus("Ready");
  } catch (err) {
    console.error("Failed to load WASM:", err);
    updateStatus("Error: " + err.message);
  }
}

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
