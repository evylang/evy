import initThemeToggle from "./module/theme.js"

initThemeToggle("#dark-theme", "theme")

// Open and close sidebar on mobile on hamburger click.
const sidebar = document.querySelector("#sidebar")
document.querySelector("#hamburger").onclick = showSidebar
document.querySelector("#sidebar-close").onclick = hideSidebar
window.addEventListener("hashchange", hideSidebar)

// Expand sidebar subsection on chevron expander click.
const expanders = [...document.querySelectorAll("#sidebar .expander")]
expanders.map((el) => (el.onclick = expanderClick))

// Highlight current page (sub-)heading in sidebar.
highlightCurrent()
window.addEventListener("hashchange", highlightCurrent)

// Utilities
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

function expanderClick(e) {
  e.target.nextElementSibling.classList.toggle("show")
  e.target.classList.toggle("show")
}

function highlightCurrent() {
  const highlighted = document.querySelectorAll("#sidebar .highlight")
  highlighted.forEach((el) => el.classList.remove("highlight"))

  const item = getCurrentItem()
  if (!item) {
    return
  }
  item.classList.add("highlight")
  const els = getShowing(item)
  els.forEach((el) => el.classList.add("show"))
  const last = els.pop()
  last && last.previousElementSibling.classList.add("highlight-within")
}

function getCurrentItem() {
  const href = normalizedHref()
  const last = href.split("/").pop()
  const nodes = document.querySelectorAll(`#sidebar a[href$="${last}"]`)
  return [...nodes].find((n) => n.href === href)
}

function normalizedHref() {
  let href = window.location.href
  let hash = window.location.hash
  if (hash) {
    href = href.replace(hash, "")
  }
  if (href.endsWith("/")) {
    href = href + "index.html"
  }
  return href + hash
}

function getShowing(item) {
  const parents = []
  let n = item
  if (n.nextElementSibling) {
    n = n.nextElementSibling.nextElementSibling // sibling UL
  }
  while (n) {
    if (n.tagName === "UL") {
      const expander = n.previousElementSibling // div with ".expander" class
      if (!expander.classList.contains("expander")) {
        break
      }
      parents.push(n)
      parents.push(expander)
    }
    n = n.parentElement
  }
  return parents
}
