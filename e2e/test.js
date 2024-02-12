// Landing page tests
import { test, expect } from "@playwright/test"
import config from "./playwright.config.js"

test("title", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await expect(page).toHaveTitle("evy · Intro")
})

test("landing", async ({ page, baseURL }, testInfo) => {
  await page.goto(baseURL)
  await expect(page).toHaveScreenshot("landing-top.png")

  if (testInfo.project.name != "ios") {
    // let's make sure we don't take snapshots of the gif
    // Set height to a small value and scroll to bottom
    await page.setViewportSize({ width: 1280, height: 250 })
    // Programmatic scrolling does not work on mobile / ios in playwright.
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
    await expect(page).toHaveScreenshot("landing-bottom.png")
    // reset to default size
    await page.setViewportSize({ width: 1280, height: 720 })
  }
  await page.waitForLoadState("networkidle")
  await page.getByRole("link", { name: "Try It Out" }).click()
  await page.waitForLoadState("networkidle")

  await expect(page).toHaveTitle("evy · Playground")
  await expect(page).toHaveScreenshot("playground.png")
})
