const { test, expect } = require("@playwright/test")

test("title", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await expect(page).toHaveTitle("evy | Playground")
})

test("console output", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await page.waitForLoadState("networkidle")
  await page.getByRole("button", { name: "Run" }).click()
  const console = page.locator("css=#console")
  await expect(console).toContainText("x: 12")
  await expect(console).toContainText("🍦 big x")
  await expect(page).toHaveScreenshot("console-output.png")
})

test("header navigation", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await page.waitForLoadState("networkidle")
  const modal = page.locator("css=#modal")
  await expect(modal).toBeHidden()
  await page.getByRole("button", { name: "Welcome" }).click()
  await expect(modal).toBeVisible()
  await expect(modal).toContainText("🚌 Tour")
  await expect(page).toHaveScreenshot("modal.png")

  await page.getByRole("link", { name: "Coordinates" }).click()
  await expect(modal).toBeHidden()

  const editor = page.locator("css=.editor")
  await expect(editor).toContainText("on move x:num y:num")
})
