"use strict"
import Editor from "./module/editor.js"
import highlightEvy from "./module/highlight.js"
import initThemeToggle from "./module/theme.js"
import showConfetti from "./module/confetti.js"

// --- Globals ---------------------------------------------------------

let wasmModule, wasmInst
const go = newEvyGo()
const canvas = newCanvas()

let jsReadInitialised = false
let stopped = true
let animationStart
let sampleData
let currentSample = "welcome"
let actions = "fmt,ui,eval"
let editor
let errors = false
let editorHidden = false
let notesHidden = true

// --- Initialize ------------------------------------------------------

await Promise.all([initWasm(), initUI()])
await format()

// --- Wasm ------------------------------------------------------------

// initWasm loads byte-code and initializes execution environment.
async function initWasm() {
  wasmModule = await WebAssembly.compileStreaming(fetch(wasmImports["./module/evy.wasm"]))
  const runButton = document.querySelector("#run")
  const runButtonMob = document.querySelector("#run-mobile")
  runButton.onclick = handleRun
  runButton.classList.remove("loading")
  runButtonMob.onclick = handlePrimaryClick
  runButtonMob.classList.remove("loading")
}

function newEvyGo() {
  // evyEnv contains JS functions from this file exported to wasm/go
  const evyEnv = {
    jsPrint,
    jsCls,
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
    stroke,
    fill,
    dash,
    linecap,
    text,
    font,
    gridn,
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
    f.grid ||
    f.gridn ||
    f.poly ||
    f.ellipse ||
    f.stroke ||
    f.fill ||
    f.dash ||
    f.linecap ||
    f.text ||
    f.font ||
    f.down ||
    f.up
  )
}

// jsPrint converts wasmInst memory bytes from ptr to ptr+len to string and
// writes it to the output textarea.
function jsPrint(ptr, len) {
  const s = memToString(ptr, len)
  const output = document.querySelector("#console")
  output.textContent += s
  output.scrollTo({ behavior: "smooth", left: 0, top: output.scrollHeight })
  // 🐣 Show confetti Easter egg if print argument contains literal string "confetti"
  if (s.toLowerCase().includes("confetti")) {
    showConfetti()
  }
}

