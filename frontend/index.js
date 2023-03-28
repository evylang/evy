"use strict"
import Yace from "./module/yace-editor.js"

// --- Globals ---------------------------------------------------------

let wasmModule, wasmInst
const go = newEvyGo()
const canvas = newCanvas()

let jsReadInitialised = false
let stopped = true
let animationStart
let courses
let actions = "fmt,ui,eval"
let editor
let errors = false
// --- Initialise ------------------------------------------------------

initWasm()
initUI()
initCanvas()

// --- Wasm ------------------------------------------------------------

// initWasm loads bytecode and initialises execution environment.
function initWasm() {
  WebAssembly.compileStreaming(fetch("evy.wasm"))
    .then((obj) => (wasmModule = obj))
    .catch((err) => console.error(err))
  const runButton = document.querySelector("#run")
  const runButtonMob = document.querySelector("#run-mobile")
  runButton.onclick = handleRun
  runButton.disabled = false
  runButtonMob.onclick = handleMobRun
  runButtonMob.disabled = false
}

function newEvyGo() {
  // evyEnv contains JS functions from this file exported to wasm/go
  const evyEnv = {
    jsPrint,
    jsRead,
    jsActions,
    jsPrepareUI,
    jsError,
    evySource,
    setEvySource,
    afterStop,
    registerEventHandler,
    // canvas
    move,
    line,
    width,
    circle,
    rect,
    color,
    clear,
    // advanced canvas
    poly,
    ellipse,
    curve,
    stroke,
    fill,
    dash,
    linecap,
    text,
    textsize,
    fontfamily,
    font,
  }
  const go = new Go() // see wasm_exec.js
  go.importObject.env = Object.assign(go.importObject.env, evyEnv)
  return go
}
// jsActions returns the comma separated evy actions to executed, e.g.
// fmt,ui,eval. The result string is written to wasm memory
// bytes. jsActions return the pointer and length of these bytes
// encoded into a single 64 bit number
function jsActions() {
  return stringToMemAddr(actions)
}

function jsPrepareUI(ptr, len) {
  const arr = memToString(ptr, len).split(",")
  const names = Object.fromEntries(arr.map((k) => [k, true]))
  names["read"] ? showElements(".read") : hideElements(".read")
  names["input"] ? showElements(".input") : hideElements(".input")
  needsCanvas(names) ? showElements(".canvas") : hideElements(".canvas")
}

function needsCanvas(f) {
  return (
    f.move ||
    f.line ||
    f.width ||
    f.circle ||
    f.rect ||
    f.color ||
    f.colour ||
    f.clear ||
    f.poly ||
    f.ellipse ||
    f.curve ||
    f.stroke ||
    f.fill ||
    f.dash ||
    f.linecap ||
    f.text ||
    f.textsize ||
    f.fontfamily ||
    f.font
  )
}

// jsPrint converts wasmInst memory bytes from ptr to ptr+len to string and
// writes it to the output textarea.
function jsPrint(ptr, len) {
  const s = memToString(ptr, len)
  const output = document.querySelector("#console")
  output.textContent += s
  output.scrollTo({ behavior: "smooth", left: 0, top: output.scrollHeight })
  if (s.toLowerCase().includes("confetti")) {
    showConfetti()
  }
}

// jsRead reads the content of the "read" textarea. If the textarea
// contains a newline jsRead extracts the string up until the newline
// and empties the textarea. The read stream is written to shared wasm
// memory and the address returned.
function jsRead() {
  const el = document.querySelector("#read")
  if (!jsReadInitialised) {
    getElements(".read").map((el) => el.classList.remove("hidden"))
    el.focus()
    jsReadInitialised = true
  }
  const s = el.value
  const idx = s.indexOf("\n")
  if (idx === -1) {
    return 0
  }
  el.value = ""
  return stringToMemAddr(s.slice(0, idx))
}

function jsError(ptr, len) {
  errors = true
  const code = editor.value
  const lines = code.split("\n")
  const errs = memToString(ptr, len).split("\n")
  const re = /line (?<line>\d+) column (?<col>\d+): (?<msg>.*)/
  let msgs = ""
  const errorLines = {}
  for (const err of errs) {
    const g = err.match(re).groups
    if (!errorLines[g.line]) {
      errorLines[g.line] = { col: g.col, text: lines[g.line - 1] }
    }
    msgs += `line ${g.line}: ${g.msg}\n`
  }
  const output = document.querySelector("#console")
  output.textContent = msgs
  output.scrollTo({ behavior: "smooth", left: 0, top: 0 })
  editor.update({ errorLines })
}

