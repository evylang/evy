"use strict"

let wasmModule, wasmInst
let sourcePtr, sourceLength
const go = newEvyGo()
const runButton = document.querySelector("#run")
const runButtonMob = document.querySelector("#run-mob")

// initWasm loads bytecode and initialises execution environment.
function initWasm() {
  WebAssembly.compileStreaming(fetch("evy.wasm"))
    .then((obj) => (wasmModule = obj))
    .catch((err) => console.error(err))
  runButton.onclick = handle
  runButton.disabled = false
  runButtonMob.onclick = handleMob
  runButtonMob.disabled = false
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

let jsReadInitialised = false
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

// evySource writes the evy source code into wasm memory as bytes
// and returns pointer and length encoded into a single 64 bit number
function evySource() {
  const code = document.querySelector("#code").value
  return stringToMemAddr(code)
}

// setEvySource is exported to evy go/wasm and called after formatting
function setEvySource(ptr, len) {
  const source = memToString(ptr, len)
  document.querySelector("#code").value = source
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

let stopped = true
async function handle() {
  stopped ? start() : stop()
}

async function handleMob() {
  if (onCodeScreen()) {
    // we need to wait for the slide transition to finish otherwise
    // el.focus() in jsRead() messes up the layout
    await slide()
    start()
    return
  }
  // on output screen
  if (stopped) {
    runButtonMob.innerText = "Run"
    slide()
    return
  }
  stop()
}

// start retrieves the input string from the code pane and
// converts it to wasm memory bytes. It then calls the evy main()
// function running the evaluator after parsing.
async function start() {
  stopped = false
  wasmInst = await WebAssembly.instantiate(wasmModule, go.importObject)
  clearOutput()
  runButton.innerText = "Stop"
  runButton.classList.add("running")
  runButtonMob.innerText = "Stop"
  runButtonMob.classList.add("running")
  go.run(wasmInst)
}

// stop terminates program in execution via exports.stop
function stop() {
  stopped = true
  wasmInst ? wasmInst.exports.stop() : onStopped()
}

// onStopped is exported to evy go/wasm and called when execution finishes
function onStopped() {
  removeEventHandlers()
  stopped = true
  animationStart = undefined
  jsReadInitialised = false
  wasmInst = undefined
  runButton.classList.remove("running")
  runButton.innerText = "Run"
  runButtonMob.classList.remove("running")
  runButtonMob.innerText = onCodeScreen() ? "Run" : "Code"
}

function newEvyGo() {
  const evyEnv = {
    jsRead,
    jsPrint,
    evySource,
    setEvySource,
    move,
    line,
    width,
    circle,
    rect,
    color,
    onStopped,
    registerEventHandler,
    sourcePtr: () => sourcePtr,
    sourceLength: () => sourceLength,
  }
  const go = new Go() // see wasm_exec.js
  go.importObject.env = Object.assign(go.importObject.env, evyEnv)
  return go
}

function clearOutput() {
  document.querySelector("#console").textContent = ""
  resetCanvas()
}

// --------------------------------------------------
// confetti easter egg
// When code input string contains the sub string "confetti"
// show confetti when clicking Run button.
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

// graphics
const canvas = {
  x: 0,
  y: 0,
  ctx: null,
  factor: 10,
  scale: { x: 1, y: -1 },
  width: 100,
  height: 100,

  offset: { x: 0, y: -100 }, // height
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

// move is exported to evy go/wasm
function move(x, y) {
  movePhysical(transformX(x), transformY(y))
}

function movePhysical(px, py) {
  canvas.x = px
  canvas.y = py
}

// line is exported to evy go/wasm
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

// color is exported to evy go/wasm
function color(ptr, len) {
  const s = memToString(ptr, len)
  canvas.ctx.fillStyle = s
  canvas.ctx.strokeStyle = s
}

// width is exported to evy go/wasm
function width(n) {
  canvas.ctx.lineWidth = scaleX(n)
}

// rect is exported to evy go/wasm
function rect(dx, dy) {
  const { ctx, x, y } = canvas
  const sDX = scaleX(dx)
  const sDY = scaleY(dy)
  canvas.ctx.fillRect(x, y, sDX, sDY)
  movePhysical(x + sDX, y + sDY)
}

// circle is exported to evy go/wasm
function circle(r) {
  const { x, y, ctx } = canvas
  ctx.beginPath()
  ctx.arc(x, y, scaleX(r), 0, Math.PI * 2, true)
  ctx.fill()
}

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
    document.addEventListener("keydown", keydownListener)
  } else if (s === "input") {
    addInputHandlers()
  } else if (s === "animate") {
    window.requestAnimationFrame(animationLoop)
  } else {
    console.error("cannot register unknown event", s)
  }
}

function logicalX(e) {
  const scaleX = (canvas.width * canvas.scale.x) / e.target.offsetWidth
  return e.offsetX * scaleX - canvas.offset.x
}

function logicalY(e) {
  const scaleY = (canvas.height * canvas.scale.y) / e.target.offsetHeight
  return e.offsetY * scaleY - canvas.offset.y
}

function keydownListener(e) {
  if (e.target.id == "code") return // skip for source code input
  const { ptr, len } = stringToMem(e.key)
  wasmInst.exports.onKey(ptr, len)
}

const inputQuerySelector = "input#sliderx,input#slidery"

function addInputHandlers() {
  getElements(".input").map((el) => el.classList.remove("hidden"))
  const exp = wasmInst.exports
  for (const el of document.querySelectorAll(inputQuerySelector)) {
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
  for (const el of document.querySelectorAll(inputQuerySelector)) {
    el.onchange = null
  }
  document.removeEventListener("keydown", keydownListener)
}

let animationStart
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

async function initUI() {
  document.addEventListener("keydown", ctrlEnterListener)
  window.addEventListener("hashchange", fetchSource)
  showHideControls()
  fetchSource()
}

function ctrlEnterListener(e) {
  if ((e.metaKey || e.ctrlKey) && event.key === "Enter") {
    handle()
  }
}

async function showHideControls() {
  const opts = parseHash()

  for (const el of getElements(opts.show)) {
    el.classList.remove("hidden")
  }
  for (const el of getElements(opts.hide)) {
    el.classList.add("hidden")
  }
}

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

async function fetchSource() {
  const opts = parseHash()
  if (!opts.source) {
    return
  }
  try {
    const response = await fetch(opts.source)
    if (response.status < 200 || response.status > 299) {
      throw new Error("invalid reponse status", response.status)
    }
    const source = await response.text()
    document.querySelector("#code").value = source
    stop()
    clearOutput()
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
  if (entries.length === 1 && entries[0].length === 1 && entries[0][0]) {
    // shortcut for example.com#draw loading example.com/samples/draw.evy
    return { source: `samples/${entries[0][0]}.evy` }
  }
  return Object.fromEntries(entries)
}

initUI()
initWasm()
initCanvas()