// jsCls clears output textarea.
function jsCls() {
  document.querySelector("#console").textContent = ""
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
  const re = /line (?<line>\d+) column (?<col>\d+):( runtime error:)? (?<msg>.*)/
  let msgs = ""
  const errorLines = {}
  for (const err of errs) {
    const m = err.match(re)
    if (!m) {
      msgs += err + "\n"
      continue
    }
    const g = m.groups
    if (!errorLines[g.line]) {
      errorLines[g.line] = { col: g.col, text: lines[g.line - 1] }
    }
    msgs += `line ${g.line}: ${g.msg}\n`
  }
  const output = document.querySelector("#console")
  output.textContent += msgs
  output.scrollTo({ behavior: "smooth", left: 0, top: output.scrollHeight })
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

function stringToMemAddr(s) {
  const ptrLen = stringToMem(s)
  return ptrLenToBigInt(ptrLen)
}

function stringToMem(s) {
  if (s === "") {
    // We cannot use `{ ptr: 0, len: 0 }`, encoded into ptrLen 0,
    // because we use 0 as sentinel for "nothing read" in wasm read polling
    // so use any non-0 pointer with length 0 for empty string.
    return { ptr: 1, len: 0 }
  }
  const bytes = new TextEncoder("utf8").encode(s)
  const e = wasmInst.exports
  const len = bytes.length
  const ptr = e.alloc(len)
  const mem = new Uint8Array(e.memory.buffer, ptr, len)
  mem.set(new Uint8Array(bytes))
  return { ptr, len }
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

// handlePrimaryClick handles view states (mobile, code, output) on mobile.
async function handlePrimaryClick() {
  // single column layout: run <-> stop
  if (editorHidden && notesHidden) {
    handleRun()
    return
  }
  const view = getView()
  if (view == "view-notes" && !editorHidden) {
    await slide("view-code")
    toggleNotesButtonVisiblity(true)
    return
  }
  if (view === "view-notes" || view === "view-code") {
    // we need to wait for the slide transition to finish otherwise
    // el.focus() in jsRead() messes up the layout
    await slide("view-output")
    toggleNotesButtonVisiblity(true)
    start()
    return
  }
  // on output view, running
  if (!stopped) {
    stop()
    return
  }
  // on output view, stopped
  document.querySelector("#run-mobile").innerText = "Run"
  const nextScreen = editorHidden ? "view-notes" : "view-code"
  slide(nextScreen)
}

function getView() {
  const cl = document.querySelector("main.main").classList
  if (cl.contains("view-output")) return "view-output"
  if (cl.contains("view-code")) return "view-code"
  if (cl.contains("view-notes")) return "view-notes"
}

function showNotes() {
  if (notesHidden) return
  if (!stopped) stop()
  slide("view-notes")
  toggleNotesButtonVisiblity(false)
}

function toggleNotesButtonVisiblity(show) {
  const showNotesBtn = document.querySelector("#show-notes")
  if (!showNotesBtn) return
  const hamburgerBtn = document.querySelector("#hamburger")
  if (!notesHidden && show) {
    showNotesBtn.classList.remove("hidden")
    hamburgerBtn.classList.add("hidden")
    return
  }
  showNotesBtn.classList.add("hidden")
  hamburgerBtn.classList.remove("hidden")
}

// start calls evy wasm/go main(). It parses, formats and evaluates evy
// code and initialize the output ui.
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
  runButton.classList.remove("running")
  runButton.innerText = "Run"
  updateMobilePrimaryButton()

  const readEl = document.querySelector("#read")
  document.activeElement === readEl && readEl.blur()
}

function updateMobilePrimaryButton() {
  const classList = document.querySelector("#run-mobile")
  classList.classList.remove("running")
  classList.innerText = mobilePrimaryButtonText()
}

function mobilePrimaryButtonText() {
  if (editorHidden && notesHidden) return "Run"
  const view = getView()
  if (view === "view-notes" && !editorHidden) return "Code"
  if (view === "view-notes" && editorHidden) return "Run"
  if (view === "view-code") return "Run"
  // output screen
  if (editorHidden) return "Notes"
  return "Code"
}

async function slide(view) {
  const el = document.querySelector("main.main")
  const cl = el.classList
  return new Promise((resolve) => {
    el.ontransitionend = () => setTimeout(resolve, 100)
    el.onanimationend = () => cl.remove("animate")
    cl.add("animate")
    setView(view)
  })
}

function setView(view) {
  const cl = document.querySelector("main.main").classList
  const viewClasses = ["view-code", "view-notes", "view-output"]
  viewClasses.map((c) => cl.remove(c))
  cl.add(view)
  updateMobilePrimaryButton()
}

function clearOutput() {
  document.querySelector("#console").textContent = ""
  resetCanvas()
}

// --- UI: initialization ----------------------------------------------

async function initUI() {
  initCanvas()
  initThemeToggle("#dark-theme", "theme")
  document.addEventListener("keydown", ctrlEnterListener)
  window.addEventListener("hashchange", handleHashChange)
  document.querySelector("#modal-close").onclick = hideModal
  document.querySelector("#sidebar-about").onclick = showAbout
  document.querySelector("#sidebar-share").onclick = share
  document.querySelector("#sidebar-icon-share").onclick = share
  const shareBtn = document.querySelector("#share")
  if (shareBtn) shareBtn.onclick = share
  const showNotesBtn = document.querySelector("#show-notes")
  if (showNotesBtn) showNotesBtn.onclick = showNotes
  await fetchSamples()
  await handleHashChangeNoFormat() // Evy wasm for formatting might not be ready yet
  initModal()
  initSidebar()
  initShareDialog()
}

async function fetchSamples() {
  const resp = await fetch("samples/samples.json")
  sampleData = await resp.json()
  sampleData.byID = {}
  let previous = null
  for (const section of sampleData.sections) {
    const listedSamples = section.samples.filter((s) => !s.unlisted)
    const sectionTotal = listedSamples.length
    let index = 1
    for (const sample of section.samples) {
      const sampleByID = {
        ...sample,
        sectionTitle: section.title,
        sectionID: section.id,
      }
      sampleData.byID[sample.id] = sampleByID
      if (sample.unlisted) {
        continue
      }
      sampleByID.sectionTotal = sectionTotal
      sampleByID.sectionIndex = index
      sampleByID.previous = previous
      if (previous) {
        sampleData.byID[previous].next = sample.id
      }
      previous = sample.id
      index++
    }
  }
}

function ctrlEnterListener(e) {
  if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
    document.querySelector(".editor textarea").blur()
    handleRun()
  }
}