// evySource writes the evy source code into wasm memory as bytes
// and returns pointer and length encoded into a single 64 bit number
function evySource() {
  const code = editor.value
  return stringToMemAddr(code)
}

// setEvySource is exported to evy go/wasm and called after formatting
function setEvySource(ptr, len) {
  const source = memToString(ptr, len)
  editor.update({ value: source, errorLines: {} })
}

function memToString(ptr, len) {
  const buf = new Uint8Array(wasmInst.exports.memory.buffer, ptr, len)
  const s = new TextDecoder("utf8").decode(buf)
  return s
}

function stringToMem(s) {
  const bytes = new TextEncoder("utf8").encode(s)
  const e = wasmInst.exports
  const len = bytes.length
  const ptr = e.alloc(len)
  const mem = new Uint8Array(e.memory.buffer, ptr, len)
  mem.set(new Uint8Array(bytes))
  return { ptr, len }
}

function stringToMemAddr(s) {
  const ptrLen = stringToMem(s)
  return ptrLenToBigInt(ptrLen)
}

function ptrLenToBigInt({ ptr, len }) {
  const ptrLen = (BigInt(ptr) << 32n) | (BigInt(len) & 0x00000000ffffffffn)
  const ptrLenNum = Number(ptrLen)
  return ptrLenNum
}

// --- UI: handle run --------------------------------------------------

async function handleRun() {
  stopped ? start() : stop()
}

// handleMobRun handles three states for mobile devices:
// run -> stop -> code
async function handleMobRun() {
  if (onCodeScreen()) {
    // we need to wait for the slide transition to finish otherwise
    // el.focus() in jsRead() messes up the layout
    await slide()
    start()
    return
  }
  // on output screen
  if (stopped) {
    const runButtonMob = document.querySelector("#run-mobile")
    runButtonMob.innerText = "Run"
    slide()
    return
  }
  stop()
}

// start calls evy wasm/go main(). It parses, formats and evaluates evy
// code and initialises the output ui.
async function start() {
  stopped = false
  errors = false
  wasmInst = await WebAssembly.instantiate(wasmModule, go.importObject)
  editor.update({ errorLines: {} })
  clearOutput()

  const runButton = document.querySelector("#run")
  const runButtonMob = document.querySelector("#run-mobile")
  runButton.innerText = "Stop"
  runButton.classList.add("running")
  runButtonMob.innerText = "Stop"
  runButtonMob.classList.add("running")
  actions = "fmt,ui,eval"
  go.run(wasmInst)
}

// format calls evy wasm/go main() but doesn't evaluate.
async function format() {
  stopped = false
  errors = false
  wasmInst = await WebAssembly.instantiate(wasmModule, go.importObject)
  actions = "fmt,ui"
  go.run(wasmInst)
}

// stop terminates program in execution via exports.stop wasm/go then
// calls afterStop to reset UI. However, to ensure consistent state
// execute afterStop if program is already stopped.
function stop() {
  stopped = true
  wasmInst ? wasmInst.exports.stop() : afterStop()
}

// afterStop is exported to evy go/wasm and called when execution finishes
function afterStop() {
  removeEventHandlers()
  stopped = true
  animationStart = undefined
  jsReadInitialised = false
  wasmInst = undefined

  const runButton = document.querySelector("#run")
  const runButtonMob = document.querySelector("#run-mobile")
  runButton.classList.remove("running")
  runButton.innerText = "Run"
  runButtonMob.classList.remove("running")
  runButtonMob.innerText = onCodeScreen() ? "Run" : "Code"

  const readEl = document.querySelector("#read")
  document.activeElement === readEl && readEl.blur()
}

function onCodeScreen() {
  return !document.querySelector("main").classList.contains("view-output")
}

async function slide() {
  const el = document.querySelector("main")
  const cl = el.classList
  return new Promise((resolve) => {
    el.ontransitionend = resolve
    onCodeScreen() ? cl.add("view-output") : cl.remove("view-output")
  })
}

