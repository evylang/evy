"use strict"

let wasmModule, wasmInst
let sourcePtr, sourceLength
const go = newEvyGo() // see wasm_exec.js

// initWasm loads bytecode and initialises execution environment.
function initWasm() {
  WebAssembly.compileStreaming(fetch("evy.wasm"))
    .then((obj) => (wasmModule = obj))
    .catch((err) => console.error(err))
  const runButton = document.getElementById("run")
  runButton.onclick = handleRun
  runButton.disabled = false
}

// jsPrint converts wasmInst memory bytes from ptr to ptr+len to string and
// writes it to the output textarea.
function jsPrint(ptr, len) {
  const s = memString(ptr, len)
  const output = document.getElementById("output")
  output.textContent += s
  if (s.toLowerCase().includes("confetti")) {
    showConfetti()
  }
}

function memString(ptr, len) {
  const buf = new Uint8Array(wasmInst.exports.memory.buffer, ptr, len)
  const s = new TextDecoder("utf8").decode(buf)
  return s
}

// handleRun retrieves the input string from the code pane and
// converts it to wasm memory bytes. It then calls the evy main()
// function running the evaluator after parsing.
async function handleRun(event) {
  wasmInst = await WebAssembly.instantiate(wasmModule, go.importObject)
  prepareSourceAccess()
  clearOutput()
  go.run(wasmInst)
}

function newEvyGo() {
  const evyEnv = {
    jsPrint,
    move,
    line,
    width,
    circle,
    rect,
    color,
    sourcePtr: () => sourcePtr,
    sourceLength: () => sourceLength,
  }
  const go = new Go() // see wasm_exec.js
  go.importObject.env = Object.assign(go.importObject.env, evyEnv)
  return go
}

function prepareSourceAccess() {
  const code = document.getElementById("code").value
  const bytes = new TextEncoder("utf8").encode(code)
  const e = wasmInst.exports
  const ptr = e.alloc(bytes.length)
  const mem = new Uint8Array(e.memory.buffer, ptr, bytes.length)
  mem.set(new Uint8Array(bytes))
  sourcePtr = ptr
  sourceLength = bytes.length
}

function clearOutput() {
  document.getElementById("output").textContent = ""
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
  scale: { x: 10, y: -10 },
  width: 100,
  height: 100,

  offset: { x: 0, y: -100 }, // height
}

function initCanvas() {
  const c = document.getElementById("canvas")
  const b = c.parentElement.getBoundingClientRect()
  c.width = Math.abs(scaleX(canvas.width))
  c.height = Math.abs(scaleY(canvas.height))
  c.style.width = `40vh`
  c.style.height = `40vh`
  c.style.display = "block"
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
  return canvas.scale.x * x
}

function scaleY(y) {
  return canvas.scale.y * y
}

function transformX(x) {
  return scaleX(x + canvas.offset.x)
}

function transformY(y) {
  return scaleY(y + canvas.offset.y)
}

function move(x, y) {
  movePhysical(transformX(x), transformY(y))
}

function movePhysical(px, py) {
  canvas.x = px
  canvas.y = py
}

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

function color(ptr, len) {
  const s = memString(ptr, len)
  canvas.ctx.fillStyle = s
  canvas.ctx.strokeStyle = s
}

function width(n) {
  canvas.ctx.lineWidth = scaleX(n)
}

function rect(dx, dy) {
  const { ctx, x, y } = canvas
  const sDX = scaleX(dx)
  const sDY = scaleY(dy)
  canvas.ctx.fillRect(x, y, sDX, sDY)
  movePhysical(x + sDX, y + sDY)
}

function circle(r) {
  const { x, y, ctx } = canvas
  ctx.beginPath()
  ctx.arc(x, y, scaleX(r), 0, Math.PI * 2, true)
  ctx.fill()
}

initWasm()
initCanvas()
