// querySPALinks returns all relative links, and links within rootDir. Links
// containing `:`, e.g. https://..., http://..., mailto:.... are excluded.
export function querySPALinks(rootDir) {
  const internalLinks = document.querySelectorAll('a:not([href*=":"])')
  return [...internalLinks].filter((link) => {
    // use literal href="..." contents, not derived, absolute link.href
    const href = link.getAttribute("href")
    // relative link or course root
    return !href.startsWith("/") || href.startsWith(rootDir)
  })
}

// Load only a site fragment, set inner HTML of target element and update
// browser history and address bar  with a new URL.
//
// wireSPALinks intercepts clicks on links, fetches the links href with an "f"
// suffix, e.g. print/index.htmlf, and replace the innerHTML of the target to
// the fragment provided in the response.
//
// If a scrollTop element is provided, wireSPALinks will scroll to the top of
// the element after replacement.
export async function wireSPALinks(links, target, scrollTop, afterNavigate) {
  links.forEach((link) => {
    link.addEventListener("click", async (e) => {
      e.preventDefault()
      await onNavigate(link.href, target, scrollTop)
      afterNavigate && afterNavigate()
    })
  })
  window.addEventListener("popstate", async (e) => {
    const href = e.state && e.state.fragment
    if (href) {
      await replaceFragment(href, target, scrollTop)
      afterNavigate && afterNavigate()
    }
  })
  const normalizedHref = window.location.href.replace(/\/(#|$)/, "/index.html$1")
  window.history.replaceState({ fragment: normalizedHref }, "", window.location.href)
}

async function onNavigate(href, target, scrollTop) {
  const normalizedHref = window.location.href.replace(/\/(#|$)/, "/index.html$1")
  if (href === normalizedHref) {
    return
  }
  window.history.pushState({ fragment: href }, "", href)
  if (href.split("#").shift() !== normalizedHref.split("#").shift()) {
    await replaceFragment(href, target, scrollTop)
  } else {
    scrollToHeading(href, scrollTop)
  }
}

async function replaceFragment(href, target, scrollTop) {
  const fragmentHref = href.split("#").shift() + "f"
  const response = await fetch(fragmentHref)
  target.innerHTML = await response.text()
  scrollToHeading(href, scrollTop)
}

async function scrollToHeading(href, scrollTop) {
  // this should be default browser behavior for onhashchange,
  // but in combination with these CSS rules it does not seem to work.
  //
  //    display: grid | flex;
  //    height: 100% | inherit;
  //    overflow: auto;
  const hash = href.split("#").pop()
  if (hash && hash !== href) {
    const scrollTarget = document.querySelector("#" + hash)
    scrollTarget && scrollTarget.scrollIntoView(true)
    return
  }
  scrollTop && scrollTop.scrollTo(0, 0)
}