async function stopAndSlide() {
  if (!onCodeScreen()) {
    await slide()
  }
  stop()
}

function clearOutput() {
  document.querySelector("#console").textContent = ""
  resetCanvas()
}

// --- UI: initialisation ----------------------------------------------

async function initUI() {
  document.addEventListener("keydown", ctrlEnterListener)
  await fetchCourses()
  window.addEventListener("hashchange", handleHashChange)
  document.querySelector("#modal-close").onclick = hideModal
  document.querySelector("#share").onclick = share
  initModal()
  handleHashChange()
  initEditor()
}

async function fetchCourses() {
  const resp = await fetch("courses/courses.json")
  courses = await resp.json()
  courses.units = {}
  for (const course of courses.courseList) {
    for (const unit of course.units) {
      courses.units[unit.id] = { ...unit, courseTitle: course.title, courseID: course.id }
    }
  }
}

function ctrlEnterListener(e) {
  if ((e.metaKey || e.ctrlKey) && event.key === "Enter") {
    document.querySelector(".editor textarea").blur()
    handleRun()
  }
}

// --- UI: URL-hash change handling ------------------------------------

async function handleHashChange() {
  hideModal()
  await stopAndSlide() // go to code screen for new code
  let opts = parseHash()
  if (!opts.source && !opts.unit && !opts.content) {
    opts = { unit: "welcome" }
  }
  if (opts.content) {
    const decoded = await decode(opts.content)
    editor.update({ value: decoded, errorLines: {} })
    return
  }
  let crumbs = ["Evy"]
  if (opts.unit) {
    const unit = courses.units[opts.unit]
    opts.source = `courses/${unit.courseID}/${opts.unit}.evy`
    crumbs = [unit.courseTitle, unit.title]
  }
  try {
    const response = await fetch(opts.source)
    if (response.status < 200 || response.status > 299) {
      throw new Error("invalid response status", response.status)
    }
    const source = await response.text()
    editor.update({ value: source, errorLines: {} })
    document.querySelector(".editor-wrap").scrollTo(0, 0)
    updateBreadcrumbs(crumbs)
    clearOutput()
    format()
  } catch (err) {
    console.error(err)
  }
}

function parseHash() {
  // parse url fragment into object
  // e.g. https://example.com#a=1&b=2 into {a: "1", b: "2"}
  // then fetch source from URL and write it to code input.
  const strs = window.location.hash.substring(1).split("&") //  ["a=1", "b=2"]
  const entries = strs.map((s) => s.split("=")) // [["a", "1"], ["b", "2"]]
  if (entries.length === 1 && entries[0].length === 1) {
    // shortcut for evy.dev#draw loading evy.dev/courses/sample/draw/draw.evy
    const unit = entries[0][0]
    if (courses && courses.units[unit]) {
      return { unit }
    }
  }
  return Object.fromEntries(entries)
}

// --- Canvas graphics -------------------------------------------------

function newCanvas() {
  return {
    x: 0,
    y: 0,
    ctx: null,
    factor: 10,
    scale: { x: 1, y: -1 },
    width: 100,
    height: 100,
    fill: true,
    stroke: true,

    offset: { x: 0, y: -100 }, // height
  }
}

function initCanvas() {
  const c = document.querySelector("#canvas")
  const b = c.parentElement.getBoundingClientRect()
  c.width = Math.abs(scaleX(canvas.width))
  c.height = Math.abs(scaleY(canvas.height))
  canvas.ctx = c.getContext("2d")
}

function resetCanvas() {
  const ctx = canvas.ctx
  ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height)
  ctx.fillStyle = "black"
  ctx.strokeStyle = "black"
  ctx.lineWidth = 1
  canvas.fill = true
  canvas.stroke = true
  ctx.lineCap = "butt"
  ctx.setLineDash([])
  ctx.font = "16px regular"
  textsize(6)
  move(0, 0)
}

function scaleX(x) {
  return canvas.scale.x * canvas.factor * x
}

function scaleY(y) {
  return canvas.scale.y * canvas.factor * y
}

function transformX(x) {
  return scaleX(x + canvas.offset.x)
}

