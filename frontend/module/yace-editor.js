// https://github.com/petersolopov/yace - MIT licensed
// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/index.js
export default class Yace {
  constructor(selector, options = {}) {
    if (!selector) {
      throw new Error("selector is not defined")
    }

    this.root = selector instanceof Node ? selector : document.querySelector(selector)

    if (!this.root) {
      throw new Error(`element with "${selector}" selector does not exist`)
    }

    const defaultOptions = {
      value: "",
      lineNumbers: true,
      styles: {},
      plugins: [preserveIndent(), history(), tab()],
      highlighter: highlightEvy,
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

    this.highlighted = document.createElement("pre")
    this.highlighted.classList.add("highlighted")
    this.lines = document.createElement("pre")
    this.lines.classList.add("lines")
    this.errorLines = {}

    this.root.replaceChildren(this.textarea, this.lines, this.highlighted)

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
        const lineText = `<span class="${errClass}txt">${escape(line)}</span>`
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

function escape(unsafe) {
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

// evy highlighter
function highlightEvy(val, errorLines) {
  const tokens = tokenize(val, errorLines)
  const span = (t) => `<span class="${t.err}${t.type}">${escape(t.val)}</span>`
  const result = tokens.map((t) => span(t)).join("")
  return result
}

const builtins = new Set([
  "atan2",
  "ceil",
  "circle",
  "clear",
  "cls",
  "color",
  "colour",
  "cos",
  "dash",
  "del",
  "ellipse",
  "endswith",
  "exit",
  "fill",
  "floor",
  "font",
  "grid",
  "gridn",
  "has",
  "index",
  "join",
  "len",
  "line",
  "linecap",
  "log",
  "lower",
  "max",
  "min",
  "move",
  "poly",
  "pow",
  "print",
  "printf",
  "rand",
  "rand1",
  "read",
  "rect",
  "replace",
  "round",
  "sin",
  "sleep",
  "split",
  "sprint",
  "sprintf",
  "sqrt",
  "startswith",
  "str2bool",
  "str2num",
  "stroke",
  "text",
  "trim",
  "upper",
  "width",
])

const keywords = new Set([
  "num",
  "string",
  "bool",
  "any",
  "true",
  "false",
  "and",
  "or",
  "if",
  "else",
  "func",
  "return",
  "on",
  "for",
  "range",
  "while",
  "break",
  "end",
])

function tokenize(str, errorLines) {
  let tokens = []
  let i = 0
  let prev = ""
  let funcs = new Set()
  let lineIdx = 1
  let lineOffset = 0
  const chars = Array.from(str)

  while (i < chars.length) {
    const start = i
    const c = chars[i]
    let type
    i++
    if (isWS(c)) {
      type = "ws"
      i = readWS(chars, i)
    } else if (isOP(c)) {
      type = "op"
      chars[i] === "=" && i++
    } else if (c === ":" && chars[i] === "=") {
      i++
      type = "op"
    } else if (isPunc(c) || (c === ":" && chars[i] !== "=")) {
      type = "punc"
    } else if (c === "/" && chars[i] == "/") {
      type = "comment"
      i = readComment(chars, i)
    } else if (c === "/" && chars[i] != "/") {
      type = "op"
    } else if (c === '"') {
      type = "str"
      i = readString(chars, i)
    } else if (isDigit(c)) {
      type = "num"
      i = readNum(chars, i)
    } else if (isLetter(c)) {
      type = "ident"
      i = readIdent(chars, i)
    } else if (c === "\n") {
      type = "nl"
    } else {
      type = "error"
    }
    let val = chars.slice(start, i).join("")
    if (type == "ident") {
      type = identType(val, prev, funcs)
    }
    const errLine = errorLines[lineIdx]
    let err = ""
    if (errLine) {
      const errCol = errLine.col - 1
      const startCol = start - lineOffset
      const endCol = i - lineOffset
      if (errCol >= startCol && errCol < endCol) {
        err = "err "
      }
    }
    if (type !== "ws") {
      prev = val
    }
    if (type === "nl") {
      lineIdx++
      lineOffset = i
      if (err) {
        val = " \n"
      }
    }
    tokens.push({ type, val, err })
  }
  tokens.forEach((t) => {
    if (t.type === "ident" && funcs.has(t.val)) {
      t.type = "func"
    }
  })
  return tokens
}

function isWS(s) {
  return s === " " || s === "\t" || s === "\r"
}

function readWS(s, i) {
  while (isWS(s[i])) {
    i++
  }
  return i
}

function isOP(s) {
  return (
    s === "+" ||
    s === "-" ||
    s === "*" ||
    s === "%" ||
    s === "!" ||
    s === "<" ||
    s === ">" ||
    s === "!" ||
    s === "="
  )
}

function isPunc(s) {
  return s === "(" || s === ")" || s === "[" || s === "]" || s === "{" || s === "}" || s === "."
}

function isDigit(s) {
  return s >= "0" && s <= "9"
}

function readNum(s, i) {
  while (isDigit(s[i]) || s[i] === ".") {
    i++
  }
  return i
}

function isLetter(s) {
  return (s >= "a" && s <= "z") || (s >= "A" && s <= "Z") || s === "_" || /\p{L}/u.test(s)
}

function readIdent(s, i) {
  while ((isLetter(s[i]) || isDigit(s[i])) && i < s.length) {
    i++
  }
  return i
}

function readString(s, i) {
  let escaped = false
  while (i < s.length) {
    const c = s[i]
    if (c === "\n") {
      return i
    }
    if (c === '"' && !escaped) {
      return i + 1
    }
    escaped = c === "\\" && !escaped
    i++
  }
  return i
}

function readComment(s, i) {
  while (s[i] !== "\n" && i < s.length) {
    i++
  }
  return i
}

function identType(val, prev, funcs) {
  if (keywords.has(val)) {
    return "keyword"
  }
  if (builtins.has(val) && prev !== ".") {
    return "builtin"
  }
  if (prev === "func") {
    funcs.add(val)
    return "func"
  }
  if (prev === "on") {
    return "func"
  }
  if (funcs.has(val)) {
    return "func"
  }
  return "ident"
}
