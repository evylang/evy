// initThemeToggle initializes a theme toggle that changes between dark and
// light themes, storing the user's  preference in localStorage and
// defaulting to OS dark mode preference.
//
// It returns a function for manual updates, taking a boolean parameter
// (isDark). This function sets the theme to "dark" if 'isDark' is true,
// and "light" if false.
//
// Minimal example: ../css/switch.html
export default function initThemeToggle(selector, storageKey) {
  function updateTheme(dark) {
    window.localStorage.setItem(storageKey, dark ? "dark" : "light")
    document.querySelector(selector).checked = !!dark
  }
  const toggle = document.querySelector(selector)
  toggle.addEventListener("click", (e) => updateTheme(e.target.checked))

  return updateTheme
}
