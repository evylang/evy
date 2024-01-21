const { test, expect } = require("@playwright/test")

test("title", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await expect(page).toHaveTitle("evy | Playground")
})

test("console output", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await page.waitForLoadState("networkidle")
  await page.getByRole("button", { name: "Run" }).click()
  await expect(page.locator("#console")).toContainText("x: 12 🍦 big x")
  await expect(page).toHaveScreenshot("console-output.png")
})

test("header navigation", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await page.waitForLoadState("networkidle")
  const modal = page.locator("#modal")
  await expect(modal).toBeHidden()

  await page.getByRole("button", { name: "Welcome" }).click()
  await expect(modal).toBeVisible()
  await expect(modal).toContainText("🚌 Tour")
  await expect(page).toHaveScreenshot("modal.png")

  await page.getByRole("link", { name: "Coordinates" }).click()
  await expect(modal).toBeHidden()
  await expect(page.getByRole("textbox")).toHaveValue(
    `grid
print "Move mouse or touch to print coordinates"

on move x:num y:num
    print "x:" (round x) "y:" (round y)
end
`,
  )
})

test("side menu", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await page.waitForLoadState("networkidle")

  // show side menu
  await page.locator("#hamburger").click()
  await page.getByText("About Evy Docs Discord GitHub").click()
  await expect(page).toHaveScreenshot("sidemenu.png")

  // hide side menu by click on main
  await page.getByRole("main").click()
  await expect(page).toHaveScreenshot("no-sidemenu.png")

  // show side menu again
  await page.locator("#hamburger").click()
  await page.getByText("About Evy Docs Discord GitHub").click()
  await expect(page).toHaveScreenshot("sidemenu.png")

  // hide side menu by click on top menu
  await page.getByRole("button", { name: "Welcome" }).click()
  await expect(page).toHaveScreenshot("modal.png")
})

test("dialogs", async ({ page, baseURL }) => {
  await page.goto(baseURL)
  await page.waitForLoadState("networkidle")
  await expect(page).toHaveScreenshot("no-dialog.png")

  // show side menu
  await page.locator("#share").getByText("Share").click()
  await page.locator('input[type="text"]').click()
  await page.locator('input[type="text"]').press("ArrowRight")
  await expect(page).toHaveScreenshot("share-dialog.png")
  await page.getByRole("button", { name: "Done" }).click()
  await expect(page).toHaveScreenshot("no-dialog.png")
  await page.locator("#hamburger").click()
  await page.getByRole("button", { name: "About Evy" }).click()
  await expect(page).toHaveScreenshot("about-dialog.png")
  await page.locator("header").filter({ hasText: "About" }).getByRole("button").click()
  await expect(page).toHaveScreenshot("no-dialog.png")
})