function escListener(e) {
  if (e.key === "Escape") {
    hideModal()
    hideSidebar()
  }
}

// --- UI: URL-hash change handling ------------------------------------

// handleHashChange is triggered by browser back/forwards buttons, manual
// address bar update or click on link e.g. #new-example.
// It first resets UI and waits for all previous actions to finish.
// Then, it loads new source code depending on URL hash contents.
// Finally it updates editor.
async function handleHashChange() {
  await handleHashChangeNoFormat()
  await format()
}

async function handleHashChangeNoFormat() {
  hideModal()
  await stop() // go to code screen for new code
  let opts = parseHash()
  if (!opts.source && !opts.sample && !opts.content) {
    if (sessionStorage.getItem("evy-editor") !== null) {
      loadSession()
      return
    }
    const sample = "welcome"
    opts = { sample, editor: sampleData?.byID[sample]?.editor }
    history.replaceState({}, "", "#" + sample)
  }
  const { source, notes } = await fetchSourceWithNotes(opts)
  updateNotes(notes)
  updateEditor(source, opts)
  updateSampleTitle()
  resetView()
  clearOutput()
}

function loadSession() {
  currentSample = "<UNSET>"
  !editor && initEditor()
  editor.loadSession()
  loadNotes()
  toggleEditorVisibility(true)
  resetView() // on mobile go to Notes view or Code view if no notes available.
}

// resetView resets the view based on content availability, on mobile only.
// If notes are available, navigates to the notes view.
// If no notes but the editor is present, switches to the code view.
// Otherwise, defaults to the output view
async function resetView() {
  const mainClassList = document.querySelector("main.main").classList
  mainClassList.add("no-translate-transition")
  toggleNotesButtonVisiblity(false)
  if (!notesHidden) {
    setView("view-notes")
  } else if (!editorHidden) {
    setView("view-code")
  } else {
    setView("view-output")
  }
  // sleep for 0.3 seconds, otherwise translate transition kicks in.
  await new Promise((r) => setTimeout(r, 300))
  mainClassList.remove("no-translate-transition")
}

function updateNotes(notes) {
  hasNotes(notes) ? addNotes(notes) : removeNotes()
}

function hasNotes(notes) {
  return !!notes && !!document.querySelector("#notes")
}

function removeNotes() {
  notesHidden = true
  const notesEl = document.querySelector("#notes")
  if (!notesEl) return
  notesEl.classList.add("hidden")
  notesEl.innerHTML = ""
  sessionStorage.removeItem("evy-sample-id")
}

function addNotes(notes) {
  notesHidden = false
  const notesEl = document.querySelector("#notes")
  notesEl.classList.remove("hidden")
  notesEl.innerHTML = notes
  sessionStorage.setItem("evy-sample-id", currentSample)

  // hide all notes after first "next" button:
  // <p><button class="next-btn">Next</button></p>
  let el = notesEl.querySelector(".next-btn")?.parentElement?.nextElementSibling
  while (el) {
    el.classList.add("hidden")
    el = el.nextElementSibling
  }
  notesEl.querySelectorAll(".next-btn").forEach((btn) => {
    btn.onclick = handleNotesNextClick
  })
  notesEl.querySelectorAll(".language-evy").forEach((el) => {
    el.innerHTML = highlightEvy(el.textContent)
  })
  notesEl.querySelectorAll("a").forEach((el) => {
    el.target = "_blank"
  })
  notesEl.querySelectorAll("img[title='evy:edit']").forEach((img) => {
    img.onclick = handleNotesImgEditClick
    img.title = "Click to edit"
    img.style.cursor = "pointer"
  })
  notesEl.scrollTo(0, 0)
}

async function loadNotes() {
  const sampleID = sessionStorage.getItem("evy-sample-id")
  const sample = sampleData.byID[sampleID]
  if (!sample?.notes) {
    removeNotes()
    return
  }
  currentSample = sampleID

  const notesURL = `samples/${sample.sectionID}/${sample.id}.htmlf`
  const notes = await fetchText(notesURL)
  addNotes(notes)
  toggleEditorVisibility(sample.editor !== "none")
}

function handleNotesNextClick(e) {
  const btn = e.target
  let el = btn?.parentElement?.nextElementSibling
  // show until following "next" button or end
  while (el && !el.classList.contains("next-btn")) {
    el.classList.remove("hidden")
    if (el.querySelector(".next-btn")) break
    el = el.nextElementSibling
  }
  const top = btn.offsetTop + btn.offsetHeight
  document.querySelector("#notes").scrollTo({ top, behavior: "smooth" })
}