function transformY(y) {
  return scaleY(y + canvas.offset.y)
}

// move is exported to evy go/wasm.
function move(x, y) {
  movePhysical(transformX(x), transformY(y))
}

function movePhysical(px, py) {
  canvas.x = px
  canvas.y = py
}

// line is exported to evy go/wasm.
function line(x2, y2) {
  const { ctx, x, y } = canvas
  const px2 = transformX(x2)
  const py2 = transformY(y2)
  ctx.beginPath()
  ctx.moveTo(x, y)
  ctx.lineTo(px2, py2)
  ctx.stroke()
  movePhysical(px2, py2)
}

// color is exported to evy go/wasm.
function color(ptr, len) {
  const s = memToString(ptr, len)
  canvas.stroke = s !== "none"
  canvas.fill = s !== "none"
  canvas.ctx.fillStyle = s
  canvas.ctx.strokeStyle = s
}

// width is exported to evy go/wasm.
function width(n) {
  canvas.ctx.lineWidth = scaleX(n)
}

// rect is exported to evy go/wasm.
function rect(dx, dy) {
  const { ctx, x, y, fill, stroke } = canvas
  const sDX = scaleX(dx)
  const sDY = scaleY(dy)
  fill && ctx.fillRect(x, y, sDX, sDY)
  stroke && ctx.strokeRect(x, y, sDX, sDY)
  movePhysical(x + sDX, y + sDY)
}

// circle is exported to evy go/wasm.
function circle(r) {
  const { x, y, ctx, fill, stroke } = canvas
  ctx.beginPath()
  ctx.arc(x, y, scaleX(r), 0, Math.PI * 2, true)
  fill && ctx.fill()
  stroke && ctx.stroke()
}

// clear is exported to evy go/wasm.
function clear(ptr, len) {
  const ctx = canvas.ctx
  if (len === 0) {
    ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height)
    return
  }
  const color = memToString(ptr, len)
  const prevColor = ctx.fillStyle
  ctx.fillStyle = color
  ctx.fillRect(0, 0, ctx.canvas.width, ctx.canvas.height)
  ctx.fillStyle = prevColor
}

// poly is exported to evy go/wasm.
function poly(ptr, len) {
  const { x, y, ctx, fill, stroke } = canvas
  const s = memToString(ptr, len)
  const points = parsePoints(s)
  ctx.beginPath()
  ctx.moveTo(transformX(points[0][0]), transformY(points[0][1]))
  for (const point of points.slice(1)) {
    const x = transformX(point[0])
    const y = transformY(point[1])
    ctx.lineTo(x, y)
  }
  fill && ctx.fill()
  stroke && ctx.stroke()
}

function parsePoints(s) {
  const arr = s.split(" ")
  const points = []
  for (let i = 0; i < arr.length; i += 2) {
    points.push([Number(arr[i]), Number(arr[i + 1])])
  }
  return points
}

// ellipse is exported to evy go/wasm.
// see https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/ellipse
function ellipse(x, y, radiusX, radiusY, rotation, startAngle, endAngle) {
  const rad = Math.PI / 180
  const { ctx, fill, stroke } = canvas
  ctx.beginPath()
  ctx.ellipse(
    transformX(x),
    transformY(y),
    transformX(radiusX),
    transformX(radiusY),
    rotation * rad,
    startAngle * rad,
    endAngle * rad
  )
  fill && ctx.fill()
  stroke && ctx.stroke()
}

