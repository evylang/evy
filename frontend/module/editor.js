// https://github.com/petersolopov/yace - MIT licensed
// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/index.js

export default class Editor {
  constructor(selector, options = {}) {
    this.root = document.querySelector(selector)

    if (!this.root) {
      throw new Error(`element with "${selector}" selector does not exist`)
    }

    const defaultOptions = {
      value: "",
      lineNumbers: true,
      plugins: [preserveIndent(), history(), tab()],
      id: `evy-editor-${Math.random().toString(36).slice(2)}`,
    }

    this.options = {
      ...defaultOptions,
      ...options,
    }

    this.init()
  }

  init() {
    this.textarea = document.createElement("textarea")
    this.textarea.spellcheck = false
    this.textarea.autocorrect = "off"
    this.textarea.autocomplete = "off"
    this.textarea.autocapitalize = "none"
    this.textarea.wrap = "off"
    this.textarea.id = this.options.id

    this.highlighted = document.createElement("pre")
    this.highlighted.classList.add("highlighted")
    this.lines = document.createElement("pre")
    this.lines.classList.add("lines")
    this.errorLines = {}

    const label = document.createElement("label")
    label.htmlFor = this.options.id
    label.textContent = "Evy editor"
    label.style.display = "none"

    this.root.replaceChildren(label, this.textarea, this.lines, this.highlighted)

    this.addTextareaEvents()
    this.update({ value: this.options.value })
  }

  addTextareaEvents() {
    this.handleInput = (event) => {
      const textareaProps = runPlugins(this.options.plugins, event)
      this.update(textareaProps)
    }

    this.handleKeydown = (event) => {
      const textareaProps = runPlugins(this.options.plugins, event)
      this.update(textareaProps)
    }

    this.textarea.addEventListener("input", this.handleInput)
    this.textarea.addEventListener("keydown", this.handleKeydown)
  }

  update(textareaProps) {
    const { value, selectionStart, selectionEnd, errorLines } = textareaProps
    // should be before updating selection otherwise selection will be lost
    if (value != null) {
      this.textarea.value = value
    }

    this.textarea.selectionStart = selectionStart
    this.textarea.selectionEnd = selectionEnd

    if (
      (value === this.value || value == null) &&
      Object.keys(this.errorLines).length === 0 &&
      (!errorLines || Object.keys(errorLines).length === 0)
    ) {
      return
    }
    if (value != null && value != undefined) {
      this.value = value
      const key = this.options.sessionKey
      key && value && sessionStorage.setItem(key, value) // don't set for empty value
    }
    this.errorLines = errorLines || this.errorLines
    const lines = this.value.split("\n")
    this.updateErrorLines(lines)
    const highlighted = this.options.highlighter(this.value, this.errorLines)
    this.highlighted.innerHTML = highlighted + "<br/>"

    this.updateLines(lines)

    if (this.updateCallback) {
      this.updateCallback(value)
    }
  }

  updateLines(lines) {
    const length = lines.length.toString().length

    const paddingLeft = `calc(${length}ch + 1.5rem)`
    this.root.style.paddingLeft = paddingLeft
    this.lines.style.paddingLeft = paddingLeft

    this.lines.innerHTML = lines
      .map((line, number) => {
        const num = `${number + 1}`.padStart(length)
        const errClass = this.errorLines[number + 1] ? "err " : ""
        const lineNumber = `<span class="${errClass}num"> ${num}</span>`
        const lineText = `<span class="${errClass}txt">${escapeHTML(line)}</span>`
        return `${lineNumber}${lineText}`
      })
      .join("\n")
  }

  updateErrorLines(lines) {
    for (const [idx, { text }] of Object.entries(this.errorLines)) {
      if (lines[idx - 1] !== text) {
        delete this.errorLines[idx]
      }
    }
  }

  destroy() {
    this.textarea.removeEventListener("input", this.handleInput)
    this.textarea.removeEventListener("keydown", this.handleKeydown)
  }

  onUpdate(callback) {
    this.updateCallback = callback
  }

  loadSession() {
    const value = sessionStorage.getItem(this.options.sessionKey) || ""
    this.update({ value, errorLines: {} })
  }
}

function runPlugins(plugins, event) {
  const { value, selectionStart, selectionEnd } = event.target

  return plugins.reduce(
    (acc, plugin) => {
      return {
        ...acc,
        ...plugin(acc, event),
      }
    },
    { value, selectionStart, selectionEnd },
  )
}

function escapeHTML(unsafe) {
  return unsafe
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;")
}

// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/plugins/isKey.js
const CODES = {
  backspace: 8,
  tab: 9,
  enter: 13,
  shift: 16,
  control: 17,
  alt: 18,
  pause: 19,
  capslock: 20,
  escape: 27,
  " ": 32,
  pageup: 33,
  pagedown: 34,
  end: 35,
  home: 36,
  arrowleft: 37,
  arrowup: 38,
  arrowright: 39,
  arrowdown: 40,
  insert: 45,
  delete: 46,
  meta: 91,
  numlock: 144,
  scrolllock: 145,
  ";": 186,
  "=": 187,
  ",": 188,
  "-": 189,
  ".": 190,
  "/": 191,
  "`": 192,
  "[": 219,
  "\\": 220,
  "]": 221,
  "'": 222,

  // aliases
  add: 187,
}

const IS_MAC =
  typeof window != "undefined" && /Mac|iPod|iPhone|iPad/.test(window.navigator.platform)

const MODIFIERS = {
  alt: "altKey",
  control: "ctrlKey",
  meta: "metaKey",
  shift: "shiftKey",
  "ctrl/cmd": IS_MAC ? "metaKey" : "ctrlKey",
}

