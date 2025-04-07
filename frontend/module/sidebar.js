export default class Sidebar {
  constructor(selector) {
    this.root = document.querySelector(selector)
    if (!this.root) {
      throw new Error(`element with "${selector}" selector does not exist`)
    }
    this.#wireExpansion()
    this.#wireLinkClickHighlights()
    this.highlightCurrent()

    this.show = this.show.bind(this)
    this.hide = this.hide.bind(this)
    this.highlightCurrent = this.highlightCurrent.bind(this)
  }

  show() {
    this.root.classList.add("show")
    document.addEventListener("click", (e) => this.#handleOutsideSidebarClick(e))
  }
  hide() {
    this.root.classList.remove("show")
    document.removeEventListener("click", (e) => this.#handleOutsideSidebarClick(e))
  }
  highlightCurrent() {
    const item = this.#getCurrentItem()
    this.#highlightItem(item)
  }

  #wireExpansion() {
    const expanders = [...this.root.querySelectorAll(".expander")]
    expanders.map((el) => (el.onclick = expanderClick))
  }
  #wireLinkClickHighlights() {
    const links = [...this.root.querySelectorAll("a")]
    links.map((el) => el.addEventListener("click", (e) => this.#highlightItem(el)))
  }
  #highlightItem(item) {
    // clear previous highlight
    const highlighted = this.root.querySelectorAll(".highlight")
    highlighted.forEach((el) => el.classList.remove("highlight"))
    const highlightWithin = this.root.querySelectorAll(".highlight-within")
    highlightWithin.forEach((el) => el.classList.remove("highlight-within"))
    if (!item) {
      return
    }
    // add new highlight
    item.classList.add("highlight")
    // expand and highlight sidebar hierarchy containing item.
    const els = findElementsToExpand(item)
    els.forEach((el) => el.classList.add("show")) // expand
    els.forEach((el) => el.previousElementSibling.classList.add("highlight-within")) // expand
  }

  #getCurrentItem() {
    // href is the page URL, replace trailing slash with /index.html
    const href = window.location.href.replace(/\/$/, "/index.html")
    const last = href.split("/").pop() // filename only
    // [href=...] could be relative, so find all matching links with same base
    // filename and then compare fully qualified n.href (as opposed to the
    // original n.getAttribute("href")) against the page href.
    const nodes = this.root.querySelectorAll(`a[href$="${last}"]`)
    return [...nodes].find((n) => n.href === href)
  }
  #handleOutsideSidebarClick(e) {
    if (this.root.classList.contains("show") && e.pageX > sidebar.offsetWidth) {
      this.hide()
    }
  }
}

function expanderClick(e) {
  const expander = e.target
  const siblingUL = expander.nextElementSibling
  expander.classList.toggle("show")
  siblingUL.classList.toggle("show")
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
