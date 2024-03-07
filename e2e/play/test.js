import { test, expect } from "@playwright/test"
import config from "../playwright.config.js"

const baseURLs = [`${config.use.baseURL}/play`]
if (process.env.BASEURL_PLAY) {
  baseURLs.push(process.env.BASEURL_PLAY)
}

for (const baseURL of baseURLs) {
  test.describe(baseURL, () => {
    test.use({ baseURL })

    test("title", async ({ page, baseURL }) => {
      await page.goto(baseURL)
      await expect(page).toHaveTitle("evy · Playground")
    })

    test("console-out", async ({ page, baseURL }) => {
      await page.goto(baseURL)
      await page.waitForLoadState("networkidle")
      await page.getByRole("button", { name: "Run" }).click()
      await expect(page.locator("#console")).toContainText("x: 12 🍦 big x")
      await expect(page).toHaveScreenshot("console-output.png")
    })

    test("header-nav", async ({ page, baseURL }) => {
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

    test("sidebar", async ({ page, baseURL }, testInfo) => {
      await page.goto(baseURL)
      await page.waitForLoadState("networkidle")

      // show sidebar
      await page.locator("#hamburger").click()
      await page.getByText("About Evy Docs Discord GitHub").click()
      await expect(page).toHaveScreenshot("sidebar.png")

      // hide sidebar by click on main
      if (testInfo.project.name != "ios") {
        await page.getByRole("main").click()
      } else {
        await page.locator("#sidebar-close").click()
      }
      await expect(page).toHaveScreenshot("no-sidebar.png")

      // show sidebar again
      await page.locator("#hamburger").click()
      await page.getByText("About Evy Docs Discord GitHub").click()
      await expect(page).toHaveScreenshot("sidebar.png")

      // hide sidebar by click on top menu
      if (testInfo.project.name != "ios") {
        await page.getByRole("button", { name: "Welcome" }).click()
        await expect(page).toHaveScreenshot("modal.png")
      }
    })

    test("dialogs", async ({ page, baseURL }, testInfo) => {
      await page.goto(baseURL)
      await page.waitForLoadState("networkidle")
      await expect(page).toHaveScreenshot("no-dialog.png")

      // show sidebar
      if (testInfo.project.name != "ios") {
        await page.locator("#share").getByText("Share").click()
      } else {
        await page.locator("#hamburger").click()
        await page.getByRole("button", { name: "Share" }).click()
      }
      await page.locator('input[type="text"]').click()
      await page.locator('input[type="text"]').press("ArrowRight")
      await expect(page).toHaveScreenshot("share-dialog.png", { maxDiffPixelRatio: 0.01 })
      await page.getByRole("button", { name: "Done" }).click()
      if (testInfo.project.name != "ios") {
        //TODO: there is a rendering bug for this on ios, view snapshot diff, see https://github.com/evylang/todo/issues/50
        await expect(page).toHaveScreenshot("no-dialog.png")
      }
      await page.locator("#hamburger").click()
      await page.getByRole("button", { name: "About Evy" }).click()
      await page.waitForLoadState("networkidle")
      await expect(page).toHaveScreenshot("about-dialog.png", { maxDiffPixelRatio: 0.01 })
      await page.locator("header").filter({ hasText: "About" }).getByRole("button").click()
      if (testInfo.project.name != "ios") {
        //TODO: there is a rendering bug for this on ios, view snapshot diff, see https://github.com/evylang/todo/issues/50
        await expect(page).toHaveScreenshot("no-dialog.png")
      }
    })
  })
}
