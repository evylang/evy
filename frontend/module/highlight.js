// evy highlighter
export default function highlightEvy(val, errorLines) {
  const tokens = tokenize(val, errorLines)
  const result = tokens.map(tokenToSpan).join("")
  return result
}

function tokenToSpan(token) {
  if (token.type !== "comment") {
    return `<span class="${token.type}">${escapeHTML(token.val)}</span>`
  }
  let words = escapeHTML(token.val).split(" ")
  words = words.map((w) => (w.startsWith("https://") ? `<a href=${w} target="_blank">${w}</a>` : w))
  return `<span class="comment">${words.join(" ")}</span>`
}

function escapeHTML(unsafe) {
  return unsafe
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;")
}

const builtins = new Set([
  "abs",
  "assert",
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
  "hsl",
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
  "repr",
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
  errorLines ??= {}
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
  if (keywords.has(val) && prev !== ".") {
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
