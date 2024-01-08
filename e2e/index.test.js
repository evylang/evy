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
})
