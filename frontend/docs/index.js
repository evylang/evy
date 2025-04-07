import hightlightEvy from "./module/highlight.js"
import initThemeToggle from "./module/theme.js"
import Sidebar from "./module/sidebar.js"
import { querySPALinks, wireSPALinks } from "./module/spa.js"

// --- ThemeToggle ----------------------------------------------------
initThemeToggle("#dark-theme", "theme")

// --- Sidebar -------------------------------------------------------
const sidebar = new Sidebar("#sidebar")
document.querySelector("#hamburger").onclick = sidebar.show
document.querySelector("#sidebar-close").onclick = sidebar.hide
window.addEventListener("popstate", sidebar.highlightCurrent)

// --- Spa light: load HTML fragments on any docs link clicks ------------
const overviewHref = document.querySelector("#sidebar a").getAttribute("href")
const docsDir = overviewHref.substring(0, overviewHref.lastIndexOf("/") + 1)
const docsLinks = querySPALinks(docsDir)
const target = document.querySelector("main div.max-width-wrapper")
const scrollTop = document.querySelector("body>main")
wireSPALinks(docsLinks, target, scrollTop, afterNavigate)
highlight()

// --- Hide/show about dialog ----------------------------------------
const aboutDialog = document.querySelector("#dialog-about")
const showAboutDialog = document.querySelector("#show-dialog-about")
showAboutDialog.addEventListener("click", () => {
  aboutDialog.showModal()
})

// --- Utilities -----------------------------------------------
function afterNavigate() {
  sidebar.hide()
  highlight()
}

function highlight() {
  document.querySelectorAll(".language-evy").forEach((el) => {
    el.innerHTML = hightlightEvy(el.textContent)
  })
}
