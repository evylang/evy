import hightlightEvy from "./module/highlight.js"

// --- Syntax coloring -----------------------------------------------
document.querySelectorAll(".language-evy").forEach((el) => {
  el.innerHTML = hightlightEvy(el.textContent)
})