// curve is exported to evy go/wasm.
// curve draws connected curve segments encoded as string
// representing 2 dimensional array, e.g.:
//  "1 2 3 4,5 6" => [[1,2,3,4], [5,6]]
// the curve segments here are 1 2 3 4 and 5 6
//
// curve segments are interpreted differently depending on their
// **length**:
// 2: endX, endY - Line from current position end position x, y
// 4: controlX, controlY, endX, endY: quadratic bezier curve, see
//    https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/quadraticCurveTo
// 5: control1X, control1Y, control2X, control2Y, radius: arcTo, see
//    https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/arcTo
// 6: control1X, control1Y, control2X, control2Y, endX, endY: bezierCurveTo, see
//    https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/bezierCurveTo
//
// This notation is very dense and is intended as escape hatch for advanced graphics.
function curve(ptr, len) {
  const s = memToString(ptr, len)
  // parse "1 2 3 4,5 6" into [[1,2,3,4], [5,6]]:
  const curves = s.split(",").map((s) => s.split(" ").map(Number))

  const { x, y, ctx, fill, stroke } = canvas

  ctx.beginPath()
  ctx.moveTo(x, y)

  for (const curve of courves) {
    switch (curve.length) {
      case 2:
        ctx.lineTo(transformX(curve[0]), transformY(curve[1]))
        break
      case 4:
        ctx.quadraticCurveTo(
          transformX(curve[0]), // controlPoint.x
          transformY(curve[1]), // controlPoint.y
          transformX(curve[2]), // endPoint.x
          transformY(curve[3]) // endPoint.y
        )
        break
      case 5:
        ctx.arcTo(
          transformX(curve[0]), // controlPoint1.x
          transformY(curve[1]), // controlPoint1.y
          transformX(curve[2]), // controlPoint2.x
          transformY(curve[3]), // controlPoint1.y
          transformX(curve[4]) // radius
        )
        break
      case 6:
        ctx.bezierCurveTo(
          transformX(curve[0]), // controlPoint1.x
          transformY(curve[1]), // controlPoint1.y
          transformX(curve[2]), // controlPoint2.x
          transformY(curve[3]), // controlPoint1.y
          transformX(curve[4]), // endPoint.x
          transformY(curve[5]) // endPoint.y
        )
        break
    }
  }
  fill && ctx.fill()
  stroke && ctx.stroke()
}

// stroke is exported to evy go/wasm.
function stroke(ptr, len) {
  const s = memToString(ptr, len)
  canvas.stroke = s !== "none"
  canvas.ctx.strokeStyle = s
}

// fill is exported to evy go/wasm.
function fill(ptr, len) {
  const s = memToString(ptr, len)
  canvas.fill = s !== "none"
  canvas.ctx.fillStyle = s
}

// dash is exported to evy go/wasm.
function dash(ptr, len) {
  const s = memToString(ptr, len)
  const nums = s.split(" ").map(Number).map(transformX)
  canvas.ctx.setLineDash(nums)
}

// linecap is exported to evy go/wasm.
function linecap(ptr, len) {
  const s = memToString(ptr, len)
  canvas.ctx.lineCap = s
}

// text is exported to evy go/wasm.
function text(ptr, len) {
  const { x, y, ctx } = canvas
  const text = memToString(ptr, len)
  ctx.fillText(text, x, y)
}

// textsize is exported to evy go/wasm.
function textsize(size) {
  const { width, ctx } = canvas
  const style = parsedStyle(`font: ${ctx.font}`)
  style.fontSize = (ctx.canvas.width / 100) * size + "px"
  ctx.font = style.font
}

var parsedStyle = function (cssString) {
  var el = document.createElement("span")
  el.setAttribute("style", cssString)
  return el.style // CSSStyleDeclaration object
}

// font is exported to evy go/wasm.
// see https://developer.mozilla.org/en-US/docs/Web/CSS/font
function font(ptr, len) {
  const font = memToString(ptr, len)
  canvas.ctx.font = font
}

// fontfamily is exported to evy go/wasm.
function fontfamily(ptr, len) {
  const ctx = canvas.ctx
  const s = memToString(ptr, len)
  const style = parsedStyle(`font: ${ctx.font}`)
  style.fontFamily = s
  ctx.font = style.font
}

function logicalX(e) {
  const scaleX = (canvas.width * canvas.scale.x) / e.target.offsetWidth
  return e.offsetX * scaleX - canvas.offset.x
}

function logicalY(e) {
  const scaleY = (canvas.height * canvas.scale.y) / e.target.offsetHeight
  return e.offsetY * scaleY - canvas.offset.y
}

function initEditor() {
  const value = `move 10 20
line 50 50
rect 25 25
color "red"
circle 10

x := 12
print "x:" x
if x > 10
    print "🍦 big x"
end`
  editor = new Yace(".editor", { value })
}

// --- eventHandlers, evy `on` -----------------------------------------