async function handleNotesImgEditClick(e) {
  const img = e.target
  const url = img.src.replace(".svg", ".evy")
  const evyImgSource = await fetchText(url)
  editor.update({ value: evyImgSource, errorLines: {} })
}

function updateEditor(content, opts) {
  !editor && initEditor()
  editor.onUpdate(null)
  editor.update({ value: content, errorLines: {} })
  document.querySelector(".editor-wrap").scrollTo(0, 0)
  editor.onUpdate(clearHash)
  toggleEditorVisibility(opts.editor !== "none")
}

function toggleEditorVisibility(isVisible) {
  editorHidden = !isVisible
  const classList = document.querySelector(".editor-wrap").classList
  isVisible ? classList.remove("hidden") : classList.add("hidden")
}

// parseHash parses URL fragment into object e.g.:
//
//    https://evy.dev#key1=v1&key2=v2  →
//    { key1: "v1", key2: "v2" }
//
// so, `&` separates key-value entries and `=` separates keys from values,
// just like in a query string. There is a shortcut to known evy samples:
//
//    #abc   →
//    { sample: "abc" }
function parseHash() {
  const strs = window.location.hash.substring(1).split("&") //  ["a=1", "b=2"]
  const entries = strs.map((s) => s.split("=")) // [["a", "1"], ["b", "2"]]
  if (entries.length === 1 && entries[0].length === 1) {
    // shortcut for evy.dev#abc loading evy.dev/samples/draw/abc.evy
    const sample = entries[0][0]
    if (sampleData && sampleData.byID[sample]) {
      return { sample, editor: sampleData.byID[sample].editor }
    }
  }
  return Object.fromEntries(entries)
}

async function fetchSourceWithNotes({ content, sample, source }) {
  if (sample) {
    const s = sampleData.byID[sample]
    currentSample = sample
    return await fetchSample(s)
  }
  currentSample = "<UNSET>"
  const src = await (content ? decode(content) : fetchText(source))
  return { source: src }
}

async function fetchSample(sample) {
  const evyURL = `samples/${sample.sectionID}/${sample.id}.evy`
  if (!sample.notes) {
    const source = await fetchText(evyURL)
    return { source }
  }
  const notesURL = sample.notes && `samples/${sample.sectionID}/${sample.id}.htmlf`
  const [source, notes] = await Promise.all([fetchText(evyURL), fetchText(notesURL)])
  return { source, notes }
}

async function fetchText(url) {
  let text
  try {
    const response = await fetch(url)
    if (response.status < 200 || response.status > 299) {
      throw new Error("invalid response status", response.status)
    }
    text = await response.text()
  } catch (err) {
    console.error(err, url)
    text = `Oops! Could not load sample.`
  }
  return text
}

function clearHash() {
  history.pushState({}, "", window.location.origin + window.location.pathname)
  // Clear hash only on first edit
  editor.onUpdate(null)
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
  clear(0, 0)
  const ctx = canvas.ctx
  ctx.fillStyle = "black"
  ctx.strokeStyle = "black"
  ctx.lineWidth = 1
  canvas.fill = true
  canvas.stroke = true
  ctx.lineCap = "round"
  ctx.setLineDash([])
  ctx.font = `${(ctx.canvas.width / 100) * 6}px "Fira Code", monospace`
  ctx.textAlign = "left"
  ctx.textBaseline = "alphabetic"
  ctx.letterSpacing = "0px"
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
  let color = "white"
  if (len !== 0) {
    color = memToString(ptr, len)
  }
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
    endAngle * rad,
  )
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

// gridn is exported to evy go/wasm.
function gridn(unit, ptr, len) {
  const { ctx, x, y } = canvas
  const restoreLineWidth = ctx.lineWidth
  const restoreStrokeStyle = ctx.strokeStyle
  const restoreLineDash = ctx.getLineDash()
  ctx.strokeStyle = memToString(ptr, len)
  ctx.setLineDash([])
  let lineCnt = 0
  for (let i = 0; i <= 100; i += unit) {
    ctx.lineWidth = lineCnt % 5 === 0 ? 2 : 1
    lineCnt += 1
    move(i, 0)
    line(i, 100)
    move(0, i)
    line(100, i)
  }
  ctx.lineWidth = restoreLineWidth
  ctx.strokeStyle = restoreStrokeStyle
  ctx.setLineDash(restoreLineDash)
  canvas.x = x
  canvas.y = y
}

