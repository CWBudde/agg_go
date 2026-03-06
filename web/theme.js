// --- Theme Management ---

export function initTheme() {
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
