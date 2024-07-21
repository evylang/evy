import hightlightEvy from "./module/highlight.js"
import initThemeToggle from "./module/theme.js"
import { querySPALinks, wireSPALinks } from "./module/spa.js"

// --- ThemeToggle ----------------------------------------------------
initThemeToggle("#dark-theme", "theme")

// --- Spa light: load HTML fragments on course link clicks ------------
const courseHref = document.querySelector("#sidebar h1 a").getAttribute("href")
const courseDir = courseHref.substring(0, courseHref.lastIndexOf("/") + 1)
const courseLinks = querySPALinks(courseDir)
const target = document.querySelector("main div.max-width-wrapper")
const scrollTop = document.querySelector("body>main")
wireSPALinks(courseLinks, target, scrollTop)

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

// -- Tick radio/checkbox on outer div click -------------------------
const checkables = [
  // radio and checkbox inputs
  ...document.querySelectorAll("div label input[type=checkbox],div label input[type=radio]"),
]
checkables.forEach((n) => {
  // For any click on surrounding div, update checkbox or radio button
  const div = n.parentElement.parentElement
  div.addEventListener("click", () => (n.checked = !n.checked))
})

// --- Sidebar -------------------------------------------------------
// Open and close sidebar on mobile on hamburger click.
const sidebar = document.querySelector("#sidebar")
document.querySelector("#hamburger").onclick = showSidebar
document.querySelector("#sidebar-close").onclick = hideSidebar
window.addEventListener("hashchange", hideSidebar)

// Expand sidebar subsection on chevron expander click.
const expanders = [...document.querySelectorAll("#sidebar .expander")]
expanders.map((el) => (el.onclick = expanderClick))

const sidebarLinks = [...document.querySelectorAll("#sidebar a")]
sidebarLinks.map((el) => el.addEventListener("click", (e) => highlightItem(el)))

// Highlight current page (sub-)heading in sidebar.
highlightCurrent()

// --- Utility functions ---------------------------------------------
function expanderClick(e) {
  const expander = e.target
  const siblingUL = expander.nextElementSibling
  expander.classList.toggle("show")
  siblingUL.classList.toggle("show")
}

function showSidebar() {
  sidebar.classList.add("show")
  document.addEventListener("click", handleOutsideSidebarClick)
}
function hideSidebar() {
  sidebar.classList.remove("show")
  document.removeEventListener("click", handleOutsideSidebarClick)
}
function handleOutsideSidebarClick(e) {
  if (sidebar.classList.contains("show") && e.pageX > sidebar.offsetWidth) {
    hideSidebar()
  }
}

function highlightCurrent() {
  const item = getCurrentItem()
  item && highlightItem(item)
}

function getCurrentItem() {
  // href is the page URL, replace trailing slash with /index.html
  const href = window.location.href.replace(/\/$/, "/index.html")
  const last = href.split("/").pop() // filename only
  // [href=...] could be relative, so find all matching links with same base
  // filename and then compare fully qualified n.href (as opposed to the
  // original n.getAttribute("href")) against the page href.
  const nodes = document.querySelectorAll(`#sidebar a[href$="${last}"]`)
  return [...nodes].find((n) => n.href === href)
}

function highlightItem(item) {
  // clear previous highlight
  const highlighted = document.querySelectorAll("#sidebar .highlight")
  highlighted.forEach((el) => el.classList.remove("highlight"))
  // add new highlight
  item.classList.add("highlight")
  // expand sidebar hierarchy to show highlighted item
  const els = findElementsToExpand(item)
  els.forEach((el) => el.classList.add("show")) // expand
  const last = els.pop()
  // highlight top level element
  last && last.previousElementSibling.classList.add("highlight-within")
}

// findElementsToExpand finds and returns an array of all <ul> and <div
// class="expander"> elements that need to be showed and expanded up the
// sidebar hierarchy tree for item to be visible. The <ul> adjecent to item
// is also included as clicking on an item expands its child <ul>.
function findElementsToExpand(item) {
  const parents = []
  if (item.nextElementSibling) {
    const siblingUL = item.nextElementSibling.nextElementSibling // expand next level down
    parents.push(siblingUL)
  }
  while (item) {
    if (item.tagName === "UL") {
      const expander = item.previousElementSibling // div with ".expander" class
      if (!expander.classList.contains("expander")) {
        break
      }
      parents.push(item)
      parents.push(expander)
    }
    item = item.parentElement // move upwards in the sidebar hierarchy
  }
  return parents
}