var parsedStyle = function (cssString) {
  let el = document.createElement("span")
  el.setAttribute("style", cssString)
  return el.style // CSSStyleDeclaration object
}

function textsize(size) {
  const { width, ctx } = canvas
  const style = parsedStyle(`font: ${ctx.font}`)
  style.fontSize = (ctx.canvas.width / 100) * size + "px"
  ctx.font = style.font
}

// font is exported to evy go/wasm.
// see https://developer.mozilla.org/en-US/docs/Web/CSS/font
//
// Exhaustive example of accepted properties encoded as JSON:
//
//    {
//      "family": "Georgia, serif",
//      "size": 3, // relative to canvas, numbers only no "12px" etc.
//      "weight": 100, //| 200| 300 | 400 == "normal" | 500 | 600 | 700 == "bold" | 800 | 900
//      "style": "italic", | "oblique 35deg" | "normal"
//      "baseline": "top", // | "middle" | "bottom"
//      "align": "left", // | "center" | "right"
//      "letterspacing": 1 // number, see size. extra inter-character space. negative allowed.
//    }
function font(ptr, len) {
  const propsJSON = memToString(ptr, len)
  const props = JSON.parse(propsJSON)
  const ctx = canvas.ctx
  const style = parsedStyle(`font: ${ctx.font}`)
  if (props.family !== undefined) {
    style.fontFamily = props.family
  }
  if (props.size !== undefined) {
    style.fontSize = (ctx.canvas.width / 100) * props.size + "px"
  }
  if (props.weight !== undefined) {
    style.fontWeight = props.weight
  }
  if (props.style !== undefined) {
    style.fontStyle = props.style
  }
  if (props.baseline !== undefined) {
    ctx.textBaseline = props.baseline
  }
  if (props.align !== undefined) {
    ctx.textAlign = props.align
  }
  if (props.letterspacing !== undefined) {
    ctx.letterSpacing = (ctx.canvas.width / 100) * props.letterspacing + "px"
  }
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

function leaveXY(e) {
  const x = clamp(logicalX(e), 0, 100)
  const y = clamp(logicalY(e), 0, 100)
  const dx100 = 100 - x
  const dx0 = x
  const dy100 = 100 - y
  const dy0 = y
  const min = Math.min(dx100, dx0, dy100, dy0)
  if (min === dx100) return [100, y]
  if (min === dx0) return [0, y]
  if (min === dy100) return [x, 100]
  return [x, 0]
}

function clamp(val, min, max) {
  return Math.min(Math.max(val, min), max)
}

function initEditor() {
  const options = {
    highlighter: highlightEvy,
    id: "evy-editor",
    sessionKey: "evy-editor",
  }
  editor = new Editor(".editor", options)
  document.querySelector(".editor-wrap").classList.remove("noscrollbar")
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
    c.onmouseleave = (e) => exp.onMove(...leaveXY(e)) // pointer can leave in middle of canvas
  } else if (s === "key") {
    unfocusRunButton()
    document.addEventListener("keydown", keydownListener)
  } else if (s === "input") {
    addInputHandlers()
  } else if (s === "animate") {
    window.requestAnimationFrame(animationLoop)
  } else {
    console.error("cannot register unknown event", s)
  }
}

function unfocusRunButton() {
  const runButton = document.querySelector("#run")
  const runButtonMob = document.querySelector("#run-mobile")
  document.activeElement === runButton && runButton.blur()
  document.activeElement === runButtonMob && runButtonMob.blur()
}

function keydownListener(e) {
  if (e.target.id == "evy-editor") return // skip for source code input
  document.querySelector(".output").focus()
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
  for (const section of sampleData.sections) {
    const sectionEl = document.createElement("div")
    sectionEl.classList.add("section")
    const h2 = document.createElement("h2")
    h2.textContent = `${section.emoji} ${section.title}`
    const ul = document.createElement("ul")
    sectionEl.replaceChildren(h2, ul)
    for (const sample of section.samples) {
      if (sample.unlisted) {
        continue
      }
      const li = document.createElement("li")
      const a = document.createElement("a")
      a.textContent = sample.title
      a.href = `#${sample.id}`
      a.onclick = hideModal
      li.appendChild(a)
      ul.appendChild(li)
    }
    modalMain.appendChild(sectionEl)
  }
  document.querySelector("#sample-selector").onclick = showSamples
  document.querySelector("#sample-previous").onclick = showPreviousSample
  document.querySelector("#sample-next").onclick = showNextSample
  updateSampleTitle()
}

