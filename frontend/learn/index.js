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

// --- Spa light: load HTML fragments on course link clicks ------------
const courseHref = document.querySelector("#sidebar h1 a").getAttribute("href")
const courseDir = courseHref.substring(0, courseHref.lastIndexOf("/") + 1)
const courseLinks = querySPALinks(courseDir)
const target = document.querySelector("main div.max-width-wrapper")
const scrollTop = document.querySelector("body>main")
wireSPALinks(courseLinks, target, scrollTop, sidebar.hide)

// --- Syntax coloring ------------------------------------------------
document.querySelectorAll(".language-evy").forEach((el) => {
  el.innerHTML = hightlightEvy(el.textContent)
})

// --- Hide/show login dialog ----------------------------------------
const loginDialog = document.querySelector("#dialog-login")
const showLoginDialog = document.querySelector("#show-dialog-login")
showLoginDialog.addEventListener("click", () => {
  loginDialog.showModal()
})

// --- Tick radio/checkbox on outer div click -------------------------
const checkables = [
  // radio and checkbox inputs
  ...document.querySelectorAll("div label input[type=checkbox],div label input[type=radio]"),
]
checkables.forEach((n) => {
  // For any click on surrounding div, update checkbox or radio button
  const div = n.parentElement.parentElement
  div.addEventListener("click", () => (n.checked = !n.checked))
})