// registerEventHandler is exported to evy go/wasm
function registerEventHandler(ptr, len) {
  const c = document.querySelector("#canvas")
  const s = memToString(ptr, len)
  const exp = wasmInst.exports
  if (s === "down") {
    c.onpointerdown = (e) => exp.onDown(logicalX(e), logicalY(e))
  } else if (s === "up") {
    c.onpointerup = (e) => exp.onUp(logicalX(e), logicalY(e))
  } else if (s === "move") {
    c.onpointermove = (e) => exp.onMove(logicalX(e), logicalY(e))
  } else if (s === "key") {
    unfocusRunBotton()
    document.addEventListener("keydown", keydownListener)
  } else if (s === "input") {
    addInputHandlers()
  } else if (s === "animate") {
    window.requestAnimationFrame(animationLoop)
  } else {
    console.error("cannot register unknown event", s)
  }
}

function unfocusRunBotton() {
  const runButton = document.querySelector("#run")
  const runButtonMob = document.querySelector("#run-mobile")
  document.activeElement === runButton && runButton.blur()
  document.activeElement === runButtonMob && runButtonMob.blur()
}

function keydownListener(e) {
  if (e.target.id == "code") return // skip for source code input
  const { ptr, len } = stringToMem(e.key)
  wasmInst.exports.onKey(ptr, len)
}

function addInputHandlers() {
  getElements(".input").map((el) => el.classList.remove("hidden"))
  const exp = wasmInst.exports
  const els = document.querySelectorAll("input#sliderx,input#slidery")
  for (const el of els) {
    el.onchange = (e) => {
      const id = stringToMem(e.target.id)
      const val = stringToMem(e.target.value)
      wasmInst.exports.onInput(id.ptr, id.len, val.ptr, val.len)
    }
  }
}

function removeEventHandlers() {
  const c = document.querySelector("#canvas")
  c.onpointerdown = null
  c.onpointerup = null
  c.onpointermove = null
  const els = document.querySelectorAll("input#sliderx,input#slidery")
  for (const el of els) {
    el.onchange = null
  }
  document.removeEventListener("keydown", keydownListener)
}

function animationLoop(ts) {
  if (stopped) {
    return
  }
  if (animationStart === undefined) {
    animationStart = ts
  }
  wasmInst.exports.onAnimate(ts - animationStart)
  window.requestAnimationFrame(animationLoop)
}

// --- UI: modal navigation --------------------------------------------

function initModal() {
  const modalMain = document.querySelector("#modal .modal-main")
  modalMain.textContent = ""
  for (const course of courses.courseList) {
    const item = document.createElement("div")
    item.classList.add("item")
    const h2 = document.createElement("h2")
    h2.textContent = `${course.emoji} ${course.title}`
    const ul = document.createElement("ul")
    item.replaceChildren(h2, ul)
    for (const unit of course.units) {
      const li = document.createElement("li")
      const a = document.createElement("a")
      a.textContent = unit.title
      a.href = `#${unit.id}`
      a.onclick = hideModal
      li.appendChild(a)
      ul.appendChild(li)
    }
    modalMain.appendChild(item)
  }
}

function hideModal() {
  const el = document.querySelector("#modal")
  el.classList.add("hidden")
}

function showCourses() {
  const courses = document.querySelector("#modal-courses")
  courses.classList.remove("hidden")
  const share = document.querySelector("#modal-share")
  share.classList.add("hidden")
  const modal = document.querySelector("#modal")
  modal.classList.remove("hidden")
}

function showSharing() {
  const share = document.querySelector("#modal-share")
  share.classList.remove("hidden")
  const courses = document.querySelector("#modal-courses")
  courses.classList.add("hidden")
  const modal = document.querySelector("#modal")
  modal.classList.remove("hidden")
}

function updateBreadcrumbs(crumbs) {
  const ul = document.querySelector("header ul.breadcrumbs")
  const breadcrumbs = crumbs.map((c) => breadcrumb(c))
  ul.replaceChildren(...breadcrumbs)
}

function breadcrumb(s) {
  const btn = document.createElement("button")
  btn.textContent = s
  btn.onclick = () => showCourses()
  const li = document.createElement("li")
  li.appendChild(btn)
  return li
}

// --- UI: Confetti Easter Egg -----------------------------------------
//
// When code input string contains the sub string "confetti" show
// confetti on Run button click.