function hideModal() {
  const el = document.querySelector("#modal")
  el.classList.add("hidden")
  document.removeEventListener("keydown", escListener)
}

function showSamples() {
  const samples = document.querySelector("#modal-samples")
  samples.classList.remove("hidden")
  const modal = document.querySelector("#modal")
  modal.classList.remove("hidden")
  samples.querySelectorAll("a").forEach((a) => a.classList.remove("highlight"))
  samples.querySelector(`a[href$="#${currentSample}"]`)?.classList.add("highlight")
  document.addEventListener("keydown", escListener)
}

function showPreviousSample() {
  if (sampleData.byID[currentSample].previous) {
    currentSample = sampleData.byID[currentSample].previous
    history.pushState({}, "", `#${currentSample}`)
    handleHashChange()
  }
}

function showNextSample() {
  if (sampleData.byID[currentSample].next) {
    currentSample = sampleData.byID[currentSample].next
    history.pushState({}, "", `#${currentSample}`)
    handleHashChange()
  }
}

function updateSampleTitle() {
  const titleDiv = document.querySelector("#sample-title")
  const indexDiv = document.querySelector("#sample-index")
  const prevButton = document.querySelector("#sample-previous")
  const nextButton = document.querySelector("#sample-next")

  const sample = sampleData.byID[currentSample]
  titleDiv.textContent = sample?.title || sampleData.defaultTitle
  if (!sample || sample.unlisted) {
    indexDiv.classList.add("hidden")
    prevButton.disabled = true
    nextButton.disabled = true
    return
  }
  indexDiv.textContent = `${sample.sectionIndex}/${sample.sectionTotal}`
  indexDiv.classList.remove("hidden")
  prevButton.disabled = !sample.previous
  nextButton.disabled = !sample.next
}

// --- UI: sidebar --------------------------------------------

function initSidebar() {
  document.querySelector("#hamburger").onclick = showSidebar
  document.querySelector("#sidebar-close").onclick = hideSidebar
}

function showSidebar() {
  document.querySelector(".editor textarea").style.pointerEvents = "none"
  document.querySelector("#sidebar").classList.remove("hidden")
  document.addEventListener("click", handleOutsideSidebarClick)
  document.addEventListener("keydown", escListener)
}
function hideSidebar() {
  document.querySelector(".editor textarea").style.pointerEvents = ""
  document.querySelector("#sidebar").classList.add("hidden")
  document.removeEventListener("click", handleOutsideSidebarClick)
  document.removeEventListener("keydown", escListener)
}
function handleOutsideSidebarClick(e) {
  const sidebar = document.querySelector("#sidebar")
  if (!sidebar.classList.contains("hidden") && e.pageX > sidebar.offsetWidth) {
    hideSidebar()
  }
}

// --- UI: dialog --------------------------------------------

function initShareDialog() {
  const shareDialog = document.querySelector("#dialog-share")
  const input = shareDialog.querySelector(".copy input")
  input.onclick = input.select
  const closeButton = shareDialog.querySelector(".icon-close")
  closeButton.onclick = () => shareDialog.close()
  const copyButton = shareDialog.querySelector("#copy")
  copyButton.onclick = () => {
    const url = input.value
    navigator.clipboard.writeText(url)
    input.value = "Copied!"
    setTimeout(() => shareDialog.close(), 500)
  }
}

function showAbout() {
  const about = document.querySelector("#dialog-about")
  hideSidebar()
  about.showModal()
}

// --- Share / load snippets -------------------------------------------

async function share() {
  hideSidebar()
  const note = document.querySelector("#dialog-share .dialog-note")
  await format()
  errors ? note.classList.remove("hidden") : note.classList.add("hidden")
  const baseurl = window.location.origin + window.location.pathname
  const encoded = await encode(editor.value)
  const input = document.querySelector("#dialog-share .copy input")
  input.value = `${baseurl}#content=${encoded}`
  input.setSelectionRange(0, 0)
  input.blur()
  document.querySelector("#dialog-share").showModal()
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
    console.error("getElements", error)
    return []
  }
}

function showElements(q) {
  getElements(q).map((el) => el.classList.remove("hidden"))
}

function hideElements(q) {
  getElements(q).map((el) => el.classList.add("hidden"))
}
