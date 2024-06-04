const defaultOptions = {
  duration: 10, // animation seconds
  fadeoutAfter: 8, // fadeout start in seconds
  count: 100, // confetti count
  texts: ["🦊", "🐐"],
  colors: ["red", "purple", "blue", "orange", "gold", "green"],
}

export default function showConfetti(options = {}) {
  options = { ...defaultOptions, ...options }
  const confettis = Array(options.count)
  for (let i = 0; i < options.count; i++) {
    confettis[i] = newConfetti(options)
  }
  const divs = confettis.map(newDiv)
  divs.forEach((div) => document.body.appendChild(div))
  animate(confettis, divs, options)
}

function newConfetti(options) {
  const { texts, colors } = options
  return {
    text: texts[Math.floor(Math.random() * texts.length)],
    x: Math.random() * 100,
    y: Math.random() * 100,
    r: Math.random() + 0.1, // scale factor, see top()
    color: colors[Math.floor(Math.random() * colors.length)],
  }
}

function newDiv(confetti) {
  const baseStyle = {
    height: "14vh",
    width: "14vh",
    lineHeight: "14vh",
    borderRadius: "50%",
    position: "absolute",
    fontSize: "8vh",
    userSelect: "none",
    textAlign: "center",
  }
  const confettiStyle = {
    background: confetti.color,
    top: `${top(confetti.y, 0)}%`,
    // left property offsets the center of the confetti with max radius 7vh.
    left: `calc(${confetti.x}vw - 7vh)`,
    transform: `scale(${confetti.r})`,
  }

  const div = document.createElement("div")
  div.textContent = confetti.text
  Object.assign(div.style, baseStyle, confettiStyle)
  return div
}

function animate(confettis, divs, options) {
  const fadeoutStyle = newFadeoutStyle(options)
  let fading = false
  const start = document.timeline.currentTime

  requestAnimationFrame(onFrame)

  function onFrame(ts) {
    const elapsed = (ts - start) / 1000 // elapsed seconds
    if (elapsed > options.duration) {
      // animation done
      divs.forEach((div) => div.remove())
      return
    }
    // update offset from top
    for (let i = 0; i < divs.length; i++) {
      const style = { top: `${top(confettis[i], elapsed)}%` }
      Object.assign(divs[i].style, style)
    }
    if (elapsed > options.fadeoutAfter && !fading) {
      // add fadeout style
      fading = true
      divs.forEach((div) => Object.assign(div.style, fadeoutStyle))
    }
    requestAnimationFrame(onFrame)
  }
}

function newFadeoutStyle(options) {
  const transitionDur = options.duration - options.fadeoutAfter
  return {
    opacity: 0,
    transition: `opacity ${transitionDur}s ease-in-out`,
  }
}

// top returns offset from top of viewport.
function top(confetti, elapsed) {
  // r is the scale factor [0.1,1.1). It scales down confetti size and delta y.
  const r = confetti.r
  // y is the initial position. It is [-120, 20) above viewport.
  const maxDiameter = 14 // 14vh
  const yInitial = -confetti.y - maxDiameter
  // yElapsed is the position after elapsed seconds.
  const yElapsed = yInitial + elapsed * 50 * r
  // When yElapsed is below the viewport start over from the top.
  return (yElapsed % (100 + maxDiameter)) - maxDiameter
}
