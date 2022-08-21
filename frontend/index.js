'use strict'

var wasm

// initWasm loads bytecode and initialises execution environment.
function initWasm() {
  const go = new Go() // see wasm_exec.js
  WebAssembly.instantiateStreaming(fetch('evy.wasm'), go.importObject).then(function (obj) {
    wasm = obj.instance
    go.run(wasm)
  })
  go.importObject.env = { 'main.jsPrint': jsPrint }
  const button = document.getElementById('run')
  button.onclick = handleRun
  button.disabled = false
}

// jsPrint converts wasm memory bytes from ptr to ptr+len to string and
// writes it the output pane.
function jsPrint(ptr, len) {
  const buf = new Uint8Array(wasm.exports.memory.buffer, ptr, len)
  const str = new TextDecoder('utf8').decode(buf)
  const output = document.getElementById('output')
  output.textContent += str
}

// handleRun retrieves the input string from the source pane and
// converts it to wasm memory bytes. It then calls the evy evaluate
// function.
function handleRun() {
  const source = document.getElementById('source').textContent
  const bytes = new TextEncoder('utf8').encode(source)
  const ptr = wasm.exports.alloc(bytes.length)
  const mem = new Uint8Array(wasm.exports.memory.buffer, ptr, bytes.length)
  mem.set(new Uint8Array(bytes))
  document.getElementById('output').textContent = ''
  wasm.exports.evaluate(ptr, bytes.length)

  // Debug switch: set `tokenize = true` globally, e.g. in developer
  // console, to print tokens.
  window.tokenize && wasm.exports.tokenize(ptr, bytes.length)
  window.parse && wasm.exports.parse(ptr, bytes.length)
}

initWasm()