function toKeyCode(name) {
  return CODES[name] || name.toUpperCase().charCodeAt(0)
}

function isKey(string, event) {
  const keys = string.split("+").reduce(
    (acc, key) => {
      if (MODIFIERS[key]) {
        acc.modifiers[MODIFIERS[key]] = true
        return acc
      }

      return {
        ...acc,
        keyCode: toKeyCode(key),
      }
    },
    {
      modifiers: {
        altKey: false,
        ctrlKey: false,
        metaKey: false,
        shiftKey: false,
      },
      keyCode: null,
    },
  )

  const hasModifiers = Object.keys(keys.modifiers).every((key) => {
    const value = keys.modifiers[key]
    return value ? event[key] : !event[key]
  })

  const hasKey = keys.keyCode ? event.which === keys.keyCode : true

  return hasModifiers && hasKey
}

// https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/plugins/history.js
function history() {
  let stack = []
  let activeIndex = null

  const initHistory = (record) => {
    activeIndex = 0
    stack.push(record)
  }

  const rewriteHistory = (record) => {
    stack = stack.slice(0, activeIndex + 1)
    stack.push(record)
    activeIndex = stack.length - 1
  }

  const shouldRecord = (record) => {
    return (
      stack[activeIndex].value !== record.value ||
      stack[activeIndex].selectionStart !== record.selectionStart ||
      stack[activeIndex].selectionEnd !== record.selectionEnd
    )
  }

  return (textareaProps, event) => {
    if (event.type === "keydown") {
      if (isKey("ctrl/cmd+z", event)) {
        event.preventDefault()

        if (activeIndex !== null) {
          // after applying all plugins it can be new props
          if (shouldRecord(textareaProps)) {
            stack.push(textareaProps)
            activeIndex++
          }

          const newActiveIndex = Math.max(0, activeIndex - 1)
          activeIndex = newActiveIndex
          return stack[newActiveIndex]
        }
      }

      if (isKey("ctrl/cmd+shift+z", event)) {
        event.preventDefault()

        if (activeIndex !== null) {
          const newActiveIndex = Math.min(stack.length - 1, activeIndex + 1)
          activeIndex = newActiveIndex
          return stack[newActiveIndex]
        }
      }

      if (activeIndex === null) {
        initHistory(textareaProps)
        return
      }

      if (shouldRecord(textareaProps)) {
        rewriteHistory(textareaProps)
        return
      }
    }

    if (event.type === "input") {
      if (activeIndex === null) {
        initHistory(textareaProps)
      } else {
        rewriteHistory(textareaProps)
      }
    }
  }
}

// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/plugins/preserveIndent.js
const preserveIndent = () => (textareaProps, event) => {
  const { value, selectionStart, selectionEnd } = textareaProps

  if (!isKey("enter", event)) {
    return
  }

  if (event.type !== "keydown") {
    return
  }

  const currentLineNumber = value.substring(0, selectionStart).split("\n").length - 1

  const lines = value.split("\n")
  const currentLine = lines[currentLineNumber]
  const matches = /^\s+/.exec(currentLine)

  if (!matches) {
    return
  }

  event.preventDefault()
  const indent = matches[0]
  const newLine = "\n"

  const inserted = newLine + indent

  const newValue = value.substring(0, selectionStart) + inserted + value.substring(selectionEnd)

  return {
    value: newValue,
    selectionStart: selectionStart + inserted.length,
    selectionEnd: selectionStart + inserted.length,
  }
}

// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/plugins/tab.js
const tab =
  (tabCharacter = "    ") =>
  (textareaProps, event) => {
    const { value, selectionStart, selectionEnd } = textareaProps

    if (event.type !== "keydown") {
      return
    }

    if (isKey("shift+tab", event)) {
      event.preventDefault()
      const linesBeforeCaret = value.substring(0, selectionStart).split("\n")
      const startLine = linesBeforeCaret.length - 1
      const endLine = value.substring(0, selectionEnd).split("\n").length - 1
      const nextValue = value
        .split("\n")
        .map((line, i) => {
          if (i >= startLine && i <= endLine && line.startsWith(tabCharacter)) {
            return line.substring(tabCharacter.length)
          }

          return line
        })
        .join("\n")

      if (value !== nextValue) {
        const startLineText = linesBeforeCaret[startLine]

        return {
          value: nextValue,
          // Move the start cursor if first line in selection was modified
          // It was modified only if it started with a tab
          selectionStart: startLineText.startsWith(tabCharacter)
            ? selectionStart - tabCharacter.length
            : selectionStart,
          // Move the end cursor by total number of characters removed
          selectionEnd: selectionEnd - (value.length - nextValue.length),
        }
      }

      return
    }

    if (isKey("tab", event)) {
      event.preventDefault()
      if (selectionStart === selectionEnd) {
        const updatedSelection = selectionStart + tabCharacter.length
        const newValue =
          value.substring(0, selectionStart) + tabCharacter + value.substring(selectionEnd)

        return {
          value: newValue,
          selectionStart: updatedSelection,
          selectionEnd: updatedSelection,
        }
      }

      const linesBeforeCaret = value.substring(0, selectionStart).split("\n")
      const startLine = linesBeforeCaret.length - 1
      const endLine = value.substring(0, selectionEnd).split("\n").length - 1

      return {
        value: value
          .split("\n")
          .map((line, i) => {
            if (i >= startLine && i <= endLine) {
              return tabCharacter + line
            }

            return line
          })
          .join("\n"),
        selectionStart: selectionStart + tabCharacter.length,
        selectionEnd: selectionEnd + tabCharacter.length * (endLine - startLine + 1),
      }
    }
  }
