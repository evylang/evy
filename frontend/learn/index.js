import hightlightEvy from "./module/highlight.js"

// --- Syntax coloring -----------------------------------------------
document.querySelectorAll(".language-evy").forEach((el) => {
  el.innerHTML = hightlightEvy(el.textContent)
})

// --- Hide/show about dialog ----------------------------------------
const aboutDialog = document.querySelector("#dialog-about")
const showAboutDialog = document.querySelector("#show-dialog-about")
showAboutDialog.addEventListener("click", () => {
  aboutDialog.showModal()
})
