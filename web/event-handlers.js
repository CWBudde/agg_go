// --- Event handlers for all demo controls ---

import { syncControlVisibility } from "./ui-sync.js";
import { clearDemoParams, updateURL } from "./url-state.js";

export function setupEventHandlers(
  canvas,
  selector,
  renderSelectedDemo,
  persistDemoParams,
) {
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

  // line_thickness controls
  document
    .getElementById("lineThicknessFactorSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("lineThicknessFactorSlider").value,
      );
      document.getElementById("lineThicknessFactorValue").textContent =
        val.toFixed(2);
      setLineThicknessFactor(val);
      persistDemoParams("line_thickness");
      renderSelectedDemo();
    });
  document
    .getElementById("lineThicknessBlurSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("lineThicknessBlurSlider").value,
      );
      document.getElementById("lineThicknessBlurValue").textContent =
        val.toFixed(2);
      setLineThicknessBlur(val);
      persistDemoParams("line_thickness");
      renderSelectedDemo();
    });
  document.getElementById("lineThicknessMono").addEventListener("change", () => {
    setLineThicknessMono(document.getElementById("lineThicknessMono").checked);
    persistDemoParams("line_thickness");
    renderSelectedDemo();
  });
  document
    .getElementById("lineThicknessInvert")
    .addEventListener("change", () => {
      setLineThicknessInvert(
        document.getElementById("lineThicknessInvert").checked,
      );
      persistDemoParams("line_thickness");
      renderSelectedDemo();
    });

  // line_patterns controls
  document
    .getElementById("linePatternsScaleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("linePatternsScaleSlider").value,
      );
      document.getElementById("linePatternsScaleValue").textContent =
        val.toFixed(2);
      setLinePatternScaleX(val);
      persistDemoParams("line_patterns");
      renderSelectedDemo();
    });
  document
    .getElementById("linePatternsStartSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("linePatternsStartSlider").value,
      );
      document.getElementById("linePatternsStartValue").textContent =
        val.toFixed(2);
      setLinePatternStartX(val);
      persistDemoParams("line_patterns");
      renderSelectedDemo();
    });

  // line_patterns_clip controls
  document
    .getElementById("linePatternsClipScaleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("linePatternsClipScaleSlider").value,
      );
      document.getElementById("linePatternsClipScaleValue").textContent =
        val.toFixed(2);
      setLinePatternClipScaleX(val);
      persistDemoParams("line_patterns_clip");
      renderSelectedDemo();
    });
  document
    .getElementById("linePatternsClipStartSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("linePatternsClipStartSlider").value,
      );
      document.getElementById("linePatternsClipStartValue").textContent =
        val.toFixed(2);
      setLinePatternClipStartX(val);
      persistDemoParams("line_patterns_clip");
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
  document.getElementById("dashSmoothSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("dashSmoothSlider").value);
    document.getElementById("dashSmoothValue").textContent = val.toFixed(2);
    setDashSmooth(val);
    persistDemoParams("conv_dash_marker");
    renderSelectedDemo();
  });
  document.getElementById("dashCapSelector").addEventListener("change", () => {
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

  // image_fltr_graph controls
  const buildIfgMask = () => {
    let mask = 0;
    document.querySelectorAll(".ifg-filter").forEach((el) => {
      const idx = parseInt(el.dataset.index);
      if (el.checked) {
        mask |= 1 << idx;
      }
    });
    return mask >>> 0;
  };
  document.getElementById("ifgRadiusSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("ifgRadiusSlider").value);
    document.getElementById("ifgRadiusValue").textContent = val.toFixed(1);
    setImageFltrGraphRadius(val);
    persistDemoParams("image_fltr_graph");
    renderSelectedDemo();
  });
  document.querySelectorAll(".ifg-filter").forEach((el) => {
    el.addEventListener("change", () => {
      setImageFltrGraphMask(buildIfgMask());
      persistDemoParams("image_fltr_graph");
      renderSelectedDemo();
    });
  });

  // image_filters2 controls
  const if2Selector = document.getElementById("if2FilterSelector");
  const syncIf2RadiusVisibility = () => {
    const hasRadius = parseInt(if2Selector.value) >= 14;
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
  if2Selector.addEventListener("change", () => {
    const val = parseInt(if2Selector.value);
    setImageFilters2Filter(val);
    syncIf2RadiusVisibility();
    persistDemoParams("image_filters2");
    renderSelectedDemo();
  });
  document.getElementById("if2RadiusSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("if2RadiusSlider").value);
    document.getElementById("if2RadiusValue").textContent = val.toFixed(1);
    setImageFilters2Radius(val);
    persistDemoParams("image_filters2");
    renderSelectedDemo();
  });
  document.getElementById("if2Normalize").addEventListener("change", () => {
    setImageFilters2Normalize(document.getElementById("if2Normalize").checked);
    persistDemoParams("image_filters2");
    renderSelectedDemo();
  });
  syncIf2RadiusVisibility();

  // idea controls
  document.getElementById("ideaRotate").addEventListener("change", () => {
    setIdeaRotate(document.getElementById("ideaRotate").checked);
    persistDemoParams("idea");
    renderSelectedDemo();
  });
  document.getElementById("ideaEvenOdd").addEventListener("change", () => {
    setIdeaEvenOdd(document.getElementById("ideaEvenOdd").checked);
    persistDemoParams("idea");
    renderSelectedDemo();
  });
  document.getElementById("ideaDraft").addEventListener("change", () => {
    setIdeaDraft(document.getElementById("ideaDraft").checked);
    persistDemoParams("idea");
    renderSelectedDemo();
  });
  document.getElementById("ideaRoundoff").addEventListener("change", () => {
    setIdeaRoundoff(document.getElementById("ideaRoundoff").checked);
    persistDemoParams("idea");
    renderSelectedDemo();
  });
  document.getElementById("ideaStepSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("ideaStepSlider").value);
    document.getElementById("ideaStepValue").textContent = val.toFixed(3);
    setIdeaAngleDelta(val);
    persistDemoParams("idea");
    renderSelectedDemo();
  });

  // mol_view controls
  document
    .getElementById("molViewMoleculeSlider")
    .addEventListener("input", () => {
      const val = parseInt(
        document.getElementById("molViewMoleculeSlider").value,
      );
      document.getElementById("molViewMoleculeValue").textContent = val;
      setMolViewMolecule(val);
      persistDemoParams("mol_view");
      renderSelectedDemo();
    });
  document
    .getElementById("molViewThicknessSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("molViewThicknessSlider").value,
      );
      document.getElementById("molViewThicknessValue").textContent =
        val.toFixed(1);
      setMolViewThickness(val);
      persistDemoParams("mol_view");
      renderSelectedDemo();
    });
  document
    .getElementById("molViewTextSizeSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("molViewTextSizeSlider").value,
      );
      document.getElementById("molViewTextSizeValue").textContent =
        val.toFixed(1);
      setMolViewTextSize(val);
      persistDemoParams("mol_view");
      renderSelectedDemo();
    });
  document
    .getElementById("molViewAutoRotate")
    .addEventListener("change", () => {
      setMolViewAutoRotate(
        document.getElementById("molViewAutoRotate").checked,
      );
      persistDemoParams("mol_view");
      renderSelectedDemo();
    });

  // sbool controls
  document.getElementById("sboolOpSelector").addEventListener("change", () => {
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
  document.getElementById("strokeWidthSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("strokeWidthSlider").value);
    document.getElementById("strokeWidthValue").textContent = val;
    setStrokeWidth(val);
    persistDemoParams("convstroke");
    renderSelectedDemo();
  });
  document.getElementById("strokeMiterSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("strokeMiterSlider").value);
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
  document.getElementById("gammaThickSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gammaThickSlider").value);
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

  // lion alpha slider
  document.getElementById("lionAlphaSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("lionAlphaSlider").value);
    document.getElementById("lionAlphaValue").textContent = val.toFixed(2);
    setLionAlpha(val);
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
    .getElementById("transCurveNumPointsSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("transCurveNumPointsSlider").value,
      );
      document.getElementById("transCurveNumPointsValue").textContent =
        val.toFixed(0);
      setTransCurveNumPoints(val);
      persistDemoParams("trans_curve");
      renderSelectedDemo();
    });
  document.getElementById("transCurveClose").addEventListener("change", () => {
    setTransCurveClose(document.getElementById("transCurveClose").checked);
    persistDemoParams("trans_curve");
    renderSelectedDemo();
  });
  document
    .getElementById("transCurvePreserveXScale")
    .addEventListener("change", () => {
      setTransCurvePreserveXScale(
        document.getElementById("transCurvePreserveXScale").checked,
      );
      persistDemoParams("trans_curve");
      renderSelectedDemo();
    });
  document
    .getElementById("transCurveFixedLen")
    .addEventListener("change", () => {
      setTransCurveFixedLen(
        document.getElementById("transCurveFixedLen").checked,
      );
      persistDemoParams("trans_curve");
      renderSelectedDemo();
    });
  document
    .getElementById("transCurveAnimate")
    .addEventListener("change", () => {
      toggleTransCurveAnimate();
      persistDemoParams("trans_curve");
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
  document.getElementById("blurRadiusSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("blurRadiusSlider").value);
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
  document.getElementById("circlesSizeSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("circlesSizeSlider").value);
    document.getElementById("circlesSizeValue").textContent = val.toFixed(1);
    setCirclesSize(val);
    renderSelectedDemo();
  });

  document.getElementById("circlesZLowSlider").addEventListener("input", () => {
    const low = parseFloat(document.getElementById("circlesZLowSlider").value);
    let high = parseFloat(document.getElementById("circlesZHighSlider").value);
    if (low > high) {
      high = low;
      document.getElementById("circlesZHighSlider").value = high;
      document.getElementById("circlesZHighValue").textContent =
        high.toFixed(2);
    }
    document.getElementById("circlesZLowValue").textContent = low.toFixed(2);
    setCirclesZRange(low, high);
    renderSelectedDemo();
  });
  document
    .getElementById("circlesZHighSlider")
    .addEventListener("input", () => {
      let low = parseFloat(document.getElementById("circlesZLowSlider").value);
      const high = parseFloat(
        document.getElementById("circlesZHighSlider").value,
      );
      if (high < low) {
        low = high;
        document.getElementById("circlesZLowSlider").value = low;
        document.getElementById("circlesZLowValue").textContent =
          low.toFixed(2);
      }
      document.getElementById("circlesZHighValue").textContent =
        high.toFixed(2);
      setCirclesZRange(low, high);
      renderSelectedDemo();
    });

  // gouraud_mesh controls
  document.getElementById("meshColsSlider").addEventListener("input", () => {
    const cols = parseInt(document.getElementById("meshColsSlider").value);
    const rows = parseInt(document.getElementById("meshRowsSlider").value);
    document.getElementById("meshColsValue").textContent = cols;
    setMeshSize(cols, rows);
  });
  document.getElementById("meshRowsSlider").addEventListener("input", () => {
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
      document.getElementById("compAlphaSrcValue").textContent = val.toFixed(2);
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
      document.getElementById("compAlphaDstValue").textContent = val.toFixed(2);
      setCompAlphaDst(val);
      persistDemoParams("compositing");
      renderSelectedDemo();
    });

  // multi_clip controls
  document.getElementById("multiClipNSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("multiClipNSlider").value);
    document.getElementById("multiClipNValue").textContent = val;
    setMultiClipN(val);
    persistDemoParams("multi_clip");
    renderSelectedDemo();
  });

  // distortions controls
  document
    .getElementById("distortionsImageSelector")
    .addEventListener("change", () => {
      const val = parseInt(
        document.getElementById("distortionsImageSelector").value,
        10,
      );
      setDistortionsImage(val);
      persistDemoParams("distortions");
      renderSelectedDemo();
    });

  // alpha_mask2 controls
  document.getElementById("am2EllipsesSlider").addEventListener("input", () => {
    const val = parseInt(document.getElementById("am2EllipsesSlider").value);
    document.getElementById("am2EllipsesValue").textContent = val;
    setAlphaMask2NumEllipses(val);
    renderSelectedDemo();
  });

  // image1 controls
  document.getElementById("img1AngleSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("img1AngleSlider").value);
    document.getElementById("img1AngleValue").textContent = val + "°";
    setImg1Angle(val);
    renderSelectedDemo();
  });
  document.getElementById("img1ScaleSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("img1ScaleSlider").value);
    document.getElementById("img1ScaleValue").textContent = val.toFixed(2);
    setImg1Scale(val);
    renderSelectedDemo();
  });

  // image_transforms controls
  document
    .getElementById("imgTransExampleSelector")
    .addEventListener("change", () => {
      const val = parseInt(
        document.getElementById("imgTransExampleSelector").value,
      );
      setImgTransExample(val);
      renderSelectedDemo();
    });
  document
    .getElementById("imgTransPolyAngleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("imgTransPolyAngleSlider").value,
      );
      document.getElementById("imgTransPolyAngleValue").textContent = val + "°";
      setImgTransPolygonAngle(val);
      renderSelectedDemo();
    });
  document
    .getElementById("imgTransPolyScaleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("imgTransPolyScaleSlider").value,
      );
      document.getElementById("imgTransPolyScaleValue").textContent =
        val.toFixed(2);
      setImgTransPolygonScale(val);
      renderSelectedDemo();
    });
  document
    .getElementById("imgTransImgAngleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("imgTransImgAngleSlider").value,
      );
      document.getElementById("imgTransImgAngleValue").textContent = val + "°";
      setImgTransImageAngle(val);
      renderSelectedDemo();
    });
  document
    .getElementById("imgTransImgScaleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("imgTransImgScaleSlider").value,
      );
      document.getElementById("imgTransImgScaleValue").textContent =
        val.toFixed(2);
      setImgTransImageScale(val);
      renderSelectedDemo();
    });

  // pattern_fill controls
  document
    .getElementById("patFillPolyAngleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("patFillPolyAngleSlider").value,
      );
      document.getElementById("patFillPolyAngleValue").textContent = val + "°";
      setPatFillPolygonAngle(val);
      renderSelectedDemo();
    });
  document
    .getElementById("patFillPolyScaleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("patFillPolyScaleSlider").value,
      );
      document.getElementById("patFillPolyScaleValue").textContent =
        val.toFixed(2);
      setPatFillPolygonScale(val);
      renderSelectedDemo();
    });
  document
    .getElementById("patFillPatAngleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("patFillPatAngleSlider").value,
      );
      document.getElementById("patFillPatAngleValue").textContent = val + "°";
      setPatFillPatternAngle(val);
      renderSelectedDemo();
    });
  document
    .getElementById("patFillPatSizeSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("patFillPatSizeSlider").value,
      );
      document.getElementById("patFillPatSizeValue").textContent = val;
      setPatFillPatternSize(val);
      renderSelectedDemo();
    });

  // pattern_perspective controls
  document
    .getElementById("patternPerspectiveTypeSelector")
    .addEventListener("change", () => {
      const val = parseInt(
        document.getElementById("patternPerspectiveTypeSelector").value,
      );
      setPatternPerspectiveType(val);
      persistDemoParams("pattern_perspective");
      renderSelectedDemo();
    });

  // pattern_resample controls
  document
    .getElementById("patternResampleTypeSelector")
    .addEventListener("change", () => {
      const val = parseInt(
        document.getElementById("patternResampleTypeSelector").value,
      );
      setPatternResampleType(val);
      persistDemoParams("pattern_resample");
      renderSelectedDemo();
    });
  document
    .getElementById("patternResampleGammaSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("patternResampleGammaSlider").value,
      );
      document.getElementById("patternResampleGammaValue").textContent =
        val.toFixed(2);
      setPatternResampleGamma(val);
      persistDemoParams("pattern_resample");
      renderSelectedDemo();
    });
  document
    .getElementById("patternResampleBlurSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("patternResampleBlurSlider").value,
      );
      document.getElementById("patternResampleBlurValue").textContent =
        val.toFixed(2);
      setPatternResampleBlur(val);
      persistDemoParams("pattern_resample");
      renderSelectedDemo();
    });

  // gouraud opacity slider
  document
    .getElementById("gouraudOpacitySlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("gouraudOpacitySlider").value,
      );
      document.getElementById("gouraudOpacityValue").textContent =
        val.toFixed(2);
      setGouraudOpacity(val);
      persistDemoParams("gouraud");
      renderSelectedDemo();
    });

  // gamma_tuner controls
  document.getElementById("gtRSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gtRSlider").value);
    document.getElementById("gtRValue").textContent = val.toFixed(2);
    setGammaTunerR(val);
    renderSelectedDemo();
  });
  document.getElementById("gtGSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gtGSlider").value);
    document.getElementById("gtGValue").textContent = val.toFixed(2);
    setGammaTunerG(val);
    renderSelectedDemo();
  });
  document.getElementById("gtBSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gtBSlider").value);
    document.getElementById("gtBValue").textContent = val.toFixed(2);
    setGammaTunerB(val);
    renderSelectedDemo();
  });
  document.getElementById("gtGammaSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gtGammaSlider").value);
    document.getElementById("gtGammaValue").textContent = val.toFixed(2);
    setGammaTunerGamma(val);
    renderSelectedDemo();
  });
  document
    .getElementById("gtPatternSelector")
    .addEventListener("change", () => {
      setGammaTunerPattern(
        parseInt(document.getElementById("gtPatternSelector").value),
      );
      renderSelectedDemo();
    });

  // bezier_div controls
  document.getElementById("bdAngleTolSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("bdAngleTolSlider").value);
    document.getElementById("bdAngleTolValue").textContent = val + "°";
    setBDAngleTol(val);
    renderSelectedDemo();
  });
  document
    .getElementById("bdApproxScaleSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("bdApproxScaleSlider").value,
      );
      document.getElementById("bdApproxScaleValue").textContent =
        val.toFixed(3);
      setBDApproxScale(val);
      renderSelectedDemo();
    });
  document.getElementById("bdCuspLimitSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("bdCuspLimitSlider").value);
    document.getElementById("bdCuspLimitValue").textContent = val + "°";
    setBDCuspLimit(val);
    renderSelectedDemo();
  });
  document.getElementById("bdWidthSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("bdWidthSlider").value);
    document.getElementById("bdWidthValue").textContent = val.toFixed(2);
    setBDWidth(val);
    renderSelectedDemo();
  });
  document.getElementById("bdShowPoints").addEventListener("change", () => {
    setBDShowPoints(document.getElementById("bdShowPoints").checked);
    renderSelectedDemo();
  });
  document.getElementById("bdShowOutline").addEventListener("change", () => {
    setBDShowOutline(document.getElementById("bdShowOutline").checked);
    renderSelectedDemo();
  });
  document
    .getElementById("bdCurveTypeSelector")
    .addEventListener("change", () => {
      setBDCurveType(
        parseInt(document.getElementById("bdCurveTypeSelector").value),
      );
      renderSelectedDemo();
    });
  document
    .getElementById("bdCaseTypeSelector")
    .addEventListener("change", () => {
      const val = parseInt(document.getElementById("bdCaseTypeSelector").value);
      setBDCaseType(val);
      // Sync width slider with Go's updated value (Go may change width for certain cases)
      const newWidth = getBDWidth();
      document.getElementById("bdWidthSlider").value = newWidth;
      document.getElementById("bdWidthValue").textContent = newWidth.toFixed(2);
      renderSelectedDemo();
    });
  document
    .getElementById("bdInnerJoinSelector")
    .addEventListener("change", () => {
      setBDInnerJoin(
        parseInt(document.getElementById("bdInnerJoinSelector").value),
      );
      renderSelectedDemo();
    });
  document
    .getElementById("bdLineJoinSelector")
    .addEventListener("change", () => {
      setBDLineJoin(
        parseInt(document.getElementById("bdLineJoinSelector").value),
      );
      renderSelectedDemo();
    });
  document
    .getElementById("bdLineCapSelector")
    .addEventListener("change", () => {
      setBDLineCap(
        parseInt(document.getElementById("bdLineCapSelector").value),
      );
      renderSelectedDemo();
    });

  // rasterizers controls
  document.getElementById("rastGammaSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("rastGammaSlider").value);
    document.getElementById("rastGammaValue").textContent = val.toFixed(2);
    setRasterizersGamma(val);
    renderSelectedDemo();
  });
  document.getElementById("rastAlphaSlider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("rastAlphaSlider").value);
    document.getElementById("rastAlphaValue").textContent = val.toFixed(2);
    setRasterizersAlpha(val);
    renderSelectedDemo();
  });

  // bspline controls
  document
    .getElementById("bsplineNumPointsSlider")
    .addEventListener("input", () => {
      const val = parseFloat(
        document.getElementById("bsplineNumPointsSlider").value,
      );
      document.getElementById("bsplineNumPointsValue").textContent = val;
      setBSplineNumPoints(val);
      renderSelectedDemo();
    });
  document.getElementById("bsplineClosed").addEventListener("change", () => {
    setBSplineClosed(document.getElementById("bsplineClosed").checked);
    renderSelectedDemo();
  });

  // gradients_contour controls
  document
    .getElementById("gcPolygonSelector")
    .addEventListener("change", () => {
      setGradientsContourPolygon(
        parseInt(document.getElementById("gcPolygonSelector").value, 10),
      );
      persistDemoParams("gradients_contour");
      renderSelectedDemo();
    });
  document
    .getElementById("gcGradientSelector")
    .addEventListener("change", () => {
      setGradientsContourGradient(
        parseInt(document.getElementById("gcGradientSelector").value, 10),
      );
      persistDemoParams("gradients_contour");
      renderSelectedDemo();
    });
  document.getElementById("gcReflect").addEventListener("change", () => {
    setGradientsContourReflect(document.getElementById("gcReflect").checked);
    persistDemoParams("gradients_contour");
    renderSelectedDemo();
  });
  document.getElementById("gcColorsSlider").addEventListener("input", () => {
    const val = parseInt(document.getElementById("gcColorsSlider").value, 10);
    document.getElementById("gcColorsValue").textContent = String(val);
    setGradientsContourColors(val);
    persistDemoParams("gradients_contour");
    renderSelectedDemo();
  });
  document.getElementById("gcC1Slider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gcC1Slider").value);
    document.getElementById("gcC1Value").textContent = String(val);
    setGradientsContourC1(val);
    persistDemoParams("gradients_contour");
    renderSelectedDemo();
  });
  document.getElementById("gcC2Slider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gcC2Slider").value);
    document.getElementById("gcC2Value").textContent = String(val);
    setGradientsContourC2(val);
    persistDemoParams("gradients_contour");
    renderSelectedDemo();
  });
  document.getElementById("gcD1Slider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gcD1Slider").value);
    document.getElementById("gcD1Value").textContent = String(val);
    setGradientsContourD1(val);
    persistDemoParams("gradients_contour");
    renderSelectedDemo();
  });
  document.getElementById("gcD2Slider").addEventListener("input", () => {
    const val = parseFloat(document.getElementById("gcD2Slider").value);
    document.getElementById("gcD2Value").textContent = String(val);
    setGradientsContourD2(val);
    persistDemoParams("gradients_contour");
    renderSelectedDemo();
  });

  // flash_rasterizer2 controls
  document.getElementById("fr2ShapeSlider").addEventListener("input", () => {
    const val = parseInt(document.getElementById("fr2ShapeSlider").value, 10);
    document.getElementById("fr2ShapeValue").textContent = String(val);
    setFlash2ShapeIdx(val);
    persistDemoParams("flash_rasterizer2");
    renderSelectedDemo();
  });

  // Mouse wheel zoom for flash_rasterizer2
  canvas.addEventListener(
    "wheel",
    (e) => {
      if (selector.value !== "flash_rasterizer2") return;
      e.preventDefault();
      const rect = canvas.getBoundingClientRect();
      const mx = (e.clientX - rect.left) * (canvas.width / rect.width);
      const my = (e.clientY - rect.top) * (canvas.height / rect.height);
      applyFlash2Wheel(mx, my, e.deltaY);
      renderSelectedDemo();
    },
    { passive: false },
  );

  // gpc_test controls
  document.getElementById("gpcSceneSelector").addEventListener("change", () => {
    const val = parseInt(document.getElementById("gpcSceneSelector").value, 10);
    setGPCTestScene(val);
    persistDemoParams("gpc_test");
    renderSelectedDemo();
  });
  document.getElementById("gpcOpSelector").addEventListener("change", () => {
    const val = parseInt(document.getElementById("gpcOpSelector").value, 10);
    setGPCTestOperation(val);
    persistDemoParams("gpc_test");
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
      (demoType === "idea" && document.getElementById("ideaRotate").checked) ||
      (demoType === "mol_view" &&
        document.getElementById("molViewAutoRotate").checked) ||
      demoType === "distortions"
    ) {
      renderSelectedDemo();
    }
    requestAnimationFrame(animate);
  }
  requestAnimationFrame(animate);
}
