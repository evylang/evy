export default function showConfetti() {
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

  const confettiDivs = confetti.map((c) => {
    const div = document.createElement("div")
    Object.assign(div.style, confettiStyle(c))
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
      Object.assign(div.style, { top: "" + c.y + "%" })
      return c
    })
  }

  loop()
  setTimeout(() => {
    cancelAnimationFrame(frame)
    confettiDivs.forEach((div) => div.remove())
  }, 10000)
  setTimeout(() => {
    confettiDivs.forEach((div) =>
      Object.assign(div.style, {
        opacity: 0,
        transition: "opacity 1.5s ease-in-out",
      }),
    )
  }, 8500)
}

function confettiStyle(confetti) {
  return {
    background: confetti.color,
    left: "" + confetti.x + "%",
    top: "" + confetti.y + "%",
    transform: `scale(${confetti.r})`,
    height: "7vw",
    width: "7vw",
    lineHeight: "7vw",
    borderRadius: "50%",
    position: "absolute",
    fontSize: "4vw",
    userSelect: "none",
    textAlign: "center",
  }
}
