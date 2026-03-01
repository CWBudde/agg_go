"use strict";

const go = new Go();
let wasmInstance;
let canvas;
let ctx;
let imageData;
let pixels;

async function init() {
    try {
        const result = await WebAssembly.instantiateStreaming(
            fetch("main.wasm"),
            go.importObject
        );
        wasmInstance = result.instance;
        go.run(wasmInstance);

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
        document.getElementById("demoSelector").addEventListener("change", renderSelectedDemo);
        document.getElementById("renderBtn").addEventListener("click", renderSelectedDemo);
        
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
    
    // Perform rendering in Go and copy pixels back to JS
    renderDemo(demoType, pixels);
    
    // Update canvas
    imageData.data.set(pixels);
    ctx.putImageData(imageData, 0, 0);
    
    updateStatus("Rendered " + demoType);
}

function updateStatus(msg) {
    document.getElementById("statusMsg").textContent = msg;
}

// Start initialization
init();
