import hightlightEvy from "./module/highlight.js"
import initThemeToggle from "./module/theme.js"

// --- ThemeToggle ----------------------------------------------------
initThemeToggle("#dark-theme", "theme")

const aboutDialog = document.querySelector("#dialog-about")
const showAboutDialog = document.querySelector("#show-dialog-about")
showAboutDialog.addEventListener("click", () => {
  aboutDialog.showModal()
})

// --- Syntax coloring -----------------------------------------------
document.querySelectorAll(".language-evy").forEach((el) => {
  el.innerHTML = hightlightEvy(el.textContent)
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

// Highlight current page (sub-)heading in sidebar.
highlightCurrent()
window.addEventListener("hashchange", highlightCurrent)

preventReloadOnSelfLink()

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
  scrollToHeading()
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

function scrollToHeading() {
  // this should be default browser behavior for onhashchange,
  // but in combination with these CSS rules it does not seem to work.
  //
  //    display: grid | flexbox;
  //    height: 100% | inhert;
  //    overflow: auto;
  const hash = window.location.hash
  if (hash === "" || hash === "#") {
    document.querySelector("body>main").scrollTo(0, 0)
    return
  }
  document.querySelector(hash).scrollIntoView(true)
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

function preventReloadOnSelfLink() {
  let href = normalizedHref()
  const last = href.split("/").pop()
  const nodes = document.querySelectorAll(`#sidebar a[href$="${last}"]`)
  const selfLink = [...nodes].find((n) => n.href === href)
  if (selfLink) {
    selfLink.onclick = (e) => {
      e.preventDefault()
      document.querySelector("body>main").scrollTo(0, 0)
      window.history.pushState("", "", window.location.pathname)
    }
  }
}
