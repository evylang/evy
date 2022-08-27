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
  if (str.toLowerCase().includes('confetti')) {
    showConfetti()
  }
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

// --------------------------------------------------
// confetti easter egg
// When source input string contains the sub string "confetti"
// show confetti when clicking Run button.
function showConfetti() {
  const names = ['🦊', '🐐']
  const colors = ['red', 'purple', 'blue', 'orange', 'gold', 'green']
  let confetti = new Array(100)
    .fill()
    .map((_, i) => {
      return {
        name: names[i % names.length],
        x: Math.random() * 100,
        y: -20 - Math.random() * 100,
        r: 0.1 + Math.random() * 1,
        color: colors[i % colors.length]
      }
    })
    .sort((a, b) => a.r - b.r)

  const cssText = (c) =>
    `background: ${c.color};left: ${c.x}%; top: ${c.y}%; transform: scale(${c.r})`
  const confettiDivs = confetti.map((c) => {
    const div = document.createElement('div')
    div.style.cssText = cssText(c)
    div.classList.add('confetti')
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
    confettiDivs.forEach((div) => div.classList.add('fadeout'))
  }, 8500)
}