function showConfetti() {
  const names = ["🦊", "🐐"]
  const colors = ["red", "purple", "blue", "orange", "gold", "green"]
  let confetti = new Array(100)
    .fill()
    .map((_, i) => {
      return {
        name: names[i % names.length],
        x: Math.random() * 100,
        y: -20 - Math.random() * 100,
        r: 0.1 + Math.random() * 1,
        color: colors[i % colors.length],
      }
    })
    .sort((a, b) => a.r - b.r)

  const cssText = (c) =>
    `background: ${c.color}; left: ${c.x}%; top: ${c.y}%; transform: scale(${c.r})`
  const confettiDivs = confetti.map((c) => {
    const div = document.createElement("div")
    div.style.cssText = cssText(c)
    div.classList.add("confetti")
    div.textContent = c.name
    document.body.appendChild(div)
    return div
  })

  let frame

  function loop() {
    frame = requestAnimationFrame(loop)
    confetti = confetti.map((c, i) => {
      c.y += 0.7 * c.r
      if (c.y > 120) c.y = -20
      const div = confettiDivs[i]
      div.style.cssText = cssText(c)
      return c
    })
  }

  loop()
  setTimeout(() => {
    cancelAnimationFrame(frame)
    confettiDivs.forEach((div) => div.remove())
  }, 10000)
  setTimeout(() => {
    confettiDivs.forEach((div) => div.classList.add("fadeout"))
  }, 8500)
}

// --- Share / load snippets -------------------------------------------

async function share() {
  await format()
  const el = document.querySelector("#modal-share")

  if (errors) {
    const msg = document.createElement("label")
    msg.textContent = "Fix errors first please."
    const button = document.createElement("button")
    button.innerText = "OK"
    button.onclick = hideModal
    el.replaceChildren(msg, button)
    showSharing()
    return
  }
  const encoded = await encode(editor.value)
  const msg = document.createElement("label")
  msg.textContent = "Share"
  const input = document.createElement("input")
  input.type = "text"
  input.onclick = input.select
  const baseurl = window.location.origin + window.location.pathname
  input.value = `${baseurl}#content=${encoded}`
  const button = document.createElement("button")
  button.className = "copy"
  button.innerHTML = `<svg><use href="#icon-copy" /></svg>`
  button.onclick = () => {
    navigator.clipboard.writeText(input.value)
    hideModal()
  }
  el.replaceChildren(msg, input, button)
  showSharing()
}

async function encode(input) {
  await polyfillCompression()
  const buffer = new TextEncoder().encode(input)
  const stream = readableStream(buffer).pipeThrough(new CompressionStream("gzip"))
  const compressedBuffer = await bufferFromStream(stream)
  const encoded = btoa(String.fromCharCode(...compressedBuffer))
  return encoded
}

async function decode(encoded) {
  await polyfillCompression()
  const bytes = atob(encoded).split("")
  const buffer = new Uint8Array(bytes.map((b) => b.charCodeAt(0)))
  const stream = readableStream(buffer).pipeThrough(new DecompressionStream("gzip"))
  const decompressedBuffer = await bufferFromStream(stream)
  const decoded = new TextDecoder().decode(decompressedBuffer)
  return decoded
}

async function polyfillCompression() {
  if (!window.CompressionStream) {
    await import("https://unpkg.com/compression-streams-polyfill")
  }
}

function readableStream(buffer) {
  return new ReadableStream({
    start(controller) {
      controller.enqueue(buffer)
      controller.close()
    },
  })
}

async function bufferFromStream(stream) {
  const reader = stream.getReader()
  let buffer = new Uint8Array()
  while (true) {
    const { done, value } = await reader.read()
    if (done) {
      break
    }
    buffer = new Uint8Array([...buffer, ...value])
  }
  return buffer
}

// --- Utilities -------------------------------------------------------

function getElements(q) {
  if (!q) {
    return []
  }
  try {
    return Array.from(document.querySelectorAll(q))
  } catch (error) {
    consol.error("getElements", error)
    return []
  }
}

function showElements(q) {
  getElements(q).map((el) => el.classList.remove("hidden"))
}

function hideElements(q) {
  getElements(q).map((el) => el.classList.add("hidden"))
}
