import { test, expect } from "@playwright/test"
import config from "../playwright.config.js"

const baseURLs = [`${config.use.baseURL}/docs`]
if (process.env.BASEURL_DOCS) {
  baseURLs.push(process.env.BASEURL_DOCS)
}

for (const baseURL of baseURLs) {
  test.describe(baseURL, () => {
    test.use({ baseURL })

    test("docs title", async ({ page, baseURL }) => {
      await page.goto(baseURL)
      await expect(page).toHaveTitle("evy docs · Documentation")
    })

    test("docs about", async ({ page, baseURL }, testInfo) => {
      await page.goto(baseURL)
      await expect(page).toHaveScreenshot("start.png")
      if (page.viewportSize().width <= 767) {
        return // no theme toggle on mobile
      }

      await page.getByRole("button", { name: "About" }).click()
      await expect(page).toHaveScreenshot("dialog.png")

      await page.getByRole("button", { name: "Done" }).click()
      await expect(page).toHaveScreenshot("no-dialog.png")

      await page.locator("label.theme.switch").click()
      await expect(page).toHaveScreenshot("start-theme.png")

      await page.getByRole("button", { name: "About" }).click()
      await expect(page).toHaveScreenshot("dialog-theme.png")
    })

    test("docs md", async ({ page, baseURL }) => {
      await page.goto(baseURL + "/builtins.html#printf")
      await expect(page).toHaveTitle("evy docs · Built-ins")
      await expect(page).toHaveScreenshot("printf.png")
      if (page.viewportSize().width <= 767) {
        return // no theme toggle on mobile
      }
      await page.locator("label.theme.switch").click()
      await expect(page).toHaveScreenshot("printf-theme.png")
    })

    test("docs sidebar", async ({ page, baseURL }) => {
      await page.goto(baseURL)
      await expect(page).toHaveScreenshot("start.png")
      const sreenshotOpts = {}
      if (page.viewportSize().width < 750) {
        sreenshotOpts.maxDiffPixelRatio = 0.01
        await page.locator("#hamburger").click()
        await expect(page).toHaveScreenshot("sidebar.png", sreenshotOpts)
      }
      await page.locator(".expander").nth(0).click()
      await expect(page).toHaveScreenshot("expand-0.png", sreenshotOpts)
      await page.locator(".expander").nth(1).click()
      await expect(page).toHaveScreenshot("expand-1.png", sreenshotOpts)

      const link = await page.getByRole("link", { name: "Comment" })
      await link.hover()
      await expect(page).toHaveScreenshot("comment-hover.png", sreenshotOpts)
      await link.click()
      await page.waitForLoadState("networkidle")
      await expect(page).toHaveScreenshot("comment-page.png", sreenshotOpts)
    })

    test("docs crawl", async ({ page, baseURL }) => {
      const seenURLs = new Set()
      const crawl = async (url) => {
        if (seenURLs.has(url)) {
          return
        }
        seenURLs.add(url)
        if (!url.startsWith(baseURL)) {
          return
        }
        if (url.endsWith(".pdf")) {
          return
        }
        await page.goto(url)
        await page.locator(".topnav.doc")
        const urls = await page.$$eval("a", (elements) => elements.map((el) => el.href))
        for await (const u of urls) {
          await crawl(u)
        }
      }
      await crawl(baseURL)
    })
  })
}
